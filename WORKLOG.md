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

## 2026-10-01：滑鼠零敏感度設定與視訊服務停點

- 接手時 MOO2 與隔離 dosgolem 分支皆乾淨；知識路由命中復古 remake 的原版對拍、規格閘門與文件職責。MOO2 玩法 RE 閘門仍開著，這輪只擴充原版執行器與證據，沒有修改 Go 玩法。
- 版控 `startup_probe_131.py --mouse-function-1a` 在固定 1.31 原檔、DOSBox-X 2026.07.02 SDL2 heavy debugger 中連續重播前三次滑鼠呼叫。第一次完整輸出後外層 `timeout 170s` 結束；分類為驗證逾時設定，改 `timeout 240s` 並以相同映像／命令乾淨重跑，正常退出且三份收據雜湊相同。原版 **CS:EIP** `0180:0038031B → 0038031D` 的 `AX=1Ah` 進入 BX=CX=DX=`0`，返回暫存器與旗標不變，caller 寫入 record；輸入雜湊、私有收據與證據等級見研究紀錄。
- 隔離 dosgolem 規格 230 經 DRAFT→READY→實作→CONFORMED，只接受 MOO2 設定下的三個零敏感度值；非零值及一般 FD2 仍拒絕。合成狀態測試與固定原檔第 6126 步整合測試通過。一次性 `golang:1.24-bookworm` 容器以原 ZIP 唯讀掛載，執行 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有輸出 SHA-256 `a983b838b0b84c14b442a82b127d937f4101c88c5fe6dfc0ce6db6742f536e6a`；`go run -buildvcs=false ./workplace/moo2-probe /tmp/ORION2.EXE` 的私有收據 SHA-256 `aab5a75b6ec1b381e7a7008968ae53844b0e7ce1deed221ef8d8b6872ca7073e`，下一停點為第 6256 步、dosgolem **重定位 LE 線性位址** `0x15C2B2` 的 `INT 10h`。
- 沒有正常玩家畫面或與 remake 玩法同狀態收據。下一個動態 oracle 行動是取固定原版 `INT 10h` 的實際功能號、返回與 caller 消費端；原版檔案及完整終端只留未版控工作區。
- 隔離 dosgolem 分支提交 `f96579f80a0ab249ce9a5f8d6f90e8040d8a1aa6` 並推送至 `github/codex/moo2-parity-20260930`；MOO2 現況文件同輪提交推送。相關一次性 Docker 容器清空，新增／修改檔屬目前使用者，未發現新 `.md` 同名目錄。

## 2026-10-01：視訊模式 03h 與合成缺檔路徑

- 知識路由命中復古 remake 規格閘門、DOS 平台來源與文件職責；主專案玩法 RE-first 閘門仍有效。本輪僅修改隔離 dosgolem 啟動工具及證據，未改 MOO2 Go 玩法。
- 固定 1.31 原檔的 DOSBox-X 輔助探針 `--video-mode-03` 命中 `INT 10h/AX=0003h` 並取得同次返回與 caller 記錄；私有收據、雜湊、位址基準見研究紀錄。隔離 dosgolem 規格 231 經 DRAFT→READY→實作→CONFORMED，只記錄模式 03h，不宣稱畫面已實作。
- 首次全套測試在原檔整合測試失敗：前一段滑鼠單步未更新測試局部計數，導致預期 6256、實得 6244；補齊 12 個計數後以同一 `golang:1.24-bookworm` 容器與命令乾淨重跑，含原檔 `go test -buildvcs=false ./... -count=1` 全通過，私有輸出 SHA-256 `66f262c63281b32a8d619f190b0d710956ffecdf158624cd0603f796b57921ee`。
- 初次原檔診斷走到第 10175 步的 `MOV` 報錯；回查確認第 10173 步已有 `INT 21h/AH=4Ch`，屬探針未在退出時停止的錯誤。修正 `workplace/moo2-probe` 後重跑，於第 10173 步以代碼 1 結束，主控台回 `Unable to open mox.set`，私有輸出 SHA-256 `8ee984ad0f300255f41f4259729102a07eb193461cc00b307d35b0a2e2f694d6`。下一步以既有空 `MOX.SET` 輔助收據建立受控成功分支；目前無正常玩家畫面或玩法同狀態對拍。
- dosgolem 分支 `20e869be0d10f3e39074938969be85368de00ec4` 已推送至 `github/codex/moo2-parity-20260930`。本輪 Docker 均為一次性容器；原版檔案、完整終端及診斷輸出仍只留在未版控工作區。

## 2026-10-01：空設定檔比較與正版資料啟動

- 接手時兩分支乾淨；知識路由命中復古 remake 與規格閘門，載入 `retro-remake-spec-gated-workflow.md`、專案現況及文件職責。本輪只擴充隔離 dosgolem 的 CPU／診斷能力，不改 MOO2 Go 玩法。
- 版控探針 `--empty-mox-cmp` 從固定 1.31 原檔與受控零位元組 `MOX.SET` 擷取 DOSBox-X 同次 `CMP` 收據；原始 bytes、兩套位址基準、原版／合成來源值與旗標見研究紀錄。dosgolem 規格 232 經 DRAFT→READY→實作→CONFORMED；含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過，最終乾淨重跑私有輸出 SHA-256 `a6c8319d03cbfbe8cfe6cef27011b170e03d6f2bffe64bbc42a45816bf8cf69e`。空檔後續仍以代碼 1 退出。
- 查證正版 ZIP 含 553-byte `MOX.SET`，先單檔診斷：20 萬步上限仍在前進，故調整為有界 100 萬步並降低輸出量。完整資料提取首次被三個聲卡子目錄觸發路徑 assertion；修正只取 ZIP 根層檔案後同容器工具鏈乾淨重跑。正版根層 417 檔及官方 1.31 EXE 只在一次性容器暫存區使用，dosgolem 自行執行至第 288215 步的 `66 F7 /0` word 形狀，私有診斷 SHA-256 `c032e24493aef6d936a810d10dfe109b9884de5a78cb5527fdebcaeec63cfb9e`。尚無此新指令的原版同次收據，也無正常玩家畫面或與 remake 同狀態玩法對拍。
- 隔離 dosgolem 分支提交 `07320bee8e6bfc6eaf5163146432c8d6ed3ea2d1` 並推送至 `github/codex/moo2-parity-20260930`。一次性 Docker 容器皆已結束，新增檔案屬目前使用者，未發現誤建 `.md` 目錄；私有原版資料及輸出未入版控。

## 2026-10-01：完整資料 F7 解碼與原版黑畫面收據

- 知識路由命中復古 remake 規格閘門；載入 `retro-remake-spec-gated-workflow.md`、專案現況與文件職責。隔離 dosgolem 的 `--full-data-test-word` 探針用相同正版資料／1.31 EXE 設候選斷點，未命中時在私有工作區留下 DOSBox-X 畫面、暫存器紀錄與終端。畫面仍黑，沒有主選單或原版該指令的同狀態收據；證據雜湊見研究紀錄。
- 對 dosgolem 原第 288215 步的 `66 F7 /0`，依 Intel CPU 契約完成規格 233、受限解碼與段／旗標／失敗測試。一次性 `golang:1.24-bookworm` 容器以固定 1.31 EXE 執行 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過。完整正版根層 417 檔案診斷自然前進至第 295276 步，再因合成 DOS 記憶體空間不足以代碼 1 退出；私有收據 SHA-256 `c75e200fda4251e97752600df06c84c14d2fbc27eb232afcd50f60936d79e744`。
- 隔離 dosgolem 分支提交 `774471211bd818e9325bf0653aa3f3de66b5c86a` 並推送至 `github/codex/moo2-parity-20260930`。主專案 Go 玩法未改；正常玩家畫面與玩法同狀態對拍仍未完成。下一輪先釐清合成 DOS 記憶體服務／環境和原版可見啟動結果的關係，再決定是否需要擴充執行器；不得把黑畫面或 Intel CPU 規格當玩法原版收據。一次性 Docker 容器均已結束，新增／修改產物 UID/GID 為 `1000:1000`，未發現誤建 `.md` 目錄；私有資料未入版控。

## 2026-10-01：合成記憶體耗盡原因稽核

- 依新 `AGENTS.md` 入口與知識路由重查現況。隔離 dosgolem 的退出探針加印 DPMI 呼叫與區塊帳本，以相同正版 ZIP 417 個根層檔案及 1.31 EXE 在一次性 `golang:1.24-bookworm` 容器重跑。第 295276 步退出不變；稽核收據 SHA-256 `05a3295912c263601e75dcf1933fd6ad14022edebd9e3a072e4d6f47be165d`。`0100h` 2 次、`0501h` 126 次、`0502h` 122 次，未實作服務為空；初始 LE 映像大於 640 KiB，而低位 DOS 游標從映像後起算，線性釋放也不回收位址。原版相同配置結果仍未知，詳細證據見研究紀錄。
- 隔離 dosgolem 分支提交 `6c1ffad8b0bafe65e6f37078a948a9fc91b8dc1c` 並推送至 `github/codex/moo2-parity-20260930`。主專案 Go 玩法未改；下一步以 DPMI 1.0 一手規格、原版啟動收據及現有區塊測試界定記憶體模型，再實作、重跑 dosgolem，不能靠擴大上限掩蓋合成環境限制。正常玩家畫面與玩法同狀態對拍仍未完成。

## 2026-10-01：DPMI 線性回收與原版 DOS 配置收據

