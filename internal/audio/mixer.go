package audio

// mixer.go:ebiten 音訊裝置層。與純解碼層(wav.go/bank.go)分離——只有互動模式
// 才會建立 Mixer;headless(-shot/script)不初始化裝置,避免無音效卡環境崩潰。

import (
	"bytes"
	"fmt"

	ebiaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// Mixer 管理單一 ebiten audio context 下的背景音樂(迴圈)與音效(一次性)。
type Mixer struct {
	ctx  *ebiaudio.Context
	rate int

	bgm *ebiaudio.Player
	// bgmOnce 標記目前這首是**單次播放**(PlayBGMOnce)。迴圈播放的曲子不會結束,
	// 所以 BGMFinished 必須先看這個旗標,否則「還沒開始播」會被誤判成「播完了」。
	bgmOnce bool
	sfx     map[string]*ebiaudio.Player
	bgmVol  float64
	sfxVol  float64
}

const (
	// DefaultBGMVolume / DefaultSFXVolume 是目前 remake 的預設值。
	// 音量本身沒有存檔格式需求；Mixer 生命週期涵蓋整個互動遊戲，因此切換畫面時保留。
	DefaultBGMVolume = 0.6
	DefaultSFXVolume = 0.9
)

// NewMixer 以指定取樣率建立 audio context(MOO2 音訊統一 22050 Hz)。
// 一個行程只能有一個 context,故整個遊戲共用一個 Mixer。
func NewMixer(sampleRate int) *Mixer {
	return &Mixer{
		ctx:    ebiaudio.NewContext(sampleRate),
		rate:   sampleRate,
		sfx:    make(map[string]*ebiaudio.Player),
		bgmVol: DefaultBGMVolume,
		sfxVol: DefaultSFXVolume,
	}
}

// PlayBGM 迴圈播放一首音樂;若已有 BGM 在放,先停舊的再換新曲。
func (m *Mixer) PlayBGM(c *Clip) error {
	if c == nil || len(c.PCM) == 0 {
		return fmt.Errorf("audio: 空音樂 Clip")
	}
	if m.bgm != nil {
		m.bgm.Pause()
		m.bgm = nil
	}
	loop := ebiaudio.NewInfiniteLoop(bytes.NewReader(c.PCM), int64(len(c.PCM)))
	p, err := m.ctx.NewPlayer(loop)
	m.bgmOnce = false
	if err != nil {
		return fmt.Errorf("audio: 建立 BGM player: %w", err)
	}
	p.SetVolume(m.bgmVol)
	m.bgm = p
	p.Play()
	return nil
}

// PlayBGMOnce 播一首**不迴圈**的音樂。
//
// 用途是原版 `Play_Streaming_Music_` 的 `edx = −2` 哨兵:「這首播完接隨機 STREAM 1..3」。
// 科學室(STREAMHD #17)是唯一走這條的畫面(第 78 項(音樂接線))。
// 呼叫端要自己輪詢 BGMFinished 才知道該接下一首。
func (m *Mixer) PlayBGMOnce(c *Clip) error {
	if c == nil || len(c.PCM) == 0 {
		return fmt.Errorf("audio: 空音樂 Clip")
	}
	if m.bgm != nil {
		m.bgm.Pause()
		m.bgm = nil
	}
	p, err := m.ctx.NewPlayer(bytes.NewReader(c.PCM))
	if err != nil {
		return fmt.Errorf("audio: 建立 BGM player: %w", err)
	}
	p.SetVolume(m.bgmVol)
	m.bgm = p
	m.bgmOnce = true
	p.Play()
	return nil
}

// BGMFinished 回報「上一首是單次播放,而且已經播完了」。
//
// 迴圈播放的曲子永遠回 false —— 它不會結束。沒有 BGM 在放時也回 false
// (那是「還沒開始」不是「播完了」,呼叫端不該因此接下一首)。
func (m *Mixer) BGMFinished() bool {
	return m.bgmOnce && m.bgm != nil && !m.bgm.IsPlaying()
}

// StopBGM 停止背景音樂。
func (m *Mixer) StopBGM() {
	if m.bgm != nil {
		m.bgm.Pause()
		m.bgm = nil
	}
}

// RegisterSFX 預先建立一個可重播的音效 player(以名稱索引)。
func (m *Mixer) RegisterSFX(name string, c *Clip) {
	if c == nil || len(c.PCM) == 0 {
		return
	}
	p := m.ctx.NewPlayerFromBytes(c.PCM)
	p.SetVolume(m.sfxVol)
	m.sfx[name] = p
}

// PlaySFX 播放先前註冊的音效;正在播放則倒帶重播。未註冊則靜默略過。
func (m *Mixer) PlaySFX(name string) {
	p := m.sfx[name]
	if p == nil {
		return
	}
	_ = p.Rewind()
	p.Play()
}

// Volumes 回傳目前音樂與音效音量(0..1)。
func (m *Mixer) Volumes() (bgm, sfx float64) {
	if m == nil {
		return DefaultBGMVolume, DefaultSFXVolume
	}
	return m.bgmVol, m.sfxVol
}

// ClampVolume 將 UI 或設定值夾在音訊播放器接受的 0..1 範圍。
func ClampVolume(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// SetVolumes 調整音樂/音效音量(0..1),即時套用到目前 BGM 與已註冊音效。
func (m *Mixer) SetVolumes(bgm, sfx float64) {
	if m == nil {
		return
	}
	bgm, sfx = ClampVolume(bgm), ClampVolume(sfx)
	m.bgmVol, m.sfxVol = bgm, sfx
	if m.bgm != nil {
		m.bgm.SetVolume(bgm)
	}
	for _, p := range m.sfx {
		p.SetVolume(sfx)
	}
}
