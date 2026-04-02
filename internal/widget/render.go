//go:build windows

package widget

import (
	"fmt"

	"go-desktop-utils/internal/w32"
)

// RenderState is a thread-safe snapshot of AppState.
type RenderState struct {
	CpuPercent    float64
	MemPercent    float64
	MemUsedGB     float64
	MemTotalGB    float64
	ExchangeRates map[string]float64
	ExchangeErr   bool
	ExchangeTime  string
	StockPrices   map[string]StockPrice
	StockErr      bool
	StockTime     string
}

func (a *App) snapshotState() RenderState {
	a.State.Mu.Lock()
	defer a.State.Mu.Unlock()

	rates := make(map[string]float64, len(a.State.ExchangeRates))
	for k, v := range a.State.ExchangeRates {
		rates[k] = v
	}
	stocks := make(map[string]StockPrice, len(a.State.StockPrices))
	for k, v := range a.State.StockPrices {
		stocks[k] = v
	}

	var et, st string
	if !a.State.ExchangeTime.IsZero() {
		et = a.State.ExchangeTime.Format("15:04")
	}
	if !a.State.StockTime.IsZero() {
		st = a.State.StockTime.Format("15:04")
	}

	return RenderState{
		CpuPercent:    a.State.CpuPercent,
		MemPercent:    a.State.MemPercent,
		MemUsedGB:     a.State.MemUsedGB,
		MemTotalGB:    a.State.MemTotalGB,
		ExchangeRates: rates,
		ExchangeErr:   a.State.ExchangeErr,
		ExchangeTime:  et,
		StockPrices:   stocks,
		StockErr:      a.State.StockErr,
		StockTime:     st,
	}
}

func drawShadowedText(app *App, hdc uintptr, text string, rc *w32.RECT, format uint32, color uint32) {
	if app.Config.Style.TextShadow {
		shadowRC := w32.RECT{Left: rc.Left + 1, Top: rc.Top + 1, Right: rc.Right + 1, Bottom: rc.Bottom + 1}
		w32.SetTextColor(hdc, w32.RGB(0, 0, 0))
		w32.DrawText(hdc, text, &shadowRC, format)
	}
	w32.SetTextColor(hdc, color)
	w32.DrawText(hdc, text, rc, format)
}

func (a *App) Render(hdc uintptr) {
	state := a.snapshotState()
	cfg := a.Config
	colors := a.Colors
	style := &cfg.Style
	pad := int32(style.HorizontalPad)
	width := int32(cfg.Window.Width)
	contentWidth := width - pad*2

	bgBrush := w32.CreateSolidBrush(w32.RGB(1, 0, 1))
	bgRect := w32.RECT{Left: 0, Top: 0, Right: width, Bottom: int32(cfg.Window.Height)}
	w32.FillRect(hdc, &bgRect, bgBrush)
	w32.DeleteObject(bgBrush)

	hFont := w32.CreateGDIFont(cfg.Font.Family, cfg.Font.Size, false)
	hFontBold := w32.CreateGDIFont(cfg.Font.Family, cfg.Font.Size, cfg.Font.BoldTitle)
	hFontSmall := w32.CreateGDIFont(cfg.Font.Family, cfg.Font.Size*3/4, false)
	defer w32.DeleteObject(hFont)
	defer w32.DeleteObject(hFontBold)
	defer w32.DeleteObject(hFontSmall)

	w32.SetBkMode(hdc, w32.TRANSPARENT_BK)

	y := int32(style.SectionPadding)
	lineH := int32(cfg.Font.Size) + int32(style.LinePadding)

	drawnSections := 0

	if cfg.Exchange.Enabled {
		for i, group := range cfg.Exchange.Groups {
			if i > 0 {
				y += int32(style.SectionPadding) / 2
			}
			y = a.drawCurrencyGroup(hdc, &state, colors, hFontBold, hFont, group, y, pad, contentWidth, lineH)
		}
		if state.ExchangeTime != "" {
			w32.SelectObject(hdc, hFontSmall)
			timeStr := fmt.Sprintf("Updated %s", state.ExchangeTime)
			rc := w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH*3/4}
			drawShadowedText(a, hdc, timeStr, &rc, w32.DT_RIGHT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Dim)
			y += lineH * 3 / 4
		}
		drawnSections++
	}

	if cfg.Stocks.Enabled {
		if drawnSections > 0 {
			y += int32(style.SectionPadding)
			if style.ShowSeparator {
				y = a.drawSeparator(hdc, colors, y, pad, contentWidth)
			}
			y += int32(style.SectionPadding)
		}
		y = a.drawStockSection(hdc, &state, colors, hFont, hFontBold, hFontSmall, y, pad, contentWidth, lineH)
		drawnSections++
	}

	if cfg.System.Enabled {
		if drawnSections > 0 {
			y += int32(style.SectionPadding)
			if style.ShowSeparator {
				y = a.drawSeparator(hdc, colors, y, pad, contentWidth)
			}
			y += int32(style.SectionPadding)
		}
		a.drawSysInfoSection(hdc, &state, colors, hFont, hFontBold, y, pad, contentWidth, lineH)
	}
}

