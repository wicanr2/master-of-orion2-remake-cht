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
