const http = require('http');
const fs = require('fs');
const { spawn, execSync } = require('child_process');
const path = require('path');

// Port internal tempat biner Golang berjalan
const GO_PORT = process.env.INTERNAL_PORT || 3001;
// Port yang dialokasikan oleh cPanel Passenger untuk Node.js
const PASSENGER_PORT = process.env.PORT || 3000;
const GO_BIN = path.join(__dirname, 'app-main');

let golangApp = null;
let shuttingDown = false;
let goRestarts = 0;
let goStartedAt = 0;

// JANGAN pernah biarkan error tak terduga mematikan wrapper. Bila proses Node
// mati, Passenger akan restart berulang dan request lain gagal (503/timeout).
process.on('uncaughtException', (err) => {
  console.error('[Wrapper] uncaughtException:', err && err.stack ? err.stack : err);
});
process.on('unhandledRejection', (reason) => {
  console.error('[Wrapper] unhandledRejection:', reason);
});

// 1. Jalankan biner Golang, dengan restart berjenjang bila ia keluar.
function startGolang() {
  if (!fs.existsSync(GO_BIN)) {
    console.error(`[Wrapper] Biner tidak ditemukan: ${GO_BIN}`);
  }

  goStartedAt = Date.now();
  console.log(`[Wrapper] Memulai biner Golang pada port internal ${GO_PORT}...`);

  golangApp = spawn(GO_BIN, [], {
    cwd: __dirname,
    env: {
      ...process.env,
      PORT: GO_PORT.toString()
    },
    stdio: ['ignore', 'pipe', 'pipe']
  });

  golangApp.stdout.on('data', (data) => {
    process.stdout.write(`[Golang]: ${data}`);
  });

  golangApp.stderr.on('data', (data) => {
    process.stderr.write(`[Golang Error]: ${data}`);
  });

  golangApp.on('error', (err) => {
    console.error(`[Wrapper] Gagal menjalankan biner Golang (${err.code || 'ERR'}): ${err.message}`);
    console.error('[Wrapper] Pastikan app-main executable: chmod +x app-main');
  });

  golangApp.on('close', (code) => {
    if (shuttingDown) return;

    // Reset hitungan bila proses sempat berjalan cukup lama (bukan crash-loop).
    if (Date.now() - goStartedAt > 60_000) {
      goRestarts = 0;
    }
    goRestarts += 1;
    const delay = Math.min(1_000 * goRestarts, 15_000);
    console.error(`[Wrapper] Golang keluar (code=${code}); restart #${goRestarts} dalam ${delay}ms`);
    setTimeout(() => {
      if (!shuttingDown) startGolang();
    }, delay);
  });
}

// 2. Tangani graceful shutdown saat cPanel Passenger me-restart aplikasi
const shutdown = () => {
  shuttingDown = true;
  if (golangApp && !golangApp.killed) {
    console.log('[Wrapper] Menghentikan biner Golang...');
    golangApp.kill('SIGTERM');
  }
};
process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

function killOrphanedProcesses() {
  try {
    execSync('pkill -9 -x app-main 2>/dev/null || killall -9 app-main 2>/dev/null || true');
  } catch {
    /* abaikan */
  }
}

killOrphanedProcesses();
setTimeout(() => {
  if (!shuttingDown) startGolang();
}, 600);

// 3. Reverse Proxy HTTP bawaan (tanpa dependensi npm external).
//    Semua jalur error ditangani agar satu request bermasalah tidak menjatuhkan
//    seluruh proses Node (yang membuat Passenger restart & sisanya 503).
const server = http.createServer((req, res) => {
  const options = {
    hostname: '127.0.0.1',
    port: GO_PORT,
    path: req.url,
    method: req.method,
    headers: req.headers
  };

  const proxyReq = http.request(options, (proxyRes) => {
    try {
      res.writeHead(proxyRes.statusCode || 502, proxyRes.headers);
    } catch (err) {
      console.error('[Proxy] writeHead gagal:', err.message);
      if (!res.headersSent) {
        res.writeHead(502, { 'Content-Type': 'application/json' });
      }
    }

    proxyRes.on('error', (err) => {
      console.error('[Proxy] proxyRes error:', err.message);
      res.destroy();
    });
    proxyRes.pipe(res);
  });

  proxyReq.on('error', (err) => {
    console.error(`[Proxy Error] Gagal terhubung ke Golang (${err.code}):`, err.message);
    if (!res.headersSent) {
      res.writeHead(502, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({
        status: 'error',
        message: 'Bad Gateway: Backend Golang service belum siap atau tidak merespons.',
        hint: 'Cek stderr.log; pastikan app-main executable (chmod +x app-main) dan .env valid.'
      }));
    } else {
      res.destroy();
    }
  });

  // Tangani pembatalan/stream error dari klien tanpa membuat proses crash.
  req.on('error', (err) => {
    console.error('[Proxy] req error:', err.message);
    proxyReq.destroy();
  });
  res.on('error', (err) => {
    console.error('[Proxy] res error:', err.message);
    proxyReq.destroy();
  });
  res.on('close', () => {
    if (!res.writableEnded) proxyReq.destroy();
  });

  req.pipe(proxyReq);
});

server.on('clientError', (err, socket) => {
  try {
    socket.end('HTTP/1.1 400 Bad Request\r\n\r\n');
  } catch {
    /* abaikan */
  }
});

server.listen(PASSENGER_PORT, () => {
  console.log(`[Wrapper] Node.js Proxy aktif di port ${PASSENGER_PORT}, meneruskan ke Golang di port ${GO_PORT}`);
});
