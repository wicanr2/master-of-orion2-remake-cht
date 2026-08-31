# 戰術戰鬥畫面骨架稽核（2026-08-31）

## 問題與證據

目標是先證明原版與 remake 使用同一個 640×480 固定骨架，再談同存檔、同艦隊與同動畫幀。
本輪只處理 layout-only；不把不同戰局的公開截圖宣稱為逐像素 oracle。

- 原版公開截圖：Old-Games.ru，640×480，SHA-256
  `77383b8a1ec88dfa964916863dd3a269a5139a34406e1055c96b39d6409f3354`。
  URL：`https://www.old-games.ru/games/pc/master_of_orion_2_battle_at_antares/screenshots/78_5abf72a923b03.png`
- remake 基準：`docs/screenshots/16_tactical.png`，1280×960；邏輯畫布為 640×480。
- 原版資產：`COMBAT.LBX#0` 是 640×129 控制甲板；`STARBG.LBX#0` 與
  `CMBTSHP.LBX` 共用已驗證戰鬥調色盤。資產雜湊與解碼契約見
  `docs/re/visual-oracle-20260811.md`。
- 證據等級：控制甲板尺寸與貼底位置為已證實；公開截圖中的區域用途與「不存在自製覆蓋層」
  為強推論；不同戰局的艦艇座標、武器內容與動畫相位未知。

## 640×480 區域契約

| 區域 | 邏輯矩形 | 證據／用途 |
|---|---:|---|
| 太空戰場 | `(0,0,640,351)` | 控制甲板上方完整自由空間；沒有標題列或外框 |
| 控制甲板 | `(0,351,640,129)` | `COMBAT.LBX#0` 已證實尺寸並貼底 |
| 選艦資訊 | 約 `(8,358,96,114)` | 艦名、縮圖與耐久條 |
| 武器／特殊 | 約 `(108,358,158,114)` | WEAPONS／SPECIALS 與逐項列表 |
| 七個控制鈕 | `x=274..391, y=365..462` | 英文 LBX 浮雕邊界與既有 hit test 共用 |
| Systems | 約 `(400,358,96,114)` | 選中艦 drive／shield／computer／structure／armor 等摘要 |
| 戰術縮圖 | 約 `(500,358,136,114)` | 艦艇／星體相對位置 |

## 推翻的 remake 斷言

2026-08-30 前的畫廊在戰場上額外繪製大標題、單行訊息帶、8×6 可見格線、逐艦艦名矩形卡與
常駐 HP 條。它只證明自製格子規則可操作，不能稱為原版戰鬥畫面；這些層會大幅改變資訊層級、
遮住 CMBTSHP 艦艇並壓縮太空戰場。

## Remake 對映與停止線

- 8×6 格位暫時保留為不可見的規則／點擊 adapter；本輪不冒稱原版自由座標移動已完成。
- renderer 只在太空戰場畫 STARBG、CMBTSHP、戰機、飛彈／特效與低彩度目標環。
- 艦名、武器模式與 typed 艦艇摘要回到原版控制甲板對應區域；戰機顯式出擊按鈕移入
  SPECIALS 區，仍標為可玩 adapter。
- 真正逐像素完成仍需由 DOSBox 使用已知存檔擷取未縮放 640×480 原版幀，再建立相同艦隊、
  位置、heading、動畫 tick 與色盤的 remake fixture。公開截圖只能關閉骨架，不關閉同狀態 parity。

## DOSBox-X 實機擷取勘誤（2026-08-31）

本專案現可用 `scripts/capture-original-oracle.sh` 在既有
`civ1-dosboxx-input:20260830` 映像中重播原版。原始目錄唯讀，DOSBox-X 視窗的 17px 工具選單
不屬於遊戲 framebuffer；固定裁切 `(0,17,640,480)` 後得到未縮放的 640×480 PNG。腳本同時輸出
PNG 雜湊、DOSBox-X 版本、裁切契約及輸入檔雜湊，避免日後把桌面截圖或選單列誤當原版畫布。

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `SAVE10.GAM` SHA-256：
  `ece2eb06d782078dd0a6f746020a05691355303ceb02bbfbbe2233e987272be1`
- 工具：DOSBox-X 2026.06.02 SDL1；位址空間不適用，此處是玩家可見動態 oracle，不是反組譯證據。

