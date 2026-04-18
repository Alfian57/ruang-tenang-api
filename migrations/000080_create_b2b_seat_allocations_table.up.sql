CREATE TABLE b2b_seat_allocations (
    id SERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES b2b_subscriptions(id) ON DELETE CASCADE,
    organization_member_id BIGINT NOT NULL REFERENCES organization_members(id) ON DELETE CASCADE,
    allocated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP,
    release_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_b2b_seat_allocations_subscription_member UNIQUE (subscription_id, organization_member_id)
);

CREATE INDEX idx_b2b_seat_allocations_subscription_active ON b2b_seat_allocations(subscription_id, released_at);
CREATE INDEX idx_b2b_seat_allocations_member ON b2b_seat_allocations(organization_member_id);
