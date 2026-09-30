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
- 已證實：格式解析與明列 CPU 指令子集；假說：DOS/4GW 外層啟動脈絡可能解釋下一停點。未知：首個玩家可見檢查點與 remake 同狀態結果。本輪未宣稱玩法對拍通過，也未修改 Go／Ebitengine 玩法。
- 以 1996 光碟版 `Orion2.exe` 重跑相同探針：LE 偏移、入口 bytes 與第 5,359 步的 `POP EDX` 停點均一致；排除「僅 1.31 patch 導致」的解釋。隔離 `dosgolem` 副本已在本機分支提交 `5d3c806d4ba1eec9cd4376de44edbc26c9d34552`，未修改來源倉庫。
- 用 IDA Pro 9.4 查私有 `.i64` 的拋棄式副本：資料庫輸入 MD5 與 1996 光碟 EXE 相符，但 IDA `0x10FF18` 是 `EB 76 ...`，`dosgolem` 同數值的 LE 線性位址是 `66 3B ...`。兩種位址基準不能混用；詳細來源及雜湊補在同一份接入收據。

## 2026-09-30：內嵌 MZ 的 LE 資料頁基址勘誤

- 原檔 `0x26654` 的第二層 MZ 與 `e_lfanew=0x2C90` 證實 LE 標頭位於 `0x292E4`；`DataPagesOffset` 必須加上第二層 MZ 基址。上一節的 `66 3B`、5,359 步、`POP EDX` 與「缺少外層堆疊」均由錯頁造成，已撤回其原版啟動結論；舊診斷保留以追溯錯誤來源。
- `dosgolem` 分支依 `docs/spec/197-bound-mz-le-file-base.md` 新增明示內嵌 MZ 載入；兩版未修改原版的 object 1 入口均是 `EB 76 WATCOM`，與 IDA 1996 版原檔 `0x1955AC` 映射一致。`go test ./...`、兩版固定雜湊入口測試及 `leprobe -mz-base 0x26654` 均通過。
- 兩版從真入口自然執行第 11 步，皆在 LE 線性位址 `0x10FFB5` 要求 `INT 21h/AH=30h`（`EAX=0x3000`、`EBX=0x50484152`），目前探針未接 MOO2 DOS 服務；未到玩家畫面，也沒有玩法同狀態對拍。詳細收據與新物件雜湊見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 另用 FD2 專屬服務層作拋棄式診斷，據 `66 09 CA` 補通用 CPU 16 位 OR；診斷至第 45 步停於 `66 26 8C 1D`。因 FD2 selector／環境不能冒充 MOO2，此步數只列工具缺口，不取代正式原版收據。

## 2026-09-30：MOO2 入口段覆寫的通用 CPU 補強

- `dosgolem` 隔離分支依規格 199 補 ES 覆寫下的 `MOV r/m16,Sreg` 絕對位址寫入與 DS 前綴的 `MOV EDX,imm32`；合成測試核對描述子基址、word 界限、立即數與旗標，`go test ./internal/cpu386 ./internal/machine` 通過。
- 1.31／1996 兩版原檔在隔離 FD2 服務診斷中由第 45 步前進至第 65 步 `0x1100FA` 的未支援 `28 C0`。這不是 MOO2 DOS 服務或玩家可見對拍的收據；正式無服務原版收據仍只到第 11 步 `INT 21h/AH=30h`。證據與限制見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- `go test ./...` 全套通過；本輪檔案 UID/GID 均為 `1000:1000`。Docker `ps -a` 無 MOO2／dosgolem 殘留容器；其餘執行中容器屬其他專案，未碰觸。衛生檢查另發現並僅移除已確認為空的舊 root-owned `knowledge-router.md/` 目錄。
- `dosgolem` 再增規格 200 的 `NewMOO2StartupDOS` 暫定入口與合成 `ORION2.EXE` 環境；兩版原檔均經該入口走過兩次服務、仍於第 65 步的通用 CPU `28 C0` 停住。DOS/4G selector 與回傳值尚借用 FD2 平台近似，故這仍不是正式 MOO2 服務或玩法對拍；FD2 舊測試維持通過。
- `DOSGOLEM_MOO2_EXE=/input/ORION2-1.31.EXE go test ./...` 全套通過，1996 版固定雜湊 `TestMOO2EmbeddedMZEntryWhenProvided` 也通過。暫定服務入口與規格 200 已提交 `cc6231bd7f6578d379e1faf31ae4f1214bf917d3` 並推送；MOO2 專案僅更新證據與待辦，沒有改玩法程式。
