package routing

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type routeRuntime struct {
	Fingerprint string
	Current     map[int64]int
	Exploration map[bool]*explorationState
}

type Engine struct {
	mu       sync.Mutex
	runtimes map[string]*routeRuntime
	health   map[string]*Health
	latency  map[string]*latencyPool
	sequence int64
}

func New() *Engine {
	return &Engine{runtimes: map[string]*routeRuntime{}, health: map[string]*Health{}, latency: map[string]*latencyPool{}}
}

// Reset discards configuration-specific policy and samples. Health is keyed by
// execution revision and remains valid until the caller changes that revision.
func (engine *Engine) Reset() {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	engine.runtimes = map[string]*routeRuntime{}
	engine.latency = map[string]*latencyPool{}
}

func (engine *Engine) NextSequence() int64 {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	engine.sequence++
	return engine.sequence
}

func healthKey(alias string, candidate Candidate) string {
	return fmt.Sprintf("%s:%d:%d:%s", alias, candidate.UpstreamID, candidate.UpstreamExecution, candidate.RemoteModel)
}

func (engine *Engine) Health(alias string, candidate Candidate) *Health {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	value := engine.health[healthKey(alias, candidate)]
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func (engine *Engine) Stats() CacheStats {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	result := CacheStats{HealthEntries: len(engine.health)}
	for _, pool := range engine.latency {
		result.Samples += len(pool.Samples)
	}
	for _, health := range engine.health {
		if !health.Cooldown.IsZero() {
			result.CooldownEntries++
		}
	}
	return result
}

func (engine *Engine) Select(snapshot Snapshot, stream bool, round int, tried map[int64]bool) Selection {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	now := snapshot.ObservedAt
	key := fmt.Sprintf("%s:%d:%d", snapshot.Alias, snapshot.ConfigRevision, snapshot.Revision)
	runtime := engine.runtimes[key]
	if runtime == nil {
		runtime = &routeRuntime{Current: map[int64]int{}}
		engine.runtimes[key] = runtime
	}
	decisions := []RouteDecision{}
	eligible := []Candidate{}
	priority := -1
	for _, candidate := range snapshot.Candidates {
		decision := RouteDecision{Round: round, UpstreamID: candidate.UpstreamID, Enabled: candidate.Enabled, Priority: candidate.Priority, Weight: candidate.Weight, Observed: now.UTC().Format(time.RFC3339Nano)}
		health := engine.health[healthKey(snapshot.Alias, candidate)]
		switch {
		case candidate.Unavailable != "":
			decision.Reason = candidate.Unavailable
		case !candidate.Enabled:
			decision.Reason = "disabled_candidate"
		case health != nil && health.Cooldown.After(now):
			decision.Reason = "cooling"
			until := health.Cooldown.UTC().Format(time.RFC3339Nano)
			decision.Cooldown = &until
		case tried[candidate.UpstreamID]:
			decision.Reason = "already_attempted"
		default:
			eligible = append(eligible, candidate)
			if candidate.Priority > priority {
				priority = candidate.Priority
			}
		}
		decisions = append(decisions, decision)
	}
	group := []Candidate{}
	for _, candidate := range eligible {
		if candidate.Priority == priority {
			group = append(group, candidate)
		}
	}
	sort.Slice(group, func(left, right int) bool { return group[left].UpstreamID < group[right].UpstreamID })
	var selected *Candidate
	fingerprint := ""
	for _, candidate := range group {
		fingerprint += fmt.Sprintf("%d/%d/%d/%d;", candidate.UpstreamID, candidate.Priority, candidate.Weight, candidate.ExecutionRevision)
	}
	if fingerprint != runtime.Fingerprint {
		runtime.Fingerprint = fingerprint
		runtime.Current = map[int64]int{}
		runtime.Exploration = nil
	}
	total := 0
	for _, candidate := range group {
		runtime.Current[candidate.UpstreamID] += candidate.Weight
		total += candidate.Weight
		if selected == nil || runtime.Current[candidate.UpstreamID] > runtime.Current[selected.UpstreamID] {
			copy := candidate
			selected = &copy
		}
	}
	if selected != nil {
		runtime.Current[selected.UpstreamID] -= total
	}
	selected, observations := engine.chooseLatency(snapshot, runtime, group, stream, round, selected, decisions, now)
	for index := range decisions {
		if decisions[index].Reason != "" {
			continue
		}
		switch {
		case selected != nil && decisions[index].UpstreamID == selected.UpstreamID:
			decisions[index].Reason = "selected"
		case decisions[index].Priority < priority:
			decisions[index].Reason = "lower_priority"
		default:
			decisions[index].Reason = "not_selected_same_priority"
		}
	}
	for index := range observations {
		if observations[index].Reason == "not_selected" {
			observations[index].Reason = decisions[index].Reason
		}
	}
	return Selection{Candidate: selected, Decisions: decisions, Observations: observations}
}

// Observe receives only an already version-validated attempt from the gateway.
func (engine *Engine) Observe(snapshot Snapshot, candidate Candidate, observation Observation) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	key := healthKey(snapshot.Alias, candidate)
	previous := engine.health[key]
	if previous != nil && (previous.Sequence > observation.Sequence || previous.InvalidatedThrough >= observation.Sequence) {
		return
	}
	health := &Health{Sequence: observation.Sequence, State: observation.State, Observed: observation.ObservedAt, CandidateRevision: candidate.ExecutionRevision}
	if previous != nil {
		health.InvalidatedThrough = previous.InvalidatedThrough
	}
	if observation.Temporary {
		health.InvalidatedThrough = engine.sequence
		delete(engine.latency, latencyKey(snapshot, candidate, false))
		delete(engine.latency, latencyKey(snapshot, candidate, true))
		health.Cooldown = health.Observed.Add(observation.Cooldown)
	}
	engine.health[key] = health
}

// ForgetCandidate invalidates health for all former execution mappings.
func (engine *Engine) ForgetCandidate(alias string, upstreamID int64) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	prefix := fmt.Sprintf("%s:%d:", alias, upstreamID)
	for key := range engine.health {
		if strings.HasPrefix(key, prefix) {
			delete(engine.health, key)
		}
	}
}
