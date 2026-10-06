# Dokumentasi Deployment Golang API (`ruang-tenang-api`) ke Shared Hosting (cPanel)

Dokumen ini berisi panduan langkah demi langkah untuk melakukan deployment API backend berbasis **Golang** (`ruang-tenang-api`) ke shared hosting cPanel menggunakan **Node.js App (Phusion Passenger Wrapper)** dan **cPanel Terminal**.

> [!IMPORTANT]
> Proses cross-compile & perakitan bundle di lokal sudah diotomasi oleh **`scripts/deploy_cpanel.sh`**. Ikuti langkah di bawah; jangan merakit bundle secara manual.

---

## 📋 Prasyarat Lingkungan

- **Lokal:** Go (v1.20+), Git, Terminal, `zip`.
- **Server:**
  - Akses **cPanel Terminal** (fitur Terminal bawaan di cPanel).
  - Fitur **Setup Node.js App** (CloudLinux NodeJS Selector / Phusion Passenger).
  - Akses **File Manager** cPanel.
- **Subdomain & DNS:** `api.ruang-tenang.my.id` (A Record / CNAME telah mengarah ke IP hosting dan SSL aktif).
- **Database:** **MySQL** (versi 5.7+ / 8.0+) atau **MariaDB** (versi 10.3+) bawaan cPanel (dibuat melalui menu **MySQL Databases** / **MySQL Database Wizard** di cPanel).
  > [!TIP]
  > Backend `ruang-tenang-api` menggunakan database engine **MySQL/MariaDB** standar cPanel. Anda dapat membuat database dan user langsung dari menu cPanel tanpa perlu instalasi pihak ketiga.

---

## 1. Cross-Compile & Rakit Bundle (Diotomasi)

Shared hosting cPanel umumnya berjalan di arsitektur **Linux 64-bit (x86_64)**. Script akan melakukan *cross-compilation* dengan `CGO_ENABLED=0` agar menghasilkan biner statis tanpa dependensi library OS.

Jalankan perintah berikut dari root folder `ruang-tenang-api` pada komputer lokal:

```bash
make deploy-cpanel
```

> Target ini memanggil `scripts/deploy_cpanel.sh`. Menjalankan `bash scripts/deploy_cpanel.sh` secara langsung juga sama. Untuk menghapus artefak hasil rakitan, gunakan `make deploy-clean`.

Script akan menjalankan seluruh rangkaian berikut secara berurutan:

1. **Preflight** — memastikan `go`, `zip`, `deployment/app.js`, `migrations/`, dan `storage/` tersedia.
2. **Cross-compile** — membangun `app-main`, `app-migrate`, `app-seeder` untuk `linux/amd64` dengan `CGO_ENABLED=0`.
3. **Rakit bundle** — menyalin ketiga biner + `app.js` + `migrations/` + `storage/`, serta menyiapkan `uploads/` kosong; biner diberi izin eksekusi.
4. **Zip** — menghasilkan arsip siap upload.

### Hasil

```text
ruang-tenang-api/
├── upload-api/                <-- isi siap upload (di-ignore git)
└── upload-api.zip             <-- arsip siap upload (di-ignore git)
```

> [!NOTE]
> Folder `prompts/` **tidak perlu di-upload** karena seluruh prompt AI sudah di-embed langsung ke dalam biner `app-main` saat proses *compile* (`//go:embed`).

---

### 📦 Struktur Berkas yang Benar di Dalam `upload-api.zip`

```text
upload-api.zip
└── ruang-tenang-api/      <-- Folder utama (Application Root saat diextract)
    ├── app-main           <-- Biner API utama
    ├── app-migrate        <-- Biner Migration database
    ├── app-seeder         <-- Biner Seeder (opsional)
    ├── app.js             <-- Node.js Reverse Proxy Wrapper (dari deployment/app.js)
    ├── migrations/        <-- Folder query SQL migration (wajib ada untuk app-migrate)
    ├── storage/           <-- Folder aset bawaan (gambar kategori/badge)
    └── uploads/           <-- Folder penyimpanan berkas media user
```

---

## 2. Node.js Reverse Proxy Wrapper (`app.js`)

Karena cPanel menjalankan aplikasi via **Phusion Passenger** (yang mengawasi siklus hidup server Node.js melalui pemanggilan `http.Server.listen`), kita memakai file wrapper `app.js` yang bertindak sebagai **Reverse Proxy**:

