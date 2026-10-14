package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const verificationGateBudget = 24 * time.Minute

func executeVerificationGate(parent context.Context, requested *exec.Cmd, directory string, environment []string, output io.Writer, state *receipt, budget time.Duration) error {
	if budget <= 0 || budget > verificationGateBudget {
		return errors.New("verification gate budget must be positive and at most 24 minutes")
	}
	ctx, cancel := context.WithTimeout(parent, budget)
	defer cancel()
	deadline, _ := ctx.Deadline()
	entry := &commandReceipt{Stage: "verify-gates", StartedUTC: time.Now().UTC().Format(time.RFC3339Nano), BudgetMilliseconds: max(0, time.Until(deadline).Milliseconds()), Exit: -1}
	state.Gate, state.GateBudgetMS = entry, budget.Milliseconds()
	if ctx.Err() != nil {
		entry.BudgetMilliseconds = 0
	}
	if _, err := fmt.Fprintf(output, "GATE_START started_utc=%s configured_budget_ms=%d effective_budget_ms=%d\n", entry.StartedUTC, state.GateBudgetMS, entry.BudgetMilliseconds); err != nil {
		return err
	}
	command := exec.CommandContext(ctx, requested.Path, requested.Args[1:]...)
	command.Dir, command.Env = directory, environment
	started := time.Now()
	code, err := execute(ctx, command, output)
	entry.ExecuteMilliseconds = time.Since(started).Milliseconds()
	entry.Exit, entry.Cancelled, state.GateExit = code, ctx.Err() != nil, code
	if command.Process != nil {
		state.Cleanup = "owned_process_group_cleanup_attempted"
		checkStarted := time.Now()
		absent := ownedGroupAbsent(command.Process.Pid)
		entry.GroupCheckMilliseconds = time.Since(checkStarted).Milliseconds()
		entry.GroupAbsent, state.OwnedGroupAbsent = &absent, &absent
		if !absent {
			err = errors.Join(err, errors.New("owned process group remains after bounded cleanup"))
		}
	} else {
		state.Cleanup = "gate_process_not_started"
	}
	_, writeErr := fmt.Fprintf(output, "GATE_END exit=%d cancelled=%t execute_ms=%d group_check_ms=%d\n", code, entry.Cancelled, entry.ExecuteMilliseconds, entry.GroupCheckMilliseconds)
	return errors.Join(err, writeErr)
}
