# TODO API — RePurpose.ai Backend

Dokumen ini memetakan **6 titik integrasi API** yang ditandai dengan komentar
`🔗 [TITIK INTEGRASI API #n]` di repo `re-purpose-fe` ke endpoint backend yang sudah
di-scaffold, beserta sisa pekerjaan yang masih harus diselesaikan.

Legenda status:

- ✅ **Done** — endpoint berjalan dan kontraknya sudah cocok dengan frontend.
- 🟡 **Scaffold** — endpoint berjalan dengan implementasi baseline (mock/sederhana),
  perlu diperkuat sebelum produksi.
- ⬜ **Belum ada** — belum diimplementasikan sama sekali.

Ringkasan:

| #  | Fitur FE                          | Endpoint                          | Status | Sumber komentar (FE)              |
| -- | --------------------------------- | --------------------------------- | ------ | --------------------------------- |
| 1  | Ekstraksi konten dari URL         | `POST /api/content/extract-url`   | 🟡     | `src/components/WorkbenchInput.tsx` |
| 2  | Generasi konten multi-platform    | `POST /api/content/generate`      | 🟡     | `src/app/page.tsx`                |
| 3  | Acak ulang variasi (reroll)       | `POST /api/content/reroll`        | 🟡     | `src/app/page.tsx`                |
| 4  | Status layanan & model engine     | `GET /api/system/status`          | ✅     | `src/components/Navbar.tsx`, `Footer.tsx` |
| 5  | Data pengguna & tier langganan    | `GET /api/user/me`                | ✅     | `src/components/Navbar.tsx`       |
| 6  | Analitik klik sponsor (opsional)  | `POST /api/analytics/sponsor-click` | ✅   | `src/components/SponsoredBanner.tsx` |

Semua path di atas juga tersedia dengan prefix `/api/v1`.

---

## API #1 — Ekstraksi Konten dari URL

**Endpoint:** `POST /api/content/extract-url`
**Implementasi:** `internal/controller/http/content.go` → `ContentService.ExtractURL` →
`internal/service/extractor.go`

**Request**

```json
{ "url": "https://example.com/artikel" }
```

**Response (kontrak FE, field flat di root)**

```json
{
  "success": true,
  "title": "Judul artikel",
  "content": "Isi teks artikel yang sudah dibersihkan...",
  "wordCount": 812,
  "data": { "title": "...", "content": "...", "wordCount": 812 }
}
```

**Status:** 🟡 Scaffold. Fetch HTML + strip tag sudah jalan, termasuk validasi URL,
timeout, batas ukuran, dan guard host privat sederhana.

**Sisa TODO**

- [ ] Ganti parser regex dengan ekstraktor kelas *readability* (mis. `go-readability`,
      trafilatura, atau readability-service) agar boilerplate/navigasi tidak ikut terbawa.
- [ ] Guard SSRF lengkap: resolve DNS lalu tolak semua IP privat/loopback/link-local
      (termasuk redirect), bukan hanya cek literal host.
- [ ] Batasi redirect dan validasi scheme setiap hop.
- [ ] Dukung konten non-HTML (PDF, halaman berbasis JS) atau kembalikan error yang jelas.
- [ ] Cache hasil ekstraksi per URL (`enable_caching` + Redis).
- [ ] Simpan metadata sumber untuk atribusi di riwayat generasi.
- [ ] Rate limit per user agar tidak dipakai sebagai proxy scraping.

---

## API #2 — Generasi Konten Multi-Platform (LLM Engine)

**Endpoint:** `POST /api/content/generate`
**Implementasi:** `internal/controller/http/content.go` → `ContentService.Generate` →
`internal/service/content.go` → `internal/llm/*`

**Request**

```json
{
  "rawText": "Teks sumber...",
  "tone": "viral",              // viral | executive | deep
  "audience": "Founder & Solo Creator",
  "sourceUrl": "https://..."    // opsional
}
```

**Response (kontrak FE)**

```json
{
  "success": true,
  "message": "Success generate multi-platform content",
  "generationId": "uuid",
  "data": {
    "tweets": [{ "id": 1, "stage": "...", "charCount": 206, "content": "..." }],
    "linkedin": { "wordCount": 265, "headline": "...", "frameworkPoints": [] },
    "videoScript": { "audioGuide": "...", "blocks": [] },
    "telemetry": { "latency": "1.24 dtk", "confidence": "99.4%", "tokens": "824 Token" }
  }
}
```

> Frontend memakai `res.data` sebagai `PlatformOutputs`, sehingga `data` sengaja
> dikirim langsung sebagai objek output (bukan dibungkus lagi), dan `generationId`
> ditempatkan di root.

