# API Contract
## Base dan auth

Base path API adalah /api/v1. Route publik dan route terlindungi dibedakan di internal/router/router_api_v1_routes.go. JWT dikirim sebagai Authorization: Bearer <token>. Middleware role/ownership menentukan akses; dokumentasi frontend bukan security boundary.

## Response

Pertahankan response envelope dan error helper yang digunakan source code. Response sukses biasanya membawa data, optional pagination/meta, dan request id; error harus mempertahankan status, message/code, serta detail validation yang dipakai client.

Pagination memakai query page, limit, dan utility terkait. Jika mengubah field pagination atau error, cek normalizer web/services/http/client.ts dan parser di mobile.

Dashboard web memakai pagination database untuk `/my-articles?search=`, `/song-categories?page=&limit=`, `/playlists?page=&limit=`, `/playlists/public?kind=official|community`, `/search?type=songs&page=&limit=`, dan `/rewards?page=&limit=&reward_type=`. `/forums?circle=` memfilter lingkar dukungan di database sebelum limit/offset. Katalog hadiah paginated juga menyertakan `reward_types` sebagai facet seluruh hadiah aktif; playlist publik menyertakan `is_admin_playlist`. `/journals/public` menambahkan `total_items`/`total_pages` tanpa menghapus `total`. `/song-categories`, `/playlists`, `/rewards`, serta `/search` tanpa opt-in query tetap memakai bentuk legacy yang dikonsumsi mobile. Web menormalisasi response flat paginated ke `meta`; mobile dapat tetap memakai panggilan lama tanpa perubahan parser.

## OpenAPI workflow

Anotasi Summary, Tags, Param, Success, Failure, dan Router pada handler adalah sumber dokumentasi. Jalankan make swagger setelah perubahan lalu pastikan make swagger-check lulus. Review docs/openapi.yaml sebagai contract snapshot; jangan mengedit generated file manual.

## Route changes

`POST /upload/audio` menerima rekaman browser WebM/Opus dan OGG, selain MP3/WAV serta MP4/AAC dan M4A/AAC dari Android/iOS, hingga 10 MB. Server memeriksa magic bytes dan brand kontainer M4A, lalu memilih ekstensi dari tipe yang terdeteksi (misalnya `video/webm` menjadi `.webm`); nama dan MIME yang dikirim client tidak dipercaya. Respons unggahan tetap memakai `data.url` dan `data.filename`, sehingga client mobile tetap kompatibel.

Perubahan endpoint harus mencakup middleware/role, DTO/request validation, response/error schema, web service/schema, mobile datasource/model, migration/seed bila perlu, test handler/service, dan context. Upload, export CSV/PDF, webhook, dan download file bukan JSON biasa; dokumentasikan content type dan auth-nya.

## Refund Midtrans

Admin dapat mengajukan refund melalui `POST /admin/billing/transactions/{orderId}/refunds` dengan `amount` (IDR) dan `reason`. Endpoint hanya menerima transaksi Midtrans yang telah dibayar, memeriksa sisa nominal refund, dan untuk top-up memastikan saldo koin saat ini mencukupi sebelum permintaan dikirim ke Midtrans. Kunci refund unik disimpan sebelum panggilan provider; jika koneksi timeout atau respons ambigu, transaksi masuk antrean rekonsiliasi dan admin harus memeriksa Midtrans sebelum mengulang.

Webhook refund diproses terpisah dari status pembayaran. `refunds[]`/`refund_chargeback_id` mengidentifikasi refund, event webhook dibedakan berdasarkan payload, dan setiap refund provider hanya dicatat satu kali. Refund dianggap terkonfirmasi saat detail provider memiliki `bank_confirmed_at`; jumlah terkonfirmasi tersedia sebagai `refunded_amount`, sedangkan jumlah dari payload provider tersedia sebagai `provider_refund_amount_reported`. Detail status refund dan rekonsiliasi juga muncul pada daftar transaksi user/admin.

Refund top-up mengurangi koin secara proporsional terhadap nominal refund, termasuk bonus koin. Pengurangan bersifat atomik dan tidak membuat saldo negatif. Jika saldo tidak cukup setelah refund terkonfirmasi, transaksi ditandai `refund_reconciliation_status=pending`; admin dapat menarik sisa saldo atau mencatat koin yang sudah digunakan sebagai write-off dengan catatan audit. Refund sebagian/chargeback pada langganan tidak mengubah masa premium otomatis; admin harus memilih jumlah hari akses yang dicabut atau mempertahankan akses sebagai goodwill, lalu menutup kasus melalui `POST /admin/billing/transactions/{orderId}/refund-reconciliation`. Penjadwalan ulang menjaga langganan berurutan. Refund penuh yang terkonfirmasi mencabut langganan sumber dan menghitung ulang akses langganan lain.

## Musik dan konten presentasi

Objek lagu menyertakan `attribution`, `source_url`, dan `license_url`; client harus menampilkan kredit dan tautan tersebut saat menyajikan lagu berlisensi. Seeder musik presentasi memakai katalog Incompetech dan CC BY 4.0. Cakupan akun, kuota demo freemium, sumber editorial, dan kebijakan isi seed dijelaskan di `docs/SEED_CONTENT.md`.

## Verifikasi nomor saat login

`POST /auth/register` menerima `whatsapp_number` (nomor Indonesia, dinormalisasi menjadi `62...`). `POST /auth/login` untuk semua role dengan nomor belum terverifikasi mengembalikan `verification_required: true`, `verification_token`, dan `phone_required` tanpa JWT penuh. Jika `phone_required` benar, `POST /auth/verification/phone` menerima `verification_token` dan `whatsapp_number` untuk akun lama yang belum memiliki nomor. `POST /auth/verification/verify` menerima `verification_token` dan kode OTP enam digit, lalu mengembalikan respons login biasa. Nomor yang berubah lewat `PUT /auth/profile` perlu diverifikasi ulang pada login berikutnya. `POST /auth/forgot-password` tetap menerima email untuk menemukan akun, tetapi kode reset dikirim hanya ke nomor WhatsApp terverifikasi.
