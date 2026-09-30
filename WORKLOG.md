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
- 使用既有 DOSBox-X 除錯映像與固定 1.31 原檔，在自身 `0180:00333FB5／00333FB7` 擷取 `AH=30h` 前後，在 `0180:00334054／00334056` 擷取 `AX=FF00h` 前後；第一個返回 `EAX=5`、`EBX=5048FF00`、`DS=SS=0188`，直接推翻先前 FD2 `1606／0160` 近似。第二個返回 `EAX=4734FFFF`、CF 清除。終端原始資料與命令留在未版控 `workplace/dosbox-moo2/`，精確位址基準與限制見既有接入收據的同日勘誤；這是 DOSBox-X 輔助基準，還不是 dosgolem 正式對拍。
- `dosgolem` 依規格 201 修正兩次 MOO2 啟動回傳，再依規格 202 增量支援第 65 步的通用 `SUB r/m8,r8` 暫存器形狀。`go test ./internal/cpu386 ./internal/machine` 通過；固定 1.31 原檔隔離診斷移至第 72 步 `0x110102` 的 ES 覆寫段載入缺口。這仍只有工具診斷，沒有改 remake 玩法。
- DOSBox-X 輔助回放腳本與設定已納入 `dosgolem/apps/moo2/tools/`，再由該版控腳本重生兩次呼叫前後的 `startup-registers.json`，SHA-256 `341ce45f695a350958a74c7a219b2bcf621ff1d86ac3cfb390d0726356da0a44`。原始終端、EXE 與 JSON 輸出留在本機，使用方式與證據限制見規格 201 及接入收據。
- 收尾驗證：`DOSGOLEM_MOO2_EXE=/input/ORION2-1.31.EXE go test ./...` 全通過，1996 版固定雜湊入口測試亦通過；回放腳本重生四快照成功，診斷於第 72 步依預期失敗即關閉。新寫入檔均為 `1000:1000`，已無錯掛 `.md` 目錄；`docker ps -a` 查 `golang:1.24-bookworm` 與 `fd2-dosbox-x:debug-0d7b272b` 均無殘留容器。
- `dosgolem` 的固定原版返回訂正與版控回放已提交並推送 `44e488536bbcfaf079c07779aaa22e91c4f9fe2f`；MOO2 專案保留逆向收據、現行待辦與歷程，不把 DOSBox-X 輔助快照寫成 dosgolem 正式玩家對拍。

## 2026-09-30：MOO2 啟動診斷的通用 CPU 指令續驗

- 在隔離的 `dosgolem` 分支依規格 203–206 補 ES 覆寫 16 位段載入、DS 前綴暫存器立即數、ES 覆寫 byte 比較及 `PUSH GS`。`go test ./internal/cpu386 ./internal/machine` 通過，固定 1.31 真檔有界探針由第 72 步推進到第 419 步、dosgolem LE 線性位址 `0x13CBAA`，停於未支援的 `66 83 E7 FC`。
- 這四項僅驗收通用 CPU 指令。MOO2 服務雖參考 DOSBox-X 輔助返回，PSP／環境仍是合成的；正常玩家路徑與 remake 同狀態對拍尚無收據。原始輸入雜湊、規格及各停點見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。本輪沒有改 Go／Ebitengine 玩法。
- `DOSGOLEM_MOO2_EXE=/input/ORION2-1.31.EXE go test ./...` 全通過；1996 版 `TestMOO2EmbeddedMZEntryWhenProvided` 亦通過。隔離工具分支提交為 `fde9ff7`；本輪新寫入檔均為目前使用者 UID/GID，無錯掛 `.md` 目錄，`golang:1.24-bookworm` 無殘留容器。

## 2026-09-30：診斷前進至 DOS 記憶體服務缺口

- 隔離 `dosgolem` 分支依規格 207 補通用 `66 83 /4` 的 16 位暫存器 AND；CPU／machine 測試通過。固定 1.31 真檔在合成 PSP／環境的診斷中至第 554 步、dosgolem LE 線性位址 `0x15E07C`，停在 `INT 21h/AH=4Ah`，未加入猜測服務返回。
- 嘗試用 DOSBox-X 在已核對的兩次啟動服務之後擷取該呼叫；除錯終端在顯示模式切換後未再提供可解析快照，已撤回腳本增量。此失敗不構成原版服務語意或正常玩家路徑證據；原始暫存器、位址基準與限制見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- `DOSGOLEM_MOO2_EXE=/input/ORION2-1.31.EXE go test ./...` 與 1996 版固定入口測試均通過；隔離工具提交 `d200533`。新寫入檔為目前使用者 UID/GID，無錯掛 `.md` 目錄；兩種本輪測試映像均無殘留容器。

## 2026-09-30：原版服務觀測改路與雙語固定鍵稽核

- DOSBox-X 在兩次已證實返回後設定候選直接斷點 `0180:0038207C`；顯示模式切換後沒有可解析的寄存器快照，因此無法判斷原版是否經過該位址，已撤回實驗腳本增量。`AH=4Ah` 返回保持未知，沒有讓 dosgolem 猜值繼續。
- 轉查中文化活表與程式，確認 production `cmd/moo2` 已無 `.tr(` 呼叫；為 `uiText` 的固定字串鍵新增語法樹檢查，要求 `assets/i18n/ui.json` 同時有繁中與英文欄位。首次測試因隔離容器未掛既有 Go 模組快取而失敗，改用唯讀 `.docker-cache/go`、同一 `cmd/moo2` 測試重跑通過。動態組合鍵及其他直接繪製文字仍須各畫面稽核。

## 2026-09-30：戰術「系統」欄雙語文案補漏

