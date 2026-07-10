package main

// audiohook.go:把原版 MOO2 的音樂/音效接進互動畫面。
//
// 音訊格式見 docs/tech/audio-format.md:全部是 LBX 內的 22050 Hz PCM WAV,
// 原封播放即與原版一致。只有互動模式才初始化(headless 無音效卡,略過)。

import (
	"fmt"

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
//   結論:menu/diplo 維持原推定(有時長輪廓佐證);galaxy 由「短曲」改選「另一條
//   獨立長曲」(較合理的長迴圈背景樂);combat 由「與配對曲同長」改選「無配對的
//   短曲」(較符合 Play_Combat_Music_ 獨立分派的假設)。**四項皆非曲名級別確證**,
//   曲目↔場景/種族的精確身分待人耳聆聽定案。見 docs/tech/audio-track-map.md 完整表。
const (
	bgmMenu   = 0  // 中信心。entry1,38.54s,屬「長曲」群,首條、契合片頭主題慣例(不變)
	bgmGalaxy = 2  // 低信心(改自 1)。entry3,42.66s,獨立長曲,較合星系圖長時間播放
	bgmDiplo  = 3  // 中信心。entry4,24.05s,落在「中長」曲群(疑似各族主題區間)(不變)
	bgmCombat = 16 // 低信心(改自 17)。entry17,14.61s,無同時長配對的短曲,較像獨立戰鬥曲
)

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
