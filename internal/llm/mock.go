package llm

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

// MockClient is a deterministic, dependency-free engine used for local
// development and tests. Replace with a real provider for production output.
type MockClient struct{}

func NewMock() *MockClient { return &MockClient{} }

func (m *MockClient) Name() string  { return "mock" }
func (m *MockClient) Model() string { return "mock-repurposer-v1" }

func (m *MockClient) Generate(_ context.Context, input Input) (*entity.PlatformOutputs, Metrics, error) {
	outputs, metrics := buildMockOutputs(input)
	return outputs, metrics, nil
}

func buildMockOutputs(input Input) (*entity.PlatformOutputs, Metrics) {
	raw := utils.TruncateRunes(utils.NormalizeWhitespace(input.RawText), 4000)
	sentences := splitSentences(raw)
	if len(sentences) == 0 {
		sentences = []string{"Ide utama konten belum tersedia."}
	}

	hook := sentences[0]
	second := pickSentence(sentences, 1, "Ubah cara pandang audiens terhadap topik ini.")
	third := pickSentence(sentences, 2, "Kerangka kerja praktis yang bisa langsung dicoba.")
	fourth := pickSentence(sentences, 3, "Bagikan pengalaman Anda di kolom komentar.")

	toneLabel := map[string]string{
		"viral":     "Singkat & Viral",
		"executive": "Otoritas Eksekutif",
		"deep":      "Analisis Mendalam",
	}[input.Tone]
	if toneLabel == "" {
		toneLabel = input.Tone
	}

	if input.Reroll {
		hook = "Sudut pandang alternatif: " + hook
	}

	twitterStage := "1/5 • Hook Pembuka"
	tweets := []entity.TweetItem{
		newTweet(1, twitterStage, hook),
		newTweet(2, "2/5 • Geser Sudut Pandang", second),
		newTweet(3, "3/5 • Kerangka Utama", "3 langkah eksekusi:\n\n1. "+third+"\n2. Uji ke audiens kecil\n3. Ukur retensi, bukan sekadar reach"),
		newTweet(4, "4/5 • Diskusi & Aksi", fourth+"\n\nKalau bermanfaat, simpan dan bagikan utas ini."),
	}

	linkedinContext := []string{
		fmt.Sprintf("Waktu informasi masih langka, merangkum tren sudah cukup untuk mendapat perhatian. Sekarang kondisinya berbalik: informasi melimpah, tetapi wawasan praktis justru langka."),
		fmt.Sprintf("%s\n\nIni alasan mengapa spesialisasi vertikal lebih bernilai dibanding positioning generik.", second),
	}

	frameworkPoints := []entity.LinkedInFrameworkPoint{
		{Bold: "1. Ciptakan kerangka berpikir sendiri:", Desc: "Beri nama pada metodologi kerja Anda agar tidak sekadar membebek ide orang lain."},
		{Bold: "2. Bagikan realita di balik layar:", Desc: "Tunjukkan proses nyata, eksperimen yang gagal, dan pelajaran konkret."},
		{Bold: "3. Fokus pada relevansi, bukan sensasi:", Desc: "Komunitas kecil yang tepat lebih bernilai daripada audiens besar yang pasif."},
	}

	linkedinText := strings.Join([]string{
		hook,
		strings.Join(linkedinContext, " "),
		strings.Join([]string{frameworkPoints[0].Bold, frameworkPoints[1].Bold, frameworkPoints[2].Bold}, " "),
		fourth,
	}, " ")
	linkedinWords := utils.WordCount(linkedinText)

	readTime := fmt.Sprintf("~%.1f mnt baca", float64(linkedinWords)/200)
	if linkedinWords < 200 {
		readTime = "~1 mnt baca"
	}

	videoScript := entity.VideoScript{
		AudioGuide:     "Rekomendasi Musik: 120 BPM Lo-Fi / Upbeat Chill",
		TargetDuration: "Durasi Target: ~52 Detik",
		Blocks: []entity.ScriptBlock{
			{
				ID:        1,
				TimeRange: "[00:00 - 00:05]",
				Title:     "HOOK STOP-SCROLL",
				Tag:       "Anti-Skip Hook",
				TagColor:  "sky",
				VisualCue: "Tatap kamera dengan ekspresi serius, tunjukkan layar HP sebentar, lalu geser ke samping.",
				Dialogue:  fmt.Sprintf("%q", utils.TruncateRunes(hook, 140)),
			},
			{
				ID:        2,
				TimeRange: "[00:05 - 00:30]",
				Title:     "ISI UTAMA",
				Tag:       toneLabel,
				TagColor:  "indigo",
				VisualCue: "Potongan cepat (jump cut) dengan teks kunci di layar setiap 3-4 detik.",
				Dialogue:  fmt.Sprintf("%q", utils.TruncateRunes(second, 220)),
			},
			{
				ID:        3,
				TimeRange: "[00:30 - 00:48]",
				Title:     "KERANGKA KERJA",
				Tag:       "B-Roll + On Screen Text",
				TagColor:  "slate",
				VisualCue: "Tampilkan b-roll aktivitas kerja dan overlay 3 poin kerangka.",
				Dialogue:  fmt.Sprintf("%q", utils.TruncateRunes(third, 220)),
			},
			{
				ID:        4,
				TimeRange: "[00:48 - 01:00]",
				Title:     "KONVERSI & CTA",
				Tag:       "Comment Trigger",
				TagColor:  "primary",
				VisualCue: "Kembali ke close-up kamera, arahkan ke kolom komentar.",
				Dialogue:  fmt.Sprintf("%q", utils.TruncateRunes(fourth, 180)),
			},
		},
	}

	runeCount := utf8.RuneCountInString(raw)
	latencyMS := int64(900 + runeCount%500)
	if input.Reroll {
		latencyMS += 60
	}
	tokens := 700 + runeCount/4
	wordCount := utils.WordCount(raw)
	confidence := "99.4%"
	if input.Reroll {
		confidence = "98.8%"
	}

	outputs := &entity.PlatformOutputs{
		Tweets: tweets,
		LinkedIn: entity.LinkedInPost{
			WordCount:       linkedinWords,
			ReadTime:        readTime,
			Headline:        fmt.Sprintf("%s: %s", toneLabel, utils.TruncateRunes(hook, 120)),
			Hook:            utils.TruncateRunes(second, 220),
			FoldNote:        `[...pembaca klik "lihat lainnya" di sini...]`,
			Context:         linkedinContext,
			FrameworkTitle:  "3 pilar eksekusi yang bisa Anda tiru mulai hari ini:",
			FrameworkPoints: frameworkPoints,
			ClosingQuestion: "Bagaimana dengan Anda: masih nyaman dengan topik umum, atau sudah mulai memetakan posisi unik Anda sendiri?",
			Hashtags:        []string{"#PersonalBranding", "#StrategiKonten", "#KreatorIndonesia", "#B2BMarketing"},
		},
		VideoScript: videoScript,
		Telemetry: entity.Telemetry{
			Latency:           fmt.Sprintf("%.2f dtk", float64(latencyMS)/1000),
			Confidence:        confidence,
			Tokens:            fmt.Sprintf("%d Token", tokens),
			SemanticDeviation: "0.02",
			WordCount:         wordCount,
			ChannelCount:      15,
		},
	}

	metrics := Metrics{
		LatencyMS:         latencyMS,
		Tokens:            tokens,
		Confidence:        confidence,
		SemanticDeviation: "0.02",
	}

	return outputs, metrics
}

func newTweet(id int, stage, content string) entity.TweetItem {
	content = utils.TruncateRunes(content, 280)
	return entity.TweetItem{
		ID:        id,
		Stage:     stage,
		CharCount: utf8.RuneCountInString(content),
		Content:   content,
	}
}

func splitSentences(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?' || r == '\n'
	})

	sentences := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}
	return sentences
}

func pickSentence(sentences []string, index int, fallback string) string {
	if index < len(sentences) && strings.TrimSpace(sentences[index]) != "" {
		return sentences[index]
	}
	return fallback
}
