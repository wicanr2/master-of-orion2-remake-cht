package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// planetEnvironmentLabels 回傳行星的四個環境顯示值，供星圖、殖民地與行星列表
// 共用。Planet 同時保留舊存檔相容用的中文字串與新生成器的 enum ID。
func planetEnvironmentLabels(lang i18n.Lang, p shell.Planet) (climate, gravity, minerals, size string) {
	climate, gravity, minerals, size = p.Climate, p.Gravity, p.Mineral, p.Size
	if lang == i18n.Traditional {
		return
	}
	if p.Gen > 0 {
		return climateName(lang, p.ClimateID), gravityName(lang, p.GravityID),
			mineralsName(lang, p.MineralID), planetSizeName(lang, p.SizeID)
	}
	// 舊存檔只有中文字串；用既有雙語表反查，未知值才保留原文以免
	// 把一個無法辨識的值誤翻成另一種行星。
	return englishEnumName(climateNames[:], p.Climate),
		englishEnumName(gravityNames[:], p.Gravity),
		englishEnumName(mineralNames[:], p.Mineral),
		englishEnumName(planetSizeNames[:], p.Size)
}

func planetSpecialLabel(lang i18n.Lang, special gamedata.PlanetSpecial) string {
	if lang == i18n.Traditional {
		return gamedata.PlanetSpecialName(special)
	}
	return gamedata.PlanetSpecialNameEN(special)
}

// colonyPlanetRows 是殖民地上方面板的行星資訊。
//
// Planet 同時保留舊存檔相容用的中文字串與新生成器的 enum ID。繁中畫面沿用
// 舊字串，避免只為換語言而改動既有顯示；英文畫面則優先使用 Gen=1 的 ID，
// 不把中文字串當成英文回傳。
func colonyPlanetRows(lang i18n.Lang, p shell.Planet) []string {
	if lang == i18n.Traditional {
		rows := []string{p.Climate, p.Size, "礦產" + p.Mineral, "重力" + p.Gravity}
		if sp := planetSpecialLabel(lang, p.SpecialID); sp != "" {
			rows = append(rows, "★"+sp)
		}
		return rows
	}

	climate, gravity, minerals, size := planetEnvironmentLabels(lang, p)
	rows := []string{climate, size, "Minerals " + minerals, "Gravity " + gravity}
	if sp := planetSpecialLabel(lang, p.SpecialID); sp != "" {
		rows = append(rows, "★"+sp)
	}
	return rows
}

// englishEnumName 將一組 [中文,英文] 值反查成英文；舊存檔若帶入未知中文 key，
// 使用明確的 Unknown 安全 fallback，不讓英文畫面把未知資料原樣畫回中文。
func englishEnumName(table [][2]string, zh string) string {
	for _, pair := range table {
		if pair[0] == zh {
			return pair[1]
		}
		if pair[1] == zh {
			return zh
		}
	}
	if strings.TrimSpace(zh) == "" {
		return "Unknown"
	}
	return englishSafeFallback(zh, "Unknown")
}

// englishSafeFallback 保留已經是英文／自訂 ASCII 名稱的資料；只有無法翻譯的
// 漢字資料才改成 generic label。這是顯示層的降級，不改動存檔內的原始 key。
func englishSafeFallback(text, fallback string) string {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return fallback
		}
	}
	return text
}

// colonyBuildingLabel 將殖民地建造佇列使用的中文 key 轉成英文名稱。
// 建築與 Special action 分屬兩張資料表，兩者都必須查，否則英文畫面會在
// Terraforming、Colony Ship 等項目重新露出中文。
func colonyBuildingLabel(lang i18n.Lang, name string) string {
	if name == shell.CapitolBuildName {
		return techCatalog(lang).Translate(name)
	}
	if lang == i18n.Traditional {
		return name
	}
	if b, ok := gamedata.BuildingByNameZH(name); ok {
		return b.NameEN
	}
	if a, ok := gamedata.SpecialActionByNameZH(name); ok {
		return a.NameEN
	}
	return englishSafeFallback(name, "Unknown Build")
}

// buildItemLabel 轉換建造佇列裡的常駐建築、Special action 與兩個特殊設定。
// 「貿易品」與「住宅」不是 gamedata.Buildings，因此不能只呼叫
// colonyBuildingLabel；它們仍是 shell 的識別 key，不能改動資料層字串。
func buildItemLabel(lang i18n.Lang, name string) string {
	if lang == i18n.Traditional {
		return name
	}
	switch name {
	case shell.TradeGoodsBuildName:
		return "Trade Goods"
	case shell.HousingBuildName:
		return "Housing"
	default:
		return colonyBuildingLabel(lang, name)
	}
}