- Menjalankan biner Golang `app-main` pada port internal (misal `3001`).
- Menerima request HTTP dari Passenger di port publik yang dialokasikan (`process.env.PORT`) lalu mem-forward request tersebut ke biner Golang.
- Menghentikan proses Golang secara otomatis ketika Passenger me-restart atau mematikan Node.js.

> File ini **sudah tersedia** dan dilacak git di `ruang-tenang-api/deployment/app.js`. Script `scripts/deploy_cpanel.sh` menyalinnya ke `upload-api/app.js`. Anda **tidak perlu** membuatnya manual.

---

## 3. Konfigurasi Node.js App di cPanel

1. Login ke **cPanel** -> Pilih menu **Setup Node.js App**.
2. Klik tombol **Create Application**:
   - **Node.js Version:** Pilih versi LTS (misal `20.x` atau `22.x`).
   - **Application Mode:** `Production`.
   - **Application Root:** `ruang-tenang-api` (nama folder di root home user cPanel).
   - **Application URL:** Pilih domain/subdomain `api.ruang-tenang.my.id`.
   - **Application Startup File:** `app.js`
3. Klik **Create** / **Save**.
4. Di bagian atas halaman aplikasi, cPanel akan menampilkan perintah aktivasi *Virtual Environment*. Simpan perintah ini untuk digunakan di Terminal.
   *Contoh:*
   ```bash
   source /home/username/nodevenv/ruang-tenang-api/20/bin/activate && cd /home/username/ruang-tenang-api
   ```

---

## 4. Unggah & Ekstrak File di Server

1. Buka **File Manager** cPanel -> Masuk ke direktori **home** (`/home/username/`).
2. Unggah file `upload-api.zip` ke direktori home.
3. Klik kanan pada file `upload-api.zip` -> pilih **Extract**. Akan terbentuk folder `ruang-tenang-api/` yang berisi seluruh aplikasi (extract akan merge bila folder tersebut sudah ada).
4. Buat file `.env` di dalam `~/ruang-tenang-api/.env` (bisa lewat tombol **+ File** di File Manager atau via Terminal `nano .env`).

> [!WARNING]
> **WAJIB jalankan `chmod +x` SETELAH setiap kali extract.** Ekstraksi zip lewat File Manager cPanel sering menghilangkan bit *executable* pada biner. Bila `app-main` tidak executable, `app.js` gagal men-*spawn* backend dan aplikasi tidak akan berjalan.
> ```bash
> cd ~/ruang-tenang-api
> chmod +x app-main app-migrate app-seeder
> chmod -R 775 uploads storage
> ```

### Template Isi `.env` Production:
```env
# Konfigurasi Environment (Wajib)
APP_ENV=production
APP_TIMEZONE=Asia/Jakarta

# Port Internal yang digunakan Golang (di-proxy oleh app.js)
INTERNAL_PORT=3001

# Database Configuration (DATABASE_URL adalah satu-satunya konfigurasi database)
DATABASE_URL=mysql://db_user:db_password@tcp(127.0.0.1:3306)/db_name?charset=utf8mb4&parseTime=True

# JWT & Security (Wajib)
JWT_SECRET=ganti-dengan-kunci-rahasia-jwt-yang-sangat-kuat-dan-acak
JWT_EXPIRY_HOURS=24

# CORS & Domains (Wajib)
# Masukkan origin frontend produksi (pisahkan dengan koma jika lebih dari satu)
CORS_ALLOWED_ORIGINS=https://ruang-tenang.my.id,https://www.ruang-tenang.my.id
FRONTEND_URL=https://ruang-tenang.my.id
API_PUBLIC_URL=https://api.ruang-tenang.my.id

# AI Service (DeepSeek - Opsional jika menggunakan fitur Chat/Jurnal AI)
DEEPSEEK_API_KEY=your_deepseek_api_key
DEEPSEEK_BASE_URL=https://api.deepseek.com

# Payment Gateway (Duitku - Opsional)
DUITKU_SANDBOX=false
DUITKU_MERCHANT_CODE=
DUITKU_API_KEY=

# WhatsApp Gateway (Fonnte - Opsional untuk reset password WA)
FONNTE_TOKEN=
```

> [!NOTE]
> `PORT` tidak perlu ditulis di `.env`. Biner Golang menerima `PORT=3001` dari `app.js` (via `INTERNAL_PORT`). Variabel wajib yang divalidasi backend: `APP_ENV`, `JWT_SECRET`, `CORS_ALLOWED_ORIGINS`, dan `DATABASE_URL`.

