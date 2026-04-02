//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// StockPrice holds price data for a single ticker.
type StockPrice struct {
	Price         float64
	Change        float64
	ChangePercent string
	Error         bool
}

// StockFetcher abstracts stock price fetching so providers can be swapped.
type StockFetcher interface {
	Fetch(symbol string) (StockPrice, error)
}

// NewStockFetcher returns a StockFetcher based on config provider name.
func NewStockFetcher(cfg *StocksConfig) StockFetcher {
	switch cfg.Provider {
	case "alphavantage":
		return &AlphaVantageFetcher{APIKey: cfg.APIKey}
	default:
		return &AlphaVantageFetcher{APIKey: cfg.APIKey}
	}
}

// --- Alpha Vantage ---

type AlphaVantageFetcher struct {
	APIKey string
}

type alphaVantageResponse struct {
	GlobalQuote struct {
		Symbol        string `json:"01. symbol"`
		Price         string `json:"05. price"`
		Change        string `json:"09. change"`
		ChangePercent string `json:"10. change percent"`
	} `json:"Global Quote"`
}

func (f *AlphaVantageFetcher) Fetch(symbol string) (StockPrice, error) {
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s",
		symbol, f.APIKey,
	)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return StockPrice{}, err
	}
	defer resp.Body.Close()

	var data alphaVantageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return StockPrice{}, err
	}

	if data.GlobalQuote.Price == "" {
		return StockPrice{}, fmt.Errorf("empty response for %s (rate limit or invalid key)", symbol)
	}

	price, _ := strconv.ParseFloat(data.GlobalQuote.Price, 64)
	change, _ := strconv.ParseFloat(data.GlobalQuote.Change, 64)

	return StockPrice{
		Price:         price,
		Change:        change,
		ChangePercent: strings.TrimSpace(data.GlobalQuote.ChangePercent),
	}, nil
}

// --- Exchange Rate ---

type exchangeRateResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// fetchExchangeRates fetches rates and returns them as "JPY per 1 unit of target".
// The API returns rates relative to baseCurrency (JPY), so we invert them.
func fetchExchangeRates(cfg *Config) (map[string]float64, error) {
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", cfg.Exchange.BaseCurrency)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var data exchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Result != "success" {
		return nil, fmt.Errorf("API result: %s", data.Result)
	}

	// Collect all needed targets from all groups
	targets := cfg.AllTargets()

	result := make(map[string]float64)
	for _, target := range targets {
		if rawRate, ok := data.Rates[target]; ok && rawRate > 0 {
			// API: 1 JPY = rawRate target_currency
			// Invert: 1 target = 1/rawRate JPY
			result[target] = 1.0 / rawRate
		}
	}
	return result, nil
}

// --- Market Data Loop ---

func marketDataLoop(hwnd uintptr, cfg *Config) {
	var exchangeTicker *time.Ticker
	if cfg.Exchange.Enabled {
		exchangeTicker = time.NewTicker(time.Duration(cfg.Exchange.RefreshMinutes) * time.Minute)
		defer exchangeTicker.Stop()
	}

	var stockTicker *time.Ticker
	var stockFetcher StockFetcher

	if cfg.Stocks.Enabled && cfg.Stocks.APIKey != "" && len(cfg.Stocks.Symbols) > 0 {
		stockFetcher = NewStockFetcher(&cfg.Stocks)
		stockTicker = time.NewTicker(time.Duration(cfg.Stocks.RefreshMinutes) * time.Minute)
		defer stockTicker.Stop()
	}

	// Immediate first fetch
	if cfg.Exchange.Enabled {
		fetchExchange(hwnd, cfg)
	}
	if stockFetcher != nil {
		fetchStocks(hwnd, cfg, stockFetcher)
	}

	for {
		switch {
		case exchangeTicker != nil && stockTicker != nil:
			select {
			case <-exchangeTicker.C:
				fetchExchange(hwnd, cfg)
			case <-stockTicker.C:
				fetchStocks(hwnd, cfg, stockFetcher)
			}
		case exchangeTicker != nil:
			<-exchangeTicker.C
			fetchExchange(hwnd, cfg)
		case stockTicker != nil:
			<-stockTicker.C
			fetchStocks(hwnd, cfg, stockFetcher)
		default:
			return
		}
	}
}

func fetchExchange(hwnd uintptr, cfg *Config) {
	rates, err := fetchExchangeRates(cfg)
	appState.mu.Lock()
	if err != nil {
		appState.exchangeErr = true
	} else {
		appState.exchangeRates = rates
		appState.exchangeErr = false
		appState.exchangeTime = time.Now()
	}
	appState.mu.Unlock()
	postRefresh(hwnd)
}

func fetchStocks(hwnd uintptr, cfg *Config, fetcher StockFetcher) {
	for _, symbol := range cfg.Stocks.Symbols {
		sp, err := fetcher.Fetch(symbol)

		appState.mu.Lock()
		if err != nil {
			if existing, ok := appState.stockPrices[symbol]; ok {
				existing.Error = true
				appState.stockPrices[symbol] = existing
			} else {
				appState.stockPrices[symbol] = StockPrice{Error: true}
			}
		} else {
			appState.stockPrices[symbol] = sp
			appState.stockTime = time.Now()
		}
		appState.mu.Unlock()

		if len(cfg.Stocks.Symbols) > 1 {
			time.Sleep(15 * time.Second)
		}
	}
	postRefresh(hwnd)
}
