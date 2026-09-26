# Setup Duitku Pop untuk Ruang Tenang

Integrasi memakai Duitku Pop. Backend membuat invoice, web membuka popup Duitku dengan `reference` dan menyediakan `payment_url` sebagai fallback, sedangkan mobile membuka `payment_url` di browser dalam aplikasi. Status pembayaran dan pemberian benefit hanya ditetapkan dari callback server yang diverifikasi.

## Kredensial dan endpoint

Buat project/merchant di Merchant Portal Duitku, lalu salin merchant code dan API key. Sandbox dan production memiliki kredensial serta endpoint yang berbeda.

| Environment | `DUITKU_BASE_URL` | `NEXT_PUBLIC_DUITKU_ENV` |
| --- | --- | --- |
| Sandbox | `https://api-sandbox.duitku.com` | `sandbox` |
| Production | `https://api-prod.duitku.com` | `production` |

Isi konfigurasi backend berikut:

```env
API_PUBLIC_URL=https://<domain-api>
FRONTEND_URL=https://<domain-web>
DUITKU_BASE_URL=https://api-sandbox.duitku.com
DUITKU_MERCHANT_CODE=<merchant-code>
DUITKU_API_KEY=<api-key>
```

Di web, atur `NEXT_PUBLIC_DUITKU_ENV` ke `sandbox` atau `production` sesuai environment backend. Simpan API key hanya di backend. Konfigurasi kedua aplikasi harus memakai environment Duitku yang sama.

## Callback dan redirect

Daftarkan URL callback berikut pada pengaturan project Duitku untuk masing-masing environment:

```text
https://<domain-api>/api/v1/billing/webhooks/duitku
```

Callback harus dapat menerima `POST application/x-www-form-urlencoded` dari Duitku. `API_PUBLIC_URL` harus merupakan origin API publik yang dapat dijangkau provider. Backend mengirim `FRONTEND_URL/payment/success` sebagai `returnUrl`; halaman ini hanya memberi informasi bahwa status sedang diverifikasi. Redirect browser dan callback popup tidak menjadi bukti pembayaran.

Callback HMAC diverifikasi menggunakan API key dengan formula `HMAC_SHA256(merchantCode + amount + merchantOrderId, apiKey)`. Handler kemudian mencocokkan merchant code, order ID, dan nominal pada transaksi lokal sebelum memproses `resultCode=00` sebagai berhasil. Callback yang sama aman dikirim ulang.

## Perilaku checkout

Web memakai Duitku JavaScript Pop: Sandbox memuat `https://app-sandbox.duitku.com/lib/js/duitku.js`, sedangkan production memakai `https://app-prod.duitku.com/lib/js/duitku.js`. Checkout dimulai dengan `checkout.process(provider_reference, options)`. Jika script gagal dimuat atau popup tidak tersedia, web membuka `payment_url`.

Mobile membuka `payment_url` dengan browser dalam aplikasi. Pengguna kembali ke aplikasi setelah menyelesaikan atau membatalkan pembayaran; aplikasi memperbarui status dari API billing.

## Pemeriksaan operasional

Gunakan dashboard Duitku untuk memeriksa status invoice dan mengirim ulang callback jika callback awal gagal. Setelah itu, periksa `GET /api/v1/billing/status` atau `GET /api/v1/billing/transactions`. Jangan menandai transaksi sebagai paid berdasarkan halaman return, popup JavaScript, atau konfirmasi dari client.

Referensi resmi: [Duitku Pop API dan callback](https://docs.duitku.com/pop/id/), [integrasi JavaScript Pop](https://docs.duitku.com/pop/id/#integrasi-frontend-atau-view).