> [!IMPORTANT]
> **`APP_ENV=production` wajib.** Bila `APP_ENV` bernilai `development`, backend mengaktifkan CORS wildcard (`Access-Control-Allow-Origin: *`) dan mengekspos route dev-only `POST /dev/cache/clear`. Router otomatis memakai `gin.ReleaseMode` untuk `APP_ENV` selain `development`, sehingga route dev tidak terekspos.

---

## 5. Amankan Berkas Aplikasi (`.htaccess`)

Karena Application Root sama dengan document root domain, LiteSpeed dapat menyajikan berkas aplikasi (`app-main`, `app-migrate`, `app-seeder`, `app.js`) sebagai file statis yang bisa diunduh publik. Tambahkan aturan deny ke `.htaccess` **tanpa menghapus blok Passenger** yang dibuat cPanel:

```apache
# Blokir akses HTTP ke berkas aplikasi (Passenger tetap bisa membacanya dari disk)
<FilesMatch "^(app-main|app-migrate|app-seeder|app\.js)$">
  Require all denied
</FilesMatch>
Options -Indexes
```

> [!CAUTION]
> JANGAN menghapus blok `# DO NOT REMOVE. CLOUDLINUX PASSENGER CONFIGURATION` di `.htaccess`. Hanya **tambahkan** aturan deny di atas. Jika Passenger dinonaktifkan, domain akan kembali menyajikan directory listing.

Verifikasi: `https://api.ruang-tenang.my.id/app-main` harus mengembalikan **403**, dan `/health` tetap **200**.

---

## 6. Eksekusi Permission, Migration, & Seed via cPanel Terminal

1. Buka menu **Terminal** di cPanel.
2. Jalankan perintah aktivasi environment yang telah disalin pada Langkah 3:
   ```bash
   source /home/username/nodevenv/ruang-tenang-api/20/bin/activate && cd /home/username/ruang-tenang-api
   ```
3. Berikan izin eksekusi (*executable permission*) ke biner Golang dan izin tulis ke folder upload & storage:
   ```bash
   chmod +x app-main app-migrate app-seeder
   chmod -R 775 uploads storage
   ```

---

### 📦 Panduan Lengkap Perintah Database Migration (`./app-migrate`)

Biner `app-migrate` dirancang sebagai CLI tool mandiri untuk mengelola skema database MySQL di server hosting tanpa memerlukan instalasi Go maupun migrate CLI bawaan host.

Perintah ini membaca konfigurasi koneksi langsung dari file `.env` di direktori yang sama (`DATABASE_URL`) dan file skrip SQL dari folder `migrations/`.

#### A. `migrate up` (Menerapkan Migrasi)
Gunakan perintah ini saat deployment pertama kali atau setiap kali ada file skrip migrasi baru yang diunggah ke folder `migrations/`:
```bash
./app-migrate up
# Atau cukup:
./app-migrate
```
- **Fungsi:** Menerapkan seluruh file migrasi `.up.sql` yang belum tercatat pada tabel `schema_migrations`.
- **Output jika berhasil:**
  ```text
  Migrations applied successfully (up-to-date)
  ```
- **Catatan:** Jika seluruh migrasi sudah diterapkan, perintah ini aman dijalankan ulang (bersifat *idempotent* dan tidak akan menduplikasi atau merusak data).

#### B. `migrate down` (Rollback / Membatalkan Migrasi)
Gunakan perintah ini jika ingin membatalkan skema migrasi sebelumnya:
```bash
# 1. Rollback 1 file migrasi terakhir (default):
./app-migrate down

# 2. Rollback sejumlah step tertentu (misal 2 migrasi terakhir):
./app-migrate down 2

# 3. Rollback SELURUH migrasi hingga database kosong:
./app-migrate down all
```
- **Output jika berhasil:**
  ```text
  Rolled back 1 migration step(s) successfully
  ```

> [!WARNING]
> Perintah `down` mengeksekusi file `.down.sql` terkait. Berhati-hatilah karena operasi seperti `DROP TABLE` atau `DROP COLUMN` pada file `.down.sql` akan menghapus tabel/kolom terkait beserta seluruh data di dalamnya!

#### C. `migrate fresh` (Reset Total Database & Migrasi Ulang dari Nol)
Gunakan perintah ini saat ingin mereset total struktur database dan membangun ulang seluruh tabel dari awal:
```bash
./app-migrate fresh
```
- **Fungsi:**
  1. Melakukan `DROP` pada semua tabel di database MySQL.
  2. Menjalankan ulang seluruh migrasi dari file `000001` hingga versi terakhir secara berurutan.
- **Output jika berhasil:**
  ```text
  Dropping all database tables...
  All tables dropped. Re-applying all migrations from scratch...
  Fresh migrations applied successfully
  ```

