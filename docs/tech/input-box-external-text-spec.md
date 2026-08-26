# 共用文字輸入彈窗外部文案與安全框規格

證據來源：[`docs/re/input-box-ui-text-audit-20260827.md`](../re/input-box-ui-text-audit-20260827.md)。

## API 契約

`sceneBuilder.inputBox(under, titleKey, initial, max, onOK)` 的第二個參數必須是穩定 `ui.json`
鍵，不得傳入已翻譯句子。畫面建立後依目前語言即時查詢：

- `inputbox.title.game_name`
- `inputbox.title.host_address`
- `inputbox.button.accept`
- `inputbox.hint.accept_cancel`
- `network.host.default_name`

若未來新增用途，先新增 `inputbox.title.*` 鍵，再接 caller；不得恢復自由字串 API。

## 邏輯座標與安全框

彈窗 `(x,y)=(177,125)`、尺寸 288×151：

| 欄位 | 安全框 | 字級 | 策略 |
|---|---:|---:|---|
| 標題 | `(x+10,y+3,268,54)` | 15 | 單行置中省略 |
| 輸入內容 | 原版欄 `(x+34,y+54,234,26)` 內縮 | 14 | 單行省略，另預留游標寬 |
| ACCEPT | OK 熱區 `(x+96,y+100,98,28)` 內縮 5 | 12 | 單行置中省略，中心與熱區相同 |
| 鍵盤提示 | `(x+10,y+80,268,20)` | 11 | 單行置中省略，位於輸入欄與按鈕之間 |

標題與按鈕的 center 必須由對應框計算，不以字級數字手算 y。輸入欄左右邊距 34／20 是原版
立即數造成的不對稱，不得為美觀改成對稱。

## 輸入行為

- 長度以 rune 計，呼叫端上限夾在 `1..205`；控制字元不進緩衝區。
- 初值超限先截斷；accept 前後去空白。
- Enter／Numpad Enter 接受、Esc 取消；框外點擊不穿透 modal。
- Ebitengine `AppendInputChars` 與 30-frame caret 是明標的平台近似。

## 驗收

- `inputbox.go` 不含 `.tr(`、玩家句子或直接字型繪製；三個現有 caller 都傳 title key。
- 五個外部鍵在英文與繁中存在；host 預設名不再內嵌於 `choosenetplyrs.go`。
- 實際 bitmap font 量測證明標題、輸入內容＋游標、ACCEPT 與提示不超出各安全框。
- ACCEPT 文字框與 98×28 熱區中心一致；提示不得侵入輸入欄底緣或按鈕頂緣。
- 原有長度、rune、控制字、退格、trim、Enter／Esc 與畫廊輸入框路徑測試保持通過。
