# 戰術戰鬥畫面骨架規格

狀態：READY（layout-only）

證據：[`tactical-screen-layout-audit-20260831.md`](../re/tactical-screen-layout-audit-20260831.md)

## 契約

1. 邏輯畫布固定 640×480；`COMBAT.LBX#0` 固定貼在 `(0,351)`，形成 640×129 控制甲板。
2. `(0,0,640,351)` 是無標題、無對話框、無可見格線的太空戰場。隱形格位可暫供規則與點擊，
   但 renderer 不得畫格線、逐艦卡、艦名框或常駐 HP 條。
3. 一般艦艇使用 `CMBTSHP` 原生比例與已證實 `45*color+rawPicture` 映射；解碼必須先套
   `COMBAT.LBX#11` 完整基底，再疊該色塊的 `CMBTSHP` palette-holder，不能讓局部色盤的透明項
   抹掉共用灰階。六種太空怪物改讀 `MONSTER.LBX#7..12`，同樣以 `COMBAT#11` 為基底並疊
   `MONSTER#13`。選中艦只可使用不遮圖的低彩度環，未知資產採小型 fail-safe token。
4. 選中艦名／縮圖、武器／特殊、控制鈕、Systems 與縮圖各自留在控制甲板既有區域；繪製與
   hit test 共用座標。繁中文案必須在區域內量測、截斷，不以新面板覆蓋原版框線。
5. 同狀態 Amoeba／Trilar III oracle 已固定守方首幀自由座標、行星與五幀我方選艦環；怪物戰首幀必須使用
   `CMBTPLNT` 真資產與自由座標。自由座標移動、同設計艦隊、縮圖、傷害數字與其餘動畫時序仍不得
   由單幀猜值。
6. 原版 framebuffer 擷取固定排除 DOSBox-X 的 17px 工具選單；保存 640×480 PNG、輸入檔雜湊、
   工具版本與裁切矩形。既有 `SAVE10.GAM` 是新局而非戰鬥前狀態，不能充當戰術 fixture。

## 驗收

- 幾何測試固定戰場／控制甲板邊界，所有文字與熱區不跨區。
- Docker＋Xvfb 產生 640×480 與 1280×960 戰術截圖；人工確認沒有標題、訊息帶、格線與艦名卡。
- 保存原版參考雜湊、remake PNG、layout-only 標籤及已知差異；不得宣稱逐像素相同。
- 同狀態驗收只接受正常玩家路徑建立的戰鬥前存檔；direct-entry 與不同戰局截圖只可作診斷。
- Trilar III fixture 驗收固定 `CMBTPLNT#32/#35`、行星 `(109,168,108,108)`、Frigate 中心
  `(412,133)/(412,174)`、Star Base 中心 `(340,201)`，我方選艦環使用 `COMBAT#34` 五幀資產；
  Amoeba 首幀座標未證實，不得再與基地中心混為一談。
- 控制甲板的武器與 Systems 文字必須量測在各自框內；Star Base fixture 的四列武器使用原版短式
  數量／改造／射界／彈藥格式，不得以除錯明細取代。此項只驗證版面與可見值，不代表縮圖、
  Star Base 精確縮放或 Amoeba 首幀座標已對齊。
