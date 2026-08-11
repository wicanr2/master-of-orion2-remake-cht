package shell

import (
	"fmt"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// RefitBuildName 是建造佇列中改裝工作的穩定名稱。真正的艦名／目標資料在
// ColonyBuild.Refit，不把艦名拼進名稱可避免存檔、翻譯或網路指令依賴顯示字串。
const RefitBuildName = "艦艇改裝"

// enqueueBuildValue 是 EnqueueBuild 與 QueueRefit 共用的無記錄入口。所有由玩家直接
// 觸發的公開方法都必須先自行記錄 PlayerCommand；這個小入口只處理佇列不變量。
func (s *GameSession) enqueueBuildValue(i int, b ColonyBuild) bool {
	if i < 0 || i >= len(s.Builds) || b.Name == "" {
		return false
	}
	s.ensureBuildQueue()
	if s.Builds[i].Name == "" {
		s.Builds[i] = b
		return true
	}
	if len(s.BuildQueue[i]) >= buildQueueBacklogMax {
		return false
	}
	s.BuildQueue[i] = append(s.BuildQueue[i], b)
	return true
}

func (s *GameSession) canEnqueueBuild(i int) bool {
	if i < 0 || i >= len(s.Builds) {
		return false
	}
	s.ensureBuildQueue()
	return s.Builds[i].Name == "" || len(s.BuildQueue[i]) < buildQueueBacklogMax
}

// discardQueuedBuild 處理從佇列拿掉的特殊工作。原版手冊明說移除改裝工作會毀掉該艦；
// 來源艦已在 QueueRefit 時移離艦隊，因此這裡刻意不放回去。
func (s *GameSession) discardQueuedBuild(i int, b ColonyBuild) {
	if b.Refit != nil {
		s.LastBuilt = append(s.LastBuilt,
			fmt.Sprintf("殖民地 %d 取消改裝:%s 已報廢", i+1, b.Refit.Source.Name))
	}
	if i >= 0 && i < len(s.RepeatBuild) && sameRepeatBuild(s.RepeatBuild[i], b) {
		// 拿掉被指定重複的那一項就是明確取消，不讓它在之後靜默重生。
		s.RepeatBuild[i] = ColonyBuild{}
	}
}

func sameRepeatBuild(a, b ColonyBuild) bool {
	return a.Refit == nil && b.Refit == nil && a.Name != "" &&
		a.Name == b.Name && a.Cost == b.Cost
}

// ColonyAutoBuild 回傳殖民地 i 的 AUTO BUILD 開關。
func (s *GameSession) ColonyAutoBuild(i int) bool {
	if i < 0 || i >= len(s.PlayerColonies) {
		return false
	}
	s.ensureBuildQueue()
	return s.AutoBuild[i]
}

// RepeatBuildFor 回傳殖民地 i 目前指定的重複建造項。零值代表未設定。
func (s *GameSession) RepeatBuildFor(i int) ColonyBuild {
	if i < 0 || i >= len(s.PlayerColonies) {
		return ColonyBuild{}
	}
	s.ensureBuildQueue()
	return s.RepeatBuild[i]
}

// SetAutoBuild 切換 AUTO BUILD。原版已證實這是一個切換鈕，但「殖民者認為最好」的選擇
// 演算法尚未從原版安全取得；refreshAutoBuild 採固定、可預測的 remake 優先順序。
func (s *GameSession) SetAutoBuild(i int, on bool) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSetAutoBuild, Args: []int{i, boolInt(on)}})
	if i < 0 || i >= len(s.PlayerColonies) {
		return false
	}
	s.ensureBuildQueue()
	s.AutoBuild[i] = on
	if on {
		s.refreshAutoBuild(i)
	}
	return true
}