// systemBodyCountLabel 將星系中其他天體的窄欄摘要轉成當前語言。shell 的
// SystemBodyCountText 是中文短字串，不能直接拿到英文畫面；數量則由同一份
// PlanetsAt 真值計算，避免解析中文文字。
func systemBodyCountLabel(lang i18n.Lang, sess *shell.GameSession, star int) string {
	if sess == nil {
		return ""
	}
	if lang == i18n.Traditional {
		return sess.SystemBodyCountText(star)
	}
	n := len(sess.PlanetsAt(star)) - 1
	if n <= 0 {
		return ""
	}
	// 這一欄只有 74px；「%d more」保留原版「另有 N」語意且不被截成半句。
	return fmt.Sprintf("%d more", n)
}

// historyMetricLabel 是歷史圖表的雙語指標名稱。
func historyMetricLabel(lang i18n.Lang, metric shell.HistoryMetric) string {
	if lang == i18n.Traditional {
		return shell.HistoryMetricName(metric)
	}
	switch metric {
	case shell.HistoryTechnology:
		return "Technology"
	case shell.HistoryFleet:
		return "Fleet"
	case shell.HistoryBuildings:
		return "Buildings"
	default:
		return "Population"
	}
}

// newGameDifficultyLabel、newGameGalaxySizeLabel、newGameGalaxyAgeLabel、
// newGameTechLevelLabel 是 NEW GAME 值列的英文顯示名。shell 的 Name 欄是
// UI 中文資料，不是規則 key；這裡集中維持英文對照，避免把中文名稱改掉而
// 影響存檔或設定索引。
func newGameDifficultyLabel(lang i18n.Lang, index int) string {
	if index < 0 || index >= len(shell.Difficulties) {
		return "Unknown"
	}
	if lang == i18n.Traditional {
		return shell.Difficulties[index].Name
	}
	return [...]string{"Tutor", "Easy", "Average", "Hard", "Impossible"}[index]
}

func newGameGalaxySizeLabel(lang i18n.Lang, index int) string {
	if index < 0 || index >= len(shell.GalaxySizes) {
		return "Unknown"
	}
	if lang == i18n.Traditional {
		return shell.GalaxySizes[index].Name
	}
	return [...]string{"Small", "Medium", "Large", "Huge"}[index]
}

func newGameGalaxyAgeLabel(lang i18n.Lang, index int) string {
	if index < 0 || index >= len(shell.GalaxyAges) {
		return "Unknown"
	}
	if lang == i18n.Traditional {
		return shell.GalaxyAges[index].Name
	}
	return [...]string{"Young", "Average", "Mature"}[index]
}

func newGameTechLevelLabel(lang i18n.Lang, index int) string {
	if index < 0 || index >= len(shell.TechLevels) {
		return "Unknown"
	}
	if lang == i18n.Traditional {
		return shell.TechLevels[index].Name
	}
	return [...]string{"Pre-warp", "Average", "Advanced"}[index]
}

// historyEmpireLabels 是歷史圖表圖例用名稱。玩家的「你」不是種族名，英文
// 固定顯示 You；AI 優先取已保存的 RaceIndex，並對沒有該欄位的舊存檔以名稱
// 反查作為相容退路。
func historyEmpireLabels(lang i18n.Lang, s *shell.GameSession) []string {
	if s == nil {
		return nil
	}
	names := s.HistoryEmpireNames()
	if lang == i18n.Traditional {
		return names
	}
	if len(names) > 0 {
		names[0] = "You"
	}
	for i := 1; i < len(names) && i-1 < len(s.AIPlayers); i++ {
		names[i] = aiEmpireEnglishName(s.AIPlayers[i-1])
	}
	return names
}

func aiEmpireEnglishName(ai shell.AIOpponent) string {
	if ai.RaceIndex >= 0 && ai.RaceIndex < len(shell.Races) {
		// RaceIndex=0 可能是舊存檔的零值；只有在名稱也沒有更可靠的
		// 種族線索時才使用 Humans。
		if ai.RaceIndex != 0 || !hasKnownRaceName(ai.Name) {
			return shell.Races[ai.RaceIndex].EnName
		}
	}
	for _, r := range shell.Races {
		if strings.Contains(ai.Name, r.Name) || strings.Contains(ai.Name, r.EnName) {
			return r.EnName
		}
	}
	return ai.Name
}

