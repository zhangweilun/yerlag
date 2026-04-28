package alert

// Feature: eth-valuation-dashboard, Property 2: 预警规则评估正确性
//
// For any metric value, any alert rule (with condition type "gt", "lt",
// "gt_percent_change", or "lt_percent_change"), and any enabled/disabled state,
// the AlertService.EvaluateAlerts function SHALL trigger an alert if and only if
// (1) the rule is enabled AND (2) the metric value satisfies the rule's condition
// against its threshold. Disabled rules SHALL never produce alerts regardless of
// metric values.
//
// **Validates: Requirements 2.5, 3.6, 5.6, 8.5, 8.6, 9.5, 10.7, 11.6, 15.2, 15.4, 17.5**

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// Generators
// ---------------------------------------------------------------------------

// genConditionType generates one of the four valid condition types.
func genConditionType() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"gt", "lt", "gt_percent_change", "lt_percent_change"})
}

// genSeverity generates one of the three valid severity levels.
func genSeverity() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"high", "medium", "low"})
}

// genMetricValue generates a finite float64 suitable for metric values.
func genMetricValue() *rapid.Generator[float64] {
	return rapid.Float64Range(-1e9, 1e9)
}

// genThreshold generates a finite float64 suitable for thresholds.
func genThreshold() *rapid.Generator[float64] {
	return rapid.Float64Range(-1e6, 1e6)
}

// genPositiveReference generates a positive reference value (non-zero) for
// percent-change conditions.
func genPositiveReference() *rapid.Generator[float64] {
	return rapid.Float64Range(1e-6, 1e9)
}

// genMetricKey generates a random metric key string.
func genMetricKey() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		"gas_avg_gwei",
		"burn_daily_pct_of_avg",
		"tvl_dominance_weekly_change",
		"etf_daily_inflow_pct_of_avg",
		"etf_daily_outflow_pct_of_avg",
		"grayscale_discount_pct",
		"validator_daily_exits",
		"missed_slots_rate_pct",
		"exchange_weekly_outflow_pct",
		"price",
		"tvl",
	})
}

// genAlertRule generates a random AlertRule with the given enabled state.
func genAlertRule(enabled bool) *rapid.Generator[AlertRule] {
	return rapid.Custom[AlertRule](func(t *rapid.T) AlertRule {
		condType := genConditionType().Draw(t, "conditionType")
		cond := AlertCondition{Type: condType}
		if condType == "gt_percent_change" || condType == "lt_percent_change" {
			cond.ReferenceValue = genPositiveReference().Draw(t, "referenceValue")
		}
		return AlertRule{
			ID:        rapid.StringMatching(`^rule-[a-z0-9]{4}$`).Draw(t, "ruleID"),
			MetricKey: genMetricKey().Draw(t, "metricKey"),
			Condition: cond,
			Threshold: genThreshold().Draw(t, "threshold"),
			Severity:  genSeverity().Draw(t, "severity"),
			Enabled:   enabled,
			Message:   "test alert",
		}
	})
}

// ---------------------------------------------------------------------------
// Helper: expected trigger logic (mirrors EvaluateCondition)
// ---------------------------------------------------------------------------

func shouldTrigger(rule AlertRule, value float64) bool {
	if !rule.Enabled {
		return false
	}
	switch rule.Condition.Type {
	case "gt":
		return value > rule.Threshold
	case "lt":
		return value < rule.Threshold
	case "gt_percent_change":
		if rule.Condition.ReferenceValue == 0 {
			return false
		}
		pct := (value - rule.Condition.ReferenceValue) / rule.Condition.ReferenceValue * 100
		return pct > rule.Threshold
	case "lt_percent_change":
		if rule.Condition.ReferenceValue == 0 {
			return false
		}
		pct := (value - rule.Condition.ReferenceValue) / rule.Condition.ReferenceValue * 100
		return pct < -rule.Threshold
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// Property tests
// ---------------------------------------------------------------------------

// TestProperty2_GT_Condition verifies that for any random metric value and "gt"
// condition, the alert triggers if and only if value > threshold.
func TestProperty2_GT_Condition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		threshold := genThreshold().Draw(t, "threshold")
		value := genMetricValue().Draw(t, "value")

		cond := AlertCondition{Type: "gt"}
		result := EvaluateCondition(cond, threshold, value)

		if value > threshold {
			assert.True(t, result, "gt: value %v > threshold %v should trigger", value, threshold)
		} else {
			assert.False(t, result, "gt: value %v <= threshold %v should not trigger", value, threshold)
		}
	})
}

// TestProperty2_LT_Condition verifies that for any random metric value and "lt"
// condition, the alert triggers if and only if value < threshold.
func TestProperty2_LT_Condition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		threshold := genThreshold().Draw(t, "threshold")
		value := genMetricValue().Draw(t, "value")

		cond := AlertCondition{Type: "lt"}
		result := EvaluateCondition(cond, threshold, value)

		if value < threshold {
			assert.True(t, result, "lt: value %v < threshold %v should trigger", value, threshold)
		} else {
			assert.False(t, result, "lt: value %v >= threshold %v should not trigger", value, threshold)
		}
	})
}

