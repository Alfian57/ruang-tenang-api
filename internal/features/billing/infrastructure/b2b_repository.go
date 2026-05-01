package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOrganizationNotFound       = errors.New("organization not found")
	ErrB2BPlanNotFound            = errors.New("b2b plan not found")
	ErrB2BSubscriptionNotFound    = errors.New("b2b subscription not found")
	ErrOrganizationMemberNotFound = errors.New("organization member not found")
)

type B2BRepository struct {
	db *gorm.DB
}

type MitraOrganizationListRow struct {
	OrganizationID         uint
	Code                   string
	Name                   string
	BusinessType           string
	ContactEmail           string
	OrganizationStatus     string
	RequiresMemberApproval bool
	MemberRole             string
	MemberStatus           string
}

func NewB2BRepository(db *gorm.DB) *B2BRepository {
	return &B2BRepository{db: db}
}

func (r *B2BRepository) RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *B2BRepository) FindUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *B2BRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *B2BRepository) CreateOrganization(ctx context.Context, organization *model.Organization) error {
	return r.db.WithContext(ctx).Create(organization).Error
}

func (r *B2BRepository) GetOrganizationByID(ctx context.Context, organizationID uint) (*model.Organization, error) {
	var organization model.Organization
	err := r.db.WithContext(ctx).Where("id = ?", organizationID).First(&organization).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}
	return &organization, nil
}

func (r *B2BRepository) GetOrganizationByCode(ctx context.Context, code string) (*model.Organization, error) {
	var organization model.Organization
	err := r.db.WithContext(ctx).Where("code = ?", strings.TrimSpace(code)).First(&organization).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}
	return &organization, nil
}

