-- ################################################################
-- SECTION for mailbox
-- ################################################################

-- name: AddMessageToMailbox :exec
INSERT INTO mailbox (
  name, email, message, date, approval, IsRead
) VALUES (
  ?, ?, ?, ?, ?, ?
);

-- name: GetCompleteMailbox :many
SELECT id, name, email, message, date, approval, IsRead FROM mailbox ORDER BY date ASC;

-- name: GetUnreadMailboxItems :many
SELECT id, name, email, message, date, approval, IsRead FROM mailbox WHERE IsRead IS NULL ORDER BY date ASC;


-- ################################################################
-- SECTION for events
-- ################################################################

-- name: AddSingleEvent :exec
INSERT INTO events (
  name, date, doors, begin, place, city, link, pic
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: LoadEventsDesc :many
SELECT * FROM events ORDER BY date DESC;
-- name: LoadEventsAsc :many
SELECT * FROM events ORDER BY date ASC;
-- name: LoadEventsRandom :many
SELECT * FROM events ORDER BY RANDOM();
