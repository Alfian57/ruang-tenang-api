# Ecosystem Ruang Tenang
| Consumer | Base URL | Catatan |
| --- | --- | --- |
| Web | NEXT_PUBLIC_API_BASE_URL lengkap /api/v1 | Fetch melalui services/http/client.ts |
| Mobile | BASE_URL host-only; client menambah /api/v1 | Dio datasource dan repository Dart |
| API | /api/v1 | Sumber response, auth, persistence, dan OpenAPI |

Web mendukung member/admin/mitra; mobile hanya member. Contract yang berubah harus diuji pada route API serta service web dan datasource mobile. Periksa timezone, pagination, upload URL, error envelope, JWT expiry, entitlement, dan role middleware saat melakukan perubahan lintas repo.

Route web yang dikirim melalui push notification, rekomendasi wellness, dan context AI harus memakai hub member kanonis: `/dashboard/community`, `/dashboard/journey`, dan `/dashboard/billing`. Detail forum memakai slug pada `/dashboard/community/forum/[slug]`; kisah memakai `/dashboard/community/stories/[id]`.

Repository sibling tidak menjadi dependency filesystem. Gunakan remote/revision yang disepakati dan dokumentasikan perubahan contract pada ketiga repo bila diperlukan.
