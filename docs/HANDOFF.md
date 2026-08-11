# 交接文件

## 2026-08-11 人口回寫勘誤

- `TestPopulationGrowthWriteback` 已重新驗證並通過。先前「新人口必為工人」不是規則：
  `assignNewColonist` 會先嘗試工人，若造成食物赤字則改派農夫。固定 demo 的 30 回合結果是
  人口 8→12、農／工／科 4/2/2→8/2/2；測試現在檢查人口上限與人口／職務總數同步，避免把
  缺糧保護誤報為人口沒有回寫。下方較早的失敗敘述是歷史記錄，不可再當成目前 blocker。

## 2026-08-11 發行候選與收尾抽樣

- 依使用者要求，本輪採抽樣驗證，不重跑 `go test ./...` 全套；測試員 A 的 remake
  抽樣以 Docker + Xvfb 產生 35 張 `-gamegallery` PNG，退出碼 0、最小檔案
  `39,409` bytes、全部為 `1000:1000`，涵蓋主選單、新局、種族／命名、星系／回合、
  艦隊／殖民地、外交、戰術、存檔／載入、艦艇設計與輸入框。
- 測試員 B 的原版 DOSBox／正版資料對照已完成，報告見 [`docs/PLAYTEST-2026-08-10.md`](PLAYTEST-2026-08-10.md)：
  動態啟動被缺少 `VESA.COM` 阻塞，成功部分限於 `MAINMENU`／`NEWGAME`／`STARBG`／`COMBAT`／
  `CMBTSHP` 等正版 LBX 靜態解碼；原版無法取得的畫面或規則只記為阻塞，不以 remake 截圖推論。
- 兩位遊戲測試代理的保守可玩度報告見 [`docs/GAME-TEST-REPORT-2026-08-11.md`](GAME-TEST-REPORT-2026-08-11.md)：
  remake 暫定 70／100；客製種族完整開局、RACES 間諜、LEADERS 管理、魚雷 NR 距離效果、20 回合
  體感與 `.GAM` fixture 仍待新鮮玩家路徑證據。測試員 B 收束前沒有重跑新命令，因此不把舊產物
  冒充目前工作樹的多人／包裝驗證。
- 三平台完整版本機測試包已備妥：`MasterOfOrion2-cht-full-x86_64.AppImage`、
  `MasterOfOrion2-cht-full-windows-amd64.zip` 與 `MasterOfOrion2-cht-full-macos-universal.tar.gz`
  均由本輪 Docker 依最新程式碼建立；macOS universal 的 `moo2`／`moo2sim` 由 osxcross
  `lipo` 驗證含 `arm64`／`x86_64` 兩 slice。雜湊見 `dist/SHA256SUMS`；完整版包含使用者私有
  資料子集、原版音訊與 CJK 字型，不可視為可公開散布包。
- 公開 Release candidate 三平台包已另行建立，不含原版資料／音訊／字型；檔名、大小與雜湊見
  [`RELEASE_NOTES-v0.1.0.md`](../RELEASE_NOTES-v0.1.0.md)、`docs/tech/packaging.md` §5.4 與
  `dist/PUBLIC-SHA256SUMS`。公開包的 Windows／Linux 可執行檔由目前工作樹重建，macOS 使用已由
  osxcross `lipo` 驗證的 universal binaries 組成，未宣稱真機執行。
- 推廣影片：`dist/promo/master-of-orion-2-remake-trailer.mp4`，含 12 張 remake 畫面、
  標題／功能字幕／版面輪換／CTA，H.264／AAC、1280×720、72 秒；背景音樂取自使用者正版
  資料的 `STREAM.LBX`，權利限制見 `dist/promo/README.md`，不可把影片或原版音樂當作可自由
  再散布素材；可重跑流程為 `scripts/make_promo.sh`。

## 2026-08-11 本輪 gameplay／視覺增補

### 本輪新增（CMBTSHP／爆炸／SABOTAGE／領袖 ETA）

