package relay

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ResourceLimits struct {
	Calls       *int64 `json:"calls"`
	RPM         *int64 `json:"rpm"`
	TPM         *int64 `json:"tpm"`
	Concurrency *int64 `json:"concurrency"`
	Budget      *int64 `json:"budget_micro_usd"`
	Hard        bool   `json:"hard"`
	Revision    int64  `json:"revision"`
}

type ResourceUsage struct {
	Calls       int64 `json:"calls"`
	RPM         int64 `json:"rpm"`
	TPM         int64 `json:"tpm"`
	Concurrency int64 `json:"concurrency"`
	Fee         int64 `json:"fee_micro_usd,string"`
	Reserved    int64 `json:"reserved_micro_usd,string"`
	Pending     int64 `json:"pending"`
}

type PriceSnapshot struct {
	Upstream int64  `json:"upstream_id"`
	ID       string `json:"price_id"`
	CostIn   int64  `json:"cost_input_rate"`
	CostOut  int64  `json:"cost_output_rate"`
	FeeIn    int64  `json:"tenant_input_rate"`
	FeeOut   int64  `json:"tenant_output_rate"`
	Input    int64  `json:"max_input"`
	Output   int64  `json:"max_output"`
}

func (price PriceSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"upstream_id": price.Upstream, "price_id": price.ID, "cost_input_rate": strconv.FormatInt(price.CostIn, 10), "cost_output_rate": strconv.FormatInt(price.CostOut, 10), "tenant_input_rate": strconv.FormatInt(price.FeeIn, 10), "tenant_output_rate": strconv.FormatInt(price.FeeOut, 10), "max_input": price.Input, "max_output": price.Output})
}

func microCharge(input, output, inputRate, outputRate int64) (int64, error) {
	return microAmount(input, output, inputRate, outputRate, 500000)
}

func microReserve(input, output, inputRate, outputRate int64) (int64, error) {
	return microAmount(input, output, inputRate, outputRate, 999999)
}

func microAmount(input, output, inputRate, outputRate, rounding int64) (int64, error) {
	if input < 0 || output < 0 || inputRate < 0 || outputRate < 0 {
		return 0, errors.New("invalid_amount")
	}
	amount := new(big.Int).Mul(big.NewInt(input), big.NewInt(inputRate))
	amount.Add(amount, new(big.Int).Mul(big.NewInt(output), big.NewInt(outputRate)))
	amount.Add(amount, big.NewInt(rounding))
	amount.Div(amount, big.NewInt(1000000))
	if !amount.IsInt64() {
		return 0, errors.New("amount_overflow")
	}
	return amount.Int64(), nil
}

func limitsFor(transaction *sql.Tx, ownerType, ownerID string) (ResourceLimits, error) {
	var value ResourceLimits
	err := transaction.QueryRow("SELECT calls,rpm,tpm,concurrency,budget,hard,revision FROM resource_limits WHERE owner_type=? AND owner_id=?", ownerType, ownerID).Scan(&value.Calls, &value.RPM, &value.TPM, &value.Concurrency, &value.Budget, &value.Hard, &value.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return ResourceLimits{Hard: true}, nil
	}
	return value, err
}

