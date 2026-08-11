# 殖民地生產控制

本頁記錄 BUILD QUEUE 的 BUY、AUTO BUILD、REFIT 與 REPEAT BUILD。它區分手冊已證實
的規則與為了讓重製版可玩、可保存而採取的近似；不可把近似敘述成原版逐值行為。

## 來源與證據分級

### 已證實

- 官方手冊第 67 頁說明 BUY 可買下目前顯示的建造項目，資金不足時按鈕不可用。
- 第 70 至 71 頁說明七格 BUILD QUEUE、AUTO BUILD 的開關、REFIT，以及 REPEAT BUILD；
  REPEAT BUILD 不適用於 Housing 或 Trade Goods。
- 第 131 至 132 頁說明改裝必須是同一艦體，Cruiser 以上需要 Star Base 或更高等級基地；
  改裝中的艦艇不可使用，從佇列移除會毀掉來源艦。
- 同頁改裝成本為 max(2 × (新設計成本 − 舊設計成本), floor(標準艦體成本 / 4))。

以上引用的主要來源為 MOO2 官方手冊：
https://moo2mod.com/patch/GAME_MANUAL.PDF

### 強推論

BUY 的兩個可查邊界是未開工時每剩餘 PP 4 BC、完成一半後每剩餘 PP 2 BC。重製版以
max(2 × remaining PP, 4 × total PP − 6 × completed PP) 作為連續且有界的價格函式，
半機械族以 half-PP 計算。這個函式精確命中兩個邊界，但中段原版公式尚未取得可重現
runtime 證據，不能升格為已證實。

### remake 近似

AUTO BUILD 的「殖民者認為最好」沒有在本輪重開 Win95 反組譯。現行規則是：

1. 人口未滿時先選住宅。
2. 人口已滿時依自動工廠、機器人採礦廠、研究實驗室、太空港的固定順序。
3. 之後依可建資料表順序選尚未完成的一般建築。
4. 沒有一般建築可選時維持貿易品。

REFIT 的原版會從持久化設計庫挑同艦體設計。重製版目前沒有那個設計庫，故在玩家選到
停泊艦後，凍結「當前已解鎖、能塞進同艦體的自動最佳模板」。它保留艦名、軍官、經驗、
損傷與可相容的武器改造／火線角，只替換元件。這是透明的替代 UX，不是原版設計庫的
聲稱。

REPEAT BUILD 在現行殖民地模型只接受可重複的 Special：殖民船、前哨船、運輸艦隊與
地形行動。一般建築只會第一次產生長期效果，住宅與貿易品是持續模式，因此拒絕成為
重複目標。這比讓一般建築花掉 PP 卻沒有第二次效果更誠實。

## 重製版行為

- BUY 立即扣 BC，但在本回合 EndTurn 才套用建築、Special 或改裝完成效果。
- AUTO BUILD、REPEAT BUILD 與改裝工作均與殖民地平行保存，進 JSON 存檔、熱座席位、
  TCP lockstep 快照與玩家指令重播。
- REFIT 選取同星系、非航行中的戰鬥艦。來源艦排入後會離開艦隊；完成時回到原殖民地
  星系，若原艦隊已離開則建立該星系的新艦隊。
- 取消改裝會報廢來源艦；這不是重製版懲罰性新增規則，而是手冊已證實的結果。
- 原版七格佇列不變：一格目前工作加六格後續工作。

## 實作與驗證

- 規則與保存：internal/shell/production_controls.go、buildqueue.go、session.go、
  persist.go、hotseat.go、command.go。
- 介面：cmd/moo2/colonyscreen.go 的 BUY，以及 cmd/moo2/buildqueue.go 與
  cmd/moo2/refit.go。
- 規則抽樣：production_controls_test.go 驗證 BUY 邊界與 EndTurn、AUTO BUILD 保存、
  REPEAT Special、REFIT 保存／完成／取消報廢／Star Base 門檻。
- UI 抽樣：cmd/moo2/production_controls_ui_test.go 驗證 AUTO／REPEAT 熱區與改裝預覽
  走同一條 shell 規則。

這一頁的未證實部分只保留作為 oracle 差異；不阻塞 remake 的殖民地生產流程。