- 一般艦艇的六列 `HP／ARM／SHD／ATK／DEF／DRV` 原先直接由 Go 英文格式字串繪製；現改為 `ui.json` 的穩定雙語鍵，保留英文原樣，繁中顯示船體、裝甲、護盾、攻擊、防禦與引擎。這是既有 remake 欄位的文案外部化，不改戰鬥數值或規則。
- `cmd/moo2` 完整測試通過，雙語列值與欄寬有回歸測試。以唯讀本機遊戲資料、Docker＋Xvfb 重生中文畫廊並檢視 `16_tactical.png`，六列在「系統」欄內未見重疊或裁切。含原版資產的重生圖片只留在本機 `/tmp/moo2-gallery-systems-20260930/`，不加入版控；這不是 dosgolem 原版對拍收據。

## 2026-09-30：MOO2 DPMI 綁定勘誤與通用 SBB 補強

- 固定 1.31 原版的 DOSBox-X 輔助快照證實 `INT 31h/AX=0006h` 返回 `CX:DX=0`，而舊 dosgolem 私有探針漏掉 `services.AttachMachine(m)`，導致第 554 步的 `AH=4Ah` 錯誤分支。補綁後，零基底分支與原版一致，舊停點已從活表撤回；私有資料的環境字串長度仍與原版不同。
- 版控 DOSBox-X 腳本新增 DPMI 與 `19 C0` 前後快照，分別以預設及 `--sbb` 重生；兩份私有 JSON SHA-256、地址基準與輸入雜湊記於 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。原版 `19 C0` 的 `EAX=0501h → 0`、旗標不變；Intel 指令規格與合成 CF=1 測試支撐通用 32 位暫存器 SBB 實作。
- `dosgolem` 規格 208 已完成；固定原檔診斷現停於 dosgolem LE 線性 `0x15171D` 的 `87 FA`，DOSBox-X 有界記錄亦經過對應 CS:EIP。`DOSGOLEM_MOO2_EXE=/input/ORION2-1.31.EXE go test ./...` 全通過；此結果仍只屬工具啟動診斷，沒有玩家畫面或 remake 同狀態對拍。
- 隔離 `dosgolem` 分支提交並推送 `3babfcf9462e9bbf7c7a1c1d67ded0fa0b30bb6b`；MOO2 主專案只更新活表、研究收據與歷程，未改玩法程式。

## 2026-10-01：`dosgolem` 的 MOO2 通用 XCHG 指令

- 固定 1.31 原檔、既有 DOSBox-X 映像與版控 `--xchg` 探針，擷取 `0180:0037571D → 0037571F` 的輔助前後快照；私有收據 SHA-256 `69deda4f6e766c9ffba973bea26dcade18e6f6b0b5b6ff5a780036bf6c6fb18d`。兩側 EDX 與 EDI 都是 `003EC028`，因此交換方向依 Intel 規格和不同值測試確認。
- 隔離 `dosgolem` 分支依規格 209 補 `87 FA` 的通用 32 位暫存器 `XCHG`，`go test ./internal/cpu386 ./internal/machine -count=1`、`go test ./... -count=1` 通過；固定原檔、已綁定 DPMI 的合成診斷推進至第 760 步 `0x1515CA` 的 `F5` 缺口。這不是玩家路徑或同狀態對拍，也未修改 remake 玩法。
- 證據與未知界線見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)；下一個最小工具步驟是核對 `F5` 的原版前後狀態，並延續 PSP／環境消費端調查。
- 隔離 `dosgolem` 分支提交並推送 `812c2b63ac95aea5256117598b705d10105b6456`；主專案僅更新活表、研究與歷程。

## 2026-10-01：`dosgolem` 啟動診斷再通過 `F5`／`21 C8`

- 固定 1.31 原檔與既有 DOSBox-X 映像重生 `F5` 前後快照及 `21 C8` 的同次有界 `LOG 2`。後者推翻獨立 `EV` 快照的 `ECX=0Fh` 值；連續 LOG 顯示 `ECX=FFFFFFFFh`，與 `EAX=80h` 相與後不變一致。證據雜湊與位址基準見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 隔離工具分支依規格 210／211 增補通用 `CMC` 與 32 位暫存器 `AND r/m32,r32`；CPU／機器與全套 Go 測試通過。固定原檔合成診斷現停第 803 步、dosgolem 重定位 LE 線性 `0x151648` 的 `83 0E 01`。未修改 remake 玩法，尚無正常玩家畫面或同狀態對拍。
- 隔離 `dosgolem` 分支提交並推送 `af3f83a30491430ff4f8525b9bc6229b1c5e7297`；主專案只更新活表、研究收據與歷程。

## 2026-10-01：`dosgolem` 的 MOO2 記憶體 OR 指令

- 以固定 1.31 原檔及既有 DOSBox-X 映像，同次 `LOG 2` 加 DS:[ESI] 前後四位元組擷取證實 `83 0E 01` 將 `90h` 寫成 `91h`；私有輸出雜湊與工具位址基準見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 隔離 `dosgolem` 分支依規格 212 接入無前綴 32 位記憶體 `83 /1 ib`，CPU／機器與全套 Go 測試通過。固定原檔合成診斷進至第 819 步、重定位 LE 線性 `0x13CC50` 的 `0F A9` 停點。未修改 remake 玩法，尚無正常玩家畫面或同狀態對拍。
- 隔離 `dosgolem` 分支提交並推送 `d837b424bc86aa5ff043716a560da50dae5e7788`；主專案只更新活表、研究收據與歷程。

## 2026-10-01：`POP GS` 與 MOO2 啟動選擇子續驗

