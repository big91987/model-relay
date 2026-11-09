package installation

import (
	"testing"
	"time"
)

func interruptedConsole(t *testing.T) (string, string) {
	t.Helper()
	store, _, directory, keyFile := newInstallation(t, true)
	accepted := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store.SetClock(func() time.Time { return accepted })
	if err := store.AdvanceClock(); err != nil {
		t.Fatal(err)
	}
	_, err := store.DB.Exec(`INSERT INTO tenants VALUES('recovery','Recovery',1,'2026-01-01T00:00:00Z');
INSERT INTO app_keys VALUES('key','Application',x'1234','hint',1,'2026-01-01T00:00:00Z');
INSERT INTO key_controls(key_id,tenant_id,application) VALUES('key','recovery','Application');
INSERT INTO models VALUES('alpha',1,1,NULL);
INSERT INTO upstream VALUES(1,'Provider','https://public.example/v1',x'',1,1,1);
INSERT INTO requests(id,time,key_id,alias,model,stream,revision,state,error_code,evidence_version) VALUES('known','2026-01-01T00:00:00Z','key','alpha','remote',0,1,'started','',2),('unknown','2026-01-01T00:00:00Z','key','alpha','remote',0,1,'started','',2);
INSERT INTO request_attempts(request_id,ordinal,upstream_id,upstream_name,remote_model,upstream_revision,upstream_execution_revision,candidate_execution_revision,started_at,state,error_code,prompt_tokens,completion_tokens,total_tokens) VALUES('known',1,1,'Provider','remote',1,1,1,'2026-01-01T00:00:00Z','started','',3,4,7),('unknown',1,1,'Provider','remote',1,1,1,'2026-01-01T00:00:00Z','started','',NULL,NULL,NULL);
UPDATE attempt_dispatches SET possible=1;
INSERT INTO admissions VALUES('known','recovery','key','2026-01','2026-01-01T00:00','2026-01-01T00:00:00Z',1,20,50,0,1),('unknown','recovery','key','2026-01','2026-01-01T00:00','2026-01-01T00:00:00Z',1,20,50,0,1);
INSERT INTO request_prices VALUES('known',1,'price',1000000,1000000,2000000,3000000,10,10),('unknown',1,'price',1000000,1000000,2000000,3000000,10,10);
INSERT INTO upstream_checks(upstream_id,execution_revision,state,started_at,error_code) VALUES(1,1,'checking','2026-01-01T00:00:00Z','');`)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	return directory, keyFile
}

func TestOpenRecoversInterruptedAccountingOnce(t *testing.T) {
	directory, keyFile := interruptedConsole(t)
	for repeat := 0; repeat < 2; repeat++ {
		store, err := Open(directory, keyFile)
		if err != nil {
			t.Fatal(err)
		}
		var running, pending, tokens, fees int64
		if err = store.DB.QueryRow("SELECT running,pending,tokens_reserved,fee_reserved FROM admissions WHERE request_id='known'").Scan(&running, &pending, &tokens, &fees); err != nil || running != 0 || pending != 0 || tokens != 0 || fees != 0 {
			t.Fatal("known interrupted usage was not settled", running, pending, tokens, fees, err)
		}
		if err = store.DB.QueryRow("SELECT running,pending,tokens_reserved,fee_reserved FROM admissions WHERE request_id='unknown'").Scan(&running, &pending, &tokens, &fees); err != nil || running != 0 || pending != 1 || tokens != 20 || fees != 50 {
			t.Fatal("unknown interrupted usage lost reservations", running, pending, tokens, fees, err)
		}
		var cost, fee, entries, audits int64
		if err = store.DB.QueryRow("SELECT sum(cost),sum(fee),count(*) FROM charge_entries WHERE request_id='known'").Scan(&cost, &fee, &entries); err != nil || cost != 7 || fee != 18 || entries != 1 {
			t.Fatal("known recovery ledger changed", cost, fee, entries, err)
		}
		if err = store.DB.QueryRow("SELECT count(*) FROM console_audit WHERE actor='system' AND action='request.recover'").Scan(&audits); err != nil || audits != 2 {
			t.Fatal("recovery audit repeated or missing", audits, err)
		}
		var requests, attempts, checks int
		if store.DB.QueryRow("SELECT count(*) FROM requests WHERE state='process_interrupted' AND error_code='process_interrupted' AND duration_ms IS NULL AND attempts_complete=0").Scan(&requests) != nil || requests != 2 {
			t.Fatal("interrupted request terminal changed")
		}
		if store.DB.QueryRow("SELECT count(*) FROM request_attempts WHERE state='process_interrupted' AND error_code='process_interrupted'").Scan(&attempts) != nil || attempts != 2 {
			t.Fatal("interrupted attempt terminal changed")
		}
		if store.DB.QueryRow("SELECT count(*) FROM upstream_checks WHERE state='unknown' AND error_code='process_interrupted'").Scan(&checks) != nil || checks != 1 {
			t.Fatal("interrupted check incorrectly remained active")
		}
		store.Close()
	}
}

func TestRecoveryAuditFailureRollsBackSettlementAndCanResume(t *testing.T) {
	directory, keyFile := interruptedConsole(t)
	db, err := database(directory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TRIGGER reject_recovery BEFORE INSERT ON console_audit WHEN NEW.action='request.recover' BEGIN SELECT RAISE(FAIL,'controlled'); END;`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if store, err := Open(directory, keyFile); err == nil {
		store.Close()
		t.Fatal("failed recovery audit accepted")
	}
	db, err = database(directory)
	if err != nil {
		t.Fatal(err)
	}
	var running, entries int
	if db.QueryRow("SELECT count(*) FROM admissions WHERE running=1").Scan(&running) != nil || running != 2 || db.QueryRow("SELECT count(*) FROM charge_entries").Scan(&entries) != nil || entries != 0 {
		t.Fatal("partial settlement escaped recovery rollback", running, entries)
	}
	if _, err = db.Exec("DROP TRIGGER reject_recovery"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	store, err := Open(directory, keyFile)
	if err != nil {
		t.Fatal("recovery retry failed", err)
	}
	defer store.Close()
	if store.DB.QueryRow("SELECT count(*) FROM admissions WHERE running=1").Scan(&running) != nil || running != 0 {
		t.Fatal("recovery retry left admissions running")
	}
}