// aiEmpireLabel 是 AI 帝國在自繪資訊面板上的顯示名。AIPlayers.Name 是
// 規則／存檔使用的名稱，不能為了英文畫面直接改寫。
func aiEmpireLabel(lang i18n.Lang, ai shell.AIOpponent) string {
	if lang == i18n.Traditional {
		return ai.Name
	}
	return aiEmpireEnglishName(ai)
}

// englishKnownRaceText 只翻譯已知種族名，保留星名、玩家自訂名稱等未知字串。
func englishKnownRaceText(text string) string {
	for _, r := range shell.Races {
		if strings.Contains(text, r.Name) {
			out := strings.Replace(text, r.Name, r.EnName, 1)
			// AI 名稱的中文顯示會在種族名後加「人」；英文種族名已經
			// 自帶完整複數／族名，不再保留這個中文後綴。
			out = strings.Replace(out, r.EnName+string(rune(0x4eba)), r.EnName, 1)
			return out
		}
	}
	return englishSafeFallback(text, "Unknown Empire")
}

// enemyDisplayName 將外交／戰鬥標籤轉成當前語言。對手名稱可能來自舊存檔，
// 因此先依 AI 的 RaceIndex 取英文名，再以種族名反查作相容退路。
func enemyDisplayName(lang i18n.Lang, sess *shell.GameSession, name string) string {
	if lang == i18n.Traditional || sess == nil {
		return name
	}
	for _, ai := range sess.AIPlayers {
		raw := ai.Name
		if strings.HasPrefix(raw, "AI (") && strings.HasSuffix(raw, ")") {
			raw = raw[len("AI (") : len(raw)-1]
		}
		if name == raw || name == ai.Name {
			return aiEmpireEnglishName(ai)
		}
	}
	return englishSafeFallback(englishKnownRaceText(name), "Unknown Empire")
}

// combatShipLabel 是敵艦格子與掃描戰報的顯示名。CombatShip.Name 仍是
// 「席隆人艦1」這類中文規則標籤；英文畫面只在最後一刻轉成「Sakkra Ship 1」。
func combatShipLabel(lang i18n.Lang, sess *shell.GameSession, name string) string {
	if lang == i18n.Traditional {
		return name
	}
	for _, r := range shell.Races {
		shipMark := string(rune(0x8266)) // 艦：CombatShip 的中文顯示標記
		prefix := r.Name + shipMark
		peoplePrefix := r.Name + string(rune(0x4eba)) + shipMark
		if strings.HasPrefix(name, peoplePrefix) {
			return fmt.Sprintf("%s Ship %s", r.EnName, strings.TrimPrefix(name, peoplePrefix))
		}
		if strings.HasPrefix(name, prefix) {
			return fmt.Sprintf("%s Ship %s", r.EnName, strings.TrimPrefix(name, prefix))
		}
	}
	return englishSafeFallback(englishKnownRaceText(name), "Unknown Ship")
}

// fighterKindLabel 是戰術戰機摘要的顯示名；shell 內的名稱保持中文，因為
// 它同時是規則資料與測試的穩定 key。
func fighterKindLabel(lang i18n.Lang, k shell.FighterKind) string {
	if lang == i18n.Traditional {
		return shell.FighterKindName(k)
	}
	switch k {
	case shell.FighterHeavy:
		return "Heavy Fighter"
	case shell.FighterBomber:
		return "Bomber"
	case shell.FighterAssaultShuttle:
		return "Assault Shuttle"
	default:
		return "Interceptor"
	}
}

// hotseatNameLabel 把 shell 的席位名稱轉成英文。熱座接管 AI 後的名稱格式是
// 「第 2 位玩家(布拉西人)」，這裡只處理這個穩定格式，其他自訂玩家名原樣保留。
func hotseatNameLabel(lang i18n.Lang, name string) string {
	if lang == i18n.Traditional {
		return name
	}
	marker := " " + string([]rune{0x4f4d, 0x73a9, 0x5bb6}) // 位玩家：shell 的席位 fallback
	prefix := string(rune(0x7b2c)) + " "                   // 第：shell 的席位 fallback
	if strings.HasPrefix(name, prefix) {
		rest := strings.TrimPrefix(name, prefix)
		if at := strings.Index(rest, marker); at > 0 {
			n := rest[:at]
			suffix := rest[at+len(marker):]
			if suffix != "" {
				return fmt.Sprintf("Player %s %s", n, englishKnownRaceText(suffix))
			}
			return fmt.Sprintf("Player %s", n)
		}
	}
	return englishSafeFallback(englishKnownRaceText(name), "Unknown Player")
}

func hasKnownRaceName(name string) bool {
	for _, r := range shell.Races {
		if strings.Contains(name, r.Name) || strings.Contains(name, r.EnName) {
			return true
		}
	}
	return false
}
