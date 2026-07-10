# 交接文件（給接手的 Claude / 開發者）

> 這份是「重啟 session 後第一個要讀的檔」。目的:讓全新的 Claude 5 分鐘內接手,不重犯前一輪的錯。
> 最後更新:2026-07-04。搭配讀:[`HONEST-STATUS.md`](HONEST-STATUS.md)(現況真相)、根目錄 `CLAUDE.md`(專案目標)、`WORKLIST.md`(細項)、`PLAN.md`(階段)。

## 0. 先認清現況(最重要)

**還原度約 20%(使用者實測)。目前不是能玩的 MOO2,是「載入原版美術的中文畫面導覽器 + 幾個自製玩具系統」。**

**前一輪最大的錯:用「單元測試綠 + 新增自製系統」謊報還原進度。** 那些 gameplay 係數大多是自編的 remake 近似值,不是手冊真值。**接手後鐵律:還原度只用「對原版實測比對」評估,不看測試套件是否綠。** 測試只防自己的回歸。細節見 `HONEST-STATUS.md`。

## 1. 環境 / 路徑(這台機器)

| 項目 | 位置 |
|---|---|
| 工作目錄 | `/home/anr2/moo2` |
| GitHub repo(只放碼+文件+譯文,無版權資產) | `github.com/wicanr2/master-of-orion2-remake-cht`(分支 `main`) |
| 私有遊戲資料(玩家正版,**絕不入 repo**) | `/home/anr2/moo2-private-build/gamedata/mastori2`(324M 全套;`-game` 只用 ~18 個 LBX) |
| CJK 字型(Noto,OFL 可散布) | `/home/anr2/moo2-private-build/fonts/NotoSansCJK-Regular.ttc`(或 `/usr/share/fonts/opentype/noto/`) |
| 手冊 | `moo2_patch1.5/MANUAL_150.html`(**有文字**,patch 1.5 變更日誌+部分數值)、`moo2_patch1.5/GAME_MANUAL.pdf`、`original_game/…CD Manual.pdf`(**掃描圖,無文字,需 OCR**) |
| C++ 參考(**只有渲染,無遊戲引擎**) | `openorion2/src/`(對照解碼/座標用,不含戰鬥/AI/數值) |
| 1oom(MOO1 AI 參考) | 見記憶 `ebiten-cht-reference-paths` |

## 2. 建置 / 測試 / 執行(一律 docker,[HARD])

```bash
# 編譯 -game(需 CGO/X11 → moo2-ebiten image)
docker run --rm -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" -w /src moo2-ebiten \
  bash -c 'go build -buildvcs=false -o /dev/null ./cmd/moo2 && echo OK'

# 純 Go 單元測試(shell/engine/save/lbx/i18n/gamedata)
docker run --rm -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" -w /src moo2-ebiten \
  bash -c 'go test ./internal/...'   # 注意:internal/uifont 需 display,用 xvfb 才過

# headless 跑 -game 並截圖驗證(對原版比對的唯一自動化手段)
DATA=/home/anr2/moo2-private-build/gamedata/mastori2
FONT=/home/anr2/moo2-private-build/fonts/NotoSansCJK-Regular.ttc
docker run --rm -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" -v "$DATA:/data:ro" -v "$FONT:/font.ttc:ro" \
  -w /src moo2-ebiten bash -c \
  'go build -buildvcs=false -o /tmp/moo2 ./cmd/moo2 && xvfb-run -a /tmp/moo2 -game -data /data -lang zh -font /font.ttc -shot /src/out.png -frames 6'
# 然後用 Read 工具看 out.png,和原版對照(原版可用 DOSBox/openorion2 截圖)

# 完整本機測試 AppImage(自帶資料/字型,免 -data;產物在 dist/,gitignore)
bash scripts/package-appimage-full.sh
```

docker images:`moo2-ebiten`(CGO+X11+xvfb,已存在)、`golang:1.25-bookworm`(純 Go)。

## 3. 程式碼地圖:哪些是真的、哪些是自製

**扎實(有逆向/驗證基礎,可信可留)**
- `internal/lbx/`:LBX 解碼(容器/RLE/多幀 delta/調色盤鏈)。對照 openorion2 逐位元組驗證。
- `internal/save/`:原版存檔唯讀解析(SAVE10.GAM 全區段,有回歸護欄)。
- `internal/gamedata/`:少數真手冊公式(人口成長 colony.go、研究成本表)。**其他很多是估計。**
- `cmd/moo2/interactive.go`:16 畫面載入真 LBX 美術 + 擦底疊字中文化。畫面像原版。
- `assets/i18n/*.tsv`:UI 譯表(數百條)。

**自製 / 不忠實(當心,別信它的測試綠)**
- `internal/shell/session.go`:回合系統(經濟/研究/艦隊/人口/建築/事件/安塔蘭/種族/AI)。**大量係數是自編 remake 近似值**(程式碼註解自己標了「remake 調校值」)。與原版規則差距極大。
- `internal/shell/persist.go`:remake 自身 JSON 存檔(非原版 .GAM)。
- 各 `*_test.go`:只證明自製邏輯自洽,**不證明對齊原版**。

