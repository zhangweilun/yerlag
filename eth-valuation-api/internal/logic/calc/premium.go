package calc

// PremiumDiscountRate computes the premium/discount rate of a market price
// relative to the net asset value (NAV).
// Formula: (marketPrice - nav) / nav * 100
// A positive value indicates a premium; a negative value indicates a discount.
// Returns nil when nav is zero.
func PremiumDiscountRate(nav, marketPrice float64) *float64 {
	if nav == 0 {
		return nil
	}
	result := (marketPrice - nav) / nav * 100
	return &result
}