`SAVE10.GAM` 經主選單 `CONTINUE` 的正常玩家路徑載入後，畫面是星曆 3500.0 的新局，仍停在
「Enter Home Star Name」對話框；它不是戰鬥前存檔。這項觀察推翻「可直接用 SAVE10 擷取同狀態
戰鬥幀」的工作假設，但不推翻它作為開局資料 oracle 的既有用途。戰術逐像素 gate 因此仍為
**未知／阻塞於戰鬥 fixture**：下一步必須在原版的可寫副本中，經正常玩家操作建立戰鬥前存檔，
再以同一腳本擷取；不得以這份新局畫面或公開不同戰局截圖升格 parity。

## 同存檔戰術 oracle（2026-08-31）

已用 `tools/gamfixture` 從上述 `SAVE10.GAM` 產生不入版控的受控副本。工具依 129-byte Ship wire
格式只把既有 `ship 16`（Amoeba，owner 10）的 `Star/X/Y` 從 `23/43/560` 改成玩家母星
`29/60/35`，並依 `Load_Game_ @ 0x10E2F` 把 `.GAM` offset 262 從 strategic `1` 改成 tactical
`0`。全檔實際只有 offsets `262、141261、141263、141265、141266` 五個 bytes 改變；manifest
保存 before／after 與雜湊。`MOX.SET` 沒有修改。

- fixture `.GAM` SHA-256：
  `dbf3f05d60562c33f23ca90737de48d0794fd76392a0c8c255485ef0c80bbb35`
- 原版戰術首幀 SHA-256：
  `b7ce37465617bc6feb19651d23863e44b800f9f35f3f5f9c54fe3c7f3272ec2e`
- 動態路徑：主選單 `CONTINUE` → 接受母星名 → `TURN` → 原版顯示 Amoeba 攻擊 Trilar III →
  進入 `sub_47939` 格子戰術畫面。這證明 fixture 由原版正常 battle-search consumer 消費，
  不是 remake importer 或 direct-entry 自行解釋。

首幀量測與資產交叉驗證：

- 戰場／控制甲板仍是 `(0,0,640,351)`／`(0,351,640,129)`。
- Trilar III 行星矩形 `(109,168,108,108)`；真存檔 planet 66 為 `Climate=5、Size=2`，精確對應
  `CMBTPLNT.LBX#32` 與同 block 調色盤 holder `#35`。
- 兩艘 Trilarian Frigate 中心約為 `(412,133)`、`(412,174)`；Star Base 中心約 `(340,201)`。
  這直接證明原版 renderer 使用自由座標，而不是 remake 舊 8×6 格心。
- Star Base 的藍色五幀我方選艦環來自 `COMBAT.LBX#33`（50×50）；另兩級環為 `#32`（28×28）
  與 `#34`（77×77）。早期只依 Star Base 的抽象 size class 選到 `#34`，但同座標放大裁切顯示
  原版外框寬度與 `#33` 一致；基地本體尺度本來已接近原版，不應連本體一起縮小。
  2026-08-31 早期判讀曾把中心艦誤認成 Amoeba；後續將中心與控制甲板 portrait 放大裁切，兩者
  是同一個 Star Base sprite，足以推翻舊「敵方目標環」解釋。Amoeba 的首幀座標仍未知。

remake 已接入上述 CMBTPLNT 映射、三組原版我方選艦環與怪物戰已證實的守方自由座標。規則層仍保留 8×6
adapter；第一次格子移動後 renderer 會明確回退格心，未冒稱自由座標移動規則已完成。Amoeba 座標、
下方面板逐字／逐值、縮圖內容與同艦艇設計仍待後續同狀態 fixture 對齊。

remake 同狀態診斷由 `-game -tactical-oracle-save <fixture.GAM> -tactical-oracle-monster amoeba
-shot <out.png> -uiscale 1` 觸發；它走正式 `.GAM` importer、`StartMonsterCombat` 與 renderer，但
直接進戰術畫面，所以只算 renderer 診斷，不取代上面的原版正常玩家路徑。2026-08-31 第二張輸出
SHA-256 為 `284b8520b34431bd1dfc4a0ea7e6c8518d16ade4a7b629767241c7693b5913a5`（僅存 `/tmp`，不入版控）。
該圖證實行星、兩艘 Frigate、Star Base 中心、選艦環與控制甲板 portrait 已對位；當時剩餘可見
差異是 Frigate／Star Base 調色偏暗、Amoeba sprite／首幀位置，以及控制甲板字型／逐值排列。

