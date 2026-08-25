# 議會／安塔蘭／戰術畫面視覺 oracle（2026-08-11）

本筆記把「原版畫面對照」與「重製已完成的可玩畫面」分開記錄。外部截圖只能作
構圖、色彩與資訊層級參考；沒有同一存檔、同一幀、同一縮放條件時，不把它升格成
逐像素 oracle。逐像素基準優先使用原版 LBX 解碼出的像素與本專案 640×480 內部畫布。

## 輸入與工具定位

| 輸入 | SHA-256 | 結構證據 |
|---|---|---|
| `COUNCIL.LBX` | `78b3a8daafafb62106839e715a9af22c4f6c4c32f6c975748c8f23cc5437ce77` | 資產 1 = 640×480、10 幀；資產 0 作調色盤提供圖 |
| `ANTAROOM.LBX` | `81898ffba3c81fde252b80c5c557dc4049acf67becfef790b56ae2239ebde1f3` | 資產 1 = 640×480、55 幀；資產 0 作調色盤提供圖 |
| `CMBTSHP.LBX` | `ae731ad1d7e09f6dcaa573d22291a7713af8897047b1bbd73e7d06e383f8bb1e` | 360 資產、8 色塊 × 45；每色塊 44 sprite + 1 palette-holder |

結構探針在隔離 Docker 容器內執行：`/private/bin/lbxdump`，輸入唯讀掛載自
`/home/anr2/moo2-private-build/gamedata/mastori2`，輸出只寫入暫存目錄；工具的
`--help` 沒有提供獨立版本字串，因此不把未記錄的版本號填成事實。位址型反組譯
證據仍以 `docs/re/oracle-static-ida-20260811.md` 的 IDA Pro 9.4、IDA 線性位址
契約為準。

## 已接入的幀序列

| 畫面 | 原版資產 | 重製接線 | 證據等級／留白 |
|---|---|---|---|
| 銀河議會 | `COUNCIL.LBX#1`，10 幀 | `loadOverlayAnimationFrames` 依 `AccumulatedUpToRGBA` 逐幀累積，約每 3 個 redraw 播一幀，播完停最後一幀 | 幀數與尺寸**已證實**；原版實際幀停留時間仍未由 runtime oracle 量測 |
| 安塔蘭王座廳 | `ANTAROOM.LBX#1`，55 幀 | `loadAntaranRoomFrames` 同樣逐幀累積，進入畫面時從第 0 幀播放至第 54 幀 | 幀數、尺寸與 `sub_14C83` 載入鏈**已證實**；重製按 redraw 的時間比例是近似 |
| 戰術戰鬥 | `COMBAT.LBX` + `CMBTSHP.LBX` | 戰場底圖、raw picture 色塊映射、艦體 frame adapter 與 CMBTSFX 特效均有 fallback | `sub_30062 @ 0x30062` 已證實 `45*color+picture`；20 frame 的原版 timer／中間幀仍未知 |
| 地面戰 | `COLGCBT.LBX` | 原版地面戰底圖與格子／戰力摘要接入；流程走 `ResolveGroundCombatOrig` 與原版 LCG adapter | 傷亡／亂數、AI 裝甲營與人口保留已接入；原版全局 save seed 尚未接軌，事件漂移／爆炸連鎖僅能接已證實公式 |

## 對照清單

重製截圖以 2× 輸出保存，因此 `docs/screenshots/*.png` 是 1280×960，內部畫布仍是
640×480。抽樣對照時固定以下四張，避免只看主選單而漏掉戰鬥與終局畫面：

- [`08_antaranroom.png`](../screenshots/08_antaranroom.png)：安塔蘭王座廳與 55 幀終局入口。
- [`15_diplomacy.png`](../screenshots/15_diplomacy.png)：外交房間、使節與操作文字。
- [`16_tactical.png`](../screenshots/16_tactical.png)：戰場、艦體大小／色塊、血條與特效。
- [`17_groundcombat.png`](../screenshots/17_groundcombat.png)：地面戰畫面與低優先 oracle 入口。

檢查順序是：先確認 640×480 邊界與資產縮放，再確認文字沒有穿出框線，最後才比
單一像素的色彩差異。原版資產的英文烘字被重製中文覆蓋時，文字區不作「像素相同」
要求；背景、框線、透明邊界與 sprite 位置才是可比區域。

## 網路畫面參考

- [MobyGames 的 Master of Orion II 截圖索引](https://www.mobygames.com/game/182/master-of-orion-ii-battle-at-antares/screenshots/)：可按類別看到 Leaders、Diplomacy、Battle、Race 與結局畫面。
- [Old-Games 戰術畫面截圖](https://www.old-games.ru/games/pc/master_of_orion_2_battle_at_antares/screenshots/78_5abf72a923b03.png)：用來比對戰術戰場的資訊密度、艦體佈局與控制列層級。
- [Rengels 的安塔蘭畫面](https://rengels.de/computer/orion2/image/antares_1.png) 與 [Orion Guardian 畫面](https://rengels.de/computer/orion2/image/orion2_guardian.png)：用來比對終局色調、中央主體比例與威脅呈現，不作同一幀像素基準。
- [StrategyWiki 外交／間諜頁](https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Diplomacy_and_intelligence)：補充外交面板中提案、餽贈與間諜操作的資訊層級。

## 未宣稱完成的部分

1. 原版議會／安塔蘭 10／55 幀的**時間曲線**尚未以 DOSBox／實機錄影逐幀量測；目前是資產幀序正確、時間比例可玩的 remake 實作。
2. `CMBTSHP` 的**標準 raw picture 映射已由 `sub_30062 @ 0x30062` 證實**；沒有 raw picture 的抽象敵艦仍使用視覺 fallback。`sub_3F5F1`／`sub_3F628` 證實 16 向 heading 與最短 ±1 轉向，但 20 frame 的原版 timer／停留曲線仍待 runtime。
3. 地面戰的兩次 `Random(100)` 比較、AI 駐軍裝甲／陸戰隊／民兵、戰後駐軍回寫與被俘人口保留已接入；`sub_22D57 @ 0x22D57` 的總人口極值排除／差平方權重、`sub_586D4 @ 0x586D4` 的反覆減半抽樣也已建立純 oracle。`sub_3868F` 的爆炸 roll／20 點連鎖／引擎潛勢屬戰鬥／殖民地爆炸鏈，已證實不被隨機事件 8 呼叫；其 `sub_39985` 完整旗標／行星消費與原版全局 save seed尚未映射，不把它們宣稱成完整 runtime parity。