- 固定 1.31 原檔的 DOSBox-X 同次 `LOG 2` 與堆疊擷取核對 `0F A9` 的 32 位彈出、`ESP+4`、GS=`0020h` 及旗標不變；原版 GS 前後相同，故不同 selector 載入另依 Intel 指令契約與合成 CPU 測試。完整雜湊與兩工具位址基準見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 隔離 `dosgolem` 分支先依 READY 規格補通用 `POP GS`；固定原檔診斷仍在第 819 步拒絕 `0020h`，遂回到 DRAFT 補查，審查後只為 MOO2 profile 允許 `0020h → GS`。FD2 profile 與其他目的維持拒絕；不推測 GS 描述子屬性。`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1`、固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過；10,000 步有界診斷至第 2460 步、重定位 LE 線性 `0x153E84`，停在帶前綴 `SBB`。這仍非正常玩家路徑或玩法同狀態對拍，也未改 remake 玩法。
- 工具分支提交並推送 `ff8f8b65d558b94e672854e5ab258e9a23a5bc22` 至 `github/codex/moo2-parity-20260930`；原版檔、完整終端輸出與堆疊擷取留在未版控工作區。Docker 掛載前確認來源形態；本輪 `golang:1.24-bookworm` 與 DOSBox-X 映像無殘留容器，工作目錄檢查無 root 擁有檔及誤掛 `.md` 目錄。下一步先核對帶前綴 `SBB` 的原版指令形狀，並查合成 PSP／環境的資料消費端。

## 2026-10-01：16 位 `SBB` 的原版輔助驗證

- 固定 1.31 原檔在 DOSBox-X 的 `0180:00377E84 → 00377E87` 同次 `LOG 2` 命中 `66 19 C0`，原版 AX=`0600h → 0000h`、CF=0、ZF=`0 → 1`；新版 `--sbb-word` 探針與私有輸出 SHA-256、原版位址基準見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。Intel 契約與合成測試補 CF=1、不同暫存器、高 16 位保留，沒有把未實測組合冒稱原版結果。
- 隔離 `dosgolem` 分支依規格 214 補 `66 19 /r mod=11`，`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1`、固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 通過。已綁定 DPMI 的合成 PSP／環境診斷由第 2460 步前進至第 2475 步，停在重定位 LE 線性 `0x1005B` 的 `C8`；新低位址路徑尚未與原版獨立核對，不作下一個 CPU 規格的已證實依據。這一輪沒有改 remake 玩法，仍無正常玩家路徑或同狀態對拍。
- 隔離工具分支提交並推送 `c21627b2031c9e6f842ec3aca08224fee3bb2778`；原檔與完整終端輸出留本機。新檔 UID/GID 為 `1000:1000`，沒有 root 擁有檔或誤掛 `.md` 目錄，`golang:1.24-bookworm`、DOSBox-X 無殘留容器。下一步核對 `0x13EF5C → 0x10018 → 0x10057` 的原版控制流及 `C8` 指令位置，再處理工具缺口。

## 2026-10-01：原版低位址控制流與層級 0 `ENTER`

- 固定 1.31 原檔在 DOSBox-X 的有界 `LOG 20` 證實 `0180:00362F5C → 00234018 → 00234057 → 0023405B`；同次 `LOG 2` 加堆疊前後擷取確認 `ENTER 00AC,00` 將舊 EBP 寫到 `SS:003EBC8C`，EBP／ESP 依 4+`00AC` 位移，旗標不變。位址空間、原檔與私有收據 SHA-256 詳見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。這只證實指令與控制流，原版與合成 PSP／環境仍非同狀態。
- 隔離 `dosgolem` 分支規格 215 先 DRAFT→READY，首版實作測試越過第 2475 步；審查發現為預檢頁面映射額外讀取堆疊，可能改變通用匯流排副作用，因此退回 DRAFT、修訂為段描述子預檢，再 READY→實作→CONFORMED。修正版新增禁止額外堆疊讀取的測試，並涵蓋配置大小、堆疊往返、前綴、段界限及下溢。固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過；同一有界診斷至第 2512 步、重定位 LE 線性 `0x109FF` 的 `66 3B 4D CE` 停點。新比較指令尚未由原版獨立核對；本輪未改 remake 玩法，也沒有正常玩家路徑或玩法同狀態對拍。
- 工具分支提交並推送 `ac9a129513e4382145fc67fabe7e41b0c690fb6e` 到 `github/codex/moo2-parity-20260930`。原版檔、堆疊檔與完整終端留在未版控工作區；新檔 UID/GID `1000:1000`，沒有 root 擁有檔或誤掛 `.md` 目錄，`golang:1.24-bookworm` 與 DOSBox-X 無殘留容器。下一步先核對 `0x109FF` 的原版記憶體讀值與旗標，再判斷是否建立新規格。

## 2026-10-01：原版 `CMP` 記憶體來源與 dosgolem 規格 216

- 固定 1.31 原檔在 DOSBox-X 的 `0180:002349FF → 00234A03` 同次 `LOG 2` 命中 `66 3B 4D CE`；原版 `CX=0`、`SS:[EBP-32h]=1`、旗標 `0246h → 0297h`，來源不變。初版腳本錯用 DS 擷取，依 `ss:[...]` 註記與解碼器修正為 SS 並重生；原版 DS=SS，故另以 Intel 契約與不同段基址測試驗證預設段。固定輸入、工具版本、位址基準與私有收據雜湊見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 隔離 dosgolem 分支依 READY 規格 216 補通用 16 位暫存器與 32 位址記憶體來源的 `CMP`；`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1`、固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過。已綁定 DPMI 的合成 PSP／環境診斷至第 2600 步、重定位 LE 線性 `0x126570` 的 `66 A9 89 CF` 停點；下一步先核對原版 `TEST` 的位置、輸入與旗標。沒有改 remake 玩法，仍無正常玩家路徑或玩法同狀態對拍。
- 工具分支提交並推送 `8f30a7e9b18e1fc1023378d8adfd9fa2d4fa551f` 至 `github/codex/moo2-parity-20260930`。原版檔與完整收據維持私有；新檔 UID/GID `1000:1000`，無 root 擁有檔或誤掛 `.md` 目錄，`golang:1.24-bookworm` 與 DOSBox-X 無殘留容器。

