package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"model-relay/internal/relay"
	"model-relay/tests/latencyfixture"
)

type cell struct {
	Alias    string
	Scenario string
	Round    int
	Stream   bool
	Strategy string
	Fixture  *latencyfixture.Scenario
	Count    int
}
type sample struct {
	Version      string                       `json:"experiment_version"`
	Cell         string                       `json:"cell"`
	Scenario     string                       `json:"scenario"`
	Round        int                          `json:"round"`
	Stream       bool                         `json:"stream"`
	Strategy     string                       `json:"strategy"`
	Index        int                          `json:"index"`
	Phase        string                       `json:"phase"`
	Started      string                       `json:"started_utc"`
	Finished     string                       `json:"finished_utc"`
	ID           string                       `json:"request_id"`
	Selected     int64                        `json:"selected_upstream"`
	Status       int                          `json:"status"`
	State        string                       `json:"state"`
	Attempts     int                          `json:"attempt_count"`
	AttemptFacts []relay.Attempt              `json:"attempts"`
	Latency      *float64                     `json:"metric_ms"`
	Duration     float64                      `json:"client_duration_ms"`
	Upstream     latencyfixture.Timing        `json:"upstream_timing"`
	Decisions    []relay.SelectionObservation `json:"selection_evidence"`
	Observations []relay.LatencyResult        `json:"latency_observations"`
	Charges      int                          `json:"charge_entry_count"`
}
type distribution struct {
	Mean float64 `json:"mean"`
	P50  float64 `json:"p50"`
	P95  float64 `json:"p95"`
}
type summary struct {
	Cell         string       `json:"cell"`
	Scenario     string       `json:"scenario"`
	Round        int          `json:"round"`
	Stream       bool         `json:"stream"`
	Strategy     string       `json:"strategy"`
	Phase        string       `json:"phase"`
	Started      string       `json:"started_utc"`
	Finished     string       `json:"finished_utc"`
	Count        int          `json:"requests"`
	Successful   int          `json:"successful"`
	Attempts     int          `json:"attempts"`
	Missing      int          `json:"missing_metric"`
	FastFraction float64      `json:"fast_fraction"`
	Latency      distribution `json:"metric_ms"`
	Duration     distribution `json:"duration_ms"`
}