func usageFor(transaction *sql.Tx, ownerType, ownerID, period, minute string) (ResourceUsage, error) {
	column := "tenant_id"
	if ownerType == "key" {
		column = "key_id"
	}
	var value ResourceUsage
	err := transaction.QueryRow("SELECT coalesce(sum(CASE WHEN period=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN minute=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN running=1 THEN 1 ELSE 0 END),0),coalesce(sum(tokens_reserved),0),coalesce(sum(fee_reserved),0),coalesce(sum(pending),0) FROM admissions WHERE "+column+"=?", period, minute, ownerID).Scan(&value.Calls, &value.RPM, &value.Concurrency, &value.TPM, &value.Reserved, &value.Pending)
	if err != nil {
		return value, err
	}
	var knownTokens int64
	err = transaction.QueryRow("SELECT coalesce(sum(CASE WHEN a.minute=? THEN c.input_tokens+c.output_tokens ELSE 0 END),0),coalesce(sum(CASE WHEN c.period=? THEN c.fee ELSE 0 END),0) FROM charge_entries c JOIN admissions a ON a.request_id=c.request_id WHERE a."+column+"=?", minute, period, ownerID).Scan(&knownTokens, &value.Fee)
	if err != nil {
		return value, err
	}
	value.TPM += knownTokens
	var legacyCalls, legacyRPM int64
	if transaction.QueryRow("SELECT coalesce(sum(CASE WHEN l.period=? THEN 1 ELSE 0 END),0),coalesce(sum(CASE WHEN substr(r.time,1,16)=? THEN 1 ELSE 0 END),0) FROM legacy_pending l JOIN requests r ON r.id=l.request_id JOIN key_controls c ON c.key_id=r.key_id WHERE l.resolved=0 AND c."+column+"=?", period, minute, ownerID).Scan(&legacyCalls, &legacyRPM) != nil {
		return value, errors.New("historical usage scan failed")
	}
	value.Calls += legacyCalls
	value.RPM += legacyRPM
	if ownerType == "tenant" {
		var adjustment int64
		err = transaction.QueryRow("SELECT coalesce(sum(c.amount),0) FROM bill_corrections c JOIN bills b ON b.id=c.bill_id WHERE b.tenant_id=? AND c.period=?", ownerID, period).Scan(&adjustment)
		value.Fee += adjustment
	}
	return value, err
}

func overLimit(limit *int64, used, additional int64) bool {
	return limit != nil && (used > *limit || additional > *limit-used)
}

func admissionClock(transaction *sql.Tx, now time.Time) time.Time {
	var latest sql.NullString
	if transaction.QueryRow("SELECT accepted FROM admissions ORDER BY julianday(accepted) DESC LIMIT 1").Scan(&latest) == nil && latest.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, latest.String)
		if err == nil && parsed.After(now) {
			return parsed
		}
	}
	return now
}

