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

### 2026-09-30：`AH=4Ah` 停點的探針設定勘誤

**舊推論已撤回**：上文在 dosgolem LE 線性位址 `0x15E07C` 的第 554 步 `INT 21h/AH=4Ah`，只發生於未呼叫 `services.AttachMachine(m)` 的合成診斷探針；它不能再列為 MOO2 原版下一個待取回傳。固定 1.31 原檔 SHA-256 仍為 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。原版 DOSBox-X 2026.07.02 SDL2 heavy debugger（映像 `fd2-dosbox-x:debug-0d7b272b`，ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）在 **CS:EIP** `0180:00334072` 的 DPMI `INT 31h/AX=0006h` 返回 `CX:DX=0`、CF 清除；`0180:00334079` 的零基底分支有 ZF=1。修正探針綁定後，dosgolem **重定位 LE 線性** `0x110079` 亦走到 `0x11007D`，原先 `AH=4Ah` 停點消失。新增的固定原檔回歸測試 `TestMOO2AttachedDPMIBaseProbeWhenProvided` 已通過。

版控 [`startup_probe_131.py`](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/apps/moo2/tools/startup_probe_131.py) 只用固定 1.31 EXE 即可重生六個啟動／DPMI 快照；私有輸出 `workplace/dosbox-moo2-dpmi/startup-registers.json` SHA-256 `a334e483d86396c9d026b14c0c580bdae7de46f555cfd6b6914ef68217018562`。前述舊四快照 SHA-256 與錯誤停點仍留在歷史段落，以保留錯誤形成原因。另一次拋棄式畫面試驗才把唯讀 1996 ZIP 的完整資料與唯讀 1.31 EXE 在一次性容器內組裝；DOSBox-X 可顯示片頭，但該畫面只是輔助觀測，不是 dosgolem 正式玩家路徑收據。

由兩次握手返回後擷取的 DOSBox-X 4,096 指令 `CS:EIP` 記錄，在跳過原版環境字串與清空迴圈的長度差後，與已綁定 DPMI 的 dosgolem 順序持續對到其後續啟動程式；原版還命中 `0180:003759EF` 的 `19 C0`。版控腳本 `--sbb` 重生的前後快照為 `EAX=0501h → 0`、`EFLAGS=0246h → 0246h`，私有 `sbb-registers.json` SHA-256 `93082156064ceb921571b445e19b7646b27e809133bd3186c35e2808b27b297b`。通用 CPU [規格 208](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/208-cpu386-sbb-rm32-register.md) 已實作並通過全套 `go test ./...`；固定原檔合成診斷現於 dosgolem LE 線性 `0x15171D` 的 `87 FA` 失敗即關閉，對應 DOSBox-X `CS:EIP 0180:0037571D` 也在有界指令記錄中。兩工具的位址各有基準，不把數字相同視為獨立證據；控制流、原始 bytes 與固定輸入才是對照依據。PSP／環境仍是近似，**尚無玩法或畫面同狀態對拍**。

上述 DPMI 勘誤、版控觀測腳本、固定原檔回歸測試與通用 `SBB` 已在 `dosgolem` 隔離分支提交並推送 `3babfcf9462e9bbf7c7a1c1d67ded0fa0b30bb6b`。下一個最小工具切片是核對並支援已觀測的 `87 FA` 暫存器交換形狀；其後仍須以 dosgolem 真正到達玩家畫面與可重播輸入，才能製作正式對拍收據。此勘誤與既有 197／201／207 規格的歷史說法應一併閱讀；現行待辦只以 `WORKLIST.md` 活表為準。

### 2026-10-01：`87 FA` 通用 CPU 切片

**已證實，限指令形狀**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，dosgolem 重定位 LE 線性 `0x15171D` 原始 bytes `87 FA`；DOSBox-X 2026.07.02 SDL2 heavy debugger 在其 CS:EIP `0180:0037571D → 0037571F` 的兩斷點，`EDX=EDI=003EC028`、`EFLAGS=0216` 前後相同。版控 `apps/moo2/tools/startup_probe_131.py --xchg` 的私有 `xchg-registers.json` SHA-256 `69deda4f6e766c9ffba973bea26dcade18e6f6b0b5b6ff5a780036bf6c6fb18d`；原版兩寄存器相等，不能單憑此樣本證實交換方向。依 Intel 手冊的 `87 /r` 編碼與不同值合成測試，`FA` 是 `XCHG EDX,EDI`，旗標不變。工具位址空間、輸入與版本詳見 dosgolem [規格 209](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/209-cpu386-xchg-register-register.md)。