- 路由命中復古 remake 的規格閘門、dosgolem oracle 與文件職責，載入 `retro-remake-spec-gated-workflow.md`、dosgolem 能力入口及 `project-document-responsibilities.md`。MOO2 的 RE-first 玩法閘門仍開啟；本輪沒有修改 remake Go 玩法。
- 隔離 dosgolem 規格 234 依 DPMI 1.0 `0501h／0502h` 建立 DRAFT→READY→實作→CONFORMED；補升序空閒區首次適配、合併、獨立控制代號與重用清零。合成測試及固定 1.31 原檔的 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE GOCACHE=/tmp/go-cache go test -buildvcs=false ./... -count=1` 全通過，私有測試輸出 SHA-256 `0c4aa5818a10ca07c604b44f0daf3bfee775b07a02d0718232aeec1666a1e0e7`。
- 相同正版根層 417 檔案的 dosgolem 完整資料診斷於第 3,947,961 步自然退出；線性空間仍餘 39,150,168 bytes、DOS 空間 0，首筆 `0100h` 請求 513 段落失敗。DOSBox-X 2026.07.02 SDL2 heavy debugger 輔助擷取同一 1.31 EXE 的首次 `0100h` 513 段落請求成功，返回實模式段 `0FE3h`。第一次 `EV` 快照解析錯取舊值，修正解析後以同一容器／資料重跑，保留錯誤私有紀錄供勘誤。兩個執行器的第二筆請求大小已分歧，不宣稱同狀態玩家對拍；完整雜湊、位址空間與兩側數值見研究紀錄。
- 直接把 DOS 低位游標改小會覆蓋現行 LE 映像；下一個最小可重現行動是先界定低位記憶體與 LE 映像隔離、雙模式存取及防覆蓋測試，再修正 dosgolem `0100h` 並重跑完整資料。正常玩家畫面與玩法同狀態對拍仍未完成；原版素材及完整終端只在本機私有工作區。

## 2026-10-01：高位 LE 與低位 DOS 記憶體隔離

- 沿用本輪已載入的規格閘門及文件路由；隔離 dosgolem 規格 235 經 DRAFT→READY，明示高位 LE loader 將 1.31 物件與 fixup 平移 `0xF0000`，為 DOS 記憶體保留低位 arena。原 loader 路徑不變，合成測試驗證 513／176 段落配置、實模式 bus／selector 資料往返、LE 防覆蓋與拒絕邊界；固定原檔測試核對所有支援的 object fixup。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有輸出 SHA-256 `6ed111b7c3a51af6330c947b8a447ce2a065fcc5f7c91da5137040c8f2582de1`。
- 完整正版根層 417 檔的明示高位診斷於第 189,580 步成功完成首筆 `0100h` 513 段落配置；第 189,582 步停在 `0300h` 要轉呼叫的實模式 `INT 10h` 未註冊，底層錯誤與未實作計數已記錄。私有輸出 SHA-256 `5dbf8b22cda56a2f77fd0ff498daa3dd5d1a14f7b194f074fe748013708023e8`。原檔第二筆配置尚未自然抵達，故規格 235 維持 READY；下一步從原版取這筆視訊呼叫的輸入／返回／consumer，再決定受限服務。詳細兩套位址基準與前輪勘誤見研究紀錄。

## 2026-10-01：原版 VBE 查詢與隔離執行器前進

- 固定 1.31 EXE／正版根層 417 檔，以 DOSBox-X 2026.07.02 SDL2 heavy debugger 作輔助原版基準，擷取 `4F00h` 的緩衝、返回及 caller，直接 `4F07h` 的零座標返回與 caller，以及兩次 `4F00h` 後 `4F01h/CX=0101h` 的封包和 256 位元組緩衝。第二筆查詢最初被誤設為 `4F01h`；依實際封包改為逐筆記錄並用同一容器乾淨重跑。私有 `real-video-4f01-registers.json` SHA-256 `baafd3f5178954e1d5d24224bf7e98233f62b701cda0088c2554c40f018a1454`，詳細輸入、位址基準、資料效果及雜湊見研究紀錄。
- 隔離 dosgolem 規格 236、237、238 逐一經 DRAFT→READY→受限實作→CONFORMED；MOO2 專屬路徑回傳 VBE 控制器資訊、零座標顯示起點及模式 0101h 資訊。全套含官方 1.31 EXE 的 `go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-238.txt` SHA-256 `84406686fb97b0c27f8aa42fe41190c08bc256c6f375a45c7b71b5ec94e4088d`；模式緩衝以原版 256 位元組 SHA-256 比對，其他模式／越界拒絕。
- `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 從 LE entry 自生執行至第 190,517 步，下一停點為直接 `INT 10h/AX=4F02h`、EBX=`0101h`，私有 `workplace/moo2-probe-238-full-game.txt` SHA-256 `6f5df2e26b65c5935db63cff4751b103182ff75bab09c18314d8a42ed2c70284`。這僅驗收隔離平台服務；正常玩家畫面、第二筆 DOS 配置與玩法同狀態尚未完成。下一步擷取原版 `4F02h` 返回／consumer，再決定最小平台支援。
- 同一原版輸入與 DOSBox-X 工具擷取直接 `INT 10h/AX=4F02h、BX=0101h` 的 EAX=`004Fh` 返回、其餘暫存器／旗標保持及 caller record 消費。私有 `video-mode-4f02-registers.json` SHA-256 `5b6d94309ba8bcc5e9de221ed6b8545c7a5c8f1cb0b86ceae7dac2b9de762dea`；位址及來源限制見研究紀錄。
- 隔離 dosgolem 規格 239 經 DRAFT→READY→受限模式設定→CONFORMED；含官方 1.31 EXE 的全套 Go 測試通過，私有 `workplace/full-test-239.txt` SHA-256 `e78f1025caaaacd4ec286585cfc18a0e92afe07dad03de08e18aad186322cb5f`。完整資料自 LE entry 重跑至第 1,151,730 步的 opcode `04h` 停點，私有 `workplace/moo2-probe-239-full-game.txt` SHA-256 `96af49b38b75df1cfd62c08ea9d30ba2f57624841bd9db6851e0b78f13507da6`。這是工具前進，不是玩家畫面或玩法同狀態對拍；下一步核對原檔 bytes 與 CPU 指令契約。
- 本輪無 MOO2 Go 玩法變動；正常玩家畫面、玩法同狀態對拍及正式封包均未完成。原版素材與探針終端留在本機私有工作區，不入版控。
- 隔離 dosgolem 分支的線性回收提交 `c26db33215b0e8df06d4ac748305181edae0ca4d`、高位映射提交 `c6cd74fe187fc7badd12969f821b5bc14ca8b448` 均已推送至 `github/codex/moo2-parity-20260930`。`git diff --check` 通過；新增／修改檔 UID/GID 為 `1000:1000`，dosgolem 工作樹未見 root-owned 檔或 `.md` 同名目錄。`docker ps -a` 無本專案或 dosgolem 殘留容器；主專案既有 root-owned 快取與檔案未動，其他專案容器未碰觸。

## 2026-10-01：dosgolem CPU、實模式平台埠與第二 DMA 遮罩

