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


## 2026-10-02：byte 記憶體 SUB 與原版自然後態

- 開工主庫 673acf6c6e414b29c301ef2d7bfc473702a3d207、工具 c58709c5ffc8841a22112ad1ea8016890f87d9e5；沿用規格閘門／平台優先／dosgolem、文件職責及結論回填路由與逆向重製技能。284 同次 DRAFT＋索引，唯讀有限原始 byte 與 Intel SUB 契約審查 READY 後才實作。主庫玩法 RE 閘門保持關閉。
- 完整 CPU 初次通過，審查補強前綴負例為所有段均可寫的初態，再以同命令乾淨重跑通過；固定官方 EXE 全套通過。兩個原版自然排程自行完成 16h→0Eh、0Dh→05h，完整 R／段保持，六旗標符合契約；284 限定 CONFORMED。下一停點為高位 LE 0x25425F 的 byte 記憶體 ADD；主選單／正常玩家路徑與整款 remake 尚未完成。
- 工具八檔提交 c1c8e731c4bcb0a5fda529f43e066508db74ba49 已推送 github/codex/moo2-parity-20260930，未推本機 origin。283 舊停點／索引與回填護欄同步維護，27 個回填函式／新四負例與 CLI／兩排程原版後態稽核通過。git add 對已受版控探針的 ignored 父目錄提示，核對八檔暫存完整後提交，未加入原始輸出。
- 主庫只更新四份現況／歷程／既有研究入口；精確命令與輸入／收據雜湊連到鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。一次性容器已退出移除，未清理其他專案或映像。


## 2026-10-02：接續 byte 記憶體 ADD

- SUB 主庫 895375bb17c6757248055a9ba601030e431709c8／工具 c1c8e731c4bcb0a5fda529f43e066508db74ba49 推送並回讀後繼續。285 同次 DRAFT＋索引，公開 ADD 契約與唯讀 byte=03h 審查 READY 後才實作；CPU 全套與固定官方 EXE 全套通過。兩個自然排程自行完成 byte=03h→1Bh／flags=297h→206h、完整 R／段保持，限定 CONFORMED。第 20,637,105 步轉停 byte 暫存器 XOR；主庫玩法 RE 閘門保持關閉。
- 首次 ADD 繞回測試誤留 SUB 的 ModRM，改正後同命令乾淨重跑通過，失敗收據保留。索引舊摘要與初建時沿用的 byte／imm 值按實際收據修正；首次繁體字稽核找到單字，修正並加 set -e 後重跑通過。28 個回填函式、新四負例／CLI、兩排程原版後態及索引／語法／UID 通過。主庫前一輪稽核錯要求 WORKLIST 複製工具提交，依文件職責修正後通過，文件本身沒有缺件。
- 工具八檔提交 96aba2440ce181a1808a508ee897985a2fbcdc99 已推送 github/codex/moo2-parity-20260930，未推本機 origin；主庫只更新四份現況／歷程／既有研究入口。精確命令與輸入／收據雜湊連到鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除。整款 remake 尚未完成，下一步為公開 XOR 定義／未定義旗標的窄 CPU 規格。


## 2026-10-02：byte 暫存器 XOR 與 dword 暫存器 TEST

- 開工主庫 d51d7a67771d328d54924d9b28e5b0e10a0ae67d／工具 96aba2440ce181a1808a508ee897985a2fbcdc99 乾淨；上一輪 SUB／ADD 已推送，屬於實際進展。沿用規格閘門／dosgolem、文件職責與後續回填路由及逆向重製技能。286／287 各自同次 DRAFT＋索引、唯讀原版完整初態與公開 CPU 契約審查 READY 後實作；主庫玩法 RE 閘門保持關閉。
- 全部 CPU 與固定官方 EXE 全套通過。兩原版自然排程自行完成三筆 CL XOR、dword TEST 與第一 JNZ 不跳續行，目的外完整狀態與定義旗標已驗。AF 清除只屬工具近似，未做原版硬體未定義值全旗標聲明。286／287 限定 CONFORMED；下一停點為高位 LE 0x254510 的 dword ROL／imm8。IRQ0 自行完成 1,598 次、等待值4，主選單／正常玩家路徑與整款 remake 未完成。
- 兩項舊停點／索引與護欄同步回填，全部 30 個回填函式、各四項缺證據負例／CLI、兩排程原版後態及索引／語法／繁體字／UID 通過。讀取輸出截短後以實際文件段落與有限 gzip 行補讀；沒有退回主機分析或新建映像。
- 工具八檔提交 30faf6eb5643aac18d197c64804f81c18293876b、八檔提交 4d7ae6c2176feda4c61e47ee2db41695a25c2208 已推送 github/codex/moo2-parity-20260930，未推本機 origin；git add 的 ignored 父目錄提醒後核對暫存完整，沒有強制加入原始輸出。主庫只更新四份現況／歷程／既有研究入口，證據連結鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除，未清理其他專案或映像。

首次主庫繁體字稽核找到新增段落的兩個字形，源於串接文字只轉換末段；修正完整新增段落後，以同一命令乾淨重跑通過，未改 CPU／遊戲收據。該輪稽核失敗後提交步驟仍執行，故字形修正另以追加提交保留，未改寫已推送歷史。


## 2026-10-02：接續 dword 暫存器 ROL

- 開工主庫 ab4ea4bdab5611a31e4b0f9c32c4556b39695fc5／工具 4d7ae6c2176feda4c61e47ee2db41695a25c2208 乾淨；沿用規格閘門／dosgolem、文件職責與後續回填路由及逆向重製技能。288 同次 DRAFT＋索引、唯讀完整 R／段與公開 Intel 契約審查 READY 後實作；主庫玩法 RE 閘門保持關閉。
- 全部 CPU 通過，審查補強保持旗標全清／全設後，同容器／命令乾淨重跑通過；固定官方 EXE 全套通過。兩自然排程自行完成三筆 ROL／CF 與下一 word 比較消費，完整目的外 R／段保持。多位 OF 保留明標工具近似，限定 CONFORMED；啟動推進到第 26,396,706 步，下一停點 byte CL NEG。IRQ0 自行完成2810次、等待值1216，主選單／正常玩家路徑與整款 remake 未完成。
- 31 個回填函式、新四項缺證據負例／CLI、兩排程三筆完整後態及索引／語法／繁體字／UID 通過。首輪讀取猜錯索引與稽核入口，改實際路徑後繼續；兩次文件尾空白檢查均停止提交，第二次重查文件職責／回填路由，修正並以暫存差異檢查重跑通過。沒有退回主機分析或重建映像。
- 工具八檔提交 c605a7f13e00c061d09f14bbf7583c201335f980 已推送 github/codex/moo2-parity-20260930，回讀遠端提交一致，未推本機 origin。主庫只更新四份現況／歷程／既有研究入口；證據連結鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除，未清理其他專案或映像。


## 2026-10-02：接續 byte 暫存器 NEG

- 288 工具 c605a7f13e00c061d09f14bbf7583c201335f980／主庫 5bc5b26e2846232e54b9a8e52af54b01bb13475b 已推送並回讀一致。沿用規格閘門／dosgolem、文件職責與後續回填路由及逆向重製技能，289 同次 DRAFT＋索引、唯讀完整初態／公開六旗標審查 READY 後實作；主庫玩法 RE 閘門保持關閉。
- 自動核准審查逾時，唯讀核對確認 READY 修改未執行，重試後成功。首次 CPU 全套遇舊未知 group /3 DL 與新 READY 支援範圍衝突，改用仍未支援的 /2 byte NOT 保留拒絕護欄；同映像／命令乾淨重跑全部 CPU 與固定官方 EXE 全套通過，失敗收據保留。八目的／256來源／64算術旗標初態、完整保持／拒絕／既有路徑回歸已驗。
- 兩自然排程自行完成三筆 byte NEG、完整目的外 R／段／六旗標與 MOV 消費，289 限定 CONFORMED。啟動推進到第 38,427,368 步，下一停點 D2 E5 的 SHL CH,CL；IRQ0 自行完成5346次、等待值3752，主選單／正常玩家路徑與整款 remake 未完成。32 個回填函式、新四項缺證據負例／CLI、兩排程原版完整後態及索引／語法／繁體字／UID 通過。
- 工具九檔提交 78738dc77eeaee6a1578f7ee98027d679408226f 已推送 github/codex/moo2-parity-20260930，回讀遠端提交一致，未推本機 origin。主庫只更新四份現況／歷程／既有研究入口，證據連結鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除，未清理其他專案或映像。


## 2026-10-02：接續 CL byte 左移與記憶體 OR

- 開工主庫 0bdc75854b86eb8311eabea5002e314cbba007ee／工具 78738dc77eeaee6a1578f7ee98027d679408226f 乾淨，上輪 ROL／NEG 已推送。沿用規格閘門／dosgolem、文件職責與後續回填路由及逆向重製技能，主庫玩法 RE 閘門保持關閉。
- 290 同次 DRAFT＋索引、原版完整 CH／CL／下一 OR 初態與公開 CPU／未定義近似審查 READY 後實作；全部 CPU／固定 EXE 全套通過，但下一記憶體 OR 缺件，290 保持 READY。291 再以兩份未改 OR 的完整初態審查 READY 後實作，兩自然排程三組 SHL→OR→ADD 完整資料鏈通過，才一併限定 CONFORMED。兩項所有回歸首次通過；沒有修改主庫玩法或猜用途。
- 八 byte 目的／全部來源與 CL／別名、八 OR 來源／全部 byte 配對、全 ModRM／SIB／DS／SS／地址繞回及拒絕／保持已驗。34 個回填函式、各四項缺證據負例／兩 CLI、兩排程完整資料鏈與索引／語法／繁體字／UID 通過。啟動推進到第 39,983,174 步，下一 F3 AF 的 REPE SCASD 缺件；IRQ0 自行完成5675次、等待值4081，主選單／正常玩家路徑與整款 remake 未完成。
- 工具十檔提交 0c6d53b869cb52fd716d95c868326244614abbf1 已推送 github/codex/moo2-parity-20260930，回讀遠端提交一致，未推本機 origin。主庫只更新四份現況／歷程／既有研究入口，證據連結鎖定工具提交。原版素材與完整終端／記憶體／gzip／PNG 留本機；本輪來源／輸出 UID/GID=1000:1000，工具無 root-owned／誤建 .md 目錄，主庫歷史 go.sum／lbxinfo／.docker-cache 不動。本批一次性容器已退出移除，未清理其他專案或映像。

## 2026-10-02：接續 REPE SCASD 與原版真實讀取

- 主庫 fbedb135060c021b9dac8ff70007f669c784c68c／工具0c6d53b869cb52fd716d95c868326244614abbf1 乾淨且遠端一致。沿用平台規格優先／規格閘門、dosgolem、文件職責與回填路由及逆向重製技能。292同次 DRAFT＋索引，未改CPU的完整R／段／8192 bytes掃描資料與公開Intel契約審查 READY 後才實作；主庫玩法 RE 閘門保持關閉。
- Intel80386 REP偽碼與文字退出條件矛盾，以正式SDM確認。READY文件寫入的自動權限審查逾時且未執行，改用工作區檔案編輯完成；主庫首次提交遇sandbox的Git中繼資料唯讀，依既有授權用主機Git權限完成。分析／測試／原版執行仍全在Docker。
- 全部CPU／固定官方EXE全套首次通過，兩自然排程比較59次，完整ECX=7C5h／EDI=6BBD4Ch／flags=206h，再由SUB EDI,4／MOV EAX,[EDI]真實讀取FFFFFBFFh，其他完整R／段保持。六旗標全定義；單次Step／內部IRQ與Error／restart等工具模型界限仍明示。292限定CONFORMED，第39,983,178步轉停高位LE0x25489C的83 F0 FF。IRQ0 completed=5675、等待值4081，主選單／正常玩家路徑與整款remake未完成。
- 35個回填函式、新四項缺證據負例／CLI、兩排程完整SCASD→SUB→MOV資料鏈與索引／語法／繁體字／UID通過。工具八檔提交e914184166c2397dba5041617793a58e42f2f927已推送github隔離分支，遠端回讀一致，未推本機origin。主庫四份現況／歷程／既有研究入口與鎖定證據同次更新。原版素材及完整終端／記憶體／gzip／PNG留本機，來源／輸出UID:GID=1000:1000；工具無root-owned／誤建.md目錄，主庫歷史root-owned不動。本批容器已退出移除，未清理其他專案／映像。

## 2026-10-02：接續 dword XOR 與 BSF 資料鏈

- 主庫aa598d09ff9d0394d4e0f2b0bf5485ebe3e30cf3／工具e914184166c2397dba5041617793a58e42f2f927乾淨，上輪SCASD已推送並完成原版消費驗證，屬實際進展。沿用平台規格優先／規格閘門、dosgolem、文件職責與回填路由及逆向重製技能，主庫玩法RE閘門保持關閉。
- 293同次DRAFT＋索引、完整唯讀初態及公開符號延伸／定義五旗標／AF近似審查READY後實作。首次CPU失敗是新附帶ADD回歸AF預期算錯，CPU未改；修正低四位無進位的207h後同映像／命令全部重跑通過，失敗收據保留。固定EXE與兩自然XOR後態通過，但BSF缺件，293保持READY。
- 294同次DRAFT＋索引，以未改BSF的兩自然完整初態與正式Intel SDM審查READY後實作。全64來源目的／別名、全部低16位、置1位置／旗標與拒絕／保持通過；全部CPU、固定EXE全套及兩自然XOR→BSF→ADD→word MOV完整資料鏈通過，293／294才一併限定CONFORMED。原版把索引Ah消費成74Ah並存回DS:002726D0，第41,223,220步轉停0x23C36B的記憶體dword OR，IRQ0 completed=5936／等待值4342。主選單／正常玩家路徑及整款remake未完成。
- 37個回填函式、各四項缺證據負例／兩CLI、兩自然完整資料鏈與索引／語法／繁體字／UID通過；原始ZIP／patch／EXE雜湊重查一致。工具十檔提交a899e9e7e4f397d17054bdccae97faa0f6cf963c已推送github隔離分支，遠端回讀一致，未推本機origin。主庫只更新四份現況／歷程／既有研究入口，原版素材與完整終端／記憶體／gzip／PNG留本機。本輪來源／輸出UID:GID=1000:1000，工具無root-owned／誤建.md目錄，主庫歷史root-owned不動；本批一次性容器已退出移除，未清理其他專案／映像。

## 2026-10-02：接續記憶體 dword OR 與完整讀取

- 開工主庫0b0faf6df8f46823eab81f620aa380e670e56923／工具a899e9e7e4f397d17054bdccae97faa0f6cf963c乾淨且遠端一致；上一輪XOR／BSF已推送。沿用平台規格優先／規格閘門、dosgolem、文件職責與回填路由及逆向重製技能，主庫玩法RE閘門保持關閉。
- 295同次DRAFT＋索引，未改CPU的完整R／段／目的dword與公開OR、五定義旗標／AF及逐byte錯誤模型審查READY後，只改通用記憶體解碼。全部CPU及固定EXE全套通過，兩自然40h→2040h寫回與MOV EAX真正讀取完整2040h後才限定CONFORMED；沒有遊戲位址特例，沒有修改主庫玩法。
- 首次CPU失敗是測試Bus拒絕全部byte，修正指定byte的失敗測試，CPU不變乾淨重跑通過。首次原版診斷包裝Bus破壞DPMI的身分驗證，提早INT31/0500h收據不計OR驗收；改觀察SegmentRead8、原樣轉送既有返回值並保持Bus後重跑。未變高byte的讀取不當成新增bit13消費，再收窄診斷取得真正完整dword MOV。CPU與DPMI不再改；完整失敗／有效收據與精確命令集中在295。
- 38個回填函式、兩份舊規格各四項缺證據負例／CLI、兩自然完整OR／JMP／MOV／TEST與新停點的獨立算術核對通過；原始ZIP／patch／固定EXE雜湊、索引／語法／繁體字／UID及git diff --check通過。第42,347,254步停實模式OUT 022Ch／C6h，IRQ0 completed=6173／等待值4579，主選單／正常玩家路徑、音效／受控亂數與整款remake未完成。
- 工具九檔提交418ca3cf6d874e24127da24c0ba66e9fafecf6e7已推送github/codex/moo2-parity-20260930，回讀遠端一致，未推本機origin。主庫更新唯一現況表與既有歷程／研究入口，鎖定工具提交；原版素材與完整終端／記憶體／gzip／PNG留本機。下一步公開SB16 DSP／DMA與sample duration契約加有限原版參數，審查READY再實作；不深挖driver／ISR／硬體時序。Docker按Go映像與moo2名稱分開核對，本批一次性容器已退出移除，未清理其他專案或映像。

## 2026-10-02：接續 SB16 C6h 自動初始化 DMA

- 開工主庫eac4a998ba1a4dc4d0a515eb9ab510c3670652da／工具418ca3cf6d874e24127da24c0ba66e9fafecf6e7乾淨，上一輪OR／完整MOV消費已推送。沿用平台規格優先／規格閘門、dosgolem、文件職責與後續回填路由及逆向重製技能，主庫玩法RE閘門不變。
- 296同次DRAFT＋索引，有限非自然參數探針取得C6 20 FF 07／22050Hz／4096byte DMA ring，公開Creative命令、每channel／TimeConstant與sample duration契約審查READY後實作。只接DSP／8位DMA，不改CPU或主庫玩法，不追driver／ISR／硬體逐週期。
- 全部CPU／機器層與固定官方EXE全套首次通過。條件時鐘獨立取樣數、46440µs首DSP block／92880µs ring重載、8／16位IRQ隔離、確認／EOI／IRET、mask／reset及拒絕／來源邊界已驗。兩自然原版接受完整C6、實模式333步返回，MOV EBX成功值0／CMP／JZ進0x2454E7成功分支；完整外層R／段／封包差異通過，296才限定CONFORMED。
- 首份正式來源診斷誤取呼叫前未設定的DMA，改取返回後base／page，同映像／命令重生兩自然，執行結果不變，舊收據保留而不當C6來源證據。返回21µs／信用926100尚未到第一sample，PCMBytes=0；條件pattern不能替代原版保護模式連續PCM／IRQ7、人耳或正常玩家路徑驗收。
- 全部39個回填函式、三份舊停點共15項缺證據負例／CLI、兩自然獨立時間／封包／caller／新停點稽核通過，原始ZIP／patch／EXE雜湊再核對。第42,347,639步轉停高位LE0x257662的dword ROR立即數10h；IRQ0 completed=6173／等待值4579、PNG同已檢視黑圖。255／主選單／正常玩家路徑、保護模式音訊／IRQ7、受控亂數與整款remake未完成。
- 工具12檔提交56e1ff979b517740bf056668ee90800bfac27321推送github/codex/moo2-parity-20260930，未推本機origin；主庫只更新四份現況／歷程／既有研究入口，證據鎖定工具提交。原版素材及完整終端／記憶體／gzip／PNG留本機，來源／輸出UID:GID1000:1000，工具無root-owned／誤建.md目錄，主庫歷史root-owned不動。本批一次性容器已退出移除；Go映像的0da5d78f687c掛載皆屬fd2，保留其他專案工作與映像。

主庫四檔現況／鎖定證據／繁體字／UID／唯一AGENTS與git diff --check通過，歷史root-owned仍2709筆。工具遠端回讀56e1ff979b517740bf056668ee90800bfac27321與本機一致。

## 2026-10-02：接續 dword ROR 全部立即數與真實消費

- 開工主庫813ced925045f3bb756983378af2a0551da18e0d／工具56e1ff979b517740bf056668ee90800bfac27321乾淨，上一輪C6已推送並驗原版返回／成功caller，屬實際進展。沿用平台規格優先／規格閘門、dosgolem、文件職責與結論回填路由及逆向重製技能，主庫玩法RE閘門不變。
- 297同次DRAFT＋索引，以未改CPU的兩自然完整R／段與公開Intel計數／CF／OF契約審查READY後，擴充裸C1 /1的全部imm8，零計數保持／單位OF更新／多位OF保留近似分開驗。舊07h負例改由全部計數正例驗，其他拒絕護欄保持；全部CPU／固定官方EXE全套首次通過。
- 兩自然各兩組ROR令CF=0、其他完整R／段保持，MOV AX,DX真實消費後完整EAX=00340AFFh／0A0A0AFFh，第二ROR也完成。獨立單bit整除oracle及完整自然收據核對後297限定CONFORMED；沒有原版位址特例或未定義硬體OF聲明。222／288舊計數範圍與293／294／295／296停點同步回填，40個回填函式／24項缺證據負例／CLI通過。
- 前態命令末端SHA摘要誤列不存在的全套檔而exit1，兩原版執行／gzip／PNG完整，獨立核對後繼續，沒有當產品或CPU失敗。首次猜錯288／fixture檔名後依實際路徑補讀；所有分析、測試與原版執行皆在Docker，未另建映像。
- 第42,349,111步外層高位LE0x2571C9的IRQ0呼叫內，原版停高位LE0x2520B7的D1 E0、SHL EAX,1缺件；IRQ0 started6174／completed6173／failed=true，等待值仍4579，PNG同已檢視黑圖。保留277舊返回樣本邊界，只按公開CPU契約補缺件，不深入ISR／driver硬體時序。255／主選單／正常玩家路徑、保護模式PCM／IRQ7、音訊／人耳／受控亂數及整款remake未驗收。
- 工具14檔提交4cdf20e347500a5c996b955831e35ef68ca556e0已推送github/codex/moo2-parity-20260930，不推本機origin。主庫只更新四份現況／歷程／既有研究入口，深層證據鎖定工具提交；原版素材及完整終端／記憶體／gzip／PNG留本機。來源／輸出UID:GID1000:1000，工具無root-owned／誤建.md目錄，主庫歷史root-owned不動；Go映像與moo2名稱皆無容器殘留，本批一次性容器已退出移除，未清理其他專案或映像。

主庫四檔現況／鎖定證據／繁體字／UID／唯一AGENTS及git diff --check通過，歷史root-owned仍2709筆。工具遠端回讀4cdf20e347500a5c996b955831e35ef68ca556e0與本機一致。

## 2026-10-02：IRQ0 單位 SHL 寫回與 C1 旗標修正

- 開工主庫9835225d47eda25f726c87e624051362669d9f62／工具4cdf20e347500a5c996b955831e35ef68ca556e0乾淨，上一輪ROR已推送。路由命中平台契約／規格閘門、dosgolem、文件職責與結論回填，沿用已載入入口及逆向重製技能；主庫玩法RE閘門不變。
- 298同次DRAFT＋索引，未改CPU的兩自然真正IRQ0前態及公開Intel五定義旗標／AF模型審查READY後才補裸D1 /4。第一次完整CPU回歸找到既有C1單位OF漏設，D1新測試均通過；保留失敗收據，沒有改正確預期。299獨立公開CPU反例／規格READY後，窄修正既有C1單位OF；零計數與未定義AF／多位OF模型保持。
- 全部CPU同映像／命令乾淨重跑通過，固定EXE全套通過。兩自然IRQ0內EAX0／flags46h、EBX1→2／flags2；TEST來源3 AND 8=0，JZ跳過第二組SHL，原版A3／89 1D把結果存DS:00272D40／DS:00272D44，八bytes由全0到00000000 02000000。完整R／段保持及真實寫回獨立核對後298限定CONFORMED；299只限公開CPU契約，未有MOO2自然OF=1同狀態收據。
- 同第42,349,111步外層0x2571C9的IRQ0內，下一停點為高位LE0x25179F的13 ED、ADC EBP,EBP。IRQ0 started6174／completed6173／failed=true，完整IRQ0返回未知。等待4579／PNG同已檢視黑圖；保護模式PCM／IRQ7、人耳、255／主選單／正常玩家路徑／受控亂數與整款remake未驗收。只補標準CPU缺件，不深入ISR／driver硬體時序。
- 293–297舊停點及186／298的C1契約／回歸發現同步回填，42個回填函式／28項新缺證據負例與兩個CLI、兩自然完整鏈核對通過。收據稽核腳本正規表示式跳脫修正後，以同一收據重跑通過，未當產品失敗。原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致；命令／來源／測試／有效與失敗收據雜湊集中工具298／299。
- 工具16檔提交a44c7eaae7f4ac3143a04f183f62ecf91190e124已推送github/codex/moo2-parity-20260930並讀回一致，不推本機origin。主庫只更新四份現況／歷程／既有研究入口，原版素材及完整終端／記憶體／gzip／PNG留本機。工具來源／輸出UID:GID1000:1000、無root-owned／誤建.md目錄，主庫歷史root-owned不動。Go映像與moo2名稱皆無容器殘留，本批一次性容器已退出移除，未清理其他專案／映像。

## 2026-10-02：IRQ0 ADC 與原版索引消費

- 開工主庫df423d7dd09a9ba202c8051bd40045ed513a4ca0／工具a44c7eaae7f4ac3143a04f183f62ecf91190e124乾淨，上一輪兩CPU修正與真實寫回已推送，屬實際進展。路由命中平台契約／規格閘門、dosgolem、文件職責與結論回填，已載入入口／上游能力矩陣及分層規格，沿用逆向重製技能；主庫玩法RE閘門不變。
- 300同次DRAFT＋索引，未改CPU的兩自然真正IRQ0完整前態及公開Intel ADC六定義旗標／別名與保持契約審查READY後才補裸13 /r、mod11。原兩操作數與CF一起計算，64來源／目的、byte配對／兩CF、77項bit／符號／繞回／nibble邊界與64初旗標獨立驗證，保持／拒絕及既有ADD／word ADD／SUB／CMP／SBB回歸通過。全部CPU與固定官方EXE全套首次通過。
- 兩自然各三組完整ADC→索引ADD→MOVSX相同，原CF1／0／1令EBP1／0／1，ADC flags2／46h／2；原版ADD以×4索引讀完整dword2／0／2，ESI=0071E1D2h／0071E1D2h／0071E1D4h、flags6，後續MOVSX才覆寫EBP0。目的外R／段及來源八bytes保持，獨立核對真正來源／索引與完整結果後300限定CONFORMED。299原版自然OF=1收據仍未知，不因ADC驗證解除。
- 同第42,349,111步外層0x2571C9的IRQ0內，下一停點為高位LE0x24678C的66 83 F7 01、XOR DI,1。IRQ0 started6174／completed6173／failed=true，等待4579，PNG同已檢視黑圖，完整IRQ0返回未知。255／主選單／正常玩家路徑、保護模式PCM／IRQ7、人耳／受控亂數及整款remake未驗收；不深入ISR／driver硬體時序。
- 293–299舊停點與索引同步回填，43個回填函式／七份28項缺證據負例、CLI及兩自然三組資料鏈獨立稽核通過。初讀006檔名不符後依實際檔名補讀，未當產品失敗。原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致；命令／來源／CPU／probe／測試與前態及有效收據雜湊集中工具300。
- 工具14檔提交752abc607e04b87d830d4a6451af9e51dfc26b78已推送github/codex/moo2-parity-20260930並讀回一致，不推本機origin。主庫只更新四份現況／歷程／既有研究入口，原版素材及完整終端／記憶體／gzip／PNG留本機。工具來源／輸出UID:GID1000:1000、無root-owned／誤建.md目錄，主庫歷史root-owned不動；Go映像與moo2名稱皆無殘留容器，本批一次性容器已退出移除，未清理其他專案／映像。


## 2026-10-02：word XOR 原版堆疊消費與 IRQ0 返回

- 開工主庫702b6b39f18312b28603d7a7334feb1143ada2e7／工具752abc607e04b87d830d4a6451af9e51dfc26b78，上一輪ADC索引消費已推送。路由命中平台契約／規格閘門、dosgolem、文件職責與結論回填，沿用已載入入口及逆向重製技能；主庫玩法RE閘門不變。
- 301同次DRAFT＋索引，未改CPU的兩自然真正完整IRQ0初態與公開word XOR／五定義旗標／AF模型、PUSH契約審查READY後才補66 83 /6、mod11。全部低word／imm8配對、八目的／45邊界／64初旗標，獨立逐bit比較／整除符號延伸／PF位元計數核對高16位、完整保持與未知拒絕。四固定位址各最多三筆唯讀hook原樣轉送，不替換CPU／Bus／IRQ橋接或時計、不注入資料或跳指令。
- 首次301新指令及隔離PUSH測試通過，舊293負例仍拒絕已合法word形式、新byte XOR回歸誤期望保留AF而失敗。依既有286清AF模型修正新增期望，267／293的word XOR舊負例由全部word正例接替；CPU未再次修改。失敗收據保留，同映像／命令乾淨重跑全部CPU通過，固定官方EXE全套通過。
- 兩自然各一組完整XOR→PUSH EDI→PUSH EAX一致，XOR令EDI=1／flags2；PUSH EDI寫SS:002723FC完整dword由003D6978h變1，PUSH EAX寫SS:002723F8由00246752h變00325048h。ESP每次減4、其他完整R／段與flags保持；真正原版寫入及初態匹配獨立核對後301限定CONFORMED。
- 這次IRQ0已返回，終態active=false／failed=false／started6228／completed6228，等待4634。兩自然前進254181步，在第42603292步停根CPU高位LE0x256171的66 93、XCHG AX,BX。VBE indexed已變化，PNG仍同已檢視黑圖；工具時鐘不稱硬體wall-clock一致。C6返回／來源收據保持；255／主選單／正常玩家路徑、保護模式PCM／IRQ7、人耳／受控亂數、299原版自然OF=1與整款remake限制保持。
- 293–300舊停點及索引同步回填，44個回填函式／32項新缺證據負例／CLI及兩自然完整XOR／PUSH寫入／IRQ0返回／新停點獨立核對通過。工具17檔提交d013029fcddf4da8e8d7650660897223e6fa67ca已推送github/codex/moo2-parity-20260930並回讀一致，不推本機origin。完整來源、有效／失敗收據雜湊與精確命令集中鎖定301；原始ZIP／patch／417根檔／EXE／MOX.SET再核對一致，原版素材及完整終端／記憶體／gzip／PNG留本機。
- 主庫只更新四份現況／歷程／既有研究入口，不修改Go／Ebitengine玩法。來源／輸出UID:GID1000:1000，工具無root-owned／誤建.md目錄，主庫歷史root-owned不動。本批一次性容器均已退出移除；Go映像清查曾見短暫容器，讀取掛載前已自行移除，未確認歸屬，未停止或刪除它。下一步只保存XCHG真正完整前態與下一ROR消費，按公開CPU契約審查窄word暫存器形式，不追helper／ISR／driver硬體內部。

主庫四檔現況／鎖定證據／繁體字／UID／唯一AGENTS及git diff --check通過，歷史root-owned仍2709筆。工具遠端回讀d013029fcddf4da8e8d7650660897223e6fa67ca與本機一致。最後docker ps -a依Go映像與moo2名稱核對皆空。


## 2026-10-02：word XCHG 原版消費與星空畫面

- 開工主庫b0b5a4ab86594466673df7f2f6f192cead05f3b8／工具d013029fcddf4da8e8d7650660897223e6fa67ca乾淨，上一輪word XOR／PUSH及IRQ0返回已推送。路由命中平台契約／規格閘門、dosgolem、文件職責與結論回填，沿用逆向重製技能；主庫玩法RE閘門不變。
- 302同次DRAFT＋索引，以未改CPU兩自然真正完整初態及公開Intel交換／兩高16位／全部旗標保持契約審查READY後，只補66 91–97。七目的／全部AX低word與對側同值、補集／45邊界配對／64初算術旗標與32旗標bit、補集，以整除拆原高低word獨立核對完整保持及拒絕／既有byte、dword XCHG與NOP。三固定位址各最多三筆唯讀hook原樣轉送，不替換CPU／Bus／IRQ橋接或時計、不注入資料或跳指令；全部CPU及固定EXE全套首次通過。
- 兩自然各三組完整XCHG→ROR→RET邊界相同，第一真正初態與未改CPU相同。XCHG令完整EAX0A0A0A0Ah／EBX2E0A0A2Eh、flags206h，兩高16位與其他完整R／段保持；下一ROR EBX,8真實消費完整來源，結果2E2E0A0Ah。後兩組完整來源100A0A0Ah／AX0A10h亦驗交換與完整ROR結果10100A0Ah；原值交換及逐次整除循環核對後302限定CONFORMED，多位OF未定義保留只屬297工具模型，RET只記邊界。
- 兩自然持續至50M診斷上限、根CPU高位LE0x22FCD2，沒有step_error／guest_cpu_stop，完整R／段／flags246h相同。IRQ0 started7789／completed7789／failed=false，等待6195；VBE兩PNG相同，已檢視星空／星雲片段，尚未見主選單。C6命令／handled返回／caller及來源SHA保持，保護模式PCM／IRQ7、人耳、255／正常玩家路徑／受控亂數、299原版自然OF=1與整款remake限制保持。工具時計不當硬體wall-clock證據，上限不當CPU拒絕。
- 首次收據稽核誤要求無事件／受控事件unique_sites相同，實際17667／17697；依兩輸入實際分支覆蓋值修正後以同一收據重跑通過，沒有修改CPU／重跑原版或當產品失敗。293–301九份舊停點及索引回填，45個回填函式／36項新缺證據負例／CLI與兩自然完整消費、IRQ持續返回與上限獨立核對通過。
- 工具16檔提交9e6ee8cea7e40fdf13528fdae6b7959361708aaa已推送github隔離分支並回讀一致，不推本機origin。原始ZIP／patch／417根檔／EXE／MOX.SET再核對一致，精確來源／CPU／probe／測試與有效收據雜湊及命令集中鎖定302；原版素材／完整終端／記憶體／gzip／PNG留本機。主庫只更新四份現況／歷程／既有研究入口，不修改Go／Ebitengine玩法。
- 來源／輸出UID:GID1000:1000，工具無root-owned／誤建.md目錄，主庫歷史root-owned不動。本批一次性容器均已退出移除，Go映像與moo2名稱核對皆空，未清理其他專案／映像。下一步核對正常鍵鼠入口，有限唯讀觀測尾端caller狀態與平台音訊／時計進度；不猜修等待、盲提高上限或深入helper／driver／ISR硬體內部。

主庫四檔現況／鎖定證據／繁體字／UID／唯一AGENTS及git diff --check通過，歷史root-owned仍2709筆。工具遠端回讀9e6ee8cea7e40fdf13528fdae6b7959361708aaa與本機一致。最後docker ps -a依Go映像與moo2名稱核對皆空。


## 2026-10-02：晚期啟動平台時計的自然觀測

- 開工主庫de386e046750f146e86fd88f8da569f22c9b6d77／工具9e6ee8cea7e40fdf13528fdae6b7959361708aaa乾淨。路由命中平台契約、正常玩家路徑與dosgolem，載入入口，沿用逆向重製技能及文件職責／結論回填契約；主庫玩法RE閘門不變。
- 303同次DRAFT／索引，探針只新增四固定sample、三caller各最多兩筆與terminal的平台值快照，原樣轉送hook。CPU／平台三來源雜湊保持，沒有資料／鍵盤／時計／IRQ注入或提高50M上限。兩自然各11快照確認BIOSMicros43985659到52095937，音訊VirtualMicros固定1375，DMA8啟動後剩2048／completions0／PCMBytes0／credit926100固定。靜態接線及自然收據確認保護模式未推進音訊裝置時計；等待因果仍未知。
- 原版AH2509安裝保護模式IRQ1 8:21C4D8，工具正常鍵盤尚未接通；全部快照鍵盤未安裝、讀取／入隊及60／61／64埠讀取零。受控滑鼠回呼返回，沒有鍵盤注入，不用BIOS入隊猜補IRQ1。高位LE0x2A8E54原始四bytes皆零，只確認已觀測比較條件，欄位語意未知。先前實模式DMA16完成1／PCM16Bytes2／IRQ7派送1保持，不稱整段無IRQ7。
- 兩次完整收據與302保持，只排除新增快照／PNG輸出名／解壓mtime與DOS DTA內時間日期四bytes；原檔內容與尺寸、全部既有CPU／橋接／時計列及三組XCHG／ROR真正消費一致。兩次50M無CPU拒絕、IRQ0 started7789／completed7789，PNG仍同星空片段，主選單未見。五來源雜湊／完整收據／11快照／原版AH2509、45個既有回填函式／36項缺證據負例／CLI與索引、DRAFT、繁體字及UID已核對通過。未變CPU全套不重跑。
- 工具五檔提交90b9f4acb3c7e55829973c51a39fe12cb2129df3已推送github/codex/moo2-parity-20260930並回讀一致，不推本機origin。來源／輸入／命令／收據雜湊集中鎖定303。原版素材與完整終端／記憶體／gzip／PNG留本機，只公開自製觀測及有限文字證據。
- 既有moo2-ebiten缺pdftotext及PDF解析模組，分類為工具環境缺件，未啟動主機分析、未另建映像或猜按鍵；不阻塞303觀測。主庫只更新四份現況／歷程／研究入口，不改Go／Ebitengine玩法。303觀測已驗但平台實作仍DRAFT，等待因果、正常玩家路徑／受控亂數、人耳及整款remake未完成。下一步審查共用裝置時間與IRQ7窄契約，READY後實作，不深入driver／ISR或硬體wall-clock考古。

主庫首次繁體字稽核抓到新增活表的時鐘混字，已修正後以同一命令重跑；分類為文件校對，非產品缺陷。

主庫四檔現況／鎖定303證據／繁體字／UID／唯一AGENTS及git diff --check通過，原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致，歷史root-owned仍2709筆不動。工具遠端回讀90b9f4acb3c7e55829973c51a39fe12cb2129df3與本機一致且工作樹乾淨；來源／輸出UID:GID1000:1000。本批一次性容器均已退出移除，最後Go映像與moo2名稱清查皆空，未清理其他專案或映像。


