package main

// audiohook.go:把原版 MOO2 的音樂/音效接進互動畫面。
//
// 音訊格式見 docs/tech/audio-format.md:全部是 LBX 內的 22050 Hz PCM WAV,
// 原封播放即與原版一致。只有互動模式才初始化(headless 無音效卡,略過)。

import (
	"fmt"
	"math/rand"

	moo2audio "github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
)

// clickSound 由 overlayScreen.update 在按鈕命中時呼叫;未初始化(headless)則為 nil。
var clickSound func()

// moo2SampleRate 是 MOO2 全部音訊的取樣率(見格式研究文件)。
const moo2SampleRate = 22050

// 場景 BGM 曲目索引(STREAMHD.LBX 內序,0-based,對應 musicClips[i])。
//
// 2026-07-10 第二輪定案:openorion2 完全未實作音樂(零 provenance 可查,見
// docs/tech/audio-track-map.md 附錄);改用三路資料交叉推定:
//   1. Orion2.exe(DOS 版,Watcom 除錯字串未 strip)反映真實程式架構:
//      Play_Background_Music_ / Play_Combat_Music_ 是「各自獨立」的函式(戰鬥音樂
//      非背景樂延續),Start_Diplomacy_Music_ 搭配 _diplomacy_good_music /
//      _diplomacy_bad_music 兩個變數(外交音樂依「該族當下關係好壞」切換,非固定單曲)。
//   2. 官方 2016 重製版原聲帶(Steam App 468020,Dave Govett 掛名同一作曲家)曲名
//      實證 "XXX Race-Peace" / "XXX Race-War" 命名法——證實「每族一對和平/戰爭曲」
//      是跨代設計慣例(佐證用,非本檔案 byte 級證據)。
//   3. 本機 STREAMHD.LBX 20 條實測時長(entry 1-20 對應 musicClips[0..19]):
//      發現多組近乎相同時長的配對(如 idx2/idx7 皆 42.66s、idx12/idx17 皆 21.32s、
//      idx0/idx13 為 38.54/38.41s),與 (1)(2) 的「和平/戰爭配對」結構吻合,佐證
//      STREAMHD 內確有配對存在——但**無法**由時長反推「哪一族/哪個配對」,此為本輪
//      duration-clustering 的已知上限(自產訊號,非外部 oracle,見 rulebook/65)。
//   結論:menu 維持原推定(有時長輪廓佐證);galaxy 由「短曲」改選「另一條
//   獨立長曲」(較合理的長迴圈背景樂);combat 由「與配對曲同長」改選「無配對的
//   短曲」(較符合 Play_Combat_Music_ 獨立分派的假設)。**這三項皆非曲名級別確證**,
//   曲目↔場景/種族的精確身分待人耳聆聽定案。見 docs/tech/audio-track-map.md 完整表。
//
// 2026-07-10 第三輪(反組譯呼叫點硬證,見 docs/tech/audio-track-map.md 第七節):
// 反組譯 Orion2.exe 找到 `_diplomacy_bad_music`(外交關係差時要播的曲)的實際賦值碼
// (object1+0x90ad..0x90c0):
//
//	mov eax, 3
//	call Get_Random   ; 0x111b10,已驗證是標準 rand()-style LCG(乘數 0x41C64E6D、
//	                   ; 加數 0x3039,經典 minstd 常數),回傳 [0,N-1] 均勻亂數
//	add  eax, 13
//	mov  _diplomacy_bad_music, ax   ; = Get_Random(3) + 13 → track 13/14/15 三選一
//
// 這是本輪唯一取得的「呼叫點引數值」硬證(信心:高,反組譯無歧義)。
// `_diplomacy_good_music`(關係好時播的曲)同一函式內證實是「該族 empire 記錄
// offset 0x25 欄位 + 1」——逐族資料驅動,不是單一常數,本輪未追出該欄位的
// 逐族靜態表,故無法給出各族實際 track index。
// `Play_Background_Music_`/`Play_Combat_Music_`(menu/galaxy/combat 用哪個函式
// 播)在全檔案(含 fixup 表)找不到任何呼叫點或位址參照——判斷為這個 build 裡的
// 死碼,menu/galaxy/combat 三項無法升級,維持第二輪時長啟發式,信心不變。
const (
	bgmMenu   = 0  // 中信心(第二輪,時長啟發式,未變)。entry1,38.54s
	bgmGalaxy = 2  // 低信心(第二輪,時長啟發式,未變)。entry3,42.66s
	bgmCombat = 16 // 低信心(第二輪,時長啟發式,未變)。entry17,14.61s
)

