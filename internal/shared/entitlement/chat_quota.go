package entitlement

import "context"

type ChatQuotaResult struct {
	Allowed     bool
	Limit       int
	Used        int
	Remaining   int
	IsUnlimited bool
}

type ChatQuotaChecker interface {
	ConsumeChatQuota(ctx context.Context, userID uint) (*ChatQuotaResult, error)
}
