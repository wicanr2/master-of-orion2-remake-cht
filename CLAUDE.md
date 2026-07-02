# 銀河霸主2 中文化

- 使用 openorion2 @~/moo2/openorion2 
- github repo (工作repo) 只放 patch

# 工作目標 master-of-orion-2 remake in golang/ebiten

* 基於 openorion2 為參考基底改寫變成 go/ebiten 版本
* 使用 patch 1.3 & 1.5 建立對應數據檔案
* 允許 在主選單選擇版本 (1.3 or 1.5) 
* 允許 在主選單選擇中文/英文 不同的語言
* 完整中文化所有訊息 包括 button 
* 重新檢視 openorion2 確保遊戲規則都有實作進去 (可以參考手冊, patch 1.5 有額外的手冊)
* github repo 要撰寫致謝 
* 要從網路收集銀河霸主2的相關中文討論資訊, 整理到專案一個獨立章節討論(角色: 遊戲歷史考究專家)
* 討論銀河霸主2在華人圈中的文化現象 (要用文案作家的角色)

# 工作順序 
* 遇到沒有把握的問題基於**第一性原理**討論, 不要假設或者過度依賴先前的記憶
 - 第一性原理的參考請參考目前的 skill & rule 
* 先研究可行性 每一輪都更新到 github repo (worklist 允許擴充)
 - 遊戲規則研究
 - 字型選擇
 - 參考先前的中文化經驗, 統整中文化策略
 - 按鈕的中文化一定要參考 先前的專案經驗, 避免重蹈覆轍
 - 建立 kick-off 知識庫
* 定義 PLAN.md
* 定義 WORKLIST.md
* 逐步建立技術文件知識庫 ex: ebiten porting 心得 / 音樂整合 / keyboard / mouse 整合 / patch 如何處理 / 選單如何擴展 
* 每一輪過做完都要盤點 這個 round 建立的技術文件 與 audit or worklist markdown 有沒有衝突部份, 有衝突的, 要清除錯誤的斷言避免佔 token

# 其他需求
* 建立一個 markdown 討論 sprite / tile 畫質優化或者重新設計的可行性
* 建立一個 markdown 討論 UI 界面調整的可行性 

# 遊戲原始檔(包含手冊)
- @original_game

# patch 1.31
- @moo2_patch1.31

# patch 1.5(包含手冊)
- @moo2_patch1.5

# 參考先前的專案  
- @~/master-of-orion (這是一個 github repo, 銀河霸主中文化(基於sdl2))
- @~/master-of-magic (魔法大帝中文化, 使用 go/ebiten)

# 放置 github repo 
https://github.com/wicanr2/master-of-orion2-remake-cht.git

# 本機授權
- 遊戲測試請在背景測試 (in docker) 
- 遊戲編譯請在 (docker)
- 允許使用 gh 命令 (github cli)
