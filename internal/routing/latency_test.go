package routing

import (
	"testing"
	"time"
)

func TestLatencySamplesExpireFromMeasurement(t *testing.T) {
	origin := time.Now()
	pool := &latencyPool{}
	for index := 0; index < 3; index++ {
		pool.Samples = append(pool.Samples, latencySample{Measured: origin.Add(time.Duration(index) * time.Second), Milliseconds: float64(10 + index), Sequence: int64(index + 1)})
	}
	state, count, score, _, _ := pool.describe(origin.Add(60 * time.Second))
	if state != "ready" || count != 3 || score == nil {
		t.Fatal("60s boundary", state, count)
	}
	state, count, score, _, _ = pool.describe(origin.Add(60*time.Second + time.Nanosecond))
	if state != "insufficient" || count != 2 || score != nil {
		t.Fatal("individual expiry", state, count)
	}
	pool.Samples = append(pool.Samples, latencySample{Measured: origin.Add(61 * time.Second), Milliseconds: 100, Sequence: 4})
	_, count, _, _, _ = pool.describe(origin.Add(63 * time.Second))
	if count != 1 {
		t.Fatal("new sample renewed old observations", count)
	}
	for index := 0; index < 2; index++ {
		state, count, score, _, _ = pool.describe(origin.Add(122 * time.Second))
		if state != "expired" || count != 0 || score != nil {
			t.Fatal("expired explanation lost", state, count)
		}
	}
}

func TestLatencyExplorationBoundAndModeIsolation(t *testing.T) {
	for _, count := range []int{1, 2, 3, 100} {
		engine := New()
		snapshot := Snapshot{Alias: "alpha", Strategy: "low_latency"}
		runtime := &routeRuntime{Current: map[int64]int{}}
		group := []Candidate{}
		reasons := []RouteDecision{}
		for index := 1; index <= count; index++ {
			group = append(group, Candidate{UpstreamID: int64(index)})
			reasons = append(reasons, RouteDecision{UpstreamID: int64(index)})
		}
		visits := map[int64]bool{}
		for index := 0; index < 5*count; index++ {
			candidate, evidence := engine.chooseLatency(snapshot, runtime, group, false, 1, &group[0], reasons, time.Now())
			for _, fact := range evidence {
				if fact.UpstreamID == candidate.UpstreamID && fact.Reason == "explore" {
					visits[candidate.UpstreamID] = true
				}
			}
		}
		if len(visits) != count {
			t.Fatal("exploration starvation", count, len(visits))
		}
		_, facts := engine.chooseLatency(snapshot, runtime, group, true, 1, &group[0], reasons, time.Now())
		if facts[0].Exploration != 1 {
			t.Fatal("mode counter shared")
		}
	}
}
