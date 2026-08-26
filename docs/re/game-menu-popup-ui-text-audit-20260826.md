# 對局內遊戲選單 IDA 與文案稽核（2026-08-26）

## 輸入與位址契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4，`metapc`；本文位址為 IDA DOS/4GW image 線性位址。
- 非破壞性匯出器：`tools/ida/audit_game_menu_popup_ui.py`；匯出保留 raw `sub_*`、邊界、caller、bytes、指令與資料參照。

## 已證實

- `sub_8012F @ 0x8012F..0x803D0` 是完整 Game Popup 外層。它呼叫
  `sub_7EA5C` 載入圖片，並依四路 switch 分流到 `sub_7DD41`主選單、
  `sub_7E00F`設定、`sub_7DA76`載入與 `sub_7E154`儲存。
- `sub_7DD41 @ 0x7DD41..0x7E00F` 是對局內主選單輸入迴圈，它在
  `0x7DFBB` 呼叫 `sub_7F701`繪製，並在音量欄命中時分別呼叫
  `sub_80892`與 `sub_80918`。
- `sub_7F701 @ 0x7F701..0x7FA28` 是主選單繪製函式；直接 caller 包含
  `sub_7DD41`、`sub_7ED66`、`sub_7EDF2` 與 `sub_8012F`。
- `sub_7EA5C @ 0x7EA5C..0x7ED4D` 從 `GAME.LBX` 依序載入資產 `0..7`；
  前八個寫入連續指標槽，對應背景、六顆按鈕與音量條。後續仍載入
  `0x1D`、`9..`等儲存／設定共用圖片，不可把整支函式誤寫為「只載主選單八張」。
- `sub_7E00F @ 0x7E00F..0x7E154` 與 `sub_7FA28 @ 0x7FA28..0x8011F` 證實原版
  SETTINGS 是 Game Popup 內的實際設定分頁，不是單一遷移連線開關。
- `sub_7EFEF @ 0x7EFEF..0x7F14C` 依序把全域 `0x199BDC`、`0x199BDD`、
  `0x199BDF..0x199BE9` 複製成 13 個畫面 word；`0x199BDE` 是 Random Events，刻意跳過。
  `sub_7F14C @ 0x7F14C..0x7F206` 以相同順序回寫。`sub_7E00F` 掃 13 個 widget，
  命中後切換對應 word；`sub_7FA28` 以 17 px 列距繪製 13 列。
- 13 個標籤依序使用字串資源 `0xA6..0xB1` 與 `0x187`；資產為
  `GAME.LBX#29`（279×378 背景）、`#9`（22×12、兩幀開關）及 `#10`（75×20 接受鈕）。
- `sub_127E1 @ 0x127E1..0x12937` 的原始指令已證實預設值：開啟回合摘要、回合等待、
  自動選艦、動畫、遷移線、GNN 與自動存檔；其餘六項關閉。`sub_12937` 把從
  `0x199BDC` 起的 `0x229` bytes 寫入 `mox.set`。這是玩家可見設定契約；檔案服務內部不深挖。
- `byte_199BE6` 唯一的非載入／回寫直接資料參照位於 `sub_B2542 @ 0xB2542..0xB2FFA` 的
  `0xB2660`。值非零時跳過 `Confirmation_Box_ @ sub_77658`，直接進入七格刪除迴圈；值為零時
  先顯示 `E_Strings(0x3B)`：「`%s is above a product...`」，玩家拒絕便重繪建造佇列而不退出。
- `sub_B09CE @ 0xB09CE..0xB0A24` 先找七格中的 raw `-3`、`-2`、`-10`，再檢查其後是否仍有
  非空格；`sub_AFC6D @ 0xAFC6D..0xAFD42` 分別把三者映射到一基底 E_Strings ID
  `0x142`、`0x21D`、`0x88`。正版 `ESTRINGS.LBX` 以 `loadStrings(offset=6)` 交叉驗證後，
  三者依序是 `Housing`、`Trade Goods`、`^ Repeat ^`；ID `0x3B` 對應零基底第 58 條確認句。
- `sub_B2150 @ 0xB2150..0xB2190` 會把指定格後方資料左移，並在移位後目前格為 raw `-10`
  時繼續移除。`sub_B2542 @ 0xB2679..0xB26AB` 則重複掃描七格，只刪除仍位於其他非空項之前的
  上述特殊模式；最後一個模式或一般產品保留。因此設定的精確語意是「離開佇列時自動確認清除
  阻塞產品的 Housing／Trade Goods／Repeat」，不是每回合或完工後無條件刪除。
- 外部符號所列 `Print_Options_To_Bitmap_ @ 0x7EDB1` 在目前 IDA 資料庫不是
  函式邊界；相鄰 `sub_7ED66` 精確結束於 `0x7EDB1`。本輪不以符號名強行建函式。

## 強推論與 remake 邊界

- openorion2 `MainMenuWindow::initWidgets` 給出視窗 `(144,25)`、六顆按鈕相對
  座標及 `GAME.LBX#0..6`資產對應；它與 IDA loader `#0..7` 連續鏈獨立吻合，
  列為強推論。
- 按鈕英文烘在圖上；現行繁中譯文、缺資產 fallback 英文與遇錯訊息均是
  remake 文案，不冒稱原版逐字資源。
- 原版座標的列距、資產尺寸與控制數量已證實；背景內的可用文字區與首列位置依實際資產
  交叉驗證，列為強推論。remake 使用固定文字安全框，不能以量圖座標反向升格為指令級證據。
- 部分開關依賴原版畫面節奏或尚未閉合的消費端；remake 可保存其值，但只有已有明確玩家
  契約的消費端可宣稱生效。Windows／DOS 平台 API 與 `mox.set` 檔案服務維持停止線。
- remake 的 Repeat Build 是 `RepeatBuild[]` typed 狀態，不把 raw `-10` 插入七格佇列；因此
  本輪精確重現 Housing／Trade Goods 的「模式在產品之前」掃描與確認流程，raw `-10` 的畫面
  表示法保留為資料模型差異，不把 typed Repeat 狀態誤刪。
