package entity

import "time"

// PlatformOutputs mirrors src/lib/presets.ts (PlatformOutputs) in re-purpose-fe.
// Keep the JSON field names in sync with the frontend contracts.
type PlatformOutputs struct {
	Tweets      []TweetItem  `json:"tweets"`
	LinkedIn    LinkedInPost `json:"linkedin"`
	VideoScript VideoScript  `json:"videoScript"`
	Telemetry   Telemetry    `json:"telemetry"`
}

type TweetItem struct {
	ID        int    `json:"id"`
	Stage     string `json:"stage"`
	CharCount int    `json:"charCount"`
	Content   string `json:"content"`
}

type LinkedInPost struct {
	WordCount       int                      `json:"wordCount"`
	ReadTime        string                   `json:"readTime"`
	Headline        string                   `json:"headline"`
	Hook            string                   `json:"hook"`
	FoldNote        string                   `json:"foldNote"`
	Context         []string                 `json:"context"`
	FrameworkTitle  string                   `json:"frameworkTitle"`
	FrameworkPoints []LinkedInFrameworkPoint `json:"frameworkPoints"`
	ClosingQuestion string                   `json:"closingQuestion"`
	Hashtags        []string                 `json:"hashtags"`
}

type LinkedInFrameworkPoint struct {
	Bold string `json:"bold"`
	Desc string `json:"desc"`
}

type VideoScript struct {
	AudioGuide     string        `json:"audioGuide"`
	TargetDuration string        `json:"targetDuration"`
	Blocks         []ScriptBlock `json:"blocks"`
}

type ScriptBlock struct {
	ID        int    `json:"id"`
	TimeRange string `json:"timeRange"`
	Title     string `json:"title"`
	Tag       string `json:"tag"`
	TagColor  string `json:"tagColor"`
	VisualCue string `json:"visualCue"`
	Dialogue  string `json:"dialogue"`
}

type Telemetry struct {
	Latency           string `json:"latency"`
	Confidence        string `json:"confidence"`
	Tokens            string `json:"tokens"`
	SemanticDeviation string `json:"semanticDeviation"`
	WordCount         int    `json:"wordCount"`
	ChannelCount      int    `json:"channelCount"`
}

// ExtractURLRequest is TITIK INTEGRASI API #1.
type ExtractURLRequest struct {
	URL string `json:"url" binding:"required" example:"https://example.com/article"`
}

type ExtractURLResponse struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	WordCount int    `json:"wordCount"`
}

// GenerateContentRequest backs TITIK INTEGRASI API #2 and #3.
type GenerateContentRequest struct {
	RawText   string `json:"rawText" binding:"required" example:"Draft artikel panjang..."`
	Tone      string `json:"tone" binding:"required" example:"viral"`
	Audience  string `json:"audience" binding:"required" example:"Founder & Solo Creator"`
	SourceURL string `json:"sourceUrl,omitempty" example:"https://example.com/article"`
}

type GenerateContentResponse struct {
	GenerationID string          `json:"generationId"`
	Data         PlatformOutputs `json:"data"`
}

type GenerationResponse struct {
	ID        string          `json:"id"`
	Status    string          `json:"status"`
	Tone      string          `json:"tone"`
	Audience  string          `json:"audience"`
	SourceURL string          `json:"sourceUrl,omitempty"`
	Data      PlatformOutputs `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

type GenerationListResponse struct {
	Items []GenerationResponse `json:"items"`
	Meta  PaginationMeta       `json:"meta"`
}
