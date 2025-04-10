-- +goose Up
ALTER TABLE payment_transactions
DROP COLUMN IF EXISTS stripe_charge_id;

ALTER TABLE payment_transactions
    ADD COLUMN stripe_payment_intent_id VARCHAR(255) NOT NULL;

-- +goose Down
ALTER TABLE payment_transactions
DROP COLUMN IF EXISTS stripe_payment_intent_id;

ALTER TABLE payment_transactions
    ADD COLUMN stripe_charge_id VARCHAR(255) NOT NULL;