package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shiangogo/minetui/game"
)

const maxHistory = 50

func (m Model) updateGame(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeCommand {
		return m.updateCommand(msg)
	}
	return m.updateNormal(msg)
}

// updateNormal 處理 vim 風格的 normal mode 按鍵。
func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// 數字前綴（0 在沒有前綴時代表「移到列首」）
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' && !(key == "0" && m.count == "") {
		if len(m.count) < 3 {
			m.count += key
		}
		return m, nil
	}

	// gg：第一個 g 先記住，保留數字前綴（如 5gg）
	if key == "g" && !m.pendingG {
		m.pendingG = true
		return m, nil
	}
	n, hasCount := m.consumeCount()
	if m.pendingG {
		m.pendingG = false
		if key == "g" {
			row := 0
			if hasCount {
				row = n - 1
			}
			m.moveTo(row, m.cur.Col)
			return m, nil
		}
		// 其他按鍵：取消 g，照常處理
	}

	b := m.board
	switch key {
	// ── 移動 ──
	case "h", "left":
		m.moveTo(m.cur.Row, m.cur.Col-n)
	case "l", "right":
		m.moveTo(m.cur.Row, m.cur.Col+n)
	case "k", "up":
		m.moveTo(m.cur.Row-n, m.cur.Col)
	case "j", "down":
		m.moveTo(m.cur.Row+n, m.cur.Col)
	case "0", "^", "home":
		m.moveTo(m.cur.Row, 0)
	case "$", "end":
		m.moveTo(m.cur.Row, b.Width()-1)
	case "G":
		row := b.Height() - 1
		if hasCount {
			row = n - 1
		}
		m.moveTo(row, m.cur.Col)
	case "H":
		m.moveTo(0, m.cur.Col)
	case "M":
		m.moveTo(b.Height()/2, b.Width()/2)
	case "L":
		m.moveTo(b.Height()-1, m.cur.Col)
	case "w", "tab":
		for i := 0; i < n; i++ {
			m.jumpHidden(1)
		}
	case "b", "shift+tab":
		for i := 0; i < n; i++ {
			m.jumpHidden(-1)
		}

	// ── 操作 ──
	case " ", "enter", "o", "x":
		return m.open(m.cur.Row, m.cur.Col)
	case "c":
		return m.chord(m.cur.Row, m.cur.Col)
	case "f", "m":
		m.flag(m.cur.Row, m.cur.Col)
	case "+", "=":
		m.zoomIn()
	case "-":
		m.zoomOut()
	case "z":
		m.resetZoom()

	// ── 遊戲 ──
	case ":":
		m.mode = modeCommand
		m.cmdBuf = ""
		m.histIdx = len(m.history)
	case "r":
		cmd := m.newGame(m.diff)
		return m, cmd
	case "d":
		m.openMenu()
	case "?":
		m.showHelp = !m.showHelp
	case "esc":
		m.showHelp = false
		m.msg = ""
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// consumeCount 取出並清除數字前綴，預設為 1。
func (m *Model) consumeCount() (int, bool) {
	if m.count == "" {
		return 1, false
	}
	n, _ := strconv.Atoi(m.count)
	m.count = ""
	if n < 1 {
		n = 1
	}
	return n, true
}

func (m *Model) moveTo(r, c int) {
	m.cur.Row = clamp(r, 0, m.board.Height()-1)
	m.cur.Col = clamp(c, 0, m.board.Width()-1)
}

// jumpHidden 依閱讀順序跳到下一個（dir=1）或上一個（dir=-1）尚未翻開、未插旗的格子。
func (m *Model) jumpHidden(dir int) {
	b := m.board
	total := b.Width() * b.Height()
	idx := m.cur.Row*b.Width() + m.cur.Col
	for i := 1; i < total; i++ {
		j := ((idx+dir*i)%total + total) % total
		r, c := j/b.Width(), j%b.Width()
		if b.Cell(r, c).State == game.Hidden {
			m.cur = game.Pos{Row: r, Col: c}
			return
		}
	}
}

func (m *Model) openMenu() {
	m.screen = screenMenu
	m.menuIdx = len(Presets)
	for i, d := range Presets {
		if d.Key == m.diff.Key {
			m.menuIdx = i
		}
	}
}

// open 翻開格子（數字格上自動 chord），並處理計時器、進度條與勝負訊息。
func (m Model) open(r, c int) (tea.Model, tea.Cmd) {
	if m.board.Over() {
		m.gameOverHint()
		return m, nil
	}
	if m.board.Cell(r, c).State == game.Flagged {
		m.setMsg(msgInfo, "這格插了旗子，先按 f 拔掉旗子才能翻開")
		return m, nil
	}
	wasReady := m.board.Status() == game.Ready
	if !m.board.Open(r, c) {
		if m.board.Cell(r, c).State == game.Revealed && m.board.Cell(r, c).Adjacent > 0 {
			m.setMsg(msgInfo, "周圍旗子數與數字不符，無法大面積翻開")
		}
		return m, nil
	}
	return m, m.afterChange(wasReady)
}

func (m Model) chord(r, c int) (tea.Model, tea.Cmd) {
	if m.board.Over() {
		m.gameOverHint()
		return m, nil
	}
	cell := m.board.Cell(r, c)
	if cell.State != game.Revealed || cell.Adjacent == 0 {
		m.setMsg(msgInfo, "大面積翻開需要在已翻開的數字格上使用")
		return m, nil
	}
	if !m.board.Chord(r, c) {
		m.setMsg(msgInfo, "周圍旗子數與數字不符，無法大面積翻開")
		return m, nil
	}
	return m, m.afterChange(false)
}

func (m *Model) flag(r, c int) {
	if m.board.Over() {
		m.gameOverHint()
		return
	}
	if !m.board.ToggleFlag(r, c) {
		m.setMsg(msgInfo, "已翻開的格子不能插旗")
		return
	}
	m.msg = ""
}

// afterChange 在盤面變動後更新計時器、進度條與狀態訊息。
func (m *Model) afterChange(wasReady bool) tea.Cmd {
	var cmds []tea.Cmd
	now := time.Now()

	if wasReady && m.board.Status() != game.Ready {
		m.start = now
		if m.board.Status() == game.Playing {
			cmds = append(cmds, tick(m.gen))
		}
	}

	switch m.board.Status() {
	case game.Won:
		m.elapsed = now.Sub(m.start)
		m.setMsg(msgSuccess, fmt.Sprintf("恭喜過關！用時 %s ── 按 r 再來一局、d 換難度", formatDuration(m.elapsed)))
	case game.Lost:
		m.elapsed = now.Sub(m.start)
		m.setMsg(msgError, "BOOM！踩到地雷了 ── 按 r 重新開始、d 換難度")
	default:
		m.msg = ""
	}

	cmds = append(cmds, m.progress.SetPercent(m.board.Progress()))
	return tea.Batch(cmds...)
}

func (m *Model) gameOverHint() {
	if m.board.Status() == game.Won {
		m.setMsg(msgSuccess, "已經過關囉！按 r 再來一局、d 換難度")
	} else {
		m.setMsg(msgError, "遊戲結束 ── 按 r 重新開始、d 換難度")
	}
}

// updateCommand 處理 : 指令列輸入。
func (m Model) updateCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		m.cmdBuf = ""
	case tea.KeyEnter:
		line := strings.TrimSpace(m.cmdBuf)
		m.mode = modeNormal
		m.cmdBuf = ""
		if line != "" {
			if len(m.history) == 0 || m.history[len(m.history)-1] != line {
				m.history = append(m.history, line)
				if len(m.history) > maxHistory {
					m.history = m.history[1:]
				}
			}
		}
		return m.execCommand(line)
	case tea.KeyBackspace:
		if m.cmdBuf == "" {
			m.mode = modeNormal // 與 vim 相同：空指令列按退格會離開
			return m, nil
		}
		rs := []rune(m.cmdBuf)
		m.cmdBuf = string(rs[:len(rs)-1])
	case tea.KeyCtrlU:
		m.cmdBuf = ""
	case tea.KeyUp:
		if m.histIdx > 0 {
			m.histIdx--
			m.cmdBuf = m.history[m.histIdx]
		}
	case tea.KeyDown:
		if m.histIdx < len(m.history)-1 {
			m.histIdx++
			m.cmdBuf = m.history[m.histIdx]
		} else {
			m.histIdx = len(m.history)
			m.cmdBuf = ""
		}
	case tea.KeySpace:
		m.cmdBuf += " "
	case tea.KeyRunes:
		m.cmdBuf += string(msg.Runes)
	}
	return m, nil
}

