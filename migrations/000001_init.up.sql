CREATE TABLE IF NOT EXISTS bank_stocks (
    name VARCHAR(255) PRIMARY KEY,
    quantity INT NOT NULL CHECK (quantity >= 0)
);

CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS wallet_stocks (
    wallet_id VARCHAR(255) REFERENCES wallets(id),
    stock_name VARCHAR(255) REFERENCES bank_stocks(name),
    quantity INT NOT NULL CHECK (quantity >= 0),
    PRIMARY KEY (wallet_id, stock_name)
);

CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    type VARCHAR(4) NOT NULL CHECK (type IN ('buy', 'sell')),
    wallet_id VARCHAR(255) NOT NULL,
    stock_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);