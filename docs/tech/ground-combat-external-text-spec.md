# 地面戰外部文案與版面規格

## 文案契約

- `groundcombat.go` 只保留座標、顏色、typed `GroundInvasionResult`、資產 ID 與
  `groundcombat.*` 鍵。固定中英文文案全部由 `assets/i18n/ui.json` 提供。
- 必要鍵涵蓋：星系轉場、無對局錯誤、一般／帶殖民地標題、攻守方、陸戰隊、裝甲營、存活合計、
  守軍、交戰回合、三種戰果及繼續按鈕。
- 殖民地名稱與數值是外部狀態，只能填入 JSON 格式模板；不可在 Go 拼接玩家可見分隔符或句子。

## 原版證據與 remake 差異

- 面板沿用 `COLGCBT.LBX#21`、攻守 `x=1/378`、首列 `y=50`、原版列距 11 與標題
  `(319,10)` 的已證實座標。
- 原版 `sub_B7491`／`sub_B7289` 是即時動畫並於完成後自動返回。remake 的一次性解算沒有原始
  時間軸，因此戰後定格與「繼續」是明標的呈現近似，不冒稱原版輸入。
- `COLGCBT` 或 palette provider 缺席時，兵種圖與面板框可安全退回純色／線框；固定文案與
  結果仍須完整顯示，不得 panic 或警告洗版。
- 攻守方單位必須分別使用 `GroundInvasionResult.AttackerColor`／`DefenderColor`，依
  `sub_B8EFB` 的 raw index 表替換 `0xC0..0xC7`、`0xE8..0xEB`；不可用轟炸格線的三色
  RGB helper 代替。raw color 2 保留原生色，`#21` 面板框不作玩家色替換。

## 文字安全框

- 標題、兩側每一列、戰果與按鈕各有獨立 `textSafeRect`；所有文字由 `drawLeft`、
  `drawCentered` 或同等受限 helper 繪製，`groundcombat.go` 禁止直接呼叫 `fnt.Draw*`。
- 面板列不得超過各自 261×149 外框、相鄰列或中間空白；中英文最長格式與 32-bit 數值都須量測。
- 按鈕文字安全框和 `(265,232,110,30)` 可見外框／熱區共用整數中心。
- 標題若含長殖民地名採單行截斷；戰果採單行截斷，不得穿入兩側面板或按鈕。

## 驗收

- 所有 `groundcombat.*` 鍵中英文都存在，來源檔不再包含 `.tr(`、直接字型繪製或代表性固定文案。
- 雙語、長殖民地名、最大整數、三種戰果與 fallback 都通過安全框測試。
- Docker + Xvfb 畫廊產生 35/35，人工抽查 `17_groundcombat.png` 的兩側面板、戰果與按鈕；
  真 `COLGCBT.LBX` 與缺資產 fallback 分開抽樣。
