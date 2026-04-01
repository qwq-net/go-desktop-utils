//go:build windows

package main

import (
	"fmt"
	"time"
)

// RenderState is a thread-safe snapshot of AppState.
type RenderState struct {
	CpuPercent    float64
	MemPercent    float64
	MemUsedGB     float64
	MemTotalGB    float64
	ExchangeRates map[string]float64
	ExchangeErr   bool
	ExchangeTime  time.Time
	StockPrices   map[string]StockPrice
	StockErr      bool
	StockTime     time.Time
}

func snapshotState() RenderState {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	rates := make(map[string]float64, len(appState.exchangeRates))
	for k, v := range appState.exchangeRates {
		rates[k] = v
	}
	stocks := make(map[string]StockPrice, len(appState.stockPrices))
	for k, v := range appState.stockPrices {
		stocks[k] = v
	}

	return RenderState{
		CpuPercent:    appState.cpuPercent,
		MemPercent:    appState.memPercent,
		MemUsedGB:     appState.memUsedGB,
		MemTotalGB:    appState.memTotalGB,
		ExchangeRates: rates,
		ExchangeErr:   appState.exchangeErr,
		ExchangeTime:  appState.exchangeTime,
		StockPrices:   stocks,
		StockErr:      appState.stockErr,
		StockTime:     appState.stockTime,
	}
}

func drawShadowedText(hdc uintptr, text string, rc *RECT, format uint32, color uint32) {
	if appConfig.Style.TextShadow {
		shadowRC := RECT{rc.Left + 1, rc.Top + 1, rc.Right + 1, rc.Bottom + 1}
		setTextColor(hdc, rgb(0, 0, 0))
		drawText(hdc, text, &shadowRC, format)
	}
	setTextColor(hdc, color)
	drawText(hdc, text, rc, format)
}

func render(hdc uintptr) {
	state := snapshotState()
	cfg := appConfig
	colors := appColors
	style := &cfg.Style
	pad := int32(style.HorizontalPad)
	width := int32(cfg.Window.Width)
	contentWidth := width - pad*2

	// Near-black background — color key eliminates this; anti-aliasing fringes blend dark
	bgBrush := createSolidBrush(rgb(1, 0, 1))
	bgRect := RECT{0, 0, width, int32(cfg.Window.Height)}
	fillRect(hdc, &bgRect, bgBrush)
	deleteObject(bgBrush)

	hFont := createGDIFont(cfg.Font.Family, cfg.Font.Size, false)
	hFontBold := createGDIFont(cfg.Font.Family, cfg.Font.Size, cfg.Font.BoldTitle)
	hFontSmall := createGDIFont(cfg.Font.Family, cfg.Font.Size*3/4, false)
	defer deleteObject(hFont)
	defer deleteObject(hFontBold)
	defer deleteObject(hFontSmall)

	setBkMode(hdc, TRANSPARENT_BK)

	y := int32(style.SectionPadding)
	lineH := int32(cfg.Font.Size) + int32(style.LinePadding)

	// Exchange Rate Groups
	for i, group := range cfg.Exchange.Groups {
		if i > 0 {
			y += int32(style.SectionPadding) / 2
		}
		y = drawCurrencyGroup(hdc, &state, cfg, colors, hFontBold, hFont, group, y, pad, contentWidth, lineH)
	}

	// Update time for exchange
	if !state.ExchangeTime.IsZero() {
		selectObject(hdc, hFontSmall)
		timeStr := fmt.Sprintf("Updated %s", state.ExchangeTime.Format("15:04"))
		rc := RECT{pad, y, pad + contentWidth, y + lineH*3/4}
		drawShadowedText(hdc, timeStr, &rc, DT_RIGHT|DT_SINGLELINE|DT_NOCLIP, colors.Dim)
		y += lineH * 3 / 4
	}

	// Separator
	y += int32(style.SectionPadding)
	y = drawSeparator(hdc, colors, y, pad, contentWidth)
	y += int32(style.SectionPadding)

	// Stocks (2-column)
	y = drawStockSection(hdc, &state, cfg, colors, hFont, hFontBold, hFontSmall, y, pad, contentWidth, lineH)

	// Separator
	y += int32(style.SectionPadding)
	y = drawSeparator(hdc, colors, y, pad, contentWidth)
	y += int32(style.SectionPadding)

	// System Info
	drawSysInfoSection(hdc, &state, cfg, colors, hFont, hFontBold, y, pad, contentWidth, lineH)
}

