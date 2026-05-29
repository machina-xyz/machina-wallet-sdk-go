package machinawallet

import (
	"encoding/json"
	"testing"
)

func TestDailySpendLimit(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.DailySpendLimit("1000000000", "USDC")
	if p.SchemaVersion != "v1" {
		t.Errorf("schema = %q", p.SchemaVersion)
	}
	if len(p.Rules) != 1 {
		t.Fatalf("rules len = %d", len(p.Rules))
	}
	r := p.Rules[0]
	if r["kind"] != "daily_spend_limit" {
		t.Errorf("kind = %v", r["kind"])
	}
	if r["max_atoms"] != "1000000000" {
		t.Errorf("max_atoms = %v", r["max_atoms"])
	}
	if r["currency"] != "USDC" {
		t.Errorf("currency = %v", r["currency"])
	}
}

func TestAllowedRecipients(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.AllowedRecipients([]string{"0xabc", "0xdef"})
	r := p.Rules[0]
	if r["kind"] != "allowed_recipients" {
		t.Errorf("kind = %v", r["kind"])
	}
	addrs, ok := r["addrs"].([]string)
	if !ok || len(addrs) != 2 {
		t.Errorf("addrs = %v", r["addrs"])
	}
}

func TestBlockContractCalls(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.BlockContractCalls()
	r := p.Rules[0]
	if r["kind"] != "block_contract_calls" {
		t.Errorf("kind = %v", r["kind"])
	}
	if r["effect"] != "deny" {
		t.Errorf("effect = %v", r["effect"])
	}
}

func TestRequireApprovalOver(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.RequireApprovalOver("10000000000")
	r := p.Rules[0]
	if r["kind"] != "require_approval_over" {
		t.Errorf("kind = %v", r["kind"])
	}
	if r["threshold_atoms"] != "10000000000" {
		t.Errorf("threshold = %v", r["threshold_atoms"])
	}
}

func TestDenyUnlimitedApprovals(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.DenyUnlimitedApprovals()
	if p.Rules[0]["kind"] != "deny_unlimited_approvals" {
		t.Errorf("kind = %v", p.Rules[0]["kind"])
	}
}

func TestPolicyJSONShape(t *testing.T) {
	t.Parallel()

	pb := PolicyBuilder{}
	p := pb.DailySpendLimit("100", "")
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var roundtrip map[string]any
	if err := json.Unmarshal(b, &roundtrip); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if roundtrip["schema_version"] != "v1" {
		t.Errorf("schema_version = %v", roundtrip["schema_version"])
	}
	if _, ok := roundtrip["rules"]; !ok {
		t.Errorf("missing rules in JSON: %s", b)
	}
}
