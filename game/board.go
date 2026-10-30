// Package game 實作與介面無關的踩地雷核心規則。
package game

import (
	"errors"
	"fmt"
	"math/rand"
)

// CellState 表示格子目前對玩家呈現的狀態。
type CellState uint8

const (
	Hidden CellState = iota
	Revealed
	Flagged
)

// Cell 是盤面上的單一格子。
type Cell struct {
	Mine     bool
	Adjacent int // 周圍八格的地雷數
	State    CellState
}

// Status 是整局遊戲的狀態。
type Status uint8

const (
	Ready   Status = iota // 尚未踏出第一步，地雷尚未佈置
	Playing               // 進行中
	Won                   // 所有安全格皆已翻開
	Lost                  // 踩到地雷
)

func (s Status) String() string {
	switch s {
	case Ready:
		return "ready"
	case Playing:
		return "playing"
	case Won:
		return "won"
	case Lost:
		return "lost"
	}
	return fmt.Sprintf("Status(%d)", s)
}

// Pos 是以 0 起算的 (列, 行) 座標。
type Pos struct{ Row, Col int }

// ErrInvalidConfig 表示盤面尺寸或地雷數不合法。
var ErrInvalidConfig = errors.New("invalid board configuration")

// Board 是一局踩地雷的完整狀態。
//
// 地雷會延遲到玩家第一次翻開格子時才佈置，並避開第一步所在的格子
// （盤面夠大時連同周圍八格也一起避開），保證第一步絕對安全且必定展開一片區域。
type Board struct {
	width, height, mines int

	cells    [][]Cell
	status   Status
	revealed int
	flags    int
	exploded *Pos

	rng *rand.Rand
}

