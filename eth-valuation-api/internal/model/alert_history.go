package model

import "time"

// AlertHistoryModel represents a triggered alert record stored in the database.
type AlertHistoryModel struct {
	ID             string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RuleID         string     `gorm:"type:varchar(36);index" json:"ruleId"`
	TriggeredAt    time.Time  `gorm:"not null;index" json:"triggeredAt"`
	Severity       string     `gorm:"type:varchar(10);not null" json:"severity"`
	Title          string     `gorm:"type:varchar(255);not null" json:"title"`
	Message        string     `gorm:"type:text" json:"message"`
	MetricKey      string     `gorm:"type:varchar(100)" json:"metricKey"`
	CurrentValue   float64    `json:"currentValue"`
	ThresholdValue float64    `json:"thresholdValue"`
	Acknowledged   bool       `gorm:"default:false" json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName returns the database table name for AlertHistoryModel.
func (AlertHistoryModel) TableName() string {
	return "alert_history"
}