func (r *B2BRepository) ListOrganizations(ctx context.Context) ([]model.Organization, error) {
	var organizations []model.Organization
	err := r.db.WithContext(ctx).
		Model(&model.Organization{}).
		Order("created_at DESC").
		Find(&organizations).Error
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (r *B2BRepository) ListOrganizationsByUserID(ctx context.Context, userID uint) ([]MitraOrganizationListRow, error) {
	rows := make([]MitraOrganizationListRow, 0)
	err := r.db.WithContext(ctx).
		Table("organization_members AS om").
		Select(`
			o.id AS organization_id,
			o.code,
			o.name,
			o.business_type,
			o.contact_email,
			o.status AS organization_status,
			o.requires_member_approval,
			om.role AS member_role,
			om.status AS member_status
		`).
		Joins("JOIN organizations AS o ON o.id = om.organization_id").
		Where("om.user_id = ? AND om.status <> ?", userID, model.OrganizationMemberStatusRemoved).
		Order("o.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *B2BRepository) CreateOrganizationMember(ctx context.Context, member *model.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *B2BRepository) SaveOrganizationMember(ctx context.Context, member *model.OrganizationMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *B2BRepository) SaveOrganizationMemberTx(tx *gorm.DB, member *model.OrganizationMember) error {
	return tx.Save(member).Error
}

func (r *B2BRepository) GetOrganizationMemberByUserID(ctx context.Context, organizationID, userID uint) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) GetOrganizationMemberByID(ctx context.Context, organizationID, memberID uint) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", organizationID, memberID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) FindOrganizationMemberByEmail(ctx context.Context, organizationID uint, email string) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND LOWER(email) = LOWER(?)", organizationID, strings.TrimSpace(email)).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) FindOrganizationMemberByInvitationToken(ctx context.Context, organizationID uint, token string) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND invitation_token = ?", organizationID, strings.TrimSpace(token)).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) FindOrganizationMemberByInvitationTokenGlobal(ctx context.Context, token string) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("invitation_token = ?", strings.TrimSpace(token)).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) ListOrganizationMembers(ctx context.Context, organizationID uint, status string) ([]model.OrganizationMember, error) {
	query := r.db.WithContext(ctx).Model(&model.OrganizationMember{}).Where("organization_id = ?", organizationID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var members []model.OrganizationMember
	err := query.Order("created_at DESC").Find(&members).Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (r *B2BRepository) CreateB2BPlan(ctx context.Context, plan *model.B2BPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *B2BRepository) UpdateB2BPlan(ctx context.Context, plan *model.B2BPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *B2BRepository) GetB2BPlanByID(ctx context.Context, planID uint) (*model.B2BPlan, error) {
	var plan model.B2BPlan
	err := r.db.WithContext(ctx).Where("id = ?", planID).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrB2BPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *B2BRepository) ListB2BPlans(ctx context.Context, activeOnly bool) ([]model.B2BPlan, error) {
	query := r.db.WithContext(ctx).Model(&model.B2BPlan{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	var plans []model.B2BPlan
	err := query.Order("base_price_per_seat ASC").Find(&plans).Error
	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *B2BRepository) CreateB2BSubscription(ctx context.Context, subscription *model.B2BSubscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *B2BRepository) UpdateB2BSubscription(ctx context.Context, subscription *model.B2BSubscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *B2BRepository) SaveB2BSubscriptionTx(tx *gorm.DB, subscription *model.B2BSubscription) error {
	return tx.Save(subscription).Error
}

func (r *B2BRepository) GetActiveSubscriptionByOrganizationID(ctx context.Context, organizationID uint) (*model.B2BSubscription, error) {
	var subscription model.B2BSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("organization_id = ? AND status = ? AND starts_at <= ? AND ends_at > ?", organizationID, model.B2BSubscriptionStatusActive, time.Now(), time.Now()).
		Order("ends_at DESC").
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrB2BSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *B2BRepository) LockActiveSubscriptionByOrganizationID(tx *gorm.DB, organizationID uint) (*model.B2BSubscription, error) {
	var subscription model.B2BSubscription
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("organization_id = ? AND status = ? AND starts_at <= ? AND ends_at > ?", organizationID, model.B2BSubscriptionStatusActive, time.Now(), time.Now()).
		Order("ends_at DESC").
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrB2BSubscriptionNotFound
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *B2BRepository) CreateSeatAllocationTx(tx *gorm.DB, allocation *model.B2BSeatAllocation) error {
	return tx.Create(allocation).Error
}

func (r *B2BRepository) CountActiveSeatAllocationsTx(tx *gorm.DB, subscriptionID uint) (int64, error) {
	var count int64
	err := tx.Model(&model.B2BSeatAllocation{}).
		Where("subscription_id = ? AND released_at IS NULL", subscriptionID).
		Count(&count).Error
	return count, err
}

func (r *B2BRepository) CountActiveSeatAllocations(ctx context.Context, subscriptionID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.B2BSeatAllocation{}).
		Where("subscription_id = ? AND released_at IS NULL", subscriptionID).
		Count(&count).Error
	return count, err
}

func (r *B2BRepository) FindActiveSeatAllocationByMemberTx(tx *gorm.DB, subscriptionID uint, memberID uint) (*model.B2BSeatAllocation, error) {
	var allocation model.B2BSeatAllocation
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("subscription_id = ? AND organization_member_id = ? AND released_at IS NULL", subscriptionID, memberID).
		First(&allocation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &allocation, nil
}

func (r *B2BRepository) ReleaseSeatAllocationTx(tx *gorm.DB, allocationID uint, reason string) error {
	now := time.Now()
	return tx.Model(&model.B2BSeatAllocation{}).
		Where("id = ? AND released_at IS NULL", allocationID).
		Updates(map[string]any{
			"released_at":    now,
			"release_reason": reason,
		}).Error
}

func (r *B2BRepository) CreateB2BPricingQuote(ctx context.Context, quote *model.B2BPricingQuote) error {
	return r.db.WithContext(ctx).Create(quote).Error
}

func (r *B2BRepository) GetB2BPricingQuoteByCode(ctx context.Context, quoteCode string) (*model.B2BPricingQuote, error) {
	var quote model.B2BPricingQuote
	err := r.db.WithContext(ctx).Where("quote_code = ?", strings.TrimSpace(quoteCode)).First(&quote).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &quote, nil
}

func (r *B2BRepository) GetActiveB2BOrganizationForUser(ctx context.Context, userID uint, at time.Time) (*uint, error) {
	var organizationID uint
	err := r.db.WithContext(ctx).
		Table("organization_members AS om").
		Select("om.organization_id").
		Joins("JOIN organizations AS o ON o.id = om.organization_id AND o.status = ?", model.OrganizationStatusActive).
		Joins("JOIN b2b_subscriptions AS bs ON bs.organization_id = om.organization_id").
		Joins("JOIN b2b_billing_histories AS bh ON bh.subscription_id = bs.id AND bh.status = ? AND bh.paid_at IS NOT NULL AND bh.billing_period_start <= ? AND bh.billing_period_end > ?", model.B2BBillingHistoryStatusPaid, at, at).
		Joins("JOIN b2b_seat_allocations AS sa ON sa.subscription_id = bs.id AND sa.organization_member_id = om.id AND sa.released_at IS NULL").
		Where("om.user_id = ? AND om.status = ?", userID, model.OrganizationMemberStatusActive).
		Where("bs.status = ? AND bs.starts_at <= ? AND bs.ends_at > ?", model.B2BSubscriptionStatusActive, at, at).
		Order("bs.ends_at DESC").
		Limit(1).
		Scan(&organizationID).Error
	if err != nil {
		return nil, err
	}
	if organizationID == 0 {
		return nil, nil
	}
	return &organizationID, nil
}

func (r *B2BRepository) IsUserEntitledB2BPremium(ctx context.Context, userID uint, at time.Time) (bool, error) {
	organizationID, err := r.GetActiveB2BOrganizationForUser(ctx, userID, at)
	if err != nil {
		return false, err
	}
	return organizationID != nil, nil
}

func (r *B2BRepository) HasPaidBillingCoverageForSubscriptionTx(tx *gorm.DB, subscriptionID uint, at time.Time) (bool, error) {
	var count int64
	err := tx.Model(&model.B2BBillingHistory{}).
		Where("subscription_id = ?", subscriptionID).
		Where("status = ? AND paid_at IS NOT NULL", model.B2BBillingHistoryStatusPaid).
		Where("billing_period_start <= ? AND billing_period_end > ?", at, at).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
