package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeRatio(t *testing.T) {
	t.Run("normal division", func(t *testing.T) {
		r := SafeRatio(10, 2)
		assert.NotNil(t, r)
		assert.Equal(t, 5.0, *r)
	})

	t.Run("zero denominator returns nil", func(t *testing.T) {
		r := SafeRatio(10, 0)
		assert.Nil(t, r)
	})

	t.Run("zero numerator returns zero", func(t *testing.T) {
		r := SafeRatio(0, 5)
		assert.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})

	t.Run("negative values", func(t *testing.T) {
		r := SafeRatio(-10, 2)
		assert.NotNil(t, r)
		assert.Equal(t, -5.0, *r)
	})
}

func TestNVTRatio(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := NVTRatio(1000000, 50000)
		assert.NotNil(t, r)
		assert.Equal(t, 20.0, *r)
	})

	t.Run("zero daily volume returns nil", func(t *testing.T) {
		r := NVTRatio(1000000, 0)
		assert.Nil(t, r)
	})
}

func TestMVRVRatio(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := MVRVRatio(200, 100)
		assert.NotNil(t, r)
		assert.Equal(t, 2.0, *r)
	})

	t.Run("zero realized value returns nil", func(t *testing.T) {
		r := MVRVRatio(200, 0)
		assert.Nil(t, r)
	})
}

func TestPriceToFeeRatio(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := PriceToFeeRatio(500000, 25000)
		assert.NotNil(t, r)
		assert.Equal(t, 20.0, *r)
	})

	t.Run("zero fee revenue returns nil", func(t *testing.T) {
		r := PriceToFeeRatio(500000, 0)
		assert.Nil(t, r)
	})
}

func TestTVLToMarketCapRatio(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := TVLToMarketCapRatio(50000, 200000)
		assert.NotNil(t, r)
		assert.Equal(t, 0.25, *r)
	})

	t.Run("zero market cap returns nil", func(t *testing.T) {
		r := TVLToMarketCapRatio(50000, 0)
		assert.Nil(t, r)
	})
}

func TestTPSRatio(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := TPSRatio(15, 30)
		assert.NotNil(t, r)
		assert.Equal(t, 0.5, *r)
	})

	t.Run("zero max TPS returns nil", func(t *testing.T) {
		r := TPSRatio(15, 0)
		assert.Nil(t, r)
	})
}

func TestStakingPercentage(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := StakingPercentage(30000000, 120000000)
		assert.NotNil(t, r)
		assert.Equal(t, 25.0, *r)
	})

	t.Run("zero total supply returns nil", func(t *testing.T) {
		r := StakingPercentage(30000000, 0)
		assert.Nil(t, r)
	})
}

func TestETFHoldingsPercentage(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := ETFHoldingsPercentage(1000000, 120000000)
		assert.NotNil(t, r)
		assert.InDelta(t, 0.8333333333333334, *r, 1e-10)
	})

	t.Run("zero circulating supply returns nil", func(t *testing.T) {
		r := ETFHoldingsPercentage(1000000, 0)
		assert.Nil(t, r)
	})
}

func TestETHDominance(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		r := ETHDominance(200000000000, 1000000000000)
		assert.NotNil(t, r)
		assert.Equal(t, 20.0, *r)
	})

	t.Run("zero total market cap returns nil", func(t *testing.T) {
		r := ETHDominance(200000000000, 0)
		assert.Nil(t, r)
	})
}

func TestAnnualizedBurnRate(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		// dailyBurn=1000, totalSupply=120000000
		// (1000 * 365 / 120000000) * 100 = 0.30416666...
		r := AnnualizedBurnRate(1000, 120000000)
		assert.NotNil(t, r)
		assert.InDelta(t, (1000.0*365.0/120000000.0)*100.0, *r, 1e-10)
	})

	t.Run("zero total supply returns nil", func(t *testing.T) {
		r := AnnualizedBurnRate(1000, 0)
		assert.Nil(t, r)
	})

	t.Run("zero daily burn returns zero", func(t *testing.T) {
		r := AnnualizedBurnRate(0, 120000000)
		assert.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})
}

func TestATHDrawdown(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		// ATH=4800, current=2400 → ((4800-2400)/4800)*100 = 50%
		r := ATHDrawdown(4800, 2400)
		assert.NotNil(t, r)
		assert.Equal(t, 50.0, *r)
	})

	t.Run("at ATH returns zero drawdown", func(t *testing.T) {
		r := ATHDrawdown(4800, 4800)
		assert.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})

	t.Run("zero ATH returns nil", func(t *testing.T) {
		r := ATHDrawdown(0, 2400)
		assert.Nil(t, r)
	})

	t.Run("current above ATH gives negative drawdown", func(t *testing.T) {
		r := ATHDrawdown(4800, 5000)
		assert.NotNil(t, r)
		assert.InDelta(t, -4.166666666666667, *r, 1e-10)
	})
}
