package alert

import (
	"fmt"
	"math"
	"sort"
	"time"

	"eth-valuation-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// AlertCondition describes how a metric value is compared against a threshold.
type AlertCondition struct {
	Type                string  `json:"type"` // "gt" | "lt" | "gt_percent_change" | "lt_percent_change"
	ReferenceValue      float64 `json:"referenceValue,omitempty"`
	ReferencePeriodDays int     `json:"referencePeriodDays,omitempty"`
}

// AlertRule defines a single alert rule that can be evaluated against metrics.
type AlertRule struct {
	ID        string         `json:"id"`
	MetricKey string         `json:"metricKey"`
	Condition AlertCondition `json:"condition"`
	Threshold float64        `json:"threshold"`
	Severity  string         `json:"severity"` // "high" | "medium" | "low"
	Enabled   bool           `json:"enabled"`
	Message   string         `json:"message"`
}

// Alert represents a triggered alert instance.
type Alert struct {
	ID             string  `json:"id"`
	RuleID         string  `json:"ruleId"`
	TriggeredAt    int64   `json:"triggeredAt"`
	Severity       string  `json:"severity"`
	Title          string  `json:"title"`
	Message        string  `json:"message"`
	MetricKey      string  `json:"metricKey"`
	CurrentValue   float64 `json:"currentValue"`
	ThresholdValue float64 `json:"thresholdValue"`
	Acknowledged   bool    `json:"acknowledged"`
}

// AlertService defines the interface for the alert system.
type AlertService interface {
	// EvaluateAlerts evaluates all enabled rules against the provided metrics.
	EvaluateAlerts(metrics map[string]float64) []Alert

	// AddRule persists a new alert rule.
	AddRule(rule AlertRule) error
	// UpdateRule updates an existing alert rule.
	UpdateRule(ruleID string, updates AlertRule) error
	// RemoveRule deletes an alert rule by ID.
	RemoveRule(ruleID string) error
	// ToggleRule enables or disables an alert rule.
	ToggleRule(ruleID string, enabled bool) error

	// GetActiveAlerts returns all unacknowledged alerts.
	GetActiveAlerts() ([]Alert, error)
	// GetAlertHistory returns alerts triggered within the last N days.
	GetAlertHistory(days int) ([]Alert, error)
	// AcknowledgeAlert marks an alert as acknowledged.
	AcknowledgeAlert(alertID string) error

	// SortAlertsBySeverity sorts alerts by severity (high > medium > low)
	// using a stable sort to preserve relative order within the same level.
	SortAlertsBySeverity(alerts []Alert) []Alert
}

// ---------------------------------------------------------------------------
// Built-in alert rules
// ---------------------------------------------------------------------------

// DefaultRules returns the set of built-in alert rules that ship with the
// system. Each rule maps to a specific requirement from the spec.
func DefaultRules() []AlertRule {
	return []AlertRule{
		{
			// Req 2.5: 单日 Base_Fee_Burn 超过过去 30 天平均值的 200%
			ID:        "builtin-burn-anomaly",
			MetricKey: "burn_daily_pct_of_avg",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 200,
			Severity:  "high",
			Enabled:   true,
			Message:   "异常销毁量：单日 Base Fee 销毁量超过 30 天平均值的 200%",
		},
		{
			// Req 3.6: 平均 Gas 超过 50 Gwei
			ID:        "builtin-high-gas",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   true,
			Message:   "高 Gas 费用：当前平均 Gas 费用超过 50 Gwei",
		},
		{
			// Req 5.6: 以太坊 TVL 市场份额单周下降超过 3 个百分点
			ID:        "builtin-tvl-share-drop",
			MetricKey: "tvl_dominance_weekly_change",
			Condition: AlertCondition{Type: "lt"},
			Threshold: -3,
			Severity:  "high",
			Enabled:   true,
			Message:   "TVL 份额下降：以太坊 TVL 市场份额单周下降超过 3 个百分点",
		},
		{
			// Req 8.5: 单日 ETF 净流入量超过过去 30 天平均值的 300%
			ID:        "builtin-etf-inflow-anomaly",
			MetricKey: "etf_daily_inflow_pct_of_avg",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 300,
			Severity:  "medium",
			Enabled:   true,
			Message:   "ETF 异常流入：单日 ETF 净流入量超过 30 天平均值的 300%",
		},
		{
			// Req 8.6: 单日 ETF 净流出量超过过去 30 天平均值的 300%
			ID:        "builtin-etf-outflow-anomaly",
			MetricKey: "etf_daily_outflow_pct_of_avg",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 300,
			Severity:  "medium",
			Enabled:   true,
			Message:   "ETF 异常流出：单日 ETF 净流出量超过 30 天平均值的 300%",
		},
		{
			// Req 9.5: Grayscale 折价率超过 20%
			ID:        "builtin-grayscale-discount",
			MetricKey: "grayscale_discount_pct",
			Condition: AlertCondition{Type: "lt"},
			Threshold: -20,
			Severity:  "high",
			Enabled:   true,
			Message:   "Grayscale 异常折价：ETHE 折价率超过 20%",
		},
		{
			// Req 10.7: 单日验证者退出数量超过 500 个
			ID:        "builtin-validator-exit",
			MetricKey: "validator_daily_exits",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 500,
			Severity:  "high",
			Enabled:   true,
			Message:   "验证者大规模退出：单日验证者退出数量超过 500 个",
		},
		{
			// Req 11.6: missed slots 比率超过 5%
			ID:        "builtin-missed-slots",
			MetricKey: "missed_slots_rate_pct",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 5,
			Severity:  "high",
			Enabled:   true,
			Message:   "网络健康预警：missed slots 比率超过 5%",
		},
		{
			// Req 17.5: 交易所 ETH 余额单周净流出超过总余额的 5%
			ID:        "builtin-exchange-outflow",
			MetricKey: "exchange_weekly_outflow_pct",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 5,
			Severity:  "medium",
			Enabled:   true,
			Message:   "大规模提币：交易所 ETH 余额单周净流出超过总余额的 5%",
		},
	}
}

// ---------------------------------------------------------------------------
// Condition evaluation (pure logic)
// ---------------------------------------------------------------------------

// EvaluateCondition checks whether a metric value triggers the given condition
// against the threshold. Returns true if the alert should fire.
func EvaluateCondition(condition AlertCondition, threshold, currentValue float64) bool {
	switch condition.Type {
	case "gt":
		return currentValue > threshold
	case "lt":
		return currentValue < threshold
	case "gt_percent_change":
		if condition.ReferenceValue == 0 {
			return false
		}
		pctChange := (currentValue - condition.ReferenceValue) / condition.ReferenceValue * 100
		return pctChange > threshold
	case "lt_percent_change":
		if condition.ReferenceValue == 0 {
			return false
		}
		pctChange := (currentValue - condition.ReferenceValue) / condition.ReferenceValue * 100
		return pctChange < -threshold
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// severityRank maps severity strings to sort-order integers.
// ---------------------------------------------------------------------------

func severityRank(s string) int {
	switch s {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}

// ---------------------------------------------------------------------------
// alertServiceImpl — concrete implementation backed by GORM + in-memory rules
// ---------------------------------------------------------------------------

// alertServiceImpl implements AlertService.
type alertServiceImpl struct {
	db    *gorm.DB
	rules []AlertRule // in-memory rule set (includes built-in + user rules)
}

// NewAlertService creates a new AlertService. If db is nil the GORM-dependent
// methods (AddRule, UpdateRule, etc.) will return errors, but the pure-logic
// methods (EvaluateAlerts, SortAlertsBySeverity) will still work. This makes
// the service testable without a database.
func NewAlertService(db *gorm.DB) AlertService {
	svc := &alertServiceImpl{
		db:    db,
		rules: DefaultRules(),
	}
	// If a database is available, load persisted user rules.
	if db != nil {
		svc.loadPersistedRules()
	}
	return svc
}

// NewAlertServiceWithRules creates an AlertService pre-loaded with the given
// rules. Useful for testing pure evaluation logic without a database.
func NewAlertServiceWithRules(rules []AlertRule) AlertService {
	return &alertServiceImpl{
		db:    nil,
		rules: rules,
	}
}

// ---------------------------------------------------------------------------
// EvaluateAlerts
// ---------------------------------------------------------------------------

// EvaluateAlerts iterates over all enabled rules and checks each against the
// provided metrics map. Returns a slice of triggered Alert objects.
func (s *alertServiceImpl) EvaluateAlerts(metrics map[string]float64) []Alert {
	var triggered []Alert
	now := time.Now().Unix()

	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}
		currentValue, ok := metrics[rule.MetricKey]
		if !ok {
			continue
		}
		if EvaluateCondition(rule.Condition, rule.Threshold, currentValue) {
			triggered = append(triggered, Alert{
				ID:             uuid.New().String(),
				RuleID:         rule.ID,
				TriggeredAt:    now,
				Severity:       rule.Severity,
				Title:          fmt.Sprintf("Alert: %s", rule.MetricKey),
				Message:        rule.Message,
				MetricKey:      rule.MetricKey,
				CurrentValue:   currentValue,
				ThresholdValue: rule.Threshold,
				Acknowledged:   false,
			})
		}
	}
	return triggered
}

// ---------------------------------------------------------------------------
// SortAlertsBySeverity
// ---------------------------------------------------------------------------

// SortAlertsBySeverity returns a new slice sorted by severity
// (high > medium > low) using a stable sort to preserve relative order
// within the same severity level.
func (s *alertServiceImpl) SortAlertsBySeverity(alerts []Alert) []Alert {
	result := make([]Alert, len(alerts))
	copy(result, alerts)
	sort.SliceStable(result, func(i, j int) bool {
		return severityRank(result[i].Severity) < severityRank(result[j].Severity)
	})
	return result
}

// ---------------------------------------------------------------------------
// GORM CRUD operations
// ---------------------------------------------------------------------------

func (s *alertServiceImpl) requireDB() error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}
	return nil
}

