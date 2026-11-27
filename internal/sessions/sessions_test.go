package sessions_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"model-relay/internal/sessions"
)

func TestSessionExpiryUsesSuppliedTimeAndDeletesExpiredSession(t *testing.T) {
	manager := sessions.New()
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	created, ok := manager.Create(now, "secret", "user", "csrf")
	if !ok || created.UserID != "user" || created.CSRF != "csrf" || !created.Expiry.Equal(now.Add(8*time.Hour)) {
		t.Fatalf("create: session=%+v accepted=%v", created, ok)
	}
	created.UserID = "changed outside owner"
	current, ok := manager.Lookup(now.Add(8*time.Hour), "secret")
	if !ok || current.UserID != "user" {
		t.Fatalf("session must remain valid exactly at expiry: %+v %v", current, ok)
	}
	if _, ok := manager.Lookup(now.Add(8*time.Hour+time.Nanosecond), "secret"); ok {
		t.Fatal("session survived expiry")
	}
	if _, ok := manager.Lookup(now, "secret"); ok {
		t.Fatal("expired session resurrected after clock moved back")
	}
}

func TestSessionSecretsAndManagersAreIsolated(t *testing.T) {
	manager, other := sessions.New(), sessions.New()
	now := time.Unix(1000, 0)
	manager.Create(now, "one", "first", "csrf-one")
	manager.Create(now, "two", "second", "csrf-two")
	other.Create(now, "one", "other", "csrf-other")
	if _, ok := manager.Lookup(now, "unknown"); ok {
		t.Fatal("unknown secret authenticated")
	}
	manager.Delete("one")
	if _, ok := manager.Lookup(now, "one"); ok {
		t.Fatal("deleted session authenticated")
	}
	if current, ok := manager.Lookup(now, "two"); !ok || current.UserID != "second" {
		t.Fatal("logout invalidated another session")
	}
	if current, ok := other.Lookup(now, "one"); !ok || current.UserID != "other" {
		t.Fatal("logout crossed manager boundary")
	}
}

func TestSessionLimitReclaimsOnlyStrictlyExpiredSessions(t *testing.T) {
	manager := sessions.New()
	now := time.Unix(1000, 0)
	for index := range 128 {
		if _, ok := manager.Create(now, fmt.Sprint(index), "user", "csrf"); !ok {
			t.Fatalf("session %d rejected before capacity", index+1)
		}
	}
	if _, ok := manager.Create(now, "overflow", "user", "csrf"); ok {
		t.Fatal("129th live session accepted")
	}
	if _, ok := manager.Create(now.Add(8*time.Hour), "exact-expiry", "user", "csrf"); ok {
		t.Fatal("session at exact expiry reclaimed capacity")
	}
	if _, ok := manager.Create(now.Add(8*time.Hour+time.Nanosecond), "replacement", "user", "csrf"); !ok {
		t.Fatal("expired sessions did not release capacity")
	}
	if _, ok := manager.Lookup(now, "0"); ok {
		t.Fatal("capacity cleanup left expired session present")
	}
}

func TestAttemptWindowsAreIndependentAndExcludeExactCutoff(t *testing.T) {
	manager := sessions.New()
	other := sessions.New()
	now := time.Unix(1000, 0)
	for _, begin := range []func(time.Time) (func(), sessions.Rejection){manager.BeginLogin, manager.BeginInvitation} {
		for index := range 10 {
			release, rejected := begin(now)
			if rejected != sessions.None || release == nil {
				t.Fatalf("attempt %d rejected: %v", index+1, rejected)
			}
			release()
		}
		for _, observed := range []time.Time{now, now.Add(time.Minute - time.Nanosecond), now.Add(-time.Hour)} {
			if release, rejected := begin(observed); rejected != sessions.AttemptLimit || release != nil {
				t.Fatalf("window limit at %s: release=%v rejection=%v", observed, release != nil, rejected)
			}
		}
		release, rejected := begin(now.Add(time.Minute))
		if rejected != sessions.None || release == nil {
			t.Fatalf("exact cutoff not pruned: %v", rejected)
		}
		release()
	}
	release, rejected := other.BeginLogin(now)
	if rejected != sessions.None || release == nil {
		t.Fatal("attempt history crossed manager boundary")
	}
	release()
}

