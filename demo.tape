# VHS 踩地雷動態展示錄製腳本
# 執行方式：vhs demo.tape（需先安裝 VHS: brew install vhs 或 go install github.com/charmbracelet/vhs@latest）

Output assets/demo.gif

Set FontSize 15
Set Width 1100
Set Height 860
Set Padding 20
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set TypingSpeed 100ms

# 啟動初級踩地雷
Type "./bin/minetui -d easy"
Enter
Sleep 1.2s

# 第一步保證安全翻開一大片！
Type " "
Sleep 1.2s

# 以 hjkl 移動游標
Type "h"
Sleep 250ms
Type "j"
Sleep 250ms
Type "2l"
Sleep 300ms

# 插旗標記地雷 (f)
Type "f"
Sleep 800ms

# 移動並再次插旗
Type "k"
Sleep 250ms
Type "h"
Sleep 250ms
Type "f"
Sleep 800ms

# 使用 : 指令快速跳轉座標
Type ":"
Sleep 400ms
Type "2 3"
Sleep 600ms
Enter
Sleep 800ms

# 示範畫面縮放 (+ / - / z：皆維持 2:1 正方形且奇數高度垂直置中)
Type "+"
Sleep 1s
Type "-"
Sleep 800ms
Type "-"
Sleep 1s
Type "z"
Sleep 800ms

# 開啟說明面板 (?)
Type "?"
Sleep 2s

# 關閉說明
Type "?"
Sleep 800ms

# 離開 (q)
Type "q"
Sleep 500ms
