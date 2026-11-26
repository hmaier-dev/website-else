// Warenkorb und PayPal-Checkout für /shop.
//
// Der Warenkorb liegt nur im Browser (localStorage). Beim Bezahlen werden nur
// Produkt-IDs, Mengen und – bei frei wählbaren Preisen – der gewählte Preis
// ans Backend geschickt. Das Backend rechnet alles mit dem Katalog nach und
// akzeptiert Wunschpreise nur innerhalb der Spanne aus content/shop.md.

const STORAGE_KEY = 'else-cart'
const MAX_QTY = 10

const data = JSON.parse(document.getElementById('shop-data').textContent)
const products = Object.fromEntries(data.products.map(p => [p.id, p]))
const money = new Intl.NumberFormat('de-DE', { style: 'currency', currency: data.currency })
const euro = cents => money.format(cents / 100)
const $ = id => document.getElementById(id)
const flexible = p => p.min_cents > 0 && p.max_cents > 0
const inRange = (p, cents) => Number.isInteger(cents) && cents >= p.min_cents && cents <= p.max_cents

// --- Warenkorb-Speicher ------------------------------------------------------
// Format: { "produkt-id": { q: Menge, p: Preis in Cent } }

function loadCart() {
	let raw = {}
	try { raw = JSON.parse(localStorage.getItem(STORAGE_KEY)) || {} } catch (_) { /* privater Modus o. ä. */ }
	const cart = {}
	for (const [id, entry] of Object.entries(raw)) {
		const p = products[id]
		// alte Speicherform { id: Menge } weiter lesen
		const q = typeof entry === 'number' ? entry : entry && entry.q
		// Artikel, die es nicht mehr gibt oder die ausverkauft sind, still entfernen
		if (!p || !p.available || !Number.isInteger(q) || q < 1) continue
		let price = p.price_cents
		if (flexible(p) && entry && inRange(p, entry.p)) price = entry.p
		cart[id] = { q: Math.min(q, MAX_QTY), p: price }
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
	else cart[id] = { q: Math.min(qty, MAX_QTY), p: cart[id] ? cart[id].p : products[id].price_cents }
	saveCart()
}

function totals() {
	let count = 0
	let subtotal = 0
	for (const { q, p } of Object.values(cart)) {
		count += q
		subtotal += p * q
	}
	const free = data.free_shipping_from_cents > 0 && subtotal >= data.free_shipping_from_cents
	const shipping = count === 0 || free ? 0 : data.shipping_cents
	return { count, subtotal, shipping, total: subtotal + shipping }
}

// --- Preis selbst wählen -----------------------------------------------------
// Ein Preis-Wähler besteht aus −/+ Knöpfen, einem Textfeld und einem Schieberegler.
// Alles rechnet in Cent (keine Kommazahl-Rundungsfehler); angezeigt wird "12,50".

const pickers = {}

const fmt = cents => (cents / 100).toFixed(2).replace('.', ',')
const parseEuro = text => {
	const t = String(text).trim().replace(/\s|€/g, '').replace(',', '.')
	if (!/^\d+(\.\d{0,2})?$/.test(t)) return null
	return Math.round(parseFloat(t) * 100)
}

function setupPicker(el) {
	const id = el.dataset.pricePicker
	const min = +el.dataset.min, max = +el.dataset.max, step = +el.dataset.step || 100
	const input = el.querySelector('[data-price-input]')
	const slider = el.querySelector('[data-price-slider]')
	const minus = el.querySelector('[data-price-step="-1"]')
	const plus = el.querySelector('[data-price-step="1"]')
	const picker = { value: +el.dataset.value }
	pickers[id] = picker

	const clamp = v => Math.min(max, Math.max(min, v))

	// Wert setzen und alle Bedienelemente nachziehen.
	// keepText: beim Tippen nicht in das Feld hineinformatieren
	function set(v, { keepText = false } = {}) {
		picker.value = clamp(v)
		if (!keepText) input.value = fmt(picker.value)
		slider.value = picker.value
		slider.style.setProperty('--fill', `${((picker.value - min) / (max - min || 1)) * 100}%`)
		minus.disabled = picker.value <= min
		plus.disabled = picker.value >= max
		input.setAttribute('aria-valuenow', picker.value)
	}

	// Auf das Raster der Schrittweite springen (12,30 → + → 13,00)
	const stepBy = dir => set(dir > 0
		? Math.floor(picker.value / step) * step + step
		: Math.ceil(picker.value / step) * step - step)

	// Gedrückt halten: ein Schritt, kurze Pause, dann gleichmäßig weiter und nur sanft schneller
	let timer = null
	const stop = () => { clearTimeout(timer); timer = null }
	const hold = (dir, n = 0) => {
		stepBy(dir)
		if ((dir > 0 && picker.value >= max) || (dir < 0 && picker.value <= min)) return stop()
		const delay = n === 0 ? 450 : Math.max(70, 170 * Math.pow(0.93, n))
		timer = setTimeout(() => hold(dir, n + 1), delay)
	}
	for (const [btn, dir] of [[minus, -1], [plus, 1]]) {
		btn.addEventListener('pointerdown', e => {
			if (e.button !== 0) return
			e.preventDefault() // kein Fokus-Sprung, kein Text markieren
			btn.setPointerCapture?.(e.pointerId)
			hold(dir)
		})
		for (const ev of ['pointerup', 'pointercancel', 'lostpointercapture']) btn.addEventListener(ev, stop)
		// Tastatur (Enter/Leertaste) löst nur "click" ohne Zeiger aus
		btn.addEventListener('click', e => { if (e.detail === 0) stepBy(dir) })
		btn.addEventListener('contextmenu', e => e.preventDefault()) // langes Drücken auf dem Handy
	}

	// Der Regler hat Cent-Auflösung (zeigt auch getippte Werte genau an),
	// beim Ziehen wird auf die Schrittweite gerundet.
	slider.addEventListener('input', () => set(min + Math.round((+slider.value - min) / step) * step))

	input.addEventListener('input', () => {
		const v = parseEuro(input.value)
		const ok = v !== null && v >= min && v <= max
		input.setAttribute('aria-invalid', String(!ok))
		input.classList.toggle('!border-lilac-600', !ok)
		if (ok) set(v, { keepText: true })
	})
	// Beim Verlassen: auf die Spanne begrenzen und schön formatieren
	input.addEventListener('blur', () => {
		const v = parseEuro(input.value)
		set(v === null ? picker.value : v)
		input.removeAttribute('aria-invalid')
		input.classList.remove('!border-lilac-600')
	})
	input.addEventListener('keydown', e => {
		if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
			e.preventDefault()
			stepBy(e.key === 'ArrowUp' ? 1 : -1)
		} else if (e.key === 'Enter') {
			input.blur()
		}
	})

	set(picker.value)
}

