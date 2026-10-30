<p align="center">
  <h1 align="center">✹ minetui</h1>
  <p align="center">
    <strong>為終端機愛好者打造的 Vim 風格現代踩地雷</strong>
    <br />
    A modern, keyboard-driven terminal Minesweeper with Vim motions, adaptive layout & delightful animations.
  </p>
  <p align="center">
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version" /></a>
    <a href="https://github.com/shiangogo/minetui/releases"><img src="https://img.shields.io/badge/release-v0.2.3-89b4fa?style=flat-square" alt="Release" /></a>
    <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/built%20with-Bubble%20Tea-cba6f7?style=flat-square" alt="Bubble Tea" /></a>
    <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-a6e3a1?style=flat-square" alt="License" /></a>
  </p>
</p>

<p align="center">
  <img src="./assets/demo.gif" alt="minetui 實際遊玩展示" width="820" />
</p>

---

## ✨ 特色亮點 (Features)

- ⌨️ **Vim 核心操作流**：全面支援 `hjkl` 移動、數字倍數前綴（如 `5j`）、`gg` / `G`、`0` / `$`、`H` / `M` / `L`，以及 `w` / `b` 智慧跳轉至下一個未翻開格子。
- ⚡ **`:` 指令列精準跳轉**：支援 `:12 5` 直接跳轉座標、`:12 5 f` 跳轉並插旗、支援 `↑` / `↓` 歷史瀏覽。
- 🛡️ **第一步絕對安全**：初次翻開後才延遲佈雷，必空出第一格與周邊八格，開局必定大面積展開！
- 💥 **大面積翻開 (Chord & Flood Fill)**：已翻開數字格周圍旗子數相符時，按 `space` 或 `c` 一次翻開鄰格；遇 0 自動連鎖擴散。
- 📐 **自適應視窗縮放 (Adaptive UI)**：拉大或全螢幕時棋盤自動平滑放大至 2x / 3x 特大磚塊，亦支援 `+` / `-` / `z` 手動縮放。
- 🎨 **Catppuccin Mocha 配色**：精準至 0.1 秒的計時器、流暢動畫漸層進度條、窄螢幕自適應緊湊盤面。
- 🎯 **豐富難度設定**：內建初級 (9×9)、中級 (16×16)、高級 (30×16)，支援自訂寬度、高度與地雷數量。

---

## 🚀 快速安裝 (Installation)

### 方式一：使用 `go install`（推薦）

若本機已安裝 Go 工具鏈：

```sh
go install github.com/shiangogo/minetui@latest
minetui
```

### 方式二：直接從原始碼編譯（支援本機 Go 或 Docker）

專案具備**智慧雙模建置系統**，本機無 Go 環境時會自動退回使用 Docker 容器編譯：

```sh
# 複製儲存庫
git clone https://github.com/shiangogo/minetui.git
cd minetui

# 編譯（自動偵測本地 Go，無 Go 則自動使用 Docker）
make build

# 啟動遊戲
./bin/minetui
```

---

## 🎮 執行參數 (Usage)

```sh
minetui                                     # 進入互動難度選單
minetui -d hard                             # 直接開始高級（easy / medium / hard）
minetui -cols 40 -rows 20 -mines 150        # 自訂盤面尺寸與雷數
minetui -seed 42                            # 固定亂數種子（挑戰特定排局）
```

---

## ⌨️ 鍵位操作表 (Keybindings)

### 移動 (Motions)

| 按鍵 | 動作說明 |
| :--- | :--- |
| `h` `j` `k` `l` / `方向鍵` | 左、下、上、右（支援數字前綴，如 `5j` 向下 5 格） |
| `gg` / `G` | 跳至第一列 / 最後一列（支援 `5G` 跳至第 5 列） |
| `0` `^` / `$` | 跳至列首 / 列尾 |
| `H` / `M` / `L` | 游標移至頂端 / 盤面中央 / 底端 |
| `w` / `b`（`tab` / `shift+tab`） | 跳到下一個 / 上一個未翻開的格子 |

### 操作 (Actions)

| 按鍵 | 動作說明 |
| :--- | :--- |
| `space` / `enter` / `o` / `x` | 翻開格子；若在數字格上即為**大面積翻開** |
| `c` | 大面積翻開（周圍旗數 = 該格數字時一次翻開其餘鄰格） |
| `f` / `m` | 插旗標記 / 取消標記 |
| `+` / `-` | 手動放大 / 縮小畫面倍率 (1x ~ 3x) |
| `z` | 重設為**自動自適應縮放 (Auto)** |
| `r` | 以相同難度重新開局 |
| `d` | 開啟難度選擇選單 |
| `?` | 開啟 / 關閉即時快捷鍵說明面板 |
| `q` / `ctrl+c` | 離開遊戲 |

### `:` 指令模式 (Command Mode)

| 指令 | 動作說明 |
| :--- | :--- |
| `:12 5`、`:12,5` | 直接跳轉至第 12 列、第 5 行 |
| `:12` | 跳轉至第 12 列（保留當前行） |
| `:12 5 f` / `o` / `c` | 跳轉座標並立即插旗 (`f`) / 翻開 (`o`) / 大面積翻開 (`c`) |
| `:easy` `:medium` `:hard` | 切換難度開局 |
| `:custom W H M` | 自訂開局（例如 `:custom 20 12 40`） |
| `:zoom [1-3\|auto]` | 設定畫面縮放倍率 |
| `:r`、`:menu`、`:help`、`:q` | 重新開始 / 開啟選單 / 說明 / 離開 |

---

## 🛠️ 開發與建置指令 (Development)

```sh
make info          # 查看當前 Go 工具鏈模式（Local 或 Docker）
make test          # 執行單元測試
make check         # 執行完整代碼品質檢查 (tidy + fmt + vet + test)
make build         # 編譯可執行檔
make build USE_DOCKER=1  # 強制使用 Docker 容器建置
```

---

## 📄 授權條款 (License)

本專案採用 [MIT License](./LICENSE) 授權。
