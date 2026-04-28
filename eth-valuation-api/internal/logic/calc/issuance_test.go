package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetIssuance(t *testing.T) {
	t.Run("positive net issuance (inflationary)", func(t *testing.T) {
		result := NetIssuance(1000, 400)
		assert.Equal(t, 600.0, result)
	})

	t.Run("negative net issuance (deflationary)", func(t *testing.T) {
		result := NetIssuance(400, 1000)
		assert.Equal(t, -600.0, result)
	})

	t.Run("zero net issuance", func(t *testing.T) {
		result := NetIssuance(500, 500)
		assert.Equal(t, 0.0, result)
	})

	t.Run("zero burn amount", func(t *testing.T) {
		result := NetIssuance(1000, 0)
		assert.Equal(t, 1000.0, result)
	})

	t.Run("zero new issuance", func(t *testing.T) {
		result := NetIssuance(0, 500)
		assert.Equal(t, -500.0, result)
	})
}

func TestIsDeflationary(t *testing.T) {
	t.Run("negative net issuance is deflationary", func(t *testing.T) {
		assert.True(t, IsDeflationary(-100))
	})

	t.Run("positive net issuance is not deflationary", func(t *testing.T) {
		assert.False(t, IsDeflationary(100))
	})

	t.Run("zero net issuance is not deflationary", func(t *testing.T) {
		assert.False(t, IsDeflationary(0))
	})
}

func TestAnnualInflationRate(t *testing.T) {
	t.Run("positive inflation rate", func(t *testing.T) {
		// netIssuance=600, totalSupply=120000000
		// (600 / 120000000) * 100 = 0.0005
		r := AnnualInflationRate(600, 120000000)
		assert.NotNil(t, r)
		assert.InDelta(t, (600.0/120000000.0)*100.0, *r, 1e-10)
	})

	t.Run("negative inflation rate (deflationary)", func(t *testing.T) {
		r := AnnualInflationRate(-600, 120000000)
		assert.NotNil(t, r)
		assert.InDelta(t, (-600.0/120000000.0)*100.0, *r, 1e-10)
	})

	t.Run("zero total supply returns nil", func(t *testing.T) {
		r := AnnualInflationRate(600, 0)
		assert.Nil(t, r)
	})

	t.Run("zero net issuance returns zero rate", func(t *testing.T) {
		r := AnnualInflationRate(0, 120000000)
		assert.NotNil(t, r)
		assert.Equal(t, 0.0, *r)
	})
}
