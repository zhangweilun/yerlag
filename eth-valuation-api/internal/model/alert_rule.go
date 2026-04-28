package model

import "time"

// AlertRuleModel represents a user-defined alert rule stored in the database.
type AlertRuleModel struct {
	ID                  string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	MetricKey           string    `gorm:"type:varchar(100);not null;index" json:"metricKey"`
	ConditionType       string    `gorm:"type:varchar(30);not null" json:"conditionType"`
	Threshold           float64   `gorm:"not null" json:"threshold"`
	ReferenceValue      float64   `json:"referenceValue"`
	ReferencePeriodDays int       `json:"referencePeriodDays"`
	Severity            string    `gorm:"type:varchar(10);not null" json:"severity"`
	Enabled             bool      `gorm:"default:true" json:"enabled"`
	CustomMessage       string    `gorm:"type:text" json:"customMessage"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the database table name for AlertRuleModel.
func (AlertRuleModel) TableName() string {
	return "alert_rules"
}
