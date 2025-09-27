-- name: AddMessage :exec
INSERT INTO mailbox (
  id, name, email, message, date, approval, IsRead
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
);

-- name: GetAllMessages :many
SELECT id, name, email, message, date, approval, IsRead FROM mailbox ORDER BY date ASC;

-- name: GetUnreadMessage :many
SELECT id, name, email, message, date, approval, IsRead FROM mailbox WHERE IsRead IS NULL ORDER BY date ASC;