func (store *Store) admit(record Record, snapshot routeSnapshot, payload map[string]json.RawMessage, request *http.Request) (string, int, string) {
	transaction, err := store.DB.Begin()
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	defer transaction.Rollback()
	credential := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	var currentlyAuthorized int
	if keyID, valid := generationIdentity(transaction, credential, store.now()); !valid || keyID != record.KeyID {
		return "unauthorized", 401, ""
	}
	if transaction.QueryRow("SELECT count(*) FROM key_model_grants k JOIN key_controls c ON c.key_id=k.key_id JOIN tenant_model_grants g ON g.alias=k.alias AND g.tenant_id=c.tenant_id WHERE k.key_id=? AND k.alias=?", record.KeyID, record.Alias).Scan(&currentlyAuthorized) != nil {
		return "storage_unavailable", 503, ""
	}
	if currentlyAuthorized != 1 {
		return "model_forbidden", 403, ""
	}
	operation := request.Header.Get("Idempotency-Key")
	var fingerprint []byte
	if operation != "" {
		if !validName(operation) {
			return "invalid_idempotency_key", 400, ""
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return "invalid_request", 400, ""
		}
		fingerprint = digest(string(encoded))
		var prior string
		var saved []byte
		err = transaction.QueryRow("SELECT fingerprint,request_id FROM generation_operations WHERE key_id=? AND operation=?", record.KeyID, operation).Scan(&saved, &prior)
		if err == nil {
			if !bytes.Equal(saved, fingerprint) {
				return "idempotency_conflict", 409, ""
			}
			return "logical_call_already_exists", 409, prior
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "storage_unavailable", 503, ""
		}
	}
	var tenantID string
	var active bool
	if transaction.QueryRow("SELECT c.tenant_id,t.active FROM key_controls c JOIN tenants t ON t.id=c.tenant_id WHERE key_id=?", record.KeyID).Scan(&tenantID, &active) != nil {
		return "storage_unavailable", 503, ""
	}
	if !active {
		return "tenant_paused", 403, ""
	}
	tenantLimit, err := limitsFor(transaction, "tenant", tenantID)
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	keyLimit, err := limitsFor(transaction, "key", record.KeyID)
	if err != nil {
		return "storage_unavailable", 503, ""
	}
	bounded := tenantLimit.TPM != nil || keyLimit.TPM != nil || tenantLimit.Hard && tenantLimit.Budget != nil || keyLimit.Hard && keyLimit.Budget != nil
	if bounded {
		var pending int
		if transaction.QueryRow("SELECT count(*) FROM legacy_pending l WHERE l.tenant_id=? AND l.resolved=0", tenantID).Scan(&pending) != nil {
			return "storage_unavailable", 503, ""
		}
		if pending > 0 {
			return "historical_usage_pending", 503, ""
		}
		if transaction.QueryRow("SELECT count(*) FROM admissions a JOIN request_prices p ON p.request_id=a.request_id WHERE a.tenant_id=? AND a.pending>0 AND (p.max_input=0 OR p.max_output=0)", tenantID).Scan(&pending) != nil {
			return "storage_unavailable", 503, ""
		}
		if pending > 0 {
			return "unbounded_usage_pending", 503, ""
		}
	}
	now := admissionClock(transaction, store.now())
	prices := []PriceSnapshot{}
	for _, candidate := range snapshot.Model.Candidates {
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		if !candidate.Enabled || !upstream.Enabled {
			continue
		}
		value := PriceSnapshot{Upstream: candidate.UpstreamID}
		err := transaction.QueryRow("SELECT id,cost_input,cost_output,tenant_input,tenant_output FROM prices WHERE alias=? AND upstream_id=? AND upstream_execution=? AND candidate_execution=? AND julianday(effective)<=julianday(?) ORDER BY julianday(effective) DESC,id DESC LIMIT 1", record.Alias, candidate.UpstreamID, upstream.ExecutionRevision, candidate.ExecutionRevision, now.Format(time.RFC3339Nano)).Scan(&value.ID, &value.CostIn, &value.CostOut, &value.FeeIn, &value.FeeOut)
		var legacy bool
		if transaction.QueryRow("SELECT legacy FROM key_controls WHERE key_id=?", record.KeyID).Scan(&legacy) != nil {
			return "storage_unavailable", 503, ""
		}
		if errors.Is(err, sql.ErrNoRows) && legacy && !bounded {
			value.ID = "legacy_unpriced"
		} else if err != nil {
			return "price_not_ready", 503, ""
		}
		var evidence string
		err = transaction.QueryRow("SELECT max_input,max_output,evidence FROM model_bounds WHERE alias=? AND upstream_id=? AND upstream_execution=? AND candidate_execution=?", record.Alias, candidate.UpstreamID, upstream.ExecutionRevision, candidate.ExecutionRevision).Scan(&value.Input, &value.Output, &evidence)
		if err != nil && (!errors.Is(err, sql.ErrNoRows) || bounded) {
			return "trusted_bound_required", 503, ""
		}
		if bounded {
			if evidence == "" || value.Input <= 0 || value.Output <= 0 {
				return "trusted_bound_required", 503, ""
			}
			var alternatives int64 = 1
			if raw := payload["n"]; raw != nil {
				if json.Unmarshal(raw, &alternatives) != nil || alternatives != 1 {
					return "unsupported_bounded_request", 400, ""
				}
			}
		}
		prices = append(prices, value)
	}
	if len(prices) == 0 {
		return "upstream_unavailable", 503, ""
	}
	if bounded {
		maximumOutput := prices[0].Output
		for _, price := range prices {
			maximumOutput = min(maximumOutput, price.Output)
		}
		if !boundedPayload(payload, maximumOutput) {
			return "unsupported_bounded_request", 400, ""
		}
	}
	tokenBounds, feeBounds := []int64{}, []int64{}
	for _, price := range prices {
		fee, err := microReserve(price.Input, price.Output, price.FeeIn, price.FeeOut)
		if err != nil || price.Input > 1000000000 || price.Output > 1000000000 {
			return "amount_overflow", 503, ""
		}
		tokenBounds = append(tokenBounds, price.Input+price.Output)
		feeBounds = append(feeBounds, fee)
	}
	if operation != "" {
		if _, err := transaction.Exec("INSERT INTO generation_operations VALUES(?,?,?,?)", record.KeyID, operation, fingerprint, record.ID); err != nil {
			return "storage_unavailable", 503, ""
		}
	}
	sort.Slice(tokenBounds, func(left, right int) bool { return tokenBounds[left] > tokenBounds[right] })
	sort.Slice(feeBounds, func(left, right int) bool { return feeBounds[left] > feeBounds[right] })
	var tokensReserved, feeReserved int64
	for index := 0; index < len(tokenBounds) && index < 3; index++ {
		tokensReserved += tokenBounds[index]
		feeReserved += feeBounds[index]
	}
	period, minute := now.Format("2006-01"), now.Format("2006-01-02T15:04")
	for _, owner := range []struct {
		kind, id string
		limits   ResourceLimits
	}{{"tenant", tenantID, tenantLimit}, {"key", record.KeyID, keyLimit}} {
		used, err := usageFor(transaction, owner.kind, owner.id, period, minute)
		if err != nil {
			return "storage_unavailable", 503, ""
		}
		if overLimit(owner.limits.Calls, used.Calls, 1) {
			return owner.kind + "_calls_exceeded", 429, ""
		}
		if overLimit(owner.limits.RPM, used.RPM, 1) {
			return owner.kind + "_rpm_exceeded", 429, ""
		}
		if overLimit(owner.limits.TPM, used.TPM, tokensReserved) {
			return owner.kind + "_tpm_exceeded", 429, ""
		}
		if overLimit(owner.limits.Concurrency, used.Concurrency, 1) {
			return owner.kind + "_concurrency_exceeded", 429, ""
		}
		if owner.limits.Hard && overLimit(owner.limits.Budget, used.Fee+used.Reserved, feeReserved) {
			return owner.kind + "_budget_exceeded", 429, ""
		}
	}
	if _, err := transaction.Exec("INSERT INTO admissions VALUES(?,?,?,?,?,?,1,?,?,0,1)", record.ID, tenantID, record.KeyID, period, minute, now.Format(time.RFC3339Nano), tokensReserved, feeReserved); err != nil {
		return "storage_unavailable", 503, ""
	}
	for _, price := range prices {
		if _, err := transaction.Exec("INSERT INTO request_prices VALUES(?,?,?,?,?,?,?,?,?)", record.ID, price.Upstream, price.ID, price.CostIn, price.CostOut, price.FeeIn, price.FeeOut, price.Input, price.Output); err != nil {
			return "storage_unavailable", 503, ""
		}
	}
	if transaction.Commit() != nil {
		return "storage_unavailable", 503, ""
	}
	return "", 200, ""
}

