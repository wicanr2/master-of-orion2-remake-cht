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

// initAudio 建立 Mixer、載入主選單背景音樂與按鈕音效,回傳 Mixer(需被持有以免 GC)。
// 任何一步失敗都不致命:音訊是加分項,絕不擋遊戲執行。
func initAudio(res *assets.Resolver) *moo2audio.Mixer {
	m := moo2audio.NewMixer(moo2SampleRate)

	// 背景音樂:STREAMHD.LBX(Win95 版採用的較完整音樂)。
	// TODO(task 4):clips[0] 只是「第一條可用曲」,主選單主題的正確 entry 待對原版聆聽定案。
	if arch, err := res.OpenLBX("streamhd.lbx"); err == nil {
		if clips, _, err := moo2audio.LoadMusic(arch); err == nil && len(clips) > 0 {
			if err := m.PlayBGM(clips[0]); err != nil {
				fmt.Println("音樂播放失敗(略過):", err)
			}
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