- 路由命中規格閘門、dosgolem 原版對拍及平台硬體規格優先；依 240–242 規格的 DRAFT→READY→實作→CONFORMED 流程，補通用 `ADD AL,imm8`、MOO2 雙模式共用既有 DSP 埠，以及第二組 DMA 控制器 `D4h` 遮罩。未改主專案 Go 玩法或發行包。
- DOSBox-X 2026.07.02 SDL2 heavy debugger 輔助擷取原版 `04 20` 的同次暫存器／旗標；dosgolem 固定官方 1.31 EXE 的 `go test -buildvcs=false ./... -count=1` 全通過。最終私有測試輸出 `workplace/full-test-242.txt` SHA-256 `63ce81a460661c68c24979ee50e1e3d3720c8c7f9fd309c627f0fd1558e3daf7`。
- `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 使用正版 ZIP 根層 417 檔與官方 EXE，自 LE 入口自然完成 513／176 段落兩筆配置，越過 `04 20`、DSP `0226h` 與第二 DMA 遮罩 `D4h/05h`，現於 `DPMI 0300h → INT 66h` 的實模式 `1201:0244` 遇到未支援的 `OUT D8h,00h`；私有輸出 SHA-256 `142f0d05f0feea086353bd9cfbd5679f0dda095b9644da82cd8007376c187c06`。詳細原始雜湊、工具、位址空間與各步收據見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- 以上只證實 dosgolem 平台服務的窄範圍前進。原版與合成 PSP／環境、完整資料流仍不同，沒有正常玩家畫面、音訊或與 remake 玩法同狀態對拍。下一最小行動是查第二 DMA `D8h` 的公開規格並建立獨立平台規格；避免展開 DAC／PIT／DMA 逐週期考古。
- 隔離 dosgolem 分支提交 `7281a38` 並推送 `github/codex/moo2-parity-20260930`。本輪 `git diff --check`、Go 格式檢查及 Docker 殘留檢查通過；新規格與私有探針輸出為目前使用者 UID/GID，未發現 dosgolem 工作樹的 root-owned 產物或誤建 `.md` 目錄。

## 2026-10-01：dosgolem 第二 DMA、SB16 B0 與 BDA 啟動狀態

- 固定 1.31 `ORION2.EXE` 與正版根層 417 檔，隔離 dosgolem 依規格 243／244 補第二 DMA 控制器暫存器及頁埠；規格 245 限定 SB16 `B0 30 00 00` 一個 16 位元 word 的單次傳輸、完成閘門與 IRQ。均採硬體規格近似，沒有逐週期或音訊聽感主張。固定原檔 Go 全套測試在三個階段均通過；收據、雜湊與來源見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- B0 後誤讀 `0006h` 的診斷由原始 bytes 追至 `0040:0063` 未初始化。規格 246 重用既有 DOS/4GW BIOS 資料區安裝函式，MOO2 啟動接線在成功安裝後才連接 DPMI 與共用埠，衝突直接拒絕。正常安裝、衝突無副作用測試通過；固定 EXE 全套 Go 測試輸出 SHA-256 `7fa4b1acb412bf35c9110a5e94a7349f0dc40a951a396b7b008708b8b5b46f08`。
- 本輪原檔自行越過 `D8h`、`8Bh`、`B0h` 與原先的 `0006h` 假停點；現行真停點為 dosgolem 實模式 `1201:073F` 的 `IN 0225h`。這些是平台執行器的有界進展，尚無正常玩家畫面或與 Go remake 同狀態玩法收據；本輪未改 remake 玩法。下一最小行動是查 `0225h` 公開平台契約與原檔讀取後的消費端。
- 隔離 dosgolem 分支提交 `642abfe` 並推送 `github/codex/moo2-parity-20260930`；`git diff --check`、Go 格式檢查及新檔 UID/GID 抽查通過。正式接線後完整資料自生收據 `workplace/moo2-probe-246-full-game.txt.gz` SHA-256 `0b3d80dc9aba63f4973089d855384e24080d8e0f529f18f4caca31e0b508667c`；原版資料與私有終端未加入 Git。

## 2026-10-01：SB16 混音器索引 `32h` 的受限讀寫

- 隔離 dosgolem 的可丟棄探針確認原檔在 `IN 0225h` 前選了索引 `32h`；Creative 原廠指南直接定義 `32h/33h` 的左右數位語音音量、高五位有效及預設 24/31。規格 247 依 DRAFT→READY→實作→CONFORMED，只補這兩個索引的狀態保存，不推論原版實機當次音量或實際聲波。
- 固定官方 EXE 全套 Go 測試通過，私有輸出 SHA-256 `fe95d9119cde6fb624038c04a132ff9df0694f0031b3e5928c3d78c329d23147`。正版資料自 LE entry 自生越過 `IN 0225h`，實模式呼叫正常返回；現行停點為 dosgolem 高位 LE 線性 `0x25221E` 的 `83 C8 10`，私有收據 SHA-256 `ad3568f83e7a85f70e682484bf392bac3bcd66dce42fd474e1d061b6af3083ca`。下一最小行動是核對 CPU 指令形狀及原版同次消費端，未改 Go remake 玩法。
- 隔離 dosgolem 分支提交 `8f6bf14` 並推送 `github/codex/moo2-parity-20260930`。`git diff --check`、Go 格式、受控新檔擁有權與專案 Docker 容器殘留檢查均通過；原版 ZIP、EXE 與私有終端未加入 Git。

## 2026-10-01：OR／XOR 原版樣本與 BIOS 時鐘接線

- 接手主專案 `200f33ed336e0712e0da2a65a7182b63547691b2`，工作樹乾淨；沿用使用者授權的 DOS 原版／dosgolem 路線與隔離分支。知識路由命中規格閘門、硬體規格優先與文件職責，載入對應入口；沒有修改 Go remake 玩法或建立發行包。
- 固定 1.31 EXE／正版根層 417 檔，以 DOSBox-X 2026.07.02 SDL2 heavy debugger 作輔助樣本，分別執行版控探針 `--or-register-imm8` 與 `--xor-register-imm32`，核對 raw bytes、暫存器與旗標。XOR 首段 LOG 的 PF 顯示與實際 EFLAGS 不同；同斷點續取後確認 PF=1，保留初次收據並記錄來源等級，未採錯誤顯示作測試預期。完整雜湊與位址空間見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- 隔離 dosgolem 規格 248／249 依 DRAFT→READY→實作→CONFORMED 補兩個 CPU 暫存器形狀；固定原檔全套 Go 測試通過。自然執行抵達八百萬步探針上限，尾端收據定位為等待 BIOS `046Ch` 改變；MOO2 尚未接已有 BIOS 時鐘，不以增加步數上限掩蓋缺口。
- 規格 250 依同一流程重用 `InstallLEBIOSClock`，測試覆蓋 tick、既有 hook、CPU 狀態保留與客製向量拒絕。首次編譯的新錯誤訊息漏套件匯入；改用已有 `errors.New`，以同一容器與命令乾淨重跑。`DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，輸出 SHA-256 `d57b7d3b9da005b9ae06105bfa2e3bfcf7b11389c92a6c830c5cda109741b91e`。
- `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 自 LE entry 離開 tick 等待；第 1,545,396 步停於 dosgolem 高位 LE 線性 `0x24C31B` 的 `INT 33h/AX=0000h`。私有收據 SHA-256 `74b73985b73a1f077f2005343fef0168df1d05dbe49d03d8f824369f5105a1b8`。下一最小行動是擷取此服務的原版輸入、返回與消費端，再審查受限滑鼠契約；目前仍無正常玩家畫面、音效或玩法同狀態對拍。
- 隔離 dosgolem 提交 `fb541129530206f16fda6edc7630b307e0d21ceb`，已推送 `github/codex/moo2-parity-20260930`。`git diff --check`、Go 格式、Python 語法、規格索引與本輪檔案 UID/GID=`1000:1000` 均通過；dosgolem 工作樹無 root-owned 產物或 `.md` 同名目錄。主專案既有 root-owned 快取／檔案未動；本專案一次性 Docker 容器均已結束，其他專案容器未碰觸。正版 ZIP、EXE、記憶體與完整終端均留本機私有工作區，未加入 Git。

## 2026-10-01：滑鼠重設與敏感度查詢

- 接手主專案 `5b59d3bb9d747d30d7a953ca54c1b2568a408a55`、隔離 dosgolem `fb541129530206f16fda6edc7630b307e0d21ceb`，兩分支起始乾淨。路由命中規格閘門、dosgolem 原版對拍與文件職責，載入對應入口；只修改隔離執行器及證據文件，未改 Go remake 玩法。
- 版控探針 `--mouse-reset`、`--mouse-sensitivity`、`--mouse-sequence` 在固定正版根層 417 檔及官方 1.31 EXE 上，以 DOSBox-X 2026.07.02 SDL2 heavy debugger 輔助擷取原版服務入口、同次返回及 caller record 消費。規格 251／252 逐一經 DRAFT→READY→實作→CONFORMED，補 `INT 33h/AX=0000h` 重設及 `001Bh` 目前敏感度查詢；模式中心是明示的平台規格近似，不宣稱完整滑鼠驅動還原。
- 初稿曾把舊缺檔探針的零敏感度設定誤套到目前完整資料。逐筆呼叫探針證明兩側目前都先走 `0000h → 001Bh`，查詢回 `50/50/50`，已修正現況並在規格及研究紀錄保留錯誤來源。兩個輔助擷取命令的收尾雜湊仍用舊 `*-logcpu.txt` 名稱而發出缺檔訊息；實際探針成功，改核對已產生的 `*-caller-logcpu.txt`，不是產品故障。
- 一次性 `golang:1.24-bookworm` 容器以 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -buildvcs=false ./... -count=1` 驗證兩階段，全套通過；最終私有測試輸出 SHA-256 `37ba0e0c1c59433302b0079af252c6b479945ff760b49f9a9d7e466ebf89cdb0`。`DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 自 LE entry 自然前進至第 1,545,719 步、dosgolem 高位 LE 線性 `0x24C31B` 的 `INT 33h/AX=0007h`，CX=`0`、DX=`04FEh`；輸出 SHA-256 `d630b6db5fc796dc91164b3e15219083ada2f8aed8bf0a7e263eb6b64708e666`。詳細原始輸入、工具版本、位址空間、私有收據及勘誤見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- 隔離 dosgolem 提交 `792b8d077d1d4611ae8246ddc9d61129e316b67a`，已推送 `github/codex/moo2-parity-20260930`。暫存差異檢查、Go 格式、Python 語法及規格索引核對通過；本輪檔案 UID/GID=`1000:1000`，dosgolem 無 root-owned 產物，主專案未見誤建 `.md` 目錄。本專案一次性 Docker 容器均已結束，其他專案容器保持原狀，原版素材與完整終端未入版控。
- 下一步只擷取原版 `0007h` 水平範圍設定的輸入、返回與後續使用，審查受限平台規格後由 dosgolem 重生驗證。正常玩家畫面、音效、受控亂數與 remake 玩法同狀態對拍仍未完成，完整 remake 目標持續進行。

## 2026-10-01：滑鼠範圍與非零設定

- 接手主專案 `af000310a2c6cbcaa96091671c0d364e95249dc9`、隔離 dosgolem `792b8d077d1d4611ae8246ddc9d61129e316b67a`，兩工作樹乾淨。上一輪為實質進展；本輪路由命中 dosgolem、規格閘門、平台規格優先、文件職責及解析回填，載入對應入口。沒有修改 Go remake 玩法或建立發行包。
- 固定正版根層 417 檔及官方 1.31 EXE，DOSBox-X 輔助探針 `--mouse-horizontal-range／--mouse-vertical-range／--mouse-set-sensitivity` 全部成功取得原版同次入口、返回與 record 消費端。規格 253／254 依 DRAFT→READY→實作→CONFORMED，補兩軸有號範圍與受控輸入限制、非零敏感度裁切與查詢。實體移動速度、粒度與完整驅動仍未知；詳細收據、工具、輸入雜湊與位址空間見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -buildvcs=false ./... -count=1` 兩階段最終全套通過。非零設定首次回歸因舊測試查位置後留下 CX=417／DX=122，卻預期下一次設定仍為零而失敗；明確設定該案例為 `1/0/0` 後，同一容器映像與命令乾淨重跑通過，初次輸出留本機私有 `workplace/full-test-254-first.txt`。一次文件補丁曾因同一檔案列兩次操作而被工具拒絕，合併該操作後正常套用，沒有部分實作或產品故障。
- `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 自 LE entry 自然越過 `7／8／1Ah`，目前第 1,546,160 步抵達 `INT 33h/AX=000Ch`；最終私有測試 SHA-256 `fe9cf5242d01be373fdea1c67d9869165446d18bdf267831524cb6af4cf0fd01`、自生診斷 SHA-256 `3f3af2c4aeef4fc15686b2a281ab35d8ac6b7c432c3016205277db6e85f9dfe7`。沒有正常玩家畫面或玩法同狀態對拍。
- 規格 230 保存歷史零值證據，現行非零拒絕邊界加明示回填及新規格連結；版控 `--check-mouse-spec-backlinks` 檢查原始鍵、狀態及舊規格標記，正常與刪標記必拒絕核對通過。Go 格式、Python 語法、規格索引與 `git diff --check` 通過。
- 隔離 dosgolem 提交 `438d6cc5971c3e212e0ce949e1ddd61de307f794`，已推送 `github/codex/moo2-parity-20260930`。本輪檔案 UID/GID=`1000:1000`，dosgolem 無 root-owned 產物，主專案無誤建 `.md` 目錄；本專案一次性 Docker 容器均已結束，其他專案容器未碰觸。原版素材、記憶體與完整終端未入版控。下一個最小行動為擷取原版回呼參數、返回與觸發鏈，再審查平台契約；完整 remake 目標持續進行。

## 2026-10-01：滑鼠事件回呼與兩條原檔啟動路徑

- 接手主專案 `7546b7855664d620c4534c2b06ab413046efbc47`、隔離 dosgolem `438d6cc5971c3e212e0ce949e1ddd61de307f794`，起始兩工作樹乾淨。路由命中 dosgolem、規格閘門、平台規格優先及文件職責，載入對應入口；收尾再次核對路由與狀態機。沒有修改 Go remake 玩法或建立發行包。
- 固定正版根層 417 檔與官方 1.31 EXE，以 DOSBox-X 輔助探針 `--mouse-callback／--mouse-callback-event／--mouse-set-position` 取得註冊、一般滑鼠移動命中、原版座標寫入及八位元組遠返回、位置設定同次返回。只保存足以接線的平台邊界；原版檔案與完整終端留本機。詳細雜湊、位址空間與工具版本見 `docs/re/dosgolem-moo2-intake-20260930.md`。
- 隔離 dosgolem 規格 255–257 各經 DRAFT→READY 後實作：受控滑鼠事件 FIFO、IF 等待、非重入、獨立有界堆疊及狀態恢復；CS 絕對 word 至 DS 的窄 CPU 形狀；`AX=4` 座標設定且不製造事件。合成回呼實際執行、寫出參數並遠返回，遮罩／按鍵／範圍／容量／重設／解除／架構恢復／既有時鐘 hook／錯誤返回等測試通過。規格 256–257 在限定範圍 CONFORMED；255 的完整座標／游標消費同狀態待驗，維持 READY。
- 三階段固定 EXE 的 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -buildvcs=false ./... -count=1` 均通過。最後一次收尾命令被檔名替換誤改為 `sha257sum`，測試與兩份診斷已成功；保留 `*-257-first*` 私有收據，修正為 `sha256sum` 後以同一映像、資料及命令乾淨重跑通過。一次規格補丁因段落行並非獨立行而拒絕，改用正確上下文後成功，未改產品契約。
- `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game` 無事件路徑第 6,216,999 步抵達 `INT 2Fh/AX=160Ah`；另加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1` 的設定後受控事件路徑回呼 started=1／completed=1，第 6,217,034 步抵達同一服務。最終全套測試 SHA-256 `33ca5589160c47f992c5cdeec08b6b65ded02dfc6a03fb705aa698c21a05f26e`、無事件診斷 SHA-256 `3b7e90bfcb64cfdbbbf9f834e7f7bd185ddc6f94fcee56cd8312c5fc812653c2`、受控事件診斷 SHA-256 `5cc4ec566b82857814f12e271e0dedd88f8f283835e9505cac0f7c6dc88324ba`。沒有正常玩家畫面或玩法同狀態完成宣稱。
- 下一個最小行動為核對 `INT 2Fh/AX=160Ah` 的原版返回與 caller；依公開介面審查最小支援，避免展開 OS 內部研究。完整 remake 與中文化目標持續進行。
- 隔離 dosgolem 提交 `7a50ae93495ed70d00e5fb2be1e129a569c5ef5f`，已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。Go 格式、Python 語法、三份規格索引、暫存差異及修改檔 UID/GID=`1000:1000` 核對通過；最初容器標準輸入未連接的稽核沒有執行，補 `docker run -i` 後確認有實際通過輸出。dosgolem 無 root-owned 產物，主專案無誤建 `.md` 目錄；本輪一次性 Docker 容器均已結束，其他專案資源未動。主庫本輪只提交現況／研究／歷程四份文件；正版輸入、記憶體及完整終端未加入 Git。

## 2026-10-01：限定未安裝 Windows 的平台查詢

- 接手主庫 `e5f874d137649c1db0abe1cfdec936c81615dda9`、隔離 dosgolem `7a50ae93495ed70d00e5fb2be1e129a569c5ef5f`，起始工作樹乾淨。路由命中 dosgolem、規格閘門、平台規格優先及文件職責，已載入入口，收尾重新核對路由。未修改 Go remake 玩法或建立發行包。
- `startup_probe_131.py --windows-version` 取得原版返回及 caller 分支，審查限定契約後接入明示 MOO2 DOS 設定；規格 258 在限定查詢範圍 CONFORMED。其他設定與未知功能仍拒絕，原版資料留本機。
- 固定原檔全套 `go test -buildvcs=false ./... -count=1` 通過，兩條自然路徑下一停點均為 `INT 31h/AX=0500h`。首次 shell 巢狀引號在容器啟動前被拒絕，修正後乾淨重跑。單項測試初態改具名暫存器索引後以 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/machine -run TestMOO2WindowsVersionAbsentPreservesState -count=1` 通過；首次重跑程序上限不足，限制平行度修正。其後讀取大型壓縮診斷一次全部展開造成記憶體不足，改逐行串流核對，原診斷不變。文件更新初次錯把反引號跳脫字元保留，定位字串未命中；修正後重跑既有文件編輯。皆為命令／驗證問題，非產品故障。
- 詳細雜湊、位址空間與工具版本見研究紀錄。完整 remake 目標持續進行；下一步只核對公開 DPMI 記憶體資訊介面及原版返回。
- 隔離 dosgolem 提交 `68e0ebca82cdcb436c7910bc676a818167215343`，已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。Go 格式、Python 語法、規格索引及暫存差異檢查通過；本輪修改檔 UID/GID=`1000:1000`，dosgolem 無 root-owned 檔案，主專案無誤建 `.md` 目錄。相關一次性 Docker 容器均已清空；其他專案資源未動，正版素材及完整終端未入版控。

