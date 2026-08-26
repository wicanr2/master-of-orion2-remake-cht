# Smacker 過場玩家路徑稽核（2026-08-27）

## 輸入與工具

- 原版：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython；位址為 IDA linear、DOS/4GW LE image。
- 非破壞性匯出：`docs/re/evidence/cutscene-player-path-ida.json`；腳本沒有改名、
  改型別或寫回資料庫。

## 已證實

1. `sub_24ED3 @ 0x24ED3..0x2518F` 是片頭流程，`0x2504A` 與 `0x25185`
   直接呼叫 `sub_14DF7`。
2. `sub_14DF7 @ 0x14DF7..0x15085` 是外部符號表所稱 `Play_Cinematic_` 的完整
   播放外層；它開啟 Smacker、取得寬高、逐幀繪製、切頁、關閉，再返回 caller。
3. 播放迴圈 `0x14F8E..0x14FE8` 每輪呼叫外部符號表所稱
   `Keyboard_Status_ @ 0x12C392`；有鍵時呼叫 `Read_Key_ @ 0x12C2E1`。
   同一迴圈也呼叫 `Mouse_Button_ @ 0x124075`，非零時離開逐幀迴圈。因此原版
   同時接受鍵盤與滑鼠跳過，不是只接受滑鼠放開。
4. `sub_14DF7` 內沒有文字列印 helper 或固定提示字串的呼叫；畫面內容來自
   Smacker frame。remake 顯示的「點擊跳過／click to skip」是後加提示，不是原版畫面。
5. 外部符號來源有位址錯位：`symbols_fixed.tsv` 把 `0x14DF7` 稱為
   `Play_Cinematic_`，符合 caller／consumer；`func_names.txt` 則把名稱往後錯放。
   本文件保留 raw 位址與原始資料庫名稱，不以任何一份名稱取代定位證據。

## Remake 對應與留白

- **已證實玩法：**按任意鍵或滑鼠動作可跳過；影片播完後回 caller。
- **remake 介面修正：**移除非原版的底部跳過提示，避免文字覆蓋影片與黑邊；轉場名稱
  改由外部 JSON 提供。
- **未知／未完成：**Smacker 音軌仍未解碼；原版 `SMACKSOUND*`／`_SmackDoPCM`
  證明原始 runtime 有音訊能力，因此「Smacker 過場已完整完成」不能涵蓋聲音。
  PCM／DAC／timer 內部依專案停止線不深挖，但 sample 解碼與人耳可見的播放完成閘門仍是
  remake 功能缺口。