## 第二輪資產與控制甲板勘誤（2026-08-31）

`CMBTSHP` palette-holder 只包含玩家色段；remake 先前直接使用它，透明的其餘索引因而把艦體
共用灰階抹成黑色。現在固定以 `COMBAT.LBX#11` 為完整基底，再疊當前色塊 holder 的 Alpha 非零
項。相同 fixture 的 Frigate、Star Base 與 portrait 因此恢復原版可見灰階。這是資產結構與
同狀態影像共同支持的**強推論**，不是 IDA 已證實的逐指令色盤呼叫順序。

以 `COMBAT#11` 基底加 `MONSTER#13` 局部色盤重建 `MONSTER.LBX` 0..12 接觸表後，可直接辨識
`7=Guardian、8=Eel、9=Crystal、10=Amoeba、11=Hydra、12=Dragon`，0..6 則是安塔蘭艦艇
尺寸／類型序列，不是殖民地軌道基地。remake 已依此**強推論**讓太空怪物走 `MONSTER.LBX`，不再誤用一般艦艇圖；Amoeba 現為
正確綠色 sprite。該接觸表只保存於 `/tmp`，未提交原版資產。

控制甲板亦改用原版可見的短式武器列與分欄 Systems 值，四列武器及七列系統資料均留在原框內，
不再讓通用戰鬥訊息覆蓋 Star Base Systems。最終 remake 診斷圖只保存於 `/tmp`，SHA-256 為
`10c18ea7062ea3e00fe06f51570a6ec7f9139bcd1c282bd8448965ffaff2dfe5`。該階段仍未關閉的玩家可見
差異由後續裁切與部署指令收斂為 raw deployment 到主畫面像素的轉換、畫面外進場移動、Star Base 外環級別、右下縮圖
與自由座標移動規則；沒有原版證據前不得以把怪物任意移出畫面來偽造首幀相似度。

## 戰術縮圖與視窗框（2026-08-31）

原版控制甲板右下裁切放大四倍後，可量測行星中心約 `(519,418)`、Star Base 約 `(532,417)`、
兩艘 Frigate 約 `(538,413)/(538,415)`，以及橘色 Amoeba 約 `(571,416)`。單靠螢幕中心曾可擬合
出近似投影，但該擬合把 sprite center offset 混入座標，不能作為原版公式。

既有 IDA Pro 9.4 非破壞性匯出現已補上語意審查：`Deploy_Ships_` 原始函式
`sub_49043 @ 0x49043..0x494A8`（bytes SHA-256
`f3fde540a1c20948dd7cfc2f9c3a57ff000eb100afeaa8d06cfcb806cfddaf95`）直接寫 313-byte combat
record `+0x21/+0x22`。Amoeba 為攻方、玩家殖民地為守方的同狀態首幀得到：

- 行星 record `(10,34)`；
- 守方一般艦從 `(25,34)` 開始，下一艘依 size 增量落在 y=36；
- `record+0x31==0` 的軌道基地由行星偏移成 `(21,35)`；
- 攻方第一艦 Amoeba 為 `(55,34)`，heading 8，確實位於目前主視窗右側。

縮圖量測與 raw 點共同收斂為約 `miniX=507+rawX*6/5`、`miniY=372+rawY*6/5`；行星、基地與
怪物圖示因自身尺寸另有數個像素的視覺中心偏移。raw 部署座標與常數寫入為**已證實**；6/5
縮圖比例及圖示偏移是**強推論**，因尚未取得縮圖繪製函式的獨立指令窗口。remake 因此改為
分開保存 raw deployment 與 640×351 螢幕中心，不再以螢幕座標冒充玩法座標。

原始指令、bytes、函式邊界與輸入雜湊可回查
[`fighter-garrison-tactical-ida-20260828.json`](evidence/fighter-garrison-tactical-ida-20260828.json)
的 `raw_deploy_ships.function`；本輪沒有可用正式 `.i64`，故只重審既有 IDA 9.4 非破壞性匯出，
沒有製造新的 IDA 執行聲明。改用 raw deployment 後的 640×480 remake 診斷圖 SHA-256 為
`1778378c057cb51518ff13a131a512236cf7b7b049152e0b501b6473e55ba076`，只保存於 `/tmp`。