## 2026-10-01：DPMI 可用記憶體資訊

- 接手主庫 `685c1ee097b5c018718e9affee339f5b056e1669`、隔離 dosgolem `68e0ebca82cdcb436c7910bc676a818167215343`，起始乾淨。路由命中 dosgolem、規格閘門、平台規格優先、文件職責及解析回填，載入各入口。未修改 Go remake 玩法或建立發行包。
- `startup_probe_131.py --free-memory` 取得原版 48-byte 返回區塊、同次架構欄位及外層第一欄位容量消費。規格 259 先 DRAFT，再依公開 DPMI 契約及 caller 審查轉 READY 後接入配置器一致的容量模型；測試與兩條原檔路徑驗收後 CONFORMED。其他分頁欄位保持規格允許的未知，不硬編原版容量。詳細雜湊與位址空間見研究紀錄。
- 一次性既有 Go 映像，以固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過。初稿測試使用不存在的 `Descriptor.Default32` 而編譯失敗，刪除多餘欄位後同命令乾淨重跑，保留第一次私有輸出。原檔無事件／設定後受控事件兩條路徑自行越過 `0500h`，下一停點均為 `C1 7D F4 04`。
- 舊規格 258 回填 259，探針 `--check-free-memory-spec-backlinks` 正常與刪除標記必拒絕驗證通過。一次補丁因段落非獨立行拒絕，改正確定位套用；一次負向護欄 shell 引號影響搜尋字串，修正後重跑，未改平台契約。
- 正常玩家畫面、音效與玩法同狀態對拍仍未完成，規格 255 維持 READY。下一步只核對新 CPU 記憶體移位指令形狀，完整 remake 目標持續進行。
- 隔離 dosgolem 提交 `a6a7b79a60c9656f93b216b539707ff9b52b5671`，已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。Go 格式、Python 語法、規格索引、正常及負向回填護欄、暫存差異核對通過；本輪修改檔及測試輸出 UID/GID=`1000:1000`。dosgolem 無 root-owned 殘留，主庫無誤建 `.md` 目錄，相關一次性 Docker 容器均已結束；其他專案資源未動。原版輸入、記憶體與完整終端未入版控。

## 2026-10-01：堆疊 dword 的 SAR 指令

- 接手主庫 `069b1afa302dcac89cdbe7f364e110da6ac20b76`、隔離 dosgolem `a6a7b79a60c9656f93b216b539707ff9b52b5671`，起始兩工作樹乾淨。路由命中 dosgolem、規格閘門、平台規格優先、文件職責及解析回填，載入各入口，沒有修改 Go remake 玩法或建立發行包。
- CPU 規格 260 先 DRAFT，再依原檔 `C1 7D F4 04` 與公開 Intel 契約轉 READY 後實作，補無前綴堆疊 dword SAR。正負值、計數遮罩、段／哨兵、旗標策略、未知形狀與錯誤存取測試通過；原版輔助探針取得同次 `32 → 2` 與 caller。原版未定義 AF=1、執行器 AF=0 明示保留，定義旗標相符。
- `startup_probe_131.py --sar-stack-memory` 首次候選位址換算少一頁而未命中，保留私有終端；修正後實際原始 bytes 與同次返回通過。加入原版樣本的第一次測試夾具把段排列錯放進 CPU 陣列，被 SS 可寫檢查拒絕；改具名索引，並修正規格 258 舊測試夾具及追加勘誤。CPU／服務程式未為此放寬檢查。
- 固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 最終全套通過；保留加入樣本前通過及夾具失敗輸出。兩條原檔自然路徑均越過 SAR，下一停點為 `INT 21h/AH=4Eh` 的當前查詢輸入。詳細雜湊、工具與位址空間見研究紀錄及規格 260。
- `--check-sar-stack-spec-backlinks` 正常與刪除舊標記必拒絕通過，規格 259 回填 260。正常玩家畫面、音效與玩法同狀態收據仍未完成，規格 255 維持 READY。下一步只擷取搜尋輸入、DTA 狀態及原版返回，完整 remake 目標持續進行。
- 隔離 dosgolem 提交 `8f4c579beda504a66aefd74186fa6d1fa8d3d072`，已推送 `github/codex/moo2-parity-20260930`。Go 格式、Python 語法、規格索引、既有及新增回填護欄、暫存差異核對通過；修改檔及最終測試輸出 UID/GID=`1000:1000`。dosgolem 無 root-owned 殘留，主庫無誤建 `.md` 目錄，相關一次性 Docker 容器均已結束；其他專案資源未動。原版輸入、記憶體與完整終端未入版控。

## 2026-10-01：目前目錄前綴的 DOS 搜尋

- 接手主庫 `f2f88e60d38050b0f7be314cc3405274271e2ce9`、隔離 dosgolem `8f4c579beda504a66aefd74186fa6d1fa8d3d072`，接手時兩工作樹乾淨。路由重新核對 dosgolem、規格閘門、平台規格優先及逆向結論回填入口；未修改 Go remake 玩法或建立發行包。
- 版控診斷探針記錄 `.\simtex.lbx` 與 DTA，原版 `--find-current-directory` 捕獲同次缺檔 `12h`／CF=1。實際 ZIP 查無該檔，不添加合成素材。規格 261 先 DRAFT，依公開平台來源及原版返回審查後 READY，再接單一前綴；提供者路徑安全不放寬。DTA 保留區原版 `+0Ch` 自增差異明示，不追 helper 內部。
- 既有容器中固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過；兩條自然路徑均越過搜尋，下一停點為高位 LE `0x114F43` 的 `66 85 C0`。測試與診斷雜湊見研究紀錄；規格 261 只在限定返回及公開結果欄 CONFORMED，255 仍 READY，正常玩家畫面／音效／玩法同狀態未完成。
- 規格 260 回填 261，219 連到延伸範圍。下一步只核對公開 CPU TEST 的位元寬度、旗標及最小支援，完整 remake 目標持續進行。
- 隔離 dosgolem 提交 `f044744b17821d429275e73c4ebc7b477a957638`，已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。Python 語法、Go 格式、索引、正常及負向回填護欄、暫存差異通過；修改檔與測試輸出 UID/GID=`1000:1000`，無 root-owned 或誤建 `.md` 目錄殘留。首次指定檔案暫存命令因 `/workplace/` 忽略規則返回非零；回讀 index 確認既有受版控探針與其餘指定七檔均已正確暫存，沒有強制加入診斷產物。本輪一次性容器均已結束；依掛載路徑確認剩餘 Go 容器屬 yuan／fd2，未動其他專案資源。正版素材、完整終端與記憶體未入 Git。

## 2026-10-01：16 位元暫存器 TEST

- 接手主庫 `ce16ab6821334f9f42f50fd3068ff84a376e8f11`、隔離 dosgolem `f044744b17821d429275e73c4ebc7b477a957638`，兩工作樹乾淨。上一輪分類為實際進展；路由命中 dosgolem、規格閘門、平台規格優先、文件職責及解析回填，已載入入口。既有 image inspect 的 Entrypoint 模板因缺 key 被拒絕，改讀 Config 核對；屬控制面查詢問題，不是工具鏈或產品故障。
- 規格 262 先 DRAFT，依公開 Intel TEST／旗標契約及固定原檔形狀轉 READY，接全部暫存器 word TEST。`startup_probe_131.py --test-word-register` 取得原版同次架構與第一個 JNZ；最小 bytes／具名暫存器輸入加入回歸，沒有追 runtime helper。
- 固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 初次及加入實際樣本後最終全套均通過；兩條自然原檔路徑自行越過 TEST，下一停點為高位 LE `0x222C9E` 的 `OUT DX,AL`，DX=`03C6h`、AL=`FFh`。詳細雜湊與位址空間見研究紀錄與規格 262。
- 規格 261 回填 262，CPU 範圍 CONFORMED；規格 255 仍 READY，正常玩家畫面／音效／玩法同狀態未完成。下一步核對公開 VGA DAC 及既有調色盤埠路由，完整 remake 目標持續進行。
- 隔離 dosgolem 提交 `f9e043fb41de8fff6e2ec3b49b0fd0695712af12`，已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。Python 語法、Go 格式、索引、正常與負向回填護欄、暫存差異通過；本輪修改檔與最終測試輸出 UID/GID=`1000:1000`。無 root-owned 或誤建 `.md` 目錄殘留，本輪一次性容器均已結束；剩餘 Go 容器依掛載確認屬 fd2，未動其他專案。主庫只提交現況／研究／歷程四份文件，原版輸入、記憶體及完整終端不入 Git。

