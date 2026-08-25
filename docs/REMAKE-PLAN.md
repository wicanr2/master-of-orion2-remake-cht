# 銀河霸主 2 重製計畫

## 法律與完整性邊界

- 原始遊戲檔、正版資料與原版存檔由使用者提供，只讀掛載，不進 Git 儲存庫。
- 重製使用乾淨的 Go／Ebiten 程式碼；`internal/save` 的原版 `.GAM` 解析器與 remake JSON 存檔分開。
- 可寫測試、截圖與轉檔產物只寫工作樹指定輸出或容器 `/tmp`；一次性工作使用 `docker run --rm`。
- 繁體中文翻譯與開源／使用者授權字型分開處理；授權不明字型不嵌入執行檔。
- Windows API／Win95 平台內部呼叫只視為外部契約；逆向到玩家可見輸入、輸出、錯誤與必要
  時序即停止，Go／Ebitengine 以現代平台近似實作，不追作業系統內部等價。
- 進度對外採 README 三項百分比儀表板，不合成單一還原度；即時工作仍只看 `WORKLIST.md`。

## 資產與證據盤點

| 類別 | 位置／格式 | 證據 | 狀態 |
|---|---|---|---|
| 執行檔 | 私有 `ORION95.EXE`／1.31、patch 1.5 | 反組譯位址與資料表索引 | 部分規則已接，候選估值仍有未知 |
| 存檔 | 私有 `SAVE*.GAM`、remake JSON | `internal/save`、`internal/shell/persist.go` | 兩條路徑分開且有測試 |
| 圖像 | LBX（背景、精靈、字串／資產） | `internal/lbx`、畫廊與資產測試 | 已解碼並接多數畫面 |
| 音訊 | `STREAM*.LBX`／`SOUND.LBX` PCM | `internal/audio`、音訊文件 | 已接場景與音效；Docker 技術抽樣通過，桌面聽感仍需驗 |
| 手冊 | `GAME_MANUAL.pdf`、`MANUAL_150.html` | `docs/knowledge-base/manual-cht` | 逐條引用，未證實項標級別 |
| 參考碼 | `openorion2/src` | 僅作次級渲染／資料交叉驗證 | 不視為遊戲引擎真值 |

## 架構

```text
平台／UI → shell 遊戲流程／規則 → gamedata 真值與公式
        → assets／lbx／save／uifont／audio 等可重用基礎層
```

## 垂直切片與狀態

| 切片 | 原版依據 | 規則／UI／存檔路徑 | 驗證 | 狀態 |
|---|---|---|---|---|
| 新遊戲／多帝國流程 | 執行檔畫面鏈、手冊、SAVE oracle | `cmd/moo2` → `GameSession` | 畫廊、流程測試、Docker 建置 | 已接；部分數值待原版校準 |
| 殖民地經濟／種族能力 | 手冊、執行檔表、部分 SAVE oracle | `engine.ColonyState` → shell 回合／JSON | engine／shell 測試 | 已接；深層客製能力仍開放 |
| 科技研究 | 手冊、執行檔選擇結構 | `research.go` → UI／保存 | 研究測試／畫廊 | 已接；C–K 估值仍未知 |
| 艦隊／兩條戰鬥路徑 | 手冊、執行檔與 gamedata 公式；IDA `sub_4D10E`、`sub_3AD57`、`sub_3AC20` 與五個 blueprint writer | `battleVolley`、`tacticalScreen`、戰機中隊、安塔蘭要塞槽位 | 固定 RNG 測試、戰機流程測試、要塞火力測試 | 已接；敵方戰機兩條 raw 傷害式與要塞已追回火力均已接入，逐彈 runtime 參數與 raw bit 名稱仍是 oracle 差異 |
| 武器改造 | 手冊 p.115–118；飛彈速度／防禦原版函式 | 設計 → `Ship.Mods` → 兩條戰鬥路徑 | `missile_mods_test.go`、`missile_test.go`、完整建置 | remake 的 ARM/FST PD、光束／魚雷 NR、ENV/OVR 已接；原版完整攔截器路徑與精確 raw 對應未知 |
| 艦艇／殖民地領袖管理 | `save.Ship.Officer`、手冊 Command Abilities／軍官管理說明；IDA `sub_10E2F`／`sub_1160B`／`sub_1307F` | HIRE → `LEADERS` → 指派／改派／POOL／DISMISS；殖民地領袖分頁 → `ColonyLeaderNames` → 技能加成／存檔；`ImportGAM` | `officer_assignment_test.go`、`officer_management_test.go`、JSON／GAM round-trip、目標測試 | remake 已接；來源 ID 與原版 `.GAM` 全局 `0x3B×0x43` 匯入／寫出已接／證實；原版任命／任期規則仍未知 |
| 火線角設計／格子戰術 | `Relative_Bearing @ 0x32AD1`、`Relative_Bearing_XY @ 0x32A20`、`Move_Ship @ 0x3F5F1`、`Ship_Can_Deploy_At @ 0x49043`；手冊 p.127–128 | 設計 `Ship.Arc` → 16 向 `CombatShip.Facing` → bearing mask → 玩家／敵方射擊合法性 | `weapon_arc_test.go`、`tacticalturn_test.go`、Docker 玩家路徑 | 格子戰術已接；原版快速結算 call graph 不消費射界，`battleVolley` 維持抽象 |
| 多人／熱座／TCP lockstep | 畫面資產、協定狀態、原版多人架構決策 | `netplay`、hotseat、`cmd/moo2` | netplay／shell／UI 目標測試 + 兩程序 loopback TCP 抽樣 | **最低鏈與可選可靠性已完成**：共同快照、席位、玩家指令、`turn_done`／`turn_ready`、指紋分歧停機、resume token 重連、心跳／寬限、challenge-HMAC 與可選 TLS 1.3；NAT 仍需外部 relay／UPnP |

