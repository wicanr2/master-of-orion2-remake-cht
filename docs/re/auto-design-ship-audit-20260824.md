# Auto_Design_Ship_ 靜態稽核

日期：2026-08-24

## 問題與舊歧義

`WORKLIST.md` 把 `Auto_Design_Ship_` 與 AI 自動設計並列，但專案另有殖民地
`AUTO BUILD`。三者不是同一功能。本次確認原版函式的 caller、輸出 record、八類分支與
直接 helper，避免拿殖民地建造政策或抽象 `FleetStrength` 冒充艦艇 blueprint。

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址基準：IDA linear，DOS/4GW LE image
- 探針：`tools/ida/audit_auto_design_ship.py`
- 操作：正式 `.i64` 唯讀掛載後複製到一次性 `/tmp` 工作目錄；不改名、不套型別、不儲存正式資料庫。

外部符號名稱採 `symbols_fixed.tsv` 的修正索引；`symbols_full.tsv` 在這段有相鄰符號錯位，
例如會把 `0x616A5` 錯標成 `Add_Beam_Weapon_To_Design_`。本文件一律保留 raw 位址與 IDA
原名，外部名稱只作導覽。

## 已證實

### 函式邊界與 caller

- raw `sub_616A5 @ 0x616A5..0x625CA`，外部修正索引為 `Auto_Design_Ship_`。
- caller 位於 `0x55088`、`0x550C1`、`0x550F9`、`0x5646B`、`0x57187`。
- `0x57187` 位於 raw `sub_57112`；修正符號索引為 `Update_Player_Ship_Designs_`。
- `0x55088` 先以 hull raw 0 呼叫，`0x550C1` 以 1..4 迴圈呼叫，`0x550F9` 以 raw 5 呼叫。
  目標指標依 `player*0xEA9 + 0x326 + hull*0x63` 計算，因此每筆輸出為 99 bytes，六種 hull
  各一筆。這證明它是玩家設計庫產生器，不是殖民地建造佇列。

### record 初始化

入口暫存器資料流證實：`eax` 是 player、`edx` 是 99-byte design 指標、`ebx` 是 hull raw。
函式寫入：

- `design+0x10 = hull raw`
- `design+0x11 = 0`
- `design+0x12`、`+0x13`、`+0x14`、`+0x15`、`+0x16` 分別由 raw
  `sub_5679E`、`sub_56726`、`sub_575D6`、`sub_5680D`、`sub_5685F` 取得。
- 修正外部索引確認 `sub_575D6 @ 0x575D6` 是 `Get_FTL_Speed_`；其餘四個欄位依附近
  `Best_Ship_*` 函式與 record consumer 可作裝備欄導覽，但本輪未逐一升格語意名稱。
- 武器槽自 `design+0x1C` 起、stride 8，helper 在加槽前檢查槽數 `< 8`；單槽數量超過
  99 時拆成多槽。飛彈／戰機 helper 以 20 為拆槽單位，炸彈與光束以 99 為拆槽單位。

### 八類設計分支

`0x61886` 比較 raw type 0..7，`jpt_61895` 的八個入口為：

| raw type | 入口 | 已證實的主要 helper 呼叫 |
|---:|---:|---|
| 0 | `0x618EF` | theme special、missile、beam×2、bomb |
| 1 | `0x61CF6` | theme special、fighters、beam×2 |
| 2 | `0x61B51` | theme special、missile、beam×2 |
| 3 | `0x623DC` | theme special、額外 special 表、missile、beam×2 |
| 4 | `0x61E99` | theme special、fighters、beam×2 |
| 5 | `0x6222D` | theme special、額外 special 表、beam×2 |
| 6 | `0x6203F` | theme special、另一 special 表、beam×3 |
| 7 | `0x6189D` | 先嘗試固定 special raw 63，再進 type 0 鏈 |

直接 helper 的修正外部索引如下；名稱是導覽標籤，證據仍是同列 raw 位址與指令：

