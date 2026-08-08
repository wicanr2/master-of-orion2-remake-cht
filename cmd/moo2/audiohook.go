package main

// audiohook.go:把原版 MOO2 的音樂/音效接進互動畫面。
//
// 音訊格式見 docs/tech/audio-format.md:全部是 LBX 內的 22050 Hz PCM WAV,
// 原封播放即與原版一致。只有互動模式才初始化(headless 無音效卡,略過)。

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	moo2audio "github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
)

// clickSound 由 overlayScreen.update 在按鈕命中時呼叫;未初始化(headless)則為 nil。
var clickSound func()

// moo2SampleRate 是 MOO2 全部音訊的取樣率(見格式研究文件)。
const moo2SampleRate = 22050

// ============ 場景 → 曲目:直接從執行檔讀出來的(第 77 項起用,取代時長啟發式)============
//
// 先前這一段寫著一長串「第二輪時長啟發式、低信心、待人耳聆聽定案」的推定。
// **那條路不必走了**:場景播哪一首是 `Orion2.exe` 裡的立即數(第 73 項(音樂場景表))。
//
// 三個入口:
//
//	Play_Streaming_Music_  @ 0x24677  指定曲目
//	Play_Background_Music_ @ 0x2484F  clock()%3+1 → STREAM 1/2/3 隨機(15 個呼叫端)
//	Play_Combat_Music_     @ 0x2496C  clock()%3+4 → STREAM 4/5/6 隨機(只有 Tactical_Combat_ 叫它)
//
// **單一編號空間**:曲目編號 ≤100 → `STREAM.LBX` 的 entry;>100 → `STREAMHD.LBX` 的
// entry(索引 = 編號 − 100)。所以本檔案要同時載入兩個 LBX,而不是像先前只載 STREAMHD。
//
// ⚠ **「主選單該放哪一首」這個問題問錯了。** 原版主選單走 `Play_Background_Music_`
// ——每次進去都在 STREAM 1/2/3 之間**重新隨機**。remake 先前固定一首,那是行為差異,
// 不是「還沒定案」。

// 曲目編號(原版的單一編號空間;>100 = STREAMHD)。
const (
	trackScienceRoom  = 117 // Science_Room_ / _Tech_Select_(播完接隨機 STREAM 1..3)
	trackEventScreen  = 118 // Start_Main_Event_ / Draw_Event_Screen_
	trackCouncil      = 119 // Main_Council_Screen_
	trackAntaranRoom  = 120 // Main_Antaran_Room_Screen_
	trackShipDesign   = 8   // Design_Screen_ / Draw_Design_Screen_
	trackColonyCombat = 10  // Colony_Combat_Screen_
)

// backgroundMusicTracks / combatMusicTracks 對應 Play_Background_Music_ 與
// Play_Combat_Music_ 的 `clock()%3 + N`。原版用 clock() 當亂數源,remake 用 math/rand
// ——兩者都是「均勻三選一」,而 clock() 在 headless 重跑時不可重現,所以不照抄。
var (
	backgroundMusicTracks = [3]int{1, 2, 3}
	combatMusicTracks     = [3]int{4, 5, 6}
)

// bgmDiploBadPool 是外交「關係差」音樂的候選集合——反組譯實證
// `_diplomacy_bad_music = Get_Random(3) + 13`(見 docs/tech/audio-track-map.md §7)。
//
// 那三個是 **STREAMHD 的編號**,所以在單一編號空間裡是 113/114/115。
var bgmDiploBadPool = [3]int{113, 114, 115}

// diploGoodMusicTrack 回傳「關係好」時該族的外交曲目編號。
//
// 反組譯給的公式是 `_diplomacy_good_music = 該族 empire 記錄 [+0x25] + 1`。
// 上一輪把它記成「逐族資料驅動,那張逐族靜態表還沒追出來」——**那張表不存在**:
// `sub_12983`(帝國建立)顯示 `[+0x25]` 就是**種族索引**本身,由
// `Random_(13,1)` 從 13 族裡挑到不重複再寫進去,接著用同一個索引查
// `dword_192630[idx*4]` 取種族名字串。所以公式就是字面上的「種族索引 + 1」,
// 沒有中間表。
//
// raceIdx 用 `diplomatRaceIndex`(0..12,原版字母序,已對 RACESEL/DIPLOMAT 逐族核實),
// 那正好與 `[+0x25]` 是同一套編號。
func diploGoodMusicTrack(raceIdx int) int {
	if raceIdx < 0 || raceIdx > 12 {
		return bgmDiploBadPool[0] // 認不出來的族(自訂種族)退回關係差那一組
	}
	return 100 + raceIdx + 1
}

