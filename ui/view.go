package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/andrewfei/minesweeper/game"
)

const (
	rowLabelWidth = 3  // 列號欄寬，如 "12 "
	minLayoutW    = 60 // 狀態列等元件的最小寬度
	chromeHeight  = 12 // 盤面以外的版面高度（標題、統計、進度條、狀態列…）
)

func (m Model) viewGame() string {
	if m.board == nil {
		return ""
	}
	if warn, ok := m.sizeWarning(); ok {
		return m.place(warn)
	}

	cw := m.cellWidth()
	var inner string
	if m.showHelp {
		inner = m.renderHelp()
	} else {
		inner = m.renderBoard(cw)
	}
	box := boardBoxStyle.BorderForeground(m.accentColor()).Render(inner)

	w := max(lipgloss.Width(box), minLayoutW)
	box = lipgloss.PlaceHorizontal(w, lipgloss.Center, box)

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(w),
		"",
		m.renderStats(w),
		box,
		m.renderProgress(w),
		"",
		m.renderStatusBar(w),
		m.renderCmdline(w),
		m.renderHints(w),
	)
	return m.place(content)
}

// cellWidth 依終端機寬度決定每格要用 3 欄（寬鬆）或 2 欄（緊湊）。
func (m Model) cellWidth() int {
	if m.width == 0 || rowLabelWidth+m.board.Width()*3+4 <= m.width {
		return 3
	}
	return 2
}

func (m Model) sizeWarning() (string, bool) {
	if m.width == 0 || m.height == 0 {
		return "", false
	}
	needW := max(rowLabelWidth+m.board.Width()*2+4, minLayoutW)
	needH := m.board.Height() + 3 + chromeHeight
	if m.width >= needW && m.height >= needH {
		return "", false
	}
	msg := lipgloss.JoinVertical(lipgloss.Center,
		errorStyle.Render("終端機視窗太小了"),
		"",
		infoStyle.Render(fmt.Sprintf("目前 %d×%d，這個難度至少需要 %d×%d", m.width, m.height, needW, needH)),
		infoStyle.Render("請放大視窗，或按 d 選擇較小的難度"),
	)
	return menuBoxStyle.BorderForeground(colRed).Render(msg), true
}

func (m Model) accentColor() lipgloss.Color {
	switch m.board.Status() {
	case game.Won:
		return colGreen
	case game.Lost:
		return colRed
	}
	return colMauve
}

func (m Model) renderHeader(w int) string {
	left := titleStyle.Background(m.accentColor()).Render("✹ MINESWEEPER")
	right := subtitleStyle.Render(m.diff.Summary())
	return spread(w, left, right)
}

func (m Model) renderStats(w int) string {
	b := m.board

	mines := chip("⚑ 地雷", fmt.Sprintf("%3d", b.RemainingMines()), colRed)
	timer := chip("時間", formatDuration(m.elapsedTime()), colBlue)

	var state string
	switch b.Status() {
	case game.Ready:
		state = chipKeyStyle.Background(colOverlay).Render("準備")
	case game.Playing:
		state = chipKeyStyle.Background(colYellow).Render("進行中")
	case game.Won:
		state = chipKeyStyle.Background(colGreen).Render("勝利")
	case game.Lost:
		state = chipKeyStyle.Background(colRed).Render("失敗")
	}
	return spread(w, mines+"  "+timer, state)
}

func chip(key, val string, c lipgloss.Color) string {
	return chipKeyStyle.Background(c).Render(key) + chipValStyle.Render(val)
}

func (m Model) renderProgress(w int) string {
	b := m.board
	label := spread(w,
		labelStyle.Render("進度"),
		labelStyle.Render(fmt.Sprintf("%d / %d 安全格", b.RevealedCount(), b.SafeTotal())),
	)
	p := m.progress
	p.Width = w
	return label + "\n" + p.View()
}

