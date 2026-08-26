# 共用文字輸入彈窗文案與版面稽核（2026-08-27）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython
- 位址基準：IDA linear，DOS/4GW LE object #1
- 唯讀匯出：`tools/ida/audit_input_box_ui.py`
- 本輪輸出：`/tmp/input-box-ui-20260827.json`（一次性大型證據，不提交）

## 已證實

本輪 IDA 重跑確認既有筆記中的五個函式皆有獨立、連續邊界：

| 原始函式 | 邊界 | caller／關係 |
|---|---|---|
| `sub_91B89` | `0x91B89..0x91BB4` | 輸入 callback 候選 |
| `sub_91BB4` | `0x91BB4..0x91BD4` | 外部符號所稱 `Remapped_Input_Box_Popup_`；由 `sub_F5777` 與 `sub_5D2BB` 呼叫 |
| `sub_91BD4` | `0x91BD4..0x91F14` | 輸入彈窗繪製候選；由 callback 與 layout 路徑呼叫 |
| `sub_91F14` | `0x91F14..0x9222A` | 輸入彈窗 layout／生命週期候選；呼叫 `sub_91BD4` |
| `sub_F5777` | `0xF5777..0xF5883` | 外部符號所稱 `Change_MP_Game_Name_`；直接呼叫 `sub_91BB4` |

原始指令立即數已由既有稽核與本輪函式匯出交叉確認：

- 彈窗起點 `(0xB1,0x7D)`；
- 標題帶 `y+3`、高 `0x36`；
- 輸入欄 `(x+0x22,y+0x36)`、高 `0x1A`；
- OK 按鈕 `(x+0x60,y+0x64)`；
- 全域長度上限 `0xCD`；
- `sub_91BD4` 使用 `288-0x36=234` 的輸入欄寬。

`INBOX.LBX#0` 為 288×151、`#1` 為 98×28 的既有資產量測，與上述幾何吻合。

## 強推論與移植決策

- 原版標題由 caller 傳入；「對局名稱」與「輸入主機位址」是 remake 各玩家路徑的語意鍵，
  不是 `INBOX.LBX` 可直接翻譯的烘字。
- 原版逐掃描碼輸入不適合現代 IME。remake 維持 Ebitengine `AppendInputChars`，並保留 Enter 接受、
  Esc 取消、rune 長度與 205 上限；這是平台移植近似，不深挖 Win95／鍵盤 API 內部。
- `INBOX.LBX#1` 烘有英文 ACCEPT；繁中需擦底疊字，英文可顯示外部 catalog 的 ACCEPT，確保缺資產
  fallback 與正版資產走同一語意來源。

## 本輪發現的版面問題

- 舊程式用 `(titleH-titleSize)/2` 計算標題中心，把字級數值當成實際 glyph 高度；這不是原版
  `54px` 標題帶的中心契約。
- 舊鍵盤提示以 `buttonBottom+4` 當文字中心，實際 11px bitmap glyph 高 16px；本輪畫廊更證實
  即使邏輯座標仍在 288×151 內，字墨也會壓到 `INBOX.LBX#0` 的下邊框。安全位置是輸入欄底緣
  `y+80` 到按鈕頂緣 `y+100` 的 20px 空帶。
- OK 鈕 14px bitmap glyph 高於擦底內框；改用可在 18px 內框放下的 12px 樣式，仍維持按鈕中心。

## Remake 映射

- `inputBox` API 收 `titleKey`，不收已翻譯句子；共用 ACCEPT、鍵盤提示與 caller 標題皆由
  `assets/i18n/ui.json` 提供。
- 標題、輸入欄、OK 與提示各有 `textSafeRect`；游標寬度預留仍由實際字型量測。
- host 預設名亦改讀外部鍵，避免「主機玩家／Host」留在多人呼叫端。
- Docker + Xvfb 中文 `-gamegallery` 於修正前後各產生 35/35 張；第二次
  `34_inputbox.png` 目視確認提示位於輸入欄與按鈕間，未再壓住下邊框。

## 未知與停止線

- 原版標題字型的逐像素 glyph metrics、掃描碼插入模式與游標閃爍週期維持未知。
- 這些差異落在現代文字輸入平台契約；依專案停止線採 IME 友善近似，不為 Win95 API 內部另開 RE。