## 2026-10-03：共用 DMA 時計與首個原版 IRQ7

- 開工主庫4e77399672e4e5988e8081ece62e5572c2e27ba2／工具90b9f4acb3c7e55829973c51a39fe12cb2129df3乾淨，前輪303唯讀觀測已推送，屬實際進展。路由命中平台規格優先、dosgolem、規格閘門／文件職責與結論回填，載入入口，沿用逆向重製技能。初讀猜錯閘門入口檔名後依路由實際檔名補讀；主庫玩法RE閘門不變。
- 304同次DRAFT／索引，以303完整原版初態、Creative取樣率／block與Intel／Open Watcom模式框架審查READY後，只補兩CPU模式的裝置時計。共用advanceDMA保留既有1µs近似／實模式順序、來源驗界／速率／block及ring，IF／PIC只阻擋派送；未建模保護模式IRQ7保留pending、三種原始向量與CPU現場明確停止，不假用實模式框架操作保護模式CPU，不丟中斷跑過等待。CPU與啟動來源未改。
- 新增測試以兩真實CPU交錯指令及整除sample數核對四mode／三rate、40h已含channels、遮罩／reset／來源超界、16位單word、IRQ0巢狀計數，以及IRQ7的IF／PIC／服務中閘門與完整CPU／FPU／堆疊保持。首輪自製fixture漏設實模式CS=0造成匯流排越界，核對cpu.Reset後只修測試初態並驗bus.err，平台來源不改，失敗收據保留。同映像／命令重跑機器層通過，固定官方EXE全套首次通過，全部CPU亦通過。
- 兩自然完成首block後，第42356668步、高位LE0x257FC9停在保護模式IRQ7。兩時計由C6返回43985659至44032078，原信用926100加46419×44100獨立整除得2048 samples、餘數4000；DMA完成1、current4800h／count7FFh，block重載2048但ring未重載。最後補有限唯讀PCM雜湊／16byte入口後兩自然同命令重生，完整收據只多兩列、解壓mtime／DTA時間日期四bytes，平台／CPU未改、未重跑未變全套。
- 實際PCM2048個80h、SHA-256 88ed1a04cb43fe65827d1cd9ef6d24a736108730b1ce6315d4d3ca79b6a0d140，與已保存原版buffer一致；靜音buffer不當人耳驗收。absolute IVT0F1201:0682／實模式線性0x12692，DPMI實模式0F及DOS保護模式0F零，入口原始16bytes保存，不追ISR內部。DSP／PIC pending=true、IRQ7Deliveries仍1，未假稱新IRQ7成功；IRQ0 started6176／completed6176／failed=false，等待原始E6 11 00 00。PNG同較早黑圖，首IRQ7在星空前停止；302／303的星空與50M屬較早未推音訊時計基線，不混作本輪終態。
- 十一份較早音訊邊界與索引同步回填，全部46個回填函式／80項缺證據負例／CLI、原版完整時間／PCM／向量與pending稽核通過。工具19檔提交089f51f13149dd2a1b2c08c6b24ec639d0c78788已推送github/codex/moo2-parity-20260930並回讀一致，不推本機origin。來源／測試／失敗與有效收據雜湊及精確命令集中鎖定304；正版輸入及原版素材、完整終端／記憶體／gzip／PNG留本機，只公開自製程式／測試與有限文字證據。
- 主庫只更新四份現況／歷程／既有研究入口。304保持READY，時間到首block已驗，正式IRQ7轉送／返回、正常鍵盤、主選單／正常玩家路徑／受控亂數、人耳及整款remake仍未完成。下一步依公開DOS/4GW及實際IVT補窄派送契約，READY後實作，不改玩法、等待值或提高上限，不深入driver／ISR／busy-wait。本輪測試及兩批自然重播的有界容器已退出移除；另FD2 oracle容器已核對掛載歸屬，保留未操作，沒有清理其他專案或映像。

主庫四檔304現況／鎖定證據／繁體字／UID／唯一AGENTS及git diff --check通過，原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致，歷史root-owned仍2709筆不動。工具遠端回讀089f51f13149dd2a1b2c08c6b24ec639d0c78788與本機一致且工作樹乾淨；來源／輸出UID:GID1000:1000。本輪有界容器均已退出移除，moo2名稱清查空，Go映像中的confident_williams為已確認另一FD2專案，保留未操作，沒有清理其他專案或映像。

最後再次清查：Go映像與moo2名稱皆空，先前另一FD2容器已自行結束，未由本輪停止或刪除。主庫d26f1b8e770e82af1148abcd9e6559e62b18477c已推送並回讀一致。


## 2026-10-03：原版 IRQ7 返回與 XOR AL 正常消費

- 開工主庫3d20cac1b3d1b2fe44922738dd97d3e2c315f42e／工具089f51f13149dd2a1b2c08c6b24ec639d0c78788乾淨，前輪304首block與明確IRQ7停點屬實際進展。路由命中平台規格優先、dosgolem、規格閘門／文件職責與結論回填，載入對應入口，沿用逆向重製技能；主庫玩法RE閘門不變。
- 305先DRAFT／索引，依Open Watcom公開passdown與核心私有16位元堆疊做可丟棄有界診斷。兩固定原版都從實際IVT1201:0682執行73步，原版OUT20h EOI、IN22Eh來源確認及IRET返回成功。診斷在304正式停止收據／PNG之後執行，額外1tick及client0100配置明標，未當正式正常路徑收據。首次編譯因Go預設平行度與64pids耗盡程序，同映像改GOMAXPROCS=2／-p2乾淨通過，分類為工具資源設定。
- READY審查後正式轉送採獨立實模式CPU、主機allocDOS私有4KiB堆疊一次配置重用、真正IVT與共用派送框架；不假呼叫DPMI0300、不增加dispatch tick，不改client DOS block／dosLast／descriptor。caller完整32位R／六段／旗標／EIP／FPU成功或失敗保持，原版資料與I/O真實寫入保留。未支援CPU／埠／INT／HLT／超時／錯誤frame或EOI明確停止並持續拒絕後續執行。寄存器映射與1µs時計明標平台近似，不反組譯ISR。
- 自製ISR／完整硬體frame、caller／FPU／原堆疊保持、9步加外層一次／堆疊重用／client帳本保持、IF／PIC／IRQ0互斥及向量／段／配置驗界、八類失敗與持續停止通過。首輪fixture沿舊測試高位映像而缺低位DOS arena，正確配置失敗；核對setLimits後只修自製fixture入口，production不加特例，保留失敗收據，同命令重跑通過。
- 固定EXE全套及兩正式自然通過首IRQ7：outer_step42356668、高位LE0x257FC9，原版73步EOI／22E／IRET後caller正常下一指令到0x257FCD；兩時計44032151=44032078+73、PCM2051／信用223300。兩自然連續14次轉送返回、PCM29175／信用60600後，第42488059步高位LE0x247BE1原始34 01 C3停在未支援XOR AL,1。工具305提交dcf764ed14d1b141948968fd8ebc596ee3dec851已推送並回讀一致。
- 306先DRAFT／索引，以305完整R／段／flags297h與Intel裸34 ib、五定義旗標及AF未定義契約審查READY後，只接CPU裸AL立即值。256×256×兩初始旗標由獨立逐bit與bit計數驗結果／五旗標，AF清除另驗模型；高24位／完整外層／FPU／記憶體、截短及11種前綴拒絕均通過。固定EXE全部CPU／機器層通過。
- 兩自然三組原版AL為1→0、0→1、1→0，flags297h→246h／202h→202h／297h→246h；RET真正讀SS:ESP的DF 1A 23 00回0x231ADF，ESP加4，caller再ADD ESP,4後，0x231AE2的MOV ESI,EAX真正得到0／1／0。全部R／段與旗標完整核對。305首XOR前收據逐列保持，僅排除解壓mtime／DTA四bytes；有限唯讀探針不注入資料／CPU／鍵盤或提高50M上限。
- 兩自然都到50M無未知CPU／平台拒絕，高位LE0x21588F、兩時計61913470。IRQ7 started386／completed386，最後73步EOI／22E／IRET成功；IRQ0 started8411／completed8411、failed=false。從C6信用與時計獨立整除取樣790617／餘391200，386個2048byte block／4096byte ring吻合，PCM只記錄前65536byte，不混稱總傳輸量。兩PNG相同且實際檢視為星空片段，主選單未見；鍵盤未安裝、60／61／64埠讀取零。
- 十二份音訊邊界與305的XOR停點同次回填，全部48回填函式／132項缺證據負例／CLI、完整原版R／段／RET／MOV鏈與DMA時計稽核通過。繁體字稽核先修新增測試訊息；後來抓到CPU舊來源的混字，稽核改只檢查本輪新增CPU行，不修改無關舊碼。工具306提交4eb97f6277121e528a2d3ce594dbb1967d16382e已推送並回讀一致，不推本機origin。精確輸入／來源／命令／有效及失敗收據雜湊集中鎖定305／306；正版素材、完整終端／記憶體／gzip／PNG留本機，只公開自製來源／測試與有限文字證據。
- 主庫只更新四份現況／歷程／既有研究入口，沒有修改Go/Ebitengine玩法。304–306限定CONFORMED，正常鍵盤IRQ1、255／299自然OF=1、人耳／主選單／正常玩家路徑／受控亂數與整款remake仍未完成。下一步按原版IRQ1 8:21C4D8補有界正常鍵盤平台契約，不以BIOS入隊代替。17:02:52 UTC清查本輪有界容器均退出移除，Go映像與moo2名稱清查皆空；先前另一FD2專案容器competent_boyd已核對掛載，保留未操作，後來自行結束，未清理其他專案或映像。

主庫原HEAD／四檔現況、限定範圍／鎖定306證據／唯一活表、繁體字／UID／唯一AGENTS及git diff --check通過，原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致。歷史root-owned仍2709筆不動；工具HEAD與遠端4eb97f6277121e528a2d3ce594dbb1967d16382e一致且工作樹乾淨。Go映像ID仍sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，最後Go映像及moo2名稱清查皆空，沒有清理其他專案或映像。

主庫3b573fd99eb75cf36b830538eded54a30c1d8b72已推送並回讀一致，兩庫工作樹乾淨。推送後再次清查，本輪容器仍已移除，moo2名稱清查空；新出現的vigilant_elgamal已核對掛載屬另一FD2專案，保留未操作，不把所有Go容器皆空當持續狀態。


### 2026-10-03 正常Esc IRQ1、FF 1D遠呼叫與原版caller續行

- 接手主庫HEAD dfdeca2e8aa6193d7ee7421dbdc2b45aeb5ccb72／工具4eb97f6277121e528a2d3ce594dbb1967d16382e，兩工作樹乾淨。路由命中平台規格優先、dosgolem對拍與結論回填，沿既有remake逆向技能。主庫玩法RE閘門不變。
- 307先DRAFT診斷，定位原版IRQ1正常入口與FF 1D CPU缺件；308公開Intel契約READY後接裸間接遠CALL，完整指標／堆疊／外層與真正CB消費通過。兩自然50M控制基線保持，私有診斷與正常輸入收據分開。
- 公開BIOS／controller／PIC／DOS/4GW契約與97／77步原始入口／返回足以READY後，實作正常Esc 01／81硬體隊列、host私有堆疊、真實default遠CALL服務、20h IRR／ISR與EOI。全部核心／FPU／兩模式／IRQ0優先巢狀、閘門／容量／非法碼、框架／descriptor污染與失敗停止通過。第一次PIC測試接錯讀取位置及BIOS診斷編譯型別錯誤，按契約修正後同容器命令重跑，失敗收據保留，未放寬斷言。
- 最終Go1.24-bookworm映像、network none／2GiB／2CPU／128pids／UID1000／600s，以417原檔與固定EXE執行go test -p 2 -buildvcs=false ./... -count=1通過。兩正常go run -buildvcs=false ./workplace/moo2-probe，DOSGOLEM_MOO2_MAX_STEPS=50000000、DOSGOLEM_MOO2_SEPARATE_DOS=1、DOSGOLEM_MOO2_HARDWARE_ESCAPE_AT_48000000=1，有／無既有受控滑鼠各一次。正常送鍵前完整基線保持，IRQ1各97／77步返回，原caller真實ADD／MOV續行。首輪與最終重跑逐列一致；50回填函式／35新負例／兩CLI通過。
- 新停點第48354467步，外層高位LE0x217AD8、IRQ7實模式1201:05DA第77步，OUT022C=D0未支援。兩時計58057009、IRQ7 started303／completed302、IRQ0完成7927，兩PNG已檢視黑色過場，主選單未見。307／308限定CONFORMED，不稱完整鍵盤或玩家流程。完整來源／命令／收據與近似集中在工具307／308；原版素材、完整RAM／終端／gzip／PNG留本機。
- 工具00ad7c645b19b51a8697e2deae85d8a5019dd657已推送github隔離分支並回讀一致，工作樹乾淨，未推本機origin。主庫只更新既有四文件。下一步公開DSP D0的8位DMA暫停契約與同一固定輸入排程，不深挖driver／ISR／DAC時序、不改玩法。17:58:40 UTC清查本輪有界Docker均退出移除，Go工具映像相關容器清查空；未清理其他專案或映像。

### 2026-10-03 DSP 8 位元 DMA 暫停恢復與正常原版續行

- 接手主庫87547eafa6ce8e198f845bfe419d6231975d8bf5／工具00ad7c645b19b51a8697e2deae85d8a5019dd657，兩工作樹乾淨。路由命中平台規格優先、dosgolem對拍與結論回填，沿既有remake逆向技能；玩法RE閘門不變。
- 309依Creative原廠D0／D4契約先DRAFT、審查READY後接active8傳輸的pause／resume，保留DMA位置、剩餘、分數信用與IRQ；idle／其他命令拒絕。只改隔離工具的DSP／裝置時計閘門與自製唯讀探針，CPU與主庫玩法不改。
- 四個DSP／DMA控制測試與固定官方EXE全部go test -p 2 -buildvcs=false ./... -count=1通過。首輪自製測試未使用import與40h信用預期單位兩次失敗，只修測試後按同一容器契約乾淨重跑，不調時鐘或放寬斷言；失敗收據保留。
- 600s／2GiB／2CPU／128pids／UID1000／network none，以417原檔與固定EXE重生兩組50M上限、48M正常Esc。真正D0／102步原版IRQ7返回、450096µs暫停來源保持、後續D4／93步INT66成功與六條真正caller已驗；原版新停點第48796894步高位LE0x240A32、DOS INT21／AH2Ah。兩時計58553364，IRQ0完成8022、IRQ7完成303；兩圖與已檢視黑色過場逐位元相同，主選單未見。獨立樣本公式及兩排程受驗完整狀態一致，51回填函式／27新增負例與CLI通過。
- 工具f6bf96a976fb31e19b19545ae438b3abcb2006fa已推送github隔離分支，回讀一致且乾淨，未推本機origin。主庫只更新既有四文件，詳細鎖定309／全部輸入、命令、收據與近似見研究紀錄。原始ZIP／patch／417根檔／EXE／MOX.SET雜湊再次通過。
- 本輪所有有界Go容器均正常退出移除；18:27:38 UTC的Go映像容器清查空，工具find . -user root -o -type d -name '*.md'空，新檔及收據1000:1000。未清理其他專案、映像或主庫既有root-owned檔。下一步公開DOS AH2Ah日期契約與可重播來源，不猜即時日期、不深入音效ISR／硬體時序。整款remake與主選單仍未完成。

### 2026-10-03 明示 DOS 日期與 word 完整立即值減法

- 接手主庫538a1f405cdc0c39c4e2d4dbec89e1010425f8f2／工具f6bf96a976fb31e19b19545ae438b3abcb2006fa，兩工作樹乾淨。路由命中平台規格優先、dosgolem對拍與結論回填，沿既有remake逆向技能；主庫玩法RE閘門保持。
- 310依Microsoft DOS AH2Ah公開契約先DRAFT、審查READY後接明示日期與共用虛擬時計。日期只能在已附掛、尚未執行時設定；不默認主機日期，無日期或時計／機器不一致即拒絕。正常執行證明下一缺件是66 81 /5 word SUB，311依Intel契約READY後接完整imm16，保留高16位與其他核心。真正原版caller形成01600101h，兩次寫入SS:ESP+8／+4。
- 六個日期／CPU測試、全部65536低word×8暫存器的獨立旗標檢查與固定官方EXE全套go test -p 2 -buildvcs=false ./... -count=1通過。最初日期caller測試因缺word SUB而失敗，保留嚴格斷言，補311後通過；自製探針loopStep宣告位置編譯失敗，只移動宣告後同命令乾淨重跑。診斷與失敗收據保留，不當作產品玩法缺陷。
- 同一Go1.24.13固定映像，600s／2GiB／2CPU／128pids／UID1000／network none，417原檔與固定EXE重生三組原版。兩組DOSGOLEM_MOO2_CALENDAR_EPOCH=1996-01-01有／無既有受控滑鼠，均固定48M正常Esc；兩次日期、16條caller與真正四bytes堆疊写入通過。第三組不設日期，除唯讀觀測與PNG名稱外完整基線保持原拒絕停點。全部53回填函式、45個新增缺證據负例與兩CLI通過；309／310的舊停點同次追加勘誤。
- 新停點第48797763步、高位LE0x210C7E，bytes 66 03 05 A4 BE 29 00，word ADD AX,DS:[0x29BEA4]尚未支援；兩時計58554306、IRQ0完成8022、IRQ7完成304。三圖逐位元保持黑色過場，主選單未見。日期來源不是亂數seed；AH2Ch既有呼叫計數占位實作不宣稱系统日期／時間一致。
- 工具ba3239ce0696f2e9bf898b543cee04a8ff455cab已推送github隔離分支並回讀一致，工作樹乾淨，未推本機origin。主庫只更新四個既有文件；完整輸入、命令、來源與成功／失敗雜湊見研究紀錄及鎖定310／311。原始ZIP／patch／EXE／MOX.SET雜湊再驗通過，私有收據與原版素材留本機。
- 本輪有界容器均退出移除；19:08:18 UTC新增的兩個Go容器核對為hr專案，保留未操作。工具root-owned／異形md目錄檢查空，新檔與收據1000:1000；未清理其他專案、映像或主庫既有root-owned檔。下一步限定word記憶體ADD，不深入helper／driver／ISR。整款remake與正常玩家路徑仍未完成。

### 2026-10-03 word ADD 記憶體來源與原版載入畫面

- 上一輪完成日期／word SUB並推送，判定為進度。接手主庫fb7461f138e7a1efc44ad7eb2afc8a83cab1adc8／工具ba3239ce0696f2e9bf898b543cee04a8ff455cab，兩工作樹乾淨。路由命中平台規格優先，載入入口及規格流程／結論回填契約，沿既有remake逆向技能；主庫玩法RE閘門保持。
- 312先DRAFT與索引，有界唯讀診斷保留原CPU，確認DS0188:0029BEA4來源word2與完整311基線保持。Intel公開03 /r、既有地址／segment框架足以READY後接CPU記憶體來源。只讀來源完整取得後才發布低word及六旗標，其他核心／RAM保持，未深挖helper或猜欄位語意。
- 三新CPU／三既有word ADD測試、八目的×全部65536低word／獨立六旗標、完整非零浮點狀態、來源只讀、DS／SS／SIB／不對齊、截短／越界／bus錯誤與真正66 A3回歸全PASS。首次自製測試誤用不存在FPU欄位，只改為實際完整欄位後同Docker命令重跑，未改CPU或放寬斷言；失敗收據保留。
- 固定EXE全套go test -p 2 -buildvcs=false ./... -count=1全PASS。同Go1.24.13映像、network none／2GiB／2CPU／128pids／UID1000／600s，三組固定原檔／相同48M Esc流程通過。兩設定日期的真正ADD、两byte寫回與四caller已驗，未設定第三完整保持日期拒絕。54回填函式、31新增負例與CLI通過；309／310／311同一ADD停點追加勘誤。
- 兩設定日期的新停點第48919460步、高位LE0x14E3DE的word CMP CX,00D4h，兩時計58965328、IRQ0完成8059、IRQ7完成312。兩圖逐位元相同，已檢視原版Loading載入畫面與中央游標，主選單未驗。受控滑鼠有／無事件現在各完成1／0回呼；首次稽核誤要求計數相同，依原先明示初態核對各自結果，其餘欄位仍嚴格一致，未改產品或注入資料。
- 工具f2d982a7d9383a2b536d9540cb5b8b9e860f6284已推送github隔離分支並回讀一致、工作樹乾淨，未推本機origin。主庫只更新四份既有文件，精確來源、命令與成功／失敗收據見研究紀錄與鎖定312；原始ZIP／patch／417根檔／官方EXE／MOX.SET雜湊再次通過。
- 本輪有界容器均退出移除，19:34:25 UTC清查只餘其他hr專案Go容器，已核對掛載並保留。工具root-owned／異形md目錄檢查空，全部新檔與收據1000:1000，主庫既有root-owned檔維持；未清理其他專案或映像。下一步限定word CMP與原有有號分支。整款remake與正常玩家路徑仍未完成。

### 2026-10-03 word CMP 完整立即值與原版標題背景

- 上一輪完成word ADD、正常原版寫回與推送，判定為進度。接手主庫b98d6f0e3f11d6a2dba7ab637ad637f5215f9c5e／工具f2d982a7d9383a2b536d9540cb5b8b9e860f6284，工作樹乾淨。路由命中平台規格優先，沿既有規格／回填契約與remake逆向技能，主庫玩法RE閘門保持。
- 313先DRAFT／索引，重核兩312正式原版收據與Intel CMP／Jcc公開契約。word 81 whitelist缺暫存器group7是根因，READY後完整讀iw再只用sub16改六旗標，不寫來源。九個CMP／ADD來源／SUB目標回歸、八來源全word與獨立旗標、完整非零FPU／R／RAM、真正JL的45組有號／相等／溢位邊界與拒絕全部通過。
- 固定官方EXE全套go test -p 2 -buildvcs=false ./... -count=1全PASS；Go1.24.13固定映像，600s／2GiB／2CPU／128pids／UID1000／network none，同417原檔／48M Esc／50M上限／兩明示日期與第三未設定流程。原版前兩CMP／JL跳轉，CX212邊界不跳，六筆完整核心／SS窗與外層212到訪已驗；未設定第三保持日期拒絕，受控滑鼠兩初態各1／0回呼分開驗。
- 新停點第49564005步、高位LE0x24C31B的INT33／AX0014h，兩時計61027457，IRQ0完成8251、IRQ7完成357。兩圖逐位元相同，已實際檢視原版標題背景與游標，主選單按鈕未見。55回填函式、33新增負例與CLI全通過；309–312的同一CMP停點追加勘誤。沒有失敗測試需放寬；私有收據與原版素材留本機。
- 工具0cc36241a3523df865d9b8f336d70c124bc7093c已推送github隔離分支、回讀一致且乾淨，未推本機origin。主庫只更新四份既有文件；原始ZIP／patch／417根檔／官方EXE／MOX.SET雜湊再次通過，來源／命令／正式收據與近似見研究紀錄及鎖定313。
- 本輪有界容器均退出移除，19:57:00 UTC清查核對剩餘為hr／fd2專案，一個其他容器已在清查間自行移除，未操作其他專案或映像。工具root-owned／異形.md目錄檢查空，新檔與收據1000:1000；主庫既有root-owned檔維持。下一步限定滑鼠AX0014h平台契約。整款remake與正常玩家路徑仍未完成。

### 2026-10-03 滑鼠回呼交換與主選單面板滑入

- 接手主庫1491008edf177b29d767e25f620de35759e7cfce／工具0cc36241a3523df865d9b8f336d70c124bc7093c，工作樹乾淨。命中平台規格優先與結論回填路由，載入入口；沿既有remake逆向技能。主庫玩法RE閘門保持。
- 314依Microsoft Mouse交換語意、Watcom完整EDX及既有255模型，先DRAFT／索引、審查READY才接INT33／AX0014h。驗證新目標後返回舊mask與ES:完整EDX。32位元交換及佇列時序保留platform-spec approximation，不深入driver或helper。
- 四新測試、六滑鼠回歸及固定官方EXE全部go test -p 2 -buildvcs=false ./... -count=1通過；兩同日期／48M Esc原版與第三不設日期流程完成。來源、輸入、命令與私有收據雜湊見研究紀錄及鎖定314。
- 正常兩組各24次交換成功、前三次真正C3返回與15步框架讀取核對。均到50M上限0x23856E，無未支援指令，時計62461366、IRQ0完成8383；畫面已檢視主選單面板部分滑入，完整展開及操作未驗。既有受控回呼0／1與allocator selector差異分開記錄，未冒稱完整終態同狀態。
- 56回填函式、31缺證據負例及CLI通過，309–313同一AX0014h停點追加勘誤。首輪文件護欄缺完整冒號形式，補正ES:EDX文字後乾淨重跑；程式與正式收據保持，沒有產品測試失敗。
- 工具9a2c7a21b0901ac8acc1f4387729ce25072b5fd9已推送github隔離分支並回讀一致，未推本機origin。主庫只更新四既有文件，原始ZIP／patch／417根檔／EXE／MOX.SET雜湊再驗通過；原版與完整終端／PNG留本機。
- 本輪有界Docker均退出移除，20:20:54 UTC其餘Go容器為hr專案，保留未操作。工具root-owned／異形.md目錄檢查空，新檔／收據1000:1000。下一步在50M內補滑入階段的唯讀觀測與正常輸入條件，完整主選單／玩家路徑及整款remake未完成。

### 2026-10-03 主選單逐換頁觀測與終態勘誤

- 前輪完成314並推送，屬進度。接手主庫11573fef0d26c84895bf98c511cf2d8f98fdc4fa／工具9a2c7a21b0901ac8acc1f4387729ce25072b5fd9，工作樹乾淨。命中平台規格優先及結論回填，載入入口／文件職責；主庫玩法RE閘門保持。
- 原314收據實際已含label=terminal與step_limit_registers，前輪只搜尋limit標籤而誤稱無完整終態。本輪核對完整R／段／flags206h及IRQ7 started388／completed388，修正目前活表並在314追加勘誤，原始收據保持。
- 315先DRAFT／索引，確認VBE快照API只讀後審查READY；只在自製探針明示加入最多16張換頁快照，CPU／平台及主庫玩法不改。新probe三正常go run編譯成功，沿314條件；既有314全套PASS適用於未改來源，不重跑CPU語料。
- 三原版流程完成：兩日期各14張，第8..21換頁；第三無日期1張，共29張，全readonly=true。全部既有314列及兩圖逐位元保持，兩14張完整序列相同；10／16／21代表畫面已檢視，像素變動區末五張左緣619／608／596／583／571，支持面板持續移入，完整主選單／點擊未驗。
- 輔助Go像素量測首輪因64pids且未限制Go平行建置出現asm fork拒絕；限制GOMAXPROCS2／-p2後同映像乾淨重跑通過，分類為分析環境，正式原版流程未失敗。57回填函式、21新增缺證據負例與CLI通過。
- 工具f1c2fa57675991080e2da3d1f5008b9b49f209bc已推送github隔離分支並回讀一致，未推本機origin。主庫只更新四既有文件；原ZIP／patch／417根檔／EXE／MOX.SET重新核對保持，29張原版快照與完整終端只留本機。
- 20:38:16 UTC Go工具映像執行中／停止容器清查空，本輪有界容器已退出移除。工具root-owned／異形.md目錄空，修改檔與收據1000:1000。下一步先定明示46M Esc獨立排程契約，再驗正常輸入，48M基準與50M上限保持。整款remake與正常玩家路徑未完成。

### 2026-10-03 明示46M Esc與原版主選單按鈕

- 前輪315換頁證據已推送，屬進度。接手主庫655dd1832a76200f02e659400757aa9b67b5941e／工具f1c2fa57675991080e2da3d1f5008b9b49f209bc，工作樹乾淨。路由載入平台規格優先；主庫玩法RE閘門保持。
- 316先DRAFT／索引，核對307控制器／原版IRQ1與315階段後READY，只改探針明示Esc步數、讀EXE前的範圍／互斥閘門，以及沿排程的只讀快照門檻。CPU／平台／玩法不改，314全套PASS仍適用。
- 11真實CLI負例均exit2、stdout空、讀EXE前拒絕；四正常原版流程完成。固定46M／48M、1996-01-01與50M cap，沒有失敗後換鍵或挑時點。完整來源／命令／雜湊與近似見鎖定316。
- 兩46M原版IRQ1各97／77步CF返回，自然caller續行；兩到50M的0x2385AC，時計64282188、IRQ0完成8401、IRQ7完成427。兩圖逐位元相同，已檢視主選單六按鈕文字完整可見，點擊／動畫停穩／新遊戲未驗。48M與無日期全部舊列保持315，47張快照均唯讀、58回填函式／26負例／CLI通過。
- 工具cbc63f6ad19f17ca8eda81c1ab44f6852fb383e9已推送github隔離分支並回讀一致，未推本機origin。主庫只更新四既有文件。原ZIP／patch／417根檔／EXE／MOX.SET雜湊再驗保持，原圖／完整終端留本機。
- 21:00:17 UTC本輪有界容器已退出移除，其餘Go容器掛載為hr，保留未操作。工具root-owned／異形.md目錄空，檔案／收據1000:1000。下一步先核對NEW GAME可點條件及正常滑鼠輸入契約，保持50M；整款remake與正常玩家路徑未完成。

### 2026-10-03 原版NEW GAME正常輸入與兩次CB返回

- 接手主庫df8d8792a482f817f1a7b38bd08ba8b1c1845c91／工具cbc63f6ad19f17ca8eda81c1ab44f6852fb383e9，兩工作樹乾淨。命中平台規格優先，載入入口及文件職責，主庫玩法RE閘門保持。
- 317／318／319依序DRAFT、索引、READY後實作，只改自製探針與證據。五正常原版流程在Docker完成：317唯讀、318點擊／無點擊基線、319初始觀測／修正後觀測。均保留46M Esc、1996-01-01、50M cap；CPU／平台未改，314全套PASS沿用。go build -p 2 -buildvcs=false與實際執行、來源與收據雜湊見鎖定規格／研究紀錄。
- 第40換頁49882420步；正常按下49882421與放開49883408，真正原版兩CB於49882522／49883496返回，完整R／六段／flags恢復。原版word寫後0100／0000已取樣。模式關閉全部317列／終圖、點擊前前綴、319全部318列／終圖保持，沒有CPU拒絕。50M終圖仍主選單，未宣稱NEW GAME成功轉移。
- 319初次觀測在startup前安裝，且誤以委派ok判斷讀取，實際會被覆蓋／一般RAM返回false。分類觀測工具缺口，保留控制收據；修正掛勾時點與8／16位元委派後，同映像／命令／輸入乾淨重跑。目標讀取零筆仍無真實正對照，不推論沒有consumer。文件編輯曾因不存在的patch上下文拒絕，未改產品或正式收據。
- 59回填函式、32缺定位／原始收據／正對照未知／狀態／回填負例及CLI通過。工具062670051203076ff688d36a390f46dd8a7883c6已推送github隔離分支並回讀一致，未推本機origin。主庫只更新四既有文件；原ZIP／patch／417根檔／固定EXE／MOX.SET重新核對保持，原素材／完整終端／PNG不入Git。
- 21:35:20 UTC本輪有界容器已退出移除，剩餘Go容器為fd2，保留未操作。工具root-owned／異形.md目錄空，修改檔及收據1000:1000；主庫既有2709 root-owned保持，未遞迴修復。下一步先補按鍵讀取正對照，完整玩家流程與整款remake未完成。
- 主庫交接稽核先因字面比對比文件多「讀取」兩字、再因活表省略按下步數而失敗。重新命中結論回填路由，核對正式收據後補明49882421步，同容器命令加失敗即停止並乾淨重跑通過。五私有收據雜湊、鎖定三規格及未知邊界核對保持；原版與探針未變。

## 2026-10-03：原版事件消費與DOS問號搜尋

- 上一輪615e967／0626700為實質進度。本輪命中平台規格優先、結論回填與文件職責路由。依320–323先READY再修改隔離dosgolem；主庫Go玩法與RE-first閘門保持。
- 320取得真正8／16位元請求與正常word MOV正對照；321沿同46M正常單次點擊，核對原版七讀、28步續行與清除待處理事件，短按下未遺失。各原有原始列／終圖保持。
- 322預定獨立44M Esc／1996-01-01／50M cap，八CLI與無點擊／單次點擊通過；新原始拒絕是47995790步的save?.gam FindFirst。323依公開DOS契約實際列舉正版根檔，缺檔返回EAX12h／CF與DTA、12個caller步通過；無點擊全部基線及點擊搜尋前完整前綴保持。不是捏造空目錄或代寫原版欄位。
- 實際命令：Go1.24.13固定Docker映像，600s／2GiB／2CPU／128pids／UID1000／network none；固定EXE go test -p 2 -buildvcs=false ./... -count=1全套PASS，CPU386194.255s／machine6.660s。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後重生原版兩44M流程。定向測試的包裝型別編譯錯誤已修，同映像／同命令乾淨重跑PASS，沒有產品失敗宣稱。60回填函式、原有32／新增49缺證據負例及兩CLI通過。
- 新CPU阻塞49442083步、高位LE17122B／00 C3，完整狀態已保存。已實際檢視終圖仍主選單，設定畫面、正常開局與remake同狀態未完成。下一步為公開byte ADD契約與READY／實作／同輸入重生，不追檔案helper內部，不重點或提高cap。
- 原ZIP／官方patch／417根檔／EXE／MOX.SET重新核對，原始LOG／PNG／素材留忽略workplace。工具e75f5aebed41cccb062609235dd2f0b07ffae371已推送github並回讀一致、工作樹乾淨，未推本機origin；來源／收據與固定連結見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。本輪Docker一次性工作均已退出，工具無root-owned或誤建.md目錄；主庫既有root-owned2709項保持，沒有遞迴修權限。交付前另核對主庫差異與最終HEAD。
- 提交前文件差異審查發現活表替換邊界過寬，已修正；原RE策略與其餘待辦全文逐字保持，四文件／七私有收據／四鎖定規格／擁有權檢查PASS。問題版本未提交。

## 2026-10-03：byte ADD與記憶體目的

- 上一輪152bd13／e75f5ae為實質進度。本輪載入平台規格優先，文件職責與結論回填沿既有入口；324先READY再改隔離工具CPU，主庫Go玩法RE閘門保持。
- 補無前綴00 /r暫存器與記憶體目的，成功寫入後才發布六旗標。七新增主測試涵蓋所有byte對／64別名／DS／SS／SIB／界限／Bus與prefix拒絕。定向PASS；固定EXE全套go test -p 2 -buildvcs=false ./... -count=1 PASS，CPU386195.653s／machine6.976s。
- Docker Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none。先go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，兩條同44M Esc／日期／50M／separate DOS、只有點擊加正常單次輸入。無點擊全部323基線及PNG保持，點擊舊拒絕前4089列與第一ADD完整核心保持。
- 原版1暫存器／7記憶體ADD、24續行與CMP／JGE兩方向逐筆PASS。來源BL皆0，這次窗口保持；非零另由平台測試覆蓋。新阻塞49501135／高位LE2130F3／F7 5D D8記憶體NEG，SS來源尚未取樣，不由EAX推測。終圖仍主選單，設定畫面／正常開局與remake同狀態未完成。
- 首次唯讀呼叫自動核准審查逾時、未建立程序，依回報重試一次成功。前綴稽核初版錯含舊失敗後快照，修正切點並加第一ADD原始核心核對後PASS；正式來源／收據不改，均未寫成產品失敗。61回填函式、原有32／49與新增25負例及兩CLI PASS。
- 工具f31793166415705c75a81339808372da24cded44已推送github、回讀一致與工作樹乾淨，未推本機origin。來源／私有收據／固定規格連結見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。本輪一次性Docker工作均已退出；工具無root-owned或誤建.md目錄，主庫既有2437檔案／272目錄保持，不遞迴修權限。最小下一步為公開dword記憶體NEG契約與同原版輸入驗證，不重開已驗ADD／CB／FindFirst。

## 2026-10-03：32位記憶體NEG與正常續行

- 上一輪主庫ea1ff9d5350794c75a50fe8bef3adbf3265f161c／工具f31793166415705c75a81339808372da24cded44已驗ADD。本輪載入平台規格優先、文件職責；325先DRAFT／READY，再改通用CPU，主庫Go玩法RE-first閘門保持。
- 補無前綴F7 /3一般32位記憶體NEG，保留描述子與逐byte Bus近似，全部寫回成功後才發布六旗標。六新增主測試與ESP／暫存器回歸通過；合法EBP SIB納入後，原錯誤SIB負例改驗LOCK，084／180／324及索引回填。
- Docker Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none。定向PASS；固定EXE DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1 全套PASS，CPU386143.926s／machine3.148s。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後，兩流程沿同44M Esc／1996-01-01／50M cap／separate DOS與單次正常NEW GAME輸入重生。
- 原版三次2130F3／F7 5D D8真正SS188:2BD9CC由FFFFFFFFh→1h，flags286h→213h；九MOV／CMP／JGE不跳續行與窗口／完整核心逐筆PASS，跳轉方向仍未知。無點擊全部324基線／兩終圖、點擊舊拒絕前4123列與第一NEG核心保持。兩CB仍191步、started2／completed2；點擊到50M／高位LE21334F，無新CPU拒絕，設定畫面仍未知。
- 62回填函式、原有32／49／25與新增27缺證據負例、三CLI PASS；本機獨立重播腳本與收據留忽略workplace。原版ZIP／patch／417根檔／EXE／MOX.SET雜湊再核對PASS。該來源稽核首次呼叫誤將已含here-document的shell腳本再包入Python，修正呼叫後在相同容器工具鏈乾淨重跑；來源與產品未改，不列產品缺陷。
- 工具31ca939e73203bafa2f55d7d23ae6115e9dc54c2已推送github隔離分支，回讀一致且工作樹乾淨，未推本機origin。精確來源／輸入／七私有收據及鎖定規格見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。一次性Docker工作已退出，無本專案遺留容器；其他專案保留。新檔UID:GID1000:1000，工具root-owned／誤建.md目錄自檢空；主庫既有2437檔案／272目錄保持，不遞迴修權限。
- 最小下一步為同50M上限與單次正常輸入的後段唯讀、有界進度／畫面觀測，查明設定畫面未出現的最小阻塞。不加cap、重點、代寫或先調整輸入；正常開局、remake同狀態與整款中文化未完成。

