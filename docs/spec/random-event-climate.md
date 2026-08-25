# 氣候改善隨機事件規格

## 適用範圍

事件 1 對目標帝國 climate 0..7 的正常殖民地適用。Terran（8）與 Gaia（9）不可成為目標；
沒有候選時事件失敗。

## 目標選擇

每次隨機殖民地抽取，依殖民地索引順序執行 reservoir sampling：第 `n` 個候選以 `1/n`
機率取代目前結果。最多重做 200 次，首度抽到 climate `<8` 的殖民地便停止；仍未命中時取
索引最小的適用殖民地。

使用事件專用可存檔亂數流。玩家、熱座與 AI 均使用相同流程。

## 回寫

選中殖民地的結果固定為 Terran，不走一般 Terraforming 階梯：

```text
oldClimate = colony.climate
newClimate = TERRAN
```

- 更新殖民地氣候、每農夫食物與人口容量。
- 更新對應星圖行星的 `ClimateID` 與顯示文字。
- 保留建築與種族造成的既有食物加成；Aquatic／Tolerant 依擁有者特性計算有效氣候差。
- 雙語訊息使用實際舊氣候與 Terran。

## 驗收

- Toxic、Radiated、Barren、Arid 都可直接變 Terran。
- Terran／Gaia 不可選；混合殖民地集合只能改 climate `<8` 的項目。
- 玩家與 AI 同步更新 colony／planet，且非目標帝國不變。
- 完整 `go test ./...` 在 Docker／Xvfb 內通過。
