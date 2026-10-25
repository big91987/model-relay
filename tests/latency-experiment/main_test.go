package main

import "testing"

func TestNearestRankStatistics(t *testing.T) {
	got := statistics([]float64{100, 20, 20, 20})
	if got.Mean != 40 || got.P50 != 20 || got.P95 != 100 {
		t.Fatal(got)
	}
}
