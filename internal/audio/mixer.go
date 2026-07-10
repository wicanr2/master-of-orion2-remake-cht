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

	bgm    *ebiaudio.Player
	sfx    map[string]*ebiaudio.Player
	bgmVol float64
	sfxVol float64
}

// NewMixer 以指定取樣率建立 audio context(MOO2 音訊統一 22050 Hz)。
// 一個行程只能有一個 context,故整個遊戲共用一個 Mixer。
func NewMixer(sampleRate int) *Mixer {
	return &Mixer{
		ctx:    ebiaudio.NewContext(sampleRate),
		rate:   sampleRate,
		sfx:    make(map[string]*ebiaudio.Player),
		bgmVol: 0.6,
		sfxVol: 0.9,
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
	if err != nil {
		return fmt.Errorf("audio: 建立 BGM player: %w", err)
	}
	p.SetVolume(m.bgmVol)
	m.bgm = p
	p.Play()
	return nil
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

// SetVolumes 調整音樂/音效音量(0..1),即時套用到目前 BGM。
func (m *Mixer) SetVolumes(bgm, sfx float64) {
	m.bgmVol, m.sfxVol = bgm, sfx
	if m.bgm != nil {
		m.bgm.SetVolume(bgm)
	}
	for _, p := range m.sfx {
		p.SetVolume(sfx)
	}
}