- `sub_5FA87 @ 0x5FA87`：`Get_Ship_Class_`
- `sub_5FE14 @ 0x5FE14`：`Add_Specials_To_Design_`
- `sub_600BF @ 0x600BF`：`Add_Theme_Special_To_Design_`
- `sub_601AC @ 0x601AC`：`Add_Fighters_To_Design_`
- `sub_60779 @ 0x60779`：`Add_Missile_Weapon_To_Design_`
- `sub_60A5D @ 0x60A5D`：`Add_Bomb_Weapon_To_Design_`
- `sub_6147F @ 0x6147F`：`Add_Beam_Weapon_To_Design_`

`sub_5FA87` 會讀 player `+0x205`，檢查是否有 fighters／special weapons／special devices，
並呼叫 raw RNG `sub_1247A0(20)`；因此八類不是「按 hull 固定輪替」。

## 強推論與未知

- **強推論**：`+0x12..+0x16` 是電腦、裝甲、FTL、護盾與燃料等自動最佳裝備欄；目前只對
  `Get_FTL_Speed_` 有本輪具名符號硬證，其餘須再追 record consumer 才能逐欄升格。
- **強推論**：raw type 1／4 是兩種戰機主題，2 是飛彈主題，3 是特殊＋飛彈，5 是特殊＋光束，
  6 是光束主題；依各分支 helper 組合命名，不代表原版 UI 名稱。
- **未知**：player `+0x205` 的完整 enum、兩種戰機 type 的差異、special 候選表每個 raw ID 的
  正式語意，以及每種 beam helper 參數如何決定 HV／PD／AF 等改造。
- **資料模型限制**：原版每個玩家有六筆 99-byte blueprint，每筆八武器槽與八特殊槽；remake
  現已有六筆持久 `ShipBlueprint`，建造會深複製八武器記錄與特殊裝置 raw ID，快速結算及
  格子戰術也已依槽序、`WorkingCount`、射界與逐槽彈藥消費。玩家 UI 可選八槽、增刪與調整
  1..99 數量，逐槽造價／空間、typed mods、命令重播亦已接線；未知 raw 武器／mods 採
  fail-closed。特殊裝置已轉為 typed 八槽，所有玩家玩法消費端讀完整集合，多種機庫亦可各自
  出擊；格子戰術已接逐武器槽可用／待命／關閉、狀態色與右鍵資訊。其中三態與待命回復有
  `Draw_Weapon_Status_Display_ @ 0x2E2CD` 與 `Do_Combat_Turn_ @ 0x42F7F` 的 IDA 證據；左鍵循環方向與右鍵彈窗外觀是 remake 操作轉接，不宣稱原版精確。2026-08-25 再依 `Load_Combat_Ship_ @ 0x4954A` 的八槽載入鏈完成逐槽 PD；紅色關閉仍自動迎擊飛彈／戰機。未解 raw mods 與 AI 仍未完成；AI 仍是
  2026-08-25 續查 caller 後，一般 AI 已依 hull 0..5／role 0..5 保存藍圖，科技完成只重建
  hull 0..4，並接上實艦生產、快速／格子戰術與快照。原版精確 AI 生產評分仍未知；部分
  消費端成立不能宣稱逐槽 parity。

## Remake 對應與驗證

- `internal/shell/auto_ship_design.go` 保存 raw 0..7 role，依已解鎖科技選擇對應武器家族、
  裝甲、護盾與特殊裝置，並以 session-aware 空間公式找第一個合法組合。
- `GameSession.ShipDesigns` 依已證實的 hull 0..5 保存六筆持久 blueprint；舊存檔依當前科技補齊。
  `cmd/moo2` 預設載入巡洋艦，點六艦體列會先保存再切換，只有底部 BUILD 會造船；元件修改、
  JSON、熱座與 TCP 快照均不再共用一份暫態巡洋艦選擇。
- 單元測試驗已解鎖、武器家族、機庫限制、艦體合法空間、逐槽開火／彈藥與三態操作；測試只證 remake 實作符合本規格，不把尚未證實的 UI 細節升格成原版逐像素一致。
- 2026-08-24 後續垂直鏈見
  [`../spec/multi-slot-build-and-quick-combat.md`](../spec/multi-slot-build-and-quick-combat.md)：建造、
  快速結算、格子戰術、造價／空間與可視化編輯已消費 typed mounts；未解 raw mods 不猜測，
  特殊裝置的原版 bitset 匯入、設計、成本／空間、建造與兩條戰鬥效果已閉合；AI 仍是 gate。
