package alert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// EvaluateCondition — unit tests
// ---------------------------------------------------------------------------

func TestEvaluateCondition_GT(t *testing.T) {
	cond := AlertCondition{Type: "gt"}
	assert.True(t, EvaluateCondition(cond, 50, 51))
	assert.False(t, EvaluateCondition(cond, 50, 50))
	assert.False(t, EvaluateCondition(cond, 50, 49))
}

func TestEvaluateCondition_LT(t *testing.T) {
	cond := AlertCondition{Type: "lt"}
	assert.True(t, EvaluateCondition(cond, 50, 49))
	assert.False(t, EvaluateCondition(cond, 50, 50))
	assert.False(t, EvaluateCondition(cond, 50, 51))
}

func TestEvaluateCondition_GTPercentChange(t *testing.T) {
	cond := AlertCondition{Type: "gt_percent_change", ReferenceValue: 100}
	// (150 - 100) / 100 * 100 = 50% > 40 → true
	assert.True(t, EvaluateCondition(cond, 40, 150))
	// (130 - 100) / 100 * 100 = 30% > 40 → false
	assert.False(t, EvaluateCondition(cond, 40, 130))
}

func TestEvaluateCondition_LTPercentChange(t *testing.T) {
	cond := AlertCondition{Type: "lt_percent_change", ReferenceValue: 100}
	// (50 - 100) / 100 * 100 = -50% < -40 → true
	assert.True(t, EvaluateCondition(cond, 40, 50))
	// (80 - 100) / 100 * 100 = -20% < -40 → false
	assert.False(t, EvaluateCondition(cond, 40, 80))
}

func TestEvaluateCondition_GTPercentChange_ZeroReference(t *testing.T) {
	cond := AlertCondition{Type: "gt_percent_change", ReferenceValue: 0}
	assert.False(t, EvaluateCondition(cond, 10, 100))
}

func TestEvaluateCondition_LTPercentChange_ZeroReference(t *testing.T) {
	cond := AlertCondition{Type: "lt_percent_change", ReferenceValue: 0}
	assert.False(t, EvaluateCondition(cond, 10, 0))
}

func TestEvaluateCondition_UnknownType(t *testing.T) {
	cond := AlertCondition{Type: "unknown"}
	assert.False(t, EvaluateCondition(cond, 50, 100))
}

// ---------------------------------------------------------------------------
// EvaluateAlerts — unit tests
// ---------------------------------------------------------------------------

func TestEvaluateAlerts_TriggersEnabledRule(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   true,
			Message:   "High gas",
		},
	}
	svc := NewAlertServiceWithRules(rules)
	alerts := svc.EvaluateAlerts(map[string]float64{"gas_avg_gwei": 60})
	require.Len(t, alerts, 1)
	assert.Equal(t, "r1", alerts[0].RuleID)
	assert.Equal(t, "medium", alerts[0].Severity)
	assert.Equal(t, 60.0, alerts[0].CurrentValue)
	assert.Equal(t, 50.0, alerts[0].ThresholdValue)
	assert.False(t, alerts[0].Acknowledged)
}

func TestEvaluateAlerts_SkipsDisabledRule(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   false,
			Message:   "High gas",
		},
	}
	svc := NewAlertServiceWithRules(rules)
	alerts := svc.EvaluateAlerts(map[string]float64{"gas_avg_gwei": 60})
	assert.Empty(t, alerts)
}

func TestEvaluateAlerts_SkipsMissingMetric(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   true,
			Message:   "High gas",
		},
	}
	svc := NewAlertServiceWithRules(rules)
	alerts := svc.EvaluateAlerts(map[string]float64{"other_metric": 100})
	assert.Empty(t, alerts)
}

func TestEvaluateAlerts_DoesNotTriggerWhenBelowThreshold(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   true,
			Message:   "High gas",
		},
	}
	svc := NewAlertServiceWithRules(rules)
	alerts := svc.EvaluateAlerts(map[string]float64{"gas_avg_gwei": 30})
	assert.Empty(t, alerts)
}

