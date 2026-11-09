package finance_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"model-relay/internal/finance"
)

func TestFrozenPriceReadDoesNotFollowCurrentRouteExecutions(t *testing.T) {
	f := install(t)
	f.exec(t, "INSERT INTO prices VALUES('old','model',1,'2026-11-01T00:00:00Z',1,2,10,20,7,11)")
	transaction, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.Exec("UPDATE upstream SET execution_revision=9 WHERE id=1; UPDATE route_candidates SET execution_revision=13 WHERE alias='model' AND upstream_id=1"); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	frozen, effective, err := finance.PriceAt(transaction, "model", 1, 7, 11, at)
	if err != nil || frozen.Upstream != 1 || frozen.ID != "old" || frozen.CostIn != 1 || frozen.CostOut != 2 || frozen.FeeIn != 10 || frozen.FeeOut != 20 || effective != "2026-11-01T00:00:00Z" {
		t.Fatal("frozen snapshot followed current execution", frozen, effective, err)
	}
	input, output, currentEffective, err := finance.CurrentModelRate(transaction, "model", at)
	if !errors.Is(err, sql.ErrNoRows) || input != nil || output != nil || currentEffective != nil {
		t.Fatal("stale execution advertised as current", input, output, currentEffective, err)
	}
	if _, err = transaction.Exec("INSERT INTO prices VALUES('current','model',1,'2026-11-02T00:00:00Z',3,4,30,40,9,13)"); err != nil {
		t.Fatal(err)
	}
	input, output, currentEffective, err = finance.CurrentModelRate(transaction, "model", at)
	if err != nil || input == nil || *input != 30 || output == nil || *output != 40 || currentEffective == nil || *currentEffective != "2026-11-02T00:00:00Z" {
		t.Fatal("current price ignored transaction-local execution", input, output, currentEffective, err)
	}
	frozen, _, err = finance.PriceAt(transaction, "model", 1, 7, 11, at)
	if err != nil || frozen.ID != "old" {
		t.Fatal("new current price replaced frozen selection", frozen, err)
	}
}

func TestPriceReadsKeepInclusiveCutoffAndEqualInstantIDOrder(t *testing.T) {
	f := install(t)
	f.exec(t, `INSERT INTO prices VALUES
('earlier','model',1,'2026-11-09T11:59:59Z',1,2,60,61,7,11),
('same-a','model',1,'2026-11-09T12:00:00Z',3,4,10,11,7,11),
('same-z','model',1,'2026-11-09T13:00:00+01:00',5,6,20,21,7,11),
('same-zz','model',2,'2026-11-09T12:00:00Z',7,8,22,23,8,12),
('future','model',1,'2026-11-09T12:00:01Z',9,10,30,31,7,11);`)
	transaction, err := f.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	for _, scenario := range []struct {
		second      int
		frozenID    string
		frozenRate  int64
		currentRate int64
	}{
		{second: -1, frozenID: "earlier", frozenRate: 60, currentRate: 60},
		{second: 0, frozenID: "same-z", frozenRate: 20, currentRate: 22},
		{second: 1, frozenID: "future", frozenRate: 30, currentRate: 30},
	} {
		at := time.Date(2026, 11, 9, 12, 0, scenario.second, 0, time.UTC)
		price, _, err := finance.PriceAt(transaction, "model", 1, 7, 11, at)
		if err != nil || price.ID != scenario.frozenID || price.FeeIn != scenario.frozenRate {
			t.Fatal("frozen cutoff/order changed", scenario.second, price, err)
		}
		input, _, _, err := finance.CurrentModelRate(transaction, "model", at)
		if err != nil || input == nil || *input != scenario.currentRate {
			t.Fatal("current candidate cutoff/order changed", scenario.second, input, err)
		}
	}
	before := time.Date(2026, 11, 9, 11, 59, 58, 0, time.UTC)
	price, effective, err := finance.PriceAt(transaction, "model", 1, 7, 11, before)
	if !errors.Is(err, sql.ErrNoRows) || price.Upstream != 1 || price.ID != "" || effective != "" {
		t.Fatal("missing frozen price lost legacy fallback identity", price, effective, err)
	}
	input, output, effectiveRate, err := finance.CurrentModelRate(transaction, "model", before)
	if !errors.Is(err, sql.ErrNoRows) || input != nil || output != nil || effectiveRate != nil {
		t.Fatal("missing current price invented rate", input, output, effectiveRate, err)
	}
}