## 2026-10-03：NEW GAME後段唯讀進度

- 上一輪主庫6cc6dc6dbc2610a041394bfda71bb829c3f4e3e3／工具31ca939e73203bafa2f55d7d23ae6115e9dc54c2已驗NEG。本輪命中平台規格優先與文件職責；326先DRAFT／READY，只改原版探針觀測，CPU／平台與主庫玩法保持。
- 六時點49500000至50M保存完整R／六段／flags／FPU、VBE、RAM雜湊與原始SS／DS窗口；前後readonly全部PASS。五區段各100000真正Step、決定性熱門位址計數通過。畫面與VBE換頁／Writes不變，R／堆疊／RAM不同；正常開局仍未驗。
- Docker固定Go1.24.13映像、600s／2GiB／2CPU／128pids／UID1000／network none。唯讀ZIP／patch乾淨重建417根檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿325的兩組同44M Esc／1996-01-01／50M cap／separate DOS與單次正常點擊重生。CPU來源逐位元保持，325固定EXE全套PASS沿用。
- 獨立python3 workplace/post-click-326-verify.py PASS：無點擊3847列、點擊4202列與兩終圖保持，六PNG CRC／RGB／SHA、完整RAM不突變、計數排序／窗口通過。五相鄰與首末不同像素皆0；首末實際檢視仍主選單，設定畫面未知。原版ZIP／patch／417根檔／EXE／MOX.SET再核對PASS。
- 真正末尾來源已定位DS:29BE74→DS:[EAX]→SS:[EBP-8]及CMP／JE／JLE；來源指標、窗口與完整flags／目的寫回待最小取樣。ESI零窗口不當作此來源，不猜素材故障。63回填函式、原有32／49／25／27與326新增27缺證據負例、兩CLI PASS。
- 工具ca437d1b17135d511b52e7fb40591b7ea12176c4已推送github隔離分支、回讀一致與工作樹乾淨，未推本機origin。[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)保存五私有收據、固定規格與來源。新來源／六PNG／收據1000:1000；工具root-owned／誤建.md目錄自檢空，主庫既有2437檔案／272目錄保持。本輪一次性Docker工作已退出，無本專案遺留容器，其他專案保留。
- 下一步沿同50M上限與正常單次輸入，核對真實來源／分支與必要的目的RAM寫回，不追完整renderer，不加cap、重點、代寫或先改輸入。正常開局與remake同狀態、整款remake／中文化目標保持。

## 2026-10-03：後段實際來源與原值寫回

- 前輪主庫a7f0c5df0b6b736766624829428a3cdab1d85d3b／工具ca437d1b17135d511b52e7fb40591b7ea12176c4。平台規格優先入口保持，327先DRAFT／READY，只改原版探針觀測。
- 同44M Esc／1996-01-01／50M cap與正常單次輸入重生兩流程，六組02h／80h／82h的指標、來源MOV、局部寫回與CMP／JE／JLE通過。160實際步保存，90步獨立核對，70步只記錄。兩次DS188:499300h／499303h寫FDh，中心原本就是FDh；80h後續與映射表來源未驗，不追完整helper。
- Docker Go1.24.13固定映像，600s／2GiB／2CPU／128pids／UID1000／network none。原ZIP／patch唯讀重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿326兩正式流程換327輸出；CPU／平台保持，325固定EXE全套PASS沿用。
- 初次獨立比較因六點完整RAM跨次雜湊不同停止；差異原因未證實。修正驗證腳本，明示正規化每次RAM雜湊，各次readonly與前後相同仍檢查；不宣稱跨次全部RAM保持。python3 workplace/post-click-source-327-verify.py乾淨重跑PASS：全部3847／4208列除明示正規化保持、六後段PNG與兩終圖逐位元保持，沒有新CPU拒絕，設定畫面仍未知。
- 64回填函式、原有32／49／25／27／27與新34缺證據負例、--check-post-click-source-spec-backlinks與--check-post-click-progress-spec-backlinks兩CLI PASS。ZIP／patch／417根檔／EXE／MOX.SET再核對，gofmt與CPU／平台逐位元保持PASS。
- 工具b6fc71c0ba5a426a7c496fc3a5bd342cca1515ec已推送github隔離分支、回讀精確一致與工作樹乾淨，未推本機origin。五私有收據與鎖定規格見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。工具root-owned／誤建.md目錄自檢空，新來源與收據1000:1000。
- 下一步查既有VBE服務與正常trace，核對目的RAM的最小畫面發布契約，不加cap／重點／代寫／先改輸入。正常開局與remake同狀態未完成，主庫玩法RE閘門與整款remake／中文化目標保持。
- 主庫四文件／五私有收據／鎖定327／其餘活表全文保持／來源雜湊與擁有權核對PASS；既有root-owned2437檔案／272目錄保持。Docker ps -a分別以主庫與隔離工具鏈掛載路徑篩選皆空，本輪沒有遺留容器，其他專案未清理。

## 2026-10-03：後段RAM取用與VBE提交對帳

- 起點主庫5bb8e0aeb1573409e9a983be81d3058874d7291d／工具b6fc71c0ba5a426a7c496fc3a5bd342cca1515ec。命中並載入平台規格優先與文件職責；328先DRAFT／READY，只改原版probe Bus觀測，不改CPU／平台或主庫玩法。
- 兩個正常50M流程重生，後段500000步Bus讀5775248／寫686978／errors0；兩目標讀回0／寫3與5、來源11byte區154讀。三源MOV與兩目的MOV正對照吻合，VBE寫／換頁差額0。後來D5h／00h寫回不當畫面發布；設定畫面仍未知。
- Docker Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none；唯讀ZIP／patch重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿327兩正式環境換328輸出名。CPU／平台來源保持，325固定EXE全套PASS沿用。
- python3 workplace/post-click-publish-328-verify.py PASS：全部3847／4369列除既定正規化保持、各次readonly與RAM前後相同、六後段PNG／兩終圖逐位元保持，沒有新CPU拒絕。65回填函式、原有32／49／25／27／27／34與新31缺證據負例，--check-post-click-publish-spec-backlinks／--check-post-click-source-spec-backlinks兩CLI PASS。
- 原ZIP／patch／417根檔／EXE／MOX.SET再核對、gofmt與CPU／平台逐位元保持PASS；最初265連結檔名誤寫已在READY前修正，未啟動錯誤實作。新來源／收據1000:1000，工具root-owned／誤建.md目錄自檢空。
- 工具963a57228f429b8e028570f9d4c1c9cfddf16837已推送github隔離分支，回讀精確一致與工作樹乾淨，未推本機origin。五私有收據與鎖定328見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 下一步另建有界續行觀測規格，保留50M正式基線與同44M Esc／單次輸入；不能因未達設定畫面而追整個renderer或猜CPU錯誤。正常開局／remake同狀態與整款remake／中文化目標保持。
- 主庫四文件／五私有收據／鎖定328與其餘活表全文保持核對PASS，既有root-owned2437檔案／272目錄保持。Docker ps -a以主庫與工具鏈掛載路徑篩選皆空，本輪無遺留容器，未清理其他專案。

## 2026-10-03：獨立100M續行與原50M保持

- 起點主庫2cb1504010ccc2a1c566e920c15244b9df294129／工具963a57228f429b8e028570f9d4c1c9cfddf16837。平台規格優先與文件分工已核對，329先DRAFT／READY。只改probe預算閘門與100M模式最多六點唯讀快照，CPU／平台與主庫玩法保持。
- Docker固定Go1.24.13映像、600s／2GiB／2CPU／128pids／UID1000／network none；唯讀ZIP／patch重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再python3 workplace/new-game-329-cli-verify.py、兩50M原基線與獨立100M單次正常點擊。14 CLI負例皆EXE讀取前exit2，三真實EXE正例通過。
- python3 workplace/new-game-329-verify.py PASS：50M全部3847／4382列與六後段PNG／兩終圖除既定正規化保持，100M同50M前綴4335列／核心／Bus／CB／VBE快照保持；六新PNG CRC／640×480 RGB／SHA、CPU／FPU／RAM／Bus不突變通過。
- 100M無新CPU拒絕，較晚兩目標各讀2次、VBE新寫52040與服務差額吻合；相鄰像素差0／186／408／484／891，變化在下方致謝文字區。50M／100M兩圖實際查看，仍主選單，NEW GAME激活與設定畫面未知。
- 66回填函式、原有32／49／25／27／27／34／31與新36缺證據負例、--check-bounded-new-game-spec-backlinks與--check-post-click-publish-spec-backlinks兩CLI PASS。原ZIP／patch／417根檔／EXE／MOX.SET再核對、gofmt與來源保持通過。
- 工具2bfb2db0f860d115cb0e96e9d1e5938a89a25c23已推送github隔離分支、回讀精確一致與工作樹乾淨，未推本機origin。七私有收據與鎖定329見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)；新來源／收據／PNG1000:1000，工具root-owned／誤建.md目錄自檢空。
- 下一步核對按下／放開與主選單事件讀取時序／最小NEW GAME激活，先證實實際阻塞，不繼續加預算，不追完整renderer／helper。正常開局與remake同狀態仍未完成，整款remake／中文化目標與主庫玩法RE閘門保持。
- 主庫四文件／七私有收據／鎖定329／其餘活表全文保持與來源雜湊核對PASS；既有root-owned2437檔案／272目錄保持。Docker ps -a以主庫與工具鏈掛載路徑篩選皆空，本輪無遺留容器，其他專案未清理。

## 2026-10-03：正常按下事件返回與上層非零分支

- 起點主庫870708cd415c38a69eed55a9d43cbf2391e19e24／工具2bfb2db0f860d115cb0e96e9d1e5938a89a25c23。命中dosgolem、平台規格優先與文件職責；結論前重核Watcom helper停止線。330先DRAFT／READY，只改原版probe唯讀觀察，主庫玩法保持。
- 初次第二組進入callee後的Jcc被標成caller，未採作正式上層收據；回到DRAFT補RET後首CALL立即停止及新stack return核對，再READY。初次收據保留本機330-initial。同Docker／命令乾淨重跑三流程；規格讀取誤用檔名與一次heredoc結尾錯誤均為工具操作，已核對落盤狀態並修正，不當產品缺陷。
- Docker固定Go1.24.13映像、600s／2GiB／2CPU／128pids／UID1000／network none；原ZIP／patch唯讀重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿329原兩50M與獨立100M環境換330輸出名。python3 workplace/new-game-330-verify.py PASS：全部3847／4382／6322舊列除既定正規化保持、72PNG逐位元保持，兩組32步獨立核算且50M／100M完整新列相同。
- 兩個放開後保留的按下事件均返回EAX1；213C83 RET到20DB69，TEST／JNE採非零臂到20DB87。另一組雙RET後CALL209325停止，不追helper。NEW GAME指令與正常開局仍未知，下一步只追20DB87按鈕判定／指令值，不需先延長點擊。
- 67回填函式、既有32／49／25／27／27／34／31／36與新增31缺證據負例、--check-event-return-spec-backlinks與--check-bounded-new-game-spec-backlinks兩CLI PASS。CPU／平台／CLI保持，325固定EXE全套與329 CLI仍有效；原ZIP／patch／417檔／EXE／MOX.SET、gofmt、git diff --check、來源與收據1000:1000核對通過。
- 工具34d758498f931d9dc155c4ca93309dd98646328e已推送github隔離分支、遠端回讀精確一致與工作樹乾淨，未推本機origin。六私有收據與鎖定330見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫RE-first與整款remake／中文化目標保持。
- 四文件／六私有收據／鎖定330／其餘活表全文保持與來源雜湊核對PASS。主庫既有root-owned2437檔案／272目錄保持，沒有新增root-owned或誤建.md目錄。Docker ps -a以主庫與工具鏈掛載路徑篩選皆空，工具root-owned／誤建.md目錄自檢空；本輪沒有遺留容器，未清理其他專案。

## 2026-10-03：caller判定範圍與第一筆跳過

- 起點主庫4c04e57bfad75c6a75ac378ba31541cd822174a7／工具34d758498f931d9dc155c4ca93309dd98646328e。命中dosgolem、規格閘門與文件職責；沿用逆向重製技能與helper停止線。331先DRAFT／READY，只改probe唯讀觀察，主庫玩法保持。
- 初次96步缺DS:26C480／29BE0E與DS:[EAX]來源bytes，回DRAFT補有界窗口，再READY。初次三收據留331-initial；同容器／命令乾淨重跑三正常流程，不改輸入。正式程序省略callee內部Step，只在原EIP／SS／ESP返回後續觀察，記明238省略步，不當作全連續trace。
- Docker固定Go1.24.13映像、600s／2GiB／2CPU／128pids／UID1000／network none；原ZIP／patch唯讀重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿330兩50M與獨立100M環境換331輸出名。python3 workplace/new-game-331-verify.py PASS：全部3847／4414／6354舊列除既定正規化保持、72PNG逐位元保持，全部96 caller步／五自然返回／原bytes與定義flags／寫回獨立核算通過，兩預算完整新列相同。
- 原版返回x500／y229，實際表指標298848／count9／55byte stride；index1四word10／20／25／35，index2為20／30／35／45。第一筆因x500>25而JLE不跳，續查index2。sample96停在index2讀取之後，caller RET與最終命中未取得，不稱整個表不命中或資料錯誤。下一步核對index2..8同輪判定、caller返回及目前主選單的關係，不改點擊時長或原流程cap。
- 68回填函式、既有32／49／25／27／27／34／31／36／31與新37缺證據負例、--check-button-branch-spec-backlinks／--check-event-return-spec-backlinks兩CLI PASS。CPU／平台／CLI保持，325固定EXE全套與329 CLI有效；原ZIP／patch／EXE／MOX.SET／417檔、gofmt、Git差異與來源／收據1000:1000核對通過。
- 工具a48f536a1132731c1b055e4419854642177b1c5e已推送github隔離分支、遠端回讀精確一致與工作樹乾淨，未推本機origin。六私有收據與鎖定331見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。正常開局／NEW GAME指令與remake同狀態仍未完成，主庫RE-first與整款remake／中文化目標保持。
- 四文件／六私有收據／鎖定331／其餘活表全文保持與來源雜湊核對PASS。主庫既有root-owned2437檔案／272目錄保持，沒有新增root-owned或誤建.md目錄。Docker ps -a以主庫與工具鏈掛載路徑篩選皆空，工具root-owned／誤建.md目錄自檢空；本輪沒有遺留容器，未清理其他專案。

## 2026-10-03：完整範圍命中與後續CALL

- 起點主庫1c114dcd23bd06d4acdf52d1950d3885576723b6／工具a48f536a1132731c1b055e4419854642177b1c5e。命中dosgolem、規格閘門、平台停止線及文件職責，沿用逆向重製技能。332先DRAFT／READY，只續原caller觀察；主庫玩法保持。
- 初次237步把已完成IRQ7777→7778誤判為轉向，回DRAFT只修觀察器的成對增量分類，再READY。同Docker命令乾淨重跑三正常流程，初次收據留332-initial，CPU／平台／CLI／原輸入與流程cap未改。
- 固定Go1.24.13映像／600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；python3 workplace/new-game-332-verify.py PASS。全部3847／4511／6451舊列與72PNG保持，舊331 terminal保持；新313步與341省略callee步兩預算一致，311完整來源、2 MOV僅低word來源已驗、1 IRQ堆疊寫回未重建，未宣稱全記憶體精確一致。
- index1..6右界拒絕、index7左界拒絕、index8全畫面範圍命中x500／y229，實際選中8。CALL208FD4正常返回EAX1，CALL209325待20DDF7返回；8192只停觀察，沒有新CPU拒絕。下一步核對原註冊表與可見主選單關係及正常返回，不猜指令、不改輸入或流程cap。NEW GAME與正常開局仍未知。
- 69回填函式、既有32／49／25／27／27／34／31／36／31／37及新增34缺證據負例、兩CLI通過；原ZIP／patch／EXE／MOX.SET／417檔、CPU／平台來源、gofmt及新來源／收據1000:1000通過。330 SS20h舊註記的SS188h勘誤追加研究紀錄，保留歷史原定位與收據。
- 工具3580b3e26181ff0978fc7ed0b2c45f8c685f76e3已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin。六私有收據及鎖定332見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫RE-first與整款remake／中文化目標保持。
- Docker兩掛載篩選空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。本輪主庫只改四現況／工作／研究文件，既有root-owned 2437檔／272目錄保持；主庫提交、遠端回讀及清理狀態於輪末核對。

## 2026-10-03：原表更換與正式選單位置

- 起點主庫178d9bf932828aeb5ddf159f9f9d75e3d44d2949／工具3580b3e26181ff0978fc7ed0b2c45f8c685f76e3。命中dosgolem／規格閘門／文件職責，333先DRAFT／READY，只新增最多16份原表快照與既有cap內的自然EIP／SS／ESP返回觀察。原CPU／平台／CLI、輸入與舊332終態不改。
- 固定Go1.24.13映像／600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔；go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿332三流程換333輸出各一次。python3 workplace/new-game-333-verify.py PASS：全部3847／4825／6765舊列與72PNG保持，兩預算共同新快照一致、readonly通過。沒有新CPU拒絕。
- 47990733原CALL正常返回20DDF7／SS188／ESP2BDAD8，開始與返回完整9筆表保持；50M／100M同指標298848已更換相同7筆表。終態第2筆範圍415／217／567／238幾何命中500／229，人工查看PNG強推論對應NEW GAME；實際NEW GAME指令未驗。下一步保留舊基線，新增7筆表就緒後的一次正常點擊，不提高100M cap，不深挖整個callee。
- 70回填函式、既有32／49／25／27／27／34／31／36／31／37／34與新增35缺證據負例、兩CLI PASS。CPU／startup／provider／matcher／CLI保持，325固定EXE全套及329 CLI有效；原ZIP／patch／EXE／MOX.SET／417檔、gofmt、Git差異及新來源／收據1000:1000核對通過。
- 工具d6688b01f7a5306bc6d271e06c4a5eb48a1eb430已推送github隔離分支，遠端回讀一致與工作樹乾淨，未推本機origin。六收據及鎖定333見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫玩法RE-first與完整remake／中文化目標保持。
- Docker兩工作區掛載篩選空，工具root-owned／誤建.md目錄自檢空，其他專案未清理；主庫本輪四文件與既有root-owned 2437檔／272目錄於輪末核對，沒有新增root-owned。

## 2026-10-03：正式選單點擊與REPNE SCASW缺口

- 起點主庫6dbb1879f0c9ab6642aa0de3d8c1b504a800248c／工具d6688b01f7a5306bc6d271e06c4a5eb48a1eb430。命中dosgolem／規格閘門／原版GUI／文件職責路由，334先DRAFT／READY，只加明示選單就緒情境；主庫玩法保持。
- 初次重播腳本更新輸出輪次誤改預期EXE雜湊，輸入檢查停止，未啟動原版；修正腳本後同Docker設定乾淨重跑。Go1.24.13固定映像／600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；沿333三基線換334名各一次，加ready情境一次。
- python3 workplace/new-game-334-verify.py PASS：無旗標全部3847／4829／6769舊列與72PNG保持；新情境50M額外輸入前4780原列保持。7筆表／385bytes符合後50000000 press、50011955 release，相隔31339微秒，callback4／4完成；61538983原20DDDB實際word0000→0200，選中第2筆已證實。
- 後續76658331原1F3640的F2 66 AF被CPU拒絕，最終PNG全黑，未到100M或新遊戲設定頁。下一步補REPNE SCASW硬體契約與CPU驗證，再沿同ready情境重跑；不增加輸入／cap，不深挖原函式。完整正常開局與remake同狀態未完成。
- 14新增CLI拒絕與有效正對照、71回填函式、既有32／49／25／27／27／34／31／36／31／37／34／35與新增33缺證據負例、兩CLI通過；原ZIP／patch／EXE／MOX.SET／417檔、gofmt／Git差異及新來源／收據1000:1000通過。CPU／startup／provider／matcher未改；325固定EXE既有全套不涵蓋新拒絕。
- 工具4501b831842f33ee5a0388b3018ae0c240949bc3已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin；鎖定334與九私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫RE-first與整款remake／中文化目標保持。
- Docker兩工作區掛載篩選空，工具root-owned／誤建.md目錄自檢空；進度讀取時原容器已自動移除，原PTY隨後回報完成，未重啟或重點。其他專案未清理，主庫本輪四文件及既有root-owned 2437檔／272目錄於輪末核對。

## 2026-10-03：REPNE SCASW修正與原版設定頁

- 起點主庫884dec265c8ecef9390d1e8d0aa3574819e9b7f7／工具4501b831842f33ee5a0388b3018ae0c240949bc3。命中CPU公開契約／dosgolem／規格閘門／文件職責路由；335先DRAFT保存未改CPU完整初態，再READY實作通用F2＋16位AF。主庫玩法保持。
- 自製公開Intel契約與獨立算術oracle涵蓋92416組合及方向／多元素／高ECX／EDI繞回／逐byte故障／前綴拒絕。CPU窄測試PASS 0.266s；固定EXE Go全套PASS，cpu386 106.777s。不稱386實機逐週期驗證。
- 同334四原版各一次乾淨重生，Go1.24.13固定Docker／600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔，輸入與100M cap不改。python3 workplace/new-game-335-verify.py PASS：三舊基線3847／4829／6769列及72PNG保持，ready入口前5833原列保持；原F266AF結果ECX3／EDI1F3587與下一MOV EAX64h獨立核算。
- 同ready原版到100M無新CPU拒絕，NEW GAME設定頁PNG人工確認；334黑屏與CPU拒絕由新收據解除。原正常press／release、第2筆store與callback4／4保持，ACCEPT、完整開局與remake同狀態未驗。下一步保存設定頁按鈕表與ACCEPT正常輸入前置。
- 72回填函式、既有負例與新增26缺證據負例、兩CLI通過；來源／新收據1000:1000、gofmt／Git差異檢查通過。startup／provider／matcher保持334；工具root-owned／誤建.md目錄自檢空。Docker兩工作區掛載篩選空，沒有本輪遺留容器，未清理其他專案。
- 工具74f574a78927f6bacdec95ea0519c83078bc71df已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin；固定335與十私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫RE-first與整款remake／中文化目標保持，主庫本輪四文件與十私有收據核對通過，既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄。

## 2026-10-03：原版正常ACCEPT與選族頁

- 起點主庫da52def1be5f3ae36ea713f2c0318054fd9a6c53／工具74f574a78927f6bacdec95ea0519c83078bc71df。命中dosgolem／GUI輸入／規格閘門／文件職責路由；336先DRAFT取得原17筆表，扣三新列後8146舊ready列與全部圖保持，935bytes三份相同與原ACCEPT候選足夠後READY。只增明示SETUP_ACCEPT_CLICK=1與80M前置，不代寫原選擇。
- 初態一次600s；正式五原版各情境一次900s外層逾時，固定Go1.24.13 Docker／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔。原版每條仍100M cap；go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；python3 workplace/new-game-336-verify.py PASS：四舊基線3847／4829／6769／8146列及102PNG保持，額外80M輸入前6145原列保持。
- 80000000 press／80011248 release，差42912微秒，callback6／6完成；80124668原20DDDB word0000→0F00，index15實際選擇已核對。原版同100M到21595F、無新CPU拒絕，SELECT RACE選族頁PNG人工確認。已保存90M／100M原16筆表供下一正常選擇；種族／完整開局與remake同狀態仍未知。
- 15新增CLI拒絕與有效正對照、同binary舊334 14拒絕與正對照、73回填函式與新增28缺證據負例、兩CLI通過。CPU／startup／provider／matcher保持335，335固定EXE全套仍有效；gofmt／Git差異及新來源／收據1000:1000通過。工具root-owned／誤建.md目錄自檢空，Docker兩工作區掛載篩選空，沒有本輪遺留容器，未清理其他專案。
- 工具7c84f3931953cf3c9ffcbdc0718852e9c700b3ba已推送github隔離分支，遠端回讀一致、未推本機origin；固定336與十三私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫RE-first與整款remake／中文化目標保持，主庫四文件與十三私有收據核對通過，既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄。下一步只核對原16筆選族表與正常輸入前置，不重開ACCEPT或SCASW。

## 2026-10-03：原版正常人類選擇與統治者名稱頁

- 起點主庫 40e3a404d22636ae85826d0a0da8da182e8ce457／工具 7c84f3931953cf3c9ffcbdc0718852e9c700b3ba。命中 dosgolem、GUI 輸入、規格閘門與文件職責路由。337 先 DRAFT，獨立核對既有 336 的原 16 筆表與 90M 核心狀態，來源足夠後 READY；只新增明示 RACE_HUMANS_CLICK=1，沒有修改 CPU／平台或主庫玩法。
- 正式六個原版情境各重生一次，Docker 外層 900s／2GiB／2CPU／128pids／UID1000／network none，Go1.24.13，原 ZIP／patch 唯讀重建 417 根檔。原版每條仍 100M cap；go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。初態沿用既有正式收據，不重跑相同初態。
- python3 workplace/new-game-337-verify.py PASS：五條舊基準 3847／4829／6769／8149／7752 列與 132 張 PNG 保持，額外 90M 輸入前 6821 原列保持獨立 ACCEPT。90000000 按下／90010495 放開，差 42293 微秒；90056672 原 20DDDB word0000→0700，callback8／8 完成。原版同 100M 到 215DEE、無新 CPU 拒絕；Enter Ruler Name、預設 Strader 與 ACCEPT 已人工確認。
- 新增 16 個 CLI 拒絕案例與有效正對照、同一 binary 的舊 336 CLI 15 個拒絕案例與正對照、74 項文件檢查及新增 27 個缺證據負例、兩個 CLI 通過。CPU／startup／provider／matcher 保持 335；本輪未重新執行 Go 全套。來源／新收據 1000:1000、gofmt 與 Git 差異通過。工具 root-owned／誤建 .md 目錄自檢空；Docker 兩工作區掛載篩選空，沒有本輪遺留容器。
- 工具 0633c346ce2ca1156ab26c1dbc533ec0e43920e0 已推送 github 隔離分支，遠端回讀一致，未推本機 origin；十三份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。主庫四文件、十三份私有收據與原始 ZIP／patch 雜湊核對通過，既有 root-owned 2437 檔／272 目錄保持，本輪未新增 root-owned 或 .md 目錄。主庫 RE-first 與整款 remake／中文化目標保持。
- 下一步保存較早時點的名稱頁、原名稱緩衝區與正常確認前置。名稱確認、typed 種族特性、完整開局、正式 RNG 與 remake 同狀態仍未驗。

## 2026-10-03：原版正常名稱確認與旗幟頁

- 起點主庫 5c1b6a4a42d3ff8fc69125d00587cd66277eef96／工具 0633c346ce2ca1156ab26c1dbc533ec0e43920e0。命中 dosgolem、GUI 還原、規格閘門與文件職責路由。338 先 DRAFT，私有唯讀探針補齊 IRQ 建置依賴後蒐證一次；原 95M／97M／100M 名稱表 165 bytes 相同，9030 原列與 30 PNG 保持，來源足夠後 READY。名稱候選的字串長度驗證修正後，直接讀既有收據，未重跑原版。
- 初版七情境由乾淨 417 根檔各重生一次，Docker 外層900s／2GiB／2CPU／128pids／UID1000／network none、Go1.24.13、原 ZIP／patch 唯讀，每條仍 100M cap。名稱按下後原 mask1，探針仍要求 mask2B，因此未放開；沒有新 CPU 拒絕。保存首次來源及收據，退回 DRAFT 核對 InjectMouseEvent 後重新 READY。修訂僅位於 rulerAccept 明示分支，不重跑六條未受影響基準。
- 新版只重生 ruler 與新舊 CLI，Docker 外層300s。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。95000000 按下／95015426 放開，差47572微秒；callback10／10 完成，同100M到228E0E、無新CPU拒絕。SELECT BANNER COLOR 八色頁人工確認，原10筆／550bytes表已保存。共享20DDDB沒有命中，正式持久名稱writer仍未知。
- python3 workplace/new-game-338-verify.py PASS：六舊情境3847／4829／6769／8149／7752／9030原列與162PNG保持，新輸入前7523原列保持獨立humans；原表／候選32bytes／核心／RGB獨立核算通過。新版17新CLI拒絕與正對照、同binary舊337 16拒絕與正對照、75項規格檢查與新338 26負例、337 27負例及兩CLI通過。CPU／平台保持335，本輪未重跑Go全套。
- 工具 0c88d04cb59d2b7bfc716e953cc529c54b5fb6e2 已推送github隔離分支，遠端回讀一致，未推本機origin；十八份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。新來源與收據1000:1000，gofmt／Git差異通過，工具root-owned與誤建.md目錄空，Docker兩工作區掛載篩選空。主庫四文件、十八份收據與原始ZIP／patch雜湊核對通過，其餘活表全文保持；既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄。
- 下一步取得較早旗幟頁、原10筆表與正常選色前置。原預設名稱確認及下一頁已驗；typed種族特性、正式名稱writer、選色、完整開局、正式RNG與remake同狀態未知。主庫RE-first及整款remake／中文化目標保持。

## 2026-10-03：旗幟正常按鍵查詢已驗，選色未完成

- 起點主庫f079ef801ab20d3ab1949ad9ee8d5e9693ee9eea／工具0c88d04cb59d2b7bfc716e953cc529c54b5fb6e2。命中dosgolem、GUI、規格閘門、文件職責與回填路由。339先DRAFT，98M／99M唯讀蒐證保持全部7874原ruler列與30PNG；98M尚一筆初始化表，99M才完整十筆。READY後不更動原初態或正式玩法預設。
- 新輸入三版各由新鮮417根檔／官方EXE重生一次：初版99M正常短按、100M仍旗幟頁；同短按延長120M、7837原100M前列保持，但仍未推進。兩次都保存來源與負收據，退回DRAFT查證，不稱選色成功。
- 收據直接確認短按期間原AX3輪詢為0次、放開後492次只讀到buttons0。重查GUI路由及既有INT33／InjectMouseEvent後重新READY，只加「首次原正常查詢確實讀到pressed」的放開閘門。最新99000000按下、99083819返回BX1／CX276／DX190、99083854首次合法mask1放開，差245291微秒；callback12／12與IRQ28866／28866完成。
- 原版120M到228DDC，仍SELECT BANNER COLOR，沒有新CPU拒絕；共享20DDDB未命中。339限定CONFORMED只涵蓋原表、正常查詢與放開，選色消費／下一頁／持久旗色仍未知，不計入玩法完成分母。下一步只保存原pressed查詢後的有界正常GUI消費與返回框架，不再調cap、換時點或盲重送。
- Docker外層300s／2GiB／2CPU／128pids／UID1000／network none，Go1.24.13，原ZIP／patch唯讀，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。新99M前7708原列及28PNG、完整表／核心／RGB／pressed順序通過；22新CLI拒絕及120M／100M正對照、同binary舊338 17拒絕及正對照、76規格回填、新339 26缺證據負例、338 26負例與兩CLI通過。未重跑未受影響六基準／Go全套；CPU／平台保持335，來源六有界區塊／五guard逆轉後逐位元保持338。
- 工具7d569e0392fd9061576da8395d3c8b4d10e0886d已推送github隔離分支，遠端回讀一致，未推本機origin；22份新私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。來源／新收據1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空。主庫RE-first及整款remake／中文化目標保持。

主庫四文件、22私有收據、精確工具HEAD與原ZIP／patch雜湊核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄。

## 2026-10-03：原GUI按鍵返回已驗，旗幟選色仍未完成

- 起點主庫3526b6da8334626a992cfa3c13689b318befb664／工具7d569e0392fd9061576da8395d3c8b4d10e0886d。命中dosgolem、GUI、規格閘門、文件職責與回填路由。340先DRAFT，私有readonly探針保存192正常步；全部10012原339列與32PNG保持，原source剝除5個trace區塊後保持339。
- 原339首個AX3讀到1之後，20DB56 caller又呼叫214075；第二次查詢與214104 RET到20DB5B均返回0，20DB5E JNZ不跳。原框架證據足夠後重新READY，只增加明示旗幟fixture的GUI按鍵返回閘門，不修改CPU／平台或主庫玩法。
- 正式340重生一次：99M按下保持，99083819首次查詢1，99083999原RET返回1，完整stack／核心／readonly已核對，99084000首次合法mask1放開，差245437微秒。120M到228E00，callback12／12及IRQ28866／28866完成，仍SELECT BANNER COLOR，無新CPU拒絕；共享20DDDB未命中。只證實原GUI按鍵返回，不能稱選色或開局完成。
- Go1.24.13 Docker每次外層300s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀，新鮮417根檔／固定EXE／MOX.SET。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。7708原列／28PNG／完整原表及RET-release核算通過；核算誤取checkpoint IRQ欄位，改讀同收據setup_table_snapshot後PASS，未重跑原版。3有界觀測及一個release前置之外全部來源保持339。
- 同binary舊338 CLI17拒絕及正對照、339 CLI22拒絕及120M／100M正對照，77項規格回填、340新增28缺證據負例、338／339各26負例及三CLI通過。未重跑未受影響六基準／Go全套。工具0b141e6cf2a00d86c028054c68d247c9e526a5c0已推送github隔離分支，遠端回讀一致，未推本機origin；15份來源／收據／核算見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 新檔1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空。主庫四文件、原來源／收據與其餘活表核對後提交；主庫RE-first與整款remake／中文化目標保持。
- 下一步只保存原20DB5B後最多192個非callback／IRQ正常步與實際20DB5E分支。原選色消費、持久名稱／旗色writer、typed種族特性、完整開局、正式RNG及remake同狀態未知，不盲調cap或重送。

主庫四文件、15份私有來源／收據、工具精確HEAD、原ZIP／patch與CPU來源核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；Docker兩工作區掛載篩選空。

## 2026-10-03：原後段按鍵返回已驗，續行揭露記憶體TEST缺口

- 起點主庫f93e17de0f1268a2bd6251da3afe1fc654ff6c4b／工具0b141e6cf2a00d86c028054c68d247c9e526a5c0。前輪已改變正式source與原證據，分類為進展；本輪命中dosgolem、GUI、規格閘門、文件職責與回填路由。341先DRAFT，兩份獨立private prototype各保存192正常步，全部9968原340列與32PNG保持。
- 原20DB5E確實走非零分支，原X139／Y190與raw locals已驗；後段20E160又CALL214075、RET20E165返回0，20E168 JZ跳20E4EB。原GUI收到第一個1仍不足以保證後段按鍵狀態。兩份source／bytes／stack／分支核算足夠後READY，只增加後段正常RET返回閘門，不改CPU／平台或主庫玩法。
- 正式fixture重生一次：99000000按下保持，99084354原RET20E165返回1、完整框架／核心／readonly valid，99084355首次合法mask1放開，差245792微秒；callback12／12完成。原版step99415524在184694 bytes85 82 19 52 26 00拒絕，error=TEST dword ModRM 82尚未支援，未達120M；終圖全黑，共享20DDDB未命中。probe exit0不當正常流程成功，原旗色結果與開局未驗。
- 每個原版情境Go1.24.13 Docker外層300s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀，新鮮417根檔／固定EXE／MOX.SET。兩份探索／一份正式重播各一次，固定日期不是seed。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。
- 正式7708原列／28PNG／完整表及兩RET／release／CPU負結果核算PASS。source三區塊／一個release逆轉後保持340；初版source核算腳本逆轉字串誤寫，修正後重讀同來源PASS，未重跑原版。舊338 CLI17拒絕及正對照、339 CLI22拒絕與120M／100M正對照，78項規格回填、341新增29缺證據負例、340 28負例、338／339各26負例及兩CLI通過。未重跑未受影響六基準／Go全套。
- 工具b07cda7188d7139da612f143ba68ea13cca21958已推送github隔離分支，遠端回讀一致，未推本機origin；24份來源／收據／核算見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。新檔1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空，沒有本輪遺留容器。
- 下一步先保存原85 82 @184694有效位址／source／mask／flags及SETNE／RET最小消費，經DRAFT→READY補CPU支援，保持同341輸入續行與固定EXE完整CPU測試。持久名稱／旗色writer、原旗色結果、下一頁／完整開局、RNG及remake同狀態未知，主庫RE-first與整款remake／中文化目標保持。

主庫四文件、24份私有來源／收據、工具精確HEAD、原ZIP／patch與CPU來源核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；Docker兩工作區掛載篩選空。

## 2026-10-03：記憶體TEST與原版宇宙生成畫面

