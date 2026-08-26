# 自動選取艦艇設定稽核（2026-08-27）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4／IDAPython，映像 `ida-pro-9.4-idapython:py312-v1`；位址均為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出器：`tools/ida/audit_game_menu_popup_ui.py`。本輪新增的 root 只匯出原始函式名、邊界、caller、bytes、指令及資料參照，不修改 `.i64` 內的名稱或註記。

## 已證實

1. `byte_199BE1 @ 0x199BE1` 是 SETTINGS 第五列 `E_Strings(0xAA)` 的旗標；`sub_127E1 @ 0x127E1..0x12937` 在 `0x12817` 將預設值設為 1。
2. 除預設、設定頁載入／回寫與 `sub_8A216` 的切換 UI 外，唯一玩法讀取位於 `sub_70875 @ 0x70875..0x70F02` 的 `0x70C4E`。
3. `0x70C45..0x70C6A` 要求暫態旗標 `byte_199F00` 非零、`byte_199BE1` 非零，且目前畫面列的 owner `word_19999C` 是現行玩家；成立時以 `eax=1, edx=1` 呼叫 `sub_7229E @ 0x7229E..0x72346`。
4. `sub_7229E` 遍歷目前畫面艦艇列，經艦艇索引取得 owner，只有現行玩家的記錄進入選取分支；`eax=1` 表示寫入選取模式，`dl=1` 時先經 `sub_724CF` 的可選 predicate，最後把艦艇記錄選取 byte 寫為 1。函式結尾亦把模式寫入 `byte_199F08`。
5. 同一個 `sub_7229E` 另由手動艦隊操作 caller 使用；因此設定的玩家契約是「進入／切換到可操作的我方艦隊時，預先選取可選艦艇」，不是每幀強制全選。若每次重繪都套用，玩家手動取消後會立即被選回，與原版獨立的手動 caller 衝突。

## Remake 對映與限制

- remake 的 `sceneBuilder.shipPick` 是目前選中艦隊內的 typed 選取集合；沒有原版逐艦 raw flag。進入艦隊畫面或切換艦隊時初始化此集合，即可保留玩家可見契約。
- remake 目前所有 `Fleet.Ships` 都是我方且可由 ALL／拆分操作，沒有與 `sub_724CF` 同構的不可選 raw 狀態；本輪選取整支目前艦隊，列為資料模型對映，不冒稱 raw predicate 逐值一致。
- `Auto Select Colony` 的直接讀取位於 `sub_12479`、`sub_825A8` 與 `sub_86188`，並會呼叫不同的殖民地／畫面路由函式；其玩家契約尚未閉合，不能由本文件類推。
- Win95 輸入與 widget 內部呼叫屬既定停止線；只重製設定、進入時機、選取結果與玩家可取消的行為。

## Remake 驗證

- Docker + Xvfb 完整測試通過；開啟、關閉、手動全不選、切換不同艦數及拆分後重建均有針對性測試。
- 正版資料與 Noto Sans CJK 字型的中文 `-gamegallery` 產生 35/35 張。第一次人工抽查
  `07_fleet.png` 發現舊 `✔` 經 runtime 字型路徑變成缺字方框；選取標記改由 `ui.json`
  提供同寬 `[x]／[ ]` 後重抓 35/35，三艘預設選取均可辨識，未侵入艦名、艦級或損傷欄。
