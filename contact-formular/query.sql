-- name: AddMessage :exec
INSERT INTO mailbox (
  id, name, email, date, approval, IsRead
) VALUES (
  ?, ?, ?, ?, ?, ?
);

-- name: GetAllMessages :many
SELECT id, name, email, date, approval, IsRead FROM mailbox ORDER BY date ASC;

-- name: GetUnreadMessage :many
SELECT id, name, email, date, approval, IsRead FROM mailbox WHERE IsRead IS NULL ORDER BY date ASC;
