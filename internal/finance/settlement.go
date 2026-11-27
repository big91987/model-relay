package finance

import (
	"database/sql"
	"errors"
	"model-relay/internal/accounting"
	"model-relay/internal/secure"
)

// Settle appends settlement deltas inside the caller's request or recovery transaction.
func Settle(transaction *sql.Tx, requestID string) error {
	var period string
	err := transaction.QueryRow("SELECT period FROM admissions WHERE request_id=?", requestID).Scan(&period)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	rows, err := transaction.Query("SELECT a.ordinal,a.prompt_tokens,a.completion_tokens,a.total_tokens,coalesce(h.price_reference,p.price_id),coalesce(h.cost_input,p.cost_input),coalesce(h.cost_output,p.cost_output),coalesce(h.tenant_input,p.tenant_input),coalesce(h.tenant_output,p.tenant_output),p.max_input,p.max_output FROM request_attempts a JOIN request_prices p ON p.request_id=a.request_id AND p.upstream_id=a.upstream_id JOIN attempt_dispatches d ON d.request_id=a.request_id AND d.ordinal=a.ordinal LEFT JOIN historical_price_evidence h ON h.request_id=p.request_id AND h.upstream_id=p.upstream_id WHERE a.request_id=? AND d.possible=1 ORDER BY a.ordinal", requestID)
	if err != nil {
		return err
	}
	type attempted struct {
		ordinal              int
		input, output, total sql.NullInt64
		price                accounting.PriceSnapshot
	}
	values := []attempted{}
	for rows.Next() {
		var value attempted
		if rows.Scan(&value.ordinal, &value.input, &value.output, &value.total, &value.price.ID, &value.price.CostIn, &value.price.CostOut, &value.price.FeeIn, &value.price.FeeOut, &value.price.Input, &value.price.Output) != nil {
			rows.Close()
			return errors.New("attempt settlement scan failed")
		}
		values = append(values, value)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var tokensHeld, feesHeld, pending int64
	for _, value := range values {
		var revision int64
		var evidence, reason string
		input, output := value.input.Int64, value.output.Int64
		reconciled := transaction.QueryRow("SELECT revision,input_tokens,output_tokens,reason,evidence FROM usage_reconciliations WHERE request_id=? AND ordinal=? ORDER BY revision DESC LIMIT 1", requestID, value.ordinal).Scan(&revision, &input, &output, &reason, &evidence)
		if reconciled != nil && !errors.Is(reconciled, sql.ErrNoRows) {
			return reconciled
		}
		knownInput := value.input.Valid || reconciled == nil
		knownOutput := value.output.Valid || reconciled == nil
		complete := reconciled == nil || value.total.Valid && knownInput && knownOutput && input <= 9223372036854775807-output && input+output == value.total.Int64
		if reconciled != nil && value.total.Valid && knownInput && knownOutput && !complete {
			knownInput = false
			knownOutput = false
		}
		if !knownInput {
			input = 0
			tokensHeld += value.price.Input
		}
		if !knownOutput {
			output = 0
			tokensHeld += value.price.Output
		}
		cost, err := accounting.Charge(input, output, value.price.CostIn, value.price.CostOut)
		if err != nil {
			return err
		}
		fee, err := accounting.Charge(input, output, value.price.FeeIn, value.price.FeeOut)
		if err != nil {
			return err
		}
		if value.price.ID == "legacy_unpriced" || !complete || !knownInput || !knownOutput {
			pending++
			unknownInput, unknownOutput := int64(0), int64(0)
			if !knownInput {
				unknownInput = value.price.Input
			}
			if !knownOutput {
				unknownOutput = value.price.Output
			}
			if !complete && knownInput && knownOutput {
				unknownInput = max(value.price.Input-input, 0)
				unknownOutput = max(value.price.Output-output, 0)
				tokensHeld += unknownInput + unknownOutput
			}
			heldTotal, err := accounting.Reserve(input+unknownInput, output+unknownOutput, value.price.FeeIn, value.price.FeeOut)
			if err != nil {
				return err
			}
			if heldTotal > fee {
				feesHeld += heldTotal - fee
			}
		}
		var priorCost, priorFee, priorInput, priorOutput int64
		if transaction.QueryRow("SELECT coalesce(sum(cost),0),coalesce(sum(fee),0),coalesce(sum(input_tokens),0),coalesce(sum(output_tokens),0) FROM charge_entries WHERE request_id=? AND ordinal=?", requestID, value.ordinal).Scan(&priorCost, &priorFee, &priorInput, &priorOutput) != nil {
			return errors.New("ledger scan failed")
		}
		var existing int
		if transaction.QueryRow("SELECT count(*) FROM charge_entries WHERE request_id=? AND ordinal=? AND revision=?", requestID, value.ordinal, revision).Scan(&existing) != nil {
			return errors.New("ledger source scan failed")
		}
		if existing == 0 {
			if _, err := transaction.Exec("INSERT INTO charge_entries VALUES(?,?,?,?,?,?,?,?,?,?,?)", secure.Random(16), requestID, value.ordinal, revision, period, input-priorInput, output-priorOutput, cost-priorCost, fee-priorFee, reason, evidence); err != nil {
				return err
			}
		}
		if value.price.Input > 0 && (input > value.price.Input || output > value.price.Output) {
			if _, err := transaction.Exec("DELETE FROM model_bounds WHERE (alias,upstream_id) IN(SELECT r.alias,a.upstream_id FROM requests r JOIN request_attempts a ON a.request_id=r.id WHERE r.id=? AND a.ordinal=?)", requestID, value.ordinal); err != nil {
				return err
			}
		}
	}
	_, err = transaction.Exec("UPDATE admissions SET running=0,tokens_reserved=?,fee_reserved=?,pending=? WHERE request_id=?", tokensHeld, feesHeld, pending, requestID)
	return err
}