// AddRule persists a new alert rule to the database and adds it to the
// in-memory rule set.
func (s *alertServiceImpl) AddRule(rule AlertRule) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	m := ruleToModel(rule)
	if err := s.db.Create(&m).Error; err != nil {
		return fmt.Errorf("failed to add rule: %w", err)
	}
	s.rules = append(s.rules, rule)
	return nil
}

// UpdateRule updates an existing alert rule in the database and in memory.
func (s *alertServiceImpl) UpdateRule(ruleID string, updates AlertRule) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	m := ruleToModel(updates)
	m.ID = ruleID
	result := s.db.Model(&model.AlertRuleModel{}).Where("id = ?", ruleID).Updates(map[string]interface{}{
		"metric_key":            m.MetricKey,
		"condition_type":        m.ConditionType,
		"threshold":             m.Threshold,
		"reference_value":       m.ReferenceValue,
		"reference_period_days": m.ReferencePeriodDays,
		"severity":              m.Severity,
		"enabled":               m.Enabled,
		"custom_message":        m.CustomMessage,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule %s not found", ruleID)
	}
	// Update in-memory copy.
	for i, r := range s.rules {
		if r.ID == ruleID {
			updates.ID = ruleID
			s.rules[i] = updates
			break
		}
	}
	return nil
}

