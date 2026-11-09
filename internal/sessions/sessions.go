// Package sessions owns in-memory management sessions, authentication attempt
// windows and the password-work capacity shared by login and invitation use.
package sessions

import (
	"crypto/sha256"
	"sync"
	"time"
)

// Session is a value snapshot. Changing it does not change the stored session.
type Session struct {
	CSRF   string
	Expiry time.Time
	UserID string
}

// Rejection distinguishes the rolling attempt limit from occupied password
// capacity. Only AttemptLimit calls for the HTTP adapter's Retry-After header.
type Rejection uint8

const (
	None Rejection = iota
	AttemptLimit
	PasswordBusy
)

// Manager is safe for concurrent use. Its times are supplied by callers so
// login's wall clock and invitation acceptance's effective clock stay distinct.
// A Manager must not be copied after first use.
type Manager struct {
	mu                 sync.Mutex
	sessions           map[[32]byte]Session
	loginAttempts      []time.Time
	invitationAttempts []time.Time
	passwords          int
}

func New() *Manager {
	return &Manager{sessions: make(map[[32]byte]Session)}
}

// Create removes strictly expired sessions before checking the 128-session
// capacity. It uses the caller's original login observation for the eight-hour
// expiry, even when password verification completed later.
func (manager *Manager) Create(now time.Time, secret, userID, csrf string) (Session, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	for key, current := range manager.sessions {
		if now.After(current.Expiry) {
			delete(manager.sessions, key)
		}
	}
	if len(manager.sessions) >= 128 {
		return Session{}, false
	}
	current := Session{CSRF: csrf, Expiry: now.Add(8 * time.Hour), UserID: userID}
	manager.sessions[sha256.Sum256([]byte(secret))] = current
	return current, true
}

// Lookup deletes a session only when now is strictly after its expiry.
func (manager *Manager) Lookup(now time.Time, secret string) (Session, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	key := sha256.Sum256([]byte(secret))
	current, ok := manager.sessions[key]
	if !ok || now.After(current.Expiry) {
		delete(manager.sessions, key)
		return Session{}, false
	}
	return current, true
}

func (manager *Manager) Delete(secret string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	delete(manager.sessions, sha256.Sum256([]byte(secret)))
}

// BeginLogin consumes a login attempt before trying to reserve password work.
// The caller must defer the returned release on every accepted path, including
// invalid input, failed authentication and cancellation. Rejections return nil.
func (manager *Manager) BeginLogin(now time.Time) (release func(), rejection Rejection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.begin(now, &manager.loginAttempts)
}

// BeginInvitation has its own rolling attempt budget but shares the two active
// password operations with BeginLogin. Its release contract is the same.
func (manager *Manager) BeginInvitation(now time.Time) (release func(), rejection Rejection) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.begin(now, &manager.invitationAttempts)
}

func (manager *Manager) begin(now time.Time, attempts *[]time.Time) (func(), Rejection) {
	cutoff := now.Add(-time.Minute)
	filtered := (*attempts)[:0]
	for _, attempt := range *attempts {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}
	*attempts = filtered
	if len(filtered) >= 10 {
		return nil, AttemptLimit
	}
	*attempts = append(*attempts, now)
	if manager.passwords >= 2 {
		return nil, PasswordBusy
	}
	manager.passwords++
	var once sync.Once
	return func() {
		once.Do(func() {
			manager.mu.Lock()
			manager.passwords--
			manager.mu.Unlock()
		})
	}, None
}
