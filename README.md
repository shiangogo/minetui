# minesweeper

以 Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) 打造、可在終端機遊玩的踩地雷，操作方式向 Vim 看齊。

## 特色

- **Vim 風格移動**：`hjkl`、數字前綴（`5j`）、`gg` / `G`、`0` / `$`、`w` / `b` 跳到下一個未翻開的格子
- **`:` 指令列**：`:12 5` 直接跳到第 12 列第 5 行，`:12 5 f` 跳過去順便插旗，支援 ↑↓ 歷史紀錄
- **插旗標記**與**大面積翻開**（在數字格上翻開，周圍旗數等於數字時一次翻開其餘鄰格；0 格自動擴散）
- **第一步絕對安全**：地雷在第一步之後才佈置，並避開第一格與周圍八格，保證一開場就展開一片
- **難度設定**：初級 / 中級 / 高級 / 自訂，可從選單、`:` 指令或命令列參數切換
- **漸層進度條**（已翻開安全格比例，含動畫）與精確到 0.1 秒的**計時器**
- 終端機太窄時自動切換成緊湊盤面

## 開發環境（只需要 Docker）

所有 Go 工具鏈都跑在 `golang:1.25-alpine` 容器裡，本機不需要安裝 Go。
模組與編譯快取存在 Docker volume（`minesweeper-gomod`、`minesweeper-gobuild`）。

```sh
make build         # 交叉編譯出本機可執行檔 bin/minesweeper
make run           # 編譯並執行
make check         # go mod tidy + gofmt 檢查 + go vet + go test
make test          # 只跑測試
make clean         # 刪除 bin/
make clean-docker  # 刪除 Go 模組 / 編譯快取 volume
```

> 遊戲本身需要真正的 TTY，因此是在容器中編譯、在本機終端機執行。

## 執行

```sh
./bin/minesweeper               # 顯示難度選單
./bin/minesweeper -d hard       # 直接開始高級
./bin/minesweeper -cols 40 -rows 20 -mines 150   # 自訂難度
./bin/minesweeper -seed 42      # 固定亂數種子
```

## 操作說明

座標以 **列（row，橫向）/ 行（column，縱向）** 表示，從 1 開始，盤面四周有標示。

### 移動

| 按鍵 | 動作 |
| --- | --- |
| `h` `j` `k` `l` / 方向鍵 | 左 / 下 / 上 / 右，可加數字前綴如 `5j` |
| `gg` / `G` | 第一列 / 最後一列（`5gg`、`5G` 跳到第 5 列） |
| `0` `^` / `$` | 列首 / 列尾 |
| `H` / `M` / `L` | 頂端 / 盤面中央 / 底端 |
| `w` / `b`（`tab` / `shift+tab`） | 下一個 / 上一個未翻開的格子 |

### 操作

| 按鍵 | 動作 |
| --- | --- |
| `space` `enter` `o` `x` | 翻開；在數字格上即大面積翻開 |
| `c` | 大面積翻開 |
| `f` / `m` | 插旗 / 拔旗 |
| `r` | 以相同難度重新開始 |
| `d` | 難度選單 |
| `?` | 說明面板 |
| `q` / `ctrl+c` | 離開 |

### `:` 指令

| 指令 | 動作 |
| --- | --- |
| `:12 5`、`:12,5` | 跳到第 12 列、第 5 行 |
| `:12` | 跳到第 12 列（保留目前的行） |
| `:12 5 f` / `o` / `c` | 跳過去並插旗 / 翻開 / 大面積翻開 |
| `:easy` `:medium` `:hard`、`:diff <名稱>` | 切換難度 |
| `:custom W H M` | 自訂寬、高與地雷數 |
| `:r`、`:menu`、`:help`、`:q` | 重開 / 選單 / 說明 / 離開 |

指令列中 `↑` / `↓` 瀏覽歷史、`ctrl+u` 清空、`esc` 取消。

## 專案結構

```
.
├── main.go          # 命令列參數與程式進入點
├── game/            # 與介面無關的踩地雷核心規則（含單元測試）
└── ui/              # Bubble Tea model：選單、輸入處理、畫面繪製
```

## 分支策略

採用 git flow：`main` 只放正式版本（以 `vX.Y.Z` 標籤標記），日常開發在 `develop`，
新功能從 `develop` 開 `feature/*` 分支、完成後以 `--no-ff` 合併回 `develop`，
發版時開 `release/*` 分支，合併進 `main` 並打標籤後再合併回 `develop`。