## 2026-10-01：標準 VGA 像素遮罩

- 接手主庫 `47ea5d6787b466736a33fdff5cfa659ac32f289c`、隔離 dosgolem `f9e043fb41de8fff6e2ec3b49b0fd0695712af12`，兩工作樹乾淨。上一輪分類為實際進展；路由核對 dosgolem、規格閘門、平台規格優先、文件職責及解析回填，載入入口，完整 remake 目標持續進行。
- 規格 263 先 DRAFT 並加入索引，再依 IBM 初始化／RGB 契約與 DOSBox-X 標準遮罩介面審查轉 READY 後實作。接 `03C6h` 讀寫、Palette／平面 RGB／LE Palette 及兩種保存，舊 gob v2 缺欄位仍保留原行為；屬硬體規格近似，不改 Go remake 玩法或平台 driver 內部。
- 有界 `--vga-pel-mask` 原版輔助探針取得實際使用點、原始 bytes、DX／AL 與下一指令。遮罩／消費端／保存相容測試及固定 EXE 全套測試通過；兩條自然原檔路徑自行越過寫入，下一停點是高位 LE `0x222CCB` 的 `F6 F3`。精確工具、位址空間、命令與雜湊見研究紀錄／規格 263。
- 建立編輯命令時一次將 JavaScript 函式放進 store 而遭序列化拒絕，發生於命令啟動前；改用可序列化字串後同一入口通過。IBM 首個鏡像／猜測 PDF 入口及短 commit 原始碼 URL 讀取失敗，改用可讀的 IBM 鏡像與官方模型頁，屬來源存取問題。串流診斷第一次篩選太寬截斷，改只輸出精確停點，不把截斷當產品缺陷。
- 規格 262 回填 263；Python 語法、索引、既有護欄及新護欄正常／缺舊標記／缺定位必拒絕通過。規格 263 限定範圍 CONFORMED，255 維持 READY，正常玩家畫面／音效／玩法同狀態未完成。下一步只核對公開 CPU byte DIV 與既有解碼，沒有新建發行包。
- 隔離 dosgolem 提交 `3cfd84be85738450b65102f46450d46ae91429f1`，已推送 `github/codex/moo2-parity-20260930`。Go 格式、Python 語法、規格索引、相關回填護欄與暫存差異核對通過；修改檔及驗證輸出 UID/GID=`1000:1000`。本輪一次性容器均已結束，剩餘 Go 容器掛載屬 fd2，未動其他專案；原版輸入、記憶體與完整終端未入 Git。
- 完整主庫擁有權自檢另查到歷史 root-owned 的 `go.sum`、`lbxinfo` 與 `.docker-cache/` 產物，本輪未寫入或改動這些路徑；隔離 dosgolem 與本輪輸出均無此問題，主庫沒有誤建的 `*.md` 目錄。沒有對儲存庫做遞迴 chown。

## 2026-10-01：位元組暫存器的無號除法

- 接手主庫 `73754675cd5ea6ea2b184ed479e9c0c091d9a675`、隔離 dosgolem `3cfd84be85738450b65102f46450d46ae91429f1`，起始兩工作樹乾淨。上一輪分類為實際進展；路由命中 dosgolem、規格閘門、平台規格優先、文件職責，建立文件／下結論前補載解析回填入口。
- 規格 264 先 DRAFT 並入索引，依 Intel DIV 與既有暫存器／錯誤策略審查 READY 後實作。接全部八種 byte 除數，保存來源別名與高半部；未知 F6 診斷保留實際 ModRM。除法錯誤與未定義旗標近似明示，沒有新玩法規格或 Go remake 玩法變更。
- 原版有界 `--div-byte-register` 捕獲同次 AX=0／BL=100、商餘皆 0 與第一個 OUT。原版 ZF 清除、執行器保留策略有差異，未隱藏或猜補。範圍規格初次檔名少了 `-and-mvp` 而讀取失敗；找到既有正確檔名後載入，屬文件定位錯誤，不是產品缺陷。
- CPU 測試、初次固定原檔全套與加入實際樣本後的最終全套均通過；兩條自然路徑自行越過 DIV 與後續調色盤寫入，下一停點為高位 LE `0x228C54` 的 `INT 10h/AX=4F05h`，BX=0、DX=5。實際命令、工具版本與輸出雜湊見研究紀錄及規格 264。
- 規格 263 回填 264，Python 語法、索引與新護欄正常／缺原始定位／缺舊標記必拒絕通過。264 限定範圍 CONFORMED，255 維持 READY；正常玩家畫面／音效／玩法同狀態尚未完成。下一步核對 VBE 視窗服務及顯存模型，完整 remake 目標持續進行，未建立發行包。
- 隔離 dosgolem 提交 `39ac121bb2f2343809a9e3ecc67836a4d510faa1`，已推送 `github/codex/moo2-parity-20260930`。Go 格式、Python 語法、規格索引、全部相關回填護欄及暫存差異核對通過；本輪修改檔與輸出 UID/GID=`1000:1000`。隔離副本無 root-owned／誤建目錄，本輪一次性 Docker 容器均已結束；主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 擁有權問題保持未動，沒有誤建 `*.md` 目錄，沒有遞迴 chown 或清理其他專案。

## 2026-10-01：VBE 視窗與實際顯存消費

- 接手主庫 `d0b8a27dbed92f017c7ab7307c9aac3c40ea16f0`、隔離 dosgolem `39ac121bb2f2343809a9e3ecc67836a4d510faa1`，兩工作樹乾淨。路由命中 dosgolem、規格閘門、平台規格優先、文件職責與逆向解析回填，載入入口；上一輪為實際進展，完整 remake 目標持續進行。
- 規格 265 先 DRAFT 並入索引，再依公開 VBE 契約及固定模式資訊審查 READY。接有限視窗 A、CPU／DPMI 同一顯存映射、區段保存、word／dword 跨界、零起點索引／RGB 消費和模式重設；未修改 Go remake 玩法或建立發行包。沒有只補成功返回，沒有追 S3／硬體時序。
- 原版有界 `--vbe-window-control` 確認第 5 個區段切換的同次返回與 POPAD／RET；原版旗標 `206h` 與 dosgolem 自然輸入 `246h` 的完整狀態差異明示。具名服務樣本及顯存測試通過，固定原檔的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過。
- 兩條自然原檔路徑自行切換區段 5–9 並寫入 307,200 bytes，下一停點為高位 LE `0x228CA7` 的非零顯示起點 `4F07h`，Y=512。目前零起點快照不是正常玩家畫面。實際命令、工具、位址空間與雜湊見研究紀錄及規格 265。
- 265 限定範圍 CONFORMED，239／264 回填延伸；新護欄正常及缺定位／缺舊標記必拒絕通過，全部既有回填函式亦通過。255 仍 READY，正常玩家畫面／音效／亂數／玩法同狀態未完成；下一步只核對非零顯示起點及顯存圖像消費。
- 接手讀檔時誤用舊檔名，依 find 與既有索引找到正確入口後讀取；一個文件編輯命令因 JavaScript 字串語法錯誤在啟動前被拒絕，改用補丁後完成。這些是文件定位／編輯控制面問題，測試與原版探針沒有因此失敗或放寬產品檢查。
- 隔離 dosgolem 提交 `dea1659135461bd75c796657cbaf202674d32b47`，已推送 `github/codex/moo2-parity-20260930`。Go 格式、Python 語法、索引、回填正負護欄與暫存差異通過；修改檔及本輪測試／診斷輸出 UID/GID=`1000:1000`。一次性 Docker 容器均已結束，隔離副本無 root-owned 或誤建 `.md` 目錄；主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 擁有權問題未動。未清理其他專案、未做遞迴 chown，原版素材／完整終端／記憶體沒有加入 Git。

## 2026-10-01：非零 VBE 起點與圖像消費

- 接手主庫 `b3d727e4165e89719efd32b524ed4f47e693daaf`、隔離 dosgolem `dea1659135461bd75c796657cbaf202674d32b47`，兩工作樹乾淨。上一輪是實際進展；路由命中 dosgolem、規格閘門、平台規格優先、文件職責與逆向回填，載入入口；完整 remake 目標持續進行。
- 規格 266 先 DRAFT 並入索引，以 VBE 公開契約與原版 Y=512 同次返回審查 READY 後接有效起點、讀回、同一起點索引／RGB。啟用前精確全零返回保留，模式重設回零；未知形狀及越界拒絕。未修改 Go remake 玩法、建立發行包或追 S3／回掃內部。
- `startup_probe_131.py --vbe-display-start` 命中實際原始 bytes、同次架構與 POPAD／CLD／RET。衍生探針最初的字串替換沒有匹配兩個保護條件；檢視生成區塊後在首次執行前訂正，探針本身一次通過。CPU／DPMI 寫不同頁、整頁消費、原版返回、邊界、拒絕與重設測試及固定原檔全套測試均通過。
- 兩條自然原檔路徑自行越過 Y=512，下一停點是高位 LE `0x234B10` 的 `66 83 C3 18`。當次有效頁索引仍全零，PNG 已實際檢視為黑圖，未聲稱正常玩家畫面完成；圖片與完整診斷留本機。工具、命令、位址空間、私有收據雜湊與限制見研究紀錄／規格 266。
- 237／265 回填 266，新護欄正常及缺定位／缺任一舊標記必拒絕通過，全部既有回填函式亦通過；修正先前 264／265 索引未跟隨實際收據的過期摘要。266 有限範圍 CONFORMED、255 仍 READY；下一步核對 word ADD，正常玩家畫面／音效／亂數／玩法同狀態仍未完成。
- 隔離 dosgolem 提交 `6da3a10b07d8b272eb55baa52ef8b52d3704ae03`，已推送 `github/codex/moo2-parity-20260930` 且回讀分支相符。Go 格式、Python 語法、索引、回填護欄、暫存差異及擁有權核對通過，本輪修改檔與測試／診斷／圖片輸出 UID/GID=`1000:1000`。本輪一次性 Docker 容器均已結束；隔離副本無 root-owned／誤建 `.md` 目錄，主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 問題未動，沒有誤建 `.md` 目錄。未遞迴 chown 或清理其他專案，原版輸入、完整終端與 PNG 未加入 Git。

