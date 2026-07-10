CREATE TABLE IF NOT EXISTS accounts (
    id               VARCHAR(255) PRIMARY KEY,
    client_id        VARCHAR(255) NOT NULL,
    balance          BIGINT       NOT NULL DEFAULT 0,
    reserved_balance BIGINT       NOT NULL DEFAULT 0,
    credit_limit     BIGINT       NOT NULL DEFAULT 0,
    status           VARCHAR(50)  NOT NULL DEFAULT 'active',
    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS transactions (
    id            VARCHAR(255) PRIMARY KEY,
    account_id    VARCHAR(255) NOT NULL,
    operation     VARCHAR(50)  NOT NULL,
    amount        BIGINT       NOT NULL,
    currency      VARCHAR(10)  NOT NULL,
    reference_id  VARCHAR(255) NOT NULL,
    status        VARCHAR(50)  NOT NULL DEFAULT 'pending',
    metadata      BYTEA,
    error_message TEXT,
    timestamp     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES accounts(id),
    CONSTRAINT uni_transactions_reference_id UNIQUE (reference_id)
    );