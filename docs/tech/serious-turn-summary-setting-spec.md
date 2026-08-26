# Serious Turn Summary 重製規格

## Typed serious 條件

`GameSession.HasSeriousTurnSummaryReport()` 在以下任一條件成立時回傳 true：

1. 任一 `LastPlayerOutput.Colonies[i].Starving`。
2. `LastRebellions` 非空。
3. `LastBankruptcy` 非空，包含被迫出售、裁撤或報廢。
4. 本回合有安塔蘭攻擊／抵達結果或敵方帝國突襲結果。

前三項由官方 help 直接支持；第四項標為玩家威脅語意的強推論。一般經濟數字、研究完成與普通建造完成不屬於 serious gate。

## 畫面路由

1. `EndOfTurnSummary=false` 時通常略過摘要。
2. `EndOfTurnSummary=true` 且 Serious 關閉時顯示摘要。
3. `EndOfTurnSummary=true` 且 Serious 開啟時，只有 typed serious 條件成立才顯示摘要。
4. Show GNN Report 關閉而本回合有特殊事件時，仍強制顯示摘要，優先於前述兩個選項。
5. 從 GNN／勘查畫面按繼續時也必須呼叫同一個路由判定，不能繞過 Serious gate。

## 顯示與文字

- 摘要一旦顯示，不過濾原有一般欄位。
- 飢荒與叛亂補一列外置摘要計數，文字模板只放 `assets/i18n/ui.json`；Go 僅保存 key 與數值。
- 最多各一列，使用既有 320px 換行寬度與 19px 行距，避免新增內容穿出摘要框。

## 驗證

- shell 純規則測試涵蓋一般回合、飢荒、叛亂、破產、攻擊。
- cmd 路由測試涵蓋 End Summary／Serious／GNN 三設定的優先順序。
- 完整 Ebitengine 測試與中文畫廊抽樣不得退化。