## 2026-10-01：word 暫存器的帶符號立即值加法

- 接手主庫 `4e90f02242cbaea206fe53035f1cbb072dd50baf`、隔離 dosgolem `6da3a10b07d8b272eb55baa52ef8b52d3704ae03`，兩工作樹乾淨。上一輪是實際進展；路由命中 dosgolem、規格閘門、平台規格優先、文件職責與逆向回填，載入入口；完整 remake 目標持續進行。
- 規格 267 先 DRAFT 並入索引，依 Intel ADD 與原檔形狀／同次樣本審查 READY 後接八種 word 暫存器帶符號立即值。沿用 add16 與既有拒絕邊界，保留高半部與其他狀態；未修改 Go remake 玩法、建立發行包或追 helper 內部。
- `startup_probe_131.py --add-word-register` 取得同次 `F0h → 108h`、旗標 `246h → 202h` 與 CMP／JL。八目的／全部 imm8／算術邊界／別名與高半部／拒絕／既有操作／原版消費端測試通過；包含實際樣本後固定原檔全套測試通過。
- 兩條自然原檔路徑自行越過 ADD／CMP／JL，下一停點是高位 LE `0x234B43` 的 `66 F7 EB`。圖片雜湊與上輪已檢視黑圖相同，沒有重跑無必要的目視檢查或聲稱玩家畫面完成；精確命令、工具、位址空間與收據雜湊見研究紀錄／規格 267。
- 266 回填 267，新增正負護欄與全部既有回填函式通過，索引同步目前結果。267 有限 CPU／首個消費範圍 CONFORMED，255 仍 READY；正常玩家畫面／音效／亂數／玩法同狀態仍未完成。下一步只核對公開 IMUL 的 word 有號乘積、CF／OF、未定義旗標與原版使用點。
- 收尾語言自檢首次誤把繁體也使用的「果」列入簡體字元集合而失敗；修正檢查集合後，以同一容器入口乾淨重跑通過。這是驗證腳本誤判，沒有修改 CPU 或放寬算術測試。
- 隔離 dosgolem 提交 `c0c1487222b5d0544f9551e78bf76c0a9e7fae43`，已推送 `github/codex/moo2-parity-20260930` 且回讀相符。Go 格式、Python 語法、索引、回填正負護欄與暫存差異核對通過，修改檔及本輪 CPU／全套測試／診斷／圖片輸出 UID/GID=`1000:1000`。一次性 Docker 容器均已結束，隔離副本無 root-owned 或誤建 `.md` 目錄；主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 擁有權問題未動，沒有誤建 `.md` 目錄。未遞迴 chown 或清理其他專案，原版輸入、完整終端與 PNG 未加入 Git。

## 2026-10-01：單運算元的 16／32 位元有號乘法

- 接手主庫 `6d9e3b7d9f31baed3b4ac5ecdff8b6f735448b4e`、隔離 dosgolem `c0c1487222b5d0544f9551e78bf76c0a9e7fae43`，兩工作樹乾淨。上一輪是實際進展；路由命中規格閘門、平台規格優先、文件職責與逆向回填並載入入口，本輪推進完整 remake 目標。
- 規格 268 先 DRAFT 並入索引，依 Intel IMUL、固定原版架構與有限使用點審查 READY 後實作 word 有號乘法。自然原版前進到 dword 同類缺口，再以規格 269 另走 DRAFT／READY，實作完整 EDX:EAX。保留既有 word／dword 無號及兩／三運算元乘法與拒絕界線；沒有修改 Go remake 玩法或建立發行包。
- 兩種輔助擷取均命中固定指令、同次乘積／CF／OF 與第一個存值／XOR。原版未定義 ZF 清除，dosgolem 保存；資料布局與最小測試重定位也明示，不稱完整狀態或旗標逐值一致。八來源、正負邊界、完整積、別名與其他狀態哨兵、拒絕及回歸，CPU 和固定原檔全套測試均通過。
- 兩條自然原版路徑自行越過兩種乘法，下一停點為高位 LE `0x234C9A` 的 CMP dword ModRM 0D；顯存多切換一次，Bank=7、BankSets=6。PNG 與先前已檢視黑圖逐位元相同，不重做不必要目視檢查或宣稱玩家畫面完成。實際命令、工具、原始位址、完整輸入與收據雜湊見研究紀錄／規格 268、269。
- 267／268 後續停點回填，兩個新護欄的正例與缺定位／缺舊標記必拒絕、全部既有回填函式、索引／格式／語法／繁體中文新增文字檢查通過。268／269 僅有限 CPU 與消費端 CONFORMED，255 仍 READY；正常玩家畫面／音效／亂數／玩法同狀態仍未完成，下一步核對實際 CMP 輸入與第一個分支。
- 接手讀檔初次誤用目錄／檔案路徑，依隔離副本實際入口訂正後讀取；屬定位問題，沒有修改產品行為或放寬測試。

- 隔離 dosgolem 提交 `25b8e79e332bb8bb1359284e8d10822462a70a9b`，已推送 `github/codex/moo2-parity-20260930`，回讀分支相符；上游本機 origin 未回寫。只提交自製 CPU、最小樣本、規格與擷取入口，原版素材、完整終端／記憶體、gzip 和 PNG 未加入 Git。
- 本輪修改檔與 CPU／全套測試／自然診斷／PNG 輸出 UID/GID=`1000:1000`。一次性 Docker 容器均已結束，專案掛載篩選無執行中或停止殘留；隔離副本無 root-owned／誤建 .md 目錄。主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 擁有權問題未動，沒有誤建 .md 目錄；未遞迴 chown、未清理其他專案資源。

## 2026-10-01：目的比較與實際分支收據

- 接手主庫 `0aaab1ada4fe6b1a356a748502c29238073eb577`、隔離 dosgolem `25b8e79e332bb8bb1359284e8d10822462a70a9b`，兩工作樹乾淨。上一輪為實際進展；路由命中 dosgolem、規格閘門、平台規格優先、文件職責與逆向回填並載入入口，完整 remake 目標持續進行。
- 270 先 DRAFT 並入索引，擷取兩側目的 dword 與原版完整旗標／實際 JGE；READY 後接 39 記憶體目的比較。原版自然前進到 word 形式，再以 271 另走 DRAFT／READY，接 66 39 的記憶體與暫存器目的。不寫回比較資料，沿用地址／分段與算術層；未改 Go remake 玩法或建立發行包。
- 初次 LOG 1 只記錄分支前，改成有界 LOG 2／分支後定位再重跑，才據以 READY。270 首次 CPU 失敗為局部 F2 前綴判斷誤讀，回 DRAFT 依全域拒絕重新審查 READY；第二次為既有 3B 回歸樣本漏設資料段描述子，補齊環境後以同一容器命令乾淨重跑。沒有放寬正式前綴或刪除回歸，失敗輸出與原因保留在 270。
- 八來源／word 八目的、16 邊界值、六旗標、精確寬度／高半部、分段／索引／別名、非寫回與部分讀取／前綴拒絕均通過；原版架構樣本與第一個 JGE／JL 實際不取分支已核對。兩次固定原檔全套 Go 測試與兩條自然重跑均完成，下一停點為高位 LE `0x21CA3D` 的 word INC。實際命令、工具／原始位址、完整輸入與雜湊見研究紀錄及工具規格。
- 269 → 270 → 271 回填，正例與缺定位／缺舊標記必拒絕及全部既有回填函式通過。271 的正式 word 正例替代 270 原 word 前綴拒絕負例，其餘拒絕不放寬；270／271 僅有限 CPU／首個分支 CONFORMED，255 仍 READY。
- 依 LOG 實際範圍，訂正工具規格 267 把 JL 分支前資料寫成原版實際跳轉的過度聲明。原 ADD／旗標／CMP 收據及定位保留，JL 目標位址改標為相同架構的 dosgolem 測試／Intel 條件所支持；主庫歷史研究追加勘誤，不抹除舊結論來源。
- 四份最終 PNG 與先前已檢視黑圖逐位元相同，不重做無必要目視檢查或聲稱可玩完成。正常玩家畫面、音效、受控亂數與玩法同狀態仍未完成，下一步只核對 word INC 目的、CF 保存與後續消費。
- 映像檔首次 inspect 使用不存在的 Entrypoint 欄位而失敗，改讀 Config 後確認既有工具鏈；未另建映像或改動 runtime。

