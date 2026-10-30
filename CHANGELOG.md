# Changelog

## [0.2.4] - 2026-10-31

### Fixed

- 回滾單元格強制鎖定尺寸與全形字元改動，全面恢復視窗動態自適應縮放機制（Responsive Auto-scaling）
- 修復中級與高級難度在一般終端機高度下無法自適應放大之問題
- 恢復單元格寬度隨視窗寬度即時調整（1x: 2~5 欄、2x: 4~6 欄、3x: 6~8 欄），使棋盤完整利用可用視野

### Changed

- 使用本機原生 VHS 環境重新錄製 `assets/demo.gif`，避免容器缺少字型導致的缺字與符號異常

## [0.2.3] - 2026-10-30

### Fixed

- 調整單元格渲染尺寸為視覺正方形比例（`cw = 2 * ch`）並採用全形數字字元，徹底解決數字格偏心問題
- 鎖定單元格高度為奇數（1x: 2×1, 2x: 6×3, 3x: 10×5），使單元格文字於終端機內達到水平與垂直絕對幾何置中

### Changed

- 重新錄製 Catppuccin Mocha 主題之高畫質展示動圖 `assets/demo.gif`，完整呈現正方形單元格與置中效果

## [0.2.2] - 2026-10-30

### Fixed

- 修復 `make dist` 於 Linux 環境下 SHA-256 校驗指令之相容性 (`sha256sum` / `shasum`)
- 自動化觸發 GitHub Actions 發布並生成 Release 頁面

## [0.2.1] - 2026-10-30

### Added

- 加入 GitHub Actions 自動化發布工作流程 (`.github/workflows/release.yml`)
- `Makefile` 新增 `make dist` 跨平台發布目標，支援一鍵編譯 macOS (ARM64/Intel)、Linux (x86_64/ARM64)、Windows 發布包與 SHA-256 校驗檔
- `minetui` 新增 `-v` / `-version` 命令列參數
- 納入正式遊玩動態預覽圖 `assets/demo.gif`

## [0.2.0] - 2026-10-30

### Changed

- 專案正式命名為 **`minetui`**
- Go 模組更名為 `github.com/shiangogo/minetui`
- 編譯輸出可執行檔更名為 `bin/minetui`
- 更新介面標頭與選單標題為 `✹ MINETUI`

### Added

- 新增 Charm [VHS](https://github.com/charmbracelet/vhs) 錄製腳本 `demo.tape`，支援一鍵產生 Catppuccin Mocha 主題的高畫質展示動圖 `assets/demo.gif`
- 全新高質感開源專案 README（含動態展示圖示、狀態徽章、完整 Vim 鍵位指南、快速安裝說明）
- 新增 MIT License

## [0.1.3] - 2026-10-30

### Added

- Makefile 智慧雙模支援：自動偵測本地是否有 Go 工具鏈
  - 本地有 Go 則直接原生編譯與測試，大幅提升速度且免去拉取 Docker image 的等待
  - 本地無 Go 則自動退回使用 `golang:1.25-alpine` Docker 容器執行
  - 支援手動參數 `USE_DOCKER=1` 強制切換
  - 新增 `make info` 指令查看當前 Go 環境狀態

## [0.1.2] - 2026-10-30

### Changed

- 更新 Go module 路徑為 `github.com/shiangogo/minesweeper`
- 更新專案 Git author / committer 身分為 `shiangogo`

## [0.1.1] - 2026-10-30

### Added

- 終端機視窗自適應縮放（Responsive Auto-scaling）：隨終端機視窗拉大或全螢幕，單元格與棋盤自動切換 1x / 2x / 3x 多行多欄大小
- 手動畫面縮放控制：`+`/`-` 放大縮小、`z` 重設為自動，或使用 `:zoom [1-3|auto]` 指令
- 狀態列即時顯示縮放倍率或單元格尺寸，空訊息行維持高度防抖動

## [0.1.0] - 2026-10-30

### Added

- 踩地雷核心引擎：延遲佈雷保證第一步安全、flood fill 展開、chord 大面積翻開、插旗與勝負判定
- Bubble Tea 終端機介面
  - Vim 風格移動（`hjkl`、數字前綴、`gg`/`G`、`0`/`$`、`H`/`M`/`L`、`w`/`b`）
  - `:` 指令列：座標跳轉、跳轉後動作、切換難度、自訂難度、歷史紀錄
  - 難度選單（初級 / 中級 / 高級 / 自訂）與命令列參數
  - 漸層動畫進度條、0.1 秒精度計時器、剩餘地雷數
  - 說明面板、狀態列、終端機尺寸不足提示與緊湊盤面
- 以 Docker 執行的 Makefile 開發流程
