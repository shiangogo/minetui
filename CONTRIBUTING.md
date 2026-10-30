# 開發者指南 (Contributing & Development)

感謝你對 `minetui` 的關注！這份文件記載專案內部的開發與維護流程。

## 🛠️ 本地開發指令

專案 `Makefile` 具備自動雙模偵測：

```sh
make info          # 查看當前 Go 工具鏈（Local 或 Docker）
make test          # 執行單元測試
make check         # 執行代碼品質檢查 (tidy + fmt + vet + test)
make build         # 編譯本地二進位檔 (bin/minetui)
make dist          # 交叉編譯所有平台發布包與校驗檔 (dist/)
make clean         # 清除編譯產物
```

## 🎥 錄製遊玩動態預覽圖 (Demo GIF)

專案使用 [Charm VHS](https://github.com/charmbracelet/vhs) 進行無頭虛擬終端錄製：

```sh
# 1. 安裝 VHS
brew install vhs
# 或使用 Go：go install github.com/charmbracelet/vhs@latest

# 2. 執行錄製（會自動讀取 demo.tape 並輸出至 assets/demo.gif）
vhs demo.tape
```

## 🌿 Git 分支策略 (Git Flow)

- `main`：僅存放正式發布版本，每一個版本均對應一個附註標籤（如 `v0.2.1`）。
- `develop`：日常開發主分支。
- `feature/*`：新功能分支，完成後以 `--no-ff` 合併回 `develop`。
- `release/*`：發版準備分支，合併進 `main` 打標籤後，再合併回 `develop`。