// renderBoard 繪製含座標標籤的盤面。
func (m Model) renderBoard(cw int) string {
	b := m.board
	var sb strings.Builder

	sb.WriteString(strings.Repeat(" ", rowLabelWidth))
	for c := 0; c < b.Width(); c++ {
		active := c == m.cur.Col
		style := labelStyle
		if active {
			style = labelActiveStyle
		}
		sb.WriteString(style.Render(colLabel(c+1, cw, active)))
	}

	for r := 0; r < b.Height(); r++ {
		sb.WriteByte('\n')
		style := labelStyle
		if r == m.cur.Row {
			style = labelActiveStyle
		}
		sb.WriteString(style.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < b.Width(); c++ {
			sb.WriteString(m.renderCell(r, c, cw))
		}
	}
	return sb.String()
}

func colLabel(n, cw int, active bool) string {
	if cw >= 3 {
		return fmt.Sprintf("%2d ", n)
	}
	// 緊湊模式下兩位數會黏在一起，只標示 1、5 的倍數與游標所在行
	if n == 1 || n%5 == 0 || active {
		return fmt.Sprintf("%2d", n)
	}
	return "  "
}

func (m Model) renderCell(r, c, cw int) string {
	b := m.board
	cell := b.Cell(r, c)
	parity := (r + c) % 2
	status := b.Status()

	var glyph string
	st := lipgloss.NewStyle()

	switch {
	case cell.State == game.Revealed && cell.Mine: // 引爆的那顆
		glyph = "✹"
		st = st.Background(colRed).Foreground(colBase).Bold(true)

	case cell.State == game.Revealed:
		st = st.Background(revealedBg[parity])
		if cell.Adjacent > 0 {
			glyph = fmt.Sprint(cell.Adjacent)
			st = st.Foreground(numberColors[cell.Adjacent]).Bold(true)
		}

	case cell.State == game.Flagged:
		switch {
		case status == game.Lost && !cell.Mine: // 插錯的旗
			glyph = "✗"
			st = st.Background(revealedBg[parity]).Foreground(colPeach).Bold(true)
		case status == game.Won:
			glyph = "⚑"
			st = st.Background(hiddenBg[parity]).Foreground(colGreen).Bold(true)
		default:
			glyph = "⚑"
			st = st.Background(hiddenBg[parity]).Foreground(colRed).Bold(true)
		}

	default: // 隱藏格
		if status == game.Lost && cell.Mine {
			glyph = "✱"
			st = st.Background(revealedBg[parity]).Foreground(colText)
		} else {
			st = st.Background(hiddenBg[parity])
		}
	}

	if r == m.cur.Row && c == m.cur.Col {
		st = st.Background(cursorBg).Foreground(cursorFg).Bold(true)
	}
	return st.Render(padCell(glyph, cw))
}

func padCell(glyph string, cw int) string {
	if glyph == "" {
		glyph = " "
	}
	if cw >= 3 {
		return " " + glyph + " "
	}
	return " " + glyph
}

func (m Model) renderStatusBar(w int) string {
	bar := lipgloss.NewStyle().Background(colSurface0)

	var mode string
	if m.mode == modeCommand {
		mode = modeCommandStyle.Render("COMMAND")
	} else {
		mode = modeNormalStyle.Render("NORMAL")
	}
	left := mode + bar.Foreground(colSubtext).Render(" "+m.diff.Name)

	pending := m.count
	if m.pendingG {
		pending += "g"
	}
	right := ""
	if pending != "" {
		right += bar.Inherit(pendingStyle).Render(pending + "  ")
	}
	right += bar.Inherit(posStyle).Render(fmt.Sprintf("列 %02d  行 %02d ", m.cur.Row+1, m.cur.Col+1))

	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + bar.Render(strings.Repeat(" ", gap)) + right
}

func (m Model) renderCmdline(w int) string {
	if m.mode == modeCommand {
		return infoStyle.Render(":"+m.cmdBuf) + cmdCursor.Render(" ")
	}
	if m.msg == "" {
		return ""
	}
	style := infoStyle
	switch m.msgStyle {
	case msgError:
		style = errorStyle
	case msgSuccess:
		style = successStyle
	}
	return style.Width(w).Render(m.msg)
}

func (m Model) renderHints(w int) string {
	var pairs [][2]string
	switch {
	case m.showHelp:
		pairs = [][2]string{{"?", "關閉說明"}, {"q", "離開"}}
	case m.board.Over():
		pairs = [][2]string{{"r", "再來一局"}, {"d", "換難度"}, {"?", "說明"}, {"q", "離開"}}
	default:
		pairs = [][2]string{{"hjkl", "移動"}, {"space", "翻開"}, {"f", "插旗"}, {":", "跳轉"}, {"?", "說明"}}
	}
	return lipgloss.NewStyle().Width(w).Render(renderHints(pairs))
}

type helpSection struct {
	title string
	rows  [][2]string
}

var helpSections = []helpSection{
	{"移動", [][2]string{
		{"h j k l / 方向鍵", "左 下 上 右（可加數字前綴，如 5j）"},
		{"gg / G", "第一列 / 最後一列（5G 跳到第 5 列）"},
		{"0 / $", "列首 / 列尾"},
		{"H / M / L", "頂端 / 盤面中央 / 底端"},
		{"w / b", "下一個 / 上一個未翻開的格子"},
	}},
	{"操作", [][2]string{
		{"space / enter", "翻開；在數字格上即大面積翻開"},
		{"c", "大面積翻開（周圍旗數 = 數字時）"},
		{"f / m", "插旗 / 拔旗"},
	}},
	{"指令", [][2]string{
		{":12 5", "跳到第 12 列、第 5 行"},
		{":12", "跳到第 12 列"},
		{":12 5 f", "跳過去並插旗（o 翻開、c 大面積翻開）"},
		{":easy :medium :hard", "切換難度"},
		{":custom W H M", "自訂寬、高與地雷數"},
		{":r  :menu  :q", "重開 / 難度選單 / 離開"},
	}},
	{"其他", [][2]string{
		{"r / d", "重新開始 / 難度選單"},
		{"? / q", "說明 / 離開"},
	}},
}

func (m Model) renderHelp() string {
	keyCol := lipgloss.NewStyle().Width(22)
	var sb strings.Builder
	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(helpHeadStyle.Render(sec.title))
		for _, row := range sec.rows {
			sb.WriteString("\n  ")
			sb.WriteString(keyCol.Render(helpKeyStyle.Render(row[0])))
			sb.WriteString(helpDescStyle.Render(row[1]))
		}
	}
	return sb.String()
}

// spread 把 left 靠左、right 靠右排在寬度 w 的一行內。
func spread(w int, left, right string) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func formatDuration(d time.Duration) string {
	d = d.Truncate(100 * time.Millisecond)
	h := int(d / time.Hour)
	mm := int(d/time.Minute) % 60
	ss := int(d/time.Second) % 60
	tenth := int(d/(100*time.Millisecond)) % 10
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%d", h, mm, ss, tenth)
	}
	return fmt.Sprintf("%02d:%02d.%d", mm, ss, tenth)
}
