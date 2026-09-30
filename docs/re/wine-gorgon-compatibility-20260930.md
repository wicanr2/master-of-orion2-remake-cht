# `wine-gorgon` 對《銀河霸主 II》1.31 的相容性收據

日期：2026-09-30。範圍：執行檔格式與載入器入口；**尚未形成原版與 remake 的玩法或畫面對拍**。

## 輸入與工具

| 項目 | 定位與 SHA-256 |
|---|---|
| 官方 1.31 ZIP | `moo2_patch1.31/MOO2-1.31.en.zip`；`908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5` |
| DOS 主程式 | ZIP 內 `ORION2.EXE`；`4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |
| Win95 主程式 | ZIP 內 `ORION95.EXE`；`6e19afdc98f1aedcb8d2f974d5b658b0c855f54529bdabdde193f5266e275185` |
| 工具 | `/home/anr2/cht/wine-gorgon` 的獨立複本，Git `b3e7037cd253d7318242ab5e14db9e6140282823`，分支 `codex/moo2-parity-20260930`；`cmd/neinfo`，Go 1.25.12，容器 `dsds-go:1.25` |

原版 ZIP 唯讀；只將上述兩個 EXE 解到被 Git 忽略的 `workplace/oracle-input/`。`wine-gorgon` 複本位於被 Git 忽略的 `workplace/wine-gorgon/`。兩者均由目前使用者擁有，不屬公開交付物。

## 可重現探針

在 `alpine:3.22` 一次性 Docker 容器內執行 `unzip -p ... ORION95.EXE` 與 `unzip -p ... ORION2.EXE`，核對上述兩個 EXE 的 SHA-256；再於 `dsds-go:1.25` 一次性容器以 `GOPROXY=off` 執行：

```text
go run ./cmd/neinfo /repo/workplace/oracle-input/ORION95-1.31.EXE
ne: 位移 0x80 不是 NE 簽章（讀到 "PE"）
exit status 1

go run ./cmd/neinfo /repo/workplace/oracle-input/ORION2-1.31.EXE
ne: 位移 0x0 不是 NE 簽章（讀到 "MZ"）
exit status 1
```

**已證實**：`wine-gorgon` 目前的 `internal/ne` 載入器拒絕兩個 1.31 執行檔；Win95 檔在 DOS 標頭指向的 `0x80` 讀到 PE 簽章，DOS 檔的 NE 檢查落在 `0x0` 並讀到 MZ。位址基準均為檔案偏移，非 IDA 線性位址。工具原有的 CPU／系統層面向 Win16／NE；上述結果只證明載入器不相容，沒有驗證任何遊戲規則。

**待決定**：若堅持以 `wine-gorgon` 執行原版，就需要新增 PE32／Win32 或 DOS／DOS4GW 執行層，屬工具架構擴充。若目標是先取得可重播的原版玩家路徑與畫面收據，需改由支援對應格式的原版執行器產生，再與 remake 同狀態比較；選用前須先核對該執行器的 MOO2 能力。不得把本次格式探針寫成對拍通過。
