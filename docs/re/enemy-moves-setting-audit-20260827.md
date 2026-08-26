# Enemy Moves 設定稽核（2026-08-27）

## 問題

`Enemy Moves` 已在設定頁、`.GAM` 與 JSON 存檔中往返，但 remake 沒有玩家可見消費端。本輪要確認原版 byte 的讀取鏈，以及在精確動畫鏈不足時可實作到哪個證據邊界。

## 證據契約

- DOS/4GW 輸入 `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- DOS IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- Win95 輸入 `ORION95.EXE` SHA-256：`6e19afdc98f1aedcb8d2f974d5b658b0c855f54529bdabdde193f5266e275185`。
- Win95 IDA 資料庫 SHA-256：`c5445bf0fed8250409d5c69ebe3282853864749a9def412ad33c7290470f65e4`。
- 工具：IDA Pro 9.4／IDAPython；DOS 記錄使用 DOS/4GW 映像 IDA linear address，Win95 記錄使用 PE IDA linear address，兩者不可混稱同一位址。
- `tools/ida/audit_game_menu_popup_ui.py` 匯出設定 byte 的直接交叉參照；`tools/ida/audit_enemy_moves.py` 匯出字串、原始函式邊界、caller、bytes 與兩個敵方在途 predicate。私有執行檔與 `.i64` 不進 Git。

## 已證實

1. DOS 版設定值是 `byte_199BDF @ 0x199BDF`，`sub_127E1 @ 0x12809` 將預設設為 0。
2. 直接交叉參照只有：`sub_7EFEF @ 0x7F085` 讀入設定頁、`sub_7F14C @ 0x7F170` 回寫，以及 `sub_8A216 @ 0x8A29F..0x8A2C7` 的 Alt-F3 toggle／說明字串 `0xA8`。沒有直接玩法讀取端。
3. `assets/i18n/help.json` 保存原版 help 資源：「If ON this option allows you to see the enemy moves.」因此玩家可見契約是顯示敵方移動，不是改變 AI 行為、航速或偵測規則。
4. Win95 launcher 的 `.i64` 未找到 `Enemy Moves`／`Enemy Move` 字串；玩家設定與遊戲狀態仍由 DOS/4GW 主程式承擔。Windows API／視窗訊息內部不另開 RE。
5. 原版確有敵方在途資料消費者，但名稱不能直接當證據：
   - `sub_77F5D @ 0x77F5D..0x77F73` 由 `sub_70875`／`sub_7109B` 呼叫，透過兩張索引表取回一個 byte；外部符號稱它為 `Enemy_Ship_Going_To_My_Colony_`。
   - `sub_A13B0 @ 0xA13B0..0xA1455` 由 `sub_A080D` 呼叫，讀取在途記錄、owner、座標與字串 ID `0x133` 後格式化輸出；外部符號稱它為 `Enemy_Ship_Heading_Toward_Our_Colony_`。
   - 兩者都沒有讀取 `byte_199BDF`，故只能證明原版有敵方在途資料與訊息鏈，不能證明設定 byte 如何閘住精確動畫。

## 強推論與未知

- 原版 help、1996 年快捷鍵紀錄與後續介面說明一致把此開關描述為敵艦移動顯示；把它實作成純顯示閘門是強推論。
- `byte_199BDF` 可能經未被直接 data xref 捕捉的基址／指標消費，也可能是特定版本未完整接線的設定。靜態資料不足以宣稱原版逐幀動畫、顏色、停留時間或是否顯示整條目的航線。
- 網路輔證只用於確認玩家語意，不升格成指令級證據：
  - patch 1.50 manual：<https://moo2mod.com/manual/MANUAL_150.html>
  - StrategyWiki controls：<https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Controls>

## Remake 映射

remake 已有 AI 主力艦隊的起點、目的地、ETA 與玩家偵測判定，但沒有原版逐 frame 在途座標。故採最小玩家可見近似：

- 設定開啟時，只對起點與目的地都在玩家目前可見範圍內的在途 AI 艦隊，於星圖畫出敵方航線與移動標記。
- 設定關閉時不畫；不改 AI 決策、ETA、戰鬥或存檔世界狀態。
- 不以航線洩漏霧區目的地；全知種族沿用既有全星圖可見契約。
- 線色與 marker timing 標為 remake 視覺近似，不宣稱原版逐幀 parity。

## 驗證

- 規則測試：開／關、在途／停泊、可見／霧區目的地與全知分支。
- UI 測試：純幾何測試驗證有效／越界航線、星圖座標及 marker 隨 tick 位移；實際繪製由畫廊確認不破壞星圖。
- 正常中文畫廊抽查星圖與設定頁；畫廊固定狀態若沒有可見在途敵艦，不把「沒有線」誤報為功能失敗。
- 以同一正版資料、字型與畫廊命令比較本輪工作樹及提交 `6929049`：兩份 `30_netwait.png` SHA-256 都是 `6b73ab5550a398cf352f5e0da3eb23aeb1e8e4236747dd572a5bcc033bf133fa`，畫面內狀態指紋同為 `2728fd7f`，證實純顯示查詢未改持久化狀態。`docs/screenshots` 的舊 fallback 圖使用不同資產條件，不拿來作本輪逐位元對照。