> [!CAUTION]
> **PERINGATAN KERAS:** `migrate fresh` akan **MENGHAPUS SEMUA TABEL DAN DATA** di database secara permanen! Hanya gunakan perintah ini di lingkungan *development*, *testing*, atau saat inisialisasi awal database demo. **JANGAN PERNAH** menjalankan perintah ini di database *production* yang sudah memiliki data pengguna aktif.

> [!TIP]
> Jika Anda menjalankan `fresh` pada lingkungan demo/staging, bersihkan juga berkas media lama di `uploads/` agar sinkron dengan database baru:
> ```bash
> rm -rf uploads/*
> ```

#### D. Perintah Bantu Migration (`version` & `force`)
- **Mengecek Versi Skema Database Saat Ini:**
  ```bash
  ./app-migrate version
  ```
  *Output:* `Current migration version: 28 (dirty: false)`

- **Memperbaiki Status Dirty State (`force`):**
  Jika terjadi kegagalan eksekusi SQL di tengah proses migrasi (misal karena sintaks query salah atau koneksi terputus), database akan berada dalam status `dirty: true` dan menolak migrasi berikutnya. Untuk mengatasinya:
  ```bash
  # Paksa status versi database kembali ke versi terakhir yang berhasil (misal versi 27):
  ./app-migrate force 27

  # Lalu jalankan kembali migrasi up:
  ./app-migrate up
  ```

---

### 4. Eksekusi Seeder Database (Opsional, tetapi disarankan untuk data demo)
Setelah struktur tabel terbentuk (melalui `up` atau `fresh`), Anda dapat mengisinya dengan data awal (katalog, badges, artikel, serta akun demo):
```bash
./app-seeder
```
*(Akun demo bawaan: `admin@ruang-tenang.com`, `mitra@ruang-tenang.com`, `gading@gmail.com` dengan password default `password`.)*

> [!IMPORTANT]
> Seeder juga **menyalin aset gambar** dari `storage/<tipe>/<file>` ke `uploads/<tipe>/<file>` (`internal/seed/presentation/utils.go`). Bila `uploads/` kosong setelah deploy, artikel/kategori/reward akan tampil tanpa gambar dan URL `/uploads/images/...` mengembalikan **404**. Jalankan `./app-seeder` (idempotent, aman diulang) untuk mengisinya, dan pastikan `uploads/` writable:
> ```bash
> chmod -R 775 uploads storage
> ./app-seeder
> ```

---

### 5. Verifikasi Biner Utama Manual
Uji coba jalankan biner utama secara manual sejenak di Terminal untuk memastikan tidak ada kendala koneksi DB maupun file `.env`:
```bash
./app-main
```
Jika muncul log:
```text
Database connected successfully
Server running on http://localhost:8080
```
Tekan tombol **`Ctrl + C`** untuk menghentikan pengujian manual.

---

## 7. Menjalankan & Memverifikasi Server

1. Kembali ke **cPanel** -> **Setup Node.js App**.
2. Pada baris aplikasi `api.ruang-tenang.my.id`, klik tombol **RESTART**.
3. Uji coba akses endpoint health check via browser atau Postman:
   ```text
   https://api.ruang-tenang.my.id/health
   ```
   **Respon yang diharapkan:**
   ```json
   {
     "service": "ruang-tenang-api",
     "status": "ok",
     "version": "1.0.0"
   }
   ```
4. Uji coba akses dokumentasi Swagger (jika diaktifkan):
   ```text
   https://api.ruang-tenang.my.id/swagger/index.html
   ```

---

## 🛠️ Troubleshooting Cepat