document.querySelectorAll('[data-price-picker]').forEach(setupPicker)

// gewählter Preis in Cent (feste Preise: Katalogpreis)
function chosenPrice(id) {
	const picker = pickers[id]
	if (!picker) return products[id].price_cents
	return inRange(products[id], picker.value) ? picker.value : null
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

	$('cart-lines').innerHTML = Object.entries(cart).map(([id, { q, p: price }]) => {
		const p = products[id]
		return `
			<li class="flex items-center gap-3 py-3">
				<div class="min-w-0 flex-1">
					<p class="font-semibold leading-tight">${esc(p.name)}</p>
					<p class="text-sm text-ink/60">${euro(price)}${flexible(p) ? ' · dein Preis' : ''}</p>
				</div>
				<div class="flex items-center rounded-full border-2 border-ink">
					<button type="button" data-dec="${esc(id)}" class="h-8 w-8 rounded-full font-bold hover:bg-lilac-200" aria-label="Eins weniger">−</button>
					<span class="w-6 text-center font-bold" aria-label="Menge">${q}</span>
					<button type="button" data-inc="${esc(id)}" class="h-8 w-8 rounded-full font-bold hover:bg-lilac-200 disabled:opacity-30" aria-label="Eins mehr" ${q >= MAX_QTY ? 'disabled' : ''}>+</button>
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
		// Ein gerade getippter Preis zählt, auch wenn das Feld noch den Fokus hat
		document.activeElement?.matches?.('[data-price-input]') && document.activeElement.blur()
		const price = chosenPrice(id)
		if (price === null) return
		// Bei frei wählbarem Preis gilt der zuletzt gewählte Preis für alle Exemplare
		cart[id] = { q: Math.min((cart[id] ? cart[id].q : 0) + 1, MAX_QTY), p: price }
		saveCart()
		add.textContent = 'Hinzugefügt ✓'
		setTimeout(() => { add.textContent = 'In den Warenkorb' }, 1200)
		return
	}
	const qty = id => (cart[id] ? cart[id].q : 0)
	const inc = e.target.closest('[data-inc]')
	if (inc) return setQty(inc.dataset.inc, qty(inc.dataset.inc) + 1)
	const dec = e.target.closest('[data-dec]')
	if (dec) return setQty(dec.dataset.dec, qty(dec.dataset.dec) - 1)
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
				const items = Object.entries(cart).map(([id, { q, p }]) => ({ id, quantity: q, price_cents: p }))
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
