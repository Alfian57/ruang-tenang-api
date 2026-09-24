# Konten seeder presentasi

Seeder presentasi mengisi database lokal atau staging dengan konten berbahasa Indonesia untuk alur artikel, musik, playlist, forum, cerita, jurnal, gamifikasi, dan billing. Ini adalah konten editorial dan ilustrasi untuk demonstrasi; bukan pengalaman pasien, testimoni klinis, atau pengganti tenaga profesional. Artikel mengutip sumber kesehatan publik dan memakai bahasa edukasi yang tidak menjanjikan hasil pengobatan.

## Menjalankan seeder

```bash
make migrate-up
make seed
```

Pada database baru, `make setup` juga menyiapkan migration dan seed. Untuk database yang sudah ada, jalankan migration terlebih dahulu agar kolom atribusi lagu (`000119_add_song_license_metadata`) tersedia. Jalankan seeder hanya terhadap database development atau staging yang memang ditujukan untuk demo. Seeder ini tidak dijalankan oleh perubahan konten. Akun presentasi dibatasi pada alamat email demo yang ditentukan di `internal/seed/presentation/default_accounts.go`; pembersihan konten lama juga menggunakan daftar judul dan pemilik fixture yang spesifik. Seeder tidak seharusnya dipakai sebagai migrasi atau mekanisme pembersihan data produksi.

### Akun dan kuota freemium

`andhika@gmail.com` disiapkan sebagai akun gratis dengan pemakaian chat tepat satu pesan di bawah `CHAT_DAILY_MESSAGE_LIMIT` (default 100). Kunci jendela pemakaian dihitung memakai fungsi yang sama dengan enforcement API dan menghormati `CHAT_QUOTA_RESET_INTERVAL`; pesan berikutnya menghabiskan kuota, lalu permintaan sesudahnya mencapai batas normal. `gading@gmail.com` adalah akun premium untuk membandingkan entitlement. Nilai pemakaian di-seed ulang agar demo skenario batas dapat diulang.

Akun login demo yang dibuat seeder diberi nomor WhatsApp fixture dan status terverifikasi agar dapat dipakai pada alur demo tanpa mengirim OTP. Nomor fixture tersebut bukan nomor penerima aktif.

## Artikel dan panduan kesehatan

Artikel dalam seeder membahas stres, mindfulness, kebiasaan tidur, dukungan sosial, dan mencari pertolongan. Referensi yang dipakai:

