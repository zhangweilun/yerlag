package model

import "time"

// ShareLinkModel represents a shareable dashboard snapshot stored in the database.
type ShareLinkModel struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DashboardState string    `gorm:"type:text;not null" json:"dashboardState"` // JSON serialized
	SnapshotData   string    `gorm:"type:longtext" json:"snapshotData"`        // JSON serialized
	ExpiresAt      time.Time `gorm:"not null;index" json:"expiresAt"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName returns the database table name for ShareLinkModel.
func (ShareLinkModel) TableName() string {
	return "share_links"
}
