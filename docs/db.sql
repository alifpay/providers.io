CREATE TABLE partners (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    id SERIAL PRIMARY KEY,
    country CHAR(2),
    name TEXT NOT NULL,
    ref_id TEXT,
    info JSONB
);

CREATE TABLE providers (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active BOOLEAN DEFAULT true,
    id SERIAL PRIMARY KEY,
    partner_id INT, -- The partner owning this provider
    min_amount DECIMAL(18,2),
    max_amount DECIMAL(18,2),
    currency CHAR(3),
    name TEXT NOT NULL,
    gate TEXT
);

CREATE TABLE payments (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    canceled_at TIMESTAMP WITH TIME ZONE,
    pay_date DATE DEFAULT CURRENT_DATE,
    id UUID DEFAULT uuidv7(),
    provider_id INT NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    fee DECIMAL(18,2) DEFAULT 0.00,
    status SMALLINT DEFAULT 1,
    error_code SMALLINT,
    agent_id TEXT NOT NULL, -- The partner triggering the transaction
    ref_id TEXT NOT NULL,            -- External reference ID
    PRIMARY KEY(id, pay_date)
) WITH (
    tsdb.hypertable,
    tsdb.partition_column = 'pay_date',
    tsdb.segmentby        = 'provider_id',
    tsdb.orderby          = 'pay_date DESC'
);
CREATE UNIQUE INDEX idx_payment_service_ref_num ON payments(agent_id, ref_id, pay_date);

-- our balance in partner side
CREATE TABLE balance_log (
    log_date DATE DEFAULT CURRENT_DATE,
    id UUID DEFAULT uuidv7(),
    partner_id INT NOT NULL,
    provider_id INT,
    amount DECIMAL(18,2) NOT NULL,
    balance DECIMAL(18,2),    -- Negative for payments, Positive for topups/cancels
    ref_id TEXT NOT NULL,
    type TEXT NOT NULL, -- payment, topup, cancel, fix
    provider_name TEXT NOT NULL,
    partner_name TEXT NOT NULL,
    PRIMARY KEY(id, log_date)
) WITH (
    tsdb.hypertable,
    tsdb.partition_column = 'log_date',
    tsdb.segmentby        = 'partner_id',
    tsdb.orderby          = 'id DESC'
);
CREATE UNIQUE INDEX idx_balance_log_ref_id ON balance_log(ref_id, type, log_date);
CREATE INDEX idx_balance_log_lookup ON balance_log (partner_id, provider_id, log_date);

CREATE TABLE daily_stats (
    stat_date DATE DEFAULT CURRENT_DATE,
    provider_id INT NOT NULL,
    total_pay DECIMAL(18,2),  -- Total payments success
    count_pay INT,
    PRIMARY KEY(stat_date, provider_id)
) WITH (
    tsdb.hypertable,
    tsdb.partition_column = 'stat_date',
    tsdb.segmentby        = 'provider_id',
    tsdb.orderby          = 'stat_date DESC'
);