func TestEvaluateAlerts_MultipleRules(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "gas_avg_gwei",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 50,
			Severity:  "medium",
			Enabled:   true,
			Message:   "High gas",
		},
		{
			ID:        "r2",
			MetricKey: "missed_slots_rate_pct",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 5,
			Severity:  "high",
			Enabled:   true,
			Message:   "Missed slots",
		},
		{
			ID:        "r3",
			MetricKey: "validator_daily_exits",
			Condition: AlertCondition{Type: "gt"},
			Threshold: 500,
			Severity:  "high",
			Enabled:   false, // disabled
			Message:   "Validator exits",
		},
	}
	svc := NewAlertServiceWithRules(rules)
	metrics := map[string]float64{
		"gas_avg_gwei":          60,
		"missed_slots_rate_pct": 6,
		"validator_daily_exits": 600,
	}
	alerts := svc.EvaluateAlerts(metrics)
	// r1 and r2 should trigger; r3 is disabled
	require.Len(t, alerts, 2)
	ruleIDs := []string{alerts[0].RuleID, alerts[1].RuleID}
	assert.Contains(t, ruleIDs, "r1")
	assert.Contains(t, ruleIDs, "r2")
}

func TestEvaluateAlerts_PercentChangeConditions(t *testing.T) {
	rules := []AlertRule{
		{
			ID:        "r1",
			MetricKey: "price",
			Condition: AlertCondition{
				Type:           "gt_percent_change",
				ReferenceValue: 1000,
			},
			Threshold: 20,
			Severity:  "medium",
			Enabled:   true,
			Message:   "Price up >20%",
		},
		{
			ID:        "r2",
			MetricKey: "tvl",
			Condition: AlertCondition{
				Type:           "lt_percent_change",
				ReferenceValue: 50000,
			},
			Threshold: 10,
			Severity:  "high",
			Enabled:   true,
			Message:   "TVL down >10%",
		},
	}
	svc := NewAlertServiceWithRules(rules)

	// price: (1250 - 1000)/1000*100 = 25% > 20 → triggers
	// tvl: (44000 - 50000)/50000*100 = -12% < -10 → triggers
	alerts := svc.EvaluateAlerts(map[string]float64{
		"price": 1250,
		"tvl":   44000,
	})
	require.Len(t, alerts, 2)
}

// ---------------------------------------------------------------------------
// SortAlertsBySeverity — unit tests
// ---------------------------------------------------------------------------

func TestSortAlertsBySeverity_BasicOrder(t *testing.T) {
	svc := NewAlertServiceWithRules(nil)
	alerts := []Alert{
		{ID: "1", Severity: "low"},
		{ID: "2", Severity: "high"},
		{ID: "3", Severity: "medium"},
		{ID: "4", Severity: "high"},
	}
	sorted := svc.SortAlertsBySeverity(alerts)
	require.Len(t, sorted, 4)
	assert.Equal(t, "high", sorted[0].Severity)
	assert.Equal(t, "high", sorted[1].Severity)
	assert.Equal(t, "medium", sorted[2].Severity)
	assert.Equal(t, "low", sorted[3].Severity)
}

func TestSortAlertsBySeverity_StableSort(t *testing.T) {
	svc := NewAlertServiceWithRules(nil)
	alerts := []Alert{
		{ID: "a", Severity: "medium"},
		{ID: "b", Severity: "high"},
		{ID: "c", Severity: "medium"},
		{ID: "d", Severity: "low"},
		{ID: "e", Severity: "high"},
	}
	sorted := svc.SortAlertsBySeverity(alerts)
	require.Len(t, sorted, 5)

	// high alerts should preserve original order: b, e
	assert.Equal(t, "b", sorted[0].ID)
	assert.Equal(t, "e", sorted[1].ID)
	// medium alerts should preserve original order: a, c
	assert.Equal(t, "a", sorted[2].ID)
	assert.Equal(t, "c", sorted[3].ID)
	// low
	assert.Equal(t, "d", sorted[4].ID)
}

func TestSortAlertsBySeverity_EmptySlice(t *testing.T) {
	svc := NewAlertServiceWithRules(nil)
	sorted := svc.SortAlertsBySeverity(nil)
	assert.Empty(t, sorted)
}

func TestSortAlertsBySeverity_SingleElement(t *testing.T) {
	svc := NewAlertServiceWithRules(nil)
	alerts := []Alert{{ID: "1", Severity: "low"}}
	sorted := svc.SortAlertsBySeverity(alerts)
	require.Len(t, sorted, 1)
	assert.Equal(t, "1", sorted[0].ID)
}

func TestSortAlertsBySeverity_DoesNotMutateOriginal(t *testing.T) {
	svc := NewAlertServiceWithRules(nil)
	alerts := []Alert{
		{ID: "1", Severity: "low"},
		{ID: "2", Severity: "high"},
	}
	_ = svc.SortAlertsBySeverity(alerts)
	// Original should be unchanged.
	assert.Equal(t, "low", alerts[0].Severity)
	assert.Equal(t, "high", alerts[1].Severity)
}

// ---------------------------------------------------------------------------
// DefaultRules — unit tests
// ---------------------------------------------------------------------------

