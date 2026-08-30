# AI 對真人間諜事件規格

狀態：CONFORMED（未嫁禍 STEAL／SABOTAGE 可表示垂直切片）

1. 只有玩家任務成功且資產確實改變時產生事件：STEAL 必須取得科技，SABOTAGE 必須移除建築。
2. 基礎關係變化依序消耗 `Random(15)`、`Random(5)`，兩者採原版 1..N 契約，總和取負。
3. 未嫁禍 STEAL 使用 reason 1；未嫁禍 SABOTAGE 使用 reason 3。沒有 attributed third party
   時不得產生 reason 2／4。
4. `Change_Relations_` 的有效變化同時更新 raw 關係；只有絕對值嚴格大於既有 pending magnitude
   時才覆蓋 pending reason／magnitude。
5. pending 在下一回合依 `OriginalNPCIncidentMemoryStep` 推進；正式 policy 1..3、貿易、研究，
   或 AI→真人的負向納貢視為 protected agreement。AI 政府與永久違約記憶採該 AI→真人方向。
6. remembered reason 與 memory 必須保存，供 `OriginalHumanTargetScore` 消費；舊 snapshot
   缺 pending 欄位時維持零值，不反推不存在的歷史事件。

驗證：正常玩家任務呼叫端只讀 `MissionSucceeded`；該旗標只有取得科技或實際刪除建築時寫入。
無資產可改變時不消耗事件的 `Random(15)／Random(5)`；SABOTAGE 正向測試固定 reason 3、
亂數順序、raw 關係與 pending magnitude。
