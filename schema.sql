-- Enable pgcrypto for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE "users" (
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
    admin_id UUID NOT NULL
);

CREATE TABLE expense (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description TEXT,
    amount DECIMAL(19, 4) NOT NULL,
    split JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'SETTLED', 'REOPENED')),
    settled_by UUID,
    created_by UUID NOT NULL,
    payee JSONB NOT NULL,
    group_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata'),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Kolkata')
);

-- Index for faster status-based queries
CREATE INDEX idx_expense_status ON expense(status);

-- Index for creator lookup
CREATE INDEX idx_expense_created_by ON expense(created_by);

-- Index for group-based expense queries
CREATE INDEX idx_expense_group ON expense(group_id);

CREATE TABLE expense_mapping (
    expense_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (expense_id, user_id)
);


-- Index for fetching all expenses of a users
CREATE INDEX idx_expense_mapping_user ON expense_mapping(user_id);


CREATE TABLE group_members (
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (user_id, group_id)
);


CREATE TABLE friends (
    user_id UUID NOT NULL,
    friend_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, friend_id)
);

-- Index for efficient searching
CREATE INDEX idx_user_friends ON friends(user_id);
CREATE INDEX idx_friend_users ON friends(friend_id);