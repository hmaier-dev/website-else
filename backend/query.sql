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
  name, date, doors, begin, place, city, coordinates, link, pic, hash
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetEventByID :one
SELECT name, date, doors, begin, place, city, coordinates, link, hash FROM events WHERE id = ?;
-- name: GetEventByHash :one
SELECT id, name, date, doors, begin, place, city, coordinates, link FROM events WHERE hash = ?;

-- name: GetAllEventsDesc :many
SELECT id, name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events ORDER BY id DESC;

-- name: CountEvents :one
SELECT count(*) FROM events;
-- name: LoadEventsDesc :many
SELECT name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events ORDER BY date DESC LIMIT ?;
-- name: LoadEventsAsc :many
SELECT name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events ORDER BY date ASC LIMIT ?;
-- name: LoadEventsRandom :many
SELECT name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events ORDER BY RANDOM() LIMIT ?;
-- name: LoadEventsNext :many
SELECT name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events WHERE DATE(date) > DATE('now', '-1 day') ORDER BY date ASC LIMIT ?;
-- name: LoadEventsPrev :many
SELECT name, date, doors, begin, place, city, coordinates, link, pic, hash FROM events WHERE DATE(date) < DATE('now') ORDER by date DESC LIMIT ?;


-- If there are duplicates get the id-list starting with the oldest entry, so we can delete it.
-- name: GetIDsFromEvent :many
SELECT id FROM events WHERE name = ? AND date = ? AND city = ? AND place = ? ORDER by id DESC;

-- name: DeleteEventByID :exec
DELETE FROM events
WHERE id = ?;

-- ##################################################
-- ## Add changes to a single value

-- name: UpdateNameByID :exec
UPDATE events SET name = ? WHERE id = ?;

-- name: UpdateDateByID :exec
UPDATE events SET date = ? WHERE id = ?;

-- name: UpdateDoorsByID :exec
UPDATE events SET doors = ? WHERE id = ?;
 
-- name: UpdateBeginByID :exec
UPDATE events SET begin = ? WHERE id = ?;
 
-- name: UpdatePlaceByID :exec
UPDATE events SET place = ? WHERE id = ?;

-- name: UpdateCityByID :exec
UPDATE events SET city = ? WHERE id = ?;
 
-- name: UpdateCoordinatesByID :exec
UPDATE events SET coordinates = ? WHERE id = ?;
 
-- name: UpdateLinkByID :exec
UPDATE events SET link = ? WHERE id = ?;
 
-- name: UpdatePicByID :exec
UPDATE events SET pic = ? WHERE id = ?;
 
-- name: UpdateHashByID :exec
UPDATE events SET hash = ? WHERE id = ?;

 

-- name: GetEventFull :one
SELECT id, name, date, begin, place, city, link FROM events WHERE id = ?;

-- name: ListEventsUpcoming :many
SELECT id, name, date, begin, place, city, link FROM events WHERE DATE(date) >= DATE('now') ORDER BY date ASC, id ASC;

-- name: ListEventsPast :many
SELECT id, name, date, begin, place, city, link FROM events WHERE DATE(date) < DATE('now') ORDER BY date DESC, id DESC;

-- name: UpdateEvent :exec
UPDATE events SET name = ?, date = ?, doors = ?, begin = ?, place = ?, city = ?, link = ? WHERE id = ?;

-- ################################################################
-- SECTION for mailbox (Verwaltung)
-- ################################################################

-- name: MarkMessageRead :exec
UPDATE mailbox SET IsRead = 1 WHERE id = ?;

-- ################################################################
-- SECTION for orders
-- ################################################################

-- name: CreateOrder :exec
INSERT INTO orders (
  paypal_order_id, status, items, item_total_cents, shipping_cents, total_cents, currency, created_at
) VALUES (
  ?, 'created', ?, ?, ?, ?, ?, ?
);

-- name: GetOrderByPayPalID :one
SELECT * FROM orders WHERE paypal_order_id = ?;

-- name: MarkOrderPaid :exec
UPDATE orders SET status = 'paid', payer_name = ?, payer_email = ?, shipping_address = ?, capture_id = ?, paid_at = ?
WHERE paypal_order_id = ? AND status = 'created';

-- name: MarkOrderShipped :exec
UPDATE orders SET status = 'shipped', shipped_at = ? WHERE id = ? AND status = 'paid';

-- name: ListOrdersByStatus :many
SELECT * FROM orders WHERE status = ? ORDER BY COALESCE(paid_at, created_at) DESC;
