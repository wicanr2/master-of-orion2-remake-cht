# 帝國投降規格

證據來源：[`../re/empire-surrender-audit-20260825.md`](../re/empire-surrender-audit-20260825.md)。

## 狀態與時序

- 投降者只能是仍存活的 AI 帝國；接收者可以是玩家／目前熱座席位或另一個存活 AI。
- 建立時保存投降者與接收者的穩定 target kind／index，立即排入事件 34；資產不在 setter
  中移動。pending surrender 必須進 JSON 與多人快照。事件 34 不在規則層保存成品句子；顯示契約見
  [`empire-surrender-notice-external-text.md`](empire-surrender-notice-external-text.md)。
- 下一個投降 consumer 依建立順序處理。AI slice 不刪除，只把投降者清成 inactive，避免所有
  既有索引漂移。已失效、自己投給自己或接收者已滅亡的 record 採失敗即關閉，不做半套轉移。

## 可表示的原版資產結果

1. 殖民地連同 star／planet 平行索引、建築、駐軍、建造狀態與人口群組移交；owner race
   產出快取改成接收帝國，但人口群組的生物種族不被改寫。
2. 接收者取得投降者已完成的研究主題、明確科技與額外授予科技；每項新增科技走既有設計更新
   callback。研究中的進度與未完成選擇不移交。
3. 國庫與抽象貨運艦全數相加；投降者歸零。
4. 領袖移交但全部解除艦艇／殖民地任命；待聘 offer 不移交。
5. 投降者所有艦艇、艦型建造進度與進行中的艦隊任務移除，不得送給接收者。
6. 所有涉及投降者的戰爭、政策、關係、貿易／研究協議與玩家 Treaty 清空。
7. 原版逐來源帝國的間諜重新定向需要完整 target matrix；remake 現有玩家／AI 間諜模型無同構
   矩陣時，保留為明示資料模型限制，不得把投降者自己的 Spy／Agent 當成接收者所得。

## 觸發近似邊界

- 自動觸發只在 AI 對 AI 現代外交已啟用、投降者正處於戰爭、且其可觀察國力顯著低於合法
  接收者時執行；平手依穩定帝國順序。
- 這個 gate 標為 trigger approximation，直到 raw `+0x717` 與 `sub_27A3D` 全部評分欄位閉合；
  資產 consumer 則依上節已證實契約實作。

## 驗收

- AI→AI 與 AI→玩家各測 pending／事件 34／下一 consumer 的完整順序。
- 測殖民地所有平行陣列、科技聯集、BC／貨運艦、領袖解除任命、艦艇刪除與外交矩陣清理。
- 測非法 record 不造成半套轉移、投降者索引不漂移、JSON／多人快照往返。
- 全專案測試、格式、擁有權與 Docker 清理通過。
