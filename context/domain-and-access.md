# Domain dan Access Control
## Role

- user/member: fitur self-care, konten, komunitas, journal, gamifikasi, billing pribadi, dan chat sesuai entitlement.
- admin: dashboard operasional, user/content management, billing, rewards, broadcast, dan konfigurasi moderation.
- mitra: organisasi, B2B plan/subscription, seat, onboarding, insight, payment, dan settings.
- Moderator dan status suspend/block/ban membatasi operasi sesuai middleware serta policy feature.

Role pada token/context hanya input authorization; setiap handler tetap memeriksa ownership dan status akun.

Pengguna baru memberikan nomor WhatsApp saat registrasi. Login pertama dan login setelah nomor berubah menghasilkan challenge OTP; token JWT penuh hanya diterbitkan setelah verifikasi. Pengguna lama tanpa nomor dapat menambahkan nomor melalui challenge login. Kode OTP berlaku 10 menit dengan batas 5 percobaan. Reset kata sandi dikirim lewat Fonnte hanya ke nomor yang sudah diverifikasi.
JWT lama tanpa klaim verifikasi nomor ditolak oleh middleware sehingga pemegang sesi lama harus login lagi dan menyelesaikan OTP.
Ketika nomor diubah, cache status akun dihapus dan middleware menolak sesi lama sampai nomor baru diverifikasi melalui login.

## Data sensitif

Journal, AI context, chat, mood, profile, moderation report, dan billing harus diperlakukan sebagai data pribadi. Endpoint public hanya boleh mengembalikan projection yang memang public. Toggle AI sharing dan journal settings adalah consent boundary.

## AI dan keselamatan

AI model dikonfigurasi per area melalui AI_*_MODEL; prompt terpusat di prompts/. Crisis keyword/moderation dan disclaimer bukan pengganti emergency service. Perubahan alur AI harus mempertahankan quota, logging akses, moderation, error fallback, dan disclaimer.

## Billing dan entitlement

Premium, top-up, payment transaction, subscription, feature usage, dan webhook memiliki state transitions. Callback Duitku harus memvalidasi HMAC, mencocokkan merchant/order/nominal/status, dan idempotent. Client tidak boleh menentukan status paid/premium sendiri. Kuota chat gratis dicatat per window dengan operasi database atomik agar permintaan bersamaan tidak melewati limit.

## Community dan gamifikasi

XP/activity, badge, leaderboard, reward claim, progress, dan moderation saling terhubung. Perubahan reward/XP harus menjaga idempotensi dan mencegah privilege dari payload user.
