-- Enable pgcrypto for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    is_verified BOOLEAN NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata'),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata')
);

CREATE TABLE "group" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    admin_id UUID NOT NULL,
    FOREIGN KEY (admin_id) REFERENCES "user"(id)
);

CREATE TABLE expense (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description TEXT,
    amount DECIMAL(19, 4) NOT NULL,
    split JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'SETTLED', 'REOPENED')),
    settled_by UUID REFERENCES "user"(id),
    created_by UUID REFERENCES "user"(id),
    payee JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata'),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata')
);

-- Index for faster status-based queries
CREATE INDEX idx_expense_status ON expense(status);

-- Index for creator lookup
CREATE INDEX idx_expense_created_by ON expense(created_by);

CREATE TABLE expense_mapping (
    expense_id UUID NOT NULL REFERENCES expense(id),
    group_id UUID REFERENCES "group"(id),
    user_id UUID NOT NULL REFERENCES "user"(id),
    PRIMARY KEY (expense_id, user_id)
);

CREATE INDEX idx_expense_mapping_group ON expense_mapping(group_id);

-- Index for fetching all expenses of a user
CREATE INDEX idx_expense_mapping_user ON expense_mapping(user_id);

-- Composite Index for user+group queries 
CREATE INDEX idx_expense_mapping_user_group ON expense_mapping(user_id, group_id);

CREATE TABLE group_members (
    user_id UUID NOT NULL REFERENCES "user"(id),
    group_id UUID NOT NULL REFERENCES "group"(id),
    PRIMARY KEY (user_id, group_id)
);