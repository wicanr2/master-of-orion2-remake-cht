# 安塔蘭陸戰隊難度分支 caller 稽核（2026-08-26）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro／Hex-Rays 9.4.0.260610，IDAPython 映像
  `ida-pro-9.4-idapython:py312-v1`
- 位址基準：IDA linear，DOS/4GW LE image
- 方法：把正式 `.i64` 唯讀掛載並複製到一次性容器的 `/tmp`，列出直接 code xref 與原始指令；
  未修改正式資料庫或原始執行檔。

## 已證實

`sub_EC15C @ 0xEC15C..0xEC2AF` 是共用的參戰者地面／陸戰隊加成建立器：

- `0xEC177` 先比較傳入的 player slot 是否小於 8。
- slot `<8` 時，真人標記 `player+0x28 == 100` 得 0；其餘帝國在
  `0xEC251..0xEC25A` 寫入 `difficulty-2`。
- slot `>=8` 時，`0xEC29F..0xEC2A8` 寫入 `2*difficulty-4`。
- 結果固定寫在輸出 block `+0x0F`，既有 `GroundDifficultyBonus` 的五級整數式正確。

直接 caller 不只殖民地入侵：`sub_EC65A`、`sub_EC767`、`sub_EC831` 會分別替雙方建立
0x37-byte 戰鬥資料；後兩者又被 `Boarding_Action_Type_ @ 0x2C129`、戰術戰鬥與殖民地戰鬥
caller 使用。`sub_DC47C` 也在建立 AI 回合資料時取用同一加成 block。

## 勘誤

官方表格的「Antaran Marines」與 slot≥8 分支證明的是共用陸戰隊／登艦資料規則，不能單獨
證明原版另有一個「安塔蘭登陸殖民地」事件。現行 remake 的安塔蘭週期入侵與母星反攻都只有
太空戰，且沒有可到達的 owner≥8 `CombatShip`／`GroundSide` typed 單位。為了消費這個常數而
新增地面戰入口會創造原版未證實玩法。

因此目前狀態應分開寫：

- 數值與原版輸出欄位：已證實，typed helper 已實作。
- 一般 AI 殖民地守軍：已有非零正常玩家路徑，使用 `difficulty-2`。
- owner≥8 的登艦／地面單位：受目前安塔蘭戰鬥資料模型限制；不是獨立功能缺漏，不可用假入口補洞。