**Status:** 🟡 Scaffold. Endpoint berjalan dengan mock engine deterministik atau
OpenAI-compatible client. Hasil disimpan ke repository (memory/PostgreSQL) beserta telemetri.

**Sisa TODO**

- [ ] Prompt engineering final per tone + audience (few-shot, guardrail anti-halusinasi).
- [ ] Validasi ketat output LLM: pastikan jumlah tweet, batas 280 karakter, field wajib.
- [ ] Retry + fallback model saat provider error/timeout.
- [ ] Streaming SSE (`text/event-stream`) agar UI bisa menampilkan progres per platform.
- [ ] Pemotongan kredit (harian/bulanan) dan pencatatan usage per generasi.
- [ ] Idempotency key untuk mencegah double-charge saat retry.
- [ ] Simpan `rawText` secara aman (enkripsi/retensi) dan hormati kebijakan privasi.
- [ ] Batasi panjang input + moderasi konten.
- [ ] Telemetry nyata (token, latency, confidence) dari provider, bukan placeholder.
- [ ] Job async via asynq untuk input besar, dengan status `PENDING/PROCESSING/COMPLETED`.

---

## API #3 — Acak Ulang Variasi Konten (Reroll)

**Endpoint:** `POST /api/content/reroll`
**Implementasi:** sama dengan #2, dengan flag `Reroll = true`

**Request/Response:** identik dengan API #2.

**Status:** 🟡 Scaffold. Sudah membedakan prompt/temperature dan menandai hasil di
riwayat sebagai generasi baru.

**Sisa TODO**

- [ ] Simpan relasi `parentGenerationId` agar variasi bisa dilacak.
- [ ] Kirim hasil sebelumnya sebagai konteks agar variasi benar-benar berbeda.
- [ ] Batasi jumlah reroll per generasi/kuota.
- [ ] Deteksi duplikasi/similarity sebelum menyimpan variasi baru.

---

## API #4 — Status Layanan & Model Engine

**Endpoint:** `GET /api/system/status`
**Implementasi:** `internal/controller/http/system.go` → `internal/service/system.go`

**Response**

```json
{
  "success": true,
  "data": {
    "status": "operational",
    "version": "1.0.0",
    "environment": "dev",
    "uptimeSeconds": 120,
    "engine": { "provider": "mock", "model": "mock-repurposer-v1", "status": "operational" },
    "dependencies": [
      { "name": "database", "status": "disabled", "latencyMs": 0, "detail": "not configured" },
      { "name": "redis", "status": "disabled", "latencyMs": 0, "detail": "caching disabled" }
    ],
    "checkedAt": "2026-11-10T23:34:26+07:00"
  }
}
```

**Status:** ✅ Done. Ping DB/Redis dengan timeout 2 detik, status agregat
`operational`/`degraded`, plus status engine. FE dapat polling tiap 60 detik.

**Sisa TODO**

- [ ] Health check LLM nyata (probe ringan/kuota) agar status engine tidak statis.
- [ ] Tambahkan dependency lain bila dipakai (queue, object storage).
- [ ] Endpoint `/healthz` & `/readyz` terpisah untuk Kubernetes.
- [ ] Sertakan `region`/`maintenance` dan cache hasil singkat agar tidak membebani DB.

---

## API #5 — Data Pengguna & Status Tier Langganan

**Endpoint:** `GET /api/user/me`
**Implementasi:** `internal/controller/http/user.go` → `internal/service/user.go` →
`internal/repository/user.go`

**Response**

```json
{
  "success": true,
  "data": {
    "id": "usr_demo",
    "name": "Demo Creator",
    "email": "demo@repurpose.ai",
    "avatarUrl": "",
    "tier": "PRO",
    "dailyCredits": 25,
    "monthlyCredits": 250
  }
}
```

**Status:** ✅ Done untuk pembacaan profil dari repository (memory/PostgreSQL), dengan
demo user ter-seed. `auth.bypass=true` membuat endpoint langsung bisa dicoba.

**Sisa TODO**

- [ ] Endpoint login/register/refresh/logout + penerbitan JWT (`auth.bypass=false`).
- [ ] Integrasi identity provider (Google/OAuth) bila diperlukan.
- [ ] Sinkronisasi kuota dengan pemakaian nyata (reset harian/bulanan, timezone).
- [ ] Endpoint update profil & upload avatar.
- [ ] Integrasi billing/subscription (Midtrans/Xendit/Stripe) dan webhook tier.
- [ ] Cache profil singkat + invalidasi saat kuota berubah.
- [ ] Audit trail perubahan tier/kredit.