隔離 dosgolem 工具分支已只補 `87 /r` 的 32 位暫存器形式，保留原記憶體形式；未支援的 16 位暫存器形式拒絕。`go test ./internal/cpu386 ./internal/machine -count=1` 與 `go test ./... -count=1` 通過。已綁定 DPMI 的**合成環境診斷**越過先前第 673 步停點，於第 760 步、dosgolem 重定位 LE 線性 `0x1515CA` 的 `F5` 失敗即關閉。`F5` 的原版前後狀態尚未核對；合成 PSP／環境、完整資料消費端及 dosgolem 正常玩家畫面仍待驗證。此輪未改 remake 玩法，也**沒有玩法同狀態對拍**。

### 2026-10-01：`F5` 與 `21 C8` 通用 CPU 切片

**已證實，限指令形狀**：固定 1.31 原檔 SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，dosgolem 重定位 LE 線性 `0x1515CA` 的 `F5` 與 `0x1515CD` 的 `21 C8`，分別對應 DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）的 **CS:EIP** `0180:003755CA` 與 `0180:003755CD`。原版 `F5` 的前後 `EFLAGS=0202h → 0203h`，由版控 `--cmc` 探針重生的私有 `cmc-registers.json` SHA-256 `a77196163f3432ffb787d577d5f761c6f78b8f0877a76dbd2b6c34dec958fe36`；Intel 手冊定義此為只翻轉 CF 的 `CMC`。

**觀測勘誤**：`21 C8` 的兩個獨立 `BP`／`EV` 快照曾顯示 `ECX=0Fh`、前後 `EAX=80h`，無法解釋相與結果。舊私有 JSON SHA-256 `955d273f733e19d92994fead6adfd3acf3eb65ff6b79a1817ad0bfac9c06d018` 只保留追溯，不作值對拍。改在同一次 DOSBox-X 執行中從 `0180:003755CD` 用 `LOG 2` 記錄到 `0180:003755CF`：原版進入 `EAX=80h`、`ECX=FFFFFFFFh`、CF=1／ZF=0／SF=1／PF=1，離開時 `EAX=80h`、`ECX=FFFFFFFFh`、CF=0／ZF=0／SF=0／PF=0，符合 `AND EAX,ECX`。權威私有 `and-logcpu.txt` SHA-256 `1bdb7c5df2801e62c4bd7aee1a7b7bb533b71fe8c1aa19fb92afa9263536755d`；修正版 `and-registers.json` SHA-256 `c5131b8aea15512c7c1deeb42964c1261e4ea8c182a07dd125a818a50c0d7796` 將 `EV` 限為地址定位，另載兩行連續 LOG。原始 bytes、工具位址空間、Intel 指令契約與限制詳見 dosgolem [規格 210](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/210-cpu386-cmc.md)／[規格 211](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/211-cpu386-and-rm32-register.md)。

隔離 `dosgolem` 分支只補無前綴 `F5` 與 `21 /r mod=11` 的通用 CPU 形狀；合成測試與固定原檔輸入的 `go test ./internal/cpu386 ./internal/machine -count=1`、`go test ./... -count=1` 通過。已綁定 DPMI 的**合成環境診斷**由第 760 步推進至第 803 步，在 dosgolem 重定位 LE 線性 `0x151648` 的原始 bytes `83 0E 01` 失敗即關閉；`83 0E 01` 尚未核對原版前後狀態。完整 PSP／環境及遊戲資料消費端仍未閉合，**沒有正常玩家畫面或玩法同狀態對拍**；本輪未改 remake 玩法。

### 2026-10-01：`83 0E 01` 的原版記憶體寫回

