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
-- name: CountEvents :one
SELECT count(*) FROM events;
-- name: LoadEventsDesc :many
SELECT name, date, doors, begin, place, city, link, pic FROM events ORDER BY date DESC LIMIT ?;
-- name: LoadEventsAsc :many
SELECT name, date, doors, begin, place, city, link, pic FROM events ORDER BY date ASC LIMIT ?;
-- name: LoadEventsRandom :many
SELECT name, date, doors, begin, place, city, link, pic FROM events ORDER BY RANDOM() LIMIT ?;
-- name: LoadEventsNext :many
SELECT name, date, doors, begin, place, city, link, pic FROM events WHERE DATE(date) > DATE('now') ORDER BY date ASC LIMIT ?;
-- name: LoadEventsPrev :many
SELECT name, date, doors, begin, place, city, link, pic FROM events WHERE DATE(date) < DATE('now') ORDER by date DESC LIMIT ?;

-- If there are duplicates get the id-list starting with the oldest entry, so we can delete it.
-- name: GetIDsFromEvent :many
SELECT id FROM events WHERE name = ? AND date = ? AND city = ? AND place = ? ORDER by id DESC;

-- name: GetEventByID :one
SELECT name, date, doors, begin, place, city, link FROM events WHERE id = ?;

-- name: DeleteEventByID :exec
DELETE FROM events
WHERE id = ?;
