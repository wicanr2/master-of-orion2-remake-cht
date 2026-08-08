package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// tacticalfighter.go:格子戰場上的**戰機中隊**(規則面在 internal/shell/fighter.go)。
//
// remake 的戰術戰鬥先前只有艦艇 token;戰機只以「母艦戰力 +N」的形式存在於快速結算裡。
// 手冊給了戰機一整套與艦艇不同的規則(飛到目標身上打、打完返航、只有攔截機能纏鬥…),
// 那些規則要能被看見,就得讓中隊在格子上有自己的位置。
//
// ============ 這一層負責什麼 ============
//
// shell 那一檔是純狀態機(移動一步、吃一發、消耗一輪、返航補給),不知道戰場上有誰。
// 這一檔負責每回合替中隊挑目標、推進、結算,以及把它們畫出來。
//
// ⚠ 誠實留白(同 shell 檔頭那三條之外的):
//   - 原版的戰機是**逐架**在格子上飛的動畫;remake 一隊一個 token,顯示剩幾架。
//     手冊說「You cannot separate the fighters in a squadron」——一隊是一個指令單位,
//     所以用一個 token 代表整隊不違反規則,只是視覺上比原版簡略。
//   - 敵方不會派戰機:`genEnemyFleet` 產出的敵艦沒有設計資料(見 CombatShip.Kind 註解的
//     同款簡化),沒有「這艘敵艦帶不帶戰機庫」可讀。不臆造一個。

// 每架戰機一次射擊的近似傷害(同 gamedata 那兩個 approx 常數的立場:
// 手冊給的是武裝**類型**不是固定數,實際傷害隨當局科技)。
const (
	fighterBeamHit = 3
	fighterBombHit = 5
)

// fighterHitPerCraft 回傳一架這型戰機一次出手打多少(重戰機是一光束 + 一炸彈)。
func fighterHitPerCraft(k shell.FighterKind) int {
	if k == shell.FighterHeavy {
		return fighterBeamHit + fighterBombHit
	}
	return fighterBeamHit
}

// launchRect 是「出擊」鈕(選中的我方艦帶戰機庫時才畫)。
//
// ⚠ 不是原版版面:原版的戰機出擊在艦艇指令列裡,而 remake 的控制列是烘死的美術,
// 那幾顆鈕(AUTO/SCAN/BOARD/RETREAT/WAIT/DONE/OPTIONS)各有其原本的意思,
// 不能拿其中一顆假裝是出擊。先擺在標題列右側的空白處(格線本身佔到 x=600,右邊放不下),
// 原版指令列做出來之後要搬過去。
func launchRect() (x, y, w, h int) { return 544, 18, 90, 26 }

// canLaunchFrom 回傳選中的我方艦這一刻能不能派中隊出擊。
//
// 條件:①有戰機庫 ②這艘船的中隊還沒在場上(手冊:一個戰機庫帶**一個**中隊,
// 返航補給後才能再出擊 —— 所以場上同時只會有一隊)。
func (t *tacticalScreen) canLaunchFrom(idx int) bool {
	if idx < 0 || idx >= len(t.player) || !t.player[idx].Bay {
		return false
	}
	for i := range t.squads {
		if !t.squads[i].Enemy && t.squads[i].Carrier == idx && !t.squads[i].Dead() {
			return false
		}
	}
	return true
}

// launchFrom 讓第 idx 艘我方艦派出一隊戰機。
func (t *tacticalScreen) launchFrom(idx int) {
	s := t.player[idx]
	// ⚠ 2026-08-08(第 69 項(戰鬥速度與引擎階)):上一版寫著「remake 的艦艇設計還沒有把『目前最佳引擎/裝甲』
	// 餵進戰鬥層,先用最保守的 1 / 0…等那兩項接上來,這裡換成真值即可」——**接上來了**。
	// 那個硬編的 1 讓所有戰機不論科技多高都跑得一樣慢,而且第 65 項(種族特性31格)的參數掃描器
	// (只掃 gamedata.X(...))看不到它,因為它在 cmd/ 這一側。
	t.squads = append(t.squads, shell.NewFighterSquadron(
		s.BayKind, false, idx, s.Col, s.Row, s.DriveLevel, s.ArmorLevelAboveTitanium))
	t.log = fmt.Sprintf(t.b.tr("%s 派出一隊%s(%d 架)", "%s launches a %s squadron (%d craft)"),
		s.Name, shell.FighterKindName(s.BayKind), t.squads[len(t.squads)-1].Alive)
}

