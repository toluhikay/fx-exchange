-- Initial database schema for Operation Borderless
-- Creates tables for wallets, transactions, fx_rates, and audit_logs
-- Uses UUID data type for IDs with uuid-ossp extension
-- Uses NUMERIC for monetary amounts and rates for fintech-grade precision
-- Includes cascading behavior for foreign keys to ensure data integrity

-- Enabling uuid-ossp extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Creating users table to store users information
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()),    
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL
);

-- Creating wallets table to store user wallet information
-- Parent table for transactions and audit_logs
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    balances JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Creating transactions table to store wallet operation history
-- Foreign key to wallets with ON DELETE CASCADE to remove transactions when wallet is deleted
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    from_currency VARCHAR(10),
    to_currency VARCHAR(10),
    amount NUMERIC(19,4) NOT NULL,
    converted_amount NUMERIC(19,4),
    rate NUMERIC(19,4),
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

-- Creating fx_rates table to store historical FX rates
-- No foreign keys, independent of other tables
CREATE TABLE fx_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency VARCHAR(10) NOT NULL,
    to_currency VARCHAR(10) NOT NULL,
    rate NUMERIC(19,4) NOT NULL,
    timestamp TIMESTAMP NOT NULL
);

-- Creating audit_logs table for operation auditing
-- Foreign key to wallets with ON DELETE SET NULL to preserve audit records
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID,
    operation VARCHAR(50) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    request_method VARCHAR(10),
    request_path TEXT,
    request_body TEXT,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL
);

-- Creating indexes for performance
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_fx_rates_timestamp ON fx_rates(timestamp);
CREATE INDEX idx_audit_logs_wallet_id ON audit_logs(wallet_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);