var (
	theMixer *moo2audio.Mixer
	// musicByTrack 是**原版編號空間** → 已解碼曲目。
	// 鍵是曲目編號(≤100 = STREAM 的 entry id;>100 = 100 + STREAMHD 的 entry id),
	// 不是「第幾條 WAV」——`LoadMusic` 會略過非 WAV 的槽(STREAM entry 0 是彩蛋字串、
	// 7/9 是空槽),所以序號與 entry id 並不一致,一定要用 LoadMusic 回傳的 entryIDs 配對。
	musicByTrack = map[int]*moo2audio.Clip{}
	curBGM       = -1
)

// playSceneBGM 切換背景音樂到指定**原版曲目編號**(headless / 該曲不存在時為 no-op)。
// 同曲重播則略過,避免每次進同場景就從頭。
func playSceneBGM(track int) {
	c := musicByTrack[track]
	if theMixer == nil || c == nil || track == curBGM {
		return
	}
	if err := theMixer.PlayBGM(c); err == nil {
		curBGM = track
	}
}

// playBackgroundMusic 對應 `Play_Background_Music_`(0x2484F):STREAM 1/2/3 均等三選一。
//
// **每次呼叫都重擲**,與原版一致——主選單、星圖、多數畫面都走這裡,所以同一局裡進出
// 主選單會聽到不同的曲子。曲目相同時 playSceneBGM 會略過,不會從頭重播。
func playBackgroundMusic() { playSceneBGM(backgroundMusicTracks[rand.Intn(3)]) }

// playCombatMusic 對應 `Play_Combat_Music_`(0x2496C):STREAM 4/5/6 均等三選一。
// 原版只有 `Tactical_Combat_` 呼叫它——戰鬥音樂是獨立的一組,不是背景樂的延續。
func playCombatMusic() { playSceneBGM(combatMusicTracks[rand.Intn(3)]) }

// playSceneBGMOnce 播一首**不迴圈**的曲子;播完由 tickBGM 接上隨機 STREAM 1..3。
//
// 對應 `Play_Streaming_Music_` 的 `edx = −2` 哨兵(第 78 項(音樂接線))。原版只有科學室走這條。
func playSceneBGMOnce(track int) {
	c := musicByTrack[track]
	if theMixer == nil || c == nil || track == curBGM {
		return
	}
	if err := theMixer.PlayBGMOnce(c); err == nil {
		curBGM = track
	}
}

// tickBGM 每幀檢查單次播放的曲子是不是播完了,播完就接隨機 STREAM 1..3。
//
// 由互動主迴圈呼叫(interactiveApp.Update)。headless / 沒有音訊裝置時 theMixer 為 nil,
// 整個函式是 no-op。
func tickBGM() {
	if theMixer == nil || !theMixer.BGMFinished() {
		return
	}
	curBGM = -1 // 這樣接下來擲到同一首也會真的重播
	playBackgroundMusic()
}

// playDiplomacyMusic 播外交音樂:關係好走該族的專屬曲,關係差走三選一。
//
// 原版的 `Start_Diplomacy_Music_` 搭配 `_diplomacy_good_music` / `_diplomacy_bad_music`
// 兩個變數,依「當下與該族的關係」切換——那是這個函式兩個分支的來源,不是 remake 自創的。
//
// ⚠ **「好/壞」的門檻是 remake 的**:原版用什麼條件切換這兩個變數還沒追出來
// (那是 `Start_Diplomacy_Music_` 的呼叫端,不是這兩個變數的賦值處)。
// 這裡用關係分數 >= 0 當分界,是 remake 的讀法。
func playDiplomacyMusic(raceIdx, relation int) {
	if relation >= 0 {
		playSceneBGM(diploGoodMusicTrack(raceIdx))
		return
	}
	playSceneBGM(bgmDiploBadPool[rand.Intn(len(bgmDiploBadPool))])
}

