package catalog

import (
	"context"
	"strconv"
	"strings"

	"model-relay/internal/fault"
	"model-relay/internal/names"
	"model-relay/internal/operations"
	"model-relay/internal/secure"
)

func validateUpstream(input UpstreamInput) error {
	if input.Enabled == nil || !names.Valid(strings.TrimSpace(input.Name)) || len(input.Token) > 8192 || strings.ContainsAny(input.Token, "\r\n\x00") {
		return fault.New(fault.Invalid, "invalid_request")
	}
	if input.SourceKind != "" && input.SourceKind != "provider" && input.SourceKind != "self_hosted" && input.SourceKind != "unspecified" {
		return fault.New(fault.Invalid, "invalid_source_kind")
	}
	return nil
}
func (service *Service) CreateUpstream(ctx context.Context, actor string, input UpstreamInput) (Upstream, error) {
	if err := validateUpstream(input); err != nil {
		return Upstream{}, err
	}
	normalized, err := service.policy.URL(ctx, input.URL)
	if err != nil {
		return Upstream{}, fault.New(fault.Invalid, "invalid_destination")
	}
	service.state.Lock()
	defer service.state.Unlock()
	if input.Token == "" {
		return Upstream{}, fault.New(fault.Invalid, "new_credential_required")
	}
	return service.writeUpstream(ctx, actor, Upstream{}, input, normalized)
}
func (service *Service) UpdateUpstream(ctx context.Context, actor string, id int64, input UpstreamInput) (Upstream, error) {
	if err := validateUpstream(input); err != nil {
		return Upstream{}, err
	}
	if input.Expected == nil {
		return Upstream{}, fault.New(fault.Invalid, "invalid_request")
	}
	normalized, err := service.policy.URL(ctx, input.URL)
	if err != nil {
		return Upstream{}, fault.New(fault.Invalid, "invalid_destination")
	}
	service.state.Lock()
	defer service.state.Unlock()
	current, err := service.UpstreamFacts(ctx, id)
	if err != nil {
		return Upstream{}, readError(err)
	}
	if *input.Expected != current.Revision {
		return Upstream{}, &RevisionConflict{CurrentRevision: current.Revision}
	}
	if input.Token == "" && (normalized != current.URL || !current.CredentialConfigured) {
		return Upstream{}, fault.New(fault.Invalid, "new_credential_required")
	}
	return service.writeUpstream(ctx, actor, current, input, normalized)
}

// writeUpstream commits credentials, profiles, audit and configuration revision
// together. Its caller holds coordination from validation through publication.
func (service *Service) writeUpstream(ctx context.Context, actor string, current Upstream, input UpstreamInput, normalized string) (Upstream, error) {
	transaction, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return Upstream{}, unavailable()
	}
	defer transaction.Rollback()
	var duplicate bool
	if err = transaction.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM upstream WHERE name=? AND id<>?)", strings.TrimSpace(input.Name), current.ID).Scan(&duplicate); err != nil {
		return Upstream{}, unavailable()
	}
	if duplicate {
		return Upstream{}, fault.New(fault.Conflict, "name_conflict")
	}
	id, execution, revision := current.ID, current.ExecutionRevision, current.Revision+1
	if id == 0 {
		result, err := transaction.ExecContext(ctx, "INSERT INTO upstream(name,url,token,enabled,revision,execution_revision) VALUES(?,?,?, ?,1,1)", strings.TrimSpace(input.Name), normalized, []byte{}, *input.Enabled)
		if err != nil {
			return Upstream{}, unavailable()
		}
		id, err = result.LastInsertId()
		if err != nil {
			return Upstream{}, unavailable()
		}
		execution = 1
	} else if normalized != current.URL || input.Token != "" {
		execution++
	}
	encrypted := current.Encrypted
	if input.Token != "" {
		encrypted, err = secure.Seal(service.key, input.Token, secure.UpstreamAAD(id, normalized))
		if err != nil {
			return Upstream{}, unavailable()
		}
	}
	if _, err = transaction.ExecContext(ctx, "UPDATE upstream SET name=?,url=?,token=?,enabled=?,revision=?,execution_revision=? WHERE id=?", strings.TrimSpace(input.Name), normalized, encrypted, *input.Enabled, revision, execution, id); err != nil {
		return Upstream{}, unavailable()
	}
	if service.schema == 3 {
		kind := input.SourceKind
		if kind == "" {
			kind = current.SourceKind
		}
		if kind == "" {
			kind = "unspecified"
		}
		if _, err = transaction.ExecContext(ctx, "INSERT INTO service_profiles VALUES(?,?) ON CONFLICT(upstream_id) DO UPDATE SET source_kind=excluded.source_kind", id, kind); err != nil {
			return Upstream{}, unavailable()
		}
		if err = operations.Audit(transaction, actor, "", "upstream.write", strconv.FormatInt(id, 10)); err != nil {
			return Upstream{}, unavailable()
		}
	}
	if configChanged(transaction) != nil || transaction.Commit() != nil {
		return Upstream{}, unavailable()
	}
	service.committed()
	value, err := service.UpstreamFacts(ctx, id)
	return value, readError(err)
}