架構配方見記憶 `moo2-game-interactive-architecture`(但那是「畫面導覽器」架構,不是忠實引擎)。

## 4. 接手 worklist(依「對玩家體驗的影響」排序,不是好做程度)

> 每項的驗收 = **對原版實測比對**,不是加一個系統+測試綠。做之前先 Read `HONEST-STATUS.md`。

### 優先 1 — 音樂 / 音效(Phase 6,基礎已打通,2026-07-10)
> ⚠ **翻案**:MOO2 **沒有 XMI/MIDI 音樂**。全部音樂/音效是 LBX 內的 22050Hz 8-bit PCM WAV,原封播即與原版 bit-identical,**不需 SoundFont/OPL 合成**。定案見 `docs/tech/audio-format.md`。
- [x] 格式逆向 + ebiten 音訊整合(`internal/audio`)+ 主選單 BGM(STREAMHD)+ 按鈕音效(BUTTON1),`cmd/moo2/audiohook.go`;單元/真檔測試綠。
- [~] **曲目/UI 事件對應待對原版聆聽定案**:BGM 暫用 clips[0]、點擊暫用 BUTTON1(哪條是主選單主題、哪個 BUTTONx 對哪類按鈕,需 oracle)。
- [ ] 星系/戰鬥各場景的 BGM 對應;`CMBTSFX/SPHERSFX` 巢狀音庫(戰鬥音效)逆向。
- [ ] 桌面實測驗收(headless 停用音訊,聽感只能桌面驗)。

### 優先 2 — 忠實的新遊戲流程(目前跳程序生成假星系)
- [ ] 獨立**種族選擇畫面**:原版有 13 族肖像 + 自訂種族點數(RACEOPT.LBX)。目前擠在設定畫面一格,要拆成真畫面。
- [ ] 真實星系生成 + 母星配置(對照原版新遊戲產生的初盤,不是我的 genGalaxy 亂數)。
- [ ] 進遊戲後載入**真實殖民地初始狀態**(母星人口/建築/科技),不是 demo colonies。

### 優先 3 — 按鍵 / 熱區逐畫面像素對齊
- [ ] 現在多數熱區是估計座標,不少是「整畫面當返回鍵」。要逐畫面用截圖 + 原版對照,量每個按鈕真實座標。
- [ ] 方法見 `re-retro-cht-rulebook` skill(逆向/老遊戲中文化路由)+ PIL 單欄掃描量測法(見記憶 `moo2-game-interactive-architecture`)。

### 優先 4 — 忠實 gameplay 規則(主體工作量,對應 PLAN「從零重建引擎」軌)
- [ ] 殖民地:格子地形 + 每格產出 + 30+ 建築全表 + 污染/食物/貿易真公式(手冊為權威)。
- [ ] 科技樹:每主題在數科技間**抉擇** + 真 RP 成本表(gamedata 已有 cResearchCosts)。
- [ ] 戰鬥:真實武器機制(命中/傷害/射程/防禦/飛彈躲避/球狀傷害/地面戰)。
- [ ] 艦艇設計:艦體空間格 + 每元件佔格 + 改造 mod。
- [ ] 把 session.go 裡自編的近似係數逐一換成手冊/逆向真值,或在 UI/文件明標 remake 近似。

## 5. 鐵律(接手必守)

1. **驗收看原版實測,不看測試綠**(前一輪的教訓,記憶 `moo2-fidelity-20pct-not-test-green`)。
2. 編譯/測試**一律 docker**([HARD]);Python 用 docker uv.venv。
3. **版權資產絕不入 repo**:遊戲 LBX/存檔只在 `/home/anr2/moo2-private-build/`;`.gitignore` 已擋;`dist/` 已 ignore。
4. **不臆造數值**:每個係數標來源;查不到就說查不到(見 `docs/tech/component-values.md` 的 provenance 做法),不憑印象填。第一性原理查手冊/逆向。
5. 預設繁體中文;commit message 結尾 `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`。
6. 每輪 push 到 `main`;每輪盤點新文件與 worklist/audit 有無衝突,清掉過期斷言(rule 63)。
7. 機械工作(翻譯/移植/打包)可派便宜 subagent;但**還原度判斷要對原版實測**,別只信 subagent 回報(記憶 `moo2-cheap-subagent-division`)。

## 6. 怎麼接續這個 Claude session

- 記憶自動載入:`~/.claude/projects/-home-anr2-moo2/memory/MEMORY.md`(索引)。關鍵:`moo2-fidelity-20pct-not-test-green`、`moo2-game-interactive-architecture`、`openorion2-is-renderer-not-engine`、`ebiten-cht-reference-paths`。
- 若要在**別台機器**接續同一對話(`claude -r`),用 `dev-setup-bundle` skill 打包(含 `claude-session/`)。同機重啟不需要,直接讀本檔即可。
- 最近進度:`git log --oneline -15`。HEAD = `40216fc`(誠實現況評估)。
