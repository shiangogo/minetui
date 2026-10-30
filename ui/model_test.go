package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shiangogo/minetui/game"
)

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func press(t *testing.T, m Model, keys ...string) Model {
	t.Helper()
	for _, k := range keys {
		next, _ := m.Update(keyMsg(k))
		m = next.(Model)
	}
	return m
}

// typeCmd 輸入一整行 : 指令並按 enter。
func typeCmd(t *testing.T, m Model, line string) Model {
	t.Helper()
	m = press(t, m, ":")
	for _, r := range line {
		m = press(t, m, string(r))
	}
	return press(t, m, "enter")
}

func newEasy(t *testing.T) Model {
	t.Helper()
	d := Presets[0]
	return New(Options{Difficulty: &d, Seed: 42})
}

func TestVimMotions(t *testing.T) {
	m := newEasy(t) // 9x9，游標從 (4,4) 開始
	if m.cur != (game.Pos{Row: 4, Col: 4}) {
		t.Fatalf("start cursor = %v", m.cur)
	}

	cases := []struct {
		keys []string
		want game.Pos
	}{
		{[]string{"h"}, game.Pos{Row: 4, Col: 3}},
		{[]string{"2", "k"}, game.Pos{Row: 2, Col: 3}},
		{[]string{"9", "9", "j"}, game.Pos{Row: 8, Col: 3}}, // 超出邊界會被夾住
		{[]string{"$"}, game.Pos{Row: 8, Col: 8}},
		{[]string{"0"}, game.Pos{Row: 8, Col: 0}},
		{[]string{"g", "g"}, game.Pos{Row: 0, Col: 0}},
		{[]string{"G"}, game.Pos{Row: 8, Col: 0}},
		{[]string{"3", "g", "g"}, game.Pos{Row: 2, Col: 0}},
		{[]string{"7", "G"}, game.Pos{Row: 6, Col: 0}},
		{[]string{"1", "0", "l"}, game.Pos{Row: 6, Col: 8}},
		{[]string{"M"}, game.Pos{Row: 4, Col: 4}},
	}
	for _, tc := range cases {
		m = press(t, m, tc.keys...)
		if m.cur != tc.want {
			t.Fatalf("after %v cursor = %v, want %v", tc.keys, m.cur, tc.want)
		}
	}
}

func TestCommandJump(t *testing.T) {
	m := newEasy(t)

	m = typeCmd(t, m, "3 7")
	if m.cur != (game.Pos{Row: 2, Col: 6}) {
		t.Fatalf(":3 7 -> %v", m.cur)
	}
	m = typeCmd(t, m, "9,1")
	if m.cur != (game.Pos{Row: 8, Col: 0}) {
		t.Fatalf(":9,1 -> %v", m.cur)
	}
	m = typeCmd(t, m, "5")
	if m.cur != (game.Pos{Row: 4, Col: 0}) {
		t.Fatalf(":5 -> %v", m.cur)
	}
	m = typeCmd(t, m, "1 1 f")
	if m.board.Cell(0, 0).State != game.Flagged {
		t.Fatalf(":1 1 f should flag (0,0)")
	}

	before := m.cur
	m = typeCmd(t, m, "42 1")
	if m.cur != before || m.msgStyle != msgError {
		t.Fatalf("out of range jump should error and keep cursor")
	}

	if m.mode != modeNormal {
		t.Fatalf("should return to normal mode after command")
	}
}

func TestCommandModeEditing(t *testing.T) {
	m := newEasy(t)
	m = press(t, m, ":", "1", "2", "backspace")
	if m.cmdBuf != "1" {
		t.Fatalf("cmdBuf = %q, want %q", m.cmdBuf, "1")
	}
	m = press(t, m, "esc")
	if m.mode != modeNormal || m.cmdBuf != "" {
		t.Fatalf("esc should cancel the command")
	}
	m = press(t, m, ":", "backspace")
	if m.mode != modeNormal {
		t.Fatalf("backspace on empty command line should leave command mode")
	}
}

func TestDifficultyCommands(t *testing.T) {
	m := newEasy(t)

	m = typeCmd(t, m, "hard")
	if m.board.Width() != 30 || m.board.Height() != 16 || m.board.Mines() != 99 {
		t.Fatalf(":hard -> %dx%d/%d", m.board.Width(), m.board.Height(), m.board.Mines())
	}
	m = typeCmd(t, m, "diff medium")
	if m.diff.Key != "medium" {
		t.Fatalf(":diff medium -> %s", m.diff.Key)
	}
	m = typeCmd(t, m, "custom 20 10 30")
	if m.diff.Key != "custom" || m.board.Width() != 20 || m.board.Height() != 10 || m.board.Mines() != 30 {
		t.Fatalf(":custom -> %+v", m.diff)
	}
	m = typeCmd(t, m, "custom 20 10 999")
	if m.msgStyle != msgError || m.board.Mines() != 30 {
		t.Fatalf("invalid custom should error and keep the current game")
	}
	m = typeCmd(t, m, "nope")
	if m.msgStyle != msgError {
		t.Fatalf("unknown command should error")
	}
}

