package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPremiumDiscountRate(t *testing.T) {
	t.Run("premium (market price above NAV)", func(t *testing.T) {
		// NAV=100, marketPrice=110 → (110-100)/100*100 = 10%
		r := PremiumDiscountRate(100, 110)
		assert.NotNil(t, r)
		assert.Equal(t, 10.0, *r)
	})

	t.Run("discount (market price below NAV)", func(t *testing.T) {
		// NAV=100, marketPrice=80 → (80-100)/100*100 = -20%
		r := PremiumDiscountRate(100, 80)
		assert.NotNil(t, r)
		assert.Equal(t, -20.0, *r)
	})

	t.Run("at NAV (no premium or discount)", func(t *testing.T) {
		r := PremiumDiscountRate(100, 100)
		assert.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})

	t.Run("zero NAV returns nil", func(t *testing.T) {
		r := PremiumDiscountRate(0, 110)
		assert.Nil(t, r)
	})

	t.Run("realistic Grayscale scenario", func(t *testing.T) {
		// NAV=$25.50, marketPrice=$22.95 → (22.95-25.50)/25.50*100 = -10%
		r := PremiumDiscountRate(25.50, 22.95)
		assert.NotNil(t, r)
		assert.InDelta(t, -10.0, *r, 1e-10)
	})
}
