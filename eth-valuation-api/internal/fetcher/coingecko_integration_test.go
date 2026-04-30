//go:build integration
// +build integration

package fetcher

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// 这些测试直接调用 CoinGecko 真实 API，用于验证接口是否畅通。
// 运行方式：
//   go test -tags=integration -v -run TestCoinGecko ./internal/fetcher/
//
// 如果有 API Key，可以通过环境变量传入：
//   COINGECKO_API_KEY=xxx go test -tags=integration -v -run TestCoinGecko ./internal/fetcher/

func newTestClient() *CoinGeckoClient {
	apiKey := os.Getenv("COINGECKO_API_KEY")
	return NewCoinGeckoClient("https://api.coingecko.com/api/v3", apiKey)
}

// 每个测试之间等一下，避免触发速率限制
func rateLimitPause() {
	time.Sleep(2 * time.Second)
}

func TestCoinGecko_SimplePrice(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	prices, err := client.GetSimplePrice(ctx, []string{"ethereum", "bitcoin"}, []string{"usd", "btc"})
	if err != nil {
		t.Fatalf("GetSimplePrice 失败: %v", err)
	}

	ethPrices, ok := prices["ethereum"]
	if !ok {
		t.Fatal("返回数据中缺少 ethereum")
	}

	usdPrice, ok := ethPrices["usd"]
	if !ok || usdPrice <= 0 {
		t.Fatalf("ETH USD 价格异常: %v", usdPrice)
	}

	btcPrice, ok := ethPrices["btc"]
	if !ok || btcPrice <= 0 {
		t.Fatalf("ETH BTC 价格异常: %v", btcPrice)
	}

	fmt.Printf("✅ SimplePrice — ETH: $%.2f / ₿%.6f\n", usdPrice, btcPrice)

	if _, ok := prices["bitcoin"]; !ok {
		t.Fatal("返回数据中缺少 bitcoin")
	}
	fmt.Printf("✅ SimplePrice — BTC: $%.2f\n", prices["bitcoin"]["usd"])

	rateLimitPause()
}

func TestCoinGecko_MarketData(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	markets, err := client.GetMarketData(ctx, []string{"ethereum"}, "usd")
	if err != nil {
		t.Fatalf("GetMarketData 失败: %v", err)
	}

	if len(markets) == 0 {
		t.Fatal("GetMarketData 返回空数组")
	}

	eth := markets[0]
	if eth.CurrentPrice <= 0 {
		t.Fatalf("CurrentPrice 异常: %v", eth.CurrentPrice)
	}
	if eth.MarketCap <= 0 {
		t.Fatalf("MarketCap 异常: %v", eth.MarketCap)
	}
	if eth.TotalVolume <= 0 {
		t.Fatalf("TotalVolume 异常: %v", eth.TotalVolume)
	}
	if eth.CirculatingSupply <= 0 {
		t.Fatalf("CirculatingSupply 异常: %v", eth.CirculatingSupply)
	}
	if eth.ATH <= 0 {
		t.Fatalf("ATH 异常: %v", eth.ATH)
	}

	fmt.Printf("✅ MarketData — 价格: $%.2f | 市值: $%.0f | 排名: #%d | 24h: %.2f%% | ATH: $%.2f\n",
		eth.CurrentPrice, eth.MarketCap, eth.MarketCapRank,
		eth.PriceChangePercentage24h, eth.ATH)

	rateLimitPause()
}

func TestCoinGecko_MarketChart(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	chart, err := client.GetMarketChart(ctx, "ethereum", "usd", 7)
	if err != nil {
		t.Fatalf("GetMarketChart 失败: %v", err)
	}

	if len(chart.Prices) == 0 {
		t.Fatal("Prices 数组为空")
	}
	if len(chart.MarketCaps) == 0 {
		t.Fatal("MarketCaps 数组为空")
	}
	if len(chart.TotalVolumes) == 0 {
		t.Fatal("TotalVolumes 数组为空")
	}

	// 验证数据点格式 [timestamp, value]
	first := chart.Prices[0]
	if len(first) < 2 {
		t.Fatalf("价格数据点格式异常: %v", first)
	}
	if first[0] <= 0 || first[1] <= 0 {
		t.Fatalf("价格数据点值异常: timestamp=%v, price=%v", first[0], first[1])
	}

	fmt.Printf("✅ MarketChart (7d) — %d 个价格点 | %d 个市值点 | %d 个交易量点\n",
		len(chart.Prices), len(chart.MarketCaps), len(chart.TotalVolumes))

	rateLimitPause()
}

func TestCoinGecko_OHLCV(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ohlcv, err := client.GetOHLCV(ctx, "ethereum", "usd", 7)
	if err != nil {
		t.Fatalf("GetOHLCV 失败: %v", err)
	}

	if len(ohlcv) == 0 {
		t.Fatal("OHLCV 数组为空")
	}

	first := ohlcv[0]
	if first.Timestamp <= 0 {
		t.Fatalf("Timestamp 异常: %v", first.Timestamp)
	}
	if first.Open <= 0 || first.High <= 0 || first.Low <= 0 || first.Close <= 0 {
		t.Fatalf("OHLC 值异常: O=%v H=%v L=%v C=%v", first.Open, first.High, first.Low, first.Close)
	}
	if first.High < first.Low {
		t.Fatalf("High < Low: %v < %v", first.High, first.Low)
	}

	fmt.Printf("✅ OHLCV (7d) — %d 根 K 线 | 最新: O=%.2f H=%.2f L=%.2f C=%.2f\n",
		len(ohlcv), ohlcv[len(ohlcv)-1].Open, ohlcv[len(ohlcv)-1].High,
		ohlcv[len(ohlcv)-1].Low, ohlcv[len(ohlcv)-1].Close)

	rateLimitPause()
}

func TestCoinGecko_GlobalData(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	global, err := client.GetGlobalData(ctx)
	if err != nil {
		t.Fatalf("GetGlobalData 失败: %v", err)
	}

	ethPct, ok := global.Data.MarketCapPercentage["eth"]
	if !ok {
		t.Fatal("MarketCapPercentage 中缺少 eth")
	}
	if ethPct <= 0 || ethPct >= 100 {
		t.Fatalf("ETH Dominance 异常: %.2f%%", ethPct)
	}

	btcPct, ok := global.Data.MarketCapPercentage["btc"]
	if !ok {
		t.Fatal("MarketCapPercentage 中缺少 btc")
	}

	fmt.Printf("✅ GlobalData — ETH Dominance: %.2f%% | BTC Dominance: %.2f%%\n", ethPct, btcPct)

	rateLimitPause()
}

func TestCoinGecko_CoinData(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	coin, err := client.GetCoinData(ctx, "ethereum")
	if err != nil {
		t.Fatalf("GetCoinData 失败: %v", err)
	}

	if coin.ID != "ethereum" {
		t.Fatalf("ID 不匹配: got %q, want %q", coin.ID, "ethereum")
	}
	if coin.Symbol != "eth" {
		t.Fatalf("Symbol 不匹配: got %q, want %q", coin.Symbol, "eth")
	}

	usdPrice, ok := coin.MarketData.CurrentPrice["usd"]
	if !ok || usdPrice <= 0 {
		t.Fatalf("CurrentPrice[usd] 异常: %v", usdPrice)
	}

	supply := coin.MarketData.CirculatingSupply
	if supply <= 0 {
		t.Fatalf("CirculatingSupply 异常: %v", supply)
	}

	fmt.Printf("✅ CoinData — %s (%s) | 价格: $%.2f | 流通量: %.0f\n",
		coin.ID, coin.Symbol, usdPrice, supply)
}
