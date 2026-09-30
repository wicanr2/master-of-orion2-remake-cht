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
