# 古代遺骸科技事件規格

## 目標

事件 0 必須授予一至兩個明確科技 application，不得增加研究進度。玩家、熱座與 AI 使用
同一候選算法和科技授予 callback。

## 帝國集合

候選基準只納入仍有殖民地的真人／熱座與 AI 帝國。科技擁有權使用逐 application 判定，
包含主要選擇與額外授予集合。

## 光束 application

1. 對每個帝國從原版武器 ID 0..39 中找已知、類別為光束且最大傷害最高者；初始基準為
   武器 ID 3 Laser Cannon。相同最大傷害保留較小 ID。
2. 從各帝國結果找全銀河最大傷害基準；相同傷害保留帝國順序先出現者。
3. 掃原版武器 ID 1..39，候選必須是光束、具有非 Xenon topic、目標未知，且
   `0 <= candidate.maxDamage - benchmark.maxDamage <= 50`。
4. 差值小於或等於目前最佳時更新，故一般同差由較大 ID 勝出；若候選 ID 正是基準 ID，
   更新後內部最佳差強制為 1。
5. 有候選才把其 application 放入授予清單第一格。

## 護盾 application

依 Class I、III、V、VII、X 的順序找全銀河最高已知護盾。若不是 `TECH_NONE` 且目標未知，
附加於授予清單；不得重複第一格。

授予清單為空時事件不適用。

## 回寫

- 依清單順序逐項呼叫共用 application 授予，保留同主題多來源科技與特殊 callback。
- 不修改 `ResearchProgress`、目前研究 topic 或預選 application。
- 真人／熱座更新玩家艦艇設計；AI 更新自己的 blueprint。
- 雙語事件訊息列出實際授予的科技名稱。

## 驗收

- 純算法測試覆蓋最強光束、傷害差 50／51、未知條件、同差覆蓋與最高護盾。
- shell 測試證明玩家和 AI 取得 application、研究進度不變，且無候選時拒絕。
- 完整 `go test ./...` 在 Docker／Xvfb 內通過。
