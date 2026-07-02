# 對手 AI 策略(kick-off)

> 使用者指示(2026-07-02):MOO2 的 AI,**先上 GameFAQs 找相關文獻,或參考一代(1oom)的 AI 移植;沒必要不要從頭逆向原版 EXE**(有必要才 re)。
> 對齊工作順序:先用現成文獻/程式碼理解規則,逆向是最後手段(rulebook 64 精神:別在單一路徑死磕)。

## 1. 三層參考來源(由省力到費力)

| 優先 | 來源 | 內容 | 取得 |
|---|---|---|---|
| 1 | **1oom `game_ai_classic.c`**(一代 AI 逆向重製) | 3620 行、可運作的 MOO1 AI,C 原始碼 | `~/master-of-orion/1oom/src/game/game_ai_classic.c` |
| 2 | **GameFAQs / 社群 AI 文獻** | AI 難度加成、性格、決策行為的玩家考據 | 見 §3 連結 |
| 3 | **逆向原版 MOO2 EXE** | 精確數值/演算法(最費力) | **僅在 1+2 不足時才做** |

MOO1 與 MOO2 同為 Simtex 4X,設計 DNA 相近,1oom 的 AI **架構**可當強力模板,即使數值/細節需依 MOO2 手冊與文獻調整。

## 2. 1oom AI 的可借鏡架構(已查證)

- **可插拔 AI 介面**:`game_ai` 是 vtable(`const struct game_ai_s *game_ai = &game_ai_classic;`,`game_ai.c`)。
  → 我們可照此設計「AI 介面 + 可換實作」,天然支援不同難度/性格,甚至 1.3 vs 1.5 的 AI 差異。
- **回合分階段處理**(`game_ai_classic.c`):
  - **p1(戰略/艦隊)**:送偵察(`send_scout`)、送殖民船(`send_colony_ships`)、進攻/防禦/待命艦隊調度(`send_attack`/`send_defend`/`send_idle`)、部隊運輸、造防禦艦、發展撥款(`fund_developing`)、稅率(`tax`)。
  - **p2(艦艇自動設計)**:`design_ship_base` → 選武器/引擎/裝備(`design_ship_weapons` 等)。
- **決策風格**:以「當前威脅/機會 + 可用資源」逐星/逐艦隊評分後行動。

MOO2 AI 需額外處理:更複雜的殖民地管理、科技選擇、多樣種族特性、間諜、外交理事會、戰術戰鬥 AI —— 這些以 §3 文獻 + 手冊補足。

## 3. GameFAQs / 社群文獻(待精讀)

- MOO2 AI FAQ(ProBoards Challenge Takers):難度加成、AI 性格拆解。
- _The_General_ 的 MOO2 策略指南(GameFAQs):AI 行為觀察。
- rpgcodex「MOO2 for Dummies」、LP Archive 實戰解說。
- 連結見本輪回報訊息的 Sources。

> 難度語意(社群共識):Tutor/Easy 給玩家加成;Average 中性;Hard/Impossible 給 AI「AI Bonus」(戰鬥+外交劣勢給玩家)。→ 我們的 AI 難度應做成**加成係數**疊在同一套決策邏輯上,而非寫多套 AI。

## 4. 落地做法(排入 Phase 5)

1. 先精讀 1oom `game_ai_classic.c` 的分階段結構,抽成語言無關的「AI 決策流程」筆記。
2. 對照 GameFAQs 文獻補 MOO2 特有行為(科技/種族/外交/戰術)。
3. 設計可插拔 AI 介面 + 難度加成係數。
4. **只有**當數值/演算法在文獻與 1oom 都查不到、又影響正確性時,才逆向 MOO2 EXE(記錄卡在哪、試過哪些路,不寫「暫緩」)。

## 待辦
- [ ] 精讀 1oom `game_ai_classic.c`,產出「AI 決策流程」語言無關筆記。
- [ ] 精讀 GameFAQs MOO2 AI FAQ + 策略指南,萃取 MOO2 特有 AI 行為。
- [ ] 設計可插拔 AI 介面 + 難度加成係數模型。
- [ ] 標示「必須逆向才能確定」的項目清單(若有)。
