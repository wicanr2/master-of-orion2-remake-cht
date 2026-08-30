# 七個玩家玩法 RE 群組收斂稽核（2026-08-30）

## 證據契約

- 輸入：`Orion2.exe` 1.31，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；IDA linear、DOS/4GW LE object #1。正式 `.i64` 唯讀，所有匯出由
  `/tmp` 副本產生。外部符號只供導覽，原始函式名、位址、bytes 與運算元仍是定位依據。
- 本頁是既有逐函式證據的收斂索引，不以摘要取代連結文件。RE 已閉合只代表 1.31 玩家可見
  資料流已達最小充分證據；remake 實作與驗證等級仍以 `remake-traceability.tsv` 為準。
- 本頁對應 immutable roots：`0x63D92`（安塔蘭）、`0xA16BF`（事件怪獸）、`0x2552D`
  （AI 外交）與 `0x233AB`（狀態播報）。

## 收斂結果

1. **隨機事件 record（已證實）**：`sub_23DFE @ 0x23DFE` 比較的是事件 16／24 固定 record
   的 state 2／6 與目標，不是 raw event type 2／6。29 種事件建立、主要 consumer 及 typed
   回寫已閉合。見 `event-colony-research-diversion-audit-20260825.md`。
2. **原版 AI 回合（已證實）**：`Compute_AI_Data_ @ 0xD3D34` 三個 cache 指標的全部直接
   consumer 已完成 census，均歸入殖民職務、建造、艦隊目的地、軍官或運輸鏈；沒有剩餘的
   泛用未知玩法群。見 `compute-ai-data-consumer-census-20260830.md`。
3. **AI 生產與艦艇產品（已證實）**：`sub_D10EE @ 0xD10EE` 的六路抽選、48 個 building ID、
   Agent／貨運艦與三種支援船 pseudo-product 轉 ship slot 已閉合。舊清單誤把 pseudo-product
   當成各自需要固定完成 callback。見 `ai-ship-products-audit-20260828.md`。
4. **安塔蘭週期入侵（已證實）**：`sub_643A0` 建立 owner 8 的一般 129-byte ship record；
   `sub_FF799/sub_EBB0C/sub_FFDDA` 以速度 1 保存並推進逐座標 route。抵達後由
   `Search_For_Battles_ @ 0xE9D62`、`sub_E8029` 與共同 battle side builder 進快速／戰術鏈。
   `sub_E87D2 @ 0xE87D2` 的 owner 8 尾端逐存活 ship 以 hull class 呼叫
   `sub_6485F @ 0x6485F`，再以 `sub_A163A` 刪除 deployed record；同一回收呼叫也出現在一般
   艦隊清理鏈，與 offensive count table consumer 交叉支持「存活艦歸還出征 pool」。raw 8
   不進 raw 10..14 的怪獸殖民轟炸分支。見 `antaran-periodic-invasion-audit-20260825.md`、
   `fleet-interstellar-movement-audit-20260828.md`、`event-monster-colony-battle-audit-20260825.md`
   與 `antaran-defense-fleet.md`。
5. **事件怪獸移動／戰術（已證實＋資料模型停止線）**：owner 10..14 共用上述逐座標 route，
   只有停泊於星系後才進 `Search_For_Battles_`，不存在另一個可追的「途中逐座標截擊」入口。
   Plasma Flux 已閉合到 96px 半徑、距離平方衰減、友軍艦艇，以及 300 筆 26-byte 飛行物的
   飛彈／戰機傷亡 consumer；怪獸共同 battle entry、快速解算與戰術特殊武器亦已有逐函式證據。
   remake 的飛彈同步命中而無在途 record，因此只能標資料模型近似，不能把它留成無限 RE。
   見 `event-monster-route-audit-20260825.md`、`event-monster-plasma-flux-spread-audit-20260825.md`
   與 `event-monster-tactical-entry-audit-20260825.md`。
6. **AI 協議與政府演化（已證實）**：`sub_2552D @ 0x2552D` 已閉合 ordered AI pair 的頻率、
   政府表、正式 policy、貿易／研究協議、納貢與四槽記憶；宣戰／停戰另由既有專題閉合。
   `Player_Gets_Tech_App_ @ 0xE4204` 對四種政府科技寫入 1／3／5／7，`player+0x89F` 的 0..7
   政體及生產、產出、士氣、BC、指揮點、間諜、同化與協議 consumer 均已閉合。見
   `npc-treaty-negotiations-audit-20260827.md`、`diplomacy-residual-fields-audit-20260828.md` 與
   `government-trait-audit-20260828.md`。
7. **狀態播報與投降（已證實＋不存在直接 caller）**：29..35 九位元組 record、dispatcher、
   29／30／31／32／34 的 caller 已閉合；完整直接交叉參照證實 33／35 setter 在 1.31 沒有
   direct caller，故結論是「契約存在、1.31 未直接使用」，不是待猜的玩法。`sub_2670A` 的投降
   gate、`sub_E4D06` pending、事件 34，以及 `sub_E4DC9/sub_E4B5F` 的殖民地、科技、國庫、
   貨運艦、領袖、刪艦與外交清理均已閉合。remake 的 33／35 與接收者選擇仍應標 trigger
   approximation。見 `event-status-broadcasts-audit-20260825.md` 與
   `empire-surrender-audit-20260825.md`。

## 停止線

- 不因函式存在卻沒有 caller，就自創原版觸發器。
- 不把 remake 缺乏原版 in-flight／逐座標資料模型寫成「原版公式未知」。
- 不把 `PARTIAL` 實作、內部測試或可重播近似升格為原版 `CONFORMED`。
- 1.50 binary、逐像素呈現、原版 PRNG 位元序與未消費 raw 欄位不阻塞 1.31 gameplay RE gate。
