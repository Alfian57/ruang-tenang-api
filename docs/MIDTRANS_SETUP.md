# Setup dashboard Midtrans untuk Ruang Tenang

Integrasi ini memakai **Midtrans Snap**. Backend membuat transaksi Snap, web dapat membuka Snap dengan Client Key, dan mobile membuka `snap_url`. Status `paid` dan akses premium hanya berasal dari webhook yang diverifikasi backend.

## 1. Pilih lingkungan dan ambil kunci

1. Masuk ke [Midtrans Dashboard](https://dashboard.midtrans.com/), pilih **Sandbox** untuk pengujian.
2. Buka **Settings → Access Keys**. Salin **Server Key** ke konfigurasi backend dan **Client Key** ke konfigurasi web. Kunci Sandbox dan Production berbeda. [Dokumentasi Access Keys](https://docs.midtrans.com/docs/access-keys)
3. Isi konfigurasi sesuai lingkungan:

| Lingkungan | Backend API `.env` | Web `.env.local` |
| --- | --- | --- |
| Sandbox | `MIDTRANS_BASE_URL=https://app.sandbox.midtrans.com` dan `MIDTRANS_SERVER_KEY=<sandbox server key>` | `NEXT_PUBLIC_MIDTRANS_ENV=sandbox` dan `NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=<sandbox client key>` |
| Production | `MIDTRANS_BASE_URL=https://app.midtrans.com` dan `MIDTRANS_SERVER_KEY=<production server key>` | `NEXT_PUBLIC_MIDTRANS_ENV=production` dan `NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=<production client key>` |

Simpan Server Key hanya di backend. Client Key memang dipakai browser untuk Snap. Pastikan konfigurasi web dan backend menunjuk lingkungan yang sama. [Panduan Snap](https://docs.midtrans.com/docs/snap-snap-integration-guide), [panduan pindah ke Production](https://docs.midtrans.com/docs/switching-to-production-mode)

## 2. Atur notifikasi pembayaran

Di dashboard lingkungan yang sedang dipakai, buka **Settings → Payment Settings** dan isi **Payment notification URL** dengan URL HTTPS publik berikut:

```text
https://<domain-api>/api/v1/billing/webhooks/midtrans
```

Simpan pengaturan ini terpisah di Sandbox dan Production. Endpoint harus dapat menerima POST dari Midtrans tanpa autentikasi pengguna. Jangan gunakan URL localhost sebagai URL notifikasi dashboard. Ini adalah **Payment notification URL** untuk transaksi Snap, bukan **BI SNAP Notification URL**. [Dokumentasi Payment Settings](https://docs.midtrans.com/docs/payment-settings)

Backend memeriksa `signature_key`, `order_id`, `gross_amount`, status pembayaran, dan status fraud. Callback halaman selesai bukan bukti pembayaran. Setelah transaksi selesai, periksa `GET /api/v1/billing/status` atau `GET /api/v1/billing/transactions`; perubahan status akan tampak setelah webhook diproses. [Dokumentasi webhook Midtrans](https://docs.midtrans.com/docs/https-notification-webhooks)

## 3. Atur halaman selesai Snap

Set `FRONTEND_URL` backend ke origin web, misalnya `https://<domain-web>`. Saat membuat transaksi, backend mengirim callback finish ke `<FRONTEND_URL>/payment/success`. Di dashboard, pengaturan fallback ada di **Settings → Snap Preference → System Settings → Redirection URL / Finish URL**. URL dari permintaan transaksi mendapat prioritas atas pengaturan dashboard. [Dokumentasi Snap Preference](https://docs.midtrans.com/docs/snap-preference-snap-checkout-settings)

## 4. Uji Sandbox

1. Pastikan migration, API, dan web berjalan dengan kunci Sandbox yang sesuai.
2. Login sebagai pengguna dan buat checkout paket premium atau top-up dari aplikasi.
3. Selesaikan pembayaran memakai metode simulasi Sandbox yang tersedia.
4. Buka **Transactions** dan periksa Notification Log untuk transaksi itu. Pastikan backend menerima webhook 200 dan status transaksi berubah dari `pending` ke `paid`.
5. Periksa `GET /api/v1/billing/status`: premium harus aktif setelah pembayaran paket berhasil; kuota chat gratis harus berubah menjadi tak terbatas. Untuk transaksi pending, premium tetap tidak aktif.

Jika webhook terlambat atau gagal, lihat **Payment notification history / Notification Log** di dashboard. Cocokkan lingkungan kunci, URL HTTPS, status HTTP endpoint, dan `order_id`. Midtrans juga menyediakan [Get Transaction Status API](https://docs.midtrans.com/reference/get-transaction-status) untuk rekonsiliasi operasional. Jangan mengubah status menjadi `paid` dari callback frontend.

## 5. Mengajukan dan memantau refund

Gunakan dashboard admin Ruang Tenang untuk mengajukan refund; backend mengirim permintaan ke API Midtrans dengan `refund_key` unik. Nomor transaksi harus berstatus `paid`, alasan wajib diisi, dan jumlah refund tidak boleh melebihi sisa jumlah yang belum direfund. Untuk top-up, backend juga menolak permintaan jika saldo koin saat ini tidak cukup untuk menanggung pengurangan koin proporsional.

Setelah Midtrans menerima permintaan, status tampil sebagai **Menunggu konfirmasi Midtrans**. Backend belum mengurangi koin atau mencabut premium pada tahap ini. Midtrans mengirim notifikasi awal dan notifikasi lanjutan dengan `bank_confirmed_at`; status dan dampak refund diterapkan setelah konfirmasi provider tersebut. Backend menyimpan setiap `refund_chargeback_id`/`refund_key`, jadi refund parsial berulang tidak menggantikan catatan refund sebelumnya. Tombol **Sinkronkan Midtrans** mengambil status dan rincian terbaru dari API Midtrans jika notifikasi belum masuk. [Referensi API refund](https://docs.midtrans.com/reference/refund-transaction), [Get Transaction Status](https://docs.midtrans.com/reference/get-transaction-status)

Jika permintaan timeout atau provider memberikan respons ambigu, dashboard admin menandai transaksi untuk rekonsiliasi. Periksa **Transactions** dan status refund pada Midtrans sebelum mengambil keputusan; jangan mengajukan ulang sampai status permintaan pertama diketahui. Jika Midtrans melaporkan nominal refund yang belum cocok dengan rincian tersimpan, kasus tetap terkunci sampai tombol **Sinkronkan Midtrans** mendapatkan rincian yang cocok. Tindakan **Tandai ditolak** hanya boleh dipakai setelah operator memeriksa dashboard Midtrans dan mencentang konfirmasi bahwa provider menolak/tidak memproses refund. Untuk kasus refund terkonfirmasi tetapi saldo koin berkurang sebelum webhook diproses, admin dapat menarik sisa koin jika saldo masih cukup atau mencatat koin terpakai sebagai write-off. Setiap keputusan meminta catatan operator dan tercatat untuk audit.

Refund sebagian/chargeback langganan masuk antrean tinjauan karena durasi premium yang dicabut perlu diputuskan operator. Pada detail rekonsiliasi, admin dapat mengurangi sejumlah hari premium atau mempertahankan akses sebagai goodwill. Jika durasi dikurangi, backend menjadwalkan ulang langganan berikutnya agar masa aktif tetap berurutan. Refund penuh yang sudah dikonfirmasi otomatis mencabut langganan dari transaksi tersebut dan menghitung ulang langganan lain.

Riwayat transaksi menampilkan jumlah yang diminta, jumlah terkonfirmasi, jumlah yang dilaporkan Midtrans, dan status rekonsiliasi. Filter **Butuh Rekonsiliasi** pada dashboard admin menampilkan kasus yang perlu ditinjau. `refund_reconciliation_status=pending` juga dapat digunakan sebagai filter pada `GET /api/v1/admin/billing/transactions`.

## Catatan operasi

- Harga dan paket diambil dari katalog backend; pengguna tidak mengirim nominal harga.
- Nominal webhook harus tepat sama dengan transaksi lokal. Jika transaksi Midtrans sengaja menerima nominal berbeda, proses selisih itu perlu aturan bisnis terpisah sebelum produk diberikan.
- Refund top-up mengurangi koin secara proporsional dengan jumlah uang yang dikembalikan, termasuk bonus koin. Jika saldo tidak cukup saat konfirmasi, pengurangan masuk antrean operator dan saldo tidak dibuat negatif.
- Refund atau chargeback penuh pada langganan mencabut hak premium dari transaksi terkait dan menghitung ulang masa aktif langganan lain. Refund sebagian langganan memerlukan penyesuaian akses oleh operator.
- Migration `000118_add_midtrans_refund_tracking` menandai refund top-up lama untuk pemeriksaan saldo koin dan refund langganan lama untuk pemeriksaan hak premium karena rincian refund sebelumnya belum tercatat per refund.
