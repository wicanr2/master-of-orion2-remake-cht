# 驗證矩陣

> 本矩陣區分「重製內部一致」與「原版對齊」。綠色單元測試不能替代原版 oracle。
> 2026-08-25 README 的發行驗證 33% 由三個互不替代的閘門計算：打包工程抽樣已通過；最新
> 工作樹重打包／三平台真機及外部音訊逐曲人耳驗收未完成。Windows API／Win95 內部行為不作
> RE 驗收，只驗玩家可見契約與 remake 近似的穩定性。

| 子系統 | 單元／真資料測試 | 原版 oracle | 視覺 oracle | 正常玩家路徑 | 目前結果 |
|---|---|---|---|---|---|
| 啟動／新遊戲 | `cmd/moo2`、設定與版本測試 | 執行檔立即數、SAVE10 開局資料 | 1.31／1.5 畫廊紀錄 | 有完整流程測試 | 通過；少數開局數值待校準 |
| 回合／殖民地 | `internal/engine`、`internal/shell` 全測 | 手冊、部分 SAVE oracle | 殖民地畫廊紀錄 | 多帝國 headless 探針 | 通過內部驗證；逐格產出仍非完整原版 |
| 科技／研究 | `research_test.go`、創造力測試 | 手冊、執行檔選擇表 | 研究畫廊紀錄 | 研究→解鎖元件 | 通過；`Calc_Tech_Value_` C–K 未證實 |
| 艦隊／戰鬥／AI | shell 戰鬥、戰機、登艦測試；`point_defense_fighter_test.go`、`tacticalfighter_test.go` | 手冊 p.117／p.157、`Fighter_Pilot_Bonus @0x35EAE`、IDA `sub_4D10E`／`sub_55161`／`sub_5542C`／`sub_55738`／`sub_55B12`／`sub_55F67` | 戰術／結果畫廊紀錄 | 宣戰→戰術／快速結算→結果 | remake 玩家／敵方戰機、PD、最弱護盾面、返航與 profile 通過；原版五級 blueprint raw 槽位已證實，敵方戰機精確命中／傷害仍未知 |
| 武器改造 | `weapon_mods_test.go`、`missile_mods_test.go`、`point_defense_fighter_test.go`、`missile_test.go` | 手冊 p.115–118；`sub_3CD21`／`sub_3DFE0`／`sub_35EAE`／`sub_3E095`／`sub_3A0B9`；完整光束改造 `sub_394F7`／射程輔助 `sub_39434` | 艦艇設計畫廊紀錄 | 設計→造艦→兩條戰鬥路徑 | remake ECCM/EMG/MV/ENV/OVR、光束／魚雷 NR、ARM/FST PD 與戰機接戰前 PD 通過；原版 raw 對應與完整攔截器仍未知 |
| 火線角／格子戰術 | `internal/gamedata/weapon_arc_test.go`、`internal/shell/weapon_arc_test.go`、`cmd/moo2/tacticalturn_test.go` | `Relative_Bearing @ 0x32AD1`、`Relative_Bearing_XY @ 0x32A20`、`Move_Ship @ 0x3F5F1`、`Ship_Can_Deploy_At @ 0x49043`、手冊 p.127–128；快速 call graph 見 `docs/re/weapon-arcs.md` | `-gamegallery` 戰術畫面；射界提示由固定測試驗證 | 設計→選弧→成本／佔格→建造→保存→格子戰術射擊／轉向 | 設計與格子戰術通過；原版快速結算不消費空間射界，抽象路徑保留 |
| 艦艇軍官 | `officer_assignment_test.go`、`officer_management_test.go`、導航／工程師回歸測試、`docs/re/officer-ids.md` | `save.Ship.Officer`、`gamestate.cpp:1724-1725,1870-1871,2372-2401`、IDA `sub_10E2F`／`sub_1160B`／`sub_1307F` | 軍官畫面／艦隊畫面本輪未新增截圖 | HIRE 指定候選→指派／改派→POOL 或 DISMISS→ID／名稱存檔→快速／格子戰鬥 | 逐艦 Weaponry／Helmsman、Navigator、Engineer、HIRE／POOL／DISMISS 與來源 ID 已接；原版 `.GAM` 全局 `0x3B×0x43` 讀寫已證實，重製原生 importer 未做 |
| 間諜／逐對手外交／條約 | `spy_test.go`、`spy_mission_test.go`、`original_spy_oracle_test.go`、`treaty_test.go`、餽贈／RACES／外交畫面測試 | 間諜手冊 p.174-175；IDA `sub_100A3E`／`sub_100A83`／`sub_101483`；外交 raw 證據 | Docker `-gamegallery` 的 `15_diplomacy.png` 與 UI 熱區測試 | 種族關係→選 AI 列→切換 STEAL/SABOTAGE/HIDE→提議餽贈／特殊貿易／條約→終止→回合結算→存檔 | SABOTAGE 帝國攻防表與 slot 分層已閉合；訓練／維護與 AI 任務政策、特殊貿易表及完整 AI end-to-end 接受表仍未閉合 |
| 事件／勝利 | `events_test.go`、勝利測試 | 手冊、執行檔表 | 報告／王座廳畫廊紀錄 | 回合→事件／三勝利路徑 | 通過內部流程；部分數值仍為近似 |
| 存檔／讀檔 | `persist_test.go`、各切片 round-trip | remake JSON schema；原版 GAM 只讀 | Load 畫面已納入本輪畫廊 | 寫入可寫 overlay 後讀回 | 單元通過；本輪 Docker 畫廊含 `18_loadgame.png`，未宣稱原版逐像素一致 |
| 圖像／主題 | LBX／資產／畫廊工具 | LBX 位元組與執行檔索引 | 本輪 Docker 畫廊 37 張 | 各畫面正常載入 | 37 張均成功產生且非零；修正 1.5／30 資產目錄的 `NEWGAME#31` 越界警告；逐張原版比對仍開放 |
| 中文／英文／字型 | i18n、英文 labels、`englishSafeFallback`、lang gap 測試 | 原版烘字與譯表 | 中英畫廊紀錄 | 語言切換→完整流程 | 英文 38/38 畫廊抽查通過；未知值保留安全 fallback，13 條漢字棘輪例外為六個規則 key 與七個 dev-only 標題；另有純英文來源契約 |
| 打包 | Docker build／跨編 | 無原版需求 | 正常路徑 smoke 截圖 | 從任意目錄啟動 | Linux／Windows 重新產出並驗證；macOS 使用既有 CI 產物，Linux 容器不宣稱真機執行 |

