// Warenkorb und PayPal-Checkout für /shop.
//
// Der Warenkorb liegt nur im Browser (localStorage). Beim Bezahlen werden nur
// Produkt-IDs und Mengen ans Backend geschickt – die Preise rechnet das Backend
// selbst aus dem Katalog nach.

const STORAGE_KEY = 'else-cart'
const MAX_QTY = 10

const data = JSON.parse(document.getElementById('shop-data').textContent)
const products = Object.fromEntries(data.products.map(p => [p.id, p]))
const money = new Intl.NumberFormat('de-DE', { style: 'currency', currency: data.currency })
const euro = cents => money.format(cents / 100)
const $ = id => document.getElementById(id)

// --- Warenkorb-Speicher ------------------------------------------------------

function loadCart() {
	let raw = {}
	try { raw = JSON.parse(localStorage.getItem(STORAGE_KEY)) || {} } catch (_) { /* privater Modus o. ä. */ }
	const cart = {}
	for (const [id, qty] of Object.entries(raw)) {
		// Artikel, die es nicht mehr gibt oder die ausverkauft sind, still entfernen
		if (products[id] && products[id].available && Number.isInteger(qty) && qty > 0) {
			cart[id] = Math.min(qty, MAX_QTY)
		}
	}
	return cart
}

let cart = loadCart()

function saveCart() {
	try { localStorage.setItem(STORAGE_KEY, JSON.stringify(cart)) } catch (_) { /* egal */ }
	render()
}

function setQty(id, qty) {
	if (qty <= 0) delete cart[id]
	else cart[id] = Math.min(qty, MAX_QTY)
	saveCart()
}

function totals() {
	let count = 0
	let subtotal = 0
	for (const [id, qty] of Object.entries(cart)) {
		count += qty
		subtotal += products[id].price_cents * qty
	}
	const free = data.free_shipping_from_cents > 0 && subtotal >= data.free_shipping_from_cents
	const shipping = count === 0 || free ? 0 : data.shipping_cents
	return { count, subtotal, shipping, total: subtotal + shipping }
}

// --- Anzeige -----------------------------------------------------------------

function esc(s) {
	return String(s).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]))
}

function render() {
	const t = totals()
	$('cart-count').textContent = t.count
	$('cart-empty').hidden = t.count > 0
	$('cart-filled').hidden = t.count === 0
	$('cart-fab').hidden = t.count === 0
	$('cart-fab-total').textContent = `(${t.count}) · ${euro(t.total)}`

	$('cart-lines').innerHTML = Object.entries(cart).map(([id, qty]) => {
		const p = products[id]
		return `
			<li class="flex items-center gap-3 py-3">
				<div class="min-w-0 flex-1">
					<p class="font-semibold leading-tight">${esc(p.name)}</p>
					<p class="text-sm text-ink/60">${euro(p.price_cents)}</p>
				</div>
				<div class="flex items-center rounded-full border-2 border-ink">
					<button type="button" data-dec="${esc(id)}" class="h-8 w-8 rounded-full font-bold hover:bg-lilac-200" aria-label="Eins weniger">−</button>
					<span class="w-6 text-center font-bold" aria-label="Menge">${qty}</span>
					<button type="button" data-inc="${esc(id)}" class="h-8 w-8 rounded-full font-bold hover:bg-lilac-200 disabled:opacity-30" aria-label="Eins mehr" ${qty >= MAX_QTY ? 'disabled' : ''}>+</button>
				</div>
			</li>`
	}).join('')

	$('cart-subtotal').textContent = euro(t.subtotal)
	$('cart-shipping').textContent = t.shipping === 0 ? 'kostenlos' : euro(t.shipping)
	$('cart-total').textContent = euro(t.total)

	const hint = $('cart-free-hint')
	const missing = data.free_shipping_from_cents - t.subtotal
	hint.hidden = !(data.free_shipping_from_cents > 0 && missing > 0)
	hint.textContent = `Noch ${euro(missing)} bis zum kostenlosen Versand.`
}

function showMessage(text) {
	const el = $('checkout-message')
	el.textContent = text
	el.hidden = !text
}

document.addEventListener('click', e => {
	const add = e.target.closest('[data-add]')
	if (add) {
		const id = add.dataset.add
		setQty(id, (cart[id] || 0) + 1)
		add.textContent = 'Hinzugefügt ✓'
		setTimeout(() => { add.textContent = 'In den Warenkorb' }, 1200)
		return
	}
	const inc = e.target.closest('[data-inc]')
	if (inc) return setQty(inc.dataset.inc, (cart[inc.dataset.inc] || 0) + 1)
	const dec = e.target.closest('[data-dec]')
	if (dec) return setQty(dec.dataset.dec, (cart[dec.dataset.dec] || 0) - 1)
})

// --- PayPal ------------------------------------------------------------------

let paypalLoaded = null

async function errorText(res, fallback) {
	try { return (await res.json()).error || fallback } catch (_) { return fallback }
}

function loadPayPal() {
	if (paypalLoaded) return paypalLoaded
	paypalLoaded = (async () => {
		const res = await fetch('/shop/api/config')
		const cfg = await res.json()
		if (!cfg.enabled) throw new Error('Bezahlen ist gerade leider nicht möglich. Schreib mir gern über das Kontaktformular.')

		await new Promise((resolve, reject) => {
			const s = document.createElement('script')
			const params = new URLSearchParams({ 'client-id': cfg.clientId, currency: cfg.currency, intent: 'capture', locale: 'de_DE', components: 'buttons' })
			s.src = `https://www.paypal.com/sdk/js?${params}`
			s.onload = resolve
			s.onerror = () => reject(new Error('PayPal konnte nicht geladen werden. Bitte prüfe deinen Werbeblocker.'))
			document.head.appendChild(s)
		})

		await window.paypal.Buttons({
			style: { shape: 'pill', color: 'black', label: 'paypal', height: 48 },

			createOrder: async () => {
				showMessage('')
				const items = Object.entries(cart).map(([id, quantity]) => ({ id, quantity }))
				const r = await fetch('/shop/api/orders', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ items }),
				})
				if (!r.ok) {
					const msg = await errorText(r, 'Die Bestellung konnte nicht angelegt werden.')
					showMessage(msg)
					throw new Error(msg)
				}
				return (await r.json()).id
			},

			onApprove: async ({ orderID }) => {
				const r = await fetch(`/shop/api/orders/${encodeURIComponent(orderID)}/capture`, { method: 'POST' })
				if (!r.ok) {
					showMessage(await errorText(r, 'Die Zahlung konnte nicht abgeschlossen werden.'))
					return
				}
				const result = await r.json()
				cart = {}
				saveCart()
				$('shop').hidden = true
				$('success-number').textContent = result.orderNumber
				$('shop-success').hidden = false
				$('shop-success').scrollIntoView({ behavior: 'smooth', block: 'center' })
			},

			onCancel: () => showMessage('Bezahlung abgebrochen. Dein Warenkorb ist noch da.'),
			onError: err => {
				console.error(err)
				if (!$('checkout-message').textContent) showMessage('Bei PayPal ist etwas schiefgelaufen. Bitte versuch es nochmal.')
			},
		}).render('#paypal-buttons')
	})()
	return paypalLoaded
}

$('cart-terms').addEventListener('change', async e => {
	const box = $('paypal-buttons')
	if (!e.target.checked) {
		box.hidden = true
		return
	}
	showMessage('')
	box.hidden = false
	try {
		await loadPayPal()
	} catch (err) {
		paypalLoaded = null
		box.hidden = true
		e.target.checked = false
		showMessage(err.message)
	}
})

render()