- [WHO — Doing What Matters in Times of Stress](https://www.who.int/publications/i/item/9789240003927), panduan keterampilan praktis untuk menghadapi stres.
- [WHO — Stress](https://www.who.int/news-room/questions-and-answers/item/stress), pengenalan stres dan cara mengelolanya.
- [WHO EMRO — Dealing with stress](https://www.emro.who.int/mhps/dealing_with_stress.html), saran umum pengelolaan stres.
- [NCCIH — Meditation and Mindfulness: Effectiveness and Safety](https://www.nccih.nih.gov/health/meditation-and-mindfulness-effectiveness-and-safety), bukti dan batasan mindfulness serta meditasi.
- [CDC — About Sleep](https://www.cdc.gov/sleep/about/), informasi dasar tidur dan kesehatan.
- [Kementerian Kesehatan RI — FAQ Healing119.id](https://kesprimkom.kemkes.go.id/assets/uploads/contents/others/FAQ_Cegah_Bunuh_Diri%2C_Dukung_Kesehatan_Jiwa__Kenali_Layanan_Healing119.id.pdf), informasi layanan dukungan krisis di Indonesia.

Artikel bukan diagnosis atau saran medis individual. Informasi layanan dan URL sumber perlu ditinjau kembali sebelum publikasi produksi.

## Musik, sumber, dan lisensi

Daftar musik menggunakan komposisi Kevin MacLeod dari katalog resmi Incompetech. Setiap lagu menyimpan tautan karya, atribusi, dan lisensi CC BY 4.0. Aplikasi menampilkan kredit dan tautan sumber/lisensi pada pemutar. Berikan atribusi kepada Kevin MacLeod, tautkan halaman karya dan lisensi, serta pertahankan informasi tersebut jika musik digunakan ulang atau diadaptasi.

| Lagu | Kategori | Halaman sumber resmi |
| --- | --- | --- |
| Meditation Impromptu 01 | Piano | [Incompetech — ISRC USUAN1100163](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100163) |
| Meditation Impromptu 02 | Piano | [Incompetech — ISRC USUAN1100162](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100162) |
| Meditation Impromptu 03 | Piano | [Incompetech — ISRC USUAN1100161](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100161) |
| Friday Morning | Piano | [Incompetech — ISRC USUAN1100224](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100224) |
| Starry | Piano | [Incompetech — ISRC USUAN1100062](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100062) |
| Water Lily | Piano | [Incompetech — ISRC USUAN1400035](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1400035) |
| Ethereal Relaxation | Meditasi | [Incompetech — ISRC USUAN2100031](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN2100031) |
| That Zen Moment | Meditasi | [Incompetech — ISRC USUAN2400001](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN2400001) |
| Ever Mindful | Meditasi | [Incompetech — ISRC USUAN1700033](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1700033) |
| Organic Meditations Three | Meditasi | [Incompetech — ISRC USUAN1100759](https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1100759) |

Semua track memakai [Creative Commons Attribution 4.0 International](https://creativecommons.org/licenses/by/4.0/) (`https://creativecommons.org/licenses/by/4.0/`) dan atribusi `Kevin MacLeod (incompetech.com)`. Audio diputar langsung dari host Incompetech; pemutaran memerlukan koneksi internet dan bergantung pada ketersediaan host tersebut. Pastikan ketentuan katalog dan tautan masih berlaku saat rilis.

Panduan katalog dan format file tersedia pada [Incompetech Agent Section](https://incompetech.com/agent-section/). Kesesuaian track dipilih dari deskripsi katalog yang menyebut karakter tenang, santai, meditasi, atau piano lembut. Musik diposisikan sebagai pendamping relaksasi, bukan terapi atau klaim manfaat klinis.

## Ilustrasi

Empat ilustrasi orisinal dibuat untuk katalog ini, diproses menjadi WebP, dan disimpan di `storage/images/`:

| File | Penggunaan | Deskripsi visual |
| --- | --- | --- |
| `article-calm-start.webp` | Artikel dan sampul lagu | Jurnal terbuka, teh, tanaman, dan cahaya pagi yang lembut. |
| `article-sleep-routine.webp` | Artikel dan sampul lagu | Kamar malam yang tenang dengan lampu amber dan suasana istirahat. |
| `article-pause.webp` | Artikel dan sampul lagu | Seseorang berhenti sejenak di dekat jendela saat hujan. |
| `story-community-support.webp` | Cerita dan sampul lagu | Dua orang berbincang di taman setelah hujan, dilihat dari belakang. |

Palet ilustrasi netral (ivory, sage, biru lembut, terracotta, dan amber) agar terbaca bersama tema aplikasi merah, biru, maupun oranye. Cerita komunitas yang disemai diberi label sebagai ilustrasi editorial komposit, bukan kesaksian nyata. Ilustrasi dapat diganti pada direktori tersebut tanpa mengubah sumber audio.

## Interaksi komunitas dan privasi jurnal

Seeder menyiapkan diskusi dengan balasan yang relevan, suara dukungan, tanda jawaban, serta cerita editorial dan reaksinya. Aktivitas ini hanya menggunakan akun presentasi yang ditentukan seeder. Jurnal publik contoh memakai isi refleksi ringan yang aman untuk dibagikan dan diberi penanda contoh; jurnal privat tetap privat. Izin berbagi jurnal dengan AI tidak dinyalakan secara umum.

Saat mengganti konten demo, jaga agar nama, isi, dan interaksi tetap fiktif/komposit kecuali ada persetujuan publikasi yang terdokumentasi. Jangan masukkan data pribadi atau catatan kesehatan orang sungguhan ke seeder.
