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
7. `CombatShip` 必須分開保存原版 313-byte combat record `+0x21/+0x22` 的部署座標與 remake
   螢幕中心座標；不得再由其中一種反推另一種後宣稱精確。Amoeba／Trilar III 首幀依
   `Deploy_Ships_ @ 0x49043` 固定：行星 `(10,34)`、同側一般艦由 `(25,34)` 起排、軌道基地
   `(21,35)`、敵側第一艦 `(55,34)`。右下縮圖以 raw 部署座標的 6/5 比例繪製標記，圖示中心
   偏移只屬顯示；不得改變射程或移動。
8. 原版 renderer 的已證實基準公式為 `baseX=(rawX-cameraX)*20`、
   `baseY=(rawY-cameraY)*20`；開場 camera 固定 `(0,0)`，X/Y 捲動界限分別為 `0..49`／
   `0..50`，指定位置置中先減 `(16,9)`。`sub_34454` 另依 heading 與圖形尺寸加入 sprite
   anchor。這是繪製基準而非視覺中心；LBX frame 內部原點尚未閉合前，不得用截圖中心替代
   anchor，也不得直接刪除目前供玩法使用的自由座標 adapter。
9. `Draw_ @ 0x12A478` 的 x/y 是完整 LBX frame 畫布左上角；底層不會再套中心或 hotspot。
   一般 `CMBTSHP` 畫布為 59×60，Amoeba `MONSTER#10` 為 59×59。視覺中心只能由
   `base + anchor + frameSize/2` 計算，透明像素外框不得被裁掉後重新置中。首幀 cameraY
   gate 尚未閉合，故目前不得用單張截圖硬編 Y 原點。

## 驗收

- 幾何測試固定戰場／控制甲板邊界，所有文字與熱區不跨區。
- Docker＋Xvfb 產生 640×480 與 1280×960 戰術截圖；人工確認沒有標題、訊息帶、格線與艦名卡。
- 保存原版參考雜湊、remake PNG、layout-only 標籤及已知差異；不得宣稱逐像素相同。
- 同狀態驗收只接受正常玩家路徑建立的戰鬥前存檔；direct-entry 與不同戰局截圖只可作診斷。
- Trilar III fixture 驗收固定 `CMBTPLNT#32/#35`、行星 `(109,168,108,108)`、Frigate 中心
  `(412,133)/(412,174)`、Star Base 中心 `(340,201)`，Star Base 我方選艦環使用 `COMBAT#33`
  五幀資產；
  Amoeba raw deployment 已證實為 `(55,34)`；其主畫面像素中心仍不得與基地中心混為一談。
- 控制甲板的武器與 Systems 文字必須量測在各自框內；Star Base fixture 的四列武器使用原版短式
  數量／改造／射界／彈藥格式，不得以除錯明細取代。此項只驗證版面與可見值，不代表原版
  動畫相位、raw deployment 到主畫面像素的 camera／sprite center 轉換或畫面外進場移動已對齊。
- 靜態證據驗收另固定 `sub_47939` 開場 camera `(0,0)`、`sub_2F4EE/sub_30062` 的 20px
  基準公式與 `sub_34454` anchor consumer。Amoeba `(55,34)` 在首幀位於右側畫面外；在追回
  frame 內部原點與進場時序以前，此項只關閉「為何首幀不可見」，不宣稱自由座標玩法完成。