// execCommand 執行 : 指令。
//
//	:12 5 / :12,5   跳到第 12 列第 5 行
//	:12             跳到第 12 列（保留目前的行）
//	:12 5 f|o|c     跳過去並插旗 / 翻開 / 大面積翻開
//	:easy :medium :hard :custom W H M :diff <名稱>
//	:r :new :menu :help :q
func (m Model) execCommand(line string) (tea.Model, tea.Cmd) {
	fields := strings.FieldsFunc(line, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	})
	if len(fields) == 0 {
		return m, nil
	}

	if _, err := strconv.Atoi(fields[0]); err == nil {
		return m.execJump(fields)
	}

	name := strings.ToLower(fields[0])
	switch name {
	case "q", "q!", "quit", "qa", "qa!", "wq", "x", "exit":
		return m, tea.Quit
	case "r", "restart", "new", "e!":
		cmd := m.newGame(m.diff)
		return m, cmd
	case "menu", "d":
		m.openMenu()
		return m, nil
	case "help", "h":
		m.showHelp = !m.showHelp
		return m, nil
	case "zoom", "scale":
		if len(fields) == 1 {
			cw, ch := m.cellSize()
			mode := "自動"
			if m.scale > 0 {
				mode = fmt.Sprintf("手動 %dx", m.scale)
			}
			m.setMsg(msgInfo, fmt.Sprintf("目前縮放：%s（單元格 %d×%d，按 + / - 調整，z 重設為自動）", mode, cw, ch))
			return m, nil
		}
		switch strings.ToLower(fields[1]) {
		case "auto", "0", "reset", "a":
			m.resetZoom()
		case "1", "1x":
			m.scale = 1
			m.setMsg(msgInfo, "畫面縮放設為 1x")
		case "2", "2x":
			m.scale = 2
			m.setMsg(msgInfo, "畫面縮放設為 2x")
		case "3", "3x":
			m.scale = 3
			m.setMsg(msgInfo, "畫面縮放設為 3x")
		case "+", "in":
			m.zoomIn()
		case "-", "out":
			m.zoomOut()
		default:
			m.setMsg(msgError, "用法：:zoom [1|2|3|auto|+|-]")
		}
		return m, nil
	case "custom", "c":
		if len(fields) != 4 {
			m.setMsg(msgError, "用法：:custom 寬 高 地雷數，例如 :custom 20 12 40")
			return m, nil
		}
		vals := make([]int, 3)
		for i := range vals {
			v, err := strconv.Atoi(fields[i+1])
			if err != nil {
				m.setMsg(msgError, "自訂難度的參數必須是數字")
				return m, nil
			}
			vals[i] = v
		}
		d := Custom(vals[0], vals[1], vals[2])
		if err := d.Validate(); err != nil {
			m.setMsg(msgError, err.Error())
			return m, nil
		}
		cmd := m.newGame(d)
		return m, cmd
	case "diff", "difficulty", "level":
		if len(fields) < 2 {
			m.setMsg(msgError, "用法：:diff easy|medium|hard")
			return m, nil
		}
		name = fields[1]
	}

	if d, ok := PresetByName(name); ok {
		cmd := m.newGame(d)
		return m, cmd
	}
	m.setMsg(msgError, fmt.Sprintf("未知的指令：%s（按 ? 查看說明）", line))
	return m, nil
}

