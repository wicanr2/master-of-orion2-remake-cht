# MOO2 編譯器與執行期輔助函數指紋

> 日期：2026-08-27。用途：辨識並排除編譯器生成的輔助函數，讓逆向工作集中於玩家可見機制。
> 本文件的位址均為 IDA 線性位址（linear EA），不是檔案偏移。

## 輸入與工具

| 執行檔 | SHA-256 | IDA Pro | 位址空間 |
|---|---|---|---|
| `Orion2.exe` | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` | 9.4／Hex-Rays 9.4.0.260610 | IDA linear EA |
| `ORION95.EXE` | `6e19afdc98f1aedcb8d2f974d5b658b0c855f54529bdabdde193f5266e275185` | 9.4／Hex-Rays 9.4.0.260610 | IDA linear EA |

可重生的逐函式輸出位於：

- `compiler-runtime-probe-20260827.json`：DOS 版代表性 helper 的 bytes、SHA-256、組語與呼叫端。
- `compiler-runtime-probe-win95-20260827.json`：Win95 版代表性 helper。
- `orion2-compiler-inventory-20260827.json`、`orion95-compiler-inventory-20260827.json`：IDA library function inventory。

## 編譯器結論

### DOS 版 `Orion2.exe`

**已證實：Watcom C/C++ 32-bit，連結 Watcom C/C++ runtime，使用 DOS/4GW Professional。**

證據包括執行檔中的 `WATCOM C/C++32 Run-Time system`、`DOS/4GW Professional Protected Mode Run-time`、
Watcom header 路徑，以及 IDA compiler id `3`／FLIRT runtime 命中。版權字串止於 1995 只能限制 runtime
年代，**不能單獨證明精確 Watcom 小版本**；精確版本目前為未知。

### Win95 版 `ORION95.EXE`

**已證實：Microsoft Visual C++ 編譯器／runtime 家族。** PE32 i386，PE linker version 為 `4.20`，
執行檔含 `Microsoft Visual C++ Runtime Library`，IDA compiler id `1`，並辨識出
`__alloca_probe`、`___CxxFrameHandler`、`__CxxThrowException@8` 與 `__except_handler3`。
linker version 不等於完整編譯器身分；**精確 Visual C++ 版本目前仍列未知**。

## 可排除 pattern

以下函數本身不列入 remake 玩法完整度，也不需在 Go 重做。分析呼叫端時跨過 helper，繼續追蹤其前後的
玩法資料流即可。

| 平台 | 原始函式／位址 | 可重現 pattern | 分類與處置 |
|---|---|---|---|
| DOS | `__CHK @ 0x13F1E0..0x13F1F0` | `xchg eax,[esp+4]` → `call __STK` → `retn 4`；743 個 xref | Watcom 函式序言堆疊檢查；排除 |
| DOS | `__STK @ 0x13F1F3..0x13F211` | 以 `ESP`、需求量及 `dword_184A04` 比較剩餘堆疊，失敗跳 overflow | Watcom stack probe；排除 |
| DOS | `__STKOVERFLOW_ @ 0x13F211..0x13F220` | 載入 `Stack Overflow!`，呼叫 fatal runtime | runtime 失敗路徑；排除 |
| DOS | `stackavail_ @ 0x153E1B..0x153E24` | `eax = esp - dword_184A04` | runtime 容量查詢；除非呼叫端用結果改變玩家機制，否則排除 |
| DOS | `__fatal_runtime_error_ @ 0x153E92..0x153EAF` | 進入文字模式後呼叫 runtime 終止訊息路徑 | runtime 終止；排除 |
| DOS | `__chk8087_ @ 0x1640E4..0x16412E` | x87 偵測／初始化 | 啟動 runtime；排除 |
| Win95 | `__alloca_probe @ 0x5405C0..0x5405EF` | 每 `0x1000` bytes 觸碰一頁，再調整 `ESP`；20 個 caller | Microsoft stack probe；排除 |
| Win95 | `___CxxFrameHandler @ 0x542E80..0x542EBC` | 整理 frame/context 後呼叫 `___InternalCxxFrameHandler` | C++ 例外處理；排除 |
| Win95 | `__CxxThrowException@8 @ 0x5436B0..0x5436F7` | 建立 exception record 後呼叫 `KERNEL32!RaiseException` | C++ 例外拋出；排除 |
| Win95 | `__except_handler3 @ 0x5486BC..0x548779` | SEH scope table、global/local unwind | Microsoft SEH runtime；排除 |

### 比對規則

不能只因名稱像 runtime 就排除。至少同時核對：

1. 原始位址與函式邊界；
2. bytes SHA-256 或上述穩定指令形狀；
3. IDA FLIRT／library flag 或內建 runtime 名稱；
4. 呼叫端角色，確認 helper 沒有承載玩家規則。

若版本差異造成位址改變，使用指令形狀與 caller context 對齊，不以固定位址跨版本套用。Miles／AIL、
圖形、音訊與檔案中介層不是編譯器 runtime；即使它們與 remake 平台層不同，仍須先檢查 cue 選擇、
等待閘門或錯誤處理是否會改變玩家可見行為，不能整批誤排除。

## remake 停止線

Go runtime 已負責堆疊成長、堆疊溢位與語言層錯誤處理；不得照譯上述 x86 helper。只有當某個呼叫端將
`stackavail_` 等結果用於資料容量、遊戲分支或玩家可見訊息時，才重新開啟該**呼叫端**的窄切片，
而不是深挖 helper 本身。Windows API／SEH 的內部行為同樣依現代平台契約近似，不納入玩法 parity。

## 符號表勘誤

**已證實：舊 `symbols_ea.tsv` 在這段 runtime 區域仍沿用錯位配對，不可作定位依據。** 例如它把
`__STKOVERFLOW_` 指到 `0x13F220`（實為 `strcat_`）、把 `stackavail_` 指到 `0x153E24`
（實為 `__CommonInit_`）。修正後的 `symbols_fixed.tsv`、IDA 函式邊界、FLIRT 名稱及函式內容互相吻合。
舊表保留作錯誤形成的歷史證據；新研究必須使用修正版，且仍以 bytes／xref 做獨立錨定。
