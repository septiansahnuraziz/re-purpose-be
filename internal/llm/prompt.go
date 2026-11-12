package llm

import "fmt"

const outputSchema = `{
  "tweets": [
    {"id": 1, "stage": "1/5 • Hook Pembuka", "charCount": 206, "content": "..."}
  ],
  "linkedin": {
    "wordCount": 265,
    "readTime": "~1,2 mnt baca",
    "headline": "...",
    "hook": "...",
    "foldNote": "[...pembaca klik \"lihat lainnya\" di sini...]",
    "context": ["...", "..."],
    "frameworkTitle": "...",
    "frameworkPoints": [{"bold": "1. ...", "desc": "..."}],
    "closingQuestion": "...",
    "hashtags": ["#PersonalBranding"]
  },
  "videoScript": {
    "audioGuide": "Rekomendasi Musik: 120 BPM Lo-Fi / Upbeat Chill",
    "targetDuration": "Durasi Target: ~52 Detik",
    "blocks": [
      {
        "id": 1,
        "timeRange": "[00:00 - 00:05]",
        "title": "HOOK STOP-SCROLL",
        "tag": "Anti-Skip Hook",
        "tagColor": "sky",
        "visualCue": "Visual: ...",
        "dialogue": "\"...\""
      }
    ]
  },
  "telemetry": {
    "latency": "1.24 dtk",
    "confidence": "99.4%",
    "tokens": "824 Token",
    "semanticDeviation": "0.02",
    "wordCount": 1200,
    "channelCount": 15
  }
}`

func buildSystemPrompt() string {
	return `Anda adalah mesin repurposing konten untuk kreator Indonesia.
Ubah satu teks panjang menjadi aset media sosial siap tayang:
- Utas X/Twitter (4-5 tweet, tiap tweet maksimal 280 karakter)
- Postingan LinkedIn berbasis otoritas
- Naskah video pendek 60 detik (TikTok/Reels/Shorts)

Aturan:
1. Pertahankan fakta dari teks sumber, jangan menambah klaim baru.
2. Tulis dalam Bahasa Indonesia yang natural, bukan terjemahan kaku.
3. Patuhi batas karakter platform dan isi field charCount dengan jumlah karakter asli.
4. Balas HANYA dengan JSON valid tanpa markdown, sesuai skema berikut:
` + outputSchema
}

func buildUserPrompt(input Input, reroll bool) string {
	tone := map[string]string{
		"viral":     "Singkat & Viral (kalimat pendek, hook kuat, energik)",
		"executive": "Otoritas Eksekutif (profesional, berbasis data, tanpa clickbait berlebihan)",
		"deep":      "Analisis Mendalam (edukatif, terstruktur, menjelaskan konteks)",
	}[input.Tone]
	if tone == "" {
		tone = input.Tone
	}

	prompt := fmt.Sprintf(`Gaya nada: %s
Target audiens: %s

Teks sumber:
"""%s"""
`, tone, input.Audience, input.RawText)

	if reroll {
		prompt += `
Buat variasi BARU yang berbeda dari hasil sebelumnya: gunakan sudut pandang
pembuka yang lain, analogi berbeda, dan susunan argumen yang segar tanpa
mengubah fakta inti.
`
	}

	return prompt
}
