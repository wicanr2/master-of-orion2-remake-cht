# `CMBTSHP.LBX` 戰鬥艦 sprite 映射與幀序（2026-08-11）

這份文件只把原版執行檔直接讀出的欄位與資產結構標成「已證實」；沒有 raw
picture 的 remake 舊 JSON／抽象敵艦才使用視覺 fallback。

## 已證實的資產結構與圖片索引

`CMBTSHP.LBX` 的 SHA-256 是
`ae731ad1d7e09f6dcaa573d22291a7713af8897047b1bbd73e7d06e383f8bb1e`，共有 360 個
資產，分成 8 個色塊、每塊 45 個資產。IDA Pro 9.4 對 `Orion2.exe.i64`
（SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`）的
`sub_30062 @ 0x30062` 直接讀出：

```text
asset = 45 * player_color + raw_ship_picture
```

其中 `player_color` 是玩家記錄 `+0x26` 的 0..7 色塊，`raw_ship_picture` 是艦艇
記錄 `+0xC4` 的原始欄位。標準艦的合法 CMBTSHP 圖片為色塊內 0..43；索引 44
（十六進位 `0x2C`）是原版進入 `MONSTER.LBX`／調色盤路徑的 sentinel，不能當作
標準艦 sprite。每色塊最後一個資產因此是 palette-holder：

```text
標準艦 sprite：45*k + 0 .. 45*k + 43
palette-holder：45*k + 44
```

這條公式已接到 `CMBTSHPSpriteIndex`、`.GAM` 匯入的 `Ship.CombatPicture` 與
戰術 `StartCombat`。raw picture=0 也是合法圖片，不能以 Go 零值判定「未知」。

## 顏色與 fallback

玩家的真實 raw picture 直接套用上述公式；敵方若有 `.GAM` 藍圖與顏色欄位也走同一
條公式。沒有 raw picture 的舊存檔、程序化 demo 或只有抽象戰力的敵艦，才使用
`CombatSpriteForClass`／`CombatSpriteForStrength` 的小艦到大艦代表索引。代表索引
（3、12、20、28、36、43）是**降級顯示**，不是原版艦級→picture 表，不能反向
覆寫已匯入的 raw picture。

## 幀與轉向

LBX 解碼抽樣確認每個 CMBTSHP 資產含 20 個 frame；這是資產格式事實。原版
`sub_3F5F1 @ 0x3F5F1` 將幾何方向轉成 0..15 的 raw heading；
`sub_3F628 @ 0x3F628` 讀艦艇記錄 `+0x23`，依最短方向把 heading 每次只改變
`+1` 或 `-1`，再以 16 取餘。原版 loader `sub_30062 @ 0x30062` 確實使用
`45*color + picture`，但本次靜態輸出沒有找到「20 個 frame 如何隨時間遞增」的
獨立 timer／tick 常數，也沒有足夠證據把 20 frame 命名成 20 個方向。

因此 remake 的責任邊界是：

- `CMBTSHPSpriteIndex` 是原版精確圖片映射；
- `CMBTSHPFrameForHeading` 是 16 向 heading 到 20 frame 的最近角度顯示 adapter；
- 戰術畫面在艦艇移動後以 `CMBTSHPFrameAtTick` 播放固定、可重播的短掃掠，停止後回到
  `CMBTSHPFrameForHeading` 的固定幀；不以 wall-clock 讓靜止艦自行旋轉；
- `CMBTSHPFrameHoldTicks=4`、`CMBTSHPMotionFrameCount=4` 與 `[0,1,2,1]` 是 remake
  顯示近似。IDA 對 `sub_30062`／`sub_30631`／`sub_31F25`／`sub_3F628` 的深度窗口
  沒有找到 frame counter、clock／timer 呼叫或中間幀停留常數；原版 frame timer 仍標為
  未知，只有取得可啟動的原版 runtime（目前缺 `VESA.COM`）才值得逐值校正。

2026-08-11 的 runtime 接線位於 `cmd/moo2/interactive.go`：點擊移動時記錄本次
`shipMotionStart`，繪圖只在 `CMBTSHPMotionDurationTicks` 內消費 timer；這使 CMBTSHP
動畫有玩法觸發點，同時不把近似 timer 誤報成原版已證實常數。

這樣不會把「精確 picture 映射已完成」誤報成「精確動畫時序也已完成」。

## 可回查證據

- `tools/ida/late_oracle.idc`：非破壞式輸出 `0x30062`、`0x49F99`、`0x3F5F1`、
  `0x3F628`；原始名稱、運算元、IDA 線性位址均保留。
- `internal/shell/session.go`：`CMBTSHPSpriteIndex`、`CombatSpriteForShip`、
  `CMBTSHPFrameForHeading`。
- `cmd/moo2/interactive.go`：依色塊 palette-holder 與 frame cache 載入資產。
- [`docs/re/oracle-static-ida-20260811.md`](../re/oracle-static-ida-20260811.md)：
  輸入雜湊、工具版本、原始匯出與證據等級。
