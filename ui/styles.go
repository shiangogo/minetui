package ui

import "github.com/charmbracelet/lipgloss"

// 配色取自 Catppuccin Mocha。
var (
	colBase     = lipgloss.Color("#1e1e2e")
	colMantle   = lipgloss.Color("#181825")
	colSurface0 = lipgloss.Color("#313244")
	colSurface1 = lipgloss.Color("#45475a")
	colSurface2 = lipgloss.Color("#585b70")
	colOverlay  = lipgloss.Color("#7f849c")
	colSubtext  = lipgloss.Color("#a6adc8")
	colText     = lipgloss.Color("#cdd6f4")

	colBlue   = lipgloss.Color("#89b4fa")
	colGreen  = lipgloss.Color("#a6e3a1")
	colRed    = lipgloss.Color("#f38ba8")
	colMauve  = lipgloss.Color("#cba6f7")
	colPeach  = lipgloss.Color("#fab387")
	colTeal   = lipgloss.Color("#94e2d5")
	colYellow = lipgloss.Color("#f9e2af")
	colPink   = lipgloss.Color("#f5c2e7")
	colLav    = lipgloss.Color("#b4befe")
)

// 盤面格子
var (
	// 隱藏格以棋盤格交錯兩種底色，看起來像凸起的方塊
	hiddenBg = [2]lipgloss.Color{colSurface1, colSurface2}
	// 已翻開格是較深的底色
	revealedBg = [2]lipgloss.Color{colBase, colMantle}

	numberColors = [9]lipgloss.Color{
		colText, // 0（不會顯示）
		colBlue,
		colGreen,
		colRed,
		colMauve,
		colPeach,
		colTeal,
		colYellow,
		colSubtext,
	}

	// numberGlyphs 使用全形數字 (width=2)，在終端機中可精準置中於偶數欄寬的單元格內
	numberGlyphs = [9]string{
		"", // 0（不會顯示）
		"１",
		"２",
		"３",
		"４",
		"５",
		"６",
		"７",
		"８",
	}

	cursorBg = colPink
	cursorFg = colBase
)

// 版面元件
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colBase).
			Background(colMauve).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().Foreground(colSubtext)

	labelStyle       = lipgloss.NewStyle().Foreground(colOverlay)
	labelActiveStyle = lipgloss.NewStyle().Foreground(colPink).Bold(true)

	chipKeyStyle = lipgloss.NewStyle().
			Foreground(colBase).
			Bold(true).
			Padding(0, 1)
	chipValStyle = lipgloss.NewStyle().
			Foreground(colText).
			Background(colSurface0).
			Padding(0, 1)

	boardBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	modeNormalStyle = lipgloss.NewStyle().
			Bold(true).Foreground(colBase).Background(colBlue).Padding(0, 1)
	modeCommandStyle = lipgloss.NewStyle().
				Bold(true).Foreground(colBase).Background(colPeach).Padding(0, 1)

	infoStyle    = lipgloss.NewStyle().Foreground(colSubtext)
	errorStyle   = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	pendingStyle = lipgloss.NewStyle().Foreground(colYellow)
	posStyle     = lipgloss.NewStyle().Foreground(colLav)
	cmdCursor    = lipgloss.NewStyle().Background(colText).Foreground(colBase)

	helpKeyStyle  = lipgloss.NewStyle().Foreground(colPink).Bold(true)
	helpDescStyle = lipgloss.NewStyle().Foreground(colOverlay)
	helpHeadStyle = lipgloss.NewStyle().Foreground(colMauve).Bold(true)

	menuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colMauve).
			Padding(1, 3)
	menuItemStyle     = lipgloss.NewStyle().Foreground(colSubtext)
	menuSelectedStyle = lipgloss.NewStyle().Foreground(colPink).Bold(true)
	menuDimStyle      = lipgloss.NewStyle().Foreground(colOverlay)
)
