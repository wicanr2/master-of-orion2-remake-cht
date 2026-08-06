package gamedata

// cutscene.go:過場影片的檔名與用途對映。
//
// MOO2 的過場是**裸的 Smacker 檔**,只是沿用了 `.LBX` 副檔名(解碼器見 internal/smk)。
//
// ============ 這份對映是怎麼定出來的 ============
//
// **不是照檔名猜的**——專案紀律明訂「一手資料勝過檔名推論」。三條獨立證據:
//
// ① **反組譯**(唯一的直接證據,但只涵蓋三個檔):執行檔字串表裡只有三個 `*FIN` 名字,
//    而且呼叫端明白告訴我們它們不是結局:
//      AMEBAFIN.LBX ← `Bomb_Results_Popups_` @ 0xE85F7、`Do_Attacker_Beat_Colony_Stuff_` @ 0xE87D2
//      PLNTDFIN.LBX ← 同上兩個
//      DIMTVFIN.LBX ← `Tactical_Combat_` @ 0x47939
//    其餘六個名字**執行檔完全沒有字面引用**,存在 `ESTRINGS.LBX` 的字串池裡
//    (按字母排序的資產名池,前後文認不出語意)。
//
// ② **尺寸分群**(內容自明的結構訊號):九個檔剛好分成兩群——
//      480×160:INTRO、WININFIN、GENWINFN、LOSERFIN、ANWINFIN
//      288×208:AMEBAFIN、PLNTDFIN、DIMTVFIN、ORIONFIN、ANATKFIN
//    ①已證實的三個遊戲內事件動畫**全都是 288×208**,而片頭是 480×160。
//    所以 ORIONFIN / ANATKFIN 跟著事件那群走,不是結局。
//
// ③ **末幀內容**(解出來直接看):GENWINFN 收在**遊戲標題 logo**(完整結局片的收尾)、
//    LOSERFIN 收在**燃燒中的廢墟城市**(敗北)、ORIONFIN 收在**行星表面的城市**
//    (與 288×208 事件群一致)。
//
// ⚠ **仍未定**:`WININFIN` 與 `GENWINFN` 都在結局群,但「哪一個對應哪一種勝利」沒有證據
// ——挑選它們的程式碼不在執行檔的字面引用裡。remake 目前一律用 `GENWINFN`(已由末幀
// 確認是完整結局片),把 `WININFIN` 標為待定,不臆測。

// CutsceneKind 是一段過場的用途。
type CutsceneKind int

const (
	// CutsceneIntro 是開場片頭。
	CutsceneIntro CutsceneKind = iota
	// CutsceneWin 是勝利結局。
	CutsceneWin
	// CutsceneAntaranWin 是擊敗安塔蘭的勝利結局。
	CutsceneAntaranWin
	// CutsceneDefeat 是敗北結局。
	CutsceneDefeat
)

// cutsceneFile 是各用途對應的檔名(依據見檔頭)。
var cutsceneFile = map[CutsceneKind]string{
	CutsceneIntro:      "intro.lbx",
	CutsceneWin:        "genwinfn.lbx",
	CutsceneAntaranWin: "anwinfin.lbx",
	CutsceneDefeat:     "loserfin.lbx",
}

// CutsceneFileFor 回傳某用途的過場檔名;沒有對應的回空字串。
func CutsceneFileFor(k CutsceneKind) string { return cutsceneFile[k] }

// UnmappedCutscenes 是已知存在、但**尚未確定對應到哪個流程**的過場檔。
// 留著是為了讓後人知道「這些不是漏掉,是還沒有證據可以定案」:
//
//   - wininfin.lbx  結局群(480×160),與 genwinfn 的分工未定
//   - orionfin.lbx  事件群(288×208),推測與攻下獵戶座有關,但 remake 還沒有獵戶座星系
//   - anatkfin.lbx  事件群(288×208),推測與安塔蘭突襲有關
//   - amebafin.lbx  反組譯證實:轟炸結果 / 攻方攻陷殖民地時播
//   - plntdfin.lbx  同上(名字像 PLaNet Destroyed,但用途以呼叫端為準)
//   - dimtvfin.lbx  反組譯證實:戰術戰鬥中播
var UnmappedCutscenes = []string{
	"wininfin.lbx", "orionfin.lbx", "anatkfin.lbx",
	"amebafin.lbx", "plntdfin.lbx", "dimtvfin.lbx",
}
