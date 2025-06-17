-- name: FetchUserByEmail :one
SELECT * FROM "users" WHERE email = $1;

-- name: InsertUser :one
INSERT INTO "users" (id, name, email, is_verified, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUser :one
UPDATE "users"
SET name = $2, email = $3, is_verified = $4, password = $5, updated_at = NOW() AT TIME ZONE 'Asia/Kolkata'
WHERE id = $1
RETURNING *;

-- name: FetchUserById :one
SELECT * FROM "users" WHERE id = $1 LIMIT 1;

-- name: FetchGroupsByUser :many
SELECT g.* FROM "group" g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1;

-- name: FetchUsersInGroup :many
SELECT u.* FROM "users" u
JOIN group_members gm ON u.id = gm.user_id
WHERE gm.group_id = $1;

-- name: FetchGroupById :one
SELECT * FROM "group" WHERE id = $1 LIMIT 1;

-- name: CreateOrUpdateGroup :one
INSERT INTO "group" (id, name, description, admin_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    admin_id = EXCLUDED.admin_id
RETURNING *;

-- name: AddUserInGroup :one
INSERT INTO group_members (user_id, group_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING TRUE;

-- name: RemoveUserFromGroup :one
DELETE FROM group_members
WHERE user_id = $1 AND group_id = $2
RETURNING TRUE;

-- name: CreateOrUpdateExpense :one
INSERT INTO expense (id, description, amount, split, status, settled_by, created_by, payee, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    description = EXCLUDED.description,
    amount = EXCLUDED.amount,
    split = EXCLUDED.split,
    status = EXCLUDED.status,
    settled_by = EXCLUDED.settled_by,
    created_by = EXCLUDED.created_by,
    payee = EXCLUDED.payee,
    updated_at = NOW() AT TIME ZONE 'Asia/Kolkata'
RETURNING *;

-- name: FetchExpense :one
SELECT * FROM expense WHERE id = $1 LIMIT 1;

-- name: FetchExpenseAssociatedGroup :one
SELECT group_id FROM expense_mapping WHERE expense_id = $1 AND group_id != NULL LIMIT 1;

-- name: DeleteExpense :one
DELETE FROM expense WHERE id = $1 RETURNING TRUE;

-- name: FetchGroupExpenses :many
SELECT e.*
FROM expense e
WHERE e.id IN (
  SELECT DISTINCT em.expense_id
  FROM expense_mapping em
  WHERE em.group_id = $1
  LIMIT $3 OFFSET (($2 - 1) * $3)
)
ORDER BY (e.status='DRAFT') DESC, e.created_at ASC;

-- name: CheckUserExistsInGroup :one
SELECT EXISTS(
    SELECT 1 FROM group_members
    WHERE user_id = $1 AND group_id = $2
);

-- name: AttachExpenseToGroup :one
INSERT INTO expense_mapping (expense_id, group_id, user_id)
SELECT $1, $2, unnest($3::uuid[])
ON CONFLICT DO NOTHING
RETURNING TRUE;

-- name: AddUserExpenseMapping :one
INSERT INTO expense_mapping (expense_id, user_id)
SELECT $1, $2 
ON CONFLICT DO NOTHING
RETURNING TRUE;

-- name: RemoveUserFromExpense :one
DELETE FROM expense_mapping
WHERE expense_id = $1 AND user_id = ANY($2::uuid[])
RETURNING TRUE;

-- name: AddFriend :one 
INSERT INTO friends (user_id, friend_id) 
VALUES ($1, $2)
ON CONFLICT (user_id, friend_id) DO NOTHING
RETURNING *; 

-- name: RemoveFriend :one 
DELETE FROM friends 
WHERE (user_id = $1 AND friend_id = $2) 
   OR (user_id = $1 AND friend_id = $2)
   RETURNING TRUE;  -- Remove friendship in both directions

-- name: GetFriends :many
SELECT u.id, u.name, u.email, u.is_verified, u.created_at, u.updated_at
FROM users u
JOIN friends f ON u.id = f.friend_id
WHERE f.user_id = $1;

-- name: GetFriend :one
SELECT u.id, u.name, u.email
FROM users u
JOIN friends f ON u.id = f.friend_id
WHERE (f.user_id = $1 AND f.friend_id = $2) OR (f.user_id = $2 AND f.friend_id = $1);

-- name: FetchExpenseCountByGroup :one
SELECT COUNT(*) AS count
FROM expense e
JOIN expense_mapping em ON e.id = em.expense_id
WHERE em.group_id = $1;

-- name: FetchExpenseByUserAndStatus :many
SELECT e.* from expense_mapping em
JOIN expense e ON em.expense_id = e.id
where em.user_id = $1 AND e.status = $2
ORDER BY e.created_at DESC
LIMIT $3 OFFSET (($4 - 1) * $3);