// New 建立一個 width × height、含 mines 顆地雷的新盤面。
// rng 為 nil 時使用全域亂數來源。
func New(width, height, mines int, rng *rand.Rand) (*Board, error) {
	if width < 2 || height < 2 {
		return nil, fmt.Errorf("%w: board must be at least 2x2, got %dx%d", ErrInvalidConfig, width, height)
	}
	if mines < 1 || mines > width*height-1 {
		return nil, fmt.Errorf("%w: mines must be between 1 and %d, got %d", ErrInvalidConfig, width*height-1, mines)
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	cells := make([][]Cell, height)
	for r := range cells {
		cells[r] = make([]Cell, width)
	}
	return &Board{
		width:  width,
		height: height,
		mines:  mines,
		cells:  cells,
		rng:    rng,
	}, nil
}

func (b *Board) Width() int     { return b.width }
func (b *Board) Height() int    { return b.height }
func (b *Board) Mines() int     { return b.mines }
func (b *Board) Flags() int     { return b.flags }
func (b *Board) Status() Status { return b.status }

// Over 回傳遊戲是否已結束（勝利或失敗）。
func (b *Board) Over() bool { return b.status == Won || b.status == Lost }

// RemainingMines 回傳「地雷總數 − 旗子數」，可能為負數。
func (b *Board) RemainingMines() int { return b.mines - b.flags }

// RevealedCount 回傳已翻開的安全格數。
func (b *Board) RevealedCount() int { return b.revealed }

// SafeTotal 回傳盤面上的安全格總數。
func (b *Board) SafeTotal() int { return b.width*b.height - b.mines }

// Progress 回傳 0~1 的進度（已翻開安全格 / 安全格總數）。
func (b *Board) Progress() float64 {
	return float64(b.revealed) / float64(b.SafeTotal())
}

// InBounds 回傳座標是否位於盤面內。
func (b *Board) InBounds(r, c int) bool {
	return r >= 0 && r < b.height && c >= 0 && c < b.width
}

// Cell 回傳指定格子的副本。
func (b *Board) Cell(r, c int) Cell { return b.cells[r][c] }

// Exploded 回傳引爆的地雷位置（僅在 Lost 狀態下有值）。
func (b *Board) Exploded() (Pos, bool) {
	if b.exploded == nil {
		return Pos{}, false
	}
	return *b.exploded, true
}

// Open 是玩家的「翻開」動作：
//   - 隱藏格：翻開；若為 0 則以 flood fill 大面積展開。
//   - 已翻開的數字格：等同 Chord（周圍旗數等於數字時翻開其餘鄰格）。
//   - 插旗格：不動作。
//
// 回傳盤面是否有變化。
func (b *Board) Open(r, c int) bool {
	if b.Over() || !b.InBounds(r, c) {
		return false
	}
	switch b.cells[r][c].State {
	case Flagged:
		return false
	case Revealed:
		return b.Chord(r, c)
	}

	if b.status == Ready {
		b.placeMines(Pos{r, c})
		b.status = Playing
	}

	if b.cells[r][c].Mine {
		b.explode(r, c)
		return true
	}
	b.flood(r, c)
	b.checkWin()
	return true
}

// Chord 在已翻開的數字格上，若周圍旗子數恰等於該數字，便一次翻開周圍所有未插旗的格子。
// 若旗子插錯，會因此踩到地雷。回傳盤面是否有變化。
func (b *Board) Chord(r, c int) bool {
	if b.Over() || !b.InBounds(r, c) {
		return false
	}
	cell := b.cells[r][c]
	if cell.State != Revealed || cell.Adjacent == 0 {
		return false
	}

	flags := 0
	b.forNeighbors(r, c, func(nr, nc int) {
		if b.cells[nr][nc].State == Flagged {
			flags++
		}
	})
	if flags != cell.Adjacent {
		return false
	}

	changed := false
	var hit *Pos
	b.forNeighbors(r, c, func(nr, nc int) {
		n := &b.cells[nr][nc]
		if n.State != Hidden {
			return
		}
		changed = true
		if n.Mine {
			if hit == nil {
				hit = &Pos{nr, nc}
			}
			return
		}
		b.flood(nr, nc)
	})

	if hit != nil {
		b.explode(hit.Row, hit.Col)
	} else {
		b.checkWin()
	}
	return changed
}

// ToggleFlag 切換隱藏格的旗標。回傳是否有變化。
func (b *Board) ToggleFlag(r, c int) bool {
	if b.Over() || !b.InBounds(r, c) {
		return false
	}
	cell := &b.cells[r][c]
	switch cell.State {
	case Hidden:
		cell.State = Flagged
		b.flags++
	case Flagged:
		cell.State = Hidden
		b.flags--
	default:
		return false
	}
	return true
}

// placeMines 隨機佈雷並避開 safe（盤面允許時連同其周圍八格）。
func (b *Board) placeMines(safe Pos) {
	all := make([]Pos, 0, b.width*b.height)
	zone := make([]Pos, 0, b.width*b.height)
	for r := 0; r < b.height; r++ {
		for c := 0; c < b.width; c++ {
			if r == safe.Row && c == safe.Col {
				continue
			}
			all = append(all, Pos{r, c})
			if abs(r-safe.Row) > 1 || abs(c-safe.Col) > 1 {
				zone = append(zone, Pos{r, c})
			}
		}
	}

	candidates := zone
	if len(candidates) < b.mines {
		// 地雷太密，無法讓出整個 3x3：至少保證第一格本身安全。
		candidates = all
	}
	b.rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	for _, p := range candidates[:b.mines] {
		b.cells[p.Row][p.Col].Mine = true
	}
	b.computeAdjacent()
}

func (b *Board) computeAdjacent() {
	for r := 0; r < b.height; r++ {
		for c := 0; c < b.width; c++ {
			n := 0
			b.forNeighbors(r, c, func(nr, nc int) {
				if b.cells[nr][nc].Mine {
					n++
				}
			})
			b.cells[r][c].Adjacent = n
		}
	}
}

// flood 從 (r, c) 開始翻開，遇到 0 便持續向外擴散。插旗格不會被翻開。
func (b *Board) flood(r, c int) {
	stack := []Pos{{r, c}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		cell := &b.cells[p.Row][p.Col]
		if cell.State != Hidden || cell.Mine {
			continue
		}
		cell.State = Revealed
		b.revealed++

		if cell.Adjacent == 0 {
			b.forNeighbors(p.Row, p.Col, func(nr, nc int) {
				if b.cells[nr][nc].State == Hidden {
					stack = append(stack, Pos{nr, nc})
				}
			})
		}
	}
}

func (b *Board) explode(r, c int) {
	b.status = Lost
	b.cells[r][c].State = Revealed
	b.exploded = &Pos{r, c}
}

func (b *Board) checkWin() {
	if b.revealed != b.SafeTotal() {
		return
	}
	b.status = Won
	// 勝利時自動把剩下的地雷全部插上旗子
	for r := range b.cells {
		for c := range b.cells[r] {
			if b.cells[r][c].Mine {
				b.cells[r][c].State = Flagged
			}
		}
	}
	b.flags = b.mines
}

func (b *Board) forNeighbors(r, c int, fn func(nr, nc int)) {
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			nr, nc := r+dr, c+dc
			if b.InBounds(nr, nc) {
				fn(nr, nc)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
