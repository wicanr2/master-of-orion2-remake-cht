# `dosgolem` 接入 MOO2 1.31 DOS 原版的首輪收據

日期：2026-09-30。**範圍僅是格式、載入與原版入口前 11 步；尚無原版與 remake 的玩法或畫面同狀態對拍。** 使用者決定以 DOS 版 `ORION2.EXE` 和 `dosgolem` 進行後續對拍。

## 固定輸入與工具

| 來源 | SHA-256／版本 |
|---|---|
| `moo2_patch1.31/MOO2-1.31.en.zip` | `908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5` |
| ZIP 內 `ORION2.EXE`，2,612,010 bytes | `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |
| 1996 光碟 ZIP 內 `mastori2/Orion2.exe`，2,644,842 bytes | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| `dosgolem` 工具 | `/home/anr2/cht/dosgolem` 起點 Git `cc1ef5611b5eb288cc489ae504f0a7a96fc526bf`；隔離修改在 `workplace/dosgolem/` 的 `codex/moo2-parity-20260930` 分支，格式與入口修正 `a2905dbcc9c14ed0716d618d2f2450a495b4ee78`，CPU 補強 `c32e854bca43c7e3b3ebbdc5caeba4d0ef6e9960`，暫定服務入口 `cc6231bd7f6578d379e1faf31ae4f1214bf917d3`，固定原版返回訂正與版控回放 `44e488536bbcfaf079c07779aaa22e91c4f9fe2f` |
| 執行環境 | `golang:1.24-bookworm` Docker，Go 1.24.13，`--network none`、原版輸入唯讀、目前使用者 UID/GID |

原版 EXE 存於被 Git 忽略的 `workplace/oracle-input/`，原 ZIP 與萃取的 EXE 均未改動。只為診斷曾建立一份標明 `header-probe` 的本機合成副本，將 MZ `0x3C..0x3F` 四 bytes 改成明示的 LE 偏移；正式解析收據已由未改動原版重生，不依賴合成副本。

## 內嵌 MZ／LE 的正確載入

**已證實（原始檔案偏移）**：外層 MZ 的 `0x3C..0x3F` 為 `00 00 B4 09`，不能當標準 `e_lfanew`。兩版原檔均在 `0x26654` 有第二層 `MZ`，其 `e_lfanew=0x2C90`，指向 `0x292E4` 的 `LE 00 00`。1996 版 `DataPagesOffset=0x6F040`，第一個資料頁為 `0x26654+0x6F040=0x95694`；1.31 則為 `0x6F000`、檔案位置 `0x95654`。1996 版首資料頁 `CC EB FD 90` 與公開 Watcom `BEGTEXT` 入口形狀相符。以第二層 MZ 基址載入後，兩版 object 1 `+0xFFF18` 均為 `EB 76 WATCOM`；1996 版對應原檔偏移 `0x1955AC`，亦由 IDA 檔案映射獨立核對。

`dosgolem` 已新增 `InspectLEInMZ(data, 0x26654)`、`LoadLEInMZ` 及 `cmd/leprobe -mz-base 0x26654`。其原檔收據位於未版控的 `workplace/dosgolem/workplace/moo2-leprobe-131.txt` 與 `moo2-leprobe-1996.txt`；兩版均為 2 個 objects、51,363 筆 internal fixup。1.31 有 365 pages、資料頁原檔起點 `0x95654`、重定位 object SHA-256 為 `d85e5536e7181c3bfd332d43284522448af22fca2133101869e034d8de29be85`／`49bc4ff5fb8da9841f42bc25c9571168ac2df0ef2a75531d444f9f10d2577767`；1996 有 373 pages、資料頁起點 `0x95694`，object SHA-256 為 `ea08d9ce3c7592e7615967fa2c2ee6c2f92418c7e2095fe71dd1fdc3af0ea0ef`／`6e4f2635f74716b99331530ec2cdd05c90c571c9bf59c18353f65c3e976a11a6`。格式契約、勘誤與驗收見 `dosgolem` 分支的 [`197`](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/197-bound-mz-le-file-base.md)。

**已證實（dosgolem 重定位 LE 線性位址）**：兩版均由 `0x10FF18` 的 `EB 76` 自然跳過 Watcom 標記，走到第 11 步 `0x10FFB5`，要求 `INT 21h/AH=30h`；`EAX=0x3000`、`EBX=0x50484152`（`PHAR`）。目前探針未接 DOS 服務，所以在此失敗即關閉。1.31 初始 ESP 是 `0x1CDCD0`，1996 是 `0x1D5CD0`。`go test ./...` 及兩版固定雜湊真檔入口測試通過；此結果不代表正常玩家畫面或玩法已對拍。

**診斷，非正式收據**：暫掛 dosgolem 既有的 FD2 專屬 DOS 服務層，兩次啟動呼叫可前進；該服務層固定 FD2 的 selector 與 `FD2.EXE` 環境，不能冒稱 MOO2 的 DOS/4GW 狀態。診斷顯示第 36 步 `0x110076` 的 `66 09 CA`（`OR DX,CX`）需要通用 16 位指令支援；此形狀已依 `dosgolem` [198](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/198-cpu386-or-word-register.md) 增量補上並測試。診斷再到第 45 步 `0x1100B2`，停在 `66 26 8C 1D ED 01 11 00` 的 segment store。第 45 步不作正常原版路徑、玩法或畫面對拍證據；下一正式閘門仍是建立 MOO2 的 DOS 服務契約。

**後續工具補強，同樣僅是診斷**：依 `dosgolem` [199](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/199-cpu386-moo2-segment-prefixes.md) 補 `66 26 8C 1D` 的 ES 覆寫 16 位段暫存器寫入，以及第 50 步 `0x1100CF` 的 `3E BA 50 A1 1C 00`（無記憶體運算元的 `MOV EDX,0x001CA150`，DS 前綴無作用）。合成測試包含不同 ES／DS 描述子 base、word 越界原子拒絕與不變的旗標。1.31／1996 兩版的隔離診斷現在均至第 65 步、dosgolem LE 線性位址 `0x1100FA`，停在 `28 C0 AA AA...` 的未支援 opcode `28`。**前 11 步的正式無服務收據沒有改變**；第 65 步仍借 FD2 專屬服務，只能用來找通用 CPU 缺口，不能說 MOO2 DOS 啟動或玩家流程已通過。

**暫定 MOO2 服務入口**：`dosgolem` [200](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/200-moo2-provisional-protected-dos.md) 新增 `NewMOO2StartupDOS`，把合成最小環境的執行檔名設為 `ORION2.EXE`，並保持 FD2 舊設定不變。兩版真檔經此入口均請求 `AH=30h`、`AX=FF00h`，診斷仍在第 65 步 `0x1100FA` 停住。這個服務入口**暫沿用 FD2 的 DOS/4G selector 與回傳值**，尚無 MOO2 原版服務回傳的獨立證據；`ORION2.EXE` 環境亦是明示的合成輸入。因此它只讓下一個 CPU 缺口可定位，不能將第 65 步升格為正式原版或玩法 parity。正式待查項是 MOO2 的服務回傳與首個玩家可見狀態。

### 2026-09-30：MOO2 DOS/4G 啟動服務的輔助基準訂正

上段「沿用 FD2 回傳、MOO2 回傳未知」是建立暫定入口當下的狀態，**現已被固定 MOO2 1.31 caller 的 DOSBox-X 返回快照訂正**。使用原版 `ORION2-1.31.EXE`，SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；輔助執行器 `fd2-dosbox-x:debug-0d7b272b`、image ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`、DOSBox-X 2026.07.02 SDL2 heavy debugger，`core=normal`、`cycles=fixed 12000`、32 MB。原版 EXE 唯讀掛載，僅複製到一次性容器 `/tmp/game/ORION2.EXE`；Xvfb 與 Python PTY 有時限及清理 trap。原始終端與命令在未版控的 `workplace/dosbox-moo2/`；公開庫不加入遊戲 bytes 或完整記憶體擷取。

