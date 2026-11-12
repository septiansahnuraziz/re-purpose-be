package entity

import (
	"encoding/json"
	"time"
)

// Generation is the persisted record of one repurposing run.
type Generation struct {
	ID                string          `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	UserID            string          `gorm:"column:user_id;type:varchar(64);index" json:"userId"`
	SourceURL         string          `gorm:"column:source_url;type:varchar(1024)" json:"sourceUrl,omitempty"`
	Tone              string          `gorm:"column:tone;type:varchar(32)" json:"tone"`
	Audience          string          `gorm:"column:audience;type:varchar(255)" json:"audience"`
	Status            string          `gorm:"column:status;type:varchar(32)" json:"status"`
	RawText           string          `gorm:"column:raw_text;type:longtext" json:"-"`
	ResultJSON        string          `gorm:"column:result_json;type:longtext" json:"-"`
	Result            PlatformOutputs `gorm:"-" json:"result"`
	LatencyMS         int64           `gorm:"column:latency_ms" json:"latencyMs"`
	Tokens            int             `gorm:"column:tokens" json:"tokens"`
	Confidence        string          `gorm:"column:confidence;type:varchar(16)" json:"confidence"`
	SemanticDeviation string          `gorm:"column:semantic_deviation;type:varchar(16)" json:"semanticDeviation"`
	Model             string          `gorm:"column:model;type:varchar(64)" json:"model"`
	CreatedAt         time.Time       `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (Generation) TableName() string { return "generations" }

// SetResult serializes the generation output into ResultJSON for persistence.
func (g *Generation) SetResult(outputs *PlatformOutputs) error {
	raw, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	g.ResultJSON = string(raw)
	if outputs != nil {
		g.Result = *outputs
	}
	return nil
}

// ParseResult hydrates Result from ResultJSON.
func (g *Generation) ParseResult() (*PlatformOutputs, error) {
	if g.ResultJSON == "" {
		return &g.Result, nil
	}
	var outputs PlatformOutputs
	if err := json.Unmarshal([]byte(g.ResultJSON), &outputs); err != nil {
		return nil, err
	}
	g.Result = outputs
	return &outputs, nil
}

// ToResponse builds the API-safe representation.
func (g *Generation) ToResponse(outputs *PlatformOutputs) GenerationResponse {
	data := g.Result
	if outputs != nil {
		data = *outputs
	}

	return GenerationResponse{
		ID:        g.ID,
		Status:    g.Status,
		Tone:      g.Tone,
		Audience:  g.Audience,
		SourceURL: g.SourceURL,
		Data:      data,
		CreatedAt: g.CreatedAt,
	}
}
