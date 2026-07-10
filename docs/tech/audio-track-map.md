# MOO2 音訊曲目/音效對應表

> 目的:把 `STREAM.LBX` / `STREAMHD.LBX`(音樂)與 `SOUND.LBX`(音效)的 entry,對應到「這是哪首曲/哪個 UI 事件」,供遊戲各場景選對背景音樂與音效。
> 日期:2026-07-10。搭配讀:[`audio-format.md`](audio-format.md)(格式)。
> **驗收原則(鐵律)**:曲名↔entry 的最終對應**必須對原版聆聽確認**。本檔嚴格區分「已驗證(指紋比對)」與「推定(待聆聽)」,不把推理當已對齊(記取專案 `rulebook/65` 教訓)。

## 一、已驗證:STREAM.LBX 的 8 條 = 官方原聲帶(byte-size 指紋)

archive.org 的《MOO2: Battle at Antares OST》(Laura Barratt,自遊戲檔抽 22kHz WAV)其 8 個檔名格式為 `cut<LBX偏移>-<byte大小>`。**其 byte 大小逐一等於我方 dump 的 STREAM 條目**,連 LBX offset 遞增順序都吻合 → 確認 STREAM.LBX 就是官方原聲帶本體。

| 我方 dump | bytes | 時長 | archive.org OST 檔(`cut<off>-<size>`) | 對上 |
|---|---|---|---|---|
| stream_01 | 5012208 | 113.65s | cut**2070**-**5012208** | ✓ |
| stream_02 | 5527010 | 125.33s | cut5014278-**5527010** | ✓ |
| stream_03 | 1887988 | 42.81s | cut10541288-**1887988** | ✓ |
| stream_04 | 4910256 | 111.34s | cut12429276-**4910256** | ✓ |
| stream_05 | 5994270 | 135.92s | cut17339532-**5994270** | ✓ |
| stream_06 | 2466700 | 55.93s | cut23333802-**2466700** | ✓ |
| stream_08 | 2583936 | 58.59s | cut25800524-**2583936** | ✓ |
| stream_10 | 846862 | 19.20s | cut28384482-**846862** | ✓ |

(entry 0 = 22-byte 彩蛋槽 `cats rule dogs drool`;07/09 為非 WAV 空槽,故序號跳號。)

**已驗證**:STREAM 這 8 條=玩家熟知的 MOO2 原聲帶長版。**尚未驗證**:每條的曲名(archive.org metadata 無標題)。

## 二、曲名池(last.fm,順序未證)

last.fm 的曲目清單(依播放數排序,非必為 LBX 順序):

```
Theme 1 · Theme 2 · Theme 3            ← 一般銀河/選單主題
Psilon · Meklar · Gnolam · Darlok · Trilarian · …  ← 各種族外交主題
Battle 1 · Battle 2 · Battle 3          ← 戰鬥
```

3 主題 + ~14 種族 + 3 戰鬥 ≈ **20 條 → 正好等於 `STREAMHD.LBX` 的 20 條**。故推定:
- **`STREAMHD.LBX`(20 條,較短 13–42s)= 完整具名音樂集**(含每個種族的外交主題)。
- **`STREAM.LBX`(8 條,較長 42–136s)= 精選/長版**(官方 OST 收錄的就是這組)。

## 三、推定對應(★待聆聽確認,勿當定案)

假設 `STREAMHD.LBX` 的 entry 順序 = last.fm 曲目順序(**未證**):

| STREAMHD entry | 時長 | 推定曲名 | 用於場景(推定) |
|---|---|---|---|
| streamhd_01 | 38.54s | Theme 1 | **主選單/標題** |
| streamhd_02 | 19.17s | Theme 2 | 星系圖/一般 |
| streamhd_03 | 42.66s | Theme 3 | 一般 |
| streamhd_04..17 | 19–24s | 各種族主題(Psilon/Meklar/…) | 對應種族外交畫面 |
| streamhd_18..20 | 15–21s | Battle 1/2/3 | 戰鬥 |

> ⚠ 上表**每一列都是待驗證假設**。streamhd_01 是否真為「Theme 1 主選單曲」、各種族順序是否吻合,**都要對原版聆聽或逐條試聽 dump 檔才能定案**。目前程式暫以 streamhd_01 當主選單 BGM(見 `cmd/moo2/audiohook.go`),正確與否待確認。

## 四、SOUND.LBX 音效(名稱已知,語意用途待確認)

entry0 名稱表已解出 68 個具名音效(見 dump 清單)。UI 相關候選:

| 名稱 | entry | 時長 | 推定用途 |
|---|---|---|---|
| BUTTON1 | 34 | 0.39s | 一般按鈕點擊(**目前接線用此**) |
| BUTTON2 | 35 | 0.88s | 按鈕變體 |
| BUTTON4 | 37 | 0.19s | 短促點擊 |
| BUTTON9 | 51 | 0.39s | 按鈕變體 |
| BUTTONA/B/E | 39/40/41 | 0.10–0.27s | 極短 UI 回饋 |
| SCRENMEC/SCRNMEC3 | 48/49 | — | 畫面切換機械音 |
| XPORTIN/XPORTOUT | 55/56 | ~3s | 傳送進/出 |

其餘(NRG*/EXPL-*/PHOTON/TORP*/MONSTR*/KABOOM…)為武器/爆炸/怪物音,屬戰鬥期,待戰鬥系統接線時對應。

## 五、如何驗收(給使用者/後續)

1. 抽出全部音檔到可試聽目錄:
   ```bash
   docker run --rm -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" \
     -v /home/anr2/moo2-private-build/gamedata/mastori2:/data:ro -w /src moo2-ebiten \
     bash -c 'go build -buildvcs=false -o /tmp/m ./cmd/moo2 && xvfb-run -a /tmp/m -audiodump /out -data /data'
   # 或桌面直接跑 AppImage 後用 -audiodump 到家目錄
   ```
2. 用播放器聽 `streamhd_01..20` 與 `stream_01..10`,把「哪條是主選單曲、哪條是戰鬥曲、各種族主題」回填第三節,把推定改成已驗證。
3. 聽 `sound_034_BUTTON1` 等,確認 UI 點擊音選得對。

## 待辦

- [ ] 對原版/dump 聆聽,定案 STREAMHD 20 條的曲名與場景。
- [ ] 定案 STREAM 8 長版各自曲名(可與 STREAMHD 同名長版對照)。
- [ ] 決定遊戲各場景用 STREAM(長版)或 STREAMHD(完整集):主選單、星系圖、各種族外交、戰鬥。
- [ ] SOUND 各 BUTTONx 的實際 UI 用途區分。