**已證實，限通用 CPU 指令形狀**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，dosgolem 重定位 LE 線性 `0x151648` 的 `83 0E 01`，在 DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）以其 CS:EIP `0180:00375648 → 0037564B` 的同次 `LOG 2` 核對。原版 `DS=0188h`、`ESI=003EC034h`；`MEMDUMPBIN` 擷取 DS:[ESI] 四位元組 `90 00 00 00 → 91 00 00 00`，符合 `OR dword [ESI],01h`。完整位址、符號擴展、旗標與測試邊界見 dosgolem [規格 212](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/212-cpu386-or-rm32-imm8-memory.md)。版控 `startup_probe_131.py --or-memory` 可重生私有收據：`or-memory-registers.json` SHA-256 `333b36bbc93717e6b74a0f0abde417658dcd40b3db47cd750d6fe7b6bc9901d6`，連續 `or-memory-logcpu.txt` SHA-256 `33bff7dea3164d1af83c2c588354fd9c47f078737ea276b2653b8d3e6581781f`，前後位元組檔 SHA-256 分別為 `0e6c738e4fe755a64a276418309bb5dc7e6bf36772bc0238010718e091a780da`／`f00061f6703ccf02a5d5d1ad9d83f2d4db90c9481268364eccdada3e2214d0fe`。原版檔與完整輸出未加入 Git。

隔離 dosgolem 工具分支依 READY 規格增加無前綴 32 位記憶體 `83 /1 ib`，`go test ./internal/cpu386 ./internal/machine -count=1` 與固定原檔輸入的 `go test ./... -count=1` 通過。已綁定 DPMI 的**合成環境診斷**越過此處，由第 803 步至第 819 步，在 dosgolem 重定位 LE 線性 `0x13CC50` 的 `0F A9` 失敗即關閉；下一指令的原版行為仍待核對。PSP／環境與完整資料消費端仍未閉合，**沒有正常玩家畫面或玩法同狀態對拍**；本輪未改 remake 玩法。

### 2026-10-01：`0F A9` 的原版堆疊與 MOO2 GS 許可

**已證實，限通用指令形狀與本次 selector 載入**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；dosgolem 重定位 LE 線性 `0x13CC50` 的 bytes `0F A9`，對應 DOSBox-X 2026.07.02 SDL2 heavy debugger 的 **CS:EIP** `0180:00360C50 → 0180:00360C52`。輔助映像 `fd2-dosbox-x:debug-0d7b272b` ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。同次 `LOG 2` 配合前後堆疊四位元組擷取：`SS=0188h`、`ESP=003EBC50h → 003EBC54h`、`GS=0020h → 0020h`、`SS:[ESP]` 原始 bytes `20 00 00 00`、旗標不變。版控 `apps/moo2/tools/startup_probe_131.py --pop-gs` 可重生私有 `pop-gs-registers.json` SHA-256 `d2b6f5a108d1bf8eb035f20eb024871e5b680efc4a6f95777f09f5ad5a35ce1b`、`pop-gs-logcpu.txt` SHA-256 `4636abf3ae7a64765c94b013718b146dfd56f47a13467987299753f3099cd39e`；前後堆疊檔 SHA-256 均為 `8d71b3faab8201459ad37ef499beb336ba88bdcfa0f51ee6f0a46ec3192d750a`。原始檔及終端輸出仍只在本機。

原版 GS 在指令前後都為 `0020h`，因此不同 selector 的載入效果**不能由這個樣本單獨證明**；以 [Intel 64／IA-32 軟體開發手冊第 2B 卷](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf) 的 `POP GS` 契約與不同值合成測試補足。實作先因服務層拒絕 `0020h → GS` 再停於第 819 步；回到規格 [213](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/213-cpu386-pop-gs.md) 補充審查後，只為 MOO2 profile 開此組合，FD2 profile 及 `0020h → DS／FS` 均維持拒絕。這是載入許可，**GS 描述子 base、limit 與權限未知**。

