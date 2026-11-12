# RePurpose.ai Backend (`re-purpose-be`)

Boilerplate backend untuk **RePurpose.ai** — mesin yang mengubah satu teks panjang menjadi
aset konten multi-platform (utas X/Twitter, post LinkedIn, naskah video pendek).

Struktur project mengikuti pola `go-partner` (cmd/cobra → controller → service → repository,
entity interface, config Viper, commons, migrations, worker, Dockerfile), sehingga mudah
dikembangkan ke arah production.

> Fokus repo ini adalah **scaffold + kontrak API**. Logika LLM dan ekstraksi URL sudah
> berjalan (mock/nyata) tetapi masih berupa baseline yang perlu diperkuat sebelum produksi.

---

## 1. Stack

| Komponen     | Pilihan                                              |
| ------------ | ---------------------------------------------------- |
| HTTP         | Gin                                                  |
| Config       | Viper (YAML + override env)                          |
| CLI          | Cobra                                                |
| Logging      | Logrus                                               |
| Database     | PostgreSQL + GORM (`db.driver: postgres`) / in-memory (demo) |
| Migration    | sql-migrate                                          |
| Cache/Queue  | Redis + asynq                                        |
| Auth         | JWT HS256 (`golang-jwt/v5`)                          |
| LLM          | Mock (default) / OpenAI-compatible endpoint          |
| Dokumentasi  | `docs/openapi.yaml` + Swagger UI di `/docs`          |

---

## 2. Struktur Project

```
re-purpose-be/
├── main.go                     # entrypoint → internal/cmd
├── config.yaml                 # config default (dev, in-memory DB, mock LLM)
├── config.prod.yaml            # contoh config production
├── .env.example                # daftar environment variable
├── docs/openapi.yaml           # spesifikasi OpenAPI 3
├── commons/                    # konstanta domain (tone, status, tier, task)
├── config/                     # loader Viper + accessor bertipe
├── utils/
│   ├── helper.go               # util umum (word count, truncate, dll)
│   ├── string.go               # normalisasi & validasi URL
│   ├── context.go              # user/request id di context
│   ├── httpclient/             # HTTP client factory
│   └── httpresponse/           # envelope response + katalog error EN/ID
├── internal/
│   ├── cmd/                    # root, server, worker, migrate, injection
│   ├── controller/http/        # handler per domain + route registration
│   ├── controller/middleware/  # request id, access log, recovery, auth
│   ├── service/                # business logic (content, system, user, sponsor)
│   ├── repository/             # GORM repository + implementasi memory
│   ├── entity/                 # model, request/response, config, claims, interface
│   ├── llm/                    # abstraksi engine + mock + OpenAI client
│   ├── database/               # koneksi PostgreSQL/Redis + migrations
│   ├── job/                    # asynq task queue & processor
│   └── logger/                 # setup logrus, hooks, context logger
├── Dockerfile                  # image API
├── Dockerfile.worker           # image worker
├── docker-compose.yml          # postgres + redis + api + worker
└── Makefile                    # shortcut perintah harian
```

---

## 3. Menjalankan Cepat (tanpa PostgreSQL)

Mode default memakai `db.driver: memory` dan `llm.provider: mock`, jadi API langsung jalan
tanpa dependency eksternal. Data hilang saat proses berhenti (cocok untuk demo/development).

```bash
cd re-purpose-be
cp .env.example .env        # opsional, config.yaml sudah cukup
go mod download
go run . server
```

Server berjalan di `http://localhost:8080`.

```bash
# health check
curl -i http://localhost:8080/api/v1/ping

# status layanan
curl -s http://localhost:8080/api/v1/system/status | jq

# profil demo user (auth.bypass=true → user usr_demo)
curl -s http://localhost:8080/api/v1/user/me | jq

# generate konten (mock engine)
curl -s -X POST http://localhost:8080/api/content/generate \
  -H 'Content-Type: application/json' \
  -d '{"rawText":"Teks sumber minimal sepuluh kata agar lolos validasi mesin repurposing konten.","tone":"viral","audience":"Founder & Solo Creator"}' | jq
```

Swagger UI: <http://localhost:8080/docs> · spec mentah: `/docs/openapi.yaml`.

---

## 4. Menjalankan dengan PostgreSQL + Redis

```bash
# 1. siapkan database
createdb -U postgres repurpose
# atau: psql -U postgres -c "CREATE DATABASE repurpose;"

# 2. jalankan migration
DB_DRIVER=postgres go run . migrate --direction=up

# 3. jalankan API
DB_DRIVER=postgres ENABLE_CACHING=true go run . server

# 4. (opsional) jalankan worker
go run . worker
```

