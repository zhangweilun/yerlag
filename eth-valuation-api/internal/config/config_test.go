package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestConfigLoad(t *testing.T) {
	const configContent = `
Name: valuation-api
Host: 0.0.0.0
Port: 8888
Database:
  Driver: mysql
  DataSource: "root:password@tcp(127.0.0.1:3306)/eth_valuation?charset=utf8mb4&parseTime=True&loc=Local"
Redis:
  Host: 127.0.0.1:6379
  Pass: ""
  DB: 0
DataSources:
  Etherscan:
    APIKey: "test-key"
    BaseURL: "https://api.etherscan.io/api"
  CoinGecko:
    APIKey: ""
    BaseURL: "https://api.coingecko.com/api/v3"
  DefiLlama:
    BaseURL: "https://api.llama.fi"
  Glassnode:
    APIKey: ""
    BaseURL: "https://api.glassnode.com/v1"
  BeaconChain:
    BaseURL: "https://beaconcha.in/api/v1"
  TradFi:
    APIKey: ""
    BaseURL: "https://api.tradfi-data.example.com"
CacheTTL:
  Price: 10
  GasData: 300
  OnChainMetrics: 300
  TVLData: 300
  InstitutionalData: 3600
  NetworkData: 300
  MacroData: 3600
  ValuationScore: 300
`

	var c Config
	err := conf.LoadFromYamlBytes([]byte(configContent), &c)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify basic server config
	if c.Name != "valuation-api" {
		t.Errorf("expected Name=valuation-api, got %s", c.Name)
	}
	if c.Host != "0.0.0.0" {
		t.Errorf("expected Host=0.0.0.0, got %s", c.Host)
	}
	if c.Port != 8888 {
		t.Errorf("expected Port=8888, got %d", c.Port)
	}

	// Verify database config
	if c.Database.Driver != "mysql" {
		t.Errorf("expected Database.Driver=mysql, got %s", c.Database.Driver)
	}
	if c.Database.DataSource == "" {
		t.Error("expected non-empty Database.DataSource")
	}

	// Verify Redis config
	if c.Redis.Host != "127.0.0.1:6379" {
		t.Errorf("expected Redis.Host=127.0.0.1:6379, got %s", c.Redis.Host)
	}

	// Verify data source configs
	if c.DataSources.Etherscan.APIKey != "test-key" {
		t.Errorf("expected Etherscan.APIKey=test-key, got %s", c.DataSources.Etherscan.APIKey)
	}
	if c.DataSources.CoinGecko.BaseURL != "https://api.coingecko.com/api/v3" {
		t.Errorf("expected CoinGecko.BaseURL, got %s", c.DataSources.CoinGecko.BaseURL)
	}
	if c.DataSources.DefiLlama.BaseURL != "https://api.llama.fi" {
		t.Errorf("expected DefiLlama.BaseURL, got %s", c.DataSources.DefiLlama.BaseURL)
	}
	if c.DataSources.Glassnode.BaseURL != "https://api.glassnode.com/v1" {
		t.Errorf("expected Glassnode.BaseURL, got %s", c.DataSources.Glassnode.BaseURL)
	}
	if c.DataSources.BeaconChain.BaseURL != "https://beaconcha.in/api/v1" {
		t.Errorf("expected BeaconChain.BaseURL, got %s", c.DataSources.BeaconChain.BaseURL)
	}
	if c.DataSources.TradFi.BaseURL != "https://api.tradfi-data.example.com" {
		t.Errorf("expected TradFi.BaseURL, got %s", c.DataSources.TradFi.BaseURL)
	}

	// Verify cache TTL config
	if c.CacheTTL.Price != 10 {
		t.Errorf("expected CacheTTL.Price=10, got %d", c.CacheTTL.Price)
	}
	if c.CacheTTL.GasData != 300 {
		t.Errorf("expected CacheTTL.GasData=300, got %d", c.CacheTTL.GasData)
	}
	if c.CacheTTL.OnChainMetrics != 300 {
		t.Errorf("expected CacheTTL.OnChainMetrics=300, got %d", c.CacheTTL.OnChainMetrics)
	}
	if c.CacheTTL.TVLData != 300 {
		t.Errorf("expected CacheTTL.TVLData=300, got %d", c.CacheTTL.TVLData)
	}
	if c.CacheTTL.InstitutionalData != 3600 {
		t.Errorf("expected CacheTTL.InstitutionalData=3600, got %d", c.CacheTTL.InstitutionalData)
	}
	if c.CacheTTL.NetworkData != 300 {
		t.Errorf("expected CacheTTL.NetworkData=300, got %d", c.CacheTTL.NetworkData)
	}
	if c.CacheTTL.MacroData != 3600 {
		t.Errorf("expected CacheTTL.MacroData=3600, got %d", c.CacheTTL.MacroData)
	}
	if c.CacheTTL.ValuationScore != 300 {
		t.Errorf("expected CacheTTL.ValuationScore=300, got %d", c.CacheTTL.ValuationScore)
	}
}
