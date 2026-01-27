CREATE TABLE IF NOT EXISTS trading_pairs (
    symbol TEXT PRIMARY KEY,    
    base_asset TEXT NOT NULL,     
    quote_asset TEXT NOT NULL,     
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO trading_pairs (symbol, base_asset, quote_asset) VALUES 
    ('BTC-USD', 'BTC', 'USD'),
    ('ETH-USD', 'ETH', 'USD')
ON CONFLICT DO NOTHING;
