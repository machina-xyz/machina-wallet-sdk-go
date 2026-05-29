package machinawallet

// PolicyBuilder is a typed factory for common SpendPolicy shapes. The rule
// JSON shapes mirror the canonical policy_schema_v1 used by the wallet
// management service.
//
// Use the zero value:
//
//	pb := machinawallet.PolicyBuilder{}
//	p := pb.DailySpendLimit("1000000000", "USDC")
type PolicyBuilder struct{}

// DailySpendLimit caps total spend per UTC day at maxAtoms of the given
// currency. currency is a token mint string or an empty string for the native
// asset.
func (PolicyBuilder) DailySpendLimit(maxAtoms, currency string) NewSpendPolicy {
	rule := map[string]any{
		"kind":      "daily_spend_limit",
		"max_atoms": maxAtoms,
		"currency":  currency,
		"window":    "1d",
	}
	return NewSpendPolicy{
		Name:          "Daily spend limit",
		SchemaVersion: "v1",
		Rules:         []map[string]any{rule},
	}
}

// AllowedRecipients restricts outgoing transactions to the given addresses.
func (PolicyBuilder) AllowedRecipients(addresses []string) NewSpendPolicy {
	rule := map[string]any{
		"kind":  "allowed_recipients",
		"addrs": addresses,
	}
	return NewSpendPolicy{
		Name:          "Allowed recipients",
		SchemaVersion: "v1",
		Rules:         []map[string]any{rule},
	}
}

// BlockContractCalls denies any transaction that targets a non-EOA address.
func (PolicyBuilder) BlockContractCalls() NewSpendPolicy {
	rule := map[string]any{
		"kind":   "block_contract_calls",
		"effect": "deny",
	}
	return NewSpendPolicy{
		Name:          "Block contract calls",
		SchemaVersion: "v1",
		Rules:         []map[string]any{rule},
	}
}

// RequireApprovalOver routes transactions above valueAtoms to manual approval
// instead of immediate signing.
func (PolicyBuilder) RequireApprovalOver(valueAtoms string) NewSpendPolicy {
	rule := map[string]any{
		"kind":             "require_approval_over",
		"threshold_atoms":  valueAtoms,
		"approval_channel": "default",
	}
	return NewSpendPolicy{
		Name:          "Require approval over threshold",
		SchemaVersion: "v1",
		Rules:         []map[string]any{rule},
	}
}

// DenyUnlimitedApprovals denies ERC-20 approve transactions with allowance
// equal to the max-uint sentinel — a common phishing vector.
func (PolicyBuilder) DenyUnlimitedApprovals() NewSpendPolicy {
	rule := map[string]any{
		"kind":   "deny_unlimited_approvals",
		"effect": "deny",
	}
	return NewSpendPolicy{
		Name:          "Deny unlimited approvals",
		SchemaVersion: "v1",
		Rules:         []map[string]any{rule},
	}
}
