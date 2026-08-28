# AI 外交剩餘方向欄位稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；正式資料庫唯讀，分析
  `/tmp` 副本。
- 非破壞性腳本：[`audit_diplomacy_residual_fields.py`](../../tools/ida/audit_diplomacy_residual_fields.py)。
- raw instructions、bytes、owner 與 caller：
  [`evidence/diplomacy-residual-fields-ida-20260828.json`](evidence/diplomacy-residual-fields-ida-20260828.json)。
- 下列欄位語意不取代原始 offset；外部符號只供導覽。

## `+0x60E`：存檔輸入旗標，沒有直接 runtime writer

**已證實**：全程式只有兩個直接 operand site，且都是讀取：

- `sub_25DF1 @ 0x26073`：值為 1 時建立 reason 113 特殊宣戰候選。
- `sub_544A1 @ 0x54593`：值為 1 且來源 raw 類型為 2，或來源人口超過目標容量一半時，
  把 AI 對真人 score 覆寫為 -150／reason 114。

沒有初始化清零、遞增或一般 runtime 直接 writer。完整 Player record 仍可能由 `.GAM` 載入或
整塊複製，因此結論不是「此 byte 永遠為零」，而是「1.31 gameplay code 沒有可追的直接
producer」。remake 保留 `.GAM` raw、JSON 往返及兩個 consumer 即已達靜態停止線；不得為它
自創遊戲內觸發器。

## `+0x6AF／+0x6BF`：四槽持久外交 modifier 的後兩槽

**已證實**：它們與 `+0x68F／+0x69F` 同為方向性 signed-word modifier：

- `sub_4D78E @ 0x4D853／0x4D85D` 初始化為 0。
- `sub_4DAB2 @ 0x4DB79..0x4DC14` 每回合各加 10，正值夾回 0。
- `Adjust_Diplomat_Modifiers_ @ 0x524C3` 四槽各減 10。
- `Modifier_Diplomacy_Adjustments_ @ 0x50DAC` 以 `Random(5)` 回復並把 `+0x6BF`
  夾在 -200..200；其他槽的完整上下限依各自 raw 分支。
- `Find_Worst_Modifier_ @ 0x50FDF` 取四槽 signed 最小值，供玩家需求評分。
- `Diplomacy_Test_ @ 0x53146` 依 proposal 類別讀其中一槽，並以
  `Random(30／50)+20／50` 寫負面持久副作用。
- `sub_51078` 宣戰把雙向 `+0x6AF=-200`、`+0x6BF=-130`；相鄰正式違約 writer
  `sub_5138E` 則把四槽雙向設為 -200。
- `Diplomacy_Exchange_Technology_ @ 0x1CB4D` 另使 `+0x6AF -= 20`；
  `Diplomacy_Propose_Treaty_ @ 0x17D5B` 使 `+0x6BF -= 20`；`+0x6BF` 也有已閉合的
  封鎖積怨 producer／consumer。

四槽的 UI 正式名稱仍未由文字資產閉合，因此本頁只稱 raw modifier slot，不以推測名稱升格。
但「宣戰 writer 的 `+0x6AF／+0x6BF` 沒有 consumer」已被上述資料流直接推翻。

## `+0x6FF`：殖民破壞／易手怨值與 reason 22

**已證實**：此欄是方向性 signed word，直接鏈如下：

- `sub_DCEBD @ 0xDCFD2／0xDD128`：破壞建築或人口候選時加 10／1。
- `sub_DD13E @ 0xDD1E7／0xDD2BD`：特殊殖民地傷亡加 1／10。
- `sub_ECBF7 @ 0xECC39`：人口 ownership 變更時加入被處理人口數。
- `sub_ECF41 @ 0xECF63`：殖民地易手前固定加 10，再進 ownership 回寫。
- `sub_25DF1 @ 0x26239..0x2624B`：在接觸、方向 gate、冷卻與尚無候選成立時，
  signed 值大於 0 就建立 type 2／reason 22；type 2 是立即宣戰候選。
- `sub_51078 @ 0x51178／0x51184` 與 `sub_524FB @ 0x525D9／0x525E3` 分別在宣戰、
  停戰時雙向清零。
- `sub_27A3D @ 0x27C56` 亦把候選接收者對其他帝國的 `+0x6FF` 加入評分；該函式完整
  接收者評分另由狀態播報／投降列管理。

因此 reason 22 的 raw producer、consumer 與清除端已閉合。現行 remake 抽象 AI 戰爭是否能
在相同時序產生此欄，屬實作資料模型差異，不是原版 RE 未知。

## `+0x737`：本回合關係變動影子

**已證實**：`Clear_Diplomacy_Messages_ @ 0x5090C` 在主回合第 8 步把方向
`+0x65F` 複製至 `+0x737`，隨即清 `+0x65F` 與訊息欄。第 10 步
`Diplomacy_Growth_ @ 0x4DD6B` 只有在 `+0x737==0` 時，才讓 current relation
向種族目標自然漂移；正式戰爭的 -90 壓制不受此 gate 影響。

所以 `+0x737` 不是永久鎖定 writer，而是保留剛被清除之 pending relation magnitude，避免同一
世界回合立刻被自然漂移抵銷。全程式只有上述一寫一讀；下一回合會由新的 `+0x65F` 覆蓋。

## 收斂判定

- `+0x60E`：直接 consumer 已閉合；無直接 gameplay writer，停於存檔／整體 record 邊界。
- `+0x6AF／+0x6BF`：主要 producer、回復、夾值、最差值與提案／需求 consumer 已閉合；正式
  文案名稱未知但不阻塞玩法 RE。
- `+0x6FF`：殖民破壞、人口／殖民地易手、reason 22 與宣戰／停戰清除已閉合。
- `+0x737`：單一 producer、單一 consumer 與主回合先後已閉合。

本切片只關閉原版靜態玩法資料流，不宣稱 remake 已具備完整方向矩陣或逐事件 producer。
