CREATE TABLE IF NOT EXISTS trades (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  bid_order_id UUID NOT NULL REFERENCES orders(id),
  ask_order_id UUID NOT NULL REFERENCES orders(id),
  
  quantity INT NOT NULL,
  price DECIMAL(12, 2) NOT NULL,
  
  timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);
