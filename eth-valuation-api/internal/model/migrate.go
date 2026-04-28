package model

import "gorm.io/gorm"

// AutoMigrate runs GORM auto-migration for all registered models.
// It creates or updates tables to match the current model definitions.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&AlertRuleModel{},
		&AlertHistoryModel{},
		&ShareLinkModel{},
	)
}