// SetRepeatBuild 指定或清除 REPEAT BUILD 的目標。remake 現有殖民地佇列裡，能有可觀
// 遊戲效果地反覆生產的是 Special（殖民船、前哨船、運輸艦隊與地形行動）；一般建築
// 本來就只能生效一次，住宅／貿易品則是持續模式，故一律拒絕當重複目標。
func (s *GameSession) SetRepeatBuild(i int, name string, cost int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSetRepeatBuild, Args: []int{i, cost}, Text: name})
	if i < 0 || i >= len(s.PlayerColonies) {
		return false
	}
	s.ensureBuildQueue()
	if name == "" {
		s.RepeatBuild[i] = ColonyBuild{}
		return true
	}
	if !repeatableBuild(name, cost) {
		return false
	}
	target := ColonyBuild{Name: name, Cost: cost}
	alreadyPresent := sameRepeatBuild(target, s.Builds[i])
	if !alreadyPresent {
		for _, q := range s.BuildQueue[i] {
			if sameRepeatBuild(target, q) {
				alreadyPresent = true
				break
			}
		}
	}
	if !alreadyPresent && !s.enqueueBuildValue(i, target) {
		return false
	}
	s.RepeatBuild[i] = target
	return true
}

func repeatableBuild(name string, cost int) bool {
	if name == "" || cost <= 0 || name == TradeGoodsBuildName || name == HousingBuildName {
		return false
	}
	_, ok := gamedata.SpecialActionByNameZH(name)
	return ok
}

// refreshAutoBuild 在自動建造啟用、且當前項目空白或不再適用時補上一個選項。
// 它不碰正在執行的一般建築與玩家已手動排在後面的項目，避免 AUTO BUILD 把玩家
// 已經明確做出的佇列決定覆寫掉。
func (s *GameSession) refreshAutoBuild(i int) {
	if i < 0 || i >= len(s.PlayerColonies) || i >= len(s.Builds) {
		return
	}
	s.ensureBuildQueue()
	if !s.AutoBuild[i] {
		return
	}
	cur := s.Builds[i]
	if cur.Name != "" {
		if cur.Name == HousingBuildName && s.PlayerColonies[i].Population >= s.PlayerColonies[i].PopMax {
			// 住宅在滿人口後不再有用；若玩家另排了項目，尊重該佇列。
			if len(s.BuildQueue[i]) > 0 {
				return
			}
		} else if cur.Name == TradeGoodsBuildName && len(s.BuildQueue[i]) == 0 {
			// 可在研究完成後從貿易品切換到新可建的設施。
		} else {
			return
		}
	}
	s.Builds[i] = s.autoBuildChoice(i)
}

// autoBuildChoice 是明示的 remake 近似：先補人口空間，再依固定的早期經濟優先項，
// 接著依資料表順序挑第一棟尚未完成的可建建築；沒有建築可蓋時才轉貿易品。
// 不把一次性地形／艦隊 Special 交給自動模式，避免它在玩家不知情時消耗行星或國庫節奏。
func (s *GameSession) autoBuildChoice(i int) ColonyBuild {
	if i < 0 || i >= len(s.PlayerColonies) {
		return ColonyBuild{}
	}
	c := s.PlayerColonies[i]
	if c.Population < c.PopMax {
		return ColonyBuild{Name: HousingBuildName}
	}
	opts := s.AvailableBuildOptions()
	priority := []string{"自動工廠", "機器人採礦廠", "研究實驗室", "太空港"}
	for _, name := range priority {
		for _, o := range opts {
			if o.Name == name && !s.buildAlreadyDone(i, o.Name) {
				return ColonyBuild{Name: o.Name, Cost: o.Cost}
			}
		}
	}
	for _, o := range opts {
		if o.Name == "" || o.Name == TradeGoodsBuildName || o.Name == HousingBuildName {
			continue
		}
		if _, isSpecial := gamedata.SpecialActionByNameZH(o.Name); isSpecial {
			continue
		}
		if !s.buildAlreadyDone(i, o.Name) {
			return ColonyBuild{Name: o.Name, Cost: o.Cost}
		}
	}
	return ColonyBuild{Name: TradeGoodsBuildName}
}