- 起點主庫34ace3c4ed5b4418882f6ab83cdf176a3cd325c7／工具b07cda7188d7139da612f143ba68ea13cca21958。路由命中dosgolem、CPU平台契約、規格閘門、文件職責與逆向回填；342先DRAFT，未改CPU原來源1／mask0／stack返回1846E3與完整readonly取得，全部7849原341列／29PNG保持後審查READY。
- 通用memory85唯讀TEST已接線，原99415524交集0／flags246h／SETNE AL0／RET1846E3三步核算通過；正式新入口前7789原341列／28PNG、341完整正常輸入／120M cap保持。原版實際進到「Generating Universe...」，113628909於17D536的0F9E拒絕；原完整開局未驗，沒有代寫遊戲資料、重送或提高cap。
- CPU窄測0.909s與固定EXE乾淨來源Go全套通過，cpu38655.710s、machine2.914s。初版CPU簽名錯誤建置即拒絕；全套初跑受忽略workplace探索main污染，乾淨版控來源加新測試後相同命令重跑PASS。初核算把舊stop診斷當入口前列，修正邊界後重讀同收據PASS，未重跑原版。8088外部實機語料缺檔，不當作硬體語料通過。
- Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，Go1.24.13／固定映像、原ZIP／patch唯讀與新鮮417根檔。新CPU全ModRM／SIB、唯讀及fault／prefix／寄存器保持／SETNE及RET兩方向已驗，79項回填、342新增25負例及340另兩負例、341的29／340的28／338及339各26負例與兩CLI通過。
- 工具0c6e871167259c382c6dac2288ace552c33e2319已推送github隔離分支，遠端回讀一致，未推本機origin。18份原來源／收據／核算hash、官方原ZIP／patch／EXE／MOX.SET、實際命令與未知邊界見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。公開只提交通用CPU／自製測試／spec與probe，原素材留本機。
- 主庫僅更新CONTEXT一行／WORKLIST正常路徑活表、追加本WORKLOG與既有研究紀錄，其他活表及玩法保持。source／新收據1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，main既有2437檔／272目錄保持；兩工作區掛載篩選Docker容器空。
- 下一步保存原SETLE完整初態與下一原byte store，依公開CPU契約走DRAFT→READY並固定同輸入續行。正式writer、生成完成／完整開局、RNG與remake同狀態未驗，主庫RE-first與整款remake／中文化目標保持。

## 2026-10-03：SETLE與原SS byte實際寫入

- 起點主庫ea13175382677216babef8a9b21801e2273d7f44／工具0c6e871167259c382c6dac2288ace552c33e2319。命中CPU平台契約、規格閘門、dosgolem、文件職責與逆向回填。343先DRAFT，原flags206h／完整R／六段／SS目的窗與stack唯讀保存，全部8237原342列／31PNG保持；Intel公開SETLE採ZF=1或SF≠OF，指出舊鏡像and錯，契約與來源足夠後READY。
- 補通用SETLE register，原113628909 AL28→0、EIP17D539、flags206h與其他核心／RAM保持；113628910下一SS byte實際Bus write linear2BDA34／value0／error nil、EIP17D53C、目的0→0及鄰居保持。因原byte0，額外保存真實write，不以RAM不變冒充消費。private三／正式五有界observer逆轉後source保持342，Bus只forward一次，原input／calendar／120M cap不改。
- 正式新CPU前8177原342列／30PNG與全部正常輸入保持。113628944原17D5A0的0F9F拒絕、flags293h，終圖逐位元保持宇宙生成圖，原生成完成／完整開局與writer未驗；probe exit0不能當完整開局通過。沒有重送、調GUI或深入helper。
- 2,097,152組byte／truth／flags與47,432組數學signedCMP→SETLE、core／SETE／SETNE保持、nextSS byte兩向及失敗回歸PASS。窄測4.352s、固定EXE乾淨來源Go全套PASS，cpu386189.431s、machine5.732s；外部386／8088實機語料未取得，不稱硬體語料驗收。80項回填、343新增26缺證據負例、342的25與340另兩負例、341的29／340的28／338及339各26負例與兩CLI通過。
- Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，Go1.24.13固定映像、原ZIP／patch唯讀，各次新鮮417根檔／官方EXE／MOX.SET。原版private初態／READY後正式各一次，未重跑六舊情境，固定日期不是seed。
- 工具eddee109e0d0d59e311f5c30e26961f84ee36560已推送github隔離分支，遠端回讀一致，未推本機origin。17份來源／收據／核算hash、官方輸入與實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。公開只提交CPU／自製測試／spec／probe，不含原素材。
- 主庫只更新CONTEXT一行／WORKLIST正常路徑活表，追加WORKLOG與既有研究紀錄，其他活表與玩法保持。新來源／收據1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，主庫既有2437root-owned檔／272目錄保持；兩工作區掛載篩選Docker容器空。
- 下一步以原SETG停點補剩餘標準SETcc register條件、完整初態與下一88 C2，走DRAFT→READY、同輸入／120M續行與固定EXE全套。正式writer、生成完成／完整開局、RNG與remake同狀態未驗，主庫RE-first與整款remake／中文化目標保持。

## 2026-10-03：完整標準SETcc暫存器條件與原SETG／MOV

- 起點主庫95dc2470e396bd902fa527c083d86a6f4a220d6c／工具eddee109e0d0d59e311f5c30e26961f84ee36560。命中CPU平台契約、規格閘門、dosgolem、文件職責與逆向回填。344先DRAFT，直接重查343不可變完整停止收據與0F9F拒絕路徑；未fetch ModRM或發布目的／flags，完整輸入充分後READY，不再重跑同一拒絕。
- 補完整16個裸register SETcc。原113628944 SETG AL E6→0、EIP17D5A3；下一113628945 MOV DL FA→0、EDX1FA→100、EIP17D5A5，完整其餘R／段／flags293h／FPU／RAM／VBE保持。兩筆readonly真，callback12／12、IRQ26735／26735非活動，budget一次臂兩步。
- 新CPU入口前8180原343列／30PNG保持，原343 SETLE／實際SS write與342 TEST三步保持。CPU小區塊逆轉後逐byte保持343，Jcc未改；probe三有界observer逆轉後保持343，原input／calendar／120M cap／Bus／hooks未改。舊343其他條件拒絕負例明確改為字面真值正例，其餘全部保持。
- 524,288組完整旗標／別名／unused欄、65,536組全部初byte、118,580組數學signed／unsigned CMP→SETcc與512組Jcc字面真值核算PASS；原SETLE／SETE／SETNE／SS byte兩方向及失敗回歸PASS。窄測5.386s，固定EXE乾淨來源Go全套PASS，cpu386143.652s／machine5.679s。81項回填、344新增25缺證據與其餘三份回填6負例、343的26／342的25與340另兩／341的29／340的28／338及339各26負例、兩CLI通過。
- 原版只重生一次，120M到17FCE4且無新CPU拒絕，終圖仍宇宙生成；生成完成／完整開局未驗。最新PNG不同於343，已重新人工檢視並保存新hash；不把游標／畫面bytes變化當新玩家頁。最後32步保存17FCC3..17FD13迴圈，但未推定其用途或宣稱正常生成完成。
- Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，Go1.24.13固定image、原ZIP／patch唯讀、新鮮417根檔／官方EXE／MOX.SET。14份私有來源／收據及實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。外部386／8088實機語料未取得，固定日期不是RNG seed。
- 工具d5127adc64a04af79796d933aec73413bbcfd824已推送github隔離分支，遠端回讀一致，未推本機origin。公開只提交CPU／自製測試／spec／probe，四份較早規格及索引／guard同次回填，原素材留本機。來源／新收據1000:1000、gofmt／Git差異與原官方輸入雜湊核對通過，工具root-owned／誤建.md目錄空。
- 主庫僅更新CONTEXT一行／WORKLIST正常路徑活表，追加本WORKLOG與既有研究紀錄，其他活表與玩法保持。下一步有界唯讀核對生成迴圈進度與退出條件，先判斷進展或阻塞，再決定續跑預算。正式writer、生成完成／完整開局、RNG與remake同狀態未驗，主庫RE-first與整款remake／中文化目標保持。

主庫四文件、14份私有來源／收據、工具精確HEAD與官方輸入雜湊核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；兩工作區掛載篩選Docker容器空，未留背景程序。

## 2026-10-03：宇宙生成迴圈有界進度與兩個正常返回

- 起點主庫33fafba87b14e1bcec47a89df79dc33641bdc6a0／工具d5127adc64a04af79796d933aec73413bbcfd824。上一輪有實際CPU與原續行進展。本輪命中dosgolem、正常玩家路徑、規格閘門、文件職責與逆向回填；345先DRAFT，只蒐證正常原迴圈與返回，不解完整helper。
- private首次單檔build缺IRQ原型，原版未啟動；沿341三檔入口修正。外層300s／450s兩次逾時，第二次持續保存raw LOG，495原步／8444可比列與60partial PNG保存，不當CPU或玩法失敗。原EBP+4候選是2BDA74，原PUSH／ENTER框架顯示返回應在+12；錯誤候選及未捕捉RET的負結果保留。
- 65,818,624-byte RAM的12次SHA-256實測1.029894684s，逐步2304次額外約3m17.739779328s。改為各group arm／首末步／RET前後明示抽樣、其他步ram_checked=false，cold buffers移至state。CPU／input／120M cap不改，相同300s容器乾淨重跑成功，不以提高遊戲步數掩蓋逾時。
- 三組各192步共576，原比較界值18／66／87，45次INC／51次CMP／51次JL／60次MOVSX／48次MOVZX數學核算PASS。114058778與117106151的原17FD1E C21400真正返回17F037，原stack37F01700、ESP2BD9F8→2BDA10、完整其他R／六段／flags246h保持。第三組120M仍pending，生成完成／完整開局未驗。
- 全部8503原344列／32PNG保持，原完整正常輸入與SETG／MOV、SETLE實際SS write、TEST三步保持。兩份private／正式三區塊逆轉後保持344；READY後正式source逐byte等於已驗v2，不重跑相同充分原流程。CPU／平台未改，不重跑無關Go全套；上輪固定EXE全套只是既有回歸。
- 82項回填、新345的27缺證據／抽樣／pending／狀態／索引負例與其餘四份較早回填8負例、所有舊負例通過。公開source同binary的338 CLI17無效／正對照、339 CLI22無效／100M及120M正對照通過。18份私有來源／收據／核算hash與命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 工具a666ae584ba4df9468c233af2a07823229b12e09已推送github隔離分支，遠端回讀一致，未推本機origin。公開只提交有界probe／spec／索引／回填／guard，不含原素材。來源與收據1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，所有容器已清理。
- 主庫只更新CONTEXT一行／WORKLIST正常路徑活表，追加本WORKLOG與既有研究紀錄；玩法與其餘活表保持。下一步保留120M基準，先為固定160M明示診斷分支指定同狀態／有界終態規格，READY後續行。若仍同頁先查生成狀態，不盲加cap。正式writer、生成完成／完整開局、RNG與remake同狀態未驗，主庫RE-first與整款remake／中文化目標保持。

主庫四文件、18份私有來源／收據、工具精確HEAD與官方輸入雜湊核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；兩工作區掛載篩選Docker容器空，成功與逾時工作均已清理。

## 2026-10-03：固定160M續行、第三正常返回與配置母星階段

- 起點主庫8f5dbc7b3eb2b3e93f8bd7262e2ff476ea2bd660／工具a666ae584ba4df9468c233af2a07823229b12e09。命中dosgolem、正常玩家路徑、規格閘門、逆向回填與文件職責路由，沿既有入口。346先DRAFT，核對345不可變收據後READY，才實作明示固定診斷預算。
- 新旗標DOSGOLEM_MOO2_UNIVERSE_CONTINUE_160M只接受1及完整160M情境。四區塊／五guard以外source保持345，CPU／平台未改。兩側均fresh417正版根檔、官方1.31 EXE與原MOX.SET；原ZIP／patch唯讀。
- 第一次啟動腳本把CAP字串換成數字時誤改HARDWARE_ESCAPE_STEP；兩側在讀EXE前exit2，原版未啟動。原rejected raw／gzip保存，修正腳本後兩個450s有界容器各乾淨完成，不當CPU／玩法失敗。
- 同公開source重生預設120M與明示160M。預設9085原345列／32PNG保持；明示分支到120M前9028可比原列／31原frame保持，只允許cap配置header與新增觀察紀錄。新120M完整核心／FPU／VBE／clock／callback／IRQ／第三pending通過同狀態核對，RAM唯讀；DTA與每次RAM雜湊差異沿既有mtime正規化，不稱全RAM跨次相等。
- 第三group於120083995原17FD1E C21400返回17F037，ESP2BD9F8→2BDA10，其他R／六段／flags246h不變。130M／140M／150M／160M快照通過；150M仍Generating Universe...，160M人工原圖已Placing home worlds...。160M無新CPU拒絕、到固定上限17FD04／unique_sites39434；不是完整開局通過。
- 新CLI29負例與160M正對照、舊338的17／339的22負例及100M／120M正對照通過。83項公開規格回填、新346的23缺證據／狀態／回填／索引負例及全部舊負例通過；345只更新過期下一步，保留其歷史證據索引。CPU未改，未重跑無關Go全套。
- 實際命令、25份本機來源／核算雜湊及68張PNG manifest見研究紀錄。公開只含probe／spec／索引／guard，原LOG／PNG／ZIP／EXE不入Git。工具92d25f387d2ff18919ec65da2d85c7ce1566d7db已推送github隔離分支並核對遠端，未推本機origin；主庫本輪四文件隨後提交推送。
- Docker：兩個原版與檢查容器皆移除，工具掛載相關容器空；工具樹無root-owned／.md目錄。本輪新收據1000:1000，主庫歷史root-owned不修動。停止前核對HEAD與兩工作樹。
- 下一步維持160M預算，追150M至160M配置母星文字producer／caller和最小狀態，不直接提高cap或重送。正式writer、完整生成／開局、RNG、人耳與remake同狀態仍未知，主庫RE-first保持，整款remake／中文化目標活躍。

## 2026-10-03：原進度文字來源、NUL複製與配置母星呼叫端

- 起點主庫6e32fabf24833d6fa2811160d0bfcaed46420e56／工具92d25f387d2ff18919ec65da2d85c7ce1566d7db。上一輪取得160M配置母星圖與第三RET，屬進展；本輪命中dosgolem、正常玩家路徑、IDA9.4、規格閘門、逆向回填與文件職責，沿既有入口。
- 347先DRAFT。核對兩LBX的archive count1、實際header／邊界，原索引161／242與檔案offset；正式1.31與原ZIP的兩檔相同。主庫既有IDA DB輸入7ae2…不符固定4e11…，拒用；以官方1.31重建一次性DB，非空JSON／schema／5365函式／input SHA／UID1000通過。idat exit1不取代產物驗證。
- 初次可丟棄LE匯出器猜400000 RAM上限而越界，未執行CPU；IDAPython初次API模組錯誤修正為ida_loader。原腳本與失敗紀錄保留，原正式DB唯讀、不改名。
- private v1取得兩原查詢返回，但目的緩衝區原有NUL讓生成文字copy_complete提早八步。獨立ESI／EDI核算拒絕；v2必須到原POP EDI且指標跨NUL／AL0／flags246h。同160M正常輸入乾淨重跑，原索引242複製102875330完成，161複製152598741完成。
- 六事件readonly，全部10523原346列／36PNG保持。配置母星原五PUSH再ENTER、SS188:EBP＋24的8AB91600保存返回位址16B98A，定位實際caller16B985；尚未捕捉callee最終RET。160M無新CPU拒絕，仍配置母星圖，完整生成／開局未驗。
- 證據審查READY後才把兩有界區塊接到正式診斷；正式只比v2多旗標關閉defer守衛。正式關閉8M的1693原列／PNG保持、68個舊CLI負例與正對照通過。84項回填、新347的28負例、345另兩負例與全部舊負例通過；兩區塊逆轉逐byte保持346，CPU／平台不改。
- 實際命令及37份本機來源／核算雜湊見研究紀錄。未重跑120M或無關Go全套，歷史346同狀態與344固定EXE全套不冒稱本輪新跑。原EXE／LBX／LOG／PNG／私有IDAPython留忽略workplace，不入Git。
- 工具53243f6d5633380f456a9c27faedf908cbade675已推送github隔離分支並核對遠端，未推本機origin；主庫本輪四文件隨後提交推送。新來源／收據1000:1000，工具樹無root-owned／.md目錄；原版、IDA與驗證容器有界並收尾清理，主庫歷史root-owned不修動。
- 下一步維持160M預算，觀察16C78E正常返回16B98A及後續word[EBP-8]分支。欄位用途、正式writer、母星配置規則、完整生成／開局、RNG、人耳與remake同狀態未知；主庫RE-first保持，整款remake／中文化目標活躍。

## 2026-10-03：原進度函式RET與上層零分支

- 起點主庫c6b990d5246414cf895422b8df14cd30cc03a5db／工具53243f6d5633380f456a9c27faedf908cbade675。上一輪完成347並推送，屬進展。本輪命中dosgolem、IDA9.4、規格閘門、逆向回填及文件職責，沿既有入口。
- 348先DRAFT，固定原EXE與正常160M輸入重播。兩次一次性IDA DB保存原sub_7C78E邊界／103指令、共享LEAVE／五POP／C3及caller CMP／JNZ；非空JSON／schema1／固定input SHA／UID1000通過，idat exit1不當失敗，正式.i64唯讀。
- private觀察捕捉149825343 CALL、次步原入口、153214295 C3 RET與次步16B98A。原保存五暫存器恢復、ESP＋4／caller EBP恢復；SS188:2BDB98 word0000使16B98F原JNZ不跳，153214298到16B995。callee local7779是另一框架，不混用。
- 全部10530原347列／36PNG保持，11事件核心／Bus／FPU／VBE與完整RAM readonly。160M仍原配置母星圖，完整生成／開局未驗。終態返回槽16BAE3及固定IDA的16BADE→16AD13只標強推論，下一輪需實際CALL證據。
- READY證據審查後正式兩區塊只更名348標記，執行語句與private相同；來源逆轉逐byte保持347，CPU／平台不改。正式關閉8M的1693原列／PNG、68舊CLI負例與正對照通過。85項回填、新348的27缺證據負例與既有負例通過，347已回填，限定CONFORMED。
- 初次來源逆轉漏移除兩個新增空行；diff只剩空行，修正核算後通過，原執行語句未改。Docker metadata的Go User欄位不存在，改獨立核對image ID，--user仍明定UID1000；不作產品缺陷。未重跑120M或無關CPU全套。
- 命令及21份忽略來源／收據雜湊見研究紀錄。工具9afe6570dc3e49e354b9a9ec07360125682b3d67已推送github隔離分支並核對遠端，未推本機origin；主庫本輪四文件隨後提交推送。原EXE／LBX／RAM／LOG／PNG／私有IDA腳本不入Git。
- Docker原版、IDA與驗證程序皆有界並結束移除；新來源／收據1000:1000，工具樹無root-owned／.md目錄，主庫歷史root-owned不修動。下一步維持160M捕捉16BADE→16AD13正常參數與生成邊界；完整生成／開局、正式writer、RNG、人耳與remake同狀態未知，整款remake／中文化目標活躍。

## 2026-10-03：後續生成原入口與直接返回，外層仍等待

- 起點主庫ff2c377b2e3a08faba7f3d600a0e547759c4bcdc／工具9afe6570dc3e49e354b9a9ec07360125682b3d67。上一輪完成348並推送，屬進展；本輪沿dosgolem／IDA9.4／規格閘門／回填／文件職責入口。
- 349先DRAFT，由固定1.31建立三次一次性IDA DB，正式.i64唯讀。原sub_7AD13有216指令、唯一caller7BADE；直接距離callee7B0B4有28指令，第三子呼叫末尾8F052 C21800清理24byte。原名、EA、file offset、bytes與operand保持，非空JSON／schema1／input SHA／UID1000通過，idat exit1不當失敗。
- private正常160M捕捉153878499原16BADE CALL、次步16AD13入口、四PUSH／ENTER後框架。EAX=caller BP-28h、EDX147D9保持，原slot16BAE3；三直接子呼叫返回，首次距離AX1／DX3→EAX1CC4D在29原步返回。外層RET及16BAE3返回未見。
- 全部10542原348列／36PNG保持，13事件readonly，160M仍配置母星圖。初版核算猜28指令為29而拒絕，腳本／stderr保留並重現exit1；另按原C21800修正第三CALL的24byte清理，不改CPU或private執行語句。
- READY審查後正式只更名349標記，逆轉兩區塊逐byte保持348，CPU／平台不改。正式關閉8M1693原列／PNG、68舊CLI負例與正對照、86回填／新349的32負例與既有負例通過，限定CONFORMED。348已回填，並刪除尾端仍把已命中CALL列未知的殘留斷言。
- 命令、26份本機來源／收據雜湊見研究紀錄。原EXE／LBX／RAM／LOG／PNG／私有IDA腳本不入Git。工具a279ce5的診斷及07611f5的現況勘誤已推送，完整HEAD07611f5808e53eb40a61cc8a94489045b0ff0686核對遠端，未推本機origin；主庫四文件隨後提交推送。
- 原版、IDA與驗證容器有界並已結束移除，輸出1000:1000，工具樹無root-owned／.md目錄，主庫歷史root-owned不修動。未重跑120M或無關CPU全套。下一步維持160M核對外層迭代／實際上限／重繪，再決定續行預算；母星配置與完整開局、正式writer、RNG、人耳及remake同狀態未知，整款remake／中文化目標活躍。

## 2026-10-03：350 原生成首四迭代與實際比較界限

起點主庫10f8dccc8d33cd311bc45b5d4c21cd24f06a4ffa／工具07611f5808e53eb40a61cc8a94489045b0ff0686。路由命中dosgolem對拍、規格閘門與回填，沿已載入入口及固定IDA349原匯出。主庫RE-first保持，未修改玩法、CPU、平台、正常輸入或160M cap。

- 已證實：原16AD7D首四head SI1..4，CMP時2..5；原relocated CMP讀DS188:28199A raw2400／signed36，前3組JL後下一原步回head。首四次head至CMP151324、151553、166358、167587原步。
- 已證實：13事件readonly；第5個SI後full=true。首四組重繪未見，不能稱後續皆未重繪；160M外層returned=false／17FD04，無新CPU拒絕，原final畫面保持。
- 已證實：全部10556原349列／36PNG、正式關閉8M1693原列／PNG、68舊CLI負例與正對照、87項回填／新350的28負例／既有負例通過。正式source與已驗private只更名350標記，逆轉逐byte保持349；本輪不重跑120M或無關CPU全套。
- 強推論：樣本顯示有進度；四次耗時無法保證所有後段／重入的完成預算。未知：全部迭代出口／RET、後續重繪、正式writer、完整配置／開局、RNG與remake同狀態。
- 工具025629e9fe5de8e29abe81aa08c67ad37a8ff161已推送github/codex/moo2-parity-20260930，未推本機origin。18份忽略收據、原地址基準、實際命令與雜湊見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)；公開僅自製診斷／規格／索引／守衛。
- 下一步維持160M，取原16AF29出口／16AF35後段及16B01F RET／16BAE3返回邊界，依實際工作量再決定預算。remake與中文化尚未完成。

收尾核對精確HEAD與兩庫工作樹、遠端分支及專案Docker清理；既有主庫root-owned2437檔／272目錄保持，本輪不新增root-owned或.md目錄。

## 2026-10-03：351 原連續迭代與160M截斷

起點主庫ac2784ad7fcac666da2b6f2aa0779d88163d6b79／工具025629e9fe5de8e29abe81aa08c67ad37a8ff161。路由命中dosgolem對拍、規格閘門與回填，沿已載入入口與固定IDA349原匯出。主庫RE-first保持，只接有界readonly診斷，不改CPU／平台／玩法／正常輸入或160M cap。

- 已證實：26個原16AD7D head的SI1..26連續，無回跳／重複。153880145到159926951共6046806原步，間隔151326..338347；160M只再續第26次73049步。原bound signed36保持，full=false，出口／後段／RET全部未見，沒有新CPU拒絕。
- 已證實：全部10570原350列／36PNG、27事件readonly、正式來源逆轉保持350、關閉8M1693原列／PNG、68舊CLI負例及正對照、88項回填／新351的25負例／既有負例通過。沒有新IDA或120M／CPU全套重跑。
- 強推論：連續進度與耗時增長支持下一次固定180M有限續行。加20M是探索餘量，不保證原後段／重入或完整開局完成；沒有guest代寫／跳過原生成。
- 工具ffc7e16a6ed34a279418982dbe5131e328ec7162已推送github/codex/moo2-parity-20260930，未推本機origin；18份忽略收據、原地址基準與命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 未知：SI27以後、全部出口／後段／RET、正式writer、完整配置／開局、RNG與remake同狀態。下一步以固定180M核對160M同一原狀態，再追原出口／返回與正常玩家畫面；不深入runtime或繪圖helper。remake／中文化目標仍活躍。

收尾核對兩庫精確HEAD／遠端／工作樹與專案Docker清理。本輪UID1000、工具root-owned零；主庫既有2437檔／272目錄保持，沒有新增root-owned或.md目錄。

## 2026-10-03：352 原生成返回與新word記憶體AND停點

起點主庫6379a13d3121fbaaca0eb82137c08a4bc3849ecb／工具ffc7e16a6ed34a279418982dbe5131e328ec7162。路由命中dosgolem對拍、規格閘門與回填，沿已載入入口與固定IDA349匯出。主庫RE-first保持；本輪只增加明示180M診斷契約，CPU／平台／玩法與正常輸入不改。

- 已證實：原SI1..35連續，163755070比較signed36，原JL不跳並退出迴圈；後段DX72／AX34，原XOR AL後AL0。163778787原C3 RET於次步回16BAE3，caller比較後原JZ到16BB00。只確認此callee返回，不宣稱完整生成或開局。
- 已證實：163795435於原103BF9的66 81 63 0C 7F FE拒絕，屬word記憶體AND尚未支援。EBX5AA044／DS188，目標offset5AA050／immFE7F；拒絕後103BFC是解碼停點。probe exit0並非180M cap或完成，final觀察在163795436。
- 已證實：160M前10532共通列／36PNG、同一原CPU／FPU／輸入等狀態、readonly checkpoint及46事件保持；正式來源逆轉保持351。關閉8M1693列／PNG、68舊CLI＋32新CLI負例、89項回填／新352的30負例／較早另4負例通過。沒有新IDA、120M或CPU全套重跑。
- 初次核算腳本的IRQ迴圈覆寫共通列變數，10532列斷言拒絕；保留腳本／輸出並重現exit1。只改變數名，嚴格斷言、原收據與CPU不改，乾淨重播全部通過。分類為驗證腳本問題。
- 工具8ef52b1fc372bf96267d964f8b1ba903594d95c6已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin。349／350／351現況已回填，352限定CONFORMED。24份忽略收據、命令與雜湊見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 未知：原word輸入／寫回、正式writer、資料語意、完整配置／開局、RNG與remake同狀態。下一步建立word記憶體AND窄CPU規格並以相同180M正常路徑驗證。remake／中文化目標仍活躍。

本輪輸出1000:1000、工具root-owned零；主庫既有2437檔／272目錄不修動，沒有新增root-owned或.md目錄。收尾核對兩庫精確HEAD／遠端／工作樹與專案Docker清理。

## 2026-10-03：353 接通word記憶體AND與原三步消費

起點主庫ed55a4b87cbd31ab8d2bf55cb5dbcc013b20ed47／工具8ef52b1fc372bf96267d964f8b1ba903594d95c6。上一輪已推送原35迭代與RET，屬實際進展。本輪命中dosgolem對拍、CPU規格閘門、回填與文件職責，沿已載入技能／入口；依Intel 80386原廠AND及第3.4.1節契約。主庫RE-first保持，不修改玩法。

- DRAFT未改CPU正常180M量原DS188:5AA050 word0000，全部10793原352列／36PNG及readonly／RAM保持。原來源可讀、ISA充分後審查READY，才新增21行通用CPU分支與三個observer；沒有位址特例或guest代寫。
- 已證實：原163795435..163795437的AND word0000／SHL EDX0／OR dword0均返回nil，EIP依序103BFF／103C02／103C05、flags246h，完整R／段／來源相鄰／stack及RAM保持。原零→零不證明Bus寫次數，受控CPU成功零／非零兩bytes與逐byte拒絕另驗；AF與多位SHL OF只驗工具模型。
- 已證實：入口前10723原共通列／36PNG、source逆轉、窄測0.312s、固定原EXE乾淨Go全套CPU38677.939s／machine1.449s、關閉8M1693列／PNG、68舊＋32新CLI負例、90項回填／新353的33負例及較早負例通過。8088實機語料缺檔，不算386硬體驗收；沒有新IDA或120M整流程重跑。
- 原164321317在223E93的86 06 memory byte XCHG拒絕，after223E95只解碼，未達180M。原ESI2BD8A8／ALFF、EDI2BD97A已記錄；目的byte與STOSB初態待下一窄切片捕捉。完整生成／開局、正式writer、RNG與remake同狀態未驗。
- 初版formal核算將四項stop診斷算進原AND前，10727對10723拒絕。保留腳本／輸出並重現exit1；按實際stop切點與四項精確類型修正後嚴格核算通過，未改原收據／CPU或重擲結果。屬驗證腳本邊界問題。
- 工具c7086292bf476be63a406131632c9cf8c4780ab6已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin；353限定CONFORMED、352未知已回填。30份本機收據與實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。下一步原byte memory XCHG／STOSB最小契約，維持180M。

本輪來源1000:1000、工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持。原版、全套與回歸容器均有界結束移除；收尾核對兩庫精確HEAD／遠端與工作樹。remake／中文化目標仍活躍。

## 2026-10-03：354 接通byte記憶體XCHG與原STOSB寫回

起點主庫fc4c448a1c2b53d2eb990db1d45f44166b884ce6／工具c7086292bf476be63a406131632c9cf8c4780ab6。上一輪完成word AND並推送，屬實際進展。本輪命中dosgolem對拍、CPU規格閘門、回填與文件職責，沿已載入技能／入口與Intel 80386原廠XCHG契約。主庫RE-first保持，不改玩法。

- DRAFT未改CPU取原DS188:2BD8A8 byte0E／ALFF與ES188:2BD97A byteFF／DF0，全部10812原353列／36PNG及readonly／RAM保持。資料可讀與ISA充分後審查READY，才用16行通用memory交換替換1行拒絕與三個observer；原register路徑、平台不改。
- 已證實：原164321317 DS0E→FF／ALFF→0E、EIP223E95；164321318原STOSB ESFF→0E／EDI增1／EIP223E96，flags202h與六段保持。兩步完整RAM差異各限一byte，相鄰資料保持，非零原寫回與後續消費已驗。未驗硬體lock波形／多CPU仲裁，顯式F0仍拒絕。
- 已證實：原入口前10742共通正常列／35frames保持；獨立窄測0.710s、固定原EXE乾淨Go全套CPU386150.111s／machine1.864s、關閉8M1693列／PNG、68舊＋32新CLI負例、91項回填／新354的34負例／352另2與較早負例通過。缺8088語料不算386實機驗收。沒有新IDA、120M整流程或失敗後挑選重跑。
- 原164560803在1CDD0F的02 45 F8 ADD AL,SS:[EBP-8]拒絕，after1CDD11只解碼，來源byte未知。較晚finalPNG已變，人工確認640×480主要黑底與小型方形圖形，未見完整地圖，不算正常開局；尚未達180M。
- 工具1f155175b2c77e6ee133ef609f43b758d7e0eed8已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin；354限定CONFORMED、352／353未知已回填。27份本機來源／收據、原地址基準與命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。
- 下一步原02 /r byte ADD記憶體來源，先取SS188:2BDB3C／AL0及後續原消費，沿相同180M。正式writer、資料語意、完整生成／開局、RNG與remake同狀態未知，remake／中文化目標仍活躍。

本輪來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持。原版、全套與回歸容器均有界結束移除；收尾核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-03：355 接通byte記憶體來源ADD與原五步零值消費

起點主庫2031d07b7890ede231f21e8ff0e5e956116a3d7a／工具1f155175b2c77e6ee133ef609f43b758d7e0eed8。上一輪XCHG／STOSB與推送完成，屬實際進展。本輪命中dosgolem對拍、CPU規格閘門、證據回填與文件職責，沿已載入逆向技能與Intel 80386原廠ADD契約，主庫RE-first保持，不改玩法。

- DRAFT未改CPU正常180M取四SS來源byte00／AL00與DS目的00，全部10829原354列／36PNG／readonly與RAM保持。ISA與原資料充分後READY，再以16行通用02 memory分支替換拒絕；三observer逆轉逐byte保持354，不改22 memory或平台。
- 原164560803..164560807四memory來源ADD及第五目的ADD零結果、flags202h→246h／EIP1CDD1E／RAM與相鄰窗口保持已驗；沒有原非零加法／進位或Bus寫次數trace，不把00→00當非零寫回。八byte register全部配對及六flags、地址別名、唯讀無寫、拒絕不發布與非零FPU另以獨立oracle驗證。
- 正式10759共通正常列／35frames、窄測4.089s、乾淨固定原EXE Go全套CPU386123.004s／machine2.959s、關閉8M1693列／PNG、68舊＋32新CLI負例與正對照、92項回填／新355的32＋4負例與較早負例通過。缺8088語料不算386實機驗收。
- 8M首次建置誤納並行formal的暫存Go main而拒絕，未跑原版；保存attempt1後改明列來源，同image／同8M與CLI乾淨重跑通過。屬共享暫存檔／驗證腳本問題，CPU全套與正式原版收據未挑選替換。
- 原164561579在1CE387的0F94 memory目的拒絕，after1CE38A只解碼、SS188:2BD834 byte未知，flags246h。尚未達180M。finalPNG與354逐byte相同，主要黑底、未見完整地圖，不算完整開局。
- 工具8273d887f5387c23d9ae13056ae2ad1e263aeee0已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin；355限定CONFORMED、354／353／352目前未知與索引已回填。29份本機來源／收據與實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。下一步取原0F94目的byte與正常消費，不增加cap或深入helper。

正式writer、資料語意、完整生成／開局、RNG與remake同狀態仍未知；remake／中文化目標活躍。本輪來源／收據1000:1000，工具root-owned／.md目錄零；主庫既有2437檔／272目錄保持。原版、全套與回歸容器有界結束移除，交接核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-03：356 接通標準SETcc記憶體目的與原SETE寫回

起點主庫ba3800608f7e55ad0a293104a57df9b41dc78669／工具8273d887f5387c23d9ae13056ae2ad1e263aeee0。上一輪ADD與推送完成，屬實際進展。本輪路由命中dosgolem對拍、CPU規格閘門、回填與文件職責，沿已載入逆向技能與Intel SETcc契約。主庫RE-first保持，不改玩法。

- DRAFT未改CPU正常180M取SS188:2BD834 byte41／相鄰004100、flags246h，全部10834原355列／36PNG與readonly／RAM保持。byte與ISA充分後READY，CPU移除3行早拒絕並加11行通用純write，原條件表／register／Jcc保持；兩舊memory負例明確限定未知selector，完整新正例補有效memory。
- 原164561579 SS41→01／唯一下標2BD834變更、EIP1CE38B、flags246h保持與164561580下一JMP1CE61E已驗；第三CMP只觀測，dword來源與byte1 reader未驗，不宣稱原三步規則全對拍。
- 4194304個memory真值／初byte組合、全地址形狀／純write／相同值仍寫／拒絕保持／非零FPU及舊register、Jcc通過。窄測7.994s，固定原EXE乾淨Go全套CPU386137.976s／machine1.755s，8M1693列／PNG，68舊＋32新CLI負例及正對照，93項回填／新356的37＋10負例與較早負例通過；缺8088語料不算386實機驗收。
- 原164567987在1CF90A的66 6B word IMUL拒絕，after1CF90C只解碼，DS188:5A2084來源word未知、flags206h；尚未達180M。finalPNG保持355主要黑底與小型方形圖形，未見完整地圖，不算完整開局。
- 初態兩步probe與原run內容已保存，READY後正式診斷三步，舊LOG／PNG不重寫；原命令內容／新重生入口分列。沒有CPU／正式原版／驗證失敗後挑選收據；明列來源建置避免並行暫存main。
- 工具442eef487fa03be9ef0f793e233396120c56973b已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin；356限定CONFORMED、六份較早unknown與索引已回填。29份本機來源／收據及實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。下一步取word IMUL來源／DI／imm05及正常消費，沿同180M，不增加cap或深挖helper。

第三CMP數值／byte1 reader、資料語意、正式writer、完整生成／開局、RNG與remake同狀態未知，remake／中文化目標活躍。本輪來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；原版、全套與回歸容器有界結束移除，交接核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-04：357 接通word立即值IMUL並驗原零積與兩MOV

起點主庫c7bd82f131c9afa5152466ce58ff1ea35b7e4b40／工具442eef487fa03be9ef0f793e233396120c56973b。上一輪SETcc與推送完成。本輪路由命中dosgolem對拍、CPU規格閘門、回填與文件職責，沿既有逆向技能與Intel IMUL契約；主庫RE-first保持，不改玩法。