func TestBusyPasswordsStillCountAttemptsAndShareCapacity(t *testing.T) {
	manager := sessions.New()
	now := time.Unix(1000, 0)
	login, rejected := manager.BeginLogin(now)
	if rejected != sessions.None {
		t.Fatal("first password operation rejected", rejected)
	}
	defer login()
	invitation, rejected := manager.BeginInvitation(now)
	if rejected != sessions.None {
		t.Fatal("second password operation rejected", rejected)
	}
	defer invitation()
	for index := range 9 {
		if release, rejected := manager.BeginLogin(now); rejected != sessions.PasswordBusy || release != nil {
			t.Fatalf("busy attempt %d: release=%v rejection=%v", index+2, release != nil, rejected)
		}
	}
	if release, rejected := manager.BeginLogin(now); rejected != sessions.AttemptLimit || release != nil {
		t.Fatalf("busy password work did not consume attempt budget: %v", rejected)
	}
	login()
	if release, rejected := manager.BeginLogin(now); rejected != sessions.AttemptLimit || release != nil {
		t.Fatal("slot release reset attempt history")
	}
	release, rejected := manager.BeginInvitation(now)
	if rejected != sessions.None || release == nil {
		t.Fatal("independent invitation history could not reuse slot", rejected)
	}
	release()
}

func TestCanceledOperationReleasesPasswordCapacity(t *testing.T) {
	manager := sessions.New()
	now := time.Unix(1000, 0)
	held, rejected := manager.BeginLogin(now)
	if rejected != sessions.None {
		t.Fatal(rejected)
	}
	defer held()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		release, rejected := manager.BeginInvitation(now)
		if rejected != sessions.None {
			t.Error("invitation failed to acquire second slot", rejected)
			return
		}
		defer release()
		<-ctx.Done()
	}()
	cancel()
	<-done
	release, rejected := manager.BeginInvitation(now)
	if rejected != sessions.None || release == nil {
		t.Fatal("canceled password operation leaked its slot", rejected)
	}
	release()
}

func TestConcurrentSessionAdmissionRespectsLimit(t *testing.T) {
	manager := sessions.New()
	now := time.Unix(1000, 0)
	var created atomic.Int64
	var wait sync.WaitGroup
	for index := range 256 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, ok := manager.Create(now, fmt.Sprint(index), "user", "csrf"); ok {
				created.Add(1)
			}
		}()
	}
	wait.Wait()
	if created.Load() != 128 {
		t.Fatalf("concurrent session admissions: %d", created.Load())
	}
}

func TestConcurrentPasswordAdmissionRespectsCapacityAndAttemptLimit(t *testing.T) {
	manager := sessions.New()
	now := time.Unix(1000, 0)
	var admitted atomic.Int64
	var wait sync.WaitGroup
	finished := make(chan struct{})
	decisions := make(chan struct{}, 10)
	for range 10 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			release, rejected := manager.BeginLogin(now)
			if rejected == sessions.None {
				admitted.Add(1)
			}
			decisions <- struct{}{}
			if rejected == sessions.None {
				<-finished
				release()
			}
		}()
	}
	for range 10 {
		<-decisions
	}
	if release, rejected := manager.BeginLogin(now); rejected != sessions.AttemptLimit || release != nil {
		t.Errorf("concurrent attempts exceeded rolling budget: %v", rejected)
	}
	close(finished)
	wait.Wait()
	if admitted.Load() != 2 {
		t.Fatalf("concurrent password admissions: %d", admitted.Load())
	}
}
