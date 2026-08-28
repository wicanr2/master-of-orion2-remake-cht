# 殖民／前哨站完整垂直鏈稽核（2026-08-28）

## 結論

原版 1.31 的殖民與前哨站鏈已由正常玩家入口、AI 入口、資格判定、艦船／殖民基地消耗、
共用建立器、前哨站升級、人口與原住民初始化、衍生重算及玩家通知閉合。此結論只代表
玩家可見玩法的靜態 RE 證據已完整，不代表 remake 已對齊；依 RE-first gate，本輪不修改 Go。

## 證據身分

- 輸入：`Orion2.exe`
- 輸入 SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 外部符號表 SHA-256：`f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`
- 工具：IDA Pro 9.4／IDAPython；位址均為 DOS/4GW LE object #1 的 IDA 線性位址
- 非破壞性匯出：[`colonization-full-ida-20260828.json`](evidence/colonization-full-ida-20260828.json)
- 腳本：[`audit_colonization_full.py`](../../tools/ida/audit_colonization_full.py)

外部符號名稱只供導覽；以下結論以原始位址、指令、交叉參照與資料流為證據。

## 資格與正常玩家路徑

### 可殖民行星

`sub_714A1 @ 0x714A1` 要求 planet index 非負、`planet+0x04 == 3`，且 colony index 為
`-1`，或該 colony 的 `+0x06 != 0`。`sub_E60C8 @ 0xE60C8` 再加入 owner gate：已存在的
前哨站只有屬於目前玩家時才算可殖民。因此不能直接把敵方前哨站升級成殖民地。

`sub_E6132 @ 0xE6132` 計算可放前哨站的空 planet slot；它只接受 colony index `-1`。
planet 類型的進一步畫面限制不在這個計數器內，不能把此 helper 單獨解讀成完整 UI 規則。

### 按鈕、選星與消耗

`sub_71F35 @ 0x71F35` 只在選取艦隊屬於目前玩家、位於目前星系、至少有 type 1 殖民船，
且五個行星中至少一個通過 `sub_714A1` 時開啟殖民按鈕。`sub_8B2DE @ 0x8B2DE` 經正常
選星畫面取得 planet，找到選取清單中的 type 1 艦，呼叫 `sub_E6071`，再把該 ship
`+0x64` 寫成 5 並遞減選取與玩家艦數；殖民船確實被消耗。

`sub_FDB01 @ 0xFDB01` 是回合殖民 dispatch。它同時處理 type 1 殖民船、type 4 前哨船、
殖民地產品 raw 11 的 Colony Base、自動殖民與多人回傳。選擇成功後呼叫 `sub_E6071` 或
`sub_E607F`；用 Colony Base 時呼叫 `sub_145EA @ 0x145EA` 移除 raw building 11，使用艦船時
走 ship removal。取消或失去合法目標時，Colony Base 退還 `dword_17EC16 / 2` BC 並清除
auto-colonize flag。

## 共用建立器

`sub_E6071 @ 0xE6071` 令第三參數為 0，`sub_E607F @ 0xE607F` 令第三參數為 1；原始
`0xE6071..0xE608A` 指令顯示兩者共用 `0xE6078` 的 call tail，避免反編譯器把後者誤顯示成
不明跳轉。兩者都進入 `sub_E5EB3 @ 0xE5EB3`。

### 新前哨站

建立器配置空 colony record，經 `sub_12D75 @ 0x12D75` 清零並連回 planet，然後寫：

- `colony+0x06 = 1`：前哨站狀態；
- `colony+0x0A = 0`：沒有人口；
- `colony+0xE2 = planet+0x08`；
- `colony+0x115 = -1`，最後重建星系 colony masks/cache。

### 新殖民地與前哨站升級

