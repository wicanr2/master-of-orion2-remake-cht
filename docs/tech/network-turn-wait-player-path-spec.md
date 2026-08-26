# 正式網路回合等待玩家路徑規格

## 目標

保留 remake 的現代 TCP、兩階段決定性鎖步、重連與失敗即關閉規則，但讓正式
`networkWaitScreen` 使用已逆向的 `Net_Next_Turn` 面板、玩家列、聊天記錄與輸入列，
不再維護一張只在正式路徑出現的自製替代畫面。

## 證據與近似邊界

- **已證實**：原版 `sub_FC470` 分流主機 `sub_FBFE2` 與客戶端 `sub_FC2D2`；客戶端呼叫
  `sub_F7E95` 送單，主客戶端都使用 `sub_F1075` renderer。
- **已證實**：面板尺寸、輸入列 `0xBB／0x11`、玩家列步距 `0x19`、14 則聊天、speaker 8
  的 GNN 分支與兩種前綴格式。
- **現代轉接**：第二階段重播指紋、resume token、心跳、HMAC 與 TLS 不冒稱原版協定；
  它們只改同步／錯誤狀態，不另建替代版面。
- **未知**：第一個玩家列 y 與原版旗色取值鏈仍未閉合，維持既有標註近似。

## 正式路徑契約

1. `submitNetworkTurn` 後建立的正式等待畫面必須持有 `Net_Next_Turn` renderer；其 table、
   roster、目前席位與 tick 每幀同步正式 `networkTurnState`。
2. 等待期間接受文字輸入、Backspace 與 Enter；本機訊息立即進聊天記錄，並以獨立
   `KindChat` 封包送出，不進鎖步指令表。
3. 正式 update loop 是唯一網路封包 consumer；renderer 不得再 poll session，避免聊天畫面
   搶走 `turn_done`、`turn_ready` 或 `desync`。
4. 動態錯誤仍採失敗即關閉：在原版面板上以受限錯誤框顯示，點擊後才清理網路並返回多人設定。
5. 所有固定轉場、錯誤模板與後備玩家名由 `assets/i18n/ui.json` 的 `network.*` 鍵提供；
   `networkgame.go` 不保存中英文玩家句子。

## 驗收

- Host／client 共同開局、送單、重播指紋、分岔與聊天測試均走正式 `networkWaitScreen`。
- 靜態測試禁止 `networkgame.go` 出現 `.tr(`、直接字型繪製或固定玩家文案。
- runtime 字型測試驗證錯誤標題、細節及返回提示均在原版下段面板內。
- 完整 Go 測試通過；正常雙 peer 測試證明聊天與鎖步訊息不互相吞掉。

