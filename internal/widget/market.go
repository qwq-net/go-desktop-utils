//go:build windows

package widget

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-desktop-utils/internal/w32"
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

	targets := cfg.AllTargets()
	result := make(map[string]float64)
	for _, target := range targets {
		if rawRate, ok := data.Rates[target]; ok && rawRate > 0 {
			result[target] = 1.0 / rawRate
		}
	}
	return result, nil
}

// --- Market Data Loop ---

func (a *App) MarketDataLoop() {
	cfg := a.Config

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

	if cfg.Exchange.Enabled {
		a.fetchExchange()
	}
	if stockFetcher != nil {
		a.fetchStocks(stockFetcher)
	}

	for {
		switch {
		case exchangeTicker != nil && stockTicker != nil:
			select {
			case <-exchangeTicker.C:
				a.fetchExchange()
			case <-stockTicker.C:
				a.fetchStocks(stockFetcher)
			}
		case exchangeTicker != nil:
			<-exchangeTicker.C
			a.fetchExchange()
		case stockTicker != nil:
			<-stockTicker.C
			a.fetchStocks(stockFetcher)
		default:
			return
		}
	}
}

func (a *App) fetchExchange() {
	rates, err := fetchExchangeRates(a.Config)
	a.State.Mu.Lock()
	if err != nil {
		a.State.ExchangeErr = true
	} else {
		a.State.ExchangeRates = rates
		a.State.ExchangeErr = false
		a.State.ExchangeTime = time.Now()
	}
	a.State.Mu.Unlock()
	w32.PostRefresh(a.Hwnd)
}

func (a *App) fetchStocks(fetcher StockFetcher) {
	cfg := a.Config
	for _, symbol := range cfg.Stocks.Symbols {
		sp, err := fetcher.Fetch(symbol)

		a.State.Mu.Lock()
		if err != nil {
			if existing, ok := a.State.StockPrices[symbol]; ok {
				existing.Error = true
				a.State.StockPrices[symbol] = existing
			} else {
				a.State.StockPrices[symbol] = StockPrice{Error: true}
			}
		} else {
			a.State.StockPrices[symbol] = sp
			a.State.StockTime = time.Now()
		}
		a.State.Mu.Unlock()

		if len(cfg.Stocks.Symbols) > 1 {
			time.Sleep(15 * time.Second)
		}
	}
	w32.PostRefresh(a.Hwnd)
}