一般殖民地從一名 owner population 開始。`planet+0x0B == 0`、Lithovore
`player+0x8B1 != 0` 或 Cybernetic `player+0x8B0 != 0` 時第一人為工人，否則為農夫；packed
record 同時寫 owner/race bits。planet raw `+0x0F == 6` 時，額外建立三名 raw race 9 的
Native farmer，總人口為四，並清除 planet 與 star 的 Native marker。

若目標原本是自己前哨站，建立器重用同一 colony record，並把 `colony+0x14C` 寫 1。
該 offset 的其他 producer／consumer 與既有 building table 共同證實它是 raw building 22
Marine Barracks；這不是未命名的「前哨升級旗標」，而是升級後附帶的實際建築狀態。

建立後依序呼叫 `sub_E2A70 @ 0xE2A70` 重算 colony derived state，寫
`colony+0x123 = -39`，再呼叫 `sub_E5296 @ 0xE5296` 重建 star owner、事件與能力 masks。
政府沒有在建立器內複製成獨立 colony 欄位；後續重算由 owner player 狀態動態取得。

## `colony+0x123` 的語意勘誤

`sub_C035E @ 0xC035E` 是此欄位的玩家通知 consumer。它讀到 `-39` 時取得 colony planet 名，
使用 raw text 64 顯示訊息，之後一律清回 `-1`。同一個 `-39` 也由
`sub_E9927 @ 0xE9927` 在系統發現事件建立新 colony 後寫入。因此可證實：

- `+0x123` 是待顯示的殖民地產品／狀態通知碼，不是建築或永久旗標；
- `-39` 是「新 colony 已建立」類通知碼，涵蓋殖民與系統發現建立器；
- raw text 64 的逐字文案尚未從 JIMTEXT 資產核對，故正式顯示字串仍列未知，但不阻塞玩法規則。

同一 consumer 另把 `-40`、`-41` 導向 raw text 65、66，其他非負值先進產品資訊 helper 再顯示
raw text 57；這些交叉參照進一步排除 `+0x123` 是永久 colony 狀態的舊解釋。

## 殖民候選、衝突與 AI

`sub_E6170 @ 0xE6170` 每回合從停泊的 type 1／4 艦與活動 Colony Base 建立 `star+0x37`
候選 owner。不同玩家同時具備殖民能力時設為 `-1`；候選還必須分別通過可殖民／可前哨 planet
計數。星系中另一玩家的停泊艦依方向外交 raw policy 4..6 也可使候選失效。失效的 Colony Base
會退還半價並清除活動旗標。

`sub_E65F8 @ 0xE65F8` 對 AI 候選星系掃描五顆空 planet，以 `sub_D27A7` 的最高分選擇目標，
呼叫同一個 `sub_E5EB3(..., 0)`。沒有目標時退還並移除 Colony Base；成功時消耗 Colony Base
或殖民船。AI 的 planet score 公式由獨立 AI 殖民評分證據列管理，本鏈只閉合其 consumer。

## 證據等級與剩餘邊界

- **已證實**：上述資格、owner gate、艦種、消耗／退款、共用 wrapper、前哨與殖民初始化、
  Native、前哨升級 Marine Barracks、重算 callback、通知碼 consumer、衝突候選及 AI consumer。
- **強推論**：`planet+0x0B` 的 typed 名稱為「自然食物為零」；它已與氣候食物表逐值吻合，
  但文件仍保留 raw offset。
- **未知但不阻塞玩法 RE**：raw text 64 的逐字英文文案、畫面 helper 的像素級動畫時序。
- **不在範圍**：`memset`、`sprintf`、檔案／平台 API、編譯器 stack／runtime helper 的內部行為。

## 對 remake 的影響

remake 已接一般殖民船消耗、起始人口／職務與 Native 分流，但須等 RE gate 全部關閉後另立
READY spec，統一補上前哨站 owner gate 與升級、Marine Barracks、Colony Base 半價退款、
候選衝突、通知碼、AI 共用建立器及 star/colony derived callback。不能只因既有測試通過就把本列
標成原版對齊。