Atau semuanya lewat Docker:

```bash
docker compose up --build
```

`docker-compose.yml` menjalankan PostgreSQL 17, Redis 7, API (migrate lalu server), dan worker.

---

## 5. Konfigurasi

Semua key di `config.yaml` bisa dioverride lewat environment variable dengan aturan:
titik diganti underscore dan huruf besar. Contoh `db.host` → `DB_HOST`, `llm.api_key` → `LLM_API_KEY`.

| Key                     | Default                        | Keterangan                                        |
| ----------------------- | ------------------------------ | ------------------------------------------------- |
| `port`                  | `8080`                         | Port HTTP API                                     |
| `mode`                  | `dev`                          | `dev`/`prod` (prod → Gin release mode)            |
| `log_level`             | `debug`                        | Level logrus (`debug`/`info`/`warn`/`error`)      |
| `log_format`            | `text`                         | `text` untuk dev, `json` untuk production         |
| `db.driver`             | `memory`                       | `memory` atau `postgres`                          |
| `db.host/port/name/...` | —                              | Kredensial PostgreSQL                             |
| `db.ssl_mode`           | `disable`                      | `disable`/`require`/`verify-full` (sesuai DB)     |
| `enable_caching`        | `false`                        | Mengaktifkan koneksi Redis                        |
| `redis.cache_host`      | `redis://127.0.0.1:6379/0`     | Redis untuk cache                                 |
| `redis.worker_host`     | `redis://127.0.0.1:6379/1`     | Redis untuk queue worker                          |
| `auth.bypass`           | `true`                         | `true` → semua request menjadi demo user          |
| `auth.demo_user_id`     | `usr_demo`                     | User yang dipakai saat bypass                     |
| `auth.jwt.secret`       | `change-me-in-production`      | Secret JWT HS256                                  |
| `auth.jwt.iss`          | `repurpose-ai`                 | Issuer JWT                                        |
| `auth.jwt.exp`          | `24h`                          | Masa berlaku token                                |
| `llm.provider`          | `mock`                         | `mock` atau `openai`                              |
| `llm.base_url`          | `https://api.openai.com/v1`    | Base URL OpenAI-compatible                        |
| `llm.api_key`           | _(kosong)_                     | Jika `provider=openai` tapi kosong → fallback mock |
| `llm.model`             | `gpt-4o-mini`                  | Nama model                                        |
| `llm.timeout`           | `120s`                         | Timeout inferensi                                 |
| `extractor.timeout`     | `15s`                          | Timeout fetch artikel                             |
| `extractor.max_content_length` | `200000`                | Batas ukuran HTML yang dibaca                     |
| `cors.allowed_origins`  | `http://localhost:3000`        | Origin frontend yang diizinkan                    |

`config.prod.yaml` bisa dipakai dengan `CONFIG_FILE=config.prod.yaml`.

### Logging

Semua log melewati `internal/logger` (Logrus) dan otomatis membawa field
`service`, `version`, `build`, serta `environment`. Gunakan logger berbasis
context agar `requestId`, `traceId`, dan `userId` ikut tercatat:

```go
logger.WithFields(ctx, logrus.Fields{
    "generationId": id,
    "tone":         tone,
}).Info("generation completed")
```

- `log_format: text` → output ramah dibaca saat development.
- `log_format: json` → satu baris JSON per log, cocok untuk log collector.
- Middleware `RequestID` membuat/meneruskan `X-Request-ID`, `Logger` menulis
  access log (method, path, status, latency, client IP), dan `Recovery`
  menangkap panic tanpa mematikan proses.
- Gunakan `logger.FromContext(ctx)` di handler/service/worker; hook otomatis
  menambahkan `requestId`, `traceId`, dan `userId` dari context.

---

## 6. Endpoint

Semua endpoint tersedia di `/api/v1/...` dan alias `/api/...` (agar path di komentar
frontend bisa dipakai apa adanya).