- 隔離 dosgolem 提交 `a1a80765b6b1d527e2874057377afd1caab7a920`，已推送 `github/codex/moo2-parity-20260930` 並回讀相符；本機上游 origin 未回寫。只提交自製 CPU／最小樣本、規格、回填與只讀觀測，原版素材、完整終端／記憶體、gzip 與 PNG 未加入 Git。
- 新增文字繁體中文、Go 格式、Python 語法、索引、回填正負護欄及全部既有回填函式通過；修改檔與本輪 CPU／全套／兩側原版觀測／PNG 輸出 UID/GID=`1000:1000`。一次性 Docker 容器均已結束，專案掛載篩選無執行中或停止殘留；隔離副本無 root-owned／誤建 .md 目錄。主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache/` 擁有權問題未動，沒有誤建 .md 目錄；未遞迴 chown、未清理其他專案資源。

## 2026-10-01：接通 word 記憶體遞增，原版首次顯示 Simtex 標誌

- 開工主庫 `3d88548eb9deec73c81f72befb0497fbf05351b6`、隔離工具 `a1a80765b6b1d527e2874057377afd1caab7a920`，兩庫工作樹乾淨並與遠端一致。重新比對知識路由，套用規格閘門／公開平台契約／dosgolem 原版對拍，玩法 RE 閘門保持關閉。
- 建規格 272 同次加入索引；只讀觀測原版 word=0、CF=0，DOSBox-X 有界樣本確認 0→1、282h→202h、下一 A1 實際載入與旗標保持。DRAFT 審查→READY 後才接 word FF /0，保留 /1 DEC。
- Docker 內 CPU 與固定原檔全套 Go 測試、原版重定位樣本、CF 兩初值與五旗標、分段／索引／環繞／別名、精確兩 bytes、受控讀寫失敗／不發布旗標、既有 DEC／dword INC／DEC 均通過。部分匯流排寫入模型明示，未冒稱原子例外；回填護欄正負例、全部既有回填函式、Python 語法、Git 差異及擁有權均通過。精確命令、雜湊與地址基準見研究入口及規格。
- 兩條 dosgolem 原版路徑均再前進 861,701 指令，無事件第 7,684,074 步、受控事件第 7,684,109 步，停於高位 LE `0x21C2D6` 的短 JS。擷取本次 EAX=2、EBX=Eh、ECX=10000h、EDX=1、flags=202h；受控回呼完成。
- 640×480 PNG 已實際檢視為白底黑色 Simtex 啟動標誌，首次非黑圖；StartY=0、Bank=2、BankSets=17、Writes=921600、DisplaySets=2。PNG／原版完整資料／終端皆留本機，未加入 Git；尚未進入主選單或玩家操作，不把 CPU 通過寫成完整 remake。
- 隔離工具已提交 `3a0c6b4cb475acc75588badff6436fd3612800d4`（7 檔），並推送 `github/codex/moo2-parity-20260930`；未推本機來源 `origin`。主庫只更新 CONTEXT／WORKLIST／WORKLOG／既有研究紀錄，研究連結鎖定該工具提交。規格 272 限定 CONFORMED，255 仍 READY，下一步為公開 Jcc／短 JS 實際 SF 與首分支；不深挖 runtime／圖形 helper。
- 本批 Docker 一次性容器皆結束並清空；隔離副本無 root-owned 檔與錯誤 `.md` 目錄，本輪來源及全部輸出 UID/GID=1000:1000。主庫既有 `go.sum`、`lbxinfo`、`.docker-cache` 的歷史 root 擁有權未改動。

## 2026-10-01：短符號分支與兩條 20.1M 原版路徑

- 開工主庫 `07a8554e2da7ef4915cd94dfd26fafe5b59d9734`、工具 `3a0c6b4cb475acc75588badff6436fd3612800d4`，均乾淨並與遠端一致；命中並載入規格閘門／平台契約／dosgolem 路由，玩法 RE 閘門保持。
- 規格 273 建檔同次入索引；公開 Jcc 與原版 JS 不取／JNS 取分支、全擷取狀態保持具備後，DRAFT→READY 才新增兩 opcode。兩工具 EAX／ECX／後續旗標初值不同，明列差異；只比 SF 條件，不宣稱整段時間／布局一致。
- Docker 內 CPU、全套固定原檔 Go 測試、分支全旗標／全位移／環繞／截短／前綴拒絕／狀態保持、原版最小控制流、既有近分支均通過；回填護欄正負例、全部既有回填函式、Python 語法、差異與擁有權核對通過。精確命令與收據在研究入口／規格。
- 首輪探針碰 8M 上限且該路徑未擷取 PNG，外層雜湊命令退出 1；分類為驗證腳本問題，原始失敗收據另存。修正診斷有限步數及共用 VBE 擷取，無效／超界設定拒絕；相同映像乾淨重跑全套與 16M 兩條路徑通過，實際檢視 MicroProse 展開動畫。16M 仍達上限，續以明示 50M 觀測，不改 CPU／原版輸入／虛擬時間，不重跑已通過全套。
- 兩條原版路徑在 20,100,561／20,100,596 步停於高位 LE `0x239A42` 的 word XOR，較原 JS 再前進 12,416,487 指令；受控事件回呼已完成。途中 Simtex／MicroProse 圖像成立，當前 PNG 與已檢視黑圖相同；主選單、玩家操作、音效／受控亂數與完整 remake 同狀態未完成。
- 工具提交 `5307099b55cd239f9ea5badd2ca42a245d951404`（7 檔）已推送 `github/codex/moo2-parity-20260930`，未推本機 `origin`；主庫僅四份現況／歷程文件，研究連結鎖定該工具提交。273 有限 CONFORMED、255 READY；下一步公開 word XOR、DI／旗標及下一 MOV ES，不追平台內部。
- 本批一次性 Docker 容器收尾時清空；隔離副本無 root-owned 與錯誤 `.md` 目錄、本輪檔案及輸出 UID/GID=1000:1000。主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache` 的 root 擁有權未改。

## 2026-10-01：接通 word 暫存器 XOR 與 MOV ES 消費

- 開工主庫 `9faa19e1676ccefddbc672b9410102a6f54b9d2d`、工具 `5307099b55cd239f9ea5badd2ca42a245d951404` 均乾淨並與遠端一致；命中規格閘門／平台契約／dosgolem 及收尾文件職責路由。玩法 RE 閘門保持。
- 規格 274 同次建檔入索引，原版 word XOR 的完整目的／高半部、定義旗標與下一 MOV ES 已取得後，DRAFT→READY 才實作。AF 清除為既有模型近似，兩工具初始布局不同，不冒稱全旗標或完整同狀態。
- 首輪新 word 測試通過，舊 dword 回歸樣本誤填 SF 造成假失敗；依公開契約確認 CPU 正確，只修測試期望並保存首輪失敗收據。相同映像／命令乾淨重跑 CPU 通過；固定原版全套 Go 測試及兩條 50M 上限自然路徑通過，下一停點高位 LE `0x239A47`、INT 2Fh/AX=1684h。精確命令、全輸入、原始定位與雜湊見研究入口及規格。
- 回填 273→274，缺定位／缺舊標記拒絕、全部既有回填函式、Python 語法、差異、繁體文字及擁有權通過。274 限定 CONFORMED、255 READY；主選單／玩家操作、音效、受控亂數與完整 remake 未完成。
- 工具提交 `45b46688666ff21d38402870d8f585b563dc2d58`（7 檔）推送 `github/codex/moo2-parity-20260930`；未推本機 origin。主庫只更新四份現況／歷程文件，研究連結鎖定工具提交；原版資料、完整終端、gzip、PNG 不加入 Git。
- 本批一次性 Docker 容器收尾清空；隔離副本無 root-owned／誤建 .md 目錄，本輪檔案及輸出 UID/GID=1000:1000。主庫歷史 `go.sum`／`lbxinfo`／`.docker-cache` 擁有權未改。下一步只核對平台查詢邊界與有限 caller 契約。

## 2026-10-01：接通未安裝 VTD 的空入口查詢

- 開工主庫 `9d539282180c10f5485fee01aa4d0b7733aceb9f`、工具 `45b46688666ff21d38402870d8f585b563dc2d58`，均乾淨並與遠端一致；命中規格閘門／平台規格／dosgolem／文件職責路由，沿用逆向重製技能，玩法 RE 閘門保持。
- 規格 275 同次建檔入索引，原版完整空入口返回與 caller 保存／讀取後，DRAFT→READY 才實作限定查詢。首輪 LOG 2 僅實際執行第一筆保存，保留原收據後用同映像 LOG 8 補足，沒有把待執行列升格成原版結果。
- 服務狀態／拒絕測試通過；最小 caller 首輪缺 DS 的可寫描述子，CPU 拒絕正確，只修測試環境並保留失敗收據。相同映像／命令乾淨重跑通過；固定原版全套與兩條自然路徑通過，查詢前後完整觀測欄位保持，高半部與原版布局差異明示。
- 兩條較舊停點再前進 835 指令，新的高位 LE `0x239AE8` 拒絕 OUT 43h/AL=34h。原版 record 保存／讀取與兩筆 word 最小測試不代表自然 caller 的 record 已做逐位元對拍；主選單／玩家操作、音效、受控亂數及完整 remake 未完成。精確命令、完整輸入、原始定位、雜湊與限制見研究入口／工具規格。
- 274→275 回填正負護欄、全部既有回填函式、Python 語法、繁體新增文字、差異與擁有權通過。工具 `3c4bace68829bd16377aeaf30e22631c6ce3cefb`（7 檔）推送 `github/codex/moo2-parity-20260930`，未推本機 origin；主庫只更新四份現況／歷程文件，研究連結鎖定工具提交，版權輸入與完整終端／記憶體／gzip／PNG 未加入 Git。
- 初次 image inspect 讀取不存在欄位、舊規格檔名猜錯皆為控制面／讀取問題；改讀 Config／實際 258 檔名後繼續，未建重複映像或修改工具 runtime。
- 本批一次性 Docker 容器收尾清空；隔離工具無 root-owned／誤建 .md 目錄，本輪檔案及輸出 UID/GID=1000:1000。主庫既存 `go.sum`／`lbxinfo`／`.docker-cache` 擁有權未改。下一步依公開硬體契約核對新埠，不深挖 timer driver、ISR／busy-wait 或硬體時鐘。

## 2026-10-01：PIT 模式 2、共享時鐘及探針容量修正

- 開工主庫 `c283b049c6a7b862d9c53658cbe9df3b4ceaf777`、隔離工具 `3c4bace68829bd16377aeaf30e22631c6ce3cefb`，均乾淨。路由命中 dosgolem、規格閘門、平台規格優先及文件分工；沿用逆向重製技能與既有 Docker 映像，未改 Go remake 玩法或建立發行包。
- 276 同次建檔入索引；原版有限 34h／4Eh／17h 與第一讀取、公開 Intel 8254 合法除數足夠後 READY，再接模式 2 與共享週期。模式編號保留，非法 1 拒絕；時鐘微秒、相位與 IRQ 合併保持明示近似，沒有 timer driver／ISR 考古。186 模式 3 的未測非法邊界與 275 舊停點均回填。
- `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/machine -run 'Test(PIT0|BIOSClock)' -count=1`、固定官方 EXE 的 `go test -p 2 -buildvcs=false ./... -count=1` 通過。原版輔助有限擷取、全編碼／共享 CPU／週期／拒絕邊界、全部回填正負例與索引／語法／繁體文字／擁有權驗證通過；精確雜湊見研究紀錄及工具規格。
- 兩次自然探針非零返回，一次完整保存 `signal: killed`；首輪暫存輸出隨容器消失，未保留 cgroup 計數。查明最後 32 筆的 slice 縮容量後持續擴容；100,000 步可重現並驗證固定容量修正，不改原版輸入／CPU 步進。killed 與該缺陷的關聯為強推論；同一 2 GiB 工具鏈乾淨重跑兩條 50M 收據成功，oom／oom_kill=0。控制面曾有路由命令／trap 引號、Go 快取未設與稽核舊檔名錯誤，訂正後通過；不是玩法缺陷。
- 兩條自然路徑越過設定，仍在高位 LE `0x239B03／0x239B09` 的 CMP／JE 等待，畫面同已檢視黑圖。額外 20.12M 只讀觀測確認預設時鐘多三次、DPMI 實模式 08h／1Ch 與 IVT 皆零、等待來源四 bytes 仍零。276 限定設定／共享週期近似 CONFORMED，255 READY；玩家路徑與 remake 玩法同狀態未完成。下一步只核對 DOS `AH=25h／35h` 向量參數及平台保存／派送，禁止猜補等待結果。
- 工具提交 `7b288510f8ff578079e73be0435784a9a2e7f720` 已推送 `github/codex/moo2-parity-20260930`，工作樹乾淨。主庫本輪只更新現況、活表、研究與本歷程四檔；公開 Git 不含正版素材、原始終端／記憶體或 PNG。修改檔與輸出 UID/GID=1000:1000，隔離工具無 root-owned／誤建 .md 目錄；主庫歷史 go.sum／lbxinfo／.docker-cache 擁有權問題保持未動。所有本輪一次性容器已結束，掛載篩選無殘留，未清理其他專案映像或容器。

## 2026-10-01：DOS 向量與受限 IRQ0 轉送

