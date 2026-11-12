package entity

import "time"

// SponsorClickRequest is TITIK INTEGRASI API #6.
type SponsorClickRequest struct {
	SponsorID string     `json:"sponsorId" binding:"required" example:"supascale-01"`
	Event     string     `json:"event" binding:"required" example:"click"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type SponsorClick struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	SponsorID  string    `gorm:"column:sponsor_id;type:varchar(64);index" json:"sponsorId"`
	Event      string    `gorm:"column:event;type:varchar(32)" json:"event"`
	UserID     string    `gorm:"column:user_id;type:varchar(64);index" json:"userId,omitempty"`
	IP         string    `gorm:"column:ip;type:varchar(64)" json:"ip,omitempty"`
	UserAgent  string    `gorm:"column:user_agent;type:varchar(512)" json:"userAgent,omitempty"`
	OccurredAt time.Time `gorm:"column:occurred_at" json:"occurredAt"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (SponsorClick) TableName() string { return "sponsor_clicks" }

type SponsorClickResponse struct {
	ID        string    `json:"id"`
	SponsorID string    `json:"sponsorId"`
	Event     string    `json:"event"`
	TrackedAt time.Time `json:"trackedAt"`
}