// nearestEnemyShip 回傳離 (col,row) 最近的敵艦索引;沒有敵艦回 −1。
func (t *tacticalScreen) nearestEnemyShip(col, row int) int {
	best, bestD := -1, 1<<30
	for i, e := range t.enemy {
		d := abs(e.Col-col) + abs(e.Row-row)
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

// advanceSquadrons 推進所有中隊一回合:返航的往母艦飛(到了就補給),
// 其餘的飛向最近的敵艦,貼身就開火。
//
// 手冊 p.157:「Fighter craft fly to their target and use whatever weapons they have at
// point-blank range… fighters will attempt to return to their carrier once they are out of
// shots. Once safely back, any surviving fighters get repairs, rearm, refuel, and can be
// launched again.」
func (t *tacticalScreen) advanceSquadrons() (damage int) {
	for i := range t.squads {
		f := &t.squads[i]
		if f.Dead() {
			continue
		}
		if f.Returning {
			// 母艦沒了就回不去了——手冊沒寫這種情況,remake 讓它留在原地繼續飄
			// (不憑空讓它消失,也不讓它變成無限彈藥的幽靈)。
			if f.Carrier < 0 || f.Carrier >= len(t.player) {
				continue
			}
			c := t.player[f.Carrier]
			if f.StepToward(c.Col, c.Row) {
				f.Recover()
				t.log = t.b.tr("戰機返航補給,可再次出擊", "Fighters recovered — ready to launch again")
			}
			continue
		}
		target := t.nearestEnemyShip(f.Col, f.Row)
		if target < 0 {
			continue
		}
		e := &t.enemy[target]
		if !f.StepToward(e.Col, e.Row) {
			continue // 這回合只飛得到一半
		}
		// 突擊艇(第 80 項(登艦戰)):抵達目標時**放下陸戰隊**,不開火。
		// 手冊:「Once launched, Assault Shuttles fly to the target ship and drop off their
		// Marines, which board and attempt capture. After the marines are dropped,
		// unpiloted shuttles are set adrift to be picked up after the battle.」
		// ——所以放完人這一隊就結束了(Alive 歸零),不是返航補給再來一次。
		if f.Kind == shell.FighterAssaultShuttle {
			t.resolveShuttleBoarding(f, e)
			continue
		}
		// 貼身開火:一輪 = 隊裡每架各開一次火。護盾照扣(戰機不是無視護盾的),
		// 但不走 ResolveShot 的命中判定——手冊說戰機是 point-blank 開火的,
		// 而 remake 沒有「戰機命中率」的公式可抄,不臆造一個骰。
		per := fighterHitPerCraft(f.Kind) - e.ShieldReduction
		if per < 1 {
			per = 1 // 護盾再厚也擋不到零(同艦砲的最低傷害立場)
		}
		dmg := shell.FighterSquadronDamage(f.Alive, per)
		if e.ArmorHP > 0 {
			absorbed := dmg
			if absorbed > e.ArmorHP {
				absorbed = e.ArmorHP
			}
			e.ArmorHP -= absorbed
			dmg -= absorbed
		}
		e.HP -= dmg
		damage += dmg
		f.SpendShot()
	}
	return damage
}

// enemyFiresAtSquadrons 讓貼身的敵艦對中隊開火。
//
// 手冊 p.157:「Like missiles, fighter craft are vulnerable to beam weapons」——
// 戰機**會**被打下來,不是無敵的裝飾。
func (t *tacticalScreen) enemyFiresAtSquadrons() (killed int) {
	for i := range t.squads {
		f := &t.squads[i]
		if f.Dead() {
			continue
		}
		for j := range t.enemy {
			e := t.enemy[j]
			if abs(e.Col-f.Col)+abs(e.Row-f.Row) > 1 {
				continue // 只有貼身的敵艦打得到(戰機是繞著目標飛的)
			}
			// 傷害取敵艦單發的中間值:remake 沒有「對戰機的命中判定」可抄
			// (手冊只說戰機的速度與裝甲都算進防禦,沒給公式),不臆造一個骰。
			killed += f.TakeHit((e.WeaponMin + e.WeaponMax) / 2)
			if f.Dead() {
				break
			}
		}
	}
	return killed
}

// dropDeadSquadrons 把全滅的中隊移出場。
func (t *tacticalScreen) dropDeadSquadrons() {
	alive := t.squads[:0]
	for _, f := range t.squads {
		if !f.Dead() {
			alive = append(alive, f)
		}
	}
	t.squads = alive
}

// squadColor 依型別給 token 顏色(攔截機偏青、重戰機偏橘)。
func squadColor(k shell.FighterKind) color.RGBA {
	if k == shell.FighterHeavy {
		return color.RGBA{240, 170, 90, 255}
	}
	return color.RGBA{120, 230, 240, 255}
}

// drawSquadrons 畫場上的中隊:一隊一個小 token,標型別與剩餘架數。
//
// 畫在格子的**右下角**,不佔艦艇 token 的位置——一個格子裡可能同時有艦與中隊
// (戰機是繞著目標飛的,本來就重疊)。
func (t *tacticalScreen) drawSquadrons(dst *ebiten.Image) {
	for _, f := range t.squads {
		if f.Dead() {
			continue
		}
		cx, cy, cw, ch := cellRect(f.Col, f.Row)
		x, y := float32(cx+cw-26), float32(cy+ch-20)
		col := squadColor(f.Kind)
		fillPanel(dst, x, y, 24, 16, color.RGBA{10, 14, 24, 220}, false)
		vector.StrokeRect(dst, x, y, 24, 16, 1, col, false)
		if t.fnt == nil {
			continue
		}
		mark := "◇" // 攔截機
		if f.Kind == shell.FighterHeavy {
			mark = "◆"
		}
		if f.Returning {
			mark = "↩"
		}
		t.fnt.DrawCentered(dst, fmt.Sprintf("%s%d", mark, f.Alive), float64(x)+12, float64(y)+12, 10, col)
	}
}

// drawLaunchButton 畫「出擊」鈕(選中的艦帶戰機庫、且它的中隊不在場上時)。
func (t *tacticalScreen) drawLaunchButton(dst *ebiten.Image) {
	if !t.canLaunchFrom(t.sel) || t.fnt == nil {
		return
	}
	x, y, w, h := launchRect()
	fillPanel(dst, float32(x), float32(y), float32(w), float32(h),
		color.RGBA{30, 60, 80, 235}, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1,
		color.RGBA{120, 220, 235, 255}, false)
	label := t.b.tr("▶ 出擊", "▶ LAUNCH")
	t.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2)+4, 12, color.RGBA{225, 245, 250, 255})
}

