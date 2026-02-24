To implement this in a **CQRS** pattern with Go and Postgres, we will split the schema into two parts: the **Write Side** (Normalised tables for data integrity) and the **Read Side** (Denormalized tables for high-performance queries).

### 1. The Write Side (Source of Truth)

These tables are highly normalized. We use a `balance_log` to ensure every change to a partner's balance is auditable and traceable to a specific event (Transaction, Top-up, or Cancellation).

```sql
-- Core Configuration
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    country CHAR(2),
    info JSONB,
    account_number VARCHAR(50) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    partner_id UUID REFERENCES partners(id), -- The partner owning this provider connection
    gate VARCHAR(100),
    currency CHAR(3),
    active BOOLEAN DEFAULT true,
    min_amount DECIMAL(18,2),
    max_amount DECIMAL(18,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transaction Ledger
CREATE TYPE transaction_status AS ENUM ('pending', 'success', 'failed', 'canceled');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id),
    agent_id UUID REFERENCES partners(id), -- The partner triggering the transaction
    ref_id VARCHAR(100) UNIQUE,            -- External reference ID
    amount DECIMAL(18,2) NOT NULL,
    fee DECIMAL(18,2) DEFAULT 0.00,
    status transaction_status DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    canceled_at TIMESTAMP WITH TIME ZONE
);

-- Audit Log for all balance movements
CREATE TYPE movement_type AS ENUM ('payment', 'topup', 'cancellation');

CREATE TABLE balance_log (
    id BIGSERIAL PRIMARY KEY,
    partner_id UUID REFERENCES partners(id),
    provider_id UUID REFERENCES providers(id),
    transaction_id UUID REFERENCES transactions(id), -- NULL if it's a manual top-up
    amount DECIMAL(18,2) NOT NULL,                   -- Negative for payments, Positive for topups/cancels
    type movement_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

```

---

### 2. The Read Side (Projections)

These tables are optimized for the queries you mentioned. They are updated via **database triggers** or **Go application logic** within the same transaction as the write.

```sql
-- 1. Query: "Remain of balance in partners by providers"
CREATE TABLE partner_provider_balances (
    partner_id UUID REFERENCES partners(id),
    provider_id UUID REFERENCES providers(id),
    current_balance DECIMAL(18,2) DEFAULT 0.00,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (partner_id, provider_id)
);

-- 2. Query: "Counts by day, total amount by day"
CREATE TABLE daily_stats_projection (
    day DATE,
    provider_id UUID REFERENCES providers(id),
    total_amount DECIMAL(18,2) DEFAULT 0.00,
    total_fee DECIMAL(18,2) DEFAULT 0.00,
    success_count INT DEFAULT 0,
    cancel_count INT DEFAULT 0,
    PRIMARY KEY (day, provider_id)
);

```

---

### 3. How the Go logic handles the 3 triggers

In your Go code, you should wrap the "Write" and the "Projection Update" in a single `sql.Tx`.

| Trigger | Action on `transactions` | Action on `partner_provider_balances` |
| --- | --- | --- |
| **New Transaction** | `INSERT` status 'success' | `UPDATE balance = balance - (amount + fee)` |
| **Top-up (Пополнение)** | N/A (Only `balance_log`) | `UPDATE balance = balance + amount` |
| **Cancellation** | `UPDATE` status 'canceled' | `UPDATE balance = balance + (amount + fee)` |

### Example SQL for "UPSERT" Daily Stats (Go Logic)

When a transaction succeeds, you run this alongside the insert:

```sql
INSERT INTO daily_stats_projection (day, provider_id, total_amount, success_count)
VALUES (CURRENT_DATE, $1, $2, 1)
ON CONFLICT (day, provider_id) DO UPDATE SET
    total_amount = daily_stats_projection.total_amount + EXCLUDED.total_amount,
    success_count = daily_stats_projection.success_count + 1;

```

---

### Why this works for your project:

* **Filtering:** To get all transactions with filters, you query the `transactions` table (Write side).
* **Performance:** To show a dashboard of balances, you query `partner_provider_balances`. Even with 10 million transactions, this query returns in **<1ms** because it's only one row per partner/provider.
* **Integrity:** The `balance_log` ensures that if a partner asks "Why is my balance $500?", you can sum the log and prove it.

**Would you like me to generate the Go structs and a `GORM` or `sqlx` function to handle a "Payment" transaction?**
