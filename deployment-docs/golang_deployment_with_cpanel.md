# Dokumentasi Deployment Golang API (`ruang-tenang-api`) ke Shared Hosting (cPanel)

Dokumen ini berisi panduan langkah demi langkah untuk melakukan deployment API backend berbasis **Golang** (`ruang-tenang-api`) ke shared hosting cPanel menggunakan **Node.js App (Phusion Passenger Wrapper)** dan **cPanel Terminal**.

---

## 📋 Prasyarat Lingkungan

- **Lokal:** Go (v1.20+), Git, Terminal / PowerShell.
- **Server:** 
  - Akses **cPanel Terminal** (fitur Terminal bawaan di cPanel).
  - Fitur **Setup Node.js App** (CloudLinux NodeJS Selector / Phusion Passenger).
  - Akses **File Manager** cPanel.
- **Subdomain & DNS:** `api.ruang-tenang.my.id` (A Record / CNAME telah mengarah ke IP hosting dan SSL aktif).
- **Database:** **MySQL** (versi 5.7+ / 8.0+) atau **MariaDB** (versi 10.3+) bawaan cPanel (dibuat melalui menu **MySQL Databases** / **MySQL Database Wizard** di cPanel).
  > [!TIP]
  > Backend `ruang-tenang-api` menggunakan database engine **MySQL/MariaDB** standar cPanel. Anda dapat membuat database dan user langsung dari menu cPanel tanpa perlu instalasi pihak ketiga.

---

## 1. Cross-Compile Biner Golang di Lokal

Shared hosting cPanel umumnya berjalan di arsitektur **Linux 64-bit (x86_64)**. Lakukan kompilasi silang (*cross-compilation*) dari komputer lokal dengan `CGO_ENABLED=0` agar menghasilkan biner statis tanpa dependensi library OS.

Buka terminal di root folder proyek lokal (`ruang-tenang-api`):

### Linux / macOS:
```bash
# 1. Build Biner Server Utama API
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app-main ./cmd/server/main.go

# 2. Build Biner Migration
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app-migrate ./cmd/migrate/main.go

# 3. Build Biner Seeder (Opsional jika butuh inisialisasi data)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app-seeder ./cmd/seeder/main.go
```

### Windows (PowerShell):
```powershell
# 1. Build Biner Server Utama API
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o app-main ./cmd/server/main.go

# 2. Build Biner Migration
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o app-migrate ./cmd/migrate/main.go

# 3. Build Biner Seeder
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"; go build -o app-seeder ./cmd/seeder/main.go
```

---

## 2. Buat Node.js Reverse Proxy Wrapper (`app.js`)

Karena cPanel menjalankan aplikasi via **Phusion Passenger** (yang mengawasi siklus hidup server Node.js melalui pemanggilan `http.Server.listen`), kita membuat file wrapper `app.js` yang bertindak sebagai **Reverse Proxy**:
- Menjalankan biner Golang `app-main` pada port internal (misal `3001`).
- Menerima request HTTP dari Passenger di port publik yang dialokasikan (`process.env.PORT`) lalu mem-forward request tersebut ke biner Golang.
- Menghentikan proses Golang secara otomatis ketika Passenger me-restart atau mematikan Node.js.

Buat file bernama `app.js` di komputer lokal (di root folder `ruang-tenang-api`):

```javascript
const http = require('http');
const { spawn } = require('child_process');
const path = require('path');

// Port internal tempat biner Golang berjalan
const GO_PORT = process.env.INTERNAL_PORT || 3001;
// Port yang dialokasikan oleh cPanel Passenger untuk Node.js
const PASSENGER_PORT = process.env.PORT || 3000;

console.log(`[Wrapper] Memulai biner Golang pada port internal ${GO_PORT}...`);

// 1. Eksekusi biner Golang dengan mengoper environment variable PORT ke GO_PORT
const golangApp = spawn(path.join(__dirname, 'app-main'), [], {
  cwd: __dirname,
  env: {
    ...process.env,
    PORT: GO_PORT.toString()
  }
});

golangApp.stdout.on('data', (data) => {
  process.stdout.write(`[Golang]: ${data}`);
});

golangApp.stderr.on('data', (data) => {
  process.stderr.write(`[Golang Error]: ${data}`);
});

golangApp.on('close', (code) => {
  console.log(`[Wrapper] Golang app keluar dengan kode: ${code}`);
  process.exit(code || 0);
});

// 2. Tangani graceful shutdown saat cPanel Passenger me-restart aplikasi
const shutdown = () => {
  if (golangApp && !golangApp.killed) {
    console.log('[Wrapper] Menghentikan biner Golang...');
    golangApp.kill('SIGTERM');
  }
};
process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);
process.on('exit', shutdown);

// 3. Reverse Proxy HTTP bawaan (tanpa perlu dependensi npm external)
const server = http.createServer((req, res) => {
  const options = {
    hostname: '127.0.0.1',
    port: GO_PORT,
    path: req.url,
    method: req.method,
    headers: req.headers
  };

  const proxyReq = http.request(options, (proxyRes) => {
    res.writeHead(proxyRes.statusCode, proxyRes.headers);
    proxyRes.pipe(res, { end: true });
  });

  proxyReq.on('error', (err) => {
    console.error(`[Proxy Error] Gagal terhubung ke Golang (${err.code}):`, err.message);
    res.writeHead(502, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      status: 'error',
      message: 'Bad Gateway: Backend Golang service belum siap atau tidak merespons.'
    }));
  });

  req.pipe(proxyReq, { end: true });
});

server.listen(PASSENGER_PORT, () => {
  console.log(`[Wrapper] Node.js Proxy aktif di port ${PASSENGER_PORT}, meneruskan ke Golang di port ${GO_PORT}`);
});
```

