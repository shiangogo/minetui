package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/shiangogo/minetui/game"
)

const (
	rowLabelWidth = 3  // 列號欄寬，如 "12 "
	minLayoutW    = 60 // 狀態列等元件的最小寬度
	chromeHeight  = 13 // 盤面以外的版面高度（標題、統計、外框、進度條、狀態列、提示等）
	sidePanelW    = 34 // 寬螢幕側欄 (32) 加上間距 (2) 所需寬度
)

func (m Model) viewGame() string {
	if m.board == nil {
		return ""
	}
	if warn, ok := m.sizeWarning(); ok {
		return m.place(warn)
	}

	cw, ch := m.cellSize()
	bh := m.board.Height()

	// 當啟用多行高單元格（ch >= 3，確保數字垂直正中），但視窗高度不足以同時堆疊完整上下資訊列時，
	// 自動切換為寬螢幕「左側資訊欄 + 右側緊湊棋盤」或嵌入式上下邊框版面，讓 50 行高的全螢幕也能容納中級/高級正方形大格子。
	if !m.showHelp && m.height > 0 && m.height < bh*ch+chromeHeight && ch >= 3 {
		box := m.renderBoardBoxCompact(cw, ch, false)
		boxW := lipgloss.Width(box)
		if m.width >= boxW+sidePanelW {
			side := m.renderSidePanel(sidePanelW - 2)
			return m.place(lipgloss.JoinHorizontal(lipgloss.Center, side, "  ", box))
		}
		if m.height >= bh*ch+4 {
			w := max(boxW, minLayoutW)
			box = lipgloss.PlaceHorizontal(w, lipgloss.Center, box)
			top := m.renderStats(w)
			bottom := m.renderStatusBar(w)
			if m.mode == modeCommand || m.msg != "" {
				bottom = m.renderCmdline(w)
			}
			return m.place(lipgloss.JoinVertical(lipgloss.Left, top, box, bottom))
		}
		return m.place(m.renderBoardBoxCompact(cw, ch, true))
	}

	var inner string
	if m.showHelp {
		inner = m.renderHelp()
	} else {
		inner = m.renderBoard(cw, ch)
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

// cellSize 依終端機視窗尺寸與棋盤大小，自適應計算單元格寬度 (cw) 與高度 (ch)。
// 終端機字元寬高比約為 1:2，因此單元格必須滿足 cw = 2 * ch 才是視覺正方形；
// 同時為確保數字永遠在垂直正中央（上下對稱留白），高度 ch 必須為奇數 (1, 3, 5)：
// 1x = 2×1、2x = 6×3、3x = 10×5。
func (m Model) cellSize() (cw, ch int) {
	if m.board == nil {
		return 2, 1
	}

	bw := m.board.Width()
	bh := m.board.Height()

	if m.width <= 0 || m.height <= 0 {
		switch m.scale {
		case 3:
			return 10, 5
		case 2:
			return 6, 3
		default:
			return 2, 1
		}
	}

	targetScale := m.scale
	if targetScale <= 0 {
		// 自動模式：優先選擇能容納的奇數高正方形比例（cw = 2 * ch）
		switch {
		case m.height >= bh*5+chromeHeight && m.width >= bw*10+rowLabelWidth+4:
			targetScale = 3
		case m.canFit(6, 3):
			targetScale = 2
		default:
			targetScale = 1
		}
	}

	switch targetScale {
	case 3:
		if m.canFit(10, 5) {
			return 10, 5
		}
		if m.canFit(6, 3) {
			return 6, 3
		}
		return 2, 1
	case 2:
		if m.canFit(6, 3) {
			return 6, 3
		}
		return 2, 1
	default: // 1x
		return 2, 1
	}
}

func (m Model) canFit(cw, ch int) bool {
	bw := m.board.Width()
	bh := m.board.Height()
	boxW := bw*cw + rowLabelWidth + 4
	return m.height >= bh*ch+2 && m.width >= boxW
}

func (m Model) sizeWarning() (string, bool) {
	if m.width == 0 || m.height == 0 || m.board == nil {
		return "", false
	}
	needW := max(rowLabelWidth+m.board.Width()*2+4, minLayoutW)
	needH := m.board.Height() + chromeHeight
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
	left := titleStyle.Background(m.accentColor()).Render("✹ MINETUI")
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

// renderSidePanel 為寬螢幕模式繪製左側資訊欄（整合標題、地雷數、時間、進度條、狀態列與快捷鍵）。
func (m Model) renderSidePanel(w int) string {
	b := m.board
	title := titleStyle.Background(m.accentColor()).Render("✹ MINETUI")

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

	topLine := spread(w, title, state)
	diffLine := subtitleStyle.Render(m.diff.Summary())

	mines := chip("⚑ 地雷", fmt.Sprintf("%3d", b.RemainingMines()), colRed)
	timer := chip("時間", formatDuration(m.elapsedTime()), colBlue)
	statsLine := spread(w, mines, timer)

	var pairs [][2]string
	if b.Over() {
		pairs = [][2]string{{"r", "再來一局"}, {"d", "換難度"}, {"+/-", "縮放"}, {"?", "說明"}, {"q", "離開"}}
	} else {
		pairs = [][2]string{{"hjkl", "移動"}, {"space", "翻開"}, {"f", "插旗"}, {":", "跳轉"}, {"+/-", "縮放"}, {"?", "說明"}}
	}
	var hintLines []string
	for i := 0; i < len(pairs); i += 2 {
		end := min(i+2, len(pairs))
		hintLines = append(hintLines, renderHints(pairs[i:end]))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		topLine,
		diffLine,
		"",
		statsLine,
		"",
		m.renderProgress(w),
		"",
		m.renderStatusBar(w),
		m.renderCmdline(w),
		"",
		strings.Join(hintLines, "\n"),
	)
}

// renderBoardBoxCompact 繪製高度緊湊的棋盤外框，將欄號直接嵌入上邊框（使 16×3=48 行的棋盤加框後僅佔 50 行）。
// 當 embedFooter 為 true 時（如全螢幕高級 30×16 在 6×3 模式下），亦將地雷、計時與狀態列嵌入下邊框。
func (m Model) renderBoardBoxCompact(cw, ch int, embedFooter bool) string {
	b := m.board
	bw := b.Width()
	bh := b.Height()

	if !embedFooter && m.height >= bh*ch+3 {
		return boardBoxStyle.BorderForeground(m.accentColor()).Render(m.renderBoard(cw, ch))
	}

	bStyle := lipgloss.NewStyle().Foreground(m.accentColor())
	vBar := bStyle.Render("│")

	var lines []string

	// 當高度 >= bh*ch+2（例如 50 行），將欄號嵌入上邊框 ╭──── 1 ──── 2 ... ─╮
	if m.height >= bh*ch+2 {
		var top strings.Builder
		top.WriteString(bStyle.Render("╭" + strings.Repeat("─", 1+rowLabelWidth)))
		for c := 0; c < bw; c++ {
			active := (c == m.cur.Col)
			numStr := strconv.Itoa(c + 1)
			token := " " + numStr + " "
			tw := lipgloss.Width(token)
			leftD := (cw - tw) / 2
			rightD := cw - tw - leftD
			if leftD < 0 {
				leftD = 0
				rightD = 0
			}
			st := labelStyle
			if active {
				st = labelActiveStyle
			}
			top.WriteString(bStyle.Render(strings.Repeat("─", leftD)))
			top.WriteString(st.Render(token))
			top.WriteString(bStyle.Render(strings.Repeat("─", rightD)))
		}
		top.WriteString(bStyle.Render("─╮"))
		lines = append(lines, top.String())
	} else if m.height == bh*ch+1 {
		// 高度恰為 bh*ch+1（49 行）：保留欄號列，省略上下邊框線
		var hdr strings.Builder
		hdr.WriteString(vBar + " " + strings.Repeat(" ", rowLabelWidth))
		for c := 0; c < bw; c++ {
			active := (c == m.cur.Col)
			st := labelStyle
			if active {
				st = labelActiveStyle
			}
			hdr.WriteString(st.Render(colLabel(c+1, cw, active)))
		}
		hdr.WriteString(" " + vBar)
		lines = append(lines, hdr.String())
	}

	for r := 0; r < bh; r++ {
		isCurRow := (r == m.cur.Row)
		for sub := 0; sub < ch; sub++ {
			var rowSb strings.Builder
			rowSb.WriteString(vBar + " ")
			if sub == ch/2 {
				st := labelStyle
				if isCurRow {
					st = labelActiveStyle
				}
				rowSb.WriteString(st.Render(fmt.Sprintf("%2d ", r+1)))
			} else {
				rowSb.WriteString(strings.Repeat(" ", rowLabelWidth))
			}
			for c := 0; c < bw; c++ {
				rowSb.WriteString(m.renderCellSubLine(r, c, cw, ch, sub))
			}
			rowSb.WriteString(" " + vBar)
			lines = append(lines, rowSb.String())
		}
	}

	if m.height >= bh*ch+2 {
		innerW := 1 + rowLabelWidth + bw*cw + 1
		if embedFooter {
			lines = append(lines, m.renderEmbeddedFooterBorder(innerW, cw, ch, bStyle))
		} else {
			lines = append(lines, bStyle.Render("╰"+strings.Repeat("─", innerW)+"╯"))
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderEmbeddedFooterBorder(innerW, cw, ch int, bStyle lipgloss.Style) string {
	b := m.board
	var leftToken string
	switch {
	case m.mode == modeCommand:
		leftToken = " " + modeCommandStyle.Render("COMMAND") + " " + infoStyle.Render(":"+m.cmdBuf) + cmdCursor.Render(" ") + " "
	case m.msg != "":
		st := infoStyle
		switch m.msgStyle {
		case msgError:
			st = errorStyle
		case msgSuccess:
			st = successStyle
		}
		leftToken = " " + st.Render(m.msg) + " "
	default:
		mines := chip("⚑ 地雷", fmt.Sprintf("%3d", b.RemainingMines()), colRed)
		timer := chip("時間", formatDuration(m.elapsedTime()), colBlue)
		leftToken = " " + mines + "  " + timer + " "
	}

	rightToken := " " + modeNormalStyle.Render("NORMAL") +
		subtitleStyle.Render(fmt.Sprintf(" %s %d×%d ", m.diff.Name, cw, ch)) +
		posStyle.Render(fmt.Sprintf("列 %02d 行 %02d", m.cur.Row+1, m.cur.Col+1)) + " "

	leftW := lipgloss.Width(leftToken)
	rightW := lipgloss.Width(rightToken)
	dashes := innerW - 2 - leftW - rightW
	if dashes < 1 {
		return bStyle.Render("╰" + strings.Repeat("─", innerW) + "╯")
	}
	return bStyle.Render("╰─") + leftToken + bStyle.Render(strings.Repeat("─", dashes)) + rightToken + bStyle.Render("─╯")
}

// renderBoard 繪製含座標標籤的棋盤，支援多行多欄自適應單元格。
func (m Model) renderBoard(cw, ch int) string {
	b := m.board
	var sb strings.Builder

	// 欄號標籤
	sb.WriteString(strings.Repeat(" ", rowLabelWidth))
	for c := 0; c < b.Width(); c++ {
		active := (c == m.cur.Col)
		style := labelStyle
		if active {
			style = labelActiveStyle
		}
		sb.WriteString(style.Render(colLabel(c+1, cw, active)))
	}

	// 棋盤列
	for r := 0; r < b.Height(); r++ {
		isCurRow := (r == m.cur.Row)
		for sub := 0; sub < ch; sub++ {
			sb.WriteByte('\n')

			// 列號標籤：固定置於單元格垂直正中間行 (sub == ch/2)
			if sub == ch/2 {
				st := labelStyle
				if isCurRow {
					st = labelActiveStyle
				}
				sb.WriteString(st.Render(fmt.Sprintf("%2d ", r+1)))
			} else {
				sb.WriteString(strings.Repeat(" ", rowLabelWidth))
			}

			for c := 0; c < b.Width(); c++ {
				sb.WriteString(m.renderCellSubLine(r, c, cw, ch, sub))
			}
		}
	}
	return sb.String()
}

func colLabel(n, cw int, active bool) string {
	if cw == 2 {
		// 緊湊模式下兩位數會黏在一起，只標示 1、5 的倍數與游標所在行
		if n == 1 || n%5 == 0 || active {
			return fmt.Sprintf("%2d", n)
		}
		return "  "
	}
	return padCenter(strconv.Itoa(n), cw)
}

func (m Model) renderCellSubLine(r, c, cw, ch, sub int) string {
	glyph, st := m.cellAppearance(r, c)
	if sub == ch/2 {
		return st.Render(padCenter(glyph, cw))
	}
	return st.Render(strings.Repeat(" ", cw))
}

func (m Model) cellAppearance(r, c int) (string, lipgloss.Style) {
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
	return glyph, st
}

func padCenter(s string, width int) string {
	sw := lipgloss.Width(s)
	if sw >= width {
		return s
	}
	left := (width - sw) / 2
	right := width - sw - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (m Model) renderStatusBar(w int) string {
	bar := lipgloss.NewStyle().Background(colSurface0)

	var mode string
	if m.mode == modeCommand {
		mode = modeCommandStyle.Render("COMMAND")
	} else {
		mode = modeNormalStyle.Render("NORMAL")
	}

	scaleInfo := ""
	if m.scale > 0 {
		scaleInfo = bar.Foreground(colYellow).Render(fmt.Sprintf(" %dx", m.scale))
	} else {
		cw, ch := m.cellSize()
		scaleInfo = bar.Foreground(colOverlay).Render(fmt.Sprintf(" %d×%d", cw, ch))
	}
	left := mode + bar.Foreground(colSubtext).Render(" "+m.diff.Name) + scaleInfo

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
		return " " // 留出 1 行高度，避免跳出訊息時造成版面抖動
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
		pairs = [][2]string{{"r", "再來一局"}, {"d", "換難度"}, {"+/-", "縮放"}, {"?", "說明"}, {"q", "離開"}}
	default:
		pairs = [][2]string{{"hjkl", "移動"}, {"space", "翻開"}, {"f", "插旗"}, {":", "跳轉"}, {"+/-", "縮放"}, {"?", "說明"}}
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
		{"+ / -", "手動放大 / 縮小畫面"},
		{"z", "重設為自動縮放"},
	}},
	{"指令", [][2]string{
		{":12 5", "跳到第 12 列、第 5 行"},
		{":12", "跳到第 12 列"},
		{":12 5 f", "跳過去並插旗（o 翻開、c 大面積翻開）"},
		{":easy :medium :hard", "切換難度"},
		{":custom W H M", "自訂寬、高與地雷數"},
		{":zoom [1-3|auto]", "設定畫面縮放倍率"},
		{":r  :menu  :q", "重開 / 難度選單 / 離開"},
	}},
	{"其他", [][2]string{
		{"r / d", "重新開始 / 難度選單"},
		{"? / q", "說明 / 離開"},
	}},
}

func (m Model) renderHelp() string {
	renderSec := func(sec helpSection) string {
		var sb strings.Builder
		sb.WriteString(helpHeadStyle.Render(sec.title))
		keyCol := lipgloss.NewStyle().Width(20)
		for _, row := range sec.rows {
			sb.WriteString("\n  ")
			sb.WriteString(keyCol.Render(helpKeyStyle.Render(row[0])))
			sb.WriteString(helpDescStyle.Render(row[1]))
		}
		return sb.String()
	}

	// 寬度足夠時以兩欄排版（節省縱向高度）
	if m.width >= 78 {
		left := lipgloss.JoinVertical(lipgloss.Left,
			renderSec(helpSections[0]),
			"",
			renderSec(helpSections[1]),
		)
		right := lipgloss.JoinVertical(lipgloss.Left,
			renderSec(helpSections[2]),
			"",
			renderSec(helpSections[3]),
		)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", right)
	}

	var sb strings.Builder
	for i, sec := range helpSections {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(renderSec(sec))
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
