-- +goose Up

-- Enable pgcrypto for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create the users table
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       stripe_customer_id VARCHAR(255),
                       created_at TIMESTAMP NOT NULL DEFAULT now(),
                       updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Create the subscriptions table
CREATE TABLE subscriptions (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id UUID NOT NULL,
                               stripe_subscription_id VARCHAR(255) NOT NULL,
                               status VARCHAR(50) NOT NULL, -- e.g., 'trialing', 'active', 'canceled'
                               trial_start TIMESTAMP,
                               trial_end TIMESTAMP,
                               started_at TIMESTAMP,
                               next_billing_date TIMESTAMP,
                               created_at TIMESTAMP NOT NULL DEFAULT now(),
                               updated_at TIMESTAMP NOT NULL DEFAULT now(),
                               CONSTRAINT fk_user_subscription
                                   FOREIGN KEY(user_id)
                                       REFERENCES users(id)
                                       ON DELETE CASCADE
);

-- Create the payment_transactions table
CREATE TABLE payment_transactions (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      user_id UUID NOT NULL,
                                      stripe_charge_id VARCHAR(255) NOT NULL,
                                      amount INTEGER NOT NULL,  -- Amount in pence (or your smallest currency unit)
                                      currency VARCHAR(10) NOT NULL, -- e.g., 'GBP'
                                      status VARCHAR(50) NOT NULL,  -- e.g., 'succeeded', 'failed'
                                      created_at TIMESTAMP NOT NULL DEFAULT now(),
                                      CONSTRAINT fk_transaction_user
                                          FOREIGN KEY(user_id)
                                              REFERENCES users(id)
                                              ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pgcrypto";