// BuildBuyCostBC 回傳 BUY 現在需要的 BC；0 代表目前項目不可買或已完成。
//
// 已證實的兩個邊界是未開工時每剩餘 PP 4 BC、完成一半後每剩餘 PP 2 BC。
// 中間連續公式尚未有可再現的原版證據，因此採能精確命中兩個邊界、且之後保持
// 2 BC/PP 的有界 remake 推導：max(2*remainingPP, 4*cost-6*progressPP)。
// 以 half-PP 整數計算以保留半機械族的半單位進度。
func (s *GameSession) BuildBuyCostBC(i int) int {
	if i < 0 || i >= len(s.Builds) {
		return 0
	}
	b := s.Builds[i]
	if b.Name == "" || b.Cost <= 0 {
		return 0
	}
	progressHalf := b.Progress*2 + b.ProgressHalf
	remainingHalf := b.Cost*2 - progressHalf
	if remainingHalf <= 0 {
		return 0
	}
	earlyCost := 4*b.Cost - 3*progressHalf
	if earlyCost < remainingHalf {
		return remainingHalf
	}
	return earlyCost
}

// CanBuyCurrentBuild 回傳玩家國庫能否買下目前建造。
func (s *GameSession) CanBuyCurrentBuild(i int) bool {
	cost := s.BuildBuyCostBC(i)
	return cost > 0 && s.Player.BC >= cost
}

// BuyCurrentBuild 扣除 BC 並把建造進度標成完成。效果不在按鈕當下套用，而是在本回合
// EndTurn 的建造結算完成，讓「買完再按回合」的流程與手冊描述一致。
func (s *GameSession) BuyCurrentBuild(i int) (int, bool) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdBuyBuild, Args: []int{i}})
	cost := s.BuildBuyCostBC(i)
	if cost <= 0 || s.Player.BC < cost {
		return 0, false
	}
	s.Player.BC -= cost
	s.Builds[i].Progress = s.Builds[i].Cost
	s.Builds[i].ProgressHalf = 0
	return cost, true
}

// RefitCandidate 是某殖民地所在星系中可改裝艦艇的穩定選取座標。
type RefitCandidate struct {
	FleetIndex int
	ShipIndex  int
	Ship       Ship
}

// RefitCandidates 回傳停泊在該殖民地星系、未航行且屬於可設計戰鬥艦體的艦艇。
func (s *GameSession) RefitCandidates(colony int) []RefitCandidate {
	star := s.PlayerColonyStarIndex(colony)
	if star < 0 {
		return nil
	}
	out := make([]RefitCandidate, 0)
	for fi := range s.Fleets {
		f := s.Fleets[fi]
		if f.AtStar != star || f.ETA != 0 || f.DestStar >= 0 {
			continue
		}
		for si, sh := range f.Ships {
			if refittableClass(sh.Class) && s.refitFacilityAllows(colony, sh.Class) {
				out = append(out, RefitCandidate{FleetIndex: fi, ShipIndex: si, Ship: sh})
			}
		}
	}
	return out
}

func refittableClass(class string) bool {
	switch class {
	case "巡防艦", "護衛艦", "驅逐艦", "巡洋艦", "戰艦", "泰坦", "末日之星":
		return true
	}
	return false
}

// refitFacilityAllows 實作手冊的容量門檻：Cruiser 或更大艦體必須在具有 Star Base
// 或更高等級軌道基地的殖民地改裝。較小的 Frigate / Destroyer 不受此限制。
func (s *GameSession) refitFacilityAllows(colony int, class string) bool {
	switch class {
	case "巡洋艦", "戰艦", "泰坦", "末日之星":
		// 總覽資料已把三種基地記為互斥的建築標記；任一存在即可。
		return colony >= 0 && colony < len(s.ColonyBuildings) &&
			(s.ColonyBuildings[colony]["星基"] ||
				s.ColonyBuildings[colony]["戰鬥站"] ||
				s.ColonyBuildings[colony]["星辰要塞"])
	}
	return true
}

