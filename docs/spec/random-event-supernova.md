# 隨機事件 24「超新星」規格

## 目標

依 `docs/re/random-event-supernova-audit-20260825.md`，把手冊推導的需求公式與玩家限定目標，
改成 1.31 已證實的全銀河目標、難度倒數、逐回合研究消費與全 owner 爆發結果。

## 建立

- elapsed turn 至少 200。
- 從全銀河 `0..len(Stars)-1` 每次等機率抽一星，最多 1,000 次；候選需至少一座 active
  玩家／熱座／AI 殖民地，且不能與彗星、瘟疫、人口暴增、時空異象、海盜活動或另一超新星
  同星。`colony+0x13F` 已證實為 Capitol 建築槽且已有 typed 狀態；事件星系抽選器
  尚未接入「至少一個無 Capitol 的有效殖民地」條件。
- 倒數=`roll1Based(5)+10-difficulty`；需求=`該星建立當下所有帝國殖民地 RP 總和×倒數`。
- 保存 `StarIndex`、`Countdown`、`ResearchNeeded`、`ResearchDone=0`，可 JSON 往返。

## 每回合

- 把該星所有玩家、熱座與 AI active 殖民地的當回合 RP 加入 `ResearchDone`；這些 RP 不得同時
  加入一般研究。
- 若 `ResearchDone >= ResearchNeeded`，立即成功並移除事件，不遞減倒數。
- 否則倒數減一；仍大於零則保留。
- 倒數歸零時，逐一把該星所有 active 殖民地對應行星改成 Radiated，移除殖民地與全部平行
  陣列；若該星已無其他殖民地，清除星系 owner。

## 版本與近似

- 本規格預設 1.31 profile。1.50 手冊的 6–14 敘述不覆蓋 1.31 指令；取得 1.50 binary 前保留
  版本差異待辦。
- remake 以 `engine.RunColonyTurn` 對當前 typed colony 重算 RP，對應原版已算好的 `+0xEB`；
  熱座回合架構不是原版 player[] 同步陣列，固定亂數與存檔測試需確保不重複計入目前席位。

## 驗收

1. 五級難度與 roll 1／5 的倒數邊界。
2. 需求嚴格等於全星系 RP×倒數，不得 `+1`。
3. 目標可落在 AI 或非目前熱座殖民星，且不固定目前玩家。
4. 多帝國同星時共同投入 RP；成功不摧毀殖民地。
5. 失敗時玩家／熱座／AI 殖民地與平行陣列均移除，各自殖民行星變 Radiated。
6. JSON 往返、研究轉用、全專案測試、格式、擁有權與 Docker 清理通過。
