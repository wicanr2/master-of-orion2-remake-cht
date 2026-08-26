# 戰略戰鬥動態結果文案稽核（2026-08-27）

## 可回查證據

- 軌道轟炸畫面與流程：`docs/re/colony-bombing-screen-ui-audit-20260826.md` 已以 IDA Pro 9.4 證實 `sub_B4D02` 外層、`sub_B4800` 逐幀回呼及 `sub_B4606` 炸彈記錄。
- 地面戰畫面與流程：`docs/re/ground-combat-screen-ui-text-audit-20260826.md` 已證實 `sub_B7491` 生命週期、`sub_B7289` 逐幀回呼及 `sub_B771D` 的真實 helper 邊界。
- 心靈控制：手冊規則與 `MindControlColony` 的垂直鏈已要求心靈感應種族、抵達敵殖民地及至少巡洋艦級艦艇。
- 怪獸戰術：`StartMonsterCombat` 以 typed 怪獸 blueprint 建立格子戰術雙方；本輪不改 blueprint、傷害或戰果回寫。

## 本輪 source 稽核

**已證實（remake source）**：四條玩家入口均已存在，但 `GroundBombardResult.Reason`、`GroundInvasionResult.Reason` 與 `StartMonsterCombat` 第三回傳值仍是內嵌中文自由字串，並由星圖直接顯示。

**強推論（介面轉接）**：這些中文／英文完整句子是 remake 等義提示，不是從原版字串資產逐句追回。規則 gate 保持不變，改以 typed code 連到外部 JSON；此變更不宣稱原版逐字 parity。

**停止線**：不因文案分離重開轟炸傷亡公式、地面戰即時動畫、怪獸 blueprint 或 Win95 內部呼叫。