func componentIndexByName(opts []Component, name string) int {
	for i, c := range opts {
		if c.Name == name {
			return i
		}
	}
	return 0
}

func (s *GameSession) bestUnlockedComponent(opts []Component) int {
	best := 0
	for i, c := range opts {
		if s.ComponentUnlocked(c) {
			best = i
		}
	}
	return best
}

// bestRefitTarget 建立同艦體的自動最佳模板。這是持久化設計庫尚未存在時的 bounded
// remake 替代：不改艦體、保留艦名／軍官／經驗／損傷，只把已解鎖元件升到可塞入艦體的
// 最高組合；原有、仍適用的武器改造與火線角會保留。
func (s *GameSession) bestRefitTarget(source Ship) (Ship, bool) {
	if !refittableClass(source.Class) {
		return Ship{}, false
	}
	weapons := BuildWeaponOptions(s.RuleProfile)
	bestWeapon := s.bestUnlockedComponent(weapons)
	bestArmor := s.bestUnlockedComponent(ArmorOptions)
	bestShield := s.bestUnlockedComponent(ShieldOptions)
	bestSpecial := s.bestUnlockedComponent(SpecialOptions)

	for wi := bestWeapon; wi >= 0; wi-- {
		if !s.ComponentUnlocked(weapons[wi]) {
			continue
		}
		for spi := bestSpecial; spi >= 0; spi-- {
			if !s.ComponentUnlocked(SpecialOptions[spi]) {
				continue
			}
			target := source
			target.Weapon = weapons[wi].Name
			target.Armor = ArmorOptions[bestArmor].Name
			target.Shield = ShieldOptions[bestShield].Name
			target.Special = SpecialOptions[spi].Name
			target.Mods = FilterWeaponModsForWeapon(target.Weapon, source.Mods)
			target.Arc = NormalizeWeaponArc(target.Weapon, source.Arc)
			target.WeaponAttack = weapons[wi].Value
			if SpecialOptions[spi].Name == "戰鬥電腦" {
				target.WeaponAttack += SpecialOptions[spi].Value
			}
			target.BonusHP = ArmorOptions[bestArmor].Value + ShieldOptions[bestShield].Value
			if !s.DesignFitsWithModsAndArc(target.Class, wi, bestArmor, bestShield, spi, target.Mods, target.Arc) {
				continue
			}
			if sameRefitEquipment(source, target) {
				return Ship{}, false
			}
			return target, true
		}
	}
	return Ship{}, false
}

func sameRefitEquipment(a, b Ship) bool {
	return a.Class == b.Class && a.Weapon == b.Weapon && a.Armor == b.Armor &&
		a.Shield == b.Shield && a.Special == b.Special && a.Arc == b.Arc &&
		strings.Join(a.Mods, "\x00") == strings.Join(b.Mods, "\x00")
}

func shipProductionCost(sh Ship) int {
	weapon := componentIndexByName(WeaponOptions, sh.Weapon)
	armor := componentIndexByName(ArmorOptions, sh.Armor)
	shield := componentIndexByName(ShieldOptions, sh.Shield)
	special := componentIndexByName(SpecialOptions, sh.Special)
	mods := FilterWeaponModsForWeapon(sh.Weapon, sh.Mods)
	arc := NormalizeWeaponArc(sh.Weapon, sh.Arc)
	return DesignCostWithModsAndArc(sh.Class, weapon, armor, shield, special, mods, arc)
}

// RefitCostPP 是手冊明示的改裝成本公式：max(2 * (新設計成本 - 舊設計成本),
// floor(標準艦體成本 / 4))。本函式的輸入設計成本使用 remake 現有的元件成本模型，
// 並把結果作為殖民地佇列的 PP 成本；「自動最佳模板」的選擇是近似，公式本身不是。
func RefitCostPP(source, target Ship) int {
	cost := 2 * (shipProductionCost(target) - shipProductionCost(source))
	minimum := ShipCost(source.Class) / 4
	if cost < minimum {
		return minimum
	}
	return cost
}