## 本輪截圖 metadata

| 圖像 | 建置 | 存檔／情境 | 座標 | 種子／時間 | 主題 | 比對類型 |
|---|---|---|---|---|---|---|
| `normal-path.png`（暫存） | 目前工作樹／Docker 最終建置 | `-data /private/gamedata/mastori2 -game -shot`，私有資料唯讀 | 640×480，`uiscale=1`；148,252 bytes | Docker + Xvfb；容器內產生後供人工檢查 | 繁中正常路徑成功；未作原版逐像素宣稱 |
| `01_menu.png`、`01b_newgame.png`、`02`–`35_research.png`，含 `02b_customrace.png`、`15a_races.png`（共 38 張） | 目前工作樹／Docker 畫廊建置 | `-gamegallery`，含 `18_loadgame.png` | 38/38 非零；統一由畫廊流程產生 | Docker + Xvfb；私有資料 `:ro` | 中文與英文均 38/38；抽查自訂種族、RACES、外交、艦艇設計、輸入框、F9 測距、星圖、研究領域；尚未逐張原版比對 |
| `15_diplomacy.png`（同上畫廊中的暫存畫面） | 目前工作樹／Docker 畫廊建置 | 無存檔，外交畫面初始路徑 | 1280×960，預設 `uiscale=2`；36/36 PNG 非零 | Docker + Xvfb；輸出在 `/tmp` | 已檢查八格提案、固定納貢按鈕、協議摘要與終止區不與結束按鈕重疊；尚未逐像素對原版比對 |

## 發行門檻

- [x] Docker 完整測試：`xvfb-run ... go test -buildvcs=false ./...`
- [x] Docker 建置：`go build -buildvcs=false ./cmd/moo2`
- [x] 軍官指派目標測試與 JSON round-trip：Docker `go test ./internal/shell -run 'OfficerAssignment|Navigator|Engineer|RouteBlackHole'`
- [x] 翻譯／編碼完整掃描在本輪重新執行：24 份 TSV、5,049 筆、1 筆刻意保留的 `BC`
- [x] 不使用 debug hook 的正常玩家路徑本輪重新執行：私有資料唯讀掛載，`-game -shot` 成功
- [x] 可寫存檔 overlay 與 pristine original 隔離本輪重新執行：截圖廊 `18_loadgame.png` 成功，原始資料 `:ro`
- [ ] 原版／重製新截圖逐張審查（本輪只完成正常路徑人工檢查）
- [x] 封裝產物 smoke test：從 `/tmp` 啟動 Docker 內建置產物並成功產生截圖