- `CMBTSHPFrameAtTick` 與 tactical `shipMotionStart` 已接線：移動後播放固定 4 tick/frame、
  `[0,1,2,1]` 短掃掠，16 tick 後回固定朝向幀。這是 remake 近似，不宣稱原版 timer。
- 事件 8 已呼叫 `resolveStrategicShipExplosion`，主艦仍受「至少留一艘」護欄；oracle 的
  `Random(201)+74`、20 點連鎖與 `OriginalExplosionDamageConsumer` 會對最多三艘倖存艦寫入
  尺度化 `Ship.Damage`。raw fleet／colony record 未命名部分仍保留未知。
- `spyMissionScore` 明列 SABOTAGE 的 AB／DB／有效門檻與成功率來源；IDA 同時確認
  `sub_101483 @ 0x101483` 的 raw slot helper、`sub_1014A4 @ 0x1014A4` 的 packed
  relationship byte／score table 讀取位置／門檻分支；Agent 上限、訓練／解除、AI 週期補充
  與 Spy-vs-Spy 實際扣除均已接線，remake 以 AB／DB／E 完成可重播近似。
- `RawStatus=1` 的 ETA `1→0` 且 location=1 會呼叫 `applyLeaderETACallback`，重整已指派
  殖民地領袖增量、刷新所有殖民地士氣但保留任職；IDA 已確認 `sub_E2AB1`→
  `sub_E1D59`／`sub_DF8F0`／`sub_E2710` raw callback 鏈，raw 設計／帝國欄位逐值 parity
  仍是 oracle 差異。
- 對應說明與抽樣命令見 [`docs/re/remake-consumer-closure-20260811.md`](re/remake-consumer-closure-20260811.md)。

- 外交 `FoodForCredits`／`ResearchExchange` 特殊貿易已有回合收益與存檔狀態；IDA 另追回活動 Trader 的政府／神級商人／經驗 bucket／tier 1/2 最大加成，並接入 GAM raw experience 與舊 demo fallback。AI→玩家 `SABOTAGE` 依 remake 性格政策選任務並實際讀玩家建築池；49 筆原版建築成本表、slot 9 skip、`+8` 權重與 raw slot helper 已接，結構化 AB／DB／E 與防守 Agent 消費也已接。這些是 remake 主流程，不冒充原版兩張 score table 的上游填值、特殊 raw 上游、完整 raw 分數或防守策略。
- 食物複製機已在半機械族半食物帳本上補足半食物／半 BC 轉換，半 BC 以 `PlayerState.FoodReplicatorBCHalfRemainder` 跨回合保存；這是手冊與 half-unit 資料形狀支持的強推論，碎片付款時機仍未知。
- 原版 `RawStatus=4` 的 `+0x37`／30 門檻已接入領袖清理：解除殖民地／艦艇指派並移出人才池；一般在職領袖 ETA 的 remake callback 近似已接，`sub_E2AB1` 六槽與三個下游 raw 函式已由 IDA 確認，完整 raw 任命／任期／設計欄位逐值仍保留 oracle 留白。
- `COUNCIL.LBX#1` 的 10 幀與 `ANTAROOM.LBX#1` 的 55 幀已由原版資產逐幀播放；`CMBTSHP` 已使用 `45*color+rawPicture` 精確圖片映射，移動後 timer 以 remake 固定 tick 近似接線。地面戰實機 seed、事件 record／漂移與原版爆炸 raw 下游見 [`docs/re/visual-oracle-20260811.md`](re/visual-oracle-20260811.md)。
- 地面戰靜態 oracle 抽樣：`internal/gamedata` 的 LCG／平手雙方受擊／逐類型切換／終止測試，以及 `internal/shell` 的入侵勝率、可重現性、AI 裝甲營、守方兵種回寫與 captured population 保留測試均通過；這不取代尚缺的 DOSBox 實機 global seed 與戰後人口 consumer 校準。
- 本輪 Docker 抽樣：`internal/gamedata`／`internal/engine` 的食物複製機測試通過；`internal/shell` 的 AI SABOTAGE 與兩個領袖任期測試通過。`TestPopulationGrowthWriteback` 仍在非本輪人口寫回路徑失敗（新人口未分配成工人），因此不把本輪抽樣寫成全包測試綠燈。
- 本輪已在 Docker 重建 Linux 完整 AppImage（97,434,104 bytes）、Windows amd64 完整 ZIP（100,661,428 bytes）與 macOS universal tar.gz（108,752,832 bytes）；Linux 完整包實際啟動 `-gamegallery` 產生 35/35 張 PNG。推廣影片以本輪新畫廊抽樣 12 張重錄為 4,807,495 bytes、72 秒、1280×720；`dist/SHA256SUMS` 已更新並驗證四項產物／影片。完整版含使用者私有資料子集與原版音訊，權利限制見 `dist/promo/README.md`。

