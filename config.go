//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Window   WindowConfig   `json:"window"`
	Font     FontConfig     `json:"font"`
	Style    StyleConfig    `json:"style"`
	Exchange ExchangeConfig `json:"exchange"`
	Stocks   StocksConfig   `json:"stocks"`
	System   SystemConfig   `json:"system"`
}

type WindowConfig struct {
	Monitor   int    `json:"monitor"`
	Alignment string `json:"alignment"`
	MarginX   int    `json:"marginX"`
	MarginY   int    `json:"marginY"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Opacity   int    `json:"opacity"`
}

type FontConfig struct {
	Family    string `json:"family"`
	Size      int    `json:"size"`
	BoldTitle bool   `json:"boldTitle"`
}

type StyleConfig struct {
	TextColor      string `json:"textColor"`
	LabelColor     string `json:"labelColor"`
	DimColor       string `json:"dimColor"`
	ErrorColor     string `json:"errorColor"`
	PositiveColor  string `json:"positiveColor"`
	NegativeColor  string `json:"negativeColor"`
	BarBgColor     string `json:"barBgColor"`
	BarNormalColor string `json:"barNormalColor"`
	BarWarnColor   string `json:"barWarnColor"`
	BarCritColor   string `json:"barCritColor"`
	SeparatorColor string `json:"separatorColor"`
	SectionPadding int    `json:"sectionPadding"`
	LinePadding    int    `json:"linePadding"`
	HorizontalPad  int    `json:"horizontalPad"`
	BarHeight      int    `json:"barHeight"`
	TextShadow     bool   `json:"textShadow"`
	ShowSeparator  bool   `json:"showSeparator"`
}

type ExchangeConfig struct {
	Enabled        bool            `json:"enabled"`
	BaseCurrency   string          `json:"baseCurrency"`
	Groups         []CurrencyGroup `json:"groups"`
	RefreshMinutes int             `json:"refreshMinutes"`
}

type CurrencyGroup struct {
	Name    string   `json:"name"`
	Targets []string `json:"targets"`
}

type StocksConfig struct {
	Enabled        bool     `json:"enabled"`
	Provider       string   `json:"provider"`
	APIKey         string   `json:"apiKey"`
	Symbols        []string `json:"symbols"`
	RefreshMinutes int      `json:"refreshMinutes"`
	Columns        int      `json:"columns"` // 1 or 2
}

type SystemConfig struct {
	Enabled bool `json:"enabled"`
}

type ParsedColors struct {
	Text      uint32
	Label     uint32
	Dim       uint32
	Error     uint32
	Positive  uint32
	Negative  uint32
	BarBg     uint32
	BarNormal uint32
	BarWarn   uint32
	BarCrit   uint32
	Separator uint32
}

func defaultConfig() Config {
	return Config{
		Window: WindowConfig{
			Monitor: 0, Alignment: "topRight", MarginX: 30, MarginY: 30,
			Width: 420, Height: 1200, Opacity: 245,
		},
		Font: FontConfig{
			Family: "Segoe UI", Size: 24, BoldTitle: true,
		},
		Style: StyleConfig{
			TextColor:      "#FFFFFF",
			LabelColor:     "#FFFFFF",
			DimColor:       "#FFFFFF",
			ErrorColor:     "#FF8888",
			PositiveColor:  "#88FF88",
			NegativeColor:  "#FF8888",
			BarBgColor:     "#444444",
			BarNormalColor:  "#44AAFF",
			BarWarnColor:    "#FFAA44",
			BarCritColor:    "#FF4444",
			SeparatorColor: "#888888",
			SectionPadding: 16,
			LinePadding:    4,
			HorizontalPad:  16,
			BarHeight:      8,
			TextShadow:     false,
			ShowSeparator:  false,
		},
		Exchange: ExchangeConfig{
			Enabled:      true,
			BaseCurrency: "JPY",
			Groups: []CurrencyGroup{
				{Name: "ASEAN", Targets: []string{"IDR", "THB", "MYR", "PHP", "SGD"}},
				{Name: "Americas", Targets: []string{"USD", "CAD", "MXN", "BRL", "CLP"}},
				{Name: "Europe", Targets: []string{"EUR", "GBP", "CHF", "NOK", "SEK"}},
				{Name: "Asia Pacific", Targets: []string{"CNY", "INR", "KRW", "AUD", "HKD"}},
			},
			RefreshMinutes: 60,
		},
		Stocks: StocksConfig{
			Enabled:        false,
			Provider:       "alphavantage",
			APIKey:         "",
			Symbols:        []string{"AAPL", "GOOGL", "MSFT"},
			RefreshMinutes: 240,
			Columns:        2,
		},
		System: SystemConfig{
			Enabled: true,
		},
	}
}

func configPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func LoadConfig() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			if writeErr := writeDefaultConfig(path, &cfg); writeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not write default config: %v\n", writeErr)
			}
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := defaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	clampConfig(&cfg)
	return &cfg, nil
}

func writeDefaultConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func clampConfig(cfg *Config) {
	if cfg.Window.Width < 100 {
		cfg.Window.Width = 100
	}
	if cfg.Window.Height < 100 {
		cfg.Window.Height = 100
	}
	if cfg.Window.Opacity < 0 {
		cfg.Window.Opacity = 0
	}
	if cfg.Window.Opacity > 255 {
		cfg.Window.Opacity = 255
	}
	if cfg.Font.Size < 8 {
		cfg.Font.Size = 8
	}
	if cfg.Font.Size > 72 {
		cfg.Font.Size = 72
	}
	if cfg.Style.SectionPadding < 0 {
		cfg.Style.SectionPadding = 0
	}
	if cfg.Style.LinePadding < 0 {
		cfg.Style.LinePadding = 0
	}
	if cfg.Style.HorizontalPad < 0 {
		cfg.Style.HorizontalPad = 0
	}
	if cfg.Style.BarHeight < 2 {
		cfg.Style.BarHeight = 2
	}
	if cfg.Exchange.RefreshMinutes < 1 {
		cfg.Exchange.RefreshMinutes = 60
	}
	if cfg.Stocks.RefreshMinutes < 1 {
		cfg.Stocks.RefreshMinutes = 240
	}
	if cfg.Stocks.Columns < 1 || cfg.Stocks.Columns > 4 {
		cfg.Stocks.Columns = 2
	}
}

func (cfg *Config) ParseColors() ParsedColors {
	return ParsedColors{
		Text:      parseColor(cfg.Style.TextColor, rgb(255, 255, 255)),
		Label:     parseColor(cfg.Style.LabelColor, rgb(187, 187, 187)),
		Dim:       parseColor(cfg.Style.DimColor, rgb(119, 119, 119)),
		Error:     parseColor(cfg.Style.ErrorColor, rgb(255, 102, 102)),
		Positive:  parseColor(cfg.Style.PositiveColor, rgb(102, 255, 102)),
		Negative:  parseColor(cfg.Style.NegativeColor, rgb(255, 102, 102)),
		BarBg:     parseColor(cfg.Style.BarBgColor, rgb(51, 51, 51)),
		BarNormal: parseColor(cfg.Style.BarNormalColor, rgb(68, 170, 255)),
		BarWarn:   parseColor(cfg.Style.BarWarnColor, rgb(255, 170, 68)),
		BarCrit:   parseColor(cfg.Style.BarCritColor, rgb(255, 68, 68)),
		Separator: parseColor(cfg.Style.SeparatorColor, rgb(85, 85, 85)),
	}
}

// AllTargets returns a deduplicated list of all target currencies across all groups.
func (cfg *Config) AllTargets() []string {
	seen := make(map[string]bool)
	var targets []string
	for _, g := range cfg.Exchange.Groups {
		for _, t := range g.Targets {
			if !seen[t] && t != cfg.Exchange.BaseCurrency {
				seen[t] = true
				targets = append(targets, t)
			}
		}
	}
	return targets
}

func parseColor(hex string, fallback uint32) uint32 {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return fallback
	}
	return rgb(r, g, b)
}

func parseHexColor(s string) (r, g, b byte, err error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid color: %q", s)
	}
	rv, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	gv, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	bv, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}
	return byte(rv), byte(gv), byte(bv), nil
}
