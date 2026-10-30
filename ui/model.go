// Package ui 是踩地雷的 Bubble Tea 終端機介面。
package ui

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/shiangogo/minesweeper/game"
)

type screen int

const (
	screenMenu screen = iota
	screenGame
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeCommand
)

const tickInterval = 100 * time.Millisecond

// tickMsg 驅動計時器重繪；gen 用來丟棄上一局遺留的 tick。
type tickMsg struct{ gen int }

func tick(gen int) tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{gen: gen} })
}

// Options 是啟動參數。
type Options struct {
	// Difficulty 不為 nil 時跳過選單直接開局。
	Difficulty *Difficulty
	// Seed 為 0 時使用隨機種子。
	Seed int64
}

// Model 是整個應用程式的 Bubble Tea model。
type Model struct {
	screen        screen
	width, height int
	rng           *rand.Rand

	// 選單
	menuIdx     int
	customOpen  bool
	customFocus int
	customIn    [3]textinput.Model
	menuErr     string

	// 遊戲
	diff  Difficulty
	board *game.Board
	gen   int
	cur   game.Pos

	// vim 風格輸入
	mode     inputMode
	count    string // 數字前綴，如 5j
	pendingG bool   // 已按下第一個 g，等待 gg
	cmdBuf   string
	history  []string
	histIdx  int

	// 狀態列訊息
	msg      string
	msgStyle int // 0 info, 1 error, 2 success

	// 計時器與進度條
	start    time.Time
	elapsed  time.Duration
	progress progress.Model

	// 畫面縮放倍率：0 為自動，1 為 1x，2 為 2x，3 為 3x
	scale int

	showHelp bool
}

const (
	msgInfo = iota
	msgError
	msgSuccess
)

// New 建立初始 model。
func New(opts Options) Model {
	seed := opts.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	m := Model{
		rng: rand.New(rand.NewSource(seed)),
		progress: progress.New(
			progress.WithScaledGradient(string(colBlue), string(colPink)),
			progress.WithSpringOptions(18, 1),
		),
	}

	placeholders := [3]string{"20", "12", "40"}
	for i := range m.customIn {
		ti := textinput.New()
		ti.CharLimit = 4
		ti.Width = 6
		ti.Prompt = ""
		ti.SetValue(placeholders[i])
		ti.TextStyle = menuSelectedStyle
		m.customIn[i] = ti
	}

	if opts.Difficulty != nil {
		m.newGame(*opts.Difficulty)
	}
	return m
}

// Init 實作 tea.Model。
func (m Model) Init() tea.Cmd { return nil }

// Update 實作 tea.Model。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		if msg.gen == m.gen && m.board != nil && m.board.Status() == game.Playing {
			return m, tick(m.gen)
		}
		return m, nil

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.screen == screenMenu {
			return m.updateMenu(msg)
		}
		return m.updateGame(msg)
	}

	// 其餘訊息（例如游標閃爍）交給自訂難度的輸入框
	if m.screen == screenMenu && m.customOpen {
		var cmd tea.Cmd
		m.customIn[m.customFocus], cmd = m.customIn[m.customFocus].Update(msg)
		return m, cmd
	}
	return m, nil
}

// View 實作 tea.Model。
func (m Model) View() string {
	if m.screen == screenMenu {
		return m.viewMenu()
	}
	return m.viewGame()
}

// newGame 以指定難度開新局。
func (m *Model) newGame(d Difficulty) tea.Cmd {
	b, err := game.New(d.Width, d.Height, d.Mines, m.rng)
	if err != nil {
		m.setMsg(msgError, err.Error())
		return nil
	}
	m.diff = d
	m.board = b
	m.gen++
	m.screen = screenGame
	m.cur = game.Pos{Row: d.Height / 2, Col: d.Width / 2}
	m.mode = modeNormal
	m.count, m.pendingG, m.cmdBuf = "", false, ""
	m.start, m.elapsed = time.Time{}, 0
	m.setMsg(msgInfo, "第一步絕對安全，放心踩下去吧！")
	return m.progress.SetPercent(0)
}

// elapsedTime 回傳目前計時器顯示的時間。
func (m Model) elapsedTime() time.Duration {
	if m.board != nil && m.board.Status() == game.Playing {
		return time.Since(m.start)
	}
	return m.elapsed
}

func (m *Model) setMsg(style int, s string) {
	m.msg, m.msgStyle = s, style
}
