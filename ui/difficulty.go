package ui

import (
	"fmt"
	"strings"
)

// 自訂難度的尺寸上限（避免盤面大到終端機放不下）。
const (
	MinSize   = 5
	MaxWidth  = 60
	MaxHeight = 40
)

// Difficulty 描述一種盤面設定。
type Difficulty struct {
	Key    string // easy / medium / hard / custom
	Name   string
	Width  int
	Height int
	Mines  int
}

// Presets 是經典的三種難度。
var Presets = []Difficulty{
	{Key: "easy", Name: "初級", Width: 9, Height: 9, Mines: 10},
	{Key: "medium", Name: "中級", Width: 16, Height: 16, Mines: 40},
	{Key: "hard", Name: "高級", Width: 30, Height: 16, Mines: 99},
}

// PresetByName 以名稱或別名查詢預設難度。
func PresetByName(s string) (Difficulty, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "easy", "e", "beginner", "1", "初級":
		return Presets[0], true
	case "medium", "m", "normal", "intermediate", "2", "中級":
		return Presets[1], true
	case "hard", "h", "expert", "3", "高級":
		return Presets[2], true
	}
	return Difficulty{}, false
}

// Custom 建立自訂難度。
func Custom(width, height, mines int) Difficulty {
	return Difficulty{Key: "custom", Name: "自訂", Width: width, Height: height, Mines: mines}
}

// Validate 檢查難度設定是否合法。
func (d Difficulty) Validate() error {
	if d.Width < MinSize || d.Width > MaxWidth {
		return fmt.Errorf("寬度需介於 %d ~ %d", MinSize, MaxWidth)
	}
	if d.Height < MinSize || d.Height > MaxHeight {
		return fmt.Errorf("高度需介於 %d ~ %d", MinSize, MaxHeight)
	}
	if max := d.Width*d.Height - 1; d.Mines < 1 || d.Mines > max {
		return fmt.Errorf("地雷數需介於 1 ~ %d", max)
	}
	return nil
}

// Summary 回傳像「中級 16×16 · 40 顆地雷」的描述。
func (d Difficulty) Summary() string {
	return fmt.Sprintf("%s %d×%d · %d 顆地雷", d.Name, d.Width, d.Height, d.Mines)
}