- 2026-10-03的DRAFT未改CPU沿同180M，取DS188:5A2084 word0000與SS frame二十bytes，全部10837原356列／36PNG、只讀與RAM保持。來源／ISA充分後READY；新增16位立即值IMUL，既有32位與其他乘法保持。
- 原164567987 word0000×5=0，EDI005AA5F4→005A0000／高word005A保持、CF／OF0；後面兩MOV dword0→0、EIP1CF91D與RAM保持已驗。兩MOV不消費DI，未定義旗標保存只屬工具近似；原非零／overflow與正式DI reader未驗。
- 33554432個imm8案例、2097152個imm16案例、全地址形狀／別名／nonzero FPU與拒絕保持通過。第一次截短測試誤用CS描述符，取指直接讀Bus；改成第一個缺byte Bus失敗後同命令4.372s通過，保留7.631s失敗輸出。READY初稿高word分組誤寫已更正005A，獨立byte視圖與正式原輸出一致；CPU不受文字錯誤影響。
- 固定原EXE乾淨Go全套CPU386121.065s／machine1.556s；正式10767正常前綴／35frames、8M1693列／PNG、68舊＋32新CLI負例／正對照、94項回填／新357的38＋16負例及較早負例通過。首次守衛缺完整164567989步號，補文件縮寫後重跑，摘要保留；CPU與原收據不改。缺8088語料不算386實機驗收。
- 原164568139在1CFD3F的66 F7 /3 word memory NEG拒絕，after1CFD42只解碼／未取disp8或source，DS188:5A207C word未知，尚未達180M。finalPNG逐byte保持356，沿354已檢視主要黑底與小型方形圖形，未見完整地圖。
- 工具a7f175f3bfadcc6d9d56a78e23c0eabf2cbe2442已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin。357限定CONFORMED、九份較早入口與現行unknown、索引／守衛已回填；29份本機來源／收據及命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。CPU與probe逆轉保持356，六舊測試不變。

下一步擷取原1CFD3F word NEG來源與自然分支／消費，沿同180M，不增加cap或深入helper。正式DI reader、資料語意、正式writer、完整生成／開局、RNG與remake同狀態未知，remake／中文化目標持續。本輪來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；各本輪容器有界結束並移除，收尾核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-04：358 接通word NEG與原零值／自然分支

起點主庫3b9ec0266baac64d5d6cc80ee4405afd73c4d6e9／工具a7f175f3bfadcc6d9d56a78e23c0eabf2cbe2442。上一輪IMUL與推送完成，屬實際進展。路由命中dosgolem對拍、CPU規格閘門、回填與文件職責；沿既有逆向技能，主庫RE-first保持，不改玩法。

- DRAFT未改CPU沿同180M，取DS188:5A207C word0000與相鄰十二bytes、SS二十bytes，全部10840原357列／36PNG／readonly與RAM保持。原來源／Intel契約充分才READY；新增24行word NEG register／memory，六算術flags完整定義。
- 原164568139 word0000→0000／flags246h與EIP1CFD43，164568140下一EB09自然到1CFD4E已驗。第三CMP來源SS:[EBP-564]未取，數值／NEG word reader未驗；第四JE按觀測ZF0不跳到1CFD5B，flags206h保持。四步R／段／RAM與source／frame保持、callback12／12、IRQ41948／41948非活動。原同值寫Bus次數未取；非零NEG與8000溢位由工程測試覆蓋，不升格原動態。
- 4194304個memory×flags、1048576個register×來源×flags案例、全地址／高word／nonzero FPU、每byte失敗與晚期部分寫不發布flags、prefix／截短通過。325舊66負例限定未知selector，268舊word NEG register負例加segment prefix，新全值域正例接合法word；兩修改與CPU／probe可逆轉回357，七舊測試不改。
- 窄測6.224s、固定原EXE乾淨Go全套CPU38666.221s／machine1.331s，正式10770正常前綴／35frames、8M1693列／PNG、68舊＋32新CLI負例／正對照、95項回填／新358的41＋20負例與較早負例通過。缺8088語料不算386硬體驗收，沒有CPU／原版／驗證失敗後挑收據。
- 原164610300在1D0944的66 29 word SUB memory目的拒絕，after1D0946只解碼／未取ModRM或source，DS188:5AA6D1目的word未知／來源AX0；未達180M。finalPNG逐byte保持357，主要黑底與小型方形圖形，未見完整地圖。
- 工具a995e5d62249aef97f73cc52e11176acc6e3218b已推送github/codex/moo2-parity-20260930並核對遠端，未推本機origin。358限定CONFORMED、十一較早入口與現行unknown／索引／守衛已回填。27份本機來源／收據與實際命令見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。

下一步擷取1D0944目的word／AX來源、正常POP／RET消費並審查word SUB，沿同180M，不增加cap或深入helper。原非零NEG／溢位、CMP數值／NEG word reader、正式writer、資料語意、完整生成／開局、RNG與remake同狀態未知，remake／中文化目標持續。本輪來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；本輪容器有界結束並移除，交接核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-04：359限定word SUB與原三POP／RET

- 工具11d9aad0d10bcf51ac75f9ee23611acfe2fd7e9b已推送github的codex/moo2-parity-20260930，精確遠端HEAD相同。只接cpu386 66 29 /r word目的，既有ADD／其他SUB與十一舊測試逐byte保持；主庫Go／Ebitengine玩法不改。
- 原164610300的DS188:5AA6D1 word0003-AX0與六flags206h、164610301–164610304三POP／RET真正SS槽／EIP1D1E0B／ESP2BDB5C已驗，五步RAM與固定窗口保持；POP／RET不是目的word reader，原同值Bus寫次數未取。
- DRAFT10845原列／36PNG、正式10775正常列／35frames、窄測0.685s／固定原EXE乾淨Go全套CPU38682.239s／machine3.104s、關閉8M1693列／PNG、CLI與96守衛／新41＋22負例通過。三項測試／環境失敗已修正後乾淨重跑，原正式guest結果未重跑挑選。
- 原164984957於1D2A33拒絕66 99 CWD，AX0001／DX0000，after1D2A35／flags246h，未達180M。下一步取CWD與下一SUB／shift，維持輸入和cap；原非零SUB來源／借位／溢位、目的word reader、正式writer／RNG／完整開局與remake同狀態未知。
- 四文件更新與27份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。工具樹乾淨、專案容器0；既有root-owned檔案2437／目錄272維持，沒有新增.md目錄。


## 2026-10-04：360限定word CWD與原SUB／SAR

- 工具a16e94c4c15e5c78760cc670cb4f6766457a5f67已推送github的codex/moo2-parity-20260930，遠端精確HEAD相同。只接cpu386 66 99 word CWD，裸CDQ／word SUB與SAR／flags helper及十三舊測試保持，主庫玩法不改。
- 原164984957 AX1／DX0與完整flags246h、164984958 SUB讀DX得AX1／六flags202h、164984959 SAR讀AX1→0與五定義flags／EIP1D2A3B已驗。三步RAM／FPU原bits保持，SAR AF不列原版parity。
- DRAFT10859列／36PNG、正式10789正常列／35frames、16777216工程全值域、高word與64flags、窄測0.828s／固定原EXE乾淨Go全套CPU38699.385s／machine1.679s、前輪359原8M收據／CLI、97守衛及新42＋24負例通過。
- 新原168496272在2376CB拒絕memory dword ROR，DS188:270FC4來源未知／imm08，after2376CD未取disp／imm或source，未達180M。下一步取原dword與下一A1真正load，維持輸入和cap，不猜目的。
- 四文件與27份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。工具樹乾淨／專案容器0；原負AX／EDX高word、新ROR、正式writer／RNG／完整開局與remake同狀態未知；主庫RE-first保持。

## 2026-10-04：361限定memory ROR與原A1消費

- 工具5a2170cfc18b40e890a21dcdcbb10605f084b57d已推送github隔離分支，遠端精確HEAD相同。只加C1 /1 memory dword ROR，既有register分支與flags helper保持，297三舊負例明示未知DS；主庫玩法不改。
- 原DS188:270FC4的000B1818 ROR8→18000B18／CF0／EIP2376D2，下一A1真正load到EAX18000B18／EIP2376D7已驗。RAM只差270FC5／270FC6／270FC7，其他狀態保持；多位OF只驗工具模型。
- DRAFT11019列／36PNG、正式10949正常列／35frames、1212416工程矩陣、窄測1.025s／固定原EXE乾淨Go全套CPU38660.001s／machine1.824s、前輪360原8M收據／CLI、98守衛及新39＋28負例通過。核對腳本空陣列問題修正後讀同收據，guest未重跑。
- 原版自然到180M無CPU拒絕，終圖顯示旗色選單的Error saving game／Permission denied。下一步取失敗DOS呼叫／檔名／mode／errno，再審查既有隔離覆蓋層，維持cap與唯讀原版來源；完整開局仍未驗。
- 四文件與27份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。工具樹乾淨／專案容器0；正式writer／RNG與remake同狀態未知，主庫RE-first保持。

## 2026-10-04：362定位原SAVE10唯讀開寫拒絕

- 工具4f9c45be2017904ea42d86ef9b7ae692388eee08已推送github隔離分支，遠端精確HEAD相同。公開只加兩個有界唯讀DOS診斷區塊，全部internal／CPU／DOS／provider及舊測試保持361。
- 原165025480／dosgolem_high_le:237024的3D01／DS188:2BDB68／SAVE10.GAM回AX5／CF1、flags202h→203h與RAM保持，已定位provider缺WriteFileProvider；不稱AH40寫入失敗。
- 128診斷、14498正規化原列／38PNG、99守衛及新29＋32負例通過。mtime直比、NUL長度與逆轉空行屬驗證問題，修正後讀同guest收據／乾淨重跑測試，未挑選原版結果。
- 可寫試作窄測0.053s／固定原EXEGo全套CPU386130.397s／machine1.643s通過，但原版在80M完整表guard拒絕，未送ACCEPT／未到存檔，沒有state寫入收據。五個+44四byte窗口各增8000h，RGB相同且前段執行已改變，所指內容未知。試作退回本機DRAFT，公開維持唯讀，原失敗收據保留。
- 下一步取初段開檔與80M五窗口候選位址的內容，不改guard或點擊時刻。36份本機來源／收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。工具樹乾淨／專案容器0；正式存檔／完整開局／RNG與remake同狀態未驗，主庫RE-first保持。

## 2026-10-04：363覆蓋層前段開檔與五窗口

已證實原1192795／dosgolem_high_le:237024的SOUND.LBX 3D02，兩側同原R／段／flags與RAM，唯讀拒絕AX5 CF1、overlay真handle5 CF0。80M record11–15的+44候選值各增8000h，但原descriptor所指前128bytes逐byte相同；完整物件／角色／消費未知。額外32KiB配置造成位移為強推論，沒有追allocator。親看dosgolem原overlay 80M設定頁，ACCEPT完整且原熱區包含480,400。

Docker內執行 `bash workplace/new-game-363-pair-run.sh`、`python3 workplace/new-game-363-pair-verify.py`、`python3 workplace/new-game-363-state-source-verify.py`、`python3 workplace/new-game-363-backlink-verify.py`，全部通過。各側一次guest止於80M只讀快照；6148／6271舊列依352既有正規化、各27PNG保持，八開檔／五窗口／全狀態自檢只讀。PNG數量初設28，修正成明示27集合後讀同收據，未重跑。兩側原418來源檔完全保持，state只sound.lbx且與原ZIP4250888bytes逐byte相同。100項回填及新37＋34缺證據負例通過，沒有新的Go／CPU／DOS／provider／公開probe變更。

隔離工具04a96f09538a6b01907685f6f56a8d7de154cde1已推送github；本輪主庫只改DOS活表／CONTEXT單行並追加工作與研究紀錄。原版PNG／LOG／資產窗口與sound副本留忽略workplace，不提交。主庫起始HEAD088df139965dfc699b56efb4b577d87eae4e0f77；精確私有收據見研究入口。下一步是獨立可寫profile的ACCEPT規格審查，原336唯讀guard保持。正式存檔／完整開局／RNG／remake同狀態未驗，主庫RE-first保持。

## 2026-10-04：364可寫ACCEPT與90M選族頁

沿363原收據建立獨立DRAFT工具profile，實際原表／五窗口／CPU／FPU／callback／IRQ及唯一熱區審查後READY。隔離試作只送一次正常ACCEPT，press80000000／133462143µs、release80013765／133507499µs，差45356µs且回呼5／5完成。原80119128在dosgolem_high_le:20DDDB實際66A3A6C42600寫word0000→0F00，R／段／flags保持，callback6／6，IRQ17934／17934保持；90M原SELECT RACE圖片親看確認，與舊原選族頁PNG逐byte相同。未送race，未驗存檔或完整開局。

Docker命令 `python3 workplace/new-game-364-ready-review.py`、`bash workplace/new-game-364-run.sh`、`python3 workplace/new-game-364-verify.py` 全部通過。6270原共通列依352正規化、27PNG保持，原precondition所有欄位保持，原336唯讀guard不改，新profile核對實際完整表／五窗口。四CLI拒絕與合法參數正對照通過；原418檔前後SHA-256保持，state仍僅與原ZIP相同的sound.lbx。產生腳本括號筆誤在guest前修正，原guest未重跑；16byte觀察窗口只取該指令六byte核對。公開CPU／DOS／provider／probe與所有internal不變，三私有區塊可逆。

工具98950c28961c0f89ed63304cd131477d083c85b7已推送github；起始主庫f731b75b01861a08e2e2b0df03ec86009a145a7b。主庫本輪只更新DOS活表／CONTEXT單行與追加歷程及研究，11份私有收據雜湊見研究入口，原LOG／PNG／資產不入Git。原版執行容器已自動移除，兩個專案掛載filter均無執行中或停止容器。下一步依新90M完整880byte表審查可寫Humans輸入。主庫RE-first保持，正式存檔／完整開局／RNG／remake同狀態未驗。

## 2026-10-04：365正常Humans與366名稱readiness

365依364原90M真實880byte表／唯一Humans幾何審查後READY。原正常press90000000／162623172µs、release90008107／162662935µs，差39763µs，callback7／7後首次放開。原90066074於dosgolem_high_le:20DDDB寫word0000→0700，R／段／flags保持、callback8／8、IRQ20889／20889非活動且非failed，95M原Strader名稱頁已親看並與338原PNG相同。6956舊列／28PNG與原來源保持，原337唯讀guard不改。

366只讀追95M flags12h／IF關閉與mask1：每原指令重算平台安全條件，95000065／234A49／flags216h自然恢復readiness。globals／header／165bytes原表／32byte候選與RGB保持，CPU／FPU／RAM只讀；64筆邊界樣本有界，offset64未列，不偽造逐步完整欄位。停止收據與受審查的first-stop程式共同核對，同guestdefer終態另取mask1／callback8／8，終點精確IRQ計數未直接保存。7662舊列／29PNG保持，未送名稱／旗色輸入。初次CLI核對把ready的read子字串誤判為讀檔，原exit2正確，確認尚未啟動guest後修正腳本；原guest只執行一次。

Docker實際命令 `python3 workplace/new-game-365-ready-review.py`、`bash workplace/new-game-365-run.sh`、`python3 workplace/new-game-365-verify.py`、`bash workplace/new-game-366-run.sh`、`python3 workplace/new-game-366-verify.py` 全部通過，各項四CLI拒絕與合法值正對照通過。原418檔SHA-256保持，state仍僅與原ZIP相同的sound.lbx，沒有SAVE10.GAM。公開CPU／DOS／provider／probe及全部internal不變，私有修改可逆。

工具bdac0e0與94b15847606ff4c90635ba90d2c38388a86a2535已推送github隔離分支；主庫起始aa44de78f60bfbcf4ac9dc61e35a3e37305e5301。本輪主庫只更新DOS活表／CONTEXT單行並追加工作與研究，23份本機收據雜湊見研究入口，原LOG／PNG／資產不提交。原版容器已自動移除，兩個專案掛載filter均無執行中或停止容器。下一步依95000065真實初態另立名稱確認契約；持久名稱／正式存檔／完整開局／RNG與remake同狀態未驗，主庫RE-first保持。

## 2026-10-04：367可寫名稱正常確認與旗色頁

沿366首個自然ready審查獨立profile後READY。原95000065正常ACCEPT，press／release相差45384µs，回呼9／9完成；原95008897、dosgolem_high_le:20DDDB六byte66A3A6C42600寫DS188:26C4A6 word0000→0100，R／段／flags與Strader候選保持。99M原SELECT BANNER COLOR已親看，完整550byte表／CPU／FPU／RAM只讀，callback10／10、IRQ23503／23503；未送旗色。原338與339唯讀guard保持。

首次生成器匹配defer與較早if分支，診斷錯置50M，尚未送名稱輸入；原32份收據保留。改以唯一執行期條件並核對99M及550byte取樣，同Docker／參數乾淨重跑，名稱輸入只送一次，不挑結果。驗證初次把95M後一筆自然INT33h併入舊365停止前綴，核對共通快照後修正比較終點，未重跑guest。7662原共通列／29PNG保持，366首個ready初態保持，九私有替換可逆，公開CPU／DOS／provider／probe及全部internal保持。

Docker命令 `python3 workplace/new-game-367-ready-review.py`、`bash workplace/new-game-367-run.sh`、`python3 workplace/new-game-367-verify.py` 通過。四CLI無效值／缺依賴及合法缺EXE正對照通過，原418檔前後SHA-256保持，state仍僅與原ZIP相同的sound.lbx，沒有SAVE10.GAM。14份新私有收據與首次32檔索引見研究入口，原LOG／PNG／RAM與版權素材不入Git。

工具a649d0b9d035eec8d8f57235b7e54a30cd7848bd已推送github隔離分支，主庫起始f32ec3e01d7be09b9f6c1e50212607e64acb07eb。本輪主庫只改DOS活表／CONTEXT單行並追加歷程／研究。原版執行容器已自動移除，專案掛載filter無執行中或停止容器；既有root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。下一步以99M真實旗色初態另立正常紅旗契約，保留339guard。名稱持久writer／正式存檔／旗色選取／完整開局／RNG與remake同狀態未驗，主庫RE-first保持。

## 2026-10-04：368可寫旗色正常輸入、原成功寫檔與新20D0拒絕

依367真實99M完整550byte表／CPU／FPU／VBE／clock／callback／IRQ審查後READY，保留339唯讀guard與339–341原poll／兩RET／release。原99103163正常pressed查詢、99103343及99103698兩GUI返回AX1，99103699首次合法放開，持按329080µs。原160M Placing home worlds已親看；原SAVE10.GAM 3D01返回handle9／CF0，truncate／完整208000bytes寫入／正常close；另兩次MOX.SET553bytes正常寫入與close。同guestDocker exec有界只讀state監測，副本與終態SHA-256／bytes一致，原418來源保持。共享旗色store未命中，持久語意未知。

原165113094在dosgolem_high_le:169E49 bytes20 D0拒絕opcode20，完整核心與下一bytes已取；CPU未改，終圖黑底游標已親看。exit0包含真正guest_cpu_stop／step_error，不稱180M或完整開局。7846舊列與30PNG保持、四CLI拒絕及正對照、三私有變更逆轉與公開全部internal／probe保持通過。初次READY審查錯拼診斷名稱，在guest前依原source修正，原guest一次沒有重擲。

實際Docker命令 `python3 workplace/new-game-368-ready-review.py`、`bash workplace/new-game-368-run.sh`、`python3 workplace/new-game-368-verify.py` 通過；同步捕捉state的精確monitor命令等價保存為私有new-game-368-state-capture.py，不公開原state／LOG／PNG／RAM與素材。18份收據雜湊見研究入口。工具62cd4911727f17042cb5f8ce98e10b0fe80331ac已推送github隔離分支，主庫起點2e2562f31d9541bf62af63a667af4b47ed3302a4。

本輪主庫只更新DOS活表／CONTEXT單行，歷程與研究追加。原版與有界monitor均已terminal，容器自動移除，兩個專案掛載filter無執行中或停止容器；既有root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。下一步為20D0 byte AND建立READY契約後補CPU及原consumer。主庫玩法RE-first保持，正式讀檔／旗色持久語意／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：369 byte AND工具能力與原版180M續行

原368的169E49／20D0阻塞經Intel80386契約及原核心審查後READY，新增裸20 /r分支與獨立窄測。原165113094正常AND AL00,DL01=00，下一POP ECX與RET依同guest真實stack通過；五定義flags已證實，AF清0僅工具模型。完整核心／FPU／RAM保持，三有界observer可逆；其他公開internal／probe不變。窄測1.011s、固定官方EXE Go全套CPU38660.343s／machine1.861s通過。

相同正常輸入一次重跑真正達180M／EIP235AA3，無guest_cpu_stop／step_error／dos_exit。終圖親看星圖上的Enter Home Star Name與Sol候選，尚未正常確認。368拒絕前11017共通列及37PNG保持，本輪39PNG；原418來源保持，state監測副本與368終態一致，四CLI拒絕與正對照通過。首次驗證把只讀observer誤當DOS3F服務也不改RAM，修正為保留服務改RAM關係；另將比較窗口止於舊拒絕專屬四terminal列前。僅重讀同收據，guest未重跑。

實際Docker命令 `python3 workplace/new-game-369-ready-review.py`、`bash workplace/new-game-369-full-run.sh`、`bash workplace/new-game-369-run.sh`、`python3 workplace/new-game-369-verify.py` 通過；有界monitor與原guest由同run擁有與清理。工具eca6a803aba5176b87f27c5defd1146ff121c226已推送github隔離分支，主庫起點283b8c63b195938220eda3a91a2cb8a1e6bf12ca。公開自製CPU／測試／spec／索引／五份回填，23份私有收據雜湊見研究入口；原LOG／PNG／RAM／state不入Git。

主庫只改DOS活表及CONTEXT單行，歷程與研究追加。兩個掛載filter無執行中或停止容器；既有root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。下一步核對原180M實際165byte命名表、候選字串與返回端，先READY再接正常ACCEPT。主庫玩法RE-first保持，正式讀檔／持久語意／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：370母星Sol候選與兩正常時點只讀觀察

沿369實際170M／180M原表審查後READY，只補兩個只讀snapshot，不送新輸入或改CPU／cap。原DS188:298848 index2+24指向28439D，32byte為Sol補零，兩時點保持。index1+24→261AC2的16byte首byte00且含BUFFER0，否定直接ACCEPT標籤推定，label_*只供導覽。170M正常命名畫面已親看，完整核心／FPU／clock／VBE／RGB／code／stack與target8:2136D1、mask2B／callback12／12、IRQ44492／44492已取。

369全14282共通列按既有正規化保持，39frames及final PNG逐byte保持，三取樣patch可逆且公開internal／probe保持。原guest一次真180M／EIP235AA3，無CPU停止；state最終副本與369終態一致、原418來源保持。初次驗證誤要求170M／180M globals相同，原26C4C6自然02→01，兩時點各與369同原時點一致；修正該跨時點假設後重讀同收據通過，未重跑guest。

實際Docker入口 `python3 workplace/new-game-370-ready-review.py`、`bash workplace/new-game-370-run.sh`、`python3 workplace/new-game-370-verify.py` 通過。工具46ae96f788d4242da161d833add26cdedd9f4c32已推送github隔離分支；主庫起點b1ce76439da49ac05cf120108be0bd04cd48bac0。19份私有收據雜湊見研究入口，原LOG／PNG／RAM／state不入Git。公開370規格／索引及369回填，主庫只更新DOS活表及CONTEXT單行，歷程與研究追加。

原guest與監測均terminal，兩專案掛載filter無殘留容器，root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。下一步依已取170M完整前置審查正常母星ACCEPT／release READY契約，保持180M，不代寫核心／RAM。主庫玩法RE-first保持，正式名稱及旗色持久writer、讀檔、完整開局與remake同狀態未驗。

## 2026-10-04：371正常母星名稱確認與新F6EC拒絕

371沿370真實170M完整前置審查後READY，只新增一次正常母星ACCEPT。原170015047 pressed查詢讀到BX1／CX550／DX260，170015082首安全release，持按47435µs。170024419原20DDDB／66A3A6C42600實寫DS188:26C4A6 word0000→0100，Sol候選及核心／FPU保持；三筆正常RET按真實stack核算，AX0不寫成AX1。命名視窗消失、星圖顯示Sol／3500.0已親看；共享store不等於正式名稱writer。

後續174213914在dosgolem_high_le:1749C0 bytesF6EC拒絕有號byte IMUL。原完整R／段／flags246h／FPU已取，CPU未改；exit0有guest_cpu_stop／step_error，未達180M。終態count23超過既有觀察器count≤16，完整表未取，這是觀察界限，不是產品資料失敗。

370點擊前11435共通原列與38PNG保持，六私有patch逆轉、公開internal／CPU／DOS／probe保持，六CLI拒絕及兩正對照通過。原guest一次，原418來源保持，state副本／終態與370一致、UID GID1000。實際Docker入口 `python3 workplace/new-game-371-ready-review.py`、`bash workplace/new-game-371-run.sh`、`python3 workplace/new-game-371-verify.py` 通過。公開371規格／索引與367／369／370回填，18份私有收據見研究入口；原LOG／PNG／RAM／state不入Git。

工具c39af543efa47387b1fd96f86038d08f940f42c6已推送github；主庫起點fd7d4e72fb3a7197a7ba999a0eb23f0b03d71fcb。主庫只更新DOS活表及CONTEXT單行，歷程與研究追加。原guest及有界監測均terminal，兩專案掛載filter無殘留容器，root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。下一步為F6EC建立CPU READY契約、獨立窄測及原consumer後相同輸入續跑。主庫玩法RE-first保持，正式名稱與旗色持久writer／讀檔／完整開局／seed與remake同狀態未驗。

## 2026-10-04：372 byte IMUL、原consumer與180M正常星圖

372經DRAFT／原371完整核心及Intel80386契約審查後READY，只新增工具裸F6 /5。獨立重複加法／有號範圍／little-endian oracle驗八register別名及memory來源，窄測0.212s通過。首次固定官方EXE Go全套只有289舊NEG拒絕fixture把F6E8當未知group；移除唯一過期樣本，原F6 /1、/2、/7與memory NEG護欄保持。相同容器及命令乾淨重跑CPU38657.925s／machine1.699s通過，首失敗收據保留。

原guest沿371同正常輸入只跑一次，180M／1996日期保持。174213914原1749C0 F6EC得到AXFFFF，下一A2只寫DS188:281F06 byte01→FF，原E9返回173CFF；全R／段／flags／FPU及全RAM效果有原三步只讀收據。實際達180M／2176C5／unique_sites54239，無CPU拒絕，終圖正常星圖已親看。既有dumpSetupTable取23物件1265bytes，未放寬舊count≤16 guard。原完整開局、正式存讀語意與remake同狀態仍未驗。

371拒絕前11645共通原列與38PNG保持，原418來源／state保持；本輪39PNG。CPU新增分支及三私有區段可逆，289測試單項更新，其餘公開internal／DOS／probe保持；六CLI拒絕與兩正對照通過。367／369／370／371追加原F6EC及完整表回填，289追加測試契約回填，歷史正文保持。

