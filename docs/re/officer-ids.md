# 艦艇軍官來源 ID 證據

本筆記記錄原版 `.GAM` 的艦艇軍官索引如何映射到重製資料。顯示名稱仍是介面文字，
不再被當成唯一的存檔識別。

## 輸入與定位契約

| 輸入 | SHA-256 | 工具／定位基準 |
|---|---|---|
| `openorion2/src/gamestate.cpp` | `04a758ba9ff9b3c331fb368032a0025de631c34b106b9c395e5a35fde01a50cd` | 原始碼行號；`GameState::_leaders[]` 固定陣列 |
| `openorion2/src/gamestate.h` | `5c66c053c9f9870a83ae961295f28a8d353c804b3e599237655350a0588671a7` | 原始碼行號；`LEADER_COUNT`／`Ship::officer` 型別 |
| `openorion2/src/officer.cpp` | `114b40f0bd1bff0f0b478943f5b92c952cda39ef9eed107045b1cffbebc42cd9` | 原始碼行號；畫面只把 ID 當陣列索引讀取 |

工具是專案內的原始碼檢視與 `sha256sum`；沒有把 `.asm` 的檔案偏移冒充成程式位址，
本項沒有 IDA 位址。

## 已證實的原版資料流

- `gamestate.h:27` 定義 `LEADER_COUNT 67`，`gamestate.h:1277` 定義
  `Ship::officer` 為 `int16_t`。
- `gamestate.cpp:1870-1871` 以 `i=0..66` 依序載入 `_leaders[i]`；因此領袖陣列
  序號就是原版資料的來源 ID。
- `gamestate.cpp:1724-1725` 驗證 `officer < LEADER_COUNT`，
  `gamestate.cpp:2372-2373` 與 `2400-2401` 直接以 `sptr->officer` 索引
  `_leaders[]` 取得 Weaponry／Helmsman 加成。
- `HERODATA.LBX` 的 parser 對同一筆固定長記錄保存 `ID=0..66`。這個欄位不是
  翻譯後的姓名，也不因 UI 語言改變。

證據等級：`.GAM` `Officer` 是 `_leaders[]` 序號為**已證實**；HERODATA 記錄序號
  與原版 `_leaders[]` 來源序號一一對應，是由相同固定數量、相同 59-byte `Leader`
  記錄與載入順序支持的**強推論**。

## 2026-08-11 同一輸入 IDA：`.GAM` 全局匯入已證實

本節使用目前 `Orion2.exe.i64`（SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）與同源
`Orion2.exe`（SHA-256
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`），工具為 IDA Pro 9.4
`idat`／Hex-Rays `9.4.0.260610`，位址基準是 IDA 線性位址；沒有執行原版。

- `sub_10E2F @ 0x10E2F` 在 `0x10E80`／`0x10E88` 組出 `.GAM` 與 `"rb"`，
  `0x10EE9` 讀取 4 bytes，`0x10EEE` 驗證版本 `0xE0`。領袖區塊讀取鏈位於
  `0x110DB..0x110F1`：`edx=0x3B`、`eax=dword_1930DC`，每輪 `edi += 0x3B`，
  直到 `edi == 0xF71`；即 59 bytes × 67 records。
- `sub_1160B @ 0x1160B` 在 `0x11673`／`0x1167B` 使用 `.GAM`／`"wb"`，
  `0x116E9` 寫版本 `0xE0`；`0x11821..0x11832` 以 `ebx=0x43`、`edx=0x3B`、
  `eax=dword_1930DC` 寫出相同區塊。讀寫指標、元素大小與筆數一致。
- `sub_1307F @ 0x1307F` 以 `0x13090`／`0x13096` 傳入 `0x3B`／`0x43`，
  `0x130A6` 設 `ecx=0xF71` 複製到 `dword_1930DC`，並在 `0x130B9..0x13169`
  以 `si < 0x43` 逐筆寫入／消費 `+0x26`、`+0x2A`、`+0x2E`、`+0x2F`、`+0x32`。
  這是全局 67 筆資料的初始化與下游消費證據。

因此原版 `.GAM` 直接匯入全局領袖／軍官記錄已是**已證實**，不只是目前選中的艦艇軍官。
外部索引對 `0x1930DC` 分別使用 `_leaders` 與 `_save_officers` 兩個別名；本報告保留
IDA raw 名稱 `dword_1930DC`，不以別名取代地址／資料流。重製已由 `ImportGAM` 保存
`save.Leader.Eta`／`Status` 到 `shell.Leader.RawETA`／`RawStatus`；原版任命／任期
門檻仍是匯入後的下游規則，不能由這段 `fread_` 單獨推出。

## 重製接線

- `internal/herodata.Leader.ID` 保存來源序號。
- `shell.Leader.ID` 與 `shell.Ship.OfficerID` 以 JSON 保存同一個序號；
  `AssignOfficerToShip`、`OfficerForShip` 先以 ID 與名稱交叉確認。
- `OfficerName` 仍保留，讓加入欄位前的舊 JSON 與人工編輯資料可以用名稱回退；
  這是相容策略，不是把名稱重新宣稱為原版 ID。
- `LoadGAMSession` 已直接載入原版 `.GAM` 全局狀態；本切片完成的是軍官欄位在
  重製 JSON 內的穩定身份、逐艦消費，以及 raw ETA／status 的保存。原版 `status=4`
  的 `Deassign_Officer @ 0x934CF` 每回合遞增 raw `+0x37`，達 **30** 後進入
  `Check_Officer_Fields @ 0x933F2` 清除欄位，重製由 `advanceLeaderLimbo` 接入。
  `status=1` 活躍任命的 ETA／船型重算仍標為待 runtime oracle，不把 30 誤寫成所有
  領袖的固定活躍任期。

固定護欄：`internal/shell/officer_assignment_test.go` 驗證 ID 指派、JSON round-trip
與缺少 `officer_id` 時的名稱回退；真實 `HERODATA.LBX` 測試驗證每筆 parser ID 與
記錄序號一致（需使用者自備資料時啟用）。
