package game

import (
	"errors"
	"math/rand"
	"testing"
)

// newWithMines 建立指定地雷位置的盤面（跳過隨機佈雷），方便撰寫決定性測試。
func newWithMines(t *testing.T, w, h int, mines ...Pos) *Board {
	t.Helper()
	b, err := New(w, h, len(mines), rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, p := range mines {
		b.cells[p.Row][p.Col].Mine = true
	}
	b.computeAdjacent()
	b.status = Playing
	return b
}

func TestNewValidation(t *testing.T) {
	cases := []struct {
		name     string
		w, h, m  int
		wantFail bool
	}{
		{"beginner", 9, 9, 10, false},
		{"expert", 30, 16, 99, false},
		{"max density", 3, 3, 8, false},
		{"too small", 1, 5, 1, true},
		{"no mines", 9, 9, 0, true},
		{"too many mines", 3, 3, 9, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.w, tc.h, tc.m, nil)
			if tc.wantFail != (err != nil) {
				t.Fatalf("New(%d,%d,%d) err = %v, wantFail %v", tc.w, tc.h, tc.m, err, tc.wantFail)
			}
			if err != nil && !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("expected ErrInvalidConfig, got %v", err)
			}
		})
	}
}

func TestFirstMoveIsSafeAndOpensArea(t *testing.T) {
	for seed := int64(0); seed < 500; seed++ {
		rng := rand.New(rand.NewSource(seed))
		b, _ := New(30, 16, 99, rng)
		r, c := rng.Intn(16), rng.Intn(30)

		if !b.Open(r, c) {
			t.Fatalf("seed %d: first Open returned false", seed)
		}
		if b.Status() == Lost {
			t.Fatalf("seed %d: lost on first move at (%d,%d)", seed, r, c)
		}
		if got := b.Cell(r, c).Adjacent; got != 0 {
			t.Fatalf("seed %d: first cell should be a zero, got %d", seed, got)
		}
		mines := 0
		for rr := 0; rr < 16; rr++ {
			for cc := 0; cc < 30; cc++ {
				if b.Cell(rr, cc).Mine {
					mines++
				}
			}
		}
		if mines != 99 {
			t.Fatalf("seed %d: placed %d mines, want 99", seed, mines)
		}
	}
}

func TestFirstMoveSafeOnDenseBoard(t *testing.T) {
	// 3x3 放 8 顆雷：無法讓出 3x3 安全區，但第一格本身仍須安全。
	for seed := int64(0); seed < 50; seed++ {
		b, _ := New(3, 3, 8, rand.New(rand.NewSource(seed)))
		b.Open(1, 1)
		if b.Status() != Won {
			t.Fatalf("seed %d: status = %v, want won", seed, b.Status())
		}
		if b.Cell(1, 1).Adjacent != 8 {
			t.Fatalf("seed %d: center adjacent = %d, want 8", seed, b.Cell(1, 1).Adjacent)
		}
	}
}

func TestFloodFillAndWin(t *testing.T) {
	b := newWithMines(t, 5, 5, Pos{0, 0})
	b.Open(4, 4)
	if b.RevealedCount() != 24 {
		t.Fatalf("revealed = %d, want 24", b.RevealedCount())
	}
	if b.Status() != Won {
		t.Fatalf("status = %v, want won", b.Status())
	}
	if b.Cell(0, 0).State != Flagged || b.Flags() != 1 {
		t.Fatalf("mines should be auto-flagged on win")
	}
	if b.Progress() != 1 {
		t.Fatalf("progress = %v, want 1", b.Progress())
	}
}

func TestFloodFillStopsAtFlags(t *testing.T) {
	b := newWithMines(t, 5, 5, Pos{0, 0})
	b.ToggleFlag(4, 0)
	b.Open(4, 4)
	if b.Cell(4, 0).State != Flagged {
		t.Fatalf("flagged cell should not be revealed by flood fill")
	}
	if b.Status() != Playing {
		t.Fatalf("status = %v, want playing", b.Status())
	}
}

func TestOpenMineLoses(t *testing.T) {
	b := newWithMines(t, 4, 4, Pos{2, 2})
	b.Open(2, 2)
	if b.Status() != Lost {
		t.Fatalf("status = %v, want lost", b.Status())
	}
	if p, ok := b.Exploded(); !ok || p != (Pos{2, 2}) {
		t.Fatalf("exploded = %v %v, want (2,2)", p, ok)
	}
	if b.Open(0, 0) || b.ToggleFlag(0, 0) {
		t.Fatalf("no action should be possible after game over")
	}
}

func TestChord(t *testing.T) {
	t.Run("correct flag reveals neighbours", func(t *testing.T) {
		b := newWithMines(t, 3, 3, Pos{0, 0})
		b.Open(1, 1)
		if b.RevealedCount() != 1 {
			t.Fatalf("revealed = %d, want 1", b.RevealedCount())
		}
		b.ToggleFlag(0, 0)
		if !b.Open(1, 1) { // Open 在數字格上等同 Chord
			t.Fatalf("chord should change the board")
		}
		if b.Status() != Won {
			t.Fatalf("status = %v, want won", b.Status())
		}
	})

	t.Run("not enough flags does nothing", func(t *testing.T) {
		b := newWithMines(t, 3, 3, Pos{0, 0})
		b.Open(1, 1)
		if b.Chord(1, 1) {
			t.Fatalf("chord without flags should do nothing")
		}
	})

	t.Run("wrong flag explodes", func(t *testing.T) {
		b := newWithMines(t, 3, 3, Pos{0, 0})
		b.Open(1, 1)
		b.ToggleFlag(0, 1)
		b.Chord(1, 1)
		if b.Status() != Lost {
			t.Fatalf("status = %v, want lost", b.Status())
		}
		if p, _ := b.Exploded(); p != (Pos{0, 0}) {
			t.Fatalf("exploded = %v, want (0,0)", p)
		}
	})
}

func TestToggleFlag(t *testing.T) {
	b := newWithMines(t, 4, 4, Pos{0, 0}, Pos{3, 3})
	if !b.ToggleFlag(1, 1) || b.Flags() != 1 || b.RemainingMines() != 1 {
		t.Fatalf("flag should be placed")
	}
	if b.Open(1, 1) {
		t.Fatalf("flagged cell must not open")
	}
	if !b.ToggleFlag(1, 1) || b.Flags() != 0 {
		t.Fatalf("flag should be removed")
	}
	b.Open(0, 1)
	if b.ToggleFlag(0, 1) {
		t.Fatalf("revealed cell cannot be flagged")
	}
}
