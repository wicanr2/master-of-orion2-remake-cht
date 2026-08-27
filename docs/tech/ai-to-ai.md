# AI 對 AI 可選強化

## 目前狀態

2026-08-11 已完成 remake 端的可選 AI-to-AI 路徑。`NewDemoSession` 預設開啟
`EnableAIVsAI`；舊 JSON 存檔若沒有這個欄位，載入時保持關閉，避免把既有對局靜默改成
另一套規則。

這一層的政策選擇、資源交換與抽象戰爭仍是可玩性模型，不宣稱已還原原版
正式外交接受表、逐艦設計或每一發武器。2026-08-27 關係演化底層已改接
`Diplomacy_Growth_` 的條約成長、14×14 種族目標與逐回合漂移；原版逐艦 blueprint
與精確談判門檻仍列在
[`docs/re/oracle-static-ida-20260811.md`](../re/oracle-static-ida-20260811.md) 的未知項目。

## 接線內容

- `AIRelations` 是原版 raw current 的 `-40..40` 投影；raw 與 Known 另以矩陣保存。
  每對依原版鏡射順序保留高槽 observer→低槽 target，再形成 `AIWars`、
  `AIPolicies`、`AITrade`、`AIResearch` 四組可保存狀態。
  這些矩陣不再只供議會顯示，會進入每回合外交與艦隊決策。
- 2026-08-27 已移除平均關係 `-25／+12／+25／+8` 四個 remake 門檻。
  互不侵犯、同盟、貿易、研究與納貢改由原版 `sub_2552D` 的難度頻率、政府表、
  raw 關係／聲望／談判記憶、第三方戰爭與亂數分數建立；現有戰爭只按正式 policy
  保留，不再因顯示關係自動宣戰或停戰。原版宣戰／停戰建立函式仍待獨立閉合。
- 貿易協議每回合最多轉移 1 BC；研究協議每回合最多分享 1 點研究進度。雙方進入戰爭後，
  兩種協議都會清除。
- 每個 AI 維持一支抽象艦隊。宣戰後，艦隊選擇對手人口最高的殖民地（同人口取穩定索引），
  依星圖距離飛行；抵達後以艦隊強度、殖民地人口防禦與回合／帝國索引的確定性擾動解算。
  勝方接收殖民地、行星與建築資料，敗方保留其餘殖民地；結果寫入 `LastAIAIBattle` 供 UI／測試
  顯示。
- AI 選星本身已使用行星價值／距離模型，不是舊文件所稱的純星圖索引順序。議會也已先選基礎票
  前兩名候選人，再用玩家↔AI 與 AI↔AI 關係分配第三方搖擺票；相關說明見
  [`docs/tech/victory-conditions.md`](victory-conditions.md)。

## 保存與測試

`internal/shell/persist.go` 保存 `EnableAIVsAI`、政策矩陣、raw current／Known 與 AI
艦隊目標欄位；熱座切換會同步過濾矩陣，避免 AI 索引在換席位後錯位。

抽樣護欄位於 `internal/shell/ai_vs_ai_test.go`：

- 注入固定亂數後，原版分數可建立同盟／研究，且談判記憶正確扣 30；
  正式戰爭不因舊 `+12` 門檻自動停戰。
- 抽象艦隊抵達後能確定性結算，並把殖民地／建築從防守 AI 轉給攻擊 AI。

這條路徑不改變原版舊傳輸，也不提供 NAT 穿透；網路可靠性另見
[`docs/tech/multiplayer-architecture.md`](multiplayer-architecture.md)。