// TestProperty2_GTPercentChange_Condition verifies that for any random metric
// value and "gt_percent_change" condition with non-zero reference, the alert
// triggers if and only if percent change > threshold.
func TestProperty2_GTPercentChange_Condition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		threshold := genThreshold().Draw(t, "threshold")
		value := genMetricValue().Draw(t, "value")
		ref := genPositiveReference().Draw(t, "reference")

		cond := AlertCondition{Type: "gt_percent_change", ReferenceValue: ref}
		result := EvaluateCondition(cond, threshold, value)

		pctChange := (value - ref) / ref * 100
		if pctChange > threshold {
			assert.True(t, result,
				"gt_percent_change: pct %v > threshold %v should trigger (value=%v, ref=%v)",
				pctChange, threshold, value, ref)
		} else {
			assert.False(t, result,
				"gt_percent_change: pct %v <= threshold %v should not trigger (value=%v, ref=%v)",
				pctChange, threshold, value, ref)
		}
	})
}

// TestProperty2_LTPercentChange_Condition verifies that for any random metric
// value and "lt_percent_change" condition with non-zero reference, the alert
// triggers if and only if percent change < -threshold.
func TestProperty2_LTPercentChange_Condition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		threshold := genThreshold().Draw(t, "threshold")
		value := genMetricValue().Draw(t, "value")
		ref := genPositiveReference().Draw(t, "reference")

		cond := AlertCondition{Type: "lt_percent_change", ReferenceValue: ref}
		result := EvaluateCondition(cond, threshold, value)

		pctChange := (value - ref) / ref * 100
		if pctChange < -threshold {
			assert.True(t, result,
				"lt_percent_change: pct %v < -threshold %v should trigger (value=%v, ref=%v)",
				pctChange, -threshold, value, ref)
		} else {
			assert.False(t, result,
				"lt_percent_change: pct %v >= -threshold %v should not trigger (value=%v, ref=%v)",
				pctChange, -threshold, value, ref)
		}
	})
}

// TestProperty2_DisabledRules_NeverTrigger verifies that disabled rules never
// trigger regardless of metric values.
func TestProperty2_DisabledRules_NeverTrigger(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rule := genAlertRule(false).Draw(t, "disabledRule")
		value := genMetricValue().Draw(t, "value")

		svc := NewAlertServiceWithRules([]AlertRule{rule})
		alerts := svc.EvaluateAlerts(map[string]float64{rule.MetricKey: value})

		assert.Empty(t, alerts,
			"disabled rule %s should never trigger (value=%v, threshold=%v, cond=%s)",
			rule.ID, value, rule.Threshold, rule.Condition.Type)
	})
}

// TestProperty2_MissingMetric_NeverTrigger verifies that rules for missing
// metrics never trigger.
func TestProperty2_MissingMetric_NeverTrigger(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rule := genAlertRule(true).Draw(t, "enabledRule")

		svc := NewAlertServiceWithRules([]AlertRule{rule})
		// Provide metrics that do NOT contain the rule's metric key.
		alerts := svc.EvaluateAlerts(map[string]float64{"nonexistent_metric": 999})

		assert.Empty(t, alerts,
			"rule %s should not trigger when its metric key %q is missing from metrics map",
			rule.ID, rule.MetricKey)
	})
}

// TestProperty2_EvaluateAlerts_CorrectTriggerSet verifies that EvaluateAlerts
// returns only alerts for rules that should trigger, given a random mix of
// enabled/disabled rules and random metric values.
func TestProperty2_EvaluateAlerts_CorrectTriggerSet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		numRules := rapid.IntRange(1, 10).Draw(t, "numRules")

		rules := make([]AlertRule, numRules)
		metrics := make(map[string]float64)

		for i := 0; i < numRules; i++ {
			enabled := rapid.Bool().Draw(t, "enabled")
			rule := genAlertRule(enabled).Draw(t, "rule")
			// Ensure unique rule IDs by appending index.
			rule.ID = rule.ID + "-" + rapid.StringMatching(`^[0-9]{2}$`).Draw(t, "suffix")
			rules[i] = rule

			// Randomly decide whether to include this rule's metric in the map.
			includeMetric := rapid.Bool().Draw(t, "includeMetric")
			if includeMetric {
				metrics[rule.MetricKey] = genMetricValue().Draw(t, "metricValue")
			}
		}

		svc := NewAlertServiceWithRules(rules)
		alerts := svc.EvaluateAlerts(metrics)

		// Build expected set of rule IDs that should trigger.
		expectedTriggerIDs := make(map[string]bool)
		for _, rule := range rules {
			val, hasMetric := metrics[rule.MetricKey]
			if hasMetric && shouldTrigger(rule, val) {
				expectedTriggerIDs[rule.ID] = true
			}
		}

		// Verify: every triggered alert corresponds to a rule that should trigger.
		triggeredIDs := make(map[string]bool)
		for _, a := range alerts {
			triggeredIDs[a.RuleID] = true
			assert.True(t, expectedTriggerIDs[a.RuleID],
				"alert for rule %s should not have triggered", a.RuleID)
			assert.False(t, a.Acknowledged, "new alerts should not be acknowledged")
			assert.NotEmpty(t, a.ID, "alert ID should not be empty")
			assert.NotZero(t, a.TriggeredAt, "alert TriggeredAt should not be zero")

			// Verify alert metadata matches the rule.
			for _, rule := range rules {
				if rule.ID == a.RuleID {
					assert.Equal(t, rule.Severity, a.Severity)
					assert.Equal(t, rule.MetricKey, a.MetricKey)
					assert.Equal(t, rule.Threshold, a.ThresholdValue)
					assert.Equal(t, metrics[rule.MetricKey], a.CurrentValue)
					break
				}
			}
		}

		// Verify: every rule that should trigger has a corresponding alert.
		for ruleID := range expectedTriggerIDs {
			assert.True(t, triggeredIDs[ruleID],
				"rule %s should have triggered but did not", ruleID)
		}

		// Verify count matches.
		assert.Equal(t, len(expectedTriggerIDs), len(alerts),
			"number of triggered alerts should match expected count")
	})
}