## 2026-10-01：`TEST AX,imm16` 與 DTA 呼叫邊界

- 固定 1.31 原檔在 DOSBox-X `0180:0034A570 → 0034A574` 同次 `LOG 2` 命中 `66 A9 89 CF` 對應的 `test ax,CF89`；EAX 不變、旗標 `0206h → 0286h`。依原版輔助收據與 Intel 契約完成 dosgolem 規格 217 的 DRAFT→READY→實作→CONFORMED，新增 `66 A9 iw` 及高半字、旗標、截短立即數、非法前綴測試。`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 與固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過；合成 PSP／環境診斷移至第 4062 步、重定位 LE 線性 `0x139A53` 的 `CD 21` 停點。
- DOSBox-X 原版另一有界探針命中對應 `0180:0035DA53 → 0035DA55` 的 `AH=1Ah` 呼叫及返回，暫存器與旗標未變。原版 DTA 指標 `003C3828h`、合成診斷 `001A5828h` 明顯不同；本輪只登記服務邊界，不猜補返回或接入 `AH=4Eh／4Fh`。輸入雜湊、工具版本、兩種位址空間及私有收據 SHA-256 見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。仍無 dosgolem 正常玩家路徑或玩法同狀態對拍；未改 remake 玩法。
- 隔離 dosgolem 分支提交並推送 `2e4730a` 至 `github/codex/moo2-parity-20260930`。原版與完整擷取只留本機；Docker 來源掛載前已核對、寫入者為目前使用者，相關一次性容器無殘留。下一步核對受保護模式 DTA 指標及首次搜尋消費端，再建立服務規格。

## 2026-10-01：受保護模式 DTA 與精確檔名首次搜尋

- 原版 DOSBox-X 有界擷取分別確認 `AH=1Ah` 保留完整暫存器／旗標、缺檔 `AH=4Eh` 返回 AX=`0012h` 且改寫 DTA 搜尋標頭，以及受控空 `MOX.SET` 的清 CF 與 DTA 結果欄。腳本 `--dta-find`／`--dta-find-present` 可重生；輸入、版本、位址基準與私有輸出 SHA-256 見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。兩側 PSP／環境和 DTA 位址不同，這是服務形狀證據，不是同狀態驗收。
- 隔離 dosgolem 工具分支依規格 218–219 的 DRAFT→READY→實作→CONFORMED 流程，保存 DTA selector／32 位偏移並接入僅 CX=0、單一精確 8.3 檔名的 FindFirst。非零舊結果欄、固定時間的成功檔案、萬用字元及越界均有合成測試。`go test -buildvcs=false ./internal/machine -count=1` 與固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 通過；同一真檔合成啟動診斷從第 4062 步至第 4168 步，在 dosgolem 重定位 LE 線性 `0x148224` 的 `38 10` 停止。未修改 remake 玩法；正常玩家畫面及玩法同狀態對拍仍未知。
- 下一步先核對該 byte 比較的原版位置與狀態，再判斷通用 CPU 規格；持續查合成 PSP／環境及完整資料消費端。原版 EXE、合成 `MOX.SET` 與完整除錯輸出僅留在未版控工作區。
- 工具分支已提交並推送 `fd9082c` 至 `github/codex/moo2-parity-20260930`；文件變更另在 MOO2 分支。變更後全套 Go 測試重跑通過，寫入檔均為目前使用者 UID/GID。一次性容器已清空；檢查未發現 `.md` 同名誤掛目錄。專案既有 `go.sum`、`lbxinfo` 與部分 `.docker-cache` 為 root 所有，本輪未改動，沒有對整個工作樹執行擁有權修復。

## 2026-10-01：`38 10` 原版比較與通用 byte CMP

- 接手確認兩個分支乾淨，知識路由命中復古 remake 生命週期、dosgolem 原版對拍與規格閘門；讀取現行 `AGENTS.md`、`CONTEXT.md`、誠實現況、活表、dosgolem 能力矩陣及文件職責。發現 dosgolem 規格索引把已 `CONFORMED` 的 219 留成 `DRAFT`，於本輪修正。
- 新增 `startup_probe_131.py --cmp-byte`，在既有 DOSBox-X 隔離映像中以固定 1.31 真檔擷取 `0180:0036C224` 的 `cmp [eax],dl`。原版兩來源均零，記憶體 `00h → 00h`，旗標 `0202h → 0246h`；完整位址、雜湊、工具版本與證據等級見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。原版證據僅涵蓋此零差樣本。
- dosgolem [規格 220](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/220-cpu386-cmp-rm8-register.md) 經 DRAFT→READY→實作→CONFORMED；無前綴 `38 /r` 記憶體目的改用既有 `decodeAddress32`，另以合成測試驗異值方向、SS／DS、SIB、越界與截短拒絕。`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1`、固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 均通過。合成 PSP／環境診斷由第 4168 步至第 4944 步，停於重定位 LE 線性 `0x146903` 的 `26 8A 1E 42 84`；此新停點仍待原版核對，並非正常玩家路徑或玩法同狀態收據。下一步是核對其 ES 覆寫載入，並持續查 PSP／環境及完整資料消費端；本輪未改 remake 玩法。
- 工具分支已提交並推送 `941e76e` 至 `github/codex/moo2-parity-20260930`。本輪新檔與修改檔、私有擷取均為目前使用者 UID/GID；無 `.md` 同名目錄誤掛，相關一次性 Docker 容器無殘留。原版 EXE 仍只在本機唯讀輸入，完整除錯輸出未加入 Git。

## 2026-10-01：種族資訊頁舊存檔態勢與譯文隔離

- 接手後依復古遊戲路由載入中文顯示／語意隔離入口，文件寫入前核對專案文件職責。發現 `AIOpponent.StanceName` 是既有存檔與 AI 規則使用的中文值，但資訊頁原先用目前 `ui.json` 的繁中譯文辨識它；譯文改字後，舊存檔會落入「未知」。這是程式資料流可證實的顯示缺陷；原版外交規則對齊度未因此改變。
- 在 `internal/shell` 將五種既有態勢名稱映射為穩定代碼，`cmd/moo2` 資訊頁再以該代碼讀取目前語系文案；不改 `StanceName`、JSON 格式或 AI 判定。新增五種態勢、未知值及譯文從「宣戰」改為「交戰狀態」後仍顯示舊存檔值的回歸測試。
- `moo2-ebiten` 一次性 Docker 容器以目前 UID/GID、無網路、3 GiB／2 CPU／128 PID，在 Xvfb 下執行 `go test -buildvcs=false ./internal/i18n ./internal/shell ./cmd/moo2 -count=1`；三套件全通過。這是 remake 內部回歸測試，沒有原版同狀態對拍，也沒有 GUI 正常玩家路徑截圖。下一步仍按活表處理 `dosgolem` 正常路徑與玩法證據，不把本次顯示修正算成原版忠實度完成。
- MOO2 分支 `codex/moo2-wine-gorgon-parity-20260930` 本輪提交並推送；相關一次性容器已清空，六個修改檔均為目前使用者 UID/GID，沒有 `.md` 同名誤掛目錄。原有 root 擁有的 `go.sum`、`lbxinfo` 及部分 `.docker-cache` 未改動。

## 2026-10-01：`dosgolem` ES byte 載入與原版輔助收據

- 知識路由命中原版決定性對拍，讀取 dosgolem 的 `README.md`、`CLAUDE.md`、能力矩陣與規格索引；MOO2 與隔離 dosgolem 工作樹開工時均乾淨。固定 1.31 原檔從本機 ZIP 唯讀擷取至一次性 DOSBox-X 容器，SHA-256 核對後以 `startup_probe_131.py --es-byte-load` 擷取原版 **CS:EIP** `0180:0036A903 → 0036A906` 的 `mov bl,es:[esi]`。來源 `30h`、BL 變 `30h`、旗標與來源不變；ES=DS，異段覆寫須另由處理器規格與合成測試驗證。首輪探針誤期待五 byte 指令，依原版實際三 byte 長度修正後以同一隔離映像乾淨重跑；地址、工具版本、輸入與私有收據雜湊見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。
- 隔離 dosgolem 分支新增 [規格 221](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/221-cpu386-mov-r8-es-memory.md)，按 DRAFT→READY→實作→CONFORMED，讓帶 ES 覆寫的 `8A /r` 記憶體來源使用既有 32 位址解碼器。合成測試核對異段、EBP／SIB、低／高 byte 暫存器及失敗即關閉。`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 通過；以 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 重跑全部套件通過，兩個需要固定 MOO2 原檔的測試另以 `-v` 確認執行且通過。
- 合成 PSP／環境的真檔診斷由第 4944 步至第 5392 步，在 dosgolem **重定位 LE 線性位址** `0x14822D` 的 `C1 CA 08` 停住；新停點原版對應未知。下一個動態 oracle 行動是先核對此指令的原版位置、來源與結果，再評估通用 CPU 規格，同時追查 PSP／環境與完整資料消費端。這輪沒有修改 remake 玩法，沒有 dosgolem 正常玩家畫面或玩法同狀態收據。原版 EXE、完整記憶體與終端資料只留未版控工作區。
- MOO2 與隔離 dosgolem 分支本輪分別提交並推送；新檔與修改檔均為目前使用者 UID/GID，沒有 `.md` 同名誤掛目錄，相關一次性容器已清空。既有 root 擁有的 `go.sum`、`lbxinfo` 與部分 `.docker-cache` 未改動。

