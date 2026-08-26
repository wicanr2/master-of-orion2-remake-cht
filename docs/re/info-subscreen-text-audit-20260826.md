# INFO 五子畫面文字證據稽核（2026-08-26）

## 結論

- **已證實**：原版 INFO 是單一畫面下的五個子畫面；既有符號索引保留
  `Draw_History_Subscreen_`、`Draw_Tech_Review_Subscreen_`、
  `Draw_Race_Stats_Subscreen_`、`Draw_Turn_Summary_Subscreen_` 與
  `Draw_Reference_*_Subscreen_`。本輪沒有用語意改名覆蓋原始定位。
- **已證實**：玩家自備 `BILLTEXT.LBX` 的 `str_id 5..13` 依序包含五個 INFO
  標題及 `Categories`、`How to?`、`Category:`、`How to:`。來源、抽取方法與
  openorion2 呼叫位置已記錄於 `docs/tech/screen-spec-info-research.md`。
- **已證實**：原版回合摘要可見字串至少包含 `MESSAGE SUMMARY AS OF SD:`、
  `Net Income:`、間諜、同盟、戰爭、條約與進貢欄位；現行 remake 的收入拆分與
  事件列不是原版逐欄復刻。
- **強推論**：現行 Reference 的十四個分類與十個操作項目符合原版雙欄用途，
  但尚無逐項 `str_id` 或原版動態畫面證據，因此只能標為 remake 補充目錄。

## 本輪訂正

先前 `infosubscreens.go` 把中英文文案成對寫在程式內，並把專案內部文件路徑顯示
給玩家。這兩者都不是原版證據。本輪將 INFO 顯示文案改由 `assets/i18n/ui.json`
的穩定語意鍵供應，並把底部提示改成玩家可理解的「參閱遊戲手冊」。

AI 對 AI 的五級關係仍是 remake 設計；本輪只修正英文模式與外部文案邊界，沒有
把它升格為原版精確外交量表。

## 證據邊界

本輪沿用已版控的 LBX 抽取結果與符號索引，沒有新增 IDA 資料庫改名或推測註記。
原版完整 Reference 內容與 Turn Summary 動態組裝仍屬未知；在取得新的可回查
位址、交叉參照或原版實機證據前，不得把現行補充內容標成精確還原。