func TestFirstOpenStartsTimerAndProgress(t *testing.T) {
	m := newEasy(t)
	if m.board.Status() != game.Ready {
		t.Fatalf("status = %v, want ready", m.board.Status())
	}
	m = press(t, m, " ")
	if s := m.board.Status(); s != game.Playing && s != game.Won {
		t.Fatalf("status after first open = %v", s)
	}
	if m.start.IsZero() {
		t.Fatalf("timer should start on first open")
	}
	if m.board.RevealedCount() == 0 {
		t.Fatalf("first open should reveal cells")
	}
}

func TestFlagKey(t *testing.T) {
	m := newEasy(t)
	m = press(t, m, "f")
	if m.board.Cell(4, 4).State != game.Flagged || m.board.RemainingMines() != 9 {
		t.Fatalf("f should flag the cursor cell")
	}
	m = press(t, m, " ")
	if m.board.Status() != game.Ready {
		t.Fatalf("opening a flagged cell must do nothing")
	}
	m = press(t, m, "f")
	if m.board.Cell(4, 4).State != game.Hidden {
		t.Fatalf("f again should remove the flag")
	}
}

func TestViewRenders(t *testing.T) {
	m := newEasy(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	m = next.(Model)
	m = press(t, m, " ")
	if v := m.View(); v == "" {
		t.Fatalf("empty view")
	}
	m = press(t, m, "?")
	if v := m.View(); v == "" {
		t.Fatalf("empty help view")
	}

	menu := New(Options{Seed: 1})
	if v := menu.View(); v == "" {
		t.Fatalf("empty menu view")
	}
}

func TestMenuCustomFlow(t *testing.T) {
	m := New(Options{Seed: 1})
	m = press(t, m, "4") // 自訂
	if !m.customOpen {
		t.Fatalf("custom form should open")
	}
	// 清空寬度欄位後輸入 12（夾雜非數字會被忽略）
	for i := 0; i < 4; i++ {
		m = press(t, m, "backspace")
	}
	m = press(t, m, "1", "x", "2", "enter")
	if m.screen != screenGame || m.board.Width() != 12 {
		t.Fatalf("custom game not started: screen=%v err=%q", m.screen, m.menuErr)
	}
}

func TestResponsiveWindowScaling(t *testing.T) {
	m := newEasy(t) // 9x9 board

	// 80x24 標準小終端機：高度僅容納 1x
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	cw1, ch1 := m.cellSize()
	if ch1 != 1 {
		t.Fatalf("at 80x24: ch = %d, want 1", ch1)
	}

	// 拉大至 100x35：高度足夠容納 2x，格子自動變大！
	next, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 35})
	m = next.(Model)
	cw2, ch2 := m.cellSize()
	if ch2 != 2 || cw2 < cw1 {
		t.Fatalf("at 100x35: cw=%d ch=%d, expected ch=2 and cw >= %d", cw2, ch2, cw1)
	}

	// 全螢幕 140x50：高度足夠容納 3x，格子進一步放大！
	next, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m = next.(Model)
	cw3, ch3 := m.cellSize()
	if ch3 != 3 || cw3 < cw2 {
		t.Fatalf("at 140x50: cw=%d ch=%d, expected ch=3 and cw >= %d", cw3, ch3, cw2)
	}

	// 縮小回 80x24：自動縮回 1x
	next, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	_, chBack := m.cellSize()
	if chBack != 1 {
		t.Fatalf("back at 80x24: ch = %d, want 1", chBack)
	}
}

func TestManualZoomKeysAndCommands(t *testing.T) {
	m := newEasy(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 50})
	m = next.(Model)

	// 預設為 auto (0)
	if m.scale != 0 {
		t.Fatalf("initial scale = %d, want 0", m.scale)
	}

	// 指令設定 :zoom 1
	m = typeCmd(t, m, "zoom 1")
	if m.scale != 1 {
		t.Fatalf(":zoom 1 -> scale = %d, want 1", m.scale)
	}
	_, ch := m.cellSize()
	if ch != 1 {
		t.Fatalf("forced 1x: ch = %d, want 1", ch)
	}

	// 按 + 放大
	m = press(t, m, "+")
	if m.scale != 2 {
		t.Fatalf("+ pressed -> scale = %d, want 2", m.scale)
	}

	// 按 z 重設為自動
	m = press(t, m, "z")
	if m.scale != 0 {
		t.Fatalf("z pressed -> scale = %d, want 0", m.scale)
	}

	// 指令 :zoom auto
	m = typeCmd(t, m, "zoom 3")
	if m.scale != 3 {
		t.Fatalf(":zoom 3 -> scale = %d", m.scale)
	}
	m = typeCmd(t, m, "zoom auto")
	if m.scale != 0 {
		t.Fatalf(":zoom auto -> scale = %d, want 0", m.scale)
	}
}

func TestSquareCellProportions(t *testing.T) {
	m := newEasy(t)

	// 驗證各 scale 單元格比例皆維持視覺正方形（cw == 2 * ch）
	for scale := 1; scale <= 3; scale++ {
		m.scale = scale
		cw, ch := m.cellSize()
		if cw != 2*ch {
			t.Fatalf("scale %d: cw=%d, ch=%d; expected square proportion cw == 2 * ch", scale, cw, ch)
		}
	}
}