func TestDefaultRules_Count(t *testing.T) {
	rules := DefaultRules()
	// 9 built-in rules as specified
	assert.Len(t, rules, 9)
}

func TestDefaultRules_AllEnabled(t *testing.T) {
	for _, r := range DefaultRules() {
		assert.True(t, r.Enabled, "built-in rule %s should be enabled", r.ID)
	}
}

func TestDefaultRules_UniqueIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, r := range DefaultRules() {
		assert.False(t, seen[r.ID], "duplicate rule ID: %s", r.ID)
		seen[r.ID] = true
	}
}

func TestDefaultRules_ValidSeverities(t *testing.T) {
	valid := map[string]bool{"high": true, "medium": true, "low": true}
	for _, r := range DefaultRules() {
		assert.True(t, valid[r.Severity], "invalid severity %q for rule %s", r.Severity, r.ID)
	}
}

func TestDefaultRules_ValidConditionTypes(t *testing.T) {
	valid := map[string]bool{"gt": true, "lt": true, "gt_percent_change": true, "lt_percent_change": true}
	for _, r := range DefaultRules() {
		assert.True(t, valid[r.Condition.Type], "invalid condition type %q for rule %s", r.Condition.Type, r.ID)
	}
}

func TestDefaultRules_SpecificRules(t *testing.T) {
	rules := DefaultRules()
	ruleMap := make(map[string]AlertRule)
	for _, r := range rules {
		ruleMap[r.ID] = r
	}

	// Req 2.5: burn anomaly
	r, ok := ruleMap["builtin-burn-anomaly"]
	require.True(t, ok)
	assert.Equal(t, "burn_daily_pct_of_avg", r.MetricKey)
	assert.Equal(t, "gt", r.Condition.Type)
	assert.Equal(t, 200.0, r.Threshold)
	assert.Equal(t, "high", r.Severity)

	// Req 3.6: high gas
	r, ok = ruleMap["builtin-high-gas"]
	require.True(t, ok)
	assert.Equal(t, "gas_avg_gwei", r.MetricKey)
	assert.Equal(t, "gt", r.Condition.Type)
	assert.Equal(t, 50.0, r.Threshold)

	// Req 10.7: validator exits
	r, ok = ruleMap["builtin-validator-exit"]
	require.True(t, ok)
	assert.Equal(t, "validator_daily_exits", r.MetricKey)
	assert.Equal(t, 500.0, r.Threshold)

	// Req 11.6: missed slots
	r, ok = ruleMap["builtin-missed-slots"]
	require.True(t, ok)
	assert.Equal(t, "missed_slots_rate_pct", r.MetricKey)
	assert.Equal(t, 5.0, r.Threshold)

	// Req 9.5: grayscale discount
	r, ok = ruleMap["builtin-grayscale-discount"]
	require.True(t, ok)
	assert.Equal(t, "grayscale_discount_pct", r.MetricKey)
	assert.Equal(t, "lt", r.Condition.Type)
	assert.Equal(t, -20.0, r.Threshold)
}

// ---------------------------------------------------------------------------
// NewAlertService — integration-like test (no DB)
// ---------------------------------------------------------------------------

func TestNewAlertService_NilDB_HasDefaultRules(t *testing.T) {
	svc := NewAlertService(nil)
	// Should be able to evaluate with default rules.
	alerts := svc.EvaluateAlerts(map[string]float64{
		"gas_avg_gwei": 60,
	})
	require.Len(t, alerts, 1)
	assert.Equal(t, "builtin-high-gas", alerts[0].RuleID)
}

func TestNewAlertService_NilDB_CRUDReturnsError(t *testing.T) {
	svc := NewAlertService(nil)
	assert.Error(t, svc.AddRule(AlertRule{}))
	assert.Error(t, svc.UpdateRule("x", AlertRule{}))
	assert.Error(t, svc.RemoveRule("x"))
	assert.Error(t, svc.ToggleRule("x", true))
	_, err := svc.GetActiveAlerts()
	assert.Error(t, err)
	_, err = svc.GetAlertHistory(7)
	assert.Error(t, err)
	assert.Error(t, svc.AcknowledgeAlert("x"))
}

// ---------------------------------------------------------------------------
// severityRank — unit tests
// ---------------------------------------------------------------------------

func TestSeverityRank(t *testing.T) {
	assert.Equal(t, 0, severityRank("high"))
	assert.Equal(t, 1, severityRank("medium"))
	assert.Equal(t, 2, severityRank("low"))
	assert.Equal(t, 3, severityRank("unknown"))
}