---

## API #6 — Analitik Klik Sponsor / Partner (Opsional)

**Endpoint:** `POST /api/analytics/sponsor-click`
**Implementasi:** `internal/controller/http/analytics.go` → `internal/service/sponsor.go` →
`internal/repository/sponsor_click.go`

**Request**

```json
{ "sponsorId": "supascale-01", "event": "click", "timestamp": "2026-11-10T12:00:00Z" }
```

**Response**

```json
{
  "success": true,
  "data": { "id": "uuid", "sponsorId": "supascale-01", "event": "click", "trackedAt": "..." }
}
```

**Status:** ✅ Done. Menerima JSON maupun body `navigator.sendBeacon` (text/plain),
memakai optional auth sehingga pengunjung anonim tetap tercatat.

**Sisa TODO**

- [ ] Deduplikasi event (per sesi/waktu) dan validasi `sponsorId` terhadap daftar partner.
- [ ] Simpan `sessionId`/UTM/referrer untuk atribusi lebih baik.
- [ ] Antre ke worker (asynq) agar request tracking tidak menunggu DB.
- [ ] Agregasi metrik (klik/impresi/CTR) + endpoint laporan untuk admin/partner.
- [ ] Retensi data & kepatuhan privasi (hash IP, consent).
- [ ] Pastikan endpoint tetap menerima `text/plain` di balik proxy/ingress.

---

## Pekerjaan Backend Lain (di luar 6 titik integrasi)

Fitur pendukung yang dibutuhkan agar produk siap produksi, belum masuk scaffold:

### Auth & Keamanan

- [ ] Tabel `refresh_tokens`, rotasi token, revoke per device.
- [ ] Middleware role/permission (admin, partner, user).
- [ ] Rate limiting (per IP & per user) + quota guard.
- [ ] CORS production yang ketat, security headers, request size limit.
- [ ] Secret management (jangan simpan API key di repo/config publik).

### Konten & Riwayat

- [ ] Endpoint berbagi hasil (share link publik, signed URL).
- [ ] Ekspor (Markdown/PDF/DOCX) dan integrasi publish langsung ke platform.
- [ ] Search & filter riwayat (tone, platform, tanggal).
<!-- markdownlint-disable-next-line -->
- [ ] Soft delete + retensi data.
- [ ] Bookmark/favorit generasi.

### Billing & Kuota

- [ ] Model paket (FREE/PRO), harga, dan masa aktif.
- [ ] Pemotongan kredit atomik (transaksi DB) + ledger usage.
- [ ] Webhook pembayaran + rekonsiliasi.
- [ ] Notifikasi kuota hampir habis.

### Observability & Operasional

- [ ] Structured logging dengan request id di semua layer (sudah sebagian).
- [ ] Tracing (OpenTelemetry) + metrics (Prometheus).
- [ ] Alerting untuk error rate LLM/ekstraktor.
- [ ] Graceful shutdown worker + dead-letter queue.
- [ ] CI: `go vet`, `go test`, lint, build image, scan dependency.
- [ ] Integration test dengan PostgreSQL/Redis (testcontainers).
- [ ] Load test endpoint generasi.

### Data & Migration

- [ ] Tabel `generations`: tambah `parent_generation_id`, `deleted_at`, index tambahan.
- [ ] Tabel `usage_ledger` untuk kredit.
- [ ] Tabel `sponsors` + `sponsor_events` teragregasi.
- [ ] Seed data tambahan untuk demo (beberapa generasi contoh).
- [ ] Backup & strategi migrasi rollback yang teruji.

---

## Catatan Integrasi Frontend

Frontend saat ini masih memakai simulasi (`setTimeout`). Untuk menyambungkan:

1. Buat `src/lib/api.ts` di `re-purpose-fe` dengan fungsi:
   `apiExtractUrl`, `apiGenerateContent`, `apiRerollContent`, `apiGetSystemStatus`,
   `apiGetUserProfile` (dan `apiTrackSponsorClick` opsional).
2. Arahkan base URL ke `http://localhost:8080` (env `NEXT_PUBLIC_API_BASE_URL`).
3. Ganti blok simulasi di `WorkbenchInput.tsx`, `page.tsx`, `Navbar.tsx`, `Footer.tsx`,
   dan `SponsoredBanner.tsx` dengan pemanggilan fungsi di atas.
4. Karena backend mengirim `success` + `data`/field flat sesuai komentar FE, tipe
   `PlatformOutputs` di `src/lib/presets.ts` bisa dipakai langsung.
