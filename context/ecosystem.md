# Ecosystem Ruang Tenang
| Consumer | Base URL | Catatan |
| --- | --- | --- |
| Web | NEXT_PUBLIC_API_BASE_URL lengkap /api/v1 | Member, admin, moderator, dan mitra; fetch melalui services/http/client.ts |
| Mobile | BASE_URL host-only; client menambah /api/v1 | Member (`user`); Dio datasource dan repository Dart |
| API | /api/v1 | Sumber response, auth, persistence, dan OpenAPI |

## Invariant lintas repo

- API dan snapshot OpenAPI adalah sumber contract HTTP. Web dan mobile mengonsumsi `/api/v1` yang sama dengan konfigurasi base URL masing-masing.
- Perubahan route/DTO harus menjaga kompatibilitas kedua client dan ditinjau pada handler/test/OpenAPI API, service/schema web, serta datasource/model mobile.
- Web melayani member, admin, moderator, dan mitra; mobile hanya member (`user`). Pemeriksaan role dan ownership yang melindungi data harus tetap dilakukan API. API menyimpan state consent AI; client menjaga gate/disclaimer chat dan tidak menganggap gate UI sebagai kontrol akses server.
- Journal, chat/context AI, mood, profil, laporan moderasi, dan billing adalah data pribadi. Consent AI dan pengaturan privasi journal harus dihormati di semua client.
- Prompt/model AI, kuota, moderasi, dan kebijakan krisis berada di API. Client mempertahankan consent/disclaimer serta menangani response/error menurut contract.

Perubahan contract juga perlu memeriksa timezone, pagination, upload URL, error envelope, JWT expiry, entitlement, dan role middleware.

Route web yang dikirim melalui push notification, rekomendasi wellness, dan context AI harus memakai hub member kanonis: `/dashboard/community`, `/dashboard/journey`, dan `/dashboard/billing`. Detail forum memakai slug pada `/dashboard/community/forum/[slug]`; kisah memakai `/dashboard/community/stories/[id]`.

Pagination daftar dashboard baru bersifat opt-in pada endpoint yang sebelumnya selalu mengirim seluruh array. Mobile tetap memanggil endpoint tersebut tanpa query baru sehingga respons legacy tidak berubah; jika mobile kelak mengadopsi pagination, parser harus mendukung envelope flat `page`, `limit`, `total_items`, dan `total_pages`.

Repository sibling tidak menjadi dependency filesystem. Gunakan remote/revision yang disepakati dan dokumentasikan perubahan contract pada ketiga repo bila diperlukan.