func (m Model) execJump(fields []string) (tea.Model, tea.Cmd) {
	b := m.board
	row, _ := strconv.Atoi(fields[0])
	col := m.cur.Col + 1
	action := ""

	rest := fields[1:]
	if len(rest) > 0 {
		if v, err := strconv.Atoi(rest[0]); err == nil {
			col = v
			rest = rest[1:]
		}
	}
	if len(rest) > 0 {
		action = strings.ToLower(rest[0])
		rest = rest[1:]
	}
	if len(rest) > 0 {
		m.setMsg(msgError, "用法：:列 行 [f|o|c]，例如 :8 12 或 :8 12 f")
		return m, nil
	}

	if row < 1 || row > b.Height() || col < 1 || col > b.Width() {
		m.setMsg(msgError, fmt.Sprintf("座標超出範圍：列 1~%d、行 1~%d", b.Height(), b.Width()))
		return m, nil
	}
	m.cur = game.Pos{Row: row - 1, Col: col - 1}
	m.msg = ""

	switch action {
	case "":
		return m, nil
	case "f", "flag", "m":
		m.flag(m.cur.Row, m.cur.Col)
		return m, nil
	case "o", "open", "x":
		return m.open(m.cur.Row, m.cur.Col)
	case "c", "chord":
		return m.chord(m.cur.Row, m.cur.Col)
	}
	m.setMsg(msgError, fmt.Sprintf("未知的動作：%s（可用 f / o / c）", action))
	return m, nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m *Model) zoomIn() {
	curr := m.currentScale()
	if curr >= 3 {
		m.setMsg(msgInfo, "已達最大縮放比例 (3x: 10×5)")
		return
	}
	next := curr + 1
	if m.width > 0 && m.height > 0 {
		wantW, wantH := 6, 3
		if next == 3 {
			wantW, wantH = 10, 5
		}
		if !m.canFit(wantW, wantH) {
			cw, ch := m.cellSize()
			m.setMsg(msgInfo, fmt.Sprintf("目前視窗尺寸已達最大可用正方形比例 (%d×%d)", cw, ch))
			return
		}
	}
	m.scale = next
	cw, ch := m.cellSize()
	m.setMsg(msgInfo, fmt.Sprintf("畫面縮放：%dx（%d×%d 格，按 - 縮小、z 重設為自動）", m.scale, cw, ch))
}

func (m *Model) zoomOut() {
	curr := m.currentScale()
	if curr <= 1 {
		m.setMsg(msgInfo, "已達最小縮放比例 (1x: 2×1)")
		return
	}
	m.scale = curr - 1
	cw, ch := m.cellSize()
	m.setMsg(msgInfo, fmt.Sprintf("畫面縮放：%dx（%d×%d 格，按 + 放大、z 重設為自動）", m.scale, cw, ch))
}

func (m *Model) resetZoom() {
	m.scale = 0
	cw, ch := m.cellSize()
	m.setMsg(msgInfo, fmt.Sprintf("已重設為自動縮放（目前：%d×%d 格）", cw, ch))
}

func (m Model) currentScale() int {
	if m.scale > 0 {
		return m.scale
	}
	_, ch := m.cellSize()
	switch {
	case ch >= 5:
		return 3
	case ch >= 3:
		return 2
	default:
		return 1
	}
}