func drawSeparator(hdc uintptr, colors ParsedColors, y, x, width int32) int32 {
	if appConfig.Style.TextShadow {
		shadowBrush := createSolidBrush(rgb(0, 0, 0))
		shadowRC := RECT{x + 1, y + 1, x + width + 1, y + 2}
		fillRect(hdc, &shadowRC, shadowBrush)
		deleteObject(shadowBrush)
	}
	brush := createSolidBrush(colors.Separator)
	rc := RECT{x, y, x + width, y + 1}
	fillRect(hdc, &rc, brush)
	deleteObject(brush)
	return y + 1
}

func drawCurrencyGroup(hdc uintptr, state *RenderState, cfg *Config, colors ParsedColors, hFontBold, hFont uintptr, group CurrencyGroup, y, pad, contentWidth, lineH int32) int32 {
	selectObject(hdc, hFontBold)
	rc := RECT{pad, y, pad + contentWidth, y + lineH}
	drawShadowedText(hdc, group.Name, &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Label)
	y += lineH

	y = drawSeparator(hdc, colors, y, pad, contentWidth)
	y += int32(cfg.Style.LinePadding)

	selectObject(hdc, hFont)

	if len(state.ExchangeRates) == 0 {
		rc = RECT{pad, y, pad + contentWidth, y + lineH}
		drawShadowedText(hdc, "Loading...", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Dim)
		y += lineH
		return y
	}

	for _, target := range group.Targets {
		if target == cfg.Exchange.BaseCurrency {
			continue
		}
		rate, ok := state.ExchangeRates[target]

		rc = RECT{pad, y, pad + contentWidth/3, y + lineH}
		drawShadowedText(hdc, target, &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Text)

		if ok && rate > 0 {
			valText := formatRate(rate)
			rc = RECT{pad, y, pad + contentWidth, y + lineH}
			drawShadowedText(hdc, valText, &rc, DT_RIGHT|DT_SINGLELINE|DT_NOCLIP, colors.Text)
		}

		if state.ExchangeErr && ok {
			rc = RECT{pad + contentWidth - 20, y, pad + contentWidth, y + lineH}
			drawShadowedText(hdc, "!", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Error)
		}

		y += lineH
	}

	return y
}

func formatRate(rate float64) string {
	switch {
	case rate >= 100:
		return fmt.Sprintf("%.2f", rate)
	case rate >= 1:
		return fmt.Sprintf("%.3f", rate)
	default:
		return fmt.Sprintf("%.4f", rate)
	}
}

func drawStockSection(hdc uintptr, state *RenderState, cfg *Config, colors ParsedColors, hFont, hFontBold, hFontSmall uintptr, y, pad, contentWidth, lineH int32) int32 {
	selectObject(hdc, hFontBold)
	rc := RECT{pad, y, pad + contentWidth, y + lineH}
	drawShadowedText(hdc, "Stocks", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Label)
	y += lineH
	y = drawSeparator(hdc, colors, y, pad, contentWidth)
	y += int32(cfg.Style.LinePadding)

	selectObject(hdc, hFont)

	if cfg.Stocks.APIKey == "" && cfg.Stocks.Provider != "" {
		rc = RECT{pad, y, pad + contentWidth, y + lineH}
		drawShadowedText(hdc, "Set apiKey in config.json", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Dim)
		y += lineH
		return y
	}

	if len(state.StockPrices) == 0 && len(cfg.Stocks.Symbols) > 0 {
		rc = RECT{pad, y, pad + contentWidth, y + lineH}
		drawShadowedText(hdc, "Loading...", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Dim)
		y += lineH
		return y
	}

	cols := int32(cfg.Stocks.Columns)
	if cols < 1 {
		cols = 2
	}
	colGap := int32(12)
	colWidth := (contentWidth - colGap*(cols-1)) / cols

	startY := y
	symbols := cfg.Stocks.Symbols

	for i, symbol := range symbols {
		col := int32(i) % cols
		row := int32(i) / cols
		cx := pad + col*(colWidth+colGap)
		ry := startY + row*lineH

		sp, ok := state.StockPrices[symbol]

		// Symbol (left-aligned in column)
		rc = RECT{cx, ry, cx + colWidth, ry + lineH}
		drawShadowedText(hdc, symbol, &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Text)

		if ok && sp.Price > 0 {
			// Price (right-aligned in column), colored by change direction
			priceColor := colors.Text
			if sp.ChangePercent != "" {
				if sp.Change > 0 {
					priceColor = colors.Positive
				} else if sp.Change < 0 {
					priceColor = colors.Negative
				}
			}
			priceText := fmt.Sprintf("%.2f", sp.Price)
			rc = RECT{cx, ry, cx + colWidth, ry + lineH}
			drawShadowedText(hdc, priceText, &rc, DT_RIGHT|DT_SINGLELINE|DT_NOCLIP, priceColor)

			if sp.Error {
				rc = RECT{cx + colWidth - 16, ry, cx + colWidth, ry + lineH}
				drawShadowedText(hdc, "!", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Error)
			}
		} else if ok && sp.Error {
			rc = RECT{cx, ry, cx + colWidth, ry + lineH}
			drawShadowedText(hdc, "ERR", &rc, DT_RIGHT|DT_SINGLELINE|DT_NOCLIP, colors.Error)
		}
	}

	totalRows := (int32(len(symbols)) + cols - 1) / cols
	y = startY + totalRows*lineH

	// Update time
	if !state.StockTime.IsZero() {
		selectObject(hdc, hFontSmall)
		timeStr := fmt.Sprintf("Updated %s", state.StockTime.Format("15:04"))
		rc := RECT{pad, y, pad + contentWidth, y + lineH*3/4}
		drawShadowedText(hdc, timeStr, &rc, DT_RIGHT|DT_SINGLELINE|DT_NOCLIP, colors.Dim)
		y += lineH * 3 / 4
	}

	return y
}