套用上述兩項後的 640×480 remake 診斷圖 SHA-256 為
`0d9492909c15e9d32001f8873ad74c84a896d03aa15e8cda85ca15572f20b7de`。本輪對拍關閉的是縮圖
可見內容與 Star Base 外環級別；較早診斷圖所列的 Star Base「精確縮放」疑慮已由裁切分離為
本體尺度接近、外環選錯，故不再作為獨立缺口。仍未關閉的是 raw deployment 到主畫面像素的
camera／sprite center 轉換、畫面外進場移動、縮圖定點公式的 IDA 證據，以及自由座標玩法模型。

## 戰術鏡頭與 sprite 繪製基準（IDA Pro 9.4，2026-08-31）

本輪以正式 `Orion2.exe.i64` 的唯讀副本重新匯出，輸入 `Orion2.exe` SHA-256 為
`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`，資料庫 SHA-256 為
`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；工具是 IDA Pro 9.4，
位址均為 DOS/4GW LE image 的 IDA linear address。可重生匯出入口為
`tools/ida/audit_tactical_deployment_renderer.py`，不改名、不改型別，也不寫回正式資料庫。

下列為**已證實**：

- `sub_47939 @ 0x47939` 在 `0x479B2/0x479BB` 將 `word_1998F0/F2` 設為 `(0,0)`；這兩個
  word 是戰術鏡頭 raw 原點。舊候選解釋「`sub_465D0(1)` 初始化鏡頭」已被真正寫入端否定；
  `sub_465D0` 是重繪／狀態 consumer，不能再當成鏡頭 setter。
- `sub_2F4EE @ 0x2FB25..0x2FB4B` 從 313-byte combat record `+0x21/+0x22` 讀取 raw X/Y，
  分別減去 `word_1998F0/F2` 並乘 `0x14`，再呼叫 `sub_30062`。因此傳入 renderer 的基準為
  `baseX=(rawX-cameraX)*20`、`baseY=(rawY-cameraY)*20`。
- `sub_34454 @ 0x34454` 依 combat record `+0x23` heading 與 `+0x25` 尺寸／圖形欄位回傳
  每軸 anchor。已見的基礎值為 `-16`、`-6`、`+4`，再依四組 heading 區段調整 `0..3` 像素；
  `sub_30062 @ 0x30121..0x30151` 把 anchor 加到上述 20px 基準後交給底層繪圖。
- `sub_49A41 @ 0x49A41` 是邊緣捲動 consumer，X/Y 鏡頭分別夾在 `0..49` 與 `0..50`；
  `sub_49D09 @ 0x49D09` 是指定位置置中 consumer，先算 `requestedX-16`、`requestedY-9`，
  再套同一邊界。這些是 raw 單位，不是像素。
- 開場鏡頭為 `(0,0)` 時，Amoeba raw `(55,34)` 的 X 基準為 `1100`，即使套用目前追回的
  anchor 仍在 640px 戰場右側；原版首幀不顯示 Amoeba 是資料與 renderer 共同導出的結果，
  不再只是截圖猜測。

仍為**未知**的是 LBX frame 解碼後的內部原點／透明裁切如何把上述「繪製基準」轉成肉眼所見的
sprite 中心，以及敵艦由畫面外進場的逐 tick 移動與鏡頭追蹤時序。先前量測的 Frigate
`(412,133)/(412,174)` 與 Star Base `(340,201)` 是視覺中心，不可拿來反推 camera 或覆蓋
已證實的 raw 基準公式；remake 在這一層閉合前保留既有可玩自由座標 adapter。

### 底層 Draw 座標與固定 frame 畫布

後續同一份 IDA 資料庫又閉合下列**已證實**關係：

- `Draw_` 原始函式 `sub_12A478 @ 0x12A478..0x12A914` 把 register `eax/edx` 保存為
  x/y，直接以 `x+image[0]-1`、`y+image[2]-1` 算裁切右下角；`sub_12ACA4 @
  0x12ACA4..0x12AE00` 也以 `(y+row)*640+x` 寫 framebuffer。傳入座標因此是完整
  frame 畫布左上角，不是中心，也沒有底層隱藏 hotspot。
- runtime image `+4` 是目前幀、`+6` 是幀數、`+8` 是 frame delay；`Draw_ @
  0x12A860..0x12A909` 在繪製後遞增並回繞目前幀。LBX decoder 舊註解把檔內
  `data[4:6]` 寫成未知，現已訂正其 runtime 對應，但 remake 仍由畫面物件持有播放狀態。
- 真實 `CMBTSHP.LBX` 的 0..43 一般艦圖均為 59×60、20 幀；`MONSTER.LBX#10`
  Amoeba 為 59×59、20 幀。這是 archive 解碼所得的已證實資產形狀，不是版本參數猜測。