**已證實，限固定 DOSBox-X 輔助環境及其 `CS:EIP` 位址空間**：以 `BPINT 21 30` 過濾載入器呼叫，在 MOO2 自身 `0180:00333FB5` 的 `CD 21`，`EAX=00003000`、`EBX=50484152`、`DS=SS=0188`、`ES=0028`、`GS=0020`；返回點 `0180:00333FB7` 是 `EAX=00000005`、`EBX=5048FF00`，selector 不變。再於自身 `0180:00334054` 的 `AX=FF00h`、`EDX=0078h` 進入，返回點 `0180:00334056` 為 `EAX=4734FFFF`、`GS=0020`、`EFLAGS=0296`（進入時 `0297`，CF 清除）。這些 DOSBox-X `CS:EIP` 數值**不能**與 dosgolem 的 LE 線性位址 `0x10FFB5` 當同一種位址比較；bytes／呼叫順序是交叉定位依據。詳細固定暫存器表及勘誤見 `dosgolem` [201](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/201-moo2-dos4g-startup-returns.md)。

**被推翻的假設與剩餘限制**：暫定入口原先借 FD2 的 `AX=1606h`、`DS=SS=0160h`，與這次固定 MOO2 返回直接矛盾，現已改成 1.31 的 `EAX=5`、`EBX=5048FF00`、`DS=SS=0188`；第二次呼叫的回傳也已獨立核對。1.31 dosgolem 診斷仍停第 65 步 `0x1100FA` 的未支援 `28 C0`，但此步數不取代 dosgolem 正式正常玩家路徑收據。`ES:[2Ch]` 環境 selector、PSP bytes、合成環境內容及 1996 版獨立回傳仍未知或近似；待 dosgolem 擴充後重生正式收據。**尚無原版與 remake 的玩法或畫面同狀態對拍。**

