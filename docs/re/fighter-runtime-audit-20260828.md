# 戰機炸彈／光束 runtime 垂直證據鏈（2026-08-28）

## 範圍與追溯

本輪只閉合玩家可見的戰機出擊、攻擊、改目標、返航與母艦回收；動畫、繪圖、音效裝置、
C runtime 與平台 helper 不納入玩法分母。輸入為 `Orion2.exe`（SHA-256
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`）及
`Orion2.exe.i64`（SHA-256
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）；工具為 IDA Pro
9.4／IDAPython，位址是 IDA linear、DOS/4GW LE object #1。非破壞性匯出見
[`fighter-bomb-ida-20260828.json`](evidence/fighter-bomb-ida-20260828.json)，腳本見
[`audit_fighter_bomb.py`](../../tools/ida/audit_fighter_bomb.py)。

`symbols_fixed.tsv` 把 `0x3AC20` 稱為 `Fire_Fighter_Bomb_`、`0x3AD57` 稱為
`Fire_Fighter_Beam_`，歷史 `func_names.txt` 卻互相衝突；所以下文只用 raw 位址。名稱不是證據。

## 建立 runtime（已證實）

`sub_3C892 @ 0x3C892..0x3CC24` 配置一筆 26-byte 記錄，從 313-byte 艦艇設計與
11-byte 武器槽複製：

| 偏移 | 已證實用途 |
|---:|---|
| `+0` | raw weapon／fighter kind |
| `+2` | 發射方 owner |
| `+4` | 母艦設計／戰鬥記錄索引 |
| `+6` | 目前目標索引 |
| `+8` | 目標類型；0 是艦艇，非 0 是 runtime 目標 |
| `+15` | 存活數量 |
| `+17` | 武器槽 raw flags |
| `+23` | 剩餘攻擊次數 |
| `+25` | active flag |

武器類別表 `byte_17F80F[28*kind] == 4` 時，數量改為槽內數量乘 4。raw kind 28、30 的
攻擊次數是 1，29 是 2，31 是 4。依手冊型別、武器 ID 與下游分流交叉驗證，28 為 Assault
Shuttle、29 為 Heavy Fighter、30 為 Bomber、31 為 Interceptor；此型別對照為**強推論**，
數字與分支本身為**已證實**。

## 兩條攻擊 consumer（已證實）

`sub_3D2DF @ 0x3D2DF..0x3DB8F` 是兩條 raw 攻擊函式的唯一直接 caller：

- kind 28 呼叫 capture consumer，扣一次攻擊，不進炸彈／光束公式。
- kind 30 略過 `sub_3AD57`，以 `sub_3AC20` 攻擊一次。
- kind 29 先走 `sub_3AD57`，但此處不扣次數，再走 `sub_3AC20` 並扣一次；兩次迴圈形成
  一發光束加一枚炸彈。
- kind 31 只走 `sub_3AD57`，每次扣一，永不走 `sub_3AC20`。

`sub_3AC20 @ 0x3AC20..0x3AD57` 依 runtime `+4` 找母艦 owner／設計，再由
`word_199242` 選該帝國的炸彈 weapon record。它對 runtime `+15` 的每一架擲 1..100，令
`S=max-min`：

```text
S > 0: D = min + floor((floor(100/(2*S)) + R) * S / 100)
S <= 0: D = min
```

每架結果直接送 `sub_39985 @ 0x39985` 的目標分面、護盾、裝甲與結構 consumer；兩個輸出
累加實際傷害及次要吸收值。因此它不是孤立的數值 helper。

`sub_3AD57 @ 0x3AD57..0x3AFB3` 由 `word_199254` 選光束 weapon record。每架擲
1..100；若 `R>95`，修正值直接設為 100，否則取 `min(R+OCV-DCV,100)`，沒有下界夾限。
修正值小於 40 miss；命中則為：

```text
D = min                                  if max-min <= 0
D = min + floor((max-min+1)*(M-40)/60)   otherwise
```

`M=100` 可能得到 `max+1`。艦艇目標 0 再將傷害向零除 2。艦艇路徑進 `sub_39985`；
runtime 目標則進 `sub_3A0B9 @ 0x3A0B9` 的耐久／餘數 consumer，再扣其 `+15` 數量。
既有 remake 的「96..100 保留原 roll」註解與實作因此被原始指令**反駁**；依 RE-first gate
本輪只登記差異，不修改 Go。

## 目標、返航與回收（已證實）

- `sub_3D299 @ 0x3D299` 依 `+8` 分流，檢查艦艇仍存活或 runtime `+15>0`。
- runtime 目標被擊毀後，`sub_3AD57` 呼叫 `sub_3DDD8 @ 0x3DDD8` 重新選目標並擊殺舊
  runtime。重選會掃描合法敵艦，排除已毀、不可選及目前目標，取距離值較小者；沒有合法目標時，
  一般飛彈銷毀，戰機則把目標改回母艦。
- 尚有攻擊次數但艦艇目標失效時，巨型 consumer 同樣重新選目標。
- 攻擊次數歸零後，若母艦不存在／失效、owner 不符，或 kind 28，runtime 銷毀；否則寫
  `+8=0`、`+25=1`、`+6=母艦`，進入返航。
- 返航抵達（`+4 == +6` 且目標類型 0）時，以 `存活數量/4` 掃母艦八個 11-byte 槽，依相同
  raw kind 回填可用槽，然後銷毀 runtime。
- `Check_For_Launched_Fighters @ 0x29480` 掃 300 筆 runtime；其唯一 caller 是母艦彈藥
  判定，證明在外戰機仍參與「是否耗盡武器」的玩家可見狀態。
- `Draw_Fighter @ 0x3DBD1` 在視覺抖動前後保存／還原 RNG seed；動畫不消耗玩法亂數。

## 結論與剩餘邊界

戰機炸彈與光束的 runtime 參數、數量、攻擊次數、分流、命中／傷害、艦艇與 runtime
consumer、失效改目標、返航及母艦回收已形成可回查的垂直鏈，RE 列可關閉。仍保留：

- raw flags 的正式遊戲名稱，以及 `sub_3DF8D` 若干來源欄位的 UI 名稱是**未知**；原始讀值與
  算術已保存，不妨礙重建玩家可見契約。
- 初始由玩家／AI 選定哪艘艦為目標的策略屬更上游命令／AI 決策；本輪證實建立函式照單保存
  傳入目標，以及目標失效後的重選，不把 AI 偏好猜成戰機公式。
- `Fighter_Combat_SFX @ 0x3E1A4` 與平台繪圖／音訊只保留「攻擊後播放」邊界；其內部不進
  remake 或 RE 完成分母。
