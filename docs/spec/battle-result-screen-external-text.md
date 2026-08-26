# 艦隊戰果摘要外部文案與版面規格

1. `BattleResult` 逐回合記錄使用 round、敵方損失、玩家損失三個純數值，不保存完整句子。
2. 特殊固定敵方使用 typed `BattleEnemyKind`；一般帝國名仍由對局資料傳入。
3. 標題／關閉、勝敗、開戰雙方、逐回合與總損失模板只能存在 JSON。
4. 標題、敵方、結果、開戰數、最多六列回合與總損失各有不相交的 `textSafeRect`；過長敵方名稱與模板採單行省略。
5. 雙語 catalog、格式參數、typed log、來源無 `.tr(` 及最長字墨 containment 必須有測試。