func drawSysInfoSection(hdc uintptr, state *RenderState, cfg *Config, colors ParsedColors, hFont, hFontBold uintptr, y, pad, contentWidth, lineH int32) int32 {
	barH := int32(cfg.Style.BarHeight)

	selectObject(hdc, hFontBold)
	rc := RECT{pad, y, pad + contentWidth, y + lineH}
	drawShadowedText(hdc, "System", &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Label)
	y += lineH
	y = drawSeparator(hdc, colors, y, pad, contentWidth)
	y += int32(cfg.Style.LinePadding)

	selectObject(hdc, hFont)

	// CPU
	cpuLabel := fmt.Sprintf("CPU  %.1f%%", state.CpuPercent)
	rc = RECT{pad, y, pad + contentWidth, y + lineH}
	drawShadowedText(hdc, cpuLabel, &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Text)
	y += lineH

	drawBarGraph(hdc, pad, y, contentWidth, barH, state.CpuPercent, colors)
	y += barH + int32(cfg.Style.LinePadding)*2

	// Memory
	memLabel := fmt.Sprintf("MEM  %.1f%%  %.1f/%.1f GB", state.MemPercent, state.MemUsedGB, state.MemTotalGB)
	rc = RECT{pad, y, pad + contentWidth, y + lineH}
	drawShadowedText(hdc, memLabel, &rc, DT_LEFT|DT_SINGLELINE|DT_NOCLIP, colors.Text)
	y += lineH

	drawBarGraph(hdc, pad, y, contentWidth, barH, state.MemPercent, colors)
	y += barH + int32(cfg.Style.LinePadding)

	return y
}

func drawBarGraph(hdc uintptr, x, y, width, height int32, percent float64, colors ParsedColors) {
	if appConfig.Style.TextShadow {
		shadowBrush := createSolidBrush(rgb(0, 0, 0))
		shadowRC := RECT{x + 1, y + 1, x + width + 1, y + height + 1}
		fillRect(hdc, &shadowRC, shadowBrush)
		deleteObject(shadowBrush)
	}

	bgBrush := createSolidBrush(colors.BarBg)
	bgRect := RECT{x, y, x + width, y + height}
	fillRect(hdc, &bgRect, bgBrush)
	deleteObject(bgBrush)

	fillWidth := int32(float64(width) * percent / 100.0)
	if fillWidth < 1 && percent > 0 {
		fillWidth = 1
	}
	if fillWidth > width {
		fillWidth = width
	}

	var barColor uint32
	switch {
	case percent >= 95:
		barColor = colors.BarCrit
	case percent >= 80:
		barColor = colors.BarWarn
	default:
		barColor = colors.BarNormal
	}

	if fillWidth > 0 {
		fgBrush := createSolidBrush(barColor)
		fgRect := RECT{x, y, x + fillWidth, y + height}
		fillRect(hdc, &fgRect, fgBrush)
		deleteObject(fgBrush)
	}
}
