package latencyfixture

import "testing"

func TestControlledSetupUsesPublicManagement(t *testing.T) {
	fixture, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Close()
	if err = fixture.SetupUpstreams(); err != nil {
		t.Fatal(err)
	}
	fixture.Register("controlled", &Scenario{})
	if err = fixture.Configure("controlled", "low_latency"); err != nil {
		t.Fatal(err)
	}
	if err = fixture.IssueKey([]string{"controlled"}); err != nil {
		t.Fatal(err)
	}
}