func boundedPayload(payload map[string]json.RawMessage, maximumOutput int64) bool {
	allowed := map[string]bool{"model": true, "messages": true, "stream": true, "stream_options": true, "max_tokens": true, "max_completion_tokens": true, "n": true, "temperature": true, "top_p": true, "seed": true, "stop": true}
	for field := range payload {
		if !allowed[field] {
			return false
		}
	}
	var messages []map[string]json.RawMessage
	if json.Unmarshal(payload["messages"], &messages) != nil || len(messages) == 0 {
		return false
	}
	for _, message := range messages {
		var role, content string
		if len(message) != 2 || json.Unmarshal(message["role"], &role) != nil || json.Unmarshal(message["content"], &content) != nil || (role != "system" && role != "user" && role != "assistant") {
			return false
		}
	}
	if payload["max_tokens"] != nil && payload["max_completion_tokens"] != nil {
		return false
	}
	field := "max_tokens"
	if payload["max_completion_tokens"] != nil {
		field = "max_completion_tokens"
	}
	if raw := payload[field]; raw != nil {
		var requested int64
		return json.Unmarshal(raw, &requested) == nil && requested > 0 && requested <= maximumOutput
	}
	payload[field], _ = json.Marshal(maximumOutput)
	return true
}

func settleAdmission(transaction *sql.Tx, requestID string) error {
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
		price                PriceSnapshot
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
		cost, err := microCharge(input, output, value.price.CostIn, value.price.CostOut)
		if err != nil {
			return err
		}
		fee, err := microCharge(input, output, value.price.FeeIn, value.price.FeeOut)
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
			heldTotal, err := microReserve(input+unknownInput, output+unknownOutput, value.price.FeeIn, value.price.FeeOut)
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
			if _, err := transaction.Exec("INSERT INTO charge_entries VALUES(?,?,?,?,?,?,?,?,?,?,?)", randomSecret(16), requestID, value.ordinal, revision, period, input-priorInput, output-priorOutput, cost-priorCost, fee-priorFee, reason, evidence); err != nil {
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