func statistics(values []float64) distribution {
	if len(values) == 0 {
		return distribution{}
	}
	ordered := append([]float64{}, values...)
	sort.Float64s(ordered)
	total := 0.0
	for _, value := range ordered {
		total += value
	}
	return distribution{Mean: total / float64(len(values)), P50: ordered[int(math.Ceil(float64(len(ordered))*.5))-1], P95: ordered[int(math.Ceil(float64(len(ordered))*.95))-1]}
}
func summarize(samples []sample) summary {
	first := samples[0]
	result := summary{Cell: first.Cell, Scenario: first.Scenario, Round: first.Round, Stream: first.Stream, Strategy: first.Strategy, Phase: first.Phase, Started: first.Started, Finished: samples[len(samples)-1].Finished, Count: len(samples)}
	latencies := []float64{}
	durations := []float64{}
	fast := 0
	for _, value := range samples {
		result.Attempts += value.Attempts
		if value.Status == 200 && value.State == "success" {
			result.Successful++
		}
		if value.Latency == nil {
			result.Missing++
		} else {
			latencies = append(latencies, *value.Latency)
		}
		durations = append(durations, value.Duration)
		fastID := int64(1)
		if value.Phase == "adaptation" || value.Phase == "recovered" {
			fastID = 2
		}
		if value.Selected == fastID {
			fast++
		}
	}
	result.FastFraction = float64(fast) / float64(len(samples))
	result.Latency = statistics(latencies)
	result.Duration = statistics(durations)
	return result
}
func perform(fixture *latencyfixture.Fixture, current cell, index int) (sample, error) {
	phase := "learning"
	if current.Scenario == "tail" {
		phase = "tail"
	} else if index >= 40 {
		phase = "steady"
	}
	if current.Scenario == "core" && index >= 160 {
		current.Fixture.Swap.Store(true)
		phase = "adaptation"
		if index >= 200 {
			phase = "recovered"
		}
	}
	started := time.Now()
	result := sample{Version: "issue63-v2", Cell: current.Alias, Scenario: current.Scenario, Round: current.Round, Stream: current.Stream, Strategy: current.Strategy, Index: index, Phase: phase, Started: started.UTC().Format(time.RFC3339Nano)}
	request, err := http.NewRequest("POST", fixture.Gateway.URL+"/v1/chat/completions", strings.NewReader(fmt.Sprintf(`{"model":%q,"stream":%t,"messages":[{"role":"user","content":"controlled fixed input"}]}`, current.Alias, current.Stream)))
	if err != nil {
		return result, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+fixture.Key)
	response, err := fixture.Client.Do(request)
	if err != nil {
		return result, errors.New("controlled proxy transport failed")
	}
	_, readErr := io.Copy(io.Discard, response.Body)
	response.Body.Close()
	result.Duration = float64(time.Since(started)) / float64(time.Millisecond)
	result.Finished = time.Now().UTC().Format(time.RFC3339Nano)
	result.Status = response.StatusCode
	result.ID = response.Header.Get("X-Request-ID")
	if readErr != nil {
		return result, errors.New("controlled proxy body incomplete")
	}
	detail, err := fixture.Admin("GET", "workspace/requests/"+result.ID, nil, 200)
	if err != nil {
		return result, err
	}
	var typed struct {
		State        string                       `json:"state"`
		Attempts     []relay.Attempt              `json:"attempts"`
		Decisions    []relay.SelectionObservation `json:"routing_decisions"`
		Observations []relay.LatencyResult        `json:"latency_observations"`
		Charges      []json.RawMessage            `json:"charge_entries"`
	}
	raw, _ := json.Marshal(detail)
	if err = json.Unmarshal(raw, &typed); err != nil {
		return result, err
	}
	result.State = typed.State
	result.Attempts = len(typed.Attempts)
	result.AttemptFacts = typed.Attempts
	result.Charges = len(typed.Charges)
	result.Decisions = typed.Decisions
	result.Observations = typed.Observations
	if result.Attempts > 0 {
		result.Selected = typed.Attempts[0].UpstreamID
	}
	if len(typed.Observations) > 0 {
		result.Latency = typed.Observations[0].Value
	}
	result.Upstream = fixture.Timing(result.ID)
	return result, nil
}
func run() (err error) {
	started := time.Now()
	root := filepath.Join("test-results", "latency-experiment", started.UTC().Format("20060102T150405.000000000Z"))
	if err = os.MkdirAll(root, 0700); err != nil {
		return errors.New("cannot create experiment evidence")
	}
	fixture, err := latencyfixture.New()
	if err != nil {
		return err
	}
	defer fixture.Close()
	if err = fixture.SetupUpstreams(); err != nil {
		return err
	}
	cells := []cell{}
	aliases := []string{}
	for round := 1; round <= 2; round++ {
		strategies := []string{"weighted", "low_latency"}
		if round == 2 {
			strategies = []string{"low_latency", "weighted"}
		}
		for _, stream := range []bool{false, true} {
			for _, strategy := range strategies {
				alias := fmt.Sprintf("core-%d-%t-%s", round, stream, strategy)
				cells = append(cells, cell{Alias: alias, Scenario: "core", Round: round, Stream: stream, Strategy: strategy, Fixture: &latencyfixture.Scenario{}, Count: 280})
			}
		}
	}
	for _, strategy := range []string{"weighted", "low_latency"} {
		alias := "heartbeat-" + strategy
		cells = append(cells, cell{Alias: alias, Scenario: "heartbeat", Round: 1, Stream: true, Strategy: strategy, Fixture: &latencyfixture.Scenario{Heartbeat: true}, Count: 80})
	}
	for _, stream := range []bool{false, true} {
		for _, strategy := range []string{"weighted", "low_latency"} {
			alias := fmt.Sprintf("tail-%t-%s", stream, strategy)
			cells = append(cells, cell{Alias: alias, Scenario: "tail", Round: 1, Stream: stream, Strategy: strategy, Fixture: &latencyfixture.Scenario{Tail: true}, Count: 20})
		}
	}
	for _, current := range cells {
		fixture.Register(current.Alias, current.Fixture)
		if err = fixture.Configure(current.Alias, current.Strategy); err != nil {
			return err
		}
		aliases = append(aliases, current.Alias)
	}
	if err = fixture.IssueKey(aliases); err != nil {
		return err
	}
	summaries := []summary{}
	failures := []string{}
	sampleCount := 0
	for _, current := range cells {
		cellStarted := time.Now()
		fmt.Printf("EXPERIMENT_CELL state=start cell=%s count=%d utc=%s\n", current.Alias, current.Count, cellStarted.UTC().Format(time.RFC3339Nano))
		file, err := os.OpenFile(filepath.Join(root, current.Alias+".jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return errors.New("cannot create raw samples")
		}
		encoder := json.NewEncoder(file)
		collected := []sample{}
		for index := 0; index < current.Count; index++ {
			value, callErr := perform(fixture, current, index)
			if err = encoder.Encode(value); err != nil {
				file.Close()
				return err
			}
			sampleCount++
			collected = append(collected, value)
			if callErr != nil {
				file.Close()
				return callErr
			}
			if value.Status != 200 || value.State != "success" || value.Attempts != 1 || value.Latency == nil || value.Charges != 1 {
				failures = append(failures, fmt.Sprintf("%s request=%d success/attempt/metric/charge contract", current.Alias, index))
			}
			if current.Scenario == "heartbeat" && value.Selected == 2 && value.Latency != nil && (*value.Latency < value.Upstream.OutputMS-5 || value.Upstream.BodyMS >= value.Upstream.OutputMS) {
				failures = append(failures, current.Alias+" heartbeat scored as output")
			}
			if current.Scenario == "tail" && value.Latency != nil && *value.Latency >= value.Upstream.OutputMS+30 {
				failures = append(failures, current.Alias+" full duration used as score")
			}
		}
		if err = file.Close(); err != nil {
			return err
		}
		visits := map[int64]int{}
		unknown := 0
		learningEnd := 40
		if current.Count < learningEnd {
			learningEnd = current.Count
		}
		for _, value := range collected[:learningEnd] {
			if len(value.Observations) == 1 && value.Observations[0].Acceptance == "accepted" {
				visits[value.Selected]++
			}
			for _, decision := range value.Decisions {
				if decision.State == "unknown" {
					unknown++
				}
				if decision.Count < 3 && decision.EWMA != nil {
					failures = append(failures, current.Alias+" insufficient sample used as score")
				}
			}
		}
		if unknown == 0 || visits[1] == 0 || visits[2] == 0 {
			failures = append(failures, current.Alias+" cold/exploration contract")
		}
		for offset := 0; offset < len(collected); {
			end := offset + 1
			for end < len(collected) && collected[end].Phase == collected[offset].Phase {
				end++
			}
			report := summarize(collected[offset:end])
			summaries = append(summaries, report)
			offset = end
		}
		if current.Scenario == "core" && current.Strategy == "low_latency" {
			adapted := false
			for _, value := range collected[160:200] {
				for _, decision := range value.Decisions {
					if decision.UpstreamID == 2 && value.Selected == 2 && decision.Reason == "lowest_latency" {
						adapted = true
					}
				}
			}
			if !adapted {
				failures = append(failures, current.Alias+" did not adapt within 40 selections")
			}
		}
		fmt.Printf("EXPERIMENT_CELL state=end cell=%s elapsed_ms=%d samples=%d\n", current.Alias, time.Since(cellStarted).Milliseconds(), len(collected))
	}
	for _, report := range summaries {
		if report.Phase != "steady" && report.Phase != "recovered" {
			continue
		}
		if report.Strategy == "weighted" {
			if report.FastFraction < .45 || report.FastFraction > .55 {
				failures = append(failures, report.Cell+" weighted distribution")
			}
		} else {
			threshold := .8
			if report.Phase == "recovered" {
				threshold = .75
			}
			if report.FastFraction < threshold {
				failures = append(failures, report.Cell+" low latency distribution")
			}
			if report.Phase == "steady" {
				for _, control := range summaries {
					if control.Strategy == "weighted" && control.Round == report.Round && control.Stream == report.Stream && control.Scenario == report.Scenario && control.Phase == report.Phase && (control.Latency.Mean <= 0 || 1-report.Latency.Mean/control.Latency.Mean < .35) {
						failures = append(failures, report.Cell+" mean improvement below 35%")
					}
				}
			}
		}
	}
	receipt := map[string]any{"experiment_version": "issue63-v2", "provider": "controlled_http_not_supplier", "started_utc": started.UTC().Format(time.RFC3339Nano), "finished_utc": time.Now().UTC().Format(time.RFC3339Nano), "elapsed_ms": time.Since(started).Milliseconds(), "samples": sampleCount, "core_samples": 2240, "heartbeat_samples": 160, "tail_samples": 80, "windows": summaries, "failures": failures, "trace": "not_captured_privacy", "supplier": "Not Run"}
	raw, _ := json.MarshalIndent(receipt, "", "  ")
	if err = os.WriteFile(filepath.Join(root, "summary.json"), append(raw, '\n'), 0600); err != nil {
		return err
	}
	var visual strings.Builder
	visual.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="920" height="980"><rect width="100%" height="100%" fill="#f4f7fb"/><text x="20" y="30" font-size="20">Issue63 controlled experiment: startup latency (ms), not production</text>`)
	row := 0
	for _, report := range summaries {
		if report.Phase != "steady" {
			continue
		}
		row++
		y := 50 + row*60
		fmt.Fprintf(&visual, `<text x="20" y="%d" font-size="13">%s</text><rect x="330" y="%d" width="%.2f" height="22" fill="#2563eb"/><text x="730" y="%d" font-size="13">%.2f ms / fast %.0f%%</text>`, y, report.Cell, y-16, report.Latency.Mean*4, y, report.Latency.Mean, report.FastFraction*100)
	}
	visual.WriteString(`</svg>`)
	if err = os.WriteFile(filepath.Join(root, "comparison.svg"), []byte(visual.String()), 0600); err != nil {
		return err
	}
	fmt.Printf("EXPERIMENT_END samples=%d failures=%d elapsed_ms=%d evidence=%s\n", sampleCount, len(failures), time.Since(started).Milliseconds(), filepath.ToSlash(root))
	if len(failures) > 0 {
		return fmt.Errorf("controlled experiment failed %d assertions; see retained summary", len(failures))
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
