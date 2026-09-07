# Domain dan Access Control
## Role

- user/member: fitur self-care, konten, komunitas, journal, gamifikasi, billing pribadi, dan chat sesuai entitlement.
- admin: dashboard operasional, user/content management, billing, rewards, broadcast, dan konfigurasi moderation.
- mitra: organisasi, B2B plan/subscription, seat, onboarding, insight, payment, dan settings.
- Moderator dan status suspend/block/ban membatasi operasi sesuai middleware serta policy feature.

Role pada token/context hanya input authorization; setiap handler tetap memeriksa ownership dan status akun.

## Data sensitif

Journal, AI context, chat, mood, profile, moderation report, dan billing harus diperlakukan sebagai data pribadi. Endpoint public hanya boleh mengembalikan projection yang memang public. Toggle AI sharing dan journal settings adalah consent boundary.

## AI dan keselamatan

AI model dikonfigurasi per area melalui AI_*_MODEL; prompt terpusat di prompts/. Crisis keyword/moderation dan disclaimer bukan pengganti emergency service. Perubahan alur AI harus mempertahankan quota, logging akses, moderation, error fallback, dan disclaimer.

## Billing dan entitlement

Premium, top-up, payment transaction, subscription, feature usage, dan webhook memiliki state transitions. Webhook Midtrans harus idempotent dan diverifikasi. Client tidak boleh menentukan status paid/premium sendiri.

## Community dan gamifikasi

XP/activity, badge, leaderboard, guild, reward claim, progress, dan moderation saling terhubung. Perubahan reward/XP harus menjaga idempotensi dan mencegah privilege dari payload user.