隔離工具分支 `go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 及固定原檔作輸入的 `go test -buildvcs=false ./... -count=1` 全通過。已綁定 DPMI、仍用合成 PSP／環境的診斷從第 819 步跨過 `0x13CC50`，至第 2460 步、dosgolem 重定位 LE 線性 `0x153E84` 停於帶前綴 `SBB` 不支援；這不是正常玩家路徑或玩法同狀態對拍。下一步核對該指令形狀及對應原版證據，並持續查合成環境與完整資料消費端。

### 2026-10-01：`66 19 C0` 的原版 16 位借位減法

**已證實，限本次通用指令樣本**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，dosgolem **重定位 LE 線性位址** `0x153E84` 的原始 bytes `66 19 C0`；DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）於其 **CS:EIP** `0180:00377E84 → 0180:00377E87` 同次 `LOG 2` 顯示 `sbb ax,ax → pop edi`。原版進入 `EAX=00000600h`、CF=0、ZF=0、AF=1、PF=1，離開 `EAX=0`、CF=0、ZF=1、AF=0、PF=1；樣本沒有驗證 CF=1 或非零高 16 位。版控 `apps/moo2/tools/startup_probe_131.py --sbb-word` 可重生私有 `sbb-word-registers.json` SHA-256 `9c894626819a99de98bde84a95a6365a4544eb0f4779f56ea88e0e4f0b858eb5` 與 `sbb-word-logcpu.txt` SHA-256 `19ba3515c9b318bef01351e27c6d335405e6695c3d6649311ec9f9dc78f5883a`。原版檔與完整終端留在本機。

[Intel 手冊的 `SBB` 條目](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf) 與 dosgolem [規格 214](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/214-cpu386-sbb-rm16-register.md) 限定 `66 19 /r mod=11` 的 16 位暫存器目的形狀。合成測試核對 CF=1、不同暫存器方向、高 16 位保留及未支援形式拒絕；`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 與固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過。已綁定 DPMI、仍用合成 PSP／環境的診斷越過第 2460 步，至第 2475 步、dosgolem 重定位 LE 線性 `0x1005B` 的 `C8 AC 00 00` 停於未支援 opcode。前段經 `0x13EF5C → 0x10018 → 0x10057`；**這條低位址控制流與 C8 停點尚未由原版獨立核對**，不可直接當下個原版行為需求。正常玩家畫面與 remake 同狀態對拍仍沒有收據。

### 2026-10-01：低位址控制流與 `ENTER 00AC,00`

**已證實，限固定原版指令形狀與本次執行狀態**：1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）的 **CS:EIP** `0180:00362F5C` 呼叫 `0180:00234018`，經 `00234057` 及四次 push，進入 `0180:0023405B` 的 `enter 00AC,00`。版控 `apps/moo2/tools/startup_probe_131.py --low-entry` 重生的私有 `low-entry-registers.json` SHA-256 `2fd82053b138d260ed49b1675377f3106f9eb4ccaea8be20fbb0dabcfbc0b6fc`，有界 `low-entry-logcpu.txt` SHA-256 `9d242f4d359cc331fbef1aad02a646ddf72e30a88eb09e7dd7e0641f883b5855`。這直接核對前輪 dosgolem **重定位 LE 線性位址** `0x13EF5C → 0x10018 → 0x10057 → 0x1005B` 的指令順序；兩工具位址基準不同，不能只比數字。

原版同次 `--enter` 的 **CS:EIP** `0180:0023405B → 0180:0023405F`：`SS=0188h`、`ESP=003EBC90h → 003EBBE0h`、`EBP=003EBCA4h → 003EBC8Ch`、`EFLAGS=0246h` 不變，`SS:003EBC8C` 四位元組 `90 20 3A 00 → A4 BC 3E 00`，寫入舊 EBP。私有 `enter-registers.json` SHA-256 `f0081c2f07edfaed125a5dc6b95de4fd404200953c009db55046b3239f4d7169`、`enter-logcpu.txt` SHA-256 `1d458daf1ebe9d383baf105f9e386277909f6443419a21a16ae3beea56331ecb`，前後堆疊檔 SHA-256 `1a2db19b7f6380c9eed28e130e8224fe9245ab0eb0613af0f8254c02341269bf`／`ec5fc2e2555302df262bcda9052310184eda069fbd22cd9ffa4df4a01874d205`。原版 EXE 與完整終端留在未版控工作區。

