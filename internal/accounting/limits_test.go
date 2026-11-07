package accounting

import "testing"

func TestAdmissionRulesKeepOrderingAndZeroBudget(t *testing.T) {
	ptr := func(value int64) *int64 { return &value }
	for _, scenario := range []struct {
		name        string
		limits      ResourceLimits
		used        ResourceUsage
		tokens, fee int64
		want        string
	}{
		{name: "unlimited"},
		{name: "calls precede rpm", limits: ResourceLimits{Calls: ptr(1), RPM: ptr(1)}, used: ResourceUsage{Calls: 1, RPM: 1}, want: "calls_exceeded"},
		{name: "token reservation", limits: ResourceLimits{TPM: ptr(20)}, used: ResourceUsage{TPM: 10}, tokens: 11, want: "tpm_exceeded"},
		{name: "zero soft budget", limits: ResourceLimits{Budget: ptr(0)}, want: "budget_exceeded"},
		{name: "positive soft budget", limits: ResourceLimits{Budget: ptr(1)}, used: ResourceUsage{Fee: 10}, fee: 1},
		{name: "hard reservation", limits: ResourceLimits{Hard: true, Budget: ptr(10)}, used: ResourceUsage{Fee: 5, Reserved: 4}, fee: 2, want: "budget_exceeded"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if got := scenario.limits.CheckAdmission(scenario.used, scenario.tokens, scenario.fee); got != scenario.want {
				t.Fatal(got, scenario.want)
			}
		})
	}
}