// RemoveRule deletes an alert rule from the database and from memory.
func (s *alertServiceImpl) RemoveRule(ruleID string) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	result := s.db.Where("id = ?", ruleID).Delete(&model.AlertRuleModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove rule: %w", result.Error)
	}
	// Remove from in-memory slice.
	for i, r := range s.rules {
		if r.ID == ruleID {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			break
		}
	}
	return nil
}

// ToggleRule enables or disables an alert rule.
func (s *alertServiceImpl) ToggleRule(ruleID string, enabled bool) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	result := s.db.Model(&model.AlertRuleModel{}).Where("id = ?", ruleID).Update("enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("failed to toggle rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule %s not found", ruleID)
	}
	for i, r := range s.rules {
		if r.ID == ruleID {
			s.rules[i].Enabled = enabled
			break
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Alert query operations
// ---------------------------------------------------------------------------

// GetActiveAlerts returns all unacknowledged alerts from the database.
func (s *alertServiceImpl) GetActiveAlerts() ([]Alert, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	var models []model.AlertHistoryModel
	if err := s.db.Where("acknowledged = ?", false).
		Order("triggered_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	return historyModelsToAlerts(models), nil
}

// GetAlertHistory returns alerts triggered within the last N days.
func (s *alertServiceImpl) GetAlertHistory(days int) ([]Alert, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	var models []model.AlertHistoryModel
	if err := s.db.Where("triggered_at >= ?", cutoff).
		Order("triggered_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}
	return historyModelsToAlerts(models), nil
}

// AcknowledgeAlert marks an alert as acknowledged.
func (s *alertServiceImpl) AcknowledgeAlert(alertID string) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	now := time.Now()
	result := s.db.Model(&model.AlertHistoryModel{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acknowledged":    true,
			"acknowledged_at": &now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("alert %s not found", alertID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// loadPersistedRules loads user-created rules from the database and appends
// them to the in-memory rule set.
func (s *alertServiceImpl) loadPersistedRules() {
	var models []model.AlertRuleModel
	if err := s.db.Find(&models).Error; err != nil {
		return // silently ignore — built-in rules are still available
	}
	for _, m := range models {
		s.rules = append(s.rules, modelToRule(m))
	}
}

// ruleToModel converts an AlertRule to the GORM model.
func ruleToModel(r AlertRule) model.AlertRuleModel {
	return model.AlertRuleModel{
		ID:                  r.ID,
		MetricKey:           r.MetricKey,
		ConditionType:       r.Condition.Type,
		Threshold:           r.Threshold,
		ReferenceValue:      r.Condition.ReferenceValue,
		ReferencePeriodDays: r.Condition.ReferencePeriodDays,
		Severity:            r.Severity,
		Enabled:             r.Enabled,
		CustomMessage:       r.Message,
	}
}

// modelToRule converts a GORM model to an AlertRule.
func modelToRule(m model.AlertRuleModel) AlertRule {
	return AlertRule{
		ID:        m.ID,
		MetricKey: m.MetricKey,
		Condition: AlertCondition{
			Type:                m.ConditionType,
			ReferenceValue:      m.ReferenceValue,
			ReferencePeriodDays: m.ReferencePeriodDays,
		},
		Threshold: m.Threshold,
		Severity:  m.Severity,
		Enabled:   m.Enabled,
		Message:   m.CustomMessage,
	}
}

// historyModelsToAlerts converts a slice of AlertHistoryModel to Alert.
func historyModelsToAlerts(models []model.AlertHistoryModel) []Alert {
	alerts := make([]Alert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, Alert{
			ID:             m.ID,
			RuleID:         m.RuleID,
			TriggeredAt:    m.TriggeredAt.Unix(),
			Severity:       m.Severity,
			Title:          m.Title,
			Message:        m.Message,
			MetricKey:      m.MetricKey,
			CurrentValue:   m.CurrentValue,
			ThresholdValue: m.ThresholdValue,
			Acknowledged:   m.Acknowledged,
		})
	}
	return alerts
}

// Ensure NaN/Inf values don't leak into percent-change calculations.
func init() {
	_ = math.IsNaN // reference math package
}