[Intel 手冊的 `ENTER` 條目](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2a-manual.pdf) 與 dosgolem [規格 215](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/215-cpu386-enter-level-zero.md) 將實作限定為無前綴、32 位堆疊、巢狀層級 0。首版為預檢頁面映射額外讀取堆疊，經審查認定不符通用匯流排契約，退回 DRAFT 後移除；修正版以段描述子檢查權限與界限，並以會拒絕堆疊讀取的合成匯流排測試。固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過；已綁定 DPMI、仍用合成 PSP／環境的診斷越過第 2475 步，至第 2512 步、dosgolem 重定位 LE 線性 `0x109FF` 的 `66 3B 4D CE` 停下。**該記憶體比較尚未由原版獨立核對**；沒有正常玩家路徑或與 remake 同狀態玩法收據。

### 2026-10-01：`66 3B 4D CE` 的原版 16 位記憶體比較

**已證實，限固定原版指令樣本**：1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）於其 **CS:EIP** `0180:002349FF → 0180:00234A03` 的同次 `LOG 2` 記錄 `cmp cx,[ebp-0032] → jl`。進入時 CX=`0000h`、EBP=`003EBB3Ah`、SS=`0188h`，來源 `SS:003EBB08` 兩位元組 `01 00`；離開時 CX／EBP 與來源 bytes 不變，EFLAGS=`0246h → 0297h`。dosgolem 對應的**重定位 LE 線性位址**是 `0x109FF`，bytes `66 3B 4D CE`；兩工具位址基準不同。

版控 `apps/moo2/tools/startup_probe_131.py --cmp-word` 可重生私有 `cmp-word-registers.json` SHA-256 `b4df0818c2059df4bef4a29c76508908e10e05ba9f8730b7102197da1a9b3969`、`cmp-word-logcpu.txt` SHA-256 `eaf818725b21cc1b76ca9830b81f0da46587ac91401a4dd0ee2d652a8f9a4c51`；`cmp-word-before.bin` 與 `cmp-word-after.bin` SHA-256 均為 `47dc540c94ceb704a23875c11273e16bb0b8a87aed84de911f2133568115f254`。首輪擷取曾把來源寫為 DS；檢查位址解碼與 DOSBox-X 的 `ss:[...]` 註記後，腳本改用 SS 並重生全部收據。原版恰好 DS=SS=`0188h`，故數值一致本身不能證明預設段；[Intel 手冊的 `CMP` 條目](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2a-manual.pdf)、既有 `decodeAddress32` 與不同 DS／SS base 的合成測試補足此界線。原版檔與完整輸出留在未版控工作區。

隔離 dosgolem 工具分支依 [規格 216](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/216-cpu386-cmp-word-register-memory.md) 接入無段覆寫／repeat 的 `66 3B /r` 16 位記憶體來源。合成測試核對 EBP 選 SS、非 EBP 選 DS、記憶體與暫存器不變、段界限及截短位移拒絕；`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 與固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過。已綁定 DPMI、仍用合成 PSP／環境的診斷由第 2512 步推進至第 2600 步，於 dosgolem 重定位 LE 線性 `0x126570` 的 `66 A9 89 CF` 停下；該 `TEST` 尚未由原版獨立核對。這不是正常玩家路徑或 remake 玩法同狀態對拍，本輪沒有改 remake 玩法。

## 舊資料頁基址的勘誤

## 2026-09-30 通用 CPU 指令續驗

**已證實（僅限所列指令形狀）**：`dosgolem` 隔離分支依規格 203–206 增量支援 `26 66 8E 1D` 的 ES 覆寫 16 位絕對記憶體段載入、`3E B9` 的 DS 前綴暫存器立即數、`26 3A 10` 的 ES 覆寫 byte 比較，以及 `0F A8` 的 32 位堆疊 `PUSH GS`。合成測試分別核對描述子 base／界限、前綴不影響無記憶體運算元、比較旗標、堆疊寬度與失敗原子性。`go test ./internal/cpu386 ./internal/machine` 通過。輸入仍是 SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` 的未修改 1.31 EXE；工具的位址基準是 dosgolem 重定位 LE 線性位址。

