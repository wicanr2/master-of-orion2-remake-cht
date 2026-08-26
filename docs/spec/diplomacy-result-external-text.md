# 外交結果碼與外部文案規格

## 證據邊界

- `docs/re/diplomacy-gifts.md` 已證實原版現金餽贈的國庫轉移方向，以及 raw
  `sub_539D9 @ 0x539D9` 的兩組回應門檻與訊息 ID `0xA5..0xA9`。
- 各 remake 行動目前顯示的完整中文／英文句子沒有逐句原版證據，屬於**等義轉接文案**；
  不得留在規則層或標成原版原文。

## 分層契約

- `internal/shell` 只回傳 `DiplomacyResult`：穩定 `DiplomacyResultCode`、對手名及必要的
  金額、可用金額、科技名或殖民地名參數。
- 空 `Code` 表示找不到對手或未知 action；狀態修改與玩家指令記錄行為不變。
- `cmd/moo2` 將 code 映射到 `diplomacy.response.<code>`，固定模板只存在
  `assets/i18n/ui.json`。缺鍵時必須使用 `diplomacy.response.unknown`，不可把鍵值畫上畫面。
- 規則層不得 import UI／i18n；UI 不得依句子內容推回規則結果。

## 格式參數

- 一般結果模板只有一個 `%s` 對手名。
- 現金不足依序使用對手名、目前 BC、要求 BC；成功使用對手名、金額。
- 科技與殖民地成功依序使用對手名、外部資料名稱。

## 驗證

- shell 測試只檢查 typed code 與狀態，不依中文句子判斷成功。
- UI 測試逐鍵檢查雙語存在、格式參數數量與未知 code fallback。
- 靜態測試禁止外交規則檔保存既有玩家回應句子。

