//go:build windows

package widget

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-desktop-utils/internal/w32"
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
	Columns        int      `json:"columns"`
}

type SystemConfig struct {
	Enabled bool `json:"enabled"`
	GPU     bool `json:"gpu"`
	Disk    bool `json:"disk"`
	Network bool `json:"network"`
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

func DefaultConfig() Config {
	return Config{
		Window: WindowConfig{
			Monitor: 0, Alignment: "topRight", MarginX: 30, MarginY: 30,
			Width: 420, Height: 1200, Opacity: 245,
		},
		Font: FontConfig{
			Family: "Inter", Size: 24, BoldTitle: true,
		},
		Style: StyleConfig{
			TextColor:      "#FFFFFF",
			LabelColor:     "#FFFFFF",
			DimColor:       "#FFFFFF",
			ErrorColor:     "#FF8888",
			PositiveColor:  "#88FF88",
			NegativeColor:  "#FF8888",
			BarBgColor:     "#444444",
			BarNormalColor: "#44AAFF",
			BarWarnColor:   "#FFAA44",
			BarCritColor:   "#FF4444",
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
			GPU:     true,
			Disk:    true,
			Network: true,
		},
	}
}

func ConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func LoadConfig() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if writeErr := writeDefaultConfig(path, &cfg); writeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not write default config: %v\n", writeErr)
			}
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
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
	clamp := func(v *int, min, max int) {
		if *v < min {
			*v = min
		}
		if max > 0 && *v > max {
			*v = max
		}
	}
	clamp(&cfg.Window.Width, 100, 0)
	clamp(&cfg.Window.Height, 100, 0)
	clamp(&cfg.Window.Opacity, 0, 255)
	clamp(&cfg.Font.Size, 8, 72)
	clamp(&cfg.Style.SectionPadding, 0, 0)
	clamp(&cfg.Style.LinePadding, 0, 0)
	clamp(&cfg.Style.HorizontalPad, 0, 0)
	clamp(&cfg.Style.BarHeight, 2, 0)
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
		Text:      parseColor(cfg.Style.TextColor, w32.RGB(255, 255, 255)),
		Label:     parseColor(cfg.Style.LabelColor, w32.RGB(255, 255, 255)),
		Dim:       parseColor(cfg.Style.DimColor, w32.RGB(255, 255, 255)),
		Error:     parseColor(cfg.Style.ErrorColor, w32.RGB(255, 136, 136)),
		Positive:  parseColor(cfg.Style.PositiveColor, w32.RGB(136, 255, 136)),
		Negative:  parseColor(cfg.Style.NegativeColor, w32.RGB(255, 136, 136)),
		BarBg:     parseColor(cfg.Style.BarBgColor, w32.RGB(68, 68, 68)),
		BarNormal: parseColor(cfg.Style.BarNormalColor, w32.RGB(68, 170, 255)),
		BarWarn:   parseColor(cfg.Style.BarWarnColor, w32.RGB(255, 170, 68)),
		BarCrit:   parseColor(cfg.Style.BarCritColor, w32.RGB(255, 68, 68)),
		Separator: parseColor(cfg.Style.SeparatorColor, w32.RGB(136, 136, 136)),
	}
}

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
	return w32.RGB(r, g, b)
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