## 目前狀態

- 目標：完成 `Master of Orion II` 的跨平台 Go／Ebiten 重製與繁體中文化。
- 分支：`main`。
- 最近已知 HEAD：`e48ac3edee415cebd9d44c81dd549147f4587249`。
- 工作樹：刻意保留大量未提交使用者／代理變更；不要 reset、checkout 或清理。
- 原始資料：只讀於 `/home/anr2/moo2-private-build/`，不可加入 repo。
- 容器：最近一輪一次性 Docker 容器已清理；沒有名稱符合 `moo2` 的專案容器殘留，建置 image `moo2-ebiten:latest` 保留供重現。

## 本輪已驗證

### 2026-08-11 GAM／戰機／要塞補充

- `.GAM` 已不再只有 `internal/save` 唯讀解析：`internal/shell` 新增 `ImportGAM`／`LoadGAMSession`，`LoadSession` 看到 little-endian `0xE0` 會自動分流。星系、行星、殖民地／前哨站、玩家／AI、外交旗標、67 筆領袖、艦隊、建築與建造佇列會轉入 remake 工作階段；研究完成 byte、special bit、原版任命／任期下游保留報告式未知。
- 戰機命中／傷害已完成有證據邊界的消費端：ID 31 第二組 `1..4 / 4..16 / 2..7`，`sub_3AD57 @ 0x3AD57` 的攻防修正／`roll <=95`／40 命中門檻，以及相鄰 `sub_3AC20 @ 0x3AC20` 的直接插值式，均由 `internal/gamedata/fighter_damage.go`／`internal/shell/fighter_attack.go` 分開接到逐架最弱護盾／裝甲／結構鏈。未完成的是原版 raw record 逐彈參數與 raw flag 正式命名，不是 remake 的命中／傷害消費流程。
- 星際要塞已接入 `sub_4D18E @ 0x4D18E` 追回的四個非空武器槽 seed／cap／raw：`(11,375,99,2)`、`(4,187,198,0)`、`(4,187,198,4)`、`(4,375,99,2)`；光束／魚雷／直接傷害特殊武器進入 `AssaultAntares` 齊射，炸彈／戰機艙分流。`sub_6EE8E` divisor、raw 2/4 百分比與 live tech 分支已記錄；快速戰鬥使用 full-cap policy，不把 99/198 宣稱成固定 runtime 數量。

- 多人網路最低可玩鏈已完成：主機由大廳名冊建立共同新局並廣播設定／種子／席位快照，
  客戶端套用同一快照；玩家操作入口會記錄可序列化 `PlayerCommand`，每回合以
  `turn_done` 收集、依玩家編號從共同基準重播，再以 `turn_ready` 與
  `NetworkStateHash` 驗證；未知指令、錯席位、重複封包或分歧均失敗即關閉。
  `internal/netplay` 另有 resume token、心跳／寬限、challenge-HMAC 與可選 TLS 1.3，
  `MOO2_NET_AUTH`／`MOO2_NET_TLS=1` 控制；NAT 穿透仍需外部 relay 或 UPnP。已有兩個實際程序的
  loopback TCP 與 resilience 抽樣，`internal/shell` 有快照／席位重播與 UI 指紋正規化測試，
  `cmd/moo2` 另以 Docker + Xvfb 通過目標測試。原版 IPX／數據機／序列／TEN 不恢復。