**工具診斷續記**：`dosgolem` [202](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/202-cpu386-moo2-sub-byte-register.md) 補 `28 C0` 所屬通用 `SUB r/m8,r8` 暫存器形狀，合成測試驗證 byte 方向、旗標與未支援 memory 形狀拒絕。固定 1.31 真檔隔離診斷現至第 72 步、dosgolem LE 線性位址 `0x110102`，停在 `26 66 8E 1D 29 CA 17 00`（ES 覆寫的 16 位段載入）缺口。前述第 65 步為診斷當時停點，保留作歷程；PSP／環境仍有合成假設，**72 步也不能當玩家路徑對拍**。

輔助服務快照已由版控中的 `dosgolem/apps/moo2/tools/startup_probe_131.py` 重新產生；本機 `workplace/dosbox-moo2/startup-registers.json` SHA-256 `341ce45f695a350958a74c7a219b2bcf621ff1d86ac3cfb390d0726356da0a44`，包含兩次呼叫前後的固定雜湊、位址空間與暫存器順序。回放腳本與 `dosbox.conf` 入工具分支，原版 EXE、終端 raw 與私有畫面仍只在本機。每次掛載來源須先確認形態，容器使用 `--rm --network none`、有界資源、目前使用者 UID/GID、Xvfb 清理 trap 及 `timeout 150s`。

## 舊資料頁基址的勘誤

## 2026-09-30 通用 CPU 指令續驗

**已證實（僅限所列指令形狀）**：`dosgolem` 隔離分支依規格 203–206 增量支援 `26 66 8E 1D` 的 ES 覆寫 16 位絕對記憶體段載入、`3E B9` 的 DS 前綴暫存器立即數、`26 3A 10` 的 ES 覆寫 byte 比較，以及 `0F A8` 的 32 位堆疊 `PUSH GS`。合成測試分別核對描述子 base／界限、前綴不影響無記憶體運算元、比較旗標、堆疊寬度與失敗原子性。`go test ./internal/cpu386 ./internal/machine` 通過。輸入仍是 SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` 的未修改 1.31 EXE；工具的位址基準是 dosgolem 重定位 LE 線性位址。

**診斷，非原版玩家路徑收據**：沿用規格 201 的兩次啟動返回與合成 PSP／環境，固定真檔探針依序從第 72 步跨至第 125 步 `0x11014C`、第 393 步 `0x163A90`、第 406 步 `0x13CB80`，本輪至第 419 步 `0x13CBAA` 停於 `66 83 E7 FC`（未支援的 16 位 `83` 形狀）。步數是 CPU 工具缺口定位，不能推論 DOS 啟動、畫面或玩法已完成對拍；無服務 hook 的正式 fail-closed 入口仍是第 11 步 `INT 21h/AH=30h`。後續先核對 MOO2 環境／PSP 消費端與服務狀態，再以 dosgolem 重生玩家檢查點，不能把 DOSBox-X 輔助基準當正式收據。

上一輪 `LoadLEAt(original, 0x292E4)` 雖讀對 LE **標頭**與 fixup 表，卻把 `DataPagesOffset` 當整個檔案的絕對偏移，錯從 `0x6F000／0x6F040` 複製物理頁面。先前 `0x10FF18` 的 `66 3B 15...`、重定位 object SHA-256 `c32b38c5...`／`137ee072...`、5,359 步及 `0x10FF43 POP EDX` 均是錯頁診斷，**不得當原版啟動收據**。原先「缺少 DOS/4GW 外層堆疊」假說因此撤回；通用 `66 3B 15` CPU 指令測試仍是有效的獨立工具測試。這次沒有接入音訊、輸入、固定 seed 或 remake 同狀態比較。

## IDA 位址基準核對

使用 `ida-pro-9.4-idapython:locked-v1` 查唯讀原版資料庫的拋棄式副本。輸入為私有 `Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA 記錄的原始輸入 MD5 是 `bacb10a92454d2f9b211eb9fe67ec099`，與上述 1996 光碟版 `Orion2.exe` 的 MD5 相同。查詢輸出留在未版控的 `workplace/ida-probe/cstart.json`。

**已證實（IDA 線性位址空間與原檔偏移）**：IDA 在 `0x10FF18` 顯示 `EB 76 57 41 54 43 ...`，標為 `start`，並映至原檔 `0x1955AC`。以內嵌 MZ 基址修正 dosgolem 載入後，其重定位 LE 線性位址 `0x10FF18` 的 bytes 相同；舊版 `66 3B 15...` 差異已由錯誤資料頁基址解釋。IDA 的函式名稱仍只作導航；這次原檔偏移、bytes、LE object+offset 與兩工具位址已逐項核對，後續交叉參照仍須附上各自位址基準。
