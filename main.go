package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shiangogo/minetui/ui"
)

var version = "v0.2.7"

func main() {
	var (
		diff        = flag.String("d", "", "難度：easy | medium | hard（不指定則顯示選單）")
		cols        = flag.Int("cols", 0, "自訂寬度（需同時指定 -rows 與 -mines）")
		rows        = flag.Int("rows", 0, "自訂高度")
		mines       = flag.Int("mines", 0, "自訂地雷數")
		seed        = flag.Int64("seed", 0, "亂數種子，0 表示隨機")
		showVersion = flag.Bool("v", false, "顯示版本資訊")
	)
	flag.BoolVar(showVersion, "version", false, "顯示版本資訊")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法：%s [選項]\n\n終端機踩地雷，進入遊戲後按 ? 查看操作說明。\n\n選項：\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("minetui %s\n", version)
		return
	}

	opts := ui.Options{Seed: *seed}
	switch {
	case *cols != 0 || *rows != 0 || *mines != 0:
		d := ui.Custom(*cols, *rows, *mines)
		if err := d.Validate(); err != nil {
			fail("自訂難度不合法：%v", err)
		}
		opts.Difficulty = &d
	case *diff != "":
		d, ok := ui.PresetByName(*diff)
		if !ok {
			fail("未知的難度 %q（可用 easy / medium / hard）", *diff)
		}
		opts.Difficulty = &d
	}

	if _, err := tea.NewProgram(ui.New(opts), tea.WithAltScreen()).Run(); err != nil {
		fail("%v", err)
	}
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
