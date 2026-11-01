package relay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
)

func TestQAIndependentBearerRFC6750PositiveForms(t *testing.T) {
	forms := []string{"Bearer %s", "bearer %s", "BEARER %s", "bEaReR %s", "BeArEr %s", "Bearer  %s", "Bearer    %s"}
	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			fixture := governedInstallation(t, false)
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:5544/v1/models", nil)
			request.Header.Set("Authorization", sprintf(form, fixture.key))
			writer := httptest.NewRecorder()
			fixture.server.ServeHTTP(writer, request)
			if writer.Code != http.StatusOK {
				t.Fatalf("RFC 6750 Bearer form %q status = %d, want 200; body=%s", form, writer.Code, writer.Body.String())
			}
			if fixture.hits.Load() != 0 {
				t.Fatalf("models request dispatched upstream %d times", fixture.hits.Load())
			}
		})
	}
}

func sprintf(format, value string) string {
	for index := 0; index+1 < len(format); index++ {
		if format[index:index+2] == "%s" {
			return format[:index] + value + format[index+2:]
		}
	}
	return format
}

func TestQAIndependentClockPersistenceFailures(t *testing.T) {
	tests := []struct {
		name   string
		stage  clockFaultStage
		counts [4]int64
	}{
		{"normal-delegation", clockFaultNone, [4]int64{1, 1, 1, 1}},
		{"begin", clockFaultBegin, [4]int64{1, 0, 0, 0}},
		{"read", clockFaultRead, [4]int64{1, 1, 0, 0}},
		{"update", clockFaultUpdate, [4]int64{1, 1, 1, 0}},
		{"commit", clockFaultCommit, [4]int64{1, 1, 1, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newClockFaultFixture(t, test.stage)
			if test.stage != clockFaultCommit {
				fixture.state.stage = clockFaultNone
			}
			var persistedBefore string
			if err := fixture.store.DB.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&persistedBefore); err != nil {
				t.Fatal(err)
			}
			fixture.store.clockMu.Lock()
			memoryBefore := fixture.store.effectiveTime
			fixture.store.clockMu.Unlock()
			observed := memoryBefore.Add(2 * time.Hour)
			fixture.store.clock = func() time.Time { return observed }
			fixture.state.counts.begin.Store(0)
			fixture.state.counts.read.Store(0)
			fixture.state.counts.update.Store(0)
			fixture.state.counts.commit.Store(0)
			fixture.state.commitHookHits.Store(0)
			fixture.state.stage = test.stage

			err := fixture.store.advanceClock()
			got := [4]int64{fixture.state.counts.begin.Load(), fixture.state.counts.read.Load(), fixture.state.counts.update.Load(), fixture.state.counts.commit.Load()}
			if got != test.counts {
				t.Fatalf("phase counts = %v, want %v", got, test.counts)
			}
			if test.stage == clockFaultNone {
				if err != nil {
					t.Fatal(err)
				}
			} else if test.stage == clockFaultCommit {
				var sqliteError sqlite3.Error
				if !errors.As(err, &sqliteError) || sqliteError.ExtendedCode != sqlite3.ErrConstraintCommitHook || fixture.state.commitHookHits.Load() != 1 {
					t.Fatalf("commit failure = %#v, hook hits=%d", err, fixture.state.commitHookHits.Load())
				}
			} else if !errors.Is(err, errClockFault) || fixture.state.commitHookHits.Load() != 0 {
				t.Fatalf("phase failure = %#v, hook hits=%d", err, fixture.state.commitHookHits.Load())
			}

			fixture.state.stage = clockFaultNone
			var persistedAfter string
			if err := fixture.store.DB.QueryRow("SELECT effective FROM console_clock WHERE id=1").Scan(&persistedAfter); err != nil {
				t.Fatal(err)
			}
			fixture.store.clockMu.Lock()
			memoryAfter := fixture.store.effectiveTime
			fixture.store.clockMu.Unlock()
			if test.stage == clockFaultNone {
				if persistedAfter != observed.Format(time.RFC3339Nano) || !memoryAfter.Equal(observed) {
					t.Fatalf("successful clock not published: persisted=%q memory=%s", persistedAfter, memoryAfter)
				}
				return
			}
			if persistedAfter != persistedBefore || !memoryAfter.Equal(memoryBefore) {
				t.Fatalf("failed clock changed state: persisted %q -> %q, memory %s -> %s", persistedBefore, persistedAfter, memoryBefore, memoryAfter)
			}

			fixture.state.stage = test.stage
			server := New(fixture.store, Config{Authority: "relay.test", Clock: func() time.Time { return observed }})
			writer := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "http://relay.test/v1/models", nil)
			request.Host = "relay.test"
			server.ServeHTTP(writer, request)
			if writer.Code != http.StatusServiceUnavailable || !server.degraded.Load() {
				t.Fatalf("handler status=%d degraded=%v body=%s", writer.Code, server.degraded.Load(), writer.Body.String())
			}
		})
	}
}

func TestQAIndependentRestoreFinalRenameAtomicity(t *testing.T) {
	archive, keyFile := validRestoreBackup(t)
	sentinel := errors.New("qa final rename sentinel")
	for _, existing := range []bool{false, true} {
		for iteration := 0; iteration < 10; iteration++ {
			name := "absent"
			if existing {
				name = "existing"
			}
			t.Run(name+"/"+time.Unix(int64(iteration), 0).UTC().Format("05"), func(t *testing.T) {
				parent := t.TempDir()
				target := filepath.Join(parent, "target")
				var before restoreTargetSnapshot
				if existing {
					if err := os.Mkdir(target, 0750); err != nil {
						t.Fatal(err)
					}
					before = snapshotRestoreTarget(t, target)
				}
				calls := 0
				err := restoreWithRename(archive, target, keyFile, func(source, destination string) error {
					calls++
					if filepath.Dir(source) != parent || destination != target || filepath.Base(source)[:len(".restore-")] != ".restore-" {
						t.Fatalf("unexpected rename %q -> %q", source, destination)
					}
					return sentinel
				})
				if !errors.Is(err, sentinel) || calls != 1 {
					t.Fatalf("error=%v calls=%d, want sentinel and one call", err, calls)
				}
				if existing {
					afterInfo, statErr := os.Stat(target)
					if statErr != nil {
						t.Fatalf("existing target removed: %v", statErr)
					}
					after := snapshotRestoreTarget(t, target)
					if !os.SameFile(before.info, afterInfo) || before.mode != after.mode || before.digest != after.digest || !reflect.DeepEqual(before.entries, after.entries) {
						t.Fatalf("existing target changed: before=%v/%x after=%v/%x", before.entries, before.digest, after.entries, after.digest)
					}
				} else if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
					t.Fatalf("absent target installed: %v", statErr)
				}
				assertNoRestoreTemporary(t, parent)
			})
		}
	}
}
