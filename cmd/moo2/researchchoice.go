package main

// researchchoice.go:MOO2 招牌「每主題數科技間抉擇」的玩家 UI。
// 回合結束後,若剛完成的研究主題有多個可選科技(PendingResearchChoice),進此畫面選一項解鎖。
// 資料為真值(gamedata.researchChoices);科技名經 TECHNAME/tech.tsv 中文化。
// 版面合成近似,尚未對原版 SCIENCE.LBX 像素對齊。

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

type researchChoiceScreen struct {
	b       *sceneBuilder
	fnt     *uifont.Font
	bg      *ebiten.Image
	cat     *i18n.Catalog
	topic   gamedata.ResearchTopic
	choices []gamedata.Technology
	hover   int
	next    func() (*overlayScreen, error) // 選定後續往的畫面(通常回合摘要)
}

// researchChoice 建抉擇畫面;讀 session 的待決抉擇。onDone 為選定後續往的畫面。
func (b *sceneBuilder) researchChoice(onDone func() (*overlayScreen, error)) (origScreen, error) {
	topic, choices, _ := b.session.PendingResearchChoice()
	s := &researchChoiceScreen{b: b, fnt: b.fnt, topic: topic, choices: choices, hover: -1, next: onDone}
	if im, err := decodeAsset(b.res, "raceopt.lbx", 0); err == nil && im.Embedded != nil {
		s.bg = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
	}
	s.cat = i18n.New(b.lang)
	if f, err := os.Open("assets/i18n/tech.tsv"); err == nil {
		defer f.Close()
		_, _ = s.cat.LoadTSV(f)
	}
	return s, nil
}

// techZh 回傳科技的中文名(查無回英文名)。
func (s *researchChoiceScreen) techZh(t gamedata.Technology) string {
	en := gamedata.TechnologyName(t)
	if s.cat != nil {
		return s.cat.Translate(en)
	}
	return en
}

func (s *researchChoiceScreen) rowRect(i int) (int, int, int, int) {
	return 140, 150 + i*56, 360, 46
}

func (s *researchChoiceScreen) update(in shell.InputState) *origTransition {
	s.hover = -1
	for i := range s.choices {
		if x, y, w, h := s.rowRect(i); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hover = i
		}
	}
	if !in.ClickReleased || s.hover < 0 {
		return nil
	}
	// 選定該科技解鎖 → 繼續往回合摘要。
	s.b.session.ChooseResearchTech(s.choices[s.hover])
	if clickSound != nil {
		clickSound()
	}
	if s.next != nil {
		return s.b.goTo(s.next, "回合摘要")
	}
	return s.b.goTo(s.b.galaxy, "星系主畫面")
}

func (s *researchChoiceScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		dst.DrawImage(s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}
	dim := color.RGBA{150, 160, 180, 255}

	s.fnt.DrawCentered(dst, "研究突破:選擇要解鎖的科技", 320, 70, 18, gold)
	topicName := topicNameZh(s.b.lang, s.topic)
	s.fnt.DrawCentered(dst, "主題:"+topicName+"(僅能擇一,其餘放棄)", 320, 104, 12, dim)

	for i, t := range s.choices {
		x, y, w, h := s.rowRect(i)
		bgc := color.RGBA{26, 34, 50, 200}
		bord := color.RGBA{90, 120, 170, 255}
		if i == s.hover {
			bgc = color.RGBA{44, 60, 90, 230}
			bord = gold
		}
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), bgc, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, bord, false)
		s.fnt.DrawCentered(dst, s.techZh(t), float64(x+w/2), float64(y+h/2), 16, body)
	}
}