// TestProperty2_PercentChange_ZeroReference_NeverTrigger verifies that
// percent-change conditions with zero reference value never trigger.
func TestProperty2_PercentChange_ZeroReference_NeverTrigger(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		value := genMetricValue().Draw(t, "value")
		threshold := genThreshold().Draw(t, "threshold")

		condGT := AlertCondition{Type: "gt_percent_change", ReferenceValue: 0}
		condLT := AlertCondition{Type: "lt_percent_change", ReferenceValue: 0}

		assert.False(t, EvaluateCondition(condGT, threshold, value),
			"gt_percent_change with zero reference should never trigger")
		assert.False(t, EvaluateCondition(condLT, threshold, value),
			"lt_percent_change with zero reference should never trigger")
	})
}

// TestProperty2_UnknownConditionType_NeverTrigger verifies that unknown
// condition types never trigger.
func TestProperty2_UnknownConditionType_NeverTrigger(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		value := genMetricValue().Draw(t, "value")
		threshold := genThreshold().Draw(t, "threshold")
		unknownType := rapid.SampledFrom([]string{"eq", "ne", "gte", "lte", "", "unknown"}).Draw(t, "unknownType")

		cond := AlertCondition{Type: unknownType}
		assert.False(t, EvaluateCondition(cond, threshold, value),
			"unknown condition type %q should never trigger", unknownType)
	})
}

// TestProperty2_EvaluateCondition_Deterministic verifies that EvaluateCondition
// is deterministic: calling it twice with the same inputs produces the same result.
func TestProperty2_EvaluateCondition_Deterministic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		condType := genConditionType().Draw(t, "conditionType")
		threshold := genThreshold().Draw(t, "threshold")
		value := genMetricValue().Draw(t, "value")
		ref := genPositiveReference().Draw(t, "reference")

		cond := AlertCondition{Type: condType, ReferenceValue: ref}

		result1 := EvaluateCondition(cond, threshold, value)
		result2 := EvaluateCondition(cond, threshold, value)

		assert.Equal(t, result1, result2,
			"EvaluateCondition should be deterministic for cond=%s, threshold=%v, value=%v",
			condType, threshold, value)
	})
}

// TestProperty2_AlertMetadata_Correctness verifies that triggered alerts carry
// correct metadata from their source rules.
func TestProperty2_AlertMetadata_Correctness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rule := genAlertRule(true).Draw(t, "enabledRule")

		// Generate a value that will definitely trigger the rule.
		var value float64
		switch rule.Condition.Type {
		case "gt":
			value = rule.Threshold + math.Abs(rule.Threshold) + 1
		case "lt":
			value = rule.Threshold - math.Abs(rule.Threshold) - 1
		case "gt_percent_change":
			// Need pctChange > threshold, so value = ref * (1 + (threshold+1)/100)
			value = rule.Condition.ReferenceValue * (1 + (math.Abs(rule.Threshold)+1)/100)
		case "lt_percent_change":
			// Need pctChange < -threshold, so value = ref * (1 - (|threshold|+1)/100)
			value = rule.Condition.ReferenceValue * (1 - (math.Abs(rule.Threshold)+1)/100)
		}

		svc := NewAlertServiceWithRules([]AlertRule{rule})
		alerts := svc.EvaluateAlerts(map[string]float64{rule.MetricKey: value})

		require.Len(t, alerts, 1, "exactly one alert should trigger")
		a := alerts[0]

		assert.Equal(t, rule.ID, a.RuleID)
		assert.Equal(t, rule.Severity, a.Severity)
		assert.Equal(t, rule.MetricKey, a.MetricKey)
		assert.Equal(t, rule.Message, a.Message)
		assert.Equal(t, value, a.CurrentValue)
		assert.Equal(t, rule.Threshold, a.ThresholdValue)
		assert.False(t, a.Acknowledged)
		assert.NotEmpty(t, a.ID)
		assert.NotZero(t, a.TriggeredAt)
	})
}