**診斷，非原版玩家路徑收據**：沿用規格 201 的兩次啟動返回與合成 PSP／環境，固定真檔探針依序從第 72 步跨至第 125 步 `0x11014C`、第 393 步 `0x163A90`、第 406 步 `0x13CB80`，本輪至第 419 步 `0x13CBAA` 停於 `66 83 E7 FC`（未支援的 16 位 `83` 形狀）。步數是 CPU 工具缺口定位，不能推論 DOS 啟動、畫面或玩法已完成對拍；無服務 hook 的正式 fail-closed 入口仍是第 11 步 `INT 21h/AH=30h`。後續先核對 MOO2 環境／PSP 消費端與服務狀態，再以 dosgolem 重生玩家檢查點，不能把 DOSBox-X 輔助基準當正式收據。

**續驗，仍為合成環境診斷**：規格 207 補 `66 83 E7 FC` 的通用 16 位暫存器 `AND`，合成 CPU 測試及 `go test ./internal/cpu386 ./internal/machine` 通過。固定 1.31 原檔進至第 554 步，在 dosgolem LE 線性位址 `0x15E07C` 停於 `CD 21`；呼叫前 `EAX=4A88h`、`EBX=1CF00h`、`ECX=1CDCD0h`、`EDX=1CECD0h`、`ES=DS=SS=0188h`、`EFLAGS=0206h`。`AH=4Ah` 屬 DOS 記憶體區塊縮放服務，但此處 protected-mode 返回、區塊 owner 與合成 PSP 的可比性仍**未知**；沒有新增猜測回傳。

曾在既有 DOSBox-X 輔助腳本的兩次已證實啟動返回後暫掛 `BPINT 21 4A`，期望擷取同一 `AX=4A88h`。這次固定原檔執行在設定斷點後切換顯示模式，除錯終端不再回傳可解析的 `EV` 快照，12 次有界 `RUN` 均無匹配；容器以失敗結束。這只證明**該擷取方法在此路徑失效**，不能證明原版沒呼叫 `AH=4Ah` 或其返回值。已撤回新增擷取段，保留原有可重播的兩次服務腳本。失敗的私有 `terminal.raw`／`commands.json` 留在未版控的 `workplace/dosbox-moo2/`，未升格為正式收據。

第二次受控輔助嘗試改在兩次已證實返回後，依已核對位址差 `+0x224000` 設候選斷點 DOSBox-X `0180:0038207C`，對應 dosgolem LE 線性 `0x15E07C`。顯示模式切換後 `EV` 仍回空快照，無法區分候選位址未經過與除錯終端失焦；同樣撤回實驗性腳本增量。**此位址差只由前兩處交叉定位支持，第三處尚未證實。** 現行正式輔助腳本仍只重生兩次啟動服務返回，`AH=4Ah` 返回維持未知。

上一輪 `LoadLEAt(original, 0x292E4)` 雖讀對 LE **標頭**與 fixup 表，卻把 `DataPagesOffset` 當整個檔案的絕對偏移，錯從 `0x6F000／0x6F040` 複製物理頁面。先前 `0x10FF18` 的 `66 3B 15...`、重定位 object SHA-256 `c32b38c5...`／`137ee072...`、5,359 步及 `0x10FF43 POP EDX` 均是錯頁診斷，**不得當原版啟動收據**。原先「缺少 DOS/4GW 外層堆疊」假說因此撤回；通用 `66 3B 15` CPU 指令測試仍是有效的獨立工具測試。這次沒有接入音訊、輸入、固定 seed 或 remake 同狀態比較。

## IDA 位址基準核對

使用 `ida-pro-9.4-idapython:locked-v1` 查唯讀原版資料庫的拋棄式副本。輸入為私有 `Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA 記錄的原始輸入 MD5 是 `bacb10a92454d2f9b211eb9fe67ec099`，與上述 1996 光碟版 `Orion2.exe` 的 MD5 相同。查詢輸出留在未版控的 `workplace/ida-probe/cstart.json`。

**已證實（IDA 線性位址空間與原檔偏移）**：IDA 在 `0x10FF18` 顯示 `EB 76 57 41 54 43 ...`，標為 `start`，並映至原檔 `0x1955AC`。以內嵌 MZ 基址修正 dosgolem 載入後，其重定位 LE 線性位址 `0x10FF18` 的 bytes 相同；舊版 `66 3B 15...` 差異已由錯誤資料頁基址解釋。IDA 的函式名稱仍只作導航；這次原檔偏移、bytes、LE object+offset 與兩工具位址已逐項核對，後續交叉參照仍須附上各自位址基準。