func (a *App) drawSeparator(hdc uintptr, colors ParsedColors, y, x, width int32) int32 {
	if a.Config.Style.TextShadow {
		sb := w32.CreateSolidBrush(w32.RGB(0, 0, 0))
		sr := w32.RECT{Left: x + 1, Top: y + 1, Right: x + width + 1, Bottom: y + 2}
		w32.FillRect(hdc, &sr, sb)
		w32.DeleteObject(sb)
	}
	brush := w32.CreateSolidBrush(colors.Separator)
	rc := w32.RECT{Left: x, Top: y, Right: x + width, Bottom: y + 1}
	w32.FillRect(hdc, &rc, brush)
	w32.DeleteObject(brush)
	return y + 1
}

func (a *App) drawCurrencyGroup(hdc uintptr, state *RenderState, colors ParsedColors, hFontBold, hFont uintptr, group CurrencyGroup, y, pad, contentWidth, lineH int32) int32 {
	cfg := a.Config
	w32.SelectObject(hdc, hFontBold)
	rc := w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
	drawShadowedText(a, hdc, group.Name, &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Label)
	y += lineH

	if cfg.Style.ShowSeparator {
		y = a.drawSeparator(hdc, colors, y, pad, contentWidth)
	}
	y += int32(cfg.Style.LinePadding)

	w32.SelectObject(hdc, hFont)

	if len(state.ExchangeRates) == 0 {
		rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
		drawShadowedText(a, hdc, "Loading...", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Dim)
		y += lineH
		return y
	}

	for _, target := range group.Targets {
		if target == cfg.Exchange.BaseCurrency {
			continue
		}
		rate, ok := state.ExchangeRates[target]

		rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth/3, Bottom: y + lineH}
		drawShadowedText(a, hdc, target, &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Text)

		if ok && rate > 0 {
			rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
			drawShadowedText(a, hdc, formatRate(rate), &rc, w32.DT_RIGHT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Text)
		}

		if state.ExchangeErr && ok {
			rc = w32.RECT{Left: pad + contentWidth - 20, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
			drawShadowedText(a, hdc, "!", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Error)
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

func (a *App) drawStockSection(hdc uintptr, state *RenderState, colors ParsedColors, hFont, hFontBold, hFontSmall uintptr, y, pad, contentWidth, lineH int32) int32 {
	cfg := a.Config
	w32.SelectObject(hdc, hFontBold)
	rc := w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
	drawShadowedText(a, hdc, "Stocks", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Label)
	y += lineH
	if cfg.Style.ShowSeparator {
		y = a.drawSeparator(hdc, colors, y, pad, contentWidth)
	}
	y += int32(cfg.Style.LinePadding)

	w32.SelectObject(hdc, hFont)

	if cfg.Stocks.APIKey == "" && cfg.Stocks.Provider != "" {
		rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
		drawShadowedText(a, hdc, "Set apiKey in config.json", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Dim)
		y += lineH
		return y
	}

	if len(state.StockPrices) == 0 && len(cfg.Stocks.Symbols) > 0 {
		rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
		drawShadowedText(a, hdc, "Loading...", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Dim)
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

	for i, symbol := range cfg.Stocks.Symbols {
		col := int32(i) % cols
		row := int32(i) / cols
		cx := pad + col*(colWidth+colGap)
		ry := startY + row*lineH

		sp, ok := state.StockPrices[symbol]

		rc = w32.RECT{Left: cx, Top: ry, Right: cx + colWidth, Bottom: ry + lineH}
		drawShadowedText(a, hdc, symbol, &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Text)

		if ok && sp.Price > 0 {
			priceColor := colors.Text
			if sp.ChangePercent != "" {
				if sp.Change > 0 {
					priceColor = colors.Positive
				} else if sp.Change < 0 {
					priceColor = colors.Negative
				}
			}
			rc = w32.RECT{Left: cx, Top: ry, Right: cx + colWidth, Bottom: ry + lineH}
			drawShadowedText(a, hdc, fmt.Sprintf("%.2f", sp.Price), &rc, w32.DT_RIGHT|w32.DT_SINGLELINE|w32.DT_NOCLIP, priceColor)
		} else if ok && sp.Error {
			rc = w32.RECT{Left: cx, Top: ry, Right: cx + colWidth, Bottom: ry + lineH}
			drawShadowedText(a, hdc, "ERR", &rc, w32.DT_RIGHT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Error)
		}
	}

	totalRows := (int32(len(cfg.Stocks.Symbols)) + cols - 1) / cols
	y = startY + totalRows*lineH

	if state.StockTime != "" {
		w32.SelectObject(hdc, hFontSmall)
		rc := w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH*3/4}
		drawShadowedText(a, hdc, fmt.Sprintf("Updated %s", state.StockTime), &rc, w32.DT_RIGHT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Dim)
		y += lineH * 3 / 4
	}

	return y
}

func (a *App) drawSysInfoSection(hdc uintptr, state *RenderState, colors ParsedColors, hFont, hFontBold uintptr, y, pad, contentWidth, lineH int32) int32 {
	cfg := a.Config
	barH := int32(cfg.Style.BarHeight)

	w32.SelectObject(hdc, hFontBold)
	rc := w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
	drawShadowedText(a, hdc, "System", &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Label)
	y += lineH
	if cfg.Style.ShowSeparator {
		y = a.drawSeparator(hdc, colors, y, pad, contentWidth)
	}
	y += int32(cfg.Style.LinePadding)

	w32.SelectObject(hdc, hFont)

	rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
	drawShadowedText(a, hdc, fmt.Sprintf("CPU  %.1f%%", state.CpuPercent), &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Text)
	y += lineH

	drawBarGraph(a, hdc, pad, y, contentWidth, barH, state.CpuPercent, colors)
	y += barH + int32(cfg.Style.LinePadding)*2

	rc = w32.RECT{Left: pad, Top: y, Right: pad + contentWidth, Bottom: y + lineH}
	drawShadowedText(a, hdc, fmt.Sprintf("MEM  %.1f%%  %.1f/%.1f GB", state.MemPercent, state.MemUsedGB, state.MemTotalGB), &rc, w32.DT_LEFT|w32.DT_SINGLELINE|w32.DT_NOCLIP, colors.Text)
	y += lineH

	drawBarGraph(a, hdc, pad, y, contentWidth, barH, state.MemPercent, colors)
	y += barH + int32(cfg.Style.LinePadding)

	return y
}

func drawBarGraph(app *App, hdc uintptr, x, y, width, height int32, percent float64, colors ParsedColors) {
	if app.Config.Style.TextShadow {
		sb := w32.CreateSolidBrush(w32.RGB(0, 0, 0))
		sr := w32.RECT{Left: x + 1, Top: y + 1, Right: x + width + 1, Bottom: y + height + 1}
		w32.FillRect(hdc, &sr, sb)
		w32.DeleteObject(sb)
	}

	bgBrush := w32.CreateSolidBrush(colors.BarBg)
	bgRect := w32.RECT{Left: x, Top: y, Right: x + width, Bottom: y + height}
	w32.FillRect(hdc, &bgRect, bgBrush)
	w32.DeleteObject(bgBrush)

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
		fgBrush := w32.CreateSolidBrush(barColor)
		fgRect := w32.RECT{Left: x, Top: y, Right: x + fillWidth, Bottom: y + height}
		w32.FillRect(hdc, &fgRect, fgBrush)
		w32.DeleteObject(fgBrush)
	}
}
