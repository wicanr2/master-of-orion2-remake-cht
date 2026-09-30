# 工作歷程

## 2026-09-30：合併代理規範與 `wine-gorgon` 相容性探針

- 以 `384e655782b51ae264fe3882dacbf5823b01fed6` 為起點，在 `codex/moo2-wine-gorgon-parity-20260930` 分支合併 `CLAUDE.md` 的專案目標、來源與工作規範到 `AGENTS.md`，刪除舊檔並修正現行文件入口；歷史 RE 日誌仍保留當時檔名。
- 複製 `/home/anr2/cht/wine-gorgon` 到 `workplace/wine-gorgon/`，在複本建立 `codex/moo2-parity-20260930`，原始倉庫未修改。原版 1.31 執行檔解到 `workplace/oracle-input/`，兩個工作區均不納入 Git。
- Docker 內的 `unzip -p` 與 `sha256sum` 通過；`dsds-go:1.25` 內 `go run ./cmd/neinfo` 對兩個 EXE 均以格式不符結束。首輪工具容器沒有 `unzip`，未產生有效輸入；改由 `alpine:3.22` 解壓後重跑。詳細輸入雜湊、工具版本、位址基準與輸出見 [`docs/re/wine-gorgon-compatibility-20260930.md`](docs/re/wine-gorgon-compatibility-20260930.md)。
- 已驗證事實：目前 `wine-gorgon` 不能直接載入這兩個 1.31 EXE。未知：其他原版執行器對本機 MOO2 的正常玩家路徑、固定狀態和畫面擷取支援程度。本輪沒有原版與 remake 的同狀態玩法收據，也未更動 Go 玩法。
- 待決定最小下一步：選定原版執行器路線後，用同一原版輸入與一個明確玩家檢查點製作第一份對拍收據。

## 2026-09-30：採用 `dosgolem` 的 DOS 原版路線

- 使用者選定 DOS 版對拍；保留前次 `wine-gorgon` 相容性證據，但不擴充其 PE32 執行層。以 `dosgolem` Git `cc1ef5611b5eb288cc489ae504f0a7a96fc526bf` 複製隔離副本並開 `codex/moo2-parity-20260930` 分支，來源倉庫未修改。
- 在隔離副本依 READY 規格補明示 LE 偏移入口與 MOO2 首個 386 CMP 指令形狀；`go test ./internal/machine ./cmd/leprobe`、`go test ./internal/cpu386 ./internal/machine` 及 `go test ./...` 均通過。用未改動原版 EXE 重生格式收據，並自然執行 5,359 步；詳細 bytes、雜湊、位址基準與停止原因見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 已證實：格式解析與明列 CPU 指令子集；強推論：下一阻塞是 DOS/4GW 外層啟動脈絡。未知：首個玩家可見檢查點與 remake 同狀態結果。本輪未宣稱玩法對拍通過，也未修改 Go／Ebitengine 玩法。
- 以 1996 光碟版 `Orion2.exe` 重跑相同探針：LE 偏移、入口 bytes 與第 5,359 步的 `POP EDX` 停點均一致；排除「僅 1.31 patch 導致」的解釋。隔離 `dosgolem` 副本已在本機分支提交 `5d3c806d4ba1eec9cd4376de44edbc26c9d34552`，未修改來源倉庫。