// PreviewRefit 在不改變狀態的前提下驗證一筆改裝並回傳凍結後的目標模板與成本。
// 介面以它在玩家按下排入前顯示同艦體限制和實際成本；QueueRefit 則重用它，
// 讓預覽與真正入列不會各自維護一份驗證規則。
func (s *GameSession) PreviewRefit(colony, fleetIndex, shipIndex int) (RefitJob, error) {
	if colony < 0 || colony >= len(s.PlayerColonies) {
		return RefitJob{}, fmt.Errorf("殖民地不存在")
	}
	if fleetIndex < 0 || fleetIndex >= len(s.Fleets) {
		return RefitJob{}, fmt.Errorf("艦隊不存在")
	}
	f := &s.Fleets[fleetIndex]
	star := s.PlayerColonyStarIndex(colony)
	if star < 0 || f.AtStar != star || f.ETA != 0 || f.DestStar >= 0 {
		return RefitJob{}, fmt.Errorf("艦隊未停泊在該殖民地星系")
	}
	if shipIndex < 0 || shipIndex >= len(f.Ships) {
		return RefitJob{}, fmt.Errorf("艦艇不存在")
	}
	source := f.Ships[shipIndex]
	if !s.refitFacilityAllows(colony, source.Class) {
		return RefitJob{}, fmt.Errorf("巡洋艦以上改裝需要星基或更高等級軌道基地")
	}
	target, ok := s.bestRefitTarget(source)
	if !ok {
		return RefitJob{}, fmt.Errorf("%s 沒有可套用的同艦體升級", source.Name)
	}
	return RefitJob{Source: source, Target: target, ReturnStar: star}, nil
}

// QueueRefit 從停泊艦隊取走一艘艦，排入同星系殖民地的改裝佇列。
func (s *GameSession) QueueRefit(colony, fleetIndex, shipIndex int) (RefitJob, error) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdQueueRefit, Args: []int{colony, fleetIndex, shipIndex}})
	job, err := s.PreviewRefit(colony, fleetIndex, shipIndex)
	if err != nil {
		return RefitJob{}, err
	}
	if !s.canEnqueueBuild(colony) {
		return RefitJob{}, fmt.Errorf("建造佇列已滿")
	}
	cost := RefitCostPP(job.Source, job.Target)
	f := &s.Fleets[fleetIndex]
	source := f.Ships[shipIndex]
	f.Ships = append(f.Ships[:shipIndex], f.Ships[shipIndex+1:]...)
	if !s.enqueueBuildValue(colony, ColonyBuild{Name: RefitBuildName, Cost: cost, Refit: &job}) {
		// 理論上 canEnqueueBuild 已保證成功；萬一未來佇列規則變動，不能默默吃掉艦艇。
		f.Ships = append(f.Ships, source)
		return RefitJob{}, fmt.Errorf("建造佇列無法加入改裝工作")
	}
	return job, nil
}

// completeRefitJob 把改裝完成的艦艇放回原殖民地星系；找不到停泊艦隊時建立新艦隊，
// 避免既有艦隊已移動後，改裝艦被錯送到另一顆星。
func (s *GameSession) completeRefitJob(job RefitJob) {
	s.ensureFleet()
	if job.ReturnStar >= 0 {
		for i := range s.Fleets {
			f := &s.Fleets[i]
			if f.AtStar == job.ReturnStar && f.ETA == 0 && f.DestStar < 0 {
				f.Ships = append(f.Ships, job.Target)
				return
			}
		}
		f := NewFleet(job.ReturnStar)
		f.Ships = append(f.Ships, job.Target)
		s.Fleets = append(s.Fleets, f)
		return
	}
	s.Fleet().Ships = append(s.Fleet().Ships, job.Target)
}
