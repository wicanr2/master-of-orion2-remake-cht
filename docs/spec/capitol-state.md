# Capitol 持久狀態與重建規格

## 狀態

每個真人與 AI 帝國保存：

- `CapitolPlanet`：原版 `player+0x29` 的 typed 行星索引；無殖民地為 `-1`。
- `CapitolPlanetKnown`：區分合法行星 0 與缺少此欄位的舊 JSON。
- `CapitolRebuildRequired`：只由首都攻陷／移除鏈設置，Capitol 完工清除；不得因測試夾具、
  舊資料或不完整建築集合缺少鍵值就自行推導失都事件。
- 殖民地是否已有 Capitol 使用 `ColonyBuildings[colony]["首都"]`，不由殖民地順序推導。

新局在各帝國母星行星設 `CapitolPlanet`，並在母星建築集合加入「首都」。舊 JSON 缺少
known 欄位時，以仍存在的第一座殖民地作一次性相容遷移；這是舊 remake 存檔遷移，不能
宣稱原版 `.GAM` 證據。

## 攻陷與重建

攻陷殖民地時：

1. 若該殖民地有「首都」，先移除。
2. 若它是舊擁有者的指定行星，從其餘殖民地選人口最高者；同人口選較低的原殖民地
   索引。沒有剩餘殖民地則設 `-1`。
3. 新擁有者原本為 `-1` 時，把被攻陷行星設為指定行星。
4. 被攻陷殖民地的其他建築隨殖民地過戶；typed 殖民地既有長期效果一併保留。
5. 指定行星沒有「首都」且政體不是統一系時，每座殖民地士氣加入
   `gamedata.MoraleCapitalCapturedPenalty`。Capitol 完工後重算全帝國士氣。

## 建造

- 「首都」成本固定 200 生產點，不計維護、不佔一般開局建築上限。
- 只有指定行星、尚未建成 Capitol、且非 Unification／Galactic Unification 時可出現在
  真人建造選項。
- AI 在同樣條件下加入 raw 9 候選，精確分數 100；其他殖民地或統一系精確分數 0。
- AI／真人完工都把「首都」寫入該殖民地建築集合，更新 `CapitolPlanet`，並重算士氣。
- 一般自動建造應優先重建 Capitol，不能先蓋住宅或經濟建築。

## 驗收

- 新局玩家與所有 AI 的指定行星、建築集合與行星平行陣列一致。
- 攻陷 AI Capitol 後，AI 依人口與 tie-break 指定另一殖民地，原 Capitol 不轉移給玩家。
- AI 指定殖民地的 raw 9 是唯一高分候選時會選「首都」，完工後解除士氣懲罰。
- 真人失都狀態可在指定殖民地看到「首都」選項；其他殖民地看不到。
- JSON 存讀檔與熱座換席保留 `CapitolPlanet`、known flag 及 Capitol 建築狀態。
- 測試只證 remake 遵守已證實靜態規格，不升格為原版動態逐回合 oracle。