- 飛彈／魚雷改造：ECCM、EMG、MIRV、魚雷 ENV／OVR，以及 ARM/FST 的 raw flag／攔截垂直切片，已接入設計、存檔、快速結算與格子戰術；魚雷 NR 的射程衰減已取消，保留原版精確 oracle 為差異追蹤，不再列為 remake 漏接。
- 艦艇軍官管理：艦隊畫面選船後按 `LEADERS`，點艦艇軍官列可指派／改派，再點同一列解除；`HIRE` 可挑指定候選，`POOL` 回人才庫，`DISMISS` 解雇艦艇軍官；`Ship.OfficerName` 與原版 `_leaders[]` 來源序號 `Ship.OfficerID` 會保存，Weaponry／Helmsman／Navigator／Engineer 均讀逐艦欄位。舊 JSON 沒有 ID 時仍以名稱回退。IDA 已證實原版 `sub_10E2F`／`sub_1160B` 對 `dword_1930DC` 直接讀寫 `0x3B×0x43` 全局記錄；remake 的 `ImportGAM`／`LoadGAMSession` 已完成原生存檔匯入，載入畫面可探測同槽 `SAVE*.GAM` 並另存 JSON，任命／任期下游規則仍未追回。
- 火線角切片：設計畫面可循環選擇 FWD／FWD EXT／BACK EXT／BACK／360；`Ship.Arc` 會保存，+25%／+25%／+50% 會進入佔格、成本與建造判定。原版 `Relative_Bearing`／`Move_Ship`／部署函式已證實 16 向朝向，格子戰術玩家／敵方射擊與移動轉向已接；快速結算的原版 call graph 不消費射界，維持抽象戰鬥，詳見 `docs/re/weapon-arcs.md`。
- 戰術戰機：玩家中隊已可出擊、貼身攻擊、被敵艦射擊、返航補給；本輪接上手冊要求的最弱護盾分面、主要目標鎖定（目標失效才重選）、戰機接戰前先消費近迫防禦（PD），以及參戰種族 Ship Defense／Fighter Pilot 加成。IDA 現已追回敵方五級 blueprint writer 與非空武器槽；ID 31 的 Interceptor／Heavy／Bomber 傷害範圍、`sub_3AD57` 命中式與 `sub_3AC20` 直接插值式已分開接入逐架護盾／裝甲／結構下游。完整 raw record 逐彈參數仍是差異，不宣稱逐值 runtime 對齊。
- 安塔蘭終局防禦：原版已證實的 `Intruder ×3`／`Interdictor ×2`／`Harbinger ×7`／星際要塞已保存為具名艦級；三個實際防禦艦的完整非空 raw weapon ID／數量／flags 已同步，並修正 Interdictor／Harbinger 第 4 槽 `raw flags=0x0004`。Intruder 的 3 個、Harbinger 的 6 個標準 `Fighter Bays` 與星際要塞四個直接火力槽已接入終局快速戰鬥，`AssaultAntares` 會把尺寸傳入球形武器的目標級數消費端；要塞 class 6=`0x17F69C=900`、P=750、seed／raw／cap 與 `sub_6EE8E` divisor 已記錄，99/198 不是固定 runtime 數量；raw bit 正式語意與 live tech 當下數量仍未知。
- 反組譯證據衛生：新增 [`docs/re/weapon-mod-flags.md`](re/weapon-mod-flags.md)，保存 `Orion2.exe`／`.asm`／`.i64`／`orion2_all.c` 雜湊、IDA Pro 9.4 與 IDA 位址基準；把完整光束改造 caller 校正為 `sub_394F7`，`sub_39434` 明確標成射程輔助。安塔蘭戰鬥 loader 也訂正為 `sub_55738`／`sub_55B12`／`sub_55F67`，重製模型以 raw 位址、標準艦非空 weapon ID／數量／raw flags 與 ID 31 戰機艙測試鎖定；ARM/FST 的垂直切片已接，但仍未被誤列為原版全路徑等價。
- 戰機逆向增補：`sub_3C892`／`sub_3CD21`／`sub_3DFE0` 的 FST runtime 速度下游、`sub_2A46A` 排除戰機記錄的目標評估，以及 raw `sub_3AC20`／`sub_3AD57` 呼叫鏈已加入 [`docs/re/weapon-mod-flags.md`](re/weapon-mod-flags.md)；重製只接入有證據的主要目標鎖定與兩條傷害公式，沒有新增敵方抽象艦隊的虛構戰機。
- 間諜／逐對手外交：RACES 內嵌區可訓練並逐對手保存 STEAL/SABOTAGE/HIDE；HIDE 每回合跳過偷科技、套用 SpyVsSpy +20；SABOTAGE 已依原版 raw `0x1014A4` 的 70 門檻與 `0x10130A`／`0x145EA` 建築旗標清除接入玩家 → AI 路徑，候選按原版建造成本加權，remake 的結構化 AB／DB／E 與防守 Agent 消費已接。外交畫面可提議／終止和平、互不侵犯、同盟、貿易、研究、固定 5%／10% 週期納貢，以及 `贈送 10 BC` 的現金餽贈玩家切片；現金餽贈核心函式接受任意正整數，國庫轉移方向與不足邊界已測試。協議進度與納貢轉移會進 BC／研究並保存。IDA 已追回外交評分桶 `-75/-50/0` 與回應分支 `-100/-50/-25/0`、特殊路徑 `-150/-75/0`；完整 AI end-to-end 接受表、一次性科技／星系餽贈、特殊貿易表／創造力係數與 AI 防守 Agent／政體資料仍是原版 oracle 留白，詳見 [`docs/re/diplomacy-gifts.md`](re/diplomacy-gifts.md)。
- 本輪依使用者要求採抽樣，不宣稱 `go test ./...` 全套通過；抽樣 gamedata／shell 測試通過。完整 suite 仍有既有 `TestPopulationGrowthWriteback`（人口分配工人預期差異）失敗，未把它誤算成本輪地面戰／外交回歸。
- `go build -buildvcs=false -o /tmp/moo2-remake-final ./cmd/moo2`：Docker 退出碼 0，產物 UID/GID `1000:1000`。
- 歷史基線（2026-08-09）曾在 Docker + Xvfb 完成全套測試與建置：`go test -buildvcs=false ./...`、`go build -buildvcs=false -o /tmp/moo2-arm-fst-final ./cmd/moo2`，退出碼 0；本輪依使用者要求改採抽樣，不重跑全套。乾淨映像缺少鎖定模組時只開放一次受限網路容器補依賴，測試與建置仍在容器內完成。
- `git diff --check`：通過。
- 翻譯掃描：Docker 讀取 24 份 TSV 得 5,049 筆、1 筆刻意保留的 `BC`。
- 私有遊戲資料唯讀掛載的正常 `-game -shot` 成功，輸出 PNG `148,252` bytes；同一輪 `-game -gamegallery` 產生 35 張非空 PNG（最小 4,043 bytes），含 `18_loadgame.png` 存檔／讀檔、`16_tactical.png` 戰術與 `25_shipdesign.png` 艦艇設計畫面；本輪未逐張人工審查。
- 本輪在 Docker + Xvfb 以私有資料唯讀掛載重跑 `-game -gamegallery`：35 張 PNG、最小 39,409 bytes，輸出目錄與檔案均為 `1000:1000`；戰機防禦加成切片未改動畫面流程，畫廊正常收尾。
- 領袖技能收尾：26/27 項技能已有 remake 消費端；Tactics 依原版不實作，Famous 招募機率與 Diplomat 獨立接受門檻保留明確 oracle 留白。新增效果已由 `internal/shell/leader_effects_test.go` 及既有 Fighter Pilot／軍官測試抽樣護欄。
- AI-to-AI 可選強化：`EnableAIVsAI` 開啟後，AI 彼此會依 `AIRelations` 形成戰爭、停戰／互不侵犯／同盟、貿易／研究協議，並以抽象艦隊攻擊最高人口殖民地；AI 選星與議會第三方搖擺票已接入。此模型保存於 JSON，但不冒充原版逐艦 blueprint，細節見 [`docs/tech/ai-to-ai.md`](tech/ai-to-ai.md)。
- IDA Pro 靜態 oracle：使用同一份 `Orion2.exe.i64`、IDA Pro 9.4、非破壞性 IDC 探針確認音樂分派與外交音樂賦值鏈，並追回敵方五級 blueprint raw writer／槽位、`sub_3AD57` 的戰機 1..100／95／40／插值／下游式、相鄰 `sub_3AC20` 的直接插值式、要塞 raw flag／seed／容量 divisor 中間鏈、外交精確比較常數、`.GAM` 全局 `0x3B×0x43` 讀寫對稱。兩份外部符號名稱一律保留 `sub_3AC20`／`sub_3AD57` raw 位址；live raw record 輸入、要塞 raw bit 正式語意與 live tech 導出的數量仍是非阻塞差異；class table 的直接 byte stride 勘誤與證據分級見 [`docs/re/oracle-static-ida-20260811.md`](re/oracle-static-ida-20260811.md)。
- IDA Pro 靜態 oracle：除既有戰機／要塞／外交／`.GAM` 證據外，本輪追加地面戰 `0xEC4FE`、事件 `0x22D57`／`0x586D4`、爆炸 `0x3868F`／`0x39985`／`0x40C2A`／`0x494A8`、CMBTSHP `0x30062`／`0x3F5F1`／`0x3F628`，活動 Trader `0x101BA4`／`0x94951`／`0x93D4B`，以及 SABOTAGE `0x1014A4`／`0x101483`／`0x1026CF`／`0x1026F1`／`0x10278D`、領袖 callback `0x934CF`／`0xE2AB1`／`0xE1D59`／`0xDF8F0`／`0xE2710`。兩份外部符號名稱一律保留 raw 位址；CMBTSHP timer、score table 上游填值、raw event record、爆炸 strategic consumer、raw callback 設計／帝國欄位與 live raw input 仍是非阻塞差異，詳見 [`docs/re/oracle-static-ida-20260811.md`](re/oracle-static-ida-20260811.md)。
- 英文模式收尾：`englishSafeFallback` 只改顯示值、不改查表 key；英文畫廊 35/35 張通過，抽看主選單、星圖、外交、艦艇設計、輸入框無越界字串。代表圖已同步 `docs/screenshots/en/` 與 `README.md`。
- 20 回合開局探針：無事件固定開局 BC 50→264、人口 8→11、士氣 0% 全程、食物輸出 0→1、工業 6→8、研究 6 維持；只作平衡基線，不冒充玩家主觀體感。
- 本輪 Docker 清理稽核：沒有名稱符合 `moo2` 的執行中／已停止專案容器；一個兩週前、無專案標籤的 dangling image 未碰觸，以免清理其他工作。

