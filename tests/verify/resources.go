package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type resourceSnapshot struct {
	GOOS           string            `json:"goos"`
	MetricsStatus  string            `json:"cgroup_metrics_status"`
	MetricsReason  string            `json:"cgroup_metrics_reason,omitempty"`
	StartedUTC     string            `json:"sample_started_utc"`
	FinishedUTC    string            `json:"sample_finished_utc"`
	CgroupIdentity string            `json:"cgroup_identity_sha256,omitempty"`
	Values         map[string]string `json:"values"`
	Counters       map[string]uint64 `json:"counters"`
	RuntimeCPUs    int               `json:"runtime_num_cpu"`
	GOMAXPROCS     int               `json:"runtime_gomaxprocs"`
	Errors         []string          `json:"errors,omitempty"`
}

type resourceDelta struct {
	FromUTC  string            `json:"from_utc"`
	ToUTC    string            `json:"to_utc"`
	Scope    string            `json:"scope"`
	Counters map[string]uint64 `json:"counters"`
}

func sampleResources() resourceSnapshot {
	return readResources(os.ReadFile)
}

func readResources(read func(string) ([]byte, error)) (snapshot resourceSnapshot) {
	return readResourcesForPlatform(runtime.GOOS, read)
}

func readResourcesForPlatform(goos string, read func(string) ([]byte, error)) (snapshot resourceSnapshot) {
	snapshot = resourceSnapshot{GOOS: goos, MetricsStatus: "available", StartedUTC: time.Now().UTC().Format(time.RFC3339Nano), Values: map[string]string{}, Counters: map[string]uint64{}, RuntimeCPUs: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0)}
	defer func() {
		snapshot.FinishedUTC = time.Now().UTC().Format(time.RFC3339Nano)
		if len(snapshot.Errors) != 0 {
			snapshot.MetricsStatus = "error"
		}
	}()
	if goos != "linux" {
		snapshot.MetricsStatus = "unsupported"
		snapshot.MetricsReason = "cgroup metrics are not supported on this operating system"
		return snapshot
	}
	readValue := func(name, path string) string {
		data, err := read(path)
		if err != nil || len(data) > 64*1024 {
			snapshot.Errors = append(snapshot.Errors, name+": read failed or oversized")
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	membership := readValue("cgroup membership", "/proc/self/cgroup")
	var cgroupPath string
	for _, line := range strings.Split(membership, "\n") {
		if strings.HasPrefix(line, "0::") {
			cgroupPath = strings.TrimPrefix(line, "0::")
		}
	}
	if !filepath.IsAbs(cgroupPath) || filepath.Clean(cgroupPath) != cgroupPath {
		snapshot.Errors = append(snapshot.Errors, "current cgroup v2 membership unavailable")
		return snapshot
	}
	snapshot.CgroupIdentity = fmt.Sprintf("%x", sha256.Sum256([]byte(cgroupPath)))
	for _, name := range []string{"cpu.max", "cpu.stat", "memory.max", "memory.events", "cpuset.cpus.effective"} {
		snapshot.Values[name] = readValue(name, filepath.Join("/sys/fs/cgroup", cgroupPath, name))
	}
	status := readValue("process affinity", "/proc/self/status")
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "Cpus_allowed_list:") {
			snapshot.Values["process_allowed_cpu_list"] = strings.TrimSpace(strings.TrimPrefix(line, "Cpus_allowed_list:"))
		}
	}
	if snapshot.Values["process_allowed_cpu_list"] == "" {
		snapshot.Errors = append(snapshot.Errors, "process affinity unavailable")
	}
	positive := func(value string) bool {
		number, err := strconv.ParseUint(value, 10, 64)
		return err == nil && number > 0
	}
	cpuMax := strings.Fields(snapshot.Values["cpu.max"])
	if len(cpuMax) != 2 || (cpuMax[0] != "max" && !positive(cpuMax[0])) || !positive(cpuMax[1]) {
		snapshot.Errors = append(snapshot.Errors, "cpu.max: invalid quota/period")
	}
	if value := snapshot.Values["memory.max"]; value != "max" && !positive(value) {
		snapshot.Errors = append(snapshot.Errors, "memory.max: invalid limit")
	}
	groups := map[string][]string{
		"cpu.stat":      {"usage_usec", "user_usec", "system_usec", "nr_periods", "nr_throttled", "throttled_usec"},
		"memory.events": {"low", "high", "max", "oom", "oom_kill", "oom_group_kill"},
	}
	for name, required := range groups {
		for _, counter := range required {
			found := false
			for _, line := range strings.Split(snapshot.Values[name], "\n") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[0] == counter {
					number, err := strconv.ParseUint(fields[1], 10, 64)
					if err == nil && !found {
						snapshot.Counters[name+"."+counter] = number
						found = true
					} else {
						snapshot.Errors = append(snapshot.Errors, name+"."+counter+": invalid or duplicate counter")
					}
				}
			}
			if !found {
				snapshot.Errors = append(snapshot.Errors, name+"."+counter+": missing valid counter")
			}
		}
	}
	return snapshot
}

func resourceSampleError(phase string, snapshot resourceSnapshot) error {
	if len(snapshot.Errors) != 0 {
		return fmt.Errorf("resource sample %s failed: %s", phase, strings.Join(snapshot.Errors, "; "))
	}
	if snapshot.MetricsStatus == "available" && snapshot.GOOS == "linux" {
		return nil
	}
	if snapshot.MetricsStatus == "unsupported" && snapshot.GOOS != "linux" && snapshot.GOOS != "" && snapshot.MetricsReason != "" && snapshot.CgroupIdentity == "" && len(snapshot.Values) == 0 && len(snapshot.Counters) == 0 {
		return nil
	}
	return fmt.Errorf("resource sample %s failed: invalid platform/metrics status", phase)
}

func resourceDifference(before, after resourceSnapshot) (*resourceDelta, error) {
	if err := errors.Join(resourceSampleError("before", before), resourceSampleError("after", after)); err != nil {
		return nil, err
	}
	if before.GOOS != after.GOOS || before.MetricsStatus != after.MetricsStatus {
		return nil, errors.New("resource delta unavailable: platform or metrics availability changed")
	}
	if before.MetricsStatus == "unsupported" {
		return nil, nil
	}
	if len(before.Errors) != 0 || len(after.Errors) != 0 || before.CgroupIdentity == "" || before.CgroupIdentity != after.CgroupIdentity {
		return nil, errors.New("resource delta unavailable: incomplete samples or changed cgroup")
	}
	for _, name := range []string{"cpu.max", "memory.max", "cpuset.cpus.effective", "process_allowed_cpu_list"} {
		if before.Values[name] != after.Values[name] {
			return nil, errors.New("resource delta unavailable: changed limits or affinity")
		}
	}
	delta := &resourceDelta{FromUTC: before.FinishedUTC, ToUTC: after.FinishedUTC, Scope: "shared visible cgroup between samples; not per-process or sole causal attribution", Counters: map[string]uint64{}}
	for name, initial := range before.Counters {
		final, exists := after.Counters[name]
		if !exists || final < initial {
			return nil, errors.New("resource delta unavailable: missing or decreased counter")
		}
		delta.Counters[name] = final - initial
	}
	return delta, nil
}
