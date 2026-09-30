# `dosgolem` 接入 MOO2 1.31 DOS 原版的首輪收據

日期：2026-09-30。**範圍僅是格式、載入與原版入口執行；尚無原版與 remake 的玩法或畫面同狀態對拍。** 使用者決定以 DOS 版 `ORION2.EXE` 和 `dosgolem` 進行後續對拍。

## 固定輸入與工具

| 來源 | SHA-256／版本 |
|---|---|
| `moo2_patch1.31/MOO2-1.31.en.zip` | `908d6b7b37ad580039c5d108bab2c64b28f51ba735485287d284d5f5242b98e5` |
| ZIP 內 `ORION2.EXE`，2,612,010 bytes | `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |
| 1996 光碟 ZIP 內 `mastori2/Orion2.exe`，2,644,842 bytes | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| `dosgolem` 工具 | `/home/anr2/cht/dosgolem` 起點 Git `cc1ef5611b5eb288cc489ae504f0a7a96fc526bf`；實測修改在 `workplace/dosgolem/` 的 `codex/moo2-parity-20260930` 分支，Git `5d3c806d4ba1eec9cd4376de44edbc26c9d34552` |
| 執行環境 | `golang:1.24-bookworm` Docker，Go 1.24.13，`--network none`、原版輸入唯讀、目前使用者 UID/GID |

原版 EXE 存於被 Git 忽略的 `workplace/oracle-input/`，原 ZIP 與萃取的 EXE 均未改動。只為診斷曾建立一份標明 `header-probe` 的本機合成副本，將 MZ `0x3C..0x3F` 四 bytes 改成明示的 LE 偏移；正式解析收據已由未改動原版重生，不依賴合成副本。

## 格式與載入

**已證實（原始檔案偏移）**：MZ `0x3C..0x3F` 為 `00 00 B4 09`，不能當標準 `e_lfanew`；內嵌 `LE 00 00` 位於 `0x292E4`。`dosgolem` 原有 `leprobe` 因超界 `0x09B40000` 拒絕；在隔離副本新增明示偏移入口後，對未改動原版執行 `go run ./cmd/leprobe -exe .../ORION2-1.31.EXE -offset 0x292E4`，得到：

- 2 個 LE objects、365 pages、358 個含 fixup 的 pages、51,363 筆 internal fixup；source type 全為 7，target type 全為 0。
- LE entry 為 object 1 `+0xFFF18`，`dosgolem` 重定位線性位址 `0x10FF18`；初始 stack 為 object 2 `+0x5DCD0`，線性位址 `0x1CDCD0`。此線性位址基準只屬 `dosgolem`，不可直接當 IDA 位址。
- 重定位預覽的 object 1 SHA-256：`c32b38c5a68299b4707abcd41cb0515c304f2879000c053cba551e2e3883d078`；object 2：`137ee072bb9277a7264f0677a28ecbd06ca715e5fa36e370f49637c045e1e5c1`。合成標頭副本與未改動原版結果相同。

本機隔離副本的 `docs/spec/195-moo2-explicit-le-offset.md` 與 `docs/spec/196-cpu386-moo2-word-cmp-absolute.md` 分別記錄明示偏移載入及入口首指令；兩者只對明列的工具子集標 `CONFORMED`。副本尚未公開，這兩份規格的路徑不作公開連結。`go test ./...` 全部套件通過，並不代表 MOO2 玩家流程通過。

## 自然執行停點

可丟棄原版探針從 `LoadLEAt(original, 0x292E4)` 起跑，未改寫遊戲 bytes。入口 `0x10FF18` 的原始 bytes 為 `66 3B 15 96 19 02 00`（16 位元 `CMP DX, DS:[0x21996]`）；補足該通用指令形狀後，原版自然前進到第 5,359 步。`0x10FF18..0x10FF30` 的 8 個執行站點重複，EDX 逐次加一；第 5,357 步到 `0x10FF32`，第 5,358 步到 `0x10FF39`，第 5,359 步在 `0x10FF43` 的 `POP EDX` 因 ESP 仍在 `0x1CDCD0`、無可讀 stack 內容而失敗。後續 bytes `5A 5B C3` 顯示會 `POP EDX`、`POP EBX`、`RET`。

**假說**：DOS/4GW 進入 LE 時可能另有暫存器或堆疊脈絡。現有停點只能證實 `dosgolem` 的目前入口狀態無法繼續，尚不能排除入口位址解讀、重定位或 CPU 指令語意的缺口。不能靠任意塞返回位址稱為正常啟動。下一個可重現行動是核對 MZ stub／DOS/4GW 到 LE 入口的位址與堆疊契約，再由 `dosgolem` 真檔重跑至首個玩家可見檢查點。這一輪未接入音訊、輸入、固定 seed 或 remake 同狀態比較；不得宣稱玩法 parity。

**交叉驗證（已證實）**：1996 光碟版同樣在 MZ `0x3C..0x3F` 讀到 `0x09B40000`，內嵌 LE 同樣位於檔案偏移 `0x292E4`，入口同為 object 1 `+0xFFF18`，首 16 bytes 相同。用相同 `dosgolem` 探針自然執行，也在第 5,359 步、`0x10FF43` 的 `POP EDX` 因初始 stack 頂端無資料而停止；該版初始 ESP 是 `0x1D5CD0`。因此停點不是 1.31 patch 才產生的差異，但外層契約仍未知。

## IDA 位址基準核對

使用 `ida-pro-9.4-idapython:locked-v1` 查唯讀原版資料庫的拋棄式副本。輸入為私有 `Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA 記錄的原始輸入 MD5 是 `bacb10a92454d2f9b211eb9fe67ec099`，與上述 1996 光碟版 `Orion2.exe` 的 MD5 相同。查詢輸出留在未版控的 `workplace/ida-probe/cstart.json`。

**已證實（IDA 線性位址空間）**：IDA 在 `0x10FF18` 顯示 `EB 76 57 41 54 43 ...`，並標為 `start`；`dosgolem` 的 LE 重定位線性位址 `0x10FF18` 則是 `66 3B 15 96 19 02 00`。兩者雖數值相同，指向不同 bytes；IDA 這個地址不能用來解釋前述 LE 停點。先前符號索引裡的 `_cstart_` 僅可當導航線索，不能當已核對的函式身分。後續需先建立原始檔案偏移、IDA 位址及 LE object+offset 的明確對照，才使用資料庫的交叉參照。
