package configstate_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"model-relay/internal/configstate"
	"model-relay/internal/routing"
)

func TestConfigurationChangeInvalidatesSuccessAndRejectsStaleCompletion(t *testing.T) {
	state := configstate.New()
	routes := routing.New()
	state.Lock()
	defer state.Unlock()
	original := state.Epoch()
	state.CompleteRequest("old", original, true)
	if state.SuccessID() != "old" {
		t.Fatal("current successful request was not published")
	}
	state.Changed(routes)
	if state.SuccessID() != "" || state.Epoch() == original {
		t.Fatal("configuration change retained success evidence or generation")
	}
	state.CompleteRequest("late", original, true)
	if state.SuccessID() != "" {
		t.Fatal("late request entered new configuration's success evidence")
	}
	state.CompleteRequest("current", state.Epoch(), true)
	state.CompleteRequest("unrelated", original, false)
	if state.SuccessID() != "current" {
		t.Fatal("unrelated completion cleared current evidence")
	}
	state.CompleteRequest("current", state.Epoch(), false)
	if state.SuccessID() != "" {
		t.Fatal("failed repeat completion retained its success evidence")
	}
	state.CompleteRequest("invalidated", state.Epoch(), true)
	state.ClearSuccess()
	if state.SuccessID() != "" {
		t.Fatal("invalidated persisted evidence remained visible")
	}
}

func TestConfigurationStateAndCheckLifetimesAreIndependent(t *testing.T) {
	state, other := configstate.New(), configstate.New()
	routes := routing.New()
	state.Lock()
	defer state.Unlock()
	other.Lock()
	defer other.Unlock()
	initial := other.Epoch()
	other.CompleteRequest("other", initial, true)
	if !state.BeginCheck(7) || !state.Checking(7) || state.BeginCheck(7) {
		t.Fatal("active check did not exclude duplicates")
	}
	if !state.BeginCheck(8) || !other.BeginCheck(7) {
		t.Fatal("checks for unrelated upstreams or owners conflicted")
	}
	state.Changed(routes)
	if !state.Checking(7) || !state.Checking(8) {
		t.Fatal("configuration change erased running checks")
	}
	if other.Epoch() != initial || other.SuccessID() != "other" {
		t.Fatal("configuration mutation crossed owner boundary")
	}
	state.EndCheck(7)
	if state.Checking(7) || !state.Checking(8) || !other.Checking(7) {
		t.Fatal("check release affected an unrelated check")
	}
	if !state.BeginCheck(7) {
		t.Fatal("finished check prevented retry")
	}
}

func TestConcurrentCheckAdmissionDeduplicatesUntilRelease(t *testing.T) {
	state := configstate.New()
	var admitted atomic.Int64
	var wait sync.WaitGroup
	for range 64 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			state.Lock()
			defer state.Unlock()
			if state.BeginCheck(7) {
				admitted.Add(1)
			}
		}()
	}
	wait.Wait()
	if admitted.Load() != 1 {
		t.Fatalf("duplicate upstream checks admitted: %d", admitted.Load())
	}
	state.Lock()
	state.EndCheck(7)
	if !state.BeginCheck(7) {
		t.Fatal("check completion did not allow retry")
	}
	state.Unlock()
}

func TestConcurrentConfigurationChangesPublishMatchingEvidence(t *testing.T) {
	state := configstate.New()
	routes := routing.New()
	var wait sync.WaitGroup
	for range 64 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			state.Lock()
			state.Changed(routes)
			state.CompleteRequest("current", state.Epoch(), true)
			state.Unlock()
			state.RLock()
			defer state.RUnlock()
			if state.SuccessID() != "current" {
				t.Error("read observed configuration without its corresponding evidence")
			}
		}()
	}
	wait.Wait()
	state.RLock()
	defer state.RUnlock()
	if state.Epoch() != 64 {
		t.Fatalf("configuration changes lost: %d", state.Epoch())
	}
}
