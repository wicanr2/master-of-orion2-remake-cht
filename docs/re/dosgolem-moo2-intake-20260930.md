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

### 2026-10-01：`66 A9 89 CF` 與下一個 DTA 服務停點

**已證實，限固定原版指令樣本**：1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`（ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）在其 **CS:EIP** `0180:0034A570 → 0180:0034A574` 的同次 `LOG 2` 顯示 `test ax,CF89 → pop es`；原版 EAX=`003EBB00h` 不變、EFLAGS=`0206h → 0286h`。dosgolem 對應 **重定位 LE 線性位址** `0x126570` 的 bytes 為 `66 A9 89 CF`。`BB00h & CF89h = 8B00h`，SF=1、ZF=0、PF=1；CF／OF 的一般規則與 AF 未定義見 [Intel 手冊 TEST 條目](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf) 及 dosgolem [規格 217](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/217-cpu386-test-ax-imm16.md)。版控 `startup_probe_131.py --test-word` 重生私有 `test-word-registers.json` SHA-256 `b907d497708f9af5e4f671ab5b4d3d3d15041aac383b4620c7ba6473ed59f2a5`、`test-word-logcpu.txt` SHA-256 `e624bc4be5bd0f955568077c3f4fcfc711927baf90ba94ca6a85254be805ceab`。

隔離 dosgolem 工具分支補無段覆寫／repeat 的 `66 A9 iw`；合成測試核對高 16 位保留、旗標、截短立即數與前綴拒絕。`go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 及固定原檔輸入的 `go test -buildvcs=false ./... -count=1` 全通過。已綁定 DPMI、仍用合成 PSP／環境的真檔診斷由第 2600 步至第 4062 步，在 dosgolem **重定位 LE 線性位址** `0x139A53` 的 `CD 21` 停下；進入 EAX=`00171A99h`（AH=`1Ah`）、EDX=`001A5828h`，服務層回報未處理。這是工具診斷，不是同狀態玩法收據。

**已證實，限原版 DOSBox-X 輔助呼叫邊界**：版控 `startup_probe_131.py --dta` 在原版 **CS:EIP** `0180:0035DA53 → 0180:0035DA55` 命中 `INT 21h/AH=1Ah`；原版 EAX=`00381A99h`、EDX=`003C3828h`、DS=`0188h`、EFLAGS=`0246h` 前後不變。私有 `dta-registers.json` SHA-256 `3e58143819ae32406a06027591a19bb14abfdf99eec644c8363886ef4048d68b`。兩側呼叫形狀相符，但 EAX 高位及 DTA 位址明顯不同；合成 PSP／環境與完整資料消費端尚未對齊，不能照抄原版指標或從未變旗標推論所有呼叫都成功。下一步須先定義 DTA 指標保存、後續 `AH=4Eh／4Fh` 消費及檔案來源的受保護模式契約，再進入服務規格審查。原版 EXE 與完整終端保持私有，沒有正常玩家畫面或 remake 玩法同狀態對拍。

## 2026-10-01：動態服務與 CPU 補證

### 2026-10-01：受保護模式 DTA 首次搜尋與缺檔勘誤

**已證實，限 DOSBox-X 2026.07.02 輔助原版樣本**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；映像 `fd2-dosbox-x:debug-0d7b272b` ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。原版 **CS:EIP** `0180:0035DA53 → 0035DA55` 的 `AH=1Ah` 設 DS:`003C3828h` DTA，接著 `0180:0035DA59 → 0035DA5B` 以 DS:`0038E099h` 的 `MOX.SET`、CX=0 執行 `AH=4Eh`。此處若缺檔，EAX=`00384E99h → 00000012h`、EFLAGS=`0246h → 0247h`；DTA 前 12 bytes 由全零改為 `02 4D 4F 58 00 00 00 00 00 53 45 54`，後 31 bytes 維持零。故「失敗不動 DTA」的 16 位路徑假設**不可套在此受保護路徑**。`--dta-find` 可重生私有 `dta-find-registers.json` SHA-256 `1416ebb092555aa7c40a64ef7972caff19b311934797d8792e9fe7b5b18c529e`、前後 DTA SHA-256 `859732b97382a08583d6a67f5842486505e50bee754bd9b57ac3abf81b9714f2`／`3a702e4d5eedf367556099638460764faa3d587607db9c313319c3ceb1c8f1a7`；下一段有界控制流 SHA-256 `a120eceae2a33819ceb3a6fe2d2c8b2b50e82e2e5604314337745488a55c1060`。

**已證實，限受控合成檔案存在樣本**：一次性容器除原版 EXE 外僅新增零位元組 `MOX.SET`（SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`），修改時間固定 `1996-01-01 00:00 UTC`。同一呼叫回 EAX=`00380000h`、EFLAGS=`0246h`；DTA `+15h` 屬性 `20h`、`+16h` 時間 `0000h`、`+18h` 日期 `2021h`、`+1Ah` 大小 0、`+1Eh` 名稱 `MOX.SET\0`。`--dta-find-present` 可重生私有暫存器 SHA-256 `ca68fb5618b84b667b749e5fe6874fc2c42a26e3b460c13d6c00cd71a1898e2b`、DTA 後態 `a36f4b577def2164c9b875189aaf1519863de0b7d115f0c7ecaac4f389effe4e`、後續有界控制流 `2a7c41a13ce549af44afff0e9da2943d434a15141f76902652f13342a80b4b7a`。完整擷取與原檔不入 Git。版本化入口在 dosgolem [`startup_probe_131.py`](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/apps/moo2/tools/startup_probe_131.py)；服務契約與未知範圍見 [規格 218](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/218-protected-dos-set-dta.md)／[規格 219](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/219-protected-dos-findfirst-exact.md)。

**強推論／近似**：缺檔保留既有 DTA 結果欄由 [DOSBox-X `SetupSearch`／`SetResult`](https://github.com/joncampbell123/dosbox-x/blob/master/src/dos/dos_classes.cpp) 支持，原版這次舊欄位恰為零；工具另以非零合成欄位測試。檔案修改時間採 UTC 打包，只是明示的環境近似。dosgolem **重定位 LE 線性位址** `0x139A53`、`0x139A59` 對應服務呼叫，數值不與上方 DOSBox-X CS:EIP 混用。固定真檔、合成 PSP／環境診斷經 `AH=1Ah`、`4Eh` 進至第 4168 步，於重定位 LE 線性 `0x148224` 的 `38 10` 停在未支援 CPU byte 比較。PSP／環境與完整遊戲資料未對齊，**沒有正常玩家路徑或 remake 同狀態玩法收據**；`CMP` 的原版對應狀態仍未知。

### 2026-10-01：`38 10` 的原版 byte 比較

**已證實，限本次原版樣本**：同一固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b` ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，於其 **CS:EIP** `0180:0036C224 → 0180:0036C226` 同次 `LOG 2` 顯示 `cmp [eax],dl → test al,03`。進入 EAX=`003EC0DCh`、EDX=`0`、DS=`0188h`、EFLAGS=`0202h`；DS:`003EC0DCh` 一位元組 `00h → 00h`，離開 EAX／EDX／DS 不變、EFLAGS=`0246h`。dosgolem 對應 **重定位 LE 線性位址** `0x148224` 的 bytes `38 10 A8 03`；位址基準不同。版控 `startup_probe_131.py --cmp-byte` 重生私有 `cmp-byte-registers.json` SHA-256 `c9763a12d3aa3af48b8c878e97c983ede9c8fb771471f31c1ef8ef90e142b72d`、`cmp-byte-logcpu.txt` SHA-256 `12adee632c97e7ae0d6ddcbd09a393c988602d0329ef7b00c90fe45a3e05e32b`；記憶體前後檔 SHA-256 均為 `6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d`。原檔與完整終端仍在未版控工作區。

**通用契約與界線**：[Intel 手冊的 `CMP` 指令表](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2a-manual.pdf) 定義 `38 /r` 為 `CMP r/m8,r8`；不同值、EBP／SIB 預設段與失敗處理由手冊、既有位址解碼器及合成測試驗證，**不是這次原版零差樣本證實**。隔離 dosgolem 分支依 [規格 220](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/220-cpu386-cmp-rm8-register.md) 將無前綴記憶體目的路由到 `decodeAddress32`；固定原檔 `go test -buildvcs=false ./... -count=1` 通過。合成 PSP／環境診斷至第 4944 步，在重定位 LE 線性 `0x146903` 的 `26 8A 1E 42 84` 停於未支援 ES 覆寫載入；其原版對應尚**未知**，無正常玩家路徑或同狀態玩法對拍。

### 2026-10-01：`26 8A 1E` 的原版 ES 覆寫 byte 載入

**歷史取樣勘誤：下段的 EV 前態與 LOG 後態配對已撤回，應以下節的連續 LOG 為準。**

**已證實，限本次原版指令形狀及執行狀態**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b` ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，於其 **CS:EIP** `0180:0036A903 → 0180:0036A906` 的同次 `LOG 2` 擷取 `mov bl,es:[esi] → inc edx`。執行前 EBX=`0000000Fh`、ESI=`003EBA74h`、ES=DS=`0188h`、EFLAGS=`0246h`，ES:`003EBA74h` 的來源 byte `30h`；執行後 EBX=`FFFFFF30h`，ESI、ES、DS、旗標及來源 byte 不變。版控 dosgolem `apps/moo2/tools/startup_probe_131.py --es-byte-load` 重生私有 `es-byte-load-registers.json` SHA-256 `8312772e1b0de3fff25a493cbdd99847ee2b9a8b97d62b24b5ed22061b428511`、`es-byte-load-logcpu.txt` SHA-256 `f751a6bc439f1380b3eaa5adeccb196d6e18e4f0a2000dc2398bec340aed9837`；來源前後 byte 檔 SHA-256 均為 `5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9`。原版檔與完整擷取留未版控工作區。第一次探針誤把 `26 8A 1E` 當成五 byte、期待 `0036A908`；原版實到 `0036A906`，修正為三 byte 後以同一映像乾淨重跑。

**通用契約與限制**：dosgolem 對應 **重定位 LE 線性位址** `0x146903`，其中 `26 8A 1E` 才是本條指令，後面的 `42 84` 屬後續位元組；兩工具的位址空間不得混同。由於原版 ES=DS，這個樣本**不能單獨證明 ES 覆寫的異段效果**；[Intel 手冊第 2A／2B 卷](https://www.intel.com/content/www/us/en/developer/articles/technical/intel-sdm.html)及異段基址合成測試補足通用 CPU 契約。隔離 dosgolem 分支的 [規格 221](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/221-cpu386-mov-r8-es-memory.md) 經 DRAFT→READY→實作→CONFORMED；固定原檔的 `go test -buildvcs=false ./internal/cpu386 ./internal/machine -count=1` 與 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過。合成 PSP／環境診斷由第 4944 步至第 5392 步，於 **dosgolem 重定位 LE 線性位址** `0x14822D` 的 `C1 CA 08` 停住；該新指令尚未由原版獨立核對。私有診斷 `workplace/moo2-probe-221.txt` SHA-256 `7f1d74911be0ea3b1ebf4cf16e02661cefc54b1d3bcfda41eabbbf4923b3ebc1`，當時全套測試輸出 SHA-256 `1babab5727e97dfbb7db6d396ded1c4b3c7fc3479ae3d4aa82d5e176a23e5a99`。仍無正常玩家路徑或玩法同狀態對拍；原版 ESP／環境與合成 PSP／環境差異未解。

### 2026-10-01：ES byte 載入的 `EV`／`LOG` 取樣勘誤與有限檢查點

**勘誤，已證實的矛盾**：上節把候選 `EV` 前斷點 EBX=`0000000Fh` 與 `LOG 2` 下一行 EBX=`FFFFFF30h` 寫成同一次 `MOV BL,[ESI]` 的前後態，這會無故改變 EBX 高 24 位，不能成立。固定 1.31 原檔 SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，同一 DOSBox-X 2026.07.02 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582` 的獨立 `--es-byte-load-ev` 前後斷點重播仍得 `0Fh → FFFFFF30h`；私有 `es-byte-load-ev-registers.json` SHA-256 `e33ff4052f0202d66a23959fc1f0221c9a6d5cd27fc514f3f5f74a6f29680606`。兩個 `EV` 斷點是否為同一次動態執行及其不一致原因仍**未知**，不可把它們升格成指令前後收據。修正版探針 `--es-byte-load` 在私有 JSON 中明寫候選 EV 前態與 LOG 首行 EBX 衝突；該 JSON SHA-256 `3d3fd5422f193a9a0e1246ffb682d473383351302f71cfcfd1c30cb10d4930f5`。

**已證實，限連續 LOG 的單一指令樣本**：原版同次 `LOG 2` SHA-256 `f751a6bc439f1380b3eaa5adeccb196d6e18e4f0a2000dc2398bec340aed9837` 在 **DOSBox-X CS:EIP** `0180:0036A903 → 0036A906` 顯示 EBX=`FFFFFFFFh → FFFFFF30h`、ESI=EDX=`003EBA74h`、ESP=`003EBA40h`、來源 `30h`、ES=DS=SS=`0188h`、旗標不變。版控 dosgolem `internal/machine/TestMOO2ESByteLoadCheckpointWhenProvided` 由固定原檔自行重生**重定位 LE 線性位址** `0x146903` 第 4944 步：EBX=`FFFFFFFFh → FFFFFF30h`、ESI=EDX=`001CDA94h`、ESP=`001CDA60h`、來源 `30h`、ES=DS=SS=`0188h`、EFLAGS=`0246h`。兩側 ESI／ESP 在此檢查點各差 `0021DFE0h`，ESI−ESP 同為 `34h`；這只支持**有限啟動指令的正規化狀態對照**。加入整合測試後固定原檔重跑 `go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-222.txt` SHA-256 `c3f703ceca457a2303656286d89a443f7737f09c5a9b8ba623accd26cf8fd9ba`。原版環境、DTA 指標、完整資料、玩家畫面及玩法同狀態仍未知；不能用此局部一致宣稱正常玩家路徑已通過。

### 2026-10-01：零輸入 ROR 立即數 8 的有限啟動對照

**已證實，限原版零輸入樣本**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，於其 **CS:EIP** `0180:0036C22D → 0036C230` 的同次 `LOG 2` 顯示 `ror edx,08 → dec ecx`；EDX=`0 → 0`，可見旗標進出均為 `0206h`。版控 dosgolem `apps/moo2/tools/startup_probe_131.py --ror-imm8` 重生私有 `ror-imm8-logcpu.txt` SHA-256 `2b49614290d647c1ed8ead3336ee49a705d5eb2581d2a9fd5213a79b23136bcc`、`ror-imm8-registers.json` SHA-256 `391c4afc435f8aac41c7316003e2208c1114e7a6748a99ee5feab4c19b9dfd41`。候選 EV 前態只供定位；前後值以同次連續 LOG 為準。原版非零輸入尚未實測，不能由零樣本推論其 CF 或未定義 OF。

**已證實，限合成環境下的原檔診斷**：dosgolem [規格 222](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/222-cpu386-ror-r32-imm8.md) 依 DRAFT→READY→實作→CONFORMED 限定 `C1 /1`、32 位暫存器、立即數 `08h`；非零契約依 [Intel 指令手冊](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf)與合成測試，OF 未定義時保留舊值只是決定性約定。固定原檔第 5392 步在**重定位 LE 線性位址** `0x14822D` 的 EDX=`0`、EFLAGS=`0206h`，單步到 `0x148230` 且狀態不變；其後至第 5529 步停於同位址基準 `0x14701B` 的 `08 E0`，尚未由原版獨立核對。私有 `workplace/moo2-probe-222.txt` SHA-256 `3801d8c21e2a3f9251b766817de4946c953340aa6e5cac74ec449b43bf123eca`。含固定原檔的 `go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-223.txt` SHA-256 `e885b326c78ac7c279c7021f0a56202ead68326f3ead9b6d2d055eb2d4cb25ca`；新整合測試另以 `-v` 確認執行通過。兩工具位址空間不同；合成 PSP／環境與原版仍不同，**沒有正常玩家路徑或玩法同狀態收據**。

### 2026-10-01：AH 零輸入 OR 的有限啟動對照

**已證實，限原版當次 AH 零輸入**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，於其 **CS:EIP** `0180:0036B01B → 0036B01D` 的同次 `LOG 2` 顯示 `or al,ah → mov edi,FFFFFFFF`；EAX=`00000001h → 00000001h`、EFLAGS=`0202h → 0202h`。版控 dosgolem `apps/moo2/tools/startup_probe_131.py --or-al-ah` 可重生私有 `or-al-ah-logcpu.txt` SHA-256 `7dc04fd5696dc56a37915d397cd758b0067c75c2b3ff64f5434077ef4a747a8b` 與 `or-al-ah-registers.json` SHA-256 `a7006a6055ecebffb66c8dc6f989a48b3ecbd5e7191c03a5a8346cd4e407fbfa`。候選 EV 前態僅供定位；以同次連續 LOG 判斷前後。非零來源、記憶體形式與未定義 AF 均不由此原版樣本證實。

**已證實，限合成環境下的原檔診斷**：dosgolem [規格 223](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/223-cpu386-or-rm8-register.md) 經 DRAFT→READY→實作→CONFORMED，僅接無前綴 `08 /r`、`mod=11` 的 8 位元暫存器來源與目的；非零行為依 [Intel OR 指令表](https://cdrdv2-public.intel.com/835752/253667-sdm-vol-2b.pdf)與合成測試，AF 清零是既有決定性約定。固定原檔第 5529 步在**重定位 LE 線性位址** `0x14701B` 的 EAX=`1`、EFLAGS=`0202h`，單步至 `0x14701D` 後不變。診斷接著在第 5806 步停於同位址基準 `0x15C1DF` 的 `2E 8D 86 82 C2 15 00`，該新停點尚未由原版獨立核對；私有 `workplace/moo2-probe-223.txt` SHA-256 `7bd4203641b216f68fd8117a1772a24cca26441d1937033127f5b88bcbed16bc`。含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-224.txt` SHA-256 `a7e03d6798acdf2df438d0bd94ed99827b2f7c858f3efec2338ae0eec4fe8f61`；新 CPU 與原檔整合測試另以 `-v` 確認執行。兩工具位址空間不同，合成 PSP／環境與原版仍不同；**沒有正常玩家路徑或玩法同狀態收據**。

### 2026-10-01：CS 前綴 LEA 與合成環境暫存器差異

**已證實，限原版當次 LEA 指令**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，於其 **CS:EIP** `0180:003801DF → 003801E6` 的同次 `LOG 2` 顯示 `lea eax,cs:[esi+00380282] → push eax`；進入 ESI=`93h`、EAX=`31h`、EFLAGS=`0212h`，離開 EAX=`00380315h`，正是 `93h+00380282h`，旗標不變。版控 dosgolem `apps/moo2/tools/startup_probe_131.py --lea-cs` 重生私有 `lea-cs-logcpu.txt` SHA-256 `44697865de21f50b0417f3924b52878aec2988a62dbc045ab4113befdacc7f2c` 及 `lea-cs-registers.json` SHA-256 `5e6bf786def83d5bcc73bb49f015bf42c901c34e4355f244ca05e01a71362adb`。候選 EV 前態只供定位，前後以同次 LOG 為準。原版只實測 CS 前綴；其他段前綴依 [Intel LEA 指令契約](https://cdrdv2-public.intel.com/789581/325383-sdm-vol-2abcd.pdf)與合成測試，不冒稱原版實測。

**已證實，限合成 PSP／環境下的原檔診斷**：dosgolem [規格 224](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/224-cpu386-lea-segment-prefix.md) 經 DRAFT→READY→實作→CONFORMED，只讓 `8D /r` 的單一 CS／DS／ES／SS 段前綴通過，不將段基址加入 LEA 的有效位址。固定原檔第 5806 步在**重定位 LE 線性位址** `0x15C1DF` 進入時 ESI=`99h`、EAX=`33h`、EFLAGS=`0016h`，以自身狀態算出 EAX=`15C31Bh`、下一 EIP=`0x15C1E6`，非目的暫存器與旗標不變。前態私有 `workplace/moo2-probe-224-before.txt` SHA-256 `029288f59181b5ab9299cfde239b176d6923fad1b90a805d9f37b75c0c31606a`。原版與合成環境 ESI 差 6、EAX／旗標也不同；原因未知，合成 PSP／環境長度僅是待查候選，**不能以這筆宣稱同狀態**。兩工具位址空間不同。

含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-225.txt` SHA-256 `0a64b5aaa486fe09688bfa7453952c1f75866b6932c45d3d606af965d7f00469`；新 CPU／原檔整合測試另以 `-v` 確認執行。合成環境診斷在第 5808 步停於同位址基準 `0x15C1E7` 的 `8E 03`，原版對應待查；私有 `workplace/moo2-probe-224-after.txt` SHA-256 `23bb76d25b0ce27fb659f558402547cce708676dacda849a9ae5206ad16f12f0`。原版 EXE 與完整終端留未版控工作區；仍無正常玩家路徑或玩法同狀態收據。

### 2026-10-01 勘誤：LEA 取樣到不同呼叫次數

**已證實，限固定 1.31 原檔的這段連續啟動記錄**：先前把 `--lea-cs` 單點斷點所得 EAX=`31h`、ESI=`93h` 與 dosgolem 第 5806 步 EAX=`33h`、ESI=`99h` 比較，並猜測差異可能來自環境長度；該筆原版收據本身仍有效，但它不是對應的**第一次**呼叫。DOSBox-X 2026.07.02 SDL2 heavy debugger 同一映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，從 **CS:EIP** `0180:00348105` 執行版控 `startup_probe_131.py --startup-value` 的連續 `LOG 80h`：`0180:0034816D` 的立即數 `mov eax,33h`，到 `0180:00363258` 的 `mov esi,eax`，再到 `0180:003801DC` 的 `lea esi,[esi+esi*2]`，於**第一次** `0180:003801DF` 得 EAX=`33h`、ESI=`99h`，下一 `0180:003801E6` EAX=`0038031Bh`。`0180:003801E7` 確為 `mov es,[ebx]`，是 dosgolem 下一停點 `8E 03` 的原版對應；當次來源 DS:[EBX] 低字為 `0188h`，ES 原值 `0188h`。私有 `startup-value-logcpu.txt` SHA-256 `f209eb9b55ec3c9184f19f335142880e747bb68848ddf091c6764066f7dc954c`、`startup-value-registers.json` SHA-256 `bf96d8bebcb9169c43466c4b3216c0f294ffe783478610c02677e4075c28f25c`；輸入仍為 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 的 `LOG 80` 參數按十六進位解讀，實得 128 行。dosgolem 的重定位 LE 線性 `0x15C1DF` 第 5806 步也由固定原檔重生 EAX=`33h`、ESI=`99h`；其私有補充追蹤 `workplace/moo2-esi-changes-225.txt` 記錄立即數 `0x12416D`、傳入 `0x13F258`、乘三 `0x15C1DC` 的同一路徑。這只訂正來源值與呼叫次數；兩邊的堆疊、EFLAGS、記憶體佈局及合成 PSP／環境仍不同，**不是完整同狀態或玩法對拍**。原先「ESI 差 6 源於環境長度」的假說撤回；舊單點記錄不刪除，保留其取樣脈絡。

### 2026-10-01：DS:[EBX] 載入 ES 與滑鼠中斷停點

**已證實，限原版啟動指令及 dosgolem 合成診斷**：前段固定原檔、DOSBox-X 2026.07.02 SDL2 heavy debugger（映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`）的連續 LOG，在其 **CS:EIP** `0180:003801E7 → 003801E9` 顯示 `mov es,[ebx]`，DS:[EBX] 低 word=`0188h`，ES 保持 `0188h`。隔離 dosgolem [規格 225](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/225-cpu386-mov-es-ds-ebx.md) 經 DRAFT→READY→實作→CONFORMED，僅支援無前綴 `8E 03`，使用既有段描述子與 selector 驗證。固定 1.31 原檔由 LE entry、合成 PSP／環境自然至第 5808 步、**dosgolem 重定位 LE 線性位址** `0x15C1E7`；EBX=`0x1CDB48`、來源低 word=`0188h`，單步至 `0x15C1E9`，ES、通用暫存器及 EFLAGS 符合規格。合成拒絕／來源越界測試及固定原檔的全套 `go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-226.txt` SHA-256 `8c5966816d8635d21bf5c0f2c763a9c8412302e37b373ca4ff4e10743d5c578e`。有界探針接著於第 5818 步、同位址基準 `0x15C31B` 的 `CD 33` 失敗即關閉；私有 `workplace/moo2-probe-225-after.txt` SHA-256 `9a2ecd6a1d4d9237df587490139ccc4529889753dfed727686e0b929ab1edb17`。原版連續 LOG 的 **CS:EIP** `0180:0038031B` 也進入 `int 33`；其呼叫／返回、裝置狀態與玩家結果尚未審查。這不是正常玩家畫面或玩法同狀態收據。

### 2026-10-01：受控滑鼠查詢、record 消費端與 POP 記憶體目的

**已證實，限固定 1.31 原檔與指定 DOSBox-X 輔助環境**：輸入 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。版控 `apps/moo2/tools/startup_probe_131.py --mouse-query` 從原版 **CS:EIP** `0180:00348105` 連續執行到 `0180:0038031B int 33`，前態 EAX=`3`、EBX=ECX=EDX=`0`、EFLAGS=`0006h`；返回斷點 `0180:0038031D` 的 EAX=`3`、EBX=`0`、ECX=`0140h`（320）、EDX=`0064h`（100）、EFLAGS=`0006h`。接著連續 `LOG 20h` 於 `0180:003801B8–003801C0` 把 EAX／EBX／ECX／EDX 寫到原版 record `+0/+4/+8/+0Ch`。私有 `mouse-query-logcpu.txt` SHA-256 `f209eb9b55ec3c9184f19f335142880e747bb68848ddf091c6764066f7dc954c`、`mouse-query-return-logcpu.txt` SHA-256 `f78b6b183041b58bdebbfdcd7c0a9495126b3a8eb8b756c5582ac80528cbb7bd`、`mouse-query-registers.json` SHA-256 `a3d2148d16441bc034176c6c09e5b52031b754b0d08b48c8e79a1ca10b98abcc`。固定原版回傳只證實此時沒有按鍵、座標 320／100；不證實其他解析度或實機的普遍初始座標。

**已證實，限 dosgolem 合成 PSP／環境診斷與局部原檔消費端**：隔離 dosgolem [規格 226](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/226-moo2-protected-mouse-query.md) 經 DRAFT→READY→實作→CONFORMED，只對明示 MOO2 設定回應 `INT 33h/AX=3`，受控初值為 buttons=`0`、X=`320`、Y=`100`；可明示改動輸入，其他功能及 FD2 設定仍拒絕。固定原檔自 LE entry 在第 5818 步、**重定位 LE 線性位址** `0x15C31B` 單步得 EBX=`0`、ECX=`0140h`、EDX=`0064h`，EAX 與 EFLAGS=`0016h` 不變；隨後 dosgolem 原檔整合測試在 record `+0/+4/+8/+0Ch` 自行讀回 `3、0、320、100`。這是受控服務與資料消費的有限對照，原版與合成環境 EFLAGS、堆疊和 PSP 仍不同。全套 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-228.txt` SHA-256 `47dc1f8bd19688d9f2bf5d38637770bba02cdc74cd7f380016b76ffb6c3ea109`。

**已證實，限 CPU 指令形狀；原版非零來源仍未知**：原版同次返回 LOG 的 **CS:EIP** `0180:003801C6 → 003801C9` 為 `pop dword [edi+0014]`，ESP=`003EB96Ch → 003EB970h` 且旗標不變；來源堆疊 dword 未獨立擷取，不能冒稱非零搬移原版實測。dosgolem [規格 227](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/227-cpu386-pop-edi-disp8.md) 依 CPU 手冊與非零合成測試實作無前綴 `8F 47 disp8`，固定原檔第 5830 步在**重定位 LE 線性位址** `0x15C1C6` 自行核對 SS:[ESP] 至 DS:[EDI+14h] 的搬移及 ESP 加 4。合成診斷現至第 5838 步、同位址基準 `0x15C1D6` 的 `66 8C 03` 失敗即關閉，私有 `workplace/moo2-probe-227-after.txt` SHA-256 `3becc0caf1cd574f0818da4e265a4358c163caa08f86c1f4b6b91a2ae468db50`；原版對應 **CS:EIP** `0180:003801D6 mov [ebx],es`，該指令仍待受限規格。這些原版啟動／滑鼠查詢收據並非正常玩家畫面或玩法同狀態對拍。

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

## 2026-10-01：ES selector 寫回與下一滑鼠服務停點

**已證實，限固定原版這次啟動指令**：輸入仍為 1.31 ZIP 內 `ORION2.EXE`，SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b`，ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，版控 `startup_probe_131.py --es-store` 由同次啟動序列重生 **CS:EIP** `0180:003801D6 → 003801D9` 的 `66 8C 03`／`mov [ebx],es`。當次 DS=ES=`0188h`、EBX=`003EB9A0h`、EFLAGS=`0046h`；DS:[EBX] 的兩位元組前後都是 `88 01`。私有 `es-store-before.bin`／`es-store-after.bin` SHA-256 均為 `cc808bee2be109604fc5c47d2ad89282d6ced79ee8fd598a5f7ada73ddab4a81`，`es-store-logcpu.txt` 為 `7a0e2d9e565f3c677bbee16d20ea6bb5d84f43edb7b1fe3d3a49c21537e66340`，`es-store-registers.json` 為 `b122d06aa4017c04594f1489b70b07c5c664e65c91d3b22ee4aa045014a4b07f`。原版當次來源等於目的，**不能**由這份前後相同快照推出不同值的寫入結果。

隔離 dosgolem [規格 228](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/228-cpu386-mov-es-to-ds-ebx-word.md) 依 DRAFT→READY→實作→CONFORMED，僅接無段覆寫 `66 8C 03`，不同值與非零 DS base 由 Intel 指令契約及合成測試驗證。固定原檔從 LE entry 在合成 PSP／環境下自然抵達第 5838 步、**dosgolem 重定位 LE 線性位址** `0x15C1D6`，DS:[EBX] 前後皆 `0188h`，單步至 `0x15C1D9`。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過；私有 `workplace/full-test-229.txt` SHA-256 `78c15a85c6a0902893dfde545d0db82ea8ebc4dfe275812eefed49cee5d1e0d1`。有界原檔探針現於第 6007 步、同一 **重定位 LE 線性位址** `0x15C31B` 的 `INT 33h`、AX=`21h` 失敗即關閉；私有 `workplace/moo2-probe-228-after.txt` SHA-256 `dd4ec7d5ac8b04bfc8bb2c8acd977b342f2d6f10654dc6657d379d702f065da7`。此服務的原版返回未知；下一輪先取同次輸入、返回與 caller 消費端，不猜補。兩側 PSP／環境與暫存器狀態尚未全等，沒有 dosgolem 正常玩家畫面或與 remake 玩法同狀態的收據。

## 2026-10-01：滑鼠軟體重設與單點快照勘誤

**已證實，限固定 1.31 原檔和同次 DOSBox-X 輔助執行**：輸入 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。版控 `startup_probe_131.py --mouse-function-21` 在 **DOSBox-X CS:EIP** `0180:0038031B` 取 `INT 33h` 前態 `AX=21h、BX=CX=DX=0`；經 512 指令的中斷內部有界 LOG，再於 **同一 CS:EIP 位址空間** `0180:0038031D` 取返回連續 LOG，得 `AX=FFFFh、BX=3、CX=DX=0、EFLAGS=0006h`。caller 在 `0180:003801B8..003801C3` 將四個回傳值寫入 record；後續讀取 EAX 亦為 `FFFFh`。私有 `mouse-function-21-call-logcpu.txt` SHA-256 `2acd8b3e8bb5cdbdaf4ddf82ffe31cd38d307e5bf27b80dcbed05115b1aae4ba`；`mouse-function-21-return-logcpu.txt` SHA-256 `587eb82e218255df69c9f09bc49ac9b61058db51a9fe496f6a44888426d5be12`；`mouse-function-21-registers.json` SHA-256 `67f63ba0d3301e080848c8e8e808ec2a0b5eb1868655dd8b8fb1f15f12f75797`。原版私有 bytes、終端與完整內部 LOG 未公開散布。

**勘誤**：首輪候選 `EV` 返回快照顯示 `AX=3、BX=0、CX=320、DX=100`，與後續連續 LOG 及 caller 寫入矛盾；該候選撤回，混入原因未知。第二輪 `LOG 20` 從中斷入口只記到 DOS extender 內部，未涵蓋返回；第三輪從入口記 512 指令並跳到返回斷點後，忘了刪斷點，後續 LOG 只有一行。修正斷點清理後以同一 Docker 映像／命令乾淨重跑，取得上述完整收據。舊私有輸出保留以追溯錯誤形成原因，不拿它當 `AX=21h` 回傳證據。

[DOSBox-X 滑鼠原始碼](https://dosbox-x.com/doxygen/html/mouse_8cpp_source.html) 將 `AX=21h` 列為軟體重設，回 AX=`FFFFh`、BX=按鍵數，重設按鍵與座標等內部狀態。隔離 dosgolem [規格 229](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/229-moo2-protected-mouse-software-reset.md) 依 DRAFT→READY→實作→CONFORMED，只在 MOO2 專屬設定支援此呼叫；受控按鍵歸零、座標回本設定的 `320,100` 是**平台契約近似**，不是原版後續查詢位置的實測。固定原檔由 LE entry 自然抵達第 6007 步、**dosgolem 重定位 LE 線性位址** `0x15C31B`，服務返回與原檔 record 消費端通過；合成測試另驗非零按鍵／座標後的受控重設。含真檔的 `go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-230.txt` SHA-256 `ff619b50ad8cf1c146c56b3bdc675c7da01d81f5d55cc1d1cdd6c12f8f63a47e`。有界診斷下一停點是第 6126 步、同一 **重定位 LE 線性位址** `0x15C31B` 的 `INT 33h/AX=1Ah`；私有 `workplace/moo2-probe-229-after.txt` SHA-256 `25ef4143e67248625582e718a6f03e5d677b0233cd68bacd3ecf019b6f270643`。沒有正常玩家畫面、完整資料消費或與 remake 玩法同狀態收據。

## 2026-10-01：零敏感度設定與下一視訊服務停點

**已證實，限固定 1.31 原檔的同次 DOSBox-X 輔助執行**：輸入 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。版控 `startup_probe_131.py --mouse-function-1a` 從前兩次已核對的滑鼠呼叫延續，於 **DOSBox-X CS:EIP** `0180:0038031B` 進入 `INT 33h/AX=1Ah`，BX=CX=DX=`0`、EFLAGS=`0012h`；返回 **同一位址空間** `0180:0038031D` 時，AX／BX／CX／DX 與旗標不變。caller 在 `0180:003801B8..003801C3` 把 `1Ah、0、0、0` 寫入 record。第一次完整擷取恰在外層 170 秒逾時後完成輸出，改用 240 秒上限、相同映像和命令乾淨重跑，正常退出且輸出雜湊相同。私有 `mouse-function-1a-registers.json` SHA-256 `5d877c4ccaaeaf841c28ef7f66ead671e3a81fe1d9691d091b5d7a62ea6723e0`、`mouse-function-1a-call-logcpu.txt` SHA-256 `9405aafaff6e2bdb1ac39c626f618d9b9906eb2293cd8549e63602fba0fe1e62`、`mouse-function-1a-return-logcpu.txt` SHA-256 `02127b37ba6ce8e1347a92bdb2bd789a86a84fff6a1281b35e588a9c49453f82`。

[DOSBox-X 滑鼠原始碼](https://dosbox-x.com/doxygen/html/mouse_8cpp_source.html) 明示 `AX=1Ah` 儲存水平／垂直敏感度與倍速值；兩軸皆非零才改動移動係數。隔離 dosgolem [規格 230](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/230-moo2-protected-mouse-zero-sensitivity.md) 依 DRAFT→READY→實作→CONFORMED，只允許固定原版已觀測的三個零值，並保存 raw 設定；非零值需要未建模的移動效果，仍失敗即關閉。初始 `50,50,50` 來自 DOSBox-X 初始化原始碼，屬**平台契約近似**，固定原版沒有直接讀回。原檔從 LE entry 在合成 PSP／環境下自行重生第 6126 步、**dosgolem 重定位 LE 線性位址** `0x15C31B` 的服務返回與 record 寫入；合成測試驗證零輸入、位置／按鍵不變及非零拒絕。`DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-231.txt` SHA-256 `a983b838b0b84c14b442a82b127d937f4101c88c5fe6dfc0ce6db6742f536e6a`。有界診斷現於第 6256 步、**dosgolem 重定位 LE 線性位址** `0x15C2B2` 的 `CD 10` 停下；私有 `workplace/moo2-probe-230-after.txt` SHA-256 `aab5a75b6ec1b381e7a7008968ae53844b0e7ce1deed221ef8d8b6872ca7073e`。該次原版 `INT 10h` 的功能號／返回未知，沒有正常玩家畫面或玩法同狀態對拍。

## 2026-10-01：模式 03h 呼叫與 DOS 退出邊界

**已證實，限固定 1.31 原檔的同次 DOSBox-X 輔助執行**：輸入 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。版控 `startup_probe_131.py --video-mode-03` 在 **DOSBox-X CS:EIP** `0180:003802B2` 擷取 `INT 10h` 前態 EAX=`00000003`、EFLAGS=`0216h`，同次於 `0180:003802B4` 返回，EAX／EBX／ECX／EDX、DS／ES／SS 與 EFLAGS 不變；caller 在 `0180:003801B8..003801C3` 寫入返回暫存器記錄。私有 `video-mode-03-registers.json` SHA-256 `8b7a2695732e8768acae0df0cb0dd25327f69b338a13d39265408cd751cb8cda`、`video-mode-03-call-logcpu.txt` SHA-256 `adff52520b95af86bee1bbb85a2db09d5ee43f9d59b20c16c6fe936b4ea11181`、`video-mode-03-return-logcpu.txt` SHA-256 `d94c025a57c2bb806cebdf3bca8f86e56cb449eae9c7c224922109e5ac7703d2`。這不證明整段與 dosgolem 同狀態。平台模式語意依 [DOSBox-X `INT10_Handler` 原始碼](https://github.com/joncampbell123/dosbox-x/blob/master/src/ints/int10.cpp)；原版只證實此一呼叫與返回，未驗證顯示記憶體或文字畫面。

隔離 dosgolem [規格 231](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/231-moo2-protected-video-mode-03.md) 經 DRAFT→READY→實作→CONFORMED；只在 MOO2 啟動設定接受精確 `EAX=3`，記錄模式 03h，不以該欄位假裝已呈現畫面。固定原檔整合測試自行抵達第 6256 步、**dosgolem 重定位 LE 線性位址** `0x15C2B2`，單步至 `0x15C2B4`，暫存器、段、旗標不變；全套含原檔 Go 測試通過，私有 `workplace/full-test-231.txt` SHA-256 `66f262c63281b32a8d619f190b0d710956ffecdf158624cd0603f796b57921ee`。

**勘誤與現行停止線**：初次診斷未檢查服務的 `Exited`，於第 10173 步 `INT 21h/AH=4Ch` 之後仍執行到第 10175 步、**dosgolem 重定位 LE 線性位址** `0x1101E4` 的 `66 2E 8E 1D`，誤看成新 CPU 缺口。該舊私有探針 `workplace/moo2-probe-231-after.txt` SHA-256 `53d72ba413dd262c5efc4dd5c2679369f35459634563bd685db2babd62ef3ac0` 保留供追溯，不當作待補功能。探針修正後從固定原檔重跑，於第 10173 步正確停在 DOS 結束服務，代碼 `1`，主控台 `Unable to open mox.set\r\n`；私有 `workplace/moo2-probe-231-exit.txt` SHA-256 `8ee984ad0f300255f41f4259729102a07eb193461cc00b307d35b0a2e2f694d6`。這是合成輸入缺檔路徑，不是正常玩家畫面。下一步可使用已另行擷取的固定空 `MOX.SET` 輔助收據作明示輸入，再由 dosgolem 自行重生成功分支；不能把 DOSBox-X 畫面代替正式收據。

## 2026-10-01：空設定檔比較與正版資料目錄

**已證實，限固定原版與明示零位元組設定檔**：1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；空 `MOX.SET` SHA-256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`，修改時間固定 1996-01-01 00:00:00 UTC。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`；版控探針 `--empty-mox-cmp` 在 **DOSBox-X CS:EIP** `0180:002340CF → 002340D8` 同次 `LOG 2` 觀測 `66 81 3D ... 82 00`，DS:`003AFCBE` 的 word 值為 `0000h`，立即數 `0082h`，原版進入 EFLAGS=`0202h`、離開算術旗標 CF=1、ZF=0、SF=1、OF=0、AF=1、PF=1。LOG 誤把指令印成 `dword`；前綴、9-byte 指令長度及 [Intel 第 2A 卷 `CMP` 契約](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2a-manual.pdf) 一致支持 `r/m16,imm16`。私有 `empty-mox-cmp-logcpu.txt` SHA-256 `60a5d3e3e3ff7e138dd762698e93280d4f6af81f4acf764f573adb6d2daef1c6`；`empty-mox-cmp-registers.json` SHA-256 `c6d90860f417e3d8d6a1a095df9aff3fe44abcb856b3c63a1d9d50c6e1146134`。該記憶體欄位玩法語意**未知**。

隔離 dosgolem [規格 232](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/232-cpu386-cmp-rm16-imm16.md) 經 DRAFT→READY→實作→CONFORMED。明示空檔、合成 PSP／環境在第 6504 步的 **重定位 LE 線性位址** `0x100CF` 讀線性 `0x191CBE`=`0000h`，單步至 `0x100D8`、EFLAGS=`0297h`；六個算術旗標與原版同次 LOG 一致，其他暫存器及來源未寫回。合成測試另覆蓋非零來源、相等、段基址、截短及越界。含固定原檔的 `go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-232.txt` SHA-256 `deb2e6e33a467ce73eb775189728366c4f0a11a26a03706adcd95e63c77b81db`；空檔診斷私有 `workplace/moo2-probe-232-after.txt` SHA-256 `af4c6ac4d965091ced033267f7b6a7112bfa2c6b9f19caf8fe80db462fae75c4`，第 15040 步仍以代碼 1、同一句 `Unable to open mox.set` 結束。故空檔只是一個受控 CPU 樣本，不是正常遊玩設定。

**正版資料輸入，尚無原版同次完整路徑驗收**：本機 `original_game/Master_of_Orion_II_-_Battle_at_Antares_1996.zip` SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 內含 `mastori2/MOX.SET`，大小 553、SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`；官方 `moo2_patch1.31/MOO2-1.31.en.zip` SHA-256 `908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5` 提供固定 EXE。先以真實 `MOX.SET` 單檔診斷，來源 word 為 `0082h`，20 萬步上限仍有新位址進展，不能標成卡死。改為 100 萬步有界診斷並只在容器暫存區抽取 ZIP 根層 417 個原版檔案；首次提取把 `MT32/SB16/SC55` 三個子目錄視為平面檔而觸發腳本 assertion，修正過濾條件後以同一工具鏈乾淨重跑。完整資料的 dosgolem 診斷在第 288215 步、**重定位 LE 線性位址** `0x151A21` 的 bytes `66 F7 05 52 1C 1C 00 F0 FF` 停於未支援的 `F7` word 形狀；私有 `workplace/moo2-probe-232-full-game.txt` SHA-256 `c032e24493aef6d936a810d10dfe109b9884de5a78cb5527fdebcaeec63cfb9e`。這是工具前進而非原版同狀態玩家收據，下一指令須另取原版呼叫前後證據；本機原版素材不加入 Git 或公開封包。

## 2026-10-01：完整資料的 TEST 指令與記憶體退出

**已證實，限隔離 dosgolem 的合成 PSP／環境**：上述固定 1.31 EXE 與正版壓縮檔根層 417 檔不變。dosgolem 原第 288215 步在**重定位 LE 線性位址** `0x151A21` 遇到 `66 F7 05 52 1C 1C 00 F0 FF`；[Intel IA-32 第 2B 卷 `TEST` 條目](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf)支持 16 位元記憶體來源與立即數形狀，來源的遊戲語意仍**未知**。dosgolem 規格 233 依 CPU 契約補此形狀；合成測試涵蓋 DS／SS 選段、非零／零結果、不寫回、截短與越界，全套含原檔測試通過。原資料探針自行越過舊停點，於第 295276 步以 DOS `AH=4Ch` 代碼 1 結束，主控台顯示 `Insufficient Memory!`、要求 8192 bytes、DOS space remaining 0 bytes。私有 `workplace/moo2-probe-233-full-game.txt` SHA-256 `c75e200fda4251e97752600df06c84c14d2fbc27eb232afcd50f60936d79e744`。記憶體不足只能歸於目前合成環境，尚不可推論原版正常啟動也不足。

**原版輔助基準未閉合**：DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 `fd2-dosbox-x:debug-0d7b272b` ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，從同一 EXE 與完整正版資料跑至前述零段基底分支，設候選 **DOSBox-X CS:EIP** `0180:00375A21` 及退出服務斷點後有界續跑。候選未命中，`EV` 無可解析快照；Xvfb 擷取當時仍是黑色 DOSBox-X 圖形視窗，沒有主選單。私有 `full-data-test-word-registers.json` SHA-256 `a36a0848eac58ee5647e1903e74c6ccab328879d44420e6e4f0292604b5be6cf`、`full-data-test-word-screen.png` SHA-256 `2b96a6124ce858301efab47c54aee2537ccebca77501e638d39f38d4b2836849`、終端 SHA-256 `ab883c767a181ca62bbfdc6af864ab400f331a19120869bf68a09acd19a26dac`。`+0x224000` 位址對應仍只是假說，這次**沒有**證實候選指令的原版來源／旗標／consumer；黑畫面也不能判定遊戲已停止。沒有 dosgolem 自生的正常玩家畫面或與 remake 玩法同狀態驗證。

**後續記憶體稽核，已證實執行器事實**：相同 EXE、`MOX.SET` 與正版 ZIP 根層 417 檔案，在 dosgolem **重定位 LE 線性位址** `0x1101E1` 的退出點記錄：初始映像長 1,891,536 bytes；DPMI `0100h` 配置 DOS 記憶體呼叫 2 次、`0501h` 配置線性記憶體 126 次、`0502h` 釋放 122 次，未實作 DPMI 功能計數為空。退出時機器記憶體長 66,285,568 bytes，僅兩個線性區塊仍在帳本，沒有 DOS 區塊。私有 `workplace/moo2-probe-memory-audit.txt` SHA-256 `05a3295912c263601e75dcf1933fd6ad14022edebd9e3a072e4d6f47be165d`；使用隔離 dosgolem 提交 `774471211bd818e9325bf0653aa3f3de66b5c86a` 加只讀退出診斷重生。

**已證實的執行器限制、原版因果仍未知**：`internal/machine/dpmi.go` 的 `setLimits` 把 DOS 低位配置游標放在 LE 映像尾端，起點已大於 `dosMemTop=0xA0000`，所以這份合成環境的 DOS 可用段數為零；`0502h` 只刪除帳本，`0501h` 的 `brk` 單調增加且限制在 64 MiB。程式印出的兩種剩餘空間與這些限制一致，但沒有同次原版配置呼叫收據，故「原版如何配置」及兩種限制各自對退出的影響仍**未知**。平台修正應先對照 [DPMI 1.0 規格](https://docs.pcjs.org/specs/dpmi/1991_03_12-DPMI_Spec_v10.pdf)，分開處理 640 KiB 以下 DOS 區塊與線性區塊回收，不能只調高記憶體上限讓探針通過。

## 2026-10-01 勘誤：線性回收後的獨立 DOS 低位配置阻塞

**已證實，限 dosgolem 合成 PSP／環境**：輸入仍為官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、原版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 的根層 417 檔案、`MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。隔離 dosgolem 規格 234 依 [DPMI 1.0 `0501h／0502h` 契約](https://docs.pcjs.org/specs/dpmi/1991_03_12-DPMI_Spec_v10.pdf)補線性區塊回收；一次性 `golang:1.24-bookworm` 容器以 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE GOCACHE=/tmp/go-cache go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-234.txt` SHA-256 `0c4aa5818a10ca07c604b44f0daf3bfee775b07a02d0718232aeec1666a1e0e7`。完整資料探針由先前第 295276 步前進，自然於第 3,947,961 步退出，仍顯示 `Insufficient Memory! Attempted to allocate 8192 bytes`。退出時線性可用 39,150,168 bytes、DOS 可用 0 bytes；DPMI `0501h` 4,779 次、`0502h` 4,775 次、`0100h` 2 次，未實作功能計數為空。私有 `workplace/moo2-probe-234-8m.txt` SHA-256 `843006eb6fc957cc8d8326fd9c6a1c825a534220ae9e4da663adc480f109611b`。首筆 `0100h` 在 **dosgolem 重定位 LE 線性位址** `0x15C317` 請求 513 段落（8,208 bytes），返回 AX=`8013h`、CF=1、BX=0；其後的 1 段落備援請求亦失敗。私有 `workplace/moo2-probe-234-dos-alloc.txt` SHA-256 `efe498bae3d15aab2b87bfcd1b8a5982341e3c94b56e142f8632b0d5c26a6942`。**訂正前段未知項**：線性釋放不重用確實造成原來的提前上限，但單修之後仍有低位 DOS 配置阻塞；不能把原退出完全歸因於線性空間。

**已證實，限 DOSBox-X 輔助原版執行**：版控 `apps/moo2/tools/startup_probe_131.py --dos-memory-0100`，以同一正版資料、1.31 EXE、DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582` 擷取首次兩筆 `INT 31h/AX=0100h`。第一筆在 **DOSBox-X CS:EIP** `0180:00380315 → 0180:00380317` 請求 BX=`0201h`（513 段落），返回 AX=`0FE3h`（實模式段，物理起點 `0xFE30`）、DX 低 word=`01D8h`（selector）、CF=0。第二筆同一呼叫位址請求 BX=`00B0h`（176 段落），返回 AX=`11E5h`（起點 `0x11E50`）、DX 低 word=`01E0h`、CF=0。私有 `workplace/dosbox-moo2/dos-memory-0100-registers.json` SHA-256 `1d96749e2d0811ffc506f228c553f8cd0976e24c4a0e8341b6bf2525b1b30600`，終端 SHA-256 `af529718a5c001d04058c0b5b2ece242498bb3a931ebea0f8fe857048694e909`。首次解析曾把較早的 `EV` 快照誤選為第二筆，舊 `dos-memory-0100-stale-registers.json` 保留在私有工作區；改取最後符合呼叫／返回位址的快照，並以相同容器與原始輸入重跑，才取得上述收據。這是 DOSBox-X **輔助基準**，不能代替 dosgolem 自行重生的正常玩家路徑。第一筆請求大小可局部比較；dosgolem 首筆失敗後第二筆改為 1 段落，兩側後續狀態已分歧。

**已證實的工程限制與下一閘門**：現行 `LEMachine.Mem` 在 LE 重定位位址 `0x10000` 起直接存放原版程式，映像長 1,891,536 bytes；`setLimits` 將 DOS 配置游標放在映像尾端，超過 `0xA0000`。若只將游標調回 `0xFE30` 附近，`allocDOS` 的資料寫入會覆蓋 LE 影像；此修法不能採用。依 [DPMI 1.0 `0100h` 契約](https://docs.pcjs.org/specs/dpmi/1991_03_12-DPMI_Spec_v10.pdf)，需先建立低位 DOS 記憶體與 LE 載入影像可並存、實模式段與保護模式 selector 可相互存取的受限記憶體映射規格，測試防覆蓋與資料往返，再由 dosgolem 重跑。原版 DOSBox-X 的段號與 selector 不要求合成執行器數值逐一相同；玩家畫面與玩法同狀態驗證仍**未知**。

## 2026-10-01：高位 LE 載入後的首筆 DOS 配置與實模式視訊停點

**已證實，限隔離 dosgolem 的高位映射模式**：輸入仍為前節固定 1.31 EXE、正版 ZIP 根層 417 檔與 `MOX.SET`；規格 235 建立 DRAFT→READY，明示 `LoadLEInMZWithDOSArena` 將 LE 物件及已支援 fixup 整體平移 `0xF0000`，使原物件起點 `0x10000` 對應至 **dosgolem 高位重定位 LE 線性位址** `0x100000`，入口 `0x1FFF18`；DOS arena 從 `0x10000` 起。原 EXE bytes 與一般 loader 不變。合成測試核對先後配置 513／176 段落、selector 與實模式 bus 雙向讀寫、防 LE 覆蓋、溢位拒絕；固定原檔測試遍歷 LE fixup 並核對平移。一次性 `golang:1.24-bookworm` 容器執行 `DOSGOLEM_MOO2_EXE=/tmp/ORION2.EXE go test -buildvcs=false ./... -count=1` 全通過，私有 `workplace/full-test-235.txt` SHA-256 `6ed111b7c3a51af6330c947b8a447ce2a065fcc5f7c91da5137040c8f2582de1`。這驗證執行器記憶體映射，不證明原版絕對載入地址相同。

完整資料在明示 `DOSGOLEM_MOO2_SEPARATE_DOS=1` 下，自 LE entry 自行執行至第 189,580 步；於 **dosgolem 高位重定位 LE 線性位址** `0x24C317` 首筆 `INT 31h/AX=0100h` 請求 513 段落成功，合成返回 AX=`1000h`、DX=`0100h`、CF=0。第 189,582 步的 `INT 31h/AX=0300h` 未處理，底層 `RealModeLast` 明示 `Interrupt:16`、`Entry:0000:0000`、`Error:DPMI0300中斷10未註冊`；輸入封包的實模式 AX 低 word 為 `004Fh`，其他功能語意與原版返回仍**未知**。私有 `workplace/moo2-probe-235-real-mode.txt` SHA-256 `5dbf8b22cda56a2f77fd0ff498daa3dd5d1a14f7b194f074fe748013708023e8`。dosgolem `Unimplemented` 只記 `0300h`。首次 DOS 配置失敗已消除；第二筆原檔配置尚未自然抵達，因此規格 235 維持 READY，不能宣稱其整段驗收或玩家路徑完成。下一步只追這個 `0300h/INT 10h` 呼叫的原版輸入、返回及可見 consumer，再決定受限視訊服務；不得填入猜測成功值。

## 2026-10-01：原版 VBE 4F01h 模式 0101h 與 dosgolem 後續停點

**已證實，限 DOSBox-X 輔助原版執行**：輸入仍為官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 的根層 417 檔、`MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`；版控探針 `apps/moo2/tools/startup_probe_131.py --real-video-4f01`。首次探針預設第二筆即 `4F01h`，實際發現另一筆 `4F00h`，因此失敗的原始終端保留於私有工作區；修正為逐筆記錄並重跑。正式輔助收據 `real-video-4f01-registers.json` SHA-256 `baafd3f5178954e1d5d24224bf7e98233f62b701cda0088c2554c40f018a1454`、終端 SHA-256 `776f36857384c197a82d2153241be09284071853f6ca9b745f90377b1d517c0c`。

**已證實，原版呼叫與寫入**：在 **DOSBox-X CS:EIP** `0180:00380315 → 0180:00380317`，前兩筆 DPMI `0300h/INT 10h` 封包 AX=`4F00h`，第三筆 AX=`4F01h`、CX=`0101h`、ES=`0FE3h`、DI=`0`；返回 AX=`004Fh`、flags=`0002h`。對應 **DOSBox-X 實模式物理線性** `0xFE30` 的 256 位元組緩衝，原先全零，返回後前 44 位元組部分欄位改變、後 212 位元組保持。輸入／輸出私有 SHA-256 分別為 `5341e6b2646979a70e57653007a1f310169421ec9bdd9f1a5648f75ade005af1`、`d0eafd13c290b998a4d0d3be332d9be06eabde2a19764f88ab68d483e64e35c3`；caller LOG SHA-256 `9a541abd20036f3f77f1b17e92d6c5e5290e0b793b9f7ca2ddf788278a85b158`。依 [VESA VBE 2.0 Function 01h](https://www.phatcode.net/res/221/files/vbe20.pdf)及原版回傳，dosgolem 規格 238 只重現固定模式資訊；虛擬硬體欄位屬 **hardware-spec approximation**，不代表真機或原版素材的通用定值。

**已證實，限 dosgolem 自生平台路徑**：規格 236–238 受限視訊服務通過含原檔全套 Go 測試，規格 238 私有測試輸出 SHA-256 `84406686fb97b0c27f8aa42fe41190c08bc256c6f375a45c7b71b5ec94e4088d`。明示高位 LE 載入、正版資料與相同 EXE 自 LE entry 自行執行，於第 190,517 步、**dosgolem 高位重定位 LE 線性位址** `0x24C2B2` 停在直接 `INT 10h/AX=4F02h`、EBX=`0101h`；私有 `workplace/moo2-probe-238-full-game.txt` SHA-256 `6f5df2e26b65c5935db63cff4751b103182ff75bab09c18314d8a42ed2c70284`。此原版模式設定的返回與 consumer 尚未擷取；規格 235 的原檔第二筆 DOS 配置仍未自然抵達。**未知／不阻塞規格 238**：兩側 PSP／環境、堆疊、VRAM 和正常玩家畫面仍不同或未觀測，不能聲稱玩法同狀態對拍。

## 2026-10-01：模式 0101h 設定返回與 CPU 下一停點

**已證實，限 DOSBox-X 輔助原版執行**：同一 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 根層 417 檔、`MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`，DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。版控 `startup_probe_131.py --video-mode-4f02` 於 **DOSBox-X CS:EIP** `0180:003802B2 → 0180:003802B4` 捕得直接 `CD 10`：輸入 EAX=`4F02h`、EBX=`0101h`、ECX=EDX=ESI=EDI=`0`、EFLAGS=`0216h`，返回 EAX=`004Fh`、其餘擷取欄位及旗標不變；caller 於 `0180:003801B8` 保存 EAX，`0180:00368A5C` 讀取。私有 `video-mode-4f02-registers.json` SHA-256 `5b6d94309ba8bcc5e9de221ed6b8545c7a5c8f1cb0b86ceae7dac2b9de762dea`、caller LOG SHA-256 `f86782937cadc66d2988dbbfdeed7fdcb90cc872eb6f42f178a8e9f289259be8`、終端 SHA-256 `74496864a917395ed27ae73de730e840d076a809376dbefc36171b608ebf26b3`。依 [VESA VBE 2.0 Function 02h](https://www.phatcode.net/res/221/files/vbe20.pdf)與此返回，隔離 dosgolem 規格 239 只接受模式 `0101h`；未擷取 VRAM 及玩家畫面，合成模式狀態屬平台近似。

**已證實，限 dosgolem 自生執行**：規格 239 全套含原檔 Go 測試通過，私有 `workplace/full-test-239.txt` SHA-256 `e78f1025caaaacd4ec286585cfc18a0e92afe07dad03de08e18aad186322cb5f`。固定正版資料及明示高位 LE 載入自然越過第 190,517 步，於第 1,151,730 步、**dosgolem 高位重定位 LE 線性位址** `0x25025F` 停在未支援 opcode `04h`；私有 `workplace/moo2-probe-239-full-game.txt` SHA-256 `96af49b38b75df1cfd62c08ea9d30ba2f57624841bd9db6851e0b78f13507da6`。該 opcode 的原始 bytes、立即數與對應 DOSBox-X 位址尚未核對，不先把它推定成玩法語意。第二筆 DOS 配置及正常畫面仍未到達。

## 2026-10-01：`04 ib` 原版樣本、第二筆 DOS 配置與音效初始化平台停點

**已證實，DOSBox-X 輔助原版樣本**：官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f`、`MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，版控 `startup_probe_131.py --add-al-imm8` 於 **DOSBox-X CS:EIP** `0180:0038425F → 0180:00384261` 取得 `04 20`：EAX=`003E4150h → 003E4170h`，EFLAGS=`0297h → 0202h`。私有暫存器、LOG 與終端 SHA-256 分別 `e116d1761ae34b83b18f970059290fa06d64264de6c65bd4f027e9e9cb75a557`、`1a13f3c836b94373807964886df093d2034e1d9af8f330850d0a9f4bec71b652`、`833baa62abfd4be1c8a8d29a68fb2bdc75fc1eac5cf37dd0885e280249a66f19`。dosgolem 停點的 EAX 不同，不能稱同狀態；通用 CPU 語意另依 [Intel IA-32 指令手冊](https://cdrdv2-public.intel.com/843829/325383-sdm-vol-2abcd-dec-24.pdf)。

**已證實，dosgolem 自生結果**：隔離 dosgolem 規格 240 將無前綴 `04 ib` 接入既有 `add8`，固定 EXE 全套測試通過，私有 `workplace/full-test-240.txt` SHA-256 `a9ceacbb357b48b98070cb4548aa2dde31c17d68a44d87af3b7960b963b49b86`。相同正版資料從明示高位 LE 入口自然越過第 1,151,730 步；第二筆 `INT 31h/AX=0100h` 請求 176 段落，合成回 AX=`1201h`、DX=`0108h`、CF=0，第一筆 513 段落回 AX=`1000h`、DX=`0100h`。因此規格 235 的兩筆配置驗收已閉合，原來的 READY／「第二筆尚未到達」只是舊 checkpoint。第 1,166,995 步在 **dosgolem 高位 LE 線性位址** `0x2454AE` 執行 `DPMI 0300h → INT 66h`，**dosgolem 實模式** `1201:0611` 寫 DSP `OUT 0226h,01h` 因未接平台埠停下。私有 `workplace/moo2-probe-240-full-game.txt` SHA-256 `0f00b6396e07eece1707f25932eb5450c46b385add27c4a8d5637171ab88d064`。這不表示原版正常啟動在同一時點選用相同音效設定。

**已證實，受限平台接線與遮罩前進**：隔離 dosgolem 規格 241 只讓 MOO2 保護／實模式共用既有 `LEOPLPorts`，正式原檔自生越過 `0226h`，實模式第 299 步在 `1201:0220` 寫 `OUT 00D4h,05h` 時拒絕；私有 `workplace/full-test-241.txt` SHA-256 `06c75ae6fa9adec96d5b6153f731499bea5cdc3b9fb46abe6ae3153fb0a6894a`，`workplace/moo2-probe-241-full-game.txt` SHA-256 `7338276b9565381915d397c6bdebdfc3d26aafa11e56b22e4396a63744f0faf9`。依 [IBM PC AT 技術參考手冊](https://www.minuszerodegrees.net/manuals/IBM_5170_Technical_Reference_1502243_MAR84.pdf)的第二 DMA 控制器 `C0h..DFh`、[Intel 8237A 資料表](https://www.pcjs.org/documents/datasheets/intel/INTEL_8237A_DMA.pdf)的單通道遮罩位元，以及 [DOSBox-X `dma.cpp`](https://github.com/joncampbell123/dosbox-x/blob/master/src/hardware/dma.cpp)的偶數埠映射，`D4h/05h` 表示設置第二控制器相對通道 1、全域通道 5 的遮罩。此語意是**平台規格近似**，不是 MOO2 逐週期音效實測。規格 242 只保存該遮罩；固定 EXE 全套測試私有 SHA-256 `63ce81a460661c68c24979ee50e1e3d3720c8c7f9fd309c627f0fd1558e3daf7`。同一正版資料自生越過 `D4h`，實模式第 305 步在 **dosgolem segment:offset** `1201:0244` 的 `OUT 00D8h,00h` 拒絕；私有 `workplace/moo2-probe-242-full-game.txt` SHA-256 `142f0d05f0feea086353bd9cfbd5679f0dda095b9644da82cd8007376c187c06`。`D8h` 尚未實作；僅追公開平台契約，不挖 DAC／PIT／DMA wall-clock。

**未知且阻塞正式對拍**：兩側 PSP／環境、堆疊與完整狀態未對齊；沒有 dosgolem 正常玩家畫面、音訊輸出、受控亂數或與 remake 同狀態的玩法收據。上述原版執行與平台模型只支持其各自窄範圍結論。

## 2026-10-01：第二 DMA、SB16 `B0h` 與 BIOS 資料區

**已證實，限 dosgolem 固定原檔的合成執行**：輸入仍為官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 的根層 417 檔及 `MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。工具為隔離 dosgolem 分支的 `golang:1.24-bookworm` 一次性容器；實際版控版本與本輪差異見規格 243–246 及 Git 提交。`DPMI 0300h → INT 66h` 的 **dosgolem 實模式** `1201:0244` 起先寫第二 DMA `D8h/00h`，後於 `1201:028F` 寫頁埠 `8Bh/00h`，再於 `1201:05D9` 向 DSP `022Ch` 發送 `B0h`。受控可丟棄探針記得完整序列 `B0 30 00 00`、第二 DMA 通道 5 的 mode=`48h`、mask=`0Dh`、base=`92AFh`、count=`0001h`、page=`00h`；私有 `workplace/moo2-probe-245-b0-state.txt.gz` SHA-256 `151a0edf2db27e3f7df8d5fc386e2c2bd7be751258fd7cf14b106960c9800485`。這是原版 bytes 在合成平台的可重播輸入，未核對原版實際聲波。

**平台契約近似**：dosgolem 規格 [243](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/243-secondary-dma-register-programming.md)／[244](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/244-secondary-dma-page-registers.md) 按 [Intel 8237A 資料表](https://www.pcjs.org/documents/datasheets/intel/INTEL_8237A_DMA.pdf)、[IBM PC AT 技術參考手冊](https://www.minuszerodegrees.net/manuals/IBM_5170_Technical_Reference_6280070_SEP85.pdf)與 [DOSBox-X `dma.cpp`](https://github.com/joncampbell123/dosbox-x/blob/master/src/hardware/dma.cpp)實作第二 DMA 獨立暫存器與頁埠；規格 [245](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/245-sb16-b0-single-cycle-probe.md)依 [Creative Sound Blaster 硬體指南](https://www.ardent-tool.com/sound/Sound_Blaster_HW_Programming_Guide_1st.pdf)限定 `B0 30 00 00` 的單次 16 位元 D/A、完成閘門及 16 位元 IRQ。虛擬時鐘以每指令 `1 µs` 推進，sample duration 是 **hardware-spec approximation**，無逐週期、波形、聽感或玩家路徑宣稱。三次自生停點收據 SHA-256 分別為 `8b967af9e5c7b8c994e586b3e0229ca1510f7cab6493196ac3f1389b5d9ceb38`（243）、`337e94af41018d907487430fc80bab5db6a965ce93972048bc1a19152b33e96b`（244）、`ba5aff0b0052eb66e77fb28ca43def2a0f33dd0b7cd8257a95d0a3d0b09a54a4`（245）；相應固定 EXE 全套 Go 測試輸出 SHA-256 分別 `bfc30c39cbdb8ba04f060ac45749ae61bf3806d027c3d9f2125ed25056425f00`、`6b91e417a4d0839fe83500e40834a58672b6367a8d0999e73701008d22662e6b`、`2eca390bbd4ef2b35f6436c49c6293c7b7e624ea622eef2aa03ed8de405d4`。

**勘誤，已證實的位址來源**：B0 後初見 **dosgolem 實模式** `1201:0317` 的 `IN 0006h`，不應被視為 DMA 埠需求。原始 bytes `B8 40 00 8E D8 8B 16 63 00 80 C2 06 EC A8 08` 先從 BIOS 資料區 `0040:0063` 讀基底，合成初態為零才得 `0006h`。可丟棄位元組視窗收據 SHA-256 `6a7f46d0afebd74ab0de6f729425cc7656380bc2ade14c0cadac065ba7057654`。dosgolem 既有規格 186 批次 38 已定義 `InstallDOS4GWBIOSData` 的 `03D4h` 彩色 CRTC 基底與 selector `0040`，MOO2 接線漏呼叫；規格 [246](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/246-moo2-bios-data-attach.md) 重用它並保留衝突拒絕。正式接線後由原版自生越過 `0006h`，於 **dosgolem 實模式** `1201:073F` 停在 DSP `IN 0225h`；私有 `workplace/moo2-probe-246-full-game.txt.gz` SHA-256 `0b3d80dc9aba63f4973089d855384e24080d8e0f529f18f4caca31e0b508667c`。早期可丟棄試跑 `workplace/moo2-probe-246-bda-trial.txt.gz` SHA-256 `ade62e024f0553d23c14d910ea1784651edf0571d07cf98bce75641fbb26b265` 僅供勘誤。正式接線後固定 EXE 全套 `go test -buildvcs=false ./... -count=1` 通過，輸出 SHA-256 `7fa4b1acb412bf35c9110a5e94a7349f0dc40a951a396b7b008708b8b5b46f08`。上述地址均為 dosgolem segment:offset；不可與 DOSBox-X CS:EIP 或 LE 線性位址混用。

**未知與下一步**：`IN 0225h` 的原檔讀取後消費端尚未擷取；須先查公開 Sound Blaster 契約並觀察自然後續，勿把未知硬體埠直接回傳成功。兩側 PSP／環境、堆疊及遊戲完整狀態不一致，沒有 dosgolem 正常玩家畫面、原版音效、受控亂數或與 Go remake 的玩法同狀態對拍。

## 2026-10-01：SB16 數位語音音量混音器與實模式返回

**已證實，限 dosgolem 固定原檔合成初態**：隔離 dosgolem `642abfe` 加可丟棄埠狀態探針，在 **dosgolem 實模式** `1201:073F` 的 `IN 0225h` 前記到 `OUT 0224h,32h`，當時索引 `32h`；私有 `workplace/moo2-probe-247-mixer-diagnostic.txt.gz` SHA-256 `0b8ea234f8b901976818c4acf1836958ca993ee5bb12a3733dbd329e3c1b11b0`。輸入仍為官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 及 `MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`，工具為 `golang:1.24-bookworm` 一次性容器。

**平台規格近似與驗收**：[Creative《Sound Blaster Series Hardware Programming Guide》第 4 章 CT1745 混音器](https://www.ardent-tool.com/sound/Sound_Blaster_HW_Programming_Guide_1st.pdf) 定義 `32h/33h` 為左右數位語音音量，D7–D3 有效，預設 24/31，原始高位 `C0h`。隔離 dosgolem [規格 247](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/247-sb16-voice-volume-mixer.md) 經 DRAFT→READY→受限讀寫→CONFORMED；僅接受這兩個索引並保存有效高位，原版物理卡當次音量及聲波未驗。固定 EXE 全套 `go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-247.txt` SHA-256 `fe95d9119cde6fb624038c04a132ff9df0694f0031b3e5928c3d78c329d23147`。正版資料自 LE entry 越過 `IN 0225h`，`DPMI 0300h → INT 66h` 的實模式呼叫於 135 步後正常返回；第 1,170,054 步自然停在 **dosgolem 高位 LE 線性位址** `0x25221E`，原始 bytes `83 C8 10 83 7B 18 01`，CPU 回報 `83` 的 ModRM `C8h` 尚未支援。私有 `workplace/moo2-probe-247-full-game.txt.gz` SHA-256 `ad3568f83e7a85f70e682484bf392bac3bcd66dce42fd474e1d061b6af3083ca`。

**未知**：`83 C8 10` 的原版同次暫存器／旗標與後續 consumer 尚未核對；兩側 PSP／環境、堆疊、混音器設定及完整資料流不同。這是執行器平台前進，不是正常玩家畫面、音訊聽感或 remake 玩法同狀態對拍。

## 2026-10-01：OR／XOR 原版樣本與 BIOS tick 等待

**已證實，固定原版輔助樣本**：輸入沿用官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 根層 417 檔與 `MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，版控探針 `--or-register-imm8` 在 **DOSBox-X CS:EIP** `0180:0038621E` 直接讀得 `83 C8 10`，同次 LOG 的下一指令 `0180:00386221` 確認 EAX=`0 → 10h`、ZF／PF=`1 → 0`、CF／OF／SF=0、IF=1。私有 JSON SHA-256 `19e379e706dc80f1100c254f68f145b11555da57796525d772b3e8abbfebda36`、LOG SHA-256 `dd2e0788f4e33ccbbf32886ab62faecace8468362056eaccad0f368887ffb047`。下一 CMP 來源值此次為 2，其結構語意未知。隔離 dosgolem [規格 248](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/248-cpu386-or-register-imm8.md) 依 [Intel OR 契約](https://cdrdv2-public.intel.com/868141/253667-089-sdm-vol-2b.pdf)補無前綴暫存器立即數形式；固定 EXE 全套測試 SHA-256 `fa42fdfa526f73f528f57088473c79af840a19e1a5ca245bd850fe11a2da89d2`，自生下一停點為第 1,170,076 步、**dosgolem 高位 LE 線性** `0x250722` 的 `81 F2 00 80 00 00`，私有輸出 SHA-256 `5ca332efc5cc54d6ebf0fb8921ac8bad10cc38c577cf0f3576fdd764084c8731`。

**已證實，XOR 與除錯旗標顯示差異**：同一工具／資料的 `--xor-register-imm32` 於 **DOSBox-X CS:EIP** `0180:00384722` 直接讀得 `81 F2 00 80 00 00`，EDX=`0 → 8000h`。首段 LOG 將 PF 印成 0，重跑相同；但下一指令 `0180:00384728` 的實際 `EV EFLAGS=0206h` 與續段 LOG 皆 PF=1。初次 JSON 另存，SHA-256 `4e49b5b650d4923192389da79fea9a0c51a7a840bf19badb29aa9f491ec82335`；重跑 JSON SHA-256 `368364c1b18d62c58ba6a381858b44fbd69d56edd32cfefb6c881e91dd10a68e`、首段 LOG SHA-256 `bec46426929c6ff31d4ec525563dbf7b972b779fe1af4664a82b3ea949bf575c`、續段 LOG SHA-256 `cc608d6504ee24b1b001a6980c34b5b6e301135c76343f712642b9fbc7fb693e`。**強推論**：[DOSBox-X 公開 flags.cpp](https://raw.githubusercontent.com/joncampbell123/dosbox-x/master/src/cpu/flags.cpp) 的 `get_PF` 取 dword 同位而 `FillFlags` 取低 byte，可解釋顯示差異；尚未逐值核對映像來源，不稱內部已證實。CPU 契約採 [Intel 80386 XOR，第 411 頁](https://read.seas.harvard.edu/~kohler/class/aosref/i386.pdf)及實際 EFLAGS，沒有把錯誤 LOG 旗標寫入預期值。[規格 249](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/249-cpu386-xor-register-imm32.md) 限無前綴暫存器完整立即數，八個目的與截短拒絕等測試通過；固定 EXE 全套測試輸出 SHA-256 `42bba15ca1049b603dc409a74635cd8a94e0287c543b341de0a2329d7ecc2732`。

**已證實，dosgolem 自生等待**：Go 1.24.13／`golang:1.24-bookworm`，隔離 `8f6bf14` 加 248／249 實作。原檔越過 XOR，第三筆 DOS 配置 512 段落自然成功，於探針八百萬步上限停止，未實作 DPMI 清單為空；私有 `workplace/moo2-probe-249-full-game.txt.gz` SHA-256 `a0e0c714e5cf33bcf612becf50e5c7186494b44259ab0c8d789e68f5890a947b`。補尾端探針重跑，**dosgolem 高位 LE 線性位址** `0x222AAD..0x222AB5` 的 `BE 6C 04 00 00 AD 3B C3 74 F6` 重複讀取 BIOS `046Ch`、比較 EBX、相同便回圈；EAX=EBX=0、IF=1。私有 `workplace/moo2-probe-249-loop.txt.gz` SHA-256 `03cf7b8a41f86a6192b91fc4beb6977b81e1a3378b47ef397abb8cfbe53bb2a2`。MOO2 啟動未安裝現有 `InstallLEBIOSClock`，證據與有界接線契約見 [規格 250](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/250-moo2-bios-clock-attach.md)；依公開平台與既有規格 186 批次 95 處理，不挖 PIT／ISR 逐週期內部。

**未知與阻塞邊界**：兩側 PSP／環境、堆疊、音效設定與完整狀態尚不可比；上述 CPU 及平台收據不構成 dosgolem 正常玩家畫面、音效、受控亂數或與 Go remake 同狀態對拍。本輪未更動 Go remake 玩法。

**已證實，BIOS 時鐘接線後自然前進**：規格 250 重用 `InstallLEBIOSClock`，保留既有 hook、IF／PIC 與客製向量拒絕契約；時間仍是每道指令一微秒的硬體規格近似。固定 EXE 的全套 `go test -buildvcs=false ./... -count=1` 通過，私有 `workplace/full-test-250.txt` SHA-256 `d57b7d3b9da005b9ae06105bfa2e3bfcf7b11389c92a6c830c5cda109741b91e`。相同正版資料自 LE entry 離開 BIOS 等待，於第 1,545,396 步、**dosgolem 高位 LE 線性位址** `0x24C31B` 的原始 `CD 33` 拒絕；EAX／EBX／ECX／EDX=`0`，EFLAGS=`0012h`，DPMI 未實作清單為空。私有 `workplace/moo2-probe-250-full-game.txt.gz` SHA-256 `74b73985b73a1f077f2005343fef0168df1d05dbe49d03d8f824369f5105a1b8`。這推翻了把八百萬步上限當 CPU 缺口的解釋；原始等待收據仍保留。下一步擷取 `INT 33h/AX=0000h` 的原版輔助返回與 consumer，再審查受限滑鼠平台契約。

## 2026-10-01：滑鼠重設、敏感度查詢與完整資料序列

**已證實，原版同次輔助服務**：官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 根層 417 檔及 `MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。DOSBox-X 2026.07.02 SDL2 heavy debugger 映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，版控 `--mouse-reset` 在 **DOSBox-X CS:EIP** `0180:0038031B` 直接取得 `CD 33 C3`，功能 `0000h` 由零暫存器輸入返回 AX=`FFFFh`、BX=`3`，其餘擷取欄位與 EFLAGS=`0016h` 不變；下一指令 `0180:0038031D` 與連續 LOG 確認返回寫入 record `003D18E0h +0/+4/+8/+0Ch`，`0180:00347527` 比較驅動是否存在。私有 JSON SHA-256 `2be392ca0015dbf77cf73de2fb3881daee97d0c71428deb5d0b519acab859b18`、caller LOG SHA-256 `d6030bb492e573e64ec239faeb92dff7701b273e19dba12b5fc46c9cb8116327`、終端 SHA-256 `8cea3130bbf2dfdd8e7bf6a271c91411e5c3f65ffbaa46fa946f29a5a2a00085`。[規格 251](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/251-moo2-protected-mouse-driver-reset.md) 依 [DOSBox-X 平台來源](https://dosbox-x.com/doxygen/html/mouse_8cpp_source.html)實作受控按鍵清除與目前模式中心；`0101h` 中心 `320,240` 僅平台規格近似，未取得原版重設後位置的同狀態收據。

**已證實，敏感度返回及消費端**：同一工具／資料的 `--mouse-sensitivity` 在 **DOSBox-X CS:EIP** `0180:0038031B → 0180:0038031D` 記錄功能 `001Bh` 返回 BX／CX／DX=`0032h`（50）、AX 與其餘擷取欄位、EFLAGS=`0012h` 不變。返回寫入相同 record，後續重用它準備 `AX=3` 位置查詢。私有 JSON SHA-256 `3becc83160dd11465db3a11e647f31703c8a1b4750639ee601d6edffd9b14e05`、caller LOG SHA-256 `9122014bd929cb6ed0caa21ad8d174d150d847c987a45e1a2212f255dda7b049`、終端 SHA-256 `4b8d8343ce60dc87743da49c610f201de76b05fe3a2da7acf5d430819a2cdbd1`。[規格 252](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/252-moo2-protected-mouse-sensitivity-query.md) 從既有敏感度狀態查詢，沒有固定回 50。

**勘誤，已證實的控制條件**：初稿曾把舊缺檔探針的 `3 → 21h → 1Ah` 零設定套到現行完整資料，誤判 dosgolem 目前為零。新增正式探針逐筆印出滑鼠服務，完整資料實際先走 `0000h → 001Bh` 並回 `50/50/50`。原版 `--mouse-sequence` 從固定 DPMI 零基底分支逐筆擷取，亦先走 `0000h → 001Bh` 並回三個 50；JSON SHA-256 `a39e0dbb15d77c67a531c0f652c40234a45c8f4e7af1f843f9a60a7ef529c6a5`、caller LOG SHA-256 `9122014bd929cb6ed0caa21ad8d174d150d847c987a45e1a2212f255dda7b049`、終端 SHA-256 `530142ef79e2de1b0b9c34c44f8d3b807d02d9eb124c2056e82dae5026145755`。舊缺檔證據仍可由規格 229／230 回查，但不能用來推定本輪完整資料初態；兩側旗標、堆疊與完整狀態仍不同。

**已證實，dosgolem 自生驗收**：Go 1.24.13／`golang:1.24-bookworm`，隔離 `fb54112` 加本輪受限實作。規格 251／252 的欄位保留、模式中心、敏感度狀態往返與一般 FD2 拒絕測試通過。固定 EXE 全套測試輸出 SHA-256 分別為 `f7ac2730974ecee63a4b2e4146f9d5850e7aa829d903a14525c62c3c731b16b5`、`37ba0e0c1c59433302b0079af252c6b479945ff760b49f9a9d7e466ebf89cdb0`。原檔自然越過重設後，第 1,545,515 步抵達敏感度查詢，私有收據 SHA-256 `e8d6b4527a940ea1aeb0c4391e7931cd41ed6f8e6e763b951d21bacfa0961c43`；查詢接線後第 1,545,719 步於 **dosgolem 高位 LE 線性位址** `0x24C31B` 的 `INT 33h/AX=0007h` 拒絕，CX=`0`、DX=`04FEh`（1278）、EFLAGS=`0012h`，私有 `workplace/moo2-probe-252-full-game.txt.gz` SHA-256 `d630b6db5fc796dc91164b3e15219083ada2f8aed8bf0a7e263eb6b64708e666`。這僅驗收平台服務；下一步核對原版範圍設定與後續使用，不宣稱正常玩家畫面、音效、受控亂數或與 Go remake 的玩法同狀態對拍。

## 2026-10-01：座標範圍與非零敏感度設定

**已證實，原版實際參數及同次返回**：沿用官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP SHA-256 `3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f` 根層 417 檔及 `MOX.SET` SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，版控探針 `--mouse-horizontal-range／--mouse-vertical-range／--mouse-set-sensitivity` 於 **DOSBox-X CS:EIP** `0180:0038031B` 直接核對 `CD 33 C3`，於下一指令 `0180:0038031D` 核對同次返回。三筆分別為 AX=`7`、CX=`0`、DX=`04FEh`（1278）；AX=`8`、CX=`0`、DX=`01DFh`（479）；AX=`1Ah`、BX／CX=`0064h`（100）、DX=`01DFh`。三筆均保持擷取的暫存器、段、堆疊與 EFLAGS=`0016h`；`0180:003801B8..003801C0` 將返回寫入 record `003D18E0h +0/+4/+8/+0Ch`。水平 caller 接著準備垂直呼叫；垂直 caller 接著準備非零敏感度；非零設定 caller 回到位置查詢鏈。

| 私有收據前綴 | JSON SHA-256 | caller LOG SHA-256 | 終端 SHA-256 |
|---|---|---|---|
| `mouse-horizontal-range` | `4fb53b8b5813d60729960cb46cc8dd7cae6feeb62105467a62430de98ea9c998` | `f9f230d1c5b1a75a89917dda9b02f4eb27bcb5d4d1951046c8f3dd75c26774ba` | `4b8df77b6b79d7b4f10f00bc867f4892a33d4ee37cba420de614fc74f013537c` |
| `mouse-vertical-range` | `bf3df1ba3de00614f1754f5b8fb9f73cf68b0a1cc4e909e34b89fd7f25676853` | `fb71a9140aa4ecc56bdd9f15428c5d7be4c5eddc3f80da32ad7b3457d3f60d58` | `8f1783e1c319338ecf1d5404a469735bd691bdde16e462b7f1786b405872ef95` |
| `mouse-set-sensitivity` | `6fcbabf3b96c68eda2bea2032090f5e2a2fa25e098a16999229b8d407eb0d57f` | `01d24378d6ffa0b5c7fda1532177ab6f3e623f64994907aa383cce5f96d06268` | `a829f04c341c922bdbc8fc46ed01b83a7aeef983c107c243f8a6c423b8385826` |

**平台規格近似與原版邊界**：[規格 253](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/253-moo2-protected-mouse-coordinate-ranges.md) 依 [DOSBox-X 平台來源](https://dosbox-x.com/doxygen/html/mouse_8cpp_source.html) 保存有號 16 位範圍、限制目前及受控注入座標；[規格 254](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/254-moo2-protected-mouse-sensitivity-settings.md) 保存各軸／倍速上限裁切至 100 的設定值供 `1Bh` 讀回。負值、反序、邊界位置、非零讀回與高位保留只按公開平台契約及合成測試驗收；沒有原版設定後移動實驗，不能宣稱原版實機的座標粒度、移動速度或完整驅動。舊缺檔零值樣本仍有效，規格 230 的歷史非零拒絕邊界已回填規格 254，版控 `--check-mouse-spec-backlinks` 正常與缺勘誤拒絕核對通過。

**已證實，dosgolem 自生前進**：隔離 `792b8d077d1d4611ae8246ddc9d61129e316b67a` 加本輪實作、Go 1.24.13／`golang:1.24-bookworm`。固定 EXE 全套 `go test -buildvcs=false ./... -count=1` 的範圍階段輸出 SHA-256 `261da4ea09ba0cbd7753a49d07ed0fabae28f1c470748a1e4f707edfa13a28d3`，自然越過兩軸後於第 1,545,911 步抵達 `1Ah`，私有診斷 SHA-256 `76521df20677fe92c685fe909a1fd13fdd41b25ffc2955051d4c9359d8ddd04d`。非零設定階段的最終全套測試輸出 SHA-256 `fe9cf5242d01be373fdea1c67d9869165446d18bdf267831524cb6af4cf0fd01`；完整資料自 LE entry 自然越過設定及後續位置查詢，第 1,546,160 步在 **dosgolem 高位 LE 線性** `0x24C31B` 的 `INT 33h/AX=000Ch` 拒絕，CX=`1`、EDX=`0x2136D1`、EFLAGS=`0012h`，DPMI 未實作清單為空。私有 `workplace/moo2-probe-254-full-game.txt.gz` SHA-256 `3f3af2c4aeef4fc15686b2a281ab35d8ac6b7c432c3016205277db6e85f9dfe7`。下一步核對原版回呼的完整參數、返回與觸發鏈，不把回呼靜默當成完成；PSP／環境、堆疊與旗標仍不同，尚無正常玩家畫面、音效、受控亂數或與 Go remake 的玩法同狀態收據。

## 2026-10-01：滑鼠回呼、CS 段載入與位置設定

**已證實，原版輔助事件鏈**：固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔及固定 `MOX.SET`，工具沿用 DOSBox-X 2026.07.02 SDL2 heavy debugger／`fd2-dosbox-x:debug-0d7b272b`。版控探針 `--mouse-callback` 於 **DOSBox-X CS:EIP** `0180:0038031B` 擷取 AX=`000Ch`、CX=`1`、ES:EDX=`0180:003476D1`；同次下一指令所擷取欄位保持。`--mouse-callback-event` 在 caller 重新啟用中斷後，以一般 X11 滑鼠移動 `8/6` 命中登錄入口，AX=`1`、BX=`0`、CX／DX=`657/189`、SI／DI=`0`。原版將 CX 右移一位寫 `003D1A38h`、DX 寫 `003D1A36h`、BX 寫 `003CF21Ah`；收尾 `0180:003477EC..003477F2` 的 `89 EC 5D 5F 5E 1F CB` 執行遠返回。堆疊前八 bytes 是 32 位元 offset `E3h` 與四位元組 selector `98h`，下一步 `0098:000000E3` 且 ESP 增八。只保存最小橋接邊界，不挖 extender／IRQ 內部。

**已證實，原版同次平台服務**：回呼 LOG 的 `0180:003341E4` 從 `CS:[003341ED]` word=`0188h` 載入 DS，下一步 DS=`00A8h → 0188h` 且所列旗標／一般暫存器不變。另以 `--mouse-set-position` 擷取 `0180:0038031B` 的 AX=`4`、ECX=`0280h`、EDX=`003400F0h`，同次 `0180:0038031D` 所列欄位保持，caller 繼續初始化；EDX 高位不是 Y 座標。限定原始定位、工具映像 ID、輸入與所有私有收據雜湊見 [規格 255](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/255-moo2-protected-mouse-callback.md)、[256](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/256-cpu386-cs-absolute-ds-load.md)、[257](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/257-moo2-protected-mouse-position-setting.md)。原版記憶體與完整終端不入公開版控。

**平台模型與自生驗證**：隔離 dosgolem `438d6cc5971c3e212e0ce949e1ddd61de307f794` 加本輪實作，Go 1.24.13／`golang:1.24-bookworm`。按 [DOSBox-X 滑鼠平台來源](https://dosbox-x.com/doxygen/html/mouse_8cpp_source.html) 補遮罩、受控事件 FIFO、IF 閘門、非重入及座標設定，按 [Intel 80386 手冊](https://read.seas.harvard.edu/~kohler/class/aosref/i386.pdf) 補 CS 絕對 word 至 DS 的窄 CPU 形狀。私有有界堆疊與架構恢復屬 **platform-spec approximation**，不宣稱 selector／IRQ 時序精確。合成回呼實際寫出六個參數；遮罩、按鍵、範圍、佇列上限、重設、解除、狀態恢復、舊 hook／BIOS 時鐘與錯誤框架拒絕通過。固定 EXE 全套回歸最終輸出 SHA-256 `33ca5589160c47f992c5cdeec08b6b65ded02dfc6a03fb705aa698c21a05f26e`。

**已證實，兩條原檔啟動路徑**：無事件原檔由 LE entry 自然越過 `AX=000Ch／4`，第 6,216,999 步停在 **dosgolem 高位 LE 線性** `0x217888` 的 `CD 2F`、AX=`160Ah`，私有 `workplace/moo2-probe-257-full-game.txt.gz` SHA-256 `3b7e90bfcb64cfdbbbf9f834e7f7bd185ddc6f94fcee56cd8312c5fc812653c2`。另在原檔自然設定座標後首次 IF=1 排入明示 `657/189/buttons=0/mickey=0/0`，回呼 started=1／completed=1，返回後第 6,217,034 步抵達同一服務；私有 `workplace/moo2-probe-257-mouse-event.txt.gz` SHA-256 `5cc4ec566b82857814f12e271e0dedd88f8f283835e9505cac0f7c6dc88324ba`。註冊／位置／回呼都由 dosgolem 自行執行，不將 DOSBox-X 圖片包裝成正式收據。

**未知與停止線**：上述只證原檔早期回呼返回，尚未以相同完整初態核對座標／游標消費支線；規格 255 維持 READY，256–257 僅限定指令／位置平台模型 CONFORMED。PSP／環境、堆疊、旗標、實體速度與完整輸入不同，仍無正常玩家畫面、音效、固定亂數或與 remake 的玩法同狀態收據。下一最小行動是核對 `INT 2Fh/AX=160Ah` 的原版返回與 caller，依公開介面契約補受限支援；不追 Windows 平台內部或硬體逐週期考古。

## 2026-10-01：DOS 環境的 Windows 版本查詢

**已證實，原版輔助返回**：固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版根層 417 檔及固定 MOX.SET；DOSBox-X 2026.07.02 SDL2 heavy debugger，既有 `fd2-dosbox-x:debug-0d7b272b` 映像。版控探針 `--windows-version` 捕獲 **DOSBox-X CS:EIP** `0180:0034B888` 的 `CD 2F`、AX=`160Ah`；同次下一指令 `0034B88A` 的全部擷取一般暫存器、段、堆疊及旗標保持。caller 比較 EAX 後跳至 `0034B89C`，略過 Windows 版本保存，恢復暫存器並返回零後繼續初始化。JSON SHA-256 `a35f451af240449b66a8762a05d6798bd4dc49a901fa53b28490e83e92d7904b`、caller LOG SHA-256 `6d787297c2f2aa296b7c4681c6d53fcf0a79bac9749f096dbdb94b8b7ae0a001`、終端 SHA-256 `fc0d6f75a86828ed1c4aabfbf4e7ef04209fe869f51d948c21b5f6f246bf2a77`。原始 bytes、工具版本與定位見 [規格 258](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/258-moo2-windows-version-absence.md)。

**限定平台契約**：[DOSBox-X 多工服務來源](https://dosbox-x.com/doxygen/html/dos__misc_8cpp_source.html) 與原版同次返回相符。隔離 dosgolem `7a50ae93495ed70d00e5fb2be1e129a569c5ef5f` 加本輪實作，只在明示 MOO2 DOS 設定接受低 AX=`160Ah`，保持完整架構欄位，不偽造 AX=0 的 Windows 成功返回；其他設定與未知多工服務仍拒絕。不展開 Windows 或 DOS 多工鏈內部逆向。規格 258 在此限定範圍 CONFORMED，不代表通用 Windows 支援。

**已證實，自生兩條路徑**：Go 1.24.13／既有 `golang:1.24-bookworm`，固定原檔全套 `go test -buildvcs=false ./... -count=1` 通過，私有測試 SHA-256 `1a7ddbe2c287b573ea39b942b954bf9a69ec5a95e7c1842b060c6edd0241196c`。無事件第 6,217,167 步、設定後受控事件第 6,217,202 步停於 **dosgolem 高位 LE 線性** `0x24C315` 的 `CD 31`、AX=`0500h`；無事件診斷 SHA-256 `3522de797c533648583a633a451530536bdc43a6b945edadd5ee76a7341f31d1`、事件診斷 SHA-256 `9fd435c4c06ca6ee936bebfdee48efb0a6eeca360ccc328bee44e642e2303743`。原始資產、記憶體與完整終端留本機私有。

**未知與下一步**：規格 255 仍待完整座標／游標消費同狀態驗證；尚無正常玩家畫面、音效、受控亂數或 remake 玩法收據。下一個最小行動為公開 DPMI 記憶體資訊契約與原版 `0500h` 返回核對，不挖 extender 內部。

## 2026-10-01：DPMI 可用記憶體資訊及容量消費

**已證實，原版平台返回與最小 consumer**：固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版根層 417 檔及 MOX.SET，既有 DOSBox-X 2026.07.02 SDL2 heavy debugger／`fd2-dosbox-x:debug-0d7b272b`。版控 `startup_probe_131.py --free-memory` 在 **DOSBox-X CS:EIP** `0180:00380315` 擷取 `CD 31`、AX=`0500h`，ES:EDI=`0188:003EBB38` 的 48 bytes 由全零變成記憶體資訊；同次 `00380317` 擷取架構欄位保持。第一個 dword=`01913000h`，另八欄及保留區見 [規格 259](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/259-dpmi-free-memory-information.md)。`0180:00334FD2` 的原始 `[ebp-007C]` 取此第一欄位，加到 EAX=`E810h`，右移十位後由 `0180:00234D5C..00234D68` 做容量檢查。不深挖前一個 runtime 記憶體 helper。
私有 JSON SHA-256 `0e5b4b0d324a843ce4f4aa06b70d47e7b8d0bfd7951806acb75b3e9a8e7e14ed`、caller LOG `929e2c6aa680b6e3932bc4692131c3b08e37c00c56ae09fd4f59b1d0dd378f80`、外層 consumer LOG `af617b57c1baaf6c28c785774f476f48fb05dec1b6dfec63ae76644a3df286bc`、終端 `83dceb963801edfa98e9e8c014f712627309d4135d09add5ff51f9271660926c`。原版容量是輔助環境返回值，不寫成遊戲規則或 dosgolem 常數。

**限定平台實作與正式自生診斷**：[DPMI 1.0 原始規格第 99–100 頁](https://docs.pcjs.org/specs/dpmi/1991_03_12-DPMI_Spec_v10.pdf) 已定義 48-byte 介面、CF 與未知欄位標記。隔離 dosgolem `68e0ebca82cdcb436c7910bc676a818167215343` 加本輪實作，只回報既有 64 MiB 配置器可兌現的最大連續區塊，其餘未模型化的分頁欄位填未知；屬明示的平台規格近似。有效寫入只清 CF；不支援 buffer 形狀先原子拒絕。合成測試實際配置全部回報容量、釋放／重用、容量耗盡與各種越界／唯讀／不同 backing 拒絕通過。Go 1.24.13，固定 EXE 全套 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 通過；測試 SHA-256 `4f556544b654a81516bd019606d95945855e71aa6f1eae55d44c7563b421fc6d`。

**已證實，兩條自然新停點**：無事件第 6,713,034 步、設定後受控事件第 6,713,069 步，均於 **dosgolem 高位 LE 線性** `0x200E5A` 的 `C1 7D F4 04` 明確拒絕，受控路徑回呼 started=1／completed=1。無事件診斷 SHA-256 `e6ce6f569119d6ef82e0b319c6ff98c65acc80c8b17ae2a25f9ded0c5f2596a9`、事件診斷 SHA-256 `7b6bedae196683e1f67dc0af1f2356cadde7e593ea54e92f95bfdce75ae645b7`。規格 258 的舊停點已回填至 259，正常護欄與刪標記必拒絕通過；保留舊收據，不重寫歷史。下一步只核對此窄 CPU 記憶體移位形狀。

**仍未知**：完整座標／游標消費、正常玩家畫面、音效、受控亂數及與 Go remake 的玩法同狀態收據；規格 255 不升格。未比較原版 extender 的容量逐值或全遊戲分頁欄位使用。

## 2026-10-01：堆疊 dword 的立即數算術右移

**已證實，原版同次 CPU 樣本**：固定 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版根層 417 檔與 MOX.SET，既有 DOSBox-X 2026.07.02 SDL2 heavy debugger。版控 `--sar-stack-memory` 捕獲 **DOSBox-X CS:EIP** `0180:00334E5A` 的 `C1 7D F4 04`。原始運算元 `SS:[EBP-0Ch]`，EBP=`003EBB84h`、SS=`188h`；`0188:003EBB78` 的 dword=`32 → 2`，下一指令 `00334E5E` 的擷取一般暫存器及段保持。EFLAGS=`246h → 212h`；CF／PF／ZF／SF 與結果相符，AF 與多位 OF 依 [Intel 原始手冊](https://read.seas.harvard.edu/~kohler/class/aosref/i386.pdf) 未定義。原版 AF=1，執行器沿既有策略清 AF，不宣稱完整旗標逐位元相等。下一無條件跳躍到 `0180:00334E6A` 的 CMP 重算旗標；不追所在 helper 的語意。
私有 JSON SHA-256 `c6030ecb0371d9a5521f4581934d459423a0f2832e9591ad750247eb04e3136b`、caller LOG `7ef34efe2fa2add0e082fd7e4f2b4b9aa19726ffcd68c059d2e23bb02d0af093`、終端 `0be4a8d736a44d0699d876865c96497ce48cbc32d7af218b570ead7f3528bbaa`。首次候選位址換算少 `1000h` 未命中，修正後由實際 bytes 與同次返回確認；原始失敗終端留本機，工具間位址差不單獨當定位證據。

**CPU 實作與測試**：隔離 dosgolem `a6a7b79a60c9656f93b216b539707ff9b52b5671` 加本輪實作，規格 [260-cpu386-sar-stack-dword-immediate.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/260-cpu386-sar-stack-dword-immediate.md) 依公開 CPU 契約及原檔形狀先 READY，再接無前綴 ModRM=`7Dh`、SS／EBP／有號 disp8、count 低五位、定義旗標及有界 dword 存取。正負／零／端點、計數 `0/1/4/31/32/33/255`、SS 與 DS 分離、鄰接哨兵、唯讀／越界／未知形狀／截短／寫入拒絕測試通過。原版同次 CPU 輸入的值與定義旗標亦通過。加入樣本的第一次測試用了錯誤段排列，被 SS 可寫檢查拒絕；改具名索引，一併修正規格 258 的同類測試夾具並追加勘誤，服務／CPU production 行為不改。固定 EXE、Go 1.24.13 最終全套 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 通過，測試 SHA-256 `12e3a3f78ea25fa3c1c4745b78faaf2b008b0466837bee54e06393b4445c23a4`。

**已證實，兩條自生新停點**：無事件第 6,725,897 步、設定後受控事件第 6,725,932 步，均於 **dosgolem 高位 LE 線性** `0x229A59` 的 `CD 21` 拒絕，EAX=`002B4E38h`、DS:EDX=`0188:002BDB38`、CX=`0`。現有受限 `4Eh` 服務在這次輸入未通過，原因與搜尋字串待查，不把它改寫成整項未實作。受控路徑回呼 started=1／completed=1。無事件診斷 SHA-256 `66b324e1fce7d3e260d550eeb7e8c16b7058528074ec3bb7130ed7fdab94eac7`、事件診斷 SHA-256 `dda9b0d9cc2b663a8bef3c4cf6735f19e7c8327c187df6864175d7eb97593732`。

**回填與未知**：規格 259 回填 260，定位／狀態／連結及刪除標記必拒絕護欄通過。規格 258 的段夾具勘誤保留舊輸出來源。仍無正常玩家畫面、音效、受控亂數或 remake 玩法同狀態收據；255 不升格。下一步只抓 `4Eh` 搜尋輸入、DTA 狀態及原版返回，不深挖標準檔案服務內部。

## 2026-10-01：目前目錄前綴的 DOS 搜尋

**已證實，自生輸入與原版輔助返回**：隔離 dosgolem `8f4c579beda504a66aefd74186fa6d1fa8d3d072` 的固定官方 1.31 EXE（SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`）於高位 LE 線性 `0x229A59` 搜尋 `.\simtex.lbx`，CX=0、DS:EDX=`0188:002BDB38`，DTA=`0188:00295828`。正版 ZIP 沒有此檔；既有服務支援不帶前綴的精確檔名，這次因 `'.\'` 被拒絕。私有診斷 SHA-256 `f1865f4101ce2b6e5a3054e2d4d7875a3503a24105c45898e0c717489e0321c1`。

既有 DOSBox-X 2026.07.02 SDL2 heavy debugger 的 `--find-current-directory` 實際捕獲 **CS:EIP** `0180:0035DA59 → 0180:0035DA5B`，原始 `CD 21 E8 AC 65 01 00 89 DA E8 21 00 00 00 59 C3`；同一字串位於 `0188:003EBB18`，DTA=`0188:003C3828`。EAX=`003E4E18h → 12h`，EFLAGS=`246h → 247h`；其餘已擷取暫存器保持。DOSBox-X 是輔助基準，PSP／堆疊／位址與 dosgolem 不同，不宣稱同狀態。DTA 公開結果區保留 DIPLOMAT.LBX；原版保留區 `+0Ch` 由 `04h → 05h`，執行器仍保留此區，差異明示為未知。第一個 caller 依 CF 使用錯誤 `12h`，不深挖檔案 helper。JSON／caller／終端 SHA-256 分別 `d3d2ee93dc1f9f812d81d2ac37ee55e312d427e530b664825c71388389ad987d`／`6b05980307ee4e21811c58da5f8832c0e9a594c224b6ea4bb35892402ec793c6`／`bc81ece31155571252f243c1d1ad0027f94164fc42b476eebf5f9447cf148fcb`；原始資料皆留本機。

**限定實作與驗收**：規格 [261-moo2-dos-findfirst-current-directory.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/261-moo2-dos-findfirst-current-directory.md) 先 DRAFT，依 [DOSBox-X 平台來源](https://github.com/joncampbell123/dosbox-x/blob/master/src/dos/dos_files.cpp) 與上述原版返回轉 READY，再實作只在 MOO2 設定消去一次目前目錄前綴；唯讀提供者仍只收到單一 8.3 檔名。成功／缺檔、最大長度、提供者輸入、哨兵、保留欄與拒絕路徑測試通過。固定 EXE、Go 1.24.13 全套 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 通過，SHA-256 `c498ec94eab675212c95c76d394bacada8ca6d51f8c88112d9488139b68d6fd9`；CONFORMED 限於前綴、返回與公開結果欄，不包含原版 DTA 保留區逐位元相等。

**已證實，兩條自生新停點**：無事件第 6,725,985 步、設定後受控事件第 6,726,020 步，均到 **dosgolem 高位 LE 線性** `0x114F43`，原始 `66 85 C0 0F 85 F5 00 00 00 B8 17 0C 26 00 31 D2` 因 TEST 的運算元大小前綴 拒絕；EAX=0，受控回呼 started=1／completed=1。診斷 SHA-256 `872b6e8979491ade99b79c5d2aa3e1734bf257b22fc0d9a9dc4293ec288e1f49`／`c27af6189cc321d44b7455abb6f7fa016e575d1ffe97fe40eeef9aa0ecd770c0`。規格 260 回填 261，219 連到延伸範圍；原停點保留為歷史證據。仍無正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態，255 維持 READY。下一步只依公開 CPU 契約核對此指令的寬度與定義旗標。

## 2026-10-01：16 位元暫存器 TEST

**已證實，原版同次 CPU 使用樣本**：固定官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，正版根層 417 檔與固定 MOX.SET，DOSBox-X 2026.07.02 SDL2 heavy debugger；版控 `--test-word-register` 實際捕獲 **CS:EIP** `0180:00248F43 → 0180:00248F46`，原始 `66 85 C0 0F 85 F5 00 00 00 B8 17 EC 38 00 31 D2`。EAX=0，word TEST 的暫存器／段／堆疊保持，EFLAGS=`246h → 246h`。第一個 JNZ 未跳轉，抵達 `0180:00248F4C`。原版與 dosgolem 窗口後面的絕對立即數有重定位差異，不宣稱全部 bytes 相等。私有 JSON／分支 LOG／終端 SHA-256 分別 `67c6888c34e674e54b614556765617356fff274ac5d217b4ee3a6bebdc2a45ae`／`70d2a992af8a687163af0a98f9a7c3b8d36d996da4477f25c3d0727b54bb50ed`／`c1880f5f784129c43d4376490e610e76676f2d24f40a5a1f484624663b1bca68`；輸入與工具映像雜湊沿用前述規格，完整資料留本機。

**CPU 契約與限定實作**：依 [Intel 原始 TEST 指令頁](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/TEST.htm) 及 [旗標附錄](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/appc.htm)，規格 [262-cpu386-test-word-register.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/262-cpu386-test-word-register.md) 先 DRAFT，再 READY 後接 `66 85 /r` 全部暫存器配對。只取低 16 位，重用 word 旗標路徑，保持資料；AF 未定義，沿既有清 AF 策略，此樣本相同不外推其他值。64 配對、零／符號／parity／高位干擾／遮罩、32 位寬度、拒絕邊界與具名索引原版 CPU／第一個分支樣本通過。最終固定 EXE、Go 1.24.13 全套 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 通過，SHA-256 `869b2683fc5a8f0f400d2f2b6a5134bfc48e0e2e4d4caeecedec342e70aa2a79`；規格 262 限定 CPU 範圍 CONFORMED，不改 Go remake 玩法。

**已證實，兩條自生新停點**：無事件第 6,728,298 步、設定後受控事件第 6,728,333 步，均在 **dosgolem 高位 LE 線性** `0x222C9E` 的 `EE E8 9D FE FF FF 66 BA C8 03 8B 35 80 38 2A 00` 因 `OUT DX,AL` 的埠 `03C6h` 未接通而拒絕；AL=`FFh`，受控回呼 started=1／completed=1。診斷 SHA-256 `75c166c25246e0c8a79edfc5f4586f07c73dccccf5aca9ef82428d792577acb3`／`803916a40f99759906ac59bb207fe7528ba6848a861cd106ebb315ac17b39b79`。規格 261 回填 262，原始 TEST 停點仍可回查。下一步只核對公開 VGA DAC 與既有調色盤埠路由，不追顯示 driver 內部；仍無正常玩家畫面、音效、受控亂數或 Go remake 玩法同狀態，255 不升格。

## 2026-10-01：標準 VGA DAC 像素遮罩

**來源與限定規格**：隔離 dosgolem 起始 `f9e043fb41de8fff6e2ec3b49b0fd0695712af12`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、ZIP 根層 417 檔及 MOX.SET。規格 [263-vga-dac-pel-mask.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/263-vga-dac-pel-mask.md) 先 DRAFT，依 [IBM VGA／XGA 技術參考 2-104 至 2-106](https://bitsavers.trailing-edge.com/pdf/ibm/pc/cards/IBM_VGA_XGA_Technical_Reference_Manual_May92.pdf) 與 [DOSBox-X 公開 DAC 介面](https://dosbox-x.com/doxygen/html/vga__dac_8cpp_source.html) 審查轉 READY。IBM 表格與寫入警告的限制明示，標準遮罩讀寫及查色索引 AND 由成熟模型交叉確認；不複製控制流、不追顯示 driver。實作後只在 **hardware-spec approximation（硬體規格近似）** 範圍 CONFORMED。

**已證實，原版同次輔助使用點**：既有 DOSBox-X 2026.07.02 映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，探針 `--vga-pel-mask` 命中 **CS:EIP** `0180:00356C9E`，bytes=`EE E8 9D FE FF FF 66 BA C8 03 8B 35 80 18 3D 00`；DX=`03C6h`、AL=`FFh`，下一指令 `0180:00356C9F` 的一般暫存器、段與 EFLAGS=`246h` 保持。與 dosgolem 高位 LE `0x222C9E` 前十 bytes 相同，後續絕對位址重定位不同。未額外讀埠；非 FF 遮罩依公開規格測試，不冒稱原版已動態使用。JSON／後續 LOG／終端 SHA-256 分別 `959fa5f9e9cc867d470a937469bd7669bd10bb8de81496cd1131241f5faef267`／`933bff7d2cc559e17f97dfc64f24028cfcd17f0414785a596c9ee47235891216`／`6d2db41eaa5b77c30389809495c1aad11c278e98487f7be05086876ade2ed8fd`。完整架構、輸入／工具雜湊與限制見規格 263，原始資料留本機。

**實作與驗收**：遮罩接 `Machine` 與 LE 埠，`Palette`／平面 RGB／LE Palette 實際消費；支援的模式設定恢復 FF。記憶體快照及 gob v2 存檔同時保存，存在欄位區分合法 0 與缺欄位舊檔，舊檔沿無遮罩查色，不從 Ports 猜值。全部色號、五組遮罩、原始 DAC 不變、RGB 中途讀寫、索引環繞、兩種還原、舊格式與音訊隔離測試通過。Go 1.24.13，既有映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，輸出 SHA-256 `3255ab2d5d3bb47cc53fd0739430536048704cc1fcf602342ac88c8bedfec559`；針對遮罩／VGA 測試輸出 `ba6cdc65bff8c8fc6aa2474276d6eb7159f74a8cbc7140e81e8797f2a7254d7f`。

**已證實，兩條自生下一停點**：同一輸入與分離 DOS arena，無事件第 6,728,365 步、設定後受控事件第 6,728,400 步，均越過遮罩寫入，在 **dosgolem 高位 LE 線性** `0x222CCB` 的 `F6 F3 EE AC F6 E7 B3 64 F6 F3 EE AC F6 E7 B3 64` 拒絕。EAX=0、EBX=`64h`、ECX=`80h`、EDX=`3C9h`、DS／ES／SS=`188h`、EFLAGS=`6h`；受控回呼 started=1／completed=1。通用錯誤寫「F6 記憶體形狀未支援」，實際 ModRM 是暫存器 byte DIV，不誤分類。兩份診斷 SHA-256 `a5836830e9c15f632e22f7a110de950604f7cea87b8d3076c693d3d892ef86ba`／`9adbbef63c5fa11f133f1f16c9785a0d5a4efc5d755b4cc94b8e2b3da4cb8df3`。

**回填與未完成範圍**：規格 262 回填 263，Python 語法、索引、既有護欄與新護欄正常／缺舊標記／缺原始定位必拒絕通過。255 仍 READY，正常玩家畫面、音效、受控亂數及 Go remake 玩法同狀態未驗證；不把 VGA 接通算成玩法矩陣閉合。下一步只核對公開 CPU byte DIV、既有解碼與固定使用點，窄規格審查後接通並重跑。

## 2026-10-01：位元組暫存器的無號除法

**來源與有限規格**：隔離 dosgolem 起始 `3cfd84be85738450b65102f46450d46ae91429f1`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、ZIP 根層 417 檔及 MOX.SET。規格 [264-cpu386-div-byte-register.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/264-cpu386-div-byte-register.md) 先 DRAFT 並入索引，依 [Intel 80386 DIV 契約](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/DIV.htm) 與既有暫存器／除法錯誤策略審查 READY 後接 `F6 /6` 全部八種 byte 暫存器。先讀來源，商 AL、餘 AH，保存高半部及其餘資料；除法錯誤沿既有 CPU Error 停止，不模擬完整 `#DE`，算術旗標未定義且保留。

**已證實，原版同次商餘與消費**：既有 DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`--div-byte-register` 命中 **CS:EIP** `0180:00356CCB`，bytes=`F6 F3 EE AC F6 E7 B3 64 F6 F3 EE AC F6 E7 B3 64`，與 dosgolem 高位 LE `0x222CCB` 相同。AX=0、BL=100，下一指令 `0180:00356CCD` 商與餘皆 0、一般暫存器與段保持；第一個 OUT 寫 `03C9h` 的值 0，再至 LODSB。原版 EFLAGS=`46h → 6h`，執行器對該輸入保留 `46h`，未定義 ZF 差異明示；自然 dosgolem 停點輸入本來為 `6h`，不宣稱兩側整段同狀態。完整具名架構、位址基準及限制見規格 264。

JSON／第一個消費 LOG／終端 SHA-256 `35f4e556aedda3256a0b1b509515cfd07e25b888639e68db231e45d842cd18a7`／`de3b1aeafcba322d51f39cb73f75aa4451b6031e0f7695cbb6bf5f5a9447766c`／`bcf95c8e21a007893978e4672515577c5e5bbe73fc80fe0169ab364b44d4f667`。原版素材、完整記憶體及終端留本機；只保存最小 CPU bytes／架構樣本作測試，不追顯示 driver 內部。

**實作與驗收**：八來源、低／高 byte、AL／AH 別名、商 255／溢位、非零餘數、除以零、高半部與資料／段／記憶體保持、旗標策略、拒絕邊界與既有 MUL／TEST 回歸通過；具名原版樣本與第一個 OUT 另驗。Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，加入原版樣本後最終 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，SHA-256 `5e5a0f83ccfc9296e5748774c79dcc94d8798ac7644fdb51a64d518d3ec50742`；初次全套及 CPU 測試輸出亦保留。規格 264 在商餘、資料保持及消費範圍 CONFORMED，近似限制仍保留，不改 Go remake 玩法。

**已證實，兩條自生下一停點**：同一完整輸入、分離 DOS arena，無事件第 6,738,645 步、設定後受控事件第 6,738,680 步，均越過 byte DIV 及後續調色盤寫入，在 **dosgolem 高位 LE 線性** `0x228C54` 的 `CD 10 61 C3 60 25 FF FF 00 00 33 D2 BB 00 00 00` 拒絕。AX=`4F05h`、BX=0、CX=0、DX=5、DS／ES／SS=`188h`、EFLAGS=`246h`，拒絕後 EIP=`0x228C56`；受控回呼 started=1／completed=1。兩份 gzip 診斷 SHA-256 `1ba0ba7a2b7ee444757fd7e737ecd8a1d93501d1c5bb94150472cc3e83580500`／`fd6e7ba30851994f24da214fbdd147cb5bde144d092322a967c486dcc3371a45`。

**回填與未完成範圍**：規格 263 回填 264；Python 語法、索引、相關護欄正常／缺定位／缺舊標記必拒絕通過。255 仍 READY，正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態未完成；不把平台依賴接通算成玩法矩陣閉合。下一步只核對公開 VBE `4F05h` 視窗控制、原版使用點與既有模式資訊／顯存模型，不能用成功返回代替顯存區段（bank）切換及實際消費。

## 2026-10-01：VBE 視窗控制與共用顯存映射

**來源與規格**：隔離 dosgolem 起始 `39ac121bb2f2343809a9e3ecc67836a4d510faa1`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔及固定 MOX.SET。規格 [265-moo2-vbe-window-control.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/265-moo2-vbe-window-control.md) 先 DRAFT，依 [VESA VBE 2.0 Rev 1.1](https://www.phatcode.net/res/221/files/vbe20.pdf) 及規格 238 的固定模式資訊審查 READY 後實作。視窗 A 位於 A000h，粒度與大小各 64 KiB，容量 2 MiB；mode 0101h 的六個已報告影像頁重設時清除，頁外保存。標為 hardware-spec approximation（硬體規格近似），不追 S3 driver、實機回掃或 DAC 硬體時序。

**已證實，原版同次返回**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --vbe-window-control` 命中 **CS:EIP** `0180:0035CC54 → 0180:0035CC56`。原始 `CD 10 61 C3 60 25 FF FF 00 00 33 D2 BB 00 00 00` 與 dosgolem 高位 LE `0x228C54` 相同；AX=`4F05h`、BX=CX=0、DX=5，返回 EAX=`4Fh`，其餘擷取暫存器／段／EFLAGS=`206h` 保持。caller 首兩指令 POPAD／RET；完整具名架構與位址空間見規格 265。dosgolem 自然輸入旗標為 `246h`，不宣稱兩側完整初始狀態相同。

JSON／caller LOG／完整終端 SHA-256 `b2a8d36a2951d0105908e34769213ef9e3f7094fab135f88d874e676ab036e7b`／`e503634c31c70a0a903a3d5317001ffafcfa70ca2adfaa68595b1ac044331f9d`／`ee76924805774425f43edb214f88865dbbe1312ab388ebb2bfcee7e8106c18ff`。原始素材及完整收據留本機，僅保存最小 bytes／具名服務輸入作回歸。

**實作與驗收**：CPU 與 DPMI 實模式 byte 橋接共用活躍顯存映射；LE word／dword 讀取也處理跨視窗邊界。切換保存各區段，原始 Mem 不當作 VRAM。零起點索引及既有 DAC／遮罩 RGB 消費端使用獨立快照；模式 03h 停用映射，再設 0101h 清六頁、保留頁外，沒有新完整 LE 序列化。CPU 實際 MOV、DPMI 讀寫、0／5／31 切換、跨界、RAM 隔離、原版返回、不變欄位與未知形狀拒絕、消費端及重設測試通過。Go 1.24.13、映像 ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，輸出 SHA-256 `e253930379399d7b649170f5e72fb40f3bf1eac298f44fdeb2466d883fd0b307`。

**已證實，兩條自生顯存消費與下一停點**：同一完整輸入及分離 DOS arena，兩條原檔自然路徑均切換區段 5／6／7／8／9 並寫入 307,200 bytes。無事件第 6,738,834 步、設定後受控事件第 6,738,869 步，在 **dosgolem 高位 LE 線性** `0x228CA7` 的 `CD 10 61 FC C3 00 00 00 00 60 66 8B 1D 5E 3A 2A` 拒絕，AX=`4F07h`、BX=CX=0、DX=`200h`（Y=512）、DS／ES／SS=`188h`、EFLAGS=`246h`，DOS 呼叫 6；拒絕後 EIP=`0x228CA9`，受控回呼 started=1／completed=1。診斷命令為 `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，事件路徑加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。

兩份 gzip SHA-256 `d3fd50d6a07c913cacca03a827ef2d4558abd33a0743e9234f89cda46f4d32fd`／`13a24e91b05a4e7addbff28701ca9b3079acf9e80c7800a719eebea8ea9be558`。零起點索引快照 307,200 bytes 的 SHA-256 均為 `7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf`，尚未切到已寫影像，不能當作正常玩家畫面或逐像素收據。

**回填與未完成範圍**：規格 264 回填 265，239 的歷史返回範圍連到新延伸。Python 語法、Go 格式、索引、全部既有回填函式、新護欄正常／缺舊標記／缺定位必拒絕通過。265 僅在有限視窗／共用映射／消費端 CONFORMED，255 維持 READY；正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態未完成。SAR 的 AF、DIV 的 ZF／除法例外、DTA 保留區與平台差異仍明示。下一步核對非零顯示起點的原版返回及顯存圖像消費，不能把平台接通列為玩法矩陣閉合。

## 2026-10-01：非零 VBE 顯示起點與有效頁消費

**來源與規格**：隔離 dosgolem 起始 `dea1659135461bd75c796657cbaf202674d32b47`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔與固定 MOX.SET。規格 [266-moo2-vbe-display-start.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/266-moo2-vbe-display-start.md) 先 DRAFT，依 [VESA VBE 2.0 Rev 1.1，Function 07h，頁 29](https://www.phatcode.net/res/221/files/vbe20.pdf)、既有固定模式及原版同次返回審查 READY 後實作。Y 的單位是掃描線，Y=512 對應顯存 327,680 byte；只處理 X=0、完整頁可容納的有限垂直起點與讀回，標 hardware-spec approximation（硬體規格近似），不追回掃／S3 driver。

**已證實，同次輔助返回**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --vbe-display-start` 命中 **CS:EIP** `0180:0035CCA7 → 0180:0035CCA9`，bytes=`CD 10 61 FC C3 00 00 00 00 60 66 8B 1D 5E 1A 3D`。與 dosgolem 高位 LE `0x228CA7` 前 12 bytes 相同，末段絕對資料位址不同。EAX=`4F07h`、EBX=ECX=ESI=0、EDX=`200h`，返回 EAX=`4Fh`，其餘擷取暫存器／段／EFLAGS=`246h` 保持；caller 為 POPAD／CLD／RET，完整具名架構見 266。此樣本未擷取完整 VRAM 差分。

JSON／caller LOG／終端 SHA-256 `a4cd66a51d17c0c9d503bb36e9940236b3f5f05de539ccdfa7f265a09700c537`／`9ddaaebb3a2ab3fca0aad131a5c0e8acfa5920f96e7b24a5fa62453de5184881`／`730d9888d6a0d951d59a75ae6ed4eb9545cd2edaa61a92ee05e646e6edae1235`。完整原版輸入／終端留本機，最小 bytes／架構服務輸入加入回歸。

**實作與驗收**：有效 Y 只在完整頁／高位／子功能驗證後更新，設定不改顯存或寫入區段。CPU MOV 與 DPMI 映射寫不同頁，索引／RGB 使用同一起點與 DAC／遮罩，快照隔離；讀回高半部保持、Y=2796 最後合法值及 2797 越界、未知形狀拒絕、模式重設回零與歷史啟用前全零返回均通過。Go 1.24.13、映像 ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`go test -buildvcs=false ./internal/machine -run 'TestMOO2(VBE|ProtectedVBE)' -count=1` 及固定原檔 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，後者 SHA-256 `de01f4ef9623961307f6d7d4a432c0b2a47cd21a9d784be01416776799823bdf`。

**已證實，兩條自生下一停點**：同一完整輸入／分離 DOS arena，`DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，另一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。兩條均自行接受 Y=512，Active=true、Bank=9、StartY=512、BankSets=5、Writes=307200、DisplaySets=1；無事件第 6,738,873 步、設定後事件第 6,738,908 步，在 **dosgolem 高位 LE 線性** `0x234B10` 的 `66 83 C3 18 81 FB E0 01 00 00 7C 13 33 DB 66 BB` 拒絕。EAX=0、EBX=`F0h`、ECX=EDX=0、DS／ES／SS=`188h`、EFLAGS=`246h`、DOS 呼叫 6；拒絕後 EIP=`0x234B13`，受控回呼 started=1／completed=1。

兩份 gzip SHA-256 `edc4f60dcc59ccd1237a0381e1eda792a108f04f9b99e4cf62a95374b2a6ca4a`／`81c282173de1eae0a7161767c7f2715319048a57888379d4d745ab7afd73a019`。有效頁索引 SHA-256 均 `7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf`，RGB 均 `0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366`。兩條命令另設 `DOSGOLEM_MOO2_VBE_PNG=/src/workplace/moo2-vbe-266-full-game.png`／`moo2-vbe-266-mouse-event.png`，兩份 PNG SHA-256 均 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622`，已檢視為 640×480 黑色，索引仍全零。不同頁選擇由受控消費端測試驗證；此黑圖不能作正常玩家畫面、原版像素一致或可玩完成證據。

**回填與界線**：237 的歷史全零邊界與 265 的非零停點回填 266；正常／缺定位／缺任一舊標記必拒絕，Python 語法、Go 格式、索引、全部既有回填函式及擁有權驗證通過。上一輪 265 索引仍說原版使用點待核對，本輪依已存在的實際收據修正，沒有重開已完成項。266 僅在有限服務與消費端 CONFORMED，255 仍 READY；正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態仍未完成。DIV 的 ZF／除法例外、SAR 的 AF、DTA 保留區與平台差異仍明示。下一步核對公開 word ADD、旗標／別名保存與原版使用點。

## 2026-10-01：word 暫存器的帶符號立即值加法

**來源與規格**：隔離 dosgolem 起始 `6da3a10b07d8b272eb55baa52ef8b52d3704ae03`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔與固定 MOX.SET。規格 [267-cpu386-add-word-register-immediate.md](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/267-cpu386-add-word-register-immediate.md) 先 DRAFT，依 [Intel 80386 ADD](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/ADD.htm)、既有 add16 與固定原檔形狀審查 READY 後實作 `66 83 /0 ib` 的全部八種暫存器目的。立即值帶符號延伸，低 word 環繞，高半部保持；未列 segment／REP／記憶體形式仍拒絕，不改 Go remake 玩法。

**已證實，原版同次結果與消費**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --add-word-register` 命中 **CS:EIP** `0180:00368B10 → 0180:00368B14`，bytes=`66 83 C3 18 81 FB E0 01 00 00 7C 13 33 DB 66 BB`，與 dosgolem 高位 LE `0x234B10` 相同。BX=`F0h + 18h → 108h`，EFLAGS=`246h → 202h`，其餘擷取暫存器／段保持；後續 CMP 比較 EBX 與 480，EFLAGS=`287h`，JL 選至 `0180:00368B2F`。完整具名架構與位址基準見 267；不推論圖形 helper 用途。

JSON／第一個消費端 LOG／終端 SHA-256 `6cd17791b1d45eb584a30d10705cc56bbd0da7b0834f54036912cf5becf2ae70`／`435c7e65e91ebfa7ce1c5f103aaf9bf6f424c669f545f6e582f842b8e461fe82`／`02aae464be90675d5a66390683a120dade776c3c8e09a9e138d9c342fb0b4d07`。完整原版輸入、記憶體與終端留本機，僅最小 bytes／架構樣本作回歸。

**實作與驗收**：解碼沿用 add16，完整取立即值後才發布低 word。八目的、16 邊界值與全部 256 imm8，以無號進位／有號範圍／半位元組（nibble）進位獨立驗 CF／OF／AF；低 byte 同位、零、符號、高半部、段／非算術旗標與記憶體哨兵，未知／截短拒絕及既有 word／dword 操作回歸均通過。實際原版 ADD／CMP／JL 另有固定輸入測試。Go 1.24.13、映像 ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，`go test -p 2 -buildvcs=false ./internal/cpu386 -count=1` 通過，輸出 SHA-256 `ed4bf0766d93be14ad0b08dba6abe6d95902783cba3534ef0a51b287bf2d4623`；固定原檔 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，SHA-256 `251f0d83baf901593ff0ec2ec8e9cbe75b8c143df7cb19cc7afdf849fcac26a9`。

**已證實，兩條自生下一停點**：同一完整輸入及分離 DOS arena，`DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，另一路增加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。無事件第 6,738,883 步、設定後事件第 6,738,918 步，在 **dosgolem 高位 LE 線性** `0x234B43` 的 `66 F7 EB A3 7A 0D 27 00 33 DB 33 C9 33 C0 33 D2` 拒絕。EAX=1、EBX=5、ECX=EDX=0、DS／ES／SS=`188h`、EFLAGS=`246h`、DOS 呼叫 6；拒絕後 EIP=`0x234B46`，受控回呼 started=1／completed=1。公開 [Intel IMUL](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IMUL.htm) 定義 F7 `/5` 的 word 有號乘積與 CF／OF，其他算術旗標未定義；下一步只核對最小支援、旗標策略與固定使用點。

兩份 gzip SHA-256 `ea179f8ff25efa3f3ed64dd6f42ffda01194fa01120bd4d6c920caafcf663c9f`／`99c0094f44fb0b53cab736fa1e699b1be629f65a94874e1095061676748434df`。當次 StartY=512、Bank=9、Writes=307200，索引 SHA-256 均 `7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf`，RGB 均 `0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366`。另設 `DOSGOLEM_MOO2_VBE_PNG` 產生 `workplace/moo2-vbe-267-full-game.png`／`moo2-vbe-267-mouse-event.png`，PNG SHA-256 均 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622`，與上輪已檢視黑圖逐位元相同；不重做目視檢查，也不當作正常玩家畫面完成。

**回填與未完成範圍**：266 的 word ADD 停點回填 267，Python 語法、Go 格式、索引、全部既有回填函式、新護欄正常／缺定位／缺舊標記必拒絕通過。267 僅在 word 暫存器加法／定義旗標與首個消費範圍 CONFORMED；255 仍 READY，正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態仍未完成。DIV 的 ZF／除法例外、SAR 的 AF、DTA 保留區及平台布局差異保留；不將 CPU 支援灌入玩法矩陣分母。

## 2026-10-01：單運算元的 16／32 位元有號乘法

**來源與規格**：隔離 dosgolem 起始 `c0c1487222b5d0544f9551e78bf76c0a9e7fae43`，固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔與 553 bytes 固定 MOX.SET，來源雜湊不變。規格 [268](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/268-cpu386-imul-word-register.md) 先 DRAFT，依 [Intel 80386 IMUL](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IMUL.htm) 與同次輔助樣本審查 READY，再接 `66 F7 /5` 暫存器形式。兩條自然原版路徑越過 word 乘法，下一停點為 dword `F7 EB`；規格 [269](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/269-cpu386-imul-dword-register.md) 另行 DRAFT／READY 後接 32 位元完整有號積。八個暫存器來源皆先讀後寫，AX／DX 與 EAX／EDX 別名不受提前發布污染；只更新 CF／OF。未修改 Go remake 玩法或建立交付包。

**已證實，兩個同次原版樣本**：DOSBox-X 2026.07.02 SDL2 heavy debugger，映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。`startup_probe_131.py --imul-word-register` 命中 **CS:EIP** `0180:00368B43 → 0180:00368B46`，AX=1、BX=5 → DX:AX=0:5；`--imul-dword-register` 命中 `0180:00368B5F → 0180:00368B61`，EAX=F0h、EBX=280h → EDX:EAX=0:25800h。兩次 CF／OF 均清除，其他擷取暫存器／段保持，具名輸入與最小 bytes 見兩份規格。完整終端、記憶體與圖像留本機。

word JSON／消費 LOG／終端 SHA-256 `50c8e64a5a3905af139746bd6398f0b2a9716ce2c46b6167e5e944e45065397a`／`99185da5a3ed77b85fe0e333bbb4f0d1e860cbb28b99e0f505479adb0e254d3c`／`e35f0a6c285c7837960f1497a50ac777e560ffe260e16d195831648c9783afa2`；dword 為 `08aaefe24232f9eb5dff628b1583b2ab6dbe3dcb6ceaa5adbc1636cb90e6b3ec`／`010f07822d2dbd81a3932abd7c59151884134b469f1a29c947e579d0e6715181`／`a9ff40ee8294f0416b770cb03778eca2580a311d29bf50f21c9db1edc20bc140`。

**限制**：Intel 定義 SF／ZF／AF／PF 未定義。兩次原版 EFLAGS=206h，dosgolem 沿用既有 MUL 保存近似為 246h，未定義 ZF 差異明示；只核對定義 CF／OF。第一個 A3 不讀旗標，後續 XOR 重新定義算術旗標。A3 的 DOSBox-X 資料位址 0039ED7Ah 與 dosgolem LE 00270D7Ah 不同，後續 word MOV 的絕對位址也不同；架構最小測試重定位 A3 存值處，不能稱完整記憶體／正常玩家同狀態。不追 helper 內部或平台時序。

**測試與自然重跑**：Go 1.24.13，映像 ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。兩個切片各跑 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/cpu386 -count=1`，CPU 輸出 SHA-256 `df1dd325364a58f2ca750e8d922e3c58215fa9ee6715f8602ff7368023a697eb`／`381ddab03d13d086873980fd1b8a8af26120e448c0ba3b8e7434555a7ddeb827`。八來源、15 邊界值成對輸入、兩種旗標、完整積與別名、高位／其他狀態、未知／截短拒絕、原版架構與首個消費、既有乘法回歸皆通過；dword 預期以任意精度整數獨立生成。加入樣本後各跑固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1`，全套 SHA-256 `e75040d49e081e8551dc94d06d5edfbf7cb68e4696848cdc15aaacd192125331`／`866e3e9326063aaac6c508f8f63db64dc55a6e9fea08dee9bf8f27b5aa7b2aab`。

兩切片均用 `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`；另一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。268 無事件／受控事件第 6,738,892／6,738,927 步在高位 LE `0x234B5F` 拒絕，gzip SHA-256 `92d63d22a845d695808fad6d51a1a94faff3eb020db20a977483e0ab4c5140e6`／`391af2813ec33e2602aae500de1775797917219b1946783cdf41f30c9b571387`。269 第 6,738,950／6,738,985 步自行越過兩種乘法，在 **dosgolem 高位 LE 線性** `0x234C9A` 的 `39 0D 9B 0D 27 00 0F 8D 01 02 00 00 80 3D 94 0D` 因 CMP dword ModRM 0D 未支援而拒絕。ECX=18h、EAX=EBX=EDX=0、ESI=A5940h、DS／ES／SS=188h、EFLAGS=246h、DOS 呼叫 6，拒絕後 EIP=0x234C9C；受控回呼 started=1／completed=1。gzip SHA-256 `70be009fe10982a86b2f9c9929ffbea27fd5a02dcee685d9d737d46c4877ce54`／`67a3821afbfa52561ced7685fff1fffb51f4a59cfef74649784ce5a03f77bba4`。

**圖像與回填**：269 Active=true、Bank=7、StartY=512、BankSets=6、Writes=307200、DisplaySets=1；有效頁索引 SHA-256 `7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf`、RGB `0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366`。另設 `DOSGOLEM_MOO2_VBE_PNG`，兩切片共四份 PNG 均 SHA-256 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622`，與已檢視黑圖逐位元相同，沒有重做目視檢查或聲稱可玩畫面完成。267 → 268 → 269 停點回填，兩組護欄正例與缺定位／缺舊標記必拒絕、全部既有回填函式、Python 語法、Go 格式、索引與擁有權通過。

268／269 限定 CPU／首個消費範圍 CONFORMED；255 仍 READY，正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態未完成。DIV 旗標／例外、SAR 的 AF、DTA 保留區與平台布局差異仍保留；CPU 支援不計入玩法矩陣分母。下一步只核對公開 CMP 契約、實際記憶體輸入與第一個分支。

## 2026-10-01：32／16 位元目的比較與實際分支

**來源與範圍**：隔離 dosgolem 起始 `25b8e79e332bb8bb1359284e8d10822462a70a9b`，固定官方 1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、正版 ZIP 根層 417 檔與固定 MOX.SET，輸入雜湊不變。規格 [270-cpu386-cmp-memory-register.md](https://github.com/wicanr2/dosgolem/blob/a1a80765b6b1d527e2874057377afd1caab7a920/docs/spec/270-cpu386-cmp-memory-register.md) 先 DRAFT 並入索引，依 [Intel CMP](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/CMP.htm)、兩側目的與同次旗標／實際 JGE 審查 READY，再接 39 /r dword 記憶體目的。自然原版前進到 word 形式，再以 [271-cpu386-cmp-word-destination.md](https://github.com/wicanr2/dosgolem/blob/a1a80765b6b1d527e2874057377afd1caab7a920/docs/spec/271-cpu386-cmp-word-destination.md) 另走 DRAFT／READY，接 66 39 /r 記憶體與暫存器目的。只更新比較旗標，不寫回資料，沿用既有地址與分段層；未修改 Go remake 玩法或建立發行包。

**已證實，同次原版結果**：DOSBox-X 2026.07.02 SDL2 heavy debugger，映像 ID `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`。`startup_probe_131.py --cmp-memory-register` 命中 **CS:EIP** `0180:00368C9A → 0180:00368CA0`，DS:0039ED9Bh 的 dword 前後皆 0，ECX=18h，EFLAGS=246h → 297h；有界 LOG 2 與分支後 EV 確認 `0180:00368CA6`，JGE 實際不取分支。`--cmp-word-destination` 命中 `0180:0035CDCE → 0180:0035CDD1`，DS:[EDI] 的 word 與 AX 皆 0，EFLAGS=213h → 246h；LOG 2／EV 確認 `0180:0035CDD3`，JL 實際不取分支。其餘擷取暫存器／段保持，原始 bytes／完整具名架構、資料地址及輸入雜湊見兩規格。[Intel Jcc](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/Jcc.htm) 提供條件契約，不追下一個圖形 helper 用途。

270 JSON／消費 LOG／終端 SHA-256 `af5583d7dd5477aba5b864b1c48f345e6b8c3b8c53c3269f88a7af2ce505726d`／`a98188c1d69a6bf02a2dcfa32e7e8d716ddc7df416e4956b25d5e327f054a490`／`8171fce2bbdb15f16f5fc585bcdc41849d4e11545a2e3372b76dbef9f3935e1a`；271 為 `dabfcb5897c4d8650cfe568442bc1352674f1938572f196a98bee9f35a591e25`／`58b1476265461c2ac090d47c7e9c5e93e5ce1053a6a6b3cd0be4c9f10b26ffd9`／`6ffb3a247d37d81b0f80ecb1c3b5accbea02092d0f465fd0b96788125278171a`。完整終端／記憶體與圖像留本機，僅最小樣本進回歸；測試只重定位資料目的，絕對指標／布局不一致不隱藏。

**失敗分類與歷史勘誤**：270 初次 LOG 1 只記錄 JGE 前 EIP，改為 LOG 2／分支後定位後乾淨重跑，才用作 READY。第一次 CPU 失敗為規格／測試誤讀局部 39 分支，漏看全域 F2 拒絕；回 DRAFT，依原始全域閘門重新 READY，驗既有拒絕而不放寬產品。第二次失敗為既有 3B 回歸樣本未設 DS descriptor，補齊測試環境後同一命令重跑。失敗輸出與訂正證據保留於 270，不刪除回歸項。

本輪 LOG 行為還證明歷史 267 的 LOG 2 只直接包含 CMP 執行後、JL 執行前；先前把 `0180:00368B2F` 寫成原版實際分支超出收據。工具規格現行文字已訂正為相同架構的 dosgolem 測試與 Intel 條件所支持，原版動態分支後位置未擷取；本研究檔保留先前紀錄與原 JSON／LOG 雜湊，追加此勘誤，不重新深挖已驗 ADD／旗標／第一個 CMP。

**測試與資料觀測**：Go 1.24.13，映像 ID `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。兩切片各跑 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/cpu386 -count=1`，輸出 SHA-256 `f741b064681cdc3ff79151be22d7af90cd735dbc7c87d2bb1374dedb8fb1a4c4`／`03adc98d6c9e63eeb196207237c71e81a4ccaf8beea6a1efef1b6db17efb6bd2`。八來源／word 八目的、16 邊界值成對輸入，獨立驗借位、有號溢位與低 byte 同位；精確寬度、DS／SS／SIB、地址環繞／別名、非寫回、部分讀取拒絕與既有方向回歸通過。word 寬度正例依 271 替代 270 的原 word 拒絕負例，其餘拒絕不放寬。

加入原版最小架構樣本後，各以固定 EXE 的 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過；輸出 SHA-256 `58df4137851ac567cdc133a724ff14cbc69a38b77662f15b9d1a0856fec26b20`／`7b337b6d123caede6767cf2ca5ff314be56ac075420eecf8228ea73c6e607c37`。實作前的兩側只讀觀測與雜湊見 270／271；沒有用注入狀態替代自然路徑。

**已證實，兩條下一停點**：兩切片均以 `DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，另一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。270 在無事件／受控事件第 6,772,679／6,772,714 步於高位 LE `0x228DCE` 的 word CMP 停下，gzip SHA-256 `6c38fdc4b3ab55211ef9e53cebfa554551f069b27c4beb560263eea3ed0358f5`／`266791cd8a54111ef0053386ad52409df3c3c13ea8674cca0af8ad21fb2a0275`。271 自行越過兩種比較及有限分支，在第 6,822,373／6,822,408 步於 **dosgolem 高位 LE 線性** `0x21CA3D` 的 `66 FF 40 04 A1 A8 42 2A 00 8A 40 0B 24 20 25 FF` 因 word INC 尚未支援而拒絕。EAX=10000h、ECX=1DFh、EDX=0、DS／ES／SS=188h、EFLAGS=282h、DOS 呼叫 6，拒絕後 EIP=0x21CA40；無事件 EBX=3DC050h、受控事件 EBX=3DD050h，布局差異明示，受控回呼 started=1／completed=1。gzip SHA-256 `7fec533d3f45b0d9a8ddeb8d5807a01d895faf98f3c0e97743555a93834397f7`／`8b633c0bb5f40764ce156978a50293daacf5102f62b146ac7290499c15811a67`。

**圖像、回填與界線**：Active=true、Bank=7、StartY=512、BankSets=6、Writes=307200、DisplaySets=1；有效索引 SHA-256 `7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf`、RGB `0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366`。另設 `DOSGOLEM_MOO2_VBE_PNG`，兩切片共四份最終 PNG 均 SHA-256 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622`，與已檢視黑圖相同。269 → 270 → 271 回填，兩組新正例與缺定位／缺舊標記必拒絕、全部既有回填函式、索引／格式／語法／擁有權通過；不宣稱正常玩家畫面完成。

270／271 僅有限 CPU／首個分支 CONFORMED，255 仍 READY；正常玩家畫面、音效、受控亂數與 Go remake 玩法同狀態未完成。IMUL 未定義 ZF、DIV 旗標／例外、SAR 的 AF、DTA 保留區與平台布局差異仍保留，CPU 支援不計入玩法矩陣分母。下一步只核對 word INC 的實際目的值、CF 保存與後續消費，不追 runtime／圖形 helper 內部。

## 2026-10-01：word 記憶體遞增與首次 Simtex 啟動圖像

- 本輪接續工具 `a1a80765b6b1d527e2874057377afd1caab7a920`，主庫 `3d88548eb9deec73c81f72befb0497fbf05351b6`；固定 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 根層原版檔及 553 bytes MOX.SET（SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`）。仍在本機隔離分支，未寫上游工作樹；CPU 支援不計入玩法矩陣分母。
- [規格 272](https://github.com/wicanr2/dosgolem/blob/3a0c6b4cb475acc75588badff6436fd3612800d4/docs/spec/272-cpu386-inc-word-memory.md) 在建檔同次加入工具索引，先 DRAFT；公開 [Intel INC](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/INC.htm) 定義 word /0 加一、更新 OF／SF／ZF／AF／PF、保留 CF。dosgolem 高位 LE 線性 `0x21CA3D` 的 `66 FF 40 04` 在兩條只讀觀測皆為 DS:10004h word=0、EAX=10000h、EFLAGS=282h，未注入原版資料。輸入觀測 gzip SHA-256 `fb9b45a6f3615754d08110549ea15a0f25f962a3dad483a55b11ace702cee509`／`a9d853654074c6ef13cca527375859c785994bcdff7a695f5b8693fe2844548b`。
- **已證實的有限原版樣本**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --inc-word-memory` 實際命中 **CS:EIP** `0180:00350A3D → 0180:00350A41`；16 bytes=`66 FF 40 04 A1 A8 22 3D 00 8A 40 0B 24 20 25 FF`。EAX=FE30h、EBX=508050h、ECX=1DFh、EDX=0、ESI=0、EDI=3A2090h、EBP=3EBBA8h、ESP=3EBB74h、DS／ES／SS=188h、FS=0、GS=20h；DS:FE34h word=0→1、EFLAGS=282h→202h，其餘擷取狀態保持。LOG 2 及後續 EV 實際核對下一 A1：DS:003D22A8 dword=FE30h，**CS:EIP** `0180:00350A46`、EAX=FE30h、flags=202h。兩工具絕對指標／布局不同，只比可重定位的值與定義旗標，未宣稱完整狀態一致。下一 A1 是最小載入觀測，不代表已解出目的欄位的玩法意義。
- JSON／消費 LOG／終端 SHA-256 `227301ea8b9741ceaf2b7fa3cf0aae638db1e4db8dc6f5745d67cd1e5728ee15`／`d063afe8d016757ef4413cce60489e9d646c0305f1e3147ab40c110502696e9b`／`50a84accb42d8c1a9e63f5b1db2bcad5b25d6c6229e47ba8f6a6df68636d1e7c`；完整資料、記憶體及 PNG 留本機，不進 Git。
- 公開契約、兩側初始目的與原版完整旗標／下一載入足夠後，審查 DRAFT→READY 才改 CPU。完整讀／寫成功後才發布旗標、還原 CF；既有 word DEC 的順序不變。16 邊界×CF 兩初值×舊旗標兩組、DS／SS／SIB／負位移／環繞／別名、精確兩 bytes、鄰接與高半部保存、唯讀／界線／兩 byte 受控讀寫失敗、未知前綴／群組拒絕、原版重定位最小樣本與下一 A1 均通過。第二 byte 寫入失敗可能已改第一 byte，已明示並測試，未冒稱 x86 原子例外。
- Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/cpu386 -count=1` 與 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 均通過，CPU／全套 SHA-256 `ed4bf0766d93be14ad0b08dba6abe6d95902783cba3534ef0a51b287bf2d4623`／`164688d8b3a2698f054ed2b00cf5fc3f40cc0b288d30dbb22f26c4cfdf27a3b5`。既有 word DEC／dword INC／DEC 全部保持；`--check-inc-word-memory-spec-backlinks` 的正例及刪舊標記／原定位的拒絕例、全部既有回填函式、Python 語法、Git 差異與擁有權核對通過。
- **dosgolem 自行重生自然收據**：`DOSGOLEM_MOO2_SEPARATE_DOS=1 go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，另一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。無事件第 7,684,074 步、受控事件第 7,684,109 步，均前進 861,701 指令，停於高位 LE 線性 `0x21C2D6` 的 `78 06 2B C2 79 02 EB EC 5E 61 C3 68 24 00 00 00`；錯誤為短 JS 未支援、拒絕後 EIP=21C2D7h，EAX=2、EBX=Eh、ECX=10000h、EDX=1、DS／ES／SS=188h、EFLAGS=202h、DOS 呼叫 6。受控回呼 started=1／completed=1。gzip SHA-256 `8c2ffdb964ca9c547416f16b026911fcd559f516c176dccf0d2a5fbef982032c`／`c3ef39afef2949572ccd154f872376f456a6f2b8df33824e0b470edea844ee38`。
- **已實際檢視圖像**：Active=true、Bank=2、StartY=0、BankSets=17、Writes=921600、DisplaySets=2；索引 SHA-256 `7de0420d874f91251343b50c1d39944115b806e92f80aa0f77f3687ef8ec3e48`、RGB `4c6ae38e1537960ec75f35b60574a1a11ad54b06737ef2ad8c5e16fbd26a8244`，兩份 `workplace/moo2-vbe-272-*.png` SHA-256 皆 `dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db`。640×480 白底黑色 Simtex 啟動標誌，首次非黑圖；尚未進入主選單、玩家操作或最終畫面對拍。這是 dosgolem 自行渲染，未換用 DOSBox 圖片。
- 規格 272 僅在 word INC／旗標／下一 A1 範圍 CONFORMED，255 仍 READY；主選單、音效、受控亂數與 Go remake 玩法同狀態未完成。部分匯流排寫入、IMUL 未定義 ZF、DIV 旗標／例外、SAR 的 AF、DTA 保留區及平台布局限制仍保留。下一步只核對公開 Jcc 契約、短 JS 的 SF 與第一個分支；候選 **DOSBox-X CS:EIP** `0180:003502D6` 尚是地址假說，不追 helper 內部、不重新開啟已完成 CMP／INC。

## 2026-10-01：短 JS／JNS 與 20.1M 原版自然觀測

- 開工主庫 `07a8554e2da7ef4915cd94dfd26fafe5b59d9734`、隔離工具 `3a0c6b4cb475acc75588badff6436fd3612800d4`，均乾淨並與遠端一致。固定 1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 根層原檔及 MOX.SET（553 bytes、SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`），不改主庫玩法。
- [規格 273](https://github.com/wicanr2/dosgolem/blob/5307099b55cd239f9ea5badd2ca42a245d951404/docs/spec/273-cpu386-short-sign-branches.md) 建檔同次入索引，先 DRAFT，依 [Intel Jcc](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/Jcc.htm) 核對 JS 的 SF=1／JNS 的 SF=0、有號 rel8 與旗標保持。以原版有限樣本審查 READY 後才實作，不解 helper 或平台時鐘內部。
- **已證實的有限原版分支**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --short-sign-branches`。**CS:EIP** `0180:003502D6` 的 `78 06 2B C2 79 02 EB EC 5E 61 C3 68 24 00 00 00`，EAX=1、EBX=Eh、ECX=0、EDX=1、ESI=470h、EDI=3A2090h、EBP=3EBC06h、ESP=3EBB9Ch、DS／ES／SS=188h、FS=0、GS=20h、EFLAGS=202h；LOG 2＋後續 EV 實際不取至 `0180:003502D8`。下一 JNS 在 `0180:003502DA`、bytes=`79 02 EB EC 5E 61 C3 68`，EAX=0、EFLAGS=246h，其餘同前；LOG 2＋EV 實際取至 `0180:003502DE`。兩個分支皆保持完整擷取狀態；未注入旗標，也未冒稱原版該處 SF=1 路徑曾被觀測。
- JSON／JS LOG／JNS LOG／終端 SHA-256 `98e6846cd4ac9647c4c707b03bd7c171c0858f156a1abbceebef5dff326ebf15`／`e3321b3f57f71081c80da868cfe19f6e82bb772ad8e60ba0ce678b1631a3e197`／`471a7afdf0c463e5c1eb72be1e37d64fc4443d2a3361a5bb3ee00a1123547970`／`fe479f44d4dc56c87b5c28cedc02c376972c419c5d0226a6083c422fa4710c29`；完整原版資料與 PNG 只留本機。
- CPU 兩 opcode×六算術旗標全 64 組×全 256 rel8、六 EIP 起點環繞、精確兩 bytes 取指／截短／未知前綴拒絕／近分支保持及原版 JS→SUB→JNS 最小樣本均通過。Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，`GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/cpu386 -count=1` 的 SHA-256 `75c97fa0a346ebd8f40805126835938477e3454df7d0bd50bf2a29dc9dfa7237`。`--check-short-sign-branches-spec-backlinks` 正例、刪原定位／舊標記的拒絕例、全部既有回填函式、語法／差異／擁有權均通過。
- **驗證腳本訂正**：首輪 Go 全套成功，但兩條探針到既有 8M 上限；上限不輸出 PNG 造成外層 sha256sum 退出 1，並非 CPU／產品失敗。收據另存 `273-first-limit`，成因與雜湊保留於規格 273。診斷步數 `DOSGOLEM_MOO2_MAX_STEPS` 限制 1–50M、預設仍 8M；0／負數／超界／非整數／溢位均在讀原檔前 exit 2 拒絕。上限與 CPU 拒絕共用 VBE 擷取，不改輸入或虛擬時間。修正後相同映像與 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 乾淨重跑全套通過，SHA-256 `abf7ee78fdd09afe67ddba3c0a05f4a9a0e0c5fad5301e507b457f407b1f2815`。
- 16M 兩條收據自生成，JS 首次步數 7,684,074／7,684,109，EAX=2／ECX=10000h／flags=202h 至高位 LE `0x21C2D8`；JNS 首次 7,684,076／7,684,111，EAX=1／flags=202h 至 `0x21C2DE`，全狀態保持。這與 DOSBox-X 的 EAX=1→0／flags=202h→246h 是不同初始狀態，只確認 SF=0 的有限控制流，未藏數值差異。16M 到上限，無事件高位 LE `0x228D24`、事件 `0x21A9B8`，尚在動畫；gzip SHA-256 `98cb5eda2583cd5637c8dbffd7de575445cf00c960499b5c020b622ab575230c`／`84dfffe57f2a7f02caa90bc7c25f4b51722f0bee3727a03647fb818957d16763`。兩 PNG SHA-256 均 `98cb13ebdeda137d02d88dc49d15ccbd833cc12b82fb8bbf7fd09166dd56012b`，已實際檢視為黑底展開中的 MicroProse 啟動動畫，並非主選單或穩定完成畫面。
- **下一實際停點**：保持輸入與虛擬時間，只將上限明示為 50M：`DOSGOLEM_MOO2_MAX_STEPS=50000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_VBE_PNG=<本機輸出> go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，事件一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`；未重跑已通過的全套。無事件第 20,100,561 步、事件第 20,100,596 步，原阻塞後前進 12,416,487 指令，停在 **高位 LE 線性** `0x239A42` 的 `66 31 FF 8E C7 CD 2F 66 89 3D 22 BC 2A 00 66 C7`，word XOR 未支援、拒絕後 EIP=239A44h；EAX=F1684h、EBX=5、ECX=239F20h、EDX=1、DS／ES／SS=188h、EFLAGS=246h、DOS 呼叫 6，事件回呼 started=1／completed=1。gzip SHA-256 `5994ce355565faaefe780bac97836cce37dc9ad98f4564d7eb192f7752734a09`／`1dda8342da159dd5720e6d9f8eb81b5fdf02ade47250ff9a8b728d6f907bb48a`。
- 此時 Active=true、Bank=2、StartY=0、BankSets=441、Writes=4046164、DisplaySets=6；索引／RGB／PNG 雜湊與先前已檢視黑圖逐位元相同，精確值見規格 273。途中 Simtex／MicroProse 圖像成立，但當前仍黑圖，尚未主選單、玩家操作、音效、受控亂數或 Go remake 玩法同狀態。
- 273 僅有限 CPU／分支 CONFORMED，255 仍 READY；初始時間／布局、分支例外、部分寫入、IMUL／DIV／SAR、DTA 保留區等限制仍明示。下一步只依公開 word XOR 契約核對 DI 初值／高半部、旗標與下一 MOV ES；候選 **DOSBox-X CS:EIP** `0180:0036DA42` 尚是假說，不追 INT 2F／helper 內部，不重新開啟已完成 JS／JNS。

## 2026-10-01：word XOR 與下一段載入的有限驗證

- 開工主庫 `9faa19e1676ccefddbc672b9410102a6f54b9d2d`、工具 `5307099b55cd239f9ea5badd2ca42a245d951404` 均乾淨並與遠端一致。原始 ZIP／官方 patch 雜湊沿用本研究入口；固定 1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 根層原檔及 MOX.SET（553 bytes、SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`），不改主庫玩法。
- [規格 274](https://github.com/wicanr2/dosgolem/blob/45b46688666ff21d38402870d8f585b563dc2d58/docs/spec/274-cpu386-xor-word-register.md) 建檔同次入索引；依 [Intel XOR](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/XOR.htm) 的低 word、CF／OF 與 SF／ZF／PF 契約及原版樣本審查 READY 後才實作。AF 未定義，沿用清除的模型並分開驗，不當成硬體全旗標對拍。
- **已證實的有限原版樣本**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --xor-word-register`，未注入暫存器／旗標。**CS:EIP** `0180:0036DA42 → 0036DA45` 的 `66 31 FF 8E C7 CD 2F 66 89 3D 22 9C 3D 00 66 C7`，EDI=38EC2Bh→380000h、flags=246h，高半部及其他完整擷取狀態保持；LOG 2＋EV 實際確認下一 MOV ES 至 `0180:0036DA47`、ES=0。JSON／LOG／終端 SHA-256 `9662eb7f0ee609340a17aa067b8f5b30b3d8fd796fff4970f9eed515e38bd16c`／`866447078bda5e5a7f5b47e045099ed2d3216f2bb6ea6787a880ece43452eba3`／`94fa43208094452683a6bb10df48de6f6cd31b7579d85b93f1c4a1ef07e4bfd0`。完整輸入欄位、只讀 dosgolem 起點 EDI=260C2Bh 及布局差異見規格；未宣稱完整同狀態。
- CPU 八來源×八目的×16 邊界組合×兩舊旗標組，逐 bit 不同、有號結果／低 byte 同位獨立期望、高半部／別名／完整狀態保持、截短／前綴／記憶體拒絕、byte／dword 回歸與 XOR→MOV ES 原版最小樣本通過。首輪只因舊 dword 測試期望把 SF 誤填為 0 而失敗；B9F90003h 應為 flags=286h，既有 CPU 正確，沒有改 production。失敗收據 SHA-256 `28410c1b77d8b114077dc7d886870ef4a8a69cc23b33d74975a8630d94ed499b` 保留。依 [Intel Appendix C](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/appc.htm) 修正測試後，以同一命令乾淨重跑通過。
- Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/cpu386 -count=1` 與 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 均通過，SHA-256 `c7376ff82feaa8997b933170e88244bf43fafc7870704148b0ea1aa77a549f9b`／`4d94cf6a4c8d621baf6172315ab9c67cab338c28a7428f55b9617d76fae5ef8e`。回填 273→274，正例／缺定位／缺舊標記必拒絕、全部既有回填函式、語法／索引／差異／新增繁體文字／擁有權通過。
- **目前自然停點**：`DOSGOLEM_MOO2_MAX_STEPS=50000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_VBE_PNG=<本機輸出> go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`，事件一路另加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。兩條自生越過 XOR／MOV ES，無事件第 20,100,563 步、事件第 20,100,598 步在 **高位 LE 線性** `0x239A47` 的 `CD 2F 66 89 3D 22 BC 2A 00 66 C7 05 24 BC 2A 00` 拒絕，EIP=239A49h；EAX=F1684h、EBX=5、ECX=239F20h、EDX=1、DS／SS=188h、ES=0、EFLAGS=246h、DOS 呼叫 6，事件回呼 started=1／completed=1。gzip SHA-256 `5c0c42d06340369d93f6abb2240106e4dd7dfb63192ba4fca7b38ec4d7dbb208`／`21f0880d38e18710824bb46166692500fec7775c0b51b620520b5bbccd66ca16`。
- VBE Active=true、Bank=2、StartY=0、BankSets=441、Writes=4046164、DisplaySets=6，索引／RGB／兩 PNG 逐位元同前次黑圖，精確雜湊見規格；不重做無必要目視驗收，不把途中 Simtex／MicroProse 動畫說成主選單。274 限定 CONFORMED、255 仍 READY；正常玩家操作、音效、受控亂數與 Go remake 玩法同狀態未完成。下一步只查公開 INT 2Fh/AX=1684h 邊界及 caller 參數／返回／消費，不解平台／runtime／圖形 helper。

## 2026-10-01：未安裝 VTD 的空入口與自然返回

- 開工主庫 `9d539282180c10f5485fee01aa4d0b7733aceb9f`、隔離工具 `45b46688666ff21d38402870d8f585b563dc2d58`，均乾淨並與遠端一致。固定 1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 根層原檔及 MOX.SET（553 bytes、SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`），ZIP／patch 雜湊沿用本研究入口，不改 Go remake 玩法。
- [規格 275](https://github.com/wicanr2/dosgolem/blob/3c4bace68829bd16377aeaf30e22631c6ce3cefb/docs/spec/275-moo2-protected-vtd-entry-query.md) 同次建檔入索引；依 [RBIL 61 裝置 API 與表 02642](https://fd.lod.bz/rbil/interrup/windows/2f1684.html) 的 1684h／ID=5／ES:DI=0:0、未安裝的空入口及 [DOSBox-X 多工入口](https://raw.githubusercontent.com/joncampbell123/dosbox-x/master/src/dos/dos_misc.cpp) 作平台前提。原版完整返回／保存／讀取足夠後 DRAFT→READY 才實作，沒有深入 Windows／VTD 時鐘內部。
- **已證實的有限原版樣本**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --vtd-entry`，未注入暫存器／旗標。**CS:EIP** `0180:0036DA47 → 0036DA49`、前 16 bytes=`CD 2F 66 89 3D 22 9C 3D 00 66 C7 05 24 9C 3D 00`；EAX=F1684h、BX=5、EDI=380000h、ES=0、flags=246h，完整一般暫存器／段保持。入口 record **DS offset** `0188:003D9C22` 六 bytes 查詢前後皆 0。LOG 8＋EV 實際保存 DI、零高 word、ES，恢復 ES=188h，再載入 ECX=0／DX=0 並 TEST，停於 `0180:0036DA70`、ESP=3EBAD4h、flags=246h；下一條件分支尚未實際執行。完整輸入及 40 bytes、debugger 把 word 誤顯示 dword 的邊界見規格。
- JSON／LOG／終端 SHA-256 `e7068aa1b1d301609c58f2211e37cbbc66bb6b2588a8482f877033853bfc3e90`／`487745aa92e31e78218f9b990f2a12be5bccb6db00bdca3b8f173aa620ae1789`／`201dd5795433bb353b7dcd3a2f0bcf0d69638eb40feb4e96d6f46071ff77760c`。首輪 LOG 2 只實際執行第一筆保存，未把待執行的最後列當成完整消費；原 JSON／LOG／終端保留為 `vtd-entry-first-*`，精確雜湊及同映像擴大觀測原因在規格。
- 平台實作只限 MOO2、低 AX=1684h／BX=5／ES:DI=0:0；完整 R／段／EIP／EFLAGS／記憶體／DOS 計數保持。其他設定／裝置／功能／非零 ES 或 DI 拒絕。四高半部×六算術旗標全 64 組×連續呼叫與原版兩筆 word 保存、相鄰哨兵保持均通過。首輪最小 caller 缺可寫 DS 描述子，CPU 正確拒絕；只修測試環境，production 不變，失敗收據 SHA-256 `571785b82c6b47cd7761709a2124763500625be2f17baeae8bb6afd44226d233` 保留。
- Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，相同 `GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/machine -run 'TestMOO2(AbsentVTDEntry|WindowsVersionAbsent)' -count=1` 乾淨重跑通過；固定原檔 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 全套通過，SHA-256 `5b8e964cc09f0722c9a01cad0529acb7b11ff9982cce4c24402892467602b94a`／`e47cea1afdb99120cff4887a484aa2138901fc60cb98b6126443168893bcba03`。274→275 定位／舊停點護欄正負例、全部既有回填函式、Python 語法、索引、繁體文字／擁有權／差異通過。
- **自然收據**：`DOSGOLEM_MOO2_MAX_STEPS=50000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_VBE_PNG=<本機輸出> go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game`；事件一路加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。無事件第 20,100,563→20,100,564 步、事件第 20,100,598→20,100,599 步，由 **高位 LE** `0x239A47 → 0x239A49`，完整只讀欄位保持，EDI=260000h、ES=0、flags=246h；與 DOSBox-X 的高半部／布局不同，只比空 ES:DI 與保持契約。
- 兩條再前進 835 指令，無事件第 20,101,398 步／事件第 20,101,433 步於 **高位 LE 線性** `0x239AE8` 的 `E6 43 EB 00 88 D8 E6 40 EB 00 88 F8 E6 40 A1 48` 拒絕，error=OUT port 43 未處理、拒絕後 EIP=239AEAh；EAX=1734h、EBX／EDX=174Eh、ECX=0、DS／ES／SS=188h、flags=246h、DOS 呼叫 6，事件回呼 started=1／completed=1。gzip SHA-256 `53e73b6fd668062d43eaa21292f1ea4b248c565ad405d4dc4b79fdbb41c526f4`／`0a116f56852f469c095ac08c7ca9ba78a77cb5de8bbf3421c55fe8f66631c554`。自然 caller 已越過，但 record 未另做逐位元擷取，不能把步進成功當成完整 consumer 同狀態驗收。
- VBE／索引／RGB／兩 PNG 與先前已檢視黑圖逐位元相同，精確雜湊見規格，不重做目視驗收。275 限定平台查詢 CONFORMED、255 READY；尚未主選單／正常玩家輸入、音效、受控亂數或 Go remake 玩法同狀態。下一步只依公開硬體契約核對 43h／34h 與共享平台埠，不逆向 timer driver、ISR／busy-wait 或逐週期時鐘。原版完整資料與收據只留本機。

## 2026-10-01：PIT 模式 2 與共享週期的有限接通

- 開工主庫 `c283b049c6a7b862d9c53658cbe9df3b4ceaf777`、工具 `3c4bace68829bd16377aeaf30e22631c6ce3cefb`，均乾淨。固定原版 ZIP／官方 patch 雜湊沿用本入口；官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 根層原檔、MOX.SET 553 bytes／SHA-256 `bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80`。
- [規格 276](https://github.com/wicanr2/dosgolem/blob/7b288510f8ff578079e73be0435784a9a2e7f720/docs/spec/276-pit0-mode2-shared-clock.md) 同次建檔入索引，依 [Intel 8254，231164-005](https://www.cs.cmu.edu/~410/doc/8254.pdf) 控制字／模式 2／圖 22 及原版有限寫入審查 DRAFT→READY 才實作。模式 2／3 的最小除數為 2，0 編碼 65536；除數 1 的拒絕是公開規格邊界，未對原版注入非法值。原有模式 3 合法收據保留，186 與 275 回填。
- **已證實的有限原版樣本**：DOSBox-X 2026.07.02 SDL2 heavy debugger、映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --pit-mode2`；**CS:EIP** `0180:0036DAE8 → 0036DAF6`，原始 `E6 43 EB 00 88 D8 E6 40 EB 00 88 F8 E6 40`，三筆 OUT=43h←34h、40h←4Eh／17h，除數 5966。EAX=1734h→1717h，EBX／EDX=174Eh、ECX=0、ESI=C8h、EDI／EBP=8、ESP=3EBAD4h、DS／ES／SS=188h、flags=246h。下一 A1 讀 **DS offset** 39F148h 得 EAX=0，到 `0180:0036DAFB`；其他完整擷取欄位保持。沒有注入 CPU／記憶體，不解 driver 或 ISR。
- JSON／OUT LOG／第一讀取 LOG／終端 SHA-256 `c522c19a88f1ea333ec2491befcb3e23b1dafab284e98b39e7728bc92f93dd74`／`a135930b7be27ea0b394ace213d093487cf463fa298b7b3a4e51232b0bc9ab70`／`cdbdf8c52ad6818a174d68dc0bd0bf0c04c4c175bc33486450a58446fab277ab`／`609ca9fe3eb4e9b6d5dd394c20061670b1d85399d9a5a0c23e219c4fdedee318`。LOG 最後列未執行的邊界保留，完整原版資料與收據只留本機。
- Go 1.24.13、映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`GOMAXPROCS=2 go test -p 2 -buildvcs=false ./internal/machine -run 'Test(PIT0|BIOSClock)' -count=1` 與固定 EXE 的 `go test -p 2 -buildvcs=false ./... -count=1` 均通過，SHA-256 `f096373c82ba90e9bdc635964167dbe3d7b657c7cab4da197131350bfcf18512`／`1604659fa9545f812655752abb417fc2878a79194720dfccc7c9a54b8e923e7d`。驗收全除數編碼、半筆／重設／拒絕、實模式及保護模式共享埠、原版最小設定與讀取、近似週期／遮罩／IF／客製向量拒絕。兩回填標記或原始定位刪除必拒絕，全部既有回填函式通過。
- **驗證腳本訂正**：兩次自然探針非零結束，保存的一次為 `signal: killed`，gzip SHA-256 `dad7ce3ec4dfe714a1f8d5c77753ad3e90b0ff2555a3dfe95bbb4720865c92c2`。原尾端 slice 會縮容量並持續擴容，100,000 步實驗：舊 len=99977／cap=122880，新 len=32／cap=32 且保留最後 32 筆順序，SHA-256 `ba1b2922445398b59c65c500e05e8a09f4cc02b3f13f7ae9a54f674e4c46691e`。只改探針觀測緩衝；與 killed 的關聯是強推論，首輪未保留 cgroup 計數，不冒稱已證實 OOM。修正後同一 2 GiB 容器兩條重跑完成，峰值 775704576 bytes、oom／oom_kill=0。
- **dosgolem 自行重生自然收據**：明示 50M、分離 DOS 與完整原版資料；無事件第 20,101,398／20,101,405 步在 **高位 LE** `0x239AE8／0x239AF6`，事件一路各晚 35 指令。Mode=3→2、Reload=5966、完整世代 4→5，低位不發布；完整 CPU 可比較欄位見規格。兩條到 50M 上限，無事件 `0x239B03`、事件 `0x239B09`，末尾交替 `CMP [00271148h]／JE` 等待；EAX／ECX=0、EBX／EDX=174Eh、DS／ES／SS=188h、flags=246h，沒有新的未支援指令。gzip SHA-256 `1a740e3cb2e56eec53dff999928bc72b969a3d036f83c2a031589a07090e25f7`／`57ce3b43683a2b1d5dacb301f7f6a9f766bae9acec5eb08bd264a12b203bc29e`。
- 額外只讀 20,120,000 步收據 SHA-256 `dc92747ae37dba5d2124cc8f808e2eff0027f87525ec4873b926bf3f813d0862`：PIT=Mode 2／Reload 5966／Generation 5，共享 clock Micros=20121042／Deliveries=1538，較設定前多三次；DPMI 實模式 08h／1Ch 均 0000:0000，絕對 IVT 也零，**高位 LE DS:offset** `0188:00271148` 四 bytes 仍零。預設時鐘推進已證實；該等待來源的完整語意與正式向量派送仍未知。既有 DOS `s.dosVectors` 與實模式向量分開保存，尚未證實遊戲 08h 的實際設定，不猜補來源。
- VBE／三 PNG 同已檢視黑圖，PNG SHA-256 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622`，索引／RGB 精確雜湊見規格。276 只限設定／共享週期近似 CONFORMED，255 READY。正常玩家路徑、音效、受控亂數及 Go remake 玩法同狀態未完成；下一步只核對 DOS 向量服務參數與保存／派送接線，不反組譯 driver／ISR／busy-wait，不追逐週期時鐘。

## 2026-10-01：保護模式 DOS 向量與受限 IRQ0 接線

- 基線主庫 `fdec7c3ed3da0849739f137a052358915387706a`、隔離工具 `7b288510f8ff578079e73be0435784a9a2e7f720`。固定官方 1.31 `ORION2.EXE` SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，417 原檔與 MOX.SET 沿本入口，不變更遊戲輸入。[規格 277](https://github.com/wicanr2/dosgolem/blob/777472b02a561740e11558ecb9852b7579541454/docs/spec/277-moo2-dos4gw-protected-irq0.md) 同次建檔入索引，有限樣本與公開 DOS/4GW 契約足夠後 DRAFT→READY 才實作；主庫玩法 RE 閘門不變。
- **已證實的只讀向量樣本**：dosgolem 高位 LE `0x24501B` 的 AH3508、`0x245048` 的 AH2508 安裝 `0008:00244D9A`；既有保存表先前沒被時鐘消費。20.12M 向量探針 SHA-256 `f9fbdbd5358dbdf6683cc966607216a3850edff10ed86858dd3350344874d39b`。按 [Open Watcom DOS/4GW 25h／35h 文件](https://open-watcom.github.io/open-watcom-1.9/pguide.html) 的非空預設入口、私有核心堆疊及實模式向上轉送建立有限平台契約，不研究核心內部。
- **已證實的有限原版輔助樣本**：DOSBox-X 2026.07.02 SDL2 重型除錯器／映像 `sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582`，`startup_probe_131.py --dos-timer-vector`。CS:EIP `0180:0037901B` 的 AH3508 返回 `0080:00000D49`，`0180:00379048` 的 AH2508 安裝 `0180:00378D9A`；自然 IRQ0 用私有 SS:ESP `00D0:00006838`，框架 flags=46h。等待來源 DS offset `39F148h` 從 0 變 1，自然到 `0180:0036DB0B`；未注入時計、暫存器或記憶體。原始 bytes／完整欄位見規格；JSON SHA-256 `3d30e7a7498d7b9207bbad2b37257f6b9dc5ab2b082681db09464bec539e4727`，終端 `98bffa98c01fe466e2c1847f111c2abd043e7d5bb3a6dc6f01d0e22db2614315`，完整原版收據只留本機。
- **平台實作與驗證**：兩 CPU 模式共享 PIT／有效 DOS08h、私有 12-byte 最外層 CF 框架、完整狀態／FPU 保存與恢復、污染／未知 opcode／有界未返回拒絕；PIC 遮罩、IF、EOI／IRQ7 優先與滑鼠回呼共存已驗。跨 CPU 與被污染預設入口也拒絕。合成布局／一微秒每指令、IRQ0 合併及完整核心鏈限制明示近似，不把合成處理器成功當原版成功。
- Go 1.24.13／映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`GOMAXPROCS=2 go test ./internal/machine -run 'Test(ProtectedIRQ0|LEMouseCallback|BIOSClock|PIT0)' -count=1 -v`，與 `DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE GOMAXPROCS=2 go test -p 2 -buildvcs=false ./... -count=1` 均通過；平台／全套 SHA-256 `5f87e0db5676d44e348a359c64a07da0972cac7a887340f308544ba9706d867b`／`f63d2055a0a486999163a33fafb3c8a2b65bea964e674eb39f70542b23d68892`。276 回填護欄正負例、全部舊回填函式、Python 語法、索引與擁有權核對通過。
- **dosgolem 自行重生的目前事實**：明示 50M、分離 DOS 與完整原版資料，另一排程加 `DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1`。兩條件第 1,160,098 步進第一次 IRQ0，started=1／completed=0，**高位 LE 線性** `0x244D9A` 的 `2E 83 3D FE 19 27 00 00 0F 87 11 01 00 00 60 1E` 拒絕，原因為 CS 記憶體比較缺件；失敗後外層恢復 `0x234341`，Micros=1160100、Deliveries=21、InService=true、模式 3／Reload=19887。兩條件均在事件注入前停止，requested=true／injected=false；不是兩條玩家分支已驗收。gzip SHA-256 `30a8bc2eda975584afbac7c5d3398b389f98c4a6a2826f5040c840c74ff92148`／`6562fcd7ea5a08935c44a88b7820429298aff922397898f9d75c64f6dd59cf82`。當前 PNG 仍與已檢視黑圖一致，VBE Writes=0，不再把舊未派送的 50M 迴圈當目前前沿。
- **驗證腳本勘誤**：保留錯誤載入差值及單步 INT 只進核心的失敗終端後，以同映像／自然流程修正候選與返回斷點成功。上一輪 276 的字串替換誤改 CMP 消費端護欄，已按程式區間訂正，使用既有實際 JSON 與 AST 驗證 CMP／PIT 正例及錯位址拒絕；PIT 實際原版收據仍在正確位置。新時鐘診斷去掉每次配置不同的主機指標，實際自然重跑驗證；原版資料不變。
- **狀態／下一步**：277 仍 READY，原版返回與等待變化沒有在 dosgolem 閉合；255 READY、256–276 限定 CONFORMED 不外推。下一步只按公開 CPU 契約補此標準 CS 記憶體比較後自然重跑；不得深入 ISR、硬體 timer driver／busy-wait 或猜寫等待值。主選單、正常玩家輸入、音效、受控亂數及 Go remake 玩法對拍尚未完成。

## 2026-10-01：CS 記憶體比較與 word 載入 ES，五次原版 IRQ0 返回

- 工具基線 `777472b02a561740e11558ecb9852b7579541454`，本輪先完成 [278-cpu386-cs-memory-cmp-imm8.md](https://github.com/wicanr2/dosgolem/blob/fabbed1a8bfd1c010b66009bdffe4a6812eb9924/docs/spec/278-cpu386-cs-memory-cmp-imm8.md)，提交 `fabbed1a8bfd1c010b66009bdffe4a6812eb9924`；再完成 [279-cpu386-cs-absolute-es-load.md](https://github.com/wicanr2/dosgolem/blob/0e6a69b0ca1971bc9235fb9727bc26d6e57eb2ed/docs/spec/279-cpu386-cs-absolute-es-load.md)，現行提交 `0e6a69b0ca1971bc9235fb9727bc26d6e57eb2ed`。均先 DRAFT＋索引，有限原版與公開 CPU 契約審查後 READY 才實作，工程／自然驗證後限定 CONFORMED；主庫玩法 RE 閘門不變。
- **已證實，原版有限輔助樣本**：固定官方 1.31 EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`、417 原檔及固定 MOX.SET 不變。DOSBox-X 2026.07.02 SDL2 重型除錯器／既有映像，`startup_probe_131.py --cs-memory-cmp`：CS:EIP `0180:00378D9A` 的 CMP 讀 CS:0039F9FE 四個零，下一 EIP=378DA2，資料、所有通用暫存器／段及 flags=46h 保持。JSON SHA-256 `245fb2c1eedf83260c2f748ab66a67715a10867f73f92374dbb68ba2057d56d0`，單指令 LOG `8353f1908d8d52d176b752db49fd0963481a60cd1ed9732f2171e6c87ccffd34`。
- `--cs-word-es-load` 的原版 `0180:00378DBA`，bytes `66 2E 8E 05 06 FA 39 00`，CS:0039FA06 原始 `88 01 FF FF` 只載入 word=0188h，下一 EIP=378DC2；ES 原本就是 188h，完整其他狀態保持、flags=46h。CPU 測試另外確認 ES 初值不同時目的確實改變，不能把原版無值變化冒稱為涵蓋所有初態。乾淨 JSON SHA-256 `ba65afbaa06cfb55cf22b867fb8fb3ec0980d5dfb10a8e1faf0a1884d0886816`、LOG `ea8e7f652fc1be6aa1c04a1aef2406a1117496e8dcefb9d07ce030d296b7203a`、終端 `d4f037c00d4778f846f33f1288971d4ba0a8e3b4bba24e16e96e7aca0e485b50`。只觀察兩條標準 CPU 指令，其餘 ISR 黑箱執行；輔助原版當次是 PIT 模式 2，現行 dosgolem 首次 IRQ 是模式 3，不宣稱整段同狀態。
- **已證實，工程回歸**：Go 1.24.13／既有映像 `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。兩批 `GOMAXPROCS=2 go test -buildvcs=false ./internal/cpu386 -count=1` 及固定 EXE 的 `go test -p 2 -buildvcs=false ./... -count=1` 通過；最新 CPU／全套 SHA-256 `18f9e249b85400a59bb1e55526c443d4fb0eec029d9f14df33b96294b1d68e84`／`95a06d5dfbbbe65897ad443fc05910667b54aa54753aca15ce886683d037dfaf`。CMP 的完整立即值／獨立六旗標、全部記憶體定址、CS／DS／SS 分離與唯讀／越界；ES 的兩 bytes／前綴／selector 舊模型／未知形狀與完整保持皆有測試。限定 selector 模型不是完整 x86 權限／例外精確模擬。
- **已證實，dosgolem 自行重生的目前前沿**：兩個自然排程按上一節 50M 上限與分離 DOS 命令重跑，輸出名稱 279；原版成功完成五次 IRQ0 返回，第六次在外層第 1,231,392 步、高位 LE 線性 `0x244E9C` 的 `CB` 遠返回拒絕，started=6／completed=5，完整外層恢復 `0x246596`。PIT 模式 3／Reload=14916／Generation=3，Micros=1233786、Deliveries=26、InService=false，DS:00271148 仍零。gzip SHA-256 `f32bf14adca2a5053edcfe93854fe7e4919d57c2f0823d276a4da7e3dcb9ee3a`／`19163a9a44954cfa04054eba31f3e69682ac21a6143ee3073516f3075e702163`；事件條件 requested=true／injected=false，仍不能當兩條實際玩家分支通過。VBE Writes=0，PNG 與已檢視黑圖 SHA-256 `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` 相同。
- **驗證命令邊界**：279 首輪完整樣本／等待退出已得，但外層 240 秒最後返回 124；保留逾時輸出，沒有推定精確收尾原因或歸成 CPU 故障。只改外層 300 秒，同映像、輸入與命令乾淨重跑退出 0。PDF 抓圖 Cache miss 改用該 PDF 文字／附錄，發現 MOV 表格 8D 誤植，以附錄 8E 及原始 bytes 校準。索引路徑猜錯與容器缺 rg 皆為讀取問題，改實際索引與 pathlib 後繼續，未退回主機分析或另建映像。
- **回填與下一步**：277／278 的舊停點與 256 的範圍連回後續規格；原始定位缺失、兩份舊標記刪除皆拒絕，全部舊回填／AST 擷取護欄、Python 語法、索引、繁體與擁有權通過。277 保持 READY，理由現為完整 IRQ／模式 2 等待未閉合，不能繼續寫「原版永遠無法返回」。下一步只按公開 CPU RET 契約建立 CB 有限輸入／返回，再自然重跑；不解 ISR／driver／busy-wait 或猜寫等待值。255 READY，主選單／正常操作／音效／受控亂數及 remake 玩法同狀態未完成。

## 2026-10-02：接通 CB 遠返回，停在預設核心鏈護欄

- 開工主庫 75bb479608478aa7bedf43ca98a956384091d130、工具 0e6a69b0ca1971bc9235fb9727bc26d6e57eb2ed，兩庫乾淨。已載入規格閘門／平台規格優先、文件職責與結論回填路由，沿用 dosgolem／逆向重製入口；主庫玩法 RE 閘門不變。[規格 280](https://github.com/wicanr2/dosgolem/blob/3e260dcf2215228d840b98242c1226ab7f902456/docs/spec/280-cpu386-far-ret32.md) 同次建檔入索引，有限原版與公開契約審查後 DRAFT→READY，才實作裸 CB、32 位、同 RPL／VM=0 的已知平坦 code selector 返回。工具提交 3e260dcf2215228d840b98242c1226ab7f902456 已推送 github/codex/moo2-parity-20260930；未推本機來源 origin。
- **已證實，原版有限輔助樣本**：固定官方 1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、417 原檔及固定 MOX.SET 沿前節。DOSBox-X 2026.07.02 SDL2 重型除錯器／既有映像，startup_probe_131.py --far-ret，300 秒容器退出 0。**DOSBox-X CS:EIP** 0180:00378E9C 的 CB，SS:ESP=00D0:00006830，框架 49 0D 00 00 80 00 00 00；一條返回後 CS:EIP=0080:00000D49、ESP=6838h，其餘通用暫存器／段／flags=246h 與框架保持。只觀察這條標準返回，其他 ISR 黑箱自然執行，不搬入原版核心布局。
- **公開契約與證據限制**：[Intel 80386 RET 契約](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/RET.htm) 的同權限 32 位返回，只需讀 EIP32／CS16 共 6 bytes，跳過 selector slot 高 word，消費 8 bytes；審查在實作前訂正 DRAFT 早期全 8-byte 讀取候選。ESP 溢位拒絕、已知平坦 code 描述子及相同 RPL 是工具明示限制，不聲稱完整 x86 code／DPL／present／精確例外。原版核心入口只放可讀佔位 byte 做純 CPU 有限重播；兩個執行器完整初態／地址不同，不宣稱整段 ISR 同狀態。
- 原版 JSON／單指令 LOG／終端 SHA-256：a494d64654cec4198e065c0eac1e9981caf09d1002fbf01a6c18ea22842a17b0／d76afc149acffdcb5b7f8ece204ed9396939aec8014ff9d9ba9b9885470847e2／28cd8774ebd9000e41d45950e30a7bc978628e59cee16813814873ced69f302f；完整原版資料只留本機工具 workplace，不入公開 Git。
- **已證實，工程回歸**：Go 1.24.13／映像 sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。GOMAXPROCS=2 go test -buildvcs=false ./internal/cpu386 -count=1 及 DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1 全部通過；CPU／全套輸出 SHA-256 ccc7bfe98304b0363076b298b9b29a00a9b710ff39d39a9a9d7841ae489c7e36／00c4cb5dda024faabe5768e0855109cbd1780f16fe2da3c80fc2e443c8a600e4。測試涵蓋唯讀非平坦 SS／DS 混淆、四種 RPL／高 word 丟棄、6-byte 必要讀取／逐 byte 失敗、完整 32 位地址與 ESP 邊界、未知／null／非平坦／跨 RPL／VM／前綴拒絕、段回呼不得繞過、完整 FPU／其他狀態／記憶體保持。近返回、CA、權限切換與平台鏈不因本項放寬。
- **已證實，dosgolem 自行重生**：沿 277 的 50M 上限、分離 DOS／完整原檔自然命令，另條件加 DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1，輸出名稱 280。兩條件外層第 1,231,392 步、第六次 IRQ0 均已越過 **高位 LE 線性** 0x244E9C 的 CB，內層返回 **合成平台 CS:EIP** 0108:00326008、SS:ESP=0110:00000FF4，flags=246h；下一步進既有「IRQ0 預設核心鏈尚未建模」護欄，不是未支援 CB。外層完整恢復 0x246596，started=6／completed=5；PIT 模式 3／Reload=14916／Generation=3，Micros=1233787、Deliveries=26、InService=false，DS:00271148 仍四個零。gzip SHA-256 ee40229c9a12fff73978cf2eba833d690538fdc9a9c832033124540992231af5／d6454c281fe069a9784bb4ddcd46b78c2a7da57bfbe092253565f73596f5b45d。事件 requested=true／injected=false，不算兩條玩家分支驗收。VBE Writes=0、PNG SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 同已檢視黑圖，沒有重做目視。
- **回填與下一步**：280 限定 CPU CONFORMED；279／277 歷史停點已連回後續規格，全部 23 個舊／新回填函式、四項缺證據負例、原版實際 AST 保持護欄正負例與既有 CMP／PIT 護欄通過。首輪負例只刪首次定位而仍有第二處，改全刪後正確拒絕，屬稽核腳本問題；image inspect 缺欄位、猜不存在 read32／舊檔名也僅是讀取問題，沒有新增映像。255／277 READY，完整 IRQ／模式 2 等待尚未閉合；下一步只依公開 DOS/4GW chaining／結束鏈介面建立新受限平台規格，READY 後實作，不逆向 ISR／driver／busy-wait 或猜寫等待值。主選單／正常操作、音效、受控亂數與 Go remake 玩法同狀態尚未完成。


## 2026-10-02：預設 IRQ0 結束鏈與原版模式 2 等待退出

- 基線主庫 879f052bc188cbcd75ceea26820b9587c1d11983、工具 3e260dcf2215228d840b98242c1226ab7f902456。已載入規格閘門／平台優先、文件職責與結論回填路由，沿用逆向重製技能；玩法 RE 閘門不變。[規格 281](https://github.com/wicanr2/dosgolem/blob/9b8d1a07121fab929fc94a7f529449beb0d49742/docs/spec/281-moo2-protected-irq0-end-chain.md) 同次 DRAFT 建檔入索引，有限原版及公開平台審查後 READY 才實作。工具提交 9b8d1a07121fab929fc94a7f529449beb0d49742 已推送 github/codex/moo2-parity-20260930，未推本機 origin。
- **已證實，有限原版輔助邊界**：同固定官方 1.31 EXE、417 原檔與 MOX.SET，DOSBox-X 2026.07.02／映像 sha256:659e8abbf93646f59a1586341769bd4b8f3cd1c707859d7a5de4c56e4672b582；startup_probe_131.py --irq0-end-chain，300 秒容器退出 0。CS:EIP 0180:00378E9C 的 CB 原始 20 bytes：49 0D 00 00 80 00 00 00 16 0F 00 00 80 00 00 00 46 00 00 00，SS:ESP=00D0:6830h；CB 後 0080:00000D49／ESP=6838h。核心黑箱自然執行至框架給出的 0080:00000F16，ESP=6844h，BDA 五 bytes 41 27 02 00 00→42 27 02 00 00，通用暫存器及 DS／ES／FS／GS 保持；SS=d0→a8，框架 flags=46→246h。只觀察邊界，沒有翻譯 ISR／核心。原版 JSON／有限 CB LOG／終端 SHA-256：7959ede54826e0c4244ad9e3d25e8f9e5a49484caee477001cec26d98db386d1／394684f0e7201f309dd4c9bc5c90dd3d7ee4a64ac2ab0ec88bf3e1ae60b566b5／caf902af15b63504c87e82b9114086046cf7d974428026a3e692a3ea34e4b0ae。
- **公開平台近似與反例**：早期「只完成返回、沒有 BIOS tick」候選被上述 BDA+1 推翻，在實作前訂正。[Open Watcom chaining／dummy](https://open-watcom.github.io/open-watcom-1.9/pguide.html)、[DOSBox-X BIOS08h](https://dosbox-x.com/doxygen/html/bios_8cpp_source.html) 及 [CB_IRQ0](https://dosbox-x.com/doxygen/html/callback_8cpp_source.html) 支持受限返回／tick／非指定 EOI。原版 EOI 埠序列未另擷取，核心 SS／IF 改寫不搬入橋接，完整外層保存／恢復及公開 EOI 是 hardware-spec approximation；網站版本不冒稱同本機 binary。沒有逐週期、波形、完整核心或整段同狀態聲明。
- **已證實，工程回歸**：Go 1.24.13／映像 sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；GOMAXPROCS=2 go test -buildvcs=false ./internal/machine -run 'Test(ProtectedIRQ0|LEMouseCallback|BIOSClock|PIT0)' -count=1 -v 及固定 EXE 的 go test -p 2 -buildvcs=false ./... -count=1 通過。平台／全套收據 SHA-256 7a0fa28b388f9e66aa01d5781a3737db395e92b01ac05d2d28f6e1ca17fa7375／e700f8b9eabd39261c6e1eb2519d4d0ba21b80e0c5b930fa9932753d681517b9。兩 CPU 模式、完整外層／FPU、12-byte 污染、客製鏈拒絕、午夜 rollover／EOI 優先及 pending 不重入皆驗。276／277／280 舊停點與索引回填，24 個文件護欄、四項缺證據負例、有限邊界暫存器與 tick 負例及既有 CMP／PIT AST 護欄通過。
- **已證實，dosgolem 自行重生**：沿 277 的 50M／分離 DOS／完整原檔自然命令，第二條件加 DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1；輸出名稱 281。兩條各完成 1,595 次 IRQ0 返回，started=completed=1595、active=false／failed=false，等待來源高位 LE DS:offset 0188:00271148 自行變為 01 00 00 00、退出模式 2 等待，未注入遊戲計時值。第二條件第 1,612,067 步實際注入 x=657／y=189／buttons=0；這是受控工具事件，不能當完整座標／游標消費或真實玩家操作。兩條第 20,634,818 步停高位 LE 0x239B3A，bytes E6 43 EB 00 E4 40 88 C4 E4 40 86 C4 25 FF FF 00，EAX=0、OUT 43h 的控制字 00h 計數鎖存未處理；非 IRQ0 返回失敗。PIT Mode=2／Reload=5966／Generation=5、Micros=21092406、Deliveries=1615、Pending=false／InService=false。
- 自然 gzip SHA-256 13d30875d8eb0f90c126705fdf996d560fe24ba627410d3bd12cad50adee0346／6a2373334cbd427b62d1415ec8eefae8cf7c8943b22492d61a3e0b711eb828b1。VBE Writes=4046164，最終 PNG 同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；完整原版資料／終端／gzip／PNG 只留本機。
- **狀態與下一步**：281 與 277 限定 CONFORMED，原版返回／模式 2 等待閘門已閉合；255 READY。下一最小行動按公開 PIT count latch 契約建立窄規格，READY 後實作，再自行重生自然收據。不拆 ISR／timer driver／busy-wait 或猜補玩法；主選單、正常玩家路徑、音效、受控亂數與 Go remake 玩法同狀態尚未完成。


## 2026-10-02：模式 2 計數鎖存與 E4 立即 byte 輸入

- 281 推送後繼續同一路徑，工具基線 9b8d1a07121fab929fc94a7f529449beb0d49742；[規格 282](https://github.com/wicanr2/dosgolem/blob/c58709c5ffc8841a22112ad1ea8016890f87d9e5/docs/spec/282-pit0-mode2-count-latch.md) 與 [規格 283](https://github.com/wicanr2/dosgolem/blob/c58709c5ffc8841a22112ad1ea8016890f87d9e5/docs/spec/283-cpu386-in-al-imm8.md) 分別 DRAFT＋索引→公開契約與原版自然停點審查→READY→實作→限定 CONFORMED。提交 c58709c5ffc8841a22112ad1ea8016890f87d9e5 已推 github/codex/moo2-parity-20260930，未推本機 origin；主庫玩法 RE 閘門保持關閉。
- **公開平台／CPU 契約**：[Intel 8254](https://www.cs.cmu.edu/~410/doc/8254.pdf) 的 count latch 凍結值、低高讀取／未讀完重複忽略、重新編程取消；282 只支援已載入 Mode=2／43h←00h，以既有共享分數相位計數，不另開時計。模式 3／未鎖存直接讀／read-back 等仍拒絕，完整硬體相位仍為 hardware-spec approximation。[Intel 80386 IN](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IN.htm) 的 E4 ib 將立即埠編號補零、byte 寫 AL／旗標不變；283 只接裸 E4 與明示平台 PortIn，不建立 IOPL／TSS／VM86 或其他輸入寬度。
- **已證實，282 工程與初次自然收據**：固定官方 1.31 EXE、417 原檔／MOX.SET 及 Go 1.24.13／映像沿前節。PIT 全合法除數邊界／count16／分數相位、凍結／讀寫交錯、共享兩 CPU 及拒絕測試通過，moo2-282-platform-tests.txt SHA-256 1c7da0875eaf21fe4ba2b51051ad423b239572d90a14f3f7df86767988997013。固定 EXE 全套 SHA-256 f5fe6a43116c4138f75a3348e151d8f151064bf9f3b1fbb20625fe8705141d7a。兩個自然排程成功 OUT 43h←00h，於第 20,634,820 步實際停高位 LE 0x239B3E 的 E4 40，readCount=5681／pending=true、尚未第一讀；gzip SHA-256 1a929e805bbf6f95dd1f3b5c990990c56abede7ed967505b1d479a4328d7325a／c24408b3ab822f01e771b505e7cd143876669c8722eeae753a4bcaff7370329b。
- 新共享 CPU 測試首次樣板只有 BDA 容量而無程式空間，修正容量／I/O 接線後揭露保護 CPU E4 缺件；實模式已通過。第二次失敗依規則重查路由，282 先以已支援 EC 隔離平台驗證，實際 E4 命中與公開契約具備後才 283 READY／實作。兩份失敗 SHA-256 d9a77898a2e29222f2f30172820f6e73f737d42d11ca98f6ba99cc45b2a56130／d749bfe9fa7b1ad4b4ba179c41b2cb5904087094a3effcbc568541e6a03203f2，保留本機；沒有把環境／依賴失敗寫成 latch 產品缺陷。
- **已證實，283 回歸**：全部 256 埠×256 byte／高位與其他 CPU／FPU 狀態、EDX 故意不同、一次回呼、失敗／前綴／VM／未知寬度拒絕通過。GOMAXPROCS=2 go test -buildvcs=false ./internal/cpu386 -count=1 -v 全部通過，SHA-256 59daba763e0352594d9246d230ed7622474f53e0d177ae9b5970afc114c64057；固定 EXE go test -p 2 -buildvcs=false ./... -count=1 通過，SHA-256 a853240f01a710fae882b28f42ea795a07208e77795c76e19b3bd2afe3501154。
- **已證實，dosgolem 自行消費**：沿 277 的 50M／分離 DOS／完整原檔自然命令，輸出名稱 283，第二條件加 DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1。第 20,634,818 步 OUT 前 credit=74946000000；OUT 後 credit=75261000000、latch=5681（1631h）。高位 LE 0x239B3E 第一 E4 後 AL=31h／pending=true，0x239B42 第二 E4 後 AL=16h／pending=false；旗標246h／其他暫存器與段保持，EDX=1 未被誤當埠號。兩讀間 credit 繼續增加而凍結值不變，原版自然繼續，後續 latch=5302／5180。有限只讀邊界已加入受版控探針，沒有注入時計／遊戲值；這是本工具近似下的轉移，不聲稱跨執行器同時鐘／逐次 IN 值一致。
- **目前前沿**：兩個排程均第 20,637,037 步停高位 LE 0x254249，bytes 80 2D C0 26 27 00 08 C1 ED 08 8A C3 EB 26 8A 0D，byte 記憶體 SUB 的 80／ModRM 2D 未支援；EAX=80000000h、EBX=31488h、ECX=6BBC7Ch、EDX=272610h、DS／ES／SS=188h、flags=212h。只記標準 CPU 缺件，不推定資料欄位用途。IRQ0 started=completed=1595，等待值1；PIT Mode=2／Reload=5966／Generation=5、readPending=false、Micros=21094625／Deliveries=1615。事件條件已實際注入，但完整游標／玩家操作未驗。
- 自然 gzip SHA-256 a4cadecfd90e37dafc30fb8cd472169cd216039956a479f5ea45a933fc77f468／9a3554fc9a7023fa375c814b6685574c5fa03aa83b4b3069df06a657a0efb181；兩 PNG 同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，完整原版資料／終端／gzip／PNG 留本機。
- **回填／下一步**：26 個回填函式、新六項缺證據負例／實際 CLI、兩條件三筆 latch／兩次 IN 的實際收據與繁體字／索引／UID 稽核通過。276／281／282 舊停點已回填；282／283 限定 CONFORMED、255 READY。下一步依公開 SUB 契約建立 byte 記憶體目的／imm8 的新窄規格，READY 後實作，不深挖 driver／ISR／runtime 或猜補玩法。主選單、正常玩家路徑、音效、受控亂數與 Go remake 同狀態仍未完成。


## 2026-10-02：byte 記憶體 SUB 的原版後態

- 工具基線 c58709c5ffc8841a22112ad1ea8016890f87d9e5；[規格 284](https://github.com/wicanr2/dosgolem/blob/c1c8e731c4bcb0a5fda529f43e066508db74ba49/docs/spec/284-cpu386-sub-byte-memory-imm8.md) 保存 DRAFT／READY 審查、原始 byte 與限定 CONFORMED 收據。只補標準 CPU 記憶體目的，不改主庫玩法，未推定 DS 欄位用途或解 runtime／driver／ISR。
- **已證實，工程驗證**：Go 1.24.13／既有工具映像；全部 256×256×兩種旗標、ModRM／SIB／DS／SS／地址繞回及讀写失敗／前綴拒絕通過。前綴負例用所有段均可寫的初態排除未知段掩蓋。CPU 全套及固定官方 EXE 全套通過，SHA-256 3e507d508afaf2049fa06cff5e2eb88019155f99dfaddc79a78ab54371877d5e／92fc976dd96d5d1f659b4b26e7d620e7740f6d3abb0c41f0980c17488a077fe1。
- **已證實，dosgolem 自行重生**：沿既有 50M／分離 DOS／417 原檔／固定 MOX.SET，有無受控滑鼠事件兩條件。高位 LE 0x254249、原始 bytes 80 2D C0 26 27 00 08，DS:002726C0 byte=16h，第 20,637,038 步 0x254250 byte=0Eh；完整 R／段保持、flags=212h。第二次 byte=0Dh→05h、flags=202h→206h，同樣保持 R／段。沒有注入目的 byte、遊戲計時或亂數；此有限 CPU 後態不等於跨原版完整玩法同狀態。
- **目前前沿**：兩條第 20,637,097 步停高位 LE 0x25425F，bytes 80 05 C0 26 27 00 18 8A C3 8B 1E FE C9 8B EB D3，byte 記憶體 ADD／imm8 缺件。IRQ0 started=completed=1595、active=false／failed=false，等待值仍1，Micros=21094685。事件條件第 1,612,067 步已注入 x=657／y=189，仍不是完整座標／游標或玩家操作驗收。
- 自然 gzip SHA-256 01efe7de89d0f2968720aa52ad742093d3fc5864624b19d782b326cc91703b06／ceb730b6fb5002e0e32f02a868e6363f05a20148834399cf2582cf402783486f；兩 PNG 仍同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。原版素材、完整終端／記憶體／gzip／PNG 留本機，只有自製來源、測試與文字證據公開。
- 工具提交 c1c8e731c4bcb0a5fda529f43e066508db74ba49 已推送 github 隔離分支；全部 27 個回填函式、新四項缺證據負例／CLI 與兩排程 byte／完整狀態稽核通過。255 READY，主選單／正常玩家路徑、音效／受控亂數及整款 remake 未完成；下一步只按公開 ADD 契約與只讀 byte 建窄 CPU 規格。


## 2026-10-02：byte 記憶體 ADD 的原版後態

- 工具基線 c1c8e731c4bcb0a5fda529f43e066508db74ba49；[規格 285](https://github.com/wicanr2/dosgolem/blob/96aba2440ce181a1808a508ee897985a2fbcdc99/docs/spec/285-cpu386-add-byte-memory-imm8.md) 保存公開 Intel ADD 契約、唯讀初態與 READY 審查及限定 CONFORMED 收據。只補標準 CPU，不改主庫玩法或推定 DS 欄位用途。
- **已證實，工程回歸**：全部 byte 配對／兩種初始旗標、ModRM／SIB／DS／SS／地址繞回、完整資料保持及拒絕邊界通過。CPU 全套與固定官方 EXE 全套通過，SHA-256 047482ae709f6d2d8d949b61a3c4e54ceedc44c77ae1b0b661d257b0fb7c5ec0／8ef615163deaa99eb1a79206023e87d14294c722ba1e0a274fa685de911fd65e。首次地址繞回測試誤留 SUB 編碼，改 ADD 編碼後同命令重跑通過，失敗收據保留本機。
- **已證實，dosgolem 自行重生**：同固定官方 EXE／417 原檔／MOX.SET、有無受控滑鼠事件兩條件。第 20,637,097 步高位 LE 0x25425F、bytes 80 05 C0 26 27 00 18，DS:002726C0 byte=03h／flags=297h；下一步 0x254266 byte=1Bh／flags=206h，完整 R／段保持。自然 gzip SHA-256 6c22f3705380a934f0ff26565b3b21f88926c149147b2f6a1d143348eebeff3f／7c8a522b61673f4d145b18bc3e4d5b367f3cfba1b5e54aa28f9dcd39428f4800。
- **目前前沿**：兩條第 20,637,105 步停高位 LE 0x254275，bytes 80 F1 FF 83 C6 04 D3 ED 59 89 07 83 C7 04 83 F9；byte 暫存器 CL／imm8 FFh 的 XOR 缺件，ECX=6BBCF9h、flags=282h。IRQ0 started=completed=1595、等待值1、Micros=21094693。兩 PNG 仍同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。
- 28 個回填函式、新四項缺證據負例／CLI 與兩排程原版 ADD 後態稽核通過。索引舊摘要一併按收據修正。工具提交 96aba2440ce181a1808a508ee897985a2fbcdc99 已推送 github 隔離分支；原版素材與完整終端／記憶體／gzip／PNG 留本機。255 READY；主選單／正常玩家路徑、音效／受控亂數與整款 remake 未完成。下一步只按公開 XOR 定義／未定義旗標契約建立窄 CPU 規格。


## 2026-10-02：byte 暫存器 XOR 與 dword 暫存器 TEST

- 工具起點 96aba2440ce181a1808a508ee897985a2fbcdc99；[規格 286](https://github.com/wicanr2/dosgolem/blob/30faf6eb5643aac18d197c64804f81c18293876b/docs/spec/286-cpu386-xor-byte-register-imm8.md) 與 [規格 287](https://github.com/wicanr2/dosgolem/blob/4d7ae6c2176feda4c61e47ee2db41695a25c2208/docs/spec/287-cpu386-test-dword-register-imm32.md) 各自保存 DRAFT／READY 審查、原始完整初態及限定 CONFORMED 收據。公開 CPU 逐 bit 規則、定義五旗標與既有 AF 清除模型分開驗，不解 helper／runtime／driver／ISR 或改主庫玩法。
- **已證實，工程驗證**：286 八 byte 目的全配對／兩種旗標與完整外層保持、前綴／截短／未知 group 拒絕通過。287 八 dword 暫存器的全 bit／補集／符號／高 word 邊界與全低 byte 遮罩、完整 R／段／FPU／記憶體保持及 JNZ 兩方向通過；既有 word／記憶體 TEST 與其他 group 保持。全部 CPU 與固定官方 EXE 全套通過。CPU SHA-256 286 de88162fa7262e3ae1513d3a13ef2593465bd0647cd0549673931f4c1ddb60ea、287 e3dabc645fc4da8b16e174461054dfd82f1d0610a21b21ee1cca33a15c5aa4d0；全套 SHA-256 f0469982e157e380e3146b72922d63c4c35bc1bbb2251f464bc364d5393c28da／e517ff40941fbd770a536caf2348751aa874f42cbb806c8a48935a3a9bb34b52。
- **已證實，dosgolem 自行重生**：同固定 EXE／417 原檔／MOX.SET、有無受控滑鼠事件兩自然條件。高位 LE 0x254275→0x254278 的三筆 CL F9h→06h、FCh→03h、FDh→02h、目的外 R／段與定義旗標已驗。其後第 20,651,437 步高位 LE 0x254499 的 F7 C1 00 00 00 80，ECX=400h／flags=213h；下一步 0x25449F flags=246h、完整 R／段保持；第 20,651,439 步第一 JNZ 不跳，續行至 0x2544A1、完整 R／段與旗標保持。AF 1→0 僅為工具模型，不冒稱原版硬體未定義值。
- 286 自然 gzip SHA-256 ee93f0ae4120843bf105d14ca5c7f60971fac7f29d3129e743318c1ac6c5bafc／01118b0985ecbc60bed4cf5be32bf718a637f7353241f18d257486e8d5e76489；287 為 b3341128e31fbed5f8fd3f191a7de03967a537cb870f50377dbf59d6caf6b004／c146fc6aae759c1e114fbcd237a9c460571f9e36cfb6072b7b45c95592ed8fa5。較晚第 20,651,512 步到 0x2544CB 的狀態不混作第一分支樣本。
- **目前前沿**：兩條第 20,651,583 步停高位 LE 0x254510，bytes C1 C0 08 66 39 05 E0 27 27 00 75 07 33 C0 89 7A，dword EAX 左循環移位／imm8 缺件，EAX=02000000h／flags=286h。IRQ0 started=completed=1598、等待值4、Micros=21109931。兩 PNG 仍同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；255 完整座標／游標仍 READY，主選單／正常玩家路徑、音效／受控亂數與整款 remake 未完成。
- 工具提交 30faf6eb5643aac18d197c64804f81c18293876b 與 4d7ae6c2176feda4c61e47ee2db41695a25c2208 已推送 github 隔離分支。兩項舊停點／索引回填，各自四項缺證據負例／CLI、原版完整後態及全部 30 個回填函式通過；原版素材與完整終端／記憶體／gzip／PNG 留本機。下一最小行動按公開 ROL 計數／CF／OF 與未定義邊界建窄 CPU 規格。


## 2026-10-02：dword 暫存器 ROL 的原版後態

- 工具基線 4d7ae6c2176feda4c61e47ee2db41695a25c2208；[規格 288](https://github.com/wicanr2/dosgolem/blob/c605a7f13e00c061d09f14bbf7583c201335f980/docs/spec/288-cpu386-rol-dword-register-imm8.md) 保存公開 Intel 計數／旗標契約、完整唯讀初態與 READY 審查及限定 CONFORMED 收據。只新增裸 C1 /0 的通用 dword 暫存器形式，不改主庫玩法；多位 OF 未定義，保留只屬工具近似。
- **已證實，工程驗證**：八目的／全部256計數／32位邊界與每 bit／補集／CF／OF 四組合／保持旗標兩初態、完整外層保持與截短／前綴／記憶體拒絕通過；既有 word ROL／dword ROR 保持。全部 CPU 與固定官方 EXE 全套通過，SHA-256 1abfe7dcb15a6419bb53484c49815b8432aa3ba1e482d513c76d82202e198c22／de3e2f0f9e1a7af3ca5379ab9518b4dbecf3ac2533d81a22be1db185c4aab8db。
- **已證實，dosgolem 自行重生**：固定官方 EXE／417 原檔／MOX.SET、有無受控滑鼠事件兩自然條件。高位 LE 0x254510→0x254513，三筆 EAX=02000000h→2h、11000003h→311h、D4000000h→D4h，完整目的外 R／段保持，CF=0／1／0。下一 word 比較自行消費資料、到 0x25451A 的 flags=293h；該旗標屬比較後態。自然 gzip SHA-256 c08ea3956f0eb327e05d87a4900ced02ba29f1b1f3fab853748ba090aaff9992／8d72e98ab94cd187ba69133e23d70f088e2fa5d76741a4973d18803086875c93。
- **目前前沿**：兩條第 26,396,706 步停高位 LE 0x2545EF，bytes F6 D9 8A D1 89 54 AF FC 4D 75 A3 5D C3 87 DB 87，byte CL NEG 缺件，ECX=FFFFFFF6h／flags=297h。IRQ0 started=completed=2810、等待值1216、Micros=27167604。兩 PNG 同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；255 完整座標／游標仍 READY，主選單／正常玩家路徑、音效／受控亂數及整款 remake 未完成。
- 工具八檔提交 c605a7f13e00c061d09f14bbf7583c201335f980 已推送 github 隔離分支。287 舊停點／索引同步回填，31 個回填函式、新四項缺證據負例／CLI 與兩排程三筆完整後態通過。原版素材與完整終端／記憶體／gzip／PNG 留本機；下一最小行動按公開 NEG 契約保存完整初態並建立窄 CPU 規格。


## 2026-10-02：byte 暫存器 NEG 的原版後態

- 工具基線 c605a7f13e00c061d09f14bbf7583c201335f980；[規格 289](https://github.com/wicanr2/dosgolem/blob/78738dc77eeaee6a1578f7ee98027d679408226f/docs/spec/289-cpu386-neg-byte-register.md) 保存公開 Intel byte／六旗標契約、唯讀完整初態與 READY 審查及限定 CONFORMED 收據。只補裸 F6 /3 暫存器形式，不改主庫玩法或推定欄位用途。
- **已證實，工程驗證**：八目的／256來源／64算術旗標初態、完整外層保持／拒絕及既有 TEST／DIV／dword／堆疊 NEG 回歸通過。首次 CPU 全套遇舊未知 group 測試與新支援的 /3 衝突，改用仍未支援的 /2 byte NOT 保留護欄，同映像／命令重跑通過，失敗收據保留。全部 CPU 與固定官方 EXE 全套通過，SHA-256 96e676993d762f55ba4f5f4524730510bd3470643dce79388243d28cec0848a9／e404aab29a5b4a0e1e0e9fdd2b7e32559942c55e839ed292e8de777f2c30882b。
- **已證實，dosgolem 自行重生**：固定官方 EXE／417 原檔／MOX.SET、有無受控滑鼠事件兩自然條件。高位 LE 0x2545EF→0x2545F1，三筆完整 ECX=FFFFFFF6h→FFFFFF0Ah、FFFFFFF5h→FFFFFF0Bh、FFFFFFFBh→FFFFFF05h，完整目的外 R／段保持，flags 分別217h／213h／217h。下一 MOV DL,CL 自行消費到 0x2545F3，完整 EDX=458F00Ah／329840Bh／3D5C405h，目的外 R／段與 NEG 旗標保持。自然 gzip SHA-256 495b237c2a7af2f16deeb12bf859a0a07ea36a3b3c0248c4c0d56bb9ac49406d／a444e4bbb0e97291d82b6376d26d4d9f2484b59eee3d52481c18398136e2a9dd。
- **目前前沿**：兩條第 38,427,368 步停高位 LE 0x254A04，bytes D2 E5 08 2C 17 83 C6 04 48 75 DA C3 81 C6 68 74，SHL CH,CL 缺件，ECX=102h／flags=202h。IRQ0 started=completed=5346、等待值3752、Micros=39852240。兩 PNG 同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；255 完整座標／游標仍 READY，主選單／正常玩家路徑、音效／受控亂數及整款 remake 未完成。
- 工具九檔提交 78738dc77eeaee6a1578f7ee98027d679408226f 已推送 github 隔離分支。288 舊停點／索引同步回填，32 個回填函式、新四項缺證據負例／CLI 與兩排程三筆完整後態通過。原版素材與完整終端／記憶體／gzip／PNG 留本機；下一最小行動按公開 byte shift 契約保存完整初態並建立窄 CPU 規格。


## 2026-10-02：CL 計數 byte 左移與記憶體 OR 的完整資料鏈

- 工具基線 78738dc77eeaee6a1578f7ee98027d679408226f；[規格 290](https://github.com/wicanr2/dosgolem/blob/0c6d53b869cb52fd716d95c868326244614abbf1/docs/spec/290-cpu386-shl-byte-register-cl.md) 與 [規格 291](https://github.com/wicanr2/dosgolem/blob/0c6d53b869cb52fd716d95c868326244614abbf1/docs/spec/291-cpu386-or-byte-memory-register.md) 保存原始完整初態／公開 Intel 契約、READY 審查及限定 CONFORMED 收據。只補通用 CPU，不改主庫玩法；移位未定義 AF／OF／大計數 CF 與 OR 的 AF 明列工具近似。
- **已證實，工程驗證**：290 八 byte 目的／全部來源與 CL／別名／初始旗標、完整保持／拒絕及既有 byte／dword 移位回歸通過。291 八來源／全部 byte 配對／兩種初始旗標、全 ModRM／SIB／scale／index／base／DS／SS 分離、位移／地址繞回／段末單 byte、bus 讀寫失敗／完整保持／既有暫存器配對及原始資料鏈通過。全部 CPU 與固定官方 EXE 全套通過；291 CPU／全套 SHA-256 5bd1e18e86ed13e733fb07a34419b14bf39c95e9c9a7966a7e860760f31b773e／1bdcc7759989a46b372c57b698c214bd874acce584ac651d0bc2be1767eea084。
- **已證實，dosgolem 自行重生**：固定官方 EXE／417 原檔／MOX.SET、有無受控滑鼠事件兩自然條件，三組高位 LE 0x254A04→0x254A06→0x254A09→0x254A0C。完整 ECX=102h→402h、101h→201h、104h→1004h，CL 與完整目的外 R／段保持；下一 OR 分別寫回 DS:006BBC60／006BBCC2／006BBC7A 的00h→04h／02h／10h，完整 R／段與 flags=202h 保持；下一 ADD ESI,4 自行完成，只有 ESI 與算術旗標改變。290 首次資料鏈停在 OR 時保持 READY，291 真實消費完成後才一併驗收。自然 gzip SHA-256 f3668c23c7c88587088a792fde27e0d982f93240e74883713500d91c148c6dc5／834d501e25dc3159120740655d123f9687bc87961fbd1ccaac7993246a8f413e。
- **目前前沿**：兩條第 39,983,174 步停高位 LE 0x25488F，bytes F3 AF 83 EF 04 8B 07 2B 3D 00 2C 27 00 83 F0 FF，REPE SCASD 缺件，EAX=FFFFFFFFh／ECX=800h／ES=188h／flags=246h。IRQ0 started=completed=5675、等待值4081、Micros=41492921。兩 PNG 同已檢視黑圖 SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；255 完整座標／游標仍 READY，主選單／正常玩家路徑、音效／受控亂數及整款 remake 未完成。
- 工具十檔提交 0c6d53b869cb52fd716d95c868326244614abbf1 已推送 github 隔離分支。289／290 舊停點與索引同步回填，34 個回填函式、兩份新規格各四項缺證據負例／CLI 及兩排程三組完整資料鏈通過。原版素材與完整終端／記憶體／gzip／PNG 留本機；下一最小行動按公開 SCASD／REPE 保存完整初態並建立窄 CPU 規格。

## 2026-10-02：REPE SCASD 的完整掃描與讀取消費

- 工具基線0c6d53b869cb52fd716d95c868326244614abbf1；[規格292](https://github.com/wicanr2/dosgolem/blob/e914184166c2397dba5041617793a58e42f2f927/docs/spec/292-cpu386-repe-scasd.md)保存固定EXE／原始定位、完整R／段／8192 bytes掃描資料與雜湊、公開Intel契約、READY審查及限定CONFORMED收據。只接32位F3 AF，六算術旗標全定義；主庫玩法RE閘門不變，單次Step／內部IRQ／Error與restart模型界限明示。
- **已證實，工程驗證**：19×19個32位邊界配對／64算術旗標／兩方向、不同初始ZF／計數／退出位置、ES與DS分離、只讀／非對齊／段末dword／EDI繞回、逐byte讀取失敗／旗標恢復／進度保持、前綴／截短／既有SCASB與完整外層保持通過。全部CPU與固定官方EXE全套首次通過，SHA-256 9b8c4a93e7acb119fb45469cfe71faf486f6b87e68653b2837b73aa69412f1da／c5e4313060d6a556a5aaadf314221247c57d5f30a6b84e42f0fc3de891fa026f。
- **已證實，dosgolem自行重生**：固定官方EXE／417原檔／MOX.SET、有無受控滑鼠事件，兩條高位LE0x25488F→0x254891→0x254894→0x254896。從原始資料獨立確認前58個dword相等，第59個FFFFFBFFh；第39,983,175步完整ECX=7C5h／EDI=6BBD4Ch／flags=206h，下一SUB回6BBD48h，MOV真實讀取EAX=FFFFFBFFh，其他完整R／段保持。自然gzip SHA-256 8985ec08936482ab06857d488af15171bd12743b5ca16504b6984b917d76b9c6／3f0d8a42a931b65f6c4175b1d4eba59cf2e56b87137562d91f2ec9f7dc440251。
- **目前前沿**：兩條第39,983,178步停高位LE0x25489C，bytes 83 F0 FF C1 E7 03 0F BC D0 03 FA 66 89 3D D0 26；XOR／imm8缺件，EAX=FFFFFBFFh／ECX=7C5h／flags=206h。IRQ0 started=completed=5675、等待值4081、Micros=41492925；兩PNG同已檢視黑圖SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。255完整座標／游標仍READY，主選單／正常玩家路徑、音效／受控亂數與整款remake未完成。
- 工具八檔提交e914184166c2397dba5041617793a58e42f2f927已推送github隔離分支並回讀一致；291停點與索引同步回填，35個回填函式／新四項缺證據負例／CLI通過。原版素材與完整終端／記憶體／gzip／PNG留本機，下一最小行動是保存83／6 XOR的完整唯讀初態，再按公開契約建立窄CPU規格。

## 2026-10-02：dword XOR／BSF 與原版 word 儲存資料鏈

- 工具基線e914184166c2397dba5041617793a58e42f2f927；[規格293](https://github.com/wicanr2/dosgolem/blob/a899e9e7e4f397d17054bdccae97faa0f6cf963c/docs/spec/293-cpu386-xor-dword-register-imm8.md)與[規格294](https://github.com/wicanr2/dosgolem/blob/a899e9e7e4f397d17054bdccae97faa0f6cf963c/docs/spec/294-cpu386-bsf-dword-register.md)保存完整原始R／段／資料、公開Intel契約、READY審查、來源／收據雜湊與限定CONFORMED範圍。XOR的AF清除、BSF五未定義旗標與零來源目的保留都是工具模型；未變更主庫玩法。
- **已證實，工程驗證**：八XOR目的／全部imm8／76來源／四初始旗標及符號延伸／獨立逐bit oracle、全BSF配對／別名／低16位／置1位置／旗標與拒絕／外層保持、完整自製資料鏈通過。新附帶ADD回歸的AF預期錯誤已修正且CPU未改，同命令全部重跑通過；失敗收據保留。全部CPU／固定EXE全套通過，最終SHA-256 9f9f260317a0228cbb6b012de9546c128ae770ecb6dcf743213dfc2c016b6c90／e865e3ac8f9ca3794654a9503e91b01dc413b06398b59ca6d8672c9f578f646a。
- **已證實，dosgolem自行重生**：有無受控滑鼠事件兩條高位LE0x25489C→0x25489F→0x2548A2→0x2548A5→0x2548A7→0x2548AE，39,983,178至183連續步數。完整EAX從FFFFFBFFh經符號延伸XOR成400h，BSF讀取並產生完整EDX=Ah／定義ZF=0，再由ADD成EDI=74Ah及word MOV存回DS:002726D0=074Ah；其他完整R／段保持，未定義旗標模型另驗。293首次停BSF時保留READY，真實消費完成後才與294一併驗收。自然gzip SHA-256 76a750012abac608e9a4a481372344d014400f0ad727376ed7bb510413fc0e81／b8c34ef4981903fab0e2f25a3a1f180d65a1b14f530858a2ce3c4c1e0645eb05。
- **目前前沿**：兩條第41,223,220步停高位LE0x23C36B，bytes09 86 84 03 00 00 EB 25 8B 04 24 8B 94 86 04 04；記憶體dword OR缺件，來源EAX=2000h／flags=206h，目的DS:[ESI+384h]完整資料待唯讀觀測。IRQ0 started=completed=5936、等待值4342、Micros=42800171；兩PNG同已檢視黑圖SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。255完整座標／游標仍READY，主選單／正常玩家路徑、音效／受控亂數及整款remake未完成。
- 工具十檔提交a899e9e7e4f397d17054bdccae97faa0f6cf963c已推送github隔離分支，遠端回讀一致；37個回填函式與各四項缺證據負例／兩CLI通過，原始ZIP／patch／EXE雜湊一致。所有原版素材及完整終端／記憶體／gzip／PNG留本機，下一最小行動按公開OR的ModRM／段與旗標／拒絕邊界保存完整初態、建窄CPU規格。

## 2026-10-02：記憶體 dword OR 與原版完整 MOV 消費

已證實：工具基線a899e9e7e4f397d17054bdccae97faa0f6cf963c，295同次DRAFT＋索引、未改CPU的完整初態與公開Intel09 /r契約／錯誤模型審查READY後，接既有decodeAddress32。全部CPU與固定官方1.31 EXE全套通過，兩自然第41,223,220步高位LE0x23C36B的目的DS:00325864由40h寫成2040h，五定義旗標均0、AF清除模型後flags202h，完整R／段保持。

真正consumer：第41,255,293步高位LE0x23B693，bytes 8B 86 84 03 00 00，MOV EAX,[ESI+384h]讀到完整2040h，其他R／段／flags206h保持，OR新增bit13確實被完整dword讀取。初步只觀察未變高byte的收據不計新增位元消費；有限段讀取觀測器縮至低兩byte／dword取址端。首次診斷包裝Bus破壞DPMI身分契約已修，CPU與DPMI不變；失敗Bus測試只修測試，無效與有效收據均保留本機。完整命令、輸入／來源／所有收據雜湊與AF／部分byte寫入／Error模型見[鎖定規格295](https://github.com/wicanr2/dosgolem/blob/418ca3cf6d874e24127da24c0ba66e9fafecf6e7/docs/spec/295-cpu386-or-dword-memory-register.md)。293／294停點與索引／回填護欄已維護，38個回填函式／8項新負例與CLI及獨立兩自然資料鏈核對通過。295限定CONFORMED，沒有玩法或硬體未定義exact聲明。

新停點：兩自然第42,347,254步高位LE0x2454AE的INT31/0300h，AX0300h／BX0066h。實模式INT66入口1201:016A，前進230步停1201:05D9，OUT 022Ch／C6h未支援、Returned=false；外層INT31未處理由這個內層錯誤造成。IRQ0 started=completed=6173，等待DS:00271148=E3 11 00 00即4579；PIT模式2／Reload5966／Generation5，Micros43985557／Deliveries6193／Pending=false／InService=false。VBE Bank7／StartY512／BankSets447／Writes4353364／DisplaySets7，兩PNG同已檢視黑圖SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。受控事件已注入x657／y189；255、主選單／正常玩家路徑、音效／受控亂數及整款remake仍未驗收。

兩自然gzip SHA-256 2d8bb4b224cdbb24440177d6ae4856de345b7602b36323f2f280c4eca862dac1／3c1bd060f88f34ca1aaf9d4055571b830ecfef834dcbb416b4d1af818925fdce，CPU收據9333cc23374ef3409423be5d3b98b6210e3d2e9fe0fe7a7380f1c9fcc4f20ee3，固定EXE全套6ab325d7a816db9a928a9429be55ae94e0e56aa8e06db12d24a1b6f86d3fae11。原始ZIP／patch／EXE固定雜湊再次核對一致；素材、完整記憶體／終端與PNG不公開。工具提交418ca3cf6d874e24127da24c0ba66e9fafecf6e7已推送github隔離分支並回讀一致，未推本機origin。下一步公開SB16 DSP／DMA與sample duration契約及有限原版參數，審查READY後實作；不解driver／ISR／DAC／PIT硬體wall-clock，不猜欄位／玩法。主庫玩法RE閘門不變。

## 2026-10-02：SB16 C6h 命令與原版成功返回

已證實：基線工具418ca3cf6d874e24127da24c0ba66e9fafecf6e7，296同次DRAFT＋索引、有限非自然參數探針與公開Creative DSP／DMA／sample duration契約審查READY後實作。原版四byte C6 20 FF 07為無號立體聲、22050Hz／每block2048個8位sample，DMA1 mode58h／mask0Dh／地址4000h／count0FFFh／page1，即物理14000h的4096byte ring。四mode／全部length語法、條件實模式取樣數／40h與41h區別／block與ring／IRQ／拒絕邊界已驗；hardware-spec approximation明示，不做逐週期或人耳聲明。

兩原版自然排程高位LE0x2454AE接受完整命令，實模式INT66入口1201:016A前進333步返回FFFF:FFF0，Returned=true／Error空；完整外層R／段／flags206h保持，50byte封包只有offset29由04h→00h。原版第42,347,256步高位LE0x2454B3的MOV EBX讀取成功值0，下一CMP令flags246h，JZ自然跳0x2454E7；兩排程完整caller相同。回填293／294／295停點、39個回填函式／15項缺證據負例／CLI及兩自然獨立稽核通過，296限定CONFORMED。

來源收據更正：呼叫前第一DMA尚未設定，不能以地址0／count0作C6來源；改取返回後DMABase／page，4096byte全80h的SHA-256 78aacbc3fb34efb8ffa5467b931291ec2bdf5e19564fc45fe97b5affbc893dc6。返回VirtualMicros1375，最後參數1354，21µs×44100=926100信用，小於首sample門檻，PCMBytes=0／DMA完成0／無8位IRQ；原版PCM真實消費與保護模式連續PCM／IRQ7仍未知。條件pattern測試在46440µs首block、92880µs ring重載，不把它當原版播放完成。完整命令、公開原廠來源、實作／測試／有效及更正前收據雜湊見[鎖定規格296](https://github.com/wicanr2/dosgolem/blob/56e1ff979b517740bf056668ee90800bfac27321/docs/spec/296-sb16-c6-auto-init-dma.md)。

新停點：兩自然第42,347,639步高位LE0x257662，bytes C1 CA 10 A2 C0 26 27 00 66 8B C2 C1 CA 10 39 11，dword ROR立即數10h缺件。IRQ0 started=completed=6173／等待4579，PIT模式2／Reload5966／Generation5，BIOSClock Micros43986044／Deliveries6193／Pending=false／InService=false，兩PNG同已檢視黑圖。受控滑鼠事件仍在第1612067步注入x657／y189；255座標／游標、主選單／正常玩家路徑、保護模式音訊／IRQ7、人耳、受控亂數及整款remake未驗收。

有效兩gzip SHA-256 51a438fd826c55922c2115a1c77918be0d313ef14cf0e6c5b06c72a8689b8f42／a05791a08f8ee36df2cbc244076ffb005fed3b31cafa054291d19f1cc82272c9；CPU／機器層收據bdc4a5adc9db1fc42bb1ca13fccfb680c750146c7e4aa5feb3a1e23c128bfb2b、固定EXE全套eaa30a613877d185e9ba095078f73c32bee44c18cdd61f00cdf5a3f9688545bd。ZIP／patch／EXE固定雜湊再核對；素材、完整終端／記憶體／PNG不公開。工具提交56e1ff979b517740bf056668ee90800bfac27321沿既有授權推送github隔離分支，不推本機origin。下一步公開ROR計數／旗標契約與有限唯讀初態，審查READY後補窄CPU形式；主庫玩法RE閘門不變。

## 2026-10-02：dword ROR 全部立即數與原版 MOV 消費

已證實：工具基線56e1ff979b517740bf056668ee90800bfac27321，297同次DRAFT＋索引，未改CPU的兩自然完整初態與公開Intel契約審查READY後，擴充裸C1 /1、mod11的全部imm8遮罩計數。兩自然初態相同，高位LE0x257662／C1 CA 10，完整R依EAX ECX EDX EBX ESP EBP ESI EDI為347010 6BB370 AFF0AFF 74A 2BDB70 2215 711120 347094、段8 188 188 0 20 188、flags297h。全部CPU及固定官方EXE全套首次通過，完整R／段／FPU／記憶體保持、零計數與單位OF及拒絕邊界已驗，多位OF保留僅為工具模型。

兩自然各自重生兩組ROR→A2→MOV AX,DX→ROR。第一組第42,347,639步ROR使EDX仍0AFF0AFFh、CF=0／模型flags296h，完整其他R／段保持。第42,347,642步高位LE0x25766D已完成原版0x25766A的66 8B C2，MOV AX,DX真實消費後完整EAX=00340AFFh；其他R／段／flags保持。第二組第42,348,715步開始，相同EDX／CF與消費規則，EAX由0A0A0A06h轉0A0A0AFFh。第二ROR均執行成功；獨立單bit整除oracle核對完整兩組鏈後297限定CONFORMED，未把隔離測試當正常玩家路徑。

兩自然第42,349,111步外層高位LE0x2571C9的StepHook內，原版IRQ0分支停止高位LE0x2520B7，bytes D1 E0 D1 E3 F7 05 28 2D 27 00 08 00 00 00 74 04；D1 /4的SHL EAX,1缺件。內層解碼後CS:EIP=0008:002520B9，完整R為0 3D6978 8000 1 2723E0 2723F4 1 325048／段8 188 188 0 20 188／flags2；外層EIP不當缺件位址。IRQ0 started6174／completed6173／active=false／failed=true，等待DS:00271148仍4579；277舊返回樣本不擴張到全部IRQ0分支。PIT模式2／Reload5966／Generation5，BIOSClock Micros43987979／Deliveries6194／Pending=false／InService=false，兩PNG同已檢視黑圖。C6返回／來源收據保持，保護模式連續PCM／IRQ7、人耳、255／主選單／正常玩家路徑／受控亂數與整款remake仍未驗收。

固定輸入、公開來源、全部命令、CPU／probe／測試SHA-256與前態摘要腳本問題見[鎖定規格297](https://github.com/wicanr2/dosgolem/blob/4cdf20e347500a5c996b955831e35ef68ca556e0/docs/spec/297-cpu386-ror-dword-register-imm8.md)。有效兩自然gzip SHA-256 21fe3aeec9f3ece1b9721b7fb067c1acfb3136343ef2ce27b193bdeeb4057768／8519d0012606a941046b6504bd06be289ce6b687e4471b09d000fad07f496c1f；CPU收據ff42f53fbaf8928ca27521c3cc7999089c454e315d4597f42b00221b0e45b4d5／固定EXE全套1df3ef49982e47055e92dae7496b219992cd812e79c50b7a28f8e92b9b8badd5。222／288與293–296回填已維護，40個回填函式／24項缺證據負例／CLI及兩自然獨立資料鏈稽核通過。原始ZIP／patch／EXE雜湊再核對一致，素材及完整終端／記憶體／PNG不公開。

工具提交4cdf20e347500a5c996b955831e35ef68ca556e0已推送github隔離分支，不推本機origin；下一步依公開D1 /4單位SHL／旗標與完整唯讀初態審查READY，再補窄CPU形式。只跨過標準CPU缺件，不深入IRQ0 handler／ISR／driver／硬體時序、不猜用途或改主庫玩法。

## 2026-10-02：IRQ0 單位 SHL 的原版寫回

已證實：工具基線4cdf20e347500a5c996b955831e35ef68ca556e0，298同次DRAFT＋索引，未改CPU的兩自然真正IRQ0完整前態及公開Intel五定義旗標／AF模型審查READY後，補裸D1 /4、mod11的八dword目的。唯讀StepHook包裝原樣轉送既有hook，九個固定位址與TEST來源／兩目的各最多三筆，不替換CPU／Bus／IRQ橋接或時計、不跳指令。第一CPU回歸發現既有C1 E0 01／EAX80000001h的OF漏設，工具flags603h、正確預期E03h；保留失敗收據，299另走公開反例／READY規格才修單位OF。AF與多位OF清除仍只是工具模型，未有MOO2自然OF=1同狀態收據。

全部CPU乾淨重跑及固定官方EXE全套通過。兩自然各七筆IRQ0完整狀態相同，outer_step42349111不是IRQ內指令序號。高位LE0x2520B7真正前態完整R依EAX ECX EDX EBX ESP EBP ESI EDI為0 3D6978 8000 1 2723E0 2723F4 1 325048，段8 188 188 0 20 188、flags2；與未改CPU原版前態一致。第一SHL後0x2520B9的EAX仍0、flags46h，第二SHL後0x2520BB的EBX2、flags2，其他完整R／段保持。

同次TEST來源DS:00272D28=3，3 AND 8=0，0x2520C5的JZ跳過第二組SHL，到0x2520CB的A3；原版A3存EAX0至DS:00272D40，0x2520D0的89 1D存EBX2至DS:00272D44。0x2520D6觀測八bytes由全0到00000000 02000000，只有EBX有數值突變，MOV保持flags46h。完整保持／分支與真實寫回獨立核對後298限定CONFORMED，不猜欄位用途；第二組SHL自然分支未走，單元測試不當原版分支驗收。299限定公開CPU契約修正，不宣稱原版動態OF已對齊。

兩自然同outer_step42349111仍在外層0x2571C9的IRQ0呼叫內，轉停高位LE0x25179F，bytes13 ED 03 34 AD 40 2D 27 00 0F BF E8 01 2F 0F BF，ADC EBP,EBP缺件。內層解碼錯誤CS:EIP0008:002517A0、R為0 0 3D69C0 0 2723D8 0 71E1D0 3C6038、段8 188 188 0 20 188／flags847h；真正下一前態仍須有限唯讀保存，不把錯誤後態冒稱前態。IRQ0 started6174／completed6173／active=false／failed=true，完整返回未知。PIT mode2／Reload5966／Generation5，BIOSClock Micros43988016／Deliveries6194／Pending=false／InService=false，等待4579；37µs差為工具時鐘，不稱硬體wall-clock對齊。兩PNG同已檢視黑圖，SB16 C6實模式333步成功返回／來源收據保持，保護模式連續PCM／IRQ7及人耳未知。

CPU收據SHA-256 198c7c11abee84f928fea00f4da39d50b225ec29767bbd212d0cb92f53a16187；固定EXE全套f851638ad1926ee142d9c438160f7ba8475a59ebfdb2950dcd5565d386b7e9fb；兩自然gzip f6984adbd0f227ce8a0034c61189ae384aff2353ae9a3eb14be9872d5e51fd55／ced6336a1ae88b4d90905a4e4789451e346a1a7d94f7b56b397a7e61b7334b32；第一次CPU失敗7c672d84df30bdd041903bc7812b6bb839f000de01222b8af640520eaba52b33。來源／CPU／probe／新測試、前態、公開Intel來源及全部精確命令見[鎖定規格298](https://github.com/wicanr2/dosgolem/blob/a44c7eaae7f4ac3143a04f183f62ecf91190e124/docs/spec/298-cpu386-shl-dword-register-one.md)與[鎖定規格299](https://github.com/wicanr2/dosgolem/blob/a44c7eaae7f4ac3143a04f183f62ecf91190e124/docs/spec/299-cpu386-c1-dword-single-shift-overflow.md)。原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致；素材及完整終端／記憶體／gzip／PNG留本機。

293–297與186／298同次回填，42個回填函式／28項新缺證據負例、兩個CLI與兩自然完整鏈核對通過。工具提交a44c7eaae7f4ac3143a04f183f62ecf91190e124已推送github隔離分支並讀回一致，不推本機origin。下一步僅保存ADC真正完整初態與下一消費，依公開六定義旗標契約審查窄CPU規格；不深入IRQ0 handler／ISR／driver硬體時序／runtime、不改主庫玩法。255／主選單／正常玩家路徑、受控亂數與整款remake未完成。

## 2026-10-02：ADC 原版索引 ADD 的真實來源

已證實：工具基線a44c7eaae7f4ac3143a04f183f62ecf91190e124，300同次DRAFT＋索引，以未改CPU的兩自然真正IRQ0完整前態及公開Intel六定義旗標／來源目的別名／保持與拒絕契約審查READY後，補裸13 /r、mod11。四個固定指令位址各最多三筆唯讀StepHook原樣轉送既有hook，不替換CPU／Bus／時計／IRQ橋接、不注入資料或跳指令。全部CPU與固定官方EXE全套首次通過，六旗標／完整R／段／FPU／記憶體與非算術旗標保持及拒絕／既有ADD／SUB／CMP／SBB回歸已驗。

兩自然各三組四筆完整ADC→索引ADD→MOVSX鏈相同，全部outer_step42349111在外層0x2571C9的同次IRQ0內，不能冒稱IRQ內指令序號或總呼叫數。高位LE0x25179F第一真正前態完整R依EAX ECX EDX EBX ESP EBP ESI EDI為0 0 3D69C0 0 2723D8 0 71E1D0 3C6038，段8 188 188 0 20 188、flags847h；與未改CPU兩自然相同。第二／第三前態完整R為0 80000000 3D69C0 0 2723D8 0 71E1D2 3C6040及0 0 3D69C0 0 2723D8 0 71E1D2 3C6048，flags86h／847h。

原CF1／0／1令ADC結果EBP1／0／1，flags2／46h／2、目的外R／段保持。下一0x2517A1的03 34 AD 40 2D 27 00以EBP×4索引，讀DS:00272D44 dword2或DS:00272D40 dword0，兩來源八bytes始終00000000 02000000。真正ADD後0x2517A8的完整ESI依次0071E1D2h／0071E1D2h／0071E1D4h、flags6，其他R／段保持；0x2517AB的MOVSX才把EBP覆寫0，ESI及flags6保持。獨立無號總和／有號範圍／低nibble進位與PF計數核對來源／索引0、1與完整結果後300限定CONFORMED，不猜欄位用途。

兩自然同outer_step42349111在外層0x2571C9的IRQ0內轉停高位LE0x24678C，bytes66 83 F7 01 57 50 E8 6A BA 00 00 A1 EC 15 2B 00，word XOR DI,1缺件。內層解碼錯誤CS:EIP0008:0024678F、R為325048 3D6978 3D69C0 1 272400 FC4 0 0、段8 188 188 0 20 188、flags46h；下一真正前態仍待有限唯讀，不把錯誤後態當前態。IRQ0 started6174／completed6173／active=false／failed=true，完整返回未知。PIT mode2／Reload5966／Generation5，BIOSClock Micros44005445／Deliveries6194／Pending=true／InService=false、等待4579；工具時鐘不當硬體wall-clock對齊證據。

C6實模式333步成功返回／來源收據與兩PNG先前已檢視黑圖雜湊保持；保護模式連續PCM／IRQ7、人耳、255／主選單／正常玩家路徑／受控亂數與整款remake未完成，299原版自然OF=1同狀態收據限制保持。CPU收據SHA-256 1909949178aad7344e64788a254c4141990a4e2fcea26c90462efac491c75513；固定EXE全套f8de03a799e482ecb7490ca8876bf9f9b0198c41da5f1ce607c126b828d65302；兩自然gzip 039cae78c78a4cacd0371db655eee68c0ed1a8e537a70b10bab54ec56fc51e11／aa7729a378fdee9cc98defc0c18991e1579843de64e5261576ce4fcbeb363083。完整來源／CPU／probe／新測試、前態、公開Intel來源與全部精確命令見[鎖定規格300](https://github.com/wicanr2/dosgolem/blob/752abc607e04b87d830d4a6451af9e51dfc26b78/docs/spec/300-cpu386-adc-dword-register.md)。原始ZIP／patch／417根檔／固定EXE／MOX.SET再核對一致；素材及完整終端／記憶體／gzip／PNG留本機。

293–299舊停點與索引同次回填，43個回填函式／七份28項缺證據負例、CLI及兩自然完整資料鏈核對通過。工具提交752abc607e04b87d830d4a6451af9e51dfc26b78已推送github隔離分支並讀回一致，不推本機origin。下一步僅保存word XOR真正完整前態與PUSH消費，依公開低16位／立即數符號延伸與旗標契約審查窄CPU規格；不深入IRQ0 handler／ISR／driver硬體時序／runtime，不改主庫玩法。


## 2026-10-02：word XOR 與兩個原版 stack dword 寫入

已證實：工具基線752abc607e04b87d830d4a6451af9e51dfc26b78，301同次DRAFT＋索引，未改CPU的兩自然真正IRQ0完整初態與公開Intel word XOR五定義旗標／AF清除模型及PUSH契約審查READY後，只補66 83 /6、mod11的八word目的與全部imm8符號延伸。四固定位址各最多三筆唯讀StepHook原樣轉送既有hook，保存完整R／段／flags與既有八byte堆疊，不替換CPU／Bus／時計／IRQ橋接、不注入資料或跳指令；不追後續CALL目標／helper內部。

第一次301全部word／立即數、八目的旗標／完整保持與隔離PUSH測試通過，舊293仍拒絕已合法word形式、新byte XOR回歸誤填保留AF而失敗。依既有286清AF模型修正新增期望，267／293的word XOR舊負例由全部word正例接替，未知拒絕保持；CPU未再次修改，失敗收據保留。相同映像／命令乾淨重跑全部CPU及固定官方EXE全套通過。AF未定義清除只是工具模型，不稱硬體逐值對齊。

兩自然各一組四筆完整XOR→PUSH EDI→PUSH EAX一致，outer_step42349111在外層0x2571C9的同次IRQ0內。高位LE0x24678C真正前態完整R依EAX ECX EDX EBX ESP EBP ESI EDI為325048 3D6978 3D69C0 1 272400 FC4 0 0、段8 188 188 0 20 188、flags46h，與未改CPU兩自然相同。SS:002723F8／SS:002723FC八bytes=52672400 78693D00，兩dword原值00246752h／003D6978h。0x246790的XOR後完整EDI=1、flags2、其他R／段及八bytes保持；0x246791的PUSH EDI後ESP2723FCh、八bytes=52672400 01000000；0x246792的PUSH EAX後ESP2723F8h、八bytes=48503200 01000000。兩個原版stack dword各有真正數值突變，兩PUSH旗標及其他R／段保持；獨立逐bit比較／完整寫入及初態匹配核對後301限定CONFORMED。

這次IRQ0成功返回，兩自然終態active=false／failed=false／started6228／completed6228，等待DS:00271148=1A 12 00 00即4634。第42603292步轉停根CPU高位LE0x256171，bytes66 93 C1 CB 08 C3 90 8B C2 8A E2 8B DA C1 C8 18，word XCHG AX,BX缺件，比前停點多254181步。解碼後EIP256173h、EAX0A0A0A2Eh／EBX2E0A0A0Ah／ECX2E0A40C0h／EDX2E0A2E0Ah／flags206h；下一真正完整前態仍須有限唯讀，不用錯誤後態取代。PIT mode2／Reload5966／Generation5、BIOSClock Micros44292270／Deliveries6248／Pending=false／InService=false，工具時鐘不當硬體wall-clock一致證據。只證明這個自然IRQ0返回樣本，不擴張其他IRQ分支。

VBE Bank9／StartY512／BankSets452／Writes4660564／DisplaySets7，indexed SHA-256 9d4d567f9cbe2a0e0069255c8f979c1ba07e90fad5974e5a0cf96a6c1469ae56已有變化，但兩PNG仍同先前已檢視黑圖SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。C6實模式333步成功返回／來源收據保持，保護模式連續PCM／IRQ7、人耳、255／主選單／正常玩家路徑／受控亂數及整款remake未驗收，299原版自然OF=1限制保持。

CPU收據SHA-256 7ee5e1795791547f35233adb0b578dac033107368ca5690d6a8e9b6ae331956e；固定EXE全套4df4f49481a1f920ac0812d4694c98bf8d70a9ba9d8fa2f327b7d0bce0ffb7ac；兩自然gzip ccd184dd64aa984b77e21b6624fca6a7407c95b938dfa959203af32ef0901c13／339eb37ec3be286d6e42b95e7f87d8913ad9423cbf381052a77993fc3025b27b；第一次CPU失敗59cd8a70fec7de6e18004bad33361b885d01b8972205574e2aa11120e44fb73b。來源／CPU／probe／新測試、前態、公開Intel來源及精確命令見[鎖定規格301](https://github.com/wicanr2/dosgolem/blob/d013029fcddf4da8e8d7650660897223e6fa67ca/docs/spec/301-cpu386-xor-word-register-imm8.md)。原始ZIP／patch／417根檔／固定EXE／MOX.SET雜湊再核對一致；原版素材及完整終端／記憶體／gzip／PNG留本機。

293–300舊停點與索引同次回填，44個回填函式／32項新缺證據負例、CLI與兩自然完整XOR／真正堆疊消費／IRQ0返回及新停點獨立稽核通過。工具提交d013029fcddf4da8e8d7650660897223e6fa67ca已推送github隔離分支並回讀一致，不推本機origin。下一步只保存XCHG真正完整前態與下一ROR消費，依公開word交換／高16位與旗標保持契約審查窄CPU規格，不追helper／IRQ0 handler／ISR／driver硬體時序，不改主庫玩法。


## 2026-10-02：word XCHG 與原版星空圖像

已證實：工具基線d013029fcddf4da8e8d7650660897223e6fa67ca，302同次DRAFT＋索引，以未改CPU兩自然真正完整前態與公開Intel word交換／兩高16位及全部旗標保持契約審查READY後，只補66 91–97。三固定位址各最多三筆唯讀StepHook原樣轉送既有hook，不替換CPU／Bus／時計／IRQ橋接、不注入資料或跳指令。全部CPU與固定EXE全套首次通過，原高低word整除拆解／交換與完整保持、拒絕／既有byte及dword XCHG、NOP已驗。

兩自然各三組完整XCHG→ROR→RET邊界一致，第一前態outer_step42603292、高位LE0x256171，完整R依EAX ECX EDX EBX ESP EBP ESI EDI為A0A0A2E 2E0A40C0 2E0A2E0A 2E0A0A0A 2BDB6C 6EE4 70E2D2 3471A0、段8 188 188 0 20 188、flags206h，與未改CPU兩自然相同。XCHG後0x256173的完整EAX0A0A0A0Ah／EBX2E0A0A2Eh，兩高16位與其他R／段及全部旗標保持；下一ROR EBX,8後0x256176的完整EBX2E2E0A0Ah／flags206h。後兩組前態步數42616626／42621546，完整EBX100A0A0Ah、AX0A10h交換後EBX100A0A10h，下一完整ROR結果10100A0Ah，其他R／段保持。原值交換與逐次整除循環獨立核對後302限定CONFORMED；多位OF未定義保留只屬297工具模型，RET只記邊界，不追helper／caller內部。

兩自然都至step_limit=50000000 eip=0x22FCD2，完整R為FFFFFFFF 4000 0 0 2BDB64 2BDBB4 FFFFFFFF 3D6978、段8 188 188 0 20 188、flags246h，bytesFF 0D 40 8E 2A 00 89 F0 5D 5F 5E C3 57 55 8B 15。沒有step_error／guest_cpu_stop，不把診斷上限寫成CPU拒絕。無事件unique_sites17667，受控滑鼠事件17697；首次稽核誤要求兩分支覆蓋數相同，依各輸入實際值修正後用同一收據通過，沒有修改CPU或重跑原版。

終態IRQ0 active=false／failed=false／started7789／completed7789，等待DS:00271148=33 18 00 00即6195；PIT mode2／Reload5966／Generation5、BIOSClock Micros52095937／Deliveries7809／Pending=false／InService=false。C6命令／handled返回／成功caller與來源SHA保持，保護模式連續PCM／IRQ7及人耳仍未知；工具時間不稱硬體wall-clock一致。

VBE Bank9／StartY512／BankSets497／Writes4744640／DisplaySets7，indexed SHA-256 8c5886995a63e609e0b33bd49e49a331d7de452437485428d9919c60461fa98c；RGB SHA-256 8f688f78c811b308a51b62769add0f23e00718eda758a10e54318ef369122bf7，兩PNG同SHA-256 d648932f847a2fe5b87723b6537f76e816d64fb60e05c13321d15ede50f3b21b。已檢視圖像，中央有星空／星雲片段，未見主選單；原圖留本機，未公開原版美術。事件第1612067步注入x657／y189／buttons0，255完整座標／游標、正常玩家操作／受控亂數、299原版自然OF=1與整款remake仍未驗收。

CPU收據SHA-256 d05045f43ddfdf2021e0a6eee727ac4463aa7fe29310821169828fdd1ba75dbb；固定EXE全套263845c9b6cd1109b941dcf14740b3b19926bb0b4e94385d9073271f3b99f8cc；兩自然gzip fe3472fa0d7766a761b3b5f7cc5fab8cc22a296a16d0e947e3a093a62826be87／ae21d6b88b831f10addae20471effd45d0ce0e163de3d9067d0840829f4f44e9。來源／CPU／probe／新測試、真正前態、公開Intel契約與全部精確命令見[鎖定規格302](https://github.com/wicanr2/dosgolem/blob/9e6ee8cea7e40fdf13528fdae6b7959361708aaa/docs/spec/302-cpu386-xchg-ax-word-register.md)。原版ZIP／patch／417根檔／EXE／MOX.SET雜湊再核對一致；原始素材及完整終端／記憶體／gzip／PNG留本機。

293–301九份舊停點及索引同次回填，45個回填函式／36項新缺證據負例／CLI與兩自然完整消費、IRQ持續返回／50M上限核對通過。工具提交9e6ee8cea7e40fdf13528fdae6b7959361708aaa已推送github隔離分支並回讀一致，不推本機origin。下一步先核對既有正常鍵鼠入口與尾端caller的有限唯讀狀態／平台音訊與時計進度，確認所等條件後以原版正常輸入重播；不猜修等待或盲提高上限，不追helper／runtime／driver／ISR硬體時序，不改主庫玩法。


## 2026-10-02：晚期啟動的平台時計與正常輸入邊界

已證實：工具基線9e6ee8cea7e40fdf13528fdae6b7959361708aaa，303只加11筆有限唯讀平台快照，CPU及三份平台來源與302雜湊保持。固定官方1.31 EXE、417原檔及MOX.SET，以相同Go1.24.13映像／600秒／2GiB／2CPU／128pids／UID:GID1000:1000／network none，兩自然至50M上限，未注入按鍵／資料或改時計／IRQ橋接。完整初終態／三組XCHG與ROR／IRQ0／圖像收據與302一致，排除新增快照、PNG輸出名、解壓mtime及DOS DTA時間日期四bytes；檔案內容及尺寸未變。

| outer_step | BIOSClock.Micros | LEDeviceState.VirtualMicros | DMA8剩bytes／完成／PCMBytes |
| --- | --- | --- | --- |
| 42347255 | 43985659 | 1375 | 2048／0／0 |
| 42603292 | 44292269 | 1375 | 2048／0／0 |
| 48000000 | 49985912 | 1375 | 2048／0／0 |
| 50000000 | 52095937 | 1375 | 2048／0／0 |

DMA8啟動／自動／stereo／FIFO、rate22050/1、block2048／sampleCredit926100保持。LEOPLPorts.AdvanceRealMode才推進裝置時間，保護模式BIOSClock.advance沒有接這條鏈；自然快照證實音訊裝置進度未隨保護模式前進。先前實模式DMA16Completions1／PCM16Bytes2／IRQ7Deliveries1仍在，不稱整段無IRQ7。這些是工具缺件證據，原版目前等待是否因此造成仍未知，時間只屬既有近似。

原版高位LE0x239833的DOS AH2509以DS=8／EDX=21C4D8h安裝保護模式IRQ1。全部快照鍵盤未安裝、讀取／入隊與60h／61h／64h埠讀取零，BDA queue bytes1E001E00，實模式INT09／absolute IVT09零；工具正常鍵盤入口尚未接。受控滑鼠第1612067步注入x657／y189／buttons0，回呼started1／completed1，沒有按鍵。BIOS入隊不能代替未驗的保護模式IRQ1消費，不追handler內部。高位LE0x231AE4／0x231AEB／0x22FCD2各兩筆晚期快照的原始0x2A8E54四bytes皆零，只記比較條件，欄位語意及等待因果未知。

兩自然仍是根CPU高位LE0x22FCD2、flags246h，無CPU拒絕，IRQ0 started7789／completed7789，PNG同已檢視星空片段，主選單未見。兩gzip SHA-256 a4920940f3006218f9fdfb6d33085d7d7e18c4bc6095adffe8e31c3e62e23a28／c8f8517db2c90c06e9f109f8a9789fd392d7d430b941841ac18fca896f945db5；兩PNG同SHA-256 d648932f847a2fe5b87723b6537f76e816d64fb60e05c13321d15ede50f3b21b。原版輸入、來源、probe與全部精確命令及證據等級見[鎖定規格303](https://github.com/wicanr2/dosgolem/blob/90b9f4acb3c7e55829973c51a39fe12cb2129df3/docs/spec/303-moo2-late-startup-platform-observation.md)。完整終端／記憶體／gzip／PNG與原版素材留本機，不散布原版美術。

完整收據／五來源雜湊／11快照／AH2509核對通過，45個既有回填函式／36項缺證據負例及CLI保持。工具90b9f4acb3c7e55829973c51a39fe12cb2129df3已推送隔離分支並回讀一致，不推本機origin。303觀測已驗但平台契約仍DRAFT，未證明等待解除；下一步以公開Sound Blaster／PIC及原版IRQ7向量呼叫邊界審查共用裝置時間與正確派送，READY後實作，鍵盤另依保護模式IRQ1／埠契約處理。255、299自然OF=1、主選單／正常玩家路徑／受控亂數、人耳與整款remake仍未驗收，主庫玩法RE閘門保持。


## 2026-10-03：兩種 CPU 模式的裝置時間與首個 IRQ7

已證實：工具基線90b9f4acb3c7e55829973c51a39fe12cb2129df3，304以303的原版C6返回與固定DMA初態、Creative取樣率／block及Intel／Open Watcom模式框架公開契約審查READY後，接共用advanceDMA與保護模式裝置時計。既有1µs近似、實模式順序／來源／rate／block／ring保持，CPU與啟動來源未改。未知保護模式IRQ7保持pending、原始向量與CPU現場明確停止，不用實模式框架猜轉送或抑制中斷跑過等待。四mode／三rate、交錯兩真實CPU、40h已含channels、mask／reset／來源超界、16位單word、IRQ0巢狀時計及IRQ7完整CPU／FPU／堆疊保持已驗，機器層與固定EXE全套通過。首輪測試fixture漏設實模式CS=0後只修測試初態、另驗匯流排錯誤，同命令重跑通過，失敗收據保留。

兩自然用同一固定官方1.31 EXE、fresh417原檔／MOX.SET、Go1.24.13映像、600秒／2GiB／2CPU／128pids／UID:GID1000:1000／network none。不注入CPU／資料／鍵盤／時計／IRQ，不提高50M上限。C6返回outer_step42347255、高位LE0x2454B0的完整R／段／flags206h及DMA狀態保持303，只將音訊時計接到BIOSMicros43985659；初始信用926100、rate22050/1／stereo、block2048、物理14000h／4096byte ring保持。

第42356668步高位LE0x257FC9，完整R依EAX ECX EDX EBX ESP EBP ESI EDI為120 325B80 3258C8 120 2BDB58 2BDB6C 325BA0 325CE1、段8 188 188 0 20 188、flags206h。兩時計44032078，增加46419µs，(926100+46419×44100)整除1000000得2048 samples、餘數4000。DMACompletions1／PCMBytes2048，current4800h／count7FFh、DSP block重載2048，DMA ring未重載；首block含先前實模式21µs合計46440µs，僅為硬體規格近似。

實際PCM2048個80h，SHA-256 88ed1a04cb43fe65827d1cd9ef6d24a736108730b1ce6315d4d3ca79b6a0d140，16byte prefix同80h，獨立雜湊核對符合296的原版靜音buffer。這是已傳輸的原始byte證據，不當人耳驗收。DSPIRQPending／PICPending=true、PICInService=false、IRQ7Deliveries仍1，前一實模式DMA16Completions1／PCM16Bytes2保持，沒有新IRQ7成功派送。IRQ0 active=false／failed=false／started6176／completed6176，BIOS Deliveries6196，等待原始DS:00271148=E6 11 00 00。

平台錯誤明確記錄absolute IVT0F=12010682，即1201:0682、實模式線性0x12692；DPMI實模式0F與DOS保護模式0F皆零。入口原始16bytes為2E FF 06 30 00 E8 5D FD 72 03 E9 95 00 E8 69 FD，只作定位，不追driver／ISR內部。它是未建模的保護模式IRQ7轉送，不是未知CPU opcode。兩PNG同較早已檢視黑圖SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，首次IRQ7在星空繪製前停止；302／303的星空與50M不能作本版終態。主選單未見，等待因果／正常玩家路徑仍未知。

最後只補PCM雜湊與原始入口的唯讀診斷，同命令重生兩自然，與before-pcm-diagnostic收據僅差兩列／解壓mtime／DTA時間日期四bytes，平台／CPU未改。有效兩gzip SHA-256 ec4abf4f565e6cf1ea3ec1b210e414b459d00a6ac8e80da0307e3c323dcb3de0／d89716bc0d1482670ea1eb5cf1709475ef31bbeff0930cb86af34ef407369fda；機器層收據bff247d74ad4200f631ebd7edcccfebed4f72a47f325b7359396fc2a2d6d7c0e，固定EXE全套163b0f2bfa8e8ad2b6efe1f831c7ab35173e1682f82dbf134f4b9fe63c1c79c2。全部來源／輸入／公開平台契約／失敗及有效收據雜湊與命令見[鎖定規格304](https://github.com/wicanr2/dosgolem/blob/089f51f13149dd2a1b2c08c6b24ec639d0c78788/docs/spec/304-le-shared-device-clock.md)。

十一份較早音訊邊界同次回填，46個回填函式／80項缺證據負例／CLI及兩自然完整首block／PCM／pending／原始向量稽核通過。工具089f51f13149dd2a1b2c08c6b24ec639d0c78788已推送隔離分支並回讀一致，不推本機origin。304保持READY，首block時間接線已驗，正式IRQ7轉送／來源確認／EOI／IRET返回仍待窄契約；主庫玩法RE閘門、255／299自然OF=1、正常鍵盤／主選單／正常玩家路徑／受控亂數、人耳及整款remake限制保持。原版素材、完整終端／記憶體／gzip／PNG留本機，不公開原版資產。


## 2026-10-03：IRQ7 正式返回與 XOR AL 的原版 caller 消費

工具基線089f51f13149dd2a1b2c08c6b24ec639d0c78788。305先做明示診斷，兩固定原版從absolute IVT1201:0682／實模式線性0x12692執行73步，實際OUT20h EOI、IN22Eh確認來源及IRET返回。診斷額外1tick與client0100私有堆疊配置明標，與正式收據分開。公開DOS/4GW passdown／私有16位元堆疊與實際入口／出口足以審查READY後，正式用獨立實模式CPU與host allocDOS私有4KiB堆疊重用，保持完整protected caller／FPU／原堆疊，來源確認與EOI由原版I/O完成，失敗後持續明確拒絕。寄存器映射與1µs工具時計是平台近似，不稱DOS/4GW核心或逐週期exact。

正式兩自然第42356668步高位LE0x257FC9，73步返回後原caller下一指令到0x257FCD，兩時計44032151、PCM2051／信用223300。連續14次IRQ7返回、PCM29175／信用60600後，第42488059步高位LE0x247BE1原始34 01 C3停在XOR AL,1。來源／固定EXE／兩正式與診斷收據／失敗fixture及窄驗收見[鎖定規格305](https://github.com/wicanr2/dosgolem/blob/dcf764ed14d1b141948968fd8ebc596ee3dec851/docs/spec/305-moo2-irq7-real-mode-passdown.md)。

306以原始完整R／段／flags297h與公開Intel XOR契約審查READY後，只接裸34 ib，五定義旗標正確、AF清除保留工具近似。全部byte配對／兩初始旗標、目的外24位與完整外層／FPU／記憶體保持、前綴／截短拒絕及固定EXE全套通過。兩自然三組AL 1→0／0→1／1→0，flags297h→246h／202h→202h／297h→246h；原始C3真正讀SS:ESP的DF 1A 23 00回高位LE0x231ADF、ESP加4，caller的ADD ESP,4後，0x231AE2的MOV ESI,EAX實際得到0／1／0。完整R／段／旗標與305前段逐列保持已核對，沒有CPU／遊戲資料或鍵盤注入。

兩自然均到50M上限無未知CPU／平台拒絕，高位LE0x21588F、R=FFFFFFFF 108 178 15A 2BDB70 2BDB90 46 260C2B、段8 188 188 0 20 188、flags246h。兩時計61913470，IRQ7 started386／completed386、IRQ7Deliveries387含較早1次，IRQ0 started8411／completed8411、failed=false。C6信用926100加(61913470−43985659)×44100獨立整除得790617 sample／餘391200，386個2048byte block、ring current4059h／countFA6h、block剩1959一致；PCM只記錄前65536byte，不能外推整段人耳或音訊完成。

兩PNG SHA-256 535e27c45ba579132ef36398335f0c773aa4889473d8336f7180184290e7e975，已實際檢視星空片段，主選單未見。自然兩gzip SHA-256 7a984c68b925623589aa6dae45973ef8431b98b72df65523dbed65eaeb485720／d4526b29dc51b057566604e4538c0adb4c379b06ad005a589e38d59a2a84506e，固定EXE全套43bf923f2d3c8d7d01a9d231fd5f2cc273c9157c945108c197fac3c15bf32bf8。精確來源／輸入／命令／返回消費及範圍見[鎖定規格306](https://github.com/wicanr2/dosgolem/blob/4eb97f6277121e528a2d3ce594dbb1967d16382e/docs/spec/306-cpu386-xor-al-imm8.md)。

十二份音訊邊界與305 XOR停點回填，48個回填函式／132項缺證據負例／CLI與完整兩自然獨立時計／真正RET／MOV消費稽核通過。工具4eb97f6277121e528a2d3ce594dbb1967d16382e已推送隔離分支並回讀一致，不推本機origin。304–306限定CONFORMED，303正常輸入觀測仍DRAFT，原版AH2509的保護模式IRQ1 8:21C4D8已保存而正常鍵盤尚未接通，60／61／64埠讀取零。下一步依公開鍵盤／PIC／DOS/4GW契約保存有界原版入口／返回，達READY後補正常輸入；不以BIOS入隊替代IRQ1、跳指令或提高50M上限，不深入driver／ISR／busy-wait。255／299自然OF=1、等待欄位語意、主選單／正常玩家路徑／受控亂數、人耳及整款remake仍未驗收，主庫玩法RE閘門保持。原版素材與完整收據留本機，不公開原版資產。


## 2026-10-03：正常Esc IRQ1與間接遠呼叫

工具基線4eb97f6277121e528a2d3ce594dbb1967d16382e；原始ZIP／patch／固定1.31 EXE與417檔／MOX.SET沿既有雜湊，Go1.24.13 linux/amd64。路由再次命中平台規格優先／正常dosgolem oracle／跨規格回填，依公開IBM／Intel／OpenWatcom契約，不追driver／ISR內部。

307先DRAFT，在正式50M基線及PNG後執行有界可丟棄診斷，捕捉原版AH2509的8:21C4D8入口。兩診斷第13步拒絕高位LE0x21C4EE的FF 1D DC 42 2A 00，尚未讀鍵盤或改遊戲RAM。308依公開Intel裸FF 1D、同RPL／平坦已知CS契約審查READY後接CPU。真正DS:2A42DC六byte 09 60 32 00 08 01、8byte返回框架F4 C4 21 00 08 00 00 00與完整外層已驗；自製測試實際執行既有CB消費此框架。正式診斷抵達108:326009預設IRQ1鏈時仍拒絕，與BIOS服務證據分開。

公開IBM INT09／環形BDA與無修飾Esc近似下，兩診斷01／81走97／77步，均抵達原始8:21C573 CF；真實遠CALL後只模擬default09服務，遊戲wrapper自行寫2A42AC／AD及2A42E2／E4。證據足以READY後，正式用controller output／PIC pending／in-service與host私有4KiB堆疊，從正常外層入隊；未借用client0501、直接呼叫ISR或注入遊戲欄位。完整caller／FPU／兩CPU模式、IF／遮罩／PIC優先序／IRQ0真實巢狀、來源／向量／框架／descriptor污染、容量／非法碼及失敗持續停止通過。首次PIC接線測試失敗，只依公開契約修20h讀取位置；BIOS診斷首輪編譯誤用Enqueue的error回傳型別，只修探針，均保留收據並乾淨重跑。

**已證實，dosgolem正常入口自行重生**：兩自然固定外層48000000步排controller 01／81；正常48M前所有收據沿無鍵盤308基線逐列保持，僅排除解壓mtime／DTA四bytes。IRQ1 8:21C4D8各97／77步抵達8:21C573，讀60h與20h EOI各一次，started2／completed2。48000000步的原caller高位LE0x215880 ADD EAX,EDX實際得到2C01A8／flags202h、EIP215882；48000001步的MOV AX,[EAX]實際得到2CFFFF／EIP215885，完整其他R／六段保持。

兩自然第48354467步，外層高位LE0x217AD8、IRQ7實模式1201:05DA第77步，真正OUT022C=D0尚未支援。IRQ7 started303／completed302，已經原版EOI／22E後停於新命令；IRQ0 started7927／completed7927、failed=false。兩時計58057009，DMA完成303／block剩2045／信用461100；60h兩次、BIOS字元一次且wrapper已清head／tail。兩PNG SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，實際檢視黑色過場，主選單未見。DPMI RealModeLast是先前已完成封包，不拿它當新D0原版入口。

最終全套SHA-256 4a9eb936bc0c33fe18376e11297e806d994421c44831bcbbd5d070b88b18900a，兩正常gzip 9a8aad65b88d1748431eb8ddfb17733a7a342cf255f4f0147db6a4b49a957753／430349cf9d921862d40f73bce9d2f0e65bb9a36a571ed4b51072ba1688801e59。首輪與最後正常重跑逐列保持；完整caller／遠CALL／CB消費、50回填函式／35個新增缺證據與舊回填負例、兩CLI通過。303／305／306的有限Esc未知與307的FF停點同次回填，其他未知不冒稱閉合。來源／全部有效與失敗收據／命令及CPU／平台近似見[鎖定307](https://github.com/wicanr2/dosgolem/blob/00ad7c645b19b51a8697e2deae85d8a5019dd657/docs/spec/307-moo2-protected-keyboard-irq1.md)與[鎖定308](https://github.com/wicanr2/dosgolem/blob/00ad7c645b19b51a8697e2deae85d8a5019dd657/docs/spec/308-cpu386-call-far-indirect-absolute.md)。

工具00ad7c645b19b51a8697e2deae85d8a5019dd657已推送github隔離分支並回讀一致，未推本機origin。307／308限定CONFORMED；完整鍵盤、255／299自然OF=1、主選單／正常玩家流程／受控亂數、人耳與整款remake未完成。下一步只按公開DSP D0暫停8位DMA契約補平台缺件，再同排程重生，不深入硬體driver／ISR／忙等，不改主庫玩法RE閘門。原始素材與完整RAM／終端／gzip／PNG留本機，不提交或公開。

### 2026-10-03 DSP D0／D4、原版 IRQ7 返回與日期服務新停點

接手主庫87547eafa6ce8e198f845bfe419d6231975d8bf5／工具00ad7c645b19b51a8697e2deae85d8a5019dd657，工作樹乾淨。路由命中平台規格優先、dosgolem對拍與結論回填。原廠Creative Hardware Programming Guide印刷頁6-24／6-26明定D0停止8位DMA請求、D4恢復；309經DRAFT與證據審查READY才實作。只接active8傳輸，保留位置、block剩餘、分數信用與IRQ，兩種CPU時計繼續；idle與其他位寬明確拒絕。沒有改CPU、鍵盤橋接、主庫玩法或猜補遊戲記憶體。時間、FIFO與分數模型是hardware-spec approximation。

全部固定EXE全套通過：Go1.24.13／golang:1.24-bookworm固定映像、UID1000、network none、2GiB／2CPU／128pids、外層600s、417原檔與官方1.31 EXE。四個DMA控制測試通過；首次未使用import編譯失敗，第二次40h信用單位預期錯誤，只修自製測試後以同容器契約重跑，兩失敗收據保留。未放寬斷言或調硬體時鐘。

**已證實，dosgolem正常入口重生**：仍用50M上限與固定48000000步controller Esc 01／81，有／無既有受控滑鼠兩排程。D0前完整307正常基線保持，只排除解壓mtime／DTA四bytes與新觀測欄位。第48354467步真正D0令pause false→true，其餘完整裝置狀態保持；原版實模式IRQ7從1201:0682執行102步到FFFF:FFF0，返回核心與flags保持，started303／completed303。外層高位LE0x217AD8成功續行到0x217ADF；後三條MOV／ADD／IMUL的完整核心消費已驗。

第48782970步原版真正送D4，VirtualMicros58507105；與D0相隔450096µs，來源current4803h／count07FCh、剩2045、credit461100保持。D4只解除pause，原版INT66實模式1201:016A的93步成功返回與caller三條MOV／CMP／JZ到高位LE0x2454E7已驗。返回後21µs真正傳1sample並保留credit387200；至新終態依獨立公式傳2040samples，地址4FFBh／count0004h、剩5、credit483000。PCM快照65536是既有觀測上限，不冒稱總sample或人耳驗收。

兩自然新停點為第48796894步高位LE0x240A32的CD 21，EAX002B2AA8h，DOS AH2Ah日期服務尚未支援；EIP240A34不能當作服務成功。兩時計58553364、IRQ0 started8022／completed8022／failed=false，IRQ7仍303／303。兩PNG與已檢視黑色過場逐位元相同，SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，主選單未見。

全套SHA-256 ad62917158674edccdd3df730566735e16845a757c1fc8ac5f3c5f7d61e52da0；兩正常gzip a33b5da6a93a996cd1cf6653455c39605fc4aaad9e539cf0cefcfce5d6abe28a／07a99506c2c6789cabf3ada4d365431cf977aa37e5509114302bd28708faf25e。獨立審查驗完整前基線／僅pause快照／102步真正IRQ7返回／450096µs來源保持／D4與六條caller／2040samples及兩排程一致。51回填函式與27個新增缺證據／較早標記或連結移除負例、309 CLI通過；303／305／306／307同次追加勘誤。詳細公開來源／輸入／命令／所有有效與失敗收據見[鎖定309](https://github.com/wicanr2/dosgolem/blob/f6bf96a976fb31e19b19545ae438b3abcb2006fa/docs/spec/309-sb16-pause-resume-dma8.md)。

工具f6bf96a976fb31e19b19545ae438b3abcb2006fa已推送github隔離分支、回讀一致且乾淨，未推本機origin。309限定CONFORMED；主庫RE閘門保持，完整鍵盤、255游標／299自然OF=1、人耳、主選單／玩家流程／受控亂數及整款remake未驗收。下一步只核對DOS AH2Ah公開日期契約與可重播時計來源，READY後補平台服務、同一固定Esc排程重生；不取主機即時日期猜補、不跳指令或提高上限。原版素材與完整RAM／終端／gzip／PNG仍只留本機忽略目錄。

新停點的服務分類依[Microsoft MS-DOS 3.3 Programmer’s Reference，Function 2AH](https://www.pcjs.org/documents/books/mspl13/msdos/dosref33/)：AH2Ah取作業系統日期，以CX年、DH月、DL日、AL星期回傳。此處只確認分類，尚未指定日期來源或實作；下一輪須沿可重播平台契約審查。

### 2026-10-03 DOS AH2Ah日期、word SUB與原版日期消費

接手主庫538a1f405cdc0c39c4e2d4dbec89e1010425f8f2／工具f6bf96a976fb31e19b19545ae438b3abcb2006fa。路由載入平台規格優先；標準日期語意直接引用[Microsoft MS-DOS 3.3 Programmer’s Reference，Function 2AH](https://www.pcjs.org/documents/books/mspl13/msdos/dosref33/)，word SUB引用[Intel 80386 Programmer’s Reference，SUB](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/SUB.htm)。兩份規格經DRAFT、證據審查READY、實作與同狀態驗證才標限定CONFORMED。日期是平台規格近似，不追原版DOS driver或以主機今天猜補。

#### 固定輸入與方法

- 原始ZIP SHA-256 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，官方1.31 patch ZIP 908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；417根檔、ORION2.EXE 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，MOX.SET 553 bytes／bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。全部再驗通過。
- Go1.24.13 linux/amd64，golang:1.24-bookworm映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。容器network none、UID1000、2GiB／2CPU／128pids，外層600s；原檔／patch唯讀，新鮮解壓到容器/tmp/game。
- 正式全套DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1。三組go run -buildvcs=false ./workplace/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game，固定DOSGOLEM_MOO2_MAX_STEPS=50000000、DOSGOLEM_MOO2_SEPARATE_DOS=1、DOSGOLEM_MOO2_HARDWARE_ESCAPE_AT_48000000=1。兩組明示DOSGOLEM_MOO2_CALENDAR_EPOCH=1996-01-01，第二組加既有DOSGOLEM_MOO2_MOUSE_EVENT_AFTER_POSITION=1；第三組不設日期。PNG各寫既有workplace/。
- 日期只在AttachMachine後、兩時計零且未執行時設定，零時為UTC午夜；每日86400000000µs，以共用虛擬時計推進。DOS年1980–2099、合法Gregorian日曆、午夜與閏日測試通過。機器／時計不一致、超界或未設定即拒絕，不默默固定日期。CalendarState未設定的0只是觀測占位值，實際裝置時計仍58553364µs。
- 未固定亂數seed，也未宣稱原版與remake亂數同狀態。既有AH2Ch仍為呼叫計數秒占位實作；本切片不宣稱作業系統日期／時間一致。未用測試IRQ1直入、BDA或遊戲欄位注入。

#### 已證實的原版邊界與消費

下列位址均為dosgolem高位LE；SS堆疊位址另標selector:offset，不能當作IDA或檔案偏移。兩組有日期的正式收據在受驗完整核心與裝置狀態一致，之前309基線除明示日期輸入與解壓mtime／DTA四bytes外逐列保持。

| 外層步 | 原版位址／bytes | 真正執行結果 |
|---|---|---|
| 48796894 | 0x240A32，CD 21，AH2Ah | virtualMicros58553364；AL=1、CX=07CCh、DX=0101h，完整其他核心／flags216h保持 |
| 48796895 | 0x240A34，66 81 E9 6C 07 | SUB CX,1900；ECX07CC→0060、flags216→206h |
| 48796896–48796898 | 0x240A39，88 C5；0x240A3B，C1 E1 10；0x240A3E，66 89 D1 | MOV CH,AL、SHL ECX,16、MOV CX,DX：0160→01600000→01600101h |
| 48796899 | 0x240A41，89 4C 24 08 | 只改SS0188:002BDB90四bytes為01 01 60 01，其餘32byte觀測窗／核心保持 |
| 48796930 | 0x240A96，CD 21，AH2Ah | virtualMicros58553400，同一日期、flags246h保持，無午夜跨日 |
| 48796931 | 0x240A98，66 81 E9 6C 07 | ECX07CC→0060、flags246→206h，其他核心保持 |
| 48796932–48796934 | 0x240A9D／0x240A9F／0x240AA2 | 原版同一MOV／SHL／MOV形成01600101h |
| 48796935 | 0x240AA5，89 4C 24 04 | 只改SS0188:002BDB8C四bytes為01 01 60 01，其餘32byte觀測窗／核心保持 |

兩次SUB的CF=0、PF=1、AF=0、ZF=0、SF=0、OF=0以獨立公式檢查，所有65536低word×8暫存器另驗非零高16位保持。原版16條caller完整收據與兩次堆疊寫入已核對；未讀取的外部年份欄位或日期用途仍未知。既有SHL在16位移量下的未定義OF／AF只是平台近似，不以觀測結果證明硬體未定義旗標。

三組PNG SHA-256皆1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，保持先前已實際檢視的黑色過場，不稱主選單。第三組未設定日期仍在第48796894步、0x240A32拒絕，完整核心保持；除新增唯讀拒絕觀測／PNG檔名與mtime／DTA外，完整終端等於309基線。

#### 測試、收據與範圍

| 本機忽略目錄中的收據 | SHA-256 |
|---|---|
| workplace/moo2-311-sub-calendar-tests.txt，六個日期／CPU測試PASS | 8acc8abf35de1678e4c26994c3349036e373c0b7eb4fc955cd83d19c526256b2 |
| workplace/full-test-311.txt，固定EXE全套PASS | ee661ea0e8fb6e464337402fbeb5a15a009a5a9c4c664abc8d313550b12978b2 |
| workplace/moo2-probe-311-full-game.txt.gz | 950f0e6690aad7f7546e2dcdcfb8ddd6aeeb4b398aa001c14a483b359d5ba1f8 |
| workplace/moo2-probe-311-mouse-event.txt.gz | 09b176610ba23b5f6b221564659a3666ed7bd8d2589b2f984a86e3e66ab4bc79 |
| workplace/moo2-probe-311-unconfigured.txt.gz | 5f736d9fa2d296bc138ba0782212e0b775e78d6a258fb569c1d2dcb57eed2589 |

先前日期消費測試因尚缺word SUB失敗，未放寬斷言；補311後同條測試通過。兩310診斷成功讀日期再停SUB，與全套之後的311正式重跑分開。自製探針loopStep宣告位置編譯失敗，只移動宣告後乾淨重跑；全部成功／診斷／可重現失敗雜湊與精確命令在鎖定310／311，最初未保存的編譯stdout不冒稱已有原始收據。53個回填驗證函式、45個新增移除證據／舊標記／連結負例及兩CLI通過，309日期停點與310 SUB停點同次追加勘誤。

兩組有日期的新停點都是第48797763步、0x210C7E，bytes 66 03 05 A4 BE 29 00，word ADD AX,DS:[0x29BEA4]未支援；部分解碼後EIP0x210C81不代表成功。完整R為F／0／2BDCDC／8／2BDBD4／2BDBE0／284324／2BDCA4，六段8／188／188／0／20／188，flags202h。實際來源word與欄位用途未知；下一條66 A3 A2 BE 29 00到DS:[0x29BEA2]尚未執行。兩時計58554306，IRQ0 started8022／completed8022／failed=false；IRQ7 started304／completed304，DMA完成304／剩2011／credit25200／current4025h／count0FDAh。主選單、正常玩家流程、255游標、299自然OF=1、受控亂數、人耳與整款remake仍未驗收，主庫玩法RE閘門保持。

工具ba3239ce0696f2e9bf898b543cee04a8ff455cab已推送github隔離分支、遠端回讀一致且乾淨，不推本機origin。[鎖定310](https://github.com/wicanr2/dosgolem/blob/ba3239ce0696f2e9bf898b543cee04a8ff455cab/docs/spec/310-moo2-dos-calendar-date.md)及[鎖定311](https://github.com/wicanr2/dosgolem/blob/ba3239ce0696f2e9bf898b543cee04a8ff455cab/docs/spec/311-cpu386-sub-word-register-imm16.md)保存限定CONFORMED與來源雜湊；原版素材、完整RAM／終端／gzip／PNG只留本機。

### 2026-10-03 word ADD 來源、真正寫回與原版載入畫面

主庫基線fb7461f138e7a1efc44ad7eb2afc8a83cab1adc8／工具ba3239ce0696f2e9bf898b543cee04a8ff455cab。路由命中平台規格優先、規格流程與結論回填；來源引用[Intel 80386 Programmer’s Reference，ADD](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/ADD.htm)，標準CPU語意不深挖遊戲runtime helper。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原始ZIP／patch／417根檔／MOX.SET雜湊保持並再次核對。工具Go1.24.13／golang:1.24-bookworm固定映像，2GiB／2CPU／128pids／UID1000／network none／600s，原檔及patch唯讀、/tmp/game乾淨組合。

312經DRAFT唯讀診斷，原CPU來源保持311 SHA-256 cfe5bc387aee00907acd897c3c6b77a5e4186828250e53fbc0513d0e40b504e1。正常48M Esc／明示1996-01-01初態在高位LE0x210C7E讀DS0188:0029BEA0窗口0F00000002000200；來源word0002h、AX000Fh，其餘完整核心／RAM／拒絕保持。診斷gzip 8f0c14193c36e21a76b16af706dd7d7a3ce44d72fd1288421c95181d2b97a778，除新增觀測、PNG名稱及mtime／DTA四bytes外完整等於311。公開03 /r與既有32bit effective address／readSegment16足以READY後接word記憶體來源，完整取得才用add16，不改玩家玩法或猜資料用途。

**已證實，dosgolem正常入口重生**：兩明示日期流程第48797763步，0x210C7E的66 03 05 A4 BE 29 00真正得到EAX00000011、flags216h，獨立六旗標CF0／PF1／AF1／ZF0／SF0／OF0及全部其他核心／來源只讀已驗。第48797764步0x210C85的66 A3 A2 BE 29 00只寫DS0188:0029BEA2兩bytes11 00，窗口變0F00110002000200；來源與其餘六bytes／R／六段／flags保持。後續0x210C8B的BB 00 01 00 00令EBX8→100，0x210C90的8B 45 F4觀測EAX11→4；後者SS來源未另存窗口，其欄位用途未知。全部四caller／311前基線、核心／時計／音訊與IRQ7／VBE兩排程一致。

兩排程原先有／無既有受控滑鼠事件，現在各完成1／0回呼，不能聲稱整體初態相同。首次稽核誤要求mouse_started／completed相等，查明只有這兩個計數不同後，依已明示初態各驗1／0，其餘被比較欄位保持嚴格一致，未改CPU／平台或遊戲資料。255完整座標／游標與玩家操作仍未驗收。第三組不設定日曆，完整終端除mtime／DTA與PNG名稱外保持311的0x240A32拒絕，沒有word ADD觀測或默認日期。

兩組新停點第48919460步、高位LE0x14E3DE，bytes66 81 F9 D4 00 0F 8C 45 FF FF FF E8 D0 04 0A 00；word CMP CX,00D4h尚未支援，CX0001h、完整R为FFFFFFFF／1／958／29BE7C／2BDBA4／2BDBD0／2600CD／2BDC2C，六段8／188／188／0／20／188，flags293h。EIP14E3E1是部分解碼，非比較成功；後續0F 8C有號分支尚未執行。兩時計58965328、IRQ0 started8059／completed8059／failed=false，IRQ7 started312／completed312；DMA完成312／剩269／credit95400／current46F3h／count090Ch。裝置irq7_deliveries313含先前16位傳輸，不混用完成數。

兩PNG SHA-256 2d0d564f814e49eab6081862f52467b1a7b232c026d3632fd2c7d97c42545e65，已實際檢視原版Loading Master of Orion II載入畫面與中央游標，主選單仍未驗。VBE Bank7／StartY512／BankSets731／Writes5696552／DisplaySets9。第三PNG保持已檢視黑色過場1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，不把三圖都稱同一結果。

| 本機忽略目錄收據 | SHA-256 |
| --- | --- |
| workplace/moo2-312-add-memory-tests.txt，三新CPU／三既有ADD回歸PASS | ca5f3b75d5f344dd147483f67a9bfdefeee1afe7c5cb8fdef3e75a36bd1afe9d |
| workplace/full-test-312.txt，固定EXE全套PASS | a8ed7290acc8531000959f85d968bfa137237a9d8fa448707838d7f56ccebfc7 |
| workplace/moo2-probe-312-full-game.txt.gz | 35b61654a285c65c619fee17cd49c86720e2fc0def4d405c06552bdfe04d4fc9 |
| workplace/moo2-probe-312-mouse-event.txt.gz | 32274bc6d48f28dad1567c74ee515e3a7867f4a8c12b2764c830c79abe6ebb14 |
| workplace/moo2-probe-312-unconfigured.txt.gz | d2d1475f15c94cccd43c012f98427571475c444548decff883f34a73aa7de904 |

全部八目的／全word／49組邊界與獨立六旗標、完整非零FPU／RAM保持、地址／只讀／拒絕及真正寫回測試通過。首輪自製測試誤用不存在FPU欄位，只改為實際完整欄位後同命令重跑；保留編譯失敗收據，不當作產品玩法缺陷。正式命令沿310，輸出311改312，固定EXE全套後順序三組go run，明示日期、50M上限、48M Esc、有／無既有事件／第三未設定均保持，不啟用私有IRQ1直入。全部54回填函式、31新增缺證據／狀態／舊標記／連結負例及CLI通過，309／310／311同一ADD停點追加勘誤。

工具f2d982a7d9383a2b536d9540cb5b8b9e860f6284已推送github隔離分支、遠端回讀一致且乾淨，未推本機origin。[鎖定312](https://github.com/wicanr2/dosgolem/blob/f2d982a7d9383a2b536d9540cb5b8b9e860f6284/docs/spec/312-cpu386-add-word-memory-source.md)保存限定CONFORMED、來源／工具／輸入、精確命令、全部成功／診斷／失敗雜湊與近似。原始素材與完整RAM／終端／gzip／PNG只留本機。AH2Ch／RNG、255／299自然OF=1、人耳、完整鍵盤、主選單／正常玩家流程與remake同狀態未驗，主庫玩法RE閘門保持。下一步限定word CMP及原有有號分支，不深入helper或提高上限。

### 2026-10-03 word CMP 旗標、JL 兩方向與標題背景

基線主庫b98d6f0e3f11d6a2dba7ab637ad637f5215f9c5e／工具f2d982a7d9383a2b536d9540cb5b8b9e860f6284。路由載入平台規格優先；原始定位沿固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。原始ZIP／patch／417根檔／MOX.SET再驗保持。平台契約引用[Intel CMP](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/CMP.htm)與[Jcc](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/Jcc.htm)，不分析runtime helper或猜CX欄位用途。313 DRAFT／原版與公開契約審查READY後，只補word暫存器group7，完整iw後用sub16改六旗標、來源不寫回；其他word memory／dword／prefix不擴張。

**已證實，dosgolem高位LE的正常入口自行重生**：兩明示1996-01-01流程固定48M Esc／50M上限。第48919460與48919797步的0x14E3DE，66 81 F9 D4 00，CX1／2比212，flags293h→297h；下一0x14E3E3的0F 8C 45 FF FF FF都真正跳0x14E32E。第48992578步CX00D4h，flags293h→246h；下一JL不跳，落0x14E3E9。前兩筆CF1／PF1／AF1／ZF0／SF1／OF0，邊界CF0／PF1／AF0／ZF1／SF0／OF0，以獨立借位／popcount／有號範圍公式核對。全部R／六段／SS0188:002BDBA4的32byte窗口保持，原版目的高word0，非零高word由全word／八來源測試補足。

正常原版六筆CMP／JL、完整312前基線與外層total212／sample_groups3／boundary212_observed=true已核對。不把未捕捉迭代當逐值對拍。兩初態受控滑鼠有／無事件，各完成1／0回呼，其餘受驗核心／音訊／時計／IRQ與VBE一致。第三未設定日曆，除新增唯讀total0、PNG名稱與mtime／DTA四bytes外完整保持312的0x240A32拒絕，沒有默認日期或CMP抽樣。

新停點兩設定皆第49564005步、高位LE0x24C31B的CD 33，AX0014h尚未支援。R為14／1／2136D1／0／2BDA88／2A0000／0／0，六段8／0／8／0／20／188，flags6h；mouse_service完整輸入／輸出與flags保持handled=false，EIP24C31D僅INT fetch，下一C3尚未執行。兩時計61027457，IRQ0 started8251／completed8251／failed=false，IRQ7 started357／completed357；DMA完成357／剩1490／credit984300／current4A2Eh／count05D1h。裝置irq7_deliveries358另含先前16位傳輸，不混為保護模式完成數。

兩PNG SHA-256 5145cdfe5e66f25f9c78f9460256152cfa914d2ac22b17a89dded2b1717f3a37，逐位元相同，已實際檢視原版Master of Orion II標題背景與游標，主選單按鈕未見。VBE Bank2／StartY0／BankSets738／Writes6003978／DisplaySets10。第三圖保持黑色過場1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，不將三圖當作相同結果。

| 本機忽略目錄收據 | SHA-256 |
| --- | --- |
| workplace/moo2-313-cmp-imm16-tests.txt，九個目標／回歸PASS | fd354238907ab735a9bcae1dfd197eae39fd03757751e9b4798338d656f7100f |
| workplace/full-test-313.txt，固定EXE全套PASS | e1bf0db1fb8c96e8f971ed6635b73fd0ad0d5f601f1905480de0f924e380cf73 |
| workplace/moo2-probe-313-full-game.txt.gz | 9ecb69d4db8d3e563ecf0437aa6c94ef20bf81cae8616054f3fcb50429380785 |
| workplace/moo2-probe-313-mouse-event.txt.gz | 024cb32aacccb303f6b58054e277c5a14e989ca79d426f8a25ef02790c40f895 |
| workplace/moo2-probe-313-unconfigured.txt.gz | 1d9d785ac3189d58aab71527bb80f8335655a4d024e28e72b395f6926241e958 |

Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none，原檔／patch唯讀，/tmp/game乾淨組合固定417根檔與官方EXE。命令沿312，輸出313；全套後三組go run，日期／Esc／受控滑鼠條件保持，未啟用私有IRQ1直入。新CMP三測試及ADD／SUB六回歸、八來源全word／49邊界、獨立六旗標、完整非零FPU／核心、真正JL的45組有號／相等／溢位與拒絕全通過。55回填函式、33新增缺證據／狀態／舊標記／連結負例及CLI通過；309–312同一CMP停點追加勘誤。

工具0cc36241a3523df865d9b8f336d70c124bc7093c已推送github隔離分支、遠端回讀一致且乾淨，未推本機origin。[鎖定313](https://github.com/wicanr2/dosgolem/blob/0cc36241a3523df865d9b8f336d70c124bc7093c/docs/spec/313-cpu386-cmp-word-register-imm16.md)保存限定CONFORMED、來源雜湊、公開契約、命令與收據。原始素材與完整RAM／終端／gzip／PNG只留本機。滑鼠AX0014h、255／299自然OF=1、AH2Ch／RNG、人耳、完整鍵盤、主選單／玩家路徑與remake同狀態未驗，主庫玩法RE閘門保持。下一步公開滑鼠API與既有callback儲存，不深挖driver、代寫遊戲資料或提高上限。

### 2026-10-03 滑鼠回呼交換與主選單面板滑入

基線主庫1491008edf177b29d767e25f620de35759e7cfce／工具0cc36241a3523df865d9b8f336d70c124bc7093c。命中平台規格優先，依[Microsoft Mouse原廠手冊，頁107](https://www.bitsavers.org/pdf/microsoft/mouse/Microsoft_Mouse_Programmers_Reference_1989.pdf)的Function20交換／返回舊值契約；[Watcom11.0c，頁129–130](https://openwatcom.org/ftp/archive/11.0c/docs/cprogguide.pdf)只明示0Ch的ES:EDX，不冒稱其已證14h的32位元擴充。314依固定原版實際ES:EDX與既有255模型，審查READY後接平台服務，32位元與佇列時序標為platform-spec approximation。

**已證實，dosgolem高位LE正常入口重生**：固定1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP／patch／417根檔／MOX.SET再核對保持。兩明示1996-01-01／48M Esc／50M上限流程各24次AX0014h成功。前三次舊／新目標皆0008:002136D1，遮罩1→1、1→2B、2B→1，返回CX1／1／2Bh。三次真正C3在49564006／49564388／49602023步由0x24C31D返回0x24C1AE，ESP002BDA88→002BDA8C，R其餘／六段／flags6h保持。原始SS窗口首dword正是返回地址；後四步只觀測框架建立與參數指標讀取，未擷取返回值後續寫回，不深挖helper。

兩組均到step_limit=50000000 eip=0x23856E，無step_error／guest_cpu_stop。時計62461366、IRQ0完成8383，VBE Bank7／StartY512／BankSets815／Writes9385664／DisplaySets21；兩末尾窗口與VBE相同。沒有完整R／六段／IRQ7上限快照。受控初態各0／1回呼且allocator selector／unique_sites不同，差異未抹除。313交換前完整前綴及第三未設定日期全流程，僅mtime／DTA四bytes／PNG名稱正規化後保持。

兩PNG SHA-256 8f7791ae57649991fbf9bf3a86fdacab602e792f39a9b3d57ae691484a754d47，逐位元相同，已實際檢視原版主選單面板從右側部分滑入；完整展開／點擊未驗。第三圖保持黑色過場1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，無日期仍在0x240A32拒絕。

| 本機忽略收據 | SHA-256 |
| --- | --- |
| workplace/moo2-314-mouse-exchange-tests.txt，四新／六回歸PASS | e9baa0f79237394f0d5e217bb71d15033a713cf629ea6bfd98e53eed20fbd6f1 |
| workplace/full-test-314.txt，固定EXE全套PASS | 12df9650c1fa427bce3f0d1e73a9aeda2f81ac905a47187ac0c029218df186fb |
| workplace/moo2-probe-314-full-game.txt.gz | 41923df89d7657f6ceb015cb740c6d426e09d0892ab4968a59f414cb53141b30 |
| workplace/moo2-probe-314-mouse-event.txt.gz | a912b61514f85eb348271958666486ee991b80fc5a6b331a5acbb6ababd4dfe9 |
| workplace/moo2-probe-314-unconfigured.txt.gz | 053d2d831055b737673985b6ddf48ea50e1bf7dd2646b0c56b94de3fe83eedbb |

Go1.24.13固定映像，600s／2GiB／2CPU／128pids／UID1000／network none，原檔／patch唯讀；命令沿313，輸出314，全套後三正常go run。56回填函式、31新負例及CLI通過，309–313同一AX0014h停點同次追加勘誤。護欄首輪因位址文字缺完整冒號形式拒絕，補正文件後通過，程式與正式來源／收據不改。

工具9a2c7a21b0901ac8acc1f4387729ce25072b5fd9已推送github隔離分支，回讀一致且乾淨，未推本機origin。[鎖定314](https://github.com/wicanr2/dosgolem/blob/9a2c7a21b0901ac8acc1f4387729ce25072b5fd9/docs/spec/314-moo2-protected-mouse-callback-exchange.md)保存源碼雜湊、公開契約、精確命令與限定CONFORMED。原始素材／完整RAM／終端／PNG留本機。303整體DRAFT、255完整游標READY、299自然OF=1、AH2Ch／RNG、人耳、完整主選單／玩家路徑及remake同狀態未完成，主庫玩法RE閘門保持。下一步在50M內補主選單滑入的有界唯讀階段觀測，辨識正常等待／輸入條件，不提高上限、代寫資料或跳動畫。

### 2026-10-03 主選單逐換頁觀測與314終態勘誤

基線主庫11573fef0d26c84895bf98c511cf2d8f98fdc4fa／工具9a2c7a21b0901ac8acc1f4387729ce25072b5fd9。路由載入平台規格優先、文件職責與結論回填；固定1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP／patch／417根檔／MOX.SET再驗保持。315只補唯讀快照，不改Go remake、CPU或平台。

**追加勘誤，已證實**：上一節314所寫「沒有完整R／六段／IRQ7上限快照」錯誤。其兩原始gzip實際含late_startup_platform label=terminal、irq7_passdown_state label=terminal及step_limit_registers；前輪只搜尋limit／step_limit標籤漏讀terminal。本輪完整核對R3530C4／0／F3／1／2BDB10／2BDB68／4F6F42／353316，六段8／188／188／0／20／188，flags206h；IRQ7 started388／completed388，DMA完成388、剩1742、credit371200。裝置irq7_deliveries389另含先前16位傳輸，不混用。完整兩終態除既有受控mouse_started／completed0／1外逐列相同，allocator selector／unique_sites差異仍保留，不把不同初態說成同初態。工具314現行斷言與回填已修正，舊收據及此處歷史索引保持。

315以明示DOSGOLEM_MOO2_VBE_FRAME_PREFIX在outer_step≥48M、真正INT10／AX4F07成功換頁後擷取最多16張，並用完整R／段／EIP／flags／FPU／VBE讀取前後核對不突變。Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none；命令沿314，移除來源未改且已PASS的全套測試、輸出315並明示各條件快照prefix。兩epoch1996-01-01／48M Esc／50M cap，有／無既有事件分開，第三不設日期，原檔與patch唯讀。

三正式流程完成，兩組各14張第8..21換頁、第三1張，共29張，全readonly=true。逐張PNG雜湊與兩序列完整CPU／段／flags／時計／VBE逐列核對。只移除新增menu_slide_phase列，再按既有mtime／DTA四bytes／終圖檔名正規化後，三流程每一既有314列完全相同。主選單尚未完整展開，沒有新正常點擊驗收。

已實際檢視10／16／21畫面：49512086步／60909981µs為標題背景，49756873步／61670846µs為右側邊緣，49967220步／62395438µs為部分主選單。固定像素區域x=[500,640)、y=[110,350)，相對第10張RGBA逐點量測，末五張17..21變動區左緣619／608／596／583／571，變動點3852／6193／8718／11726／14364。支持到50M動畫仍有進展的強推論，不用量圖值補玩法規則。第21換頁圖0e5f213ed365098b07e2fe92bc5d8259fc6a13081b413d51c3edb1bc0293db4f與終圖8f7791ae57649991fbf9bf3a86fdacab602e792f39a9b3d57ae691484a754d47為不同取樣時點；不混成同狀態。

| 本機忽略收據 | SHA-256 |
| --- | --- |
| workplace/moo2-probe-315-full-game.txt.gz | 91ff5b147572fc6a7dbe706c3e19f832fdd3ab61b8c7ed6a0e2c29a9cc6ba779 |
| workplace/moo2-probe-315-mouse-event.txt.gz | 6d53ab8d999c22f89814fa2793316a59ffada212f6fe14639cfc904e270f9926 |
| workplace/moo2-probe-315-unconfigured.txt.gz | 4895e5e54aa52d12d0333f1fc1cec86bfeb5121957dbadfc0316ee9962f7a701 |

輔助Go量測首輪asm fork受64pids限制，明示GOMAXPROCS2／go run -p2後同映像乾淨重跑通過，屬工具環境；原版流程與素材不改。57回填函式、21缺定位／終態／收據／314勘誤負例與CLI通過。新probe SHA-256 546c234534a8f82e3a5e37de82e95fc46d220af0e428184ac2eafbaf64fc530d，CPU／平台source與314保持。

工具f1c2fa57675991080e2da3d1f5008b9b49f209bc已推送github隔離分支，回讀一致且乾淨，未推本機origin。[鎖定315](https://github.com/wicanr2/dosgolem/blob/f1c2fa57675991080e2da3d1f5008b9b49f209bc/docs/spec/315-moo2-menu-slide-observation.md)保存限定CONFORMED、逐幀定位／工具、命令／收據與314勘誤。原素材／快照／完整終端留本機；303整體DRAFT、255游標READY、299自然OF=1、AH2Ch／RNG、人耳、完整主選單／玩家路徑與remake同狀態未完成，主庫玩法RE閘門保持。下一步先定明示46M Esc獨立排程契約，再沿既有IRQ1驗正常輸入，與48M基線分開，不提高50M上限或代寫／跳動畫。

### 2026-10-03 明示46M Esc與原版主選單按鈕

基線主庫655dd1832a76200f02e659400757aa9b67b5941e／工具f1c2fa57675991080e2da3d1f5008b9b49f209bc。平台規格優先，沿307已證的控制器／原版IRQ1；316審查READY後只改探針輸入參數與快照門檻，CPU／平台／主庫玩法不改。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；原ZIP／patch／417根檔／MOX.SET再驗保持。

**已證實，dosgolem高位LE正常入口**：預先固定DOSGOLEM_MOO2_HARDWARE_ESCAPE_STEP=46000000，日期1996-01-01零時，50M cap，兩初態有／無既有受控滑鼠分開。46000000步0x257BC3排01／81，原版8:21C4D8各97／77步CF返回8:21C573，實際60h讀取與20h EOI；正常原caller續行257BC8／257BCA。排隊前每一原始列與315對應初態前綴保持，沒有宿主代寫欄位或直接呼叫handler。

兩46M到50M高位LE0x2385AC，無未支援指令，完整R／六段／flags246h及裝置相同，時計64282188、IRQ0完成8401、IRQ7完成427。有／無既有受控回呼1／0與allocator selector差異保留，與48M不同輸入時點不冒稱同初態。VBE Bank2／StartY0／BankSets848／Writes15212758／DisplaySets40；兩圖逐位元相同87fabf21f7be22d264f83390bdb4d39f09098516191c3a17c51815452856645e，已實際檢視CONTINUE／LOAD GAME／NEW GAME／MULTI PLAYER／HALL OF FAME／QUIT GAME六按鈕文字完整可見，動畫停穩、點擊與新遊戲未驗。完整第40換頁步數未取樣，不用終圖猜最早可點時點。

11真實CLI負例在讀不存在EXE前exit2，stdout空，含範圍／格式／溢位／8M預設超界及新舊開關互斥。四正常流程完成：兩46M／日期、第三48M／無日期，第四48M／日期／無事件legacy-baseline。48M與無日期的每一原始列及終圖完整保持315，僅mtime／DTA四bytes／PNG路徑正規化。47張快照全部readonly=true、雜湊核對；46M第8..21的索引／RGB／PNG與315同頁保持，時計與步數分開。

| 本機忽略收據 | SHA-256 |
| --- | --- |
| workplace/moo2-316-escape-config-tests.txt | 55f6fa2f53c58f507b95dd9aa62a7e935da4050e0649da4d9a3fd01c041f8379 |
| workplace/moo2-probe-316-full-game.txt.gz | 06d0fe6d2cb2e4bfa676ab67a514f5165a33e4cfc6d458e83975976626108209 |
| workplace/moo2-probe-316-mouse-event.txt.gz | c820da5c2d3e2f11be4d1fe09932d2880210a7d3946c13b8a6183b7ba78d337a |
| workplace/moo2-probe-316-unconfigured.txt.gz | e4cef3b724b933508b331a6746c3ec11bbfc7464d126a30f14b76801329363a3 |
| workplace/moo2-probe-316-legacy-baseline.txt.gz | 32306c41cd16d801f2dd99cca5c03b7584e1d2406eabf1793d8e0c6412d83a60 |

Go1.24.13固定映像、600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。命令沿315、輸出316，先建probe驗11拒絕，再四go run；CPU／平台未改，314全套PASS沿用，不重跑硬體語料。58回填函式／26缺證據負例及CLI通過，315追加新排程收據，保留其原48M部分滑入。新probe SHA-256 57399685a537099ed8871151e9d79d07c4570d94a1f3f498a77efcd0e60efdaa。

工具cbc63f6ad19f17ca8eda81c1ab44f6852fb383e9已推送github隔離分支，回讀一致且乾淨，未推本機origin。[鎖定316](https://github.com/wicanr2/dosgolem/blob/cbc63f6ad19f17ca8eda81c1ab44f6852fb383e9/docs/spec/316-moo2-configured-hardware-escape-schedule.md)保存限定CONFORMED、精確命令／來源／收據與邊界。原素材／完整RAM／終端／PNG留本機；255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG、人耳、主選單點擊／正常玩家路徑與remake同狀態未完成，主庫玩法RE閘門保持。下一步先核對第40換頁的實際步數與NEW GAME可點條件，定單次正常滑鼠輸入READY契約，保留46M／48M基線與50M cap，不猜熱區／代寫資料或跳動畫。

### 2026-10-03 原版選單正常滑鼠與CB寫後

基線主庫df8d8792a482f817f1a7b38bd08ba8b1c1845c91／工具cbc63f6ad19f17ca8eda81c1ab44f6852fb383e9。路由平台規格優先，317–319先READY再改自製探針；CPU／平台與主庫Go玩法保持。固定1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP／patch／417根檔／MOX.SET重新核對。

**已證實，dosgolem高位LE正常輸入**：46M Esc／1996-01-01／50M，無早期滑鼠事件。第40換頁49882420／63906833µs、callsite228CA7、PNG dc938ac71e2a617ec9a6c2d75029f689de6aaf985530a030713195e0c566b15d，實際檢視六按鈕全見。剩117580步，不假設動畫或熱區已驗。

沿255原版CX右移一位的已證消費，以畫面500／229送正常API1000／229／buttons1，outer49882421、caller228CA9；32350µs後49883408／caller2354D1送1002／229／buttons0。原版8:2136D1兩入口AX3／5、BX1／0、CX3E8／3EA、DXE5，真正2137F2的CB各於49882522／49883496返回原caller。191步無錯，完整R／六段／flags246h／207h恢復。原始213741的66 A3 1A 12 2A 00目的2A121A，319兩CB後兩byte實際0100／0000；保留地址，不補主選單consumer名稱。

**未知**：50M點擊終態2385AF、時計64282188，已檢視PNG 0c45ba73ddfe850693520f5aee093c1c188ab118df9c51570521cfe3fc0ddad9仍主選單，只有NEW GAME上的游標可見。mask1不證按鈕啟動。319目標讀取零筆，覆蓋沒有真實正對照，不推論按住期間沒有讀取；直接Bus與實際新遊戲設定尚未驗。

319首輪掛勾在startup前安裝且要求委派ok，經原始程式核對才確認startup覆蓋、普通RAM false後走Bus；保留原始控制收據。修正為按下時接現存8／16位元委派與OR鏈、不改Bus身分後，同映像／同命令／同輸入重跑，仍不把零取樣當成不存在consumer。這是觀測工具缺口；原版兩輪均正常到cap，沒有CPU或產品失敗。

| 本機忽略原始收據 | SHA-256 |
| --- | --- |
| workplace/moo2-probe-317-full-game.txt.gz | 35e2604c0d223e7170e6774dee33c67828992269d75ab41bb0c7607c26c1d0b7 |
| workplace/moo2-probe-318-click.txt.gz | f49d8ab7c7cc6c88c7229d0dfac0d23e01d4d02da4e286086793cba217ab744a |
| workplace/moo2-probe-318-baseline.txt.gz | f3f80f78409e490f1c4a0a0d1ffca00194eab0b0b614d7a61d1f436e3573ab24 |
| workplace/moo2-probe-319-segment8-control.txt.gz | eb850aa6e18dc4faf47134368dd4289962a2b9153cef24cf1ff940c5a31ab617 |
| workplace/moo2-probe-319-click.txt.gz | 597266974b32440f35818cc672464573202bb997ebcf3dd29a7d2fd65bf8ac9e |

316全部舊列與終圖保持317；無點擊模式全部317列與終圖保持；點擊前完整前綴保持，修正319全部318列與終圖保持。mtime／DTA四bytes／PNG路徑是唯一正規化，原始RAM／allocator差異不抹掉。59回填函式／32缺證據負例與CLI通過。

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後/tmp/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game；46M Esc／日期／50M／separate DOS固定，點擊模式只加DOSGOLEM_MOO2_NEW_GAME_CLICK_AFTER_DISPLAY40=1。CPU／平台未改，314固定EXE全套PASS仍適用，沒有以工程自洽冒稱原版玩法parity。最終probe SHA-256 2cab77b5f3ddcb4a6dcc024e4a0b07462ee63a8be660ab662d71fb46ee4f6a10。

工具062670051203076ff688d36a390f46dd8a7883c6推送github並回讀一致，未推本機origin。[鎖定317](https://github.com/wicanr2/dosgolem/blob/062670051203076ff688d36a390f46dd8a7883c6/docs/spec/317-moo2-menu-display40-observation.md)、[鎖定318](https://github.com/wicanr2/dosgolem/blob/062670051203076ff688d36a390f46dd8a7883c6/docs/spec/318-moo2-new-game-normal-click.md)、[鎖定319](https://github.com/wicanr2/dosgolem/blob/062670051203076ff688d36a390f46dd8a7883c6/docs/spec/319-moo2-new-game-button-consumer.md)保存原始定位、精確返回／寫後及明示未知。主庫玩法RE閘門保持，完整新遊戲／正常玩家路徑／remake同狀態未完成；下一步先補真正讀取請求數與路徑正對照，不反覆換鍵、點擊或提高cap。

### 2026-10-03 真正事件消費與DOS問號搜尋

本輪接續上一輪實質進度。起點主庫615e96781fba559c09e203ac94755df1641e34f5／工具062670051203076ff688d36a390f46dd8a7883c6。命中平台規格優先、結論回填與文件職責路由。固定EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP／官方patch／417根檔／MOX.SET重新核對保持；原版完整LOG／RAM／PNG／素材留本機忽略工作區。主庫Go玩法不改，RE-first閘門保持。

**追加勘誤，已證實**：319零筆讀取不表示沒有按鍵consumer。320在同46M輸入中取得normal8=103537／normal16=11288、callback8=59／callback16=15、兩寬度各兩正對照與掛勾code保持。第一byte正對照是派送器validTarget，不能泛稱客體指令；真正高位LE234AC5／234AD5的word MOV讀500／229。CPU dword讀取先嘗試byte委派，不能由掛勾width推客體寬度。

321沿相同輸入，兩CB原始2A121A..2A1229保持按下事件與座標。原版放開後七次正常讀取、28步續行；49895675的高位LE213C60讀2A1228=1，49895677的213C69清0。213BD9／213C06取500／229，213EC5取dword10001再SAR16／CMP／JG。已證短按下事件保留並被消費，沒有延長按住或重點；未補欄位高層用途。320全部既有列與終圖保持，僅新增事件列與目標計數0→7。

322預先固定獨立44M Esc／1996-01-01／50M cap，八真實CLI閘門與兩正常原版流程通過。正常controller／97與77步IRQ1及兩CB保持。第40換頁47850591／59121761µs；47850592按下、47851578移動放開，相隔32370µs，兩CB於47850693／47851666返回、191步。44M和46M是不同初態，雖同頁PNG相同仍不冒稱same-state。點擊至47995790，高位LE229A59／CD21真正AH4Eh／CX0搜尋DS188:261692的save?.gam，DTA188:295828；先前save10.gam成功。這是問號樣式未支援，不能說所有AH4Eh未接。正版根層417檔只有SAVE10.GAM／208000bytes，沒有捏造測試遊戲素材。

323依[公開DOSBox-X 8.3匹配契約](https://raw.githubusercontent.com/joncampbell123/dosbox-x/master/src/dos/drives.cpp)READY後延伸MOO2服務，只支援ASCII 8.3問號／CX0／根目錄或單一目前目錄前綴，安全邊界沿os.Root。列舉實際根檔再匹配，問號可匹配短名剩餘空位；SAVE?.GAM不匹配SAVE10.GAM。成功DTA寫實際候選，樣式區保留問號；未支援provider與其他輸入拒絕。排序、UTC日期與DOS保留區為明示platform-spec approximation，FindNext／星號仍未支援，不稱原FAT或完整DTA原版逐byte相同。

**已證實，dosgolem高位LE正常路徑**：323真正save?.gam查無匹配；47995790步R264E92／0／261692／295828／2BDB60／2BDB78／FFFFFFFF／2BDC2C只EAX→12h，六段8／188／188／0／20／188保持、flags246h→247h。DTA首12bytes由02534156453130000047414D→02534156453F00000047414D，+0Ch..+2Ah仍保留先前SAVE10.GAM結果。12個原始caller步全部無錯，24000C的73 0E因CF=1不跳，24000E的AND保存AX12h。只保存此最小輸入／返回／消費，不追Watcom檔案helper內部。

無點擊全部322列與終圖保持，點擊至問號搜尋前完整前綴保持；只正規化檔案mtime、DTA時間日期四bytes與PNG路徑。正常點擊後再進1446293步，新停點49442083、高位LE17122B／00 C3 0F BF C2 42 00 1C 06 66 83 FA 08 7D 11 EB，opcode00未支援。錯誤後EIP17122C；完整R0／0／1／0／2BDB50／2BDB84／2BDB68／2BDB68、六段8／188／188／0／20／188、flags247h，時計64287801，IRQ7完成427、正常mouse兩次完成。已實際檢視終圖仍主選單，設定畫面仍未知。

| 本機忽略收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-320-click.txt.gz | 88d0b595b250003fd11e4c1f311a23abac74cbe2605f80ebd7f4078cbb6b298c |
| workplace/dosgolem/workplace/moo2-probe-321-click.txt.gz | cf8392ef4e849c3ec268a6742fe46ee6d843850e56298d3afffafeaa7bf7e30e |
| workplace/dosgolem/workplace/moo2-probe-322-baseline.txt.gz | 417992b4c93c9092cd4366bea6ab43a5b53eeabf898c1d20f07697b4e56b7d6e |
| workplace/dosgolem/workplace/moo2-probe-322-click.txt.gz | 3285597c857f30c6e542cef70ee0a8568e4725e05f2de0c8d7ed90bcbc9ff7fb |
| workplace/dosgolem/workplace/moo2-probe-323-baseline.txt.gz | 966a964554a3fe5ca744f792baf492d1c3dc4e03a902c952174059ac55da1d46 |
| workplace/dosgolem/workplace/moo2-probe-323-click.txt.gz | 62b3344c95ed7215ff3602d4399129ba5e58d923f89585ecb98271d856e78ec3 |
| workplace/dosgolem/workplace/full-test-323.txt | a5b1ee8b15de471a587768f875ee5cb1e2761684e66c7d5a5e361dd5e5b53e8a |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch只讀。建置／原版命令沿322，只換323輸出前綴；兩44M初態、日期、50M、separate DOS預先固定，只有點擊模式加正常點擊旗標。固定EXE DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1全套通過，CPU386194.255s、machine6.660s。定向搜尋測試先修正測試包裝型別後乾淨重跑，屬測試工具問題。60回填函式、原有32與新49負例及兩CLI通過。CPU來源保持538abd53a40d65cc07b33cbeaf272d3541cd6dd4622c4fa17a5d8244a0521a77；323 startup dbdbba06c9616547ad7beaec251f4aceb8a9b07d44ed6ffab2e53911ec7e972c，probe feb3b80caa1b220aa6ee3c47856bebb70dda40d1b9a47b45d03186b955ffdbc4。

工具e75f5aebed41cccb062609235dd2f0b07ffae371已推送github隔離分支並回讀一致、工作樹乾淨，未推本機origin。[鎖定320](https://github.com/wicanr2/dosgolem/blob/e75f5aebed41cccb062609235dd2f0b07ffae371/docs/spec/320-moo2-button-read-hook-control.md)、[鎖定321](https://github.com/wicanr2/dosgolem/blob/e75f5aebed41cccb062609235dd2f0b07ffae371/docs/spec/321-moo2-menu-mouse-event-consumer.md)、[鎖定322](https://github.com/wicanr2/dosgolem/blob/e75f5aebed41cccb062609235dd2f0b07ffae371/docs/spec/322-moo2-earlier-escape-new-game-continuation.md)、[鎖定323](https://github.com/wicanr2/dosgolem/blob/e75f5aebed41cccb062609235dd2f0b07ffae371/docs/spec/323-moo2-dos-findfirst-question-pattern.md)保存中間與最終來源雜湊、原始定位、限定CONFORMED、命令／收據／近似與未知。完整新遊戲、正常玩家路徑與remake同狀態未完成；下一步依公開CPU契約補00 C3的byte ADD，先READY再實作、全CPU回歸、同44M正常單次輸入重生與有界consumer驗證，不提高cap或重新開已驗CB／FindFirst。

### 2026-10-03 byte ADD與真正記憶體消費

上一輪152bd136b0731b447b838528bf453e267a55727d／工具e75f5aebed41cccb062609235dd2f0b07ffae371為實質進度。本輪命中平台規格優先、文件職責與結論回填，324先DRAFT／READY再改通用CPU與唯讀探針；主庫Go玩法不改。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，正版417根檔與MOX.SET、同44M Esc／1996-01-01／50M cap及單次正常NEW GAME輸入保持，沒有新種子、代寫、直接進入helper或加cap。

依[Intel 80386原始ADD契約](https://people.freebsd.org/~jhb/386htm/ADD.htm)，無前綴00 /r同時支援八byte暫存器目的／來源與32位一般位址、DS／SS、disp8／disp32／SIB／no-base的記憶體目的。目的只改一byte；記憶體寫入成功後才發布六旗標。沿既有add8與decodeAddress32，不複製原版helper控制流或猜欄位名，未審查prefix仍拒絕。

**已證實，dosgolem高位LE正常原版**：49442083步17122B／00 C3到17122D，R0／0／1／0／2BDB50／2BDB84／2BDB68／2BDB68與六段8／188／188／0／20／188保持，BL0＋AL0仍0，flags247h→246h。下一MOVSX EAX,DX只EAX0→1，INC EDX只EDX1→2、flags246h→202h。七次171231／00 1C 06到171234，實際DS188:[ESI+EAX]為2BDB69..2BDB6F，BL皆0，目的D5／D4／D5／D5／D5／D5／D5及五byte鄰接窗口保持；六旗標以獨立較寬和、低nibble進位、有號範圍與popcount逐筆核對。這次沒有原版非零寫入樣本，非零由完整平台測試覆蓋，不稱另一個遊戲初態也已對拍。

24原始續行全部無錯；七個171234／66 83 FA 08的CMP DX,8六旗標及七個171238／7D 11的JGE核對，六次不跳、一次跳，最後DX8到17124B。沒有用單一分支或自製fixture代替正常原版入口。兩CB仍started2／completed2、191步、pending0／activefalse；原版問號搜尋與先前事件契約保持。

無點擊全部323列與終圖、點擊舊拒絕前4089列及第一ADD完整輸入保持，只正規化mtime／DTA時間日期四bytes／PNG路徑。首輪稽核誤以guest_cpu_stop作前綴切點，包含其前四個失敗後快照；第一差異正是label=stop。修正到首次失敗快照前並另核對第一ADD核心，乾淨重跑PASS，沒有修改正式來源／收據或放寬資料比較。

正常再前進59052步，新拒絕49501135，高位LE2130F3／F7 5D D8 8B 45 D8 66 3B 45 E0 0F 8D A8 00 00 00，F7／3記憶體NEG尚未支援；fetch後EIP2130F5不表示NEG成功。完整RFFFFFFFF／498AC0／8／47／2BD9B8／2BD9F4／495230／2BDACC，六段8／188／188／0／20／188，flags286h，時計64507253、IRQ7完成432。標準ModR/M對應SS:[EBP-28h]／SS188:2BD9CC，來源RAM未取樣，不能由EAXFFFFFFFF推測。

| 本機忽略收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/byte-add-324-tests.txt | a48db475b0157766892c6a47f7d333682b4575457fcefa954df4b7968f5a0073 |
| workplace/dosgolem/workplace/full-test-324.txt | dd00bcf329a9841fa4810d94a30f90c76245bcee66c0718c6073a6862464656a |
| workplace/dosgolem/workplace/moo2-probe-324-baseline.txt.gz | 6130a9fe5bf084b2128c700c9e10fb821d43bd6ee8a474933015e8272d7eb0e1 |
| workplace/dosgolem/workplace/moo2-probe-324-click.txt.gz | c9e8b2d3d15c4b8a64fb9ff497513b2a294b67b9a682b5584dc9156d873bf1e3 |
| workplace/dosgolem/workplace/byte-add-324-backlink-tests.txt | 7e134f37db5edeb23ee21a6258a1e5f0655654d2d29663080460d4ed43f1554b |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestByteADD00|TestByteRegisterANDADDAndAliases|TestByteADDRegisterImmediate' -count=1 -v PASS，七新增主測試、所有byte對與64暫存器別名、ModR/M／SIB／DS／SS／段末／地址繞回／Bus／prefix拒絕與R／六段／FPU／鄰接RAM通過。固定EXE DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1全套PASS，CPU386195.653s／machine6.976s。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後兩正常流程，環境沿323只換324輸出名稱。61回填函式、原有32／49與新增25負例及兩CLI通過。

CPU來源SHA-256 d3fd7c1125d6ecda2fbc05f021776a0532a820af6babea1f3693dc6608aace4e，probe118203d32391772177dd96e6da2539886618a73bb76f14ec8e161399dc98a985，測試10b6297576161f93de3c7f9b9dc6388b02e2fa1c1d985ff2404a22c8923bc58a。DOS startup／provider／matcher保持323。兩終PNG與已實際檢視的323逐位元相同，點擊仍59f76749db5232f97a6b6f969f56f848c2e89e731d7e3fb30b8786982bca4a81、主選單，設定畫面仍未知。完整LOG／PNG／RAM／原素材留本機，不進Git。

工具f31793166415705c75a81339808372da24cded44已推送github隔離分支並回讀一致、工作樹乾淨，未推本機origin。[鎖定324](https://github.com/wicanr2/dosgolem/blob/f31793166415705c75a81339808372da24cded44/docs/spec/324-cpu386-add-byte-register-memory.md)保存限定CONFORMED、正式輸入／窗口／兩方向、精確來源與未知，323停點同次回填。不將標準CPU工具能力加入玩法分母。下一步依公開NEG契約READY／實作／固定EXE全套與同44M正常單次原版重生，核對實際SS來源與最小caller，不追helper內部、不重點、代寫或加cap。主庫玩法RE閘門及整款remake／中文化目標保持，正常開局／remake同狀態未完成。

### 2026-10-03 dword記憶體NEG與原版真正寫回

上一輪主庫ea1ff9d5350794c75a50fe8bef3adbf3265f161c／工具f31793166415705c75a81339808372da24cded44為已驗ADD進度。本輪命中平台規格優先及文件職責，325先DRAFT／READY，再修改隔離工具通用CPU與原版唯讀探針；主庫Go玩法RE-first閘門保持。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、原417根檔與MOX.SET、同44M Esc／1996-01-01／50M cap和單次正常NEW GAME輸入不改。

依據[Intel 80386原始NEG契約](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/NEG.htm)及[附錄C的六旗標表](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/appc.htm)，只補無前綴F7 /3 mod0..2的32位記憶體目的，沿既有DS／SS、ModR/M／SIB與描述子。完整目的寫回後才發布flags；沿既有sequential Bus近似，晚期寫錯可留下已寫byte前綴，flags／R不發布，不稱全CPU例外或rollback。未審查prefix保持拒絕。原ESP測試「wrong SIB」實為合法EBP形狀，325明示擴充、改驗LOCK；084／180／324與索引、守衛同次回填，歷史FD2及324未取樣收據保留。

**已證實，dosgolem高位LE正常原版**：49501135／49520331／49576598步的2130F3／F7 5D D8皆到2130F6。實際SS188:2BD9CC來源FFFFFFFFh→1h、flags286h→213h，八R與六段保持；16byte窗口從SS188:2BD9C8起，僅目的四byte改變，SS188:2BD9D4比較word為0008h。完整RFFFFFFFF／498AC0／8／47／2BD9B8／2BD9F4／495230／2BDACC，六段8／188／188／0／20／188；來源由真正RAM取樣，沒有從EAX推測。

九原版續行全部無錯：2130F6／8B 45 D8的MOV EAX,SS:[EBP-28h]把EAX變1、flags213h保持；2130F9／66 3B 45 E0的CMP AX,SS:[EBP-20h]以0001h比較0008h，flags213h→297h，其餘核心與窗口保持；2130FD／0F 8D A8 00 00 00的JGE三次不跳至213103。跳轉方向未實際發生，仍未知，不稱另一個遊戲初態已對拍。

無點擊全部324列／終圖保持；點擊舊首個失敗後快照前4123列與第一NEG完整核心保持，僅mtime／DTA日期時間四bytes／PNG路徑正規化。兩CB仍191步、started2／completed2、pending0／activefalse。點擊到50M上限，高位LE21334F／88 45 F8 EB 8C FF 45 E8 0F BF 45 D0 01 45 CC E9，完整R3DD302／0／3DA3D4／42／2BD948／2BD998／4953AA／2BDACC、六段保持、flags216h、26937個unique_sites。沒有新CPU拒絕；這是有界執行終點，不猜成正常開局完成或產品卡死。無點擊仍238573／24693個sites。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/neg-dword-325-tests.txt | a48cd9c18945224c5610961b38058ade0f272b3016471b7d57e6cec6ceab8520 |
| workplace/dosgolem/workplace/full-test-325.txt | f98aa39690cc448441483c927be79f48d73dc943970819136b4d7e8ace851d03 |
| workplace/dosgolem/workplace/moo2-probe-325-baseline.txt.gz | 42adfe85b088d25bb3a4e3856d7f8c2187f001253759b9f194c4e123572c552f |
| workplace/dosgolem/workplace/moo2-probe-325-click.txt.gz | 5c269c7fbd6360ad5248da763e2f57f302ea3b837e48a41d1896cb73a16a1d7b |
| workplace/dosgolem/workplace/neg-dword-325-backlink-tests.txt | 02e589623a7b22418ead00bb74525b1622d6d8c1ab3c26cd5497f3559ffe8628 |
| workplace/dosgolem/workplace/neg-dword-325-parity-tests.txt | b7bd5bca62c91c69c0d784d57e7e1b7177f6acb0bb429c1a287ef28a4e645a84 |
| workplace/dosgolem/workplace/neg-dword-325-verify.py | f13094e9dfc2c75bd089a60b6cbb4f35f459eaf1840a2979a78be53695ea0e83 |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。定向go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestNegDword|TestNegStackDisp8Dword|TestNegRegister32|TestNegByte' -count=1 -v PASS，六新增主測試覆蓋獨立flags、值域與所有ModR/M／SIB／DS／SS／每byte讀寫失敗／界限／繞回／prefix／截短與R／六段／FPU／鄰接RAM保持。固定EXE DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1全套PASS，CPU386143.926s／machine3.148s。先go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再兩正常流程，完整環境與來源見鎖定325；本機可重播python3 workplace/neg-dword-325-verify.py獨立驗真正來源／寫回／flags／最小消費及前綴／終圖。62回填函式、原有32／49／25與325新增27負例、三CLI通過。

CPU來源SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1，probe2f58792f30d6fea061861baacef317f55a0eb503b5a0d1acbec0fca131eb8844，測試ab4330e73511963f050c82d086384cb0095658e9ed9fe200070d314bb42aa436。startup、read-only provider與問號matcher保持324。兩終圖與實際檢視的323／324逐位元相同：無點擊11ec0ed15a4c874d36c94a824af73eb71dc6937dfe3bd568450943861db89927、點擊59f76749db5232f97a6b6f969f56f848c2e89e731d7e3fb30b8786982bca4a81；仍主選單，設定畫面仍未知。原素材／完整LOG／PNG／RAM留本機，不進Git。

工具31ca939e73203bafa2f55d7d23ae6115e9dc54c2已推送github隔離分支、回讀一致與工作樹乾淨，未推本機origin。[鎖定325](https://github.com/wicanr2/dosgolem/blob/31ca939e73203bafa2f55d7d23ae6115e9dc54c2/docs/spec/325-cpu386-neg-dword-memory.md)保存限定CONFORMED、正式來源與未知。下一步在同44M單次正常輸入與50M上限下增加後段唯讀、有界進度／畫面觀測，核對最小阻塞；不追helper內部、不加cap、重點、代寫或先調整輸入。主庫玩法RE閘門及整款remake／中文化目標保持，正常開局／remake同狀態、AH2Ch／RNG與人耳未完成。

### 2026-10-03 NEW GAME後段唯讀進度與實際來源定位

上一輪主庫6cc6dc6dbc2610a041394bfda71bb829c3f4e3e3／工具31ca939e73203bafa2f55d7d23ae6115e9dc54c2已驗NEG。本輪命中平台規格優先與文件職責，326先DRAFT／READY，再只改原版探針唯讀觀測；CPU／平台及主庫玩法保持。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、原417根檔／MOX.SET、同44M Esc／1996-01-01／50M cap／separate DOS與單次正常NEW GAME輸入不改。

**已證實，dosgolem高位LE正常原版**：49500000／49600000／49700000／49800000／49900000／50000000六時點全為readonly=true；CPU、完整FPU、VBE與完整RAM前後保持。EIP依序2132B0／22F1F9／2131AD／21333E／2132BF／21334F，ESI495230／495230／495230／4952B0／4952B0／4953AA，R／堆疊與六個RAM雜湊不同。五後續區段各100000真正Step，unique_sites494／618／623／467／1326；前十位址按次數降序、地址升序固定，前四熱門2132E0／2132E2／2132E5／2132EA，各段次數2860／2021／2573／2816／2566。

六時點DisplaySets42、BankSets797、Writes16194454、Bank4／StartY0完全相同；五組相鄰及首末不同像素皆0。索引SHA b6fb8d68a422788d78e693eca398cab0e54c46127bc49de42c10d18c3aae9cb3，RGB efce8f0dd2eb63e07b25b6933808b2e18d6cc606419a867f292a8b37b44e05c6，六PNG皆59f76749db5232f97a6b6f969f56f848c2e89e731d7e3fb30b8786982bca4a81。首末兩圖已實際檢視，主選單與NEW GAME游標可見，設定畫面仍未知。這些只證明有界區間畫面未改與核心狀態有變，不證明loop正確、最終必然完成或產品掛起。

既有末尾32步已含真正213345／8B 15 74 BE 29 00讀DS:29BE74指標、21334B／01 D0、21334D／8A 00讀DS:[EAX] byte，21334F／88 45 F8寫SS:[EBP-8]；之後2132E0／31 C0、2132E2／8A 45 F8、2132E5／3D 80 00 00 00、2132EA／0F 84 64 00 00 00及JLE分支。前一輸入84h使JE／JLE未跳、低7bit4；最末AL42h已見。真正來源指標、讀取窗口、目的RAM與完整分支flags尚未另取樣，仍未知。DS:[ESI]窗口六次為零只描述該取樣地址，不當作這組DS:[EAX]指令來源，也不推字型／素材故障或追完整renderer。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-326-baseline.txt.gz | d37cf5c32298e69e38648ac9d70045d6501e188d56f1e7bb3fa2efb6a1877612 |
| workplace/dosgolem/workplace/moo2-probe-326-click.txt.gz | fa3d71704a13552e22ef0b04b5700afc9264934ed709a1c8134b5bb8b76782dc |
| workplace/dosgolem/workplace/post-click-326-parity-tests.txt | d9e9c2b7ab2386e6ae4d49b02b2070ebbd8742599a33a1424a71f5386f9da6e0 |
| workplace/dosgolem/workplace/post-click-326-verify.py | c0c40c21aa0ef315f64463cd9925a9cd7e56ec2383e1424caf4dc4aadf1f36c7 |
| workplace/dosgolem/workplace/post-click-326-backlink-tests.txt | 434fc2536d4b96a9e68e762dc36f81b712dcf0e51ac95d1b776dd01632c94f62 |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417根檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後沿325兩正式正常流程，只換326輸出名；完整環境見鎖定326。獨立python3 workplace/post-click-326-verify.py驗全部3847／4202列與兩終圖保持、六PNG CRC／640×480 RGB／SHA、窗口與每區段計數排序通過，僅mtime／DTA四bytes／PNG路徑正規化。63回填函式、原有32／49／25／27與新增27負例、兩CLI通過。

probe SHA-256 034e6a2c85a3cbbb93d9c7b762d3ba5aea070ec2fc126f82fef30bc913841512；CPU仍1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1，startup／provider／matcher保持325，325固定EXE完整Go測試仍有效，不重跑未改的CPU。原ZIP／patch／417根檔／EXE／MOX.SET再次核對PASS。全部原版素材／LOG／PNG／RAM留忽略workplace，不進Git。

工具ca437d1b17135d511b52e7fb40591b7ea12176c4已推送github隔離分支、回讀一致與工作樹乾淨，未推本機origin。[鎖定326](https://github.com/wicanr2/dosgolem/blob/ca437d1b17135d511b52e7fb40591b7ea12176c4/docs/spec/326-moo2-post-click-progress-observation.md)保存限定CONFORMED、六時點與實際來源邊界，325後段待辦同次回填。下一步沿同輸入與50M cap，核對真正DS:29BE74→DS:[EAX]→SS:[EBP-8]與CMP／JE／JLE的最小資料消費，必要時記目的RAM寫回；不追helper內部，不加cap、重點、代寫或先調整輸入。主庫玩法RE閘門、AH2Ch／RNG、人耳、正常開局與remake同狀態未完成。

### 2026-10-03 後段實際來源、三類分支與原值寫回

上一輪主庫a7f0c5df0b6b736766624829428a3cdab1d85d3b／工具ca437d1b17135d511b52e7fb40591b7ea12176c4已驗六時點觀測。本輪沿平台規格優先入口，327先DRAFT／READY，再只改原版探針唯讀觀測；CPU／平台與主庫玩法保持。官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、乾淨417根檔／MOX.SET、44M Esc／1996-01-01／50M cap／separate DOS與單次NEW GAME正常輸入完全保持。

**已證實，dosgolem高位LE正常原版**：DS188:29BE74原始D4A33D00→指標3DA3D4h；213345／8B 15 74 BE 29 00→21334B／01 D0→21334D／8A 00真正MOV與預讀來源一致，21334F／88 45 F8寫SS:[EBP-8]。六組於49500045／49500076／49500096／49500158／49500249／49500331開始，index3993h／3994h／3995h／3997h／399Ah／399Dh、DS來源3DDD67h／3DDD68h／3DDD69h／3DDD6Bh／3DDD6Eh／3DDD71h，byte依序02h／82h／02h／80h／82h／80h。分類條件執行前固定，各類前兩組，groups=[2 2 2 0]。

六組真正Step數27／20／27／33／20／33，160步連續R／六段／flags／堆疊保存，90步獨立核對，70步只保存原始續行。CMP／JE四次不跳、兩次跳；JLE兩次跳／兩次不跳，82h經兩次AND變2h與局部dword加2已驗。80h兩組在33步預定budget止於2132DB→2132DD，後續未驗，不延長追整個helper。XOR／AND未定義AF沿工具清除近似，不能稱硬體exact。

兩組02h真正213336／88 02將AL=FDh寫DS188:499300h與499303h；五byte窗口FDFDFD00FD→FDFDFD00FD及00FDFDFD00→00FDFDFD00。中心原本就是FDh，寫回與鄰接保持已驗，不稱值改變。213330／8A 80 7B BE 29 00映射表來源未另取樣，不猜素材用途。目的RAM不是畫面發布證據，六後段PNG與兩終圖保持326，DisplaySets42／Writes16194454不變，設定畫面仍未知。

初次獨立比較因六點完整RAM雜湊不同而失敗；各次ram_before_sha256=ram_after_sha256及readonly=true，R／段／flags／窗口／計數／VBE／原事件保持。未保存跨次完整RAM差異，不判定差異來源，不宣稱跨次全部RAM保持。審查後明示正規化每次RAM雜湊，只驗各次快照不突變；新來源觀測copy方向／探針局部狀態另經程式審查，沒有guest寫入或額外Bus／hook請求。修正後獨立python3 workplace/post-click-source-327-verify.py PASS：無點擊全部3847列、點擊全部4208列，除mtime／DTA日期時間四byte／PNG路徑／每次RAM雜湊外保持326，六快照各次readonly及六PNG／兩終PNG逐位元保持，沒有新CPU拒絕。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-327-baseline.txt.gz | c781b232058e3ca2c157a14c685ee44c12db851c00c6ed97d7f92344e9377d45 |
| workplace/dosgolem/workplace/moo2-probe-327-click.txt.gz | 6e46e29082ceff5cdcab14668c113b89ac8e43f0ba19f7ef2e562e613db564a5 |
| workplace/dosgolem/workplace/post-click-source-327-verify.py | f994ed0fabea7ad5eca2e51acc088c3988443670952be3258ec3f6990ab89d77 |
| workplace/dosgolem/workplace/post-click-source-327-parity-tests.txt | 15fef4b3ed0dd796418a71a1cc2c52bb1826373392ff665f7c8bd27516957664 |
| workplace/dosgolem/workplace/post-click-source-327-backlink-tests.txt | fc4f283990fae78884bce4c74056d1ed572d660118fb46b5581ae1c7c62b748a |

Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417根檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿326完整兩流程，只換327輸出名；全部環境見鎖定327及326。probe SHA-256 73ff02f6b4885207c029c99efa1c1920053e03b3fb2cbafc53db032ca2aef3d2，CPU仍1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1，startup／provider／matcher與325逐位元保持，325固定EXE全套PASS仍有效。64回填函式、原有32／49／25／27／27與327新增34缺證據負例、兩CLI PASS。原ZIP／patch／417根檔／EXE／MOX.SET再核對PASS。原版LOG／PNG／RAM留本機忽略workplace，不進Git。

工具b6fc71c0ba5a426a7c496fc3a5bd342cca1515ec已推送github隔離分支，回讀一致與工作樹乾淨，未推本機origin。[鎖定327](https://github.com/wicanr2/dosgolem/blob/b6fc71c0ba5a426a7c496fc3a5bd342cca1515ec/docs/spec/327-moo2-post-click-source-consumer.md)保存限定CONFORMED與觀測限制，326真正來源待辦同次回填。下一步先查既有VBE服務與正常trace，核對目的RAM的最小畫面發布契約及是否被消費；未找到實際玩家阻塞不擴大RE，不追完整renderer，不加cap／重點／代寫／先調輸入。正常開局、remake同狀態、主庫玩法RE閘門、AH2Ch／RNG／人耳未知保持。

### 2026-10-03 後段目的RAM取用與VBE發布監測

前輪主庫5bb8e0aeb1573409e9a983be81d3058874d7291d／工具b6fc71c0ba5a426a7c496fc3a5bd342cca1515ec。平台規格優先與文件職責已載入，328先DRAFT／READY。只改原版probe，在49500000包裝CPU Bus，每個真正請求只轉呼叫原Bus一次、值與錯誤透傳；CPU／平台／VBE服務來源逐位元保持。原版來源、417根檔、官方EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、44M Esc／1996-01-01／50M cap與單次正常輸入保持。

**已證實，dosgolem高位LE、固定正常單次點擊**：49500000..50000000恰500000真正CPU.Step；5775248讀／686978寫／errors=0。DS188 descriptor.Base=0、Limit=FFFFFFFFh，兩目標linear499300h／499303h，target_reads=[0 0]、target_writes=[3 5]；來源3DDD67h..3DDD71h共11byte、source_reads=154。前四源MOV值02h／82h／02h／02h，前三筆與327真正MOV核心吻合；兩目標最早MOV與327 AL／五byte窗口正對照吻合。包裝在真正CPU.Step內啟用，快照不計數、終態bus_matches=true。

499300h三筆寫於49500071／49575709／49910819，前兩筆213336寫FDh、第三筆2176A1寫00h。499303h前四筆於49500122／49557015／49575760／49614173，皆213336，寫值FDh／FDh／FDh／D5h；第五次只計數不猜值。前輪327兩筆原值FDh取樣不因此被推翻，後來改寫不當作素材故障或畫面已發布。

兩端VBEState皆Bank4／StartY0／BankSets797／Writes16194454／DisplaySets42；vbe_writes=0與Writes差額0對帳通過，六後段PNG與終點PNG仍59f76749db5232f97a6b6f969f56f848c2e89e731d7e3fb30b8786982bca4a81，設定畫面仍未知。零CPU Bus讀回只限此區間／路徑，平台服務直接RAM讀、較早較晚取用、完整renderer不在覆蓋內，不稱資料永不被消費或原版掛起。因真正來源與目的正對照存在，零讀回不能解釋成觀測未安裝。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-328-baseline.txt.gz | 5fecae395d6bda9def947025d4571974abc5121adea2708ad94cd213f51958d9 |
| workplace/dosgolem/workplace/moo2-probe-328-click.txt.gz | 09ae500cc236aba3c661aa967720b7fe861ffc30fbb9ca83c2de188ca8633878 |
| workplace/dosgolem/workplace/post-click-publish-328-verify.py | 90c8908f6cc691c6c464a2a9d2124b2ce2a071c6ca722a3cf2b654aee38b38f8 |
| workplace/dosgolem/workplace/post-click-publish-328-parity-tests.txt | 332849a1cbf7bb63338fbb9b0283a319eea05d088e6def28fd5f0182cfefd818 |
| workplace/dosgolem/workplace/post-click-publish-328-backlink-tests.txt | aab8c33aa44a17a15d66de7d707d38eff4f1da3eef72a4abd203c2fbd123d9b7 |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none。原ZIP／patch唯讀乾淨重建417根檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe後沿327兩正常流程，只換328輸出名。python3 workplace/post-click-publish-328-verify.py PASS：未點擊全部3847列／點擊全部4369列除327既定正規化保持，六快照各自前後RAM相同與readonly=true，六後段PNG／兩終PNG逐位元保持，沒有新CPU拒絕。未保存跨次完整RAM差異，不宣稱跨次RAM相同。65回填函式、原有32／49／25／27／27／34與328新增31缺證據負例、兩CLI PASS。

probe SHA-256 5590a5cad4663dcb91df9124648f04b2a70a6d4a25b9799d0082df262a76f0b5；CPU仍1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1，startup／provider／matcher與VBE服務逐位元保持，325固定EXE全套PASS沿用。原ZIP／patch／417根檔／EXE／MOX.SET再核對PASS，gofmt與擁有權通過；原版素材／LOG／PNG／RAM保持本機忽略，不進Git。

工具963a57228f429b8e028570f9d4c1c9cfddf16837已推送github隔離分支、回讀一致與工作樹乾淨，未推本機origin。[鎖定328](https://github.com/wicanr2/dosgolem/blob/963a57228f429b8e028570f9d4c1c9cfddf16837/docs/spec/328-moo2-post-click-publish-monitor.md)保存限定CONFORMED與覆蓋範圍，327最小發布待辦同次回填。50M基線已到上限無新CPU拒絕，沒有足夠證據把問題路由為CPU／素材／renderer故障；下一步另建DRAFT／READY有界續行觀測，保留原50M正式基線與同44M Esc／單次輸入，不能把不同終態混稱同狀態。不追整個helper或猜規則，正常開局／remake同狀態、主庫玩法RE閘門、AH2Ch／RNG／人耳未知保持。

### 2026-10-03 正常NEW GAME的獨立100M有界續行

起點主庫2cb1504010ccc2a1c566e920c15244b9df294129／工具963a57228f429b8e028570f9d4c1c9cfddf16837。命中並載入平台規格優先與文件分工，329先DRAFT／READY，只改原版probe預算閘門與最多六唯讀快照；CPU／平台／主庫玩法保持。固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、原417根檔／MOX.SET、44M Esc／1996-01-01與單次正常按下／放開不改。50M舊收據保留，100M是獨立觀測，不把不同終態混稱同狀態。

14 CLI負例PASS，拒絕發生在EXE讀取前；三真實EXE正例PASS。python3 workplace/new-game-329-verify.py核對50M未點擊全部3847列、點擊全部4382列除328既定正規化保持，六後段PNG／兩終圖逐位元保持。100M同50M前綴4335列保持，只另正規化明示預算；50M核心／VBE／Bus／CB快照與原基線吻合。跨次完整RAM一致不宣稱，各快照前後完整RAM與CPU／FPU／VBE／Bus保持readonly=true。

**已證實，dosgolem高位LE原入口與正常單次輸入**：47850592按下／47851578放開，六點50／60／70／80／90／100M全部不突變，完整FPU控制127Fh／status0／depth0／八槽0。EIP依序21334F／213311／2131BC／213239／21332A／213321，虛擬µs65660599／83672515／102038657／120423628／138808801／157193966。最終100M／flags297h／unique_sites27018，沒有新CPU拒絕，仍兩CB開始／完成2、callback_samples191。

BankSets797／797／802／807／812／817，Writes16194454／16194454／16207094／16220318／16233406／16246494，DisplaySets42、Bank4／StartY0維持。Bus VBE累計0／0／12640／25864／38952／52040，與Writes差額完全吻合；終態監測50500000真正Step，讀444340260／寫54307571／errors0，target_reads=[2 2]／target_writes=[446 423]、source_reads19481。328的50M零讀回與零提交僅限原區間，較晚確有讀回和顯存寫入，不能延伸成永不發布。

六PNG經CRC與640×480RGB／SHA核對；相鄰不同像素0／186／408／484／891，變化全落在x66..273／y414..422。50M與100M兩圖實際檢視，仍主選單六按鈕與NEW GAME游標；下方由空區變成Game Design／Steve Barcia致謝。最終PNG0ff69fc4f60431f01fb2dcadfbc7ee97b5378e87f0bdda3ff7eeff8fe2d4e363，設定畫面仍未知。較晚畫面變化不是完整renderer正確或NEW GAME指令激活的證據。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-329-baseline.txt.gz | 553cca0bcd9ba0b44bb2284877345efa1f0c6f25bb85354ee64bb1514f150bc4 |
| workplace/dosgolem/workplace/moo2-probe-329-click.txt.gz | 07aa81ac86c5e142fc11c99530907424241aadd320800c9a855625a67ab78f64 |
| workplace/dosgolem/workplace/moo2-probe-329-extended.txt.gz | c7dc2bb37a5292f588d3027cc7d99ac1e682d618bf8259dfc067bb6f533793c5 |
| workplace/dosgolem/workplace/new-game-329-cli-tests.txt | 365aaaa3a75ae05c12a430aa58310fd86ee8897e9444cc9bf931af4a95021989 |
| workplace/dosgolem/workplace/new-game-329-verify.py | 1e10a48db678caf3ed6a40ba4c6021fc6f8012729928a0c65dfc853ae240902c |
| workplace/dosgolem/workplace/new-game-329-parity-tests.txt | 8c041b02e2ac4594e2c3db9d0e47ee706bc3badd65d3c71d458e155796b97b98 |
| workplace/dosgolem/workplace/new-game-329-backlink-tests.txt | f2538841315d8d53db30eb0aaee340e72eb741bc32b5fe39c9b05eb1d9c6bb9d |

Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，ZIP／patch唯讀乾淨重建417根檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；先python3 workplace/new-game-329-cli-verify.py，然後沿328兩50M正常環境換329輸出名，再同點擊環境MAX_STEPS=100000000輸出329-extended。完整命令與環境見鎖定329。66回填函式、原有32／49／25／27／27／34／31與329新增36缺證據負例、兩CLI PASS。原ZIP／patch／417根檔／EXE／MOX.SET再核對PASS，gofmt／擁有權與來源保持PASS。

probe SHA-256 83154a870ee955744d847c26eb25ba94404eb718ec4ffcdeec5984e7a27a89bf，CPU仍1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1；startup／provider／matcher／VBE來源保持328，325固定EXE全套仍有效。原版素材／LOG／PNG／RAM保持本機忽略，不進Git。

工具2bfb2db0f860d115cb0e96e9d1e5938a89a25c23已推送github隔離分支、回讀一致與工作樹乾淨，未推本機origin。[鎖定329](https://github.com/wicanr2/dosgolem/blob/2bfb2db0f860d115cb0e96e9d1e5938a89a25c23/docs/spec/329-moo2-bounded-new-game-continuation.md)保存限定CONFORMED，328有界續行待辦同次回填。下一步改查按下／放開與首個主選單事件讀取的最小NEW GAME激活，不繼續加預算、不追整個renderer。事件覆寫、熱區錯誤或特定激活條件仍是假說，取到原始讀寫與branch證據前不修改測試輸入或玩法。正常開局／remake同狀態、主庫RE-first、AH2Ch／RNG／人耳未知保持。

### 2026-10-03 正常按下事件返回與首個上層邊界

工具34d758498f931d9dc155c4ca93309dd98646328e已推送github隔離分支、遠端回讀一致與工作樹乾淨，未推本機origin；固定規格[330](https://github.com/wicanr2/dosgolem/blob/34d758498f931d9dc155c4ca93309dd98646328e/docs/spec/330-moo2-event-return-caller.md)，較早329已回填入口。主庫起點870708cd415c38a69eed55a9d43cbf2391e19e24。

**已證實，固定1.31與原正常單次輸入的最小事件消費**：位址均dosgolem高位LE。47863846..47863861共16步，DS188:2A1228讀1／清0，213C83 C3的SS20:ESP2BDAD4指標69DB2000返回20DB69、EAX1，TEST的定義flags206h→202h，JNE751Ah非零臂跳20DB87。另一組47864834..47864849共16步，DS:2A1226讀1寫DS:26C518 word1，213A8E返回209197、20919D返回20DDF2，均EAX1；20DDF2 CALL209325的新return20DDF7與ESP-4吻合，立即停止觀察。共32步MOV／POP／三RET／TEST／JNE／CALL全部獨立核算，50M／100M的完整新列相同、readonly通過、callback始終2／2且非活動。兩事件均在47851578放開之後實際讀到，不支持短按事件被丟棄的猜測。

固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、原ZIP／patch／MOX.SET與417根檔保持。Docker Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原輸入唯讀；先go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿329兩50M與獨立100M正式環境換330輸出名。python3 workplace/new-game-330-verify.py PASS：全部3847／4382／6322列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊外保持329，72PNG逐位元保持。各次RAM前後相同、readonly仍檢查，不宣稱跨次全RAM一致。沒有新CPU拒絕，CPU／平台／CLI逐位元保持，325固定EXE全套與329 CLI結果有效。

初次探針遇callee的Jcc誤標caller；未採正式結論。330回DRAFT修訂、READY再實作首CALL停止、同容器命令乾淨重跑三流程，初次收據保留本機330-initial。67回填函式／既有負例與新增31缺證據負例、兩CLI通過。來源／收據1000:1000、gofmt與Git差異核對通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-330-baseline.txt.gz | 2e3944aef1e541d94231d0b737afd11041bee204bd376cfa5856a6a2ea824a0d |
| workplace/dosgolem/workplace/moo2-probe-330-click.txt.gz | a08c6b3c18fc1a4d89dcaedf69666d0b180e612ade7bff59bc742028025c5cdd |
| workplace/dosgolem/workplace/moo2-probe-330-extended.txt.gz | efa03837fe9dcac0be033eab345efc59cfdaa330a67030b5f82204d156ab130a |
| workplace/dosgolem/workplace/new-game-330-verify.py | a9bf44ae5cbd49df4aebba97dfd204243bc375a660065516409f739a6b262e2b |
| workplace/dosgolem/workplace/new-game-330-parity-tests.txt | d8c3fe1d846d3d0a41195c65ee38d05af7cafa3e2f471b878b11666a59c62d48 |
| workplace/dosgolem/workplace/new-game-330-backlink-tests.txt | 4423e974f97d1852c72ccd57080070b6fb6dbfdfa07aaafacffc7c6714b00584 |

probe SHA-256 931bb9d364a144460f2b358543f36e110f07e732020b80f69cadc5a6dbcdbdc1；CPU SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1。原版素材與LOG／PNG／RAM不公開，僅自製來源與有限文字證據提交。

**未知與下一步**：NEW GAME指令激活與正常開局仍未知。20DB87非零臂尚未連到按鈕命中／指令值，下一窄觀察只追這條分支，不改輸入、不重點、不繼續加預算，不追整個renderer或compiler helper。100M仍主選單；主庫玩法RE閘門、255完整座標／游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳與remake同狀態保持未知。Docker兩掛載路徑清查皆空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。

### 2026-10-03 非零事件caller與第一筆範圍跳過

工具a48f536a1132731c1b055e4419854642177b1c5e已推送github隔離分支、遠端回讀一致與工作樹乾淨，未推本機origin；固定規格[331](https://github.com/wicanr2/dosgolem/blob/a48f536a1132731c1b055e4419854642177b1c5e/docs/spec/331-moo2-button-branch-call-return.md)，330已回填入口。主庫起點4c04e57bfad75c6a75ac378ba31541cd822174a7。

**已證實，固定1.31正常單次輸入與有限caller樣本**：位址均dosgolem高位LE。47863862..47864195包含96實際caller步與238省略callee步；五次正常返回的EIP／SS／ESP吻合，省略[32,32,32,104,38]步，EAX[1,500,229,260001h,0]，只保存caller消費，不深入被呼叫函式。原版收到x500／y229，DS188:26C480原dword指標298848／DS:29BE0E word9／DS:29BE12 dword0，index*37h是55byte stride。index1實際讀DS:29887F八byte0A00140019002300，四word10／20／25／35；index2讀DS:2988B6的14001E0023002D00，四word20／30／35／45。實際來源與MOV寫回已獨立核算，不把它們直接命名為可見NEW GAME按鈕。

47864164於20DCE5 CMP EDX500,EAX25令flags216h，20DCE7 JLE不跳、20DCE9 E9到20DDAE，INC將初始SS:EBP2BDB40-44h的index1→2。**僅第一筆因x500>25被跳過已證實**，尚未走該筆y判定。sample_budget在第96樣本、index2讀完四word後停止，未達caller RET；不稱全部範圍不命中、NEW GAME被丟棄或原資料錯誤。

固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP／patch／MOX.SET與417根檔保持。Docker Go1.24.13固定映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原輸入唯讀；先go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，再沿330兩50M與獨立100M環境換331輸出名。python3 workplace/new-game-331-verify.py PASS：全部3847／4414／6354列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊外保持330，72PNG逐位元保持。全部96 caller步的原bytes／R／六段／EIP／定義flags／來源與寫回獨立核算；readonly與callback2／2、caller步IRQ無介入、兩預算完整新列相同。IMUL只驗定義CF／OF，SAR16的AF／OF未定義，不混入契約。CPU／平台／CLI保持，325固定EXE全套與329 CLI有效，沒有新CPU拒絕。

初次缺來源窗口的331-initial三收據僅作定位線索；回DRAFT／READY補直接有界RAM peek，同容器命令乾淨重跑，不改輸入。68回填函式、既有負例與新增37缺證據負例、兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-331-baseline.txt.gz | a74ea3d6d2d0af74126ba3df8a8b9e5617464883fc3617d82e5e21ba8ce0b87b |
| workplace/dosgolem/workplace/moo2-probe-331-click.txt.gz | 7ec5479ec14b80b6790b08b495e7d1ad0cd6953aa2f857e3b5b135af543a9a88 |
| workplace/dosgolem/workplace/moo2-probe-331-extended.txt.gz | ec2a35b53c8e20fa1dbecf205b5f12d241c0fe69a907d233fce8895e5d1265c8 |
| workplace/dosgolem/workplace/new-game-331-verify.py | 58094ab282f5e2e0c35212a3500b801f83c658d492fe16c7449f40dce821bbf9 |
| workplace/dosgolem/workplace/new-game-331-parity-tests.txt | 2b224220f0a6d7d5c08b30905c95ed37a229ba8cfa08222891608b38075fa251 |
| workplace/dosgolem/workplace/new-game-331-backlink-tests.txt | e1ceaca0b3593d9ae9d8b7daeee07b2b42cbe91ca7ab10e8261620885881bb15 |

probe SHA-256 05592edc1f377a163687a09f82ef2d1293e94d3577b8bf26ee5c41d2c80b3e23；CPU SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1。原版素材及LOG／PNG／RAM不公開，來源／收據1000:1000。

**未知與下一步**：正常開局／NEW GAME指令仍未知；優先同輪核對index2..8後續判定、完整caller返回／命中項與目前主選單的關係，不提高原流程cap、不改點擊時長、不重點，不追callee或完整renderer。主庫玩法RE閘門、255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳與remake同狀態保持。Docker兩掛載路徑清查皆空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。

### 2026-10-03 完整範圍命中與後續CALL

工具3580b3e26181ff0978fc7ed0b2c45f8c685f76e3已推送github隔離分支、遠端回讀一致與工作樹乾淨，未推本機origin；固定規格[332](https://github.com/wicanr2/dosgolem/blob/3580b3e26181ff0978fc7ed0b2c45f8c685f76e3/docs/spec/332-moo2-button-tail-return.md)，331已回填。主庫起點1c114dcd23bd06d4acdf52d1950d3885576723b6。

**已證實，固定1.31正常單次輸入的有界caller樣本**：位址均dosgolem高位LE，官方EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。47864196..47864849含313實際caller步、341省略callee步。DS188:26C480 dword298848、DS:29BE0E word9、DS:29BE12 dword0與55byte stride保持；index1..6右界25／35／45／55／65／75均小於x500，index7左界5000大於x500，實際分支拒絕；index8來源DS:298A00八bytes000000007F02DF01，四word0／0／639／479，命中x500／y229全部四比較。47864490的20DD3B保存局部8到初始SS188:EBP2BDB40-14h，47864503的20DDDB、66A3A6C42600保存word8到DS:26C4A6。原八項決策已取得，但全畫面項8與可見NEW GAME按鈕的關係未知。

47864507的20DDED CALL208FD4實際參數EAX8／EDX500／EBX229，return20DDF2與ESP-4已驗；47864849正常返回SS188／ESP2BDAD8／EIP20DDF2、EAX1，省略341 callee步。同一步20DDF2 CALL209325，新return20DDF7已驗，caller RET未取得。terminal為samples313／max_samples384／max_outer_steps8192、waiting=true、return_selector188／return_esp2BDAD8／return_eip20DDF7／outer_budget；觀察停止後原程式仍完成50M／100M，沒有新CPU拒絕，畫面保持主選單及credits。8192是觀察界限，不是產品失敗。

**來源限制**：313步的控制／R／六段／EIP／定義flags／數學核算通過；311完整來源、2 MOV僅低word來源已驗、1 IRQ堆疊寫回未重建。20DD2D／20DDCB的8B4006讀DS:[EAX+6] dword，原8byte窗口只含低word479，未捕捉高word；實際EAX701DFh／SAR16得到7可核算，但不稱完整typed record已證實。47864432完成IRQ7777→7778造成堆疊差異，原bytes保留而未重建ISR寫回。初次觀察器將其誤判轉向，回DRAFT／READY修正成對完成增量，同容器命令乾淨重跑；332-initial保留，不改CPU或IRQ平台服務。

Docker固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿331兩50M與獨立100M正式環境換332輸出名；python3 workplace/new-game-332-verify.py PASS：全部3847／4511／6451舊列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊外保持331，72PNG逐位元保持，舊331 terminal逐字保持。兩預算完整313新列相同；只讀快照保持，不宣稱跨次完整RAM一致。69回填函式、既有負例及新增34缺證據負例、兩CLI通過。CPU／平台／CLI未改，325固定EXE全套及329 CLI有效。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-332-baseline.txt.gz | d8d129deb83dcf71adf8cd46772e22206cbacf61be7f3723600d6e2de5bb2a55 |
| workplace/dosgolem/workplace/moo2-probe-332-click.txt.gz | f38048de3d96cc1db43b68f092ebd55fb0cf3443af57ca30a11436cb68d4e501 |
| workplace/dosgolem/workplace/moo2-probe-332-extended.txt.gz | cfb76ca5490e2dfa89cd74404f2c9a33bd48969fe4a4dcf49875273b3dbdc509 |
| workplace/dosgolem/workplace/new-game-332-verify.py | dd9ade15018780b0284232a058eec81678cf17446e1acb9979b2c19d2a3dde08 |
| workplace/dosgolem/workplace/new-game-332-parity-tests.txt | 9744908cee05cf75cf9e788cde86f96cadb1baff6a2cbb933fd48e20a50ab418 |
| workplace/dosgolem/workplace/new-game-332-backlink-tests.txt | 66e47bb4a8605e2bcf02e2886f899bb446fab9de363f33bf5d7ffeb31ecff658 |

probe SHA-256 32f37ac91a6758e6794f30882a5184228836b5ad84317058b51828a3b336a2a4；CPU SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1。原ZIP／patch／MOX.SET／417檔保持，來源／收據1000:1000，原版素材及LOG／PNG／RAM不公開。

**330歷史註記勘誤**：該節舊SS20h誤把GS當SS。internal/cpu386/cpu.go的段順序CS／DS／ES／FS／GS／SS，原段陣列[8 188 188 0 20 188]與return_selector=188證實SS188h。三次RET的原SS188偏移／bytes依序為47863859：2BDAD4／69DB2000→20DB69；47864842：2BDAB0／97912000→209197；47864848：2BDAD4／F2DD2000→20DDF2。原檔名／位址／bytes／雜湊與330歷史收據不變，本勘誤不重新推定函式語意。工具330現行規格附同一勘誤。

**未知與下一步**：正常開局／NEW GAME指令仍未知。唯讀核對當前實際註冊表／全畫面index8與可見主選單的關係，以及20DDF7正常返回；不猜title skip／輸入過早／指令意義，不提高原流程cap、不改點擊時長、不重點、不深挖209325或整個renderer。主庫玩法RE閘門、255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳及remake同狀態保持。Docker兩掛載清查空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。

### 2026-10-03 原表更換與正常20DDF7返回

工具d6688b01f7a5306bc6d271e06c4a5eb48a1eb430已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin；固定規格[333](https://github.com/wicanr2/dosgolem/blob/d6688b01f7a5306bc6d271e06c4a5eb48a1eb430/docs/spec/333-moo2-menu-table-return.md)，332已回填。主庫起點178d9bf932828aeb5ddf159f9f9d75e3d44d2949。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址均dosgolem高位LE。

**已證實，原CALL自然返回與原表**：47864850首callee Step前，R／段／flags、原globals／header／固定frame／新return bytes與code逐項對接332最後實際CALL。47990733實際EIP20DDF7／SS188／ESP2BDAD8，R=[3CE038 1DF 2100E5 E5 2BDAD8 2BDB40 0 7]、flags293h，新return bytes F7DD2000；callback2／2與IRQ7816／7816均非活動。CALL47864849與返回觀察相隔125884外層Step，未觀察callee逐指令；332原8192只是觀察界限，沒有新CPU拒絕。原caller最終RET仍未核對，不把此callee返回當整個選單命令完成。

開始與返回的DS188:26C480 pointer298848／DS:29BE0E count9／bias0／55byte stride，完整495bytes相同，SHA-256 d04abf3b20ccb6058d571a8092aa113242ddd8fd3ddefc42fc79d6cfff2a3a08。index1..8前8bytes與331／332實際來源一致：前6範圍左界10／20／30／40／50／60、右界25／35／45／55／65／75；index7四界5000、index8全畫面0／0／639／479，index0全零。完整index8在此時+8..+9為0700，不能把後來快照當332兩次MOV先前缺高word的原始收據。

50M／100M終態保持同一pointer298848但count7、bias0，完整385bytes逐位元相同，SHA-256 776c6e6e61a5b17529ff383cae79a194edc17c7cf0b9a11f3dc8fc8841339183。index0四界0；index1為415／172／567／193，index2為415／217／567／238，index3為415／240／567／260，index4為415／262／567／283，index5為415／285／567／306，index6四界5000。開始／自然返回／終態各3份新快照唯讀通過，原frame之後被正常呼叫覆用，不把終態return槽當原CALL仍活動。

**強推論，僅顯示位置**：人工查看50M原版PNG，終態index2範圍涵蓋NEW GAME字樣與500／229；以原word幾何比較只有index2命中。這不是新增原版點擊、handler語意或NEW GAME指令已執行的證據。先前唯一正常點擊實際消費時使用9筆表並選中全畫面index8；之後才觀察到7筆正式選單範圍，精確更換時點／producer與type／handler未知。

Docker固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none；原ZIP／patch唯讀重建417檔，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿332兩50M與獨立100M環境換333輸出名各一次。python3 workplace/new-game-333-verify.py PASS：全部3847／4825／6765舊列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊保持，72PNG逐位元保持，332 terminal未改；共同開始與返回快照保持。只讀RAM前後一致，不宣稱跨次完整RAM相同。70回填函式、既有負例及新增35缺證據負例、兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-333-baseline.txt.gz | 0157ca8aaf9998e13068830b69624850fbd77fe51ee64dafe9e66aca01b3161e |
| workplace/dosgolem/workplace/moo2-probe-333-click.txt.gz | f268d7b9285eb35b86246d465d6968d82e01c86eb5d2aea3ff24867a6d9ff723 |
| workplace/dosgolem/workplace/moo2-probe-333-extended.txt.gz | 690cd85a374102025fd8092950ad11d66ddb5ccc62a8305cd3f85e5e4a931c82 |
| workplace/dosgolem/workplace/new-game-333-verify.py | db1dc3e53c34a40105f34a8549c2a73caf44232e95de858d9e022e4ac192ccaa |
| workplace/dosgolem/workplace/new-game-333-parity-tests.txt | 3ab912b9a2e8cb2b80511007dc4853875617e6388e713fc88083f7fbab83f16d |
| workplace/dosgolem/workplace/new-game-333-backlink-tests.txt | b1eec641549e6ed249dcaf60a008abd3d1bb7b0d0e68d4332cba2808590282da |

probe SHA-256 fbd038f68ad029c1427cebf5baa852df0c2d90649554c38a95f10964f7ca0b57；CPU SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1。CPU／startup／provider／matcher／CLI保持，325固定EXE全套與329 CLI有效。原ZIP／patch／MOX.SET／417檔與來源／收據1000:1000再核對，原版素材／LOG／PNG／RAM不公開。

**未知與下一步**：正常開局／NEW GAME指令仍未知；保留44M與單次短按舊基線，另立實際7筆表就緒後的一次正常press／release情境，先確認原表與第2筆範圍，再驗實際caller消費與玩家可見後續。不是改正式遊戲或代寫狀態；不提高100M cap，不深挖209325或整個renderer。主庫玩法RE閘門、255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳及remake同狀態保持。Docker兩掛載清查空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。

### 2026-10-03 正式選單正常點擊與F2 SCASW拒絕

工具4501b831842f33ee5a0388b3018ae0c240949bc3已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin；固定規格[334](https://github.com/wicanr2/dosgolem/blob/4501b831842f33ee5a0388b3018ae0c240949bc3/docs/spec/334-moo2-ready-menu-normal-click.md)，333已回填。主庫起點6dbb1879f0c9ab6642aa0de3d8c1b504a800248c。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址均dosgolem高位LE。

**已證實，正式選單局部消費**：獨立ready100M情境保留44M Esc與舊47850592／47851578輸入，另在固定50000000核對原DS188:26C480 pointer298848／DS:29BE0E count7／bias0／stride55。完整385bytes逐位元對接333終態，SHA-256 776c6e6e61a5b17529ff383cae79a194edc17c7cf0b9a11f3dc8fc8841339183，第2筆415／217／567／238。R／段／EIP21334F／flags216h與333同50M checkpoint一致、peek readonly=true；callback2／2、IRQ8415／8415皆非活動才送額外press。

50000000 press：virtual_micros65660599／x1000／y229／buttons1；50011955 release：virtual_micros65691938／x1002／buttons0，差31339微秒，沿正常20ms與回呼完成契約首次可送時放開，不以固定986外層Step代替虛擬時間。全部只經InjectMouseEvent，未代寫EIP／RAM／索引。終態callback4／4、pending0／active=false。61538983於20DDDB執行66A3A6C42600，原EAX2、DS188:26C4A6 word0000→0200，下一EIP20DDE1；R／六段／flags297h保持，callback4／4與IRQ11675／11675非活動，error=nil。原store來源與寫回獨立驗證，已證實原版選中正式第2筆，不能把局部消費當完整新遊戲完成。

**新CPU阻塞已證實**：76658331於1F3640 bytes F2 66 AF 8B 45 FC 01 F0 48 2E FF 24 8D 8B 35 1F，被「F2 prefix 只支援 SCASB／MOVSB／MOVSD」拒絕。原起始EAX2／ESP2BDA0C，錯誤後EIP1F3643／ECX9／flags246h、DS／ES／SS188；protected IRQ15961／15961完成。REPNE SCASW尚未支援，是下一個CPU能力缺口；原情境在此停止、未到100M。VBE StartY512／DisplaySets43，最終640×480原PNG人工查看全黑，設定頁與完整正常開局未知。此處不深挖原函式或假設其玩法用途。

本機忽略PNG workplace/dosgolem/workplace/moo2-vbe-334-ready.png SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622；RGB SHA-256 0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366，indexed SHA-256 7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf。不是remake對照圖，不公開原版素材。

固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417檔。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；前三條沿333換334輸出名，第四條沿獨立100M增DOSGOLEM_MOO2_MENU_READY_CLICK=1，各一次。初次腳本輪次替換誤改預期EXE雜湊，輸入檢查停止／未啟動原版，修正腳本後同隔離設定乾淨重跑，不記產品缺陷。

python3 workplace/new-game-334-verify.py PASS：無旗標全部3847／4829／6769舊列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊保持333，72PNG逐位元保持；ready額外輸入前4780原列保持333獨立100M。新輸入後不同狀態不互比為同狀態。python3 workplace/new-game-334-cli-verify.py：14無效值／缺依賴在讀EXE前exit2通過，有效情境正對照越過參數閘門後缺原EXE明確失敗。71回填函式、既有負例及新增33缺證據負例、兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-334-baseline.txt.gz | 89af01a886da66931c694b5de618b1fa9f16e6adb9cae4fac11d7ca9402104bf |
| workplace/dosgolem/workplace/moo2-probe-334-click.txt.gz | 4b8eaa813db115852c5f965e7fc7f482fbcb1d76e45bdc55523ae7c67e5916f0 |
| workplace/dosgolem/workplace/moo2-probe-334-extended.txt.gz | ac175972abd192de3cf816eb1bf427b77a42c3cf32cd95419cbe81e4c2b09738 |
| workplace/dosgolem/workplace/moo2-probe-334-ready.txt.gz | 23c850a0b1f444da6975a300cd890f9d4022a5223db01a2b764617d1df21baa1 |
| workplace/dosgolem/workplace/new-game-334-verify.py | 02b8a7e1451cdbf7d608048f61bf8bce900645e809f0f6802faffc446318d229 |
| workplace/dosgolem/workplace/new-game-334-parity-tests.txt | 807639d9171e392e4c92623a2623dcb3a3447b5119a2bbbbda7720c0d8c08f6a |
| workplace/dosgolem/workplace/new-game-334-cli-verify.py | 10771963a67d238c3528a1cba08f10bcb383c4a8123b5b74794e8428822ab8fe |
| workplace/dosgolem/workplace/new-game-334-cli-tests.txt | 5b7e3ab2e595ba8396c74593efeb62daca5ef46729048bcf0625809d4ebd6ac2 |
| workplace/dosgolem/workplace/new-game-334-backlink-tests.txt | b1e9b6ea218e367510eadf5f1c1a323bd66361af655b8d001c19d4d94d857dc2 |

probe SHA-256 f58f52154c18389dea984d83d746980327529fd5d0150896d34734433efef8d3；CPU SHA-256 1448f29f24dad35e83575189055dc15dd7ac9db2b3529c6d450f03baf91c97b1。CPU／startup／provider／matcher未改，325固定EXE既有全套有效而未涵蓋新拒絕。原ZIP／patch／MOX.SET／417檔與來源／收據1000:1000核對，原版素材／LOG／PNG／RAM不公開。

**未知與下一步**：設定頁、完整正常開局、正式RNG與remake同狀態未驗。只補CPU的F2／66／AF字串指令，以既有[099-cpu386-repne-scasb](https://github.com/wicanr2/dosgolem/blob/4501b831842f33ee5a0388b3018ae0c240949bc3/docs/spec/099-cpu386-repne-scasb.md)、[292-cpu386-repe-scasd](https://github.com/wicanr2/dosgolem/blob/4501b831842f33ee5a0388b3018ae0c240949bc3/docs/spec/292-cpu386-repe-scasd.md)與處理器規格／硬體語料驗證ECX／EDI／DF／ZF與定義旗標；再重跑同一ready情境生成新原版收據，不代寫結果、不增加輸入或提高100M cap，不深挖1F3640或整個renderer。主庫RE-first／255完整游標／303整體DRAFT／299自然OF=1／AH2Ch／RNG／人耳保持。Docker兩掛載清查空，工具root-owned／誤建.md目錄自檢空，其他專案未清理。

### 2026-10-03 REPNE SCASW與正常新遊戲設定頁

工具74f574a78927f6bacdec95ea0519c83078bc71df已推送github隔離分支，遠端回讀一致、工作樹乾淨，未推本機origin。固定規格[335](https://github.com/wicanr2/dosgolem/blob/74f574a78927f6bacdec95ea0519c83078bc71df/docs/spec/335-cpu386-repne-scasw.md)，292／334與公開索引已回填。主庫起點884dec265c8ecef9390d1e8d0aa3574819e9b7f7。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；位址均dosgolem高位LE。

**334歷史停止的後續修正**：334新正常點擊已選中第2筆，但F2 66 AF不支援而黑屏。335先保存未改CPU完整初態，按[原Intel80386 SCAS契約](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/SCAS.htm)及[Intel SDM 2B](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf)印刷頁4-549..4-551審查READY，才增加通用REPNE SCASW的16位AX／ES word比較與32位計數／位址。只有F2 66 AF／66 F2 AF，其他未知形狀仍拒絕；沒有遊戲位址特例，主庫玩法不改。公開CPU契約導出的獨立規則驗證，不稱386實機語料、硬體exception重啟或逐週期對拍。

**已證實，原來源與最小消費**：76658331原1F3640的完整初態R=[2 9 D6 2BDA40 2BDA0C 2BDA14 78 1F357B]、段=[8 188 188 0 20 188]、flags246h，與未改CPU收據逐項保持。ES188:1F357B完整18bytes 03080208010800080300020001000000E936，SHA-256 8d541fc3f57f9ea2e104600161cf1973e124d7970ee34c6aea92329a2a9f8221；AX2匹配第6個word，ECX9→3、EDI1F357B→1F3587、flags246h保持、下一EIP1F3643。76658332原8B45FC讀SS188:EBP2BDA14-4即2BDA10的64000000，EAX2→64h、下一EIP1F3646，其餘核心／段／旗標與來源保持。callback4／4、IRQ15961／15961皆非活動、readonly=true、error=nil。來源用途未知，只核對這一MOV消費，不挖原helper。

**已證實，正常設定頁**：同ready情境的原正常兩組press／release、61538983原store與callback4／4保持334；100M cap到228DE8，無新CPU拒絕。終態R=[3 1DF 163 9F 2BDB40 2BDB9C 2C0DEC 2C0660]、段=[8 188 188 0 20 188]、flags246h。原640×480 PNG人工確認NEW GAME設定頁、Tutor／Medium／Average／5 Players／Average與CANCEL／ACCEPT可見，黑屏已解除。這只驗證設定頁，未驗選項操作、ACCEPT或完整開局。

本機忽略PNG workplace/dosgolem/workplace/moo2-vbe-335-ready.png SHA-256 1507dfb323dd4614fee5eb1523ab60c5652c7eb2a7fa6c53182772b5bdc64908；RGB SHA-256 3bbe6c339cfbc74e84b3210bccfdc4574073b4d308ef9b90a1720cc5c74e623e。原版素材／LOG／RAM不公開，這不是remake同狀態對照圖。

固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；Docker600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀重建417根檔與官方EXE。未改CPU初態情境一次，修正後四原版各一次；go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿334固定環境換335輸出名，不增加輸入或提高100M cap。日期1996-01-01不是seed。

python3 workplace/new-game-335-verify.py PASS：無旗標全部3847／4829／6769舊列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊保持334，72PNG逐位元保持；ready掃描入口前5833原列保持。未改CPU完整初態與原來源／掃描／下一MOV獨立核算，原正常輸入／選擇保持。CPU窄測試PASS 0.266s；DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1 PASS，cpu386 106.777s。自製92416算術／旗標／方向／前綴組合及多元素／高ECX／EDI繞回／故障恢復／拒絕測試，不嵌原版資料。72回填函式、既有缺證據負例與新增26負例、兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-335-input.txt.gz | c78e848f2577e04bbe46ed34968a9e3476d034197a4f0c860f609532af57c955 |
| workplace/dosgolem/workplace/moo2-probe-335-baseline.txt.gz | 62289373b26d4a2289fd23b6f6d7df8efb80779d3d037389767f381e1655a638 |
| workplace/dosgolem/workplace/moo2-probe-335-click.txt.gz | 15a308f1ad491c5c99f98f49d024c25d61d37e96a4ef5e6015fca5eeb584b8a3 |
| workplace/dosgolem/workplace/moo2-probe-335-extended.txt.gz | cb2a91372a22e6bcdf3786aa6060c3e27dd06890a97cd6553d2454037a65a7bf |
| workplace/dosgolem/workplace/moo2-probe-335-ready.txt.gz | 00d1c848f02c7d9fdf29fd92b7f4010a96e35f092eeb8b3f630b6e9701959233 |
| workplace/dosgolem/workplace/new-game-335-verify.py | dda63b17dfa67491bc140cc5f809c0c9b364656a8745c967000b4d30c9c7133f |
| workplace/dosgolem/workplace/new-game-335-parity-tests.txt | 9df66abe1d8bcf40295dd82919dc612be41da300a747478919d21dfb002b9b4a |
| workplace/dosgolem/workplace/moo2-335-cpu-narrow-tests.txt | d87bd1466a51b5a245fdff3125b4d2a688ba2b6e7aa7f1e508fceaebb46955a3 |
| workplace/dosgolem/workplace/full-test-335.txt | 2b8c03eb757e9a021b70885047357e77b4c931d23eec7f8f1376f56feae10e52 |
| workplace/dosgolem/workplace/new-game-335-backlink-tests.txt | 440e7f914cfc6c8fc5ba298db0e299344f759bea8f6a864ae70d4b4736aa597e |

CPU SHA-256 967de02753e0dce276413fe67a085e6df16789b52c948999d30ed26c3bb4e6a4；repne_scasw_test.go SHA-256 f8e512f4d57cc18e1eb3933f448f7c5493f3e4a2b224c3da72d00d8eecfac97d；probe SHA-256 63d47d60e38e0d2489aadda95aef334f3d66005c72cde96dd893f5d71f272979。startup／provider／matcher逐位元保持334，原輸入與新來源／收據1000:1000核對通過。工具root-owned／誤建.md目錄自檢空；Docker兩工作區掛載篩選空，沒有本輪遺留容器，未清理其他專案。

**未知與下一步**：只保存原設定頁按鈕表與ACCEPT正常輸入前置；不重開已完成SCASW、原掃描helper或renderer。ACCEPT、選族、完整正常開局、正式RNG、remake同狀態及整款中文化未驗。主庫RE-first、255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳保持。

### 2026-10-03 原版正常ACCEPT與選族頁

工具7c84f3931953cf3c9ffcbdc0718852e9c700b3ba已推送github隔離分支，遠端回讀一致、未推本機origin。固定規格[336](https://github.com/wicanr2/dosgolem/blob/7c84f3931953cf3c9ffcbdc0718852e9c700b3ba/docs/spec/336-moo2-setup-accept-normal-click.md)，335與公開索引已回填。主庫起點da52def1be5f3ae36ea713f2c0318054fd9a6c53；官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址均dosgolem高位LE。

**已證實，原設定頁表與前置**：335已正常進設定頁；336先補原333觀測器count≤16的限制，另保存80M／90M／100M原17筆表，不修改舊觀測行。原DS188:26C480 pointer298848／DS:29BE0E count17／bias0／stride55，完整935bytes三份相同，SHA-256 2a18a0213dcb1d539de3175c8356b8c8b886c1d61b38f63795b3fa1859a3f52f。原index15範圍433／392／527／414，原80M EIP22F1FB／完整R／六段／flags213h、callback4／4與IRQ16965／16965非活動、原設定頁RGB吻合。扣三新列後全部8146原335 ready列與圖片保持；原表與輸入初態足夠後336才READY。

**已證實，正常ACCEPT消費**：固定80000000 press、virtual_micros125567232、x960／y400／buttons1，正常callback座標480／400；release實際80011248、virtual_micros125610144、x962／y400／buttons0，原座標481／400，差42912微秒。只經InjectMouseEvent，兩點只命中原index15，不代寫RAM／EIP／選擇。80124668原20DDDB執行66A3A6C42600，EAXFh、DS188:26C4A6 word0000→0F00、下一EIP20DDE1；R=[F 0 339 0 2BDB14 2BDB7C 0 2B0001]／段=[8 188 188 0 20 188]／flags297h保持，callback6／6與IRQ17000／17000完成、皆非活動／非failed、error=nil。原store與下一畫面共同證實ACCEPT消費，不能只用幾何交集宣稱成功。

**已證實，SELECT RACE選族頁**：原版同100M cap到原21595F，無新CPU拒絕。終態R=[7F 1A8 341DF4 8A 2BDA7C 2BDA9C D 1A]／段=[8 188 188 0 20 188]／flags212h。實際640×480 PNG人工確認SELECT RACE與種族／Custom按鈕，滑鼠留在Custom附近不能當已選任何種族。90M／100M原count16／pointer298848／bias0，完整880bytes相同，SHA-256 f03515b12cb289bfcfe49b46d5cf8619f8f4e300ccfd1bd4ca43107f1c313ade，保存作下一正常輸入來源，欄位用途未完整推定。

本機忽略PNG workplace/dosgolem/workplace/moo2-vbe-336-accept.png SHA-256 7aec4ca6aad1f948560e184695bd3b415b1145ad9778e536da14a60e11d61861；RGB SHA-256 9d8c0a1acb3b96200296f789bb13c6877067f6165ec608a7e019506036832dfc。原版素材／LOG／RAM不公開，這不是remake同狀態對照圖。

固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac／Docker2GiB／2CPU／128pids／UID1000／network none；原ZIP／patch唯讀重建417根檔／官方EXE。初態蒐證一次600s，正式五情境各一次900s外層有界逾時，原版各條仍100M cap。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；沿335四基線換336輸出名，加DOSGOLEM_MOO2_SETUP_ACCEPT_CLICK=1獨立情境，不提高cap或改舊事件。日期1996-01-01不是seed。

python3 workplace/new-game-336-verify.py PASS：四舊情境3847／4829／6769／8146原列除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊保持335，ready只扣三新唯讀列，102PNG逐位元保持；新ACCEPT額外80M輸入前6145原列保持獨立ready，原表／RGB／核心前置與實際store核算通過。新正常輸入後不同狀態不冒充同狀態。python3 workplace/new-game-336-cli-verify.py：15無效值／缺依賴在讀EXE前exit2、有效正對照通過；同新binary舊334 CLI 14拒絕與正對照保持。73回填函式、既有缺證據負例與新增28負例、兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-336-input.txt.gz | ec8f0867e944a0323813316b3805e21c1e220c654b0f6651826b210cb6d043fb |
| workplace/dosgolem/workplace/new-game-336-input-verify.py | e22d9aa59b4b5cc855b0777f6dc5537ee05f8cf83137c9255b8167d6e8b402ff |
| workplace/dosgolem/workplace/new-game-336-input-tests.txt | 7802819d83ad8c61b7c49d63b7bed05f58e20ed0b011408855331cb5241770f3 |
| workplace/dosgolem/workplace/moo2-probe-336-baseline.txt.gz | bde1882211515e81d17abd98f9c2a49fa8cbe3f11939c80cd88680da69f3fedb |
| workplace/dosgolem/workplace/moo2-probe-336-click.txt.gz | 29445900c269635d1c47877d42eed2f3d4f8cc4effb7c0b318d33c4a280a57f1 |
| workplace/dosgolem/workplace/moo2-probe-336-extended.txt.gz | 08cbf2b2994d353033ec84c22df2f9e0a03e21227e7a817ae0447b915583c0ff |
| workplace/dosgolem/workplace/moo2-probe-336-ready.txt.gz | e3f22c215074a9d3911971a9bdadc76414e569162c868f224ffd23f6d0abbbfc |
| workplace/dosgolem/workplace/moo2-probe-336-accept.txt.gz | a8c64e4519ac73c78578e3d953257c6cd8d19b5be1345899f15cab0c0b86a252 |
| workplace/dosgolem/workplace/new-game-336-verify.py | da6336609bb79a5887b383a2505ed4f7f4f4cfa927aa59b9a46f924d9b5d7083 |
| workplace/dosgolem/workplace/new-game-336-parity-tests.txt | a7488a255580640aef0af2a8b67d1c9572ff7f81e6c9bc0007350e48eee4db0d |
| workplace/dosgolem/workplace/new-game-336-cli-verify.py | f63589c59f6a91749537d75453af4bf8da6883776e02849ae006d6c9f4bb678b |
| workplace/dosgolem/workplace/new-game-336-cli-tests.txt | 503b4534840a44fe81a235244c8ea4e2393610ed8458b2b23f2719db66efc313 |
| workplace/dosgolem/workplace/new-game-336-backlink-tests.txt | 4d79d60bd98d9b0eddf69c3a4a855e10077477f90632cf1252ae78f240d42e74 |

新probe SHA-256 cfcc89585eb163e67c3043202501f957708b4818985ddbd6d4b1f2138c635b19；CPU SHA-256 967de02753e0dce276413fe67a085e6df16789b52c948999d30ed26c3bb4e6a4。CPU／startup／provider／matcher逐位元保持335，335固定EXE Go全套仍有效，但不涵蓋整款原版玩法。來源／新收據1000:1000、gofmt與Git差異檢查通過。工具root-owned／誤建.md目錄自檢空，Docker兩工作區掛載篩選空，沒有本輪遺留容器，未清理其他專案。

**未知與下一步**：只獨立解碼已保存90M／100M的原16筆選族表與正常種族輸入前置，再依新READY規格送一次正常選擇。不重開已完成ACCEPT／SCASW，不深入renderer或原helper。種族選擇、名稱輸入、完整正常開局、正式RNG與remake同狀態未驗；主庫RE-first、255完整游標、303整體DRAFT、299自然OF=1、AH2Ch／RNG／人耳保持。

### 2026-10-03 原版正常人類選擇與統治者名稱頁

工具 0633c346ce2ca1156ab26c1dbc533ec0e43920e0 已推送 github 隔離分支，遠端回讀一致，未推本機 origin。固定規格 [337](https://github.com/wicanr2/dosgolem/blob/0633c346ce2ca1156ab26c1dbc533ec0e43920e0/docs/spec/337-moo2-race-humans-normal-click.md)，336 與公開索引已回填。主庫起點 40e3a404d22636ae85826d0a0da8da182e8ce457；官方 1.31 ORION2.EXE SHA-256 為 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間均為 dosgolem_high_le。原版素材、LOG、PNG 與 RAM 留在忽略的 workplace，不公開。

**已證實，原選族表與前置**：獨立核對既有 336 ACCEPT 收據 a8c64e4519ac73c78578e3d953257c6cd8d19b5be1345899f15cab0c0b86a252。90M／100M 的原表 count16、DS188:26C480 pointer298848、DS:29BE0E、bias0、stride55，完整 880 bytes 相同，SHA-256 f03515b12cb289bfcfe49b46d5cf8619f8f4e300ccfd1bd4ca43107f1c313ade。index7 原位址 2989C9，範圍 351／330／473／374；對照 Humans 按鈕為強推論。兩個候選點 412／352、413／352 都只命中第 7 筆。90M 原 EIP238576、R=[369321 7A 2D 0 2BD9F0 2BDA4C 502022 369321]、段=[8 188 188 0 20 188]、flags206h，callback6／6 與 IRQ19929／19929 完成且非活動、非 failed；FPU 控制字127F／status0／depth0，VBE Active=true／StartY512／DisplaySets47，RGB SHA-256 9d8c0a1acb3b96200296f789bb13c6877067f6165ec608a7e019506036832dfc。完整原 RAM、表與核心狀態唯讀核算通過後才 READY。其他欄位用途仍未知。

**已證實，正常第 7 筆選擇**：固定 press90000000、virtual_micros154693274、x824／y352／buttons1，原座標412／352；release90010495、virtual_micros154735567、x826／y352／buttons0，原座標413／352，差42293微秒。正常 callback7／7 完成後首次可送時放開，只經 InjectMouseEvent，不代寫 RAM、EIP 或種族。90056672 原20DDDB 實際執行66A3A6C42600，EAX7、DS188:26C4A6 word0000→0700、下一EIP20DDE1；完整 R=[7 502004 181 298848 2BDA2C 2BDA94 D 1A]、段=[8 188 188 0 20 188]、flags297h 保持，callback8／8 與 IRQ19946／19946 完成，error=nil。原選族表第 7 筆確實選中；typed 種族／trait producer 不在本輪驗收。

**已證實，統治者名稱頁**：同100M cap到215DEE，無新CPU拒絕。終態 R=[6 178 DDE0 7 2BD94C 2BD970 2843A5 28439D]、段=[8 188 188 0 20 188]、flags206h、callback8／8 完成。640×480 原 PNG 人工確認 Enter Ruler Name、預設 Strader 與 ACCEPT；末尾底線只視為畫面字形，不當作名稱緩衝區字元。本機 workplace/dosgolem/workplace/moo2-vbe-337-humans.png SHA-256 7f1725d8669cacd9758350420bb60e4dcf6f01c8139ad422edc1e49b3577fc61；RGB SHA-256 3e264c7fbd003a8e24cfea2e4fc0e019d9ee096a760c467634b44a6ef37a9fc9。

原100M名稱表 count3／pointer298848／bias0／stride55，完整165bytes SHA-256 f3dc28cf153edf625b154c5864b3040cddd52fc9ade2a0cfd2a256b80789d0a7、readonly=true，IRQ22901／22901 完成。index1 範圍273／225／371／253 對照 ACCEPT 為強推論；index2 範圍211／179／313／205 的文字欄位語意未知。終態不能充當同100M上限內的輸入前置，沒有送出名稱確認或代寫文字。

固定 Go1.24.13 映像 sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；Docker 外層900s／2GiB／2CPU／128pids／UID1000／network none，原 ZIP／patch 唯讀重建417根檔。六個原版情境各重生一次，原版每條仍100M cap。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；沿336五條基準改337輸出名，新增 RACE_HUMANS_CLICK=1 的獨立情境。固定日期1996-01-01不是seed。

python3 workplace/new-game-337-verify.py PASS：五條舊情境3847／4829／6769／8149／7752原列，除既定mtime／DTA四byte／PNG路徑／每次RAM雜湊正規化外保持336，全部132PNG逐位元保持。本輪不扣336已存在的setup_table_snapshot三列。新humans於90M額外輸入前6821原列保持獨立ACCEPT。原表、RGB、核心前置與原store獨立核算通過；輸入後不同狀態不冒充同狀態。

python3 workplace/new-game-337-cli-verify.py PASS：16無效值／缺依賴在讀EXE前exit2；有效正對照越過參數閘門後，缺EXE明確失敗。同新binary舊336 CLI的15個拒絕案例與正對照保持。74項文件檢查、新增27個缺證據／狀態／索引負例與兩個CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/new-game-337-input-verify.py | 3d3db6ccaf6bf48ab934f148bf31ec8f2d166c8568811bb5d032996163b36c57 |
| workplace/dosgolem/workplace/new-game-337-input-tests.txt | 4ac2d0b56904d9fba8323c5ceb184ed9e5e0d5fb40abf1ecfc9415f2b02abf2a |
| workplace/dosgolem/workplace/moo2-probe-337-baseline.txt.gz | 2b86481ae22b7c4cf675ac169890a33b63b574dad642c150a0ffd7b3a0b40b3e |
| workplace/dosgolem/workplace/moo2-probe-337-click.txt.gz | 32e007ef3550b70775faa8a89362e0c963bc2ec542b9cf88818b1ebfb3cc2c45 |
| workplace/dosgolem/workplace/moo2-probe-337-extended.txt.gz | b6e5177086d2477db952c78774cbc72e92e93c3a7a72102273ff8984682deaba |
| workplace/dosgolem/workplace/moo2-probe-337-ready.txt.gz | ca2aa612913c9ef5d90666576039ddccc28c541fc2aa7e610b048bb12de0861b |
| workplace/dosgolem/workplace/moo2-probe-337-accept.txt.gz | 5a0ad464cc663be9701bad5314f4c82308a310c19e9b8bb2bba0dcbce23f5c40 |
| workplace/dosgolem/workplace/moo2-probe-337-humans.txt.gz | 541d0032fa9711a65fe00f62018cf00bea7a78daed46dc06ce79fe8b5de20e45 |
| workplace/dosgolem/workplace/new-game-337-verify.py | 95ff52a16ae185bd989a74e4bff962758d08884680eebe7d4dfb7c462450c134 |
| workplace/dosgolem/workplace/new-game-337-parity-tests.txt | 9b41ace6e35830796bade6341a10688f7800bbd37522b5172e124826709ae30f |
| workplace/dosgolem/workplace/new-game-337-cli-verify.py | 5c12da132ba19c3139c0014f6513b6d4755d4e1c74310e4e0e8e095c32663dd6 |
| workplace/dosgolem/workplace/new-game-337-cli-tests.txt | 22d72c914180b92bec7ae33148f59702fe53d1d839221c0d51dc25aef271fce1 |
| workplace/dosgolem/workplace/new-game-337-backlink-tests.txt | 9b9c07085f859ad64350ee7ea8fc919cab4d0dee989eebbe5d6b5df4cd52844f |

新 probe SHA-256 6d13d37c144e35cea19a2decc1b623b4d24023d07cee341a25d4ff5f5ba93712；CPU SHA-256 967de02753e0dce276413fe67a085e6df16789b52c948999d30ed26c3bb4e6a4。CPU／startup／provider／matcher 逐位元保持335，335固定EXE的Go全套仍有效，本輪沒有重跑全套，也不外推整款玩法parity。新來源與收據1000:1000，gofmt／Git差異通過；工具root-owned／誤建.md目錄自檢空，Docker兩工作區掛載篩選空。

**未知與下一步**：在同humans情境較早正常時點唯讀保存名稱頁、165bytes原表、原名稱緩衝區及ACCEPT可接受輸入前置，再依新READY規格正常確認。本輪未確認名稱；typed種族特性、完整開局、正式RNG與remake同狀態未知，主庫RE-first保持。不重開已完成選族／ACCEPT／SCASW，不深挖renderer或原helper。

### 2026-10-03 原版正常名稱確認與旗幟頁

工具 0c88d04cb59d2b7bfc716e953cc529c54b5fb6e2 已推送 github 隔離分支，遠端回讀一致，未推本機 origin；固定規格 [338](https://github.com/wicanr2/dosgolem/blob/0c88d04cb59d2b7bfc716e953cc529c54b5fb6e2/docs/spec/338-moo2-ruler-name-normal-confirmation.md)，337 與索引已回填。主庫起點 5c1b6a4a42d3ff8fc69125d00587cd66277eef96；原官方 1.31 ORION2.EXE SHA-256 為 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間為 dosgolem_high_le。以下只有 raw RAM 候選清楚標為 dosgolem_ram_linear，DS188 descriptor base0／limitFFFFFFFF；未因數值相同混用工具的位址基準。原版資料、LOG／PNG／RAM 留在忽略 workplace。

**已證實，正常輸入前置**：私有唯讀探針在原 95M／97M／100M 保存完整3筆／165bytes表，三份相同，SHA-256 f3dc28cf153edf625b154c5864b3040cddd52fc9ade2a0cfd2a256b80789d0a7。原DS188:26C480 pointer298848／DS:29BE0E count3／bias0／stride55；index1位址29887F，範圍273／225／371／253，候選320／239與321／239只命中該筆。95M原EIP215D9E，R=[A0 178 2C0864 7 2BD94C 2BD970 2843A5 28439D]／段=[8 188 188 0 20 188]／flags202h、FPU127F／status0／depth0，callback8／8、IRQ21395／21395完成且非活動／非failed。95M原圖人工確認Enter Ruler Name／Strader／ACCEPT，PNG SHA-256 e1c739f5aeaf47a6cfdca4749b14509ec2592b65974fa6669e9347c491a52a36，RGB SHA-256 2646a7ef25939869ede60aa35bcd97d649b906fd2cc287d08deb6c2d8058cc12。剝除23新唯讀列後，337 humans全部9030列與30PNG保持，初態足夠後才READY。

**已證實，raw名稱候選；語意為強推論**：原index2+24 dword為28439D；該位置32bytes為5374726164657200000000000000000000000000000000000000000000000000。95M／97M的原RAM還見29871C、2BD9BC、2BDAAC、506994的Strader字串，100M的29871C已不相同。表指標與EDI28439D對應，只把文字編輯buffer稱為強推論；正式持久名稱writer仍未知，不把字串存在當存檔已確認。

**原版實測勘誤**：首次七情境來源667e3dd2a6a56e59d610ad40f4bd94e418ddaf6ee89faf2b8d611807ebe83e83，正常95000000按下後callback9／9完成，原mask從2B改為1。舊探針仍要求2B才放開，100M終態pressed=true／released=false／store_seen=false，仍停在名稱頁，無新CPU拒絕。首次收據及來源另存，未當成功。回到DRAFT核對既有 internal/machine/le_mouse_callback.go 的 InjectMouseEvent：位置移動為flags1、左鍵放開為flags4，固定x640→642與放開產生flags5，5&mask1=1可排正常回呼並更新裝置狀態。重新READY後只允許本明示分支在已按下／回呼完成後使用mask1或2B放開，其他初態與事件保持。沒有修改CPU、平台或原版mask。

**已證實，正常名稱確認與旗幟頁**：按下95000000、virtual_micros167823195、x640／y239／buttons1、mask2B；放開95015426、virtual_micros167870767、x642／y239／buttons0、mask1，差47572微秒。回呼9／9完成且至少20ms後首次可送放開，兩個原座標320／239、321／239只命中index1；終態callback10／10完成、pending0／非活動，mask回到2B。共享20DDDB未命中，不能捏造原index1 store或正式名稱寫回。

原版同100M cap到228E0E，無新CPU拒絕。終態R=[37 EE AF 4B 2BD9B8 2BDA14 2C0B20 2C0390]／段=[8 188 188 0 20 188]／flags212h，callback10／10及IRQ22854／22854完成且非活動／非failed。實際640×480 PNG人工確認SELECT BANNER COLOR與八色旗幟。workplace/dosgolem/workplace/moo2-vbe-338-ruler.png SHA-256 96efbd1ce6538c27b019cc6713fe7397d6d0fde7a63c7e01cb82023f75614c67；RGB SHA-256 8c2c573c08ebf4bce0b4e166d4268fb6fcf90f35dd27fbbb2de614b26f187f15。

原100M旗幟表count10／pointer298848／bias0／stride55，完整550bytes SHA-256 978afb91aaed9e0cb37672a66352e2ae515b382d5b6f472da9238e09fb650dfb，readonly=true。index1..8位置對照八色為強推論，index9原5000／5000矩形用途未知。終態表只作下一正常輸入來源，不從100M終態直接送新事件。

工具固定Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac、Docker2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀，每次由乾淨417根檔與官方EXE重生。首次7情境外層900s；修訂後只重播新版ruler與新舊CLI一次、外層300s，每條仍100M cap。首次六條基準的來源667e3dd2與新版1cc5f74e只差rulerAccept旗標分支的一行放開條件，唯讀來源審查逐位元核對其餘不變；六條基準不進此分支，因此不重跑未受影響情境。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；固定日期1996-01-01不是seed。

python3 workplace/new-game-338-verify.py PASS：六情境全部3847／4829／6769／8149／7752／9030原337列除既定mtime／DTA四byte／PNG路徑／各次RAM雜湊正規化保持，全部162PNG逐位元保持；新版ruler於95M額外輸入前7523原列保持獨立humans。原完整表、候選32bytes、完整核心及RGB初態對接，不把正常輸入後不同狀態冒充同狀態對拍。

python3 workplace/new-game-338-cli-verify.py PASS：17無效值／缺依賴在讀EXE前exit2，有效正對照過閘門後缺EXE明確失敗；同新版binary舊337 CLI的16拒絕與正對照保持。75項規格回填正對照、新338 26個缺證據／狀態／索引負例、337 27個負例與兩CLI通過。

| 本機忽略收據／腳本 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-338-input.txt.gz | 47172f944786e4d134091485ce73ae64d7df622fb75470fa62ddf7a660526e9e |
| workplace/dosgolem/workplace/new-game-338-input-verify.py | e29103c693b3536da56b9081c75e8edf297776982fb217a4d0688ecdbb730544 |
| workplace/dosgolem/workplace/new-game-338-input-tests.txt | b26cb2cb901c0930689c1c17596550442dd41b70e1a78ac822471ab955949e6b |
| workplace/dosgolem/workplace/moo2-probe-338-baseline.txt.gz | 94e4e8e18e6cb21fca2a263bf1431380eb5ea7d6db7f5e83b5bf1619390a8c49 |
| workplace/dosgolem/workplace/moo2-probe-338-click.txt.gz | 5afe2ab6f88283b926beaca6352e14eb87f581fcf4179082a488c55c0b5ada6b |
| workplace/dosgolem/workplace/moo2-probe-338-extended.txt.gz | 4991ba7afda8c7343e7d450d5292a8f44e54ef3196d72876cd38ff85f47191f5 |
| workplace/dosgolem/workplace/moo2-probe-338-ready.txt.gz | 260c7e5a372ffa46e1cb8cbc0c073596eacd8252605765c46a3f453b10ba93b9 |
| workplace/dosgolem/workplace/moo2-probe-338-accept.txt.gz | cd4ac24f40c5e6ab27afd6e9aa2604fce6ae95b636e710d385dc4b831642db8a |
| workplace/dosgolem/workplace/moo2-probe-338-humans.txt.gz | cd9b796834045d35ddab5afce9e40d791d0836c23e4cef6f454b566930ebe91c |
| workplace/dosgolem/workplace/moo2-probe-338-ruler-press-only.txt.gz | 1c71292dca36c0046c12b09bf1ebba9704492eaa55d72873fb2acf4d515fda44 |
| workplace/dosgolem/workplace/moo2-probe-338-ruler.txt.gz | 21e94e53b1a4f43b3a6fe736490412ea291bb2c3796c922c1940a6dbd2a5dcbb |
| workplace/dosgolem/workplace/new-game-338-verify.py | c507578eda62b34522c882e0f1a97951c10f5adb384e00fc06a96b7e24e52300 |
| workplace/dosgolem/workplace/new-game-338-parity-tests.txt | 79a5f7d0444a38afad151dc09544a1aea21339fb5ef4be8f7a36b6a5e6c531e6 |
| workplace/dosgolem/workplace/new-game-338-cli-verify.py | 872069a6fcdfe31ae5779494ff879d529d84c656a8889f4a361f2d722aaa1e79 |
| workplace/dosgolem/workplace/new-game-338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/moo2-name-338-input.go | 9aa51fcddc4ffdc44896424f3bd50757c2cf49b52390dbc90550dd130faf812a |
| workplace/dosgolem/workplace/moo2-name-338-press-only.go | 667e3dd2a6a56e59d610ad40f4bd94e418ddaf6ee89faf2b8d611807ebe83e83 |
| workplace/dosgolem/workplace/new-game-338-backlink-tests.txt | 7ceda5b0175a4d41a30c89cccde10c9d1ff6940241d4d8c41460c7459cf8e042 |

新版probe SHA-256 1cc5f74e48b63017b4dc68cfe714b2bec674e7adba0c47d8f119f095583f58c5；CPU SHA-256 967de02753e0dce276413fe67a085e6df16789b52c948999d30ed26c3bb4e6a4。CPU／startup／provider／matcher／mouse callback逐位元保持335，本輪未重新執行Go全套，不外推整款玩法。三份初態RAM的65818624bytes各自唯讀保持：95M SHA-256 6272fad6da4146af26f7513cba625ec2ad2a040d75cdf31cc2b332cffd7a463c，97M SHA-256 6e1f1ea729118153c95e716b1be96dc072d2208942d7d42510872fcc4c561cad，100M SHA-256 5de3d135522bb1a1b08d701b884b4547f287a57aba57bb10e6500a1cbf6ad673；不宣稱跨次RAM相同。

來源與新收據1000:1000、gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空，本輪容器已清理。

**未知與下一步**：只取得同ruler情境較早旗幟頁、原10筆表與正常選色前置，再依新READY規格送一次可重播的色彩fixture。預設名稱正常確認與旗幟頁已驗；正式持久名稱writer、旗幟選擇、typed種族特性、完整開局、正式RNG與remake同狀態未知。主庫RE-first保持，不重開已完成的名稱放開／設定ACCEPT／SCASW，不深入renderer／helper。

### 2026-10-03 原旗幟正常輸入與按鍵查詢，選色未完成

工具起點0c88d04cb59d2b7bfc716e953cc529c54b5fb6e2，主庫起點f079ef801ab20d3ab1949ad9ee8d5e9693ee9eea；工具7d569e0392fd9061576da8395d3c8b4d10e0886d已推送github隔離分支，遠端回讀一致，未推本機origin。公開[339旗幟正常輸入與按鍵查詢](https://github.com/wicanr2/dosgolem/blob/7d569e0392fd9061576da8395d3c8b4d10e0886d/docs/spec/339-moo2-banner-red-normal-click.md)只限定原表／正常按鍵查詢／放開CONFORMED，不把它算成原選色或主庫玩法完成。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具Go1.24.13與映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac保持。原ZIP／patch唯讀，每次重建新鮮417根檔及MOX.SET；固定日期1996-01-01不是seed。

**已證實，原輸入前置**：可丟棄唯讀probe保存98M／99M，剝除6新列後全部7874原ruler列與30PNG保持338。98M只有count1／55bytes，99M才有count10／550bytes；後者表SHA-256 978afb91aaed9e0cb37672a66352e2ae515b382d5b6f472da9238e09fb650dfb、RGB 8c2c573c08ebf4bce0b4e166d4268fb6fcf90f35dd27fbbb2de614b26f187f15。原pointer298848／bias0／stride55，index1 rect96／144／179／242。紅旗對應與原選色結果仍分開，座標命中不算原選擇成功。

**未推進收據**：初版來源4e628d389e2910f7dea1c5bdacff3372e863b49524487125c4dfff3586268662，99M按下／99019204放開、差20000微秒，原100M仍旗幟頁。回DRAFT只延長相同短按至120M，來源6c63547a372fd329bc58351928d0e8dc2a7806a45461ce9d5a049e659c05fe63，全部7837原100M前列保持，仍旗幟頁，沒有新CPU拒絕。兩份負收據及來源保留。

直接解析第二份原收據，press到release原INT33 AX3輪詢0次，release後492次皆buttons0。核對現存internal/machine/le_startup.go的AX3把正常裝置mouseButtons返回BX低word、x／y返回CX／DX；InjectMouseEvent來源internal/machine/le_mouse_callback.go不變。回DRAFT再READY，probe只等待首次原正常callsite24C31B查詢確實讀到pressed，再按既有回呼／IF／IRQ／20ms前置放開，不移動按下時點、不重送、不代寫CPU／RAM／選擇。

**已證實，原正常按鍵查詢與放開**：99000000、177207342微秒、x276／y190／buttons1／mask2B按下；99083819、177452599微秒原24C31B返回BX低word1／CX276／DX190；99083854、177452633微秒、x278／y190／buttons0／mask1首次合法放開，差245291微秒。完整原查詢前後R／六段／flags保存並核算；終態callback12／12、IRQ28866／28866完成且非活動／未failed。共享20DDDB未命中，持久旗色writer仍未知。

原版120M到228DDC，R=[74 EE CF 89 2BD9B8 2BDA14 2C0B9C 2C040E]、段=[8 188 188 0 20 188]、flags216h；原PNG仍SELECT BANNER COLOR，與初版未推進PNG逐位元相同。PNG SHA-256 fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11、RGB 8f140b5eb4f65e2ed42744f269e5e2505a395feefb91b7dc2c829a81e398cb97。120M完整十筆表／FPU／VBE與只讀RAM收據保留，不從游標更新猜正式旗色。

最新source SHA-256 d131475620cda95f77a6c9846dc7a494a5a60bae778196c6ea49555a79b5d4f3；六有界旗標區塊／五guard及訊息逆轉後逐位元保持338，CPU／startup／provider／matcher／mouse callback保持335。較早六條基準與全套Go沿既有335／338收據，本輪未重跑、不冒充339重生。新來源的正常情境只重生一次；Docker每次300s／2GiB／2CPU／128pids／UID1000／network none。

容器內主要命令：

```text
go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe
DOSGOLEM_MOO2_CALENDAR_EPOCH=1996-01-01 DOSGOLEM_MOO2_HARDWARE_ESCAPE_STEP=44000000 DOSGOLEM_MOO2_MAX_STEPS=120000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_NEW_GAME_CLICK_AFTER_DISPLAY40=1 DOSGOLEM_MOO2_MENU_READY_CLICK=1 DOSGOLEM_MOO2_SETUP_ACCEPT_CLICK=1 DOSGOLEM_MOO2_RACE_HUMANS_CLICK=1 DOSGOLEM_MOO2_RULER_NAME_ACCEPT_CLICK=1 DOSGOLEM_MOO2_BANNER_RED_CLICK=1 DOSGOLEM_MOO2_VBE_FRAME_PREFIX=/src/workplace/moo2-339-red-frame DOSGOLEM_MOO2_VBE_PNG=/src/workplace/moo2-vbe-339-red.png /tmp/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game
python3 workplace/new-game-339-verify.py
python3 workplace/new-game-339-cli-verify.py
python3 workplace/new-game-339-source-verify.py
python3 workplace/new-game-339-backlink-verify.py
```

核算PASS：99M額外輸入前7708原列及28既有PNG保持，原表／globals／header／完整核心／RGB與獨立readonly ruler一致，首次AX3輸出與正常順序吻合。22新CLI拒絕與120M／100M正對照、同binary舊338 17拒絕與正對照、76項文件回填、新339 26缺證據負例、338 26負例及兩CLI通過。只證明原裝置／查詢契約；原選色消費與下一頁仍未證實。

| 本機忽略來源／收據／核算 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-probe-339-input.txt.gz | 3f1e74f87087527db0dab08a7ae877015f021bf26b098c18e7d067c52a64d93c |
| workplace/dosgolem/workplace/moo2-banner-339-input.go | f30e296ee52f359a1cf04308ffa7c10bd643d57a9e70239d39a0eede912017e8 |
| workplace/dosgolem/workplace/new-game-339-input-verify.py | 5e86ed732e266dacec88154f2c92fcfd37196c76fa4ccbd83874e2a6b4224341 |
| workplace/dosgolem/workplace/new-game-339-input-tests.txt | 70c2f1af8e5a1edfd4a5088311d1db6108ed0c5aaa85c68cf31de637d9c14794 |
| workplace/dosgolem/workplace/moo2-banner-339-early.go | 4e628d389e2910f7dea1c5bdacff3372e863b49524487125c4dfff3586268662 |
| workplace/dosgolem/workplace/moo2-probe-339-early.txt.gz | a88b8631f68504f9a373f07a1a77910736ca4a58e09285cc9981369ee0483cc4 |
| workplace/dosgolem/workplace/moo2-vbe-339-early.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-339-early-parity-tests.txt | 9cb29e6a9e5947c177fc20d4beff735caad31af02b8c8e97f28fe04068360237 |
| workplace/dosgolem/workplace/moo2-banner-339-extended.go | 6c63547a372fd329bc58351928d0e8dc2a7806a45461ce9d5a049e659c05fe63 |
| workplace/dosgolem/workplace/moo2-probe-339-extended.txt.gz | 559b388072eb7c8d9976b8de4f9c137b2d17dde9ef748630ce7a887100b1df8b |
| workplace/dosgolem/workplace/moo2-vbe-339-extended.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-339-extended-parity-tests.txt | 2030fffc23bec8d03e485feaab79e3f97847720dfdfd149c2c3a1ed23fc8b478 |
| workplace/dosgolem/workplace/moo2-probe-339-red.txt.gz | 9505df160613d748632e9e43a7e47e73945e9ed56948a9e66f951fbd9ed49965 |
| workplace/dosgolem/workplace/moo2-vbe-339-red.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-339-verify.py | 9e0f1b9faaaf01f793b8c899146d64d285fe33e3c7327487ec3998dd46fecfcb |
| workplace/dosgolem/workplace/new-game-339-parity-tests.txt | f89eb1f61495c194d3889c4536b0bcedd720189ea957b9431dda1c108a2d26c8 |
| workplace/dosgolem/workplace/new-game-339-cli-verify.py | d4b4430e1654c5b2e1170394e19a05642ce229d4658e3ce44084c6b599a8b441 |
| workplace/dosgolem/workplace/new-game-339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |
| workplace/dosgolem/workplace/new-game-339-source-verify.py | 80b75e90c3304f4259a52697c2b0d757de7c474881d2fa6461b0a4b56e0e37d4 |
| workplace/dosgolem/workplace/new-game-339-source-tests.txt | 5b209298ebaf57ffe253112aa7357875b4443d71c126a9d9f8f1038821fc7f48 |
| workplace/dosgolem/workplace/new-game-339-backlink-tests.txt | 3499c8deb32ad69c4bfb6acea4332039308806d913a265b3cec0cd34463fc4b1 |
| workplace/dosgolem/workplace/new-game-339-backlink-verify.py | e5e11057102f59e863b5b21491248f275500049ce6933da9b102eecbc32650a5 |

原來源與22份新收據1000:1000，gofmt／Git差異通過，工具root-owned與誤建.md目錄空，Docker兩工作區掛載篩選空。原版素材／PNG／LOG／RAM留本機忽略workplace，不公開。

**未知與下一步**：只保存首個原pressed查詢後最多192個非callback／非IRQ的正常GUI步、原事件／選擇及返回框架，解釋正常輸入未推進的原因；不再提高cap、移動時點或盲重送。持久名稱／旗色writer、typed種族特性、完整開局、正式RNG及remake同狀態未知。主庫RE-first及完成整款remake／中文化目標保持。

### 2026-10-03 原GUI按鍵返回與正常放開，旗幟選色仍未完成

主庫起點3526b6da8334626a992cfa3c13689b318befb664，工具起點7d569e0392fd9061576da8395d3c8b4d10e0886d；工具0b141e6cf2a00d86c028054c68d247c9e526a5c0已推送github隔離分支，遠端回讀一致，未推本機origin。公開[340原GUI按鍵消費](https://github.com/wicanr2/dosgolem/blob/0b141e6cf2a00d86c028054c68d247c9e526a5c0/docs/spec/340-moo2-banner-pressed-consumer.md)限定CONFORMED只涵蓋原按鍵RET返回與正常放開，原選色未完成。339與索引／回填guard同次更新，不把這項工具輸入驗收算入主庫玩法分母。

來源與位址：官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，Go1.24.13與映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。正版ZIP SHA-256 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，官方patch ZIP 908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；每次唯讀掛載、重建新鮮417根檔及MOX.SET，固定日期1996-01-01不是seed。

**已證實，339的第二次查詢與原返回0**：可丟棄readonly prototype只增加五個有界trace區塊，首次pressed查詢後臂192個前後非callback／IRQ正常步。原99083851從22F26F RET到2139D2，234A48 STI後原放開；20DB56再CALL214075。99093510原24C31B第二次AX3返回BX低word0，2140D9原66A1E4382A00讀DS:2A38E4為0，2140DF AND3仍0，2140F9取原local，99093560的214104 RET到20DB5B返回0；SS188:ESP2BD998原top5BDB2000。20DB5B TEST AX後20DB5E原0F855B050000 JNZ不跳，落20DB64。這只是原按鍵返回鏈，正式旗色writer未證實。

**已證實，340同原框架返回1後才放開**：原框架證據審查後READY，正式source只增三個有界readonly觀測與一個release前置，原CPU指令自行返回。99M／177207342微秒按下仍x276／y190；99083819首次AX3仍返回1。99083999原214104 RET到20DB5B，EAX低word1、SS188:ESP2BD998 top5BDB2000，完整R只ESP加4、段=[8 188 188 0 20 188]與flags202h不變、stack readable／readonly／valid為true；99084000／177452779微秒首次合法mask1放開，x278／y190／buttons0，差245437微秒。沒有代寫原CPU／RAM／選擇，不改按下時點或重送。

**負結果與未知**：原版120M到228E00，R=[74 EE BE 89 2BD9B8 2BDA14 2C0B58 2C03CC]、段同上、flags287h，callback12／12與IRQ28866／28866完成、非活動／未failed；沒有新CPU拒絕，共享20DDDB未命中。原PNG人工確認仍SELECT BANNER COLOR，與339終圖逐位元一致，SHA-256 fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11、RGB 8f140b5eb4f65e2ed42744f269e5e2505a395feefb91b7dc2c829a81e398cb97。正式340沒有記錄20DB5E後的實際分支，不能只據返回1猜選色結果。

private原trace剝除194新增列後全部10012原339列與32PNG保持；192筆完整、readonly、前後排除callback／IRQ，budget一次臂後耗盡，剝除五區塊後source逐位元保持339。正式99M按下前7708原列與28PNG保持，完整550bytes表／核心／RGB與首次AX3一致；RET與release順序核算通過。核算首次誤從GUI checkpoint讀IRQ欄位，改讀同收據setup_table_snapshot後乾淨重讀PASS，沒有重跑原版，也不是產品缺陷。

正式source SHA-256 bef46d84f24b674d90e3e84e535485a538bf3e2dba9c0e41f4c08fc27898ab56；三觀測區塊及release閘門逆轉後全部來源保持339，CPU／平台未改。正式同binary舊338 CLI17拒絕及正對照、339 CLI22拒絕及120M／100M正對照通過；77項規格回填、340新增28缺證據負例、338／339各26負例及三CLI通過。未重跑未受影響六個舊基準／Go全套，不冒充本輪重生。

容器內主要命令：

```text
go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe
DOSGOLEM_MOO2_CALENDAR_EPOCH=1996-01-01 DOSGOLEM_MOO2_HARDWARE_ESCAPE_STEP=44000000 DOSGOLEM_MOO2_MAX_STEPS=120000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_NEW_GAME_CLICK_AFTER_DISPLAY40=1 DOSGOLEM_MOO2_MENU_READY_CLICK=1 DOSGOLEM_MOO2_SETUP_ACCEPT_CLICK=1 DOSGOLEM_MOO2_RACE_HUMANS_CLICK=1 DOSGOLEM_MOO2_RULER_NAME_ACCEPT_CLICK=1 DOSGOLEM_MOO2_BANNER_RED_CLICK=1 DOSGOLEM_MOO2_VBE_FRAME_PREFIX=/src/workplace/moo2-340-red-frame DOSGOLEM_MOO2_VBE_PNG=/src/workplace/moo2-vbe-340-red.png /tmp/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game
python3 workplace/new-game-340-verify.py
python3 workplace/new-game-340-return-verify.py
python3 workplace/new-game-340-source-verify.py
python3 workplace/new-game-340-backlink-verify.py
```

私有trace與正式新source各由乾淨輸入重生一次，Docker每次300s／2GiB／2CPU／128pids／UID1000／network none、原ZIP／patch唯讀。原素材、PNG、RAM、LOG與可丟棄probe留本機忽略workplace，不公開。

| 本機忽略來源／收據／核算 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-poll-340.go | 5bb1ec1da69f80a4b7831e05f0a48aa355d8a21383a6c890ae057289d00709ea |
| workplace/dosgolem/workplace/moo2-probe-340-trace.txt.gz | dec37ddf557a2e35092e006d91b529966064a8f3f81efa5132f19f571d2b59d9 |
| workplace/dosgolem/workplace/moo2-vbe-340-trace.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-340-verify.py | 3ecb7f2b3562c7a76967baaa2f8e599989ecaecc52971195e76171b2d1cce931 |
| workplace/dosgolem/workplace/new-game-340-tests.txt | 068dee19b4d062a8015fd0cf0cf9d339491f4b75f0ba4cd869e4814127500ade |
| workplace/dosgolem/workplace/moo2-probe-340-red.txt.gz | a57ac9a85155154f0e026391a09fd547791b95b56cf2afa488212329ee36c02c |
| workplace/dosgolem/workplace/moo2-vbe-340-red.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-340-return-verify.py | 39147738f42c0cd55e74387cab7d0326e3e0dbb65c309561f369033402cdb434 |
| workplace/dosgolem/workplace/new-game-340-return-tests.txt | a12ab0e6b44343af80eaf272c9804adfc95e29b19bb8fe03eb7c5061ee9cbe09 |
| workplace/dosgolem/workplace/new-game-340-source-verify.py | ddc521e2f33c0f9768459786ead5ebd0492024cc0052cfe5820d6e2a0fa45d57 |
| workplace/dosgolem/workplace/new-game-340-source-tests.txt | 5b856aaa7bf692c6070ed6c5a414261d7a0ad72ac44d14d67e1257ff39e904de |
| workplace/dosgolem/workplace/new-game-340-backlink-verify.py | 3f693b0fce6d2002a94d6e6ac7b7aa42adefe7b95b62ac689fa6e17f1634d1e7 |
| workplace/dosgolem/workplace/new-game-340-backlink-tests.txt | 5681c96b742c39d09b81e95d20a419670e70f1081cc30c9efb3d876d991af732 |
| workplace/dosgolem/workplace/new-game-340-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-340-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |

新檔1000:1000，gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空，本輪沒有遺留容器。

**下一步**：只保存340同輸入下原20DB5B之後最多192個非callback／IRQ正常步、實際20DB5E分支與原返回框架。原選色消費、持久名稱／旗色writer、typed種族特性、完整開局、正式RNG及remake同狀態未知；保持120M、原輸入與READY閘門，不盲調cap或重送，不深挖原helper。主庫RE-first與完成整款remake／中文化目標保持。

### 2026-10-03 原後段按鍵返回已驗，續行揭露85 82記憶體TEST缺口

主庫起點f93e17de0f1268a2bd6251da3afe1fc654ff6c4b，工具起點0b141e6cf2a00d86c028054c68d247c9e526a5c0；工具b07cda7188d7139da612f143ba68ea13cca21958已推送github隔離分支，遠端回讀一致，未推本機origin。公開[341原後段按鍵返回](https://github.com/wicanr2/dosgolem/blob/b07cda7188d7139da612f143ba68ea13cca21958/docs/spec/341-moo2-banner-after-gui-return.md)限定CONFORMED只涵蓋原正常分支／後段按鍵返回與放開，原旗色結果／下一頁／完整開局未驗。340及索引／回填guard同次修正，不計入主庫玩法分母。

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。Go1.24.13、映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。正版ZIP 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，官方patch ZIP 908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；原ZIP／patch唯讀，各次新鮮417根檔／MOX.SET／固定EXE。固定日期1996-01-01不是seed。

**已證實，原GUI非零分支與再次查詢**：第一獨立192步在原RET20DB5B返回1後臂一次，99093561 TEST AX,AX、99093562原20DB5E JNZ20E0BF確實跳。20E0BF再CALL214075，99093649原24C31B AX3返回BX0／CX278／DX190；99093699原214104 RET20E0C4返回0，原topC4E02000與完整R只ESP+4已驗。20E0CB CMP EAX,2後20E0CE JNE20E12B跳，原X getter213ABA返回139至20E130，20E133 CALL213AE7返回點是20E138。只跨過Watcom helper，不深挖內部。

**已證實，原位置返回與後段按鍵分支**：第二獨立192步從第一次原20E138入口開始。99093771原Y返回AX190，SS188:EBP2BDA04、原SS:2BD9C4 locals偏移8／12為139／190；20E13B CMP DS:26C4E2 word,0為等，20E143 JNZ不跳，20E145寫word1，20E151／20E15A寫原139／190至DS:26C4DE／26C4E0。這些只記原定位，正式旗色writer未知。

20E160 CALL214075，99093865原24C31B再查按鍵返回0；99093915原214104 RET20E165返回0，SS188:ESP2BD998 top65E12000、完整R只ESP+4、六段與flags246h保持。20E165 TEST AX,AX後，99093917原20E168 bytes0F847D030000 JZ到20E4EB。已證實按鍵0走後段分支；保持到這個原caller返回1能否選色，在實作前仍標強推論，沒有猜補持久資料。

兩份private source各剝除五個區塊後逐位元保持340；每份剝除194新增列後全部9968原340列與32PNG保持。兩份各192筆完整readable／readonly、前後非callback／IRQ、一次臂並耗盡；第二份另保存原SS:EBP-64 locals64 raw。第一次prototype的post位置在建置前修正，沒有以錯誤觀測跑原版；runtime只各重生一次。完整兩包與最小bytes／stack／分支核算通過，見24份收據表。

**READY後實作與已驗輸入**：正式source只增三有界後段RET觀測及一個本fixture release前置，不改原CPU／平台／主庫玩法。99000000按下與99083819首個AX3、99083999首個RET20DB5B保持。99084354原214104 RET20E165返回EAX低word1，SS188:ESP2BD998 top65E12000、完整R只ESP加4、六段與flags202h保持、readable／readonly／valid true、error nil。99084355／177453134微秒首次合法mask1放開，原177207342微秒按下，相差245792微秒；x276／190按下、x278／190放開、target8:2136D1與原callback11／11完整前置不變，終態callback12／12。

**實際CPU負結果**：原版未到120M上限，step99415524原184694 bytes85 82 19 52 26 00拒絕，error=TEST dword ModRM 82 尚未支援；拒絕前EAX=EDX=0、flags202h、DS／ES／SS188h。CPU讀opcode與ModRM後error行EIP184696，原輸入位置仍184694，兩位址不混用。靜態解碼為TEST dword [EDX+265219],EAX，正式source欄位與SETNE／RET消費仍待下一切片。終圖人工確認全黑，PNG SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622，RGB 0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366等於640×480×3全零像素。共享20DDDB未命中，原旗色結果／持久writer／下一頁未驗。probe exit0仍有guest_cpu_stop／step_error，不當正常流程通過。

正式source SHA-256 16f40cba5262cb7765fdfd8c8478b95d31a7ff38f40438dc255e2d96b6fcef2c，三區塊／一個release逆轉後逐位元保持340。首次source核算腳本逆轉字串誤寫，修正腳本後重讀同來源PASS，未重跑原版，不是產品缺陷。正式7708原列／28PNG／完整550bytes表、核心／RGB、兩RET／首次合法release與CPU負結果核算PASS，沒有把它稱成選色完成。

每個原版情境Docker300s／2GiB／2CPU／128pids／UID1000／network none；兩份探索與正式fixture各重生一次。容器內主要命令：

```text
go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe
DOSGOLEM_MOO2_CALENDAR_EPOCH=1996-01-01 DOSGOLEM_MOO2_HARDWARE_ESCAPE_STEP=44000000 DOSGOLEM_MOO2_MAX_STEPS=120000000 DOSGOLEM_MOO2_SEPARATE_DOS=1 DOSGOLEM_MOO2_NEW_GAME_CLICK_AFTER_DISPLAY40=1 DOSGOLEM_MOO2_MENU_READY_CLICK=1 DOSGOLEM_MOO2_SETUP_ACCEPT_CLICK=1 DOSGOLEM_MOO2_RACE_HUMANS_CLICK=1 DOSGOLEM_MOO2_RULER_NAME_ACCEPT_CLICK=1 DOSGOLEM_MOO2_BANNER_RED_CLICK=1 DOSGOLEM_MOO2_VBE_FRAME_PREFIX=/src/workplace/moo2-341-red-frame DOSGOLEM_MOO2_VBE_PNG=/src/workplace/moo2-vbe-341-red.png /tmp/moo2-probe /tmp/game/ORION2.EXE --game-dir /tmp/game
python3 workplace/new-game-341-verify.py
python3 workplace/new-game-341-tail-verify.py
python3 workplace/new-game-341-semantic-verify.py
python3 workplace/new-game-341-tail-semantic-verify.py
python3 workplace/new-game-341-formal-verify.py
python3 workplace/new-game-341-source-verify.py
python3 workplace/new-game-341-backlink-verify.py
```

同新版binary舊338 CLI17拒絕及正對照、339 CLI22拒絕與120M／100M正對照；78項規格回填、341新增29缺證據負例、340 28負例、338／339各26負例與兩CLI通過。未重跑未受影響六個舊基準／Go全套，CPU保持335，不冒充本輪重生。

| 本機忽略來源／收據／核算 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-gui-after-341.go | 567a769ac4d0419cb9f82745652705577015ea56bc83773bc893b8c312f83973 |
| workplace/dosgolem/workplace/moo2-probe-341-trace.txt.gz | 90f6c34305a588cc42759639f771065b3c59ec4128fef445c0e9b91b219cee47 |
| workplace/dosgolem/workplace/moo2-vbe-341-trace.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-341-verify.py | 250b770595bd52a5be409fab5e277392f5e0c85d214e247b500edd46bff44b4d |
| workplace/dosgolem/workplace/new-game-341-tests.txt | edde141d6857b5fbf8d77b78f92f94aea84589888c611e8bcef807536e93d6b6 |
| workplace/dosgolem/workplace/new-game-341-semantic-verify.py | ffee726e9651ebb7ed2aca47c7b0ce2e94a0dfb9acb49099087c0269cd199c32 |
| workplace/dosgolem/workplace/new-game-341-semantic-tests.txt | 8e96b2ccf83403b1ebe0f8ab3edc408bcc7774f14a291dff9d1edb41c4099da6 |
| workplace/dosgolem/workplace/moo2-gui-tail-341.go | d6c6213e03b1f427da6feef78180255ff2ccc22ea7e396f7543bc02acadae31d |
| workplace/dosgolem/workplace/moo2-probe-341-tail.txt.gz | 5add8c37fbd9849eddfca09950f4d0c4dfca077cd439f038cdf144633ec55618 |
| workplace/dosgolem/workplace/moo2-vbe-341-tail.png | fdcfce7eb7a39bfa24641a15309ec035208efa205c394173c5cdde37e5e6ec11 |
| workplace/dosgolem/workplace/new-game-341-tail-verify.py | 51ea3347af4071a2a7dc05e297a7fd74d88f46cdbaa5844375fe01eff97a25b7 |
| workplace/dosgolem/workplace/new-game-341-tail-tests.txt | 6ca1dc5ae73a71e2a9487dfe7d4bbe8d14c0702c5f2b2b48c7ba8b3279cac498 |
| workplace/dosgolem/workplace/new-game-341-tail-semantic-verify.py | 970c4d2557d1061103133c88e51f2a66d8c08cb4960f434fe3d6c6a3ddce52bc |
| workplace/dosgolem/workplace/new-game-341-tail-semantic-tests.txt | c3e9f640232e0c842ea31d7b89e324a1c561d035e8deef7c719541db0d7ccfa1 |
| workplace/dosgolem/workplace/moo2-probe-341-red.txt.gz | 9e6dc06295aaf43383350abb53e8e9a4e0f883ff261ebbc54cc728e2173e7656 |
| workplace/dosgolem/workplace/moo2-vbe-341-red.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| workplace/dosgolem/workplace/new-game-341-formal-verify.py | 8d5b4e13b276d4453c5dc6035fb38d6e5a515b15eb1f788c02fbd3aea56d4c83 |
| workplace/dosgolem/workplace/new-game-341-formal-tests.txt | ed93fd3d40d0ed61dc392c709a21903e8d6c71b08c2f1a1f454468b17fcfe760 |
| workplace/dosgolem/workplace/new-game-341-source-verify.py | 3e536a4eb29413605a84c3c641a1bc5a5c631d50dad59f73143e6b4c7e8afc13 |
| workplace/dosgolem/workplace/new-game-341-source-tests.txt | 76ee9a8e95dc52a7624e81dee4afbce0c413d51c8f2221f83034d821aed13284 |
| workplace/dosgolem/workplace/new-game-341-backlink-verify.py | 0acc309787fa1c2e1bda21ba0a8d65810bd2e55d27f0454160159d66a98be58f |
| workplace/dosgolem/workplace/new-game-341-backlink-tests.txt | 878c94271001c71f87e930d866cd28f5429285707c00484543588ded3b48f7aa |
| workplace/dosgolem/workplace/new-game-341-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-341-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |

24份來源／收據與新檔1000:1000，gofmt／Git差異通過，工具root-owned／誤建.md目錄空，Docker兩工作區掛載篩選空，沒有本輪遺留容器。原素材／PNG／LOG／RAM與private prototype留本機忽略workplace，不公開。

**下一步**：先核對CPU386 TEST與記憶體operand契約，保存原85 82 @184694有效位址／source／mask／flags及SETNE／RET最小消費，經DRAFT→READY補支援，再按同341輸入續行與固定EXE完整CPU測試。保持120M與原輸入，不因全黑調GUI座標、cap或重送。原旗色結果、持久名稱／旗色writer、typed種族特性、下一頁／完整開局、正式RNG與remake同狀態未知，主庫RE-first與整款remake／中文化目標保持。

### 2026-10-03 記憶體TEST已接通，原版正常進到宇宙生成畫面

主庫起點34ace3c4ed5b4418882f6ab83cdf176a3cd325c7，工具起點b07cda7188d7139da612f143ba68ea13cca21958；工具0c6e871167259c382c6dac2288ace552c33e2319已推送github隔離分支，遠端回讀一致，未推本機origin。公開[342記憶體TEST](https://github.com/wicanr2/dosgolem/blob/0c6e871167259c382c6dac2288ace552c33e2319/docs/spec/342-cpu386-test-dword-memory.md)限定CONFORMED，340／341與索引／守衛同次回填。這是CPU能力與原版玩家流程證據，不計入主庫玩法分母。

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。Go1.24.13固定Docker映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。正版ZIP 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，官方patchZIP 908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；各次新鮮417根檔，MOX.SET553bytes SHA-256 bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。1996-01-01固定日期不是RNG seed，來源欄位用途未知。

**已證實，未改CPU初態**：99415524原184694完整R=[0 0 0 0 2BDB68 2BDB70 0 0]，段=[8 188 188 0 20 188]、flags202h；DS188:265219四bytes01000000即1、EAX0，SS188:ESP2BDB68四bytesE3461800返回1846E3。private三區塊唯讀budget最多三步，未改CPU仍同TEST拒絕、EIP184696；全部7849原341列／29PNG保持。公開Intel TEST契約與真實來源充分後，342由DRAFT審查READY才寫通用CPU能力。

**已證實，原三步消費**：CPU85 memory沿decodeAddress32及readSegment32唯讀，不按遊戲位址特例。99415524 TEST交集0、flags246h、EIP18469A；99415525 SETNE AL0、EIP18469D；99415526 RET只ESP加4到2BDB6C、EIP1846E3，其他R／六段／FPU與RAM、四byte來源／stack保持。AF未定義，清除只沿工具模型。三筆readable／readonly／step_ram_unchanged真、error nil，callback12／12、IRQ22676／22676非活動，pending0。正式新入口前7789原341列／28PNG保持；三observer逆轉後probe保持不可變341，CPU逆轉新memory85分支後保持335。

**已證實，原正常玩家畫面**：同341原99M按下／99084355首次合法放開、原全部前置與120M cap，原版實際進到640×480「Generating Universe...」。人工確認原框、綠色格線及文字，終圖已非全黑；PNG e975c476784977be084da3abb7601363bf290c74b96b26d56fc1af0994dd7dd0，RGB353171a9bce8ad97f55e3ee444a0f9f8017bf44e531b5616f42393ff976ddf69。只確認顯示，不猜持久旗色值或生成完成。

**已證實，新CPU負收據**：113628909原17D536 bytes0F 9E C0 88 45 FC 89 C8 99 31 D0 29 D0 0F BF 55拒絕，error=0F 9E 尚未支援。指令起點17D536、解碼後EIP17D538，EAX28／EBXFFFFFFC2／ECXFFFFFFDA／EDX0、flags206h；終態完整R=[28 FFFFFFDA 0 FFFFFFC2 2BDA08 2BDA38 0 1]、段=[8 188 188 0 20 188]。probe shell exit0仍有step_error，不能當完整開局通過。callback12／12、IRQ26735／26735非活動，原共享20DDDB未命中。

**驗證與環境失敗分類**：新自製測試以287獨立逐bit交集／五旗標核算，全部來源、bit／低byte／高位、ModRM／SIB、DS／SS分離／readonly／bus逐byte／段末／繞回／非對齊、截短與未知prefix／完整核心保持與零writes通過。SETNE／RET兩方向另有測試。初CPU簽名使用錯誤於建置拒絕、未跑原版；修正後窄測0.909s PASS。首次全套被忽略workplace多份探索main污染，CPU386本身61.132s PASS而全套FAIL；同映像／固定EXE／命令於/tmp/test-src乾淨版控輸入加本輪自製測試重跑全套PASS，CPU38655.710s、machine2.914s。外部8088實機語料缺檔，CPU386是公開契約推導及原版續行證據，不冒稱硬體語料通過。初版formal核算包含舊stop診斷，修正入口前邊界後重讀同收據PASS，沒有重跑原版。

實際入口：窄測go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestTESTDwordMemory|TestTESTDwordRegisterImmediate|TestORDwordMemory' -count=1 -v；全套DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE go test -p 2 -buildvcs=false ./... -count=1。乾淨輸入以git ls-files -z | tar --null -T - -cf - | tar xf - -C /tmp/test-src、複製新自製測試、沿現存testdata建立，歷史探索檔與其hash保持。正常probe go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿341全部環境：固定calendar／44M Esc／分離DOS／原NEW_GAME／MENU_READY／SETUP_ACCEPT／RACE_HUMANS／RULER_NAME_ACCEPT／BANNER_RED旗標，MAX_STEPS120000000。先private初態、再READY後正常重播各一次。Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。

79項回填、342新增25缺證據負例及340另兩負例、341的29／340的28／338與339各26負例、舊338 CLI17拒絕與正對照、339 CLI22拒絕與120M／100M正對照通過。不另重跑未受影響六舊情境。CPU SHA-256 330baaf9917f4534906797e9eb484136b7297fc6a94343252db9b6f8da97c5f7，自製測試58bfb07a0836f04c5600a214ac4c53efac67cddc2a0ed45698622b58a1357e5f，probe ebc395db43be26ea4423639f785ceb2f0e8753f44ef5f90cf6cf7f45a5434481。全部原圖／LOG／RAM與private原source留忽略workplace，沒有提交原素材。

| 本機忽略來源／收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-test-memory-342.go | ebc395db43be26ea4423639f785ceb2f0e8753f44ef5f90cf6cf7f45a5434481 |
| workplace/dosgolem/workplace/moo2-probe-342-input.txt.gz | 923a2fe5b6f669c83579aecfb3b2b14e84fc143948600855133dbdc85e12845e |
| workplace/dosgolem/workplace/moo2-vbe-342-input.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| workplace/dosgolem/workplace/new-game-342-input-verify.py | 989a598e4839891f998b4fc0f13b2140fb6ab56d5952160a70a2b6d3950deb79 |
| workplace/dosgolem/workplace/new-game-342-input-tests.txt | 544fe5b937e8859b5d01a98c1610047b75b3e2fca75e793da0f8eeb14930c1eb |
| workplace/dosgolem/workplace/moo2-probe-342-red.txt.gz | 9181e24bfbd29931985b3e3966c5081cd6b6f367d6faaac8d13e33b9dd155181 |
| workplace/dosgolem/workplace/moo2-vbe-342-red.png | e975c476784977be084da3abb7601363bf290c74b96b26d56fc1af0994dd7dd0 |
| workplace/dosgolem/workplace/moo2-342-cpu-narrow-tests.txt | 44fe2a818c8adf0c69205fc8f6e572d21b83b18f32a40c27d3d521333957b807 |
| workplace/dosgolem/workplace/full-test-342.txt | 2877e41d8c9448a37a589fe7e6e35ed8be2059b21cb6a4c5cac7d86101ca663c |
| workplace/dosgolem/workplace/full-test-342-unclean.txt | 06d8759f8cd6a26869658099871025f69b90c7c2ca6d94493f334a91e53fbfba |
| workplace/dosgolem/workplace/new-game-342-formal-verify.py | d6598337b5a5ca39bb98650bf39b1f1065d3f02c1a5778be55072f3614e47872 |
| workplace/dosgolem/workplace/new-game-342-formal-tests.txt | c2922c3e4f2ae1fbf278e3eaffca267bafaf8fdf23e32fb7b26d07b3a3f15dde |
| workplace/dosgolem/workplace/new-game-342-source-verify.py | 990afbbaf981fb20697695cfce90bfbeb179512a9bf6fb0a97033fc170042571 |
| workplace/dosgolem/workplace/new-game-342-source-tests.txt | 600a3b0bd26b0c5f0019a819b844c50ca9cd07e8b01aad46872d5324d3059365 |
| workplace/dosgolem/workplace/new-game-342-backlink-verify.py | b783c3b7d664d386db0c9e7d05d6ccdae3ec0f6bd330abf348bb9553b483198e |
| workplace/dosgolem/workplace/new-game-342-backlink-tests.txt | 4a23e596a6b1b0adbc563ba4d22dbed8a872575bdebfd49b3158b09a891eebf1 |
| workplace/dosgolem/workplace/new-game-342-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-342-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |

工具root-owned／誤建.md目錄空，main既有2437root-owned檔／272目錄保持，本輪source／收據1000:1000，未新增root-owned／誤建.md；兩工作區掛載篩選Docker容器空。gofmt／Git差異／版權邊界通過。

**未知與下一步**：只核對0F9E SETLE公開條件與既有SETcc，保存原17D536完整初態／flags／AL與下一byte store最小消費，DRAFT→READY後補通用CPU能力，按同輸入／120M與固定EXE全套續行。正式姓名／旗色writer、typed種族特性、生成完成／完整開局、正式RNG／人耳與remake同狀態未驗；主庫玩法RE閘門保持，沒有重開已完成CPU切片。整款remake／中文化仍是活躍目標。

### 2026-10-03 SETLE與原SS byte實際寫入已驗，續行揭露SETG缺口

主庫起點ea13175382677216babef8a9b21801e2273d7f44，工具起點0c6e871167259c382c6dac2288ace552c33e2319；工具eddee109e0d0d59e311f5c30e26961f84ee36560已推送github隔離分支，遠端回讀一致，未推本機origin。公開[343 SETLE與SS byte寫入](https://github.com/wicanr2/dosgolem/blob/eddee109e0d0d59e311f5c30e26961f84ee36560/docs/spec/343-cpu386-setle-byte-register.md)限定CONFORMED，只閉合標準CPU與原兩步，生成完成／完整開局、正式writer與remake同狀態未驗。342／341／340與索引及guard同次回填，不計入主庫玩法分母。

官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，所有原位址dosgolem_high_le；Go1.24.13固定Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。原ZIP 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，官方patchZIP 908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；各次新鮮417根檔／MOX.SET553bytes SHA-256 bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。固定1996-01-01不是RNG seed。

**已證實，原初態與READY**：343先DRAFT。113628909原17D536完整R=[28 FFFFFFDA 0 FFFFFFC2 2BDA08 2BDA38 0 1]、六段=[8 188 188 0 20 188]、flags206h。SS188:EBP2BDA38-5的四bytes01000000，目的SS:2BDA34原byte0；stack SS188:ESP2BDA08四bytes040F0000。private兩步budget只臂一次，未改CPU同拒絕、EIP17D538，全部8237原342列／31PNG保持，完整CPU／FPU／VBE與RAM的observerreadonly通過。Intel現行SDM的SETLE為ZF=1或SF≠OF，所有flags保持；80386鏡像SETLE列and錯、同行SETNG用or，採現行公開契約。完整原來源與契約充分後才READY，欄位用途未知。

**已證實，原SETLE與byte消費**：CPU只增加裸0F9E八個byte寄存器。113628909原SETLE條件false，AL28→0、EAX28→0、EIP17D539；113628910原88 45 FC成功，EIP17D53C，唯一setle_bus_write為linear2BDA34／value0／error nil，位於兩筆消費紀錄間。目的是0→0，所以只有RAM不變不足以證實write；既有Bus observer在原Write8只forward一次後，於兩步budget內保存實際address／value／error。兩步其他R／六段／FPU／所有flags206h／RAM／四byte窗／stack保持，callback12／12、IRQ26735／26735非活動，pending0／readonly真。private三／正式五observer區塊逆轉後source保持342，CPU逆轉新增條件式／註解後保持342；沒有代寫資料或EIP。

正式新CPU入口前8177原342列／30PNG保持，原341全部press／release與原342 TEST三步保持。原input／1996calendar／120M cap未改；未因0→0假驗收調參，也沒有原版重送或深入helper。

**已證實，新CPU拒絕與畫面邊界**：113628944原17D5A0 bytes0F 9F C0 88 C2 80 7D FC 00 75 0E 80 7D F4 00 75拒絕，error=0F 9F 尚未支援。原起點17D5A0、解碼後EIP17D5A2，EAXE6／EBXFFFF0000／ECX3／EDX1FA、flags293h。尚未120M，probe shell exit0不能當正常開局通過。終640×480原PNG逐位元保持342「Generating Universe...」，SHA-256 e975c476784977be084da3abb7601363bf290c74b96b26d56fc1af0994dd7dd0、RGB353171a9bce8ad97f55e3ee444a0f9f8017bf44e531b5616f42393ff976ddf69；沿342相同hash的人工確認，不把同一畫面重算新玩家功能。原共享20DDDB未命中，持久writer未知。

**獨立規則驗證**：2,097,152組八byte目的／unused reg欄／256初byte／64算術flags／兩context，以字面8列真值表與little-endian byte陣列驗所有core／flags與鄰居保持、零writes。77筆32位值全部配對×八目的共47,432組CMP→SETLE，以int64有號差／溢位範圍核算ZF／SF／OF，再以數學≤核算目的，不呼叫CPU condition helper。原SETE／SETNE保持、截短／memory全ModRM／11prefix／其他SETcc仍拒絕；下一SS byte兩方向、DS分離、段末／readonly／未知／Bus寫失敗回歸通過。外部386／8088實機語料未取得，不宣稱硬體語料或逐週期對拍。

實際命令：go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestSETLERegister|TestTESTDwordMemory|TestTESTDwordRegisterImmediate' -count=1 -v PASS，cpu3864.352s；固定DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE的go test -p 2 -buildvcs=false ./... -count=1 PASS，cpu386189.431s、machine5.732s。全套從/tmp/test-src乾淨版控輸入加本輪新自製測試與現存testdata執行，以git ls-files -z | tar --null -T - -cf - | tar xf - -C /tmp/test-src與cp internal/cpu386/setle_byte_register_test.go建立，避免忽略探索main污染；沒有重寫舊探索檔或其hash。

兩次原版為private初態與READY後正式CPU，各一次fresh417根檔／固定EXE／MOX.SET。go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿341固定calendar／44M Esc／分離DOS／NEW_GAME／MENU_READY／SETUP_ACCEPT／RACE_HUMANS／RULER_NAME_ACCEPT／BANNER_RED旗標／MAX_STEPS120000000。Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。80項回填、343新增26負例、342的25與340另兩負例、341的29／340的28／338及339各26負例，舊338 CLI17無效與正對照、339 CLI22無效與120M／100M正對照通過。未重跑未受影響六舊情境。

CPU SHA-256 76f4b7b3f97156f9422d32e894e55579286f85d2c8b39a26eb90887fa22cc88f，自製setle_byte_register_test.go 91f13f6f7f5773b6f27367d3f94a9299600ce84f4437750a4f8390dbdff99d5c，正式probe c304dbac8a0689e83560529b1fb7acc4cd46ef87298c5077125ea2d2cbd6fca9。原PNG／LOG／RAM／private source留本機忽略workplace，不入Git。

| 本機忽略來源／收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-setle-343.go | f7bdd7a9f44b6b94aaa8f8430e9c0391d8bb8ff91dca988eb138805a76d49622 |
| workplace/dosgolem/workplace/moo2-probe-343-input.txt.gz | 3d7697eeec4e3bf79f6fa933c9c2461ca922a6383f2b9638f709cfaed6383aaa |
| workplace/dosgolem/workplace/moo2-vbe-343-input.png | e975c476784977be084da3abb7601363bf290c74b96b26d56fc1af0994dd7dd0 |
| workplace/dosgolem/workplace/new-game-343-input-verify.py | 8d45747a105e964ba682da28c7449403278b96e4c22a42c71060dd9a0b628f4a |
| workplace/dosgolem/workplace/new-game-343-input-tests.txt | 947caf2ec23f6bbaa20eb0892119ad8d9f304a5dddb43660b42144715a88fcb4 |
| workplace/dosgolem/workplace/moo2-probe-343-red.txt.gz | fffa5bf9edfa2a4c3fdbece2852a0069a3307fe6b96bf9ded086fb73cf35ed9b |
| workplace/dosgolem/workplace/moo2-vbe-343-red.png | e975c476784977be084da3abb7601363bf290c74b96b26d56fc1af0994dd7dd0 |
| workplace/dosgolem/workplace/moo2-343-cpu-narrow-tests.txt | 58fc527d088242338c7e26d45213492c39b3e9c5e2e6eb1b61e52deadf87c46d |
| workplace/dosgolem/workplace/full-test-343.txt | 025ca761a9bffc3a248275548ec218ceac829409b5ed63f54991057c97a8e19b |
| workplace/dosgolem/workplace/new-game-343-formal-verify.py | d894d095885725194902ad71eb43674fea881039d1b09ead0968d630a246e2d9 |
| workplace/dosgolem/workplace/new-game-343-formal-tests.txt | 02411ef852fd4264190bc32c83c36267888f0859cea877fe35fafc9ef84e17e4 |
| workplace/dosgolem/workplace/new-game-343-source-verify.py | ea31be2f0852cfb10aac272c595fc5cfc34726e833015620c9b759197dfec617 |
| workplace/dosgolem/workplace/new-game-343-source-tests.txt | 272def4e6dfc55e47d6889da34ad152d83d57d68be3fbff995794dc29ed7fcba |
| workplace/dosgolem/workplace/new-game-343-backlink-verify.py | 1cfd101d8f127aa499bbbfd768266b55dfd672ba6642202ea63cc14a60911b59 |
| workplace/dosgolem/workplace/new-game-343-backlink-tests.txt | e2ca5ba480a18704f04880ac5dbfc90cb5dbcdf2dd185996635a52833a183161 |
| workplace/dosgolem/workplace/new-game-343-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-343-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |

來源／收據1000:1000、gofmt／Git差異與版權邊界通過；工具root-owned／誤建.md目錄空，主庫既有2437root-owned檔／272目錄保持。本輪沒有新增root-owned／誤建.md；兩工作區掛載篩選Docker容器空，未留背景程序。

**未知與下一步**：以原17D5A0 SETG為入口核對剩餘標準SETcc register條件，保存完整初態／AL／flags與下一88 C2最小消費，走DRAFT→READY、同正常input／120M與固定EXE全套。正式姓名／旗色writer、typed種族特性、生成完成／完整開局、正式RNG、人耳及remake同狀態未驗。主庫玩法RE閘門保持，整款remake／中文化目標仍活躍。

### 2026-10-03 標準SETcc暫存器條件已補，原SETG／MOV通過，同120M無新CPU拒絕

主庫起點95dc2470e396bd902fa527c083d86a6f4a220d6c，工具起點eddee109e0d0d59e311f5c30e26961f84ee36560；工具d5127adc64a04af79796d933aec73413bbcfd824已推送github隔離分支且遠端回讀一致，未推本機origin。公開[344標準SETcc與原兩步](https://github.com/wicanr2/dosgolem/blob/d5127adc64a04af79796d933aec73413bbcfd824/docs/spec/344-cpu386-setcc-byte-register.md)限定CONFORMED，只涵蓋公開CPU契約與原實測SETG／MOV。340／341／342／343及索引／護欄同次回填，不計入主庫玩法分母。

官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原位址dosgolem_high_le。原ZIP 3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f，patchZIP908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；fresh417根檔、MOX.SET553bytes SHA-256 bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，固定1996日期不是RNG seed。

**已證實，輸入與原兩步**：343不可變收據fffa5bf9edfa2a4c3fdbece2852a0069a3307fe6b96bf9ded086fb73cf35ed9b已有完整R／六段／flags／FPU，344重查0F9F先fetch extended後拒絕、未fetch ModRM或發布目的／flags，資料足夠後READY；不是新StepHook之前caller捕捉，不多跑相同拒絕。正式113628944原SETG完整before R=[E6 3 1FA FFFF0000 2BDA08 2BDA38 0 1]與舊解碼器初態相同，flags293h即ZF0／SF1／OF0，條件false，AL E6→0／EIP17D5A3。下一113628945原88C2令DL FA→0、EDX1FA→100／EIP17D5A5。其他R／六段／全部flags293h、FPU control127F／status0／depth0／八stack bits全0、RAM／VBE保持，兩筆readonly真、callback12／12及IRQ26735／26735非活動、pending0、error nil。一次臂兩步，沒有目的記憶體write特例、代寫資料或EIP。

新CPU入口前8180原343列／30PNG逐位元保持，原343 SETLE兩步／唯一SS實際write與342 TEST三步、341全部正常輸入保持。CPU只擴張裸register SETcc16條件，逆轉小區塊逐byte保持343、Jcc未改；probe三有界observer逆轉後逐byte保持343、Bus／hooks／calendar／120M cap／輸入未改。343其他條件拒絕負例明確改為字面真值正例，所有原SETLE／prefix／memory／CMP／store測試保持。

**已證實，獨立CPU契約**：524,288組完整32旗標×AF兩向×兩context×16opcode×八目的×八unused欄×四代表byte；65,536組全部256初byte×16opcode×八目的×兩flags。以16個字面32bit真值位圖核對，獨立於CPU布林表，驗高低byte鄰居／其餘24位、完整R／段／FPU／flags／RAM保持與零writes。77筆32bit值全部配對×10比較條件×AL／AH，共118,580組數學signed／unsigned CMP→SETcc；512組Jcc直接對字面真值。保留原SETLE的2,097,152組與47,432 signedCMP及store失敗回歸。截短、全memory ModRM、11prefix及相鄰0FA2／0FA3保持拒絕。外部386／8088實機語料未取得，不稱硬體語料或逐週期對拍。

**已證實，實際終態**：同120M上限到17FCE4、unique_sites37433，無新CPU拒絕，終R=[48 8 1 5 2BD9D0 2BD9EC 2BDA74 F]、flags297h。最後32步包含17FCC3..17FD13迴圈及計數指令，未證實生成完成或持久writer。原640×480終圖重新人工檢視仍「Generating Universe...」，PNG d0e387dc900c9955d82dc9c6b21a533d581ac383ba36abca9ce2de28b290f13a、RGB3eeb511abe9d33ff110ce8e478b7d2c5073d36d6775a56622e64651ca8c63fce。舊343相同圖bytes的判定不沿用，不將新游標／圖bytes當新玩家頁。probe exit0不能當完整開局通過，共享20DDDB仍未命中。

實際命令：go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestSETcc|TestSETLE' -count=1 PASS，5.386s；固定DOSGOLEM_MOO2_EXE=/tmp/game/ORION2.EXE的go test -p 2 -buildvcs=false ./... -count=1 PASS，cpu386143.652s／machine5.679s。全套以git ls-files -z | tar --null -T - -cf - | tar xf - -C /tmp/test-src加cp internal/cpu386/setcc_byte_register_test.go及現存testdata建立，避免忽略探索main污染；未重寫歷史來源。

原版本輪只重生一次，go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe，沿341固定calendar／44M Esc／分離DOS／NEW_GAME／MENU_READY／SETUP_ACCEPT／RACE_HUMANS／RULER_NAME_ACCEPT／BANNER_RED旗標及MAX_STEPS120000000。Docker原版300s／全套600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。python3 workplace/new-game-344-input-verify.py／new-game-344-source-verify.py／new-game-344-formal-verify.py／new-game-344-backlink-verify.py PASS；輸入audit於未改CPU階段核對。81項回填、新344的25缺證據負例與其餘三份較早回填6負例、343的26／342的25與340另兩／341的29／340的28／338及339各26負例通過；338 CLI17無效／正對照、339 CLI22無效／120M與100M正對照保持。

CPU SHA-256 b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30，setcc_byte_register_test.go722ea60398d646b4934148ba31067493e0123daca13cc06cfe1d2a3cb93a4cf2，343測試擴張版a282a1b7b4fc188dc450d4cffcffadd9157e0d8667db5fb0c02573e852ca0b60，正式probe5909237597f3dd80599d85597180bd44e628372f92dcd62573f493d0298dfcd4。原素材／圖／LOG／RAM留本機忽略workplace，不入Git。

| 本機忽略來源／收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/new-game-344-input-verify.py | 804bdded426e4d3ff9beab8279158b73adacc0236140ce79c5a1e3f5e17619af |
| workplace/dosgolem/workplace/new-game-344-input-tests.txt | a173f81e49deb943e361509369a6fa35568ddbe776db8e02e18486dd0e146a88 |
| workplace/dosgolem/workplace/moo2-probe-344-red.txt.gz | 2724628fa913674f41df2af004c3ec4863b02cfb105695c6864f1bd35f4610a9 |
| workplace/dosgolem/workplace/moo2-vbe-344-red.png | d0e387dc900c9955d82dc9c6b21a533d581ac383ba36abca9ce2de28b290f13a |
| workplace/dosgolem/workplace/new-game-344-cpu-tests.txt | 477b24311933a5af41b1c84e2cdcec56003384c8738b7380b302860e1401a88b |
| workplace/dosgolem/workplace/full-test-344.txt | b7bcf138095b48d20b4a8ca1f43b8f60531a80ae59c6243463f3de52d6e38554 |
| workplace/dosgolem/workplace/new-game-344-formal-verify.py | ab6233d11b6cccabc11027275f4d9183aba27402205659e7e5068368f22c25fe |
| workplace/dosgolem/workplace/new-game-344-formal-tests.txt | 8de44fc29170e33c7d1bc9508405fcbf26e66b4379e5841ce3028d9f00ce8947 |
| workplace/dosgolem/workplace/new-game-344-source-verify.py | 71f92f1579a000ff8509835534e36dde253188cba1a79e80d6415ca4ea8f3c67 |
| workplace/dosgolem/workplace/new-game-344-source-tests.txt | 92d21a4cfb9882b5c2438a99150fbefc53f87235092ebad131a66e6a60245da0 |
| workplace/dosgolem/workplace/new-game-344-backlink-verify.py | b866c737eabfa1f96f8ef323d4eeb387ce8ae97673013ab5a0a06c4612962d29 |
| workplace/dosgolem/workplace/new-game-344-backlink-tests.txt | 8cd1a7082b4ef0ae71f0338672a08a9b25758462d05f275b45e50d80ee4bd85e |
| workplace/dosgolem/workplace/new-game-344-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-344-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |

**未知與下一步**：在目前120M正常流程的17FCC3..17FD13範圍保存有界唯讀計數／caller／退出條件，先判斷宇宙生成是否正常推進或有阻塞，再決定續跑預算。保持原正常輸入，不盲提高cap、重送或深挖無關helper。持久姓名／旗色writer、typed種族特性、生成完成／完整開局、正式RNG、人耳與remake同狀態未驗。主庫玩法RE閘門保持，整款remake／中文化目標仍活躍。

主庫四文件、14份私有來源／收據、工具精確HEAD與官方輸入雜湊核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；兩工作區掛載篩選Docker容器空，未留背景程序。

### 2026-10-03 宇宙生成迴圈576步與兩個原RET正常返回

主庫起點33fafba87b14e1bcec47a89df79dc33641bdc6a0，工具起點d5127adc64a04af79796d933aec73413bbcfd824；工具a666ae584ba4df9468c233af2a07823229b12e09已推送github隔離分支且遠端回讀一致，未推本機origin。公開[345有界迴圈／兩RET](https://github.com/wicanr2/dosgolem/blob/a666ae584ba4df9468c233af2a07823229b12e09/docs/spec/345-moo2-universe-loop-progress.md)限定CONFORMED；第三例pending，生成完成／完整開局與remake同狀態未驗。340..344與索引／guard同次回填，不計入主庫玩法分母。

官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原位址dosgolem_high_le；CPU完全保持344的b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30。原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patchZIP908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5唯讀；fresh417根檔／MOX.SET553bytes bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／2GiB／2CPU／128pids／UID1000。固定1996日期不是RNG seed。

**已證實，負結果與真因**：345先DRAFT。單檔private build缺IRQ原型，原版未啟動；沿341入口補齊依賴。外層300s／450s各exit124，沒有完整終態，不當CPU／玩法失敗。450s持續raw LOG保存495原步，8444可比原列保持，39次INC／44次CMP／44次JL／51次MOVSX／41次MOVZX核算PASS；第三組只有111步。60partial PNG逐位元保持相應344圖並以manifest列hash。首次EBP+4候選2BDA74非code，真正RET未捕捉，這份負結果與來源不重寫。

12次65,818,624-byte SHA-256實測1.029894684s，2304次約3m17.739779328s，證實逐步完整RAM檢查的額外成本。新v2完整RAM在arm／首末步／RET抽樣，其餘570原步明示ram_checked=false；每筆activationPeek核對R／段／EIP／flags／Bus／FPU bits與VBE。peekSourceWindow只範圍檢查及copy，不寫guest，source逆轉及全舊列／PNG另驗，不冒稱每步全RAM對拍。cold buffers移至state，未臂時不反覆分配。修正候選與成本後，相同300s容器／原120M乾淨重跑成功。

**已證實，原計數與正常返回**：114030168／117001983／119946572原17FCC3各臂192步，共576。原SS188／EBP2BD9EC窗口可讀，AX比較EBP-28界值18／66／87，BX比較DI15、DX比較8；只是原比較值，不當生成百分比。45次INC／51次CMP／51次JL／60次MOVSX／48次MOVZX以16bit signed差／溢位、SF／OF字面真值及byte資料核算PASS，未呼叫CPU helper。用途未知，不深挖完整helper。

17FC82 bytes56 57 C8 14 00 00先PUSH兩個寄存器再ENTER，17FD1B..17FD20為C9 5F 5E C2 14 00，返回在EBP+12。114058778／117106151的17FD1E C21400實際返回17F037，SS:ESP2BD9F8原四bytes37F01700，ESP→2BDA10即pop4+imm20；其他完整R／六段／flags246h保持，readonly／valid真、error nil。兩次自觀測入口至返回28,610／104,168步，證實這兩次呼叫已完成。第三組120M仍waiting=[false false true]，不宣稱第三例／全生成／所有呼叫已完成。

全部8503原344列／32PNG逐位元保持，所有正常press／release與原SETG／MOV、SETLE實際SS write、TEST三步保持。兩份private／正式三有界區塊逆轉後保持344，CPU／平台／Bus／hooks／input／calendar／120M cap保持。原資料充分後READY，正式source逐byte等於已驗v2，不因private→public重新跑相同原流程。終120M保持17FCE4／unique_sites37433／flags297h、無新CPU拒絕；原圖同344「Generating Universe...」，PNGd0e387dc900c9955d82dc9c6b21a533d581ac383ba36abca9ce2de28b290f13a、RGB3eeb511abe9d33ff110ce8e478b7d2c5073d36d6775a56622e64651ca8c63fce。沿同hash人工判讀，不當新玩家頁；共享20DDDB未命中，probe exit0不當完整開局通過。

實際build：cp workplace/moo2-universe-345-v2.go workplace/moo2-probe/probe345_input.go，trap清理；go build -p 2 -buildvcs=false -o /tmp/moo2-probe workplace/moo2-probe/probe345_input.go workplace/moo2-probe/irq1_prototype.go workplace/moo2-probe/irq7_prototype.go；正式go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。沿341固定calendar／44M Esc／分離DOS／NEW_GAME／MENU_READY／SETUP_ACCEPT／RACE_HUMANS／RULER_NAME_ACCEPT／BANNER_RED旗標及MAX_STEPS120000000。原版本輪一個建置拒絕、兩個環境逾時、修訂後一個完整成功收據，原步數都保持120M。

python3 workplace/new-game-345-partial-verify.py／new-game-345-verify.py／new-game-345-source-verify.py／new-game-345-backlink-verify.py PASS。82項回填、345新增27缺證據／抽樣／pending／狀態／索引負例與其餘四份較早回填8負例、所有舊負例通過。正式binary的338 CLI17無效與正對照、339 CLI22無效與100M／120M正對照保持。CPU／平台不改，不重跑無關Go全套；上輪固定EXE全套b7bcf138095b48d20b4a8ca1f43b8f60531a80ae59c6243463f3de52d6e38554只稱既有回歸。外部386／8088實機語料未取得，不稱硬體逐週期驗收。

private v2／正式source SHA-256 03bb2adb3414b8801f35ea25398a1cffc239befd4994f454050d1074ee659c35；舊proto67783bbbccffd894218f5ded77ecf7f1d3ff1607b7260bf1308e2f889ec931df／partial資料不重寫。原素材／LOG／PNG／raw frame／private source留本機忽略workplace，不入Git。

| 本機忽略來源／收據 | SHA-256 |
| --- | --- |
| workplace/dosgolem/workplace/moo2-universe-345.go | 67783bbbccffd894218f5ded77ecf7f1d3ff1607b7260bf1308e2f889ec931df |
| workplace/dosgolem/workplace/moo2-universe-345-v2.go | 03bb2adb3414b8801f35ea25398a1cffc239befd4994f454050d1074ee659c35 |
| workplace/dosgolem/workplace/moo2-probe-345-attempt2.txt.gz | f6535c62ddea3697415ed3f400e6d29a32a392dc514899f3d368a4ec4958be74 |
| workplace/dosgolem/workplace/moo2-probe-345-attempt2-raw.txt | 167e1cacc2710615ca1c2d9099f5565315a41b8c58399309ba2f49e8ebd28f98 |
| workplace/dosgolem/workplace/new-game-345-partial-verify.py | 9c12e2d2683f167edb5e657b9d7821a6a68a9dd2e698d6c937cd9eb3bf9c2b65 |
| workplace/dosgolem/workplace/new-game-345-partial-tests.txt | a626ff35a5249b429135bb76a3d3e57e10826327530a51481efc7211624a7441 |
| workplace/dosgolem/workplace/new-game-345-sha-cost.txt | 2cf6031e6533b102f049772b28efbecccb3cd45ba889a4f7382abba7c4d29178 |
| workplace/dosgolem/workplace/moo2-probe-345-red.txt.gz | c5ffbfda48e4ff7354452ace99bded92c4ccb31f4522e4cdc8534a34d84a4e44 |
| workplace/dosgolem/workplace/moo2-vbe-345-red.png | d0e387dc900c9955d82dc9c6b21a533d581ac383ba36abca9ce2de28b290f13a |
| workplace/dosgolem/workplace/new-game-345-verify.py | c6be67c4a1109f512c070b6166b6733b787944d3e03be6b2c4f930c2228cd180 |
| workplace/dosgolem/workplace/new-game-345-tests.txt | 14c2d70635217f3266ad9b811d624e18c805abf7f02f92726635378c173ff096 |
| workplace/dosgolem/workplace/new-game-345-source-verify.py | e0087de1adc8490943013c87b1cbca49f193b683e6ed14853909a8023960ee50 |
| workplace/dosgolem/workplace/new-game-345-source-tests.txt | d01abf123dd62e684115e636babd2b513ba71b341891a38059ef7e66363e2419 |
| workplace/dosgolem/workplace/new-game-345-backlink-verify.py | 35a5c69cffab41f10921538d1c46e1e4b54219b4b75474bdd04524c0d639bdd9 |
| workplace/dosgolem/workplace/new-game-345-backlink-tests.txt | 489053fd9832911efae521e38f859c892c1e70d14ae5ce5736854f57a93cce2a |
| workplace/dosgolem/workplace/new-game-345-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| workplace/dosgolem/workplace/new-game-345-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |
| workplace/dosgolem/workplace/new-game-345-partial-frames.json | f9f98a7be5291b00d87552a7391808cce66baa3f3e4b3f4db738b7c1408454b7 |

**未知與下一步**：保留120M基準，另建立固定160M的明示診斷分支，先指定120M同狀態收據／160M有界終態，再DRAFT→READY後同正常input續行。160M是一次性預算，不保證生成完成；若仍同頁先找生成producer／狀態變化，不連續盲加cap或重送。正式writer、typed種族特性、生成完成／完整開局、正式RNG、人耳與remake同狀態未驗，主庫RE-first與整款remake／中文化目標仍活躍。

主庫四文件、18份私有來源／收據、工具精確HEAD與官方輸入雜湊核對通過，其餘活表全文保持。既有root-owned 2437檔／272目錄保持，本輪未新增root-owned或.md目錄；兩工作區掛載篩選Docker容器空，成功與逾時工作均已清理。

## 2026-10-03：346固定160M診斷同狀態、第三RET與配置母星原圖

起點主庫8f5dbc7b3eb2b3e93f8bd7262e2ff476ea2bd660／工具a666ae584ba4df9468c233af2a07823229b12e09。規格沿工具[346](../../../workplace/dosgolem/docs/spec/346-moo2-universe-160m-normal-continuation.md)，現工具92d25f387d2ff18919ec65da2d85c7ce1566d7db已推送github隔離分支；所有原位址dosgolem_high_le，並非DOSBox位址或EXE檔案偏移。主庫RE-first保持。

### 證據等級與界線

- **已證實**：固定160M只由新明示旗標且完整正常輸入依賴鏈開放，原100M／120M契約保持。四346區塊／五guard逆轉逐byte等於345，CPU／平台未改。
- **已證實**：預設120M全部9085原345列／32PNG保持；160M到120M前全部9028可比原列／31原frame保持。依既有mtime／DTA／每次RAM SHA規則正規化，原terminal不改名當新收據。新120M觀察在下一CPU.Step之前，完整R／六段／flags／FPU／VBE／clock／callback／IRQ與原同點保持。
- **已證實**：第三group於120083995，17FD1E C2 14 00→17F037，stack37F01700、ESP2BD9F8→2BDA10，其他R／六段／flags246h保持。from119946572共137423原步。原120M第三pending仍正確，本輪延後取得RET，三waiting全false；不推論所有生成helper已完成。
- **已證實**：150M仍「Generating Universe...」，160M原640×480圖為「Placing home worlds...」，人工檢視130M／140M／150M／160M。僅知文字變化在150M至160M之間，不猜producer原位址或生成規則。
- **未知**：正式姓名／旗色writer、文字producer與caller、typed種族特性、完整生成／開局、存檔、正式RNG、人耳與remake同狀態。固定日期不是seed，probe exit0只是固定上限。

120M R=[48 8 1 5 2BD9D0 2BD9EC 2BDA74 F]／seg=[8 188 188 0 20 188]／flags297h／EIP17FCE4；FPU control127F／status0／depth0／stack0。VBE Bank4／StartY0／BankSets1685／Writes36595476／DisplaySets50，virtual_micros227500108；callback12／12，IRQ28643／28643，BIOS deliveries28663另列；RAM readonly、原第三pending保持。其餘完整platform欄位逐欄比對，無跨次全RAM逐byte聲明。

160M終態R=[6D 0 5 4 2BCEE8 2BCF04 2BDB44 F]／六段相同／flags207h／EIP17FD04／unique_sites39434，原bytes66 0F B6 4D 1C 66 39 CA 7C D3 40 66 3B 45 E4 7C。FPU control127F／status0／depth0／stack0，virtual_micros349146004；VBE Bank7／StartY512／BankSets2328／Writes46391556／DisplaySets77。130M／140M／150M／160M四checkpoint RAM before／after相同，35frame加final共36真實PNG。final等於160M原frame，RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9、PNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca。160M共10523原列、無新CPU拒絕；完整開局未驗。

### 可重生工具與實際驗證

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；原版每側timeout450s／network none／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀。ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5、EXE4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、MOX.SET553bytes bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。fresh417根檔各重生一次，guest唯讀檔案provider。

CPU SHA b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30；正式probe SHA b389f6c6b661e534b92e0be060c31921c2251735594187c7ed842a2f1f12f7e2。容器內go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe；兩次實際命令完整保存為本機workplace/new-game-346-run-120.sh與new-game-346-run-160.sh，檔案provider、原99M press／99084355 release保持。第二側只額外UNIVERSE_CONTINUE_160M=1／MAX_STEPS=160000000。沒有注入guest選擇、重送或略過生成。

python3 workplace/new-game-346-ready-verify.py／new-game-346-source-verify.py／new-game-346-cli-verify.py／new-game-346-verify.py／new-game-346-backlink-verify.py PASS；公開python3 apps/moo2/tools/startup_probe_131.py --check-universe-continuation-spec-backlinks PASS。新CLI29負例在讀EXE前exit2，舊338的17／339的22負例及100M／120M正對照保持。83項規格回填、新346的23負例與所有舊負例通過；CPU未改，本輪不重跑無關Go全套。344固定EXE全套只當既有回歸。

初次run腳本CAP代換誤改HARDWARE_ESCAPE_STEP，兩次在讀EXE前拒絕、原版未執行。原rejected raw／gzip保存；修正命令後相同容器／公開source乾淨重跑完成。這是執行腳本問題，沒有新產品缺陷。345過期下一步回填346，同時保存完整歷史證據索引。

### 本機忽略來源與收據

以下均位於workplace/dosgolem/workplace/，不入Git；68張原圖manifest只記本機雜湊，不散布PNG。

| 收據／核算 | SHA-256 |
| --- | --- |
| new-game-346-ready-verify.py | 3f57e33ac48b2f7c2d50e9f84ca72142617fd642929c6daccb9f232fa53c08a5 |
| new-game-346-ready-tests.txt | 2dc21fe750aa318c262d3d82286b37f22adf7a3f69b3f6cab7412a3f56e7db48 |
| new-game-346-source-verify.py | 0b36c270f693cda3e8568a7f545ebec8850ffd98abe6a4dfa780707c069522b7 |
| new-game-346-source-tests.txt | 3fae2a65f8aa3f5d4578fa6ae19f2c9bf1fb359ba1673638501968f85cbe673f |
| new-game-346-cli-verify.py | fa18280fad37abc1c8bab98e8d53e1d7c824bbbad640009f78eda64b77c0c79c |
| new-game-346-cli-tests.txt | 8b1443a99376737876f15ec2b99afee310602a31733b8377000feeac1e44d8f8 |
| new-game-346-old338-cli-tests.txt | c5d03a8a5858cdeb915036ef21e2a8f0f3627b6007bfbe3e54c3745fb624b6b1 |
| new-game-346-old339-cli-tests.txt | e4461c40e971daf4de8338567b57b68f0945355ed595b7c5c09620a2590be774 |
| new-game-346-run-120.sh | 67af7bb351387ccc2f58f41f2b14a58db005eafe6a407a0d2e22793d27163e66 |
| new-game-346-run-160.sh | adeb31446d7297b6c6feeded05b615d0e3dc5bed2f2f7c0a9eb5402af09eba1c |
| moo2-probe-346-120.raw.txt | 00c7be8ca72851f793879eac0414edb795f98b001aa769d34ee8c0a7c6ad4eb1 |
| moo2-probe-346-160.raw.txt | 149c5c3f99ef56be28d84dc49e940d36b41c96d4919fc42cd97ea47b9ca657a5 |
| moo2-probe-346-120.txt.gz | f8a87f0a377ccc16d53526263bbe2751cf90accfa43da9b4e9db0302385209dd |
| moo2-probe-346-160.txt.gz | 09038decb752495dfc75519e89dc8f0cedec997a42fe02c467085fd130d869af |
| moo2-vbe-346-120.png | d0e387dc900c9955d82dc9c6b21a533d581ac383ba36abca9ce2de28b290f13a |
| moo2-vbe-346-160.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-346-verify.py | 76cc426525a1a6ce04fb2c6e7ed46c50127ebb7d5285295790eba88c2436c698 |
| new-game-346-tests.txt | 51a032f667f7fd516274638474f53b01deb5a79d07f37f9f58ac02b8c192eb0b |
| new-game-346-backlink-verify.py | 1f773cef7c8db3496e749658244ce4eae5a34b21566a275609b9fc5f4f56c7e8 |
| new-game-346-backlink-tests.txt | 0b7a0ab2dfee09a5829c44e10d215d65e47fe009a910de4b53e23d3b03a2dd4a |
| moo2-probe-346-cli-rejected-120.raw.txt | 531247bf00fd36b424407cf703a6460b06e5fc7deec8d03f7e096e95cee5f8aa |
| moo2-probe-346-cli-rejected-160.raw.txt | 108e19d356c0b577c4ebb15ab6ca32aa18917196bd37f0eec2db7b234294e1b3 |
| moo2-probe-346-cli-rejected-120.txt.gz | 13783d6f70719c0ad0ec999a76cfcd695f0e9768a6cdd10994b55bb1c8643fdd |
| moo2-probe-346-cli-rejected-160.txt.gz | 56d6959c3ae0202b375c3f764047f1bd436a568ebed6c669f11118a4870ef58f |
| new-game-346-frames.json | 008b17b9ed7b033fba7a33c77e47964527f9a9ff4bcd7d6eb52d4e035da47359 |

兩個原版容器已結束並移除，工具掛載相關容器空；新來源與收據1000:1000，工具樹無root-owned／.md目錄。主庫本輪只更新四文件，其餘活表保持，歷史root-owned不修動。下一步維持160M預算，用150M至160M文字變化定位最小producer／caller與狀態，先DRAFT觀測，不直接提高cap或重送。整款remake／中文化目標仍活躍。

## 2026-10-03：347原進度文字、正常查詢與NUL複製

起點主庫6e32fabf24833d6fa2811160d0bfcaed46420e56／工具92d25f387d2ff18919ec65da2d85c7ce1566d7db。現工具53243f6d5633380f456a9c27faedf908cbade675已推送github隔離分支；詳見工具[347](https://github.com/wicanr2/dosgolem/blob/53243f6d5633380f456a9c27faedf908cbade675/docs/spec/347-moo2-home-worlds-text-source.md)。主庫RE-first保持。

### 證據等級

- **已證實，原資料**：HESTRNGS.LBX SHA-256 a3193f56d12ae512ab8a78cc50aba9df44d6d614a4cc23b14ec7b9f1b09cbf30，archive count1／entry2048..15573／raw1×13521；配置母星檔案offset9646／payload7594／NUL ordinal161，生成宇宙offset11729／payload9677／ordinal242。MSGENG.LBX ee70bd446054139101b5187861af24590bc2b46a30736a52b9a09eb27b6d3096，entry2048..410244／raw384×1063；兩文字在record161或242＋63。原ZIP與官方1.31 patch對應檔逐byte相同。這些是檔案offset，不與CPU或IDA混列。
- **已證實，原查詢與複製**：索引242在102875194的dosgolem_high_le:17DCA5 E8 E6 CC FE FF呼叫16A990，102875207返回27D9E9，102875330於17DCBA完成NUL複製。索引161在152598605的16C8A3 E8 E8 E0 FF FF呼叫16A990，152598618返回27D1C6，152598741於16C8B6完成。兩次真正CALL返回各13步；來源48-byte窗口連後一段字串與HESTRNGS相同，目的2842F4由原CPU寫入。兩來源指標等於27B41C＋原payload offset。
- **已證實，完成邊界**：原POP EDI之前才完成，兩ESI各source＋24，EDI28430C、AL0／flags246h／ZF1。原CMP與JNZ已通過、NUL已讀寫，不用既有零值冒稱複製完成。六事件核心／Bus／FPU／VBE及完整RAM readonly；只讀窗口，不注入guest狀態，copy等待各256原步。
- **已證實，實際呼叫端**：IDA linear EA的sub_7C78E prolog53 51 52 56 57 C8 10 00 00，對應已核對的dosgolem_high_le:16C78E，五PUSH再ENTER。原SS188:EBP2BDB5C＋24四bytes8AB91600保存16B98A，對應IDA caller7B985的E8 04 0E 00 00。本次實際caller16B985，另一靜態caller16AE07未當實際命中；尚未捕捉16C78E最終RET。
- **強推論，實際來源檔**：IDA sub_7A816的HESTRNGS.LBX引用／34D1h長度，加上原RAM相鄰文字布局，支持英文blob由該檔載入；本輪未另攔截file read。只有實際比對的來源bytes與正常查詢稱已證實。
- **未知**：16C78E最終RET、其後word[EBP-8]欄位用途、文字renderer、母星配置規則、持久姓名／旗色writer、完整生成／開局、存檔、RNG、人耳與remake同狀態。

IDA linear EA sub_7A990..7A9B0有12指令，讀word_1912FC並加byte_18B41C；本輪核對的dosgolem_high_le定位為16A990..16A9B0、2812FC與27B41C。IDA原名稱／原operand／file offset／bytes保持；每列工具／基準分開記載，本輪F0000映射只套用已核對的入口、RET、PUSH、CALL及資料指標，不外推其他版本或位址。lookup_call的source_offset0是尚未取得來源指標的佔位，並非空字串查詢結果；負值／397以上分支只有靜態證據，未列為本輪動態驗收。

### 保持與實際命令

全部10523原346列／36PNG保持，沿mtime／DTA／每次RAM SHA正規化，不把原terminal換名當收據。160M終態仍17FD04／unique_sites39434／無新CPU拒絕，原PNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca、RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9，沿346已人工判讀的配置母星原圖，不當新玩家頁或完整開局通過。

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；IDA9.4 image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。Go原版每次timeout450s／network none／2GiB／2CPU／128pids／UID1000；IDA一次性DB timeout90s或120s／同資源，原ZIP／patch／正式.i64唯讀，輸出只寫workplace或tmp。原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5、EXE4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、417根檔／MOX.SET553bytes bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。固定日期不是seed。

DRAFT階段private v2六事件與完整160M原流程核對通過，READY後才把兩有界區塊接到正式診斷。正式只比private v2多defer的universe160關閉守衛；旗標1觀察語句相同，兩區塊逆轉逐byte保持346。正式source ca5643c8d598cb705914cbbcfdc79ef571c6d603960b9395f552c7defe52c546，private v2 644805a449403684a68da0b2a0ee85c987d2e8a7e5e0bd4099189c0bce3059ce；CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30未改。

private建置沿IRQ1／IRQ7三檔入口，原版完整命令保存workplace/new-game-347-run.sh；正式go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe與旗標關閉8M／CLI命令保存new-game-347-off-run.sh。正式8M關閉基準1693原列／PNG保持，沒有新增觀察，終態235826／unique_sites10813；它只驗啟動與關閉守衛，不取代正常160M。python3 workplace/new-game-347-verify.py／new-game-347-source-verify.py／new-game-347-off-verify.py／new-game-347-backlink-verify.py PASS；公開python3 apps/moo2/tools/startup_probe_131.py --check-progress-text-spec-backlinks PASS。84項回填、新347的28負例與345另兩負例、所有舊負例通過。三舊CLI共68負例與100M／120M／160M正對照通過。未重跑120M或無關Go全套，歷史346同狀態與344全套不當本輪新跑。

### 失敗分類及勘誤

初次可丟棄LE匯出器猜RAM上限400000，超過初始2874576-byte容量而panic，未執行CPU。主庫既有IDA DB輸入7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5不符固定1.31，拒作oracle；重新建一次性DB，以非空JSON／schema／5365函式／固定input SHA／UID1000驗證，不以idat exit1或stdout空判失敗。初次IDAPython API放錯模組，保留attempt1並改為ida_loader後重跑。

private v1目的buffer的既有NUL讓生成文字copy_complete在102875322提早八步，獨立ESI／EDI核算拒絕；原log／腳本保留。v2在原POP EDI之前、指標跨NUL／AL0／flags246h才完成，102875330及152598741通過。同160M固定預算乾淨重跑，沒有把觀察錯誤寫成CPU或玩法缺陷。

### 本機忽略來源及收據

以下均位於workplace/dosgolem/workplace/，原EXE／LBX／LOG／PNG／私有腳本不入Git。36PNG manifest只記本機雜湊。

| 收據／核算 | SHA-256 |
| --- | --- |
| new-game-347-text-search.py | 7a83ff0e1ae7d04567e08920eb6d1835794e873d9ea44d0886b8dc135bbf5e25 |
| new-game-347-text-search.json | 5286b807104ebc75644110c27bdb62dbb265c77b3420cd6a90c245fcf43d6657 |
| new-game-347-lbx-shape.py | c4bb727ac68e940ab6a4a6fc5e633e0634a346f6157e26f4ecf212821ee718b6 |
| new-game-347-lbx-shape.json | ebc5a68fc1d68da73d5179311037d5f3265a3996b1655a08082757ceda52cbbb |
| moo2-347-ida-existing-db.json | 3521023087e238592469a6dfb2d876e3652f6f9154e02efa9b6bb7aab68ae997 |
| moo2-347-ida-min.py | b2b9ec6cdbcbc8178d107073b41a0c822cedc76e471894ab85b3bbb0810468a1 |
| moo2-347-ida-min.json | 48e6ca489165f4ebac6731e4a9dd9b4260759e78e4e1883edd48e8eeda2125cc |
| moo2-347-ida-literals.py | b548cd1bda9d744f8d64080994a845cc34cdf7f0c1429e4b55435e149b660047 |
| moo2-347-ida-literals.json | 78b4e716b1426e89d9105574397cd1e6a4f555772a1ce55bea3bcb1d9d963ea9 |
| moo2-347-ida-literals-attempt1.py | 57de283eeea29f6bb292a6b0e41d8234f5455bf23a682ebd762d1e7e557964fd |
| moo2-347-ida-literals-attempt1.log | 5227c159fd7b0eea37983aafe45f972616ce9c70669923b0a993f9b7d93b511c |
| moo2-347-ida-text-consumer.py | 4648e3fc6b470a8d3c184f3f9e6098eb7a5e6d239dd01b7b85d0285d6ffe98d2 |
| moo2-347-ida-text-consumer.json | 5970160c8aa070bcb282ad8934a3b8213ff31baff74d94749d77df8e54a64b18 |
| moo2-347-ida-init-caller.py | 7fd53c6ba4f0252f68805b692242c50ca0c269781cefc441eb92cd2aadce95c2 |
| moo2-347-ida-init-caller.json | 84465efbc4891edadf700252a35891e53054372d30eb2487fb004e7de28db92e |
| moo2-347-load.go | 5e2a78987c12072862a70f70353c44beab31edae64a92f883863c63c4a8652c6 |
| moo2-347-text-attempt1.go | cced941c35c9011bd6f45190459d77433eba16e2be7cf0a973546befdf48d943 |
| moo2-probe-347-attempt1.txt.gz | ce652c3e6f0fd8040fd53c246c4c17f95460ffdf78942cbf2fc6a5050521a3c5 |
| new-game-347-attempt1-tests.txt | 1447c07cc2cdd16c13c8281a8173fec14b4108cd7f2f1d8a681b8e5ead06be99 |
| moo2-347-text.go | 644805a449403684a68da0b2a0ee85c987d2e8a7e5e0bd4099189c0bce3059ce |
| moo2-probe-347-text.txt.gz | 3e90425b497fcb452e72a5066a88fcc402042e7fe163342fe6c697a08954d87e |
| moo2-vbe-347-text.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-347-verify.py | 5a466ebc9fa1311a12134811a994e5d3b699a6642d4048d07f2c97bd4a28c899 |
| new-game-347-tests.txt | 1799ba5f1447ed213c5b2003c75ee7bf6fe708e9c0447b6e1fff72926bb64f4c |
| new-game-347-source-verify.py | 260eea59425e2899a9b84e95139f8b3ba4174136a6eba044caa5192764a5181b |
| new-game-347-source-tests.txt | ba08223bfa0707adb11b84c60ef6df7555cb85efd218177e7c49c653f79e817a |
| new-game-347-off-verify.py | b9dff364d976d09a4f0ecf846d4b2ed09449caafbc16d2d5c7c716d65a9c9824 |
| new-game-347-off-cli-tests.txt | 12f8c09ab7c76599b2fc8ab0fdedadb9bec8c10f2b9ea2b1ed80bb5c67248505 |
| moo2-probe-347-off-old.txt | dccb3bbc64a3c1fd1b7d003274f4951676ac7c6ed0517903aedf574049930e39 |
| moo2-probe-347-off-new.txt | f651d1c0d023eed50f75fa62b191bbf7a93db4d2b6ae4ea8f8511b11becae460 |
| moo2-vbe-347-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-347-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-347-run.sh | 4755cd82aaf0ca4c5d51c7af229fd267c0fac2bcbcb966ab645671a33fe0742f |
| new-game-347-off-run.sh | c951a79e48e1645ed63ec27c9a6bf033f970df9b25dcddd2646c1229f08ecfdb |
| new-game-347-backlink-verify.py | c6f952b1823132ab87b5c353d3a863f73920edbb7166bee70ae5faf7f3e84eb8 |
| new-game-347-backlink-tests.txt | 517b4d757e0589df38f71a3eda0e74d9b34ff02b199cfd514afa3f200fea5e31 |
| new-game-347-frames.json | c5a76825087036f3581e0ebf03dd99147e834a659c323ae02502756f64d64ea3 |

新來源／收據1000:1000，工具樹無root-owned／.md目錄，原版與IDA容器均結束移除；主庫只更新四文件、其餘活表保持，歷史root-owned不修動。下一步維持160M，觀察16C78E正常返回16B98A及其後word[EBP-8]分支，原欄位用途未知，不猜補規則或提高cap。整款remake／中文化目標仍活躍。

## 2026-10-03：348原進度函式RET與上層零分支

起點主庫c6b990d5246414cf895422b8df14cd30cc03a5db／工具53243f6d5633380f456a9c27faedf908cbade675。現工具9afe6570dc3e49e354b9a9ec07360125682b3d67已推送github隔離分支；完整限定契約見[348](https://github.com/wicanr2/dosgolem/blob/9afe6570dc3e49e354b9a9ec07360125682b3d67/docs/spec/348-moo2-home-worlds-return.md)。主庫RE-first保持。

### 證據等級與保持

- **已證實，原正常返回**：149825343 dosgolem_high_le:16B985 E8 04 0E 00 00 CALL，149825344到16C78E，原SS:ESP存8AB91600。153214287於16C8E1跳共享退出、153214288執行16BF57 LEAVE、153214290開始16BD81五POP，153214295在16BD86 C3真正RET，153214296回16B98A。EDI／ESI／EDX／ECX／EBX恢復，ESP2BDB74→2BDB78、EBP恢復2BDBA0。
- **已證實，原零分支**：caller的SS188:2BDB98 word0000；153214296原CMP 66 83 7D F8 00、次步16B98F的0F 85 77 01 00 00 JNZ因flags246h／ZF1而不跳，153214298到16B995。callee EBP2BDB5C-8的raw7779與caller EBP2BDBA0-8不同，欄位用途未知，不猜語意。
- **已證實，IDA定位**：固定1.31原sub_7C78E邊界IDA linear EA7C78E..7C8E6／103指令；7C8E1 E9 71 F6 FF FF→7BF57 C9，再7BF58 E9 24 FE FF FF→7BD81的5F／5E／5A／59／5B／7BD86 C3，RET file offset1053658。原caller7B985／7B98A／7B98F／7B995 file offset1052633／1052638／1052643／1052649。原名、EA、offset、bytes與operand保留，兩工具基準分開標明，不外推其他版本。
- **強推論，較晚呼叫**：160M的原返回槽2BDB74為E3BA1600；固定IDA caller窗口7BADE E8 30 F2 FF FF呼叫sub_7AD13、返回7BAE3。支持後續dosgolem_high_le:16BADE→16AD13尚在執行，本輪未捕捉其CALL，不把保存位址當實際命中或返回通過。
- **未知**：caller raw欄位用途、後續CALL實際命中與輸入、文字renderer／母星配置規則、正式writer、生成完成／完整開局、存檔、RNG、人耳及remake同狀態。

全部10530原347列／36PNG保持，沿既有mtime／DTA／每次RAM規則正規化。11事件核心／Bus／FPU／VBE、完整RAM before／after一致，原CPU／平台／輸入與160M預算不改。終態160000000／17FD04／unique_sites39434／無新CPU拒絕。原PNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca，RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9，仍配置母星圖；只核算原圖相同，不冒稱完整開局。

### 命令、工具及驗證

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；IDA9.4 locked-v1 image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。沿347已驗UID1000組合，兩次IDA非空JSON／schema1／固定input SHA／擁有權通過，idat exit1不是失敗。正式DB唯讀，原EXE只複製到tmp建立一次性DB。

原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5、官方EXE4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，417根檔／MOX.SET553bytes bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。原ZIP／patch／正式.i64唯讀，容器network none／2GiB／2CPU／128pids／--user1000。原版timeout450s、正式8M及CLI timeout180s、IDA timeout90s。固定1996-01-01不是seed。

容器實際入口為bash workplace/new-game-348-run.sh及new-game-348-off-run.sh；private三檔建置沿IRQ1／IRQ7原入口，正式go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。python3 workplace/new-game-348-verify.py／new-game-348-source-verify.py／new-game-348-backlink-verify.py全部PASS；公開python3 apps/moo2/tools/startup_probe_131.py --check-home-return-spec-backlinks PASS。85回填、新348的27負例及既有負例通過。

READY審查後才接正式兩區塊，正式只更名private的348標記，執行語句相同；逆轉兩區塊及新增空行逐byte保持347。private9a5b920cb3c167eff8d2e472ce780fab37ccd56bc9b88fd9595078e23c8be8a7、正式2aa4d46019d0c532fef169ba686f48aaadc347ccdb9b4b49928b8054a2dc8571、CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30未改。正式關閉8M的1693原列／PNG保持，沒有新增觀察；68舊CLI負例及100M／120M／160M參數正對照保持。8M不取代160M正常流程；未重跑120M或無關CPU全套，既有回歸只稱歷史已驗。347已回填，不讓保存返回位址的舊邊界取代本輪真正RET。

初次來源逆轉漏移除新增空行，diff只有兩空行，修正核算後通過。Docker metadata Go User欄位不存在，另核對image ID通過，容器仍明定--user；均屬驗證腳本問題，不列產品缺陷。

### 本機忽略來源與收據

均位於workplace/dosgolem/workplace/。原EXE／LBX／RAM／LOG／PNG／私有IDA腳本不入Git；36PNG manifest只記本機雜湊。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-348-ida-home-return.py | 1fdd50368529695e682b09cad75f02f999a4c66908697dc884ddb36ddfbcf73b |
| moo2-348-ida-home-return.json | 2038716b21fe3332fb8044cf01d914637eceb7cbebe55c0650b4762e2784b0aa |
| moo2-348-ida-home-epilog.py | f944158607c92a47af670c271bdc6ddc7711f164f0880d594036f67c7f72a21b |
| moo2-348-ida-home-epilog.json | a0c6edcee841854579115800ff5ef0edeec7c0df07ad5d9320c351682fe0993d |
| moo2-348-home-return.go | 9a5b920cb3c167eff8d2e472ce780fab37ccd56bc9b88fd9595078e23c8be8a7 |
| moo2-probe-348-home-return.txt.gz | 33623a56218c0c0684b9513704266928c47ed49488ae880b5a55eb9181c73efa |
| moo2-vbe-348-home-return.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-348-run.sh | 82af3816ef9756888432bd3457fb9d8d93892f28964799890e642d129f0ab2f6 |
| new-game-348-verify.py | 185f38302d53008c56e18bab8cf8a12cf2bc94167de1a67605ca7678d71a5386 |
| new-game-348-tests.txt | f66285c1243f3493b6bd37ed875a82c9f057405fdb09366ff72b2e1a5af39c22 |
| new-game-348-source-verify.py | 382e38868bc40b06b7a556fdc1e126715da14de308795c0dc12b88901bf76c28 |
| new-game-348-source-tests.txt | c029e9776e92d9798835896002d816c2248fbb0c4f745e608980d1b29f01bc94 |
| new-game-348-off-run.sh | 48569ef0b1b10f22da4a688daf431647567900710ce996e4c51e847d919b78f5 |
| new-game-348-off-cli-tests.txt | 7d720ece028de31c0f318e14ac3aa0eaf3a62300093ab5410538b8b4007df7f1 |
| moo2-probe-348-off-old.txt | 46bf56e6965b90579fb24fba65ae70f5e5707110de40149bd0ebd43d75ce121c |
| moo2-probe-348-off-new.txt | 8ba812f8773be1e48cff2b5146929359e7bbabea8cf8acfbe511c672b3bf4ac5 |
| moo2-vbe-348-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-348-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-348-backlink-verify.py | fcc64554ff1312fadd0d28e67866fdac0784379ea0b36d32be950cc1db3fd2bd |
| new-game-348-backlink-tests.txt | f9d2f9a204f756743bc680bd15400a4d7e685bad3b553f39e9288d8b60215d08 |
| new-game-348-frames.json | 79a5115a26fafec11dd48bcf5c89b62ed8a86e6f0596cde819545dda104b2bcd |

原版、IDA與驗證容器皆結束移除，新來源／收據1000:1000，工具樹無root-owned／.md目錄；主庫只更新四文件，其餘活表保持，歷史root-owned不修動。下一步維持160M，觀察16BADE→16AD13正常入口參數、框架與生成呼叫邊界，不猜欄位用途或直接提高cap。整款remake／中文化目標仍活躍。

## 2026-10-03：349後續生成原入口與外層等待

起點主庫ff2c377b2e3a08faba7f3d600a0e547759c4bcdc／工具9afe6570dc3e49e354b9a9ec07360125682b3d67。現工具07611f5808e53eb40a61cc8a94489045b0ff0686已推送github隔離分支；限定契約見[349](https://github.com/wicanr2/dosgolem/blob/07611f5808e53eb40a61cc8a94489045b0ff0686/docs/spec/349-moo2-generation-entry.md)。主庫RE-first保持。

### 證據等級與原流程

- **已證實，原CALL**：153878499 dosgolem_high_le:16BADE E8 30 F2 FF FF呼叫16AD13，153878500到原入口。EAX2BDB78=caller EBP2BDBA0-28h，EDX147D9=83929；原SS188:ESP2BDB74存E3BA1600。四PUSH／ENTER 0BF8h後153878505的EBP2BDB60／ESP2BCF68，原參數保持，欄位用途未知。
- **已證實，三直接返回**：原16AD34→17E5C5於153878512呼叫、153878623回16AD39；16AD56→17E5C5於153878632呼叫、153878743回16AD5B，各111步、ESP2BCF60保持。16ADFB→17EFE1於153881393呼叫、153979600回16AE00，98207步，ESP2BCF48→2BCF60。固定IDA linear EA的8F052 C2 18 00／file offset1132198清理24byte，不能預設每次CALL返回ESP不變。
- **已證實，首次距離返回**：153994840的16AE7F E8 30 02 00 00→16B0B4，原AX1／DX3，153994869到16AE84回EAX1CC4D=117837，29原步。原IDA sub_7B0B4有28指令，以71h stride讀兩records的signed word +0Fh／+11h，差值平方再相加。只閉合距離平方形式及一次原返回，records名稱、座標單位、完整規則及remake對齊未驗。
- **已證實，原定位**：固定IDA sub_7AD13邊界7AD13..7B020／216指令，唯一direct caller7BADE／file offset1052978 E8 30 F2 FF FF；入口offset1049447。sub_7B0B4範圍7B0B4..7B0FB／28指令，RET7B0FA C3／offset1050446。原名、EA、file offset、operand、bytes保持，與dosgolem原CALL／框架逐項核對，工具與基準分開標明，不外推其他版本。
- **已證實，限定等待**：13事件，0..11及16已見、12..15未見，returned=false。160M原EIP17FD04／unique_sites39434／無新CPU拒絕，原外層slot16BAE3仍在，未捕捉16B01F RET或16BAE3返回。只稱本觀察範圍pending，不冒稱配置完成或CPU故障。
- **未知**：外層迭代實際上限與剩餘工作、16AE07重繪、參數／frame欄位用途、文字renderer／母星配置規則、持久姓名／旗色writer、完整生成／開局、存檔、RNG、人耳及remake同狀態。

全部10542原348列／36PNG及160M終態保持，沿既有mtime／DTA／每次RAM規則比較，原home_return_observation的RAM before／after各次相等，不要求跨次RAM SHA相同。新13事件核心／Bus／FPU／VBE與完整RAM readonly。原PNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca／RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9，仍配置母星圖。原99M press／99084355 release、160M cap、原日期保持，無代寫／重送；固定日期不是seed，exit0只表示cap。

### 工具、命令與驗證

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，IDA9.4 locked-v1 image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。沿已驗UID1000組合，三次原EXE一次性DB的非空JSON／schema1／固定input SHA／UID1000通過；idat exit1不當失敗，正式.i64唯讀。

原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5、官方ORION2.EXE4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，417根檔／MOX.SET553bytes bfd6855a41760b31156b96114b5b33c88f442ab8f8aae020c1740b3b486a3a80。原ZIP／patch／正式DB唯讀，network none／--user1000／2GiB／2CPU／128pids；原版timeout450s，正式8M／CLI180s，IDA90s，輸出只寫workplace／tmp。

容器內：bash workplace/new-game-349-run.sh／new-game-349-off-run.sh；private沿IRQ1／IRQ7三檔入口，正式go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。python3 workplace/new-game-349-verify.py／new-game-349-source-verify.py／new-game-349-backlink-verify.py PASS；公開python3 apps/moo2/tools/startup_probe_131.py --check-generation-entry-spec-backlinks PASS。86項回填、新349的32缺證據／狀態／較早回填／索引負例與既有負例通過。

READY證據審查後正式只更名349標記，執行語句與已驗private相同；逆轉兩區塊逐byte保持348。private fe1b7a036bcedfd0bc26705cb86b1fcd26973ae28fbac7058f5a1d64eac8f9dc／正式f2826d243f01ff4b6659aa45d4faae635d218dc779e8357b486f04b5870471be／CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30保持。正式關閉8M1693原列／PNG保持，無新觀察；68舊CLI負例及100M／120M／160M參數正對照保持。8M不取代160M，未重跑120M或無關CPU全套，既有回歸只稱歷史已驗。348已回填，尾端仍把已命中CALL列未知的殘留已刪除；歷史保存slot的證據保持。

初版核算猜28指令為29而拒絕，原腳本／stderr保留並重現exit1；另外按原C21800修正第三CALL的24byte清理。CPU／private執行語句未改，不列產品缺陷。原CALL也計一原步，29原步返回不證明callee有29指令。

### 本機忽略來源與收據

均位於workplace/dosgolem/workplace/，原EXE／LBX／RAM／LOG／PNG／私有IDA腳本不入Git。36PNG manifest只記本機雜湊。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-349-ida-generation-entry.py | c4f6fd6abb0fd757447cc4d33450057b3901ffc3c53305a169477e59e7b93a42 |
| moo2-349-ida-generation-entry.json | 46f87208bd5896c95af4598c1e1cfbea6fe23114841eb9852462b0c29b153428 |
| moo2-349-ida-generation-nested.py | 603f5a87dbeee3203ae3c0e866f82f1998fee198fb31e0b1b95d532f2b3d3f22 |
| moo2-349-ida-generation-nested.json | 07944763f6098d52665442faae7729a3b1cf1811cb4a98980c79100bdcfc8a6a |
| moo2-349-ida-ui-ret.py | f8b9754475957a5c9ca09ce98161546efc72ee4b48a38205e87a78f4c35943e7 |
| moo2-349-ida-ui-ret.json | 0c16013bacada3769016549c3a30cfe12ddc2ea72b3829d20ff77256969b0dec |
| moo2-349-generation-entry.go | fe1b7a036bcedfd0bc26705cb86b1fcd26973ae28fbac7058f5a1d64eac8f9dc |
| moo2-probe-349-generation-entry.txt.gz | c47496ca8f677245371ca239a5a945baf7f56747e0cbd7e2471d81c55e24126d |
| moo2-vbe-349-generation-entry.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-349-run.sh | 573f9a810a93136e7a5a78dfdb8d29242c3c481c98ac7d46007db3e5b981a97a |
| new-game-349-verify.py | 16d3e5bc7a4d83a015953b819ca4b5727b017887e30b0350e52c0952f932341c |
| new-game-349-tests.txt | a7e3f2e45420a231c00270496b8b4b3f6dbc19c455edde78c5281d360184ceca |
| new-game-349-attempt1-verify.py | 92335a08ba904116c43ace623338adae1decb5872986db2ee62a4e631feab304 |
| new-game-349-attempt1-tests.txt | 5915c4337c93e40c5141b6768754cc4313c1fb1fd50f533cb193e5e2a51bb5d3 |
| new-game-349-attempt1-verify-output.txt | 4c5c68a11c199e993db2d08da691f6ef2ebe4225d2f771ed694ca70509e91139 |
| new-game-349-source-verify.py | 1c291c9852cee4cd85d767f47dfc13fd33a4166b1ebb76c1a1e945bc9719faf2 |
| new-game-349-source-tests.txt | 8cc211bb15558b6950cfc0a183b951fdd8238a2860ef46fb09846281d43e4399 |
| new-game-349-off-run.sh | fb90870320e62dcc800601f43bd201be8c7259ba187d01e42ac5d2847544a43c |
| new-game-349-off-cli-tests.txt | 7d720ece028de31c0f318e14ac3aa0eaf3a62300093ab5410538b8b4007df7f1 |
| moo2-probe-349-off-old.txt | faeb64440db8b4a958ea16263211a19f5a0723f5a64dbaef12f52886901812f9 |
| moo2-probe-349-off-new.txt | 0e6bfc4639a8292e1b66db5fdd325f8ea73165ff50306c380907946bbb301d68 |
| moo2-vbe-349-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-349-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-349-backlink-verify.py | 22d2e06a78f4530bc63ad6d01537ba1c2c1d228dadc8e44e692b6e466f8cfc1c |
| new-game-349-backlink-tests.txt | 7091c1cf8e1e8435873cfa8ec2baa81b78ad531d9432fd5bbfad253badee99e9 |
| new-game-349-frames.json | 661ebbf70b113592d12355fa40b2632acbfdc68feea7116b2ce315b96584d098 |

原版、IDA與驗證容器皆結束移除，輸出1000:1000，工具樹無root-owned／.md目錄；主庫只更新四文件，其餘活表保持，歷史root-owned不修動。下一步維持160M核對dosgolem_high_le:16AD7D／16AF1B..16AF23外層迭代、IDA word_19199A實際上限與16AE07重繪，再依進度決定續行預算，不猜欄位用途或盲提高cap。整款remake／中文化目標仍活躍。

## 2026-10-03：350 原生成外層首四迭代與實際界限

路由命中dosgolem／規格閘門／回填；沿原EXE4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、417檔、MOX.SET與既有固定IDA349匯出，不重建正式DB。起點主庫10f8dccc8d33cd311bc45b5d4c21cd24f06a4ffa／工具07611f5808e53eb40a61cc8a94489045b0ff0686。工具結果[025629e](https://github.com/wicanr2/dosgolem/commit/025629e9fe5de8e29abe81aa08c67ad37a8ff161)，完整HEAD 025629e9fe5de8e29abe81aa08c67ad37a8ff161；[350規格](https://github.com/wicanr2/dosgolem/blob/025629e9fe5de8e29abe81aa08c67ad37a8ff161/docs/spec/350-moo2-generation-iteration-bound.md)限定CONFORMED，不代表整款remake完成。主庫玩法RE-first保持。

| 原head SI | head原步／dosgolem_high_le:16AD7D | CMP原步／16AF1C | CMP SI | JL原步／16AF23 | head至CMP原步 |
| --- | --- | --- | --- | --- | --- |
| 1 | 153880145 | 154031469 | 2 | 154031470 | 151324 |
| 2 | 154031471 | 154183024 | 3 | 154183025 | 151553 |
| 3 | 154183026 | 154349384 | 4 | 154349385 | 166358 |
| 4 | 154349386 | 154516973 | 5 | 154516974 | 167587 |

**已證實**：實際dosgolem_high_le:16AF1C原66 3B 35 9A 19 28 00讀取DS188:28199A raw2400／signed36。既有IDA9.4 linear EA:7AF1C原66 3B 35 9A 19 19 00／word_19199A另記；EA7AF1B 46 INC ESI／EA7AF23 0F8C54FEFFFF JL7AD7D逐項核對，位址基準不混用。原欄位語意未知，不以名稱猜星系或玩家數。

**已證實**：各CMP後一原步到JL，下一SI2..5與signed36使SF xor OF=1，前3組JL後下一步回head；frame EBP2BDB60／ESP2BCF60保持。13事件核心／Bus／FPU／VBE及完整RAM before／after相等。首四組16AE07／真正16AE0C返回未見，第5個SI後full=true，最多21事件；不能把有界未見外推到全程。

**已證實**：160000000／17FD04／unique_sites39434／outer_returned=false，無新CPU拒絕。terminal ESI2BDB44／EBP2BCF04屬nested框架，不當外層SI。finalPNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca／RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9仍「Placing home worlds...」。

**強推論**：原樣本有迭代進度，但四次耗時不保證全部生成或剩餘時間。**未知**：全部迭代出口、後段／重入／RET、後續重繪、正式writer、原frame語意、完整配置／開局、RNG與remake同狀態。固定1996-01-01不是seed，不宣稱受控亂數parity。

全部10556原349列／36PNG按既有mtime／DTA／每次RAM規則保持。DRAFT私有收據及獨立核算通過後先READY審查，再只更名350標記接正式；source逆轉逐byte保持349／CPU與平台不改。正式關閉8M1693原列／PNG、68舊CLI負例與100M／120M／160M正對照、87項回填／新350的28負例與既有負例通過。本輪沒有CPU修正、IDA重跑、120M重跑或CPU全套。

### 原入口、命令與私有收據

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，Docker --rm／network none／UID1000／2GiB／2CPU／128pids，原版timeout450s、正式8M／CLI180s；原ZIP／patch只讀。CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30，正式probe0cbe29a3e93145f2c7ee71036e0fe76902b6dbc76809571df761a3269d7e70d3。固定IDA349匯出SHA另列，input EXE／IDA版本及位址基準保持。

容器內實際入口與結果：

```text
bash workplace/new-game-350-run.sh
  原160M exit0，13新事件／groups4／indices1..4／full=true／outer_returned=false
python3 workplace/new-game-350-verify.py
  全部10556原349列／36PNG／原signed界限與JL下一步／來源逆轉 PASS
python3 workplace/new-game-350-source-verify.py
  正式與已驗private只更名350標記／兩區塊逆轉逐byte保持349／CPU平台保持 PASS
bash workplace/new-game-350-off-run.sh
  原關閉8M1693列／PNG保持，68 CLI負例及正對照 PASS
python3 workplace/new-game-350-backlink-verify.py
  87項回填／新350的28缺證據與限定範圍等負例／既有負例 PASS
python3 apps/moo2/tools/startup_probe_131.py --check-generation-iteration-spec-backlinks
  原首四迭代／signed界限36／飽和範圍／較早回填 PASS
```

原99M按下／99084355放開、calendar與160M cap保持；沒有guest代寫／重送／換Bus／CPUhook。本輪private及正式都只蒐首四組與終態，原平台／玩法不改。正式關閉測試包含 go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。

以下檔案在工具忽略workplace，原EXE／LBX／RAM／LOG／PNG／私有腳本不入Git；只公開自製診斷／規格／索引／守衛與雜湊。既有IDA檔是重用來源，不宣稱本輪新執行。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-349-ida-ui-ret.json | 0c16013bacada3769016549c3a30cfe12ddc2ea72b3829d20ff77256969b0dec |
| moo2-350-iteration.go | b7c6d33772360ad748ce19d6dc20758da3ea30dadbf9e248af0f1d701955efa7 |
| moo2-probe-350-iteration.txt.gz | d2c9130ac36766528d7a955a2711e45a863cbee5ca631bea36ad1f4085135231 |
| moo2-vbe-350-iteration.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-350-run.sh | 62cf6bd9e6e606f087dfd06ec38c64cd898aafc21a7296b20ceabb30b12cb82b |
| new-game-350-verify.py | 9a006bb260d23ce0f4e7415c24cc55d586181c87bbe5d480b728dcf08b38b265 |
| new-game-350-tests.txt | f24b7714b79603c1868f293204fe82d6a9ca42cd7b01ac7878f8407fe172f45a |
| new-game-350-source-verify.py | 1fc55322c19b2ba4bb64e84a15fc27a8786c9c9f0d2dc0fe5d64d4ff6263152c |
| new-game-350-source-tests.txt | 75ad232d938073707ae1fd1c698391bfb883538453b39813e24010cfdba0f649 |
| new-game-350-off-run.sh | db2087e5f4655be47413d6fac125b4298f9bfc03740dc0736493afed0283bb92 |
| new-game-350-off-cli-tests.txt | 7d720ece028de31c0f318e14ac3aa0eaf3a62300093ab5410538b8b4007df7f1 |
| moo2-probe-350-off-old.txt | f710a4306e74e7424e99444b9a469d4f0ccd80b51cd746dbbae4794e88cca567 |
| moo2-probe-350-off-new.txt | 288af6c93c3cd4ed571eda12995791e16cd313570ec7f7150983aec0af5b7b7e |
| moo2-vbe-350-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-350-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-350-backlink-verify.py | 7a35c9000023e5f5a9d094b29682518a9d695f4d1c6a4d763e4f36b9f820333e |
| new-game-350-backlink-tests.txt | 38f6a1121a9a05b982ddf8055ef80a0ba810a5a76da19ea0f999e75b3b0c822a |
| new-game-350-frames.json | 53151cd1754678ba3c5a702351d0c629782aa222d9cfb01b8dc3131ff6de6bc3 |

主庫只更新目前狀態表與追加歷程。下一步維持160M，取原16AF29出口／16AF35後段及16B01F RET／16BAE3返回邊界，再依實際工作量調整預算；不深挖runtime或圖形helper。兩庫收尾核對精確HEAD／遠端／乾淨工作樹與Docker容器清理；本輪輸出UID1000、工具root-owned零、主庫既有2437檔／272目錄保持，沒有.md目錄。

## 2026-10-03：351 原連續SI迭代與160M截斷證據

路由命中dosgolem對拍／規格閘門／回填，沿已載入入口及固定IDA349原匯出。起點主庫ac2784ad7fcac666da2b6f2aa0779d88163d6b79／工具025629e9fe5de8e29abe81aa08c67ad37a8ff161。工具結果[ffc7e16](https://github.com/wicanr2/dosgolem/commit/ffc7e16a6ed34a279418982dbe5131e328ec7162)，完整HEAD ffc7e16a6ed34a279418982dbe5131e328ec7162；[351規格](https://github.com/wicanr2/dosgolem/blob/ffc7e16a6ed34a279418982dbe5131e328ec7162/docs/spec/351-moo2-generation-completion-boundary.md)限定CONFORMED。主庫RE-first保持，整款remake／中文化尚未完成。

**已證實**：原dosgolem_high_le:16AD7D的26個head，SI嚴格為1..26，無回跳／重複；153880145到159926951共6046806原步。25個head間隔151326..338347，整體耗時增長但非每次嚴格增加。160M只讓第26次再續73049步，原SI36比較、16AF29出口、16AF35..16AF6E後段、16B014／16B018結果入口、16B01A epilog、16B01F RET與16BAE3 caller返回皆未見。

**已證實**：每個原head EBP2BDB60／ESP2BCF60，bound依原relocated bytes66 3B 35 9A 19 28 00讀DS188:28199A raw2400／signed36。27事件readonly、heads26／events27／full=false，最大72heads＋11邊界＋終態1筆；沒有觀察飽和。終態160000000／17FD04／unique_sites39434／outer_returned=false／無新CPU拒絕。terminal ESI2BDB44屬nested框架，不作外層SI。

固定IDA9.4 linear EA sub_7AD13的原7AF29 B9FFFFFFFF、7AF35 E884390800、7AF68 0FBFC1、7AF6E 0F8EA4000000、7B014 30C0、7B018 B001、7B01A C9、7B01F C3／file offset／operand與原名保持；本輪只重用349匯出，不執行新IDA或改正式DB。實際dosgolem_high_le定位另記，不因靜態有RET即宣稱動態完成。

**強推論**：連續26次工作量支持下一次固定180M有限續行，加20M供原SI26..35及後段探索。20M約為10個最大已量間隔338347的5.9倍，屬選定探索餘量，不是剩餘時間上限。後段／重入／完整開局仍未知，不盲提高cap或跳過生成。

**未知**：SI27以後、全部迭代出口／後段選取／重入／RET／caller16BAE3、後續重繪、正式writer、原欄位語意、完整配置／開局、RNG與remake同狀態。固定1996-01-01不是seed，不宣稱受控亂數parity。

全部10570原350列／36PNG依既有mtime／DTA／每次RAM規則保持。DRAFT private實測／獨立核算後先READY審查，正式只更名351標記；來源逆轉逐byte保持350，CPU／平台／原輸入／160M cap不改。正式關閉8M1693原列／PNG、68舊CLI負例與100M／120M／160M正對照、88項回填／新351的25負例及既有負例通過。沒有CPU修正、IDA重跑、120M重跑或CPU全套。

### 命令、環境與私有收據

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，Docker --rm／network none／UID1000／2GiB／2CPU／128pids及450s原版外層timeout；正式8M／CLI180s，原ZIP／patch只讀。原417檔／MOX.SET／99M按下與99084355放開／calendar與160M cap保持；無代寫／hook／換Bus或重送。

官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30；正式probe6cc3c65f7648e9582713a69fd51fa13a2d2a824cdec79204a7c88e90b22a183c。finalPNG c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca／RGB4a1d9efc9d15da575e330128f22d27e97f6d6616ac1b4184ec94efc2bd27f2c9仍「Placing home worlds...」。

```text
bash workplace/new-game-351-run.sh
  原160M exit0；heads26／events27／full=false／outer_returned=false
python3 workplace/new-game-351-verify.py
  全部10570原350列／36PNG／SI1..26連續／160M截斷／來源逆轉 PASS
python3 workplace/new-game-351-source-verify.py
  正式只更名351標記／兩區塊逆轉逐byte保持350／CPU平台保持 PASS
bash workplace/new-game-351-off-run.sh
  原關閉8M1693列／PNG與68舊CLI負例及正對照 PASS
python3 workplace/new-game-351-backlink-verify.py
  88項回填／新351的25缺證據與限定範圍等負例／既有負例 PASS
python3 apps/moo2/tools/startup_probe_131.py --check-generation-completion-spec-backlinks
  原連續進度／160M截斷／未見出口／較早回填 PASS
```

正式關閉測試建置使用 go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。以下收據在工具忽略workplace；原EXE／LBX／RAM／LOG／PNG／私有腳本不入Git，公開只保存自製診斷／規格／索引／守衛與雜湊。IDA檔是重用來源，不宣稱本輪新執行。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-349-ida-ui-ret.json | 0c16013bacada3769016549c3a30cfe12ddc2ea72b3829d20ff77256969b0dec |
| moo2-351-completion.go | 185d611e3cccf08c5b4a0d2c423766e7dcb5bf0e1119bcb02475a9edda3b5465 |
| moo2-probe-351-completion.txt.gz | 389a328d590e406bf2a09134cbc864476bcee5ee31e09288e82a1c51fb65db95 |
| moo2-vbe-351-completion.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| new-game-351-run.sh | 25acc19b3a0b9cf04d5eb82b1c9df952dba988c433ac463cddcb965f3165beae |
| new-game-351-verify.py | 6910cd202514cf6b5a5ac333ae3507feebd86cbdf85d012dfbf18d26825cbe92 |
| new-game-351-tests.txt | 83c52d22e97fb63c4a13bebdec6d2158af7fc5a012c8cc8894ba6c0f3fcfd2ae |
| new-game-351-source-verify.py | 90b8631a7ec629c9b305f4dfd19da6d9852fa0ea3ae6c7bd21d31e7c0f52f2b4 |
| new-game-351-source-tests.txt | c5399648a97935aa06578278eedee83a817f58f9e725a8a501688b94eef9e202 |
| new-game-351-off-run.sh | 923783a3c35a606ecc795e92dc3d3d4188d90f7f314273f4d8a2fd80c9eced0e |
| new-game-351-off-cli-tests.txt | 7d720ece028de31c0f318e14ac3aa0eaf3a62300093ab5410538b8b4007df7f1 |
| moo2-probe-351-off-old.txt | b26779d13b2b12e3bc94db2813b50f0b1920e1dc2b4d4ade45e62e480e6b44c7 |
| moo2-probe-351-off-new.txt | bd693fc70584776eadc2835d35d072a58b625781a3ae52def9e81c51a510ed69 |
| moo2-vbe-351-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-351-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-351-backlink-verify.py | 48c5f0b652b55799313da611b25391d0c239df15f1aa17b33447aab8f08fc2de |
| new-game-351-backlink-tests.txt | afd425de1526b803e70c395ff059efbf2e9eee1b63420ba946d29b44e8a1e694 |
| new-game-351-frames.json | 77c5cc266bb737831784fc814d9d3fbd356db8a74e192d53c201ab59fd27d1b7 |

主庫只更新目前狀態表與追加歷程。下一步建立固定180M正常續行，核對160M前同一原狀態／進度／畫面，再追出口、後段、RET與正常玩家畫面。兩庫收尾核對精確HEAD／遠端／乾淨工作樹及專案Docker清理；本輪UID1000，工具root-owned零，主庫既有2437檔／272目錄保持，沒有新增root-owned或.md目錄。

## 2026-10-03：352 明示180M正常續行、原RET與新CPU拒絕

路由命中dosgolem對拍／規格閘門／回填，沿已載入入口及固定IDA349匯出。起點主庫6379a13d3121fbaaca0eb82137c08a4bc3849ecb／工具ffc7e16a6ed34a279418982dbe5131e328ec7162。工具結果[8ef52b1](https://github.com/wicanr2/dosgolem/commit/8ef52b1fc372bf96267d964f8b1ba903594d95c6)，完整HEAD 8ef52b1fc372bf96267d964f8b1ba903594d95c6；[352規格](https://github.com/wicanr2/dosgolem/blob/8ef52b1fc372bf96267d964f8b1ba903594d95c6/docs/spec/352-moo2-generation-180m-continuation.md)限定CONFORMED，較早349／350／351已回填。主庫RE-first保持，整款remake／中文化尚未完成。

### 原版已證實與未知邊界

以下動態位址均為dosgolem_high_le，不與IDA linear EA混用。原16AD7D head的SI1..35連續，無重複／回跳。SI27在160267006，SI35在163345051；163755070原16AF1C CMP比較SI36與DS188:28199A raw2400／signed36，163755071原16AF23 JL不跳，163755072到16AF29出口。原16AF35後段在163755075進入，163778142到16AF68／DX48h，163778144到16AF6E／EAX22h，JLE不跳。

163778780到原16B014／EAX5，原30C0清AL，EB02跳過16B018 B001。163778782到16B01A／EAX0，163778787原16B01F C3 RET的ESP2BDB74／EBP2BDBA0／topstack E3BA1600；163778788回16BAE3／ESP2BDB78／EAX0。原163779084到16BAEC，次步JZ到16BB00。heads35／events46／full=false／outer_returned=true；AL1入口未見。AL0返回不等於完整生成、成功開局或配置規則已對拍。

新停點：163795435原103BF9 bytes66 81 63 0C 7F FE C1 E2 07 09 53 0C EB 2B A1 18，錯誤「81 word形狀尚未支援」。前六bytes解碼為AND word [EBX+0Ch], FE7Fh；EBX5AA044／DS188，目標是segment offset DS188:5AA050。拒絕後EIP103BFC僅為decode已取ModRM的位置，不表示後續指令執行。原word輸入／寫回尚未捕捉；資料語意、後續消費端、正式writer、完整生成／開局、RNG與remake同狀態仍未知。

要求預算180000000，CPU在163795435拒絕，final readonly觀察記163795436；沒有step_limit或dos_exit。probe exit0是main處理step_error後返回，不稱預算完成。finalPNG由dosgolem自行產生，SHA-256 d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6；沒有可靠的人眼檢視或完整玩家路徑驗收。

### 契約與驗證

新增DOSGOLEM_MOO2_UNIVERSE_CONTINUE_180M=1，須依賴舊160M旗標與完整正常banner路徑、MAX_STEPS=180000000。新旗標只接受1；舊160M契約保持。記錄實際budget，180M不冒稱160M。原正常99M按下／99084355放開、1996-01-01日期及硬體逃逸44M不改；日期不是seed。沒有guest代寫、跳呼叫、重送輸入或換CPU／Bus。

160M前10532共通列保持，只正規化兩個實際預算宣告、既有mtime／DTA／每次RAM／path。舊35 frames加160M final共36PNG保持；新160M checkpoint的CPU／segment／EIP／flags／完整FPUstack／code16／stack96／frame128／input64／slot4／VBE與indexed／RGB／callback／IRQ逐欄相同。每個觀察的activationPeek與完整RAM前後相同；不宣稱跨執行完整RAM SHA相同，也不說180M保留舊351所有10598列的stop footer。

DRAFT private原實測後READY審查，正式只更名352標記，執行語句與已驗private相同；new-game-352-patches.json逆轉全部替換逐byte保持351。CPU／平台保持，關閉旗標8M原1693列／PNG保持，68舊CLI及32新CLI拒絕案例／正對照通過。89項規格回填、新352的30負例與349／350另4負例及既有負例通過。未新執行IDA、未重跑120M或無關CPU全套。

初次核算在IRQ key迴圈使用a／b，覆寫共通列list變數；附加10532列斷言因而拒絕。保存初次腳本／tests／stderr，確定重現exit1；只改IRQ迴圈變數名，保留原嚴格斷言、private來源、raw、CPU與平台，乾淨重跑通過。這是驗證腳本錯誤，不是遊戲缺陷。

### 命令、環境與私有收據

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。Docker --rm／network none／UID1000／2GiB／2CPU／128pids，原版外層timeout600s，正式8M／CLI180s；原ZIP及patch只讀。固定DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，417根層檔與MOX.SET輸入保持。

正式probe SHA-256 c49edc0afcb44dfec22139043afb887a72a70b83d172a165d6869ac9a54be4a1；CPU b5c8bd900047e8f018ad66edaf5ae8e1009e9f6d04ec89dea4000c27f3326b30。固定IDA349原匯出僅重用，不宣稱本輪IDA新執行。

```text
bash workplace/new-game-352-run.sh
  probe exit0；原163795435 CPU拒絕，未達180M
python3 workplace/new-game-352-verify.py
  160M前10532共通列／36PNG／原160M同一狀態／原RET／新81拒絕 PASS
python3 workplace/new-game-352-source-verify.py
  正式只更名352標記／逆轉逐byte保持351／CPU平台保持 PASS
bash workplace/new-game-352-off-run.sh
  關閉8M1693列／PNG、68舊CLI＋32新CLI負例／正對照 PASS
python3 workplace/new-game-352-backlink-verify.py
  89項回填／新352的30負例／較早另4負例及既有負例 PASS
```

正式關閉測試建置使用go build -p 2 -buildvcs=false -o /tmp/moo2-probe ./workplace/moo2-probe。以下來源／收據均在工具忽略workplace，原EXE／LBX／RAM／LOG／PNG／私有腳本不入Git。公開只保存自製診斷／規格／索引／守衛與雜湊；IDA檔是重用來源。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-349-ida-ui-ret.json | 0c16013bacada3769016549c3a30cfe12ddc2ea72b3829d20ff77256969b0dec |
| moo2-352-continuation.go | 7e03730549d3b9d74f160becde2780e31251b9b1d1437ad3c5181240b191152b |
| moo2-probe-352-continuation.txt.gz | db7630f02daaac46f0a2d45da1813a319d27e14e48c77c7b07a75a97b338c60f |
| moo2-vbe-352-continuation.png | d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6 |
| new-game-352-run.sh | 94f2aa11cd4897af8ed9337adb0d2bcb48a35f9ef845e57eb1b6085b7f2a01b4 |
| new-game-352-verify.py | e31817cc2b71b7df2fb1944dce6166318aa1237ece1c76d8e275417135ff0dbb |
| new-game-352-tests.txt | 798aa04bf026fa1b99af5815065506f1189a7b2e56d04886414453e7d5b65754 |
| new-game-352-source-verify.py | f53b017083b2749e04ccba39da8e2307f785ca5e2af566748dc806f330c48d9c |
| new-game-352-source-tests.txt | f9ab870f8ce6d2eee6f91be4fecb86fc51eab23a23a9890826b2e2cce48a7cff |
| new-game-352-off-run.sh | e6dd0e59c2d3cdf15024697454fe138df6c42f661c1f46f3d81062e6c8740246 |
| new-game-352-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-352-off-old.txt | d5d6210c5e646dd9b39e5547c986cd552099d9ababead60d858d9a3cfe7a82cf |
| moo2-probe-352-off-new.txt | cf37120a71d5f990a3018b630c4b3a6454e397d2fa235b24679ad811defa4d36 |
| moo2-vbe-352-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-352-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-352-backlink-verify.py | 29323d3879ef89ef69b0f13f8426fa2c75b1c72aab1af638c5547cbf006c2863 |
| new-game-352-backlink-tests.txt | 72a022f287c95347684deb10b1c9dd627ebd184962880627cbd186ae77c6f66f |
| new-game-352-frames.json | e20e80211f7a9a28d974451ac23ce6f43cd4ae93e7e69d5c56ed8b1e33f13ff9 |
| new-game-352-patches.json | cdd4bd25a4836607eb32381a2fb9bdf7607910e8059cccb20c7cdfce72e2a33f |
| new-game-352-cli-verify.py | aa13e74dfee1e0b0119650bf5bcb5af87eeedba0ae236e4a7d1a64c4c02973ef |
| new-game-352-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-352-attempt1-verify.py | c88d35f658a09cf5663c16e0744e82e78eb70e22e09a93f908e57f5774ad7163 |
| new-game-352-attempt1-tests.txt | 38ff0861e4b9e06c158d148eab8af7eab4804d8ee9f5ff4c84167051611abcf9 |
| new-game-352-attempt1-verify-output.txt | 9f7b7da0a8992dd918da5c9800a341794e870d3d3bdbc28ff5129b5539963772 |

輸出1000:1000，工具root-owned／.md目錄零；主庫既有2437檔／272目錄保持，本輪不新增或遞迴修權限。原版與正式回歸容器均有界並結束移除；兩庫遠端與精確HEAD在收尾核對。

下一步建立66 81 /4 word記憶體AND窄CPU規格，驗解碼／16bit寫回／flags／相鄰bytes與拒絕邊界，再以相同180M正常輸入捕捉原word及後續消費端。全生成、完整開局、正式writer、RNG、人耳與remake同狀態未驗；不繼續提高預算或深挖繪圖／runtime helper。

## 2026-10-03：353 word記憶體AND、原三步與新memory byte XCHG停止

路由命中dosgolem對拍／CPU規格閘門／回填／文件職責，沿已載入逆向技能與入口。起點主庫ed55a4b87cbd31ab8d2bf55cb5dbcc013b20ed47／工具8ef52b1fc372bf96267d964f8b1ba903594d95c6。工具結果[c708629](https://github.com/wicanr2/dosgolem/commit/c7086292bf476be63a406131632c9cf8c4780ab6)，完整HEAD c7086292bf476be63a406131632c9cf8c4780ab6；[353規格](https://github.com/wicanr2/dosgolem/blob/c7086292bf476be63a406131632c9cf8c4780ab6/docs/spec/353-cpu386-and-word-memory-imm16.md)限定CONFORMED，352未知已回填。主庫RE-first保持，整款remake／中文化尚未完成。

### 契約、已證實與未知

[Intel 80386原廠AND](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/AND.htm)列81 /4 iw的word完整立即值與寫回；[第3.4.1節](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/s03_04.htm)明示OF／CF清零、SF／ZF／PF更新、AF未定義。不深挖driver或逐週期硬體。不以相同日期冒稱亂數seed。

DRAFT未改CPU，沿352相同180M完整正常輸入捕捉dosgolem_high_le:103BF9首遇：163795435完整R=[0 0 0 5AA044 2BDB08 2BDB2C 0 0]、段=[8 188 188 0 20 188]、flags246h。DS188:5AA04F八bytes全零，目的word DS188:5AA050為0000；SS188:2BDB08四bytes01000000。callback12／12、IRQ41715／41715非活動、pending0；readonly與RAM保持。全部10793原352列／36PNG保持，三區塊逆轉source逐byte保持352。原完整word來源可讀與ISA充分，才READY實作。

**已證實，原三步**：163795435的66 81 63 0C 7F FE執行word0000 AND FE7F→0000，EIP103BFF；163795436原C1E207的EDX0左移7仍0，EIP103C02；163795437原09530C的DS同offset dword0 OR EDX0→0，EIP103C05。三步flags246h、完整R／段／來源相鄰bytes／stack與完整RAM保持，readonly=true／error nil／ram_changes=[]。原零→零沒有Bus寫次數trace，通用成功兩byte寫入由CPU受控Bus測試另驗；AF與多位SHL的OF只驗工具模型，不宣稱硬體定值。

**已證實，保持與CPU回歸**：正式入口前10723原共通正常列／36PNG與同一原word／R／段／flags／stack保持；不把舊stop診斷算作指令前原事件。CPU僅新增21行66 81 /4 memory word分支、成功writeSegment16後才發布邏輯旗標；逆轉逐byte保持352。unknown selector／唯讀／段外／讀失敗保持旗標與RAM；第二Bus byte寫失敗保留已寫低byte且旗標保持，只沿工具模型，不宣稱硬體exception原子重啟。原81 word register／memory CMP、83 AND register保持。observer三區塊逆轉保持352，平台／8088 CPU不改。

獨立逐bit交集／五旗標oracle：全部65536 word來源配原mask、65536低byte配對、16位單bit／補數／高位／零與64個初旗標組合；全部ModRM／SIB相異DS／SS、ESP忽略index／無base DS、非對齊／負disp8／32位繞回／最後完整word、相鄰byte；prefix／截短／逐byte讀寫拒絕與成功零／非零實際兩寫全部通過。窄測0.312s，固定DOSGOLEM_MOO2_EXE的乾淨Go全套CPU38677.939s／machine1.449s通過。8088語料缺檔skip不算實機驗收，沒有386實機語料。

**已證實，新停止**：164321317原dosgolem_high_le input223E93 bytes86 06 AA 46 4A 75 F7 07 C3 56 57 06 0F A0 0F A8，錯誤「XCHG byte僅支援暫存器」。ModRM06為DS:[ESI]與AL，原交換尚未執行；after223E95只解碼。原R=[FF 0 2 2BD976 2BD730 2BD888 2BD8A8 2BD97A]、段=[8 188 188 0 20 188]、flags202h。目的DS188:2BD8A8 byte及後續ES188:EDI2BD97A的STOSB初態未捕捉，資料語意未知。requested budget180000000、actual stop164321317，尚未達180M；probe exit0只代表錯誤收尾。

原finalPNG仍d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6／RGB5a416d1db0fa55dc3523c99212ceb813056dda787a301b5e8d26bf799d058b59；本輪沒有新正常開局或人眼畫面驗收。完整母星配置／開局、正式writer、RNG、人耳與remake同狀態仍未知。

初版formal核算切在舊guest_cpu_stop，納入四項stop診斷：late_startup_platform／irq7_passdown_state／protected_dma_pcm／irq7_real_entry；10727對10723的長度斷言拒絕。初次腳本／tests／stderr保留並重現exit1；按實際late_startup_platform label=stop與四項精確類型修正後，原10723列及三步嚴格核算通過。不修改CPU、raw或原收據，不為選結果重跑原版；分類為驗證腳本邊界問題。

### 命令、環境與私有收據

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。Docker --rm／network none／UID1000／2GiB／2CPU／128pids，原版兩次各600s、全套600s、8M／CLI與窄測180s；原ZIP及patch唯讀。原ZIP根層417檔／MOX.SET／99M按下與99084355放開／1996日期／44M硬體逃逸保持，沒有資料代寫／跳呼叫／換Bus或重送。

固定DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。CPU1bfd9e0c7a63453549a7ab1e081d7d669ea8f38c0388e94743f9e0b0049793b1；自製測試02d55bbfac7846cc578101052f5186b62720b6fd0dd3faf95c2e387a56553ef5；probe848dba4c436357de62902e5f4185177bf2b5912decc832d0cd22ba1f34d8624f。沒有本輪新IDA或120M整流程重跑。

```text
bash workplace/new-game-353-input-run.sh
python3 workplace/new-game-353-input-verify.py
  未改CPU10793原列／36PNG與原word初態 PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestANDWordMemoryImmediate|TestANDWordRegisterSignedImmediate|TestCMPWord|TestSUBWord|TestADDWord' -count=1 -v
  獨立CPU窄測與舊word契約 PASS
bash workplace/new-game-353-full-run.sh
  固定原EXE乾淨Go全套 PASS
bash workplace/new-game-353-formal-run.sh
python3 workplace/new-game-353-formal-verify.py
  原10723共通前綴／36PNG／三步／新XCHG停止 PASS
python3 workplace/new-game-353-source-verify.py
  僅21行CPU與三observer區塊／逆轉逐byte保持352 PASS
bash workplace/new-game-353-off-run.sh
  關閉8M1693列／PNG、68舊CLI＋32新CLI負例及正對照 PASS
python3 workplace/new-game-353-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-and-word-memory-spec-backlinks
  90項規格回填／新353的33負例與較早負例 PASS
```

乾淨全套：git ls-files -z | tar --null -T - -cf - | tar xf - -C /tmp/test-src，再複製本輪新自製CPU測試；避免歷史探索main污染，未移動或改寫舊探索檔。關閉8M以352原probe與353新probe同新CPU比較，兩者在本輪AND入口前，source逆轉與正式正常前綴另證CPU舊行為保持。

以下30份本機來源／收據在工具忽略workplace；原EXE／LBX／RAM／LOG／PNG／私有腳本不入Git。公開只提交通用CPU、測試、診斷、規格、索引、守衛與雜湊。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-and-word-353.go | 848dba4c436357de62902e5f4185177bf2b5912decc832d0cd22ba1f34d8624f |
| moo2-probe-353-input.txt.gz | bf7cf698bf7c6cb1bc8e8e4cdf60f2f748801332b6375fdd83f0535204625c20 |
| moo2-vbe-353-input.png | d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6 |
| new-game-353-input-run.sh | d948eee16e1bd8995ae2ff705efa813a5ec3be89e19e6c9e5f74515fa95af50d |
| new-game-353-input-run-output.txt | d361590f9c5d03f14079df3afb60b85517179650af054dc4c717f91d54667963 |
| new-game-353-input-verify.py | 5c35378e2be730e0af658635c890d7dcfb4b561f2f41fad95ffeae5273d8c8b7 |
| new-game-353-input-tests.txt | e936db7ab38cc5afb5708afd958d578e0062337af64252a65f747291d8c8c92e |
| moo2-probe-353-formal.txt.gz | 5e9ca79374c2a67298872b3b2d04d210d9241035d2644899182ebff3b28433a9 |
| moo2-vbe-353-formal.png | d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6 |
| new-game-353-formal-run.sh | a98d928282e41760ac89009df9513396149cff59decf74fb2aec7d3fa94325e1 |
| new-game-353-formal-run-output.txt | 1acc8ab3b8f6ac60b66cf8193af680905f20d45b9fe497df0e9c32b69b391f99 |
| new-game-353-formal-verify.py | b1754dec76a12174c7bc0e81cca5845b0e3822bd8a1d16cbc833a6879367a667 |
| new-game-353-formal-tests.txt | 9d7617a2e3deb70e4bd7a284bbb840eda16bc3b8876f75edc09bcfa3c618a42c |
| moo2-353-cpu-narrow-tests.txt | 62d999ac687acb10705a7dfdfcd1e3ab2a0d97d9e1b548392d01d7b23d872ec7 |
| new-game-353-full-run.sh | 9acf4d81d65ecbafc793d6d57dc142e3d3b8d54e9e53ed549e27500bdfbec11d |
| full-test-353.txt | 16f9e86bd4b984eef315f5e5fb4497cf7bc77dd55b4a719061ad76ff035128fb |
| new-game-353-source-verify.py | 3f20a61abf8b010d0090e1daa423dc5cabc4e7b71f87cb4174a7638d8d6dbbb2 |
| new-game-353-source-tests.txt | b185ef0fa306a8519c42b6a59f072f5a45f14f80f869a25d24cc3838aeb0b5af |
| new-game-353-off-run.sh | a535eea06ec74fa44e764bd80d7ffab4a591ca9fdbc2b7fa7bddf48144578cd8 |
| new-game-353-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-353-off-old.txt | ca3296c1e76ebca60c1bdd7edec596b0aa8ed73611874183256b3d865c1e9cdc |
| moo2-probe-353-off-new.txt | 721b10698c61acb1cf5a86901e8e220ae1975f7e58b3e1ca9f0638d3a538e662 |
| moo2-vbe-353-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-353-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-353-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-353-attempt1-verify.py | ec899e6027caf0852ecf2fdf86f9ab40bf77bbc902dd6bbaa8098e3d481ddcdb |
| new-game-353-attempt1-tests.txt | a2a4c3afb986d052f64f24b2d54748628518941f36b592a2c44cc1b7573435fc |
| new-game-353-attempt1-verify-output.txt | 40d9fbbed66ee080b1f6d5e4107ebdb0e6ced366041f54704429354839aee198 |
| new-game-353-backlink-verify.py | 08adcb15b06cfa3436dc1640a2d4e8b4c6b0a492c36c2e55d53594a8438df934 |
| new-game-353-backlink-tests.txt | 0f8df807f7b5c18644d12f066cff38d3c7d9628b43d23fef4826cbe5fef361d2 |

來源／收據1000:1000，工具root-owned／.md目錄零；主庫歷史2437檔／272目錄保持，本輪不新增或遞迴修權限。原版、全套與回歸容器有界且已結束移除；精確HEAD／遠端與工作樹於收尾核對。

下一步為86 /r memory byte XCHG做窄CPU切片：先捕捉原DS:[ESI]／AL與ES:EDI的STOSB前狀態，審查byte交換／旗標保持／寫入邊界，再以同一180M正常輸入核對交換與下一byte store。不提高cap，不深入helper；完整生成／開局與remake同狀態未驗。

## 2026-10-03：354 byte記憶體XCHG、原STOSB與新ADD停止

路由命中dosgolem對拍／CPU規格閘門／回填／文件職責，沿已載入逆向技能與入口。起點主庫fc4c448a1c2b53d2eb990db1d45f44166b884ce6／工具c7086292bf476be63a406131632c9cf8c4780ab6。工具結果[1f15517](https://github.com/wicanr2/dosgolem/commit/1f155175b2c77e6ee133ef609f43b758d7e0eed8)，完整HEAD 1f155175b2c77e6ee133ef609f43b758d7e0eed8；[354規格](https://github.com/wicanr2/dosgolem/blob/1f155175b2c77e6ee133ef609f43b758d7e0eed8/docs/spec/354-cpu386-xchg-byte-memory-register.md)限定CONFORMED，較早352／353未知已回填。主庫RE-first保持，整款remake／中文化尚未完成。

### 公開契約與未改CPU原初態

[Intel 80386原廠XCHG](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/XCHG.htm)列86 /r為byte register與memory交換、旗標全部保持；不可寫／段外拒絕。memory XCHG硬體上即使無F0也有bus lock。本工具只處理單CPU的單Step邏輯交換，期間不插入機器tick；未驗lock電氣波形、多CPU仲裁或跨執行緒atomic，顯式F0仍拒絕。不深挖硬體實作。

DRAFT未改CPU，同353完整180M正常輸入取原dosgolem_high_le:223E93首遇。164321317原R=[FF 0 2 2BD976 2BD730 2BD888 2BD8A8 2BD97A]／段=[8 188 188 0 20 188]／flags202h，DF0；DS188:2BD8A7三bytes FF0E00，來源byte DS188:2BD8A8=0E，ALFF。ES188:2BD979三bytes00FFFF，STOSB目的ES188:2BD97A=FF。readonly與完整RAM保持、callback12／12及IRQ41873／41873非活動／pending0；原仍拒絕在after223E95。全部10812原353列／36PNG保持，三區塊逆轉source逐byte保持353；資料可讀與ISA充分，才審查READY實作。

### 原兩步、回歸與新停止

**已證實，原兩步**：164321317原86 06交換DS0E→FF／ALFF→0E，EIP223E95；來源window FF0E00→FFFF00，ES目的window00FFFF保持。原RAM只有index2BD8A8改0E→FF。164321318原AA將AL0E寫ES188:2BD97A，ESFF→0E、EDI2BD97A→2BD97B、EIP223E96；目的window00FFFF→000EFF，來源FFFF00保持。原RAM只有index2BD97A改FF→0E。兩步完整R各只變AL或EDI、段與flags202h保持；readonly=true／error nil，完整RAM差異各一byte，兩側相鄰byte保持。這是實際非零寫回，未猜欄位用途或caller語意。

**已證實，正常前綴與CPU**：原首遇前10742共通正常列／35既有frames與同一原DS來源／AL／ES目的保持。舊四項stop診斷不列正常前綴；DRAFT保持全部10812／36PNG與正式保持10742／35frames分開。CPU只以16行memory分支替換1行拒絕，readSegment8／writeSegment8成功後才發布來源reg8；有效地址和reg8來源在發布前固定，覆蓋AL／AH與base／index別名。失敗不發布R／旗標，自製fail-before-write Bus保持RAM，不假稱任意外部Bus可回滾。原86 register交換、平台與8088 CPU不改；正式probe與已驗private相同，三區塊逆轉保持353。

8來源byte register全部256×256配對，以獨立little-endian四byte視圖核算，所有ModRM／SIB／DS與SS相異／base與index別名／ESP忽略index／無base DS／負disp8／32位繞回／最後byte／相鄰資料、64旗標組合與非零FPU通過。成功相同byte仍一寫；未知／唯讀／段外／線性溢位／讀或寫Bus拒絕、prefix與截短、原86 register64配對通過。窄測0.710s，固定DOSGOLEM_MOO2_EXE的乾淨Go全套CPU386150.111s／machine1.864s通過。缺8088實機語料不算386硬體驗收。

**已證實，新停止**：164560803於原dosgolem_high_le input1CDD0F bytes02 45 F8 02 45 E4 02 45 FC 02 45 E0 00 43 07 8A，錯誤「byte運算記憶體形式尚未支援」。首三bytes為ADD AL,SS:[EBP-8]；SS188／EBP2BDB44，來源offset2BDB3C、AL0，原byte值未知。after1CDD11只解碼，ADD未執行。原R=[0 5A2044 5AA044 5AA5E8 2BDB18 2BDB44 2 0]／段=[8 188 188 0 20 188]／flags202h。actual stop164560803、requested budget180000000，尚未達180M；probe exit0只代表錯誤收尾，不是完成。

**已證實，視覺邊界**：較晚原finalPNG為1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457，RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17，確實不同於舊353final。Docker讀原PNG轉base64後視覺檢視，640×480主要為黑底，僅小型方形圖形可見，其用途未知；未見完整地圖，不算完整開局／GUI驗收。舊35frames保持只指原入口前，不把新圖包裝成同一終態。

**未知**：原SS byte加法來源／後續消費、資料語意、正式writer、完整生成／開局、RNG、人耳與remake同狀態。固定日期不是seed；沒有CPU位址特例、guest代寫／跳呼叫或重送。沒有新IDA、120M整流程重跑或失敗後挑選原版結果。

### 命令、環境與私有收據

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。Docker --rm／network none／UID1000／2GiB／2CPU／128pids，原版兩次各600s、乾淨全套600s、8M／CLI與窄測180s；原ZIP及patch唯讀。417根層檔／MOX.SET／99M按下與99084355放開／1996日期／44M硬體逃逸保持。

固定DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。CPU3b4f3dc4e3054bd92ce252f54202414c47dcc501257be7d0cf538c02ea449132；自製測試527846c29fc2b3da8043053ed8bdbf587c616db3216a6a4ccc13a1d9e51359ce；probe9c15f43b426eef78dbc983cf840df926b73817ecbbc5eea3d364a4fbb7e82d26。

```text
bash workplace/new-game-354-input-run.sh
python3 workplace/new-game-354-input-verify.py
  未改CPU10812原列／36PNG與原DS byte0E／ALFF／ES目的FF PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestXCHGByte' -count=1 -v
  獨立CPU byte交換與舊register契約 PASS
bash workplace/new-game-354-full-run.sh
  乾淨固定原EXE Go全套 PASS
bash workplace/new-game-354-formal-run.sh
python3 workplace/new-game-354-formal-verify.py
  10742共通前綴／35frames／原兩步／新ADD停止與finalPNG變化 PASS
python3 workplace/new-game-354-source-verify.py
  CPU16行替換1拒絕與三observer區塊／逆轉逐byte保持353 PASS
bash workplace/new-game-354-off-run.sh
  關閉8M1693列／PNG、68舊CLI＋32新CLI負例與正對照 PASS
python3 workplace/new-game-354-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-xchg-byte-memory-spec-backlinks
  91項回填／新354的34負例／352另2及較早負例 PASS
```

乾淨全套從git ls-files -z打包至/tmp/test-src，再複製本輪新CPU測試，未帶入歷史探索main。關閉8M以353原probe及354新probe同新CPU比較，兩者在本輪交換入口前；source逆轉與正式正常前綴另證舊行為保持。formal的新停止／PNG變化依據同一原收據補嚴格斷言重讀，沒有為選結果重跑原版。

以下27份本機來源／收據在工具忽略workplace；原EXE／LBX／RAM／LOG／PNG與私有腳本不入Git。公開只保存通用CPU、測試、診斷、規格、索引、守衛與雜湊。

| 收據／核算 | SHA-256 |
| --- | --- |
| moo2-xchg-byte-354.go | 9c15f43b426eef78dbc983cf840df926b73817ecbbc5eea3d364a4fbb7e82d26 |
| moo2-probe-354-input.txt.gz | 27f670f8c55613750722c2e6535f1f56b6a1dbf9864dda8bd4f9141853871b0c |
| moo2-vbe-354-input.png | d493c2b5628d55381176c9e676586ab8940fd62544302195b59570b6136e6ba6 |
| new-game-354-input-run.sh | d4be4a474c9dd88c40570c0825ab18bbec7c0b466f201f845fcb8a8042d09ece |
| new-game-354-input-run-output.txt | 63317953747d1219878a0a322dec995adcc33e6a389b782dee9739f843819e1b |
| new-game-354-input-verify.py | 14c3819d8c6d19c899af56b362fc85310b4172628042490c9b8223825d503564 |
| new-game-354-input-tests.txt | 1066544c02d6d09839a7f95527787382420b1a5c69db490fe59d0e0be629db1f |
| moo2-probe-354-formal.txt.gz | 7fe5e0408b1a24d44fcb8b02d3f618f218370aaa917648519d2d518cc1be5e43 |
| moo2-vbe-354-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-354-formal-run.sh | 56b50477e63a82e9156df24235b1c94aa9404027f81296cac36b92867715b838 |
| new-game-354-formal-run-output.txt | 8a2ee512fb30d871f36f75de0a6ff8fdba1bc42031be389fdb8984edd1d57d4c |
| new-game-354-formal-verify.py | 913a781bb8c07d74fa686f8ffa9e5f75e214c723b7975fa22bac82bff1ac8a73 |
| new-game-354-formal-tests.txt | 890a4ad081f9bdeb99f02628a7b0ffcafa0b1a338ed57c4b90dccb40076b18a3 |
| moo2-354-cpu-narrow-tests.txt | ac15f0602ad562d675a37433a727637d4f2c8d652ebec28545d981b139893d48 |
| new-game-354-full-run.sh | 29b29794f8b03e782186b3424e82ca2b2e030a43e1cccc7fd32bd1859f2d1925 |
| full-test-354.txt | ee150537f153e610e107754d74cb6de3abca7daad880a7927495d5fd80449393 |
| new-game-354-source-verify.py | e3db16dcf472815aab97aa3964f4a9e5ac68315a4e61d1f05d1b2aa67eccc462 |
| new-game-354-source-tests.txt | 11cb4362cd90f937579beeced8b5c8c20ec3663fb01925a7ec6ecded0f1092d0 |
| new-game-354-off-run.sh | 6cb00aa78d1c491d49d1e752f427ad65bffb57b84bbc46b93fb460bc28df4a38 |
| new-game-354-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-354-off-old.txt | 4c797d9053666d7d54370c71a72af0ddf3dc3eacd36d3c0cb6556fdf83697624 |
| moo2-probe-354-off-new.txt | 2e1bcbf17cd47568407dc42200aa8c0431ebc25ded2ba66134f76316f3b914dc |
| moo2-vbe-354-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-354-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-354-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-354-backlink-verify.py | 2800ded3b650d67cb6a16bb91ceb3a4acfffc2a349a24a4d08418ce2390130d6 |
| new-game-354-backlink-tests.txt | 1212830700af6cfbab65c597d0856168ae0b35fc00dfecc858cb5688e28f1b2b |

來源／收據1000:1000、工具root-owned／.md目錄零；主庫歷史2437檔／272目錄保持，本輪不新增或遞迴修權限。原版、全套與回歸容器有界且已結束移除，精確HEAD／遠端／工作樹於收尾核對。

下一步為02 /r byte ADD register,memory建立窄CPU切片，先捕捉原SS188:2BDB3C／AL0與後續原byte來源，再審查通用加法／六算術旗標／source唯讀／地址別名／失敗不發布。沿同一180M正常輸入驗原消費，不提高cap，不深入helper；完整生成／開局及remake同狀態仍未驗。


## 2026-10-03：355 原byte記憶體來源ADD與零值正常消費

路由命中dosgolem對拍、規格閘門、證據回填與文件職責；依既有入口與逆向技能前進。上一輪XCHG／STOSB已完成並推送，本輪為新CPU缺口，主庫RE-first保持。原檔／手冊／私有分析只讀，不改Go／Ebitengine玩法，不深挖helper。

### 來源、基準與推論等級

- 官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch ZIP908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5。根層417檔／MOX.SET與固定1996日期保持，日期不是seed。
- dosgolem原始位址基準dosgolem_high_le，工具起點1f155175b2c77e6ee133ef609f43b758d7e0eed8、主庫2031d07b7890ede231f21e8ff0e5e956116a3d7a；現工具8273d887f5387c23d9ae13056ae2ad1e263aeee0已推送github隔離分支並核對遠端，未推本機origin。使用/home/anr2/cht/dosgolem的隔離副本workplace/dosgolem，能力／用法見其README.md／CLAUDE.md。
- [Intel 80386原廠ADD](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/ADD.htm)為通用02 /r與六算術旗標契約。通用CPU沒有遊戲位址或資料代寫；沒有新IDA研究。工具CPU規格[355](https://github.com/wicanr2/dosgolem/blob/8273d887f5387c23d9ae13056ae2ad1e263aeee0/docs/spec/355-cpu386-add-byte-memory-source.md)為入口，352／353／354與000-index同次回填，限定CONFORMED。
- Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac已核對，UID1000／network none／2GiB／2CPU／128pids，原ZIP／patch唯讀。原版／乾淨全套600s，窄測／8M與CLI180s。原EXE／LOG／PNG／RAM與私有腳本留忽略workspace；公開自製CPU／測試／診斷／規格／索引／守衛與雜湊。
- **已證實**：DRAFT未改CPU10829原354列／36PNG／readonly與RAM保持；原四SS來源00／AL00與目的00可讀才READY。正式首遇前10759共通正常列／35frames保持，同一R／四source byte／目的窗口已驗。
- **已證實**：原164560803..164560806於1CDD0F／12／15／18的02 45 F8／E4／FC／E0，以SS188:2BDB3C／2BDB28／2BDB40／2BDB24四byte00加到AL00，首flags202h→246h；164560807原00 43 07把AL00加到DS188:5AA5EF byte00，EIP1CDD1E／flags246h。五步R／段／相鄰窗口／完整RAM保持，readonly=true／error nil，callback12／12、IRQ41945／41945非活動。原零結果不證原非零加法／進位；沒有Bus寫次數trace，不把00→00冒稱非零寫回。
- **已證實**：原164561579於input1CE387 bytes0F 94 45 F4 E9 8E 02 00 00 83 EF 04 F6 47 01 02拒絕；after1CE38A只解碼、SS188:[EBP-12] offset2BD834 byte未知。R=[5AA5F4 0 0 256 2BD36C 2BD840 5AA5E8 5AA614]／段=[8 188 188 0 20 188]／flags246h。actual stop164561579未達requested180000000，probe exit0是錯誤收尾。
- **已證實**：finalPNG仍1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17，逐byte保持354。沿354人工檢視主要黑底與小型方形圖形，未見完整地圖，不算完整開局。
- **已證實，工程契約**：八byte register全256×256與兩初態、64flags組合、完整ModRM／SIB與DS／SS、地址別名、負disp8／繞回／最後byte、唯讀來源零Bus寫／拒絕不發布／非零FPU；原02／22 register保持、22 memory仍拒絕。獨立較寬和／nibble進位／signed範圍／popcount與little-endian byte視圖未調CPU add8。窄測4.089s及固定原EXE乾淨全套CPU386123.004s／machine2.959s通過；缺8088語料不是386實機驗收。
- **已證實，回歸**：8M1693原列／PNG、68舊CLI＋32新CLI負例與100M／120M／160M／180M正對照、92項回填／新355的32＋4負例及較早負例通過。CPU16行02分支替換1行拒絕、三observer逆轉逐byte保持354，平台／8088 CPU不改。
- **環境／腳本修正**：8M attempt1套件建置誤納並行formal的暫存main而拒絕，未執行8M。保存失敗腳本／輸出，改明列main.go／irq1_prototype.go／irq7_prototype.go，同image／同命令乾淨重跑通過。不是產品缺陷；正式原版、CPU全套沒有失敗後挑選收據。
- **未知**：原非零ADD／進位、0F94目的byte與後續消費、資料語意、正式writer、完整母星配置／生成／開局、RNG、人耳、remake同狀態及Windows／macOS實機。主庫玩法RE閘門保持。

### 實際命令與收據

下列相對路徑均位於隔離工具workplace/dosgolem。
```text
bash workplace/new-game-355-input-run.sh
python3 workplace/new-game-355-input-verify.py
  未改CPU10829原354列／36PNG／四00來源與原拒絕 PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestADDByteSource' -count=1 -v
  全byte配對／六flags／地址別名／唯讀無寫與拒絕邊界 PASS
bash workplace/new-game-355-full-run.sh
  乾淨固定原EXE Go全套 PASS
bash workplace/new-game-355-formal-run.sh
python3 workplace/new-game-355-formal-verify.py
  10759共通正常列／35frames／原五步零值／新0F94停止／finalPNG保持 PASS
python3 workplace/new-game-355-source-verify.py
  CPU16行分支與三observer逆轉逐byte保持354 PASS
bash workplace/new-game-355-off-run.sh
  修明列來源後同8M1693列／PNG與68舊＋32新CLI負例／正對照 PASS
python3 workplace/new-game-355-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-add-byte-source-spec-backlinks
  92項回填／新355的32＋4負例與較早負例 PASS
```



CPU SHA-256 125674de469ef82d4abe457190e28eb0c75b88aca8a876349b395165e77c79e1；新測試3ccaf15df8f8ee1163465e7b41ae862ff26352f92a8f477fa554ba87e131c2c1；probe c872e72ed84611c4bb98fa8d9958e8e0036eafed28663d1172f209edf2666d54。

| 本機來源／收據 | SHA-256 |
| --- | --- |
| moo2-add-source-355.go | c872e72ed84611c4bb98fa8d9958e8e0036eafed28663d1172f209edf2666d54 |
| moo2-probe-355-input.txt.gz | 0c301fbf7d191d27e123b5a0ce826460f88e3bf9d711c2a1046eb14d8d17e443 |
| moo2-vbe-355-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-355-input-run.sh | f2ed9c42ca381d4f451864fc3ffedd2658e8ad2dacecfdcf27f8247ef1684f1a |
| new-game-355-input-run-output.txt | 3fa48b72e10393532f945d84d58a269dc42307a767ccdc38eddd6de30bde588d |
| new-game-355-input-verify.py | 903433d9d1df6e0aaa436a4d487754b32d2a9be8ea3b7464aacad99a867e8ad8 |
| new-game-355-input-tests.txt | 70fa7d7740252ef9f42c743e0c191502e5192a44e8f9c5bfd4b8fb89af47406c |
| moo2-probe-355-formal.txt.gz | 58c0834c577fe6a8a32b64ad82567f231e3ae56b314ce21a616d506475bb4454 |
| moo2-vbe-355-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-355-formal-run.sh | 1de0400f464ed73539d32a1da7370d8d803550fc6d3702e1b34edd8827448f30 |
| new-game-355-formal-run-output.txt | d216742120da1aade2164ebc5d38306dfe3979dd0c866b5a8036e0378d7765a0 |
| new-game-355-formal-verify.py | 339f7cd6d40024c31ba87369debee4b1ac63f54cffc68b30819ab352209a6d4b |
| new-game-355-formal-tests.txt | 4dc98de74d75ad7bb000a46b9ee42d19c56397e4736cbc089b65c71e8ba6ea27 |
| moo2-355-cpu-narrow-tests.txt | 1b71347efae7958e54a7fadb8169a9b8e56a57af24108a64dbf5fbccda959385 |
| new-game-355-full-run.sh | 67aed4dd8cb4d7d31f9e1fa04f1a21b7383f3c6528893cf4f465f92dc0b8ecc4 |
| full-test-355.txt | 336f8b09bdd31fba5da14c5b72a2fbc594cf0729f50b87953f548b9ff062467d |
| new-game-355-source-verify.py | dee62839f9620cee20d06b2bd772345642603f22ebc28530ebae1dcf18829568 |
| new-game-355-source-tests.txt | 14e7d812eb4f30208c887780b1c80af65e9be489436c91e8b94a5576e4abf4b6 |
| new-game-355-off-run.sh | 65df3aff5021caa6d926c3e5690b6605a3d2ba1e6e50d38b18bf70f935fe9b43 |
| new-game-355-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-355-off-old.txt | 30515f0274bda6c1bfd233531dee79e0dcedd18e11e98d12223ffa260324f028 |
| moo2-probe-355-off-new.txt | e009646a55c98984e0535808d87441277c50d0fe97a7beade9658514d1f15374 |
| moo2-vbe-355-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-355-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-355-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-355-backlink-verify.py | 2fa45fae3fbdb0e5f3cd47b6b59b86391d95a7c3993dfea4bc36a088ac2f694e |
| new-game-355-backlink-tests.txt | 2d37edfccf2a0a4d2fda48942e960387daa915af59b2604c3ba4dd6cfc096db0 |
| new-game-355-off-attempt1-run.sh | 2ac94e8e0ba001a75813c882a457dc0ac2cd58e208e594ca1abbd66580524faf |
| new-game-355-off-attempt1-output.txt | da277346d8fa7f358ee1d958e3d5c6ab6edb08bc4a13f9bba3eb54810c78eb32 |

來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持。原版、全套與回歸容器有界結束移除；收尾核對兩庫精確HEAD／遠端與工作樹。下一步只捕捉原1CE387的SS188:2BD834目的byte／相鄰資料、flags246h與後續消費，審查通用SETcc memory byte目的及寫回拒絕契約，沿同180M，不提高cap或深入helper；完整生成／開局與remake同狀態未驗。


## 2026-10-03：356 原SETE記憶體寫回與正常JMP

路由命中dosgolem對拍、CPU規格閘門、證據回填與文件職責，沿既有入口與逆向技能。上一輪ADD已完成並推送，本輪接新CPU缺口；主庫RE-first保持，不改玩法或深入helper。

### 來源、基準與推論等級

- 官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch ZIP908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5；417根檔／MOX.SET／固定1996日期保持，日期不是seed。
- 位址基準dosgolem_high_le，工具起點8273d887f5387c23d9ae13056ae2ad1e263aeee0、主庫ba3800608f7e55ad0a293104a57df9b41dc78669。現工具442eef487fa03be9ef0f793e233396120c56973b已推送github隔離分支並核對遠端，未推本機origin。使用/home/anr2/cht/dosgolem的隔離副本workplace/dosgolem，能力／用法見其README.md／CLAUDE.md。
- [Intel 80386 SETcc](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/SETcc.htm)提供SETE的ZF與byte0／1、旗標保持契約；完整16條件沿344已驗字面真值與CMP數學比較，不照鏡像SETG／SETLE的排字錯誤改條件。[Intel SDM Volume 2，SETcc 4-620](https://cdrdv2-public.intel.com/671110/325383-sdm-vol-2abcd.pdf)列ModRM.reg未用與不可寫／段外拒絕；其未用欄契約沿現模型擴至memory，不稱386實機已驗。沒有新IDA或遊戲位址特例。
- 工具規格[356](https://github.com/wicanr2/dosgolem/blob/442eef487fa03be9ef0f793e233396120c56973b/docs/spec/356-cpu386-setcc-byte-memory.md)為入口，355／354／353／352／344／343與000-index同次回填；限定CONFORMED只涵蓋CPU與原SETE／JMP，第三CMP與byte1 reader限制明示。
- Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac已核對，UID1000／network none／2GiB／2CPU／128pids，原ZIP／patch唯讀。原版與乾淨全套600s，窄測／8M及CLI180s。原EXE／LOG／PNG／RAM與私有腳本留忽略workplace，公開自製CPU／測試／診斷／spec／索引／守衛與雜湊。
- **已證實**：DRAFT未改CPU10834原355列／36PNG／readonly及RAM保持；SS188:2BD834 byte41／相鄰004100與flags246h可讀才READY。正式首遇前10764共通正常列／35frames與同一目的41／完整R／段／flags保持。
- **已證實**：原164561579在1CE387的0F9445F4，以ZF1把SS byte41→01、窗口000100，EIP1CE38B；RAM唯一變更index2BD834，R／六段／flags246h保持。164561580原E98E020000跳到1CE61E，全部R／段／flags／RAM保持；兩筆readonly=true／error nil，callback12／12、IRQ41946／41946非活動、pending0。
- **已證實，僅觀測**：原164561581的3B7DE0 CMP，EIP1CE621、目的000100與RAM保持，flags246h→202h。SS:[EBP-32] dword來源未取，不能獨立核算CMP，不列原第三步數值驗收；它不是SS:2BD834 byte1的reader。
- **已證實，新停止**：原164567987 input1CF90A bytes66 6B 7B 40 05 C7 45 D8 00 00 00 00 C7 45 E0 00拒絕word IMUL prefix，after1CF90C只解碼、未取ModRM／來源。R=[5A2044 5AA5F4 5AA5E8 5A2044 2BD920 2BDB54 5AA614 5AA5F4]／段=[8 188 188 0 20 188]／flags206h。DI目的／DS188:5A2084來源word未知／sign-extended imm05，actual stop164567987未達requested180000000；probe exit0是錯誤收尾。
- **已證實**：finalPNG逐byte保持355，SHA-256 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17。沿354人工檢視主要黑底與小型方形圖形，未見完整地圖，不算完整開局。
- **已證實，工程契約**：4194304個memory真值／初byte組合以字面位圖核算，所有ModRM／SIB、DS／SS與兩結果、負位移／繞回／最後byte／相鄰資料、純write／相同byte仍一寫／非零FPU與拒絕不發布。目的一byte Bus read拒絕仍成功；77值全部配對×10CMP條件的memory消費用signed／unsigned大小關係核算。兩舊memory負例限定未知selector，真值表／prefix／截短保持；原register／Jcc／SS store回歸通過。
- **已證實，驗證**：窄測7.994s、固定原EXE乾淨Go全套CPU386137.976s／machine1.755s，缺8088語料不算386實機驗收。關閉8M1693原列／PNG、68舊＋32新CLI負例與100M／120M／160M／180M正對照、93項回填／新356的37＋10負例與較早負例通過。
- **已證實，source**：CPU移除3行早拒絕加11行純write，原條件表／register／Jcc逐byte保持355；三observer逆轉保持355；兩舊負例只增未知selector及344單一函式名，逆轉保持。初態兩步source已保存，正式三步source只差budget；初次run內容另存original-run，重生入口改指向已保存初態。沒有重寫舊LOG／PNG或失敗後挑選收據，明列來源建置無暫存main問題。
- **未知**：第三CMP數值、byte1 reader、原其餘15條件動態逐條實測、新word IMUL來源／乘積／後續消費、資料語意、正式writer、完整母星配置／生成／開局、RNG、人耳、remake同狀態及Windows／macOS實機。主庫玩法RE閘門保持。

### 實際命令與收據

下列相對路徑均位於隔離工具workplace/dosgolem。
```text
bash workplace/new-game-356-input-run.sh
python3 workplace/new-game-356-input-verify.py
  未改CPU10834原355列／36PNG／SS目的41／原拒絕 PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestSETcc|TestSETLE' -count=1 -v
  全memory真值／初byte／純write／拒絕邊界與舊register、Jcc PASS
bash workplace/new-game-356-full-run.sh
  乾淨固定原EXE Go全套 PASS
bash workplace/new-game-356-formal-run.sh
python3 workplace/new-game-356-formal-verify.py
  10764共通正常列／35frames／原SS41→01與JMP／新word IMUL停止 PASS
  第三CMP只觀測，來源未取，不列規則數值驗收或byte1 reader
python3 workplace/new-game-356-source-verify.py
  CPU11行write／三observer及兩舊負例限定未知selector，逆轉保持355 PASS
bash workplace/new-game-356-off-run.sh
  明列source，關閉8M1693列／PNG與68舊＋32新CLI負例／正對照 PASS
python3 workplace/new-game-356-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-setcc-byte-memory-spec-backlinks
  93項回填／新356的37＋10負例與較早負例 PASS
```



DRAFT實際input-run內容保存於new-game-356-input-original-run.sh；重生入口指向moo2-set-memory-356-input.go。READY後正式三步診斷，不改180M或原輸入。CPU SHA-256 17854f07854ba4d59e70b9ac4ed4b1ec2b5e1a01b4df4336c60ef3bbed2ffe66，新memory測試ff6d3f8f64c7981391b935a67041f956efd998187dcfb742785799427f70d463，舊SETcc測試c123f9ff7e5cd167bf65240766f8053fdf9e7e220cf78f5de38bae8a60f2a93a／SETLE測試47c2013c857eb78c004b115b50031101b362f81105b6b04ea3b2911e8b845a7c，正式probe a72124624383618cc50c2ddc143efb3a92f9c3dd91cbf4bcfde9cffb658c4f55。

| 本機來源／收據 | SHA-256 |
| --- | --- |
| moo2-set-memory-356.go | a72124624383618cc50c2ddc143efb3a92f9c3dd91cbf4bcfde9cffb658c4f55 |
| moo2-probe-356-input.txt.gz | 1e0b74050f15e9c731e279ae1e5e2a843054c3fec207ad4a9f44fee118c0e9ff |
| moo2-vbe-356-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-356-input-run.sh | 01a8b807e0352e271deb2fea3d1df934ebaed1d4009b054d5f48841bdfc5b4e4 |
| new-game-356-input-run-output.txt | f7301b988f77ef2699421a569df215039f427ebf0f15407529dacf210090a8a3 |
| new-game-356-input-verify.py | 0794e79dcb65e0eaf83670d3958795659c2c69b173d8e96658018095a3e1aefa |
| new-game-356-input-tests.txt | 18b79b7ebac6885c214c3e4eba31bb096ed89425639ec1815f2a539cf7829e2b |
| moo2-probe-356-formal.txt.gz | 386ac92b1570e1f1f7d4956752f2b87606b06e41286bc1b8c11bafb87df54978 |
| moo2-vbe-356-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-356-formal-run.sh | 1a50457aea027fa68f7e5f719cf666307f63e5a2abd1df1bc2871e02155e2243 |
| new-game-356-formal-run-output.txt | f504b66c022d757ff38d36745bdb8e4071b3dcae17b30d76c5b44e7027c0273b |
| new-game-356-formal-verify.py | 2d0fbc24323252b84a0f3fc08f17673ba6b490d3f545f80143ff3478307bf8aa |
| new-game-356-formal-tests.txt | b41f7d6f821ca08b4e80ee14ca206dc16f34a46c5ba2ed53d215396f4b8bd850 |
| moo2-356-cpu-narrow-tests.txt | 5b42c680568d34298cb482dbee7785fd851048dde122815a44fa479a5b90a230 |
| new-game-356-full-run.sh | 1a6259c525794811b4942979fd5ab35f6392c59f67f8b71d87fd9869fc4dd71a |
| full-test-356.txt | 6315adceb9f636c4a9628b2850c6d753eaa6b59166e142bc12a853577259887c |
| new-game-356-source-verify.py | 0e2508824574782a0366c163dcad94b924573edebb107f9995dc1bc1226cd818 |
| new-game-356-source-tests.txt | 8c40e750c0ed61e4b28952d8bb2b67f8a04bd7a5b56442de748af902eb0bb0bf |
| new-game-356-off-run.sh | cfea9c865f9d931f96f74600effa2704456e02c42ff677d5fc056602e8709b42 |
| new-game-356-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-356-off-old.txt | 86bf4539d4864642eabccec65d8891583827743d5921ffdde9a6ff86d2e834f9 |
| moo2-probe-356-off-new.txt | 8caee365e8cc587f929964766ed721317ffd1dadf77fe139a0c29b74c7ca6512 |
| moo2-vbe-356-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-356-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-356-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-356-backlink-verify.py | ee1be0a913ee83f78185312dabb98495d9af18ee76f94d86fa71886fd30193bd |
| new-game-356-backlink-tests.txt | c74c882fe7c2ec2ebac2c325c34fb17ca76c3590ffd20f6eb882d6794aeb026c |
| moo2-set-memory-356-input.go | daac78a496a13350fbfe38961c1cfc9c7541049a05718b4ef9ac9359c5b1db11 |
| new-game-356-input-original-run.sh | 6d9362371bb4fa2701a5ee07e2481203f8fb419232268d3f6de9211e65d31ce4 |

來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；原版、全套與回歸容器有界結束移除，收尾核對兩庫精確HEAD／遠端與工作樹。下一步取DS188:5A2084來源word與DI／imm05及原消費，審查word IMUL低16bits／CF與OF、其他旗標未定義邊界／高16bits保持／拒絕不發布，沿同180M，不增加cap或深挖helper。完整生成／開局與remake同狀態未驗。


## 2026-10-04：357 原word立即值IMUL與兩MOV零寫

路由命中dosgolem對拍、CPU規格閘門、證據回填與文件職責；沿既有逆向技能，主庫RE-first保持。原版主要執行器/home/anr2/cht/dosgolem，使用隔離副本workplace/dosgolem；能力與用法見其README.md／CLAUDE.md。原EXE／LOG／PNG／RAM與私有腳本不公開，公開自製CPU／測試／probe／spec／索引／守衛與雜湊。

### 來源與已驗範圍

- 工具起點442eef487fa03be9ef0f793e233396120c56973b／主庫c7bd82f131c9afa5152466ce58ff1ea35b7e4b40；現工具a7f175f3bfadcc6d9d56a78e23c0eabf2cbe2442已推送github隔離分支並核對遠端，未推本機origin。[規格357](https://github.com/wicanr2/dosgolem/blob/a7f175f3bfadcc6d9d56a78e23c0eabf2cbe2442/docs/spec/357-cpu386-imul-word-immediate.md)為入口，356／355／354／353／352／344／343／268／269與000-index同次回填；限定CPU與原零積／兩MOV，未當完整開局。
- 官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5、417根檔及MOX.SET保持。固定1996日期不是seed；99M按下／99084355放開與180M保持。
- [Intel 80386 IMUL](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IMUL.htm)定義signed word來源×sign-extended imm8／signed imm16、低word目的與signed範圍的CF／OF；SF／ZF／AF／PF未定義。沿268與既有工具保留政策，標示工具近似，沒有新IDA／遊戲位址特例。
- Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac、UID1000／network none／2GiB／2CPU／128pids；原ZIP／patch唯讀。原版及乾淨全套600s，窄測／8M與CLI180s。位址基準均dosgolem_high_le，DS／SS selector188。
- **已證實，原擷取**：2026-10-03未改CPU的DRAFT10837原356列／36PNG、只讀／RAM與三observer逆轉保持。原164567987 DS188:5A2083四byte00000000，真正word DS188:5A2084=0000；SS188:2BDB28二十byte FFFFFFFF00000000000000000000000080091D00，來源／ISA充分才READY。
- **已證實，正式三步**：10767共通正常列／35frames／原R／六段／flags／code24／source四byte／frame二十byte保持。164567987 input1CF90A的66 6B 7B 40 05把word0000×5=0寫DI A5F4→0000、EDI005AA5F4→005A0000／高word005A保持，EIP1CF90F、CF／OF0。164567988／164567989兩C7把SS188:2BDB2C／2BDB34 dword0→0、EIP1CF916／1CF91D。三步ram_changes=[]／readonly true／error nil、callback12／12、IRQ41948／41948非活動／非failed／pending0。兩MOV不消費DI，沒有原Bus寫次數trace。
- **已證實，工程契約**：33554432個來源word×imm8×兩旗標、2097152個imm16×16 signed邊界來源×兩旗標，以獨立signed範圍與little-endian視圖核算；八目的×八register來源別名、全ModRM／SIB／DS／SS、唯讀來源／last word／線性溢位／第一與第二byte讀失敗／截短立即數／prefix及非零FPU通過。CPU移除word早拒絕加獨立width16分支，逆轉逐byte保持356；六舊測試／平台與8088 CPU不改，三observer逆轉保持356。
- **工具近似**：三步flags206h只驗定義CF／OF0，其餘旗標保存不當原硬體parity。原來源為零；非零積／負值／overflow由公開ISA工程測試覆蓋，正式原動態未驗。兩MOV零寫不當DI reader。READY初稿誤分組高word為5AA5已更正005A；CPU與獨立byte視圖及正式輸出皆為005A0000。
- **已證實，驗證與失敗**：窄測第一次7.631s在op6B cut1失敗，原因測試誤用CS描述符限制取指。取指直接讀Bus；以Bus拒絕第一個缺byte且來源完整可讀修正後，同命令4.372s通過，原失敗輸出保留。固定官方EXE乾淨Go全套CPU386121.065s／machine1.556s通過；缺8088硬體語料不算386實機驗收。8M1693列／PNG與68舊＋32新CLI負例、100M／120M／160M／180M正對照保持。94項回填、新357的38＋16缺證據負例與較早負例通過；首次文件縮寫缺完整164567989拒絕後補步號，同命令重跑，拒絕摘要保留，CPU／原收據不變。
- **已證實，新停止**：原164568139 input1CFD3F bytes66 F7 5B 38 EB 09 8B 55 E4 29 C2 66 89 53 38 83拒絕word memory NEG，after1CFD42只解碼，未取disp8或source。R=[0 0 4 5A2044 2BD920 2BDB54 0 0]／段=[8 188 188 0 20 188]／flags246h；DS188:[EBX+38h] offset5A207C word未知。actual164568139未達requested180000000；probe exit0是CPU錯誤收尾。
- **已證實，圖像界線**：finalPNG逐byte保持356，SHA-256 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17；沿354人工檢視主要黑底與小型方形圖形，未見完整地圖，不算開局完成。
- **未知**：正式DI reader、原非零IMUL／signed overflow、新NEG來源與後續消費、第三CMP數值／byte1 reader、資料語意、正式writer、完整母星配置／生成／開局、RNG、人耳、remake同狀態與Windows／macOS實機。主庫玩法RE閘門保持。

### 命令與私有收據

下列相對路徑均位於workplace/dosgolem，DRAFT兩命令是CPU442eef4時的歷史擷取，重生DRAFT須用該CPU與357只讀observer。
```text
bash workplace/new-game-357-input-run.sh
python3 workplace/new-game-357-input-verify.py
  未改CPU10837原356列／36PNG／word0000與frame／原拒絕 PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run TestIMUL -count=1
  attempt1 截短fixture失敗，Bus缺byte修正後4.372s PASS
bash workplace/new-game-357-full-run.sh
  固定原EXE乾淨Go全套 PASS
bash workplace/new-game-357-formal-run.sh
python3 workplace/new-game-357-formal-verify.py
  10767正常前綴／35frames／原word IMUL／兩MOV零寫／新NEG停止 PASS
python3 workplace/new-game-357-source-verify.py
  width16分支／三observer逆轉保持356、六舊測試不變 PASS
bash workplace/new-game-357-off-run.sh
  8M1693列／PNG與68舊＋32新CLI負例／正對照 PASS
python3 workplace/new-game-357-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-imul-word-immediate-spec-backlinks
  94項回填／新357的38＋16負例與較早負例 PASS
```

CPU SHA-256 94fab7c6e4389ce205c485dd498607f1442c20b3b7999212812f7e622fff4bf2、新測試628915d9b97a671bce0639d43e7644ec91a07563123c09e5cc4ce6c225291036、正式probe e08cf955cfb1b543f62a8cb573d2b6784630dc2bdc44f06fe47820a008cba2a7。input observer與正式相同，診斷budget均三步，沒有重寫舊LOG／PNG或提高cap。第一次截短失敗為實際輸出，第一次回填守衛拒絕存摘要，均與通過收據分列。

94項規格回填全過；新357的38項缺證據／限定範圍／狀態／356回填／索引負例、其餘八份較早回填另16及舊負例全過。首次步號縮寫拒絕摘要另存attempt1，CPU與原收據不變。來源／收據均1000:1000，工具root-owned／.md目錄零。

| 本機來源／收據 | SHA-256 |
| --- | --- |
| moo2-imul-word-357.go | e08cf955cfb1b543f62a8cb573d2b6784630dc2bdc44f06fe47820a008cba2a7 |
| moo2-probe-357-input.txt.gz | 47e10ec4128cb451881b59741a9f46ebdff374cdd27409b8834edafad4d4fc0b |
| moo2-vbe-357-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-357-input-run.sh | d8e2e70d86ed132223d836591af364cc94798a7c79e97c0c64c059d27064d0c3 |
| new-game-357-input-run-output.txt | de768f16a275eed1b0aea16376f2b912c69c608592a2f607ad2d199f4840bb39 |
| new-game-357-input-verify.py | 211548c27443824d4c0776e47045459a99ebd44e595a1575d05cc5d1fc421ff6 |
| new-game-357-input-tests.txt | 315f1767ae0a5d03a245c1d87759f9a16d9007539b6ce04cce3e472f0b977f56 |
| moo2-probe-357-formal.txt.gz | 464f05e17bb736116a05a8f18d81edf45a9edcd0f40e4d3fd82c41b937822b71 |
| moo2-vbe-357-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-357-formal-run.sh | 5baa5359f7b8a6ae721585a475e9da3d33df08809a53c800d4dcdc494aa80cf1 |
| new-game-357-formal-run-output.txt | 5c4821ee66c0580a83900be3c096977f43471946a3767e1a9adb19f88a04c89b |
| new-game-357-formal-verify.py | 765061983c4fcc3e906389efe7fd982d415916ef03106d5cbe6d1432fd2b14ce |
| new-game-357-formal-tests.txt | 2835387eca804693752d275f41e7c34a8c227ae074a788181058ee77edd9654a |
| new-game-357-unit-tests.txt | 33d4dc610a69ab1ba9663ad9ad2863e0b6c8a37cbf1e38e617ee693c3dba5883 |
| new-game-357-full-run.sh | 2277cfdca014a70dbd902dc3eae95e3268c7cb2789ff0037035e3020b30d565b |
| full-test-357.txt | 114cc77fa894252e3fcc3a20b57a4992690101f642e00d1532f9c6c733730a16 |
| new-game-357-source-verify.py | eade6f79cc4ad877e18a912ec5b5c02ac496d0cb0857922e84976d433ae03e99 |
| new-game-357-source-tests.txt | 0207dbe9a870631abca1548e17d591ba6e559d9c1f05f6537313bd2762afb4de |
| new-game-357-off-run.sh | f697bb45bc9c9bc9d71847e7ee9fc6e0a01d98083fa20e17db25c3ffaf8c50ad |
| new-game-357-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-357-off-old.txt | 3b1453f36f0ce3d75a1bee99c2d5de4aec4ac463b60417134e753130e3297b36 |
| moo2-probe-357-off-new.txt | f438eea44fd91d5267b2a6f77693434fe039a638dc21fa58f69a81a5448ffb09 |
| moo2-vbe-357-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-357-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-357-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-357-backlink-verify.py | 9c20c505262dc71b14e1f9ddbc7f07c6d85c770b340fa2773bbc52b898b9d44d |
| new-game-357-backlink-tests.txt | 0870d186f33b7775bac8d228a186e72275f5e1c95d2ff8791d4f047a84801ec6 |
| new-game-357-unit-tests-attempt1.txt | d31e60ddb220607f0f5ab62749843d20e1113f177505cc8003c514bb9f373e29 |
| new-game-357-backlink-tests-attempt1.txt | a11e1282fe16a014f6d6f7f79e9d61d37ff991d4f1122fc96c3fb3ef8e38aa9e |

來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；本輪原版／全套／回歸容器有界結束並移除。下一步擷取1CFD3F的DS188:5A207C word NEG與相鄰bytes、自然EB09分支／消費，按公開ISA審查，沿180M不增加cap／不跳指令或代寫／重送／深入helper；收尾核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-04：358 原word NEG與下一EB09

路由命中dosgolem對拍、CPU規格閘門、證據回填與文件職責，沿既有逆向技能。主庫RE-first保持，原版執行器/home/anr2/cht/dosgolem採隔離副本workplace/dosgolem，能力／用法見其README.md／CLAUDE.md。原EXE／LOG／PNG／RAM與私有腳本留本機忽略目錄，不公開；只公開自製CPU／測試／probe／spec／索引／守衛與雜湊。

### 來源與證據等級

- 主庫起點3b9ec0266baac64d5d6cc80ee4405afd73c4d6e9／工具a7f175f3bfadcc6d9d56a78e23c0eabf2cbe2442；現工具a995e5d62249aef97f73cc52e11176acc6e3218b已推送github隔離分支與核對遠端，未推本機origin。[規格358](https://github.com/wicanr2/dosgolem/blob/a995e5d62249aef97f73cc52e11176acc6e3218b/docs/spec/358-cpu386-neg-word.md)為入口，357／356／355／354／353／352／344／343／268／269與325共十一份較早入口／現行unknown及000-index已回填。原357停止保留歷史定位，舊NEG待擷取已移除。
- 官方DOS1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；原ZIP3a28a52f5953ff6d8fc251548940500236752ee19b52b51581e71ec1a3373c2f、patch908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5／417根檔與MOX.SET保持。99M按下／99084355放開、固定1996日期及180M保持，日期不是seed。
- [Intel 80386 NEG](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/NEG.htm)定義word目的的二補數取負、零來源CF0／其他CF1；[Intel SDM Volume 2B NEG 4-161](https://www.intel.cn/content/dam/www/public/cn/zh/documents/64-ia-32-architectures-software-developer-vol-2b-manual-cn.pdf)列六算術flags按減法結果定義。沿325既有逐byte Bus模型：第二byte晚期寫失敗可已改第一byte，但全部寫成功前不發布flags；不稱任意Bus回滾或硬體exception restart。不深挖遊戲helper或作新IDA。
- Go1.24.13 Docker image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac／UID1000／network none／2GiB／2CPU／128pids，原ZIP／patch唯讀；原版與乾淨全套600s，窄測／8M／CLI180s。位址基準dosgolem_high_le、DS／SS selector188。
- **已證實，未改CPU擷取**：DRAFT全部10840原357列／36PNG保持，三observer逆轉保持a7f175f，CPU仍94fab7…；原164568139 code24 66F75B38EB098B55E429C26689533883BDCCFDFFFF000F84，DS188:5A207B十二bytes全00，真正來源word5A207C=0000。SS188:2BDB28二十bytes6901000001000000000000000800000000000000，與NEG目的分開；readonly／RAM保持，原仍拒絕after1CFD42。原來源／ISA充分後READY。
- **已證實，原NEG／EB09**：首遇前10770共通正常列／35frames、原R／六段／flags／code24／source與frame逐欄保持。164568139原word0000→0000，六flags CF0／OF0／SF0／ZF1／AF0／PF1、flags246h保持，EIP1CFD43；164568140原EB09自然跳1CFD4E，全部R／段／flags／RAM保持。
- **已證實，僅觀測與條件**：164568141第三CMP 83 BD CC FD FF FF 00讀SS:[EBP-564] dword，來源offset2BD920未在擷取窗口；僅觀測flags246h→206h／R與RAM保持／EIP1CFD55，不列CMP數值驗收／NEG word reader。164568142原0F84 BC ED FF FF以觀測ZF0不跳到1CFD5B，flags206h保持，四步readonly／error nil／ram_changes=[]、callback12／12與IRQ41948／41948非活動／非failed／pending0。原非零NEG／8000溢位、真正Bus兩write次數未驗，不以零RAM差異升格寫次數parity。
- **已證實，工程契約**：4194304個memory source word×64初flags、1048576個八register×65536 source×兩初態，用獨立modulo／signed邊界／nibble借位／低byte位元計數與little-endian視圖；全部ModRM／SIB、DS／SS相異資料／地址繞回／last word、唯讀／未知selector／段外／線性溢位、兩byte各read／write失敗／晚期部分寫但flags不發布、完整目的prefix／截短定址與nonzero FPU通過；工程另驗零仍兩write。register低word與高word保持規則、其餘R／六段／FPU與非算術flags保持。
- **已證實，source與測試**：CPU只加24行word group3分支，memory寫回成功才sub16；CPU與三observer逐byte逆轉保持357。325舊66負例改明確未知selector，268舊word register NEG負例加segment prefix，兩修改可逆轉；新全值域正例接合法word，七舊測試不變。窄測6.224s、固定官方原EXE乾淨Go全套CPU38666.221s／machine1.331s、8M1693列／PNG、68舊＋32新CLI負例與100M／120M／160M／180M正對照、95項回填／新358的41＋20負例與較早負例通過；缺8088語料不算386實機驗收。沒有CPU／原版／驗證失敗後挑收據。
- **已證實，新停止**：原164610300 input1D0944 bytes66 29 83 E9 00 00 00 5A 59 5B C3 53 51 52 56 57拒絕word SUB memory目的，after1D0946只解碼／未取ModRM或source；R=[0 64 0 5AA5E8 2BDB4C 2BDBA0 5AA614 5AA5F4]／段=[8 188 188 0 20 188]／flags206h。DS188:[EBX+E9h] offset5AA6D1目的word未知、AX來源0000，actual164610300未達requested180000000；probe exit0只是CPU錯誤收尾。
- **已證實，圖像界線**：finalPNG逐byte保持357，SHA-256 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17。沿354人工檢視主要黑底與小型方形圖形，未見完整地圖，不算開局驗收。
- **未知**：原非零NEG／8000溢位、CMP數值／NEG word reader、SUB目的與正常消費、正式DI reader／前輪CMP及byte1 reader、資料語意、正式writer、完整母星配置／生成／開局、RNG／人耳、remake同狀態與Windows／macOS實機。主庫RE-first保持。

### 實際命令與收據

下列相對路徑在隔離工具workplace/dosgolem；DRAFT兩命令為CPUa7f175f時的歷史擷取，重生DRAFT須用該CPU與358只讀observer。
```text
bash workplace/new-game-358-input-run.sh
python3 workplace/new-game-358-input-verify.py
  未改CPU10840原357列／36PNG／word0000與原拒絕 PASS
go test -p 2 -buildvcs=false ./internal/cpu386 -run 'TestNEG|TestNeg|TestIMUL|TestMUL|TestF7|TestTEST' -count=1
  word NEG全值域／六flags／拒絕／高word與原乘法 6.224s PASS
bash workplace/new-game-358-full-run.sh
  固定原EXE乾淨Go全套 PASS
bash workplace/new-game-358-formal-run.sh
python3 workplace/new-game-358-formal-verify.py
  10770正常前綴／35frames／原NEG零值與EB09／CMP限制／新SUB停止 PASS
python3 workplace/new-game-358-source-verify.py
  word NEG分支／三observer及兩舊負例明確限制逆轉保持357、七舊測試不變 PASS
bash workplace/new-game-358-off-run.sh
  8M1693列／PNG與68舊＋32新CLI負例／正對照 PASS
python3 workplace/new-game-358-backlink-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-neg-word-spec-backlinks
  95項回填／新358的41＋20負例與較早負例 PASS
```

CPU SHA-256 9840611ea0e3ae22ece69fd1f6f545dd08a316d1ed87247bbe061bf3f7f09522、新測試ab96fa39e050cca58f1f0b8c46275eeb0e3f3f56d9aa3210e9148f58a51839e3，舊NEG dword測試3ec49cd5357fdb65d416b92da4f13519155c0138572f8357a0ad68ccfaf60307／word IMUL測試9edb03bb303a26ec1f3437820c59d2d7f3cd558c80a308f8f9d7ec0cd5074010，正式probe07243e3f82540f6ffb9c99a3134e187b20dc27194593661f677587010df7633c。input與正式observer均四步budget、來源相同，不改舊LOG／PNG或增加cap。

95項規格回填全過；新358的41項缺證據／限定範圍／狀態／357回填／索引負例、其餘十份較早回填另20及舊負例全過。來源／收據1000:1000，工具root-owned／.md目錄零。

| 本機來源／收據 | SHA-256 |
| --- | --- |
| moo2-neg-word-358.go | 07243e3f82540f6ffb9c99a3134e187b20dc27194593661f677587010df7633c |
| moo2-probe-358-input.txt.gz | 9c323583a0c88033b62fe9f27a73b3bf76ea15a51cc215ce8335ff613c144bfd |
| moo2-vbe-358-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-358-input-run.sh | 3e04ede51d6ab24a39508e3d78c72e20d44acc3755e8a2d387a215519a6e0c50 |
| new-game-358-input-run-output.txt | 4ac677bbdf9f4a579fc88a9a38199682c5f558cf33d906c6146e92751541d891 |
| new-game-358-input-verify.py | b1690e006c5a1be106b15814caa75033fa344511bcfaa53b81a65cdb697cc09b |
| new-game-358-input-tests.txt | c4f76167e13434999fb6090b08e2fe8340f9378b157ea4377885d98196d7f5ed |
| moo2-probe-358-formal.txt.gz | a334a424881733be645038a99b2d7f36a724ae3a833a66caa77737f45905d25e |
| moo2-vbe-358-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-358-formal-run.sh | 2bf94975bfcca9d618b4e75ca47f21d130cfc9bc6ac30d46b0124bc480f1b8c4 |
| new-game-358-formal-run-output.txt | 4b7651bd0e452ae221567624329c0afbcd94f4c6b522563602e5246fa48065a4 |
| new-game-358-formal-verify.py | d9bc9b87429f5da9a665b26cdd9425d553bbfe5d04e637545206b2f1fc8fdb3f |
| new-game-358-formal-tests.txt | b3a36d75334bee537a87f4c1c7a532694a3cd50a1e20507395cabf0bc63027f0 |
| new-game-358-unit-tests.txt | b21e27933d33b6c56def7b143991c1672e18f73842a711888c0dbda3352e8be9 |
| new-game-358-full-run.sh | c57f83b67b2ea9fe12253ea75f31e6447fe8b870b56150a38be0a765eb0b2230 |
| full-test-358.txt | 8a600babc81848f516e97569ee5098953c560ffa6cd5bc6e147278522f392020 |
| new-game-358-source-verify.py | 5f6f4a6da59b40fca52395537c8d8e891496404e0d710fc01f17b08054b0507f |
| new-game-358-source-tests.txt | d4e7e1d223bb949d809985b4f0eb6a1f9698ca78668989da21cfba3c026ae99c |
| new-game-358-off-run.sh | 4d05336977f159fe680ba9387de60e646c6e53e783edc16ac034d9276cfcfcf8 |
| new-game-358-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-358-off-old.txt | 40043f24aceea509507d0aa217cb3e349783b4acaf4ab5bf4fd00237b3631c96 |
| moo2-probe-358-off-new.txt | 27d1b0979f14131c1dd33cc8d15b959f72aee09e6ddd77d3ec0122432865876a |
| moo2-vbe-358-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-358-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-358-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-358-backlink-verify.py | aca396e00133a792a03645020d61830f4f15b5e30c3883b681639ca57cb137ce |
| new-game-358-backlink-tests.txt | 460a5e9c5b2a48db1f4dc21138eefdf436c303f1b34f787c4a10c170f6b329b9 |

來源／收據1000:1000，工具root-owned／.md目錄零，主庫既有2437檔／272目錄保持；本輪原版／全套／回歸容器有界結束並移除。下一步取1D0944的目的word／來源AX、後續POP／RET消費，審查29 /r word SUB，沿180M不增加cap或跳指令／代寫／重送／深入helper；收尾核對兩庫精確HEAD／遠端與工作樹。


## 2026-10-04：359 word暫存器來源SUB與原三POP／RET

路由：dosgolem原版oracle／spec-gated-workflow／re-resolution-backlinks／project-document-responsibilities；原版仍固定官方DOS1.31 EXE 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。沿原180M與單次正常輸入，不提高cap／跳指令／代寫／重送／深入helper。工具11d9aad0d10bcf51ac75f9ee23611acfe2fd7e9b已推送github隔離分支，遠端精確核對；未改上游/home/anr2/cht/dosgolem。

規格與完整命令：[359限定CONFORMED](https://github.com/wicanr2/dosgolem/blob/11d9aad0d10bcf51ac75f9ee23611acfe2fd7e9b/docs/spec/359-cpu386-sub-word-register-source.md)。DRAFT未改CPU先取目的與真正SS槽，再READY實作，原輸入於CPU改動前核對；其驗證後以/tmp舊CPU a995e5d快照保存重驗輸出，沒有重跑guest。主庫玩法RE-first仍關閉。

### 已證實與工程驗證

- 未改CPU10845原358列／36PNG與正式10775共通正常列／35frames保持。
- 原164610300在1D0944的66 29 83 E9 00 00 00，DS188:5AA6D1 word0003-AX0000=0003，六flags CF0／OF0／SF0／ZF0／AF0／PF1，flags206h、EIP1D094B與相鄰00／0E保持。
- 原164610301／164610302／164610303三POP按SS188:2BDB4C真正槽到EDX0000000E／ECX005AA5E8／EBX00000000，ESP逐dword；164610304 RET按001D1E0B槽，EIP1D1E0B／ESP2BDB5C。其餘R／段／flags與固定窗口／全部RAM保持。五步callback12／12／IRQ41960／41960、readonly true／error nil／pending0／非active／非failed。
- POP／RET不是目的word reader。原同值Bus寫次數未取；工程另驗來源0仍兩write，不以無RAM差異冒充原寫trace。原非零SUB來源／借位／溢位未由此路徑驗證。
- CPU只新增29 word分支與移除該早拒絕；來源／目的別名先取舊值，register高word保持，memory兩byte成功後才發布六flags。逆轉逐byte保持358，既有ADD／其他SUB／flags helper與十一舊測試不改；三只讀observer逆轉原probe。既有66 2B word原已支援，本輪不改。
- 獨立整數差／modulo／signed範圍／nibble借位／低byte位元計數與little-endian視圖，2097152 memory全來源×邊界目的×兩flags、1048576八register別名全值域×兩flags，另全src／dst邊界配對與131072 memory八來源×64flags×邊界配對。全ModRM／SIB／DS／SS、signed位移／繞回／unaligned／last word、合法memory目的prefix／截短、各byte read／write失敗與nonzero FPU通過。第二byte晚期部分寫可先改第一byte，但不發布R／flags／FPU，非硬體restart或任意Bus回滾。
- 窄測0.685s、固定官方原EXE乾淨Go全套CPU38682.239s／machine3.104s；缺8088語料不算386硬體驗收。關閉8M1693原列／PNG，68舊CLI＋32新180M負例與100M／120M／160M／180M正對照；96規格守衛與新41＋22缺證據負例、較早負例均通過。
- 首次窄測把既有2B word誤設拒絕，回查原分支／原測試後改保持性驗證。關閉腳本第一次錯用未掛原始資料入口，改完整既有入口；新停止驗證的bytes空白欄位以完整尾碼比對修正。三者都乾淨重跑，不把測試或環境問題寫成產品缺陷，也未重跑挑選原guest結果。

### 新停止與限制

原164984957在dosgolem_high_le input1D2A33 bytes66 99 66 2B C2 66 D1 F8 98 01 C7 81 FF FF 7F 00拒絕operand16 CWD，after1D2A35只取prefix／opcode。R=[1 0 0 5A2EED 2BDB24 2BDB58 5A2EED A]／段=[8 188 188 0 20 188]／flags246h，AX0001／DX0000；尚未達180M。下一步最多三步只讀診斷取CWD前後與下一SUB／shift的R／段／flags／RAM，審查99 word sign-extension，沿同180M。

finalPNG逐byte保持358，SHA-256 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17；沿354人工檢視主要黑底與小型方形圖形，未見完整地圖。固定1996日期不是seed；目的word reader／欄位語意、CWD正常消費、正式writer／RNG／完整生成開局與remake同狀態仍未知，不新增玩法spec或修改主庫玩法。

CPU SHA-256 995f059949b7caac9618ab8b2513999b64b4cb2c928bd608da96eff171d3a8d1；新測試a7a6cb7e44c56b3f49a035c0bcbf87978196c63a59f294ff88a0d9de57b6d0bd；正式probe1d72d00195ff24b6c9e2ba6e48385c6c40e222a5e2e4417f2a21e39984b8434d；舊NEG／word IMUL／SETcc／ADD／XCHG／AND測試保持。

### 私有收據與隔離狀態

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；原版與全套600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。原EXE／LOG／PNG／RAM與私有腳本留忽略workplace，只公開自製程式、規格／索引／守衛及雜湊。工具樹乾淨／專案容器0，既有root-owned2437檔／272目錄保持，不做遞迴修復。主庫只改CONTEXT的DOS單行、WORKLIST的DOS活表、追加WORKLOG與本研究檔，其他活表全文保持。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-sub-word-359.go | 1d72d00195ff24b6c9e2ba6e48385c6c40e222a5e2e4417f2a21e39984b8434d |
| moo2-probe-359-input.txt.gz | 45694965a26be50c4581fa5cc5862d45adcf17a9c6d102efc323b2ce4dd749e6 |
| moo2-vbe-359-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-359-input-run.sh | c4ad45371017abc0bf2e32552c81f73e2f86f6a538cd376c78985ae29c1b9f7f |
| new-game-359-input-run-output.txt | 93be53dc55c5eb22a08e4f683347c0965363ab7ce86ef29d9a5d558f2fc69958 |
| new-game-359-input-verify.py | 0799b673cd025fbf37ac0fb63602270899ba3d2401db362f91f968379020ba81 |
| new-game-359-input-tests.txt | 40ec9df86a3be0dc6d05a2b36adb1439cb1ce3fa6b2a50afa940891748885be3 |
| moo2-probe-359-formal.txt.gz | a5421d88e46dea38d24ad92c49d8abd1a46ed341d93e25ead73efcfcd09ec856 |
| moo2-vbe-359-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-359-formal-run.sh | 3e6927213c8f81cae873d1cb95ebb027d1a471dabb13f232ce0f93564558556b |
| new-game-359-formal-run-output.txt | 242535a7b416cfa9a90593cc8b9186aa835a80b6a2b7a3df27d781db04b5635f |
| new-game-359-formal-verify.py | 427c829e0ac23972ad1820ad2b41a5f4449fc95d5009f2e6331eaf8fdb9f3a37 |
| new-game-359-formal-tests.txt | 289344414b72db164dbe3bfe2a81a21f725bc26a0924a2d09a75f0e42ab45bc1 |
| new-game-359-unit-tests.txt | d82c13509fce2ecd11968c1b06f1e4786689734b359a0a2c532bb816d774277e |
| new-game-359-full-run.sh | 7a2b9ef3c96f6b59233f785e51090974c8becdbb1de5f675b7f1e492bd91b085 |
| full-test-359.txt | f341b54cc3b33ecc90c7900fb9b1a9a29033434b3d2c62e2a56868d799cec0c7 |
| new-game-359-source-verify.py | c6c4284f241d4ac1fb7a1327c1df25289c1d4b731003b2a39ea160ce1ede3de7 |
| new-game-359-source-tests.txt | 7d08bc04f2726e0a7f7e822f19a580eda7b98326ccba31c6d4252668913890b6 |
| new-game-359-off-run.sh | d87864b67b10fa9a206fb4ce98a29d7573ac7de562773fc149bc5c01f3b4a612 |
| new-game-359-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-359-off-old.txt | e488250f0b4f940dae079332df4ee18ad431691c799e8d0408511de55bf76c7f |
| moo2-probe-359-off-new.txt | 65d95947d29cf69464ede23a4a770b40c5dd6450ddc8dd045074b0002ca836d2 |
| moo2-vbe-359-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-359-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-359-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-359-backlink-verify.py | a1e23ad060dd3d78f8905f402fb18ce9b2c1e36fdd0ade59090f1375482210d1 |
| new-game-359-backlink-tests.txt | 10dcf5f100ca93d8aeded6fc51ab5d6a627991944d2b66c8328907d8eecd6edf |


## 2026-10-04：360 word CWD與原SUB／SAR消費

路由：dosgolem原版oracle／spec-gated-workflow／re-resolution-backlinks／project-document-responsibilities，寫結論前再核對。沿固定官方DOS1.31 EXE 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原180M與單次正常輸入，不增加cap／跳指令／代寫／重送／深入helper。工具a16e94c4c15e5c78760cc670cb4f6766457a5f67已推送github隔離分支，精確遠端核對，未改上游/home/anr2/cht/dosgolem。

完整規格與命令：[360限定CONFORMED](https://github.com/wicanr2/dosgolem/blob/a16e94c4c15e5c78760cc670cb4f6766457a5f67/docs/spec/360-cpu386-cwd-word.md)。DRAFT先以未改CPU取原AX／DX、FPU與RAM，輸入驗證通過才READY，僅實作工具word CWD；主庫玩法RE-first保持。

### 已證實與工程驗證

- 未改CPU全部10859原359列／36PNG、正式10789共通正常列／35frames保持，同一原完整R／段／flags／FPU原bits及code16核對。
- 原164984957在1D2A33的66 99，以AX0001符號正／DX0000→0000，完整flags246h、EIP1D2A35與其餘R／段／FPU／RAM保持。
- 原164984958的66 2B C2讀真正DX低word，以SUB AX0001-DX0000=0001，六定義flags CF0／OF0／SF0／ZF0／AF0／PF0、flags202h與EIP1D2A38。原164984959的66 D1 F8讀AX1，SAR AX0001→0000，CF1／OF0／SF0／ZF1／PF1、EIP1D2A3B／EAX0，其他R／段／FPU／RAM保持。SAR AF不列原版parity，完整flags202h→247h只是觀測工具模型。
- 三步FPU控制127F／status0／depth0／八stack bits0與全部RAM保持；readonly true／error nil／callback12／12／IRQ42073／42073、pending0／非active／非failed。原AX正且舊DX已0，不稱原負AX／EDX高word或非零舊DX動態已驗。
- CPU只移除99的operand16早拒絕並新增五行word CWD；逆轉逐byte為359，裸CDQ原一行／word SUB與SAR／flags helper不改。三observer逆轉11d9aad原probe，十三舊測試完全保持。
- 獨立signed數值範圍與little-endian視圖，16777216個65536 AX×64算術flags×四EAX／EDX高word哨兵，源高位符號相反也依AX，完整flags／EAX／EDX高word／其他R／段／nonzero FPU／RAM保持。全部32flags位元×16邊界AX、截短每byte／合法來源prefix拒絕、裸CDQ／十組CWD→SUB→SAR正負奇偶來源通過。
- 窄測0.828s／固定原EXE乾淨Go全套CPU38699.385s／machine1.679s；缺8088語料不算386硬體驗收。關閉8M1693原列／PNG另比前輪359原收據，68舊CLI＋32新180M負例與100M／120M／160M／180M正對照；97守衛與新360的42＋24缺證據負例及較早負例通過。本輪CPU／原版／驗證均未失敗後挑選guest結果。

### 新停止與限制

原168496272於dosgolem_high_le input2376CB bytesC1 0D C4 0F 27 00 08 A1 C4 0F 27 00 4E 74 46 25拒絕memory dword ROR，after2376CD只取opcode／ModRM，未取disp32／imm8／source。R=[18181818 0 110 5C 2BD5D8 2BD648 69 34B94C]／段=[8 188 188 0 20 188]／flags202h。目的DS188:270FC4／imm08，原dword未知，不能以EAX18181818猜source；尚未達180M。下一步取原dword／相鄰資料與下一A1同址真正load，審查C1 /1 memory ROR，沿同180M。

finalPNG逐byte保持359，SHA-256 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457／RGB04fef4b6a6d6c6c485ef1ce0a82ea71591956cdd3b7cd37b8d1082a023e20e17，沿354人工檢視主要黑底與小型方形圖形，未見完整地圖。固定1996日期不是seed；原負AX／EDX高word、新ROR來源與正常消費、正式writer／RNG、完整生成／開局與remake同狀態未知，不新增主庫玩法spec或修改玩法。

CPU SHA-256 ed94eaf7e9c363e8a2c89d5410b30a9e653a60532d31654ee91c14d042f75337；新測試bad93e84a75a9e4a7e77c74e904a10a74ed23e82c0919b0ae006581c28b72181；正式probe8d65d37030f58fb8fd034b2396d9da00c3617987211e750da517cdc8b3c1bd9c。來源helper與其他舊CPU測試保持，不把工程自洽當原版全值域硬體對拍。

### 私有收據與隔離狀態

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。原EXE／LOG／PNG／RAM與私有腳本留忽略workplace，只公開自製CPU／測試／probe／規格／索引／守衛及雜湊。工具樹乾淨／專案容器0，既有root-owned2437檔／272目錄保持，不做遞迴修復。主庫只改CONTEXT的DOS單行與WORKLIST的DOS活表、追加WORKLOG與本檔，其他活表全文保持。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-cwd-word-360.go | 8d65d37030f58fb8fd034b2396d9da00c3617987211e750da517cdc8b3c1bd9c |
| moo2-probe-360-input.txt.gz | 4143ae1c60ec3957ca373d4ac1be397ec9932daaa384ad997ab07f2b3820cd99 |
| moo2-vbe-360-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-360-input-run.sh | 26e87efec02f945cd3b7f5fe59f5782f8a3106e2251b426ac082b93994686eae |
| new-game-360-input-run-output.txt | bf41f1ab66bfc8d33bb80e36994128fd89aab7928e5e88c89af923bcff1927d2 |
| new-game-360-input-verify.py | c261f469f17d48aefa68c1253d8f221037cff2180638248b6f89e47a3f8d6990 |
| new-game-360-input-tests.txt | c1a14e9f83aab4afbe67045390dd2a317758fe6b2db3535853378fc2c488588e |
| moo2-probe-360-formal.txt.gz | a05e0f46fe83e5fb048112cef174471599185d7a1c20f11d3a623565ddb156e3 |
| moo2-vbe-360-formal.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-360-formal-run.sh | cef6c7cdf5dab4880504c41409153d8604f9a5caaa6360b4500fb54a9c69fcac |
| new-game-360-formal-run-output.txt | ceac0b93c9af0ec069eb7063359f029b6e129ed430b6062b0672afd39f187178 |
| new-game-360-formal-verify.py | 0aa89c540754008531d111fba662e1a0a8d3af3898980942f56c4f46fd3b743d |
| new-game-360-formal-tests.txt | d8696c2be76aa43b67e8003b078cb92dbb858d027b762cd4ee5a83426b12b90a |
| new-game-360-unit-tests.txt | dc141c466dd16697abf74895e64e160216a95629cabec5296b4fedb500ccbe89 |
| new-game-360-full-run.sh | 6a790107b572abfa549210758fcea49907a2e0c281f6466df0b9aa706a9659dd |
| full-test-360.txt | 7a638098bb95b91ee21387d09949d76e590ba9225be9eb0a5cbf07eba236936f |
| new-game-360-source-verify.py | 6127647b58f12bb3680f3f87e38f5dd7a941335a84ea83d42a2cbaec5cc29e5e |
| new-game-360-source-tests.txt | bf3d0a64c1d6c6b50911a3241c276a668ad1ae4e38a5ad0385bc853124d0a200 |
| new-game-360-off-run.sh | d46319f88606f40c6ad16affc6346146380fdfd0ac090cd31bcc447282228ce5 |
| new-game-360-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-360-off-old.txt | c0b9bdf2b2e95878b6d6a34053f9da63395e752b3925ab774ea2a207acaa5726 |
| moo2-probe-360-off-new.txt | 67c25bdf682de6e28276dd5437be8f412d70deed8ec3c0b6ac5a495bb29b0aff |
| moo2-vbe-360-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-360-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-360-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-360-backlink-verify.py | a12abc12f0390921488c2be3a32f5e508daf528461918e95df6fe349c650edd4 |
| new-game-360-backlink-tests.txt | 2f15caa9dd2a0509fc267e29461e2bf074c6446ca62cc43a8bc5325cb746f390 |

## 2026-10-04：361 memory ROR與原A1消費

路由：dosgolem原版oracle／spec-gated-workflow／re-resolution-backlinks／project-document-responsibilities，寫結論前再核對。沿固定官方DOS1.31 EXE 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、同180M與單次正常輸入，未增加cap／跳指令／代寫／重送／深入helper。工具5a2170cfc18b40e890a21dcdcbb10605f084b57d已推送github隔離分支，精確遠端核對，未改上游/home/anr2/cht/dosgolem。

完整規格、命令與證據：[361限定CONFORMED](https://github.com/wicanr2/dosgolem/blob/5a2170cfc18b40e890a21dcdcbb10605f084b57d/docs/spec/361-cpu386-ror-dword-memory-imm8.md)。DRAFT未改CPU取原目的與相鄰資料，11019原360列／36PNG保持才READY；主庫玩法RE-first保持。

### 已證實與工程驗證

- 原168496272／dosgolem_high_le:2376CB的C1 0D C4 0F 27 00 08，DS188:270FC4來源000B1818不同於EAX18181818，ROR8→18000B18、CF0／EIP2376D2。RAM僅270FC5／270FC6／270FC7不同，首byte及左右0000保持；不把三個差異當三次Bus寫入。
- 原168496273的A1 C4 0F 27 00真正讀新目的，EAX18181818→18000B18／EIP2376D7，其他R／六段／FPU／flags202h及全部RAM保持。兩步readonly true／error nil／callback12／12／IRQ43075／43075非active／非failed／pending0，FPU127F／status0／depth0／八stack bits0保持。
- count8 OF未定義，完整flags202h只驗沿297保留工具模型；非CF及非count1 OF的flags保持。原count0／count1 OF／CF1未由本路徑驗證。
- 正式10949共通正常列／35frames保持。CPU只加memory分支，逆轉為a16e94c原CPU；三observer與297舊負例未知DS限定也可逆轉。十四其他舊測試與297 oracle保持。1212416個74來源×256imm8×64算術flags，以原單bit整除循環oracle獨立核算；ModRM／SIB／DS／SS、masked0零write／其他count四write、失敗與截短／prefix拒絕通過。
- count0先驗可讀可寫及讀目的，是工具政策，真硬體Bus行為不列parity。四byte晚期失敗可能先寫前bytes但不發布R／flags／FPU，不稱Bus rollback或硬體restart。
- 窄測1.025s、固定原EXE乾淨Go全套CPU38660.001s／machine1.824s、關閉8M1693原列／PNG及前輪360原收據、68舊＋32新CLI負例／正對照、98守衛及新39＋28缺證據負例與全部舊負例通過。缺8088語料不算386硬體驗收。核對腳本原arr拒空[]，修正後讀同收據通過，原guest未重跑挑選。

### 180M端點與下一個玩家阻塞

同輸入達step_limit180000000／EIP215DCA／R=[FFFFFFFF D6 2C0864 A5 2BD630 2BD654 FC D8]／段=[8 188 188 0 20 188]／flags202h，bytes89 45 F4 83 7D F4 FF 74 35 8B 55 EC C1 E2 02 A1；無guest_cpu_stop／step_error／DOS exit，不稱完整開局。

finalPNG SHA-256 679f08239fe789c2f1ae82b5fa9a72a08cad884aa8d12f756a3fd9badf7f4c3a／RGB0d093a071c686fc389622b2bc2d80a30d2a051a8162ef819fb2d4976484c57c5。2026-10-04人工檢視：SELECT BANNER COLOR上有Error saving game／Permission denied及CLOSE，未見完整星圖。終圖已改變，不再將現況稱主要黑底或與360相同。

已證實工具政策：probe使用ReadOnlyFileProvider／OpenDirectoryReadOnlyFiles，le_startup.go的writeFile要求io.Writer，唯讀提供者拒寫。已有overlay_files.go的machine.OpenDirectoryOverlayFiles(basePath,statePath)及保持來源／寫入／截斷測試。終圖權限錯誤與此政策相關為強推論，真正DOS呼叫／檔名／mode／errno尚未知。下一步先有界唯讀擷取失敗契約，再審查既有覆蓋層與容器新state目錄，原版來源唯讀，不建重複檔案系統或代寫存檔。

固定1996日期不是seed；原其他count／CF1／count1 OF／硬體Bus、多位OF、存檔呼叫與內容、正式writer／RNG、完整生成／開局與remake同狀態未知。主庫RE-first保持。

CPU SHA-256 b8c1844163fddd7e3557e19fc51abcb9bcb9c1d021fd415dae376b2afbb9c72b；新測試96bee4520e99c09b5edb8cda11de95fbf745811f15c1b9545c1ccc56f4663ab1；297修改後測試8e0e36c37f5f3a0469475ec4f5b6c19693964986cf437acafc1f6e2cfc3a1bca；正式probe4fd257a6edf9a3460af7f7d792ab2624474f0fd7070f3eb5ef9877d2357cd787。

### 私有收據與隔離狀態

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。原EXE／LOG／PNG／RAM與私有腳本留忽略workplace，只公開自製CPU／測試／probe／規格／索引／守衛及雜湊。工具樹乾淨／專案容器0，既有root-owned2437檔／272目錄保持，不做遞迴修復。主庫只改CONTEXT的DOS單行與WORKLIST的DOS活表、追加WORKLOG與本檔，其他活表全文保持。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-ror-memory-361.go | 4fd257a6edf9a3460af7f7d792ab2624474f0fd7070f3eb5ef9877d2357cd787 |
| moo2-probe-361-input.txt.gz | 40eb10b9ba8e3fd07c94d2a9e2bd632b9cb990e5d96da65e6e007aa6360f6e96 |
| moo2-vbe-361-input.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-361-input-run.sh | 9192a27b6fc1479dff1d346e1151bdd2aedf493862a19814df491686fb5ff6e7 |
| new-game-361-input-run-output.txt | 283ad3d4dd868fb4d4c13a59ead2fdbf6fad7fef90314750de8bd391cb3e37c5 |
| new-game-361-input-verify.py | 51b5a2fd56c5c4f76e8cfe64af7b192caf96ebe04a820d43c5c995e73523d71b |
| new-game-361-input-tests.txt | 2da58bb8da84644d5d8e0a21e56c465d784a07875b9aa5fe1e414316d759af04 |
| moo2-probe-361-formal.txt.gz | 40fb333d1aaeb9160779b6fe56b439b390d19b9d6f6ddfb7d50ec7d577eff512 |
| moo2-vbe-361-formal.png | 679f08239fe789c2f1ae82b5fa9a72a08cad884aa8d12f756a3fd9badf7f4c3a |
| new-game-361-formal-run.sh | d35f4426c1d2e503b9767803a9ad94f49c1c7abcbf11d7a27e0728e1ef3979b5 |
| new-game-361-formal-run-output.txt | f3d230940aeea63afd71fea139439828044dfc4ef9d7ed1cc0522c1cc9b91856 |
| new-game-361-formal-verify.py | 0394ba2ae1dc00b53989458efa79609db75eb1d62f3cb0fc86d0ea9c0a1f6f7d |
| new-game-361-formal-tests.txt | 2df647f060a30cb14a75590cf058df3c5d33f43f55c366cdd3cb30918cfc050a |
| new-game-361-unit-tests.txt | 4f22548c265fcd52aa5cdc964d6ae17a6b83d979e14990f779aa635e1e2c0da6 |
| new-game-361-full-run.sh | c7301ef7b3537e267a36707097a02063b35dd9f02a82954ff74dd336c9a387dc |
| full-test-361.txt | 965899b1f79e4a78f356155613b1da31feb50f65f3104a74f3f48c040a1aa081 |
| new-game-361-source-verify.py | 9bf3f08a2f558948bdec8785b1539cee4f07db3e9964e4c550dcf2a0559fe8d0 |
| new-game-361-source-tests.txt | a20d4eca2b0a176498b862e7abef057a09483ea6aa460d0854ffd393482de270 |
| new-game-361-off-run.sh | 29534b33231463f3aa9b51374f7ab98208ae0262e61aeefb73d87a981b0b9a0e |
| new-game-361-off-cli-tests.txt | a5eabaf6f165680e4e73a808efc13d87da68f29a4c55ca63bd6bf4ec96c09dff |
| moo2-probe-361-off-old.txt | 9db2981eb8655004c1697457522407cce118170997414a5156c97e3e0fae4f7a |
| moo2-probe-361-off-new.txt | d13916fa91b68d8150a9bef1747abff488ef9944c0325fd44b8e1191ecc1aa86 |
| moo2-vbe-361-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-361-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-361-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-361-backlink-verify.py | 6791241d913126c88dfd4127634a237abd0965268b18e91949910f2dbefdd7b8 |
| new-game-361-backlink-tests.txt | 53eccf8b9d3282363b85031e24a17d83ebe14c6dd7f0bacbb3913e2e623a969b |

## 2026-10-04：362原SAVE10唯讀開寫拒絕與可寫DRAFT邊界

路由：dosgolem原版oracle／spec-gated-workflow／re-resolution-backlinks／project-document-responsibilities，寫結論前再核對。工具4f9c45be2017904ea42d86ef9b7ae692388eee08已推送github隔離分支，精確遠端核對，上游/home/anr2/cht/dosgolem未修改。固定DOS1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、417根檔與同180M，固定1996不是seed，未增加cap／注入／代寫或重送。

完整證據與實際命令：[362限定唯讀CONFORMED／可寫DRAFT](https://github.com/wicanr2/dosgolem/blob/4f9c45be2017904ea42d86ef9b7ae692388eee08/docs/spec/362-moo2-save-permission-boundary.md)。主庫玩法RE-first保持。

### 已證實

- 原165025480在dosgolem_high_le:237024的INT21 AX3D01，DS188:2BDB68字串SAVE10.GAM可讀NUL終止。原R=[3D01 270E74 2BDB68 42 2BD6AC 2BDB1E 2BDB68 FFFFFFFF]／段=[8 188 188 0 20 188]／flags202h；handled true，AX0005／CF1／flags203h，其他R／段與全部RAM保持。
- 唯讀provider無WriteFileProvider，openReadOnly型別檢查拒絕。這次是開寫失敗，不是AH40失敗，也不要求AH3C新建；ZIP已有SAVE10.GAM。拒絕早於361的168496272 ROR，後段合法CPU契約不能當生成成功證據。
- 未改CPU／DOS／provider的128有界唯讀診斷，361全部14498列依352既有mtime／DTA／每輪RAM診斷雜湊正規化保持、38PNG逐byte保持，各次診斷自身CPU／RAM／FPU／VBE保持。首次直比因重新解壓mtime不同失敗，改用既有正規化讀同收據通過，不稱跨輪RAM全相同。原guest未重跑挑選。
- 公開probe只加兩診斷區塊，可逆轉361。所有internal／CPU／DOS服務／provider及舊測試逐byte保持361。99守衛與新29＋32缺證據負例及全部較早負例通過，17較早規格回填。

### 可寫試作未通過玩家驗收

沿既有overlay接空state並委派base清單，窄測0.053s、既有overlay0.037s、固定原EXE乾淨Go全套CPU386130.397s／machine1.643s、原8M1693列／PNG／68舊＋32新CLI與五state拒絕／空state正對照通過。新列舉測試僅由窄測執行，全套測原有已追蹤程式與測試。首次窄測NUL欄位長度誤寫12byte而實際11byte，及逆轉腳本多一空行，均修正驗證後通過，不記成產品缺陷。

原版overlay試跑exit1，在80M setup_accept_precondition valid=false停止，panic「設定頁ACCEPT點擊的原表或輸入條件不符」，未送ACCEPT press、未到165M存檔、無finalPNG及state正式寫入收據。RGB3bbe6c339cfbc74e84b3210bccfdc4574073b4d308ef9b90a1720cc5c74e623e、globals／header及幾何保持。935bytes七byte不同，index=[650,651,705,760,761,815,870]；record11–15的+44四byte窗口從004AFA3C／004B0D14／004AE634／004B2084／004B3294各增8000h，其他表內容逐byte保持。第1177共通列已不同，唯讀1260000 EIP24659F／overlay1410000 EIP238291。heap地址解釋為強推論，所指內容與消費未知，不能稱已證實標籤指標或同初態。

可寫profile回DRAFT，試作與新列舉方法／測試只留忽略workplace，公開不接state旗標與新方法，原336完整表guard與點擊時刻保持。原失敗收據保留，沒有挑成功guest重跑。容器exit後state未保留，不能聲稱原版正式寫入或全部來源檔案不變已驗。下一步有界保存初段開檔mode／路徑／結果與80M五窗口候選位址的內容，每側五個最多128byte／NUL視窗，依原描述符及只讀狀態驗證，再審查READY正常輸入契約。

歷史formal/full/source/off腳本記錄當輪試作版本。未來明示重生試作使用workplace/new-game-362-prototype-rerun.sh，容器/tmp從固定361 Git源碼與已雜湊試作另建source；不覆寫舊收據，本輪未執行重生入口。正式存檔writer／內容、RNG／完整生成／開局與remake同狀態未驗。原EXE／PNG／LOG／RAM與存檔不進Git。

### 私有收據與清理

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。公開probe SHA-256 63b0182fc7311d4e0760dc29906419a1480ea3fa12faf1e58067f61775104bb1，與可寫試作f1f2be8d3a2a15cb0fa0d7013b64f8ed97d9798f5b9a2b6d85b69f570602043a分開。工具樹乾淨／專案容器0；既有root-owned2437檔／272目錄保持，本輪不新增root-owned或.md目錄。主庫只改CONTEXT的DOS單行與WORKLIST的DOS活表、追加WORKLOG與本檔，其他活表全文保持。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-save-diagnostic-362.go | 63b0182fc7311d4e0760dc29906419a1480ea3fa12faf1e58067f61775104bb1 |
| moo2-save-overlay-362.go | f1f2be8d3a2a15cb0fa0d7013b64f8ed97d9798f5b9a2b6d85b69f570602043a |
| overlay-files-362-prototype.txt | 3c775f009959f54ae6572c854e0b8b049ea7dd4f8c9228345a1297b611daf21a |
| moo2-overlay-enumeration-362-prototype.txt | 38d4f013df53e422349e67861c5d6258047210c8e040c80f74cf5b1e73a7db00 |
| moo2-probe-362-input.txt.gz | f05a445160916c9da6b70ad85d24c18733b7a6d7c6ac1a21921d7e6b9d5669b3 |
| moo2-vbe-362-input.png | 679f08239fe789c2f1ae82b5fa9a72a08cad884aa8d12f756a3fd9badf7f4c3a |
| new-game-362-input-run.sh | b1c3c4c536155b65e117b3b7b407a5d7472cfec3549b1ebc5bcbb8a7502f6c78 |
| new-game-362-input-run-output.txt | 3408cddb753c02c6e95d87091a117c6bad045d1c1153147744c690af2b5516e6 |
| new-game-362-input-verify.py | cff7c6c7ba8af38a387ac150281a1b41ca337567dc10834488f6ac02e1723869 |
| new-game-362-input-tests.txt | b3706925e325823f7a242a097307f4183bc638e1e746462119f64142104e16c1 |
| moo2-probe-362-formal.txt.gz | 5bc986900a6f8fcd2a5089f93dfbd83a7540a9bb4fac729832b8054770155a62 |
| new-game-362-formal-run.sh | 055464dec165a7d72e72f932ba785491056c59cf92e953ce9fa61c94b5527321 |
| new-game-362-formal-run-output.txt | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| new-game-362-prototype-rerun.sh | 5495873d6065a8fa898ae84255b77c067c4ef6a0131b9a4014742687ea06425e |
| new-game-362-unit-first.txt | 97cc21a2f46984c6226404181374bc6082a68f7fd502bccfc6004082d1c32c74 |
| new-game-362-unit-tests.txt | a0b221267f1587d9433d39356146322498bd0d34fc30e7a0b3d5b767ddc8831a |
| new-game-362-existing-overlay-tests.txt | 03b8dca9b60bee4420a1a8eaa1787f313051b59e7a91a69d4f36b98a04a9cdbf |
| new-game-362-full-run.sh | b705b197af9db444118f4f65298ba633262f98347e3e334c242e018f2d5fe516 |
| full-test-362.txt | 0a338e0ec95dde8f75872e4d3e2504b1ca6a512fea32310dfca6c5a161c3461a |
| new-game-362-source-verify.py | 86dc0ae849b90d8554cc9d6946913e8c6f9cb4809e5fad1035e410ac23673620 |
| new-game-362-source-tests.txt | 8ae6b6cf1db5d9d3d3849fb67933d60648e06fdf93fedfe0febf9ef6ee9565f9 |
| new-game-362-live-source-verify.py | 620be2d788ae58a3cf1d3ebba7d809114fe8bc9dde9288094d9852d5cfb50cf2 |
| new-game-362-live-source-tests.txt | 3a754db108d9bf58f783ce382bf1185c82b0d6b50afb2218e2b0be2ceb130504 |
| new-game-362-result-verify.py | 4ce178708f6d42efee1671bea33121bef400c3c68a5121ab2f00b26cc8dbffbe |
| new-game-362-result-tests.txt | 0bf1e660d5cbbd4150ddbca0aa3848f2ea636870b95d4529eb6a0085b66b7948 |
| new-game-362-off-run.sh | 599c293f3bd9d6b8359fe34bde7feefdb7e0b4c92ba7cc2deded301fcc9005cb |
| new-game-362-off-cli-tests.txt | 02243db97e3e984c4b4f4e43aa3b58aa773c4caf35ca7fc6b31c2125b0cdc70d |
| new-game-362-state-cli-verify.py | 81a4bc54cbe240756991c0dc0fb9592d51cd015c4c3a2e392c040c4fcdee8211 |
| new-game-362-state-cli-tests.txt | 13eefb0e4c8a34ff6e80e77f933661db278f9c7b98a7ec49243334069424e177 |
| moo2-probe-362-off-old.txt | 3f9009096ed960d8ce3edc5fe8c1f0a26606bcc3ad74fdd050bf965ef2dcf515 |
| moo2-probe-362-off-new.txt | 696ba37cc2a6ddb15e8adb7dedef3c2c1ff7677223719d0fa59b796882288317 |
| moo2-vbe-362-off-old.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| moo2-vbe-362-off-new.png | dd4c21dd11e57ec86ef759182fd3ebadd6b53870b54de1ef192e2553820286db |
| new-game-362-cli-tests.txt | 0c39c7d30d26388473df3e31ad7be38651d38f98677857d9e5a9de59a4aaac43 |
| new-game-362-backlink-verify.py | 8f065021b447755d27ba85e07b0a1f4f64edbe22eef2a1fc7c0d93e8d0c2d420 |
| new-game-362-backlink-tests.txt | 44b1392d7f74613becf5bf2fafe80b9537002244b9055524118642b84799aa43 |

## 2026-10-04：363覆蓋層前段開檔與設定頁五窗口

起始主庫088df139965dfc699b56efb4b577d87eae4e0f77；公開工具4f9c45be2017904ea42d86ef9b7ae692388eee08，成果[363限定規格](https://github.com/wicanr2/dosgolem/blob/04a96f09538a6b01907685f6f56a8d7de154cde1/docs/spec/363-moo2-overlay-startup-and-setup-source.md)。CPU／DOS服務／provider／全部internal及公開probe保持起始版，公開probe SHA-256 63b0182fc7311d4e0760dc29906419a1480ea3fa12faf1e58067f61775104bb1。官方ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；所有原位址為dosgolem_high_le。

**已證實**：原1192795／237024 INT21 AX3D02，DS188:26C3AF的sound.lbx，兩側R=[3D02 270E26 26C3AF 43 2BDB34 2BDB9C 26C3AF FFFFFFFF]／段=[8 188 188 0 20 188]／flags202h與RAM SHA-256 d89e8bbe89447eaa4834b9b9a9cf102146b5d77d3d4b901545d71aa3e1f84c34相同。唯讀AX5 CF1／203h拒絕，overlay真handle5 CF0／202h成功，其他R／段／RAM保持；同AX5不能視為相同。這是前2M最多64筆中的八個AH3D所捕捉最早差異。八筆無AH40，不外推2–80M無寫入。後續fonts.lbx／orioncd.ini step及handle改變。

80M表pointer298848／17筆／stride55，兩側935byte hash原2a18a0213dcb1d539de3175c8356b8c8b886c1d61b38f63795b3fa1859a3f52f／overlay49374b4c6dfd2d1d8231cfc137e1b5b7d86c7ec49f6fdf3be7417480d0da948b。record11–15的+44四byte候選值各增8000h，原DS188 descriptor所指128byte可讀且逐byte相同，首NUL在offset1，不能以NUL截斷二進位資料。窗口逐項原值／overlay值與SHA-256見上述固定363規格，不公開原bytes。RGB仍3bbe6c339cfbc74e84b3210bccfdc4574073b4d308ef9b90a1720cc5c74e623e；各側CPU／FPU／VBE／RAM自檢前後保持，不聲稱兩側完整狀態相同。**強推論**：額外32KiB配置引起位移；未追allocator。**未知**：完整物件長度／角色／消費端。

兩側原418檔guest前後SHA-256保持。overlay state只有sound.lbx，4250888bytes、UID及GID1000、SHA-256 3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d，另讀原ZIP比對完整bytes及本機保留副本通過。未有SAVE10.GAM，不稱存檔成功。私有副本／128byte窗口／所有原PNG及LOG不入Git，不散布。

2026-10-04親看原overlay 80M PNG：NEW GAME設定頁640×480，Tutor／Medium／Average／5 Players／Average，Tactical Combat未勾，Random Events與Antaran Attacks勾選；ACCEPT完整，原433,392–527,414包含480,400。只驗當前設定頁，沒有送ACCEPT。

### 命令及限定結果

既有Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；Docker600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP及patch唯讀；隔離來源及state在容器/tmp，自製輸出在忽略workplace。

```text
bash workplace/new-game-363-pair-run.sh
python3 workplace/new-game-363-pair-verify.py
python3 workplace/new-game-363-state-source-verify.py
python3 workplace/new-game-363-backlink-verify.py
```

兩側各一次原guest，解析180M情境但80M快照後明示診斷停止exit0，不是180M完成。唯讀6148／overlay6271既有列依352的mtime／DTA／每輪診斷RAMhash正規化保持，各27PNG逐byte保持，原guard唯讀true／overlay false。原核對腳本把27張寫28，修正明示27檔集合後讀同收據通過，沒有guest重跑。八早期開檔／五窗口／全狀態快照只讀通過，三診斷區塊逆轉為各自362來源通過。100項規格回填、新37＋34及較早全部缺證據負例通過。

### 私有收據

所有檔案位於 `workplace/dosgolem/workplace/`，SHA-256與UID／GID1000已核對。PNG／LOG／state只留本機。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-setup-source-363-readonly.go | 606dd5927f9315cdbf92ecf6132783137b936e44d609f38be6de69c9d87a0b5c |
| moo2-setup-source-363-overlay.go | 8d956c8851556391e01f85c3a8e3a4676acf5341e406c7773ffeda1292af0820 |
| new-game-363-pair-run.sh | 35bfda9d39e56477125bb9f6505b1ce694958666a171cece3495e4bbcf8e8432 |
| new-game-363-pair-run-output.txt | 6c7c5f09bf978b013fd5fe4d74ab33aaecd606ead07f24d4646e3819890a71c9 |
| moo2-probe-363-readonly.txt.gz | bebf7c93f2a802cc1c6241985eafe0b8955cbfc034fa76f39ef679b0f91dfe6e |
| moo2-probe-363-overlay.txt.gz | 72f5cc29637de198914526676953f53c99fb333c20f5dafa449d921fe7c11811 |
| moo2-save-state-363.json | e35322cba18956bcd245e4091932c822f9460271deda20ddda8b7722aa1375bc |
| new-game-363-pair-verify.py | 65afec78a9563a1e9df3f212fbd681af47e38b443b83034022c539d653073579 |
| new-game-363-pair-tests.txt | a74b63c1c9d88fb73a41376e7ec26508b93cdbe18150b766190275516b22c03e |
| new-game-363-state-source-verify.py | ded59fb88373f6f142a5855e79c71155528ecfe0398c93c254214ca547519e76 |
| new-game-363-state-source-tests.txt | 1e4a74c2a520e13a297711e93f87d753cd63e6db531045fe137a5df55bab5315 |
| new-game-363-backlink-verify.py | 6cac016298bcce7b2c3d5ffee8633f3b9e41461f78045c018b76d47354e4aedc |
| new-game-363-backlink-tests.txt | d3cb9da7430e763f13edd98f0979ec5cb882af3d107e1e1c7a374811b56a72d1 |
| moo2-363-overlay-frame-extended-80000000.png | 1507dfb323dd4614fee5eb1523ab60c5652c7eb2a7fa6c53182772b5bdc64908 |

### 目前閘門

工具成果04a96f09538a6b01907685f6f56a8d7de154cde1已推送github隔離分支；主庫RE-first保持。下一步另立DRAFT可寫profile正常ACCEPT規格，使用已觀察完整overlay935bytes hash及五窗口／RGB／原hotspot／callback／IRQ／IF，不改336唯讀guard、不mask位址／套任意+8000h／調時刻。證據足夠才READY，正常press／release各一次後取真正store及90M選族表，再審查後續輸入。正式存檔、完整開局、RNG、音訊與remake同狀態仍未知。

## 2026-10-04：364可寫覆蓋層正常ACCEPT與90M選族頁

主庫起始f731b75b01861a08e2e2b0df03ec86009a145a7b；工具成果[364限定規格](https://github.com/wicanr2/dosgolem/blob/98950c28961c0f89ed63304cd131477d083c85b7/docs/spec/364-moo2-overlay-setup-accept.md)。原ZIP417根檔另加官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f共418檔，MOX.SET／1996／原180M參數與正常先前輸入保持。日期不是seed，所有原位址為dosgolem_high_le。

### DRAFT → READY → 限定驗收

原363覆蓋層的完整935byte表及五窗口／RGB／完整R／段／EIP／flags／FPU／callback4／4／IRQ17901／17901、唯一ACCEPT熱區逐項核對，同原收據通過才READY。原336唯讀guard不改，另以私有profile核對真實overlay全表SHA-256 49374b4c6dfd2d1d8231cfc137e1b5b7d86c7ec49f6fdf3be7417480d0da948b；五個+44值與前128byte SHA照363逐項檢查，沒有忽略欄位／加任意位移／調點擊時刻。完整物件／角色仍未知，不追allocator／renderer考古。

**已證實**：原80000000正常press x960／y400／buttons1，virtual_micros133462143；80013765於133507499 release x962／y400／buttons0，差45356µs且正常回呼5／5完成後首次送。原80119128在20DDDB執行六byte66A3A6C42600，DS188:26C4A6 word0000→0F00，下一20DDE1；R=[F 0 339 0 2BDB14 2BDB7C 0 2B0001]／段=[8 188 188 0 20 188]／flags297h保持，callback6／6、IRQ17934／17934非活動且非failed、error=nil。16byte觀察窗口另含後續原bytes，核對只取此六byte指令。

90M完整16筆原表／pointer298848／stride55／bias0可讀，880bytes SHA-256 eca22108024dbc103f8de8257b32321b0b2e39fcd6bb8435cb4dc0ba73e29396，與唯讀336舊表不同。EIP228E0B／R=[77 116 FA 95 2BD9F0 2BDA4C 2C0C4A 2C04BC]／段=[8 188 188 0 20 188]／flags287h，FPU127F／status0／depth0／八stack0，callback6／6／mask2Bh／pending0／非活動，IRQ20869／20869非活動且非failed；原表及PNG擷取自身整RAM／CPU／FPU／VBE保持。虛擬時間162623172µs。RGB 9d8c0a1acb3b96200296f789bb13c6877067f6165ec608a7e019506036832dfc、PNG 7aec4ca6aad1f948560e184695bd3b415b1145ad9778e536da14a60e11d61861與336舊原選族頁相同。2026-10-04親看原SELECT RACE畫面，十四種族／Custom按鈕完整；游標在Custom附近不證明已選Custom。此處明示diagnostic stop，未送race，不稱180M完成。

原418來源檔guest前後逐檔SHA-256保持；state與363清單完全相同，只有sound.lbx4250888bytes／SHA-256 3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d／UID及GID1000，沒有SAVE10.GAM。本輪未達原存檔writer與內容，存檔成功、完整開局／RNG／音訊／remake同狀態仍未知。

### 命令與結果

Docker沿Go1.24.13映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，來源唯讀，工具及state在容器/tmp。三私有區塊逆轉為362原試作通過，公開CPU／DOS／provider／probe及所有internal未變。

```text
python3 workplace/new-game-364-ready-review.py
bash workplace/new-game-364-run.sh
python3 workplace/new-game-364-verify.py
python3 apps/moo2/tools/startup_probe_131.py --check-overlay-setup-source-spec-backlinks
```

全部通過。80M前6270共通列扣363早期診斷且不含precondition自身，依352既有mtime／DTA／每輪診斷RAMhash正規化保持，27PNG逐byte保持；原precondition各欄保持、原唯讀guard false，新profile依觀察表與五窗口 true。四CLI無效值／缺依賴在讀EXE前拒絕，合法值越過閘門後缺EXE明確失敗。產生腳本括號筆誤在原guest前修正，原guest一次、未重跑挑結果。公開規格只限定隔離ACCEPT／選族頁，完整可寫玩家路徑仍未驗。

### 私有收據

位於 `workplace/dosgolem/workplace/`，11份SHA-256與UID及GID1000已核對。原版LOG／PNG／資產不提交，公開只保存自製規格與雜湊。

| 本機檔名 | SHA-256 |
|---|---|
| new-game-364-ready-review.py | 2a2d04a7fb38ab94aaf7162d82693d0681f3c933be2e295ecc62fa51da152e79 |
| new-game-364-ready-review-tests.txt | 183e0384896f54b31b34fbcc6d6439b2abbc5ca901eb547ba606ea18751f46a7 |
| moo2-overlay-accept-364.go | 567ba02b76e90abc00e9bb921d73924a7a4ea99563b386223209006695b16e28 |
| new-game-364-run.sh | ddcd3450bc97e3500e612a82597a25b7784b5277bca104e7eabcad5b0eedf9db |
| new-game-364-run-output.txt | 9d3097bc1c26b3c44ce7a9620be44167be03519a9e7c0c69efb56c1d28960f50 |
| new-game-364-cli-tests.txt | a24e047c8bfbc281dd85f7c9c2cb609f5b82805b149f1d3474a87043556c0645 |
| moo2-probe-364-overlay.txt.gz | cba0542dce180527e2a6a258e7977b9c2b5b153231e1d7f89bb8ac17ce93fc8d |
| moo2-save-state-364.json | e35322cba18956bcd245e4091932c822f9460271deda20ddda8b7722aa1375bc |
| new-game-364-verify.py | d7dbfe076b5aed280b221f1fc0e9d440cbe89c074afede410ec372711bb8cb49 |
| new-game-364-tests.txt | 8a3ed1fdf3f2edb42ca7cb852d54f4c83fc9b1e912bd7a550968bcc74a6626a2 |
| moo2-364-overlay-frame-extended-90000000.png | 7aec4ca6aad1f948560e184695bd3b415b1145ad9778e536da14a60e11d61861 |

### 交接

工具98950c28961c0f89ed63304cd131477d083c85b7已推送github隔離分支，主庫玩法RE-first保持。原版執行容器已自動移除；兩個專案掛載filter均無執行中與停止容器。下一步以本次90M真實880byte表／完整初態與既有337幾何，另立可寫Humans正常輸入規格，原唯讀race guard保持；正常store與名稱頁未驗前不外推成功。

## 2026-10-04：365正常Humans與366名稱readiness

主庫起始aa44de78f60bfbcf4ac9dc61e35a3e37305e5301，工具成果[365正常選族](https://github.com/wicanr2/dosgolem/blob/94b15847606ff4c90635ba90d2c38388a86a2535/docs/spec/365-moo2-overlay-race-humans.md)與[366只讀readiness](https://github.com/wicanr2/dosgolem/blob/94b15847606ff4c90635ba90d2c38388a86a2535/docs/spec/366-moo2-name-ready-boundary.md)。原官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；ZIP417根檔另加官方EXE共418檔，MOX.SET／1996-01-01／180M參數與原先正常輸入保持，日期不是seed。原位址均dosgolem_high_le。

### 365正常選族，限定已證實

原90M完整880bytes表與唯一index7幾何351,330–473,374、完整CPU／FPU／RGB／callback6／6／IRQ20869／20869，直接用364原收據審查後READY。只送一次正常press x824／y352／buttons1於90000000／162623172µs；release x826／y352／buttons0於90008107／162662935µs，差39763µs，回呼7／7完成後首次可送。原target8:2136D1已由實際輸入收據核對，不從舊80M推定。

原90066074／20DDDB六byte66A3A6C42600，DS188:26C4A6 word0000→0700，下一20DDE1，R=[7 50A004 181 298848 2BDA2C 2BDA94 D 1A]／段=[8 188 188 0 20 188]／flags297h保持；callback8／8與IRQ20889／20889非活動且非failed，error=nil。16byte觀察窗口只取六byte指令作核對，其餘原bytes仍留私有收據，不把窗口當一條指令。

95M名稱表3筆／pointer298848／stride55／完整165bytes SHA-256 db67c471ba6a34a66146e083662f4c134c192e2371a807ec9647327558a7fd1b，與唯讀338表不同。index2+24原28439D的32byte候選為Strader補零，可讀且CPU／FPU／VBE／RAM保持；只證文字編輯候選，不稱正式持久名稱writer。2026-10-04親看原640×480 Enter Ruler Name、Strader與ACCEPT，RGB2646a7ef25939869ede60aa35bcd97d649b906fd2cc287d08deb6c2d8058cc12／PNG e1c739f5aeaf47a6cfdca4749b14509ec2592b65974fa6669e9347c491a52a36與338原95M圖相同。原95000000 EIP22F263／R=[20 2BD8C8 2A38E0 2A38E0 2BD8C8 2BD8E4 33 28439D]／段=[8 188 188 0 20 188]／flags12h，IF關閉、callback mask1／8／8、IRQ22338／22338非活動非failed。這與舊338的95M前置不同，不能只換表hash送輸入。

365前90M6956共通原364列依352既有mtime／DTA／每輪只讀診斷RAMhash正規化保持，28PNG逐byte保持；原337唯讀race guard仍false且未改，獨立profile核對真表／CPU／FPU／IRQ才true。三區塊與兩診斷條件逆轉為364，公開CPU／DOS／provider／probe及全部internal保持；四CLI拒絕與合法值正對照通過。原guest一次，未送名稱確認。

| 本機檔名 | SHA-256 |
|---|---|
| new-game-365-ready-review.py | 6a8d76b95307a3e4af245eda3b36bcb11838ed7fc9cab6d6dd5d419c07242b4f |
| new-game-365-ready-review-tests.txt | b72821308ab7330bf5032ce58805e624a5e4b3141ad2687ad908dc002b9302d8 |
| moo2-overlay-humans-365.go | 78ebe0f2180bb4ebcc8895543e3f4b3607927a5ee55572de0ca5d8a4d0e0b0a2 |
| new-game-365-patches.json | 1f9d35e403088e452da10a355c7cfab1e6edf54973bdfdc7e2cf5026481b1b84 |
| new-game-365-run.sh | f1dc742bd448941a5830cb25c886a63778469c2622e7b19dff6a4d8fb1afce09 |
| new-game-365-run-output.txt | 06632b1e5beb8577072d0a803553912f6d38ad529cb63f3af689afd883a6614a |
| new-game-365-cli-tests.txt | 7e323a6a2e8c352aeb94f038f92b5504ee495ca4b344a639b7549773dbe04a1e |
| moo2-probe-365-overlay.txt.gz | 89f17f037e6b8d840ba082e11bbb7da96992c9ded103b09b9c7be43990e83133 |
| moo2-save-state-365.json | e35322cba18956bcd245e4091932c822f9460271deda20ddda8b7722aa1375bc |
| new-game-365-verify.py | 93de3d5d145c9188bf1c13cddcb75b4bdf3de9279fe7110f84330c37d14b6ac1 |
| new-game-365-tests.txt | 5afa1124b9e65f7b04b9cbd03388f96d682ddd6f44332c5c89e392ab7fa31a0b |
| moo2-365-overlay-frame-extended-95000000.png | e1c739f5aeaf47a6cfdca4749b14509ec2592b65974fa6669e9347c491a52a36 |

### 366自然恢復平台readiness，限定已證實

從95M起最多4096原指令、每步重算既有平台安全條件：IF開、target8:2136D1／mask1或2B、pending0／callback非活動、IRQ非活動且非failed。沒有輸入或修改IF，不skip CPU.Step。前64筆95000000–95000063逐步邊界均ready=false；offset64超出64筆記錄未打印，不偽造該列。受審查的first-stop每步判定、首次true即return，實際95000065停止ready=true；來源四替換可逆及native停止收據核對。未要求逐行翻譯內部helper。

原95000065 EIP234A49／R=[3 1 210160 2BD8C8 2BD8DC 2BD8E4 2843A5 28439D]／段=[8 188 188 0 20 188]／flags216h，IF已自然開啟；FPU127F／status0／depth0／八stack bits0，原globals／header／165bytes表／32bytes候選與RGB保持95M，VBE Active／Bank4／StartY0／DisplaySets48保持。原readiness快照CPU／FPU／VBE／整RAM自身前後保持。同guest只讀defer終態另證mask1／pending0／callback非活動／8／8，沒有名稱press／release。終點精確IRQ計數未直接列出，非活動／非failed由ready判定與受審查原工具程式核對，不把95M的22338／22338寫成終點直接收據。

366前95M7662原365共通列依既有正規化保持，29PNG逐byte保持，原95M前置／候選亦保持。首次CLI驗證因ready含read子字串而誤判原正常拒絕，原exit2正確、guest未啟動；確認raw檔尚不存在後修正同腳本，原guest只首次執行一次。四CLI無效值／缺依賴在讀EXE前拒絕，合法值越過閘門後缺EXE明確失敗。核對64筆邊界與defer終態時直接讀同收據，未重跑原版。

既有平台契約internal/machine/le_mouse_callback.go:123，位置變化flags1／按下2／放開4，flags&mask非零才排回呼、裝置按鍵與位置仍更新；dispatcher遇IF關閉不派發。mask1下新位置與按下可產生flags3並排回呼，這只證平台契約，原名稱確認成功尚未驗。原338唯讀guard保持，不單純換hash放行原95M。

| 本機檔名 | SHA-256 |
|---|---|
| moo2-name-ready-366.go | 2988ba33477cc973b9daba38a94804a21903bd9a6318108024376faba65b712c |
| new-game-366-patches.json | 40501475fbfda45c7bac976fcdc2e6e89c1e5dfc981f7fd2476716b8c25157c9 |
| new-game-366-run.sh | 39131b221de304b0ca8154147e7ef262dfe39734a86706582d5be00e163fadeb |
| new-game-366-cli-first.txt | 28f43f09125af7f46276da789bfa4027a4e96b505db656dd4accb08cb28461fa |
| new-game-366-run-output.txt | bd590e0ba7de887112e246b273fbdb4f258b36d8d994d031efd8ef87e2dbe415 |
| new-game-366-cli-tests.txt | 552eda2f737327ec1d4ff956837f2291493c96a0955b80fe7b91a901215536aa |
| moo2-probe-366-overlay.txt.gz | 6e1eff8cb6b876047130faf0388c575a38096c58b2304c9315c7968212e9707d |
| moo2-save-state-366.json | e35322cba18956bcd245e4091932c822f9460271deda20ddda8b7722aa1375bc |
| new-game-366-verify.py | 0f1e17cf3e635fd49bd9659f46cac07fcdcb2c0ae7129eade7db32ea62d9b38c |
| new-game-366-tests.txt | db28d2402badc34544eddd3f3b11c6c6783af0d5fdd27f74faab19875f313ce0 |
| moo2-366-overlay-frame-extended-95000000.png | e1c739f5aeaf47a6cfdca4749b14509ec2592b65974fa6669e9347c491a52a36 |

### 命令、來源與交接

上述23份檔案均在忽略 `workplace/dosgolem/workplace/`，SHA-256與UID／GID1000已核對。原PNG／LOG／資產／state不加入Git。Go1.24.13 Docker映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀，隔離工具與state在容器/tmp。

```text
python3 workplace/new-game-365-ready-review.py
bash workplace/new-game-365-run.sh
python3 workplace/new-game-365-verify.py
bash workplace/new-game-366-run.sh
python3 workplace/new-game-366-verify.py
```

均於Docker通過。每次原418來源檔guest前後SHA-256保持；state與364相同，僅sound.lbx4250888bytes／SHA-256 3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d／UID與GID1000，原ZIP同bytes。沒有SAVE10.GAM，未達正式存檔成功、typed種族／持久名稱／旗色、完整開局、RNG、人耳或remake同狀態。

工具bdac0e0與94b15847606ff4c90635ba90d2c38388a86a2535已推送github隔離分支。原版工作容器已自動移除，兩個專案掛載filter均無執行中或停止容器。下一步依原95000065真實初態另立名稱確認DRAFT，審查按下mask1與一次正常press／release，仍實讀IF／target／callback／IRQ安全條件；不以任意時刻試到成功，不深挖helper。主庫玩法RE-first保持。

## 2026-10-04：367可寫Strader正常ACCEPT與99M旗色頁

**已證實，限定正常名稱確認**：官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，原ZIP417根檔另加官方EXE共418檔，MOX.SET／1996-01-01／180M參數保持，1996不是seed。位址空間dosgolem_high_le。工具起始94b15847606ff4c90635ba90d2c38388a86a2535，公開CPU／DOS／provider／probe與全部internal保持，362可寫overlay／365Humans與367私有profile僅在容器/tmp組合。

READY前直接審查366同guest首個95000065自然ready收據，完整CPU／FPU／165byte表／globals／header／32byte Strader／RGB與唯一ACCEPT矩形保持。95M flags12h／IF關閉不送輸入，65原指令自然恢復；獨立profile在95000065原EIP234A49／flags216h、callback mask1／8／8、IRQ22338／22338非活動且非failed通過，原338唯讀guard false保持。一次press x640／239／buttons1在175812740µs，release95013578／175858124µs／x642／239／buttons0，相差45384µs，放開前callback9／9正常完成。未改IF／RAM／選擇／EIP。

原95008897、20DDDB六byte66A3A6C42600真正寫DS188:26C4A6 word0000→0100，下一EIP20DDE1，R／六段／flags297h與候選Strader保持，callback9／9、IRQ22340／22340非活動且非failed。共享GUI選擇欄位已證實，不外推正式名稱writer。原99M SELECT BANNER COLOR已親看，RGB8c2c573c08ebf4bce0b4e166d4268fb6fcf90f35dd27fbbb2de614b26f187f15，實際count10／stride55全550bytes可讀，CPU／FPU與RAM只讀。EIP23857C、R=[347B20 0 4B 1 2BD9B8 2BDA14 6AE528 347B6C]、六段=[8 188 188 0 20 188]、flags206h，FPU127F／status0／depth0／八stack bits0，callback mask2B／10／10、IRQ23503／23503非活動且非failed。原339唯讀旗色guard false保持，99M診斷停止，BANNER_RED輸入未送。

**環境／驗證腳本勘誤**：第一次生成器匹配早期defer並把旗色診斷插到50M menu分支，尚未送名稱輸入即diagnostic return。該次Go／patches／run／CLI／原LOG／state／24PNG共32檔原樣保留，逐份SHA-256與UID／GID1000核對，以new-game-367-first-receipts.json索引。改用唯一執行期條件與99M／550bytes範圍斷言，同Docker與同參數乾淨重跑；名稱輸入只在修正後guest送一次，不重擲結果。首次不包裝為99M或正常玩家完成。驗證初次把舊365在95M停止之後一筆原INT33h列入共同前綴，7662／7663長度不同、共通列相同；按共同95M候選快照終點修正，guest未再跑。原輸入收據和結果沒有變。

**驗證**：365前95M的7662共通列按352既有mtime／DTA／每輪只讀RAMhash正規化，29PNG逐byte保持；366首個readiness CPU／完整表／候選／RGB保持。四CLI拒絕與合法值越過前置後的缺EXE正對照通過。九私有替換逆轉為365，公開internal與probe保持。原418檔guest前後SHA-256保持，state同366僅sound.lbx4250888bytes／SHA-256 3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d／UID與GID1000，原ZIP同bytes，沒有SAVE10.GAM。

### 本機私有收據SHA-256

| 文件 | SHA-256 |
|---|---|
| moo2-367-overlay-frame-extended-99000000.png | 96efbd1ce6538c27b019cc6713fe7397d6d0fde7a63c7e01cb82023f75614c67 |
| moo2-probe-367-overlay.txt.gz | 4e241f99d868c28cde3107d884b68136769d8dd9bb158021edbc2f9547d34725 |
| moo2-ruler-accept-367.go | 0f49ac8cdfb3a1850b5d32f6de86b5dd216ef4cb144e3416aa4c581bef700c38 |
| moo2-save-state-367.json | e35322cba18956bcd245e4091932c822f9460271deda20ddda8b7722aa1375bc |
| new-game-367-cli-tests.txt | 552eda2f737327ec1d4ff956837f2291493c96a0955b80fe7b91a901215536aa |
| new-game-367-data-hashes.json | 0510e9168f7137fd1a1d5551ee7b4e84122e9c4321d2db58085aa448b5e7c40c |
| new-game-367-first-receipts.json | 98710ea3cc2bf068c48bd15ab1d0ad5344060aaba07874733421418506784de2 |
| new-game-367-patches.json | 1b434c2da34079037b9240916e7135107408eec0cb28f101225c6da11be47388 |
| new-game-367-ready-review-tests.txt | c122807494b5e12fa430d945289c80c50168479dd7ef8e219a0239dd5c3e1efa |
| new-game-367-ready-review.py | ab6b08790ef6e8c925984ad397226216007e890b8186e8ed9dedb88227e7bf69 |
| new-game-367-run-output.txt | b6055c293bf323cd1f2752684b096f615843cbc9f24fea742fa269d45242dbe8 |
| new-game-367-run.sh | 746424ff56b9d4a888278b45d82be8781e2113f92b4dcff6becb090f28e43d87 |
| new-game-367-tests.txt | 458c0c10022164a9e7ff4225331bd8884cf5ec5fce0b44c11dcfd7accda7a60f |
| new-game-367-verify.py | 23cbecb0c4572aa304bfc54a55c1d1b60233fd813ece9e93886eb31af7a13969 |

上述14份在忽略workplace/dosgolem/workplace/，SHA-256與UID／GID1000核對，首次32檔由其獨立索引保存。原PNG／LOG／RAM／版權素材不入Git。工具規格367、索引與338／365／366回填已提交a649d0b9d035eec8d8f57235b7e54a30cd7848bd並推送github隔離分支。

### 命令與下一步

```text
python3 workplace/new-game-367-ready-review.py
bash workplace/new-game-367-run.sh
python3 workplace/new-game-367-verify.py
```

均於既有Go1.24.13 Docker執行，image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP與patch只讀。原版容器已自動移除，兩個專案掛載filter無執行中或停止容器；既有root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。

原資料→完整表／候選→正常press／release→共享store→旗色UI限定接通。未知：typed名稱、正式名稱持久writer／旗色選擇／正式存檔／完整開局／RNG／音訊與remake同狀態。180M參數保持但99M診斷停止不稱180M完成；主庫RE-first保持。下一步以99M完整550byte表／globals／header／CPU／FPU／RGB／callback與IRQ，另立紅旗正常press／原INT33h poll／GUI selection／release契約並審查，不只換hash或挑時刻放寬339唯讀guard。

## 2026-10-04：368正常旗色輸入與原成功寫檔，169E49的新CPU拒絕

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le；原ZIP417根檔加官方EXE共418檔，MOX.SET／1996-01-01／180M與先前正常輸入保持。日期不是seed。工具起始a649d0b9d035eec8d8f57235b7e54a30cd7848bd，公開internal／CPU／DOS／provider／probe保持。新版368為隔離輸入試作，原339唯讀guard不改。

### 已證實的正常流程與範圍

沿367真實99M完整550bytes／globals／header／CPU／FPU／VBE／RGB／clock／callback／IRQ審查後READY。一次press99000000／185561342µs，99103163原24C31B的INT33 AX3返回BX1／CX276／DX190；99103343與99103698原214104 RET20DB5B／20E165返回AX1。原SS188:ESP2BD998 stack top5BDB2000／65E12000，完整R只ESP+4、六段與flags202h保持，原observer readable／readonly／valid true、error nil。99103699／185890422µs首次合法mask1 release，持按329080µs、callback11／11完成。沒有代寫CPU／選擇／RAM、沒有重送或重跑guest。

原160M Placing home worlds已親看。165058686、原237024的SAVE10.GAM 3D01返回handle9／CF0，165058731原237093的40／CX0 truncate；13筆40含truncate，238310各筆完整EAX等於請求ECX，合計208000bytes，原3E正常close。另兩次MOX.SET開檔／truncate／40寫553bytes／正常close。99筆DOS診斷全部handled且CF0。這已解出可寫原寫檔成功的玩家阻塞，未證正式讀檔或存檔內容語意。

state終態SAVE10.GAM208000bytes／SHA-256 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f，sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d與367及原ZIP相同。三檔UID／GID1000，原418來源前後SHA-256保持。原guest存活時在同bounded容器內Docker exec只讀state，另存變動檔副本；原guest退出後最後副本逐份與run終態清單hash／bytes一致，中途0bytes／部分寫入不當終態。監測400s有界，精確命令等價保存私有new-game-368-state-capture.py，未改guest或state檔。

### 新能力缺口與未知

原165113094在169E49拒絕bytes20 D0 59 C3 53 51 89 C1；EAX18900／ECX0／EDX601／EBX0F／ESP2BDB00／EBP2BDB34／ESI1／EDIFFFFFFEC，段=[8 188 188 0 20 188]、flags216h。ModRM D0對應AND AL,DL，這個正常consumer結果未知。probe exit0但明確guest_cpu_stop／step_error存在，沒有step_limit或dos_exit；未達180M。終圖黑底游標已親看，原旗色共享20DDDB store未命中，正式旗色持久writer未知。

### 驗證與回填

367前99M的7846原共通列依352既有mtime／DTA／每輪只讀RAMhash正規化、30PNG逐byte保持；全550byte表、完整CPU／FPU／VBE／RGB／callback／IRQ與clock保持。三私有變更逆轉為367，公開internal與probe保持；四CLI拒絕與合法缺EXE正對照通過。初次READY腳本誤拼診斷名稱，guest前依原source修正後重讀同367收據通過，原guest一次。339／341／362／367已附不可變原定位與限定回填，舊唯讀收據保持，不把可寫流程當舊same-state。

| 本機忽略來源／收據 | SHA-256 |
|---|---|
| moo2-368-overlay-frame-extended-160000000.png | c7534b8f40b51b8377d255d66e6dd759dfb3d427fa6fccc6ee7d8c3999b32bca |
| moo2-banner-red-368.go | ac013d424348917b2178ccf44d425669731d19ca314ad71a249f1e47023402f4 |
| moo2-probe-368-overlay.txt.gz | 8cf7b6f1e75b7759e3e40c8a7574cbf85150a7269bd1e99fdb3fae4445d882cd |
| moo2-save-state-368-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-368-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-368.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-368-overlay.png | 1f757f5b16fe492795accf198c8db851443a062450e7f37ecf465c7eae3b6457 |
| new-game-368-cli-tests.txt | f7bc407d133dd1cdb56fcb8aa6377e508459429adc0ec583216ba2c8088e545d |
| new-game-368-data-hashes.json | 46add69fbddc4b2805b22bd6f590c8581b37aa7581d7f4ccafce02173e097b82 |
| new-game-368-patches.json | 005d803b8855a66259d60a1d767a5bb0575858c076ed59088412a4c4b84d200b |
| new-game-368-ready-review-tests.txt | c454988ebdecfc38d1c52b5b659724f672e917f7a810e43c87d994d876662430 |
| new-game-368-ready-review.py | 487fd42694ac2324aba6c8053ecd6818e246076f15af8c8c0545c2cb0ce2ba97 |
| new-game-368-run-output.txt | 4c4d2e3ed0fdbb3eeb6dd3aeb061b537367214a733afe97fc6360ab6484c99f9 |
| new-game-368-run.sh | bb101a3bf3c22781071b0c0567bd4cb7e0d5dc7f65ed1f703e90c984e76df44c |
| new-game-368-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-368-state-capture.py | 5754b764b86bc93587d84761dd3c9367c4a625e5acb3355e93d73d8d2e53a8e9 |
| new-game-368-tests.txt | cc818c2a534656e0c69d2bd9fa772a31ced7d68a7f44882b7d8614d7ebaaedf3 |
| new-game-368-verify.py | ac3864d78b6d36748cc371f8d4bddde575503a2665f6e13792afe236c1bbce52 |

18份均在workplace/dosgolem/workplace/，雜湊及UID／GID1000核對。原LOG／PNG／RAM／state bytes與版權素材不提交。公開只交368規格／索引／四舊規格回填，工具62cd4911727f17042cb5f8ce98e10b0fe80331ac已推送github隔離分支。

### 實際命令與交接

```text
python3 workplace/new-game-368-ready-review.py
bash workplace/new-game-368-run.sh
python3 workplace/new-game-368-verify.py
```

Go1.24.13 Docker映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。原版與同步監測均terminal，容器自動移除，兩專案掛載filter無執行中／停止容器；既有root-owned2437檔／272目錄保持，無新增root-owned或.md目錄。

下一步依原169E49／20D0另立byte AND規格，先READY再補CPU register-source能力、原AND／下一POP／RET與原前綴保持；同正常輸入續行，不增點擊或改cap。主庫玩法RE-first保持。正式讀檔／存檔內容／typed旗色／完整開局／RNG／音訊與remake同狀態仍未知。

## 2026-10-04：369原AND／POP／RET與180M母星命名終態

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。工具起始62cd4911727f17042cb5f8ce98e10b0fe80331ac，CPU b8c1844163fddd7e3557e19fc51abcb9bcb9c1d021fd415dae376b2afbb9c72b；READY後只增加裸20 /r分支與自製測試，CPU變為a39e5b9f74e026fc1c2514d733808e94920b6789d1b68ef8eb8960dc72df5350，窄測331d33f4a0f22ff524fa6d37d0fba73c7070729333c73d45bfc46c98019cca38。來源418檔、1996日期、可寫state、180M與368所有正常輸入保持，日期不是seed。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none、原ZIP及patch唯讀。

### 已證實與工程模型

原165113094在169E49執行20D0。R=[18900 0 601 F 2BDB00 2BDB34 1 FFFFFFEC]、段=[8 188 188 0 20 188]、flags216h；AND AL00,DL01=00，完整R不變、EIP169E4B、flags246h。CF／OF／SF0、ZF／PF1與[Intel80386 AND](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/AND.htm)一致；[附錄C](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/appc.htm)的AF未定義，清0只沿工具模型，不稱硬體逐值一致。

原165113095在169E4B執行59 POP ECX，真實SS188:ESP2BDB00給ECX0F，ESP2BDB04／EIP169E4C。下一165113096 C3依真實stack返回14DC1E、ESP2BDB08；其餘R／段／flags246h保持。三步完整FPU127F／status0／depth0／八stack bits0、RAM保持，observer readonly／step_ram_unchanged true、error nil，callback12／12及IRQ43071／43071完成且inactive。工程測試合成stack與此原native證據分開。

CPU窄測1.011s通過，memory八來源×256×256×兩flags初態1048576組，全部register重疊／ModRM／SIB／相異DS SS／wrap／段末／截短／未知段／唯讀／bus拒絕與成功發布flag契約。固定官方EXE乾淨Go全套通過，CPU38660.343s／machine1.861s；未導入8088外部資料，不外推驗收。剝除新分支回到原CPU，三私有observer逆轉為368，其他公開internal及probe保持。

### 原玩家終點與未知

原guest一次實際step_limit180000000／EIP235AA3／unique_sites53798，沒有guest_cpu_stop、step_error、dos_exit。終圖已親看：星圖背後、Enter Home Star Name視窗、Sol候選與ACCEPT，名稱尚未確認。PNG SHA-256 4ef5ef10d460497a6042c3df9f24149ceea8b9ddafd2d1d76580ef64f9094686。

180M虛擬418789381µs，R=[0 18 0 0 2BD3A8 2BD3D8 3CA527 34CFD9]、段=[8 188 188 0 20 188]、flags246h，FPU127F／status0／depth0／八stack bits0。VBE bank9／startY512／sets4023／writes52675030／display93；callback12／12、IRQ47499／47499完成且inactive，IF1。實際table298848／count3／stride55／165bytes SHA-256 d12f33061ee90916f4c35c9ee5177fc4e3760fb4ca791f572935e139c084d2ff、globals69a0c2f011924ac7d4f6960ddb98df646f158d49b48f0f43ed72c92cfa0956b9、header1eb307e414824282e81d17faf0ac5ab6752ae0f074ecd526f3cb4ce16847011c，只讀RAM前後009974bbd52333341aad90e02cb7fca96a902f8db174d5ad5df8931a937ac457保持。候選字串的記憶體與原ACCEPT返回契約待下一窄任務，不從圖像猜持久writer或重用旗色guard。

SAVE10.GAM208000／SHA-256 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d，MOX.SET553／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f，sound.lbx4250888／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d；三檔UID／GID1000。550s有界監測同guest只讀state，最後副本與終態及368一致；中途truncate與部分寫入不當終態。原418來源前後保持。正式讀檔／內容語意／名稱及旗色持久writer／完整開局／RNG與remake同狀態未驗，主庫玩法RE-first保持。

### 驗證窗口與回填

368拒絕前11017共通原列與37PNG保持，新39PNG；原mtime／DTA及每輪RAMhash沿既有正規化。save_dos_diagnostic保留DOS服務是否改RAM的關係，observer只讀不代表3F服務不寫RAM。首次驗證誤設RAM相同而失敗，修正該判準；第二次誤含舊CPU拒絕專屬四筆terminal列，修正窗口至late_startup_platform label=stop前。前11017列始終相同，兩次均為重讀同native收據，沒有重跑guest／調參挑結果。

四CLI拒絕及合法缺EXE正對照保持。369 CONFORMED只限CPU／原consumer與180M續行；339／341／362／367／368以官方EXE／dosgolem_high_le:169E49／20D0不可變鍵追加回填，歷史拒絕來源保留，不升格舊唯讀same-state。

| 本機忽略來源／收據 | SHA-256 |
|---|---|
| full-test-369.txt | 9b7928672f6dc4cf9f52698dc413b7fc572874a92df4cd869b2bf671762bfdfd |
| moo2-369-overlay-frame-extended-180000000.png | 4ef5ef10d460497a6042c3df9f24149ceea8b9ddafd2d1d76580ef64f9094686 |
| moo2-and-byte-369.go | a8a30a4b3fa124e0ffc9c2555907df8e607ce20d232f890326292983365aebbd |
| moo2-probe-369-overlay.txt.gz | 59254d0a66659a475cf28d20be1d32ec4fcd131c424cd1ea9745fa225446b631 |
| moo2-save-state-369-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-369-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-369.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-369-overlay.png | 4ef5ef10d460497a6042c3df9f24149ceea8b9ddafd2d1d76580ef64f9094686 |
| new-game-369-cli-tests.txt | f7bc407d133dd1cdb56fcb8aa6377e508459429adc0ec583216ba2c8088e545d |
| new-game-369-cpu-patch.json | 428a22c26cd5969208c1e602886b4599f8a7f2cedb53a58e6ed301bc0f1d0b04 |
| new-game-369-cpu-tests.txt | 6b7c6322229e4441a6c3a11a8573491f697732c498a420492eced9035aeafd1d |
| new-game-369-full-run-output.txt | 9b7928672f6dc4cf9f52698dc413b7fc572874a92df4cd869b2bf671762bfdfd |
| new-game-369-full-run.sh | 8764d3f3c8a115b9a824a7db31a6c93e80b3381a404b56a721c0e1e082e2ef4c |
| new-game-369-probe-patches.json | ce8336bea155a82761291246d368f28f6e2094b8d498f51902d0819eb5592ac2 |
| new-game-369-ready-review-tests.txt | 4b84db379c967b6365ebabed3b0f0d7ba3f861c9b35048baa22bdf7af94dbd87 |
| new-game-369-ready-review.py | 3180d509167ef4dda0d2720b5745a3b6cd205393d23012626f7c9b84c4ec1120 |
| new-game-369-run-output.txt | cb4550c129ac263a8190428c4a796c8c1b618186e585278940ea43f6bcf6c09b |
| new-game-369-run.sh | ed316b27c4184246c5e5b10c1e259d4bb31c53a00c668a05a99f0650c68e0c1f |
| new-game-369-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-369-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-369-state-capture.py | f9f53efa9ae68a44d5da7ce6daed219c7b3482a691b3b38cc5974a75cc4079fb |
| new-game-369-tests.txt | d54a66147244069afd12b6af35fd881d0def96124df6489c5f10ce68df0486da |
| new-game-369-verify.py | 02d4bf970503de32916908ba90b5a5bfdf142a6de1925c66fc6e30fb776b2a51 |

23份均在workplace/dosgolem/workplace/，雜湊與UID／GID1000核對。工具eca6a803aba5176b87f27c5defd1146ff121c226已推送github，公開自製CPU／測試／規格與索引／五份回填。原LOG／PNG／RAM／state留在本機忽略目錄。

下一步核對180M實際表、候選字串與原返回端，建立母星命名正常ACCEPT的READY契約後再輸入及受控release。原run／full-run／review／verify精確入口見工具369規格；不加cap、不代寫核心、不把星圖出現當整段開局完成。

## 2026-10-04：370母星候選來源與正常輸入前置

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。起始工具eca6a803aba5176b87f27c5defd1146ff121c226，CPU保持a39e5b9f74e026fc1c2514d733808e94920b6789d1b68ef8eb8960dc72df5350，公開internal／probe皆不變。沿369可寫正常輸入、180M、1996固定日期與原418來源，日期不是seed。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。

### 已證實的來源與前置

同guest170M及180M只讀原DS188:298848 count3／stride55的完整165byte表、index2+24→28439D的32byte候選均可讀；候選Sol加NUL補零，SHA-256 6919c6ec1f4751149ac2653bf5decc4881f64d5b997fb2cac85cf8f4c175f30d，兩時點保持。table SHA-256 d12f33061ee90916f4c35c9ee5177fc4e3760fb4ca791f572935e139c084d2ff、header1eb307e414824282e81d17faf0ac5ab6752ae0f074ecd526f3cb4ce16847011c保持。候選是輸入緩衝區，正式星名欄位／writer仍未知。

index1+24→261AC2原16byte窗口首byte00並含BUFFER0，SHA-256 2c9d54f6d98794161cf52a7d394dc84cca138e8c6793f82b56577d7e8c7c7699；不是直接ACCEPT字串，診斷label_*只作導覽，不給欄位已證實語意。index1矩形227,246–324,273對應畫面ACCEPT為強推論，275,260與276,260唯一落入該矩形，正常點擊consumer待實測。

170M原EIP235948／R=[0 18 0 0 2BD3C4 2BD3F8 2A206E 34F062]／段=[8 188 188 0 20 188]／flags293h／IF1、FPU127F／status0／depth0／八stack bits0，虛擬386324835µs。VBE bank9／startY512／sets2793／writes51487286／display93，target8:2136D1、mask2B／pending0／inactive／callback12／12、IRQ44492／44492非活動非failed。code16／SS:ESP stack16、完整globals／header／表／候選／RGB已取，核心／FPU／VBE／整RAM／callback IRQ前後不變。170M原Enter Home Star Name／Sol／ACCEPT親看，PNG410c764d6bc1e928af3100cae79431bc03e152d83330ebd94b29c6fa0523c862，RGB677ff4d0508dd2470b05ab122639815b8382cdc8f2146f83f0e821536c3ad16e。

180M原核心與終圖逐值／逐byte保持369，callback12／12與IRQ47499／47499已返回。170M globals SHA-256 3e2f27c3b1dfa465b3f915b3a6a14427ebd8eba8ecc18023d89513014896c3b1、180M69a0c2f011924ac7d4f6960ddb98df646f158d49b48f0f43ed72c92cfa0956b9，只有DS188:26C4C6由02→01，語意未知。兩時點各與369相同原時點保持，不要求跨時點自然狀態不變。

### 驗證與限定範圍

369全14282共通原列按既有mtime／DTA與每輪RAMhash正規化保持，39frames及final PNG逐byte保持；新兩筆snapshot均只讀且對接原時點完整前置。三私有patch逆轉為369，公開internal／probe逐byte保持eca6a80；四CLI拒絕及合法缺EXE正對照保持。原guest一次真step_limit180000000／EIP235AA3／unique_sites53798，沒有CPU停止／step_error／dos_exit，未送母星確認。

初版驗證在全共通列與畫面已保持後，因錯要求跨時點globals相同而失敗。依369／370同170M與180M收據的26C4C6差異，改為逐時點核對，僅重讀收據未重啟guest。原state副本與終態及369一致，SAVE10.GAM208000bytes、MOX.SET553bytes、sound.lbx4250888bytes，UID GID1000；原418來源前後保持。原guest及550s有界監測由同run擁有／trap清理，終止後無殘留容器。

370限定CONFORMED只讀來源；369已追加候選不可變鍵回填，私有驗證檢查原指標／byte形狀、文件及勘誤入口，缺項即失敗。主庫玩法RE-first保持；正常確認、正式名稱及旗色持久writer／正式讀檔／完整開局／seed／remake同狀態未驗。

| 本機忽略來源／收據 | SHA-256 |
|---|---|
| moo2-370-overlay-frame-extended-170000000.png | 410c764d6bc1e928af3100cae79431bc03e152d83330ebd94b29c6fa0523c862 |
| moo2-370-overlay-frame-extended-180000000.png | 4ef5ef10d460497a6042c3df9f24149ceea8b9ddafd2d1d76580ef64f9094686 |
| moo2-home-name-source-370.go | 913775ad75aaee70331a967d05c92662e465615be25582a2e42941b35c2ddb73 |
| moo2-probe-370-overlay.txt.gz | 0c650a8610291fb3d5c141fce339f8c24850a7dd0c70f0b173febb566116fed2 |
| moo2-save-state-370-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-370-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-370.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-370-overlay.png | 4ef5ef10d460497a6042c3df9f24149ceea8b9ddafd2d1d76580ef64f9094686 |
| new-game-370-cli-tests.txt | f7bc407d133dd1cdb56fcb8aa6377e508459429adc0ec583216ba2c8088e545d |
| new-game-370-patches.json | 69deb738e8fdfb1d1a44176d6bee5d9ed280779f4a8fe8c8adbbd5dff4bf08e2 |
| new-game-370-ready-review-tests.txt | dba2470062f5175805af15647c19ca4535eeeed20be825a7c6060700f74617b9 |
| new-game-370-ready-review.py | d407becfeae0b1b591ad1f1103982d933d4f337f1cdd8977c5f0966a96156ef6 |
| new-game-370-run-output.txt | d5796d3892d09d820efd19c1dbf7162345979e9752cf468ecfe44786d52cd6db |
| new-game-370-run.sh | 35aebfa0aa024c120e5dfa727cfd0e6bd9b1945d26ad6f58ab4eee3d97f43356 |
| new-game-370-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-370-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-370-state-capture.py | 5b96496ce14c52badedc493f3e3462b4b939f0d61f67a9a62acc3427a632f836 |
| new-game-370-tests.txt | a9101fd6c9fc7654a081d56d21caa861a696eb658fe6a0b754c6c7922dc0d99e |
| new-game-370-verify.py | e6252d25e2a1bcc3bb557ce7d3495b68add4ec976830a5edb14fbe297717d72f |

19份均在workplace/dosgolem/workplace/，SHA-256與UID GID1000已核對。工具46ae96f788d4242da161d833add26cdedd9f4c32已推送github，公開自製370規格／索引／369回填，原LOG／PNG／RAM／state留本機忽略目錄。精確run／review／verify入口見370規格。

下一步以170M完整前置、Sol候選與唯一index1熱區審查正常ACCEPT press x550,y260／受控release x552,y260的READY契約。保持180M與既有正常輸入，不套舊ruler／banner guard，不代寫核心／RAM；共享選取store或原返回是否命中依實測，不深挖與玩家阻塞無關的renderer／helper。

## 2026-10-04：371母星正常ACCEPT與1749C0新F6EC阻塞

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。起始工具46ae96f788d4242da161d833add26cdedd9f4c32，CPU保持a39e5b9f74e026fc1c2514d733808e94920b6789d1b68ef8eb8960dc72df5350，所有公開internal／DOS／provider／probe不變。沿370可寫正常輸入／180M／1996日期，日期不是seed。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。

### 已證實的玩家流程

原170M完整370前置通過，一次press170000000／386324835µs／x550,y260／buttons1；原170015047、24C31B正常INT33 AX3返回BX1／CX550／DX260、段／flags12h保持。170015082／386372270µs首次合法release x552,y260／buttons0，持按47435µs，mask1／IF1／pending0／inactive、callback13／13、IRQ44497／44497完成。沒有重送、代寫核心／RAM或修改候選。

原170024038在214104 C3依SS188:ESP2BD364真實stack返回20DB5B，完整R只ESP+4，AX0／flags246h／段／FPU保持。170024419原20DDDB六byte66A3A6C42600實寫DS188:26C4A6 word0000→0100，EIP20DDE1；R=[1 1 37 8 2BD368 2BD3D0 2843A1 28439D]及六段／flags297h／FPU、32byteSol候選保持，callback14／14與IRQ44498／44498非活動非failed。後兩RET 171940309→20DB5B、172067022→174742均按真實stack核算ESP+4，AX0／flags246h保持，不推定caller名稱。

正常命名視窗消失、星圖顯示Sol及3500.0已親看，終PNG SHA-256 c577873bc2e015667734ff050d4601a24ac2daac5e5b4b5a94e4650df2c1773b。原資料表／Sol候選→正常press／poll／release→共享選取store→星圖的垂直鏈已驗；正式星名持久writer、存讀及完整開局仍未知。

### 新拒絕與觀察界限

原174213914、1749C0 bytesF6 EC A2 06 1F 28 00拒絕F6 ModRM EC。after EIP1749C2，R=[FF01 1A5 2 8 2BD488 2BD4E0 171C80 2BD4E0]、段=[8 188 188 0 20 188]、flags246h，FPU127F／status0／depth0／八stack bits0。原虛擬397790869µs，VBE bank9／startY512／sets2837／writes52762744／display93；home terminal pressed／released／polled／store_seen true、三returns、callback14／14／pending0／inactive。CPU拒絕只fetch opcode與ModRM，指令結果尚未執行。

[Intel80386 IMUL](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IMUL.htm)定義F6 /5為有號AL×r/m8→AX，CF／OF依符號延伸條件、其餘算術flags未定義。EC為AH來源，ISA辨識已證實；原AL01／AHFF已取，原native結果仍待補工具後取得，不把ISA導出或綠測試稱原版結果。本輪CPU未改。probe exit0但有guest_cpu_stop／step_error，沒有step_limit或dos_exit，未達180M／完整開局。

終態header count23、stride55，既有new_game_menu_table_snapshot限制count≤16，因此table_readable=false、records空；這是觀察器拒絕擷取的界限，不是已證實表格不可讀、資產越界或產品缺陷。完整23物件未取，不放寬舊guard或改資料型別。

### 驗證與回填

370點擊前11435共通原列按既有mtime／DTA及每輪只讀RAMhash正規化保持，38PNG逐byte保持；完整170M表／候選／核心／FPU／VBE／clock／callback IRQ精確對接。六私有patch逆轉為370，公開internal／CPU／DOS／probe保持，六CLI拒絕與合法模式開／關缺EXE兩正對照通過。原guest一次，沒有驗證腳本失敗或重啟選結果。

原418來源前後保持，state副本／終態與370相同，SAVE10.GAM208000bytes、MOX.SET553bytes、sound.lbx4250888bytes，UID／GID1000，實際hash見下表。原guest與550s有界monitor同run擁有與trap清理，terminal後無殘留容器。沒有新CPU碼，沿369固定EXE全套收據，另建置本正常輸入probe。

371限定CONFORMED正常母星確認、共享store與星圖／新拒絕；367／369／370追加不可變候選及原20DDDB／bytes／DS:26C4A6回填，舊統治者與旗色上下文保持，不當same-state或正式持久名稱writer解答。私有驗證核對三回填入口與原位址／bytes，缺項即失敗。

| 本機忽略來源／收據 | SHA-256 |
|---|---|
| moo2-371-overlay-frame-extended-170000000.png | 410c764d6bc1e928af3100cae79431bc03e152d83330ebd94b29c6fa0523c862 |
| moo2-home-name-accept-371.go | 002b8406497c5c6d131d3ca059828ca918bdb829587f340f605a8c7af1c3b7d7 |
| moo2-probe-371-overlay.txt.gz | 20a23fdb18adaef0b149d490bdf7c2224fe8890c5f131dea402e4a163f80a689 |
| moo2-save-state-371-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-371-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-371.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-371-overlay.png | c577873bc2e015667734ff050d4601a24ac2daac5e5b4b5a94e4650df2c1773b |
| new-game-371-cli-tests.txt | e23e63ea4e37e0f0823fda228ce807b3cdac82cc8d2c63c2bf2d1afb1a2312af |
| new-game-371-patches.json | d8fdaa346b6aeb3d6c7bad4d9b8df847e0e8c07a23126028dd3b28ceebbfa961 |
| new-game-371-ready-review-tests.txt | 375fef9620bba6b29cec9514ddec2847ab07dbffefae55b9ea1429ad7264dfb7 |
| new-game-371-ready-review.py | 3441389db8b415559a84e148f5445abb358e7286652f82ef017ddc7c60df43fc |
| new-game-371-run-output.txt | 8209e34ebabed029b9801ca26220d1506a58144f918015d216f056f6055ae557 |
| new-game-371-run.sh | 8ee629a0fc4b985243b4a20f9f9c70ac72d06bef141314c50b9f34b786bbb0a7 |
| new-game-371-state-capture-output.txt | 4e9e700c2ba93e7b5419aa50d201f6f01e55e4e39ed41951f50610daa299cd4e |
| new-game-371-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-371-state-capture.py | 5ce23c2b2e364a9ef14f4c410806c0354d2d3b918a8483e4a79d6863d3096d75 |
| new-game-371-tests.txt | cd0a380b027e82e49993dd7cc32eb50b5eb87e201f6b0c2d05bf46c058e0ca0c |
| new-game-371-verify.py | c5c6c500c255d7ffb29143a048899d05692cb05fce406d1ccbc70366581de4fe |

18份均在workplace/dosgolem/workplace/，SHA-256與UID／GID1000核對。工具c39af543efa47387b1fd96f86038d08f940f42c6已推送github，公開自製371規格／索引與三回填；原LOG／PNG／RAM／state留本機忽略目錄。精確READY審查／run／verify入口見工具371規格。

下一步依原174213914／1749C0 F6EC與完整核心、Intel ISA建立byte IMUL READY規格，補獨立CPU窄測與原consumer，保持180M及所有既有輸入再續行。主庫玩法RE-first保持，正式名稱與旗色持久語意／讀檔／完整開局／seed與remake同狀態未驗。

## 2026-10-04：372 byte IMUL原consumer與母星確認後180M星圖

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。起始工具c39af543efa47387b1fd96f86038d08f940f42c6，交付工具e57e9e063b1713b087423a78bef1349237c3d4b0已推送github；[372限定規格](https://github.com/wicanr2/dosgolem/blob/e57e9e063b1713b087423a78bef1349237c3d4b0/docs/spec/372-cpu386-imul-byte-source.md)與五份回填保留歷史正文。公開自製CPU／窄測／NEG拒絕判準／規格／索引，原資料、LOG／PNG／RAM／state及probe不入Git。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；600s／2GiB／2CPU／128pids／UID1000／network none，原ZIP／patch唯讀。

### 已證實原正常consumer

- 原174213914、1749C0 F6EC，R=[FF01 1A5 2 8 2BD488 2BD4E0 171C80 2BD4E0]、段=[8 188 188 0 20 188]、flags246h，FPU127F／status0／depth0／八stack bits0。原AL01乘AHFF得到AXFFFF，EIP1749C2；只有EAX低16改變，CF／OF0。
- 原174213915、1749C2 A2 06 1F 28 00，DS188:281F06真正舊byte01→FF，EIP1749C7；全RAM比較只有此一byte改變，完整核心／FPU保持。
- 原174213916、1749C7 E9 33 F3 FF FF，按原signed disp獨立核算返回173CFF，完整核心／RAM保持。三observer只讀、error nil，callback14／14與IRQ45735／45735完成且inactive。
- [Intel80386 IMUL](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/IMUL.htm)及[旗標附錄](https://www.ardent-tool.com/CPU/docs/Intel/386/manuals/prref386/appc.htm)給CF／OF與未定義邊界。SF／ZF／AF／PF保留是工具模型，不能稱原硬體逐值一致。目標byte用途仍未知；工程fixture舊55不冒充本原native舊01。

### 實際180M與完整表

相同371輸入／日期／180M cap，原guest一次。實際step_limit=180000000／EIP2176C5／unique_sites54239，無guest_cpu_stop／step_error／dos_exit。R=[35017C 70 39C17C 70 2BD478 2BD4C0 39C17C 35017C]、六段同上、flags202h、FPU同上，虛擬414027099µs，IF1；callback14／14與IRQ47425／47425已返回。VBE bank9／startY512／sets2927／writes55120744／display93。

終圖已親看正常星圖、Sol／3500.0、底部COLONIES／PLANETS／FLEETS／ZOOM／LEADERS／RACES／INFO與TURN。命名視窗消失；星圖控制尚未另送正常輸入驗收，不稱完整開局完成。PNG SHA-256 beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832，RGB 9b433167360cbfa77c0422b5ba6848db3efba59cd7d168738e0f62d7dd03200f。

既有dumpSetupTable在cap取得DS188:298848、count23／stride55／1265bytes，只讀表SHA-256 4392f446efbdd96acafbfba8ee39cc0e8ac67a2388119df5bfbfef14a89ab15a，globals 6ff76fc0f447a300d6468cb76bc2acc884bd9b0e544c01f0ad101d76ee6d1dfe，header 35d7cde9f525f64e4d64ea3bdaa7bf3ee2770440c7ecb8cbe2f3fc954aff8cda。完整快照前後R／段／flags／FPU／RAM保持；沒有擴充舊count≤16 observer，後者table_readable=false仍是觀察界限。先前完整23物件未取的缺口由同native另一既有入口補齊。

### 工程驗證、來源與回填

裸F6 /5 register及memory來源新增分支逆轉後逐byte等於c39af54的CPU；CPU SHA-256 1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。獨立重複加法、[-128,127]範圍與little-endian lane窄測0.212s通過，1048576 register fixtures／131072 memory值、全ModRM／SIB、DS SS、唯讀／wrap／段末、截短／prefix及讀失敗發布。FPU用非零fixture，memory乘法不得寫入。新測試SHA-256 08e22ec33fba14abed76eb3f4c86f07ebd1e9cfcc54d66dd9103a94c15ad2dc4。

首次Go全套只有289舊NEG拒絕fixture要求F6E8失敗；移除唯一過期樣本，F6 /1、/2、/7與memory NEG護欄保持，測試SHA-256 6248ee23037b32db47c8d68baec8bf84731886cad7d95102c3540123e68fd042。相同容器／命令乾淨重跑固定官方EXE全套通過，CPU38657.925s／machine1.699s，首失敗收據保留。全套後新測試僅兩中文字改成繁體，實作與斷言保持；未掛8088外部資料，不外推其驗收。其餘公開internal及probe保持，主庫玩法未改。

三私有新增區段逆轉後精確等於371；371拒絕前11645共通原列按既有mtime／DTA與每輪RAMhash正規化保持，38PNG逐byte保持，本輪39PNG。六CLI拒絕與mode on／off缺EXE兩正對照保持；新驗證首次與擴充cap／完整表／回填檢查皆通過，原guest沒有重啟。

SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f、sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持。有界同guest副本與終態及371一致，三檔UID／GID1000；原418來源前後保持。監測owned、有界550s，outer600s；所有程序terminal，相關執行中／停止Docker容器為零。既有主庫root-owned2437檔／272目錄保持，不修復未涉檔案。

367／369／370／371追加原F6EC／正常星圖及完整表回填，289追加工程拒絕判準回填；全保留舊正文。分開原ISA定義／工具未定義flags模型／原native／工程fixture，不把新consumer當舊其它點擊上下文。

### 私有收據

以下位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| full-test-372-first-failure.txt | c8cca171857b16523756e2cfb693253f5ca801d1e359d186d4e53a251fcdd300 |
| full-test-372.txt | be9a4c6301c81c392aa52bbc9de2e345c25d837db1d688a8b29c4c14852d15a3 |
| moo2-372-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-imul-byte-372.go | 3de3b64fc461fe5e9b2cf2a29e1d4a0c9d5e230a9fdf8b9c7d959bd4c224a205 |
| moo2-probe-372-overlay.txt.gz | 3b78e609532aa8cd75936ad76128f401da3a29060e206a88e801ca6b24eed910 |
| moo2-save-state-372-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-372-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-372.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-372-overlay.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| new-game-372-cli-tests.txt | e23e63ea4e37e0f0823fda228ce807b3cdac82cc8d2c63c2bf2d1afb1a2312af |
| new-game-372-cpu-branch.txt | 55e9ad252e540e3f1e22c4460b4ce5a9de03d514d43252f1670df74e051255da |
| new-game-372-cpu-tests.txt | 29ea4110b337460eaac7d074c42c812afd6a52691b565d2f25eff1d04a362d1e |
| new-game-372-full-run.sh | a4bf19eba75c6b22df6a0b7435399874d428f8c209221f86a6a6590206163a69 |
| new-game-372-patches.json | 16a1ef0c8832e8c6ffb30f1c2fe8eb4043aaa5d8b19f58fd563edb910159f3ce |
| new-game-372-ready-review-tests.txt | f343dd68b926087145c90f86c26bcca3bc86137beee9ae48013ec6bdbb476e46 |
| new-game-372-ready-review.py | ce9864a3eff0fd7435d1a9503c80b21dbe9c1d33922f21bda61e4b018dd27472 |
| new-game-372-run-output.txt | 884fe62f0b4d63fb72e11fec0c2e93b080bbfc5511f4d1d38647bd704797b36b |
| new-game-372-run.sh | 66a1df5d644c17e4f6d929064b9aad10dac0fa2c6728f8d438e402bec4baac51 |
| new-game-372-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-372-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-372-state-capture.py | 93bcc27910bb1271e1836a0301d8987877e1580a412858a97e1a0ed0d3717940 |
| new-game-372-tests.txt | b6de5fc5ff6b5be092915da3d40d8b869652f499c7a8c081fa3d96b52750f540 |
| new-game-372-verify.py | 52eaa125a91cc85002653cd653dd8b98820b4c8aef7795dc80d184478b29cd8d |

實際Docker命令依序new-game-372-ready-review.py、裸byte IMUL窄測、new-game-372-full-run.sh、new-game-372-run.sh、new-game-372-verify.py；READY審查在CPU編輯前，原guest一次，full過期判準修正後以同命令重跑。深層契約與逐步結果見上方372工具規格。

下一步以180M正常星圖及23物件核對首個COLONIES來源、矩形、callback與安全輸入前置，先有界只讀，再READY正常裝置輸入。主庫玩法RE-first保持；正式存讀語意、typed名稱／旗色持久writer、母星配置、星圖正常控制、完整開局、RNG及remake同狀態未驗。固定1996日期不是seed。

## 2026-10-04：373正常星圖COLONIES來源與安全前置

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具基線e57e9e063b1713b087423a78bef1349237c3d4b0，CPU保持1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。工具f92dd15be1f5bf94d193d9bfc0367f2baa5793a8已推送github，見[373限定來源規格](https://github.com/wicanr2/dosgolem/blob/f92dd15be1f5bf94d193d9bfc0367f2baa5793a8/docs/spec/373-moo2-star-map-colonies-source.md)。公開本規格／索引與372追加回填，全部公開internal／DOS／probe保持；原LOG／PNG／RAM／state及probe不入Git。

### 已證實來源與未知邊界

原180M的DS188:298848完整23物件／stride55／1265bytes保持372，SHA-256 4392f446efbdd96acafbfba8ee39cc0e8ac67a2388119df5bfbfef14a89ab15a。按原前8bytes四signed word矩形，邏輯43,450只命中index10矩形17,434–79,471；與原圖COLONIES對應為強推論，正常選取consumer尚未驗。

index10+24→DS188:261716共32bytes首byte00、SHA-256 487a51bcf042e9ed41e13586fbf0e823b3fa9bc77dae18296d9d34c95739de53；+32→2801A9共16bytes全0、374708fff7719dd5979ec875d56cd2286f6d3cf7ec317a3b25632aab28ec37bb；+44→3ED894原16bytes、c36adafdfcf829c80935633c82a6fe2bf9d24469ce4ae188bc6881dfe21dade2。原+40為0，維持原offset與bytes，不推定圖片、標籤或正式資料型別。

index1–5的+24原窗口26A104／26A10D／26A113／26A119／26A12B，首NUL前文字EINSTEIN／MOOLA／MENLO／ISEEALL／SCORE已直接取得；矩形全-1，沒有當成底部COLONIES熱區。八窗口224bytes可讀，原字串用途未知，不追handler或renderer。各32byte窗口雜湊見373規格與私有驗證輸出。

原callback target8:2136D1，mask2B、pending0／inactive、callback14／14與IRQ47425／47425完成且非failed，IF1。R=[35017C 70 39C17C 70 2BD478 2BD4C0 39C17C 35017C]、段=[8 188 188 0 20 188]、flags202h、FPU127F／status0／depth0／八stack bits0，虛擬414027099µs。code16與SS188:ESP2BD478 stack16已取；快照前後完整核心／FPU／clock／VBE／callback／IRQ／全RAM保持。只有一筆star_map_source_snapshot，valid／readonly true，colonies_press_sent=false。

### 正常重播驗證

原guest一次，所有372正常輸入／日期／state／180M cap保持，實際step_limit=180000000／EIP2176C5／unique_sites54239，無guest_cpu_stop／step_error／dos_exit。372全部12047共通原列按既有mtime／DTA與每輪RAMhash正規化保持；IMUL ram_effect維持changed_bytes與hash是否相等的關係，只正規化每輪值。39PNG及final逐byte保持，終PNG beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832，RGB 9b433167360cbfa77c0422b5ba6848db3efba59cd7d168738e0f62d7dd03200f。原372已親看的正常星圖保持，沒有將觀察寫成新的COLONIES點擊。

function及單一cap call兩私有區段逆轉後逐byte等於372，全部公開internal／CPU／DOS／probe保持e57e9e0，沿372固定官方EXE全套收據，不重跑無關CPU測試。六CLI拒絕與合法mode on／off缺EXE兩正對照保持；首次驗證、完整回填與索引檢查通過，未重啟或挑結果。

SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f、sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持，同guest有界副本與終態及372一致，UID GID1000；原418來源前後SHA-256保持。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀，owned監測550s／trap清理。所有程序terminal，相關Docker執行中及停止容器為零，既有root-owned2437檔／272目錄保持。

### 私有收據

以下位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-373-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-probe-373-overlay.txt.gz | dbb1b5c634b31c1bdba2938450a0d325e274bc75969dc75a3235d33c79da0475 |
| moo2-save-state-373-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-373-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-373.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-star-map-source-373.go | a87ebf311db6ce15cb919176705ff6c89bad252b14579893150db045199ddb4d |
| moo2-vbe-373-overlay.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| new-game-373-cli-tests.txt | e23e63ea4e37e0f0823fda228ce807b3cdac82cc8d2c63c2bf2d1afb1a2312af |
| new-game-373-patches.json | 9c04d40a45cf6cc97920a0665d84c3b26befc830372828bbada3b3bf55f323b2 |
| new-game-373-ready-review-tests.txt | 30b225e902ce249aa492dc0572e5d4eee7e0da4f6b8bde187cbd473b33b0bb31 |
| new-game-373-ready-review.py | 8edd62169c447618bf6b9ebc406320f74e8054c8d217ff80a775908e36b7b108 |
| new-game-373-run-output.txt | d6dfe3cc91ba02036b3acdbc19314b7ee79676e58b00b1129ecad1340d5b6f62 |
| new-game-373-run.sh | 846f63733f1a07b4d96998ef5c9e670f3bb22a1e7d55bf2b09099f7c07ca4d65 |
| new-game-373-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-373-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-373-state-capture.py | dfd358762f7137f1e895f23623cf3c69eba025e54e5e1d0203a6123f8b38d58d |
| new-game-373-tests.txt | 4ada99ccb04431049cd714cb92a1403a0d2f5d01d7df024c0ef9405c1d5b74b2 |
| new-game-373-verify.py | 2c69315ac167cb7121eb5d59c45f18c64ef568e3d13d5391c5a1e8786f3cd71d |

實際Docker入口依序new-game-373-ready-review.py、new-game-373-run.sh、new-game-373-verify.py；READY審查在新增observer前，八窗口計數修正為224bytes後才實作。372正文同次追加回填，深層來源與每筆窗口雜湊見上方373工具規格。

下一步以本180M前置審查獨立COLONIES一次正常press、原AX3查詢、首安全release與原選取consumer。邏輯43,450依既有2:1橫向裝置尺度對應physical x86,y450，仍須READY審查，不代寫原選取word或核心／RAM。新輸入需另明示有界後續預算與模式，不修改本373的180M收據或加cap挑結果。主庫玩法RE-first保持；正式存讀語意／typed名稱與旗色持久writer／母星配置／星圖操作／完整開局／RNG與remake同狀態未驗。1996固定日期不是seed。

## 2026-10-04：374正常COLONIES輸入、原選取10與185M黑終圖

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，基線工具f92dd15be1f5bf94d193d9bfc0367f2baa5793a8，公開CPU保持1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。工具20fd1507c45a053860b0ad50ad9f9f86049566ad已推送github，見[374限定正常輸入規格](https://github.com/wicanr2/dosgolem/blob/20fd1507c45a053860b0ad50ad9f9f86049566ad/docs/spec/374-moo2-star-map-colonies-click.md)；公開規格／索引及四份回填，所有internal／CPU／DOS／probe保持。原資料、LOG／PNG／RAM／state及probe不入Git。

### 已證實正常輸入與原consumer

373完整180M來源、八窗口hash、RGB／核心／FPU／clock與target／IRQ一次對接，邏輯43,450／44,450唯一命中23物件index10矩形17,434–79,471。原180000000正常press x86,y450／buttons1、414027099µs，target8:2136D1／mask2B／pending0／inactive、callback14／14與IRQ47425／47425完成。

原180010886、24C31B的INT33 AX3真正返回BX1／CX86／DX450，段與flags16h保持。180010921／414069959µs首次符合安全條件release x88,y450／buttons0，持按42860µs，IF1、mask1、callback15／15與IRQ47428／47428完成且inactive；沒有重送、代寫原核心／RAM或選取結果。

180019489原214104 C3依SS188:ESP2BD44C stack返回20DB5B，ESP+4／AX0／flags246h保持。180020238原20DDDB／66A3A6C42600真正寫DS188:26C4A6 word0000→0A00，EIP20DDE1，R=[A 0 226 FFFFFFFF 2BD450 2BD4B8 FFFFFFFF 2BDC2C]、段／flags202h／FPU保持，原選取10已證實。180144100第二214104 RET依SS188:ESP2BD3A0 stack返回174742、ESP2BD3A4，AX0／flags246h及word0A00保持。首store與最多三RET只命中兩RET，未猜額外返回或AX10。原三consumer快照均只讀；既有raw32窗口保持，不推定新列表資料或正式名稱writer。

### 固定185M、黑畫面與觀察界限

新COLONIES模式cap固定185M，180M後只觀察5M；mode off維持373全部180M與guard，新185M由完整依賴限定，沒有增加cap求過。實際step_limit=185000000／EIP223A71／unique_sites55872，無guest_cpu_stop／step_error／dos_exit。R=[0 0 0 0 2BD720 2BD980 0 2BDC2C]、段=[8 188 188 0 20 188]、flags206h／IF1、FPU127F／status0／depth0／八stack bits0，虛擬424485517µs；callback16／16與IRQ48857／48857完成且inactive／非failed。colonies_terminal pressed／released／polled／store_seen true、returns2。

終PNG已親看全黑，SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。RGB SHA-256 0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366等於921600個零byte；indexed307200bytes SHA-256 4d46c5beedd237ddba268a74a01d8c33a4a0323fb52213a2e5d5b6ea4d688d43，與全零indexed的7818f5542a0404157573be6cffc0e0c8e68ce3c0f5d17d07ccdd9313fb700baf不同。已證實索引畫面非全0／RGB全0；色盤與轉頁邊界待查，不能稱列表正常開啟，也未證產品缺陷。

VBE bank4／startY0／sets2941／writes55575758／display94。新header count20／stride55／pointer298848，舊count≤16 observer明示不擷取，records空；目前1100bytes完整新表未取，既有cap入口的10M條件在185M未命中。舊menuFrame與return仍為較早caller，不當目前223A71 frame來源。正常選取到新header已驗，殖民地畫面與內容仍未知。

### 驗證與回填

373到180M source snapshot為止11981共通原列與39PNG保持。只將universe_continuation_config及hardware_keyboard_schedule兩處maxSteps在比較前正規化為已驗180M baseline，新增colonies_continuation_config明示baseline180M／cap185M／window5M；實際新185M另核對。舊180M cap terminal rows不混入新輸入後比較。mtime／DTA／每輪RAMhash沿既有契約，IMUL ram_effect保留changed_bytes與hash是否相等關係。

八私有patch逆轉後精確等於373，所有公開internal／CPU／DOS／probe保持f92dd15；沿372固定EXE全套，另建置本probe，不重跑無關CPU測試。13CLI拒絕、mode on185M及舊mode off180M的home on／off三正對照通過，185M只由完整新模式使用。原guest一次，沒有重啟或挑結果。

初版原輸入驗證通過；擴充終態的非空parser讀records=空失敗，改為明示空欄位判準後同收據通過，保留首失敗輸出。這是驗證parser修正，沒有修改原native結果、程式或guard。367／371／372／373追加shared word選取10的新星圖上下文，保留舊正文與其它名稱確認／持久未知。

原418來源前後SHA-256保持。SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f、sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持，終態與373一致；同guest有界副本與終態及UID GID1000核對。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。所有程序terminal，相關Docker執行中及停止容器為零；既有root-owned2437檔／272目錄保持。

### 私有收據

以下位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-374-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-colonies-click-374.go | 068ba87dbc9485f3f96607393287815318af7812e75c121ab5fd77f78d09c4f8 |
| moo2-probe-374-overlay.txt.gz | 251a8f37a090a5d37f9e02dec128f7cedd08fbc208151e49c78fc9dafb646bd7 |
| moo2-save-state-374-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-374-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-374.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-374-overlay.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-374-cli-tests.txt | 2dbf20d4bdf4311e15de13d3a416bfe1f5043e9675a59e320dd49ec4569ecb61 |
| new-game-374-patches.json | ae787147e5944122065c4ddbbe8a19333688725476ad3cc4ff1a7f96abebe072 |
| new-game-374-ready-review-tests.txt | 79574aff1345c9f9b68f211e17bb27cdaba4776ba56702c6e92cc05c367c84bd |
| new-game-374-ready-review.py | d45760eafcd72d79ee3e21e16dd92acf67496b6ac438a2c25768bbb4817073a6 |
| new-game-374-run-output.txt | 8ac679ac86da8a834c40ec8c195b1e888026cfd6fb2c2c352f6df02aeae443cb |
| new-game-374-run.sh | f5cc1112fc1da075e3eb91c2a5afe1bf42789b0c8cd0da8d058d6f004b12a67a |
| new-game-374-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-374-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-374-state-capture.py | 876a39f3725e1356ba28ff3422f0a9e689d1299656cbe2157bc8a4d62c153ef9 |
| new-game-374-tests-terminal-parser-first-failure.txt | 6d61755220687b3c5ad037bb6285ce941871e0e1cea511971774676a4a4b0a02 |
| new-game-374-tests.txt | 202724dc4a5a8836c27a920ae6a68f3e83c99f12580090467df51756dc474558 |
| new-game-374-verify.py | 9ae54694b8b30b33f4350b42f7914703f32a4d1537a2e36021df7e1ca950e53d |
| new-game-374-window-hashes.json | 6904a18ee3e8deca39c2cf0aca8d4f26fc380817e05e30fa29ee09217af08da0 |

實際Docker命令依序new-game-374-ready-review.py、new-game-374-run.sh、new-game-374-verify.py；READY審查在私有新模式前，原guest一次。原輸入與末態逐值、四份回填及深層契約見上方374工具規格。

下一步保持相同185M與正常輸入，先取20物件完整1100byte表、當前code16／SS:ESP stack16及有界索引使用集合／RGB對應。沿internal/machine/moo2_vbe_video.go的VBEIndexed／VBERGB只讀API與既有dumpSetupTable count≤64入口，另審查185M固定取樣，不放寬舊count≤16 guard。釐清正常轉頁邊界，不盲目擴cap求過、不深入DAC／PIT／driver或renderer helper。主庫玩法RE-first保持，列表內容與正常操作、正式存讀語意／typed名稱旗色持久writer／母星配置／完整開局／RNG與remake同狀態未驗；1996固定日期不是seed。

## 2026-10-04：375完整20表與原DAC全零色彩來源

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具基線20fd1507c45a053860b0ad50ad9f9f86049566ad，公開CPU保持1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。工具03dcee257142e790661be653e9ba6102136d162a已推送github，見[375限定只讀規格](https://github.com/wicanr2/dosgolem/blob/03dcee257142e790661be653e9ba6102136d162a/docs/spec/375-moo2-colonies-color-source.md)。公開本規格／索引／374回填，全部internal／CPU／DOS／probe保持；原LOG／PNG／RAM／state及私有probe／getter不入Git。

### 已證實與未知

保持374正常COLONIES press／poll／release、原共享選取10與相同185M，原guest一次；未增加cap或輸入。新DS188:298848／count20／bias0／stride55完整1100byte表取得，SHA-256 d308de8fbcf9736ce8b4edb64e7c93b3e0cad2bfb8d7024da0776fef36235387。globals192／header16保持374，未推定新物件語意。當前EIP223A71的code16=C1E0028B8014392A000345A88A0025FF，SS188:ESP2BD720 stack16全0，不使用舊menuFrame套用目前frame。

同VBE device的rawDAC768bytes全0，maskFF，mapped palette768bytes亦全0，兩者SHA-256均ef115a0e0c15cdc41958ca46b5b14b456115f4baec5e3ca68599d2a8f435e3b8。256bin histogram共307200pixels，index0有294477、其餘12723非0；使用索引集合與RGB樣本逐值核對。以逐bit除法與算術擴展獨立核算既有Palette()，全部pixel RGB符合索引映射。indexed SHA-256 4d46c5beedd237ddba268a74a01d8c33a4a0323fb52213a2e5d5b6ea4d688d43、RGB SHA-256 0b150fd32588b1daca5569992ebe559c0102c837306b1af4c44d35128ec58366、black PNG SHA-256 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622保持。

已證實黑色輸出由當前全零DAC映出，mask並未把所有索引限制到0。色盤何時歸零、是否正常轉頁中間態與後續恢復仍未知，不能稱列表開啟或產品缺陷。此契約只驗工具當前色彩映射，不稱硬體逐週期或逐波形對齊。

### 驗證與保存

實際step_limit=185000000／EIP223A71／unique_sites55872，無guest_cpu_stop／step_error／dos_exit。R=[0 0 0 0 2BD720 2BD980 0 2BDC2C]、段=[8 188 188 0 20 188]、flags206h／IF1、FPU127F／status0／depth0／八stack bits0，clock424485517µs；VBE bank4／startY0／sets2941／writes55575758／display94。callback target8:2136D1／mask2B／pending0／inactive、16／16與IRQ48857／48857完成且非failed。完整核心／FPU／clock／VBE／callback target與state／IRQ／device State／ports Reads與Writes map及Log長度4096／rawDAC／mapped palette／mask前後保持，全RAM前後hash一致。兩新observer只讀有效；getter不呼叫IO或InstallLEVideo。

374全部12253共通原列／39frames及black final保持，只有185M完整表與一筆色彩快照新增；mtime／DTA／每輪RAMhash沿既有比較契約，IMUL ram_effect保持changed_bytes與hash相等關係。三私有patch逆轉精確等於374，全部公開internal／CPU／DOS／probe保持20fd150。13CLI拒絕及三正對照保持，沒有新CPU行為，沿372固定EXE全套，另建置private probe，不重跑無關測試。374正文保留並追加375回填，索引同次更新。

原418來源保持，SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f與sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持；終態與374一致，同guest有界副本及UID GID1000核對。固定日期不是seed。

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，Docker network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。實際入口依序new-game-375-ready-review.py、new-game-375-run.sh、new-game-375-verify.py，READY在私有實作前，native一次。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

### 私有收據

以下位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-375-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-colonies-color-375.go | 20d14f37740a286cf538d606de5298bdeecaf9f77cefd3bea7ec6c21f97ca1f1 |
| moo2-probe-375-overlay.txt.gz | 0302d7e2ba4c7aa725566897d6637d4fd84128df9a5fe682269d1163304c3874 |
| moo2-save-state-375-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-375-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-375.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-375-overlay.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-375-cli-tests.txt | 2dbf20d4bdf4311e15de13d3a416bfe1f5043e9675a59e320dd49ec4569ecb61 |
| new-game-375-patches.json | c85eb5fc11a6cf3b9ff35340ee8695aa4226c6ce436181ace3e0f37f8529b54e |
| new-game-375-ready-review-tests.txt | 393dee8c97be932e7d4753b9e2bcd67199529a89cfb2bb96ea686d61ad4b05a7 |
| new-game-375-ready-review.py | 87103271d48c28d7b399b5c8cf0afc3b443b1e9de2786a84e7fc28f6bfea03fa |
| new-game-375-run-output.txt | c3c7ccedf5b7c9e04d09e7d41175ac95b61793d40d9b693f5bf9db82d0635029 |
| new-game-375-run.sh | 4e1dde60e673e4097aaeedbfa85d4e594b711b432ac141be952bfd5477336db1 |
| new-game-375-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-375-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-375-state-capture.py | 3431ddb64ea971b82ce274ef30d3d4cdaed8af463f8259840f40831ff86cc196 |
| new-game-375-tests.txt | fb2ed2604b605b365542badd8a18732f3e207b2c0bffe325632ff7bc35fb7326 |
| new-game-375-verify.py | ba133ad57d1faefb80472bc70ad9bf8046d164372c5105e1049f30373c90ff70 |
| moo2-palette-snapshot-375-prototype.txt | 69da24f32ed36f2791689bb05e2f2ffe35aa9167c5cb7d3369728fefea0e3db7 |

下一步保持相同輸入與185M，審查internal/machine/machine.go的DAC ports既有寫入入口，再以有界私有只讀觀察保存180M起到185M的DAC寫入總數、首個全零與最近寫入邊界及其原核心／clock。只觀察既有寫入，不添加IO、色盤修補、輸入或盲目擴cap，不深入DAC／PIT／driver或renderer helper。主庫玩法RE-first保持；殖民地列表內容與正常操作／正式存讀語意／typed名稱旗色持久writer／母星配置／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：376原DAC降色、歸零與末寫入邊界

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具基線03dcee257142e790661be653e9ba6102136d162a，公開CPU保持1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。工具3295ddcac19dfbbebed167a490cebed7859c86a2已推送github，見[376限定DAC序列規格](https://github.com/wicanr2/dosgolem/blob/3295ddcac19dfbbebed167a490cebed7859c86a2/docs/spec/376-moo2-colonies-dac-write-journal.md)。公開本規格／索引與374／375回填，全部internal／CPU／DOS／probe保持；原序列／rawDAC／LOG／PNG／RAM／state及private probe／getter不入Git。

### 已證實的原降色與未知恢復

180M正常COLONIES press前原DAC有592非0值，SHA-256 ddf4dd57bc57ded0c9deddf28069189396b21f3c855841debfe98a1f482d5292、maskFF。既有LEOPLPorts.Log只保存最早4096筆，本輪從同VBE device完整PortLog讀取有界增量，保存原sequence／port／value／device Step；device Step不當outer step，loop現在EIP不當內部write原執行位址。

完整11275group／11275DAC事件，3C6／3C8／3C9為11／2816／8448筆。從初始DAC／mask／index／phase獨立以除法／餘數重播，全部group rawDAC hash／非0數／mask／index／phase／ports三埠累計與終態一致。11輪均maskFF／index0..255／768色值，所有值相對初態與前次單調不增；這是數值來源已證實，不宣稱硬體逐週期或轉頁完成。

| 輪 | 起觀察step | 末觀察step | 末非0色值數 |
|---|---|---|---|
| 1 | 182482821 | 182489427 | 592 |
| 2 | 182490806 | 182497412 | 583 |
| 3 | 182498793 | 182505399 | 583 |
| 4 | 182506486 | 182513092 | 583 |
| 5 | 182514179 | 182520785 | 583 |
| 6 | 182521872 | 182528478 | 583 |
| 7 | 182529565 | 182536171 | 571 |
| 8 | 182537258 | 182543864 | 554 |
| 9 | 182544951 | 182551557 | 545 |
| 10 | 182552644 | 182559250 | 496 |
| 11 | 182560337 | 182566943 | 0 |

首次全0與末DAC事件同為sequence290512／port3C9／value0、loop觀察182566943／419464025µs、目前EIP222D1C。R=[0 1 7003C9 64 2BD99C 2BDBD0 2A3758 2BDC2C]、段=[8 188 188 0 20 188]、flags6h，FPU127F／status0／depth0／八stack bits0；target8:2136D1／mask2B／pending0／inactive、callback16／16與IRQ48162／48162已返回且非failed。VBE bank9／startY512／sets2935／writes55268558／display93。寫入後loop邊界保存目前核心／clock，未猜每write精確時間。

182566943後到185M沒有DAC寫入。原185M／223A71、flags206h／clock424485517µs與完整核心／FPU／VBE／callback16／16／IRQ48857／48857、20表及全零DAC／maskFF保持375。降色到零已證實；色盤恢復、正常轉頁與列表內容仍未知，不稱黑圖是列表完成或產品缺陷。

### 驗證與收據

375全部12255共通原列／39frames／black final與兩只讀快照保持，只有一筆新journal摘要。375色彩快照RAMhash逐輪正規化，前後相等／只讀與其餘bytes／RGB hash保持。四私有patch逆轉精確等於375，public internal／CPU／DOS／probe保持03dcee2；13CLI拒絕與三正對照、原418來源／state／同guest副本／UID GID1000通過。沒有新CPU行為，沿372固定官方EXE全套，另建置probe，不重跑無關測試。

增量getter前後完整R／段／flags／FPU／clock／VBE／callback與IRQ／device／ports map與早期Log／完整journal／DAC／mask／index／phase保持，初態和終cap另驗全RAMhash；length getter純讀，不呼叫IO或Restore，不代寫模型。最大delta4096／事件和group262144界限未超出。374／375正文保留並追加376，索引與backlink驗證通過。

原guest一次，原SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f與sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac、network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。實際入口依序new-game-376-ready-review.py、new-game-376-run.sh、new-game-376-verify.py，READY在私有實作前；相關Docker容器清理，既有root-owned2437檔／272目錄保持。固定日期不是seed。

以下21份位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-376-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-colonies-dac-journal-376.go | ef19ab41906cf84aec2a8498e82e91deae9f7c4cc0d8fd0df3be87ac15940507 |
| moo2-probe-376-overlay.txt.gz | c79298f8e2bd24283593a0c24e9c2c1005a08cba8f4cb57d52d9a1a7d350529d |
| moo2-save-state-376-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-376-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-376.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-376-overlay.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-376-cli-tests.txt | 2dbf20d4bdf4311e15de13d3a416bfe1f5043e9675a59e320dd49ec4569ecb61 |
| new-game-376-patches.json | 2eee0f5d63561346e928a9c0f38ccf6cb0d66fb087f3115960d05699b65313eb |
| new-game-376-ready-review-tests.txt | bebc250bbc1974165b414d1fd27b9e63e3a2818a7c0d033100ed5ab8858282ff |
| new-game-376-ready-review.py | 9a284d301e98ff8216b6330b2097b2b925605833e32d40b2f802f9877422b2b0 |
| new-game-376-run-output.txt | 0281065dab51441d91a7dd15b0de89726a3ae1649a1fc786edc99c8a7c8192ff |
| new-game-376-run.sh | 68a91e84eef2a66839c6515ddd80ee2f727a76e6dc32d8b4ed4df12375548e70 |
| new-game-376-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-376-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-376-state-capture.py | cc197ea5bceeb15723ace565b0bfa525f3df4e357edeb077b35f2036aa130035 |
| new-game-376-tests.txt | 8a6503c9d9e969ecf7ed5ac26d70a7eeeee9c525660e74a25b117a37286e6aa8 |
| new-game-376-verify.py | f138637397d6f65f60f3df63bb7b6b6e9539cb978661f9c62f806d1b5163f654 |
| moo2-dac-journal-376-prototype.txt | df69292f30e77a6b5b7c3c550edff42d8775c804cb2c16a5d2a4838d0ddd46e9 |
| new-game-376-dac-journal.json | 574b0bba3f916f6994a13717437f0ba9f58e0af451a6cc8565b20bada50f9e83 |
| moo2-palette-snapshot-375-prototype.txt | 69da24f32ed36f2791689bb05e2f2ffe35aa9167c5cb7d3369728fefea0e3db7 |

下一步以376已證實的11輪單調降色與182566943歸零為來源，先審查新的轉頁觀察契約：保留185M完整前置，限定追加一次10M窗口至195M，追首個恢復非0色值的DAC寫入並保存原核心／clock與可見頁；未恢復時記錄實際邊界，不以加碼重跑求過。不得代寫palette、增加玩家輸入或深入DAC／PIT／driver及renderer helper。主庫玩法RE-first保持，殖民地列表正常操作／正式存讀／typed名稱旗色持久writer／母星配置／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：377完整185M前置後的原2C17 CPU停止

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具基線3295ddcac19dfbbebed167a490cebed7859c86a2，公開CPU保持1d8a4d8252372c97d8e873ba13c3ab3670796527dcd6d74de52e8cbf226e068c。工具168b91b9a8cb08a36f9517ae031a98239f841161已推送github，見[377限定續跑規格](https://github.com/wicanr2/dosgolem/blob/168b91b9a8cb08a36f9517ae031a98239f841161/docs/spec/377-moo2-colonies-restore-continuation.md)；公開本規格／索引與376回填，public internal／CPU／DOS／probe保持，原事件／PNG／LOG／RAM／state及private probe／getter不入Git。

### 已證實的實際邊界

保持正常180M COLONIES輸入與376完整185M前置，原MAX_STEPS參數仍185M，另runSteps195M／一次10M明示。原376共通12188列至185M journal／11275 baseline groups／39frames與黑PNG保持，baseline續跑標記另核對，沒有把guest稱185M停止或把後續terminal totals當185M相同。

185M後原188259170／430366579µs至188265776／430373460µs又寫一輪maskFF／index0..255／768零值，共1025事件；rawDAC一直全0，首非0恢復與first-restore PNG均不存在。全部12300group／12300事件、三埠12／3072／9216筆獨立算術重播至終DAC／mask／index／phase／ports累計通過。

原188532362在1F455D，原16bytes2C173C080F87ED0000000FB6C02EFF24。原guest_cpu_stop與step_error明示opcode2C未支援；工具抓opcode後EIP1F455E，不能當SUB已執行。實際reason=cpu_stop／actual_boundary188532362，沒有step_limit=195M；probe exit0不是guest成功。ISA名稱已交叉核對[Intel SDM Vol.2B](https://www.intel.com/content/dam/www/public/us/en/documents/manuals/64-ia-32-architectures-software-developer-vol-2b-manual.pdf)的SUB條目4-654頁，2C ib為SUB AL,imm8，旗標與實作留下一規格，不猜後續CMP／Jcc／jump table語意。

停止R=[1A 0 2BD800 D 2BAFC8 2BD8A0 29E1D8 2BD348]、段=[8 188 188 0 20 188]、flags206h／FPU127F／status0／depth0／八stack bits0、clock430866468µs，target8:2136D1／mask2B／pending0／inactive、callback16／16與IRQ49858／49858已返回且非failed。後移1F455E的code16=173C080F87ED0000000FB6C02EFF2485，SS188:ESP2BAFC8 stack16=1AD82B00000000000D000000F5401F00；原定位與bytes保持，不補frame語意。

終indexed12723非0，DAC全0／maskFF與RGB全0；終PNG親看仍黑，indexed／RGB／PNG hashes與185M相同。VBE bank4／startY0／sets2941／writes55575758／display94保持。20表DS188:298848／stride55完整1100bytes終SHA-256 3c6bd2afa22a4ac6500ce62df53ea9bfe88dd4713423922949d23f179820471c，與185M有10個raw byte自然變化，語意未知，不把header20當整表不變。

### 驗證與失敗分類

八patch逆轉精確回376，public internal／CPU／DOS／probe保持3295ddc，18CLI拒絕與4正對照通過。baseline和終PNG以獨立標準PNG filter／CRC／RGB hash／非0數解碼，palette獨立算術及256bin histogram映色一致；首恢復不存在與實際CPU stop另核對。全新取樣核心／FPU／clock／RAM／VBE／callback與IRQ／device／ports／journal／DAC前後保持。376正文保留追加377，索引及backlink通過。

原418來源、state終態與同guest有界副本／UID GID1000保持。SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f與sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持。原guest一次；無新CPU行為，沿372固定官方EXE全套，不重跑無關測試，固定日期不是seed。

私有生成首次外層here-document與內嵌PY界符重名，Python收到截斷內容，在執行前syntax失敗；更名外層為PY_GEN_377，同內容成功且八patch逆轉通過。此為腳本界符問題，當時未實作／未起guest；首失敗摘要留存，不歸因產品／CPU。

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀、owned監測550s／trap收尾。實際入口依序new-game-377-ready-review.py、new-game-377-run.sh、new-game-377-verify.py，READY在私有實作前。相關Docker容器清理，既有root-owned2437檔／272目錄保持。

以下25份位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-377-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-colonies-restore-377.go | e7b99e7c0f93a7ef386ac09e16f9e2cb23d80a16815bdac9900a95626629c2c8 |
| moo2-probe-377-overlay.txt.gz | ac98bca6c9e88c4e546078182001ab58de76b29400060e8c7c11168bd41b68f7 |
| moo2-save-state-377-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-377-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-377.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-377-overlay.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-377-cli-tests.txt | ae64c7e607a57e1a3d334d200bfe6d52ab92a0a46fbcb955720d8200e833b19d |
| new-game-377-patches.json | 1fb387ff6ff73553869cfbd92f667365216178f0b43c8cd6d986087584030b93 |
| new-game-377-ready-review-tests.txt | d462ef12bd0fe9a1233f7612d1275211eb959bd02220ce1131277d5e0589e053 |
| new-game-377-ready-review.py | c4ea0f55950a075a418b0f9782bbe7a424ecbf608bb452eeec63e104584da748 |
| new-game-377-run-output.txt | d962358ee82d021b06981eff74601c7bbf470ad02e59897be5e05901a63775b8 |
| new-game-377-run.sh | abf96ad8ff28380bba144044d851fffdad6332d7a01e673b233ac90eb04b0121 |
| new-game-377-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-377-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-377-state-capture.py | bbbe5b1f83202ab67455bbd5d339b490b21f88e61d8f54c5537f04266078fb3b |
| new-game-377-tests.txt | db96c06842adfb491b007ce9847a4b89a12bbf852e397edf71c6a09a9e60cb3d |
| new-game-377-verify.py | 56db8cb42f72292f1b576ad854cbabd6f31459a346a52757cce427e2b5192630 |
| moo2-palette-snapshot-375-prototype.txt | 69da24f32ed36f2791689bb05e2f2ffe35aa9167c5cb7d3369728fefea0e3db7 |
| moo2-dac-journal-376-prototype.txt | df69292f30e77a6b5b7c3c550edff42d8775c804cb2c16a5d2a4838d0ddd46e9 |
| new-game-377-baseline-dac-journal.json | 21949db91355df74e56b2e1d142232bfa393c69dd3ca3a959fe97cd37693f13b |
| new-game-377-restore-journal.json | e8ec5bc7c3b368c0d5807eea2303c7c75182a01161d7cf9c62b90e75dddffd83 |
| moo2-377-baseline185.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-377-terminal.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-377-generation-heredoc-first-failure.txt | ee782398d36bf2039cb76f16f65f141424558fef9cb8b43a16102c172d66fd8d |

下一步依原dosgolem_high_le:1F455D／2C17與188532362完整來源，建立CPU386 SUB AL,imm8規格，核對Intel SDM契約及既有byte SUB旗標模型；READY後補2C、獨立256×256輸入及EAX高24bit／其它核心與非算術flags保持、立即數fetch失敗驗收，再重播相同377窗口，不再擴cap。主庫玩法RE-first保持，殖民地列表正常操作／正式存讀／typed名稱旗色持久writer／母星配置／完整開局／RNG與remake同狀態未驗。

## 2026-10-04：378 SUB AL四consumer與原列表恢復

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具來源168b91b9a8cb08a36f9517ae031a98239f841161；新工具d2df07fb795875ffdcdd2b8566ae6ecadc073070已推送github，見[378限定規格](https://github.com/wicanr2/dosgolem/blob/d2df07fb795875ffdcdd2b8566ae6ecadc073070/docs/spec/378-cpu386-sub-al-immediate.md)。公開CPU SHA-256 736d95e801e8e9658078671af8148a79896d805cc7a8baa42f8881574e562080，自製測試ad24baa3a5dc7447b8cd967e1789796465228ca1c0e3ac49d5a10b7af80b8d87；CPU僅新增2C，既有sub8／其它opcode及DOS／probe保持。

狀態：**CONFORMED，限定naked2C、原四步consumer及同195M窗口**。正式CPU只新增2C入口，既有sub8與其它opcode保持；不是整個CPU或remake玩法驗收。

舊core單一原2C17測試先RED，明示opcode尚未支援。新入口通過393216組、64種初算術flags的5184組邊界、原完整核心四步、fetch截斷與拒絕、工具prefix邊界及EIP wrap。固定官方EXE重跑internal/cpu386與internal/machine全套通過，沒有MOO2 skip。首次測試wrapper誤加Bus不存在的Write16／Write32，編譯在測試前失敗；讀回Read8／Write8契約後移除，首失敗另存。真正RED與後續綠測試分開，不把編譯失敗當CPU證據。

原188532362在dosgolem_high_le:1F455D自然執行2C17，AL1Ah→03h／flags206h；下一3C08產生293h，0F87ED000000未跳、EIP1F4567，0FB6C0後EAX03h／EIP1F456A。四步其它R／段／FPU／SS:ESP stack與RAM保持，callback16／16、IRQ49858／49858均已返回且inactive／非failed。原R與RAM未注入，沒有追加玩家輸入；這四步結果已證實，不猜jump table用途。

原377共通12393列至2C入口保持，完整185M DAC baseline、39frames及黑PNG保持；377已有12300事件前綴保持。相同一次195M上限實際到step_limit195000000／EIP22C8BA，無CPU stop或DOS exit。全部24601 DAC事件獨立重播，三埠3C6／3C8／3C9計25／6144／18432筆，終DAC／mask／index／phase及ports累計一致。

首非0DAC write為sequence292693／原189322149／432332347µs／dosgolem_high_le:222CCE／3C9=04h。首恢復快照僅一個DAC值非0，當時RGB仍全黑；不能把這個write當可見畫面恢復。195M終DAC588個非0、indexed296428非0、RGB759775非0，獨立PNG解碼與palette histogram映色一致。終PNG SHA-256 d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba，親看殖民地列表顯示Sol II與人口圖示；不稱列操作、轉入殖民地或人口調整已驗。185M、首非0與終PNG分開保留。

20物件表DS188:298848／stride55／1100bytes，終SHA-256 3c6bd2afa22a4ac6500ce62df53ea9bfe88dd4713423922949d23f179820471c，與377停止表相同；185M至377的10個raw byte變化仍語意未知。既有collector另自然取得第三個214104 RET，原189574450按SS188:2BD920實際stack返回20DB5B，AX0、flags246h與完整其它核心／FPU保持，這次共享word為0000，不把早先選取10宣稱持久不變，也不猜重設writer。

五私有patch逆轉精確回377，CPU單一2C新增區段逆轉精確回168b91b，其它公開internal與原probe保持。18CLI拒絕／4正對照、原418來源與state、同guest副本及UID GID1000保持，原guest一次。完整readonly取樣前後狀態與PNG／原序列核對通過，原LOG／PNG／journal／RAM／state不入Git。

| 不可變鍵 | 新語意與等級 | 舊規格 | 回填 |
|---|---|---|---|
| 官方1.31／同195M上限／1F455D 2C17 | SUB與四consumer自然完成，195M可見殖民地列表，已證實於本收據 | 377 | 追加恢復結果，保留舊缺opcode與黑圖歷史 |

[377](https://github.com/wicanr2/dosgolem/blob/d2df07fb795875ffdcdd2b8566ae6ecadc073070/docs/spec/377-moo2-colonies-restore-continuation.md)正文保留並追加378，索引與backlink同次核對。CPU／原平台／時序範圍保持，Docker容器收尾；收據SHA-256與實際命令連主庫既有研究入口。

下一步以195M可見Sol II與同時取得的20物件表，核對正常列表行的熱區／原選取來源及callback前置，再建立一次正常press／原AX3 poll／安全release的限定驗證。不增加cap或猜欄位，不深挖DAC／PIT／renderer helper。主庫玩法RE-first保持；列表操作／正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

原418來源與state、同guest副本保持。SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f、sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持。原資料、LOG／PNG／RAM／journal／state及private probe／getter不入公開Git。Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000；實際入口依序new-game-378-ready-review.py、new-game-378-cpu-tests.sh、new-game-378-run.sh、new-game-378-verify.py。原ZIP／patch唯讀，owned監測550s／trap收尾，native一次。提交後驗證器對新納入版控的自製測試明示其hash，舊internal仍逐byte保持；相同收據重新核對通過，沒有重跑guest。

以下31份位於workplace/dosgolem/workplace/，不入公開Git：

| 檔案 | SHA-256 |
|---|---|
| moo2-378-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-sub-al-378.go | 4f9e7a3a1e942bbcc88c59fd512fa47ee93a0e9cfe073beff8c445c1d84ae9c1 |
| moo2-probe-378-overlay.txt.gz | af0e9c8d017208017cfab0be515203175b66b295e491dc2b138cceb844615a58 |
| moo2-save-state-378-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-378-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-378.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-378-overlay.png | d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba |
| new-game-378-cli-tests.txt | ae64c7e607a57e1a3d334d200bfe6d52ab92a0a46fbcb955720d8200e833b19d |
| new-game-378-patches.json | ab7709a3b0e0cb8c02504a5bf1f84d75b11e68002d3170094e6dd908c9f3f21a |
| new-game-378-ready-review-tests.txt | 6d3463f0968c2ea411d082b17d788a55ec2d84596f50c59870af3b8e67ee5ba8 |
| new-game-378-ready-review.py | e0404e293a79d0800c36ffb2ece665f882a7656040bc83c85b7e2b9e1e5bbba8 |
| new-game-378-run-output.txt | 16447a4beee89fd24aad0a981e39dbf08a5124e16b1b490f28cfda2055ab0a40 |
| new-game-378-run.sh | bfdf357eb434494e087992e6154e87e22764250301a922c94d16f836d68cf06a |
| new-game-378-state-capture-output.txt | 9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3 |
| new-game-378-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-378-state-capture.py | dbf6926132a6d72563894b774002604a6b3265dcb5f270e9e515963822736654 |
| new-game-378-verification.txt | 86d3ad8077fec36c834d1e3dc9fa6ef9bfa83d57c07da2c063e3dc92ee91b0a7 |
| new-game-378-verify.py | a9e7e92d0411174289c0b2fbec4f727968697fa64fa216c9991fce3aaa092792 |
| new-game-378-baseline-dac-journal.json | bd49185311781922b9acc2b69d6cd0af8d5f2192875db63876bc0eaa2ff1162e |
| new-game-378-restore-journal.json | 825e88b325d7ee2b0fe5c2f99c35dd6a693216e953e7052336a5486d1943ae0c |
| new-game-378-cpu-patch.json | a50940c33ef8d340b1eabacfb2026de142027b18c7585b36c3866ee2b589407c |
| new-game-378-cpu-red-build-first-failure.txt | 06e73c58621ac476395e0a8d354c54d0d9d143c8303eb71ad5de3160448a951f |
| new-game-378-cpu-red-tests.txt | 7f3e98a8c5a7514b78805ba7df90a8d989d4f95d02234d5f7248e81d8188fb46 |
| new-game-378-cpu-tests-output.txt | f1e16ae4ab52508980aa0ec11e8140c325faa35028916ee443d86d2c063094f1 |
| new-game-378-cpu-tests.sh | de4698dd6650914a3dd53e3979fb135b127530631c790da93c0f6680deef3c5d |
| new-game-378-cpu-tests.txt | 5254cc3a60d9da21b8a766170831bd4d2082d9c202fdadedad4ab19931c2648d |
| moo2-378-baseline185.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-378-first-restore.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-378-terminal.png | d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba |
| moo2-palette-snapshot-375-prototype.txt | 69da24f32ed36f2791689bb05e2f2ffe35aa9167c5cb7d3369728fefea0e3db7 |
| moo2-dac-journal-376-prototype.txt | df69292f30e77a6b5b7c3c550edff42d8775c804cb2c16a5d2a4838d0ddd46e9 |

已證實：本CPU ISA範圍、四自然consumer、原事件序列與195M可見列表。未知：列表行操作、共享word重設writer、typed欄位與完整玩家開局。主庫玩法RE-first保持，未宣稱remake同狀態對拍或完成。

## 2026-10-04：379殖民地Sol II行的原第一命中與195M前置

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具來源d2df07fb795875ffdcdd2b8566ae6ecadc073070，新工具c6d319edf0f8a8bacfdc1a53d2eb39a5a53205c1已推送github，見[379限定來源規格](https://github.com/wicanr2/dosgolem/blob/c6d319edf0f8a8bacfdc1a53d2eb39a5a53205c1/docs/spec/379-moo2-colonies-row-source.md)。本輪只讀重用378既有收據，加固定官方EXE的IDA查詢，沒有新native、輸入、cap、RNG或state寫入；public internal／CPU／DOS／原probe保持。

**已證實，原靜態選取規則**：IDA Pro9.4／linear EA sub_11CEF5 11CEF5..11E718，11DC51從index1開始、11DC5B比較count／11DC62 JGE離開、37h stride；四signed word與原有符號bias、含端點矩形。非raw type14首次命中在11DD3B 8945EC保存index，11DD3E EB76直接離開掃描，不繼續到較後物件。raw type非11經11DDD4 JNZ，11DDDB 66A3A6C41700寫IDA word_17C4A6。原runtime定位另標dosgolem_high_le:20DDDB／66A3A6C42600，指令EA與資料operand的LE重定位均加F0000h。真正RET為runtime214104／C3，對應IDA124104／C3、sub_124075 124075..124105／38指令；runtime callback8:2136D1對應IDA sub_1236D1 1236D1..1237F3／74指令。原名、file offset、operand與bytes全保留，未改名或追runtime內部。

**已證實，raw表及靜態導出**：378的DS188:298848／20表／stride55／1100bytes hash3c6bd2afa22a4ac6500ce62df53ea9bfe88dd4713423922949d23f179820471c。index13原矩形12,35–101,65／type7，index19矩形0,0–639,479。logical43,48／44,48均命中13+19，原第一項13；四內角第一項仍13，右上角101,35另與16重疊。11,48／43,34／43,66只命中19，102,48先16。record13 +24／+28／+32／+44的261716／0132h／2801A9／4015EC只保留原值，不命名typed colony id或callback，不取新pointer窗口。

**強推論**：原矩形與可見Sol II名稱區相符。**未知**：實際pressed poll、word13 store、單／雙擊、轉入殖民地、人口調整與共享word生命週期。本來源的預期13不當實際玩家選取結果。

378完整195M actual_boundary／step_limit／EIP22C8BA、R=[264C6C 4C 3E7D8BA 26A98C 2BD4E0 2BD50C 5 1]、段[8 188 188 0 20 188]、flags283h／IF1、FPU127F／status0／depth0／八stack bits0、clock447368391µs與readonly snapshot已核對。callback8:2136D1／mask2B／pending0／inactive／16／16；IRQ51746／51746、inactive／非failed。終PNG d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba、DAC588非0／indexed296428／RGB759775非0保持。10邊界與13前置漂移拒絕、全部public internal／CPU／probe逐byte保持及378正文保留／backlink／索引通過。

首次IDA查詢anchor換算錯，把runtime214104寫成IDA114104，callback也錯位；C3不吻合，first-query完整留存。按runtime減F0000h修正到124104／1236D1後，相同官方EXE重建一次性DB，真正RET與函式邊界吻合，主選取11CEF5全部1550 rows匯出與首查詢一致。初版邊界oracle漏列101,35同時命中16，被原表解析拒絕，修正為13+16+19；第一項仍13，首失敗保留。沒有改原表或CPU來讓預期成立。

IDA9.4 image sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780，UID1000／network none／120s／2GiB／2CPU／128pids，官方patch唯讀/patch、既有工具workplace輸出/out、一次性DB在/tmp。入口new-game-379-ida-run.sh；非空JSON schema1／EXE hash／5365函式／UID1000通過，idat exit1不作判準，stdout空不作失敗。READY審查後才建立只讀驗證器，入口new-game-379-ready-review.py／new-game-379-source-verify.py，Go1.24.13 image／30s／512MiB／1CPU／64pids。沿378固定官方EXE CPU386／machine全套收據，不重跑無關測試；正式.i64、原遊戲來源與state不變。

以下16份位於workplace/dosgolem/workplace/，不入公開Git，含兩份重用的378來源：

| 檔案 | SHA-256 |
|---|---|
| moo2-379-ida-list-input.py | 9b09c3ab136cd4f4ba564c238ec75c312c455cf56f0face50df8e537a41c7ed6 |
| moo2-379-ida-list-input.json | a006c2e57f9f12c563c7c8fe3254ddc388e9db6c68221fb6eb2b4d93a548dc26 |
| moo2-379-ida-list-input.log | bd7f4000aa9de0dcfeb58d83a208bc07d79befdec5d529d8e35d92bff376a8eb |
| moo2-379-ida-list-input.stdout.txt | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| moo2-379-ida-list-input-first-query.py | 96486d5f8d0e3ab422ee6ca59771800faf86bde81b0c1735f8b911a72cb4ddc9 |
| moo2-379-ida-list-input-first-query.json | c3fd822b3689f07987ba68b6b4c0257b4a1b1e4711775dc2cde42339be417902 |
| moo2-379-ida-list-input-first-query.log | 390b6acf40ad029c30004ca301017bc66ce78acc966523feed9a8654e646a797 |
| moo2-379-ida-list-input-first-query.stdout.txt | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| new-game-379-ida-run.sh | 089969c1f0907eca1b78350cef73cdf2ee16176579c91aeece2b54437e9ba546 |
| new-game-379-ready-review.py | 98cba94fb5cb066b1bc06d496448922f1661060364688c712b37753e0883a4ec |
| new-game-379-ready-review-tests.txt | 68367e6ce908ef7db7a89a85bb0d551263af7f95a7ccb6cb39d27653d5e147d2 |
| new-game-379-source-verify.py | 1685869bb0a59313c6f9d76c93c8d1134d8a68af389fb4df7282b934b015a0b8 |
| new-game-379-source-tests.txt | 1d32e811b4d8b24a3546db4af0691ef00c6eea1d73fefbc6ff149100572faa5e |
| new-game-379-source-tests-first-failure.txt | e8dfbb0d2da4e3ccb3463fae252b1b15d6fc670f455b809a8253bf7579c37a3c |
| new-game-378-restore-journal.json | 825e88b325d7ee2b0fe5c2f99c35dd6a693216e953e7052336a5486d1943ae0c |
| moo2-378-terminal.png | d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba |

下一步依379已核對的195M完整來源與原first-match順序，建立Sol II行一次正常press／原AX3 pressed poll／首安全release的獨立READY契約，明示新玩家輸入的固定後續預算；先驗原index13選取store，再記實際畫面。logical43,48／44,48對應physical86,48／88,48，不代寫word13或跳handler，不重啟／重擲／加cap挑結果。主庫玩法RE-first保持；正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

## 2026-10-04：380正常Sol II行press／poll／release與原選取13

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le，工具來源c6d319edf0f8a8bacfdc1a53d2eb39a5a53205c1。工具5d3f5b80373c4872e1401366d3b5dd482577252a已推送github，見[380限定規格](https://github.com/wicanr2/dosgolem/blob/5d3f5b80373c4872e1401366d3b5dd482577252a/docs/spec/380-moo2-colonies-row-click.md)。公開CPU／internal／DOS／原probe保持；CPU hash736d95e801e8e9658078671af8148a79896d805cc7a8baa42f8881574e562080，本輪只改忽略workplace的觀測／裝置輸入探針，原資料與GUI未代寫。

狀態：**CONFORMED，限定正常Sol II行press／poll／release與原選取13**。殖民地畫面仍未驗，200M終圖為黑，不稱完整行操作或開局完成。

原378共通12701列至195M保持，只有新row config與195M來源標記另核對。完整185M baseline及39frames、首恢復PNG與24601 DAC前綴保持；新195M baseline-row PNG逐byte等於378可見列表終圖，完整core／FPU／clock／VBE／table／code／stack與device來源保持。沒有拿不同終點的callback18或DAC35876稱195M同狀態。

195000000／447368391µs source_ready／valid／readonly true，target8:2136D1／mask2B／pending0／inactive、callback16／16與IRQ51746／51746已返回。一次press physical86,48／buttons1／delta0,0；原195216883在24C31B的INT33 AX3返回BX1／CX86／DX48，段與flags16h保持。195216918／447838123µs首安全release physical88,48／buttons0，持按469732µs；mask1／pending0／IF1、callback17／17與IRQ51807／51807完成且inactive。只送此一次press與release，不注入結果。

原195225486、dosgolem_high_le:214104 C3按SS188:ESP2BD920原stack首word5BDB2000返回20DB5B，ESP+4、AX0／flags246h與其它核心／FPU保持，word0000保持。原195226311在20DDDB／66A3A6C42600真正寫DS188:26C4A6 word0000→0D00，EIP20DDE1；R=[D 0 2CB 400 2BD924 2BD98C 0 2BDC2C]、flags297h及完整其它核心／FPU保持。原store13已證實，Sol II名稱區的正常選取鏈已驗，不把13當typed colony id或持久欄位。raw32候選窗口仍為原Sol字串／原bytes，保持不命名正式資料。最多三RET只命中一筆，不預填額外返回。

實際step_limit200000000／EIP223A23／unique_sites60185，無guest_cpu_stop／step_error／dos_exit；完整35876 DAC事件獨立重播至實際末態，三埠3C6／3C8／3C9計36／8960／26880筆。新點擊後11275事件為11輪maskFF／index0..255／768色值，從196292723／450713469µs開始，逐component單調不增；首全0與末write同在196376553／450990327µs、device sequence315113、當前EIP222D1C。200M前未恢復色彩；只稱本收據降色，不追DAC／PIT／driver逐週期。

終indexed307200bytes全0、DAC768bytes全0／maskFF、RGB921600bytes全0，獨立PNG filter／CRC／RGB hash及palette histogram映色核對。PNG親看全黑，hash1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622。VBE bank5／startY512／sets3031／writes57388844／display97，與195M分開記錄。原表DS188:298848 count1／bias0／stride55、record0的55bytes全0，hash02779466cdec163811d078815c633f21901413081449002f24aa3e80f0b88ef7；不是舊20表不變，也不猜新的正式控制項。

原200M R=[2A375C 0 A4 15 2BD908 2BDB68 30E76A57 2B0000]、段[8 188 188 0 20 188]、flags202h／IF1、FPU127F／status0／depth0／八stack bits0、clock458239660µs，callback8:2136D1／mask2B／pending0／inactive、18／18與IRQ53180／53180完成且非failed。當前code16=0345A88A0025FF000000C1E0028A805A，SS188:ESP2BD908 stack16=F3A6210000010000AE01000000010000；只保存定位，不把首stack值21A6F3當已證實caller。

14private patches逆轉精確回378，所有公開internal／CPU／DOS／原probe保持c6d319e。18舊CLI拒絕／4正對照保持，新增10拒絕／2正對照共34通過；mode off仍195M，mode on是明示新玩家輸入的200M。完整DAC／PNG／195M前置與原store／RET只讀保護通過。原guest一次，沒有首失敗或重啟，不增加cap挑結果。沒有新CPU行為，沿378固定官方EXE CPU386／machine全套，不重跑無關測試。

原418來源前後保持，SAVE10.GAM208000bytes／0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d、MOX.SET553bytes／de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f與sound.lbx4250888bytes／3f0354ac5c1b13a3c5c4fd098c2cbc22af37b71e74fc95e582782c22024c449d保持；同guest副本與終state／UID GID1000核對。這不證正式玩家存讀內容或持久選取欄位。

| 不可變鍵 | 新語意與等級 | 舊規格 | 回填 |
|---|---|---|---|
| 官方1.31／195M表hash3c6bd2…／index13；dosgolem_high_le20DDDB／DS188:26C4A6 | 正常行press／poll／release與原writer13已證實；200M黑圖，新UI未知 | 378、379 | 分開來源預期、原選取與未驗殖民地畫面，保留舊收據 |

378／379正文保留並追加380，索引與backlink同次驗證；較早其它選取上下文不由本row13外推。原LOG／PNG／RAM／journal／state及private probe維持本機，Docker清理與收據連主庫研究入口。

下一步依200M原完整核心、count1／55零bytes、當前223A23 code與SS188:2BD908 stack16，只追正常畫面建立所需的最小上層來源；先核對21A6F3是否真為該路徑的return定位與其原呼叫邊界，再由來源決定下一個有界畫面觀察，不盲目加cap或深挖renderer／DAC／PIT helper。主庫玩法RE-first保持，人口調整／正式存讀／完整開局／RNG與remake同狀態未驗，固定日期不是seed。

Go1.24.13 image sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac，network none／600s／2GiB／2CPU／128pids／UID1000，原ZIP／patch唯讀，owned監測550s／trap收尾。READY前核對379來源，實際入口new-game-380-ready-review.py、new-game-380-run.sh、new-game-380-verify.py。收據驗證90s／1536MiB／1CPU／128pids，兩次只重讀相同收據。首次通過，沒有腳本、CPU或guest失敗。正式.i64未改，不作新IDA或無關CPU全套；原CSV／LOG／PNG／RAM／journal／state及private probe／getter不入Git。

以下29份位於workplace/dosgolem/workplace/，不入公開Git，含兩份重用的getter來源：

| 檔案 | SHA-256 |
|---|---|
| moo2-380-overlay-frame-extended-180000000.png | beb773bf0623f56bc9b2be697c6e0e472abebd8fa4c319f15e6074291bb7a832 |
| moo2-colonies-row-380.go | cd0e0ff62e9abc30bee6aae02d47801dd55c3d3f229337c3dbf2e87228d4b27c |
| moo2-probe-380-overlay.txt.gz | b8588d8984adede9b706a0b4dc9ccbc3b9d299960cc95624bde2eb9b49979fe3 |
| moo2-save-state-380-mox.set | de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f |
| moo2-save-state-380-save10.gam | 0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d |
| moo2-save-state-380.json | 50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705 |
| moo2-vbe-380-overlay.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| new-game-380-cli-tests.txt | dc58311eac04c67975f95465bcd93ae3a16168772ef171dd261e2266157272ce |
| new-game-380-patches.json | 5467aef830a138c09b8c1cb7f500112d4ab700a6ac77a716e5cd34dc7a55a35a |
| new-game-380-ready-review-tests.txt | a0dfd9d12c8b8a33cb958d6cae0c229e1e015d07fa92f4b12603e1cc5f2bf49e |
| new-game-380-ready-review.py | a6f1cf48807088a921938f296c2c6c08a75985514d76deb804966a0dd1575899 |
| new-game-380-run-output.txt | fc3e77e860916ae74083b4666effba6dfcbd4bf178ea0dfc6792285be7b08d9a |
| new-game-380-run.sh | 09f343ee3cdcd76910da893f4ae4b81825e92651480aca1f8cccfb95eeea5752 |
| new-game-380-state-capture-output.txt | 4e9e700c2ba93e7b5419aa50d201f6f01e55e4e39ed41951f50610daa299cd4e |
| new-game-380-state-capture.json | 5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d |
| new-game-380-state-capture.py | fcd66b36f2c23ac1606b5e7f0b3d816f39ac6c2946b85f0d2b8d4ba5bc9204ce |
| new-game-380-source-tests.txt | 90b2da3d6fd61bd9fc05d78c5f330f7625bed00fecb6c6ae29cae45aeb5748a8 |
| new-game-380-source-verify.py | 2838783bdfbf6907cfa631ce87c0d263a0dc488be8b28a0a8c2469137809befc |
| new-game-380-verification.txt | 86f6edb7f1aca012ea2145b02e49a210392e585aaaf6a0bba0b5ca5a89bc92b1 |
| new-game-380-verify.py | 0a527c154697ba031b40c6d4056a26ce1f8dbc388bede569924a809d94031299 |
| new-game-380-baseline-dac-journal.json | f8fd3dacba516528a5877e1db81ac547aaaaeb6679dfe9a64fc4c2127a4e48d0 |
| new-game-380-restore-journal.json | 009108c483e0c490a322d213f56938ab319547f17805c299af3047c47b5b4cab |
| new-game-380-row-baseline.json | 4159e29168b365f6fd9367e0775093d20c01588c2f9576a6e4fb79c7877f7e93 |
| moo2-380-baseline185.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-380-first-restore.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-380-row-baseline195.png | d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba |
| moo2-380-terminal.png | 1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622 |
| moo2-palette-snapshot-375-prototype.txt | 69da24f32ed36f2791689bb05e2f2ffe35aa9167c5cb7d3369728fefea0e3db7 |
| moo2-dac-journal-376-prototype.txt | df69292f30e77a6b5b7c3c550edff42d8775c804cb2c16a5d2a4838d0ddd46e9 |

已證實：本正常行選取13、原RET／writer、完整前置、降色末態與state保持。未知：新殖民地正常畫面、人口調整、typed資料語意與正式存讀。下一步為200M正常畫面建立的最小上層來源核對，未宣稱remake玩法同狀態或整款完成。

## 2026-10-04：381 同200M原框架與直接呼叫來源

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。工具來源5d3f5b80373c4872e1401366d3b5dd482577252a，新工具6e3dc0cb0e6019884bc3540a6c10598da2e1248c與繁體用字訂正a03c322d28002bb0e11d5d5109d5833a17114d04已推送github，見[381限定規格](https://github.com/wicanr2/dosgolem/blob/6e3dc0cb0e6019884bc3540a6c10598da2e1248c/docs/spec/381-moo2-colonies-frame-source.md)。公開只有新規格、索引及380回填，全部CPU／internal／DOS／原probe保持；原資料、RAM／journal、PNG、state及private probe留在忽略workplace。

### 原定位與分級

- **已證實，靜態與原bytes**：IDA linear EA sub_1338C9為1338C9..133BAE／214項；原dosgolem_high_le:223A23對應IDA 133A23、bytes0345A8、file offset1806455。原4push後sub ESP,260h，尾端mov ESP,EBP、4pop、RET。原EBP2BDB68−ESP2BD908=260h，返回槽為EBP+10h，即2BDB78；ESP首值21A6F3在區域空間。21A6F3映射IDA 12A6F3但前call12A6EE指向sub_14852C，不能當目前caller。
- **已證實，原只讀框架**：同200M原SS188:EBP2BDB68的32bytes為`9CDB2B0000002B00576AE730000000001C342200750000000100000000000000`，原返回槽2BDB78=22341C、保存EBP2BDB9C。讀取放在既有前後核心／FPU／DAC／RAM只讀保護內，原cap與全部輸入保持。
- **已證實，靜態CALL**：IDA 133417／file offset1804907／E8AD040000→sub_1338C9，下一13341C對應原22341C；caller為sub_133237，133237..1334BB／184項。**強推論**：當前caller鏈吻合此CALL。**未知**：未觀察本次自然RET或其caller133425的store執行，不把byte_1B2358命名正式欄位。
- 七個直接CALL完整保留原EA／bytes／file offset／xref。1336AD與1336F0沒有IDA函式邊界，顯式unknown；初始查詢對caller數及None邊界的假設已修正並保留first-query。其餘前後frame同一數值2BDB78的322／323／348／349／352是不同生命週期，不外推200M解釋；只有380本輪追加381回填。

### 同200M核對與停止線

381先核對固定官方EXE、IDA schema1／9.4／5365函式、原file offsets、4push／4popRET、七E8相對目標與原380完整200M，才READY及修改private觀察器。兩patch精確逆轉380；mode off無新frame讀取。private首次編譯guard的ports變數不在scope，原guest尚未啟動，改用既有只讀base的virtual_micros後同命令乾淨重跑。

原guest只跑一次，仍step_limit200000000／EIP223A23；34CLI逐byte保持。完整13032原列、35876 DAC事件及所有非新增final欄位逐項保持，只有每輪RAMhash依既有規則排除；全core／FPU／clock／裝置／表／像素不略過。新frame的原槽值由靜態堆疊契約獨立定位，未預填；七直接call唯一命中133417。所有舊PNG逐byte保持，末圖仍黑／count1空表，沒有新CPU stop、DOSexit或新UI。原418來源、SAVE10.GAM／MOX.SET／sound.lbx與同guest副本／UID GID1000保持。兩次驗證器使用同一新收據，第二次加入實際槽值及380回填護欄，沒有原guest重跑。

IDA image `sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`，Go1.24.13 image `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`。各次原ZIP唯讀、/tmp一次性DB、可寫輸出UID/GID1000、network none。idat exit1不當失敗或成功，非空JSON／schema／原hash／所有權才作驗收。正式.i64未改，不追helper內部。

實際命令：
- 120s／2GiB／2CPU／128pids的IDA容器，/patch掛moo2_patch1.31唯讀、/out掛workplace/dosgolem/workplace可寫，以`bash /out/new-game-381-ida-run.sh`及`bash /out/new-game-381-ida-boundaries-run.sh`匯出。每份未驗證查詢由固定官方EXE建一次性DB；失敗原查詢及log保留first-query。
- 30s／512MiB／1CPU／64pids的Go容器，/patch唯讀、/src為clone，執行`python3 workplace/new-game-381-ready-review.py`通過後才由`new-game-381-generate.py`建立private probe。
- 600s／2GiB／2CPU／128pids的Go容器，原ZIP及patch唯讀、/src為clone，執行`bash workplace/new-game-381-run.sh > workplace/new-game-381-run-output.txt 2>&1`；原PID監測550s與trap，沒有獨立第二guest。
- 90s／1536MiB／1CPU／128pids的Go容器，執行`python3 workplace/new-game-381-verify.py > workplace/new-game-381-verification.txt 2>&1`。完整DAC獨立重播、PNG filter／CRC／palette映色沿已驗證的380流程，再對同200M全部紀錄與新frame作獨立核對。
- Git差異／版權／UID與root-owned核對通過，Go與IDA相關容器為空，原root-owned2437檔／272目錄保持，無.md目錄。首次工具push審核拒絕後，已核實clone及原來源GitHub網址、唯讀遠端分支hash、三份自撰公開文件範圍；相同push命令通過審核及完成，不推本機origin。

下一步：只查IDA linear EA sub_133237的直接上層CALL／返回邊界與正常畫面建立入口，使用同200M原SS188:2BDB78=22341C及保存EBP2BDB9C作錨；來源充分後才決定一次有界畫面完成觀察，不延伸palette／renderer／DAC／PIT helper或盲增cap。主庫RE-first保持，殖民地正常畫面、人口調整、正式存讀、完整開局、RNG及remake同狀態未知，固定日期不是seed。

### 381 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，不公開原log、RAW／PNG／RAM／journal／state或私有工具。全部收據UID/GID1000；39個正常沿途frame另逐byte核對380，個別PNG不重複列於下表。

| 收據 | SHA-256 |
|---|---|
| `moo2-381-baseline185.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-381-first-restore.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-381-ida-frame-boundaries.first-query.log` | `081b9a5cdf748ae28681251b58a897fc1b7111973d8478669e70f209740201c8` |
| `moo2-381-ida-frame-boundaries.first-query.py` | `06221d1efd230188af83aeca4ccd910be8f36c793abd9961485cad6343ac7ee7` |
| `moo2-381-ida-frame-boundaries.first-query.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-381-ida-frame-boundaries.json` | `bd5f847debd1554f9530b56ba9505dea5619bd6e7f10f94e2f582efdf6ecb08c` |
| `moo2-381-ida-frame-boundaries.log` | `9eeba38a14dde1dfbe39cd4eb70d07d62cce05751d2eaf7ba31ded9ca4b39dbd` |
| `moo2-381-ida-frame-boundaries.py` | `29c67882b314f78cfa95f8ce1eae6aa5913cd8a013cc77c5223a4e86ca2f1d71` |
| `moo2-381-ida-frame-boundaries.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-381-ida-screen-source.first-query.log` | `e03119bec7f91c7c12fce0643aae1310d27ebed4a06432a2b5667ed675ec9e0d` |
| `moo2-381-ida-screen-source.first-query.py` | `e4701a003d0327b62fd2eb982ea7cf2fa457e74bd90a8e0a985e545be96c416a` |
| `moo2-381-ida-screen-source.first-query.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-381-ida-screen-source.json` | `0379993cb514f54e1d0a0e98bf2a5c74ffbe185bb371990a926270741d7ffbdd` |
| `moo2-381-ida-screen-source.log` | `b44bf8f8943c7bcd7e30c1937afe2d9afefbb410184c4a8efd7c912a7690e5af` |
| `moo2-381-ida-screen-source.py` | `51e63313009200a1b1bc3a5913a595306d43f13e0cae36b6a1f3ae6155bc3d02` |
| `moo2-381-ida-screen-source.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-381-row-baseline195.png` | `d3c775f1b8594e8313c27b5bd9e363dfb6e3210ca05f0f17136b36409b7a67ba` |
| `moo2-381-terminal.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-colonies-frame-381.go` | `abc0eed02e96858b1401e52b5fb75a76398379199479aae1973c3af35d00a636` |
| `moo2-colonies-frame-381.go.first-build` | `3dc75657afc322da26edcfe845696007ef21ba4c50628cc035eb403858bfb268` |
| `moo2-probe-381-overlay.txt.gz` | `04990c03b4f2a963d02415269fab5fb3764bcd9003a1306859287bd9e2b8bb9e` |
| `moo2-save-state-381-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-381-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-381.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-vbe-381-overlay.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `new-game-381-baseline-dac-journal.json` | `0f9d0b648f46882252d607612f61c5780951301c31658c256657b16317ad309f` |
| `new-game-381-cli-tests.txt` | `dc58311eac04c67975f95465bcd93ae3a16168772ef171dd261e2266157272ce` |
| `new-game-381-generate.py` | `768ceb13c711124347ce4da1afee99ec43ae9c4d0c6268344f55f7068c33b2ba` |
| `new-game-381-ida-boundaries-run.sh` | `368cb055e3cdcee60c7e75b0e320f3ea00a19f3054ad64ab35f06151aaff74a8` |
| `new-game-381-ida-run.sh` | `4810373838c8bb572d199ef5233412dfad7c50a9b6a5778783c3c0865f991376` |
| `new-game-381-patches.json` | `db9ce44d5f1a4bbfb1a2bb7851414535614b9f67e2ef73613c08ab58f31e972f` |
| `new-game-381-patches.json.first-build` | `7685813e7d94d8d56b79f0963c74ffc2cf5a7aa0cd0a7ef6ec8ba9c3e09a41b6` |
| `new-game-381-ready-review-tests.txt` | `e2c90a4ec88ebd3767bef89bede866e092fa0a7f54ac7b9db62b7d9e8b43248d` |
| `new-game-381-ready-review.py` | `7358fe92d90cb090a8472652774e74fb363cb0f57ea6ce789bc880c5feabf145` |
| `new-game-381-restore-journal.json` | `1bcb2e8890eff4e9fb7b8cb9acf7a3c7c65ce58f76263fdc92c7238ffc0eb51d` |
| `new-game-381-row-baseline.json` | `7a5e48ae0863996ea9a60f17293dbf67470d3927072d23eff90aec04c361a23c` |
| `new-game-381-run-output.txt` | `21ada60b6b243f9e36e233b8bc1eabec8f1957f8153d51c345bd7b61a77f0375` |
| `new-game-381-run-output.txt.first-build` | `a3f85d0d0b47669232596e227c903a5e087bfbdb4011b847f1b375a5e15db19f` |
| `new-game-381-run.sh` | `84969e8c97eb0415a66dcb5c16dbd6701dda23106178086ee68ccfe5faadc1c0` |
| `new-game-381-source-tests.txt` | `dfbbe997427798604f4a6a5ba10e07ad0ceff202a8ea94986b2151f120970d9a` |
| `new-game-381-source-verify.py` | `46153a2dd5a0b41b503a0fecea0b4e79c045b3c633ee4f4d53c00cbe75900196` |
| `new-game-381-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-381-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-381-state-capture.py` | `f0db3ea43987038ca0c7a2e4065dd2d4e2e5bf1d40f1ee0e0df3e625f24122b4` |
| `new-game-381-verification.txt` | `1965648ec736484bab9039fefaabc817160ab683751a685c8bb4ab8515682a37` |
| `new-game-381-verify.py` | `7463bf4e7fa48bbcb8c0dd315621458ea55a7f240090bc730a38c701a90db119` |

## 2026-10-04：382原父返回槽、自然RET與殖民地畫面

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。工具基線a03c322d28002bb0e11d5d5109d5833a17114d04，新工具1929523731e5f1af2c1bbb446cdfd89e28401af4已推送github，見[382限定規格](https://github.com/wicanr2/dosgolem/blob/1929523731e5f1af2c1bbb446cdfd89e28401af4/docs/spec/382-moo2-colonies-upper-continuation.md)。原EXE、LOG／journal／PNG／RAM／state與private probe維持本機忽略workplace；公開只有四份自撰來源與回鏈文件，CPU／internal／DOS／原probe不改。主庫RE-first閘門保持。

### 原定位與分級

- **已證實，靜態及原bytes**：IDA Pro9.4／IDA linear EA sub_133237為133237..1334BB／184項。133277比較EBP-20h與100h、133288增；133294比較EBP-1Ch與10h、1332A2增，256×16有界。1334B2／89EC及6pop後1334BA／C3退出，返回槽為父EBP+18h。26個直接E8及原file offset由官方EXE獨立核對，不深入palette演算法。
- **已證實，原200M只讀框架**：dosgolem_high_le原SS188:2BDB9C父框架32bytes為`C8DB2B0000002B00576AE730FF00000000000000000400008E071B0000000000`，原EBP-20h的raw117／1／0保持381。原槽2BDBB4=1B078E唯一吻合IDA C0789／file offset1334749／E8A92A0700→sub_133237，下一C078E映射1B078E，直接上層為sub_C058A／C058A..C0965／215項。按實際原槽定位，不以COLONY.LBX字串選caller。
- **已證實，原自然RET**：203011820／463667422µs，原2234BA C3讀真stack `8E071B00000000000000000000FFFFFF`，返回1B078E、ESP2BDBB4→2BDBB8，其他R、段、flags246h、FPU127F／status0／depth0／8bits及RAM保持，clock增1µs；callback18／18、IRQ54032／54032不活躍。此收據證明當前父自然返回；未把所有共用上層或child每次返回外推成已驗。
- **已證實，原可見畫面**：205281523／222F4F首DAC非0為1，RGB仍0。實際210000000 step_limit／228DF6／483821442µs，沒有CPU拒絕或DOSexit；終圖親看Colony of Sol II、Pop8,000k、職業列與殖民地地景。raw DAC756非0、indexed288944非0、RGB681553非0；RGB SHA-256 `1135b9bc7686966d5fdf0992aa02e35cf73cc0a82bc20ce660d864eac78f7364`，PNG SHA-256 `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03`。VBE bank4／startY0／banksets3083／writes58605678／displaysets98，callback18／18與IRQ56102／56102不活躍。
- **已證實，raw物件表**：DS188:298848 count36／bias0／stride55完整1980bytes，SHA-256 `5a6102f64d586c1978a37d8c5b034caf99783c44ba8ca5eb96896780c8b334db`。UI出現不證明操作全可用。**未知**：人口調整與typed正式狀態寫入、保存／讀取語意、完整開局、RNG與remake同狀態；固定日期不是seed。

### 同來源、驗證及停止線

382 READY之前固定26直接CALL、原bytes／file offsets、有界退出及381完整200M前置已核對。九private patch可逆回381；mode off保持200M與原34CLI，mode on明示一次210M窗口，不增加輸入。46CLI為38拒絕與8正對照，原guest只有一次。原200M完整核心／FPU／clock／表／DAC／PNG保持；12962列至舊收尾前保持，35876 DAC事件前綴逐項保持。新49246事件獨立從原DAC重播，ports counts336／82037／246111與末態逐項核對；PNG filter／CRC／palette／histogram獨立核對，包括首DAC非0仍黑。原418來源、SAVE10.GAM／MOX.SET／sound.lbx及同guest副本保持，UID/GID1000。

首次驗證器誤找不存在的200M color_source標記，改用原restore_terminal；第二次漏沿381既有row journal封裝hash正規化。真正195M核心／FPU／clock／table／DAC／PNG與journal內容已獨立核對，才正規化含每輪RAMhash的封裝hash，沒有忽略核心或像素。沿同收據通過完整核對及380／381追加回鏈護欄。原guest、收據、輸入與預算未改；第二失敗輸出與腳本保存second-check，本節保留首次失敗原因。

實際容器命令沿381同版本及掛載；IDA image `sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`，Go1.24.13 image `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`，network none／UID1000：
- IDA120s／2GiB／2CPU／128pids：`bash /out/new-game-382-ida-run.sh`，官方patch唯讀、tmp一次性DB；schema1／5365函式／184項／hash與UID通過，idat exit1不作判準。
- 來源審查30s／512MiB／1CPU／64pids：`python3 workplace/new-game-382-ready-review.py`；通過后生成private probe。
- 原guest600s／2GiB／2CPU／128pids，owned PID550s／trap：`bash workplace/new-game-382-run.sh > workplace/new-game-382-run-output.txt 2>&1`。session34285 exit0，實際guest step_limit210M獨立紀錄。
- 驗證90s／1536MiB／1CPU／128pids：`python3 workplace/new-game-382-verify.py > workplace/new-game-382-verification.txt 2>&1`，同一收據，末次只新增回鏈護欄。
- 自撰文件及版權輸入／UID差异檢查通過；一次性本項目容器結束，主库收尾核對root-owned基线，不清理共享主機其他項目。

下一步保持210M原來源，核對36筆物件表的職業列熱區、原輸入消費端與callback安全前置。來源充分后才订人口操作觀察契约，不盲增cap或深入renderer／palette／DAC／PIT helper。

### 382 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，不公開原log、RAW／PNG／RAM／journal／state或私有工具。全部UID/GID1000；39個正常沿途frame及舊200M PNG另逐byte核對，不重複列於下表。

| 收據 | SHA-256 |
|---|---|
| `moo2-382-ida-upper-source.json` | `cdcb10ddd0fdcc31da4ef8229134ce3b0fc77fb9d52a784cb11c00a21bc0543b` |
| `moo2-382-ida-upper-source.py` | `603a738627f69362e4ae4a180bd736a7c19e19917dc33394930515225de47ffb` |
| `new-game-382-ida-run.sh` | `182362154ccf863118bd8d7c8497115d01927133076033283c95b89f05523dae` |
| `new-game-382-ready-review.py` | `cf9c1240b429bff87c5b565dcd4a3e3cedf1dbcf9be14eb0c003855114e63c46` |
| `new-game-382-ready-review-tests.txt` | `8638cd8b0ac189626d3a913feb792080227fc03e9618a81964a50fa2b4858049` |
| `new-game-382-generate.py` | `49dcc40c70b8d584d33aaedcf5ff4e840c0b4836f56499dad76d37fc18f848e4` |
| `new-game-382-patches.json` | `7b3066f623bf334abca0cd2595c782861f98aebfd4865f1b0f4b00b5e6006dc9` |
| `moo2-colonies-upper-382.go` | `a9e16e7e8e0cd1f0c779b4981789aeb0930de472865eba18cb6d10ce77882d96` |
| `new-game-382-run.sh` | `0c5d62cc1f2ab9366758a956db05c6de1c85b47a4af342ec3a2af0c92d6f80eb` |
| `new-game-382-run-output.txt` | `5857f3cd75df05d395ae20120fe3a96d2b7009235b14d13a39cf759c31dfe52c` |
| `new-game-382-source-verify.py` | `487ddbe18fa652344b48a5e62355e05368eb7f087d5c99d17da95b636f58369c` |
| `new-game-382-source-tests.txt` | `e0b971a57158014ca8130431a62afb8febd4a12d546518f80a721667edfc3783` |
| `new-game-382-cli-tests.txt` | `2476578bcff4f3ef37d6caa1f76152a3f7a489e1b4761b1958ed61ccb7f03514` |
| `new-game-382-verify.py` | `bdebe1cef521cc76282eb8613c43555ef11abad911ae904ca5f0eff86f4bdb86` |
| `new-game-382-verification.txt` | `34a71953a87aa6ee37ce3a3204605fe05a4d9d071aa5760ba8ab868c810e2aca` |
| `new-game-382-verification.second-check.txt` | `a3e5401af03ed4647450b4ae8429ec993addefda4dd9a798a77b3e4265c33066` |
| `new-game-382-verify.second-check.py` | `c1a35efc9690e9df3e1f5588fc5a9929e914cdb7ce94f15b1c318b37e81ddee8` |
| `moo2-probe-382-overlay.txt.gz` | `fb6da45a85e9f431f34fc5b016e253e59bdace493bc7cc067bc3ace8f8bd2416` |
| `new-game-382-restore-journal.json` | `cfffc3f5d60bd04fb3807f256cb4998da32a92cfa2669b3e328e36192d3ffacf` |
| `new-game-382-upper-source200.json` | `0ee80b04d3d5310f9a425c051130bde5f488dc0a3e3529d02aa17b5556738ab4` |
| `new-game-382-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-382-row-baseline.json` | `47b9d059620730767e26c6a66765c29676ca69b193033473600cf4c6dfd18bb4` |
| `moo2-382-terminal.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-382-upper-source200.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-382-upper-first-restore.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-save-state-382.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-382-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |

## 2026-10-04：383殖民地職業列輸入來源與原重定位記錄

固定官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；工具基線1929523731e5f1af2c1bbb446cdfd89e28401af4，已推送工具b331bb640e4932dc59f5e4da694c63530132f4c1。完整來源與分級見[383限定來源文件](https://github.com/wicanr2/dosgolem/blob/b331bb640e4932dc59f5e4da694c63530132f4c1/docs/spec/383-moo2-colonies-job-control-source.md)，由工具索引與382回鏈進入。主庫玩法閘門保持，沒有新guest、輸入、預算或公開CPU改動。

### 原定位與證據等級

- **已證實，原表及靜態來源**：原210M DS188:298848的36筆／stride55／1980bytes保持，SHA-256 `5a6102f64d586c1978a37d8c5b034caf99783c44ba8ca5eb96896780c8b334db`。index1／2／3均kind6，矩形依次310,62..518,92／310,92..518,122／310,122..518,152，+20h pointers2879DA／2879DC／2879DE。sub_11CEF5含端點、從前列first-match，330,92先1、330,122先2；16邊界模型通過。沒有新原滑鼠輸入，不能稱原index1選取成功。
- **已證實，IDA linear EA**：sub_115478在115560／66894218存word +18h、11557B存word +1Ch、11561C存kind6、115638存pointer +20h。+18h word310與+1Ch word510是原上下界；不把+18h完整dword280136當pointer，高word40保持raw。11B0B8／0F8604120000到11C2C2，狀態1時11C2CF／E80E94FFFF到sub_1156E2；115982讀+20h，115988／668902間接寫word。8水平模型通過，原pointer值及持按時序未知。
- **已證實，靜態場景回呼**：sub_C058A的C07D2傳sub_BED21，C07E1呼叫sub_1191CA；1191EA寫dword_1A8840並啟用word_17C48C，1192F3間接呼叫。11E1F1／11E33B／11E508的kind6持按分支使用此回呼。**強推論**：BF627、B4EF6 mode3／4與B9C3D／B9E94為原人口選取／放置鏈；目前210M實際callback、mode及正式人口變更尚未觀察。
- **已證實，最小用途分類**：BC928只清word_17AAB9；B9CE3比較packed值而不寫人口；BB1FA重設游標狀態，三者不命名job writer。B9C3D在B9CAF對原pool+169h×colony+4×slot+0Dh執行AND FDh，B9E94含bitmap／record寫入及轉移分支。只保留人口觀察候選，不深挖無關轉移，不宣稱當前同殖民地換職成功。

### bytes、原地址空間與核對

IDA Pro9.4使用IDA linear EA；原檔另存file offset，執行器另用dosgolem_high_le。原EXE內嵌MZ26654／LE292E4，2objects基址10000h與170000h、365pages及51363筆原fixup record由原file bytes逐筆核對。既有Go InspectLEInMZ僅讀typed header，Python再獨立核對原header／object／page／fixup offsets、signed source offset與target，重建已重定位code bytes。4349筆紀錄／3578原EA通過，其中674筆bytes差異均有實際fixup覆蓋。私有bytes index逐項保存原file bytes、IDA bytes、raw relocation record、record file offset與原target；沒有推測性改名。

首次錯誤模型把全部IDA bytes當磁碟bytes，BF80E原IDA `803DC8AA170000`與file offset1330786的`803DC8AA000000`不同；原object2基址170000h解釋實際fixup。初版LE入口誤讀外層MZ，沿既有le_machine_test.go中的26654入口修正。失敗腳本／輸出保存first-check／first-layout；同原來源通過，未修改原EXE或收據。空switch查詢只表示此段是CMP分派樹。

末次來源驗證六項PASS：原LE records；8份IDA schema／hash／兩種bytes；kind6比較樹／欄位／場景；原210M完整核心／FPU／clock／callback IRQ／36表／PNG及24模型；public code不變／無新guest；383限定狀態／382正文與回鏈／索引。原382 journal SHA-256 `cfffc3f5d60bd04fb3807f256cb4998da32a92cfa2669b3e328e36192d3ffacf`保持。此次CONFORMED僅來源與舊收據，正常人口操作未知。

### 工具與下一個最小動作

八份IDA查詢的實際啟動命令見WORKLOG本輪，IDA image `sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`，每次120s／2GiB／2CPU／128pids。Go1.24.13 image `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；`go run workplace/moo2-383-le-inspect.go`與`python3 workplace/new-game-383-source-verify.py`只讀固定來源，沒有guest。全部Docker network none／UID/GID1000，原patch唯讀。有效非空IDA JSON、schema1／5365函式／hash及ownership核對通過，idat exit1不當產品失敗。

下一步建立384只讀觀察契約，維持同輸入與210M完整核心／畫面守衛，補讀三個pointer值、原current colony／pool／完整361byte record與場景callback。取得可重播前置後才訂一次正常人口選取／放置；不盲增cap或追renderer／palette／DAC／PIT helper。固定日期不是RNG seed，正式存讀、完整開局與remake同狀態仍未知。

### 383 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`。以下50份全部UID/GID1000；原JSON／LOG／bytes／state與私有腳本不公開。原382 PNG沒有重生。

| 收據 | SHA-256 |
|---|---|
| `moo2-383-ida-colony-consumers.json` | `5dbfc0b0af389afe10bab0a4ad4d6094dd13588f26d054cedb0f295f7316887e` |
| `moo2-383-ida-colony-consumers.log` | `1574c4442dc1d58dc5e25d025402a33ff831f3b4bdfead6872e9b5a8a188a73f` |
| `moo2-383-ida-colony-consumers.py` | `170664c48483fc3a232dae761fa87cebeaed547666288596bdc6b604f1e88378` |
| `moo2-383-ida-colony-consumers.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-colony-input.json` | `8013da440d64e8c944152236c477199433250a4c521b2e9759366b3d493bf3d0` |
| `moo2-383-ida-colony-input.log` | `e68b12be2383bf8501016ad382113cbd52c1b473afd5b3c44c22cd99813bcdb6` |
| `moo2-383-ida-colony-input.py` | `5731ca220980c81e1077dde38b5b2bca00caf2923dd60b7045891e9de2719842` |
| `moo2-383-ida-colony-input.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-input-ownership.json` | `1603c05183d35fb179a2a17ce1854acd9463b847bd1ffeda948f1e6bdebf5988` |
| `moo2-383-ida-input-ownership.log` | `87919c50ee070d175636a4cf69860239880359c96518ba5698ba87d6e5bfe0a9` |
| `moo2-383-ida-input-ownership.py` | `acf195741bc384bd793c79d03b797a4753a975e236b30b7f7001488586cfe10d` |
| `moo2-383-ida-input-ownership.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-job-mutation.json` | `fa02fe1acd0990af66b2969da7fcb3f3f06b079774e4eb551a999bbe91a24c4d` |
| `moo2-383-ida-job-mutation.log` | `12791c737e5f819a8257e3dd8caeb65512b993da49590f3a30171c4ff54260c1` |
| `moo2-383-ida-job-mutation.py` | `1ba0d8ed219e8444756b9357c0f78d4abb23a67e1edb20a7cfc2a0bac7894776` |
| `moo2-383-ida-job-mutation.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-pop-contract.json` | `391a3e91c551efcecc3dbfd1497e6c777673c9f9213ef541429caf6d2c928479` |
| `moo2-383-ida-pop-contract.log` | `dee9cdd9c7f15054252cf1a7d08561db13958eebe0875567a99f92ba1b03ce7c` |
| `moo2-383-ida-pop-contract.py` | `c5dcd9ff3701ec38437e060a7adbe219ea786166c03ce9763af4c52d0d04a6a3` |
| `moo2-383-ida-pop-contract.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-pop-pick.json` | `fa59f0b7ce497ecc93ddac1a9996072c595b4e67f6e99f03191cbdf109db8719` |
| `moo2-383-ida-pop-pick.log` | `25edc31685339ed3b300d14c70bc33b6f7679807ec1e8c152f10f59a282927bf` |
| `moo2-383-ida-pop-pick.py` | `66499c07b98a6377bede6ce181293b10eb58d13a13a1ffb8b15e9ba937958f56` |
| `moo2-383-ida-pop-pick.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-type6-dispatch.json` | `5c237411a5bfd0c3302d8de287e3a883070c5bdb9f2f28e3234ec1e380b3dd65` |
| `moo2-383-ida-type6-dispatch.log` | `25623b8f3c1156c0c470dc2d476a1b8035750aaa3827ee319b14bcb26fcfb4a7` |
| `moo2-383-ida-type6-dispatch.py` | `c59d6097ff195342856d989b6d145bebaea8ef1f091e9d3c3a2645d3bd3bd7ec` |
| `moo2-383-ida-type6-dispatch.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-ida-worker-writes.json` | `e63ce45805876fc191e88ba65edaf6034d258f05394beb07dacdb94c4ba665e4` |
| `moo2-383-ida-worker-writes.log` | `e18bf9777e5b22f2c1d6e474c4a4809ff164b6e4c21f45c5414e22864bc7531d` |
| `moo2-383-ida-worker-writes.py` | `45a10c4b08f63a96fbcabeb321826dd2217d989b45967b141ec1bd4457190be4` |
| `moo2-383-ida-worker-writes.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-383-le-inspect.first-layout.go` | `d6169c1bcbd1a53685cec97c887e7dc1934e649519a18e9a88e64a5f240aab46` |
| `moo2-383-le-inspect.go` | `9c7803100c86174b6ddf41d9337838794c220048789b75f4f063052e1642483e` |
| `moo2-383-original-le-fixups.json` | `861f46cdef3ab27dd088a72cb0c6b2b1d62ee35285cd23c1995ef24a28602aa4` |
| `moo2-383-source-byte-index.json` | `6d98888ead1909383ae3b7322f2f7518527c802d0800059ca93f3ce4f51495d9` |
| `new-game-383-ida-consumers-run.sh` | `167c21982f5ab3bec8c9f6fc71944e96c845f971fd2f9d6d22c4476b82ad3b8a` |
| `new-game-383-ida-contract-run.sh` | `144b6a6501ac829e287ba0ce41f7360155f23ed7988e7263f808a0cf7633b30b` |
| `new-game-383-ida-dispatch-run.sh` | `e60b71a84d854fabc06a76f462de0aa7988ec09dcee608315c35bde46e4a9cb9` |
| `new-game-383-ida-mutation-run.sh` | `e3a8c2cfdd6c598d900918f4ea5045869c1e5d7ec02c38c860490871497dc62d` |
| `new-game-383-ida-ownership-run.sh` | `d93a3fe7f78135617111dd442dbb7331666bcc45d85802eacb0e8fa3ac732af4` |
| `new-game-383-ida-pick-run.sh` | `ec4ac92d4f913f3a976ad0f6f0fa94b5fa6dc1367fd425791335738c45f04ed6` |
| `new-game-383-ida-run.sh` | `d3069febd5d654e30584d7da1cb3a1cc4a2556dfe8cd23e0f102044884ccad03` |
| `new-game-383-ida-worker-run.sh` | `eee0b368d23bff432291282cc082e651bee6db220d4bb4575950c850bb343334` |
| `new-game-383-le-inspect.first-layout.txt` | `a46b432f80f42dace5e473cec4f0d87a0821663faed876edc32064a4876b87ff` |
| `new-game-383-le-inspect.txt` | `ed9acf3368915a38b3853aa31ae6d7c8ce578cd6dfe65f83fcab4438f3a55dcf` |
| `new-game-383-source-tests.first-check.txt` | `ec20fcc14736f121b86a6effe074971c5da2fc097fa15b79b2ddd1899e8f0cd5` |
| `new-game-383-source-tests.txt` | `91b0f2aa5f3df89b6925a404a89d869660562f3108ad168b11edca232f5e5fd9` |
| `new-game-383-source-verify.first-check.py` | `8d880eac635a4df6dfc48a13f3260a407a7bb22681cfe623dfba95e0e06d032c` |
| `new-game-383-source-verify.py` | `21b33c85343f309f9c169c90d11d28f639da4dbc752ba35b091f47a505a37e36` |

## 2026-10-04：384原210M人口控制前置快照

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；工具基線b331bb640e4932dc59f5e4da694c63530132f4c1，新工具0a6c7f97c75262c1d36098f5d6763918e91e2305已推送github。見[384限定只讀契約](https://github.com/wicanr2/dosgolem/blob/0a6c7f97c75262c1d36098f5d6763918e91e2305/docs/spec/384-moo2-colonies-job-source-snapshot.md)，工具索引及383回鏈均可進入；主庫玩法閘門保持。

### 已驗原raw前置

- **已證實，dosgolem_high_le／DS188**：2877A8原dword `04000000`為current raw4，27AB18原pointer `44205B00`為5B2044。按原361byte stride取得record5B25E8，完整bytes SHA-256 `6fff16d8ffc032a510cf5ef745b4d5c862010078ea57a71a48537b450e111da0`。原UI選取13與current raw4各自保存，不當同一typed id。
- **已證實，原三pointer**：2879DA／DC／DE各2bytes皆 `3601`，原word310。2879D4的18bytes為 `030001000100360136013601010002000300`；保持raw，沒有將3／1／1或310命名職務數量。8個原窗口與完整record／三word均可讀，64-bit record算式與descriptor／RAM界限通過，沒有猜typed colony id上限。
- **已證實，原scene欄位**：DS188:2A8840四bytes `00000000`，26C48C兩bytes `0100`，29BE14兩bytes `0000`。enable與完整原globals重疊bytes核對，水平偏移與原header重疊核對；零pointer來自成功讀取，沒有使用fallback。
- **已證實，IDA Pro9.4 linear EA靜態先後**：383原來源中的C07C1／E890ECFFFF先呼叫sub_BF456；C07D2／B821ED0B00送sub_BED21到EAX，C07E1／E8E4890500呼叫sub_1191CA，1191EA／A340881A00寫dword_1A8840。原file offset、file bytes與已套fixup bytes留383 bytes index，不與runtime位移混用。**未知**：本次設置／還原時序、callback為0的原因與安全人口輸入前置；不能由靜態CALL宣稱實際已觸發。

### 同來源與只讀驗收

384先DRAFT→READY，才生成私有觀察器。五patch精確逆回382，mode off保持382，mode on保持同輸入與210M，12新增CLI與原46共58／48拒絕及10正對照。原guest僅一次，session6185 exit0；實際step_limit210M／228DF6／483821442µs，無新輸入或cap變更。快照於既有terminal後捕捉一次，完整terminal與固定382比較，只按既有規則跨run排除per-run RAM hash；本run整個RAM前後hash必須相等。

獨立驗證保持完整382 journal、全部原列、所有舊PNG、36表／原核心／FPU／clock／callback IRQ／DAC及palette、418來源／state與同guest副本。快照before=after、原pointer與重疊窗口及361record界限核對通過。原382 journal SHA-256 `cfffc3f5d60bd04fb3807f256cb4998da32a92cfa2669b3e328e36192d3ffacf`不變；原終PNG SHA-256 `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03`保持。公開internal／CPU／DOS／原probe不改，本輪沒有CPU行為或新IDA。

READY初版把115988／668902間接寫入當成需有relocation，來源審查拒絕後按原指令修正。驗證器首次exec來源檢查覆蓋b，日誌迭代TypeError；隔離namespace後沿同收據完成全套。首腳本／輸出保存first-check。原EXE、定位、guest收據、輸入與預算未改。CONFORMED限定只讀原前置，383舊正文保持並追加不可變定位回填，文件閘門通過；人口正常操作仍未知。

### 工具、停止線與下一步

實際命令及生成方式見WORKLOG本輪。Go1.24.13 image `sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac`；network none／UID1000，原ZIP及patch唯讀，native600s／2GiB／2CPU／128pids，owned guest及capture PID由550s監測／trap收尾。驗證90s／2GiB／1CPU／128pids；來源／文件30s／512MiB／1CPU／64pids。相關容器結束，沒有新image，私有輸出擁有權核對通過。

下一步只核對原sub_C058A的C07C1→sub_BF456返回邊界及C07D2／C07E1回呼設置，保留同210M原callback0前置；來源充分才建立一次有界等待／安全輸入契約。不直接送人口輸入、加cap或深入共享renderer／palette／DAC／PIT helper。固定日期不是seed，正式人口調整／存讀／完整開局與remake同狀態未知。

### 384 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，下表32份全部UID/GID1000。沿途39frames、完整保存的遊戲資料副本與其他舊PNG已逐byte核對，表內只列主要來源、腳本、收據與三個末段PNG；不公開原JSON／LOG／RAM／state或私有程式。

| 收據 | SHA-256 |
|---|---|
| `moo2-384-terminal.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-384-upper-first-restore.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-384-upper-source200.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-colonies-job-source-384.go` | `2a98fbd0b51cac2ff4cc87a4ff828d8aafce14c979a88398fc3ae993ca029bd2` |
| `moo2-probe-384-overlay.txt.gz` | `761fce460fc4ac4df829a44ce2bf791c43355a63b1a6872158cb3bfd87c948ad` |
| `moo2-save-state-384.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-384-baseline-dac-journal.json` | `a0dec3c1afdaa02874f6a4b3745f874114eaa4cb7a0783cace3287fbad692093` |
| `new-game-384-cli-tests.txt` | `991c4159a634cecc94e0798d8292e5883e9a28fd4e3b875850dfc44921fbb753` |
| `new-game-384-document-gate-tests.txt` | `d36630869803bba0311ac3fa52bc2cba16f995b83c50cd8ff111b271f431a4b4` |
| `new-game-384-document-gate.py` | `f3511d6336cfaebf60ab1653ddc24ec7626ce6801b5295d75dd3d2da2aabb4ca` |
| `new-game-384-generate.py` | `61667016caa3dfc4b03a8c64d74282e9efd65543d22b7f8d839a1533239b0965` |
| `new-game-384-job-source.json` | `a16f4c6d194400e1823687ef72eb969bfd7b4bcf8662a04be830877a6efe881d` |
| `new-game-384-patches.json` | `3c6cc1dac1812959c052fcc52fb8f6c2f16f6b34a92f050ce3fbc584c163a517` |
| `new-game-384-ready-review-tests.txt` | `b2592932031114de814da4082149718b95fd7e438c1cf520dac2672f7353eb06` |
| `new-game-384-ready-review.first-check.py` | `3622c8e188b275ebc08f3377562a93211ec5260d6775234d44f5b3f704035cf3` |
| `new-game-384-ready-review.first-check.txt` | `1025953f3a20e0a201aff58cc40b7c9838032319887835f7d6b3d56ae650eedf` |
| `new-game-384-ready-review.py` | `49fecf2b60c1a3da8882371f0a3d295c9e97404bfc7d3be70e43095323c8a6f5` |
| `new-game-384-restore-journal.json` | `d77fb990e1869abcb7fa305fb7b496f77778ea50fc4f146a7d5ec7bce74a99d7` |
| `new-game-384-row-baseline.json` | `8d6aef77dd334ded646e017641b9168e9e6e04e7ecddb61bf1690ce1c820bfbe` |
| `new-game-384-run-output.txt` | `d5ff2ea5ecd5125017419ec9e053d0f7ff3151f8a43935e17cd3f1dfb4a8371a` |
| `new-game-384-run.sh` | `23958d06cb57e46a79e7670c2891e137fb84a6d04f15783acc8d1a3b3b8bdc58` |
| `new-game-384-source-tests.txt` | `7d41fe9ebacde83cf809038e4a45b091f637263ae6c668f90f6e5e316cab4645` |
| `new-game-384-source-verify.py` | `e8a9628f4cee51f3c35e6cb2347e1026111965d30d1d5f043fd3c0c661a540fb` |
| `new-game-384-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-384-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-384-state-capture.py` | `fb78faca20aac783036cc376b60389d9e6e12c76cceda6f6ce9e4bfb01de9cd0` |
| `new-game-384-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-384-upper-source200.json` | `bfee503987c247d57f02d8ed413bc007e7ffe7bd99932685dc118c97a4224ae6` |
| `new-game-384-verification.first-check.txt` | `c01279a9f3fe78ff9d15ef699cef395d8d64d3bb5bad3dc8d57df507ad82203c` |
| `new-game-384-verification.txt` | `dc35e109e8dc5cb2a6aaa062d25fe53d7d917a75ec1cb5d87db25bb470211b8a` |
| `new-game-384-verify.first-check.py` | `277550d54ed0457b52fddf5ab9a5df8ec5dd181b10dc73b280237172f0116da5` |
| `new-game-384-verify.py` | `05b2d85633fc9bcc1a3b3201ba7fcd819289bd79486e431161f2facc0d3a93ba` |

## 2026-10-04：385原返回／首次正常輸入與回呼runtime地址勘誤

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；IDA Pro9.4 linear EA、原file offset及dosgolem_high_le各自標明。工具基線0a6c7f97c75262c1d36098f5d6763918e91e2305，新工具3842529eb5dff704adf3d33b6c4f5720ebe9cbb4已推送github。見[385限定觀察契約](https://github.com/wicanr2/dosgolem/blob/3842529eb5dff704adf3d33b6c4f5720ebe9cbb4/docs/spec/385-moo2-colonies-scene-ready.md)，383／384與索引同步回鏈。主庫玩法閘門保持。

### 來源、原時序與訂正等級

新增一份IDA查詢，327筆來源／325原EA、71筆fixup差異核對原LE51363 records；原來源與重定位bytes保存新bytes index。sub_BF456／C058A／1191CA原範圍及CALL／RET保存，1191CA的205 direct callers只匯出32並明示截斷，不深入renderer或平台helper。

- **已證實，dosgolem_high_le**：原BF456真RET203219421／1AF4FF→1B07C6；原scene CALL203219427／1B07E1，EAX1AED21；原store203219451／2091EA→2091EF；原setup真RET203219467／209225→1B07E6。兩次RET的真SS stack、ESP+4、其他核心／FPU／RAM保持通過。
- **已證實，地址勘誤**：原runtime code `A340882900`指定DS188:298840，符合IDA dword_1A8840加F0000h。384及385觀察器誤用2A8840，偏差10000h。歷史384收據成功讀得2A8840四bytes零的事實不改；撤回「scene callback0」欄位語意。store的其他RAM不變檢查使用錯誤排除範圍，不當CPU缺陷。
- **強推論**：原A3將EAX1AED21寫到298840。**未知**：正確位移實際讀回與全RAM僅改實際store四bytes的證明，舊收據未捕捉的bytes不補寫。
- **已證實，首正常輸入入口**：205804505／virtual471004699µs到1B0845／E861690500，完整原frame及只讀窗口保存；count36／RGB非零681553。PNG SHA-256 `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03`與原210M相同。原current raw4／pool5B2044／361byte record5B25E8、三word310及其他正確定位欄位保持。先前「尚需等待setup」推論被本次原時序否定，210M已涵蓋首輸入。

### 驗收範圍及停止線

READY先於private實作；8patch逆回384，70CLI為58拒絕及12正對照。原guest一次，明示220M窗口，無新滑鼠輸入、CPU／DOS／main玩法改動；原210M完整前置／journal／39frames／DAC／PNG／418來源／state與副本保持。獨立驗證session93305 exit0，SAFE EVIDENCE PASS限定原返回與首輸入，ADDRESS MODEL REJECTED明示地址模型失敗；385契約回DRAFT。原Go／JSON／PNG／journal不修改、不重跑求綠，383／384保留正文追加勘誤，現況表移除錯誤callback0斷言。

首次驗證重新序列化Go事件造成key順序及HTML escape雜湊差異，改核對原journal事件文字切片，同一收據通過；首腳本與失敗輸出保留。其他生成器唯一定位與唯讀mount問題、命令及資源限額見WORKLOG。所有工作在既有Go1.24.13及IDA9.4鎖版Docker內，network none／UID1000，原資料唯讀，程序有界及trap；沒有新image或遺留本輪容器。

下一步386先依原A3運算元修正DS188:298840只讀觀察與守衛，再在原205804505首輸入完整狀態驗正確callback讀值，維持原輸入與210M。人口輸入的座標契約仍待證據，不猜倍率、不盲增cap、不深入renderer／palette／DAC／PIT。固定日期不是RNG seed；正常人口變更、正式存讀、完整開局及remake同狀態仍未知。

### 385 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，下表45份UID/GID1000。只列主要來源、腳本、收據與五個末段PNG；完整39frames及遊戲副本已獨立核對。原資料／收據／private Go不公開。

| 收據 | SHA-256 |
|---|---|
| `moo2-385-ida-scene-ready.json` | `12aea3957cdb776dbdeb41e65f58b1a0213179528d8d48d6dbe55c04d367b159` |
| `moo2-385-ida-scene-ready.log` | `0e19178e36367d87358fd5ec1e406ed1a4a6e27030cd428da636efcfcb9bc460` |
| `moo2-385-ida-scene-ready.py` | `df29ffe1655e1731af0d90814548cdbc0bab7d75051597adfcab9d95a210881a` |
| `moo2-385-ida-scene-ready.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-385-scene-input.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-385-scene-source210.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-385-source-byte-index.json` | `6518e6ee91ea85407d37b9879c0cea9b1a847fc4f3d7baeb49d0d5240102f58d` |
| `moo2-385-terminal.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-385-upper-first-restore.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-385-upper-source200.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-colonies-scene-ready-385.go` | `c228e3757acd0101001bcc2970762380d0441c2205d00b58d535e83b07834d30` |
| `moo2-probe-385-overlay.txt.gz` | `e86363caf6d0adfbcf5caa98b221bb57d293fa7c6b4a853262e198eefabf1029` |
| `moo2-save-state-385.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-385-baseline-dac-journal.json` | `3c4b88a7a732630116aa8eae8d71628de2ce5b9c9a983424c991bf7e0cb66b79` |
| `new-game-385-cli-tests.txt` | `0c1619e2a6ab235eb633ec859665b6d5e0bcc7f7c5851afc5632404202644589` |
| `new-game-385-document-gate-tests.txt` | `de3c1c0825b19b2f0541a2344c2b9bc55996539a58ea354bc08ecaea10028d68` |
| `new-game-385-document-gate.py` | `710ea75eda0d1abb2fa486c6e7bda0c32eced06b6785ecf8bd3a363aae3c2129` |
| `new-game-385-generate.first-check.py` | `dc06687dcd281cc1902c828ebffd56ab9bb461dfe11af459e0befdfa05584fc4` |
| `new-game-385-generate.first-check.txt` | `282b6cfacffaf2f6f1f9f201c7e2048ff14f8fa998bcda9ce89a931d1b6a12c7` |
| `new-game-385-generate.py` | `47f316001322831cdd43b5c96f0c12234241b3d6defd02c044bcabb855c469f8` |
| `new-game-385-ida-run.sh` | `210cb2579c2eac493167135398fae6391561e99b88f13d11170f4616d2bfa4e0` |
| `new-game-385-job-source.json` | `baaad3576197182d38004615f93d00af5b221d4c46ff2b97db9f1cba15103290` |
| `new-game-385-patches.json` | `d136779cb62542128f3a8b9b1e1084e0f97c7fd7778fccbb2b6172d71e872cca` |
| `new-game-385-re-tests.txt` | `1e4b367880cb99ed61c10b2ef7ab90ebf3164c34f415c73531a2d19df65e6763` |
| `new-game-385-re-verify.py` | `182381d42db2e2d9694fdfe762a2db164f1cbbe3dbbc2c5da12b8c474e5031db` |
| `new-game-385-ready-review-tests.txt` | `257fdff3d82d633ee9c8ca45395fb7de658da7b6aefe75a5da24a471df4ba7ad` |
| `new-game-385-ready-review.py` | `5e4b4b49ec10964b25e92886c5c8e9f54e411960fcade1a6e9f9f08cc1ce3d1b` |
| `new-game-385-restore-journal.json` | `ba3072b5c04eb34f46d13adde7cbdc7b47630424d7aae35b2d49a553f70279bb` |
| `new-game-385-row-baseline.json` | `e17d39cba4a507247e0a6b46eae86fa847f364dadd8ee630c441214924125b67` |
| `new-game-385-run-output.txt` | `0f38389ce269dddc1939bc3e4093cf2de770c20fa091cccf809628edaa4883b3` |
| `new-game-385-run.sh` | `f3688b9b6ab0605c2329de2006e80b68ce869890b89f8ed3969168814d79f446` |
| `new-game-385-scene-input-source.json` | `1d863e75ef8afb9d80f756abf394de556a2d808682b1cc60f74504d559175c44` |
| `new-game-385-scene-input.json` | `4e0a112574c9fae4e1e5bb6b27a625a23ca887d07859dfa9dd1ea0abb7412f84` |
| `new-game-385-scene-source210.json` | `07314efca204bdfe67825da8be8d0705724945251072f23eeecefd0e2d27dff0` |
| `new-game-385-source-tests.txt` | `5398a6a98822756cc588bff3a820f83d320198cc7830ab2ab582c3d857e7eab9` |
| `new-game-385-source-verify.py` | `74634c4a2dc76722123cb2c0870579018da6dfc98ee57f80e6847357c16e676e` |
| `new-game-385-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-385-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-385-state-capture.py` | `9c7995624c7c15eb327d0cdaa5598513b57930fe6564d3901ef736a73f0d9758` |
| `new-game-385-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-385-upper-source200.json` | `dbc1d601442f6f5eb5a9efa08b2526ec0c12ad4d65dfea9954ffe509d2fbe37c` |
| `new-game-385-verification.first-check.txt` | `d3457904cbcc4388897f402e81b27bd90e353ead596d35514a5d151a14b482b0` |
| `new-game-385-verification.txt` | `f09541a5e585ab028afb03c953a370efda61c1708ec32421ec40bb53ed3a0105` |
| `new-game-385-verify.first-check.py` | `0dd1589cef2c987948642d877fca2093607604fb4f982bd53afd31ac820ea7c9` |
| `new-game-385-verify.py` | `07e02abddd0ab0622f1c2001288a635e11ba431cae37dc952df710c27d7c8f70` |

## 2026-10-04：386原回呼地址讀回與首輸入前置

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`；工具基線3842529eb5dff704adf3d33b6c4f5720ebe9cbb4，新工具3e2290007d5d0163346b4150c7e8cc625b1e0166已推送github。見[386限定只讀驗收](https://github.com/wicanr2/dosgolem/blob/3e2290007d5d0163346b4150c7e8cc625b1e0166/docs/spec/386-moo2-colonies-callback-read.md)，383–385與索引按同一不可變定位回填；原385錯誤觀察器仍DRAFT。主庫玩法RE閘門保持。

### 已證實範圍與歷史訂正

- **已證實，原A3**：IDA Pro9.4 linear EA 1191EA／A340881A00寫dword_1A8840；dosgolem_high_le原203219451／2091EA執行A340882900指定DS188:298840，符合加F0000h。原file offset／bytes／fixup沿385 bytes index，不混用位址基準。本輪沒有新IDA或原EXE改動。
- **已證實，原讀回與全RAM**：正確地址四bytes由55451700變21ED1A00，raw由174555變1AED21，後值等於原EAX。原Step後EIP2091EF，實際store_linear298840；對Step前整RAM的observer副本只替換這四bytes，再與Step後hash全等，其他RAM保持。observer不代寫guest。原兩個RET及四事件時序保持385。
- **已證實，首輸入前置**：205804505／1B0845完整原385 frame守衛通過，只排除跨run ram_sha256；第9只讀窗口298840為21ED1A00。原8窗口／current raw4／pool5B2044／361byte record5B25E8／三word310保持，before=after與RAM前後hash全等。PNG SHA-256 `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03`與原210M及385首次輸入相同。
- **歷史界線**：384／385的2A8840零raw讀值不改，維持未分類資料；正確首輸入callback讀回及此次store範圍已由未知／強推論升為已證實。不是原384在210M已讀到正確欄位，也不外推其他callback生命週期或人口操作。沒有CPU缺陷證據，原385 Go／journal／PNG與state保持。

### 一次原版與驗收

READY先於11個可逆private patches，mode off精確逆回385，mode on保持原輸入與210M；82CLI為68拒絕與14正對照，原70逐項保持。原guest一次，session78051 exit0，actual_boundary及run_limit210000000／step_limit；完整382／384原210M核心／FPU／clock／callback IRQ／36表／DAC／journal／39frames／PNG／418來源／state及副本保持。獨立驗證session12161 exit0，386 CONFORMED SCOPE PASS只限正確回呼、實際四bytes寫入與首輸入前置，不是remake原版整段同狀態。

兩次生成器縮排定位失敗皆在寫Go前停止，版本與分類摘要保留；回查閘門入口及全部原縮排後生成。沒有guest重跑、改原收據或增加預算。原EXE／JSON／LOG／PNG／RAM／state／private Go留本機忽略目錄，公開只有五份自撰工具文件。來源檢查／CLI、native及驗證命令、固定image及限額見WORKLOG；network none／UID1000／原資料唯讀，程序有界及trap，本輪無新image或遺留容器。

下一步387由原sub_1171AB及kind6消費端補齊座標到職業列熱區／暫存的來源鏈與按下／放開契約；證據足夠才建立一次正常人口選取／放置觀察。不猜倍率或job語意、不盲增cap、不追renderer／DAC／PIT。固定日期不是seed，人口正常操作／正式存讀／完整開局／RNG與remake同狀態仍未知。

### 386 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，下表37份全部UID/GID1000。只列主要腳本、收據及五個末段PNG；原39frames及遊戲資料副本已逐byte核對。私有內容不公開。

| 收據 | SHA-256 |
|---|---|
| `moo2-386-scene-input.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-386-scene-source210.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-386-terminal.png` | `1a0e7551173173b43897be18ccf5295755ca80b222be7dc6dbca0b1d986c8c03` |
| `moo2-386-upper-first-restore.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-386-upper-source200.png` | `1610444d26adb3135e7e933dd44912044bb636728af59c70e614945278d3c622` |
| `moo2-colonies-callback-read-386.go` | `3a050603540f9da42345ca0726b089952177e20cedfeadfd4884e33c92532e92` |
| `moo2-probe-386-overlay.txt.gz` | `5fe3cfd30f2e3e365aec16a8114bce86bb8716cb25f5897dec0f04b0654d6287` |
| `moo2-save-state-386.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-386-baseline-dac-journal.json` | `75cc0d208cc64f339fed430e66dd6b084cb4e01765f1548ee251e33cd92a9556` |
| `new-game-386-cli-tests.txt` | `7f4e01b64d98b60429a5dce665be01241aa98816271edda5669c01d07e96a5f6` |
| `new-game-386-document-gate-tests.txt` | `e42b45ece289adcf2144749256d136a59a9c53d2e6a8f76bea9f735e1f50c14c` |
| `new-game-386-document-gate.py` | `2e5eaf0eeac7a486ec0912feb8057cdc1dd61d9eb38092c314261cbd584fbe9c` |
| `new-game-386-generate.first-check.py` | `fa90858b520553a5d4de15a07e0ffc712ebd65fe6519ad4f4a97773c5ad50a77` |
| `new-game-386-generate.first-check.txt` | `f6376c1c83464f0ee8be34567090351e53b47a3b227db212cb0e854776f33d78` |
| `new-game-386-generate.py` | `ceb34be8dd3bc62e11e8d654706d4d5376d334dde8c6e0de8a84194fac665d10` |
| `new-game-386-generate.second-check.py` | `6b8aa4451fca9e65de3af493dcdc39feeb010aeaa40137c0ec59995704c13568` |
| `new-game-386-generate.second-check.txt` | `51ceb935f679909f1808c53a8ad5261d1ab42acb71794f98f5bbd5ef7a0a7e12` |
| `new-game-386-job-source.json` | `d64da30e8ed57a309883105de5a3c5113a1085c8e6b53ed45fca697d939e341c` |
| `new-game-386-patches.json` | `4905eca7b3bc120ddaefaeca752f914b9d56e0e1a70eb83ed42232c2fbffc9d9` |
| `new-game-386-ready-review-tests.txt` | `f723810a9dca9f092c7017f787097864fb712989f0395020c5ffed7c6df523d4` |
| `new-game-386-ready-review.py` | `bae96bd6274bd2a0746893ef7a2df2e38a8ca942c6c4ce31e033969d7fa567d4` |
| `new-game-386-restore-journal.json` | `26b0c793b2de68c86d61683e755a15ba5dcbc6431f0f99aec3270c71a843ca75` |
| `new-game-386-row-baseline.json` | `ca4bbf80c1c431e4d924ed48279e9d2d618fabf347f7bf66d8328cb2532c543a` |
| `new-game-386-run-output.txt` | `6f93f285ca23cd175ee1c821f880905cf615bd0e8af8d6a78f8c827463fb054d` |
| `new-game-386-run.sh` | `a1206604cc64f39928a7518acaf4fb458eac643ae2e09a3273dcecc17d0e28c5` |
| `new-game-386-scene-input-source.json` | `799c63ee4d5b30cf0c524a856546b8418a9f5e262fba52c97ea89cb72616d540` |
| `new-game-386-scene-input.json` | `aa6728601569f07b6ed5022d0a7fa5c1f5291fa3e18592e40a7e84415c5c715f` |
| `new-game-386-scene-source210.json` | `781a32c6ceaf62edb3b6ce5afc91676bfb1cd84fe2ea72d3e6fd369ee487cd7a` |
| `new-game-386-source-tests.txt` | `89f08500a9027219f59c6bb1ea5b1f9a1de88e3ccd17dfae769819940ae373e0` |
| `new-game-386-source-verify.py` | `78513a6850d732b1222d9d0ca8068086a640d4b68966ed59d593ba156d68108f` |
| `new-game-386-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-386-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-386-state-capture.py` | `717dc8f1500f58a81ce070b3b96dbf0097d619df2d7fdd3d6268eeb0b583f844` |
| `new-game-386-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-386-upper-source200.json` | `57d1a4af24af8a4cae14e1b64feb6fdc161e8b6f8aa9b36cd5568c4eac0204b5` |
| `new-game-386-verification.txt` | `0f7ed558d4ba722459ec7d03593a93165bcc42193c190d721017182d40c577d9` |
| `new-game-386-verify.py` | `3866f17927c48e400b1c1cea578bfa9faed8dddb0935029fc0129c9069221a7f` |

## 2026-10-04：387職業列條件座標、原CB與持按／放開來源

官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`，工具基線3e2290007d5d0163346b4150c7e8cc625b1e0166，新工具1d0d128c52a7de357a41a19b8bcbb2513bc70e46已推送github。見[387限定來源核對](https://github.com/wicanr2/dosgolem/blob/1d0d128c52a7de357a41a19b8bcbb2513bc70e46/docs/spec/387-moo2-colonies-input-coordinates.md)，379／383／386與索引回填；主庫玩法閘門保持。

### 原定位、來源等級與停止線

- **已證實，IDA Pro9.4 linear EA靜態**：原1236D1..1237F3／74項保存ECX及EDX；word_17C51A與+2皆0時，123729 MOVSX、12372D／D1F8 SAR1、12372F存GUI X，123738存Y。1237F2／CB為遠返回，runtime8:2136D1與既有386target一致。未修改CPU／dispatcher，不外推一般CB分支完成。
- **已證實，條件位移**：IDA1B3A38／1B3A36投影runtimeDS188:2A3A38／2A3A36；旗標26C51A／26C51C，width／height26C534／26C538。1B3A3A是區域索引暫存，不當Y。原123D53保存事件座標到1B121C／1B121E，與目前X／Y分開。原123491初始化INT33 range X=2×(width−1)、Y=height−1；不能以初始化清旗標取代首輸入實際取樣。
- **已證實，持按與放開來源**：原124075的AX3按鍵分支AND3；11DB5E非0跳持按入口11E0BF，讀目前X／Y；事件路徑則讀保存的事件座標。受17C4E4控制的11E1A7呼叫113FB9，原signed含端點及index1 first-match保持。11E160 AX3為0後到11E4EB，kind6在11E508呼叫1192D1、11E50D清共享選取。人口操作真正觸發或job變更仍未知。

五個查詢共2562筆rows／2091原EA、514筆file與IDA bytes差異，全沿原LE2objects／365pages／51363 records獨立核對；原file offset／原bytes／relocated bytes／IDA EA／runtime投影各自保存。11CEF5共1550項僅查玩家所需頭尾與切片，不稱整函式已解；caller最多32並明示截斷，直接xref不涵蓋全部間接讀寫。原385／386收據及原EXE不改。

### 只讀驗收與未知前置

READY審查先於只讀模型，13 signed X／3range／6kind6 first-match模型通過，387 SOURCE PASS及backlink通過。660,77→330,77模型先1，只在原旗標及裝置range條件成立時可作候選；本輪沒有guest或新輸入。完整386首輸入／36表／正確callback及public internal／CPU／DOS／原probe保持。

**未知，阻塞人口輸入**：386原首輸入未捕捉26C51A／26C51C、width／height、目前及事件座標／按鍵與17C4E4；不補0、不把事件座標當目前座標或由模型直接送輸入。下一步388沿同輸入與210M、205804505完整首輸入及正確callback守衛，補取raw前置與裝置range，才訂一次正常press／持按消費／安全release契約。固定日期不是seed；人口變更／正式存讀／完整開局／RNG及remake同狀態未驗。

RE verifier初次寫唯讀mount受拒，改可寫工作樹後同來源通過，屬環境問題；未寫原資料或改定位。五次外層IDA exit0，idat exit1另以有效非空JSON／schema／hash／5365函式／UID1000核對。本輪沒有新image、guest或原.i64修改。固定Docker image／有界資源及實際命令見WORKLOG，network none／原patch唯讀／UID1000，相關一次性容器已結束。

### 387 私有收據SHA-256

路徑根為本機忽略的`workplace/dosgolem/workplace/`，下表34份全部UID/GID1000。原JSON／LOG／bytes index與私有腳本不公開，公開只保存自撰來源文件。

| 收據 | SHA-256 |
|---|---|
| `moo2-387-ida-input-coordinates.json` | `46cef48565fb7dbe1cfdd24196633fa247cb5f987546d64f95c96bdf9778fb25` |
| `moo2-387-ida-input-coordinates.log` | `cfcc1f46bbea15bc1af3d3410c699b5202eb0e0a6d776541d6cd108ec5f97017` |
| `moo2-387-ida-input-coordinates.py` | `589ade57ff3368d9b5f121331b33605678d76ea82e50f8973e32225df1cf0f30` |
| `moo2-387-ida-input-coordinates.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-387-ida-input-producers.json` | `57956677515b28b1d6a13ad5c40613d37fa7ee0daa8be3bd678988d20f08f64c` |
| `moo2-387-ida-input-producers.log` | `9b1d5e17de27420a46590c755b227656ccd8904ff1ba8b073993ee2942deb23a` |
| `moo2-387-ida-input-producers.py` | `06f80a038b6e86ba69cd535bc871c622863ee4c17abdf626bdbdbd106536cd66` |
| `moo2-387-ida-input-producers.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-387-ida-mouse-values.json` | `2a8be485cb25931616b6f1d28c9f48a3389704cc259799190e37b5acefca315e` |
| `moo2-387-ida-mouse-values.log` | `00890d62902a5bb74934874de1eea4c3de5c897262b306e7b2bcd8dcb0de6e38` |
| `moo2-387-ida-mouse-values.py` | `577ca59cd10c1b70b03e300d9bcba47a1f2719de64eab1a5eb394c0dd7797a5f` |
| `moo2-387-ida-mouse-values.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-387-ida-press-route.json` | `2a43bd3323e4541dc30b9e635f4bbe9d09263b2c652d7e37eda51fa72eeb56a2` |
| `moo2-387-ida-press-route.log` | `4617401c4f99f39c1d37034f437029385d0db19027620d455dbffdda453ef2e9` |
| `moo2-387-ida-press-route.py` | `6ced1a505cae7d613d8a03187f123540d0f5ac8e852769f73591da5d7d1d1487` |
| `moo2-387-ida-press-route.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-387-ida-type6-lifecycle.json` | `28320e29d5f9528c45178724e6e56f9433f999721657cdab4615f226ed237a60` |
| `moo2-387-ida-type6-lifecycle.log` | `f5d9e5c4d36e47a2afe08c142ba3b01dba0c7ccc024adc7439c7ae1d7e01adfd` |
| `moo2-387-ida-type6-lifecycle.py` | `047258f56f0111fd78d286a026a86d4a0342371f5abeb55390b4f47a717c064f` |
| `moo2-387-ida-type6-lifecycle.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-387-source-byte-index.json` | `6cc21e6aaa326ce8879d4a4dd2c16f68efac84a59ee75a37d04158ae9764b8b9` |
| `new-game-387-document-gate-tests.txt` | `62a8d0ad9a712ae80167ce7bfe3d8f06bee32815321b7c4a252585cf68bde89b` |
| `new-game-387-document-gate.py` | `1334f1065b55cf631ee1bd6efbce4eb5a9239009294ad338d62e082f0c7c0be6` |
| `new-game-387-ida-lifecycle-run.sh` | `6967717257c6c66a67fd5b21078205f2633e01c373f13b284d5e8e837f305096` |
| `new-game-387-ida-mouse-run.sh` | `8bae69b3c1690ad2ca0470ea2667fa46c66862c3b841d76154f3b9f67d31c252` |
| `new-game-387-ida-press-run.sh` | `5ea5dacc8448bf2b0e5337ad482c8385cf89bf7f958c9de0d4254d7100ec1b8c` |
| `new-game-387-ida-producers-run.sh` | `68e3c779f701f03be7809a5c8dad21439ffafb0586661101ccd3de86369034f0` |
| `new-game-387-ida-run.sh` | `5f78eaed02d0701cb476a7887fb8a8030ffa6c7bfd26956f1aec2817a4b472bd` |
| `new-game-387-re-tests.txt` | `e9cf93e679dc3fa9770b296185635a63e681879a8048e65e8fd451cb11e2dd60` |
| `new-game-387-re-verify.py` | `7ab3bd53138a77b44d45ca38c73682335f4a3e9ed3094cac4c17d83db5706d41` |
| `new-game-387-ready-review-tests.txt` | `31a6cd90460a9228e7e4aac44d786a8f16cb577a8ffad4741cb0231422365ec9` |
| `new-game-387-ready-review.py` | `408b59e2a420cd5703668f5ebcf0bc75ea25435a19e923548598eaa1eaddf6db` |
| `new-game-387-source-tests.txt` | `865117e612de24f648f9ccd4f3f5ea3e538b50c5f0a4121e7b2198176752d442` |
| `new-game-387-source-verify.py` | `5dd05a928513bcdbb5eb960334a245af028f03b63e049b5756d6524ef0169304` |

## 2026-10-04：388原首輸入raw及裝置範圍只讀前置

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具基線1d0d128c52a7de357a41a19b8bcbb2513bc70e46，新工具011fe510aa8cf74a00d26b7bbc7d65b6094f04bc已推送github，見[388只讀契約](https://github.com/wicanr2/dosgolem/blob/011fe510aa8cf74a00d26b7bbc7d65b6094f04bc/docs/spec/388-moo2-colonies-mouse-source.md)。主庫玩法RE閘門保持。

### 原定位與證據等級

**已證實，限定此次原首輸入取樣**：原205804505／1B0845首輸入：DS188:26C51A／26C51C為0／0，width／height為640／480。目前GUI X／Y=43／48，保存事件X／Y=43／48，按鍵word=0，持按閘門26C4E4=1，共享active word26C4A6=0。裝置x／y／buttons=86／48／0，X range=0..1278，Y range=0..479，range設定旗標=True／True。

IDA Pro9.4 linear EA原17C51A／17C51C、17C534／17C538、1B3A38／1B3A36、1B121A及17C4E4的原bytes／fixups沿387 index；dosgolem_high_le資料投影加F0000h，本次DS188窗口與descriptor各自保存。2A3A34的+2為Y、+4為X、+6不當Y；完整8窗raw見工具388表，不替未知相鄰words命名。原正確callback298840保持21ED1A00。

裝置range的set／signed minimum及maximum直接取自MOO2StartupDOS嵌入的FD2StartupDOS私有欄位。私有純取值方法只存在一次性source-overlay，不取代公開Handle、InjectMouseEvent或DOS功能；device calls=13及其餘欄位前後全等。原兩旗標0及裝置範圍支持387條件X÷2，這是當次前置，不能外推所有場景或job語意。

### 驗證與停止線

READY審查先於可逆private實作。94CLI含原82逐項保持、78拒絕及16正對照。原guest一次／同原輸入／210M，原385與固定386完整首輸入守衛均通過；新8窗及descriptor可讀，整體state／RAM及device before=after，完整core／FPU／clock／callback IRQ／VBE／DAC保持。獨立核對原journal／39frames／PNG／418來源／state及副本、原9窗／record／pointer words保持，388 MOUSE SOURCE PASS。沒有新人口輸入或CPU／DOS／主庫玩法變動。

383／386／387正文及舊收據保留，追加不可變鍵回填，索引同步；385錯誤觀察器保持DRAFT。回填腳本首版在執行前SyntaxError，沒有改文件；保存first-check，修正字串後同入口通過，不重跑guest。Docker資源／image與命令見WORKLOG，network none／原輸入唯讀／UID1000，相關一次性容器已結束。

下一步389固定本次完整首輸入、原旗標及range，建立一次正常職業列press觀察契約；候選裝置660,77按原signed SAR1為GUI330,77，原36表先命中kind6 index1。先追原持按選取與1192D1／場景回呼，依實際消費點安全release；不把候選命中當人口變更或預設職業語意。 正式人口變更／存讀／完整開局／RNG及remake同狀態未知；固定日期不是seed。停止於一次可重播玩家操作所需證據，不深入平台helper或DAC／PIT。

### 388 私有收據SHA-256

路徑根為本機忽略的workplace/dosgolem/workplace/，下表35份均UID/GID1000。原JSON／LOG／PNG／RAM／state與private Go不公開，公開只提交自撰文件及雜湊索引。

| 收據 | SHA-256 |
|---|---|
| `moo2-colonies-mouse-source-388.go` | `b154cdaa17035a0a9a69025f8b7ef88972f6984eaae52a6a70ec2f27166e528f` |
| `moo2-mouse-read-388-prototype.txt` | `2579c35f375ad5b50684c2478e2961858bf9d36827cf7a1cbc9ae0ea5f44d3bf` |
| `moo2-probe-388-overlay.txt.gz` | `60692add9125290f1a4d23c324f0a9c15b957824796bdc042a0ffcac619fc7d8` |
| `moo2-save-state-388-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-388-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-388.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-388-baseline-dac-journal.json` | `96511830c84201e6f69e88132c1f68a5f6630ae1d3ff4dceecbcb014356fddbc` |
| `new-game-388-cli-tests.txt` | `6599c3a248907230accb317d93d663b8c16266e8ff11a89393ab9e4bb2b1360b` |
| `new-game-388-document-gate-tests.txt` | `2bf97a6d1cb9f18ffa655b29216e45e81372e8bf733a1e4bcd5ab0f65027dc88` |
| `new-game-388-document-gate.py` | `d1bb848eaba39ec0b019906cf332f9204dbe5e9fdcfa564889f1a60e44c801b2` |
| `new-game-388-finalize-first-check.py` | `3cd377c9f0b781d97e041c51604df1e4097f8fce1aebc04a9076eca7d55f67d1` |
| `new-game-388-finalize-first-check.txt` | `cb3af52bc9c8a0a31537a2a88227d3543e90fa0d57620af8a2d6c7571f1a4ec4` |
| `new-game-388-finalize.py` | `cfd2d007e5400130fa5e3a558f7594606b7754f7f41b69d664471987d1225712` |
| `new-game-388-generator.py` | `7da984dad2cbdfba2f0840e838c8c9cd9d8d8c75664d909dfef3412873f76ed8` |
| `new-game-388-job-source.json` | `fa7417b3e4d4840b1ed11aa9f9f1ab7ba8c7b19f514f82f29aaca204b7cdb791` |
| `new-game-388-mouse-source.json` | `650876e32fe41394fe85ae2b7dc871ccc7de551a72153fe16d2f0fee6beee49d` |
| `new-game-388-patches.json` | `77b8722606e2f26a5b4da2408890266aa6a52b71ed720a023fd3189ebe14bbac` |
| `new-game-388-ready-review-tests.txt` | `c9ba784c36af1aff3399bff04a05b08cc42eaf3c8bcf85af70ac28d53cb4f999` |
| `new-game-388-ready-review.py` | `f017245003091f7047d8b561adc21a74d9bdb27e51b36c4c299f02ec28f22e1b` |
| `new-game-388-restore-journal.json` | `242d8a9b44f5a1e522fc3f181308c86b439f61c46493c08233b4e1bafb5b2cb5` |
| `new-game-388-row-baseline.json` | `5eed5d028ba567d6787270f79fd3ac8568362890f0df45f86a5fd3a4bf682765` |
| `new-game-388-run-output.txt` | `7c4cea5c76a9f826aed89f0b81d2258e482adbbcf0a57de03ffcdc708d0710d8` |
| `new-game-388-run.sh` | `0b4470d13814b2bf72e74cd2ecc0906abd32a4ae83c57b4a758fc1a48c1da85e` |
| `new-game-388-scene-input-source.json` | `4080b7003e445827be4a977c5f46c050cd7dfc090096a1cb452404f75c4ca1d7` |
| `new-game-388-scene-input.json` | `ddf8a2eea05482b2c3abc105c05d0a38b3d2252b6da3b467851b7ffbf0f0c791` |
| `new-game-388-scene-source210.json` | `ea111350790e4cc9c288a4b62077ae959304a3f402dd760c1f28e1208eccd36c` |
| `new-game-388-source-tests.txt` | `c31311cdccc61ca2c6336fdb091ccc553c09cd9b104d380481159f3039311a20` |
| `new-game-388-source-verify.py` | `09d81a7cda83c02c3af1f1c6f926ef0659c92e476ed9859ee159f89b7c697cc6` |
| `new-game-388-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-388-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-388-state-capture.py` | `14ed1fc3ce0b0757d3ec25c52aa9808f5bc55faac629a63d14aec6d72a7e7957` |
| `new-game-388-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-388-upper-source200.json` | `c77864fdbab6c311cfc6fd56419e5135c9e3579679a258fb303b5ed8c32fcb4d` |
| `new-game-388-verification.txt` | `75460b63256feac28a00167cc383bb4c147d72b9b7e591ab23e5f202bed6fe28` |
| `new-game-388-verify.py` | `9876ed14844dfcb72c37bd577df511bf95edf535912bc501d56aa38586345fb0` |

## 2026-10-04：389正常職業列press／release與原人口record差異

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具基線011fe510aa8cf74a00d26b7bbc7d65b6094f04bc，新工具d2c3519528cc475d4c891990e72449a0698fd5bd已推送github，見[389限定正常輸入契約](https://github.com/wicanr2/dosgolem/blob/d2c3519528cc475d4c891990e72449a0698fd5bd/docs/spec/389-moo2-colonies-pop-press.md)。主庫玩法RE閘門保持。

### 原定位與已證實正常輸入

原bytes與fixups沿383／387索引，IDA Pro9.4 linear EA：11E1A7→113FB9、1192F3／FF1540881A00、BED21、11E508及11E50D；runtime及data投影加F0000h，各自標dosgolem_high_le。原1192F3的間接場景callback正確位移298840，target1AED21，舊2A8840不賦予回呼語意。

**已證實，限定此次正常輸入及生命週期**：原205804505完整首輸入／raw／device守衛後，正常裝置660,77／buttons1按下；205813953進入runtime203FB9，205814268依真SS返回20E1AC，EAX1。205815045進入1AED21，206207980依真SS及ESP+4返回2092F9；同一步IF、pending0／inactive／IRQ安全及至少20ms條件成立後送buttons0放開。206264647到20E4EB零按鍵分支，206264656到20E508 kind6 CALL；206264681再次場景入、206658139返回2092F9。206658146／147執行20E50D清共享active並到20E516。15事件與所有取樣前後狀態、RAM及device均只讀，19個裝置回呼完成。

**已證實，當次raw差異；欄位語意未知**：原DS188 current4／pool5B2044／record5B25E8的361bytes，在press至active-clear-after206658147全部保持；210M終態出現8個差異：+0Bh FF→02、+0Dh／11h／15h／19h 02→00、+C8h 49→B9、+C9h 00→FE、+E7h 08→00。三個UI pointer words由310／310／310變330／310／310。這是原raw實測，正式職務、選取群及放置語意仍未知。

正常玩家路徑的原PNG人工核對：首輸入顯示Colony of Sol II、Pop 8,000k (+73k)；210M顯示Research Colony of Sol II、Pop 4,000k (-327k)。終圖SHA-256 6be8a5cf20dbdfaca8c2471e407e9e70607c8201ff3633c261ebd9e5b71cdf26，末態EIP1A5042／482658319µs。顯示差異已觀察，不用它命名原欄位或宣稱人口配置完成。

### 驗證與邊界

DRAFT／固定來源與前置審查後READY，再生成可逆private實作；106CLI／unique original Step／原210M／兩次正常裝置輸入／24事件上限與source gate通過。15實際事件及PNG各對hash，真SS incoming stack slot與ESP+4證實兩次場景近返回；press及release服務本身不改guest RAM。原8窗與device、before／after完整狀態證明取樣只讀。13,233列共同日誌與49,210個press前DAC groups、完整原388首輸入／raw／device及舊9窗保持。輸入以後原程式產生新的末態，不宣稱與舊無人口輸入210M相同。

獨立驗證首兩次漏掉解壓縮mtime及DOS DTA時間／日期。連續同類拒絕後回查平台規格入口、既有338比較規則及le_startup.go的0x16／0x18欄位；最後只排除mtime和DTA0x16..0x19，保留attribute、size、name及其餘bytes。完整首輸入／raw／device、13,233列共同日誌與49,210個press前DAC groups另行比對；每次RAM雜湊只依既有跨run規則排除，不遮玩法結果。三次驗證都讀同一批原收據，沒有guest重跑或原observer／CPU改動。

原guest一次，state副本另行hash／size／UID1000核對，418來源及SAVE10／MOX與388保持，正式存檔語意未驗。383／387／388按不可變鍵追加回填，原正文與receipt不改，385錯誤observer仍DRAFT。公開CPU／DOS／internal／原probe與主庫玩法不改；Docker命令與資源、腳本失敗分類見WORKLOG，相關一次性容器已結束。

下一步390沿同輸入及210M，追查206658147之後原C086E→C02F9與B9C3D／B9E94的最小正式寫入鏈，定位這8個record差異與可放置狀態；取得證據才訂一次跨職業列放置。不假設8,000k→4,000k已完成換職或刪除人口，不盲增cap或深挖renderer／平台helper。 原record只在通用active清除後至210M變化，本輪沒有定位實際store或正式職務語意。固定日期不是seed，正式放置／存讀／完整開局／RNG及remake同狀態未知。

### 389 私有收據SHA-256

路徑根為本機忽略的workplace/dosgolem/workplace/，下表42份均UID/GID1000；15個PNG雜湊另綁在pop-events／terminal收據並由驗證器逐張核對。原EXE／JSON／PNG／LOG／RAM／state與private Go不公開，公開只提交自撰文件及雜湊索引。

| 收據 | SHA-256 |
|---|---|
| `moo2-colonies-pop-press-389.go` | `4395a63d157555ad918007ab816d92fccdf5d872693033253dc672d4227bff4d` |
| `moo2-probe-389-overlay.txt.gz` | `bf9487b700331c41beb2d92ecadbe21f38cb8e804a411f9bc9a4454783ef3600` |
| `moo2-save-state-389-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-389-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-389.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-389-baseline-dac-journal.json` | `39b641d6ff0087069b8d07c5d9802a8aa7db70568f2f0d41ed4cda84387c44c2` |
| `new-game-389-cli-tests.txt` | `023c572bb9f8a5419639b75e33736e1d5e97729b0a27405b06a384ee812436a9` |
| `new-game-389-document-gate-tests.txt` | `ed745d5322c0947c2b45f0f3fe4a4550eedd727988fa6fd16cf47656ad72d784` |
| `new-game-389-document-gate.py` | `43c295c2531bc68ae45c81456feaae44a8745f83e127df28c18d5cfc956e1002` |
| `new-game-389-environment-check.txt` | `1e691072207c21d77bf8d3bb872274248dba10de40c00b5238de9ac22f2b0cd6` |
| `new-game-389-finalize.py` | `006ffab16c815ba176827c0221bca48422e27831641e76594241334e0b4183ca` |
| `new-game-389-generator.py` | `6b6c4173829d33ab205c3b5b0a990c591d37fef03735e89861ffc17c8b30dcea` |
| `new-game-389-go-block.txt` | `c29bb42f6f0b4219835abeabeb82070ecbda697e7e967106745937bf50e4a0b0` |
| `new-game-389-mouse-source.json` | `5ac76d7c973d8f1ac693c929c04a939b6d3014b0ebaba5fe11b48228959eebbc` |
| `new-game-389-patches.json` | `372e850517fc234df239e66d94587a111a3f41d6dab0a2b4f22733b7a9ee25f7` |
| `new-game-389-pop-events.json` | `5c151aaa2f6aa3ac8b4ce8a81b3ff976d7e2ae7ccabfcd256ea50347270a9ed9` |
| `new-game-389-pop-terminal.json` | `3af68d6bdac519dbb2e2107b0af35c71b9b245b6ee2f87e470ddaa93b15bcfab` |
| `new-game-389-ready-review-tests.txt` | `4411bf7cc496daba6905aa3d6f3d4ccb961bc60a9c681d56852a1cd1a5c49e0a` |
| `new-game-389-ready-review.py` | `0312e702331f080e249c8d7a7fa5523b497c0a7c2694f327df7f57675effd0fe` |
| `new-game-389-restore-journal.json` | `d955c8529fc36cb8d6f59fa027b94b23af751dcc27db5e2516eee939daa136a9` |
| `new-game-389-result.json` | `123e26d1cefac6329a981ed0c80cea7a8860189d7b74b30b1f1a00091dbf55be` |
| `new-game-389-row-baseline.json` | `c4598ae2e1580a11c472acb9ebf725e502adccdc3ae4fd72369390fa65e5b17c` |
| `new-game-389-run-output.txt` | `08f122e651ebf97517c3d8d70c1bf8d0029b91bb0e0161bcb6de9a7378b3116e` |
| `new-game-389-run.sh` | `dcc69ede2ac30af9d8b4baa11c6a82feea8926828d29dcbb15bb071a7152199f` |
| `new-game-389-scene-input-source.json` | `d25854459b5f6b82d9555526649b65af585482bcc4babde1e0a850384432f556` |
| `new-game-389-scene-input.json` | `773e62221c396ca4182fed2655b038c6a9d2f2f46e9b129d1dbb16b5eb8c72c8` |
| `new-game-389-scene-source210.json` | `6c36df3be3b0b301b4bc6acb790cfd13fb103433793695d8081a23d0969fb3a2` |
| `new-game-389-source-tests.txt` | `b76bab4a43f09bbd19fa0bfee9c0933878a840b293273bde1678ef6808cd75cc` |
| `new-game-389-source-verify.py` | `7daa6011c29b9f62ae7f680cbf30434cd6ad27c60fdea6e2ec71c8cf0ffe25d1` |
| `new-game-389-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-389-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-389-state-capture.py` | `a5dc5902f45fcc253d6e7069a25d53121fe56777462ec4daae2c4ccaa66e67f9` |
| `new-game-389-state-tests.txt` | `4989d1c00d4813bda8ee4afcf2714f07fd665a108595ba5002c98fa783d62e21` |
| `new-game-389-state-verify.py` | `6daf3c42a37c1ec36ca48db493e7666544ec203805bc7f340a1b6dfd5d561c2f` |
| `new-game-389-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-389-upper-source200.json` | `dead621c2e0a7ef147a44ad78f82e9a08de985fcc5128cde5f1a39bc81519ee5` |
| `new-game-389-verification-first-check.txt` | `733286f7e9014aedad3a7237b402ac305349f12db71cfb4cdbdf64fa117a175c` |
| `new-game-389-verification-second-check.txt` | `e3dd627f71a5d1f0418a2af1e7712b454cc25893e6c454fd18618d52a51e2a96` |
| `new-game-389-verification.txt` | `adabe775299606dee0980f716c3c241b5988cbdca7aa2f9b21a298b71dfdfe9f` |
| `new-game-389-verify-first-check.py` | `d9a1de69ba98407da5d5849ebdc90aa05f95e499db1f1feb2ee7b92a182529fd` |
| `new-game-389-verify-second-check.py` | `ff3ca3dae01d950aa802552140a0a1086b4fbfb1ee751a416452da674a149d06` |
| `new-game-389-verify.py` | `c6e931ce1bc56566507ad17436294683ac11ab50020f90efaeff15e9b38e5b2d` |

## 2026-10-04：390原職業列選取的正式記錄寫入

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具基線d2c3519528cc475d4c891990e72449a0698fd5bd，新工具574f8e60998bb74c1a5add54e0cd5c362386c9ff已推送github，見[390限定寫入契約](https://github.com/wicanr2/dosgolem/blob/574f8e60998bb74c1a5add54e0cd5c362386c9ff/docs/spec/390-moo2-colonies-pop-writes.md)。原IDA9.4 linear EA／file offset／dosgolem_high_le投影分開保存；原程式與資料投影加F0000h。

### 已證實的原玩家選取路徑

修正版原210M／step_limit保持389全部15事件、完整日誌與DAC journal、末態EIP1A5042／482658319µs及原PNG。206659890到C086E、206659891到C02F9、206660054到C0337、206660055到BF627。206697288原BF681將17AABB由0設1；206697307原B9C81把17A974由FFFF設4。原B9CAF於206697426／545／664／783依序將record+0Dh／11h／15h／19h的bit1清除，02→00。BF6ED及B9E94未到達，限定此次走選取分支。

69個實際Step變更重建四個監看範圍終值；8個首末record差異均定位。+E7h的word由原DE727在206700878寫0；+C8h先由原E19C6在206704020寫FE70h，再由E1CD9在206704659加到FEB9h，後續重算保持FEB9h；+0Bh由E1E64在206704718將FF改02。另有+EFh／F2h／FCh／104h等先清除再重建的中間值，首末比較不會顯示，均保留實際byte變更。record+0Ah原08保持；原+0B／C8／E7正式名稱與職務數量仍未定型。

原記錄仍361bytes，+0Ah原08保持。已證實的是此原分支清除四槽bit1及隨後重算，不宣稱四槽正式職務名稱、人口刪除或完成放置。原PNG保持389首輸入及末態，終圖6be8a5cf20dbdfaca8c2471e407e9e70607c8201ff3633c261ebd9e5b71cdf26。

### 來源與只讀驗證

新增一次窄IDA9.4查詢，16個未索引實際writer定位，233列／212個EA／14筆重定位差異；原MZ／LE、2object／365page／51363fixup records獨立核對。保留原始函式名、EA、file offset及bytes，runtime投影分開；__STOSB／__STOSD只保存實際清除writer與呼叫邊界，不追平台helper。 全部69個實際writer皆可回查原bytes，unknown writer為空。79個新增frame／PNG、原388首輸入／raw／device及389全部正常事件、完整日誌與DAC journal保持。取樣前後RAM／core／device相同；變更由原唯一CPU.Step發生，觀察器不Step、不改Bus或guest RAM。118CLI及可逆source gate通過。原418輸入、SAVE10／MOX實際副本保持，正式存檔語意未驗。

首輪觀察器以目前DS必須188誤拒，session94008 exit1，原失敗88份產物按failed-390前綴及manifest保留；無原CPU缺陷證據。改監看固定descriptor188，實際getter隔離測試重現舊拒絕並證實三種DS切換、只讀及越界拒絕；DRAFT→修正版READY後同命令／同輸入／210M乾淨重跑。修正版session42905 exit0；獨立驗證session84523 exit0，同一批收據，無第三次guest。 所有失敗來源／日誌／圖片／state雜湊由failed-390-manifest.json綁定，舊275份主要收據保持。環境、命令、清理及版本見WORKLOG；主庫玩法RE閘門不變。

391先捕捉選取後下一個原輸入點1B0845，核對原17AABB=1、17A974=4及第二列signed熱區；條件成立才以正常裝置660,107一次按下及安全放開，驗證BF6ED→B9E94與四槽是否恢復。不得直接改bit／派送ID，不把210M中途renderer末態當可按輸入點。 原第二列來源矩形與first-match沿387；同狀態輸入還要核對當次裝置、IRQ與callback，不以局部測試代替正常放置。

### 390 私有收據SHA-256

本機忽略根workplace/dosgolem/workplace/，下列53份主要收據均UID/GID1000。79個PNG另綁在pop-writes.json並逐張核對；88個原失敗產物綁在failed-390-manifest.json。原EXE／JSON／PNG／LOG／RAM／state／IDA及private Go不公開。

| 收據 | SHA-256 |
|---|---|
| `failed-390-manifest.json` | `a0a9f5b8a8fd1570c050d35952045bf9ea86f5a06828c734b67edff5c0ef2cad` |
| `moo2-390-ida-actual-writes.json` | `480824e90a8a37712bffe29ebf79828edaed6d19dd91a9fa5a2d2f2b1c9dbdfd` |
| `moo2-390-ida-actual-writes.log` | `6abd105157da1bd1b6f04932b181be21a024687954c8a603923dec0bfb49f8a7` |
| `moo2-390-ida-actual-writes.py` | `c1a2a107b0483d428092257f21708aa358897436183f27a480746269a49d8dc1` |
| `moo2-390-ida-actual-writes.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-390-source-byte-index.json` | `1c4dfd2c492bd26ab53fc2dab0c5c00e1edd3d06798c10023eff34798388278a` |
| `moo2-colonies-pop-writes-390.go` | `ec043c8fe60c741f99a6ac8ae6a6db2295b84bd4aa306b95ca829331bbf4fa18` |
| `moo2-pop-getter-390.go` | `5cc4603d4f256207b6fcd4d3c7a2db8082da9adba0307d0d6773ab56476965ad` |
| `moo2-probe-390-overlay.txt.gz` | `aa3cc52cd6d43fea71a0e8c2debd77359c6d0474565aeba4f1fa26b1e2091cba` |
| `moo2-save-state-390-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-390-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-390.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-390-baseline-dac-journal.json` | `16c091fec130bb60270c229b27d8dabe83116dc231fb3b93b4dccc2bd350489e` |
| `new-game-390-byte-tests.txt` | `983c99673d37415caa9bc5dd438b4ec910bb9c79d796dbe230418a4c5f1634f3` |
| `new-game-390-byte-verify.py` | `2725cb87406c192f1ec090ade4dbf92d0efc33b215fb023a435c53134e84a529` |
| `new-game-390-cli-tests.txt` | `9f6a0ba025bd44cf216911c7e16fea2a366c37e52bc2c0b1c81b38dfe9740dd5` |
| `new-game-390-document-gate-tests.txt` | `c69c9cd4fcc859abd95f59cbe851d8bbc820b2b3227a52272ab121c1332f6e9e` |
| `new-game-390-document-gate.py` | `7638a8a2902f2d751aacb662177009d272edfdb9c1e08714c8aee6a65d2629ae` |
| `new-game-390-finalize.py` | `1f250c68d67afe889abc293f097b0022b1e907b8752bdc66ae012e1650254194` |
| `new-game-390-generator.py` | `fbc8c8aba3491c86710eb58b3a1ef1f7d46ab45540eab0c62316eb03e9ded60c` |
| `new-game-390-getter-test.py` | `5524530ed0edb7929595fd53aafad8158211ade37ce65a113ebea60903361c2e` |
| `new-game-390-getter-tests.txt` | `570f9f7a82ac7511b0fa1e90f682251fa93a11bb7aa1f236166679c15d03d76d` |
| `new-game-390-history-audit.txt` | `2919f8f33bf84e5e614daad4dbf45f39f215a986b2d9deaa3de29fca830b670b` |
| `new-game-390-ida-output.txt` | `453b20b6b9778963e82a54de8faea90176b0856ab91784d6b9a58817a72be75f` |
| `new-game-390-ida-run.sh` | `aec16a4150a65b26dd9f3c962ffd6fbea8b61e6309090a3fba2f6c835dccb59d` |
| `new-game-390-mouse-source.json` | `da8e74c4520d1b993df6cbfc1d35b887d8d689cd69510b039825c472c741ca36` |
| `new-game-390-patches.json` | `05b4045a1f4913d6b274a789132155d4e5343980342bb6dd9b4b8e70cc7778ed` |
| `new-game-390-pop-events.json` | `531ff59d427c8af6960dd886b1fc66e0cef3d680e23f3adaa176af2d73183454` |
| `new-game-390-pop-terminal.json` | `8510fe25f71add56aee28ea6621f3475d40fb68470c343a1ad9d140e4e49f996` |
| `new-game-390-pop-writes.json` | `fe71dc8a45063a6dd28565b40a25705796ca049e2fbec5f825922482e80431f3` |
| `new-game-390-ready-review-tests.txt` | `011c4da5f06657442796cf7d6bda594b7d41486a8794b45ce47dc51aa6ab1bf9` |
| `new-game-390-ready-review.py` | `a57fbdea98627eb19d088f00138539b7b5b96d3175ba1e571ac9f144f8815eb3` |
| `new-game-390-restore-journal.json` | `6e26225a7ae3ca08fc8796a71bc89cf56587532dc5565407e32a23a005b626b3` |
| `new-game-390-result.json` | `123e26d1cefac6329a981ed0c80cea7a8860189d7b74b30b1f1a00091dbf55be` |
| `new-game-390-row-baseline.json` | `f635eb0c1051d239525320e0e99e5c3eb5836f5852ca1f99a7b71818fc0a1295` |
| `new-game-390-run-output.txt` | `6be5d92b236aa79e4e7c3f9f17810129420a19fd1a184eb4cd9b1ed2b70992e1` |
| `new-game-390-run.sh` | `0b166d7e1010d6d1df8e6c120dea0629efd0115362da22fba8d4122cd9dc43bc` |
| `new-game-390-scene-input-source.json` | `0a1d2b69512c6e11f4000b7c1b2fdf84f55ecf1d8f4f20d520c502b1e768a813` |
| `new-game-390-scene-input.json` | `cc9832c6555edd6767d6b36f670428e27097fbe65747c2c5b76b287313a96024` |
| `new-game-390-scene-source210.json` | `788e56eb77a20cde377b31ef89301bf73078cc64c27dd3bf7e6592548c84958a` |
| `new-game-390-source-tests.txt` | `4f584001875c60ec2232069d63367692518083ebd7aa775fc02a0dbeadd08fb3` |
| `new-game-390-source-verify.py` | `15e983474d9d04d24471e3c8a215d92d0762b261e792807d72a4106e9d0616fd` |
| `new-game-390-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-390-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-390-state-capture.py` | `8e913cd876862af029f9975b4f62269f766096590961f160c108a8b30196410e` |
| `new-game-390-state-tests.txt` | `5fe79a160d39cee12b3fbfd566d877d7b44c203393c790b90521826213e2a4d5` |
| `new-game-390-state-verify.py` | `c12b94765e3274de0c7a398d1fdf90a9258e78db0079d98c062b742281c29526` |
| `new-game-390-summary.json` | `e3517341a9f0bfb43c7e93301bbf7219d100175bfbe5b290140e474453fbccc7` |
| `new-game-390-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-390-upper-source200.json` | `cf5772512116874b443534547d50c057e87794d9e095709fad1d9e7715cbd1f8` |
| `new-game-390-verification.txt` | `524195e71cbf64a74ce96a46d5d77af0dbb5796fff31ec12447b0885712645f3` |
| `new-game-390-verify.py` | `ea081e0e0f04cb4fc075d4f15195e18633b9815cec88531d728f9821e83918f7` |
| `new-game-390-writes-result.json` | `d59e5662faede544d097e7f45d9243fafdbf552041bf83a2e3f285cef63d46dd` |

## 2026-10-04：391選取後kind7職業列及按下／放開來源

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；工具基線574f8e60998bb74c1a5add54e0cd5c362386c9ff，新工具b17eb21c212deb1f95cef54f1399d7b35efd8bdf已推送github，見[391限定RE來源](https://github.com/wicanr2/dosgolem/blob/b17eb21c212deb1f95cef54f1399d7b35efd8bdf/docs/spec/391-moo2-colonies-pop-place.md)。原IDA9.4 linear EA／file offset／dosgolem_high_le分開保存，runtime投影加F0000h。

### 原390控件表反例

**已證實，原控件表及輸入幾何**：390原首輸入36表／kind6，210M選取後為37表／kind7，整表2035bytes，SHA-256 d1c13ea6764aab2848addf42cadb7166030d2a5421fd0f8533b9b33773842524。三列signed矩形為310,60..510,88／310,90..510,118／310,120..510,148，pointer沿2879DA／DC／DE。裝置660,107經SAR1成GUI330,107，當次first-hit2。 原37表由pop-terminal及scene-source210共同核對，沒有替換原畫面或RAM。原36表仍屬於205804505，不能套用到選取後。

**已證實，原來源分派與欄位來源**：原11E1EC／11E334／11E503只讓kind6呼叫1192D1，kind7略過held與release場景CALL，仍清共享active；11E582到11E69D，依var_30在11E6D6選正index或11E6E2選負index。原11DB87呼叫123C1B，123C33讀cached word_1B1222，runtime位移2A1222，123C47近返回；390末態cached1與裝置buttons0同時存在，不把cache當新press完成。snapshot的calls取s.calls，源碼證實它是啟動服務計數，不是AX3 poll次數。 原控制流是靜態證據，當次selector、按下／放開及配置結果未知，不宣稱正常放置完成。

### 來源驗證與推論邊界

初稿READY審查被37表拒絕，原DRAFT與AssertionError保存；尚未生成private observer或新guest。兩窄IDA匯出238列／238EA／27原file bytes與IDA重定位差異、MZ／LE／2object／365page／51363fixup records通過。原390表／cached與device反例、九個含端點正對照、y92 first-hit1→2反例、原getter及service counter程式碼通過；383／388／390追加不可變鍵回填，舊正文與收據保持。

**強推論，待正常輸入驗證**：392先保持完整390至210M，再有界捕捉原1B0845及當次37表／kind7／first-hit2、17AABB=1／17A974=4與安全裝置；正常660,107 press後觀察213C1B依真SS返回20DB8C的低AX1及當次新座標、callback完成、安全IRQ與至少20ms，再release。不得等待kind6 held場景或以calls增加為消費閘門；實際selector／BF6ED→B9E94、record與原畫面另驗。 此候選放開條件由已證實來源導出，仍要觀察實際原getter返回及新座標，不能只讀cached1或s.calls就標消費完成。

沒有新native／裝置輸入／Go或主庫玩法變更，不增加玩法分母。正式職務／放置、跨殖民地、存讀、完整開局、RNG及remake同狀態未知；固定日期不是seed。Docker入口及腳本失敗分類見WORKLOG，專案容器已清理。

### 391 私有收據SHA-256

本機忽略根workplace/dosgolem/workplace/，下列27份UID/GID1000收據。原EXE／IDA／JSON／原畫面及初稿只留本機，不公開。

| 收據 | SHA-256 |
|---|---|
| `moo2-391-ida-button-value.json` | `fb04bbb07f06deb274306f8003419e822ea579ca11f85bfd340703d3c32e074d` |
| `moo2-391-ida-button-value.log` | `408405b4b3e00bdc282d93f09789e07d008dc575aeb682aa590803b9f62952b6` |
| `moo2-391-ida-button-value.py` | `223dd55d5ff20c1ae56f3b45ee49516bc5dd2cc2ed86d2bda40d7e01125976b6` |
| `moo2-391-ida-button-value.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-391-ida-kind7-input.json` | `8dbb018a2720de9fed058d21173590a216c816fa0f89bbde91906a307167e804` |
| `moo2-391-ida-kind7-input.log` | `d52846499b1dd20ef7f3f97f04125ba3f1d22b0bacbb10a4904f98a4f0005132` |
| `moo2-391-ida-kind7-input.py` | `169c7536111ee134c6565d6bfaa9edd1f7a885306299cb83ccadb3a88e2e03fb` |
| `moo2-391-ida-kind7-input.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-391-source-byte-index.json` | `29efff82099bbf1335f112a3b2ac4f4661549dbbb223674d75b659b9b950ecb1` |
| `new-game-391-button-ida-output.txt` | `35951b6fe740749f8a7c0d3c53ca9c6225e049e9616ca883b6bbef53a8b3b9bd` |
| `new-game-391-button-ida-run.sh` | `75ea9c692f669f2af7b40e1f45e48a6046e63442ded3333d4a2d62ff76c8de87` |
| `new-game-391-byte-tests.txt` | `d262edf7b3a77372b779c6f6cc1bdd8a2591609b27a3c7e6ab990c2ca980b6f6` |
| `new-game-391-byte-verify.py` | `dfab9f0e749cf8577771dce442aa1db8674ef19cfc5757237533800a4f998cfb` |
| `new-game-391-document-gate-tests.txt` | `cdcff3a4c8e2db2397808c183e4ac88439ebfc5e404be6ea276bdbe5855833fd` |
| `new-game-391-document-gate.py` | `14cc2e32cf1d8598956ad7d6ae66123c9a4bb116d62eb17e9ea02154338f9a5e` |
| `new-game-391-finalize.py` | `0e6834b5cae84ac29634d3cbdf79e31b638802e23e8f5e24af7e77f302321e24` |
| `new-game-391-first-draft.txt` | `bc4e5cdcb3c44e708043cdd61e5ab760174dad61b0030b59e20468cca19ea0c6` |
| `new-game-391-ida-output.txt` | `f1e994234c36a8b469459801298258e676526a6f6a5cc45f849c831e11031d64` |
| `new-game-391-ida-run.sh` | `f3c2edd99ba8aaff6b175045fdb010fc68f6d368432cdb377c453c8f0f442d99` |
| `new-game-391-ready-review-tests.first.txt` | `b07f19ad67656984aef27cb39b2115ef63dde11ebae79144de9796bc5d6303d3` |
| `new-game-391-ready-review-tests.txt` | `b07f19ad67656984aef27cb39b2115ef63dde11ebae79144de9796bc5d6303d3` |
| `new-game-391-ready-review.first.py` | `98b48bee2a5254fea38bc2e217b6eaa81f568b67f1996ebc8f72b1c760d61b9f` |
| `new-game-391-ready-review.py` | `98b48bee2a5254fea38bc2e217b6eaa81f568b67f1996ebc8f72b1c760d61b9f` |
| `new-game-391-result.json` | `5a6ccfbb60b06a755791c11dbb6fc07df2faf6559bd762f076772def6baeb6bc` |
| `new-game-391-source-tests.txt` | `110235da89affdfd9d7e9ab060d3f079f28be2a7c91af9619470d54e9a36f63d` |
| `new-game-391-source-verify.first.py` | `12b011d11aeabb5d1f597df2f887a0f418d8cb96d4c0987f54462dc1efcb87d7` |
| `new-game-391-source-verify.py` | `84b36e8c4b9ddb46b16f53ddfd0b79ed3017fee8be749c2665ab2cebf8b928b2` |

## 2026-10-04：392／393原kind7正常放開與配置分支

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；Go1.24.13／IDA9.4，IDA linear EA／file offset／dosgolem_high_le分開，runtime投影加F0000h。工具ddaf4d80291eb33e759fb01695018c3960786b7a已推送github，見[393限定正常原路徑](https://github.com/wicanr2/dosgolem/blob/ddaf4d80291eb33e759fb01695018c3960786b7a/docs/spec/393-moo2-colonies-pop-release.md)與[392拒絕候選](https://github.com/wicanr2/dosgolem/blob/ddaf4d80291eb33e759fb01695018c3960786b7a/docs/spec/392-moo2-colonies-pop-place-input.md)。

**已證實，原正常輸入與控制流**：完整390三份原收據及392前4phase保持。210272551正常660,107 press，selector210282377真SS返20E1AC／EAX2、新GUI330,107、callback20／20及41,697µs已驗；同一步安全release660,107,0，兩次裝置輸入不直接改guest RAM。210283103 released-zero、210283112／113清active，kind7略過kind6回呼；210285765／766進原BF6ED／B9E94，210305813真SS返BF6F2，212590909下一原1B0845。原配置分支與返回已驗，不外推正式job欄位語意。

**已證實，原資料與bytes**：15phase／50個原Step變更完整重建四範圍，65frame／PNG逐項雜湊通過。record13差異：+08h 00→06，+0Bh 02→01，四槽+0Ch／10h／14h／18h 00→80及+0Dh／11h／15h／19h 00→02；+E9h 06→0C、+EDh 0F→12、+F9h F4→F1。+0Ah08保持，17AABB回0／17A974回FFFF，控件表回36。新BA6E8／BA6EF、DF0E9／DF0F0、E0717各八個鄰近指令，83列／53EA／8fixup差異；MZ／LE／2object／365page／51363fixups及全部runtime writer原bytes核對，原函式名／EA／offset／bytes保持。sub_BA5DA只保存實際writer鄰近內容，不深挖helper或renderer。

**勘誤與證據限制**：392第三次215M原held只有22個17AAB9的2／0變更、record末值保持；候選getter至20DB8C未觀測，不將cache或啟動calls當消費完成。392仍DRAFT，393以實際selector返回補驗release。兩次core拒絕由原callback4KiB尾端8-byte及原MOV SS／MOV ESP中間Step暫時不可讀解釋；被動unknown0與strict near4-byte各自標示，保留175／184份失敗產物及原390的88份。無CPU缺陷證據，公開CPU／DOS／主庫玩法保持。來源路由、實際命令與退出值見WORKLOG，IDA殼層0／idat1及非空JSON／獨立bytes分開記錄。

**強推論／未知**：slot word的正式job與flag位元consumer尚待394核對，不以raw變更直接猜enum或人口總數。原418輸入及實際SAVE10／MOX副本保持，不代表正常存讀已驗。跨殖民地、完整開局、亂數及remake同狀態仍未知；固定日期不是seed，RE-first保持。

### 392／393私有收據SHA-256

本機忽略根workplace/dosgolem/workplace/；原資料／原PNG／JSON／LOG／IDA與private Go不公開。下表記錄重生入口及主要收據，失敗manifest另綁定原447份產物。

| 收據 | SHA-256 |
|---|---|
| `failed-392-manifest.json` | `8f4e2c084a92024b08238ac0a4be2fdf143e9f77b08dc383ebea5b3d2e416a87` |
| `failed2-392-manifest.json` | `5c55a234be7361991cc451f093abd2ec3dd2cf32c23659ee7f186cb958676fdb` |
| `moo2-392-ida-actual-writes.py` | `3287d585d08cce346403e51288160f9b8504263e951049e3356ee742bcc13f19` |
| `moo2-393-ida-actual-writes.json` | `86f4f24f99c95ee833432ea082a5cd436b9bdf0726b0ca445e6d60932da4ad33` |
| `moo2-393-ida-actual-writes.py` | `cc47b50f1e878f71a41e75957153440d295a790a3da25d250f8e18a4c82c7fc6` |
| `moo2-393-ida-actual-writes.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-393-source-byte-index.json` | `655c75160a0494e5978957e0af39bee09e90a2d2b858eb964f8bbd86deeccfa3` |
| `moo2-colonies-pop-place-392.go` | `6edc3987af0e5210c26cf3790848be5f8173a29d59e6b1b38b06f46b212afa40` |
| `moo2-colonies-pop-release-393.go` | `dbff2283b2e1332f46a01f53135f3acb2de66ad2bded9a5e50b8af8e2f14bacb` |
| `moo2-place-core-getter-392.go` | `2d6993e00688525b6a627bb13540eb192f90a72f89bd3fabc0bf45a2a9e5cd68` |
| `moo2-save-state-392.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-save-state-393.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-392-baseline-dac-journal.json` | `2308441d06c0314306e208576814c59cf3ad6851b84ce6a8209b0f7b8b03e08e` |
| `new-game-392-byte-verify.py` | `a22011e2f000fc4b2fa4021312441a8e5cb6c518bdb3b097a9ffd9fd75bd101d` |
| `new-game-392-cli-tests.txt` | `90333f80dfae1745efbed8c64839e23cd6e44f6c0e21f4286804b055324275e7` |
| `new-game-392-core-binding-tests.txt` | `e065b49be3275e3bee0d1d364e48e534879601cf51b7b84f1ddd130e744bf7c3` |
| `new-game-392-core-binding.py` | `c391fdc7ae91389b59432603741c66897d48b7a1b559493e064bb86ae332b52f` |
| `new-game-392-core-prototype.txt` | `a6679e9e45b19a24809aba96e3afef119c7737e4d2858fa05fb07e9eb63656d4` |
| `new-game-392-core-test.py` | `eb2859b26d47e024f9e700d3f4b981c8745057b5d0a22e663f8059008fdd3cc8` |
| `new-game-392-core-tests.txt` | `e4628a4eff7322fe9158a16feff03e87402472846a4f2e8629f87deeda6831a6` |
| `new-game-392-failed-tests.txt` | `a1e78961072c77b803f9730dedc89809a0dc0ef2182ef1396d0b8c51766eaaad` |
| `new-game-392-failed-verify.py` | `6f6ec203c27db74020ded3f219017d40d47048694d45edbd88b311e316f583f0` |
| `new-game-392-failed2-tests.txt` | `c25cfa093389c442c6ac0f6548979905ab8e812135970501773707cebaf8bca4` |
| `new-game-392-failed2-verify.py` | `dc7a0d0f01557c96b2b990ecc2727a2de0365789f91b2b65e019c04b862f1f89` |
| `new-game-392-generator.py` | `9a9916ecdba43f768a6fe43d34b55ac0968fe62a293598ffe4d0345abf9b81f8` |
| `new-game-392-ida-run.sh` | `f7e2de15115d52cbf9084f25bbf5f00733f78796e1d499ed31d4dafd57d6053b` |
| `new-game-392-mouse-source.json` | `f61e8c78e6463a7a97f0c2a2db38debe9e5b201d5fdb77ab1eea1d40e606bbcf` |
| `new-game-392-patches.json` | `5eb5228164e06abb09531aa0719a10d2ccb61a0929f93746862c761450353e01` |
| `new-game-392-place-events.json` | `83ab1eb5a1ef2cbeb044bce833c3b0fb5122e3c11805171e0fd9d553897bf3b4` |
| `new-game-392-place-terminal.json` | `028fc6f967f968ccfc0079109b38de140760077c237d23a112abe8bf288cd197` |
| `new-game-392-place-writes.json` | `0721c9bb442c2f3a5c402f1d8a3baab20d6d8ebe3642782143542e43099ca435` |
| `new-game-392-pop-events.json` | `f820e828a8f2f8046b30fdee1bc4e8c6498e36d3665df7e8aa108a1346d0e3d5` |
| `new-game-392-pop-terminal.json` | `708dbffe8596a33ed87afa3cee4342423e4117ca25e26be75a302ab5776a9661` |
| `new-game-392-pop-writes.json` | `a5a6a2e5570a30630e61492fa68eb8e1edb48ec14b340b54bf0c2f1f55cf4346` |
| `new-game-392-ready-review-tests.txt` | `bc26aad51ad6e9c73ccc12e02124401044c1f0e89793788c3920f3fb6d5b2acd` |
| `new-game-392-ready-review.py` | `42ff8e615d188161d452ddfe4acc0b017b2da48cfa2532406ffa66ae28510feb` |
| `new-game-392-restore-journal.json` | `76f4d3c32d0ca9d1fab2cbaa1de70279f7bd49f3c4b67560b5f8fc4fc59de806` |
| `new-game-392-result.json` | `96773422729722a14ea751e7d38b97f9511315bf6376ebfa570d2b2a52bd3a48` |
| `new-game-392-row-baseline.json` | `fde4ded1330b4b557dc3eef2dd463740fb8d00028c505319b72f741c5083cbc7` |
| `new-game-392-run-output.txt` | `ad3fd5b3fe061d594dcbea85846c08711d6848cda2966599b8b49ef2df9a8c07` |
| `new-game-392-run.sh` | `25a9f7ed6331e9bf4a28fcb1f9c221addedb08442cf0ed7ecb7b864f195a7ce7` |
| `new-game-392-scene-input-source.json` | `a46de96ac69ce3aa8268ec67af81236616d13e5bbf61a5291f3ad3191f10d122` |
| `new-game-392-scene-input.json` | `03ff6fffe12bb833ff35cd295bf57c11003537c85c139ee16df7d1c1b6a3f464` |
| `new-game-392-scene-source210.json` | `2be5c6ce4cde6b33dd28523530d61b4ed5289ded1854bdc2968de6645c263def` |
| `new-game-392-source-tests.txt` | `ecd82411b240665eb44d8d518994bdd2c782ec88d37e80be6fdedbcb362d6631` |
| `new-game-392-source-verify.py` | `67aa5530eb6f78b5c274c1fa2568d40be45a294e3fe8d1880ab35d9d1c1e5bd7` |
| `new-game-392-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-392-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-392-state-capture.py` | `6b735a667538fc361a4839ed122c5fd32c84ea2b95bbcf77622b49f6a8ccebd3` |
| `new-game-392-unknown-stacks.json` | `5162444568f17029c09d90ee5a9e227b958771de7e1144c91d235f24650ae8c0` |
| `new-game-392-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-392-upper-source200.json` | `e7fb0f25f71415b7d4a691b3d565f39ae94c3dc992e12cfefbe74bbdaa5e29ba` |
| `new-game-392-verification.txt` | `e88dcecc66ed1acea602b1cc6b5a2cb40744833342f6287e7c539f3950f49ab2` |
| `new-game-392-verify.py` | `2dd4bf7ad59982caae7a3532b9eee586044c118c27b152bdc4c329d8162a9fa0` |
| `new-game-393-baseline-dac-journal.json` | `27c137ae16a49f7bd1099c88ad71808cfb3c12c4e64eab0b2b5078e747319ab3` |
| `new-game-393-byte-tests.txt` | `2517236a769f45a7b973d18f4d182dca32a26cc8c7532964cd62e60c6eaa78e6` |
| `new-game-393-byte-verify.py` | `57396fae33dc116b0fc72b5af5d70278347cd5d1b4e2545931592d0d0011baf8` |
| `new-game-393-cli-tests.txt` | `cc58feb9b3d6a8d7cee6940d477c17740120be9921c88f3864462fe673f6d3ac` |
| `new-game-393-document-gate-tests.txt` | `20cabb4eb1773f480eacfcf2576ab6e117473bcba5f674a835eb1ed049524a0f` |
| `new-game-393-document-gate.py` | `c5e62dcff73ffb33dbfafa2bd3e2a1e0357b31fc5341a769316e2e4e34f1ec3f` |
| `new-game-393-finalize.py` | `99b22335e87ee7e62d02af60cf2f2e33d52d4778073e5faabdc0df202f615e7f` |
| `new-game-393-generator.py` | `44aa41f3e0271dc0d27df4b7fa84b6533b296a5db6c6c83ca1e546d7816b2e19` |
| `new-game-393-ida-output.txt` | `896ceefe48c57047ea11f036b6bcf3895c5b3ef93ab8f169b78b1ff52114d8b6` |
| `new-game-393-ida-run.sh` | `f53ae5d80b744f77b1fb7bd87e1b396613839846f889e44a9463e8288c6e879f` |
| `new-game-393-mouse-source.json` | `2030d020a1050e17ea768cf0dbd7bc9fd495a57106ae3c09983cae2d80c6732a` |
| `new-game-393-patches.json` | `2be2c3f137a8f04a8f1b75f36d67e5d25faba80c31e05bd202c722619812b148` |
| `new-game-393-place-events.json` | `43ab1fac113c01c1a34642597d7fddbb5e8774ebdfb7784eb804202d692a0717` |
| `new-game-393-place-terminal.json` | `a92979e8b9b9badf8e8b95cbb3d472d18726eca3442c116f05a28306b0664ef0` |
| `new-game-393-place-writes.json` | `9a712de1865a6afd1b1c1d11d1c02bfcad0eb28d6161ffd36692a8ae249b5768` |
| `new-game-393-pop-events.json` | `af14e2ad3394d2991b00bfde58d22f93bb4ced9e86526a2254f5eb5014ee1c7d` |
| `new-game-393-pop-terminal.json` | `9d5bf1302ce6a50bc61d95f9bd773438db29eb9a88fba0104254d2585f8ab201` |
| `new-game-393-pop-writes.json` | `e779f51afbc9ea34bb7a21dc32302c58acb429a5d1749fa8d7c22a4ce76ac59d` |
| `new-game-393-ready-review-tests.txt` | `a0b51fc812b39662a9940626e04a4d5f943b8978ec430cccfdab20efaa683be3` |
| `new-game-393-ready-review.py` | `7a7cc466e6189abc08f2ae80dc20d3e80f4e9bf2f991b1130254215589e1408e` |
| `new-game-393-restore-journal.json` | `9837fe78c1a3e4e82bb796baba0952322fd1bddce2de273b39acb3ed23c57c0d` |
| `new-game-393-result.json` | `2bf5d1bbcf428c5588e0ff76951323b981b29e2fe2fa1f78898eee4e1a2b0364` |
| `new-game-393-row-baseline.json` | `b3368fd2f2c85e53263ae610d4a001f35288365223a7145cd7370221300b785a` |
| `new-game-393-run-output.txt` | `e6376d3c65fe1b7068c431837b172e09806f1768c024a9a52338175b3ef56c02` |
| `new-game-393-run.sh` | `dd63a9de7c9975200c9ce3c6f1c0b0a015f4692291a2e95fbb1e136ffa761e67` |
| `new-game-393-scene-input-source.json` | `efe576e2113c54537108db4beda0f42b74f9167ea0d44ff775273fd974ca57a4` |
| `new-game-393-scene-input.json` | `f019d8ff13cd297f1a83f9a505fabee95bdd9f07dfedc252407b9310ee132b0c` |
| `new-game-393-scene-source210.json` | `d30ab441113743dfdf05dd72b23453cacd743683f4be11b439614927713c2440` |
| `new-game-393-source-tests.txt` | `9aa59fe5e050fdfa031bff37e5d772a791765eaf24410000845f87962528f941` |
| `new-game-393-source-verify.py` | `e12f6ca84362355df39109ae3772d4c49ea5238db9d39c72c8d28a5fd700e244` |
| `new-game-393-state-capture-output.txt` | `9332dea0878e8d161b0c361a816571d5946567602223ec6dadb95d15d9cd9ab3` |
| `new-game-393-state-capture.json` | `5502c93ab78d1b87641a44ba6d8931b546673bdeb39ae35bb6482c037616fa3d` |
| `new-game-393-state-capture.py` | `a99bf83ba71a99cb075b9b003419ee3106c52bce0f5dfe5282eb2997ef4190c4` |
| `new-game-393-unknown-stacks.json` | `5162444568f17029c09d90ee5a9e227b958771de7e1144c91d235f24650ae8c0` |
| `new-game-393-upper-return.json` | `5079802ec98ba5702cb16565244f0b521419734c55f37cc90f7d44273930254f` |
| `new-game-393-upper-source200.json` | `601c9ba2fa5c8964a3115442525c0c23d50f00ad48522b17dd6d8244616d4715` |
| `new-game-393-verification.txt` | `c8448cf0fa8957dc7699ba537737ffa011fe474d66ebb3b4e41672dfd5e42090` |
| `new-game-393-verify.py` | `310c8c4d4133aa1bd0fe8bfa4ef4410268e8bd18fd85e3753a08e1822b018ca3` |


## 2026-10-04：394原人口槽位職務與產出gate

工具a11095c650962323492f7cab4e4603fdf62f9084已推送github，入口[394人口槽位職務與產出讀取端](https://github.com/wicanr2/dosgolem/blob/a11095c650962323492f7cab4e4603fdf62f9084/docs/spec/394-moo2-pop-slot-consumers.md)。官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；IDA Pro9.4 linear EA，dosgolem_high_le runtime code／data投影另加F0000h，file offset與IDA重定位bytes分列，原始名稱／operand保持。

**已證實**：BA5E1／BA5E3／BA5E6保存colony／slot／job；BA6DF以FE7Fh清word第7、8位，BA6DC取job兩位，BA6E5／BA6E8移7並OR回slot，BA6EF設第9位。DE393測試slot+1 byte的bit1，未設者不計產出；DE39F／DE3A6移17h再移1Eh抽職務，DE3A9比較指定值，DE3FA綁定當筆record／slot／job到DE22C。DE6FA／DEEA5／DFFE8三caller的EDX=0／1／2，接農夫／工人／科學家產出。DE376的record+0Ah及DE37A／DE390／DE61E／DE621構成4-byte slot count迴圈。

原390／393收據hash保持；四個完整原範圍重建393全部50變更，八次BA6E8／BA6EF原runtime bytes、EDX80h／SI1、每槽地址與其餘位元核對。record 5B25E8的+0Ah一直08；四槽0200h→選取0000h→配置0080h→0280h。原選取時4000k不證明刪除人口，此次是四農夫暫停計入產出後改派工人。

| 原階段 | 農夫 | 工人 | 科學家 | slot count |
| --- | ---: | ---: | ---: | ---: |
| 390選取前 | 4 | 2 | 2 | 8 |
| 390選取後／393按下前 | 0 | 2 | 2 | 8 |
| 393配置返回／下一輸入 | 0 | 6 | 2 | 8 |

兩次窄IDA300列／294EA／9fixup差異，原MZ26654／LE292E4／2object／365page／51363原fixup records獨立核對。兩次殼層exit0、idat_exit1保持；非空schema／固定hash／UID1000及全部原file bytes／重定位bytes通過。既有主庫產出研究原檔SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5，只連回16個逐bytes相符錨點，不宣稱兩版整函式相同。

**已證實的存讀來源**：802CC呼叫7E154，802C2呼叫7DA76，81136呼叫804B7。**強推論**：IDA switch註記case2／3對應遊戲中讀存，原jump-table值／正常控件輸入未驗。**未知**：job3、全部拒絕文案／helper內部、+0Bh／C8h正式名稱、跨殖民地、正常存讀與remake同狀態。本輪沒有新guest／輸入／Go改動；SAVE10／MOX保持不等於正常存讀通過。395先查COLONIES返回及options入口／安全輸入條件，不直接派送case ID或寫RAM，主庫RE-first保持。

以下20份新本機忽略收據在workplace/dosgolem/workplace，均UID/GID1000。公開只有自撰文件與hash；原EXE／IDA／JSON／PNG／LOG與private scripts不入Git。較早原版及失敗收據保持原hash。

| 本機檔名 | SHA-256 |
| --- | --- |
| `moo2-394-ida-job-binding.json` | `b94db0c36a1975fcfca69219eb51ad3efd51d8354dd6f4aa495b27434d771169` |
| `moo2-394-ida-job-binding.log` | `c81211d19f82f4a51d00eaefb64e1d18fd44fb33bac082719a482939123344ca` |
| `moo2-394-ida-job-binding.py` | `f3d7bf2a88ce5d08f4cf3bc7eb690e738bb17f6e02d520c4951160c3e483d492` |
| `moo2-394-ida-job-binding.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-394-ida-slot-consumers.json` | `fee60edc26f1a92376751f6813d243539ccaccd640bc3193d6e88207ed2c2f86` |
| `moo2-394-ida-slot-consumers.log` | `6c39549b5f793b4dd944c39190f0a6f28e9c96385af283702941c402cf175338` |
| `moo2-394-ida-slot-consumers.py` | `5effb3ca0bcfcc8d938d1b9ad83d1f6182acd7befc8142315289a895542296a2` |
| `moo2-394-ida-slot-consumers.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-394-source-byte-index.json` | `9118305ec07b39551e027af067cf0fb3fa6b8211af7b506d36ebb829ae7cf16d` |
| `new-game-394-binding-ida-output.txt` | `2837fa1495c870223d0a26fc76aba7e6fc9e138190913f3bc46012d00982e7e7` |
| `new-game-394-binding-ida-run.sh` | `84c18e7bfaddb31688e6a8e580a72fc36d35edfc3474d506bb69723c4b8c7154` |
| `new-game-394-byte-tests.txt` | `3cb0fd3b9bc8a45ea815c4421285a16c71cb4ba37f3b7bbfb949a2ef3b75e7b1` |
| `new-game-394-byte-verify.py` | `7ab8f596a166c5a1872de5764fb5eee9deb4b03d53dabb4ab801ca53bf7f93ac` |
| `new-game-394-document-gate-tests.txt` | `75340f04cae1f9106d3fc857101ddfb37552dd068107e8e9d1a8c2c9b00cf6c9` |
| `new-game-394-document-gate.py` | `d3cb64fd3e606adc34d74e0825d1bea0069efdc064069c1608ddc935f8590c29` |
| `new-game-394-ida-output.txt` | `c672c2705604f5d903cd9f7ef31cd22199387c287116f7090692fce43fd9de2f` |
| `new-game-394-ida-run.sh` | `286c15e5d0fc24bf7e53d1465eed2bfb82681eb8dfee89d479d7b96b7bbad956` |
| `new-game-394-result.json` | `8218e484a075e1769121e68581eb3705efb67d6d36a3910c01656ab6f9f79b99` |
| `new-game-394-verification.txt` | `00a7610adfa2543f9e0f18c630194abcb3ecb8d282a44c73880e0a3c7aad41b7` |
| `new-game-394-verify.py` | `4f62b42395b2b147b8f7b023bf3628d3f0753279eb6f0fa4ec6c2fc60c5b4b91` |


## 2026-10-04：395原跳表與397正常RETURN

原官方1.31 ORION2.EXE SHA-256 `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f`。IDA9.4 linear EA，runtime投影另加F0000h，file offset／原bytes／LE重定位分列。395共436列／358EA／117fixup差異，原4與43-entry跳表及九個RETURN18熱區邊界通過。8011F的mode2經802C2呼叫7DA76讀檔，mode3經802CC呼叫7E154存檔，由394強推論升為已證實。只證靜態分支，不證正常GUI存讀。

397在212606147正常release，elapsed47384µs；213103590原C058A依真SS返回1004EF，ESP+4及相同SS。16phase／PNG與按下／放開前後RAM／core相同通過，完整393的15phase／50writer／65PNG及固定396前五phase保持。原215M末態1A5051仍顯示殖民地，2071AB下一輸入沒有觀測；畫面切換、正常存讀及remake同狀態未驗。418原輸入及SAVE10／MOX副本保持，固定日期不是seed。

```json
{
  "schema": 1,
  "source396_first5_unchanged": true,
  "source393_full_unchanged": true,
  "actual_boundary": 215000000,
  "reason": "step_limit",
  "pressed": true,
  "consumed": true,
  "released": true,
  "exit_seen": true,
  "colony_returned": true,
  "next_normal_input": false,
  "true_return_target": "0x1004ef",
  "events": [
    {
      "kind": "place-return-press-before",
      "step": 212590909,
      "eip": "0x1b0845"
    },
    {
      "kind": "place-return-press-after",
      "step": 212590909,
      "eip": "0x1b0845"
    },
    {
      "kind": "place-return-selector-entry",
      "step": 212596669,
      "eip": "0x203fb9"
    },
    {
      "kind": "place-return-selector-return",
      "step": 212597711,
      "eip": "0x20e1ac"
    },
    {
      "kind": "place-return-selector-release-guard-rejected",
      "step": 212597711,
      "eip": "0x20e1ac"
    },
    {
      "kind": "place-return-press-consumed",
      "step": 212606147,
      "eip": "0x1aceab"
    },
    {
      "kind": "place-return-release-before",
      "step": 212606147,
      "eip": "0x1aceab"
    },
    {
      "kind": "place-return-release-after",
      "step": 212606147,
      "eip": "0x1aceab"
    },
    {
      "kind": "place-return-gate-20E4EB",
      "step": 213068102,
      "eip": "0x20e4eb"
    },
    {
      "kind": "place-return-gate-20E50D",
      "step": 213068111,
      "eip": "0x20e50d"
    },
    {
      "kind": "place-return-gate-20E516",
      "step": 213068112,
      "eip": "0x20e516"
    },
    {
      "kind": "place-return-gate-1B08CA",
      "step": 213102567,
      "eip": "0x1b08ca"
    },
    {
      "kind": "place-return-gate-1B0960",
      "step": 213103582,
      "eip": "0x1b0960"
    },
    {
      "kind": "place-return-near-before",
      "step": 213103589,
      "eip": "0x1aad9f"
    },
    {
      "kind": "place-return-colony-returned",
      "step": 213103590,
      "eip": "0x1004ef"
    },
    {
      "kind": "place-return-terminal",
      "step": 215000000,
      "eip": "0x1a5051"
    }
  ],
  "png": "741bb4ca62f46653dc9c65d0b3d6639b93cbf9c8b71758e3864988105de89ee1",
  "formal_save_load_verified": false,
  "remake_same_state_verified": false,
  "release_elapsed_us": 47384
}
```

首次396 overlay_probe_exit137，已到固定393結果但未送RETURN，239份失敗產物保存。兩份48036591-byte terminal在新增守衛同時展開／複製，記憶體壓力為強推論；Docker未查得OOM事件，不能稱已確證OOM。READY只把新增守衛改為逐token串流SHA-256，原固定檔SHA核對，僅略三個RAM雜湊鍵；full393正對照及value／type／nested突變拒絕通過。同資源／同入口／原215M乾淨重跑，原getter與唯一Step保持，公開CPU／DOS及主庫Go玩法不改。

工具ba7dd46630aee381f0a15051fbd103ef9c0c3ea4已推送github，395限定RE、396即時放開候選DRAFT、397限定正常RETURN；394／393／395／396解決回鏈及索引同步，385／392DRAFT保持。原資料／JSON／PNG／LOG／private Go僅留本機忽略workplace/dosgolem/workplace，自撰文件及hash公開。正常存讀、跨殖民地、remake同狀態仍未知；398先查1004EF返回後控制流及1A5051等待，不直接派送ID或寫RAM，RE-first保持。

| 本機檔名 | SHA-256 |
| --- | --- |
| `failed1-396-manifest.json` | `fcfd29dee77e26c4c360a874fff2ded1ea12352ef1601199b036c658bb185423` |
| `failed1-397-manifest.json` | `f0471b08e58e2973e5e395bd7ae8479c419b129fdbbcd3b6fc5ee636e6d199a2` |
| `moo2-395-ida-player-return-options.json` | `f8698e6a8a41d9599f78a1857932ee2cd4f06d177145472a8fea8af47376e53f` |
| `moo2-395-ida-player-return-options.log` | `abb1274f24d8cc8855fda14ace37c5bd227a7f9cf7fc328dc412bd209f83ed58` |
| `moo2-395-ida-player-return-options.py` | `c515b0bc205ce902c8077327cf8aaa547dae6f05e8149b473412c70549f608fd` |
| `moo2-395-ida-player-return-options.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-395-ida-return-switch-values.json` | `038d320fee085ddb2889f9e8cfc1d557a9fa81036f940306e1449e579aeb4f8b` |
| `moo2-395-ida-return-switch-values.log` | `183c50856abafa0cc4b5eface18085e5441e5d798b7e7a3e7dd5656137edd9f0` |
| `moo2-395-ida-return-switch-values.py` | `9473f39963cecd45d2de1131137458b7aa14a5ba1d006a9ce880bcb8deec4247` |
| `moo2-395-ida-return-switch-values.stdout.txt` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `moo2-395-source-byte-index.json` | `6f5faf5ad96d2a2344debc3c3a6cc360b47e9fa972d8bc0aa605993347c56b69` |
| `moo2-colony-return-396.go` | `ed75f3fb65ff48372ec781ce8149a5b53cdbb577bea6359f24de4df417e6ada5` |
| `moo2-colony-return-397.go` | `6ef58e19e94671a0a4281cc9147e6e9ec9d9bf78816631f52621e7e0fe201f94` |
| `moo2-probe-396-overlay.txt.gz` | `74b85d04d1b5edc0e0066e4c4cf9b9d527c8d557281338ab1cdf16580d8b04a6` |
| `moo2-probe-397-overlay.txt.gz` | `e675f05810bc191faa5f06e0161bcdcf6d6e53e887d186e22ed2839194d88b2e` |
| `moo2-save-state-396-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-396-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-396.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-save-state-397-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-save-state-397-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-397.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-395-byte-tests.txt` | `184014bf562a7b01f988f65bc44f6e63383527d34e1d724c3e2e4d26cd86496a` |
| `new-game-395-byte-verify.py` | `6e7e370c47ffea98ea848ba5a3dcd446f1965838ad28760bbb637a7f2b4cce27` |
| `new-game-395-document-gate-tests.txt` | `094cc8e62a49175169d4ea646dc44ce1049bd608ff4a305250c3924a11a78e54` |
| `new-game-395-document-gate.py` | `31be51862b17b414afb358496ac9a9373097c6b21a20bb73c88bbbaf8183f015` |
| `new-game-395-ida-output.txt` | `ce7471d4f352db93c504a7e8d90b0753368532402783a0a1d3a8174658c64e86` |
| `new-game-395-ida-run.sh` | `0b9cb9de0bdbb22a86a2c38a36219b4d060bb4fddd8229e9647d10bd729b8e6a` |
| `new-game-395-ida-validation-rejected.txt` | `ce7471d4f352db93c504a7e8d90b0753368532402783a0a1d3a8174658c64e86` |
| `new-game-395-ida-validation.txt` | `5e5a84be618c9a56b2f13e2dd65a4fcdb7cdcf5284da0511011cb7caa8229986` |
| `new-game-395-result.json` | `98e55c0eda2b655bbb4e206976e0b695507fea66f6cc22b30121bc1b7e6cb4ef` |
| `new-game-395-source-tests.txt` | `3a656b7c7b07f1636581c06e54ef33bc680234688d242d3a0c2a1bd7b0e45eb3` |
| `new-game-395-source-verify.py` | `dfe5dd84127c8d7fe175bf7d96d4f449bd0ac6d50edd24df22904d4ea9eb45fc` |
| `new-game-395-switch-ida-output.txt` | `2125e3e849a06beb90cd6abcb898f9ee39539955650da7b18e658edce0f12156` |
| `new-game-395-switch-ida-run.sh` | `4845181bc4a3701373352501ffe1328dedcdf75f6dad6c3c40acf58a69044078` |
| `new-game-396-cli-tests.txt` | `e62ad87d3ef98ae326e50ecd47f2384fe1c2a6c99a31c01a580b28beb390121d` |
| `new-game-396-generator.py` | `d7a5d6c5fe4b22b121f4acdd7171e9fe138890112f618f8f62db8ca578fa1947` |
| `new-game-396-patches.json` | `c1d6a919254f7f4e72e19380493345299ec1d4cc139bc39ca099e2e3d8baa7ca` |
| `new-game-396-ready-review-tests.txt` | `818f2b608253e9329ac09b05e99c814d9896f9c03051f72fa4b0d7e462f7a420` |
| `new-game-396-result.json` | `88a635a36890c04c4c931c8f1dc0dbfc618d2ff6c08b21932a8fa9093f95a0fd` |
| `new-game-396-return-events.json` | `e095ced17ff9cc56c7222715ab6addaceb0130172825873d4bee567b77ff2c5e` |
| `new-game-396-return-terminal.json` | `a9449f25b51366f6b00337e4913700ceaf14e1612be72ea926b16b9a332e16ff` |
| `new-game-396-run.sh` | `c7c020701a88c977831ccc382619ca0229d20e3918b6f3df4fbb6a4d350abe7b` |
| `new-game-396-source-tests.txt` | `12bf1e9415d8a24c79b53b2349d0ac93757ca68992f6a48a467102f4c1daec67` |
| `new-game-396-source-verify.py` | `67d90337d1721431abfa1e984b60f6a6bf7a127e7c8ef873ca47d111b28f1f14` |
| `new-game-396-stream-test.py` | `7590cab1fbca9ddbc808cb63c9e0a67bdc0fddbe6a31fe5ad7fd65c8a02e300f` |
| `new-game-396-stream-tests.txt` | `93f4c895b80c4b994d23fbfc8c41de6e34ef566c4167a36da7614332b1af82ba` |
| `new-game-396-verification.txt` | `4a1e5ac110161f3f92fa0626a136980d2c38ac2e9b320acf5bbcecf6dc013ae1` |
| `new-game-396-verify.py` | `fd480cfe538a78ab1d9d5d724738c19b4be4d512fbfb1642f6e58a92438aeec8` |
| `new-game-397-cli-rejected.txt` | `a603d1168134cebf9c2b5b0c7d976ba9fa825964f61d99cb8361831aef43acae` |
| `new-game-397-cli-tests.txt` | `ea75d92ae74a1342281fd0406d0889a6a963629ad056123fc8c7946766ec8891` |
| `new-game-397-document-gate-tests.txt` | `18cd78d69a0930b53047ca1316d958133a70c773dd992a5c025c05441e976672` |
| `new-game-397-document-gate.py` | `cab118ac7cf71b9317ed541fb4be32c4d51d5cec95bd404457fc738edf58bf1b` |
| `new-game-397-generator.py` | `995ab014266e900990a77300980d6ba65e196a130b555567ebf93a23831fe62d` |
| `new-game-397-patches.json` | `6b87018679240283790eb36098dada18b0f1250797a8a2d072540e2039bdc648` |
| `new-game-397-ready-review-tests.txt` | `d1a28a03ad914c18fd86d10688d9d034e7ae3b28617e69ea7c231fabfc257ed6` |
| `new-game-397-result.json` | `9965536395acaa8871df3436114ff33c7719d2dffd9b90567b718aa5e1ecadd6` |
| `new-game-397-return-events.json` | `b42e9e4000764edf45af45318de558ec2fb30a274cf76ac4663e3e5f18cb4c2d` |
| `new-game-397-return-terminal.json` | `d6ecbf67eb19ec92c36a4ff288f6faf11c02c26d1b8de64f11c2f78f14cf73f6` |
| `new-game-397-run.sh` | `afd6d5a81d8fe1ad7f3286f90d931edb7c4f4e0b501516950f7ec4b1a005d282` |
| `new-game-397-source-tests.txt` | `4ff3057494e206a4445cd9a4ae391c3da3019faa4a44fda72644f4c9bf2e32f5` |
| `new-game-397-source-verify.py` | `8d728ee0c4c87d6a929a1ddaf28299a4046e3d566db69c0ff3b31ebe9e3c17a2` |
| `new-game-397-verification.txt` | `4c809e17df7b88db5c5b7223a47e90f2fa68f084cc33b37bfc0fe9ef978b4812` |
| `new-game-397-verify.py` | `51452e7acd7570ec24ab040e22673c474bce61192fdd108bcaa0ab780432c0a3` |


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

### 本輪私有收據索引

| 本機檔名 | SHA-256 |
| --- | --- |
| `moo2-398-ida-return-parent-boundary.json` | `8c4c514be8d4fe8cbc9c213bb7438ad1a89af3f416448c0b304d4a6b7d72f747` |
| `moo2-398-ida-mode-restore-boundary.json` | `1aa344c16da92ff68cc143a9769488aa65b99836a2b098471f3d972e1e64ed25` |
| `moo2-398-source-byte-index.json` | `af5b8d8c6bc2c704d1ff48a8341ca145d13f6f134e30716c3912e3fda6df85d6` |
| `moo2-398-full-dispatch-table.json` | `f0ae8fdd28c70c3e44e85888f012f83d7dea0c5ec5beac5a17e14d1ac26c633f` |
| `new-game-398-byte-tests.txt` | `787032cf47525224bcdd34ca41e2d9de8eacffa39d3e97b84d2fce171272cb6d` |
| `new-game-398-table-tests.txt` | `78d574a8e34fd7383a160dddd68d6ec7e5deee5dd12f7378dbd017e02717c9ab` |
| `new-game-398-byte-verify.py` | `8d87bcaeb5a9b140ccbc80bba63bb4d1cfac2f1f5cb18518962059a1bb1fc872` |
| `new-game-398-table-verify.py` | `f8a1c13a366cf32ca361d4cd2c8ec547dbd24fdff29b640ee717970bbcc2d78b` |
| `moo2-colony-return-399.go` | `2256ab0cd3ba70900b773612ec2cd57bbad0d1ffd98c597df75bd63d4f24c642` |
| `new-game-399-generator.py` | `f4dd7b20020a0ab1720ca25b8a5060ae892d9d34a7cf160a2f30ce09b15cc91c` |
| `new-game-399-patches.json` | `9784598a28f86a3e24fd780fdad7b72544bca01d4ff677ecde21e4f7627eda55` |
| `new-game-399-cli-tests.txt` | `698495f4538b7b7dd785615be17c3a49dea2f4a61e2efe2728c4176ede708f97` |
| `new-game-399-ready-review-tests.txt` | `e118dd19556ac37eca81f39a74b4091665261119a02bfa6d6f56508d79c4a23b` |
| `new-game-399-source-verify.py` | `1c8ee0462066ed6a66d3c62ba80e52591ddc7d200cd0751dab950b0a6a2345f0` |
| `new-game-399-source-tests.txt` | `4ea4b65de69e1d7bc36a9c67bd1040e67aedfa7dad0c5a9b9da67fb2e10e1245` |
| `new-game-399-run.sh` | `ee06aa43c378ce48f03f7f0f21237f1e14095912f8f7f9bbec1bd4b4300d413b` |
| `new-game-399-run-output.txt` | `e6250f4effbba2c1eb0996d64132978493aa01d0103fad81fc180f665d4f45a0` |
| `moo2-probe-399-overlay.txt.gz` | `ba5e0cbf4568c8677bfbd501f51c206ff7c96a41b6bd197ab7faa1c234177398` |
| `new-game-399-mode-events.json` | `09db2d1b3c1ab6aa4fac5d387ecc3f47b901904d591782b291d19ef4b25c0dc6` |
| `new-game-399-mode-terminal.json` | `b7d1b942581e56f5de8dcc01e7f71776305d094acdd6895120e0189e261764fe` |
| `new-game-399-return-terminal.json` | `e73b88c767f2886736cb8e8162e2ec14ac7cc4dc6928da4fcbc9a7d41fe83511` |
| `new-game-399-verify.py` | `f378979895fbb171d99471b673088916f16e729e43e565d0ae72c8b79c08cf17` |
| `new-game-399-verify-tests.txt` | `3be8a37c7ee9f45038e7f9c9a6a26133e0b48ec816e9ab12b726b0611b8368cd` |
| `new-game-399-result.json` | `dc43aa4a24f115da6aa29c273b2d322b8e359577b6ea4560eca32520ff714f9c` |
| `moo2-save-state-399.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-save-state-399-save10.gam` | `0cf71fc5368861758da5a5dffbd791f8b711a48cfd295572eb9a21d27dea0f4d` |
| `moo2-save-state-399-mox.set` | `de8554b3284c9f9065efe6e4f2dfa71ab0c7e732e75eaea0d0ecf465ab86409f` |
| `moo2-400-ida-return-parent.py` | `74c9f1ac667b1e9fbb42b7e18bd3a8740d77968df4d3aeba96c1208ce041201d` |
| `moo2-400-ida-return-parent.json` | `7e0c5bdbe3db9bbe98794d74e00fbe5835dfcb048975b0304f0380a48a41bb1b` |
| `moo2-400-ida-parent-flow.py` | `061c50d3537b2b5a6039c4b8ecec63b76f13a6239cbab114b37dfae44a794aa8` |
| `moo2-400-ida-parent-flow.json` | `61de4c6de2408ff9a22ea45fd2ef222c40360aeced07d7273b8f2a8afafda86e` |
| `moo2-400-source-byte-index.json` | `8a7d5b59b1dce4aa921df4043b641026a588ef29268a5ed680405ae8dffa0dec` |
| `new-game-400-byte-verify.py` | `2aa1a36804f5813bd06bf30d2dece39723d7126aaf6bd5e3f7817c47ad83b26b` |
| `new-game-400-byte-tests.txt` | `6c8363e6d42e440b76f5bf537b6805ae41e55f6f678ba7b03ddd7409e964f5c1` |
| `new-game-400-source-verify.py` | `3938a19f9b1ea468897fa0a6f821d9bca012a83b81e5998bdde11bfe01e570e3` |
| `new-game-400-source-tests.txt` | `7bec41a8f5f7a80a18800278a8490cc43432606d440e7652e8690957c7bcc316` |
| `new-game-400-result.json` | `c95eebf9ab443a0f5f1f922e8ea6bbe950c27dc5982518fda96db2dfd8fa19df` |
| `new-game-400-document-gate.py` | `ded7035a9e2de8e1700e2dbc3fc8966d0f79c37ea8eb0648aae54ef3b87e3453` |
| `new-game-400-document-gate-tests.txt` | `dda147bddd67455f75f907a35a6dba766463ceda1de8647b8b91b2bfc34a4b19` |

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

### 401–402本機私有收據索引

| 本機檔名 | SHA-256 |
| --- | --- |
| `moo2-401-ida-slot-parent.py` | `11d4dbb2500fe13c9931782dcba56ac4bbfa9207c6a69e5609c38bb9ed7e5971` |
| `moo2-401-ida-slot-parent.json` | `99186b26f6d26b1453484b564f719965b0084b8af1b6bfa6452ab41ec81f4fea` |
| `moo2-401-source-byte-index.json` | `42a8d8382fe12782540f99159b11f38ec51e5594c56a03d954c1f3ce0c4c9a34` |
| `new-game-401-byte-verify.py` | `82f5d3355b2accaa4947b738ea6708d406d6b5d12514fba592ddcd561d8cc2d1` |
| `new-game-401-byte-tests.txt` | `baeb38b5e4a5943eeebcb9c228d9402e5f8180133151712f6d73b4b6cce1dcc3` |
| `new-game-401-ready-review.py` | `09f68e340115b65d41304f6deaf9c3aaa4a2aba49aa575905f1e181157e93be4` |
| `new-game-401-ready-review-tests.txt` | `b33499481db9825c7635e263af91b0c409b663cb2748aa79f06fc61857348a86` |
| `moo2-colony-return-401.go` | `1593aa58a873d8c2b3a2d5b961d7115e449f9b1b2bb97deab8546b76376763fa` |
| `new-game-401-generator.py` | `8900e337d0d216177b2863fd86b6f8bfc0702d31d27b1ab16785a85b1739ac33` |
| `new-game-401-patches.json` | `47c8b0fb79d06adf8be56192c08b14f8b81c61454ee0af31e39f9d9776e3ea86` |
| `new-game-401-run.sh` | `da10db1c55bdefac8b92304f01491cc7f277d3880bd2fc350b6883544120f017` |
| `new-game-401-executed-run.sh` | `f55b1e2f1380ffe6d6b588b3a04248a0d36dc71a1d8c521b2de122ece34bf6c2` |
| `new-game-401-run-output.txt` | `18f4c06197f8d7082fc4a035897a369e459ec27e84a20d31b370ec10ea05b627` |
| `new-game-401-cli-tests.txt` | `20ad383c103ac4dcb70a37ffccb4599af5dfc370d96d8b89c9670c3514b37efc` |
| `new-game-401-source-verify.py` | `71727666df61179ffbfadcf1cccbd14483229e529b663931edaf7fbbb5dc00ed` |
| `new-game-401-source-tests.txt` | `a2a40219703be9fbcf4a63b0a1cd58187c226982feb0d0427c8370e28279cc7b` |
| `moo2-probe-401-overlay.txt.gz` | `b85134b530df2a149e5e273509beeb73300641b8eb461144e3cf2def74081a00` |
| `new-game-401-parent-events.json` | `3a4a1a91ede5b3474642733e84c44298de5d33cd74446bf969e6d4e46e0b78bd` |
| `new-game-401-parent-terminal.json` | `5b340590cc6a2d3d0d11125ae770db4dc6bd5a1aa71d1d8a68d00ee8f6033f99` |
| `new-game-401-verify.py` | `7de1dbd4ef42deca38970997fa10df4fb1d052329fc0e3ae757402efa80d70aa` |
| `new-game-401-verify-tests.txt` | `8941f53555a5148ff230f63814bb676bdf0c6158dcc44bc9540b05825ffab688` |
| `new-game-401-result.json` | `f2c5b6a27388a8731433d3f5eaf5003aa1fa792e93a603b5414bafad40f206b3` |
| `failed1-401-manifest.json` | `23046edd15da4cb0da1d30ea62977716e7a9f4f69acfabb4de532310614ac064` |
| `failed1-401-new-game-401-run.sh` | `e6726ad96bb6b9da76e5f20068fa29127094631044d1d4a22372358815930ead` |
| `failed1-401-new-game-401-run-output.txt` | `5b7492b0ed8dd0645208d503ce146f89c9ff00bde03fdf01a4c49a2376f36162` |
| `failed1-401-new-game-401-cli-tests.txt` | `698495f4538b7b7dd785615be17c3a49dea2f4a61e2efe2728c4176ede708f97` |
| `held399-before401-moo2-probe-399-overlay.txt.gz` | `ba5e0cbf4568c8677bfbd501f51c206ff7c96a41b6bd197ab7faa1c234177398` |
| `held399-before401-moo2-save-state-399.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-save-state-401.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `new-game-402-ready-review.py` | `bb7a81674ae58d1926c451b645c9e3049f9f50ea29a38e9f9827326fc8fd0bcb` |
| `new-game-402-ready-review-tests.txt` | `ff710959c821cd0af3c111f4dd530eef5e922bc7ff9ed7ef38cd7d3c2c9e8783` |
| `moo2-colony-return-402.go` | `f86a2e5e97d36a0abc42f14b10de00dc91f78ad884961f2385d53f7a793b5ad4` |
| `new-game-402-generator.py` | `be1c8286ac6e65e35e34191dc05f5afcb2692a0efdce9c7737807428c68f6103` |
| `new-game-402-patches.json` | `6cdc2213cf15d0db7c6088038112b6596ce121004d2fe2d46edb31573bb17983` |
| `new-game-402-run.sh` | `f7b48a9c3b12cc4c4a328a5217349034b15294d09f596faf525959b565f5e738` |
| `new-game-402-run-output.txt` | `60d7dbaddee29a32c0e4b9c9d8a66c847805b25b0549b4467b38062e91925041` |
| `new-game-402-cli-tests.txt` | `033832d82c5f7641194fd93bf266501eb04a22c087a5460d04376c3216e68d36` |
| `new-game-402-source-verify.py` | `ef857f90e82e262fa8fd66ffb8a6794f433025894c9201e486409cb0a90dd010` |
| `new-game-402-source-tests.txt` | `795fea1acfa5010d97bbb1fc2fc698728f269908792beb1b31ee18a4c1a19519` |
| `moo2-probe-402-overlay.txt.gz` | `6b11c6ff3330bc47f847143a627e805331e97a9fe9aa2a15e2e3cbf56015abd2` |
| `new-game-402-continue-events.json` | `e9c9b89a44a9202d9318e9e4b15b14fbc91ac05cce67a7c632dfddc25888b987` |
| `new-game-402-continue-terminal.json` | `0f39a119df80abe95a1eb3c8f9f8aa62e897cc9e2f19673c33b2f6bb2f3e3e11` |
| `new-game-402-verify.py` | `0dc21d4a92e39cd4557196ec2496201a713dd376a6b2249eeeaf1a21975af2bd` |
| `new-game-402-verify-tests.txt` | `7247aa0a56c534260494bd98e2511edfb1852819a0fa7cdd0c2f65161e5da2af` |
| `new-game-402-result.json` | `e52590082e1d75606a81d3d23850f225c51e3b056569c86f1777732c70459487` |
| `new-game-402-visual-review.json` | `f149f443d2c907db66c391272ab183865c16229c39a49c279f9dc451e3e89007` |
| `new-game-402-table-verify.py` | `a4f5c5d9e40351ba30b21a990c6236eb42170cdd14e45e9f5d1c44a34a483ba8` |
| `new-game-402-table-tests.txt` | `6e1309111fca77171df737b092d9d251a116849fbde5a7b0f0bdd93d1d92f9d7` |
| `new-game-402-table-result.json` | `44634fe0a6be36d548cf05f6afcc886a75538baab608c90750d3e15c8613de80` |
| `new-game-402-document-gate.py` | `9844da537fb03d7c9402ab782296ae852300e84cfbc54af5b7b52e82646c88c4` |
| `new-game-402-document-gate-tests.txt` | `0ecfc1eedaddc5729827873432c4588267ae996ea7e40caa8eefaa6c293d6de8` |
| `moo2-save-state-402.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-402-writes-place-return-continue-terminal.png` | `299868824b0ea14741bbcd781a8aa353c8344ef29bd5997f50f5be6d211de8c2` |


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

### 403–404本機私有收據索引

| 本機檔名 | SHA-256 |
| --- | --- |
| `moo2-403-ida-list-return-binding.py` | `b74da45fe69411f83cc7681a425fa0f3fcae722b595088831057d3c769d38729` |
| `moo2-403-ida-list-return-binding.json` | `def6e610c32dd614fbd7575dea0c88f977e9761c2452149f44a9566443190932` |
| `new-game-403-ida-run.sh` | `411d02a4d47b83ac72c701cb2b68a2b6a89979dfc0b5da6f8cf0dece644a05b8` |
| `new-game-403-ida-run-output.txt` | `01552b96be7e5f5c2488fec3923ac5593139382dbefac98cb3a71677ea5c1889` |
| `moo2-403-ida-list-return-producer.py` | `e9c30c5003d34234da613a3759a590f31aee6ccb2876c4c55b9629fa41519c7c` |
| `moo2-403-ida-list-return-producer.json` | `9fa597a65e70682ee82f7ce2dee244bf770b1dfd7fbf2cadb3fcda8dbd85dafe` |
| `new-game-403-producer-ida-run.sh` | `4f55f4ccc71d44d46f39f116fd35f9c6e792b172396f55cadd0203a53fe91015` |
| `new-game-403-producer-ida-run-output.txt` | `4bf83f9d8f15d4fc7843a6e55e6ca3c92ab06571d5e5fdc9b7fc842ed4ac25bc` |
| `moo2-403-source-byte-index.json` | `c762f25ad4e5d1f074949a8cb4c6d9c38407159be5ec79dd32fce3ee2fb36166` |
| `new-game-403-byte-verify.py` | `64f41a76e471a8a55bf03ca5e27e1fb119aef2f917b6f3e019640aa80431fd05` |
| `new-game-403-byte-tests.txt` | `1bd7caf64224a42158431f8b23698eafccdcb5deac35869d383e393e45d9fc70` |
| `new-game-403-source-verify.py` | `5de41a833c167f29f50baebc50b9a6ad07b84b554067e72b92ba00dbbab8a6f1` |
| `new-game-403-source-tests.txt` | `bb160781f4de9c7db5c7a55bd6922bfcb5b2149d27b6b651ceabef9cec32739a` |
| `new-game-403-result.json` | `78fcd4aae70baf32d248df0b91d006de2eacdfff13f1174b25199859e295d8c6` |
| `new-game-404-ready-review.py` | `f3cb96a587bc68dff87730fa5a2f00b120296d6c16f5d26b9836b909342309fb` |
| `new-game-404-ready-review-tests.txt` | `2ebd3d86bf993a5782d8cf328c13c1957de0f53a59041086f8febc1b50dc1e7d` |
| `moo2-colony-return-404.go` | `f21eaf07fbd6efb6081033a080fdef0d1ac94fb8c93b47e2983f7d939306acc1` |
| `new-game-404-generator.py` | `9d2735c20f36db1927956d62339306f70f63f5197737b42bc7cb0af3d8658125` |
| `new-game-404-patches.json` | `175a05b62444180d1dcf716cd2cc8de10128e5ce4cfdf3c0ec8b2888d1a889cd` |
| `new-game-404-run.sh` | `1802dba9dcfc98549b5b53141230cdd14434f939321760bfdbb9cd19ca49ced8` |
| `new-game-404-run-output.txt` | `1df39aa9fde0674309fdf4d0b62eccb8729a97674bb3dcd62dfb94d80250ad45` |
| `new-game-404-cli-tests.txt` | `344c63a7d158e03c940c9fa12ac25699cbf6f88a22677d3580e5cf8d36cf3900` |
| `new-game-404-source-verify.py` | `9689b22ae1e48bac8d08452cce24f9d01f3507c35d5c3235452e58ef99acd892` |
| `new-game-404-source-tests.txt` | `b8beabbc0790b6df08be5f96e1c2d168356daa1bf7ef1ed984464f4602142eba` |
| `moo2-probe-404-overlay.txt.gz` | `feb023b1ba148149326ec1ed08fa283dcef5de2f03031f5067372389e5cfda22` |
| `new-game-404-continue-terminal.json` | `50cdd4f23b15e6bd720aa25aa2ec1f6d4c94598a92300d41c0ff6bd84e9017fa` |
| `new-game-404-list-events.json` | `5844911c8ae97054f81fb6058cce4bef59804895018655e05cd7432b62babf48` |
| `new-game-404-list-terminal.json` | `f215afceca9adb63c42a853251c6ed1a8f573edee0ad771ae22e714ceb2925ce` |
| `new-game-404-verify.py` | `0b98d67f4597c274862e44bda7ce3be267adffae4c65544a193409d87dca7acf` |
| `new-game-404-verify-tests.txt` | `c6cefab5a87384ed2c808e6889bc1b7c10cf732d17f0adf1ad0c8b46fa7e10e7` |
| `new-game-404-result.json` | `ad37b4150232e66a2a754cc274d36446b4e674f9d36973b590ce925e09f616d5` |
| `new-game-404-visual-review.json` | `7488bae903504e2a8741f4c3ed9b1298c898f1d6e648e7b52e7794e79b2bb14b` |
| `new-game-404-table-verify.py` | `88906f06bf38982ffdbf8a8f4de8de9a6bc2505924c2656ddcffdf049bfa5fce` |
| `new-game-404-table-tests.txt` | `260c311b81ac02acde7df1d167e6325616b0de1169cceca8fa92cec6ad27082b` |
| `new-game-404-table-result.json` | `8b7f316bd2bb48f26ad62e510ac8359da3c3dfc6d6a2511814c3026cf414f47e` |
| `new-game-404-document-gate.py` | `5d2fbc3aa75c15efdc01a5ae3089fabb313b6090dc9f24f1d049828b682299c0` |
| `new-game-404-document-gate-tests.txt` | `691303e0b3197890127d032ea8b8a55800e66ad2a8b44902bd08ab09441ada5e` |
| `moo2-save-state-404.json` | `50d4793b97dc0a0b1471ecf9f0ec52f5862976d0a1bef454675f09eef599e705` |
| `moo2-404-writes-place-return-list-terminal.png` | `8b9854fca02675829dab522a22bcf7f812fcf255aa4856854630fde357a3dd4c` |
| `failed1-404-manifest.json` | `f3ed4e7addb299e2d2551335ee884aaca70e2d8d5dbe9870cdca957984dfbdcd` |
| `failed1-404-moo2-colony-return-404.go` | `93b6596e6b6aaf77ae439a13c97454d43186fe3bd8b4630d491e9a8cb9f598be` |
| `failed1-404-new-game-404-patches.json` | `9732ab5e3192397912dc43baa03be94c2b18c2f785fd86f5cedba0a83e399152` |
| `failed1-404-new-game-404-generator.py` | `9d2735c20f36db1927956d62339306f70f63f5197737b42bc7cb0af3d8658125` |
| `failed1-404-new-game-404-run.sh` | `a5fbb30f7ae44d8c51544eacfc23440576fd9e1335ad3509850af5fe959cd70a` |
| `failed1-404-new-game-404-run-output.txt` | `0294ec2350064488959f94ba34070cb61d38bb502fc15e6a186c60d3d67d1958` |

## 2026-10-05：405–409 原GAME輸入與外層框架交接

工具已推送HEAD de7456f0d5f9f5c01232a7b0028c8488e0a998de。原1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；IDA9.4 linear EA、runtime＝EA+F0000h與原file offset分列。

- [405 GAME來源](https://github.com/wicanr2/dosgolem/blob/de7456f0d5f9f5c01232a7b0028c8488e0a998de/docs/spec/405-moo2-star-map-game-source.md)
- [406正常輸入與拒絕](https://github.com/wicanr2/dosgolem/blob/de7456f0d5f9f5c01232a7b0028c8488e0a998de/docs/spec/406-moo2-star-map-game-input.md)
- [407真選單輸入讀取端](https://github.com/wicanr2/dosgolem/blob/de7456f0d5f9f5c01232a7b0028c8488e0a998de/docs/spec/407-moo2-menu-control-input-source.md)
- [408框架來源](https://github.com/wicanr2/dosgolem/blob/de7456f0d5f9f5c01232a7b0028c8488e0a998de/docs/spec/408-moo2-star-map-frame-source.md)
- [409 READY只讀契約](https://github.com/wicanr2/dosgolem/blob/de7456f0d5f9f5c01232a7b0028c8488e0a998de/docs/spec/409-moo2-game-outer-frame-continue.md)

來源405／407／408與原byte index各自保留，計597／515／66列、475／451／57個EA、187／121／15列重定位差異。原2object／365page／51363fixup records逐筆核對。

完整404／402／401／399／397凍結、215CLI含177拒絕／38正對照及8個可反轉patch。225305800正常裝置press560,13,1，225315546原113FB9真SS返回20E1AC／EAX6／GUI280,13，225323360安全release、49572µs。七個新phase的core／device／RAM只讀、原Code16／LE fixups與PNG hash核對。當次三word6／0／34，原418檔與SAVE10／MOX保持。limited verifier exit0，實際原PNG人工檢視另存：仍為星圖，游標在GAME，沒有選單。406完整契約仍DRAFT，不稱外層返回、選單輸入、正式存讀或remake同狀態已驗。

408證實ENTER6CC與EBP減82；原首輸入ESP2BD4F8／EBP2BDB46算外層RET槽2BDBE0。226846736的tail ESP2BD378異於外層保存暫存器2BDBCC。另一活動框架為強推論；實際caller與精確停止step未知。409經DRAFT→來源／只讀審查→READY，尚無409 Go或guest；下一步凍結七phase，只讀續行、無新裝置輸入、不預填1004BC，原230M上限不延長。

第三輪3GiB／GOMEMLIMIT1GiB下cgroup峰值2119880704bytes，oom／oom_kill增量0。Go1.24.13官方runtime/extern.go的工具軟上限契約與hash在resource收據；未把第二輪SIGKILL回填成確診OOM。生成器在容器暫存區重建Go／patches／runner／source verifier／state capture／resource capture，六份bytes一致。只有三次原guest，沒有為畫面或檔名重跑。


110份新收據全在本機忽略工作區，完整hash index為workplace/408-current-receipt-index.json。failed1／2／3的406 manifest分別覆蓋318／321／333份原產物，保存所有失敗與第二輪較晚殘留分類；較早45／53／39份hash保持。原EXE／PNG／JSON／LOG／private Go及存檔副本不公開。原CPU／DOS與主庫Go未變。

| 代表性本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-406.go | 36496a3f9ca1ba8a3f126e8ee2f234454be5fdb4403bf11dc48a89f6017141ff |
| new-game-406-game-events.json | 21b367cb55a1e8508dbeb2092ca50df75d5e261b4717cc730619482916745156 |
| new-game-406-limited-result.json | 7b1567724fe1c583769c70dc45d96359260b5a17f6f5c488c718235020913e18 |
| new-game-406-resource-result.json | 6ac124b002d63310a5037b1d0bf568a39e80416fb41b83cbeb8f2ba8b3d38e78 |
| new-game-405-result.json | d16199891b36fdfae8633da52b59e6ad0eb55235ad5b372e85695a87f5014394 |
| new-game-407-result.json | 6f20a114b2d69608fa90b543338eaa9e53aea6907185044629e2b10a1b012ba9 |
| new-game-408-result.json | 18d2141c538a00c81e5f6a6eebf53edbc8021dd91f3c0ffa28cc4b9d32847a30 |
| new-game-409-ready-review.json | 076611f219b0473fee125cf2acca60d618921651ada47be0044393d68af64b4e |
| failed1-406-manifest.json | 889345ee327998bd5031442b4ea606fbc3aba337a1d463d158ab0d2d0dbd3018 |
| failed2-406-manifest.json | 9fd232d6cbd1ed7977bb0572150e4619db6faa4ff57141dd145491645c181867 |
| failed3-406-manifest.json | 7a386b6fc0069ec25724429465cfb18d10bfa3d380a04d608a15f63314f3a7c6 |

## 2026-10-05：409原外層真RET與410末態最小來源

原輸入沿官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、417根層檔與patch，合計418唯讀輸入。Go1.24.13 image SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；IDA9.4 locked-v1 image SHA-256 6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。IDA linear EA、runtime＝EA+F0000h與file offset分列。工具HEAD ee2602e8f9f7292548036670bc300224076f2b07已推送；主庫玩法與公開CPU／DOS保持。

已證實：409完整七phase與舊正常玩家前置保持，225CLI及來源／獨立數值驗證通過。原正常GAME press／selector6／release保持，新增裝置輸入0。226846742 runtime173D05／真SS ESP2BD38C讀到target174BC9，唯一下一Step226846743的ESP2BD390通過；不能拿它取代外層。227146859 runtime17651B返回EAX6與ESP2BD4FC，原191830／191A08／191A10 writer實際寫0／8／0。227148164真外層RET槽2BDBE0讀到1004BC，227148165同CS／SS及ESP＋4通過；227148175原104A6 CALL8012F，下一Step到17012F、真return1004AB與ESP−4通過。這補解較早[408框架來源](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/408-moo2-star-map-frame-source.md)的實際target；較早原收據保持，不改寫歷史失敗。

未知：230M正常上限停止在runtime21F7C1，控件建立、case0與真正7DD77正常reader未到。實際原PNG人工檢視仍為星圖，沒有GAME選單，與numeric verifier分開；正常存讀與remake同狀態未驗。[409完整觀察契約](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/409-moo2-game-outer-frame-continue.md)保持DRAFT，不用外層返回縮小完成條件。沒有panic、CPU stop或step error證據。

[410最小來源](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/410-moo2-menu-frontier-source.md)兩窄IDA678列／609EA／89fixup差異，原file bytes與IDA重定位bytes分列；2object／365page／51363fixup records核對。原84BC4 CALL87BAE與84BC9返回定位核對。原12F7C1屬sub_12F578的byte-copy循環，原入口與RET框架支持末態EBP2BD934／ESP2BD888、真RET槽2BD948。實際helper target／caller未知；static menu dependency80211→7EDF2→12F578不等於當次caller，不深入格式或renderer內部。

[411只讀續行](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/411-moo2-game-frontier-continue.md)經DRAFT→來源／只讀審查→READY，尚無411 Go或guest。先重生並完整凍結409的230M terminal與15phase，再只讀當次真RET槽及原target，按trueSS／SP握手追自然返回；新有界240M不改寫409的上限與原收據，不以步數推論會到reader。主庫RE-first保持。

原418檔與SAVE10／MOX保持。原session79227殼層exit0、15個新phase唯讀與原Code16／LE fixups核對；cgroup峰值1479024640bytes、oom／oom_kill增量0。六份生成器產物在容器暫存區逐bytes重生一致，沒有為畫面重跑guest。371份本輪私有收據索引workplace/411-current-receipt-index.json；所有原檔、Go／PNG／JSON／LOG Git忽略。兩IDA殼層exit0／idat_exit1、非空JSON／5365函式、固定hash與UID1000核對；全部handle terminal，Docker清理完成。

| 本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-409.go | 0f06450a25d71fdeebdccb9ba0b13695d3d9ff21c93d4c036d06d785daa9fa24 |
| new-game-409-outer-events.json | a12e2882d1b2a429bf2f75ee5c9da7269b6cb6fdaa86d264335d5f0b7dba3e74 |
| new-game-409-outer-terminal.json | 84af7158ef52528a32c1279e5287bbe3197864540121ab8d643e1cde82479f19 |
| new-game-409-verification-result.json | 7a8e8beaa4499869a258d6c79f4b3111f114c615de8d26434e55685e552ecd95 |
| new-game-409-visual-review.json | 09cd80a7e9238adfad7b441daa762917af92e7c21c96b22f6668c77870ce8da6 |
| new-game-410-source-result.json | 5f42165fce08afced0c13c4da4945c911768a549ac834e383ae78a0d1bdf1c6f |
| new-game-411-ready-review.json | 2a8811caa0a62aa1b2fe178f8ad1c9c619da218947d7a6deb458ea84138a44cc |
| moo2-409-writes-place-return-outer-terminal.png | 16f963bae4e60486d59eda15cca5b0747ecf34b8e68bc54fe4c6f777f0577594 |

## 2026-10-05：411原GAME正常輸入與412儲存按鈕來源

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f保持。IDA9.4 linear EA、dosgolem_high_le runtime＝EA＋F0000h與file offset分列；本輪沒有新IDA查詢。Go1.24.13 image SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac；工具HEAD 193390f1d70ba1cd34c784a18734e142d2fc378e已推送，主庫玩法及公開CPU／DOS保持。

已證實：[411原GAME續行](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/411-moo2-game-frontier-continue.md)完整凍結409的230M原末態與15phase後，230237065 runtime21F7E5真SS slot2BD948讀出16EE5E，230237066同CS／SS及ESP＋4自然返回。此實際返回與原7EE59 CALL12F578定位相符，補解410的未知caller，不追helper內部格式。236250619原8028F CALL7D061、236253032原7D891真RET、236253039原802AE CALL7DD41與236253169原7DD77 CALL1171AB全部唯一下一Step握手通過。236253170 runtime2071AB、ESP2BD704、trueSS return16DD7C為第一正常GAME輸入點，立即早停。

原11項／605bytes控件表、物件4516E0及binding保存；原PNG人工檢視可見GAME面板、SAVE GAME／LOAD GAME／RETURN，與數值驗證分開。235CLI、11個可反轉patch、原getter／唯一Step、零新裝置輸入通過。完整較早前置與原418檔／SAVE10／MOX保持；唯一原session60197 exit0，12phase全只讀，六份生成器產物逐bytes重生一致。cgroup峰值1501089792bytes，oom／oom_kill增量0。

已證實：[412正常SAVE GAME契約](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/412-moo2-menu-save-normal-input.md)核對原7D1BB物件+2Ah writer及當次3；SAVE矩形184,68..274,95，GUI229,81／device458,81，九個中心／端點／界外first-hit案例通過。原7DF12比較binding，7DF18比較word1919E4是否0，7DF20非零到提示；零才經7DF29寫mode3。原802CC CALL7E154、返回802D1來源已核對。

未知：當次enable word尚未取樣；412執行前必須只讀，非零就拒絕，不能注入0。DRAFT→來源／契約審查→READY，尚無412 Go或guest。正常SAVE press／release、case3真CALL、存檔頁／檔案內容、讀檔及remake同狀態未驗。411限定原正常選單入口CONFORMED，409原230M完整契約仍DRAFT，主庫RE-first保持。

既有Docker與UID1000、network none、原輸入唯讀、240M上限及3GiB／GOMEMLIMIT1GiB保持。全部guest與驗證handle終止，沒有新image或多跑guest。本輪373份收據由workplace/412-current-receipt-index.json索引，原EXE／PNG／Go／JSON／LOG忽略。較早371／110／45／53／39份及972份失敗原guest hash保持。

| 本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-411.go | f789844c77cc8cc0e6651d6ec4726647300b4590fdf1e480d540d79e763c49a2 |
| new-game-411-frontier-events.json | dd32a754b6fe38627153053e365a65ec7ba6e6865a2f3b672d4ffb8a26180e4c |
| new-game-411-frontier-terminal.json | ec239a46d0c0bea9edfbdeb968e11d717a5a8798776ccf7d280bf9b7b209a1f0 |
| new-game-411-verification-result.json | 9b06795bffe062d0e8e074c1fbb0d9ce77258cc8e0f2fce855892a8f3e1bb803 |
| new-game-411-visual-review.json | 953138efbb97403457922cda4c50cc0d7d6d0dbbcc5518a22ba3f3bbb6f38f8e |
| new-game-412-source-result.json | 0785d7c8072a5cd0ddf195ed6349a4763afee4947cad9b0ae3f474625179e36c |
| new-game-412-ready-review.json | c7af27ff1b32245d40420ea1da8667b3b429ccc6603c74f10446285c7e8a9698 |
| moo2-411-writes-place-return-frontier-terminal.png | 9c4e1dd9c34470082332b5eeda274a005d4a37f7cdc01e3ed22a63352cad8073 |

## 2026-10-05：412原SAVE消費與414平台查詢缺口

固定官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；IDA9.4 linear EA、runtime dosgolem_high_le＝EA＋F0000h與file offset分列。Go1.24.13 image SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac、IDA9.4 locked-v1 image SHA-256 6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780保持。工具HEAD 9e4c7bd4e2f1baaf205fbbb4bf680001f9270189已推送；主庫玩法保持，工具新增414保護模式只讀服務，CPU／公開probe保持。

已證實：[412原SAVE輸入](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/412-moo2-menu-save-normal-input.md)完整411／409及祖先保持，actual enable0、236253170正常press、236263780原selector真SS／下一Step返回3、新GUI229,81。236264659正常release、43132µs／callback24與idle；236282812原reader真SS返回16DD7C、EAX3／ESP＋4，原7DF29逐Step寫mode3。245CLI、9patches／唯一Step／原getter、兩個裝置呼叫、15只讀phase與原Code16／LE fixups通過。

已證實的限制：238069860原runtime219E75 bytes CD21、AX4300／DS188h／EDX2BD904未支援；拒絕後219E77，原檔名未知，沒有到存檔子入口。outer124與probe0分列，probe記錄cpu_stop／step_error；after resource／gzip／state manifest完整，沒有panic或OOM證據。原PNG人工仍為GAME面板／游標在SAVE，存檔頁未到，與數值分開。412仍DRAFT，389份產物及manifest保存；原418檔／SAVE10／MOX及覆蓋層保持，cgroup峰值1518264320bytes、oom增量0。

[413原存檔子頁來源](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/413-moo2-save-entry-input-source.md)窄IDA225列／135EA／26fixup獨立核對，入口7E154、控件CALL7E1E6→7D061及第一reader7E1FD→1171AB已錨定；本篇沒有新guest，不研究平台helper內部。

[414平台修正](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/414-moo2-protected-file-attributes.md)採Microsoft出版《The MS-DOS Encyclopedia》[Function43H公開契約](https://www.pcjs.org/documents/books/mspl13/msdos/encyclopedia/section5/)及既有008／findFirstExact普通檔archive20h近似。只支援MOO2保護模式AL0的真provider唯讀查詢，完整EDX與錯誤／非輸出狀態保持；不擴張AL1、FD2與實模式。12案例／三套件／核心與命令程式建置通過，僅工程CONFORMED，原FAT／當次路徑與原玩家續行未驗。測試fixture及build範圍兩種驗證問題各留失敗收據，不列產品缺陷。

[415下一原續行](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/415-moo2-save-attributes-continue.md)經DRAFT審查至READY，明示受版控le_startup.go為固定archive建置輸入，凍結原成功14phase，不重寫原拒絕。取樣真正DS:EDX／檔名／CF返回、240M上限保持；外層1200s／state1150s／kill-after15s只補工具時間邊界。尚無415 Go或guest，正式存讀與remake同狀態未驗。

本輪804份私有收據由workplace/415-current-receipt-index.json索引；較早373／371／110／45／53／39份及972份失敗原guest hash保持。原檔與私有Go／PNG／JSON／LOG忽略，所有handle terminal，Docker清理完成，無新image。

| 本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-412.go | 6ad43493b3adc0ff46669e1299e755f243f57ef1e8d138cb1137ec4a0f609a67 |
| new-game-412-save-events.json | 5fbf3cc32cc79026922a82638cf79572a0b6b7e549329673e191957f8e634b8f |
| new-game-412-save-terminal.json | 3f24102813ede84f53c9672c12908b9c1fe14c6a596648e2a4a5f1dc648a4e1b |
| new-game-412-verification-result.json | 6ca88e6a4a7cff737bc6c5fa914eb7af0141bd161248f62b750d88f893ee4b2a |
| new-game-412-visual-review.json | de4f3d69702580922a9d0ff13bdf3e490ceffe9c1ad41e94e6bb03614a6975bc |
| failed1-412-manifest.json | c38274d70df5609fa38960331e85cc2a08dcbda1e5688b584734e85d3274414b |
| new-game-413-source-result.json | 52668ebda334c699bc18b40d6ede692a9f69bee41474538d23dec0444d255b5a |
| new-game-414-target-tests.txt | 715ef67effe156f390638c152f682984c6b2afe340708536a3ca0337b3f27725 |
| new-game-414-package-tests.txt | ac9d04edb03ccd93481250fcdc3ebf8a985400901d4e2b8cbebc1d78a0383355 |
| new-game-414-build-tests.txt | 4a36a627307b672c6c2c43474a1f18b4bfa1861510af260a2beecd383944efe0 |
| new-game-415-source412-prefix.json | 659439b2d39da39d0b1cc59f2e4219a4781faaa855cf324c505c889b53710af4 |
| new-game-415-ready-review.json | dd3181243447c5e2b1baee488d6d3536ce2db7a92841f6ff891390721de57b11 |

## 2026-10-05：415缺檔分支與原存檔入口

工具HEAD 2137e39c4a599d8c64c1369492f015014739901f已推送原分支。415經READY後完成私有觀察器，7個精確反轉patch／唯一CPU.Step／原getter、255CLI含209拒絕及46正對照與六份逐bytes重生通過。固定archive明示加入公開414平台輸入，hash 0d1860f0c22c7583e25865cfa5697dc54b11061efb0a38b8e1feceb85f40b90d核對。前置建置未啟動guest，只有一次原session56037；outer及probe exit0。

完整411／409、406及所有較早前置、412成功14phase保持。238069860原runtime219E75／AX4300、DS188h／EDX2BD904，NUL路徑SAVE1.GAM，bytes 53415645312E47414D00。此檔在原417根檔及patch／覆蓋層中不存在；238069861唯一下一Step到219E77回AX2／CF1，CX與全部非輸出暫存器／segment／RAM／裝置保持。原版處理缺檔後繼續，沒有清CF或預填成功。

238113911原runtime1702CC CALL，238113912真SS到16E154、ESP−4／return1702D1；沿原SAVE press／selector3／43132µs release、reader返回與mode3 writer，沒有新增裝置輸入。獨立驗證重建LE bytes／fixups，18個SAVE及4個屬性phase全只讀；實際檔案集合與前後狀態通過。原PNG人工仍為GAME面板及SAVE游標，存檔頁尚未繪製。415僅此契約CONFORMED；412原拒絕與DRAFT不改寫，414普通檔20h仍是平台近似。主庫玩法及公開CPU／probe保持，正式存讀與remake同狀態未驗。

本輪驗證入口：

```text
python3 workplace/new-game-415-generator.py
bash /tmp/415-preflight.sh
bash workplace/new-game-415-run.sh > workplace/new-game-415-run-output.txt 2>&1
python3 workplace/new-game-415-implementation-source-verify.py
python3 workplace/new-game-415-verify.py > workplace/new-game-415-verify-tests.txt 2>&1
```

沿Go1.24.13既有image、UID/GID1000、network none、唯讀原ZIP／patch、3GiB／2CPU／128pids／GOMEMLIMIT1GiB；工具外層1200s／kill-after15s、state1150s，虛擬時間與240M上限保持。cgroup峰值1841967104bytes、oom／oom_kill增量0；418原輸入與SAVE10／MOX保持，覆蓋層無新差異。所有handle已終止，沒有重跑原guest。私有收據索引workplace/415-continuation-receipt-index.json共1195份；較早804份及其祖先、972份較早失敗原guest保持。原版資料與Go／PNG／JSON／LOG仍忽略。

下一步依[413來源](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/413-moo2-save-entry-input-source.md)，建立415完整入口到第一個正常存檔頁reader的窄觀察規格；不得以入口替代頁面或保存驗收。

原1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；Go1.24.13 image SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。IDA linear EA、dosgolem_high_le runtime＝EA＋F0000h、file offset分列，未開新IDA分析。本輪結果是原版平台邊界與玩家CALL的已證實證據，完整保存／讀取及remake同狀態仍未知。

[415限定契約](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/415-moo2-save-attributes-continue.md)與412／414回鏈已更新。

| 本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-415.go | 63237795ea000ded0675c7886738f1196db1877cabdc9cc1961eb95a91eaadfb |
| new-game-415-save-events.json | 1a2023911f78c1eb3e5bf610c8b0ae2f9f747e878fa32c8f303fdc8c1e2a2d69 |
| new-game-415-attributes-events.json | 3dedcd3ed3dd67641d50e79e43f7d48f5bb0a283725c5f4e374205203e923fce |
| new-game-415-verification-result.json | b5118f6400b9394f2dfdb962ea0bee2603aec5ef747ead9a2c88b6ca63fbf975 |
| new-game-415-visual-review.json | 43bc045f7c2a976ac04e3046d52cdc018a3f8f72715f75c89fba52dd19ab8c58 |
| new-game-415-conformance-review.json | 77e85054610445b407a356a72d611b8af0d821d400fbc3918de2102b745d58bb |

## 2026-10-05：416控件／reader已驗，417原顯示待驗

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

官方1.31 EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f；Go1.24.13 image SHA-256 1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac、IDA9.4 image SHA-256 6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。IDA linear EA、runtime＝EA＋F0000h及file offset分列，原操作與seed條件不更改，不把固定日期當RNG seed。

[416限定驗證](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/416-moo2-save-page-input-continue.md)與[417唯讀顯示契約](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/417-moo2-save-page-display-continue.md)保留畫面與完整存讀限制，413／415已回鏈。

| 本機收據 | SHA-256 |
| --- | --- |
| moo2-colony-return-416.go | d49ea3616cee52bf17638089f866c956e2f06b5387cfae691b10b6c6552205e9 |
| new-game-416-save-page-events.json | 5b039dacf6d3045206f490ded8224634572a4a0f95bb813d61d055d442a8482d |
| new-game-416-save-page-terminal.json | bfb51eaa961768b594d3379e216519842ad70fc798161bcff987928aaeabc599 |
| new-game-416-verification-result.json | a6ca744d153442ddea68545ec9fd2b77edb78a864f886f704f3eb30c9195915c |
| new-game-416-visual-review.json | 0d1d811874c9ca19d1a2b0eca9c240b00fb0a3c446514b75c9b5c5ecad6a89bb |
| new-game-416-dispatch-source-result.json | 9d9f7bd8797cd2039d82051a3f3924ffc05cc377c04e6f1539bfe1686df3e3b8 |
| new-game-417-source-result.json | 45226fa95d65df8174aff8a9873e98046f8c89eb0b0b510a420319098fe47886 |
| new-game-417-ready-review.json | 266e7e008a4a867c4782cac095fdb2daa9f2b48dc252536eaae34f2d7596d8aa |


## 2026-10-05：417原存檔頁首次顯示

### 已證實的原版續行

輸入為官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f、正版ZIP根層417檔與patch覆蓋共418輸入。Go1.24.13工具及IDA Pro9.4既有來源；IDA linear EA、dosgolem_high_le runtime＝EA＋F0000h、file offset分列，重定位獨立核對。私有原版觀察器與生成器入口在既有workplace/dosgolem/workplace/new-game-417-generator.py／run.sh／verify.py，公開契約在工具docs/spec/417-moo2-save-page-display-continue.md及000-index.md。

唯一session74131 outer／probe exit0，完整416 terminal與9phase、415及全部祖先保持。238143403在runtime207261執行C3，真SS槽16E202；238143404唯一下一Step自然返回、ESP＋4、實際EAX0，未強制結果。238143422／455／187171觀測原16E498／16E49D／16E4A2 CALL，不推定callee繪圖語意。238251948第一VBE DisplaySets107→108，末態runtime228CA9；8個phase全只讀，零新裝置輸入，240M上限保持。

實際PNG SHA-256 ee59ac6fdce06a1a7e391ea30607abc2ceef5789355e178f0a23c07d81ebfb67，人工view_image可見九個empty slot、Auto Save、SAVE／CANCEL，原星圖與GAME標題保持。原416停在首次reader之前的GAME圖片保存其實際時點，不覆寫舊證據。VBE更新與頁面可見分別核對，不宣稱remake逐像素對拍。

### 原始定位與未知

當次原DS:284038保存pointer4516E0，0x234原bytes已捕捉；相對+38h十個word為1..10、+7Ch為11..20、+232h為21。25×55byte控件表仍為hash0400259e495af6e8bf475e1d5c63e55118459ca94f0bfc9a716a3b959c44b1e1。原offset／pointer／bytes與比較consumer已證實；控件矩形與畫面可建立位置關聯，正式選格、文字編輯、SAVE結果及儲存資料語意尚未經正常輸入驗證，維持未知。

275CLI含225拒絕／50正對照、10個反轉patch、六份逐bytes重生通過。原418檔與SAVE10／MOX保持、覆蓋層無差異，cgroup峰值2012610560bytes、oom／oom_kill增量0；Docker相關容器為空，root-owned基準2437檔／272目錄與零.md目錄已核對。

收據SHA-256：

- moo2-colony-return-417.go：374fde64425b2f1407f7617c7bafd08e4ee6fbda27c1ccd36599820930631114
- new-game-417-save-display-events.json：167826fde33e1cca1647602118901831b09c009c1f416867db179c05f8879184
- new-game-417-save-display-terminal.json：c83f5d47d8099f4cb25846ea7c0d6b3bf17c1dc350385fb2901872be95acbcea
- new-game-417-verification-result.json：ad12357f0391004ae74db649a52456873d28dd33b1c3f33c3970c42652f26eb8
- new-game-417-readonly-bindings-review.json：fd91191cb66db8f44076de84697f2ed5e893423dbab39942cc16095c9747bd1b

完整收據索引workplace/417-continuation-receipt-index.json共2033份，原417-current的1619份hash保持。來源、EXE、私有Go、JSON、PNG與state本機忽略，公開僅自撰工具文件。工具HEAD 0cd0bc3b13ae7b068dcc07cde077685a7eddc442 已推送。417 CONFORMED只限首次原顯示與實際存檔頁可見；正式選格／命名／存讀與remake同狀態未驗，主庫玩法RE-first保持。下一步只核對第一空格kind11與兩組ID的輸入consumer，再審查正常有界輸入。


## 2026-10-05：418原存檔格與名稱consumer

### 輸入與已證實資料流

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f。IDA Pro9.4 linear EA、dosgolem_high_le runtime＝EA＋F0000h、原file offset與LE relocation分列。一次窄IDA保存原sub_7E154的260指令；通用input從原379來源選取，所有915列／891EA／151fixup獨立原EXE核對。原函式與runtime／檔案bytes不混用，callee與runtime helper內部未追。

417當次原4516E0物件及25項表保持；原+38h／4Ah／6Eh／70h／7Ch／232h分別為1／10／22／24／11／21。十個kind11名稱控件的+18h pointer為2816BE＋37×zero-based row，+2Ah raw1E0000h。13個inclusive first-hit案例核對端點、名稱／資訊交界、SAVE與CANCEL：裝置400,54→GUI200,54→ID1；430,373→GUI215,373→ID21。

原7E226將reader低word對第一組ID，7E23A寫DX到原[EBP+6Ah]；7E240處理另一組ID。原11E3A6設word_17C4CA＝1，11E3B2將ID寫入原word[dword_17C4CE+2]；ASCII writer11D9B2寫byte_1A871C[index]及11D9BC的後續NUL。原SAVE ID21的selected有效區間0..9、37-byte stride、空／預設／編輯名稱分支與7E3F4 CALL1160B的signed參數已錨定；runtime16E3F4→10160B、真return16E3F9。

原ID／pointer／bytes／比較與writer／CALL資料流為已證實。「保存入口」的玩家用途為強推論，內部檔案writer、成功回饋及正式正常選格／命名／存讀仍未知。沒有新guest，來源CONFORMED不等於實際保存或remake對拍。

### 419工具驗證與交付邊界

419原稿與READY來源審查分開保存，固定完整417末態與8phase及原418 bytes，只允許私有原版觀察器。下一輪等待實際正常reader後，第一格與SAVE兩次正常press／release，附原local、名稱record與文字buffer，不派送ID／代寫名稱。原CALL到10160B、真SS／ESP−4與實際參數0後停止，不執行未知callee或聲稱正式寫檔。新後段245M，原417的240M停止收據保持；尚無419 Go／guest，鍵盤命名仍另驗。

工具公開入口docs/spec/418-moo2-save-slot-name-input-source.md與419-moo2-save-first-slot-submit.md，同次000-index掛載。私有入口沿既有workplace/dosgolem/workplace/new-game-418-ida-run.sh、new-game-418-byte-verify.py與new-game-418-source-verify.py、moo2-418-ida-save-slot-name-owner.json、moo2-418-source-reused-input.json與new-game-419-ready-review.json，原碼／JSON／bytes本機忽略。IDA session29387 wrapper0／idat1、非空JSON／5365函式及固定hash保存；獨立bytes session33568 exit0，來源／文件gate通過。


| 收據 | SHA-256 |
|---|---|
| moo2-418-ida-save-slot-name-owner.json | 4214bf79697605768badc8b783ec4a0cfb0a66efdac151680a287e7c40bf718e |
| moo2-418-source-reused-input.json | 03aed99847281c8fb3885cb6d35ebe3a4e910f7f861ad899c74ffee9c32c8e79 |
| moo2-418-source-byte-index.json | be2efb07b061715792c539a87d968ac2863c868093e95db0bf288f99147fb7dc |
| new-game-418-source-result.json | 9bf8779c6c8cdd960c3707faf4af4539a6ffd4ee2318d1077a9faa5a4298853a |
| new-game-419-ready-review.json | 983a6065ddaa142dbe14b20d35ba00fda35eb2fad3bd839302b1145b6c8706d6 |

工具HEAD 4cdc5c6cd801a3ed5a892a94a658fc05d7ca7560 已推送，主庫玩法與公開CPU／probe保持。完整2052份索引workplace/419-current-receipt-index.json保持較早417-continuation的2033份及所有祖先。原418輸入、SAVE10／MOX及畫面收據未改，固定日期不當作RNG seed。Docker相關volume filter為空，root-owned基準2437檔／272目錄及零.md目錄核對通過；精確HEAD保存workplace/419-final-state.json。


## 2026-10-05：419原第一格press與工具放開守衛

官方1.31 ORION2.EXE SHA-256 4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f，位址空間dosgolem_high_le。原417完整八phase／末態、PNG日期及全部祖先保持。419先通過285CLI／六份逐bytes重生及六項私有日期回歸，再以單次session1794跑原版；outer1、實際probe2，不列為保存成功。

238285407原16E1FD CALL、238285408真SS到2071AB／ESP−4／return16E202；正常400,54,1 press已送。238295560原204176 C3、238295561依真SS返回20E1AC／ESP＋4、實際EAX1。這些握手與第一名稱控件first-hit已證實，不等於原selected local已定案或editor已啟用。

同一步原mask1、callback25／25 idle／pending0、target8:2136D1、buttons1保持。私有Input額外要求mask2Bh，與419 READY僅要求消費後idle／pending0的release契約不符，遂在放開前停止。原415已驗release-before也為mask1；沒有CPU拒絕證據。原selected仍FFFF、editor active0／focusFFFF，原reader與SAVE尚未到。

獨立verifier session76204 exit0限定確認舊前置、原press／selector RET與守衛失敗。九個只讀phase、原名稱record／文字buffer／local raw bytes保存；只送一個新裝置輸入。cgroup峰值1498697728bytes、OOM增量0，418原輸入／SAVE10／MOX保持、overlay無新內容差異。PNG45262e3491d09fb7882b8538621729a68092b598b375b040888b0b0641da49d7人工可見原SAVE頁與第一空格游標，沒有成功回饋。

原DTA有兩群時間：初始2筆AC38、SAVE頁20筆ACF5，date5D44均已證實。受控測試副本固定對應元資料，原檔不動，保留完整PNG／日期比對。base／overlay歸屬為依provider建立的強推論；修改時間不證明原guest已寫存檔內容。初次「全log只有單一時間」審查失敗發生在任何改檔或guest之前，另存收據後修正分類。

436份完整419產物按failed1-419保存並由manifest釘選；原419不覆寫或重跑。私有入口new-game-419-generator.py、run.sh、verify.py、metadata-input-review／attribution-review與visual-review.json均本機忽略。公開[419](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/419-moo2-save-first-slot-submit.md)由420接替，原失敗不改。

[420](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/420-moo2-save-release-guard-correction.md)修正已READY、實作與建置，295CLI含241拒絕／54正對照、五個可反轉patch及六份重生通過。窄wrapper只捕捉原額外守衛錯誤，完整九phase／失敗末態與PNG核對後才接受mask1正常release；只刪私有collector的已另存失敗快照，不動guest。唯一新原guest已完成，結果見下節。主庫玩法、公開CPU／probe與414平台保持；正式存讀、鍵盤命名與remake同狀態未驗。


## 2026-10-05：420修正守衛後的原正常選格與SAVE

沿用同一官方1.31 EXE、dosgolem_high_le位址空間、完整417畫面及436份原419失敗。READY來源、原稿與唯一guest前輸入SHA固定；修正只影響私有release守衛。唯一session64443 outer0／probe0，獨立verifier session61489 exit0，不重跑419、不改原CPU／RAM。

238295561送原尚未派送的400,54,0 release，原reader在238311501依真SS回16E202／EAX1。238311517原16E23A MOV word[EBP+6Ah],DX，到238311518 selected由FFFF變0。下一reader CALL在238473198到2071AB，editor active1／focus1；正常430,373,1 press、238484120 selector真RET／EAX21後送430,373,0 release。238503075原reader真RET／EAX21，238505420原16E3F4 CALL，238505421依真SS到10160B、return16E3F9／ESP−4／EAX0即停止。writer與CALL已證實，保存用途仍強推論；callee一條指令也未執行。

26個只讀phase與四次正常裝置操作、完整419九phase／失敗terminal、完整417及全部祖先通過。原第一個37-byte名稱record在記憶體生成Strader, Human, 1 colony；editor buffer仍是空格文案。實際PNG SHA-256 a6a210241532bd87f46b58fd33ecf13665b4dbe8f4bfa4aeedf78c0f5c2d836a已檢視，畫面仍顯示empty slot，游標位於SAVE，沒有保存成功回饋；不把記憶體名稱當作可見結果。名稱模板對其他狀態仍未知。cgroup峰值1679069184bytes、OOM增量0，418原輸入／SAVE10／MOX保持、overlay無新內容差異；固定日期不是亂數seed。

公開[420](https://github.com/wicanr2/dosgolem/blob/codex/moo2-parity-20260930/docs/spec/420-moo2-save-release-guard-correction.md)限定正常原版選格／SAVE到callee入口CONFORMED；417／418／419與000-index同次回鏈。私有入口沿new-game-420-generator.py、run.sh、verify.py、ready-review／one-guest-ready／conformance-review／visual-review／resolution-backlinks.json及source419收據，原碼／JSON／PNG本機忽略。工具HEAD 9fef15c4974bab36a4cdbd0097ee0328c9c81830已推送。

本輪完整索引保存workplace/420-current-receipt-index.json，審核入口workplace/420-main-audit.py及結果420-main-audit-result.json；精確HEAD與清理狀態保存workplace/420-final-state.json。下一步只追原sub_1160B玩家存檔資料流與首檔案請求，通過RE與READY後才續行。正式存讀、鍵盤命名與remake同狀態未驗，主庫RE-first不變。
