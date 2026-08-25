# 逐人口 prisoner 產出稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_colonist_production.py`
- 位址空間：IDA linear；DOS/4GW LE object #1
- 證據型態：靜態、唯讀；未改名或寫回 `.i64`

## 已證實

1. `sub_DE280 @ 0xDE280` 從 colony `+0x0C` 以 4-byte stride 反向走訪
   packed colonist；`[colonist+1] & 2` 是 raw bit 9，也就是 parser 右移九位後的
   `WORKING = 1`。只有該 bit 成立的人口參與產出。
2. 同函式以 `(raw << 23) >> 30` 取 bits 7..8 的 job，再以 `raw & 0xF`
   取 race；因此產出是逐人口、逐職務、逐種族計算，不是只乘 colony 的三個總數。
3. `[colonist+1] & 4` 是 raw bit 10，也就是右移後的 `PRISONER = 2`。
   `0xDE4CF..0xDE4DE` 對這類人口額外加入 `base * -5`；同一人口的基礎項在
   `0xDE479..0xDE485` 先乘 20。因此 prisoner 淨產出為 `15/20 = 3/4`。
4. `sub_DE22C @ 0xDE22C` 依 job dispatch，且把 colonist race 傳給農業、工業或
   研究 helper。`sub_DDF2C @ 0xDDF2C` 讀 colony gravity、colonist race 與該 race
   的 gravity tech，回傳逐人口重力調整碼；它不是 loyalty 或 3/4 helper。
5. `Apply_Colony_Pop_Growth_`（`sub_E2DCA @ 0xE2DCA`）同樣保存逐人口 race 與
   `PRISONER`。`0xE3030` 測試的 `0x180` 是 job bits，不是 loyalty。

## 強推論與未知

- openorion2 的格式註解稱 loyalty 在同化前仍指向前主，與 packed 3-bit 欄位形狀一致；
  但本輪沒有找到 loyalty 直接改變食物／工業／研究的讀取端。因此不把 loyalty 值加入
  remake 產出公式。
- 原版同化會清除某一筆 packed colonist 的 `PRISONER`，但本輪靜態切片未閉合其選人順序。
  remake 沒保存 colonist 陣列順序，採固定且可重播的職務順序，明列為近似。

## 推翻的舊斷言

`internal/gamedata/production.go` 原稱 remake 不知道未同化人口在哪個職務，只能按總人口
比例分攤。這只對舊 JSON 狀態成立；原版 `.GAM` 已保存逐人口 job 與 prisoner flag，現在可
精確匯入。比例法保留為舊存檔遷移 fallback，不再是新狀態的主要模型。