---

## 3. Rakit Bundle Deployment (`upload-api`)

Kumpulkan berkas-berkas biner dan folder yang dibutuhkan untuk server ke dalam folder sementara `upload-api`, lalu kompres menjadi `upload-api.zip`:

> [!NOTE]
> Folder `prompts/` **tidak perlu di-upload** karena seluruh prompt AI sudah di-embed langsung ke dalam biner `app-main` saat proses *compile* (`//go:embed`).

Lakukan langkah perakitan berikut di komputer lokal:

### Di Linux / macOS:
```bash
# 1. Buat folder sementara untuk bundling
mkdir -p upload-api

# 2. Salin biner hasil build dan wrapper app.js
cp app-main app-migrate app-seeder app.js upload-api/

# 3. Salin folder migrations, storage, dan siapkan folder uploads
cp -r migrations storage upload-api/
mkdir -p upload-api/uploads

# 4. Kompres seluruh isi folder upload-api menjadi upload-api.zip
cd upload-api
zip -r ../upload-api.zip .
cd ..
rm -rf upload-api
```

### Di Windows (PowerShell):
```powershell
# 1. Buat folder sementara
New-Item -ItemType Directory -Force -Path "upload-api"

# 2. Salin biner hasil build dan wrapper app.js
Copy-Item "app-main", "app-migrate", "app-seeder", "app.js" "upload-api\"

# 3. Salin folder migrations, storage, dan siapkan folder uploads
Copy-Item -Recurse -Force "migrations", "storage" "upload-api\"
New-Item -ItemType Directory -Force -Path "upload-api\uploads"

# 4. Kompres menjadi upload-api.zip
Compress-Archive -Path "upload-api\*" -DestinationPath "upload-api.zip" -Force
Remove-Item -Recurse -Force "upload-api"
```

---

### 📦 Struktur Berkas yang Benar di Dalam `upload-api.zip`

```text
upload-api.zip
├── app-main           <-- Biner API utama
├── app-migrate        <-- Biner Migration database
├── app-seeder         <-- Biner Seeder (opsional)
├── app.js             <-- Node.js Reverse Proxy Wrapper
├── migrations/        <-- Folder query SQL migration (wajib ada untuk app-migrate)
├── storage/           <-- Folder aset bawaan (gambar kategori/badge)
└── uploads/           <-- Folder penyimpanan berkas media user
```

---

## 4. Konfigurasi Node.js App di cPanel

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

## 5. Unggah & Ekstrak File di Server

1. Buka **File Manager** cPanel -> Masuk ke direktori `~/ruang-tenang-api/`.
2. Unggah file `upload-api.zip` ke dalam direktori tersebut.
3. Klik kanan pada file `upload-api.zip` -> pilih **Extract**.
4. Buat file `.env` di dalam `~/ruang-tenang-api/.env` (bisa lewat tombol **+ File** di File Manager atau via Terminal `nano .env`).

### Template Isi `.env` Production:
```env
# Konfigurasi Environment (Wajib)
APP_ENV=production
APP_TIMEZONE=Asia/Jakarta

# Port Internal yang digunakan Golang (di-proxy oleh app.js)
INTERNAL_PORT=3001

# Database Configuration (Pilih DATABASE_URL atau kombinasi DB_*)
# Contoh dengan DATABASE_URL:
DATABASE_URL=mysql://db_user:db_password@tcp(127.0.0.1:3306)/db_name?charset=utf8mb4&parseTime=True

# Atau jika menggunakan variabel terpisah (default port MySQL cPanel adalah 3306):
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=cpaneluser_dbuser
DB_PASSWORD=db_password_kuat
DB_NAME=cpaneluser_dbname

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

---

## 6. Eksekusi Permission, Migration, & Seed via cPanel Terminal

1. Buka menu **Terminal** di cPanel.
2. Jalankan perintah aktivasi environment yang telah disalin pada Langkah 4:
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

Perintah ini membaca konfigurasi koneksi langsung dari file `.env` di direktori yang sama (`DATABASE_URL` atau `DB_*`) dan file skrip SQL dari folder `migrations/`.

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

### 4. Eksekusi Seeder Database (Opsional)
Setelah struktur tabel terbentuk (melalui `up` atau `fresh`), Anda dapat mengisinya dengan data awal (katalog, badges, artikel, serta akun demo):
```bash
./app-seeder
```
*(Akun demo bawaan: `admin@ruang-tenang.com`, `mitra@ruang-tenang.com`, `gading@gmail.com` dengan password default `password`).*

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

- **Error `Dirty database version ...` saat menjalankan migration:**
  - Terjadi ketika migrasi sebelumnya gagal di tengah jalan (misal timeout atau sintaks SQL error).
  - Solusi: Periksa dan perbaiki query SQL di folder `migrations/`, lalu jalankan `./app-migrate force <versi_sebelumnya>` (misal `./app-migrate force 27`), kemudian jalankan kembali `./app-migrate up`.

- **Error `no such file or directory` pada folder migrations:**
  - Pastikan folder `migrations/` berada di direktori yang sama dengan biner `app-migrate` (`~/ruang-tenang-api/migrations/`) atau tentukan jalurnya di `.env` melalui `MIGRATIONS_PATH=./migrations`.

- **Error `503 Service Unavailable` atau `504 Gateway Timeout`:**
  - Cek file log cPanel di `~/ruang-tenang-api/stderr.log`.
  - Pastikan biner `app-main` sudah diberi izin eksekusi (`chmod +x app-main`).
  - Pastikan file `app.js` menggunakan implementasi Reverse Proxy seperti di Langkah 2 (bukan hanya `spawn` tanpa listener server).
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