## 證據入口

| 主題 | 入口 | 說明 |
|---|---|---|
| 唯一活工作表 | [`WORKLIST.md`](../WORKLIST.md) | 目前剩餘工作，不要依賴歷史方框 |
| 誠實現況 | [`HONEST-STATUS.md`](HONEST-STATUS.md) | 已接、近似、未知 |
| 逆向工程日誌 | [`docs/re/01-gap-report.md`](re/01-gap-report.md) | 位址、資料表與編號歷史，不是活工作表 |
| 研究帳本 | [`RESEARCH-LOG.md`](RESEARCH-LOG.md) | 本輪規則與證據分級 |
| 武器／飛彈反組譯索引 | [`docs/re/weapon-mod-flags.md`](re/weapon-mod-flags.md) | `sub_394F7`／`sub_39434`、`sub_3CD21`、`sub_3E095`／`sub_3A0B9`、ARM/FST 證據分級 |
| 安塔蘭艦隊反組譯 | [`docs/re/antaran-defense-fleet.md`](re/antaran-defense-fleet.md) | 防禦艦級、戰鬥 loader 與 compact design 位址勘誤 |
| 重製計畫 | [`REMAKE-PLAN.md`](REMAKE-PLAN.md) | 架構、垂直切片與剩餘交付 |
| 驗證矩陣 | [`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md) | 測試、原版 oracle、畫廊與玩家路徑狀態 |

## 下一個精確動作

1. 發行前改用不含原版資料／音樂的公開 release 包，並於有音訊輸出的桌面逐曲確認
   `STREAM`／`STREAMHD` 的音色與場景切換；本輪本機完整版與影片不直接作公開散布。
2. 公網部署若需要跨 NAT，配置外部 relay 或 UPnP；本專案目前只負責 TCP、身份驗證、TLS、心跳與重連。
3. 只在取得可啟動 `VESA.COM` 的原版 runtime 後，才補 `sub_3AC20`／`sub_3AD57` 相鄰函式的外部名稱解析、live tech／逐彈逐槽數值、外交特殊表與原版逐值 oracle；靜態 Fire／要塞公式已完成，不重新深挖與玩家行為無關的反組譯內部功能。

## 不要重做

> **本輪勘誤**：上方較早的交接句仍把「敵方戰機下游命中／傷害、要塞完整火力、原生
> `.GAM` importer」列成未完成；它們已由本輪補充接線。現在只剩 `sub_3AC20`／`sub_3AD57`
> 相鄰函式的外部名稱／live 逐彈輸入、要塞 raw flags 正式名稱／live 數量、外交特殊表
> 與原版任命／任期等 oracle 差異；靜態中間式已寫入證據文件。

- 不要把 `openorion2` 當成遊戲引擎；它主要是渲染／資產參考。
- 不要用測試綠或新功能數量宣稱原版還原度。
- 不要在主機執行遊戲、Go 測試、Python、DOSBox、Xvfb 或 GUI。
- 不要把 ARM/FST 的 remake 垂直切片寫成原版全路徑等價；魚雷 NR、心靈感應／幸運／全知／匿蹤艦等仍未證實項目也不可寫成完成。
- 不要清理整個 dirty worktree；所有既有變更都視為使用者資產。

## 2026-08-11 Fire／要塞深度追查交接

- `sub_3AC20 @ 0x3AC20` 與 `sub_3AD57 @ 0x3AD57` 的外部名稱索引互相衝突，交接一律保留 raw 位址。
- `sub_3AD57` 的 `RawFlags & 4` 分支不可達；其命中／傷害式、隨機 helper 與
  `sub_39985`／`sub_3A0B9` 下游已證實。`sub_3DF8D` 的部分 runtime 欄位仍不命名。
- `sub_4D18E` 的 class 6 直接讀址是 `0x17F69C=900`（0x0F byte stride），不是舊探針誤算的
  `0x17F6F6=25`；四槽 seed、raw `0/2/4` 百分比、`sub_6EE8E` divisor 與 `99/198`
  上限／拆槽已證實。`T=sub_6E70A(...)` 仍是 live tech 輸入，不能把 caps 當 live quantity。
- 完整 raw 指令、輸入雜湊、工具版本與證據等級集中在 [`oracle-static-ida-20260811.md`](re/oracle-static-ida-20260811.md)。