// loadMusicLBX 把一個音樂 LBX 的曲目併進 musicByTrack,鍵加上 base 偏移
// (STREAM base=0、STREAMHD base=100,即原版的單一編號空間)。
func loadMusicLBX(res *assets.Resolver, name string, base int) {
	arch, err := res.OpenLBX(name)
	if err != nil {
		return // 缺這個 LBX 不致命(玩家資料夾可能只有其中一個)
	}
	clips, entryIDs, err := moo2audio.LoadMusic(arch)
	if err != nil {
		fmt.Printf("音樂載入失敗(略過)%s: %v\n", name, err)
		return
	}
	for i, c := range clips {
		musicByTrack[base+entryIDs[i]] = c
	}
}

// initAudio 建立 Mixer、載入全部背景音樂與按鈕音效,回傳 Mixer(需被持有以免 GC)。
// 任何一步失敗都不致命:音訊是加分項,絕不擋遊戲執行。
func initAudio(res *assets.Resolver) *moo2audio.Mixer {
	m := moo2audio.NewMixer(moo2SampleRate)
	theMixer = m

	// 兩個音樂 LBX 都要載:原版的曲目編號是**單一空間**,≤100 指 STREAM、>100 指 STREAMHD。
	// 先前只載 STREAMHD,所以 STREAM 那 8 條(含背景樂與戰鬥樂的全部六首)一條都播不到。
	loadMusicLBX(res, "stream.lbx", 0)
	loadMusicLBX(res, "streamhd.lbx", 100)
	playBackgroundMusic() // 開場:主選單走 Play_Background_Music_,不是固定曲

	// UI + 戰鬥音效:SOUND.LBX 的具名音效(全部已解碼,見 docs/tech/audio-format.md;
	// CMBTSFX 是視覺特效非音效)。註冊本遊戲會用到的幾個,並接上對應的播放閉包。
	if arch, err := res.OpenLBX("sound.lbx"); err == nil {
		if sb, err := moo2audio.LoadSoundBank(arch); err == nil {
			reg := func(name string) func() {
				c := sb.Clip(name)
				if c == nil {
					return nil // 該音效不存在時回 nil,呼叫端 nil-safe
				}
				m.RegisterSFX(name, c)
				return func() { m.PlaySFX(name) }
			}
			clickSound = reg("BUTTON1")   // 按鈕點擊
			sfxFireBeam = reg("NRGBLAST") // 光束/能量武器開火
			sfxFireMissile = reg("MISLFIRE")
			sfxHit = reg("SHIPHIT1") // 命中船體
			sfxExplode = reg("EXPL-1")
		} else {
			fmt.Println("音效載入失敗(略過):", err)
		}
	}

	return m
}

// 戰鬥音效播放閉包(由 initAudio 設定;headless / 缺該音效時為 nil,呼叫端須 nil-check)。
// 戰術戰鬥 fireRound 依「武器類型 / 是否命中 / 是否擊毀」呼叫,音效來源全為 SOUND.LBX
// 現成具名音效,不需逆向任何格式(見 docs/tech/audio-format.md 2026-07-11 訂正)。
var (
	sfxFireBeam    func() // 光束/能量/球狀武器開火(NRGBLAST)
	sfxFireMissile func() // 飛彈/魚雷開火(MISLFIRE)
	sfxHit         func() // 命中敵艦船體(SHIPHIT1)
	sfxExplode     func() // 敵艦被擊毀(EXPL-1)
)

// playSFX 播放一個音效閉包,nil(headless / 未載入)則 no-op。
func playSFX(fn func()) {
	if fn != nil {
		fn()
	}
}

// fireSFX 依「首艘開火艦是否為飛彈類」回傳對應的開火音效閉包。
func fireSFX(missile bool) func() {
	if missile {
		return sfxFireMissile
	}
	return sfxFireBeam
}