## 下一個工作清單

| 優先 | 交付物 | 必要證據 | 驗收門檻 | 狀態 |
|---|---|---|---|---|
| 1 | captain／common 領袖技能消費端 | `docs/tech/leader-officer-skills.md` 與各技能對應的 gameplay 子系統 | 逐技能接入或標成 remake 差異；管理 UI 不足以算完成 | **已完成 remake 消費端**：26 項有至少一個效果；Tactics 依原版不實作；Famous 招募機率已由 `sub_9781D` 補證並接線 |
| 2 | 英文模式安全 fallback | 真正會顯示的引擎敘述、錯誤原因與自訂名稱 | 英文 `-gamegallery` 逐張抽查；不以內部查表 key grep 取代 | **已完成抽樣收尾**：35/35 英文畫廊，未知顯示值走安全 fallback |
| 3 | 開局經濟平衡 playtest | 士氣 0 起跳、前 20 回合收入／人口／食物／工業／研究軌跡 | 抽樣探針記錄基線；未取得玩家主觀回報不改公式 | **已完成 headless 基線**：BC 50→264、人口 8→11、士氣 0% 全程，無死亡螺旋 |
| 4 | 三平台發行維護 | 最新程式碼、Linux／Windows 腳本、macOS universal CI 產物 | 重建 Linux／Windows、驗證 macOS、更新雜湊與 release notes | 目前既有包需確認是否包含本輪 polish |
| 5 | 外部音訊驗收 | `STREAM`／`STREAMHD` 曲目與場景切換 | 有音訊輸出環境逐曲聽感確認 | Docker 技術檢查完成；人耳驗收待外部環境 |
| 6 | 公網多人強化（可選） | TCP 對局目前的已知邊界 | 重連、心跳／逾時 UX、加密／身份驗證 | **已完成可由 remake 控制的部分**；NAT 穿透仍需外部 relay／UPnP，不宣稱內建 |

## 刻意差異與未證實項

| 領域 | 原版 | 目前重製 | 原因／玩家影響 |
|---|---|---|---|
| AI 選星與 AI 對 AI | 原版策略互動 | AI 選星與議會搖擺票已接；可選 AI-to-AI 戰爭／外交使用可保存的抽象關係與艦隊模型 | remake 行為已接；原版逐艦 blueprint／精確 AI 門檻仍是 oracle 差異 |
| 戰術戰機 | 獨立中隊、攔截、回收、最弱護盾面 | 玩家／敵方格子戰術已接 ID 31 第二組傷害、`sub_3AD57 @ 0x3AD57` 命中下游與相鄰 `sub_3AC20 @ 0x3AC20` Bomber 直接插值；快速結算仍保留母艦貢獻近似 | 兩個 raw 函式的逐彈 runtime 參數、敵方逐艦藍圖配置與 raw bit 名稱仍有差異 |
| 字型 | 原版／使用者正版 TTC | 不嵌入授權不明字型 | 需使用者提供字型才能完整重現 CJK |
| 原版驗證 | DOSBox／原始畫廊 | 多數使用靜態證據與固定 RNG | 未有 oracle 的項目不得宣稱完成 |
