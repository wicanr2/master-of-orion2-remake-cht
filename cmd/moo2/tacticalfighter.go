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
// remake 的戰術戰鬥先前只有艦艇本體；戰機只以「母艦戰力 +N」的形式存在於快速結算裡。
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
//   - 原版一般敵方的完整逐艦藍圖未全部取回；shell 以固定的戰力→戰機庫表建立敵方
//     艦艇，這層完整消費該表。未知的是原版敵艦「哪一艘」帶哪個槽位，不是戰機中隊
//     的移動、射擊、PD、受傷、返航等規則。

// launchRect 是「出擊」鈕(選中的我方艦帶戰機庫時才畫)。
//
// ⚠ 不是原版版面:原版的戰機出擊在艦艇指令列裡,而 remake 的控制列是烘死的美術,
// 那幾顆鈕(AUTO/SCAN/BOARD/RETREAT/WAIT/DONE/OPTIONS)各有其原本的意思,
// 不能拿其中一顆假裝是出擊。先擺在標題列右側的空白處(格線本身佔到 x=600,右邊放不下),
// 原版指令列做出來之後要搬過去。
// launchRect 把 remake 的顯式出擊轉接放進 COMBAT 控制甲板的 SPECIALS 區，不再浮在
// 原版太空戰場右上角。原版逐項 special 選單尚未同構；這個 62×14 熱區明示為可玩 adapter。
func launchRect() (x, y, w, h int) { return 198, 458, 62, 14 }

// canLaunchFrom 回傳選中的我方艦這一刻能不能派中隊出擊。
//
// 條件:①有戰機庫 ②這艘船的中隊還沒在場上(手冊:一個戰機庫帶**一個**中隊,
// 返航補給後才能再出擊 —— 所以場上同時只會有一隊)。
func (t *tacticalScreen) canLaunchFrom(idx int) bool {
	if idx < 0 || idx >= len(t.player) || !t.player[idx].Bay {
		return false
	}
	for _, kind := range combatShipBays(t.player[idx]) {
		active := false
		for i := range t.squads {
			if !t.squads[i].Enemy && t.squads[i].Carrier == idx && t.squads[i].Kind == kind && !t.squads[i].Dead() {
				active = true
				break
			}
		}
		if !active {
			return true
		}
	}
	return false
}

func combatShipBays(s shell.CombatShip) []shell.FighterKind {
	if len(s.Bays) > 0 {
		return s.Bays
	}
	if s.Bay {
		return []shell.FighterKind{s.BayKind}
	}
	return nil
}

// launchFrom 讓第 idx 艘我方艦派出一隊戰機。
func (t *tacticalScreen) launchFrom(idx int) {
	s := t.player[idx]
	kind, found := s.BayKind, false
	for _, candidate := range combatShipBays(s) {
		active := false
		for i := range t.squads {
			if !t.squads[i].Enemy && t.squads[i].Carrier == idx && t.squads[i].Kind == candidate && !t.squads[i].Dead() {
				active = true
				break
			}
		}
		if !active {
			kind, found = candidate, true
			break
		}
	}
	if !found {
		return
	}
	// ⚠ 2026-08-08(第 69 項(戰鬥速度與引擎階)):上一版寫著「remake 的艦艇設計還沒有把『目前最佳引擎/裝甲』
	// 餵進戰鬥層,先用最保守的 1 / 0…等那兩項接上來,這裡換成真值即可」——**接上來了**。
	// 那個硬編的 1 讓所有戰機不論科技多高都跑得一樣慢,而且第 65 項(種族特性31格)的參數掃描器
	// (只掃 gamedata.X(...))看不到它,因為它在 cmd/ 這一側。
	squadron := shell.NewFighterSquadron(
		kind, false, idx, s.Col, s.Row, s.DriveLevel, s.ArmorLevelAboveTitanium)
	squadron.CarrierName = s.Name
	squadron.FighterRacialDefenseBonus = s.FighterRacialDefenseBonus
	squadron.FighterPilotBonus = s.FighterPilotBonus
	squadron.FighterHelmsmanBonus = s.FighterHelmsmanBonus
	t.squads = append(t.squads, squadron)
	t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.fighter.log.launch"),
		s.Name, fighterKindLabel(t.b.lang, kind), t.squads[len(t.squads)-1].Alive)
}