- **Domain menampilkan *directory listing* ("Index of /"), `/app-main` atau `/app.js` bisa di-download, dan `/health` mengembalikan 404:**
  - Ini berarti **Node.js App (Passenger) tidak aktif**. Web server menyajikan `~/ruang-tenang-api` sebagai direktori statis, sehingga request tidak pernah sampai ke `app.js`/Golang.
  - Solusi:
    1. Buka **Setup Node.js App** -> pastikan entri untuk `api.ruang-tenang.my.id` ada dengan **Application Root** `ruang-tenang-api`, **Application Startup File** `app.js`, **Application URL** `api.ruang-tenang.my.id`, dan Node.js version LTS.
    2. Jalankan `chmod +x app-main app-migrate app-seeder` di Terminal (ekstraksi File Manager sering menghilangkan izin executable).
    3. Klik **RESTART** pada aplikasi tersebut.
    4. Verifikasi ulang: `curl https://api.ruang-tenang.my.id/health` harus mengembalikan JSON `status: ok`.
  - Bila tetap statis, kemungkinan besar file `.htaccess` berisi konfigurasi Passenger di `~/ruang-tenang-api/` hilang. Perbaikan paling aman: **hapus lalu buat ulang Node.js App** di menu *Setup Node.js App* agar cPanel menulis ulang `.htaccess`-nya.
  - Fallback manual: buat `~/ruang-tenang-api/.htaccess` berikut (ganti `<USER>` dan `<VER>` sesuai perintah aktivasi cPanel, mis. `.../nodevenv/ruang-tenang-api/20/bin/node`):
    ```apache
    # DO NOT REMOVE. CLOUDLINUX PASSENGER CONFIGURATION BEGIN
    PassengerAppRoot "/home/<USER>/ruang-tenang-api"
    PassengerBaseURI "/"
    PassengerNodejs "/home/<USER>/nodevenv/ruang-tenang-api/<VER>/bin/node"
    PassengerAppType node
    PassengerStartupFile app.js
    # DO NOT REMOVE. CLOUDLINUX PASSENGER CONFIGURATION END
    ```
    Lalu `touch ~/ruang-tenang-api/tmp/restart.txt` untuk memicu restart Passenger.
  - File `.htaccess` ini juga yang mencegah `app-main`, `app.js`, dan file lain dapat di-download publik.

- **Error `502 Bad Gateway` (JSON dari wrapper `app.js`):**
  - Artinya `app.js` sudah jalan (Passenger aktif) tetapi biner Golang belum siap/crash saat startup.
  - Cek `stderr.log`, lalu jalankan `./app-main` manual di Terminal untuk melihat error (mis. `.env` kurang atau koneksi database gagal).

- **Error `Dirty database version ...` saat menjalankan migration:**
  - Terjadi ketika migrasi sebelumnya gagal di tengah jalan (misal timeout atau sintaks SQL error).
  - Solusi: Periksa dan perbaiki query SQL di folder `migrations/`, lalu jalankan `./app-migrate force <versi_sebelumnya>` (misal `./app-migrate force 27`), kemudian jalankan kembali `./app-migrate up`.

- **Error `no such file or directory` pada folder migrations:**
  - Pastikan folder `migrations/` berada di direktori yang sama dengan biner `app-migrate` (`~/ruang-tenang-api/migrations/`) atau tentukan jalurnya di `.env` melalui `MIGRATIONS_PATH=./migrations`.

- **Error `503 Service Unavailable` atau `504 Gateway Timeout`:**
  - Cek file log cPanel di `~/ruang-tenang-api/stderr.log`.
  - Pastikan biner `app-main` sudah diberi izin eksekusi (`chmod +x app-main`).
  - Pastikan file `app.js` menggunakan implementasi Reverse Proxy (bukan hanya `spawn` tanpa listener server).
  - Pastikan file `.env` sudah dibuat dengan benar dan kredensial database MySQL valid.

- **Error `502 Bad Gateway`:**
  - Ini menandakan server `app.js` aktif, namun biner `app-main` belum menyala atau mengalami crash saat startup.
  - Periksa `stderr.log` untuk melihat pesan error dari Golang (misal gagal connect database atau variabel `.env` yang kurang).
  - Coba jalankan `./app-main` di cPanel Terminal untuk melihat pesan error langsung di layar.

- **Error `bind: address already in use`:**
  - Terjadi jika ada proses `app-main` lama yang masih menggantung di port `3001`.
  - Matikan proses lama melalui cPanel Terminal:
    ```bash
    pkill -f app-main
    ```
  - Lalu klik **RESTART** di menu Setup Node.js App.

- **Error `missing required env vars: ...`:**
  - Pastikan variabel wajib di `.env` sudah terisi: `APP_ENV`, `JWT_SECRET`, `CORS_ALLOWED_ORIGINS`, serta konfigurasi database.
  - Ingat bahwa variabel mode aplikasi adalah **`APP_ENV`** (bukan `SERVER_ENV`).

- **CORS Error saat diakses dari Frontend:**
  - Pastikan domain frontend Anda telah dicantumkan pada variabel `CORS_ALLOWED_ORIGINS` di file `.env` (misal: `CORS_ALLOWED_ORIGINS=https://ruang-tenang.my.id`).

- **Gagal Upload File / Avatar:**
  - Pastikan folder `uploads/` memiliki izin tulis bagi web server:
    ```bash
    chmod -R 775 uploads storage
    ```
