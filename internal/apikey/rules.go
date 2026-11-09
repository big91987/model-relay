package apikey

import (
	"time"

	"model-relay/internal/names"
)

func validInput(input Input, now time.Time) bool {
	if !names.Valid(input.Name) || !names.Valid(input.Application) || len(input.Models) > 1000 || len(input.Diagnostics) > 1000 {
		return false
	}
	if input.Expires != nil {
		expires, err := time.Parse(time.RFC3339Nano, *input.Expires)
		if err != nil || input.Enabled && !now.Before(expires) {
			return false
		}
	}
	for _, values := range [][]string{input.Models, input.Diagnostics} {
		seen := map[string]bool{}
		for _, value := range values {
			if seen[value] {
				return false
			}
			seen[value] = true
		}
	}
	return true
}

func pausedExpansion(input Input, key ControlledKey) bool {
	return input.Enabled && !key.Enabled || !subsetStrings(input.Models, key.Models) || expiryExpands(input.Expires, key.Expires)
}

func subsetStrings(values, existing []string) bool {
	set := map[string]bool{}
	for _, value := range existing {
		set[value] = true
	}
	for _, value := range values {
		if !set[value] {
			return false
		}
	}
	return true
}

func expiryExpands(next, previous *string) bool {
	if previous == nil {
		return false
	}
	if next == nil {
		return true
	}
	nextTime, nextErr := time.Parse(time.RFC3339Nano, *next)
	previousTime, previousErr := time.Parse(time.RFC3339Nano, *previous)
	return nextErr != nil || previousErr != nil || nextTime.After(previousTime)
}