- 對 heading 8、最小 size group，`sub_34454` 的 X anchor 是 `-18`。若 cameraX=5，
  Frigate rawX=25 的 59px frame 視覺中心為 `(25-5)*20-18+59/2=411.5`，與原版量測
  約 412 收斂。這只交叉驗證 X 鏈；不能反過來把截圖量測升格成 cameraY 證據。

首幀 cameraY 仍為**未知**：`sub_42F7F @ 0x44440..0x4446A` 可依戰場邊界計算位置，
而 `0x444C8..0x444EE` 在特定 gate 下又會以選中 combat record 呼叫 `sub_49D09`。
目前尚未閉合該 gate 的初始值與兩條呼叫的先後結果，因此不能僅憑 Frigate Y 中心選定
cameraY。敵艦進場也仍需追 record `+0x21/+0x22` 的第一個 runtime writer。

### 首個可玩鏡頭與 runtime 座標 writer

後續交叉參照推翻了本節早期的兩個導覽名稱：`word_199888` 是 `sub_11438B @ 0x46C89`
建立的 UI 熱區 ID，`word_199892/94` 是右下縮圖熱區 `sub_114DCA @ 0x467DE` 回填的
點擊座標；三者都不是 camera gate 或戰場邊界。`sub_42F7F @ 0x44435..0x444EE` 比較的是
本輪事件碼：點縮圖時把座標乘 `2/3` 後置中，點選艦事件則用該艦 raw X/Y 置中。

下列控制流為**已證實**：

- `sub_47939` 在部署後呼叫 `sub_4A12A` 載入戰鬥 sprite；該函式不寫 camera。進入首個
  戰術迴圈後，`sub_4A5CE @ 0x4A6B5..0x4A6C5` 選出該方可行動 combat ship，直接讀
  record `+0x21/+0x22` 並呼叫 `sub_49D09`。因此初始化 `(0,0)` 只是載入暫態，第一個
  可玩鏡頭會以活動艦置中；具體艦號由主動方與可行動篩選決定，不能固定寫成 Star Base。
- 全程式掃描 391 個文字上帶 `+21h/+22h` 的相對運算元後，313-byte combat record 的
  成對 runtime writer 收斂為兩條。`sub_3EE0F @ 0x3F543..0x3F579` 將連續像素路徑終點
  各除以 20，寫回 byte `+0x21/+0x22`，並把該艦設為目前選中艦；這是玩家／一般移動落點。
- `sub_ABFF3 @ 0xAC025..0xAC12D` 以目前 raw 點加候選偏移，查 81×68 佔位圖，經
  `sub_3E598` 等 consumer 後把選定候選寫回同一雙軸。它沒有直接 code xref，屬於間接派送
  的戰術 AI 移動端是**強推論**；雙軸寫入與佔位圖資料流本身為已證實。

目前沒有證據支持獨立的「Amoeba 進場動畫」狀態。畫面外敵艦透過正常戰術 AI 移動改寫 raw
座標，再由 camera／renderer 顯示，是現有證據支持的模型；remake 應完成自由座標移動與 AI
consumer，不應另造只服務同狀態截圖的 scripted entrance。

Remake 映射已開始：`cmd/moo2` 新增暫態 `tacticalCamera`，直接實作已證實的活動艦置中、
兩軸夾制與 20px 基準；同狀態怪物戰建構完成 raw deployment 後會以目前選中艦初始化，玩家
改選我方艦也會更新。這一輪刻意未讓 renderer／hit test 消費 camera，因兩者若不同步切換，
會製造「畫面外仍可用舊格位點擊」的新缺陷。此接線為 **PARTIAL**，不升格為畫面 parity。

第二個 remake 小切片先修正現有 adapter 的玩家可見不對稱：`ScreenPositionKnown` 艦艇原本
依自由中心繪製，點擊卻只查 `Col/Row`。現在 hit test 先使用與 renderer 相同中心及完整
60×60 frame 半開矩形，再回退 8×6 格位。這項內部測試只證明畫面與點擊共用座標；raw camera
renderer 與捲動輸入仍未完成。