實際Docker入口依序：
```text
python3 workplace/new-game-372-ready-review.py
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestIMULByte372' -count=1
bash workplace/new-game-372-full-run.sh
bash workplace/new-game-372-run.sh
python3 workplace/new-game-372-verify.py
```
READY審查在CPU編輯前執行；全套失敗後同命令重跑，原guest沒有重跑。Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀。實際原命令／輸出／23收據雜湊掛[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。工具e57e9e063b1713b087423a78bef1349237c3d4b0已推送github，主庫只改四份現況／歷程文件。Docker專案相關執行中／停止容器為零；主庫既有root-owned2437檔／272目錄保持。主庫基線95e9325dab29a7773a5ddf6f084202b68f10d490，提交後精確HEAD見Git。

下一步核對180M星圖COLONIES控制的來源與安全正常輸入前置，再READY測正常裝置輸入。主庫玩法RE-first保持，不猜typed／持久欄位。原字串用途未知、RNG未固定；日期不是seed。

## 2026-10-04：373正常星圖COLONIES來源與callback前置

373依372原完整23表、唯一index10矩形、完整核心／FPU／cap／PNG及CPU hash審查轉READY，只在180M既有cap快照後新增一筆只讀觀察。原index10的+24／+32／+44窗口、callback target8:2136D1／mask2B／14／14及IRQ47425／47425已取，224bytes八窗口可讀；+40實際0，維持原offset與未知型別。邏輯43,450唯一命中17,434–79,471，和COLONIES對應強推論，沒有送新輸入。

372全部12047共通原列、39PNG及final保持，完整核心／FPU／clock／VBE／callback／IRQ／RAM前後保持。兩私有區段可逆，全部公開internal／CPU／DOS／probe保持e57e9e0，沿372固定官方EXE全套，未重跑無關CPU測試。六CLI拒絕與兩正對照、原418來源及state終態／同guest副本／UID GID保持，原guest一次。首次驗證及追加回填檢查均通過。

實際Docker命令：
```text
python3 workplace/new-game-373-ready-review.py
bash workplace/new-game-373-run.sh
python3 workplace/new-game-373-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀，owned監測550s與trap收尾。18份私有收據SHA-256掛[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，公開規格／索引與372追加回填。工具f92dd15be1f5bf94d193d9bfc0367f2baa5793a8已推送github，主庫只改四份現況／歷程文件；基線70043e75469a24fe7dc9d353566f89d23d9e782f，提交後精確HEAD見Git。Docker專案相關執行中／停止容器為零，既有root-owned2437檔／272目錄保持。

下一步審查COLONIES正常press／原查詢／首安全release與consumer，明示新輸入所需有界後續窗口。不修改本373的180M收據、不代寫原核心或RAM。主庫玩法RE-first保持；正式存讀語意／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：374正常COLONIES座標輸入與原選取10

374依373完整180M前置經READY，新私有模式唯一185M、固定後續5M；mode off保持原180M。一次press x86,y450、原24C31B的AX3查詢真正讀到BX1／CX86／DX450，42860µs後首安全release x88,y450。180020238原20DDDB／66A3A6C42600寫DS188:26C4A6 word0000→0A00，完整核心／FPU保持；兩214104 RET按真實stack返回20DB5B／174742，AX0與ESP+4已驗，不猜AX10或持久語意。

實際cap185M／223A71／unique_sites55872，無CPU拒絕。終圖親看全黑，RGB全零而indexed非全零，header count20但1100byte完整表未取；列表畫面／色盤與轉頁邊界未知。沒有延長預算或宣稱COLONIES列表已開啟，下一步保持同185M補只讀新表、當前stack／code與索引／RGB對應。

373新輸入前11981共通原列、39PNG保持，只正規化兩處已明示baseline budget；實際185M另核對。八patch可逆，全部公開internal／CPU／DOS／probe保持f92dd15；沿372固定EXE全套，未重跑無關CPU測試。13CLI拒絕與三正對照通過，原418來源及state／同guest副本／UID GID保持。原guest一次。初版原輸入驗證通過，擴充終態的非空parser讀空records失敗；改明示空欄位判準後同收據通過，首失敗保留，非產品或native變動。

實際Docker入口：
```text
python3 workplace/new-game-374-ready-review.py
bash workplace/new-game-374-run.sh
python3 workplace/new-game-374-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。20份私有收據SHA-256見[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，367／371／372／373同次追加新星圖上下文。工具20fd1507c45a053860b0ad50ad9f9f86049566ad已推送github，主庫只改四份現況／歷程文件；基線ec8531ee553929c5859c8f13a27c07918ccf8008，提交後精確HEAD見Git。相關執行中／停止Docker容器為零；既有root-owned2437檔／272目錄保持。主庫玩法RE-first保持，正式存讀／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：375完整20表與黑畫面色彩來源

375保持374正常COLONIES輸入與固定185M，一次原guest重播，只補兩筆只讀快照。DS188:298848 count20／stride55完整1100byte表取得，當前code16與SS:ESP stack16保存，未命名物件語意。原DAC768bytes全0、maskFF，非0像素索引12723個仍逐pixel映成黑；黑色來源已證實為當前全零DAC，轉頁原因與列表正常操作仍未知。

374全部12253共通原列／39frames／black final保持，實際185M／223A71無CPU停止。全核心／FPU／clock／RAM／VBE／callback與IRQ／device／ports／色盤前後保持；三私有patch可逆，getter僅在/tmp隔離編譯，公開internal／CPU／DOS／probe不變。13CLI拒絕與三正對照、原418來源／state／同guest副本與UID GID通過。無新CPU行為，沿372固定官方EXE全套，未重跑無關測試。

實際Docker入口：
```text
python3 workplace/new-game-375-ready-review.py
bash workplace/new-game-375-run.sh
python3 workplace/new-game-375-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。19份私有收據SHA-256見[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，工具03dcee257142e790661be653e9ba6102136d162a已推送github，374正文保留並追加回填。主庫只改四份現況／歷程文件；基線8a0985c7beabe2ecc190bf14fa546fe3653572a2，提交後精確HEAD見Git。相關Docker容器清理；既有root-owned2437檔／272目錄保持。

下一步保持相同輸入與185M，審查internal/machine/machine.go的DAC ports既有寫入入口，再以有界私有只讀觀察保存180M起到185M的DAC寫入總數、首個全零與最近寫入邊界及其原核心／clock。只觀察既有寫入，不添加IO、色盤修補、輸入或盲目擴cap，不深入DAC／PIT／driver或renderer helper。主庫玩法RE-first保持，正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：376正常轉頁窗口DAC降色與歸零序列

376保持375正常輸入與185M，原guest一次。從180M press前原DAC592非0值保存實際device寫入序列，11275事件獨立算術重播通過；11輪maskFF／index0..255／768色值均單調不增。首次全0與末DAC事件同為device sequence290512、loop觀察182566943／419464025µs／目前EIP222D1C，到185M沒有新DAC write。原降色來源已證實，恢復／轉頁完成與列表正常操作未驗。

375全部12255共通原列／39frames／黑終圖與兩只讀快照保持，只多journal摘要。每group hash／非0數／index／phase／mask／ports累計與終態核對；四patch逆轉375、private getter純讀，public internal／CPU／DOS／probe不變。原來源／state／同guest副本與UID GID1000、13CLI拒絕與三正對照通過。沿372固定官方EXE全套，不重跑無關CPU測試。

實際Docker入口：
```text
python3 workplace/new-game-376-ready-review.py
bash workplace/new-game-376-run.sh
python3 workplace/new-game-376-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。工具3295ddcac19dfbbebed167a490cebed7859c86a2已推送github，374／375正文保留並追加376回填。21份私有收據SHA-256見[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，原事件與rawDAC維持本機。主庫僅四份現況／歷程文件，基線fc54d5b2031a4d3b0690cfe0a50befaa9dd0b7c7，提交後精確HEAD見Git。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

下一步以376已證實的11輪單調降色與182566943歸零為來源，先審查新的轉頁觀察契約：保留185M完整前置，限定追加一次10M窗口至195M，追首個恢復非0色值的DAC寫入並保存原核心／clock與可見頁；未恢復時記錄實際邊界，不以加碼重跑求過。不得代寫palette、增加玩家輸入或深入DAC／PIT／driver及renderer helper。主庫玩法RE-first保持，正式存讀／完整開局／RNG與remake同狀態未驗；固定日期不是seed。

## 2026-10-04：377降色後有界續跑與原2C17停止

377先核對376完整185M前置，原MAX_STEPS guard保持185M，另runSteps195M、一次10M。原guest一次，實際188532362在1F455D／2C17因未支援2C停止；EIP抓opcode後1F455E，不是SUB已執行，沒有195M step_limit。窗口新增一輪1025筆全零DAC事件，首非0恢復未命中，終PNG親看仍黑。20表有10個raw byte自然變化，語意未知。

376共通12188列至185M journal／11275 baseline groups／39frames與黑PNG保持；後續完整12300 DAC事件獨立重播，baseline／終PNG獨立解碼RGB及palette histogram映色通過。八patch可逆回376，public internal／CPU／DOS／probe不變；18CLI拒絕4正對照、原418來源／state／同guest副本／UID GID1000保持。沿372固定官方EXE全套，未重跑無關CPU測試。

私有生成第一次外層here-document與內嵌PY重名，Python syntax在執行前失敗；更名外層PY_GEN_377後同內容成功。此為腳本界符問題，原guest尚未啟動；首失敗摘要保留，native只啟動一次。

實際Docker入口：
```text
python3 workplace/new-game-377-ready-review.py
bash workplace/new-game-377-run.sh
python3 workplace/new-game-377-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。工具168b91b9a8cb08a36f9517ae031a98239f841161已推送github，376正文保留並追加377。25份私有收據SHA-256見[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，主庫僅四份現況／歷程文件，基線8abae2ab000083e0a34fe97d7ce98d9a9905af2b，提交後精確HEAD見Git。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

下一步依原dosgolem_high_le:1F455D／2C17與188532362完整來源，建立CPU386 SUB AL,imm8規格，核對Intel SDM契約及既有byte SUB旗標模型；READY後補2C、獨立256×256輸入及EAX高24bit／其它核心與非算術flags保持、立即數fetch失敗驗收，再重播相同377窗口，不再擴cap。主庫玩法RE-first保持，列表正常操作／正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：378補SUB AL與原殖民地列表恢復

378先核對377原2C17完整停態與Intel SDM，再建READY規格。真正舊core單一RED後補naked2C，393216組、64種初flags邊界、fetch失敗／工具prefix拒絕／EIP wrap及固定官方EXE的CPU386／machine全套通過。首次wrapper誤加Read8／Write8 Bus不存在的Write16／Write32，編譯在測試前失敗；移除後才保存真正RED，不把編譯失敗當CPU證據。

相同正常輸入與一次195M窗口，原188532362自然SUB AL1Ah→03h／flags206h、CMP→293h／JA未跳／MOVZX四consumer通過，195000000達step_limit／EIP22C8BA。首非0DAC在189322149但首恢復PNG仍黑，195M終PNG親看顯示Sol II殖民地列表。全部24601 DAC事件獨立重播與PNG／palette映色核對通過；未驗列表操作或人口調整。

377至2C共通12393列、完整185M前置／39frames／黑PNG與12300 DAC事件前綴保持。五private patch可逆回377，public CPU單一2C區段可逆回168b91b，其它internal／原probe保持；18CLI拒絕4正對照與原418來源／state／同guest副本保持，原guest一次，沒有增加cap。第三RET、20表與共享word0000另保存，較早選取10不當持久結果。

實際Docker入口：
```text
python3 workplace/new-game-378-ready-review.py
bash workplace/new-game-378-cpu-tests.sh
bash workplace/new-game-378-run.sh
python3 workplace/new-game-378-verify.py
```
Go1.24.13／network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。工具d2df07fb795875ffdcdd2b8566ae6ecadc073070已推送github，公開CPU／自製測試／規格／索引與377回填；31份私有收據SHA-256見[既有研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫僅四份現況／歷程文件，基線de43636b754208aa33b4371f468fefc5b238c5b4，提交後精確HEAD見Git。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

下一步以195M可見Sol II與同時取得的20物件表，核對正常列表行的熱區／原選取來源及callback前置，再建立一次正常press／原AX3 poll／安全release的限定驗證。不增加cap或猜欄位，不深挖DAC／PIT／renderer helper。主庫玩法RE-first保持，正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：379殖民地名稱行的原第一命中來源

重用378完整195M收據與可見Sol II列表；固定官方1.31 EXE在IDA9.4 locked-v1重建一次性DB，原signed含端點矩形、index遞增與首次命中直接離開已核對。logical43,48／44,48命中13及19，原順序先13；右上角101,35另命中16，仍先13。10邊界與13前置漂移拒絕通過，完整195M核心／FPU／IF／callback IRQ與table／PNG hash保持。本輪沒有新guest或玩家輸入，未把預期13當原選取成功。

首次IDA anchor把214104錯換為114104，與原C3不符；按F0000h映射修正到124104及callback1236D1後重建一次性DB，真正RET／邊界吻合。主選取11CEF5匯出保持。初版邊界oracle漏列101,35的重疊16，被驗證拒絕；按原表修正並保留首失敗。兩者為研究定位／驗證預期問題，CPU、原版收據與輸入不改。

實際容器入口：
```text
bash /out/new-game-379-ida-run.sh
python3 workplace/new-game-379-ready-review.py
python3 workplace/new-game-379-source-verify.py
```
IDA image6f6d59af49d0／UID1000／network none／120s／2GiB／2CPU／128pids，官方patch唯讀掛/patch，既有工具workplace掛/out，tmp DB不持久。非空JSON／schema1／EXE hash／5365函式／擁有權通過，idat exit1不作判準。Python只讀核對沿既有Go1.24.13 image，30s／512MiB／1CPU／64pids。

工具c6d319edf0f8a8bacfdc1a53d2eb39a5a53205c1已推送github，公開379來源規格／索引與378回填，其它public internal／CPU／DOS／probe保持d2df07f。16份私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫僅四份現況／歷程文件，基線68532b9d3e60e7c8f49d850cfef4d6456595a982，提交後精確HEAD見Git。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

下一步依379已核對的195M完整來源與原first-match順序，建立Sol II行一次正常press／原AX3 pressed poll／首安全release的獨立READY契約，明示新玩家輸入的固定後續預算；先驗原index13選取store，再記實際畫面。logical43,48／44,48對應physical86,48／88,48，不代寫word13或跳handler，不重啟／重擲／加cap挑結果。主庫玩法RE-first保持；正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：380正常Sol II行選取13與200M黑終圖

379完整195M來源經READY審查後，private mode明示新玩家輸入的200M／一次5M窗口。原195M source通過，一次physical86,48 press／原AX3 poll／physical88,48安全release，持按469732µs。原195225486 RET按真實stack返回20DB5B／AX0，195226311在20DDDB／66A3A6C42600寫DS188:26C4A6 word0→13；正常行選取已證實。原guest一次，沒有重啟／重擲或增加cap挑結果。

實際step_limit200M／EIP223A23，無CPU拒絕；200M終PNG親看黑，indexed／DAC／RGB全0、table count1／55零bytes。新點擊後11輪11275 DAC事件單調降色，196376553／450990327µs首全0與末write；殖民地正常畫面尚未驗，沒有把轉頁中間態稱為產品缺陷或完成。

378共通12701列至195M、39frames、185M／首恢復PNG、24601事件前綴及195M可見frame／PNG保持；完整35876事件獨立重播與PNG／palette映色核對通過。14patch逆轉精確回378，所有public internal／CPU／DOS／原probe保持；34CLI拒絕與正對照、原418來源／state／同guest副本及UID GID1000保持。沒有新CPU行為，沿378固定官方EXE CPU386／machine全套，不重跑無關測試。

實際容器入口：
```text
python3 workplace/new-game-380-ready-review.py
bash workplace/new-game-380-run.sh
python3 workplace/new-game-380-verify.py
```
Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾；收據驗證90s／1536MiB／1CPU／128pids。兩次驗證只重讀同一收據，沒有重跑原版。

工具5d3f5b80373c4872e1401366d3b5dd482577252a已推送github，公開380限定規格／索引及378／379回填。29份私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，主庫僅四份現況／歷程文件，基線157f583b47b8682910eec5da7b95293b9b3ab2ba，提交後精確HEAD見Git。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

下一步依200M原完整核心、count1／55零bytes、當前dosgolem_high_le:223A23 code與SS188:2BD908 stack16，只追正常畫面建立所需的最小上層來源；先核對21A6F3是否真為該路徑的return定位及原呼叫邊界，再由來源決定下一個有界畫面觀察，不盲目加cap或深挖renderer／DAC／PIT helper。主庫玩法RE-first保持；人口調整／正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：381 同200M原堆疊框架定位

承接主庫76bd674e10745561e43c53832d51d5ff99083b4c與工具5d3f5b80373c4872e1401366d3b5dd482577252a，沿既有復古RE／文件職責／resolution-backlink路由；主庫玩法閘門保持。IDA9.4固定官方EXE的序言、尾端、七直接CALL及file offsets通過獨立核對，381先DRAFT→READY，才寫private觀察器。

原380輸入及200M cap保持；只在terminal既有只讀保護內讀32bytes原SS:EBP。真正SS188:2BDB78=22341C，保存EBP2BDB9C，吻合sub_133237內IDA 133417的直接CALL。ESP首值21A6F3位於本函式260h區域空間。原槽值／靜態CALL已證實，當前caller鏈強推論，自然RET及UI未驗。

原guest一次；完整13032原列、35876 DAC事件、非新增final、全部舊PNG、34CLI、418來源與state／副本保持。兩private patches可逆回380，公開internal／CPU／DOS／probe不改，沿378固定官方EXE CPU386／machine測試，不重跑無關測試。381限定CONFORMED，380正文保留追加回填，舊stack offset在不同生命週期的322／323／348／349／352不受影響。

例外分類：IDA第一查詢引用數上限與第二查詢缺少函式邊界，原查詢保留first-query，修成明示bounded及unknown；私有建置guard使用不可見ports而中止，沒有guest，改用既有只讀snapshot clock後同命令重跑。一次函式編排的const賦值錯誤與只讀查詢的heredoc結尾錯誤修正；不當產品缺陷。驗證器在同份原收據跑兩次，第二次加380回填與實際槽值護欄，均通過；沒有native重啟求過。

工具6e3dc0cb0e6019884bc3540a6c10598da2e1248c與繁體用字訂正a03c322d28002bb0e11d5d5109d5833a17114d04已推送github既有分支。首次push被自動審核拒絕，理由為目的地尚未核實；核對本機來源與clone皆指向wicanr2/dosgolem.git、唯讀ls-remote為原5d3f5b8，並確認三份公開文件／69行無私有收據後，相同命令審核通過。主庫本輪只更新CONTEXT的DOS行、WORKLIST唯一DOS活表與兩份歷史追加；提交／push前核對exact範圍、版權與全部私有SHA-256。

實際容器入口及46份私有收據見[研究紀錄](docs/re/dosgolem-moo2-intake-20260930.md)。所有寫入UID/GID1000；相關Go與IDA容器已清空，既有root-owned2437檔／272目錄保持，沒有.md目錄。無新image。200M仍黑／count1空表，不宣稱殖民地畫面、完整開局、正式存讀、RNG或remake同狀態完成。下一步：只查IDA linear EA sub_133237的直接上層CALL／返回邊界與正常畫面建立入口，使用同200M原SS188:2BDB78=22341C及保存EBP2BDB9C作錨；來源充分後才決定一次有界畫面完成觀察，不延伸palette／renderer／DAC／PIT helper或盲增cap。

## 2026-10-04：382原上層自然返回與可見殖民地畫面

先由固定官方EXE的IDA Pro9.4核對sub_133237有界退出、26直接CALL與原父框架契約，再READY及修改private observer。保持原Sol II行輸入；200M原父槽SS188:2BDBB4=1B078E唯一命中IDA C0789。203011820原2234BA C3自然返回1B078E／ESP+4已證實，其餘核心／FPU與RAM保持。205281523首非0DAC仍黑，210M終圖親看可見Colony of Sol II，36筆完整物件表取得。尚未驗人口操作、正式存讀或remake同狀態。

原guest一次，明示200M→210M的一次10M觀察，沒有新輸入、重啟或增加cap挑结果。原200M完整前置／12962列／35876 DAC與所有舊PNG保持；新49246 DAC獨立重播、PNG獨立解碼／CRC／palette／histogram、46CLI、九patch逆轉381與原418來源／state／同guest副本通過。public internal／CPU／DOS／原probe保持a03c322。

驗證器首次錯找不存在的200M color_source標記，第二次漏沿381既有row journal封裝hash正規化；逐項原核心與畫面已獨立驗證後才修正。原收據不改，沿同收據完成驗收及回鏈護欄。Git格式檢查另移除新文件EOF空白行，沒有功能變更。

實際Docker入口：
```text
bash /out/new-game-382-ida-run.sh
python3 workplace/new-game-382-ready-review.py
bash workplace/new-game-382-run.sh
python3 workplace/new-game-382-verify.py
```
沿381既有IDA locked-v1及Go1.24.13 image、network none、UID/GID1000；IDA120s／2GiB／2CPU／128pids、native600s及owned監測550s／2GiB／2CPU／128pids、驗證90s／1536MiB／1CPU／128pids，原ZIP與patch唯讀。來源審查／建置／CLI首次通過，原native session34285 exit0；exit0不代表guest完成，實際step_limit210M已另核對。

工具1929523731e5f1af2c1bbb446cdfd89e28401af4已推送github，公開四份自撰規格／索引／380與381回填，私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫僅改四份現況／歷程文件，基線9d0b38403fbb88c1b8aba8e4930f5237a4a8ecff，提交後精確HEAD見Git。Go及IDA本專案容器已結束；收尾稽核通過，既有root-owned2437檔／272目錄保持，無.md目錄。

下一步保持210M原來源，核對36筆物件表的職業列熱區、原輸入消費端與callback安全前置。主庫玩法RE-first保持；人口操作、正式存讀、完整開局、RNG及remake同狀態未知，固定日期不是seed。

## 2026-10-04：383殖民地職業列來源與LE重定位核對

沿既有復古GUI還原、RE技能及文件職責入口，保持主庫RE-first。固定官方1.31 EXE，八份IDA Pro9.4查詢保留原名稱／EA／運算元／bytes／file offsets與xref邊界。三個kind6控制的原熱區、+18h word與+20h pointer、11C2CF→1156E2及115988暫存寫入已錨定；場景回呼與人口選取／放置鏈僅作來源候選，正式人口操作未驗。本輪沒有guest執行、新輸入、cap變更或CPU改動。

獨立核對4349筆指令紀錄／3578原EA、674筆含fixup紀錄；原MZ26654／LE292E4、2objects／365pages／51363原重定位記錄逐筆驗證。16熱區端點與8水平模型通過，只證明來源推導。原382完整210M核心／FPU／clock／callback IRQ／36表／PNG與journal保持。公開internal／CPU／DOS／原probe保持1929523。

驗證器首次把IDA已重定位bytes直接當原file bytes，在BF80E被拒絕；以原fixup record獨立重建後同來源通過。LE初版讀錯外層MZ入口，沿既有原26654測試入口修正。原失敗腳本與輸出保留，沒有改EXE、guest收據或刪除bytes差異。空switch查詢是比較樹的查詢方式不適用，沿原CMP定位；不是原分支缺失。

實際Docker入口：
```text
bash /out/new-game-383-ida-run.sh
bash /out/new-game-383-ida-ownership-run.sh
bash /out/new-game-383-ida-consumers-run.sh
bash /out/new-game-383-ida-worker-run.sh
bash /out/new-game-383-ida-mutation-run.sh
bash /out/new-game-383-ida-dispatch-run.sh
bash /out/new-game-383-ida-pick-run.sh
bash /out/new-game-383-ida-contract-run.sh
go run workplace/moo2-383-le-inspect.go
python3 workplace/new-game-383-source-verify.py
```
沿既有IDA locked-v1 image與Go1.24.13 image，network none／UID/GID1000，原patch唯讀。IDA每次120s／2GiB／2CPU／128pids；Go來源讀取與Python核對均有外層逾時及資源限額。八份IDA schema1／5365函式／EXE hash及輸出擁有權通過；idat exit1以非空有效JSON另核對。末次驗證session85427 exit0，六項PASS，沒有重跑原版。

工具來源提交62b65ae50fefafbeaafba433216e3d3edac07e72及索引／繁體訂正b331bb640e4932dc59f5e4da694c63530132f4c1已推送github，公開只有三份自撰文件。索引382的舊UI未知已改為畫面已驗／人口操作未驗。50份私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，原EXE／JSON／LOG／PNG／RAM／state維持本機忽略目錄。

主庫僅改四份現況／歷程文件，基線4677bb9c35ede4f0ae9b661238d62e1fcae3cff5；提交後精確HEAD見Git。收尾核對工具已推送且工作樹乾淨、收據UID/GID1000、既有root-owned2437檔／272目錄及無.md目錄；本專案一次性Docker容器已結束。下一步保持同210M建立只讀pointer／current colony／pool／record／callback觀察契約，先取得原前置再訂人口正常操作；固定日期不是seed，正式人口變更與remake同狀態未驗。

## 2026-10-04：384同210M殖民地控制前置的只讀快照

上一輪383已推送並完成來源核對，分類為progress。開工核對AGENTS／CONTEXT／HONEST-STATUS／唯一活表及Git現況；命中復古GUI、spec閘門與回鏈入口，主庫RE-first保持。384先DRAFT及索引、固定382完整210M／383原bytes審查通過，才READY及生成五個可逆private patch；沒有改公開CPU、DOS或主庫玩法。

原guest僅一次，session6185 exit0；實際guest是step_limit210000000／228DF6／483821442µs。保持同輸入與cap，完整382原核心／FPU／clock／callback IRQ／36表／DAC／journal及全部舊PNG、原418來源與state／副本核對通過。58CLI為48拒絕與10正對照，原46逐項保持；快照before=after且本run RAM hash前後相等。

三個控制pointer word均310，current raw4、pool5B2044、完整361byte record5B25E8取得；不把原選取13混為current raw4，不命名job數量。scene callback2A8840原值0、enable26C48C為1均可讀。這個實際前置改變下一步：先追原C07C1→sub_BF456返回及C07D2／C07E1回呼設置，不能直接據熱區送人口操作。

兩次失敗均分類為腳本：READY初版對115988間接WORD指令誤要求relocation；修正條件後同來源通過。獨立驗證首次exec來源檢查覆蓋日誌變數b，TypeError發生於原日誌迭代；隔離namespace後同一收據通過全部驗收，session58828 exit0。首腳本／輸出保留first-check，沒有重跑guest、修改原收據或加cap。CONFORMED後再跑來源及文件回鏈閘門通過，來源閘門可接受已核對的完成狀態。

實際Docker入口：
```text
python3 workplace/new-game-384-ready-review.py
python3 - <<'PY_384_GEN'
bash workplace/new-game-384-run.sh > workplace/new-game-384-run-output.txt 2>&1
python3 workplace/new-game-384-verify.py > workplace/new-game-384-verification.txt 2>&1
python3 workplace/new-game-384-source-verify.py > workplace/new-game-384-source-tests.txt
python3 workplace/new-game-384-document-gate.py > workplace/new-game-384-document-gate-tests.txt
```
生成器由容器heredoc執行，原文保存於私有new-game-384-generate.py。沿既有Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；network none／UID/GID1000，原ZIP與patch唯讀。來源／文件檢查30s／512MiB／1CPU／64pids，native600s／2GiB／2CPU／128pids及owned PID550s／trap，驗證90s／2GiB／1CPU／128pids。本輪沒有IDA或新image。

工具0a6c7f97c75262c1d36098f5d6763918e91e2305已推送github，公開只有索引、384自撰契約與383追加回鏈。32份私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫只改四份現況／歷程文件，基線d6e4933c116fefdc8a4610773c3ed77ce810d686；提交後精確HEAD見Git。收尾核對擁有權、兩庫狀態與Docker；既有root-owned2437檔／272目錄保持，無.md目錄，本輪容器已結束。

下一步只查原sub_C058A的C07C1→sub_BF456返回邊界與回呼設置，不送人口輸入，不盲增cap或深入共享renderer。正式人口操作／存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：385原場景設置與首輸入、回呼runtime地址勘誤

384已推送並分類progress。開工核對唯一現況表與兩庫Git，命中復古GUI、spec閘門與文件回鏈入口。主庫RE-first保持，本輪只有原版來源、私有觀察與文件訂正。

沿固定官方1.31 ORION2.EXE與IDA9.4一次性資料庫核對BF456／C058A／1191CA；分別43／215／31項，1191CA的205 direct callers只匯出32並標截斷。327筆來源紀錄／325原EA、71筆含fixup差異核對通過，原LE2objects／365pages／51363 fixup保持。私有RE輸出的「8份IDA」標籤沿用舊字樣，本輪實際僅新增一份查詢。原file bytes、file offset、IDA linear EA與dosgolem_high_le分列。

先DRAFT→READY審查，才生成8個可逆private patches及70CLI。原guest僅一次，session26846 exit0，明示200–220M有界觀察；原210M守衛通過，沒有新輸入、重啟或CPU改動。原210M完整核心／FPU／clock／callback IRQ／36表／DAC／journal／39frames／PNG、418來源／state與副本保持。

原BF456於203219421由1AF4FF／C3依真SS stack返回1B07C6；203219427到1B07E1，EAX1AED21；203219451原2091EA執行A340882900，203219467由209225／C3返回1B07E6。首正常輸入CALL在205804505／1B0845，PNG與原210M相同。兩次RET的ESP+4／其他核心／FPU／RAM保持通過。

原A3實際寫DS188:298840，IDA dword_1A8840加F0000h也為298840。384及385觀察器卻讀2A8840，算錯10000h；原零raw值保持，但撤回scene callback0語意。錯誤排除範圍導致store其他RAM保持檢查失敗，不當CPU缺陷。正確位移實際讀回及全RAM僅改正確四bytes未知；EAX寫入只列強推論。原current／pool／361record／三word／enable與水平偏移保持，先前仍需等待設置的推論被原時序否定。

獨立驗證session93305 exit0，SAFE EVIDENCE PASS限返回／首輸入／舊前置，ADDRESS MODEL REJECTED保留回呼模型失敗。原385收據與private Go不改，不靠重跑取得綠色；契約READY退回DRAFT，383與384只追加勘誤，索引移除目前錯誤斷言。來源及CLI是在原READY階段通過，DRAFT後不重新啟動錯誤觀察器。

三項環境／腳本問題已保留：RE verifier最初寫入唯讀mount，改用既有可寫工作樹；生成器首版marshal定位到三處，在寫Go前拒絕，改唯一payload定位；獨立驗證首版Python重新序列化改變Go事件key順序與HTML escape，改核對原journal事件文字切片。沒有改guest或原收據，也沒有把它們記作產品問題。

實際Docker入口：
```text
bash /out/new-game-385-ida-run.sh
python3 workplace/new-game-385-re-verify.py
python3 workplace/new-game-385-ready-review.py
python3 workplace/new-game-385-generate.py
python3 workplace/new-game-385-source-verify.py
bash workplace/new-game-385-run.sh > workplace/new-game-385-run-output.txt 2>&1
python3 workplace/new-game-385-verify.py > workplace/new-game-385-verification.txt 2>&1
python3 workplace/new-game-385-document-gate.py > workplace/new-game-385-document-gate-tests.txt
```
生成器及來源檢查由容器執行，原文保存在私有workplace。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；IDA9.4 locked-v1 image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。network none／UID/GID1000，原ZIP／patch唯讀；IDA120s／2GiB／2CPU／128pids，native600s／2GiB／2CPU／128pids及owned PID550s／trap，驗證90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids。沒有新image，相關容器已結束。

工具3842529eb5dff704adf3d33b6c4f5720ebe9cbb4已推送github，只改四份自撰文件；45份主要私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫只改四份現況／歷程文件，基線68ee2f7beafd94f7a8ef87d8957ee615a9c4e0d2；提交後精確HEAD見Git。收尾核對兩庫狀態／擁有權／Docker，既有root-owned2437檔及272目錄保持，無.md目錄。

下一步386修正DS188:298840只讀觀察，依原A3及已定位首輸入完整狀態作守衛；維持原輸入與210M，不再等待或盲增cap。人口正常選取／放置、正式存讀、完整開局／RNG與remake同狀態未驗。

## 2026-10-04：386原回呼正確地址與首輸入只讀前置

385已推送，分類progress；新證據已改變下一步，回呼原零值語意撤回。開工核對AGENTS／CONTEXT／HONEST-STATUS／唯一活表、兩庫HEAD及乾淨工作樹；沿已載入復古GUI與spec閘門，新增文件前核對職責，回填前載入resolution backlink。主庫RE-first保持。

386先DRAFT／索引與固定385原A3、IDA+F0000h投影及完整首輸入收據審查，再READY及11個private patches。新CALLBACK_READ只接受1並要求原SCENE／JOB等前置；mode off逆回385，mode on取消額外10M，固定原輸入與210M。82CLI為68拒絕及14正對照，原70逐項保持。public CPU／DOS／internal／原probe不改。

原guest僅一次，session78051 exit0，actual_boundary及run_limit均210000000／step_limit；末態228DF6／483821442µs。203219451原2091EA的A340882900指定DS188:298840；原回呼由55451700變21ED1A00／raw1AED21，等於EAX。依實際moffs及descriptor Base求linear，observer自身Step前RAM副本只替換實際四bytes後與Step後整RAM hash相同。observer沒有代寫guest或修正EAX；全RAM其他bytes保持已證實。

原兩個真RET及四事件時序與385保持。首正常輸入205804505／1B0845先經固定385完整frame守衛，僅排除跨run ram_sha256；第9只讀窗298840為21ED1A00，原8窗口／current／pool／361record／三word與舊raw保持，snapshot before=after及RAM前後hash全等。獨立驗證session12161 exit0，386 CONFORMED SCOPE PASS限定此次原A3與首輸入；完整382／384原210M核心／FPU／clock／callback IRQ／36表／DAC／journal／39frames／PNG／418來源／state及副本保持。沒有原guest重跑或新增人口輸入。

生成器前兩次因原Go縮排與字串定位不符，在寫Go前拒絕；連續同類失敗後回查spec路由，檢查剩餘全部原縮排，再生成。首／次版本與分類摘要留first-check及second-check；原385 Go及收據不變。這是腳本定位問題，不當產品故障。獨立驗證第一次即通過，不追加可選CPU／renderer測試。

實際Docker入口：
```text
python3 workplace/new-game-386-ready-review.py
python3 workplace/new-game-386-generate.py
bash workplace/new-game-386-run.sh > workplace/new-game-386-run-output.txt 2>&1
python3 workplace/new-game-386-verify.py > workplace/new-game-386-verification.txt 2>&1
python3 workplace/new-game-386-document-gate.py > workplace/new-game-386-document-gate-tests.txt
```
source verify由run.sh在82CLI後、guest前執行。沿Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／UID/GID1000；原ZIP及patch唯讀。native600s／2GiB／2CPU／128pids、owned guest與capture由550s監測及trap收尾；驗證90s／2GiB／1CPU，來源／文件30s／512MiB／1CPU／64pids。本輪沒有IDA／新image，相關一次性容器已結束。

工具3e2290007d5d0163346b4150c7e8cc625b1e0166已推送github，公開五份自撰文件，383／384／385按同一不可變定位追加回填；385錯誤觀察器仍DRAFT。37份私有主要收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫只改四份現況／歷程文件，基線0910ddd90ef84b0b81533ab6ff8cf5e01672c7b0；提交後精確HEAD見Git。收尾核對擁有權／兩庫狀態／Docker，既有root-owned2437檔及272目錄保持，無.md目錄。

下一步387追原sub_1171AB與kind6座標消費及按下／放開契約，證據足夠才訂一次正常人口操作觀察。原回呼讀值完成不外推人口選取／放置、正式存讀、完整開局／RNG或remake同狀態。

## 2026-10-04：387職業列條件座標與持按／放開來源

386已推送且分類progress，正確回呼與原首輸入前置已驗。開工核對AGENTS／CONTEXT／HONEST-STATUS／唯一活表及兩庫乾淨HEAD；路由命中復古GUI、spec閘門及resolution backlink，沿已載入逆向及IDA技能。主庫RE-first保持，本輪不改玩法／CPU，不啟動guest或送人口輸入。

五個有界IDA查詢定位原1171AB、1236D1座標producer、123ABA／123AE7 getter、裝置range初始化、123D53事件座標、持按選取器113FB9與11CEF5玩家分支切片。2562筆rows／2091原EA、514筆fixup差異對原EXE及LE51363 records通過；file offsets／file bytes／IDA relocated bytes分存bytes index。11CEF5有1550項，只保留頭尾及生命週期／press切片，caller最多32並保留截斷，不深挖共享繪圖及平台helper。

原callback1236D1在兩個原旗標為0時，把ECX低word作signed SAR1成GUI X，Y保存EDX低word，1237F2／CB為遠返回；對應runtime8:2136D1與386原target一致。原X／Y runtime2A3A38／2A3A36、旗標26C51A／26C51C；Y不是2A3A3A區域暫存。原初始化range為2×(width−1)、height−1。持按選取器是113FB9，事件分支另讀123BC1／123BEE座標。AX3為0後11E4EB對kind6在11E508呼叫1192D1，11E50D清共享選取；不是送一次pressed poll就證明人口已消費。

DRAFT與索引、固定來源及完整386前置審查後READY，才建立只讀模型。13 signed X／3range／6原人口列first-match樣本通過，387 SOURCE PASS；均為來源推導，沒有原guest實測。原386首輸入／36表／正確callback及385／386收據、公開internal／CPU／DOS／原probe保持。實際旗標／寬高／座標／事件／按鍵與17C4E4未捕捉，是下一個有界raw前置；不預填0或送660,77候選。

RE verifier首次誤在唯讀mount建檔，未寫出檔案；改用既有可寫clone後相同原來源通過，分類環境問題。五次外層IDA均exit0，idat exit1以有效非空JSON／schema／固定EXE hash／5365函式／UID1000另驗，不能單靠exit1判定。本輪無新image或原始.i64改動。

實際Docker入口：
```text
bash /out/new-game-387-ida-run.sh
bash /out/new-game-387-ida-producers-run.sh
bash /out/new-game-387-ida-mouse-run.sh
bash /out/new-game-387-ida-lifecycle-run.sh
bash /out/new-game-387-ida-press-run.sh
python3 workplace/new-game-387-re-verify.py
python3 workplace/new-game-387-ready-review.py
python3 workplace/new-game-387-source-verify.py
python3 workplace/new-game-387-document-gate.py
```
IDA image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780，120s／2GiB／2CPU／128pids、patch唯讀、tmp DB一次性。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，來源／文件30s／512MiB／1CPU／64pids。全部network none／UID/GID1000，私有輸出擁有權核對；本輪容器已結束。

工具1d0d128c52a7de357a41a19b8bcbb2513bc70e46已推送github，公開五份自撰文件，379／383／386正文保留並追加不可變定位回填。34份私有主要收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)，原EXE／JSON／LOG／bytes／state不公開。主庫只改四份現況／歷程文件，基線06664fead4ccccc5d1f739196479d87d023c1f1c；提交後精確HEAD見Git。收尾核對兩庫／擁有權／Docker，既有root-owned2437檔及272目錄保持，無.md目錄。

下一步388在原205804505完整首輸入及正確callback守衛後，維持同輸入與210M，補讀動態旗標／range／座標／按鍵raw前置，再訂一次人口列press／持按消費／安全release。source-only完成不外推人口變更、正式存讀、完整開局／RNG或remake同狀態。

## 2026-10-04：388原首輸入raw與裝置範圍只讀驗證

387已推送且分類progress。開工核對AGENTS／CONTEXT／HONEST-STATUS／唯一活表及兩庫乾淨HEAD；路由命中復古GUI、spec閘門及resolution backlink，沿已載入逆向技能。文件職責先核對，主庫RE-first保持。

DRAFT及索引、固定原387 bytes／386首輸入與正確callback／裝置欄位來源審查後READY，才生成可逆private observer。新MOUSE_SOURCE只接受1，要求原CALLBACK／SCENE／JOB／upper／row／restore、185M及state前置；mode off精確逆回386，mode on固定原210M。原82CLI保持，新增10拒絕及2正對照，共94。私有overlay的MouseReadSnapshot388只取現有欄位，不Handle／Step／IO／送事件或改guest RAM。descriptor取樣在建置前改用既有Descriptors map；公開CPU／DOS／internal／原probe不改。

原guest一次，session58166 exit0；run_limit與actual_boundary均210000000／step_limit，末態228DF6／483821442µs。完整385及固定386首輸入守衛後，在205804505讀原8窗及device。原205804505／1B0845首輸入：DS188:26C51A／26C51C為0／0，width／height為640／480。目前GUI X／Y=43／48，保存事件X／Y=43／48，按鍵word=0，持按閘門26C4E4=1，共享active word26C4A6=0。裝置x／y／buttons=86／48／0，X range=0..1278，Y range=0..479，range設定旗標=True／True。

獨立核對session91475 exit0，388 MOUSE SOURCE PASS；完整core／FPU／clock／callback IRQ／VBE／DAC／RAM與device前後相同，原386的9窗及record／pointer words、原210M journal／39frames／PNG／418來源／state及副本保持。沒有新人口輸入或guest重跑。候選職業列輸入前置已可比較，人口變更仍未驗。

文件回填首版Python字串接合SyntaxError，在執行前拒絕，沒有修改文件；後續文件閘門拒絕尚為READY的狀態。保留first-check，修正字串後用同一容器入口完成回填與閘門，不改observer、原guest或收據，分類腳本問題。主庫稽核首兩版各因文件採「未知」或「未驗」而拒絕；回查文件職責後改核對兩種明示邊界，原資料與產品碼不變。

實際Docker入口：
```text
python3 workplace/new-game-388-ready-review.py
python3 -  # 生成器本文保存於 workplace/new-game-388-generator.py
bash workplace/new-game-388-run.sh > workplace/new-game-388-run-output.txt 2>&1
python3 workplace/new-game-388-verify.py > workplace/new-game-388-verification.txt 2>&1
python3 workplace/new-game-388-finalize.py
python3 workplace/new-game-388-document-gate.py > workplace/new-game-388-document-gate-tests.txt
python3 workplace/main-audit-388-final.py
```
source verify由run.sh在94CLI後、guest前執行。沿Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；network none／UID/GID1000，原ZIP及patch唯讀。native600s／2GiB／2CPU／128pids，owned guest及capture以550s監測與trap收尾；驗證90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids。沒有IDA查詢或新image，相關一次性容器已結束。

工具011fe510aa8cf74a00d26b7bbc7d65b6094f04bc已推送github，公開五份自撰文件，383／386／387按不可變鍵追加回填，385原錯誤觀察器仍DRAFT。35份主要私有收據SHA-256見[研究入口](docs/re/dosgolem-moo2-intake-20260930.md)。主庫只改四份現況／歷程文件，基線be66c792c8e3e51a7fe42a83d791046c423cbc07；提交後精確HEAD見Git。收尾核對兩庫狀態／擁有權／Docker，既有root-owned2437檔及272目錄保持，無.md目錄。

下一步389固定本次完整首輸入、原旗標及range，建立一次正常職業列press觀察契約；候選裝置660,77按原signed SAR1為GUI330,77，原36表先命中kind6 index1。先追原持按選取與1192D1／場景回呼，依實際消費點安全release；不把候選命中當人口變更或預設職業語意。 主庫玩法RE閘門保持，正式人口變更／存讀／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：389正常職業列按下／放開與人口原始差異

388已推送，分類progress。開工核對AGENTS／CONTEXT／HONEST-STATUS／唯一活表及兩庫乾淨HEAD；路由命中復古GUI、spec閘門及resolution backlink，沿已載入逆向技能。DRAFT及索引、固定388完整首輸入／raw／device、原36表first-hit1、持按CALL／callback operand／near RET／release CALL／CB與安全條件審查後READY，才生成可逆private observer。私有純只讀388裝置方法保持；建置前補入明示SS匹配，106CLI及source gate在guest前通過。mode off精確逆回388；mode on同210M，僅既有裝置方法送一次按下及一次放開。

原guest一次，session64247 exit0，210000000／step_limit。原205804505完整首輸入／raw／device守衛後，正常裝置660,77／buttons1按下；205813953進入runtime203FB9，205814268依真SS返回20E1AC，EAX1。205815045進入1AED21，206207980依真SS及ESP+4返回2092F9；同一步IF、pending0／inactive／IRQ安全及至少20ms條件成立後送buttons0放開。206264647到20E4EB零按鍵分支，206264656到20E508 kind6 CALL；206264681再次場景入、206658139返回2092F9。206658146／147執行20E50D清共享active並到20E516。15事件與所有取樣前後狀態、RAM及device均只讀，19個裝置回呼完成。

原DS188 current4／pool5B2044／record5B25E8的361bytes，在press至active-clear-after206658147全部保持；210M終態出現8個差異：+0Bh FF→02、+0Dh／11h／15h／19h 02→00、+C8h 49→B9、+C9h 00→FE、+E7h 08→00。三個UI pointer words由310／310／310變330／310／310。這是原raw實測，正式職務、選取群及放置語意仍未知。

正常玩家路徑的原PNG人工核對：首輸入顯示Colony of Sol II、Pop 8,000k (+73k)；210M顯示Research Colony of Sol II、Pop 4,000k (-327k)。終圖SHA-256 6be8a5cf20dbdfaca8c2471e407e9e70607c8201ff3633c261ebd9e5b71cdf26，末態EIP1A5042／482658319µs。顯示差異已觀察，不用它命名原欄位或宣稱人口配置完成。

獨立驗證首兩次漏掉解壓縮mtime及DOS DTA時間／日期。連續同類拒絕後回查平台規格入口、既有338比較規則及le_startup.go的0x16／0x18欄位；最後只排除mtime和DTA0x16..0x19，保留attribute、size、name及其餘bytes。完整首輸入／raw／device、13,233列共同日誌與49,210個press前DAC groups另行比對；每次RAM雜湊只依既有跨run規則排除，不遮玩法結果。三次驗證都讀同一批原收據，沒有guest重跑或原observer／CPU改動。 最終獨立驗證session90064 exit0，389 EVIDENCE PASS及389 INPUT CONTRACT PASS；42份主要私有收據hash見研究入口。另以actual state副本的hash／size／UID1000對manifest及388核對，389 SOURCE STATE PASS，418原輸入與SAVE10／MOX保持；正式存檔語意未驗。較早只讀搜尋腳本首版誤讀空functions陣列，未寫輸出；改讀既有bytes index與正確schema，屬腳本問題。

實際Docker入口：
```text
python3 workplace/new-game-389-ready-review.py
python3 workplace/new-game-389-generator.py
bash workplace/new-game-389-run.sh > workplace/new-game-389-run-output.txt 2>&1
python3 workplace/new-game-389-verify.py > workplace/new-game-389-verification.txt 2>&1
python3 workplace/new-game-389-state-verify.py > workplace/new-game-389-state-tests.txt
python3 workplace/new-game-389-finalize.py
python3 workplace/new-game-389-document-gate.py > workplace/new-game-389-document-gate-tests.txt
python3 workplace/main-audit-389-final.py
```
SS守衛及CLI私有調整以容器Python標準輸入完成，保存在可逆patches、final Go及run.sh。source verify由run.sh在106CLI後、guest前執行。沿Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／UID/GID1000，原ZIP／patch唯讀；native600s／2GiB／2CPU／128pids、capture550s及owned PID trap；驗證90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids。沒有IDA／新image，相關一次性容器已結束。

工具d2c3519528cc475d4c891990e72449a0698fd5bd已推送github，公開五份自撰文件，383／387／388追加不可變鍵回填與索引更新。主庫只改四份現況／歷程文件，基線65f93c9b957e4dde99a93361f8aa2d86a8cf8df7；提交後精確HEAD見Git。收尾核對兩庫／擁有權／Docker，既有root-owned2437檔及272目錄保持，無.md目錄。

下一步390沿同輸入及210M，追查206658147之後原C086E→C02F9與B9C3D／B9E94的最小正式寫入鏈，定位這8個record差異與可放置狀態；取得證據才訂一次跨職業列放置。不假設8,000k→4,000k已完成換職或刪除人口，不盲增cap或深挖renderer／平台helper。 主庫RE-first保持，正式人口配置／存讀／完整開局／RNG及remake同狀態未驗。

## 2026-10-04：390定位職業列選取後的原寫入鏈

基線93648beddaccdb5b41a12e4165f68c01226d2d15與工具d2c3519528cc475d4c891990e72449a0698fd5bd，兩庫乾淨。路由命中復古GUI、spec閘門、resolution backlink及文件職責；沿逆向技能，必要的實際writer補用IDA9.4技能及權威工具契約。READY前固定389完整前置／終態及383原bytes，私有Go可逆，CPU／DOS及主庫玩法保持。

首輪觀察器以目前DS必須188誤拒，session94008 exit1，原失敗88份產物按failed-390前綴及manifest保留；無原CPU缺陷證據。改監看固定descriptor188，實際getter隔離測試重現舊拒絕並證實三種DS切換、只讀及越界拒絕；DRAFT→修正版READY後同命令／同輸入／210M乾淨重跑。修正版session42905 exit0；獨立驗證session84523 exit0，同一批收據，無第三次guest。 私有getter在原Step前後直接讀已驗RAM，不掛新Bus，不新增輸入；118CLI、source gate先於每次guest。修正版監看固定descriptor188，以實際current4／pool5B2044守衛原record。首次失敗產物88份及manifest保持；275份較早主要receipt SHA／UID1000核對通過。

修正版原210M／step_limit保持389全部15事件、完整日誌與DAC journal、末態EIP1A5042／482658319µs及原PNG。206659890到C086E、206659891到C02F9、206660054到C0337、206660055到BF627。206697288原BF681將17AABB由0設1；206697307原B9C81把17A974由FFFF設4。原B9CAF於206697426／545／664／783依序將record+0Dh／11h／15h／19h的bit1清除，02→00。BF6ED及B9E94未到達，限定此次走選取分支。

69個實際Step變更重建四個監看範圍終值；8個首末record差異均定位。+E7h的word由原DE727在206700878寫0；+C8h先由原E19C6在206704020寫FE70h，再由E1CD9在206704659加到FEB9h，後續重算保持FEB9h；+0Bh由E1E64在206704718將FF改02。另有+EFh／F2h／FCh／104h等先清除再重建的中間值，首末比較不會顯示，均保留實際byte變更。record+0Ah原08保持；原+0B／C8／E7正式名稱與職務數量仍未定型。

新增一次窄IDA9.4查詢，16個未索引實際writer定位，233列／212個EA／14筆重定位差異；原MZ／LE、2object／365page／51363fixup records獨立核對。保留原始函式名、EA、file offset及bytes，runtime投影分開；__STOSB／__STOSD只保存實際清除writer與呼叫邊界，不追平台helper。 IDA session77513 exit0，獨立LE／原bytes核對通過；完整原389 journal／日誌／15人口事件保持，69Step變更重建四範圍全部終值。79個新只讀frame及PNG逐張綁定雜湊；原418輸入與SAVE10／MOX實際副本hash／size／UID1000保持。固定日期不是seed，沒有新RNG或正式存檔語意驗證。

Docker內實際入口：
```text
bash workplace/new-game-390-run.sh > workplace/new-game-390-run-output.txt 2>&1
python3 workplace/new-game-390-getter-test.py > workplace/new-game-390-getter-tests.txt 2>&1
bash /out/new-game-390-ida-run.sh > /out/new-game-390-ida-output.txt 2>&1
python3 workplace/new-game-390-byte-verify.py > workplace/new-game-390-byte-tests.txt 2>&1
python3 workplace/new-game-390-verify.py > workplace/new-game-390-verification.txt 2>&1
python3 workplace/new-game-390-state-verify.py > workplace/new-game-390-state-tests.txt 2>&1
python3 workplace/new-game-390-finalize.py
python3 workplace/new-game-390-document-gate.py > workplace/new-game-390-document-gate-tests.txt 2>&1
```
原run.sh第一輪失敗後保留，getter測試與修正版READY後同命令乾淨重跑一次；無第三次guest。初次私有生成及守衛改動用容器Python標準輸入，完整final Go／可逆patches／generator保存，文件gate在記憶體比對generator與final Go相同。byte與原結果／state驗證逐項讀取PASS，沒有把shell最後一個exit當前項成功。

Go1.24.13既有image／network none／UID/GID1000，原ZIP／patch唯讀；native600s／2GiB／2CPU／128pids、capture550s與owned PID trap；驗證90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids。IDA locked-v1既有image，120s／2GiB／2CPU／128pids，原EXE解壓到容器tmp；只輸出私有窄來源，無新image或主機工作負載。Docker專案掛載的容器清單已空，未清理其他專案。

工具574f8e60998bb74c1a5add54e0cd5c362386c9ff已推送github，公開四份自撰文件。主庫只更新一條CONTEXT、唯一DOS活表與追加歷程／證據，公開原資料及private Go未入Git。主庫本輪精確HEAD由Git提交紀錄取得，收尾檢查兩庫乾淨及Docker／檔案擁有權。

391先捕捉選取後下一個原輸入點1B0845，核對原17AABB=1、17A974=4及第二列signed熱區；條件成立才以正常裝置660,107一次按下及安全放開，驗證BF6ED→B9E94與四槽是否恢復。不得直接改bit／派送ID，不把210M中途renderer末態當可按輸入點。 主庫RE-first保持；正式配置／放置／存讀／完整開局及remake同狀態未知。

提交前執行python3 workplace/main-audit-390-final.py，390 MAIN AUDIT PASS：四份文件範圍、53份新與275份舊收據、88份失敗產物、工具已推送且乾淨、公開CPU／DOS保持，root-owned 2437檔／272目錄與零.md目錄保持。git diff --check通過。

## 2026-10-04：391確認選取後kind7與放開來源

上一輪390分類progress，已有實際writer與完整原結果。開工核對AGENTS、CONTEXT、HONEST-STATUS、唯一活表及乾淨兩庫HEAD；路由命中復古GUI、spec閘門、resolution backlink與文件職責，沿逆向及既有IDA9.4技能。主庫基線3fae9a35be452702561d4e41674f38c5f4e916af，工具基線574f8e60998bb74c1a5add54e0cd5c362386c9ff。

初稿以390前的36表／kind6提出放置，READY審查AssertionError由原390的table_count37否定；沒有生成觀察器或新guest。初稿及拒絕保留，回到來源查證，不將審查錯誤歸為CPU缺陷。390原首輸入36表／kind6，210M選取後為37表／kind7，整表2035bytes，SHA-256 d1c13ea6764aab2848addf42cadb7166030d2a5421fd0f8533b9b33773842524。三列signed矩形為310,60..510,88／310,90..510,118／310,120..510,148，pointer沿2879DA／DC／DE。裝置660,107經SAR1成GUI330,107，當次first-hit2。

原11E1EC／11E334／11E503只讓kind6呼叫1192D1，kind7略過held與release場景CALL，仍清共享active；11E582到11E69D，依var_30在11E6D6選正index或11E6E2選負index。原11DB87呼叫123C1B，123C33讀cached word_1B1222，runtime位移2A1222，123C47近返回；390末態cached1與裝置buttons0同時存在，不把cache當新press完成。snapshot的calls取s.calls，源碼證實它是啟動服務計數，不是AX3 poll次數。

兩次窄IDA查詢session65314／16408 exit0，原schema／固定EXE／非空238列與UID1000核對。原MZ／LE、2object／365page／51363fixup records及238列／238EA／27重定位差異獨立通過。SOURCE PASS另核對九個熱區含端點、原y92 first-hit由1成2、原cached1／device0及公開CPU保持。source-only READY後CONFORMED；原390及較早native receipts不改。

工具編排有兩次在工具呼叫前拒絕的JavaScript語法／sh作用域錯誤，修正後重送來源讀取與文件寫入；不影響已驗原資料，沒有guest重跑。既有技能安裝連結的權威IDA入口已於前輪載入，沿locked-v1相同image／UID。

實際Docker入口：
```text
bash /out/new-game-391-ida-run.sh > /out/new-game-391-ida-output.txt 2>&1
bash /out/new-game-391-button-ida-run.sh > /out/new-game-391-button-ida-output.txt 2>&1
python3 workplace/new-game-391-byte-verify.py > workplace/new-game-391-byte-tests.txt 2>&1
python3 workplace/new-game-391-source-verify.py > workplace/new-game-391-source-tests.txt 2>&1
python3 workplace/new-game-391-finalize.py
python3 workplace/new-game-391-document-gate.py > workplace/new-game-391-document-gate-tests.txt 2>&1
```
初稿審查另執行new-game-391-ready-review.py並保留first版與拒絕輸出；原源碼欄位搜索使用容器grep及Python。Go1.24.13／network none／UID1000，patch唯讀；來源核對90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids；IDA既有locked-v1，120s／2GiB／2CPU／128pids，原EXE解壓到容器tmp。無新image、native、輸入或Go實作。專案容器清單為空，保留其他專案。

工具b17eb21c212deb1f95cef54f1399d7b35efd8bdf已推送github，五份自撰來源／回填文件；主庫僅四份現況／歷程文件。主庫本輪精確HEAD見Git提交，交付前核對private receipts／擁有權／Docker。

392先保持完整390至210M，再有界捕捉原1B0845及當次37表／kind7／first-hit2、17AABB=1／17A974=4與安全裝置；正常660,107 press後觀察213C1B依真SS返回20DB8C的低AX1及當次新座標、callback完成、安全IRQ與至少20ms，再release。不得等待kind6 held場景或以calls增加為消費閘門；實際selector／BF6ED→B9E94、record與原畫面另驗。 正式放置／職務、存讀與remake同狀態未驗，RE-first保持。

提交前執行python3 workplace/main-audit-391-final.py，391 MAIN AUDIT PASS：四份文件範圍、27份新與328份舊收據、88份失敗產物、工具已推送且乾淨、公開CPU／DOS保持；kind7／238EA／27重定位差異，沒有新guest。root-owned 2437檔／272目錄與零.md目錄保持。git diff --check通過。

## 2026-10-04：392／393正常kind7放開與配置函式返回

主庫基線5adff55d2b6079458f0cf5ebe271be8f58883201，工具b17eb21c212deb1f95cef54f1399d7b35efd8bdf。沿復古GUI、spec閘門、resolution backlink及文件職責路由與逆向／既有IDA9.4技能。連續core守衛拒絕後重讀GUI入口及實際CPU／callback契約，不調參猜通過。

392首個來源護欄抓到新舊收據都讀390，尚未guest；六份source-rejected產物保存。第一次guest session17614 exit1，210272551正常press後，210272553原callback SS158／ESPFF8僅8-byte合法尾端，原16-byte observer拒絕；175份failed-392及manifest保存。實際getter測試後第二次guest session22864 exit1：210280976／977原1237D9 MOV SS,DX後、1237DB MOV ESP,EAX前，SS158／ESP2A1E2B暫時不可讀；184份failed2-392保存。兩次完整390凍結／正常press已驗，無CPU缺陷證據。原387 bytes、callback4KiB descriptor及實際暫存器反例證實真因，只追必要平台契約。

私有被動core記CodeBytes／StackBytes，不可讀stack標0，padding不當原bytes；strict near槽仍強制4-byte。抽取的實際getter測試、正常8／4／16-byte、unknown0／strict拒絕及RAM／CPU只讀通過，與執行中Go綁定相同。新增增量writer及至多16個unknown stack EIP。兩次建置拒絕分別為局部型別宣告順序及data陣列與JSON同名；兩次都在guest前，輸出保存。回歸fixture縮短RAM時連code移除另修正，非產品缺陷。來源護欄與142CLI均在guest前；腳本TC gate誤把「格／算」納入簡體集合亦已修正。未修改公開CPU／DOS。

392第三次guest session30226 exit0，原215M held窗口只有5phase／22個17AAB9的2／0變更，record末值不變。getter入口候選至20DB8C未觀測，consumed／released／placement_returned均false，392回DRAFT。獨立驗證session10613 exit0，不強造release通過。原selector真SS返20E1AC／EAX2、新GUI330,107、callback20／20、安全IRQ與41,697µs成393來源。393 READY先於新Go；第四次guest session89860 exit0，完整390與392前4phase保持，210282377正常release，210285765／766進BF6ED／B9E94，210305813真SS返BF6F2，212590909下一1B0845提前停止。沒有第五次guest。

393獨立驗證session84691 exit0；五個缺來源實際writer以IDA session52800補證，各八個鄰近指令，殼層exit0但idat_exit1，非空JSON／schema／固定EXE／UID及獨立原bytes均通過，保留內層退出值，不用shell0覆蓋。原LE／2object／365page／51363fixups，83列／53EA／8差異；全部50個runtime writer投影及原bytes核對。最終同一批收據驗證session56695 exit0，15phase／65frame與PNG、13個record差異重建、原state副本保持，無新guest。正式job欄位及正常存讀未驗，固定日期不是seed。

Docker實際入口：
```text
python3 workplace/new-game-392-ready-review.py
python3 workplace/new-game-392-generator.py
bash workplace/new-game-392-run.sh > workplace/new-game-392-run-output.txt 2>&1
python3 workplace/new-game-392-core-test.py > workplace/new-game-392-core-tests.txt 2>&1
python3 workplace/new-game-392-core-binding.py > workplace/new-game-392-core-binding-tests.txt
python3 workplace/new-game-392-failed-verify.py > workplace/new-game-392-failed-tests.txt 2>&1
python3 workplace/new-game-392-failed2-verify.py > workplace/new-game-392-failed2-tests.txt 2>&1
python3 workplace/new-game-392-verify.py > workplace/new-game-392-verification.txt 2>&1
python3 workplace/new-game-393-ready-review.py
python3 workplace/new-game-393-generator.py
bash workplace/new-game-393-run.sh > workplace/new-game-393-run-output.txt 2>&1
python3 workplace/new-game-393-verify.py > workplace/new-game-393-verification.txt 2>&1
bash /out/new-game-393-ida-run.sh > /out/new-game-393-ida-output.txt 2>&1
python3 workplace/new-game-393-byte-verify.py > workplace/new-game-393-byte-tests.txt 2>&1
python3 workplace/new-game-393-finalize.py
python3 workplace/new-game-393-document-gate.py > workplace/new-game-393-document-gate-tests.txt 2>&1
```
每個退出值／PASS逐項核對。392初兩次native沿600s／state550s；依測得210M約10分鐘，392第三次及393外層改900s／state850s涵蓋新增5M，指令上限仍215M，未改虛擬時序或重擲。所有native用既有Go1.24.13 image、2GiB／2CPU／128pids、UID/GID1000／network none，原ZIP／patch唯讀，owned PID trap。來源／獨立驗證90s／2GiB／1CPU，文件30s／512MiB／1CPU／64pids；IDA locked-v1既有image120s／2GiB／2CPU／128pids。無新image或主機工作負載，所有clone掛載一次性容器已結束。

工具ddaf4d80291eb33e759fb01695018c3960786b7a已推送github，僅十份自撰文件，八個舊規格回填與索引保持；393限定CONFORMED、392拒絕候選DRAFT、385舊觀察器DRAFT保持。主庫僅一條CONTEXT、唯一DOS活表及追加WORKLOG／研究紀錄；精確本輪HEAD見Git。交付前核對收據、原失敗產物、兩庫／Docker與root-owned基線。下一步394先核對原BA5DA的slot word職務位元及其實際consumer，保持原record定位／raw bytes／推論等級；再判斷正常存讀的最小原玩家路徑。不撰寫主庫玩法規格或改Go行為，RE-first仍待全部玩法證據閉合及使用者確認。

提交前python3 workplace/main-audit-393-final.py通過：四份文件精確範圍、92份新／355份舊收據、88／175／184份失敗產物保持；工具已推送乾淨，公開CPU／DOS不變。50個實際Step變更／13個record差異／83列53EA8fixup及正常下一輸入點通過，root-owned 2437檔／272目錄、零.md目錄保持。git diff --check通過。


## 2026-10-04：394人口職務位元與產出consumer

主庫基線67089eac86ea8fd26a567c45fdb7d3aa5a4db6a1、工具ddaf4d80291eb33e759fb01695018c3960786b7a，兩庫乾淨。重讀AGENTS、CONTEXT、HONEST-STATUS與活表；路由命中原版GUI、IDA位址證據與文件職責，載入retro-gui-restoration、逆向技能evidence-and-re／spec閘門及既有IDA9.4技能／工具入口。主庫玩法RE-first保持。

兩次窄IDA查BA5DA正常寫入邊界88指令、DE280篩選／輸入／迴圈、三產出caller與存讀入口caller。session60382及49346殼層exit0，idat_exit1均保存；最小非空schema／5365函式／固定EXE／UID1000與獨立原bytes通過。首次bytes核對session21065通過181列／179EA／7fixup差異；新增查詢後session71968通過300列／294EA／9fixup差異。16個既有別版原檔錨點逐bytes相符，不以相同EA宣稱兩版全部相同。

收據SOURCE + ACTUAL SLOT CONSUMER PASS：BA6DF清職務第7、8位，BA6E8回寫，BA6EF設第9位，DE393以第9位篩選產出。三caller送job0／1／2。原八槽保持，農夫／工人／科學家4／2／2→選取暫停0／2／2→配置0／6／2；八次實際writer與完整393全部50變更重建通過。文件DOCUMENT GATE PASS；job3、跨殖民地、正常存讀及remake同狀態仍未知。802CC存檔／802C2讀檔／81136主選單讀檔只核對call邊界，switch case2／3列強推論，不派送ID。

實際Docker入口：

```text
bash /out/new-game-394-ida-run.sh > /out/new-game-394-ida-output.txt 2>&1
bash /out/new-game-394-binding-ida-run.sh > /out/new-game-394-binding-ida-output.txt 2>&1
python3 workplace/new-game-394-byte-verify.py > workplace/new-game-394-byte-tests.txt 2>&1
python3 workplace/dosgolem/workplace/new-game-394-verify.py > workplace/dosgolem/workplace/new-game-394-verification.txt 2>&1
python3 workplace/new-game-394-document-gate.py > workplace/new-game-394-document-gate-tests.txt 2>&1
```
一次唯讀schema檢查錯取390 terminal的ranges而KeyError；390實際使用record，修正欄位後通過，屬讀取腳本問題。一次文件編排JavaScript因Markdown fence在工具呼叫前SyntaxError，未執行編輯；改用字元建立fence後完成。兩者均未重跑guest，不歸為原CPU缺陷。

原patch唯讀，Go1.24.13／IDA locked-v1既有image，UID/GID1000、network none；IDA120s／2GiB／2CPU／128pids，bytes核對90s／2GiB／1CPU，收據及文件核對30s／512MiB／1CPU／64pids。無新image、guest、裝置輸入或Go改動。工具a11095c650962323492f7cab4e4603fdf62f9084已推送github，五份自撰RE／索引／歷史追加；主庫只更新四份現況／歷程文件，原EXE／JSON／PNG／LOG／private scripts保持忽略。精確主庫HEAD見提交紀錄，收尾另驗工作樹、擁有權及Docker清理。395先查COLONIES返回與原options讀存控件，未達READY不送存檔事件。

提交前執行python3 workplace/main-audit-394-final.py，session34012 exit0，394 MAIN AUDIT PASS：四文件／唯一DOS活表／追加歷史、20新收據hash與5個固定原版輸入、工具已推送且乾淨、公開CPU／DOS保持、8個writer／八槽與300列294EA9fixup核對通過。root-owned 2437檔／272目錄與零.md目錄保持。專案Docker兩個volume filter皆空，沒有新image或遺留容器；git diff --check通過。


## 2026-10-04：395原存讀來源與397正常RETURN

基線主庫dfb9676ebb1435ce2927bbaec3730a9df9572738、工具a11095c650962323492f7cab4e4603fdf62f9084。路由命中原GUI、IDA、規格閘門、文件職責及逆向回鏈；沿已載入契約。395原來源436列／358EA／117fixup、跳表及九個熱區通過，394補證回鏈。主庫RE-first與Go玩法／公開CPU／DOS保持。

首次396 session19733 exit1／overlay_probe_exit137，未送RETURN，固定393末態外部完整比對相同。239份178985623-byte產物與manifest保存，hash通過。48MB JSON同時展開造成記憶體壓力為強推論，Docker未查得OOM事件；未當作CB或原CPU缺陷。READY窄改串流守衛，full393及三種非雜湊突變拒絕通過；154CLI含128拒絕／26正對照、精確反轉固定393、原getter與唯一Step保持。同入口第二次原guest session31572完整驗證原press與selector真返回，7357µs即時放開守衛拒絕，維持held至215M；未宣稱返回完成。397依固定396前五phase，首次安全20ms才正常release，原guest session63484，900s／state850s／2GiB／2CPU／128pids及原215M保持。

397在212606147正常release，elapsed47384µs；213103590原C058A依真SS返回1004EF，ESP+4及相同SS。16phase／PNG與按下／放開前後RAM／core相同通過，完整393的15phase／50writer／65PNG及固定396前五phase保持。原215M末態1A5051仍顯示殖民地，2071AB下一輸入沒有觀測；畫面切換、正常存讀及remake同狀態未驗。418原輸入及SAVE10／MOX副本保持，固定日期不是seed。

隔離執行入口及已執行命令摘要：

```text
bash /out/new-game-395-ida-run.sh
bash /out/new-game-395-switch-ida-run.sh
python3 workplace/new-game-395-byte-verify.py
python3 workplace/new-game-395-source-verify.py
python3 workplace/new-game-395-document-gate.py
python3 workplace/new-game-396-generator.py
python3 workplace/new-game-396-source-verify.py
go run /tmp/396-stream-test.go
bash workplace/new-game-396-run.sh > workplace/new-game-396-run-output.txt 2>&1
python3 workplace/new-game-396-verify.py > workplace/new-game-396-verification.txt 2>&1
python3 workplace/new-game-397-generator.py
python3 workplace/new-game-397-source-verify.py
bash workplace/new-game-397-run.sh > workplace/new-game-397-run-output.txt 2>&1
python3 workplace/new-game-397-verify.py > workplace/new-game-397-verification.txt 2>&1
```

串流測試先以inline Python取出實際helper產生/tmp/396-stream-test.go，後保存同內容workplace/new-game-396-stream-test.py供重生。395首輪IDA後處理raw target缺meta誤拒KeyError，核對既有非空JSON通過，不重跑IDA；第二次IDA殼層0／idat_exit1保存。唯讀查詢meta=None及容器內rg缺件、文件編排JavaScript SyntaxError，397首次CLI正對照誤設缺檔退出值1而原panic為2，在guest前拒絕；修正後session38119又被私有全字串改名污染固定值35039662的選族守衛拒絕，45份10398299-byte產物保存。改名限縮為檔名前綴，新增精確基底及數值保持守衛，session63484同入口乾淨重跑。以上皆屬腳本／環境；修正查詢或編排後繼續，未歸為產品缺陷。

既有Go1.24.13／IDA locked-v1，UID/GID1000、network none，原ZIP／patch唯讀。IDA120s／2GiB／2CPU／128pids，串流及驗證90s／2GiB／1CPU，來源與文件30s／512MiB／1CPU／64pids。沒有新image或主機工作負載。工具ba7dd46630aee381f0a15051fbd103ef9c0c3ea4已推送，公開僅自撰規格及回鏈。收尾稽核另記Git／Docker／擁有權；398先核對1004EF返回後控制流及1A5051等待，正常存讀及remake同狀態未驗。

提交前python3 workplace/main-audit-397-final.py，session19055 exit0，397 MAIN AUDIT PASS：四份自撰文件、唯一DOS活表、追加歷史、63份新收據hash與固定舊輸入、239／45份失敗產物保持，工具ba7dd46630aee381f0a15051fbd103ef9c0c3ea4已推送且乾淨；公開CPU／DOS與主庫Go玩法保持。原47384µs正常放開、1004EF真返回及436列358EA117fixup核對通過；原215M末態PNG SHA-256 741bb4ca62f46653dc9c65d0b3d6639b93cbf9c8b71758e3864988105de89ee1視覺確認仍為殖民地。root-owned 2437檔／272目錄及零.md目錄保持；git diff --check通過。本輪一次性容器全數結束，沒有新image，精確主庫HEAD與最終工作樹／Docker狀態於Git提交後另核對。


## 2026-10-04：398–400原RETURN分派20與父入口來源

基線主庫5d4595f69b3ad4d3a5c94fde6bfec8fd92853c8b、工具ba7dd46630aee381f0a15051fbd103ef9c0c3ea4。路由命中原GUI、IDA、規格閘門、文件職責與逆向回鏈，沿用已載入契約。398兩次窄IDA共320列／298EA／84fixup差異；原103EB..1049B完整44項與AX≤2Bh相符，從固定EXE及44個LE fixup獨立重建，395前43項保持、raw43→106AF。395先前43項是前綴，現行入口已修正，原收據不覆寫。

原輸入ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。工具IDA Pro 9.4／Go1.24.13／Python3.11；原EA使用IDA linear EA，runtime投影為EA+F0000h，原file offset與LE fixup另列，不混用位址基準。

399沿READY只讀觀察器、精確檿名改名及固定397基底；175CLI含145拒絕／30正對照、唯一原Step／原getter／裝置輸入數保持。原guest session76282 exit0，215M cap不改；11個新只讀phase的core／RAM／device前後及PNG、原MOV bytes與唯一Step差異通過，完整397的16phase／terminal逐欄位保持，僅略既定三個RAM雜湊鍵。原418輸入及SAVE10／MOX副本保持，固定日期不是seed。

已證實：原C093A runtime1B093A於213103514→515把191A08從1恢復20；原104EF runtime1004EF於213103590→591把191F19從20設1。原runtime1006A7在213103599的EAX20。完整44項原表raw20→1050C CALL C4562。強推論：下一父分支為C4562；399未直接取樣該入口，不宣稱實際entry已驗。原215M仍1A5051，PNG SHA-256 741bb4ca62f46653dc9c65d0b3d6639b93cbf9c8b71758e3864988105de89ee1與397相同。

400兩窄IDA殼層exit0／idat_exit1，非空JSON／5365函式／固定SHA／UID1000通過；原404列／229EA／113fixup差異與2object／365page／51363原fixup records核對通過。完整135指令sub_C4562 C4562..C47C1保存原輸入／返回邊界：C472A CALL1171AB、C4740 CALLC4343、C4792回輸入迴圈及C47C0 RET。原C541C CALLB4EF6位於C53C9，direct callerC54D6；只保存caller定位，不猜未取樣call chain。398證明B5051是槽位篩選，不稱輸入等待。400無新guest、Go修改或裝置輸入。

原GUI／下一輸入、正常存讀及remake同狀態未驗。主庫RE-first保持；公開CPU／DOS及主庫玩法程式不改，未把未知當CB指令缺陷。工具959e108084ad465389e9c7470a7c3408f84456bf已推送github，六份自撰RE／索引／回鏈；原JSON／EXE／PNG／LOG／private Go保持本機忽略。下一401先建最小只讀READY契約，保持完整399／397及215M，捕捉原1050C／C4562入口、B4EF6真SS caller與C472A／1171AB；不新增裝置輸入、代寫RAM或延長cap。

實際Docker命令：

```text
bash /out/new-game-398-ida-run.sh
bash /out/new-game-398-mode-ida-run.sh
python3 workplace/new-game-398-byte-verify.py
python3 workplace/new-game-398-table-verify.py
python3 workplace/new-game-399-generator.py
python3 workplace/new-game-399-source-verify.py
bash workplace/new-game-399-run.sh > workplace/new-game-399-run-output.txt 2>&1
python3 workplace/new-game-399-verify.py > workplace/new-game-399-verify-tests.txt
bash /out/new-game-400-ida-run.sh
bash /out/new-game-400-flow-ida-run.sh
python3 workplace/new-game-400-byte-verify.py
python3 workplace/new-game-400-source-verify.py
python3 workplace/new-game-400-document-gate.py
```

既有Go1.24.13與IDA9.4 locked-v1，UID/GID1000、network none、原ZIP／patch唯讀。IDA120s／2GiB／2CPU／128pids，原guest900s／state850s／2GiB／2CPU／128pids及owned PID trap，bytes90s／2GiB／1CPU，來源／文件30s／512MiB／1CPU／64pids。兩次讀取查詢使用錯誤容器路徑，及容器缺rg已改用正確/knowledge路徑與grep；399來源映射補驗及400第二次窄查詢均使用既有收據，沒有重跑399 guest。以上是讀取環境問題，未列為產品缺陷。Docker兩個專案volume filter皆空，沒有新image或遺留容器。提交前再核對工作樹、證據hash、權利分類與擁有權。

提交前執行python3 workplace/main-audit-400-final.py，exit0，400 MAIN AUDIT PASS：四份自撰文件、唯一DOS活表、追加歷史、39新收據hash及固定舊輸入、工具959e108084ad465389e9c7470a7c3408f84456bf已推送且乾淨；公開CPU／DOS與主庫玩法保持。完整44項原表、398的320列及400的404列來源、399只讀raw20與完整397保持通過。實際父入口、下一輸入、正常存讀與remake同狀態仍未知。root-owned 2437檔／272目錄及零.md目錄保持，git diff --check通過；本輪容器均已結束，無新image。精確主庫HEAD與最終工作樹／Docker狀態在提交後另核對。


## 2026-10-04：401–402原父層輸入與殖民地列表自然返回

基線主庫a18897eb5e6421726ec05a4bb5e30c122280b2f2、工具959e108084ad465389e9c7470a7c3408f84456bf；接手時兩庫乾淨。路由命中復古遊戲RE、原GUI、IDA、READY閘門、文件職責與逆向回鏈，沿用已載入契約。

401新增一份窄IDA，43列／43EA／11fixup差異與原MZ／LE、2object／365page／51363fixup records獨立核對。原C2C5A CALLB4EF6、下一C2C5F及sub_C2B72 owner保留原EA；C2259初始化只保存邊界，不猜返回時刻。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。工具Go1.24.13／Python3.11／IDA9.4 locked-v1；IDA linear EA、runtime＝EA+F0000h及原file offset分列。

401沿READY只讀觀察器與固定399 Go，原guest session93429 exit0、185CLI含153拒絕／32正對照，19phase全只讀；獨立驗證session80541 exit0。已證實：213103600原1050C CALL後，213103601真SS到C4562，同CS8／SS188h、SP−4與返回槽runtime100511通過。B4EF6有12次entry、10次RET1Ch自然返回、11823次B5051命中，所有實際RET的唯一下一Step／真SS／ESP+32通過。兩個active frame在215M保持，不把末態篩選EIP稱停死或輸入等待。原caller target為runtime1B2C5F及遞迴1A4F44，不從C53C9靜態caller補虛構call chain。

402經DRAFT→READY審查，精確反轉固定401的20個patch，唯一原CPU.Step／getter／裝置輸入數保持。195CLI含161拒絕／34正對照，保留185前綴，exact copied build input與全部新輸出名稱守衛通過。原guest session23489 exit0，獨立驗證session44403 exit0；先凍結並逐欄位保持完整401 parent／399 mode／397 return的215M前置，僅略既定三個RAM雜湊鍵，才進230M有界窗口。

已證實：222329888在runtime1B472A原CALL，222329889自然到2071AB父層正常輸入並提前停止；同CS8／SS188h、SP−4及真返回槽1B472F通過，4個新phase與原EXE／LE重定位16-byte code window獨立核對。原控件表runtime0x298848／count20／stride55，VBE StartY0／DisplaySets102。實際原PNG SHA-256 299868824b0ea14741bbcd781a8aa353c8344ef29bd5997f50f5be6d211de8c2已回COLONIES列表，Sol II可見農夫欄空、6工人／2科學家。人工圖像檢視另存new-game-402-visual-review.json；數值verifier的visual_transition_verified仍false，未假稱程式自動視覺驗證。只驗原版自身正常返回，remake同狀態及正常存讀仍未驗。原record+C8正式名稱仍未定型，不因畫面同數值就命名。

同一402末態1100原bytes的表解碼，raw3／kind0矩形531,445..615,470，GUI590,468→裝置1180,468，九個含端點／外側first-hit案例通過。強推論：raw3幾何對應原PNG RETURN；實際選取返回值尚未驗。舊18現為378,34..510,64上方欄位。下一403先核對C4343 handler及當次17B0EE／17B0F0與1985AC，再建READY正常press／release契約；正常星圖返回後才追options／存讀。不直接派送ID、代寫RAM或深挖renderer／DAC／PIT／平台helper。

401第一次runner仍編譯399 Go，session51344 exit1，parent_value CLI在原guest前拒絕；failed1-401三份產物及manifest保留。修正明確Go檿名與copied build input守衛後185CLI通過。成功401 guest執行中發現native／state輸出仍用399名稱，保留原399 gzip與manifest，未中止或重跑；terminal後核對實際401 marker／source，另存401日誌與manifest、還原原399 gzip／raw／manifest。實際執行script另存new-game-401-executed-run.sh，現行乾淨重生runner全名已修正。原399 gzip SHA-256 ba5e0cbf4568c8677bfbd501f51c206ff7c96a41b6bd197ab7faa1c234177398及39份較早400索引逐hash保持。這是runner接線問題，無原CPU缺陷證據；獨立verifier執行前核對cpu.go的ESP4／SegCS0／SegSS5。

原418輸入、SAVE10／MOX副本與394固定較早收據保持；固定日期不是seed。工具9c588a41508e038d5654903c1cb04a830875bd32已推送github，七份自撰RE／索引／回鏈文件；EXE／原PNG／JSON／private Go不公開。主庫RE-first與玩法、公開CPU／DOS保持，本輪分類progress。

實際Docker命令：

```text
bash /out/new-game-401-ida-run.sh
python3 workplace/new-game-401-byte-verify.py
python3 workplace/new-game-401-ready-review.py
python3 workplace/new-game-401-generator.py
bash workplace/new-game-401-run.sh > workplace/new-game-401-run-output.txt 2>&1
python3 workplace/new-game-401-source-verify.py
python3 workplace/new-game-401-verify.py > workplace/new-game-401-verify-tests.txt
python3 workplace/new-game-402-ready-review.py
python3 workplace/new-game-402-generator.py
bash workplace/new-game-402-run.sh > workplace/new-game-402-run-output.txt 2>&1
python3 workplace/new-game-402-source-verify.py
python3 workplace/new-game-402-verify.py > workplace/new-game-402-verify-tests.txt
python3 workplace/new-game-402-table-verify.py > workplace/new-game-402-table-tests.txt
python3 workplace/new-game-402-document-gate.py > workplace/new-game-402-document-gate-tests.txt
```

既有Go1.24-bookworm及IDA9.4 locked-v1，UID/GID1000、network none、原ZIP／patch唯讀。原guest900s／state850s／2GiB／2CPU／128pids及owned PID trap；IDA120s／2GiB／2CPU／128pids，bytes90s／2GiB／1CPU；來源／文件30s／512MiB／1CPU／64pids。401與402原guest均已terminal，沒有持續程序；Docker兩個專案volume filter皆空，沒有新image。提交前核對工作樹、hash、忽略／權利分類與擁有權。原較早root-owned基線2437檔／272目錄保持，沒有.md目錄，不做廣域chown。

收尾稽核：容器內python3 workplace/402-main-audit.py通過。四份自撰文件、唯一DOS活表與next403、53份新收據及39份舊收據hash、工具乾淨且已推送、玩法／CPU／DOS保持、UID1000與root-owned基線均通過。git diff --check通過；兩個Docker專案volume filter無容器。


## 2026-10-04：403–404正常列表RETURN與星圖返回

基線主庫811cd9303cf5e6e83882e480cae8023a322133b9、工具9c588a41508e038d5654903c1cb04a830875bd32；接手兩庫乾淨。上一輪分類progress，本輪沿復古遊戲RE／原GUI／IDA／READY閘門／文件職責／逆向回鏈入口。正常玩家路徑持續前進，主庫玩法RE-first保持。

403兩窄IDA238列／209EA／70fixup差異，原MZ／LE、2object／365page／51363fixup records及完整121指令sub_C4343 C4343..C4562逐原EXE核對。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具Go1.24.13／Python3.11／IDA9.4 locked-v1。IDA linear EA、runtime＝EA+F0000h及原file offset分列。C2F10 CALL1151B0後C2F15將AX寫17B0EE，C2F25 CALL114C72後C2F2A寫17B0F0；C43DD／C43E6原consumer比較及C4428／C442E／C4434的1985AC→191A08／191A10保存原定位。只追控件呼叫邊界，不深挖美術／renderer或平台helper。

404經DRAFT→READY審查，8個精確可反轉patch至固定402 Go f86a2e5e97d36a0abc42f14b10de00dc91f78ad884961f2385d53f7a793b5ad4；205CLI含169拒絕／36正對照、195前綴、唯一原Step／原getter與兩個新增正常裝置呼叫保持。copied build input及新404全部輸出名稱守衛通過。直到222329889原1171AB首輸入，依原順序凍結並逐欄位保持完整402 continue-terminal及401 parent／399 mode／397 return，僅略既定三個RAM雜湊鍵，才讀當次UI word並送新輸入。

修正版原guest session82189 exit0，獨立原EXE／LE重定位及數值verifier exit0。已證實：當次UI words1／2／3／4，17B0EE＝3、17B0F0＝4、1985AC＝0；222329889正常press1180,468,1。222335412原113FB9 entry，222335801真SS返回20E1AC／SP+4／EAX3、cached GUI590,468；222340506正常release1180,468,0，elapsed42556µs、callback與IRQ安全，兩次裝置事件不直接改CPU／RAM。42.556ms使用原虛擬時鐘，不宣稱實機wall-clock。

222805471到C472F，222806140到C4343，222806153到C43DD。222806164原C442E將191A08恢復0，222806165的C4434把191A10恢復0，222806166到C443A同為0。223264812於原C47C0 RET，223264813唯一下一Step自然回100511，同CS8／SS188h與ESP+4通過。225305800到下一原1171AB正常輸入並提前停止，低於230M既有上限，不再延長。16個新phase全只讀，全部原16-byte code window／core／device／四raw範圍／UI word／mode及PNG逐bytes核對。

實際原PNG SHA-256 8b9854fca02675829dab522a22bcf7f812fcf255aa4856854630fde357a3dd4c已回星圖，Sol、3500.0、頂端GAME及下方COLONIES等可辨讀。VBE StartY512／DisplaySets105，23×55＝1265原控件bytes。人工畫面收據new-game-404-visual-review.json與數值verifier分開；new-game-404-result.json的visual_transition_verified仍false，不假稱自動視覺比對。只驗原版正常玩家路徑，正常存讀與remake同狀態仍未驗；word_17B0F0＝4的讀值已證，鍵盤替代入口未實際按下。C2259返回時間未取樣，已知正常父層返回足以繼續玩家路徑，不重新深挖初始化helper。

同一404末態的raw6／kind0矩形249,5..307,21、GUI280,13→裝置560,13，九個含端點／外側first-hit案例通過；強推論為頂端GAME候選。當次raw3已是不可見kind8項，不沿用列表RETURN3。下一405先核對星圖GAME handler及原mode8→8012F來源，再建立READY正常點擊與選項觀察契約；正常存讀另驗，不直接派送ID／mode8或寫RAM。

首次私有build session78343 exit1，生成器把換行寫進Go literal，CLI／guest尚未開始，failed1-404五份產物及manifest保留。七個Go literal與兩個新Python CLI literal跳脫已修正，全部Python heredoc AST、bash -n與建置檢查通過後，同容器／同命令乾淨重跑；修復腳本初次literal數核對在寫入前拒絕，核對實際七處後完成。沒有失敗guest；這是私有生成器問題，無原CPU缺陷證據。

原418輸入、SAVE10／MOX副本、53份較早402及39份400收據hash保持，固定日期不是seed。工具b25e49f74c21bbc0abbfaa36dee5b390683c394a已推送github，九份自撰RE／索引／解決回鏈；EXE／原PNG／JSON／private Go均本機忽略。主庫玩法及公開CPU／DOS保持，本輪分類progress。

實際Docker命令：

```text
bash /out/new-game-403-ida-run.sh
bash /out/new-game-403-producer-ida-run.sh
python3 workplace/new-game-403-byte-verify.py
python3 workplace/new-game-403-source-verify.py
python3 workplace/new-game-404-ready-review.py
python3 workplace/new-game-404-generator.py
bash -n workplace/new-game-404-run.sh
bash workplace/new-game-404-run.sh > workplace/new-game-404-run-output.txt 2>&1
python3 workplace/new-game-404-source-verify.py
python3 workplace/new-game-404-verify.py > workplace/new-game-404-verify-tests.txt
python3 workplace/new-game-404-table-verify.py > workplace/new-game-404-table-tests.txt
python3 workplace/new-game-404-document-gate.py > workplace/new-game-404-document-gate-tests.txt
```

既有Go1.24-bookworm及IDA9.4 locked-v1，UID/GID1000、network none、原ZIP／patch唯讀。原guest900s／state850s／2GiB／2CPU／128pids及owned PID trap；IDA120s／2GiB／2CPU／128pids，bytes90s／2GiB／1CPU；來源／文件30s／512MiB／1CPU／64pids。兩次IDA、私有build及唯一新guest皆terminal；Docker兩個專案volume filter無容器，沒有新image。提交前核對工作樹、hash、忽略／權利分類與擁有權。

收尾稽核：容器內python3 workplace/404-main-audit.py通過。四份自撰文件、唯一DOS活表與next405、45份新收據、53份較早402與39份400收據hash、工具乾淨且已推送、玩法／CPU／DOS保持均通過。UID1000與root-owned基線2437檔／272目錄保持，沒有.md目錄；git diff --check通過，Docker專案volume filter無容器。

## 2026-10-05：原GAME正常輸入、選單輸入讀取端來源與外層作用域

工具分支codex/moo2-parity-20260930，HEAD de7456f0d5f9f5c01232a7b0028c8488e0a998de已推送github。主庫本輪只更新四份自撰文件，玩法RE-first保持；原CPU／DOS與公開probe相對a03c322d28002bb0e11d5d5109d5833a17114d04無差異。來源405／407／408分別597／515／66列、475／451／57個EA、187／121／15列原file bytes與IDA重定位bytes不同。原1.31 EXE hash與2object／365page／51363fixup records獨立重建，原始名稱、位址基準與bytes保留。7D061是控件建立，7DD77→1171AB才是mode0正常輸入讀取端，原定位名稱的分類已回填。

406三輪皆殼層exit1：7418在新press前被原34／契約0拒絕；61174在210420000後SIGKILL137，舊OOM旗標未知；10879在正常GAME press／selector6／release後，將另一SP位置的共用83D05套入外層返回位址檢查而拒絕。三輪318／321／333份產物與manifest保存；第二輪較晚檔明示為殘留。source405首查起點87643與407驗證器BX／DX編碼錯誤各保存失敗，重核同一原source。沒有原CPU缺陷證據。

已驗：完整404／402／401／399／397凍結、215CLI含177拒絕／38正對照及8個可反轉patch。225305800正常裝置press560,13,1，225315546原113FB9真SS返回20E1AC／EAX6／GUI280,13，225323360安全release、49572µs。七個新phase的core／device／RAM只讀、原Code16／LE fixups與PNG hash核對。當次三word6／0／34，原418檔與SAVE10／MOX保持。limited verifier exit0，實際原PNG人工檢視另存：仍為星圖，游標在GAME，沒有選單。406完整契約仍DRAFT，不稱外層返回、選單輸入、正式存讀或remake同狀態已驗。

408證實ENTER6CC與EBP減82；原首輸入ESP2BD4F8／EBP2BDB46算外層RET槽2BDBE0。226846736的tail ESP2BD378異於外層保存暫存器2BDBCC。另一活動框架為強推論；實際caller與精確停止step未知。409經DRAFT→來源／只讀審查→READY，尚無409 Go或guest；下一步凍結七phase，只讀續行、無新裝置輸入、不預填1004BC，原230M上限不延長。

第三輪3GiB／GOMEMLIMIT1GiB下cgroup峰值2119880704bytes，oom／oom_kill增量0。Go1.24.13官方runtime/extern.go的工具軟上限契約與hash在resource收據；未把第二輪SIGKILL回填成確診OOM。生成器在容器暫存區重建Go／patches／runner／source verifier／state capture／resource capture，六份bytes一致。只有三次原guest，沒有為畫面或檔名重跑。

實際容器內命令：

```text
python3 workplace/new-game-405-byte-verify.py
python3 workplace/new-game-405-source-verify.py
python3 workplace/new-game-406-generator.py
bash workplace/new-game-406-run.sh > workplace/new-game-406-run-output.txt 2>&1
python3 workplace/new-game-406-prepress-verify.py
python3 workplace/new-game-406-limited-verify.py > workplace/new-game-406-limited-tests.txt
bash workplace/new-game-407-ida-run.sh
bash workplace/new-game-407-mode0-ida-run.sh
python3 workplace/new-game-407-byte-verify.py
python3 workplace/new-game-407-source-verify.py
bash workplace/new-game-408-ida-run.sh
python3 workplace/new-game-408-byte-verify.py
python3 workplace/new-game-408-source-verify.py
python3 workplace/new-game-408-document-gate.py
```

405的四個IDA入口詳見source405，本輪每個原JSON與殼層輸出均存hash。沿既有Go1.24-bookworm及IDA9.4 locked-v1，UID/GID1000、network none、原ZIP／patch唯讀。guest900s／state850s／2CPU／128pids與owned PID trap；IDA120s／2GiB、bytes90s／2GiB，來源／文件30s／512MiB。所有guest、IDA及驗證handle terminal；Docker兩個專案volume filter無容器，沒有新image。110份新收據由workplace/408-current-receipt-index.json保存，較早45／53／39份hash保持。原資料／私有Go與所有收據Git忽略。

收尾稽核：容器內python3 workplace/408-main-audit.py通過。四份自撰文件、唯一DOS活表與next409、110份新收據與較早凍結收據、972份失敗原guest產物hash、工具乾淨且已推送均核對。玩法與公開CPU／DOS保持；406為DRAFT，409為READY且尚未實作。UID1000與root-owned基線2437檔／272目錄保持，沒有.md目錄。git diff --check通過，Docker專案volume filter無容器。

## 2026-10-05：原GAME真外層返回與230M末態

工具分支codex/moo2-parity-20260930，HEAD ee2602e8f9f7292548036670bc300224076f2b07已推送github。主庫本輪只改四份自撰文件，RE-first保持；公開CPU／DOS與probe相對a03c322d28002bb0e11d5d5109d5833a17114d04保持。409私有Go從固定406重生，7個反轉patch、唯一Step、原getter與零新裝置呼叫通過；225CLI含185拒絕／40正對照及完整215前綴。mode-off保持409以前的406拒絕契約，不改玩法。

唯一原guest session79227 exit0，以原230M上限正常停止。完整406七phase與404／402／401／399／397保持，15個新phase的core／device／RAM／原Code16、LE fixups與PNG hash獨立核對。226846742較低共用RET的trueSS target174BC9，227146859首輸入17651B返回EAX6，原mode writer寫0／8／0；227148164外層RET槽2BDBE0當次target1004BC、唯一下一Step同CS／SS及ESP＋4通過。227148175原主CALL8012F、下一Step真SS return1004AB與ESP−4通過。

230M末態runtime21F7C1、EBP2BD934／ESP2BD888，仍未到控件建立或正常reader。實際原PNG人工檢視仍為星圖，沒有GAME選單；人工收據與numeric verifier分開。409仍DRAFT，不以有限外層成功縮小完整選單契約。原418檔與SAVE10／MOX保持；cgroup峰值1479024640bytes，oom／oom_kill增量0。六份生成器產物於Docker暫存區逐bytes重生一致，沒有為畫面或檔名重跑。

410兩次最小IDA查詢678列／609EA／89fixup差異，原2object／365page／51363原fixup獨立核對。較低原84BC4 CALL87BAE與84BC9定位核對；230M原12F7C1在sub_12F578的byte-copy循環，入口及RET框架保存。只保留選單static dependency80211→7EDF2→12F578，不用static caller宣稱當次caller。原helper RET槽算2BD948，實際target仍未知；不深入圖像格式、driver或runtime。411經DRAFT→來源／只讀審查→READY，尚無411 Go或guest；下一步先凍結完整230M末態，再開新有界240M只讀觀察，不改409的230M收據或上限。

實際容器內命令：

```text
python3 workplace/new-game-409-generator.py
bash workplace/new-game-409-run.sh > workplace/new-game-409-run-output.txt 2>&1
python3 workplace/new-game-409-source-verify.py
python3 workplace/new-game-409-verify.py > workplace/new-game-409-verify-tests.txt 2>&1
bash new-game-410-ida-run.sh > new-game-410-ida-run-output.txt 2>&1
bash new-game-410-epilog-ida-run.sh > new-game-410-epilog-ida-run-output.txt 2>&1
python3 workplace/new-game-410-byte-verify.py
python3 workplace/new-game-410-source-verify.py
python3 workplace/new-game-411-document-gate.py
```

沿既有Go1.24.13 image與IDA9.4 locked-v1、UID/GID1000、network none、原ZIP／patch唯讀。guest900s／state850s／3GiB／2CPU／128pids、GOMEMLIMIT1GiB、owned PID trap與cgroup收據保持；IDA120s／2GiB／2CPU／128pids，bytes90s／2GiB／1CPU，其他來源／文件30s／512MiB／1CPU。兩IDA殼層exit0／idat_exit1，固定原hash、非空JSON與5365函式通過；所有handle terminal，Docker專案volume filter無容器，沒有新image。371份本輪私有收據由workplace/411-current-receipt-index.json索引，原資料與私有Go／PNG／JSON／LOG忽略。

收尾稽核：容器內python3 workplace/411-main-audit.py通過。四份自撰文件、唯一DOS活表next411、追加式歷史、371份新與先前110／45／53／39份凍結收據、972份較早失敗原guest產物hash、工具乾淨且已推送均核對。原mode writer的立即數、目的位址、唯一下一Step及實際0／8／0值獨立核對；409保持DRAFT、410限定來源CONFORMED、411 READY且尚無Go或guest。root-owned基線2437檔／272目錄保持，沒有.md目錄，git diff --check通過；Docker三個專案volume filter無容器。首次稽核器誤去除git狀態首欄空白，兩份失敗腳本／輸出保存後修正解析，重核同一收據；沒有重跑原guest。

## 2026-10-05：原GAME面板與第一正常輸入入口

工具分支codex/moo2-parity-20260930，HEAD 193390f1d70ba1cd34c784a18734e142d2fc378e已推送github。主庫只更新四份自撰文件，玩法RE-first保持；公開CPU／DOS及probe相對a03c322d28002bb0e11d5d5109d5833a17114d04無差異。411私有觀察器的11個精確反轉patch、235CLI含193拒絕／42正對照、唯一Step／原getter與零新裝置呼叫通過，六份生成器產物逐bytes重生一致。

唯一guest session60197殼層exit0，在236253170原2071AB第一正常GAME輸入點提早停止。完整409的230M terminal與15phase、406七phase與404／402／401／399／397全部保持。12個新只讀phase的core／device／RAM／原Code16與LE fixups獨立核對。230237065原helper RET槽2BD948實際讀出16EE5E，下一Step同CS／SS及ESP＋4通過；原8028F CALL7D061、7D891 RET、802AE CALL7DD41與7DD77 CALL1171AB全部握手通過。236253170真SS return16DD7C。

原實際11項／605bytes控件表、物件+28h／2Ah／30h／32h的1／3／5／6均保存。原PNG人工檢視可見GAME面板及SAVE GAME／LOAD GAME／RETURN，收據與數值驗證分開。411限定正常原選單入口CONFORMED；409舊230M完整契約仍DRAFT，不把較晚續行寫回舊原末態。正常存讀與remake同狀態仍未驗，固定日期不是RNG seed。

412從既有395／407及411原收據核對SAVE GAME的binding3、正常device458,81及九個first-hit案例，沒有新IDA或guest。原7DF18的enable word1919E4當次值未知，必須只讀通過0才點擊；原7DF29／802CC→7E154來源核對。DRAFT原稿及來源審查後READY，尚無412 Go或guest；限定正常press／安全release及case3真CALL入口，不預稱保存頁或檔案寫入成功。

實際容器內命令：

```text
python3 workplace/new-game-411-generator.py
bash workplace/new-game-411-run.sh > workplace/new-game-411-run-output.txt 2>&1
python3 workplace/new-game-411-source-verify.py
python3 workplace/new-game-411-verify.py > workplace/new-game-411-verify-tests.txt 2>&1
python3 workplace/new-game-412-source-verify.py > workplace/new-game-412-source-tests.txt
python3 workplace/new-game-412-document-gate.py
```

沿Go1.24.13既有image、UID/GID1000、network none、原ZIP／patch唯讀。guest900s／state850s／3GiB／2CPU／128pids、GOMEMLIMIT1GiB、owned PID trap與cgroup收據保持。原418檔及SAVE10／MOX保持；cgroup峰值1501089792bytes、oom／oom_kill增量0。來源／文件30s／512MiB／1CPU；獨立數值90s／2GiB／1CPU。本輪373份私有收據索引workplace/412-current-receipt-index.json；較早371／110／45／53／39份及972份失敗原guest hash保持。所有原EXE／PNG／Go／JSON／LOG Git忽略，沒有為畫面或檔名重跑guest。

收尾稽核：容器內python3 workplace/412-main-audit.py通過。四份自撰文件、唯一DOS活表next412與追加式歷史核對；373份新及371／110／45／53／39份較早收據、972份失敗原guest hash保持。工具乾淨且已推送，主庫玩法及公開CPU／DOS保持；411限定原GAME輸入CONFORMED、412 READY尚無Go或guest，409／406保持DRAFT。root-owned基線2437檔／272目錄保持，沒有.md目錄；git diff --check通過，Docker三個專案volume filter無容器，沒有新image。

## 2026-10-05：正常SAVE輸入與保護模式DOS屬性查詢

工具分支codex/moo2-parity-20260930，HEAD 9e4c7bd4e2f1baaf205fbbb4bf680001f9270189已推送github。主庫只更新四份自撰文件，玩法RE-first保持。公開工具差異相對a03c322d28002bb0e11d5d5109d5833a17114d04只新增internal/machine/le_startup.go及其測試中的414保護模式平台查詢；CPU解碼與公開probe保持。

412私有Go SHA-256 6ad43493b3adc0ff46669e1299e755f243f57ef1e8d138cb1137ec4a0f609a67；9個反轉patch、唯一Step／原getter與兩個正常裝置呼叫、245CLI含201拒絕／44正對照通過，完整235前綴保持。六份產物在容器暫存區逐bytes再生一致。原唯一session26701，完整411／409及全部祖先保持；actual enable word0後236253170正常press458,81,1，236263779 C3真SS槽20E1AC、236263780唯一下一Step返回EAX3及GUI229,81。236264659 elapsed43132µs、callback24／24且idle後正常release，CPU／RAM保持；236282811原reader C3槽16DD7C、236282812自然返回EAX3與ESP＋4，原7DF29 writer逐Step寫mode3。15phase全只讀、原Code16／LE fixups及PNG hash獨立核對。

238069860原219E75 bytes CD21／AX4300／DS188h／EDX2BD904未支援，拒絕後EIP219E77；當次檔名未取樣，存檔入口未到。outer900s exit124，原probe exit0但cpu_stop／step_error明確，resource-after／gzip／state manifest完整；Docker容器已刪除，沒有第二個guest。原PNG人工仍為GAME面板、游標在SAVE，數值與人工分開。412完整契約回DRAFT；389份原產物及manifest按failed1-412保存，完整15phase與terminal不改寫；14個已成功phase另給415凍結。

413單次窄IDA225列／135EA／26fixup，原2object／365page／51363fixup獨立重建。原7E154入口、7E1E6→7D061控件建立與7E1FD→1171AB第一正常輸入來源核對；不深挖標準helper。IDA殼層exit0／idat_exit1、固定EXE hash／非空JSON／5365函式及UID1000通過。

414經DRAFT→Microsoft公開契約及現有普通檔模型審查→READY，接通MOO2保護模式AH43／AL0唯讀查詢；完整32位EDX、NUL／260bytes、真provider與2／3／5錯誤回傳，非輸出狀態／RAM／來源保持。普通檔CX20h沿既有平台近似，不宣稱原FAT屬性exact。12案例、internal/machine／dos／dosfile套件與核心／命令程式建置通過，工程CONFORMED，原玩家續行未驗。首輪測試誤設ECX索引及constructor已初始化表，fixture／輸出與manifest保存後只修測試；go build ./...誤含忽略的私有workplace Go，失敗保存，按建置範圍重核通過。沒有重跑原guest。

實際容器內命令：

```text
python3 workplace/new-game-412-generator.py
bash workplace/new-game-412-run.sh > workplace/new-game-412-run-output.txt 2>&1
python3 workplace/new-game-412-implementation-source-verify.py
python3 workplace/new-game-412-verify.py > workplace/new-game-412-verify-tests.txt 2>&1
bash new-game-413-ida-run.sh > new-game-413-ida-run-output.txt 2>&1
python3 workplace/new-game-413-byte-verify.py
python3 workplace/new-game-413-source-verify.py
go test ./internal/machine -run TestMOO2ProtectedFileAttributes -count=1 -v
go test -p 2 ./internal/machine ./internal/dos ./internal/dosfile -count=1
go build -p 2 ./internal/... ./cmd/...
python3 workplace/new-game-415-document-gate.py
```

原418檔與SAVE10／MOX保持，覆蓋層無新差異。沿Go1.24.13與IDA9.4既有image、UID1000、network none、原ZIP／patch唯讀；guest900s／state850s／3GiB／2CPU／128pids／GOMEMLIMIT1GiB，cgroup峰值1518264320bytes、oom／oom_kill增量0。IDA120s／2GiB；bytes90s／2GiB／1CPU；測試有界180／240s與2GiB／2CPU；其餘來源／文件30s／512MiB。全部handle已終止，Docker三個專案volume filter無容器，沒有新image。本輪804份收據索引workplace/415-current-receipt-index.json；較早373／371／110／45／53／39份及972份失敗原guest hash保持，原檔與所有私有Go／PNG／JSON／LOG忽略。

415經DRAFT→來源／平台build input與邊界審查→READY，尚無Go或guest。只凍結原成功14phase，明示414平台建置輸入及真正檔名／CF返回，不改原拒絕。240M與虛擬時間保持，工具外層改1200s／state1150s、kill-after15s以補足outer124；不把工具等待時間當玩法。正常保存／讀取與remake同狀態未驗。

收尾稽核：容器內python3 workplace/415-main-audit.py通過。四份自撰文件、唯一DOS活表next415、追加式歷史與工具乾淨且已推送核對；804份新收據、較早373／371／110／45／53／39份、389份新失敗原guest及972份較早失敗hash保持。主庫玩法不變，公開工具差異只允許414平台與測試兩檔；412正常SAVE輸入已驗而完整契約DRAFT，414工程CONFORMED，415 READY尚無Go或guest。root-owned基線2437檔／272目錄保持，沒有.md目錄；git diff --check通過，Docker三個專案volume filter無容器。

## 2026-10-05：原SAVE屬性查詢與真入口續行

工具HEAD 2137e39c4a599d8c64c1369492f015014739901f已推送原分支。415經READY後完成私有觀察器，7個精確反轉patch／唯一CPU.Step／原getter、255CLI含209拒絕及46正對照與六份逐bytes重生通過。固定archive明示加入公開414平台輸入，hash 0d1860f0c22c7583e25865cfa5697dc54b11061efb0a38b8e1feceb85f40b90d核對。前置建置未啟動guest，只有一次原session56037；outer及probe exit0。

完整411／409、406及所有較早前置、412成功14phase保持。238069860原runtime219E75／AX4300、DS188h／EDX2BD904，NUL路徑SAVE1.GAM，bytes 53415645312E47414D00。此檔在原417根檔及patch／覆蓋層中不存在；238069861唯一下一Step到219E77回AX2／CF1，CX與全部非輸出暫存器／segment／RAM／裝置保持。原版處理缺檔後繼續，沒有清CF或預填成功。

238113911原runtime1702CC CALL，238113912真SS到16E154、ESP−4／return1702D1；沿原SAVE press／selector3／43132µs release、reader返回與mode3 writer，沒有新增裝置輸入。獨立驗證重建LE bytes／fixups，18個SAVE及4個屬性phase全只讀；實際檔案集合與前後狀態通過。原PNG人工仍為GAME面板及SAVE游標，存檔頁尚未繪製。415僅此契約CONFORMED；412原拒絕與DRAFT不改寫，414普通檔20h仍是平台近似。主庫玩法及公開CPU／probe保持，正式存讀與remake同狀態未驗。

實際容器內命令：

```text
python3 workplace/new-game-415-generator.py
bash /tmp/415-preflight.sh
bash workplace/new-game-415-run.sh > workplace/new-game-415-run-output.txt 2>&1
python3 workplace/new-game-415-implementation-source-verify.py
python3 workplace/new-game-415-verify.py > workplace/new-game-415-verify-tests.txt 2>&1
```

沿Go1.24.13既有image、UID/GID1000、network none、唯讀原ZIP／patch、3GiB／2CPU／128pids／GOMEMLIMIT1GiB；工具外層1200s／kill-after15s、state1150s，虛擬時間與240M上限保持。cgroup峰值1841967104bytes、oom／oom_kill增量0；418原輸入與SAVE10／MOX保持，覆蓋層無新差異。所有handle已終止，沒有重跑原guest。私有收據索引workplace/415-continuation-receipt-index.json共1195份；較早804份及其祖先、972份較早失敗原guest保持。原版資料與Go／PNG／JSON／LOG仍忽略。

下一步依[413來源](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/413-moo2-save-entry-input-source.md)，建立415完整入口到第一個正常存檔頁reader的窄觀察規格；不得以入口替代頁面或保存驗收。

收尾稽核：容器內python3 workplace/415-continuation-main-audit.py通過；四份自撰文件、1195份本輪索引及較早804／373／371／110／45／53／39份收據、root-owned基線2437檔／272目錄及零.md目錄核對。git diff --check通過，工具乾淨且已推送，主庫玩法及公開CPU／probe保持。Docker三個專案volume filter無容器，沒有新image。

## 2026-10-05：存檔控件與第一正常輸入，顯示尚待刷新

工具HEAD c92fbf3a8dce6071421f03814e4c4868b0dc7fe0已推送原分支；公開只更新五份自撰工具文件，主庫玩法與公開CPU／probe保持，DOS差異仍只允許414兩檔。416經DRAFT→來源／範圍審查→READY，私有觀察器12個反轉patch、唯一Step／原getter、零新裝置輸入、265CLI含217拒絕／48正對照、六份逐bytes重生通過。固定archive明示copy公開414平台輸入，其hash保持。

唯一原session56044，outer及probe exit0。完整415 terminal與18＋4phase、411／409及全部祖先保持；238113912只解除工具停止旗標，沒有guest寫入。238114944原16E1E6 CALL、238114945真SS到16D061，SP−4／return16E1EB。238142546原16D891 C3、238142547自然返回16E1EB，SP＋4；238142548原分支自然走零。238142550原16E1FD CALL、238142551真SS到2071AB、return16E202，9個新phase全只讀。獨立LE bytes／fixups與完整末態驗證通過。

當次25×55bytes表pointer298848／bias0，hash 0400259e495af6e8bf475e1d5c63e55118459ca94f0bfc9a716a3b959c44b1e1。包含十個kind11、十個kind7及兩個底部kind0矩形；實際語意與物件ID尚未映射。原PNG人工仍是GAME面板與SAVE游標，hash與415相同，存檔頁尚未刷新。416只限控件建立／第一reader CONFORMED，正式存讀與remake同狀態未驗。

附帶一次最小IDA查詢，reader返回後48指令與同玩家函式分支，62列／59EA／12fixup通過原LE／file bytes核對。原BX與物件+38h／+7Ch十列word及+232h比較已證實，slot與控件名稱未知。IDA殼層exit0／idat_exit1、固定EXE hash／5365函式／非空JSON及UID1000通過；不研究callee或標準helper。結果首次寫入誤用唯讀掛載，失敗收據保存，改用可寫容器同來源核對通過；沒有重跑guest或IDA。初次工具READY審核判為跨越主庫RE-first，補查主庫玩法限制、工具READY規範及活表下一步後，同操作獲准，正式玩法閘門保持。

實際容器內命令：

```text
python3 workplace/new-game-416-generator.py
bash /tmp/416-preflight.sh
bash workplace/new-game-416-run.sh > workplace/new-game-416-run-output.txt 2>&1
python3 workplace/new-game-416-verify.py > workplace/new-game-416-verify-tests.txt 2>&1
bash new-game-416-ida-run.sh > new-game-416-ida-run-output.txt 2>&1
python3 workplace/new-game-416-byte-verify.py
python3 workplace/new-game-416-dispatch-source-verify.py
python3 workplace/new-game-416-document-gate.py
```

沿Go1.24.13與IDA9.4既有image、UID/GID1000、network none、原ZIP／patch唯讀。原guest1200s／kill-after15s、state1150s／3GiB／2CPU／128pids／GOMEMLIMIT1GiB，240M邏輯上限保持；cgroup峰值1725698048bytes、oom／oom_kill增量0。IDA120s／2GiB／2CPU，bytes90s／2GiB／1CPU；418輸入與SAVE10／MOX保持，覆蓋層無新差異。所有handle已終止，沒有新image；原EXE／Go／PNG／JSON／LOG仍忽略。本輪索引workplace/417-current-receipt-index.json共1619份；較早1195／804／373／371／110／45／53／39份及原失敗收據保持。

417經既有原CALL／分支、完整416末態與VBE107／實際未刷新PNG審查至READY。只觀察第一次原VBE顯示更新，原reader若自然返回就記錄真值，不預填0、不強制分支或新增輸入，240M保持；尚無417 Go或guest。顯示計數變化本身不能證明存檔頁正確可見，須另看原PNG。

收尾稽核：容器內python3 workplace/417-main-audit.py通過；四份自撰主庫文件、工具乾淨且已推送、本輪1619及較早1195／804／373／371／110／45／53／39份收據保持。root-owned基線2437檔／272目錄、零.md目錄及git diff --check通過。416僅控件／第一reader CONFORMED，原PNG仍為GAME；417 READY尚無Go或guest，主庫玩法／公開CPU保持。Docker三個專案volume filter無容器，沒有新image。


## 2026-10-05：417原存檔頁首次顯示與可見畫面

- 目標：沿完整416第一reader末態，只讀觀察原VBE首次更新與原物件ID；不修改主庫玩法／公開CPU或新增原輸入。
- 前置：10個可反轉patch、275CLI含225拒絕／50正對照、六份逐bytes重生與唯一Step／getter核對通過。原417 READY收據凍結，追加範圍以readonly-bindings-review獨立審查。
- 命令：既有Go1.24.13 image、network none、UID/GID1000、原ZIP／patch唯讀；timeout --kill-after=15s 1200s、3GiB／2CPU／128pids、GOMEMLIMIT1GiB，執行bash workplace/new-game-417-run.sh。唯一guest session74131 outer／probe exit0，不重跑。另以90s／2GiB／1CPU容器執行python3 workplace/new-game-417-verify.py，session12375 exit0。
- 已驗：完整416末態／9phase與祖先保持；238143403／404原reader真SS返回16E202、ESP＋4、實際EAX0。238251948第一VBE107→108，8個新phase全只讀。實際PNG以view_image人工檢視，可見九個empty slot、Auto Save與SAVE／CANCEL；數值與視覺分開。
- 原4516E0物件0x234bytes與兩組word ID1..10／11..20、+232h＝21只讀捕捉；正式控件編輯／保存語意仍未知，不以位置關聯代替實際輸入。
- 原418輸入與SAVE10／MOX保持、覆蓋層無差異；cgroup峰值2012610560bytes，oom／oom_kill增量0。唯一容器moo2-save-417-20261005已由--rm移除，專案相關docker ps -a為空。收據與原版素材仍Git忽略。
- 工具分支HEAD 0cd0bc3b13ae7b068dcc07cde077685a7eddc442 已推送github/codex/moo2-parity-20260930。主庫只更新既有四份交接文件；提交／推送後精確HEAD、工作樹與清理狀態保存workplace/417-continuation-final-state.json。
- 邊界：417只限原首次顯示及頁面可見CONFORMED，未驗正式選格／命名／存讀及remake同狀態，完整remake目標仍進行中。
- 下一步：核對當次第一空格kind11、兩組ID的正常輸入分派及consumer；來源足夠後審查有界原輸入，主庫RE-first保持。

收尾稽核：python3 workplace/417-continuation-main-audit.py通過，2033份本輪索引與全部祖先合併3012份獨立hash核對。四份自撰文件與工具已推送狀態核對，root-owned基準2437檔／272目錄、零.md目錄及git diff --check通過；主庫玩法與CPU保持，沒有新增image。


## 2026-10-05：418存檔格／名稱輸入來源與419正常提交契約

- 目標：接續原可見存檔頁，錨定第一格、文字buffer、SAVE分派與原保存CALL，不改主庫玩法。
- 原來源：一次IDA Pro9.4匯出sub_7E154的260個玩家指令；通用selector／kind11／ASCII／release從原moo2-379-ida-list-input-first-query.json只讀選取，不重新分析helper。
- 命令：既有IDA image、network none、UID/GID1000、patch唯讀，timeout120s／2GiB／2CPU／128pids執行bash /out/new-game-418-ida-run.sh，session29387 wrapper exit0；idat實際exit1與非空JSON／5365函式／固定hash記錄。另以既有Go1.24.13 image、timeout90s／2GiB／1CPU執行python3 workplace/new-game-418-byte-verify.py，session33568 exit0。new-game-418-source-verify.py與document-gate.py通過，沒有新guest。
- 已證實：915列／891EA／151fixup獨立原bytes核對、13個含端點first-hit案例；控件1為kind11，control+18h對應37-byte名稱record，原選格writer與SAVE ID21接受0..9，7E3F4 CALL1160B參數資料流已錨定。
- 證據等級：原ID／bytes／pointer／比較與CALL已證實；「保存callee入口」的玩家用途為強推論，內部檔案writer與成功回饋未知。正式選格／命名／保存與remake同狀態未驗。
- 419經來源與範圍審查至READY，只允許私有原版工具正常第一格與SAVE兩次click，到原保存CALL入口後停止；新245M後段保留原417的240M收據，不猜補輸入結果或代寫RAM。尚無419 Go／guest。
- 工具HEAD 4cdc5c6cd801a3ed5a892a94a658fc05d7ca7560 已推送github/codex/moo2-parity-20260930。公開只提交五份自撰工具文件；新原碼／IDA JSON與私有收據本機忽略。主庫更新既有四份交接文件與HONEST-STATUS的過期斷言，推送後精確HEAD保存workplace/419-final-state.json。
- Docker專案volume filter為空，沒有新image。完整收據索引workplace/419-current-receipt-index.json共2052份，較早2033份保持；root-owned基準2437檔／272目錄與零.md目錄已核對，精確工作樹由最終狀態保存。
- 下一步：實作419 READY觀察器、前置建置與CLI／六份重生，再以唯一原guest正常選格與SAVE到入口；主庫RE-first保持，完整remake目標繼續。

現況勘誤：核對正式finalize呼叫端、先進文明／Money規則、AI profile權重與index及四列追溯表後，移除HONEST-STATUS仍稱MISSING／錯誤PARTIAL的舊斷言。四列目前為CONFORMED／INTERNAL，原版全域PRNG與完整同狀態仍未證明；不變更歷史儀表板數字。檔案hash與核對範圍保存workplace/419-status-source-review.json，沒有改Go或新增runtime完成聲明。

收尾稽核：python3 workplace/419-main-audit.py通過，2052份本輪索引與全部祖先合併3031份hash核對。工具已推送且工作樹乾淨、419 READY來源釘選與無Go／guest核對，五份主庫文件含HONEST-STATUS勘誤；root-owned基準2437檔／272目錄、零.md目錄與git diff --check通過。


## 2026-10-05：419第一格press與420工具守衛修正

- 命中GUI還原、規格閘門及文件分工路由。只處理私有原版工具與既有交接，主庫RE-first保持。
- 419受控SAVE10日期輸入、六項普通讀寫／日期／旗標／路徑回歸通過；285CLI／六份重生與來源檢查通過。最初DTA全log單時間假設在改檔或guest之前被拒絕，原失敗保存後按兩群修正，不改原資料。
- 命令：既有Go1.24.13 image、network none、UID/GID1000、3GiB／2CPU／128pids／GOMEMLIMIT1GiB、1200s外層與1150s owned state capture執行bash workplace/new-game-419-run.sh。唯一session1794 outer1／probe2；工具額外mask2Bh條件拒絕原mask1，未派送release／SAVE。
- 已證實：238285408正常第一格press、238295561原selector真RET到20E1AC／ESP＋4／EAX1、九個只讀phase及完整417／祖先。selected仍FFFF、editor未啟用、保存入口未到。這是工具守衛錯誤，沒有新CPU缺陷證據。
- 獨立python3 workplace/new-game-419-verify.py，session76204 exit0限定核對原press與拒絕。原PNG人工另檢視、436份完整失敗保存；cgroup峰值1498697728bytes／OOM增量0、原418檔／SAVE10／MOX保持、overlay無內容差異。原base／overlay日期歸屬為強推論，不把mtime當成寫檔證據。
- 420修正以既有415成功mask1 release與當次419來源經READY審查後實作。五個反轉patch、唯一Step／原getter、零新裝置API／鍵盤、295CLI241拒絕／54正對照、六份重生通過。完整419九phase／失敗末態先凍結，才修正私有release守衛；原419資料不改。新唯一session64443 outer0／probe0，原guest正常完成。
- 目前正式存讀／鍵盤命名／remake同狀態未驗。公開只更新自撰工具規格與四份主庫文件；沒有主庫Go或原CPU／平台修改。工具HEAD 9fef15c4974bab36a4cdbd0097ee0328c9c81830已推送；主庫精確HEAD與清理狀態保存workplace/420-final-state.json。
- 命令：同容器界限執行bash workplace/new-game-420-run.sh；獨立python3 workplace/new-game-420-verify.py，session61489 exit0。26個只讀phase、四次正常裝置操作、原selected FFFF→0、reader 1／21及238505421真CALL到10160B／參數0通過；callee未執行。原名稱僅在記憶體生成，PNG仍顯示空格，不稱為保存成功。峰值1679069184bytes／OOM增量0，原輸入／副本與overlay內容保持。
- 原419失敗不覆寫，420限定CONFORMED；417／418／419回鏈與索引更新。五份工具自撰文件檢查通過並推送，四份主庫文件完成當次核對；完整新收據索引、公開CPU兩檔既有差異與八份主庫來源雜湊保持。Docker相關容器清空、root-owned基準2437檔／272目錄與零.md目錄保持。下一步只做sub_1160B窄RE，不解除主庫玩法閘門。
- 收尾命令python3 workplace/420-main-audit.py通過，本輪3383份索引及全部祖先合併4362份雜湊保持；主庫來源八份與419／420執行前輸入釘選通過。Docker兩個volume filter均為空。