- 開工主庫 `fdec7c3ed3da0849739f137a052358915387706a`、工具 `7b288510f8ff578079e73be0435784a9a2e7f720`，原版輸入固定。命中規格閘門／平台優先／dosgolem、文件職責與結論回填路由，沿用逆向重製技能；玩法 RE 閘門保持關閉。
- 規格 277 建檔入索引，取得原版非空向量、私有框架與等待自然退出後 DRAFT→READY 才新增受限轉送。兩 CPU 模式、完整保存／返回污染／有界拒絕、PIC／EOI／IRQ7 優先及滑鼠共存均驗證；最後的跨 CPU／污染預設入口護欄完成後，平台與固定 EXE 全套／兩個自然排程條件同映像乾淨重跑。命令、輸入、原始位址／bytes 與雜湊在研究入口及規格。
- 原版第 1,160,098 步實際進 IRQ0，在高位 LE `0x244D9A` 的 CS 記憶體比較停止，started=1／completed=0；事件尚未注入，不能算滑鼠分支已驗。這是舊未派送而隱藏的更早 CPU 缺口，277 保持 READY、255 READY。沒有宣稱原版返回、主選單、玩家路徑或 remake 玩法完成；下一步只補標準 CPU 比較，禁止猜補等待值／深挖 ISR。
- 擷取候選載入差值與單步進核心的兩次失敗均保留收據，修正後同工具鏈重跑；上一輪誤落 CMP 分支的護欄已用實際 AST／既有原版收據訂正。回填稽核初次選到同名 fixture 條件，及繁體字檢查把共用字列為候選，修正檢查後通過；這些是驗證控制面問題。首次 Git 暫存因既有忽略父目錄拒絕已追蹤探針來源，確認唯獨強制加入該來源後提交；沒有加入研究輸出。
- 隔離工具提交 `777472b02a561740e11558ecb9852b7579541454`（13 檔）已推送 `github/codex/moo2-parity-20260930`，未推本機 origin。主庫本輪只改 CONTEXT／WORKLIST／WORKLOG／既有研究入口。原版素材／完整終端／gzip／PNG 只留本機；新增與修改檔及輸出 UID/GID=1000:1000，工具無 root-owned 或誤建 .md 目錄。主庫歷史 go.sum／lbxinfo／.docker-cache 不動；本批 Docker 一次性容器收尾清空，沒有刪其他專案資源或映像。

## 2026-10-01：CS 比較、ES 載入與五次原版 IRQ0 返回

- 開工主庫 `f720506116e9f2e9644a0898f4034d66fa2f48fa`、工具 `777472b02a561740e11558ecb9852b7579541454`，均乾淨。沿用規格閘門／平台優先／dosgolem、文件職責與結論回填路由及逆向重製技能；玩法 RE 閘門未開，不改主庫玩法。
- 278／279 各自 DRAFT＋索引→有限原版與公開 CPU 契約審查→READY→實作→限定 CONFORMED。補 CS 記憶體 CMP 的六旗標與符號擴展、絕對 word 的 ES 目的；全部 CPU 及固定官方 EXE 全套通過。原版兩個自然排程已完成五次 IRQ0 返回，第六次停 `0x244E9C` 的 CB，外層完整恢復。等待來源仍零，事件尚未注入，沒有把兩條排程當兩條玩家分支或正常玩家路徑驗收。
- 279 有限擷取首輪外層 240 秒收尾返回 124，完整有限樣本與等待退出已得；保留收據，只改外層 300 秒，同映像／輸入／命令乾淨重跑退出 0。PDF 抓圖 Cache miss、索引路徑猜錯與缺 rg 皆為工具／讀取問題；改 PDF 附錄文字、實際索引與 pathlib 後繼續，不增加映像。原始 MOV 表格誤植另有附錄與原始 bytes 校準；未深入 ISR。
- 256 的 ES 拒絕測試範圍由新規格取代，SS／其他未知形狀仍拒絕；277／278 舊停點與完整返回限制回填，正負例／全部舊護欄／AST 擷取、語法、索引、繁體文字與 UID 稽核通過。255／277 READY，完整等待／玩家路徑及 remake 玩法對拍未完成；下一步是公開標準 CB 返回的最小契約。
- 工具提交 `fabbed1a8bfd1c010b66009bdffe4a6812eb9924`（7 檔）、`0e6a69b0ca1971bc9235fb9727bc26d6e57eb2ed`（10 檔）已推送 github/codex/moo2-parity-20260930，未推本機 origin。主庫只更新 CONTEXT／WORKLIST／WORKLOG／既有研究入口四檔。原版資料、完整終端／記憶體／gzip／PNG 留本機；本輪檔案及輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄；主庫歷史 go.sum／lbxinfo／.docker-cache 不動。一次性 Docker 容器收尾清空，沒有清理其他專案資源或映像。

## 2026-10-02：修正 CB 遠返回並自然驗證下一停點

- 開工主庫 75bb479608478aa7bedf43ca98a956384091d130、隔離工具 0e6a69b0ca1971bc9235fb9727bc26d6e57eb2ed，兩庫乾淨並與遠端一致。沿用推送授權、逆向重製技能與規格閘門／平台優先／dosgolem 路由；結論與文件維護另載入文件職責／後續回填入口。主庫玩法 RE 閘門保持關閉。
- 280 同次建檔入索引；原版有限 CB 框架／完整返回及 Intel RET 公開契約足夠後 DRAFT→READY 才實作。審查訂正為必要 6-byte 讀取、完整消費 8，跳過 selector 高 word；已知平坦 code／同 RPL／VM=0 模型、原子提交與拒絕邊界明示，不擴充近返回、CA、權限／核心鏈。全部 CPU 與固定官方 EXE 全套通過，限定 CONFORMED；精確命令、地址空間／bytes／輸入與收據在研究入口及規格。
- 兩個自然排程都越過高位 LE 0x244E9C 的 CB，內層返回合成平台 0108:00326008、ESP=FF4h，下一步停既有預設核心鏈尚未建模護欄。外層第 1,231,392 步、started=6／completed=5，外層上下文完整恢復；等待來源仍零，事件尚未注入，PNG 同已檢視黑圖。255／277 READY；正常玩家路徑、音效、受控亂數及 remake 玩法對拍尚未完成。
- 有限擷取 300 秒命令乾淨退出 0；後加完整保持／框架護欄以既有原版收據與實際 AST 正負例驗證。279／277 歷史停點回填，全部 23 個回填函式／新四項負例、既有 CMP／PIT 護欄、Python 語法／索引／擁有權通過。初次負例只刪第一處而另一處仍在，訂正為全刪後通過；image inspect、read32／舊檔名猜測均屬控制面／讀取問題，沒有改工具 runtime 或另建映像。
- 工具提交 3e260dcf2215228d840b98242c1226ab7f902456（8 檔）已推送 github/codex/moo2-parity-20260930，未推本機 origin。主庫只更新 CONTEXT／WORKLIST／WORKLOG／既有研究入口四檔，研究連結鎖定工具提交。原版素材／完整終端／記憶體／gzip／PNG 留本機；本輪來源及輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄。主庫歷史 go.sum／lbxinfo／.docker-cache 擁有權未動；本批一次性 Docker 容器已清空，沒有清理其他專案或映像。
- 下一步只按公開 DOS/4GW chaining／結束鏈介面建立新受限平台契約，READY 後才接合法完成／拒絕條件與自然重跑，不深挖 ISR／driver／busy-wait、不追逐週期或注入等待值。


## 2026-10-02：原版 IRQ0 結束鏈與模式 2 等待閉合

- 開工主庫 879f052bc188cbcd75ceea26820b9587c1d11983、工具 3e260dcf2215228d840b98242c1226ab7f902456；沿用既有 push 授權及逆向重製技能，載入規格閘門／平台優先／dosgolem、文件職責與結論回填入口。主庫玩法 RE 閘門保持關閉。
- 281 同次 DRAFT＋索引，有限原版 BDA+1 反例推翻「只返回」候選；公開 BIOS tick／EOI 契約審查後 READY 才實作。平台及固定官方 EXE 全套通過；兩個自然排程自行完成 1,595 次原版 IRQ0 返回，等待值變為 1 並退出模式 2 等待。281／受限 277 CONFORMED，255 READY；受控事件已注入，正常玩家操作／主選單與整款 remake 仍未完成。下一停點是 PIT 通道 0 的 count latch。
- 24 個回填函式、新四項缺證據負例、有限原版邊界護欄及既有 CMP／PIT AST 護欄通過。scope／舊規格路徑猜錯及公開原始碼 Cache miss 是讀取問題，使用實際檔名與官方公開參考後繼續，沒有重建工具映像或深挖 ISR。核心布局／IF 與 EOI 仍明示平台近似。
- 工具 11 檔提交 9b8d1a07121fab929fc94a7f529449beb0d49742 已推送 github/codex/moo2-parity-20260930，未推本機 origin；主庫只更新 CONTEXT／WORKLIST／WORKLOG／既有研究入口四檔，深層連結鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機，所有本輪來源／輸出 UID/GID=1000:1000；工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除，未清理其他專案或映像。


## 2026-10-02：接續計數鎖存與 E4 原版讀取

- 281 主庫 checkpoint 03a9fde13e4dc06fc07e8ad0954b75d114b74b32／工具 9b8d1a07121fab929fc94a7f529449beb0d49742 推送後繼續同一目標。282／283 分別 DRAFT＋索引、公開 Intel PIT／IN 契約與實際自然停點審查、READY 後實作；限定 CONFORMED。平台、全部 CPU 與固定官方 EXE 全套通過，原版兩個自然排程自行消費凍結低高計數，再停 byte 記憶體 SUB，沒有修改主庫玩法或注入遊戲值。
- 初次新測試樣板容量不足，修正容量／共享 I/O 接線後揭露 E4 CPU 缺件；第二次失敗重查知識路由，先以既有 EC 隔離平台，再用自然 E4 命中與公開契約完成 CPU gate。同映像乾淨重跑通過，兩份失敗收據保留；沒有增加硬體／driver 考古。26 個回填函式、新六負例／CLI、兩排程三筆 latch／兩次 IN 狀態及索引／繁體字／UID 通過。
- 工具 13 檔提交 c58709c5ffc8841a22112ad1ea8016890f87d9e5 已推送 github/codex/moo2-parity-20260930，未推本機 origin。git add 對父目錄被忽略但已受版控的探針產生提醒，精確暫存 13 檔仍完整，核對後提交；沒有強制加入其他 workplace 原始輸出。主庫只更新四份現況／歷程／既有研究入口，深層連結鎖定工具提交。
- 原版素材、完整記憶體／終端／gzip／PNG 留本機；本輪來源與輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。一次性容器已退出移除，未清理其他專案或映像。主選單與正常玩家路徑、音效／受控亂數及整款 remake 尚未完成；下一步是公開標準 byte SUB CPU 契約。
