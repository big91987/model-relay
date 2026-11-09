// Package configstate owns the current configuration generation, its successful
// request evidence and the registry of active upstream checks.
package configstate

import (
	"sync"

	"model-relay/internal/routing"
)

// Coordinator provides the shared lock for atomic database/runtime operations.
// Callers hold Lock across a configuration mutation, request admission or check
// registration and the related persistence work. This preserves a single
// coordination boundary without owning SQL or calling back into the server.
//
// All state methods require the caller to hold Lock; Epoch, SuccessID and
// Checking may instead run under RLock. Methods never acquire the lock again.
// A Coordinator must not be copied after first use.
type Coordinator struct {
	mu        sync.RWMutex
	epoch     uint64
	successID string
	checking  map[int64]bool
}

func New() *Coordinator {
	return &Coordinator{checking: make(map[int64]bool)}
}

func (state *Coordinator) Lock()    { state.mu.Lock() }
func (state *Coordinator) Unlock()  { state.mu.Unlock() }
func (state *Coordinator) RLock()   { state.mu.RLock() }
func (state *Coordinator) RUnlock() { state.mu.RUnlock() }

func (state *Coordinator) Epoch() uint64     { return state.epoch }
func (state *Coordinator) SuccessID() string { return state.successID }

// Changed publishes a committed configuration change to the runtime generation,
// success evidence and routing policy. Checks already in progress remain
// registered until their own completion.
func (state *Coordinator) Changed(routes *routing.Engine) {
	state.epoch++
	state.successID = ""
	routes.Reset()
}

// CompleteRequest publishes success only for the configuration under which the
// request began. A failed or stale completion clears only its own prior evidence.
// Call only after the request result has been durably recorded.
func (state *Coordinator) CompleteRequest(id string, epoch uint64, successful bool) {
	if successful && epoch == state.epoch {
		state.successID = id
	} else if state.successID == id {
		state.successID = ""
	}
}

// ClearSuccess invalidates evidence found inapplicable by a persisted-state read.
func (state *Coordinator) ClearSuccess() { state.successID = "" }

func (state *Coordinator) Checking(upstreamID int64) bool { return state.checking[upstreamID] }

// BeginCheck registers an upstream only when it has no active check. Callers
// retain Lock across eligibility checks, check-row creation and registration.
func (state *Coordinator) BeginCheck(upstreamID int64) bool {
	if state.checking[upstreamID] {
		return false
	}
	state.checking[upstreamID] = true
	return true
}

// EndCheck releases the registration on every exit path, allowing a later retry.
func (state *Coordinator) EndCheck(upstreamID int64) { delete(state.checking, upstreamID) }