// launchEnemySquadrons 依 StartCombat 已建立的敵艦戰機庫開場出擊。
// 敵方沒有玩家的「點擊出擊」階段，故戰鬥開始時便把可用中隊放上棋盤；之後仍走
// 完整的移動、接戰、PD、彈藥耗盡與返航鏈。
func (t *tacticalScreen) launchEnemySquadrons() {
	for i := range t.enemy {
		s := t.enemy[i]
		if !s.Bay {
			continue
		}
		for _, kind := range combatShipBays(s) {
			f := shell.NewFighterSquadron(kind, true, i, s.Col, s.Row,
				s.DriveLevel, s.ArmorLevelAboveTitanium)
			f.CarrierName = s.Name
			f.FighterRacialDefenseBonus = s.FighterRacialDefenseBonus
			f.FighterPilotBonus = s.FighterPilotBonus
			f.FighterHelmsmanBonus = s.FighterHelmsmanBonus
			t.squads = append(t.squads, f)
		}
	}
}

// nearestEnemyShip 回傳離 (col,row) 最近的敵艦索引;沒有敵艦回 −1。
func (t *tacticalScreen) nearestEnemyShip(col, row int) int {
	best, bestD := -1, 1<<30
	for i, e := range t.enemy {
		if e.HP <= 0 {
			continue
		}
		d := abs(e.Col-col) + abs(e.Row-row)
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

// fighterTarget 先維持仍有效的主要目標；只有它不存在或已失效才重選。
//
// 手冊 p.157：「If their primary target is no longer valid while they still have
// shots available, fighters select a new target automatically.」原版的戰鬥目標
// 評估也會排除戰機記錄，這裡因此只在目前尚未建模的敵方戰機之外搜尋敵艦。
// 以名稱保存而不是陣列索引，避免敵艦被壓縮後把中隊錯接到另一艘船。
func (t *tacticalScreen) fighterTarget(f *shell.FighterSquadron) int {
	if f == nil {
		return -1
	}
	targets := t.enemy
	if f.Enemy {
		targets = t.player
	}
	if f.TargetName != "" {
		for i := range targets {
			if targets[i].Name == f.TargetName && targets[i].HP > 0 {
				return i
			}
		}
	}
	target := t.nearestEnemyShip(f.Col, f.Row)
	if f.Enemy {
		target = t.nearestPlayerShip(f.Col, f.Row)
	}
	if target >= 0 && targets[target].Name != "" {
		f.TargetName = targets[target].Name
	}
	return target
}

func (t *tacticalScreen) nearestPlayerShip(col, row int) int {
	best, bestD := -1, 1<<30
	for i, p := range t.player {
		if p.HP <= 0 {
			continue
		}
		d := abs(p.Col-col) + abs(p.Row-row)
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

func (t *tacticalScreen) carrierForSquadron(f *shell.FighterSquadron) *shell.CombatShip {
	ships := t.player
	if f.Enemy {
		ships = t.enemy
	}
	if f.CarrierName != "" {
		for i := range ships {
			if ships[i].Name == f.CarrierName {
				f.Carrier = i
				return &ships[i]
			}
		}
		f.Carrier = -1
		return nil
	}
	if f.Carrier >= 0 && f.Carrier < len(ships) {
		return &ships[f.Carrier]
	}
	return nil
}

func (t *tacticalScreen) refreshSquadronCarriers() {
	for i := range t.squads {
		if t.squads[i].CarrierName == "" {
			continue
		}
		t.carrierForSquadron(&t.squads[i])
	}
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
			carrier := t.carrierForSquadron(f)
			if carrier == nil {
				continue
			}
			if f.StepToward(carrier.Col, carrier.Row) {
				f.Recover()
				t.log = uiText(t.b.lang, "tactical.fighter.log.recovered")
			}
			continue
		}
		target := t.fighterTarget(f)
		if target < 0 {
			continue
		}
		targetShip := &t.enemy[target]
		if f.Enemy {
			targetShip = &t.player[target]
		}
		if !f.StepToward(targetShip.Col, targetShip.Row) {
			continue // 這回合只飛得到一半
		}
		// 手冊 p.117:近迫防禦武器若本回合尚未開火,
		// 必須在飛彈命中艦艇或戰機接戰前先開火。戰機不是飛彈彈頭,
		// 所以走戰機 Beam Defense 公式與 FighterSquadron.TakeHit，
		// 不把這次消費塞進 Missile_Dcv 的彈頭餘數鏈。
		// 自動 PD 依 typed 槽序處理，且刻意不讀 WeaponModes：原版說明明定
		// 紅色關閉的 PD 仍會在飛彈／戰機接觸前開火。
		for _, mount := range shell.AvailableTacticalPointDefenseMounts(targetShip) {
			shell.MarkTacticalPointDefenseMountSpent(targetShip, mount.Slot)
			for n := 0; n < mount.Count && !f.Dead(); n++ {
				beamRoll := 1
				if t.rng != nil {
					beamRoll = t.rng.Intn(100) + 1
				}
				// p.157 的完整式為 5*Speed + 種族 Ship Defense
				// + Fighter Pilot + Helmsman。種族與 Fighter Pilot 已由
				// StartCombat 依參戰艦隊證據帶入；Helmsman 的原版戰機呼叫端
				// 尚未證實，因此保留明示的零值。
				pd := shell.ResolvePointDefenseFighterShot(shell.PointDefenseFighterShot{
					BeamWeaponName:   mount.WeaponName,
					BeamAttack:       targetShip.Attack,
					BeamDamageMax:    mount.BeamDamageMax,
					BeamRangeSquares: 0, // 接戰前自動攔截，同格
					BeamRoll:         beamRoll,
					BeamSystems:      targetShip.BeamSystems,
					BeamMods:         mount.BeamMods,
					FighterBeamDefense: gamedata.CombatFighterBeamDefense(f.Speed,
						f.FighterRacialDefenseBonus, f.FighterPilotBonus, f.FighterHelmsmanBonus),
				})
				if pd.Fired && pd.Hit {
					f.TakeHit(pd.DamageToFighter)
				}
			}
			if f.Dead() {
				break
			}
		}
		if f.Dead() {
			continue
		}
		// 突擊艇(第 80 項(登艦戰)):抵達目標時**放下陸戰隊**,不開火。
		// 手冊:「Once launched, Assault Shuttles fly to the target ship and drop off their
		// Marines, which board and attempt capture. After the marines are dropped,
		// unpiloted shuttles are set adrift to be picked up after the battle.」
		// ——所以放完人這一隊就結束了(Alive 歸零),不是返航補給再來一次。
		if f.Kind == shell.FighterAssaultShuttle && !f.Enemy {
			t.resolveShuttleBoarding(f, targetShip)
			continue
		}
		// 貼身開火:一輪 = 隊裡每架各開一次火。這裡不是再用母艦武器
		// 的固定 3/5 代理，而是依 raw 函式分流：一般戰機走
		// sub_3AD57 @ 0x3AD57 的攻防差／40 門檻，轟炸機走相鄰的
		// sub_3AC20 @ 0x3AC20 直接傷害插值；兩者最後都送入最弱護盾面
		// → 裝甲 → 結構。原版每架都有自己的骰，不能把整隊壓成一發。
		carrier := t.carrierForSquadron(f)
		attack := 0
		if carrier != nil {
			attack = carrier.Attack
		}
		rng, ok := gamedata.FighterDamageRangeForKind(int(f.Kind))
		if !ok {
			// 突擊艇在抵達時走登艦路徑；未知型別 fail-closed，仍消耗
			// 本輪，避免零值戰機變成無限射擊的幽靈。
			f.SpendShot()
			continue
		}
		dmg := 0
		for craft := 0; craft < f.Alive; craft++ {
			roll := 100 // headless 測試沒有 RNG 時維持可觀測的命中路徑
			if t.rng != nil {
				roll = t.rng.Intn(100) + 1
			}
			shotDamage := 0
			if f.Kind == shell.FighterBomber {
				// IDA 的 weapon ID 0x1E 是 Bomber Bays，呼叫端在
				// sub_3D2DF 的 0x3D825 直接跳到 sub_3AC20；這裡只把
				// 已證實的公式接到 remake Bomber profile，不替外部符號
				// 索引解決名稱衝突。
				bomb := shell.ResolveFighterBomb(shell.FighterBombInput{
					DamageMin: rng.Min, DamageMax: rng.Max, Roll: roll,
				})
				shotDamage = bomb.Damage
			} else {
				shot := shell.ResolveFighterAttack(shell.FighterAttackInput{
					Attack: attack, Defense: targetShip.Defense,
					DamageMin: rng.Min, DamageMax: rng.Max, Roll: roll,
				})
				if !shot.Hit {
					continue
				}
				shotDamage = shot.Damage
			}
			structure, _ := shell.FighterDamageAtWeakestShield(targetShip, 1, shotDamage)
			if targetShip.ArmorHP > 0 {
				absorbed := structure
				if absorbed > targetShip.ArmorHP {
					absorbed = targetShip.ArmorHP
				}
				targetShip.ArmorHP -= absorbed
				structure -= absorbed
			}
			targetShip.HP -= structure
			dmg += structure
			if targetShip.HP <= 0 {
				break
			}
		}
		if dmg > 0 {
			t.spawnCombatFX(combatFXImpact, *targetShip)
		}
		if targetShip.HP <= 0 {
			t.spawnCombatFX(combatFXExplosion, *targetShip)
		}
		if !f.Enemy {
			damage += dmg
		}
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
		shooters := t.enemy
		if f.Enemy {
			shooters = t.player
		}
		for j := range shooters {
			e := shooters[j]
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
// 畫在隱形規則格位的**右下角**,不佔艦艇 sprite 的位置——同一位置可能同時有艦與中隊
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
		mark := uiText(t.b.lang, "tactical.fighter.glyph.interceptor")
		if f.Kind == shell.FighterHeavy {
			mark = uiText(t.b.lang, "tactical.fighter.glyph.heavy")
		}
		if f.Returning {
			mark = uiText(t.b.lang, "tactical.fighter.glyph.returning")
		}
		textSafeRect{x: int(x), y: int(y), w: 24, h: 16, insetX: 1, insetY: 1}.drawCentered(
			dst, t.fnt, fmt.Sprintf("%s%d", mark, f.Alive), 10, col)
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
	drawHoverBorder(dst, float32(x), float32(y), float32(w), float32(h), pointInRect(t.hoverX, t.hoverY, x, y, w, h))
	textSafeRect{x: x, y: y, w: w, h: h, insetX: 3, insetY: 1}.drawCentered(dst, t.fnt,
		uiText(t.b.lang, "tactical.fighter.button.launch"), 9, color.RGBA{225, 245, 250, 255})
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
	return fmt.Sprintf(uiText(t.b.lang, "tactical.fighter.status.active"), n, craft)
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
		StrengthBonus:    e.SecurityBonus,
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
		t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.fighter.log.boarding_success"), combatShipLabel(t.b.lang, t.b.session, e.Name))
		return
	}
	t.log = fmt.Sprintf(uiText(t.b.lang, "tactical.fighter.log.boarding_failed"), combatShipLabel(t.b.lang, t.b.session, e.Name), res.DefenderSurvived)
}
