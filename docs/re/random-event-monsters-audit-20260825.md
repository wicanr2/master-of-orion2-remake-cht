# 隨機事件 19–23 怪獸入侵靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`（MOO2 1.31），SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 資料庫：`Orion2.exe.i64` 的 `/tmp` 可寫副本；原始資料庫與執行檔唯讀。
- 工具：IDA Pro 9.4、IDAPython image `ida-pro-9.4-idapython:py312-v1`。
- 位址基準：IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_random_event_monsters.py`；本輪 JSON 位於
  `/tmp/moo2-event-monsters-work/random-event-monsters.json`，不加入 Git。
- 太空鰻分裂追加匯出器：`tools/ida/audit_space_eel_split.py`；JSON 位於
  `/tmp/moo2-surrender-audit-20260825/space-eel-split.json`，同樣不加入 Git。
- 註記契約：下列名稱只供導覽；原始函式名、線性位址、指令 bytes 與運算元均保留於 JSON。

## 已證實

1. `Determine_Event_ @ 0x2230A` 以相對星曆限制事件 19／22／20／23／21 分別不得早於
   `100／150／200／250／300` 回合；每個限制分支都另呼叫 `sub_233FA`，超空間亂流存在時
   候選失敗。這補強先前只有手冊最早回合的證據。
2. 事件 19–23 共用建立分支 `0x22C1C`：以一般受害帝國槽呼叫 `sub_23BEC @ 0x23BEC`，
   把結果寫入九位元組事件 record 的 `+0x07` 星系欄；沒有候選時清除 record。
3. `sub_23BEC` 逐星掃描目標帝國擁有的星系，再逐一檢查五個殖民槽；有效殖民地以
   `Random(candidateCount)==1` 做 reservoir sampling。因此目標是受害帝國的有效殖民星，
   不是第一顆無主星，也不是全銀河等機率星系。
4. `sub_206A2 @ 0x206A2` 的事件 19–23 case 在 record 狀態 6 時呼叫
   `sub_A16BF @ 0xA16BF`，傳入保存的目標星、旗標 1 與怪獸類型 `10／11／12／13／14`。
   `sub_A16BF` 分別呼叫五個不同 loader，再建立並路由 owner 8 艦艇。依事件表與既有 loader
   對照，順序是變形蟲、水晶、巨龍、太空鰻、九頭蛇；事件 22 絕非變形蟲替身。
5. `sub_A1A23 @ 0xA1A23` 會寫艦艇移動旗標與目的地，並呼叫既有航行座標 helper；怪獸事件
   使用真實艦隊／艦艇路徑，不是直接在目標星建立靜態守衛物件。
6. `sub_DB8D8 @ 0xDB8D8` 是怪獸逐回合 consumer。它掃 active ship `+0x63 >= 8`；
   太空鰻以 raw owner/type `+0x63 == 13` 識別，且只在停泊狀態 `+0x64 == 0` 推進。
   `0xDB9AC` 將 ship `+0x61` 加一，恰好等於 `30` 時先歸零，再檢查全銀河 active、
   status `<5` 的 type 13 數量。數量低於 4 才呼叫 `sub_57D14` 載入新太空鰻，經
   `sub_100010` 在母體目前星系座標建立，新船 `+0x61` 亦寫 0。這閉合了手冊所述
   「每 30 回合分裂」的精確 age、位置與全局上限。

## 強推論與尚未知

- **強推論：**record 狀態 6 是首次新聞後進入效果 consumer 的建立階段；精確 GNN 顯示與
  owner 8 艦艇建立的同回合先後，尚未以動態 oracle 驗證。
- **未知：**途中截擊及五個 loader 產生的完整 blueprint。初始航速、星距／ETA 與停泊轉換已由後續
  `docs/re/event-monster-route-audit-20260825.md` 閉合；原版視窗捲動 globals 未進 remake，
  故出生邊界仍是明示座標近似。殖民地戰鬥、三輪轟炸、`/40`、共用傷亡與殖民地移除已由
  `docs/re/event-monster-colony-battle-audit-20260825.md` 閉合控制流；數值仍受 blueprint 缺口限制。
- remake 現以 `MonsterGuard.TransitETA` 保存 owner 8 首次航行投影；航行中不算守衛，抵達後
  才盤據目標星。它仍不是可在中途逐座標截擊的完整 owner 8 ship record。

## 推翻的舊結論

- 「事件怪獸優先第一顆無主星」：錯；原版從受害帝國有效殖民星 reservoir sampling。
- 「太空鰻可用變形蟲代打」：錯；原版有獨立 type 13 loader。
- 「五種怪獸只有手冊最早回合」：不完整；1.31 指令亦有相同五個門檻及亂流排除。