// bgmDiploBadPool 是外交「關係差」音樂的原版候選集合——反組譯實證
// `_diplomacy_bad_music = Get_Random(3) + 13`(見上方第三輪筆記),原版每次觸發
// 關係轉壞時都重新擲骰,三條均等機率。
//
// 本檔(audiohook.go)看不到「目前跟該族關係好壞」這個遊戲狀態(那個狀態在
// interactive.go 的外交場景,本輪任務邊界只能改本檔+docs,不動 interactive.go),
// 所以無法做到「每次進外交畫面都重擲」的原版行為;改用「每次進程啟動時擲一次」
// 當作忠實度較高的近似(至少三個候選都是硬證值,不再是 duration-heuristic 猜的
// entry4)。「好關係」音樂因逐族資料未追出,暫不實作,沿用同一個 bgmDiplo 值。
var bgmDiploBadPool = [3]int{13, 14, 15}

// bgmDiplo 是外交場景播放的曲目索引;由 initAudio 在啟動時從 bgmDiploBadPool
// 隨機擲定(見上方說明),取代第二輪的 duration-heuristic 常數 3。
var bgmDiplo = bgmDiploBadPool[0]

var (
	theMixer   *moo2audio.Mixer
	musicClips []*moo2audio.Clip
	curBGM     = -1
)

// playSceneBGM 切換背景音樂到指定曲目索引(headless / 未載入音樂時為 no-op)。
// 同曲重播則略過,避免每次進同場景就從頭。
func playSceneBGM(i int) {
	if theMixer == nil || i < 0 || i >= len(musicClips) || i == curBGM {
		return
	}
	if err := theMixer.PlayBGM(musicClips[i]); err == nil {
		curBGM = i
	}
}

// initAudio 建立 Mixer、載入全部背景音樂與按鈕音效,回傳 Mixer(需被持有以免 GC)。
// 任何一步失敗都不致命:音訊是加分項,絕不擋遊戲執行。
func initAudio(res *assets.Resolver) *moo2audio.Mixer {
	m := moo2audio.NewMixer(moo2SampleRate)
	theMixer = m

	// 依硬證公式擲一次「外交關係差」曲目(見上方第三輪筆記);每次啟動重擲,
	// 不是原版的「每次進外交畫面重擲」,是本檔案邊界內能做到的最接近近似。
	bgmDiplo = bgmDiploBadPool[rand.Intn(len(bgmDiploBadPool))]

	// 背景音樂:STREAMHD.LBX 全部曲目(供各場景切換;曲目↔場景對應見常數註記)。
	if arch, err := res.OpenLBX("streamhd.lbx"); err == nil {
		if clips, _, err := moo2audio.LoadMusic(arch); err == nil && len(clips) > 0 {
			musicClips = clips
			playSceneBGM(bgmMenu) // 開場播主選單曲
		} else if err != nil {
			fmt.Println("音樂載入失敗(略過):", err)
		}
	}

	// 按鈕音效:SOUND.LBX 的具名音效 BUTTON1。
	if arch, err := res.OpenLBX("sound.lbx"); err == nil {
		if sb, err := moo2audio.LoadSoundBank(arch); err == nil {
			if c := sb.Clip("BUTTON1"); c != nil {
				m.RegisterSFX("BUTTON1", c)
				clickSound = func() { m.PlaySFX("BUTTON1") }
			}
		} else {
			fmt.Println("音效載入失敗(略過):", err)
		}
	}

	return m
}
