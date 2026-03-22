package kubeid

import "testing"

func TestIDString(t *testing.T) {
	id := ID{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Namespace: "payments",
		Name:      "ledger-api",
	}

	got := id.String()
	want := "apps/v1/Deployment:payments/ledger-api"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
