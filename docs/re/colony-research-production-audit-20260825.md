# 殖民地研究產出靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython
- 位址基準：IDA linear，DOS/4GW LE object #1
- 匯出腳本：`tools/ida/audit_research_breakthrough.py`
- 方法：唯讀開啟 `.i64`，保留原函式名、線性位址、bytes、運算元與 caller／callee；未修改資料庫。

## 已證實

### 每位科學家的產出

`sub_DE280 @ 0xDE280` 以 mode 2 計算研究人口產出。其 dispatch slot
`0x183564` 指向 `sub_DFE77 @ 0xDFE77`：

- `0xDFF11..0xDFF1C` 讀取玩家／種族 runtime record `+0x8A3` 作為該人口的研究基礎值。
- `0xDFF1C..0xDFF31` 只有「人口種族等於殖民地擁有者，且 runtime 欄位 `+0x16B == 3`」時另加 1。
- 此函式沒有讀取 Colony 的四個研究建築旗標。
- `sub_DE280` 逐一處理 job 2 人口，再於 `0xDE651..0xDE65D` 以 `(sum + 10) / 20` 回傳整數研究量。

### 研究建築的固定產出

`Colony_Research_Production_`（原名 `sub_DFF74 @ 0xDFF74`）先於
`0xDFFE1..0xDFFED` 呼叫 `sub_DE280(colony, 2)`，之後才分別檢查 Colony 建築旗標：

| Colony offset | 原版 building ID | 建築 | 固定 RP |
|---|---:|---|---:|
| `+0x159` | 35 | Research Laboratory | 5 |
| `+0x154` | 30 | Planetary Supercomputer | 10 |
| `+0x149` | 19 | Galactic Cybernet | 15 |
| `+0x13C` | 6 | Autolab | 30 |

`0xE0041..0xE0050` 將四個固定值與人口研究量相加，`0xE0050` 寫回
Colony `+0xEB`。四項加成都只套一次，沒有乘上科學家人數。

## 勘誤

舊文件把手冊摘要解讀為 Research Laboratory／Planetary Supercomputer／Galactic
Cybernet 同時增加「每位科學家」與固定研究，Go 也因此重複加成。這與原版最終消費端直接
矛盾；本輪以 `sub_DFE77` 與 `sub_DFF74` 的完整上下游資料流否定舊斷言。原始手冊頁碼仍可作
建築用途說明，但不能推翻執行檔實際結算。

## 尚未在本切片宣稱的項目

- `sub_DE280` 的全部人口忠誠、士氣、領袖、事件與種族修正尚未逐欄位完成命名。
- 本切片只閉合研究建築是否修改 per-scientist，以及四個固定值；其餘欄位維持未知或既有證據等級。
