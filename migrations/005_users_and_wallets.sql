CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL DEFAULT 'placeholder',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    asset VARCHAR(10) NOT NULL,
    balance NUMERIC(20, 8) DEFAULT 0 CHECK (balance >= 0),
    locked NUMERIC(20, 8) DEFAULT 0 CHECK (locked >= 0),
    UNIQUE(user_id, asset)
);

ALTER TABLE orders 
ADD COLUMN user_id UUID REFERENCES users(id);


WITH new_users AS (
    INSERT INTO users (username) 
    VALUES 
        ('sauron'), 
        ('frodo')   
    RETURNING id, username
)
INSERT INTO wallets (user_id, asset, balance, locked)

SELECT id, 'USD', 1000000.00, 0 FROM new_users WHERE username = 'sauron'
UNION ALL
SELECT id, 'BTC', 100.00, 0     FROM new_users WHERE username = 'sauron'

UNION ALL
SELECT id, 'USD', 1000.00, 0    FROM new_users WHERE username = 'frodo'
UNION ALL
SELECT id, 'BTC', 0.00, 0       FROM new_users WHERE username = 'frodo';
