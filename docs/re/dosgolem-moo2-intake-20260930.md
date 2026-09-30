# `dosgolem` 接入 MOO2 1.31 DOS 原版的首輪收據

日期：2026-09-30。**範圍僅是格式、載入與原版入口前 11 步；尚無原版與 remake 的玩法或畫面同狀態對拍。** 使用者決定以 DOS 版 `ORION2.EXE` 和 `dosgolem` 進行後續對拍。

## 固定輸入與工具

| 來源 | SHA-256／版本 |
|---|---|
| `moo2_patch1.31/MOO2-1.31.en.zip` | `908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5` |
| ZIP 內 `ORION2.EXE`，2,612,010 bytes | `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |
| 1996 光碟 ZIP 內 `mastori2/Orion2.exe`，2,644,842 bytes | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| `dosgolem` 工具 | `/home/anr2/cht/dosgolem` 起點 Git `cc1ef5611b5eb288cc489ae504f0a7a96fc526bf`；隔離修改在 `workplace/dosgolem/` 的 `codex/moo2-parity-20260930` 分支，修正提交 `a2905dbcc9c14ed0716d618d2f2450a495b4ee78` |
| 執行環境 | `golang:1.24-bookworm` Docker，Go 1.24.13，`--network none`、原版輸入唯讀、目前使用者 UID/GID |

原版 EXE 存於被 Git 忽略的 `workplace/oracle-input/`，原 ZIP 與萃取的 EXE 均未改動。只為診斷曾建立一份標明 `header-probe` 的本機合成副本，將 MZ `0x3C..0x3F` 四 bytes 改成明示的 LE 偏移；正式解析收據已由未改動原版重生，不依賴合成副本。

## 內嵌 MZ／LE 的正確載入

**已證實（原始檔案偏移）**：外層 MZ 的 `0x3C..0x3F` 為 `00 00 B4 09`，不能當標準 `e_lfanew`。兩版原檔均在 `0x26654` 有第二層 `MZ`，其 `e_lfanew=0x2C90`，指向 `0x292E4` 的 `LE 00 00`。1996 版 `DataPagesOffset=0x6F040`，第一個資料頁為 `0x26654+0x6F040=0x95694`；1.31 則為 `0x6F000`、檔案位置 `0x95654`。1996 版首資料頁 `CC EB FD 90` 與公開 Watcom `BEGTEXT` 入口形狀相符。以第二層 MZ 基址載入後，兩版 object 1 `+0xFFF18` 均為 `EB 76 WATCOM`；1996 版對應原檔偏移 `0x1955AC`，亦由 IDA 檔案映射獨立核對。

`dosgolem` 已新增 `InspectLEInMZ(data, 0x26654)`、`LoadLEInMZ` 及 `cmd/leprobe -mz-base 0x26654`。其原檔收據位於未版控的 `workplace/dosgolem/workplace/moo2-leprobe-131.txt` 與 `moo2-leprobe-1996.txt`；兩版均為 2 個 objects、51,363 筆 internal fixup。1.31 有 365 pages、資料頁原檔起點 `0x95654`、重定位 object SHA-256 為 `d85e5536e7181c3bfd332d43284522448af22fca2133101869e034d8de29be85`／`49bc4ff5fb8da9841f42bc25c9571168ac2df0ef2a75531d444f9f10d2577767`；1996 有 373 pages、資料頁起點 `0x95694`，object SHA-256 為 `ea08d9ce3c7592e7615967fa2c2ee6c2f92418c7e2095fe71dd1fdc3af0ea0ef`／`6e4f2635f74716b99331530ec2cdd05c90c571c9bf59c18353f65c3e976a11a6`。格式契約、勘誤與驗收見 `dosgolem` 分支的 [`197`](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/197-bound-mz-le-file-base.md)。

**已證實（dosgolem 重定位 LE 線性位址）**：兩版均由 `0x10FF18` 的 `EB 76` 自然跳過 Watcom 標記，走到第 11 步 `0x10FFB5`，要求 `INT 21h/AH=30h`；`EAX=0x3000`、`EBX=0x50484152`（`PHAR`）。目前探針未接 DOS 服務，所以在此失敗即關閉。1.31 初始 ESP 是 `0x1CDCD0`，1996 是 `0x1D5CD0`。`go test ./...` 及兩版固定雜湊真檔入口測試通過；此結果不代表正常玩家畫面或玩法已對拍。

**診斷，非正式收據**：暫掛 dosgolem 既有的 FD2 專屬 DOS 服務層，兩次啟動呼叫可前進；該服務層固定 FD2 的 selector 與 `FD2.EXE` 環境，不能冒稱 MOO2 的 DOS/4GW 狀態。診斷顯示第 36 步 `0x110076` 的 `66 09 CA`（`OR DX,CX`）需要通用 16 位指令支援；此形狀已依 `dosgolem` [198](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/198-cpu386-or-word-register.md) 增量補上並測試。診斷再到第 45 步 `0x1100B2`，停在 `66 26 8C 1D ED 01 11 00` 的 segment store。第 45 步不作正常原版路徑、玩法或畫面對拍證據；下一正式閘門仍是建立 MOO2 的 DOS 服務契約。

## 舊資料頁基址的勘誤

上一輪 `LoadLEAt(original, 0x292E4)` 雖讀對 LE **標頭**與 fixup 表，卻把 `DataPagesOffset` 當整個檔案的絕對偏移，錯從 `0x6F000／0x6F040` 複製物理頁面。先前 `0x10FF18` 的 `66 3B 15...`、重定位 object SHA-256 `c32b38c5...`／`137ee072...`、5,359 步及 `0x10FF43 POP EDX` 均是錯頁診斷，**不得當原版啟動收據**。原先「缺少 DOS/4GW 外層堆疊」假說因此撤回；通用 `66 3B 15` CPU 指令測試仍是有效的獨立工具測試。這次沒有接入音訊、輸入、固定 seed 或 remake 同狀態比較。

## IDA 位址基準核對

使用 `ida-pro-9.4-idapython:locked-v1` 查唯讀原版資料庫的拋棄式副本。輸入為私有 `Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA 記錄的原始輸入 MD5 是 `bacb10a92454d2f9b211eb9fe67ec099`，與上述 1996 光碟版 `Orion2.exe` 的 MD5 相同。查詢輸出留在未版控的 `workplace/ida-probe/cstart.json`。

**已證實（IDA 線性位址空間與原檔偏移）**：IDA 在 `0x10FF18` 顯示 `EB 76 57 41 54 43 ...`，標為 `start`，並映至原檔 `0x1955AC`。以內嵌 MZ 基址修正 dosgolem 載入後，其重定位 LE 線性位址 `0x10FF18` 的 bytes 相同；舊版 `66 3B 15...` 差異已由錯誤資料頁基址解釋。IDA 的函式名稱仍只作導航；這次原檔偏移、bytes、LE object+offset 與兩工具位址已逐項核對，後續交叉參照仍須附上各自位址基準。