| Method | Path                            | Fungsi                                      | Titik Integrasi FE |
| ------ | ------------------------------- | ------------------------------------------- | ------------------ |
| GET    | `/api/v1/ping`                  | Health check (204)                          | —                  |
| POST   | `/api/content/extract-url`      | Ekstraksi artikel dari URL                  | #1                 |
| POST   | `/api/content/generate`         | Generasi konten multi-platform              | #2                 |
| POST   | `/api/content/reroll`           | Variasi ulang hasil generasi                | #3                 |
| GET    | `/api/system/status`            | Status layanan + engine + dependency        | #4                 |
| GET    | `/api/user/me`                  | Profil, tier, kredit harian/bulanan         | #5                 |
| POST   | `/api/analytics/sponsor-click`  | Telemetri klik sponsor (sendBeacon-friendly)| #6                 |
| GET    | `/api/v1/content/generations`   | Riwayat generasi user                       | —                  |
| GET    | `/api/v1/content/generations/:id` | Detail generasi                           | —                  |
| GET    | `/api/v1/errors?lang=id`        | Katalog kode error EN/ID                    | —                  |
| GET    | `/docs`                         | Swagger UI                                  | —                  |

### Envelope Response

Sukses (mengikuti `go-partner` + flag `success`):

```json
{
  "id": "request-id",
  "appName": "RePurpose.ai API",
  "version": "1.0.0",
  "build": "1",
  "success": true,
  "message": "Success generate multi-platform content",
  "data": { }
}
```

Gagal:

```json
{
  "id": "request-id",
  "appName": "RePurpose.ai API",
  "version": "1.0.0",
  "build": "1",
  "success": false,
  "message": "content is too short to repurpose",
  "errorCode": "RPUR-400003",
  "data": null
}
```

Kode error memakai format `RPUR-<http status><serial>` (contoh `RPUR-400003`), katalog
lengkap ada di `utils/httpresponse/error.go` dan bisa diambil via `GET /api/v1/errors`.

> Catatan kontrak frontend: untuk **#1** field `title`, `content`, `wordCount` juga
> dikirim flat di root (selain di `data`), dan untuk **#2/#3** `data` berisi
> `PlatformOutputs` dengan `generationId` sebagai field root — sesuai komentar integrasi
> di repo `re-purpose-fe`.

---

## 7. Auth

- `auth.bypass: true` (default dev): setiap request diperlakukan sebagai `usr_demo`.
- `auth.bypass: false`: wajib header `Authorization: Bearer <jwt>`.
  Claims yang dibaca: `uid`, `email`, `tier`, `iss`, `exp`.
- `GET /api/analytics/sponsor-click` memakai *optional auth* agar tetap bisa mencatat
  event dari pengunjung anonim.

Endpoint auth/login belum termasuk scaffold — lihat `TODO_API.md`.

---

## 8. LLM Engine

`internal/llm/client.go` memilih implementasi berdasarkan `llm.provider`:

- `mock` — deterministik, tanpa network, berguna untuk demo & test.
- `openai` — memanggil `POST {base_url}/chat/completions` dengan `response_format: json_object`.
  Bisa dipakai untuk OpenAI, gateway Azure-compatible, OpenRouter, Ollama, dsb.
  Jika `api_key` kosong, otomatis fallback ke mock dengan warning.

Prompt dan skema output ada di `internal/llm/prompt.go`; bentuk JSON harus tetap sama
dengan `PlatformOutputs` di frontend (`src/lib/presets.ts`).

---

## 9. Worker & Job

- Worker asynq: `go run . worker` (membaca `redis.worker_host`).
- Task type: `content:generate` (`commons.TaskTypeContentGenerate`).
- Saat ini processor hanya menjalankan generasi dan menulis log — belum menyimpan hasil
  atau memotong kredit. Bagian ini ditandai `TODO` di `internal/job/task_processor.go`.

---

## 10. Perintah Umum

```bash
make run          # go run . server
make worker       # go run . worker
make build        # build binary ke bin/repurpose-be
make fmt          # gofmt
make vet          # go vet
make test         # go test ./...
make migrate-up   # DB_DRIVER=postgres go run . migrate --direction=up
make migrate-down # rollback 1 step
make docker-up    # docker compose up --build
make tidy         # go mod tidy
```

---

## 11. Status & Rencana Lanjutan

Daftar lengkap endpoint, kontrak, dan sisa pekerjaan per titik integrasi ada di
[`TODO_API.md`](./TODO_API.md). Ringkasnya:

- [x] Scaffold struktur + config + middleware + envelope + error catalog
- [x] 6 endpoint integrasi FE berjalan (mock LLM, in-memory DB)
- [x] OpenAPI + Swagger UI, migration, worker, Docker
- [ ] Auth/login + refresh token, rate limit per tier
- [ ] Pemotongan kredit & kuota, integrasi billing
- [ ] Ekstraktor readability-grade + guard SSRF berbasis DNS resolution
- [ ] Streaming (SSE) untuk generasi konten
- [ ] Observability (trace/metrics), job async end-to-end