// squadronStatusLine 回傳控制列上方要顯示的中隊摘要(沒有中隊就回空字串)。
func (t *tacticalScreen) squadronStatusLine() string {
	if len(t.squads) == 0 {
		return ""
	}
	n, craft := 0, 0
	for _, f := range t.squads {
		if f.Dead() {
			continue
		}
		n++
		craft += f.Alive
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(t.b.tr("戰機:%d 隊 %d 架在場", "Fighters: %d squadrons, %d craft"), n, craft)
}

// resolveShuttleBoarding 是突擊艇抵達目標時的登艦結算(第 80 項(登艦戰))。
//
// 手冊:「Marines on Assault Shuttles **always try to capture** the target ship.」
// ——沒有突襲那個選項,突襲是傳送器那一側的用法。
//
// ⚠ **奪船在 remake 只表現成「那艘船退出戰鬥」**。手冊說奪船之後還要贏下整場戰鬥才留得住
// (心靈感應種族除外),真的把船搬到玩家艦隊要動戰後結算與艦隊清單。這是**建模取捨**,
// 即時效果是對的(那艘船不再開火),長期歸屬沒做。
func (t *tacticalScreen) resolveShuttleBoarding(f *shell.FighterSquadron, e *shell.CombatShip) {
	party := shell.BoardingParty{
		Intent:     shell.BoardingCapture,
		Marines:    f.Alive * gamedata.AssaultShuttleMarinesEach,
		Strength:   t.player[f.Carrier].Attack,
		HitsToKill: gamedata.GroundBaseHitsToKill(false),
	}
	def := shell.BoardingDefense{
		Marines: e.Marines, Strength: e.Defense,
		HitsToKill:       gamedata.GroundBaseHitsToKill(false),
		SecurityStations: e.SecurityStations,
	}
	res := shell.ResolveBoarding(party, def, func(n int) int {
		if n <= 0 {
			return 0
		}
		return t.rng.Intn(n)
	})
	e.Marines = res.DefenderSurvived
	f.Alive = 0 // 放完人,艇就漂在那裡了
	if res.Captured {
		e.HP = 0
		e.Captured = true
		t.log = fmt.Sprintf(t.b.tr("登艦成功:%s 的守軍全滅,該艦被奪下",
			"Boarded: %s's crew is wiped out — the ship is taken"), e.Name)
		return
	}
	t.log = fmt.Sprintf(t.b.tr("登艦失敗:%s 還剩 %d 隊守軍",
		"Boarding repelled: %s still has %d marine units"), e.Name, res.DefenderSurvived)
}
