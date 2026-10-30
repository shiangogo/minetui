package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var customLabels = [3]string{"寬度 W", "高度 H", "地雷 M"}

func (m Model) menuItemCount() int { return len(Presets) + 1 }

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.customOpen {
		return m.updateCustom(msg)
	}

	n := m.menuItemCount()
	switch key := msg.String(); key {
	case "j", "down", "tab":
		m.menuIdx = (m.menuIdx + 1) % n
	case "k", "up", "shift+tab":
		m.menuIdx = (m.menuIdx - 1 + n) % n
	case "g", "home":
		m.menuIdx = 0
	case "G", "end":
		m.menuIdx = n - 1
	case "1", "2", "3", "4":
		m.menuIdx = int(key[0]-'1') % n
		return m.selectMenu()
	case "enter", " ", "l", "right":
		return m.selectMenu()
	case "esc":
		if m.board != nil { // 有進行中的對局就回去
			m.screen = screenGame
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) selectMenu() (tea.Model, tea.Cmd) {
	if m.menuIdx < len(Presets) {
		cmd := m.newGame(Presets[m.menuIdx])
		return m, cmd
	}
	m.customOpen = true
	m.menuErr = ""
	return m, m.focusCustom(0)
}

func (m *Model) focusCustom(i int) tea.Cmd {
	m.customFocus = i
	for j := range m.customIn {
		m.customIn[j].Blur()
	}
	m.customIn[i].CursorEnd()
	return m.customIn[i].Focus()
}

func (m Model) updateCustom(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.customOpen = false
		m.menuErr = ""
		for j := range m.customIn {
			m.customIn[j].Blur()
		}
		return m, nil
	case "tab", "down", "j":
		return m, m.focusCustom((m.customFocus + 1) % len(m.customIn))
	case "shift+tab", "up", "k":
		return m, m.focusCustom((m.customFocus - 1 + len(m.customIn)) % len(m.customIn))
	case "enter":
		d, err := m.customDifficulty()
		if err != nil {
			m.menuErr = err.Error()
			return m, nil
		}
		m.customOpen = false
		m.menuErr = ""
		cmd := m.newGame(d)
		return m, cmd
	}

	// 只允許輸入數字
	if msg.Type == tea.KeyRunes {
		digits := make([]rune, 0, len(msg.Runes))
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				digits = append(digits, r)
			}
		}
		if len(digits) == 0 {
			return m, nil
		}
		msg.Runes = digits
	}

	var cmd tea.Cmd
	m.customIn[m.customFocus], cmd = m.customIn[m.customFocus].Update(msg)
	m.menuErr = ""
	return m, cmd
}

func (m Model) customValues() (w, h, mines int, ok bool) {
	var vals [3]int
	for i, ti := range m.customIn {
		v, err := strconv.Atoi(strings.TrimSpace(ti.Value()))
		if err != nil {
			return 0, 0, 0, false
		}
		vals[i] = v
	}
	return vals[0], vals[1], vals[2], true
}

func (m Model) customDifficulty() (Difficulty, error) {
	w, h, mines, ok := m.customValues()
	if !ok {
		return Difficulty{}, fmt.Errorf("請填入三個數字")
	}
	d := Custom(w, h, mines)
	return d, d.Validate()
}

func (m Model) viewMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("✹  M I N E T U I  ✹"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("選擇難度"))
	b.WriteString("\n\n")

	for i, d := range Presets {
		line := fmt.Sprintf("%s  %2d × %-2d   %3d 顆地雷", d.Name, d.Width, d.Height, d.Mines)
		b.WriteString(m.menuLine(i, line))
		b.WriteString("\n")
	}
	b.WriteString(m.menuLine(len(Presets), "自訂  自行設定寬、高與地雷數"))
	b.WriteString("\n")

	if m.customOpen {
		b.WriteString("\n")
		for i, ti := range m.customIn {
			label := menuDimStyle.Render(customLabels[i])
			marker := "  "
			if i == m.customFocus {
				label = menuSelectedStyle.Render(customLabels[i])
				marker = menuSelectedStyle.Render("› ")
			}
			fmt.Fprintf(&b, "    %s%s  %s\n", marker, label, ti.View())
		}
		if w, h, mines, ok := m.customValues(); ok && w > 0 && h > 0 {
			density := float64(mines) / float64(w*h) * 100
			b.WriteString(menuDimStyle.Render(fmt.Sprintf("\n    地雷密度 %.1f%%  （寬 %d~%d、高 %d~%d）",
				density, MinSize, MaxWidth, MinSize, MaxHeight)))
			b.WriteString("\n")
		}
	}

	if m.menuErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  " + m.menuErr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.customOpen {
		b.WriteString(renderHints([][2]string{
			{"tab/j/k", "切換欄位"}, {"enter", "開始"}, {"esc", "返回"},
		}))
	} else {
		hints := [][2]string{{"j/k", "選擇"}, {"enter", "開始"}, {"1-4", "快速選擇"}, {"q", "離開"}}
		if m.board != nil {
			hints = append(hints, [2]string{"esc", "返回對局"})
		}
		b.WriteString(renderHints(hints))
	}

	return m.place(menuBoxStyle.Render(b.String()))
}

func (m Model) menuLine(i int, text string) string {
	if i == m.menuIdx {
		return menuSelectedStyle.Render("  ▸ " + text)
	}
	return menuItemStyle.Render("    " + text)
}

// place 把內容置中於終端機視窗。
func (m Model) place(content string) string {
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func renderHints(pairs [][2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = helpKeyStyle.Render(p[0]) + " " + helpDescStyle.Render(p[1])
	}
	return strings.Join(parts, helpDescStyle.Render(" · "))
}