## 2026-10-01：原版 ES 載入收據的取樣勘誤與 dosgolem 局部檢查點

- 接手確認兩個分支乾淨；知識路由命中原版決定性對拍，讀取 dosgolem 入口、能力矩陣及專案現況，文件寫入前再讀文件職責。發現前輪 `26 8A 1E` 收據把候選 `EV` 前斷點 EBX=`0Fh` 與 `LOG 2` 後態 EBX=`FFFFFF30h` 配成同次指令，與 byte MOV 的高 24 位不變契約矛盾。新增 `--es-byte-load-ev` 以前後斷點重播，矛盾仍在；原因未知。修正 `--es-byte-load` 收據明列 EV／LOG 衝突，原版指令依同次連續 LOG `FFFFFFFFh → FFFFFF30h` 限定解釋。原檔雜湊、工具版本、兩種位址基準及私有輸出雜湊見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。保留舊收據的錯誤形成原因，不把候選前態冒稱同狀態。
- 隔離 dosgolem 分支新增固定 1.31 真檔、缺檔即跳過的 `TestMOO2ESByteLoadCheckpointWhenProvided`，從 LE entry 與受控合成 PSP／環境自然跑到第 4944 步，單步核對 EBX=`FFFFFFFFh → FFFFFF30h`、來源 `30h`、旗標與來源不變。此處原版 ESI／ESP 與 dosgolem ESI／ESP 各差 `0021DFE0h`，相對差同為 `34h`；只支持單一啟動指令的正規化對照。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false -v ./internal/machine -run TestMOO2ESByteLoadCheckpointWhenProvided -count=1` 與含固定原檔的 `go test -buildvcs=false ./... -count=1` 通過；後者輸出雜湊見研究紀錄。
- MOO2 與隔離 dosgolem 分支本輪分別提交並推送。新 `C1 CA 08` 停點的原版對應、原版環境／DTA 與完整資料消費端仍待查；沒有正常玩家畫面或玩法同狀態收據，也未改 remake 玩法。原版 EXE、DOSBox-X 完整終端和 Go 診斷輸出只留未版控工作區；修改檔由目前使用者擁有，相關一次性 Docker 容器已清空，沒有 `.md` 同名誤掛目錄。

## 2026-10-01：原版零輸入 ROR 與 dosgolem 規格 222

- 重新查知識路由表第 53 列，沿用 dosgolem 原版對拍入口。以固定 1.31 原檔及 DOSBox-X 2026.07.02 同次 `LOG 2` 核對 **DOSBox-X CS:EIP** `0180:0036C22D → 0036C230` 的 `ror edx,08`；EDX=`0 → 0`、EFLAGS=`0206h`。原版只覆蓋零輸入，非零與未定義 OF 的契約另由 Intel 手冊和合成測試限定。完整輸入／輸出雜湊與位址基準記在研究紀錄。
- 隔離 dosgolem 分支依 DRAFT→READY→實作→CONFORMED 新增規格 222，僅接 `C1 /1` 的 32 位暫存器與 `08h`，保留未定義 OF 的原值。固定原檔整合測試從 LE entry 自行重生第 5392 步 `0x14822D` 的零輸入並單步至 `0x148230`；合成診斷進至第 5529 步，在**重定位 LE 線性位址** `0x14701B` 的 `08 E0` 停下，原版對應待查。含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過，新整合測試另以 `-v` 確認確實執行。此輪未改 remake 玩法，無正常玩家畫面或玩法同狀態對拍。

## 2026-10-01：原版 AH 零輸入 OR 與 dosgolem 規格 223

- 接手核對兩個分支乾淨，知識路由命中原版決定性對拍與 spec 閘門，確認專案 `AGENTS.md` 已是唯一代理人規範。玩法矩陣仍有多個未閉合的玩家機制；本輪先處理活表列出的原檔執行器停點，不把 CPU 進度計入玩法完成度。
- 以固定 1.31 原檔及 DOSBox-X 2026.07.02 的同次 `LOG 2` 核對 **DOSBox-X CS:EIP** `0180:0036B01B → 0036B01D` 為 `or al,ah`；EAX=`1`、EFLAGS=`0202h` 前後不變。原版 AH=`0`，非零／未定義 AF 依 Intel 手冊及合成測試分開標示；輸入與私有輸出雜湊見研究紀錄。
- 隔離 dosgolem 分支依 DRAFT→READY→實作→CONFORMED 新增規格 223，接無前綴 `08 /r mod=11` 的 8 位元暫存器 OR。固定原檔整合測試從 LE entry 在合成 PSP／環境下自行重生第 5529 步 `0x14701B`，單步至 `0x14701D`；有界診斷至第 5806 步，在**重定位 LE 線性位址** `0x15C1DF` 的 `2E 8D 86 82 C2 15 00` 停下，原版對應未知。新 CPU／原檔整合測試的 `-v` 輸出及含固定原檔的 `go test -buildvcs=false ./... -count=1` 均通過。此輪未改 remake 玩法，仍無正常玩家畫面或玩法同狀態對拍。
- MOO2 與隔離 dosgolem 分支分別提交並推送；原版 EXE、DOSBox-X 完整終端及 Go 診斷輸出留在未版控工作區。修改檔均為目前使用者 UID/GID，沒有 `.md` 同名目錄或殘留的相關一次性 Docker 容器。

## 2026-10-01：原版 CS 前綴 LEA 與 dosgolem 規格 224

- 接手確認兩分支乾淨；知識路由命中原版決定性對拍與 spec 閘門，舊規格回填路由只適用於被解出的 RE 語意，不把本次 CPU 支援範圍擴充誤列成玩法 RE。固定 1.31 原檔在 DOSBox-X 2026.07.02 的同次 `LOG 2`，於 **DOSBox-X CS:EIP** `0180:003801DF → 003801E6` 核對 `lea eax,cs:[esi+00380282]`；ESI=`93h`、EAX=`31h → 380315h`、EFLAGS=`0212h` 不變。輸入與私有輸出雜湊見研究紀錄。
- 隔離 dosgolem 分支依 DRAFT→READY→實作→CONFORMED 新增規格 224，只讓 `LEA` 接受單一段前綴並忽略段基址；既有 021 與 186 規格附註舊拒絕範圍的後續擴充。合成四段測試與固定原檔整合測試通過；後者在第 5806 步以自身 ESI=`99h` 算出 EAX=`15C31Bh`，與原版 ESI、EAX、旗標不同，未宣稱同狀態。含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過；有界診斷於第 5808 步、**重定位 LE 線性位址** `0x15C1E7` 的 `8E 03` 停住，原版對應待查。未改 remake 玩法，仍無 dosgolem 正常玩家畫面或玩法同狀態對拍。
- MOO2 與隔離 dosgolem 分支分別提交並推送；原版 EXE、DOSBox-X 完整終端與 Go 診斷輸出留未版控工作區。修改檔都是目前使用者 UID/GID，沒有 `.md` 同名誤掛目錄或殘留相關一次性 Docker 容器。

## 2026-10-01：LEA 呼叫次數勘誤與下一停點定位

- 目標是查上輪 ESI=`93h／99h` 差異。知識路由命中原版決定性對拍與規格閘門；在隔離 dosgolem 副本的私有探針記錄第 5806 步前暫存器變化，看到 `mov eax,33h` 經 `mov esi,eax` 與 `lea esi,[esi+esi*2]` 到 ESI=`99h`。以固定 SHA-256 的 1.31 原檔在 DOSBox-X 2026.07.02 執行版控 `startup_probe_131.py --startup-value`，由其 **CS:EIP** `0180:00348105` 連續記錄 128 指令，也得到同一來源 EAX=`33h`／ESI=`99h`；`LOG 80` 被解讀為十六進位 `80h`。先前 `--lea-cs` 單點取得 `31h／93h` 是另一呼叫，不是合成環境長度已造成的差異。保留舊收據並於研究紀錄及規格 224 追加勘誤，修正 CONTEXT／WORKLIST 的現行敘述。
- 同一連續記錄確認原版 **CS:EIP** `0180:003801E7` 為 `mov es,[ebx]`，對應 dosgolem 重定位 LE 線性 `0x15C1E7` 的 `8E 03`。隔離 dosgolem 分支建立規格 225 並完成 DRAFT→READY→實作→CONFORMED，只擴充無前綴 DS:[EBX] word 載入 ES。原版 `startup-value-logcpu.txt` SHA-256 `f209eb9b55ec3c9184f19f335142880e747bb68848ddf091c6764066f7dc954c`、`startup-value-registers.json` SHA-256 `bf96d8bebcb9169c43466c4b3216c0f294ffe783478610c02677e4075c28f25c`，dosgolem 私有 `moo2-esi-changes-225.txt` SHA-256 `3c75c3fdd3024a92e7693b183cd695a99dff44559f78f73fbe7918606d762010`；完整輸入雜湊、工具版本、位址基準與證據等級在研究紀錄。第一次 `LOG 80` 的驗收腳本錯把輸出當 80 行，失敗分類為腳本參數進位解讀，改成 128 行後以相同容器／命令乾淨重跑成功。
- 固定原檔整合測試 `TestMOO2MoveESFromDSEBXCheckpointWhenProvided` 在第 5808 步讀到來源 `0188h` 並單步至 `0x15C1E9`；合成測試含非零 DS base、越界、無效 selector 與拒絕形狀。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 在一次性 Go 容器全通過。探針又從原檔 LE entry 自然進至第 5818 步、dosgolem 重定位 LE 線性 `0x15C31B` 的 `CD 33`（`INT 33h`）失敗即關閉；原版同一連續 LOG 也命中其 **CS:EIP** `0180:0038031B int 33`，但服務結果未知。完整測試與診斷的私有雜湊見研究紀錄。
- 兩側 EFLAGS、堆疊、PSP／環境和完整資料流仍不同；沒有正常玩家畫面或玩法同狀態收據。下一最小行動是查明 `INT 33h` 這次呼叫／回傳與裝置狀態，再建立受限服務規格。原版 EXE、完整終端與 Go 診斷輸出留未版控工作區；本輪無 MOO2 Go 玩法變更。一次性 Docker 容器已清空，新增／修改檔為目前使用者 UID/GID，沒有新 `.md` 同名目錄；主儲存庫既有 root 擁有的 `go.sum`、`lbxinfo` 與 `.docker-cache` 未改動。
- 隔離 dosgolem 分支 `codex/moo2-parity-20260930` 已提交 `ba72cda30fc642abe90d22e4b28ee5f70f5e3c56` 並推送至 `github`；主儲存庫文件同輪另行提交推送。

## 2026-10-01：原版滑鼠查詢與啟動 record 消費端

- 接手時 MOO2 與隔離 dosgolem 工作樹均乾淨，知識路由命中原版決定性對拍與規格閘門；現行活表的 dosgolem 停點是第 5818 步 `INT 33h`。版控原版探針新增 `--mouse-query`，以固定 1.31 EXE 在 DOSBox-X 2026.07.02 SDL2 heavy debugger 從同一啟動序列抓前態、返回斷點與後續 32 指令。原版 **CS:EIP** `0180:0038031B` 的輸入 EAX=`3`、EBX=ECX=EDX=`0`，`0180:0038031D` 回 EAX=`3`、EBX=`0`、ECX=`320`、EDX=`100`，EFLAGS=`0006h` 不變；後續 caller 寫入 record `+0/+4/+8/+0Ch`。原版輸入、工具版本、位址空間與私有輸出雜湊見 [`docs/re/dosgolem-moo2-intake-20260930.md`](docs/re/dosgolem-moo2-intake-20260930.md)。這是固定 DOSBox-X 裝置狀態的輔助基準，非普遍預設座標。
- 隔離 dosgolem 依 DRAFT→READY→實作→CONFORMED 建立規格 226，只讓 MOO2 保護模式設定處理 `INT 33h/AX=3`；初始座標與按鍵可由 `SetMouseState` 控制，其他功能和一般 FD2 設定仍拒絕。固定原檔第 5818 步自行重生回傳，整合測試再讀回原檔寫入 record 的 `3、0、320、100`。下個停點第 5830 步 `8F 47 14` 經規格 227 僅擴充受限 `POP dword` 記憶體目的形狀；合成測試驗證非零搬移、段基址、位移與拒絕路徑，固定原檔第 5830 步自行核對 SS:[ESP] 到 record `+14h`、ESP 加 4。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有輸出 SHA-256 `47dc1f8bd19688d9f2bf5d38637770bba02cdc74cd7f380016b76ffb6c3ea109`；新增 record／POP 整合斷言另以 `-v` 通過。
- 有界原檔探針現停第 5838 步、dosgolem **重定位 LE 線性位址** `0x15C1D6` 的 `66 8C 03`；原版 **CS:EIP** `0180:003801D6` 的連續 LOG 顯示 `mov [ebx],es`，下一步先核對記憶體效果，再建立受限規格。兩側 EFLAGS、PSP／環境、堆疊位置與 DTA 指標不同；沒有正常玩家畫面或玩法同狀態收據。本輪未改 MOO2 Go 玩法，也未散布原版輸入。原版 EXE、DOSBox-X 完整終端與 Go 探針輸出留未版控工作區；Docker 容器及檔案擁有權於提交前複核。
- 隔離 dosgolem 分支 `codex/moo2-parity-20260930` 已提交 `27d1f989c660e04f48d8194af8372e1fb1608610` 並推送 `github`；MOO2 主分支文件同輪另行提交推送。改動檔案均屬目前使用者 UID/GID，未發現新 `.md` 同名目錄，檢查時相關一次性 Docker 容器清空。

## 2026-10-01：ES selector 寫回與後續滑鼠服務停點

- 接手兩工作樹均乾淨；知識路由命中復古 remake、原版決定性對拍及 spec 閘門，讀取相應入口與專案現況。固定 1.31 原檔在 DOSBox-X 2026.07.02 SDL2 heavy debugger 的同次啟動 `LOG 2`，於 **CS:EIP** `0180:003801D6` 執行 `66 8C 03`；DS:[EBX] 前後皆 `0188h`。版控 `startup_probe_131.py --es-store` 可重生 LOG、暫存器與前後兩位元組；雜湊、輸入及位址基準已記在研究紀錄。來源等於原目的，沒有宣稱原版的非相等寫入已實測。
- 隔離 dosgolem 規格 228 經 DRAFT→READY→實作→CONFORMED，只接該單一 ModRM 形狀。合成測試覆蓋不同值、小端序、非零 DS base、界限拒絕與未列形狀拒絕；固定原檔整合測試自行抵達第 5838 步並單步到下一指令。一次性 `golang:1.24-bookworm` 容器以原 ZIP 唯讀掛載，在容器內擷取 EXE，執行 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過；私有輸出 SHA-256 `78c15a85c6a0902893dfde545d0db82ea8ebc4dfe275812eefed49cee5d1e0d1`。同容器執行 `go run -buildvcs=false ./workplace/moo2-probe /tmp/ORION2.EXE`，私有收據 SHA-256 `dd4ec7d5ac8b04bfc8bb2c8acd977b342f2d6f10654dc6657d379d702f065da7`，下一停點第 6007 步、dosgolem **重定位 LE 線性位址** `0x15C31B` 的 `INT 33h/AX=21h`。
- 這輪未改 MOO2 Go 玩法；仍無正常玩家畫面或玩法同狀態對拍。下一最小行動是取固定原版 `INT 33h/AX=21h` 的輸入、返回與 caller 消費端，審查是否足以建立受限工具規格。原版 EXE、完整終端及私有診斷留未版控工作區。
- 隔離 dosgolem 分支提交 `c3cc5d90b235e3466f4a86484c0b3d907432da7d` 並推送至 `github/codex/moo2-parity-20260930`；MOO2 現況文件同輪提交與推送。相關一次性 Docker 容器清空；新增檔案屬目前使用者，未發現新 `.md` 同名目錄。

## 2026-10-01：滑鼠 `INT 33h/AX=21h` 固定原檔收據

- 接手時兩分支乾淨；知識路由命中復古 remake、原版決定性對拍與 spec 閘門，讀取專案現況、dosgolem 能力入口及文件職責。本輪只擴充原版執行器；MOO2 玩法 RE 閘門仍開啟，未改 Go 玩法。
- 版控 `startup_probe_131.py --mouse-function-21` 以固定 1.31 原檔在 DOSBox-X 2026.07.02 SDL2 heavy debugger 擷取第二次滑鼠呼叫。首輪 `EV` 候選返回與 LOG 矛盾；第二輪 32 指令只進 DOS extender；第三輪返回斷點未清除，LOG 只得一行。清理斷點後在相同隔離映像及命令乾淨重跑，得到同次 **DOSBox-X CS:EIP** `0180:0038031B → 0038031D` 的 `AX=21h → FFFFh`、`BX=0 → 3`，呼叫端寫入 record。舊候選撤回而保留私有輸出，完整雜湊、位址基準與未知邊界見研究紀錄。
- 隔離 dosgolem [規格 229](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/229-moo2-protected-mouse-software-reset.md) 經 DRAFT→READY→實作→CONFORMED，僅允許 MOO2 設定的 `AX=21h`；回傳依同次原版 LOG，內部按鍵／座標重設依 DOSBox-X 原始碼列為平台契約近似。合成測試與固定原檔第 6007 步整合測試均通過；一次性 `golang:1.24-bookworm` 容器以原 ZIP 唯讀掛載，執行 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有輸出 SHA-256 `ff619b50ad8cf1c146c56b3bdc675c7da01d81f5d55cc1d1cdd6c12f8f63a47e`。同容器執行 `go run -buildvcs=false ./workplace/moo2-probe /tmp/ORION2.EXE`，私有收據 SHA-256 `25ef4143e67248625582e718a6f03e5d677b0233cd68bacd3ecf019b6f270643`，下一停點第 6126 步、dosgolem **重定位 LE 線性位址** `0x15C31B` 的 `INT 33h/AX=1Ah`。
- 這輪沒有正常玩家畫面或玩法同狀態對拍。下一個動態 oracle 行動是擷取原版 `AX=1Ah` 的輸入、回傳和 caller 消費，再審查是否足以建立受限服務規格；原版輸入、終端及中間 LOG 只留未版控工作區。
- 隔離 dosgolem 分支提交 `fc3fa4a5e1e3e5ae335186d10b7d4691c5596369` 並推送至 `github/codex/moo2-parity-20260930`；MOO2 現況文件同輪提交推送。相關一次性 Docker 容器清空，新增／修改檔均屬目前使用者，未發現新 `.md` 同名目錄。
