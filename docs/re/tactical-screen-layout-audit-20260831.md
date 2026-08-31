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
- Star Base 的藍色五幀我方選艦環來自 `COMBAT.LBX#34`（77×77）；較小兩級環為 `#32/#33`。
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
`7=Guardian、8=Eel、9=Crystal、10=Amoeba、11=Hydra、12=Dragon`，0..6 則是軌道基地尺寸／類型
序列。remake 已依此**強推論**讓太空怪物走 `MONSTER.LBX`，不再誤用一般艦艇圖；Amoeba 現為
正確綠色 sprite。該接觸表只保存於 `/tmp`，未提交原版資產。

控制甲板亦改用原版可見的短式武器列與分欄 Systems 值，四列武器及七列系統資料均留在原框內，
不再讓通用戰鬥訊息覆蓋 Star Base Systems。最終 remake 診斷圖只保存於 `/tmp`，SHA-256 為
`10c18ea7062ea3e00fe06f51570a6ec7f9139bcd1c282bd8448965ffaff2dfe5`。仍未關閉的玩家可見差異是
Amoeba 首幀座標、Star Base 精確縮放、戰術
縮圖內容與自由座標移動規則；沒有原版證據前不得以把怪物任意移出畫面來偽造首幀相似度。
