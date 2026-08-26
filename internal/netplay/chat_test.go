package netplay

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// 原版的環是「滿 14 則就 memmove 掉最舊一則」——不是覆寫、不是停止接收。
func TestChatLogDropsTheOldestWhenFull(t *testing.T) {
	var g ChatLog
	for i := 0; i < ChatLogMax+5; i++ {
		g.Append(0, string(rune('a'+i)))
	}
	if g.Len() != ChatLogMax {
		t.Fatalf("記錄應維持 %d 則,得到 %d", ChatLogMax, g.Len())
	}
	lines := g.Lines()
	// 丟的是最舊的:第 0..4 則消失,留下第 5..18 則。
	if got, want := lines[0].Text, string(rune('a'+5)); got != want {
		t.Errorf("最舊一則應是 %q(前 5 則已被擠掉),得到 %q", want, got)
	}
	if got, want := lines[len(lines)-1].Text, string(rune('a'+ChatLogMax+4)); got != want {
		t.Errorf("最新一則應是 %q,得到 %q", want, got)
	}
}

// 正對照:沒滿之前不該掉任何一則。
// (少了這條,「Append 什麼都不加」也會讓上面那個測試通過。)
func TestChatLogKeepsEverythingBelowTheCap(t *testing.T) {
	var g ChatLog
	for i := 0; i < ChatLogMax; i++ {
		g.Append(0, string(rune('a'+i)))
	}
	if g.Len() != ChatLogMax {
		t.Fatalf("剛好 %d 則不該掉,得到 %d", ChatLogMax, g.Len())
	}
	if g.Lines()[0].Text != "a" {
		t.Errorf("最舊一則應仍是 %q,得到 %q", "a", g.Lines()[0].Text)
	}
}

// 82 byte 的格子:1 byte 發話者 + 字串 + NUL → 內文最多 80 byte。
func TestChatTruncateHonoursTheOriginalBuffer(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := ChatTruncate(long)
	if len(got) != ChatTextMax {
		t.Fatalf("ASCII 應剛好截到 %d byte,得到 %d", ChatTextMax, len(got))
	}
	if ChatTextMax != ChatEntryStride-2 {
		t.Fatalf("上限應由 stride 推出來:%d − 2 = %d,得到 %d",
			ChatEntryStride, ChatEntryStride-2, ChatTextMax)
	}
}

// UTF-8 不能切在半個字上:守住 80 byte 的同時,結果必須是合法字串。
func TestChatTruncateCutsOnRuneBoundaries(t *testing.T) {
	// 每個中文字 3 byte,27 個 = 81 byte > 80 → 只能留 26 個。
	got := ChatTruncate(strings.Repeat("霸", 27))
	if len(got) > ChatTextMax {
		t.Fatalf("截斷後仍超過 %d byte:%d", ChatTextMax, len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatalf("截斷產生了非法 UTF-8:%q", got)
	}
	if n := utf8.RuneCountInString(got); n != 26 {
		t.Errorf("80 byte 放得下 26 個三 byte 字,得到 %d 個", n)
	}
}

// 控制字元(尤其換行)不進聊天列:原版是單行 C 字串,換行會把版面撐爛。
func TestChatTruncateDropsControlCharacters(t *testing.T) {
	got := ChatTruncate("  hello\nworld\t!  ")
	if got != "helloworld!" {
		t.Errorf("控制字元應被丟掉並去掉頭尾空白,得到 %q", got)
	}
}

// 空訊息不進記錄(原版 `cmp byte_1AAC54, 0 / jz` 就擋掉了)。
func TestChatLogIgnoresEmptyMessages(t *testing.T) {
	var g ChatLog
	g.Append(0, "   ")
	g.Append(0, "\n\n")
	if g.Len() != 0 {
		t.Fatalf("空訊息不該進記錄,得到 %d 則", g.Len())
	}
	g.Append(0, "ok")
	if g.Len() != 1 {
		t.Fatalf("非空訊息應進得去,得到 %d 則", g.Len())
	}
}

// 發話者 >= 8 走 GNN 那一路,< 8 走玩家那一路——兩個 sprintf 的格式照抄。
func TestChatPrefixMatchesTheOriginalFormats(t *testing.T) {
	prefix := func(speaker int, name string) string {
		return FormatChatPrefix(speaker, name, "( GNN )  ", "(%s)  ", "#%d")
	}
	if got, want := prefix(ChatGNNSpeaker, "無所謂"), "( GNN )  "; got != want {
		t.Errorf("GNN 前綴應是 %q,得到 %q", want, got)
	}
	if got, want := prefix(ChatGNNSpeaker+3, ""), "( GNN )  "; got != want {
		t.Errorf("超過門檻一律 GNN,應是 %q,得到 %q", want, got)
	}
	if got, want := prefix(0, "薩科拉"), "(薩科拉)  "; got != want {
		t.Errorf("玩家前綴應是 %q,得到 %q", want, got)
	}
	if !strings.HasSuffix(prefix(0, "a"), "  ") {
		t.Error("前綴後面是**兩個**空格(原版格式字串在右括號後有兩格),不是一個")
	}
}

// 門檻本身:7 是玩家、8 是 GNN。寫成測試是因為那個 8 是 `cmp ax, 8 / jge` 讀出來的,
// 改成 <= 會讓最後一位玩家變成新聞台。
func TestChatGNNThresholdIsInclusive(t *testing.T) {
	if (ChatLine{Speaker: ChatGNNSpeaker - 1}).IsGNN() {
		t.Errorf("發話者 %d 應是玩家", ChatGNNSpeaker-1)
	}
	if !(ChatLine{Speaker: ChatGNNSpeaker}).IsGNN() {
		t.Errorf("發話者 %d 應是 GNN", ChatGNNSpeaker)
	}
}

// Lines 回複本:呼叫端改它不該動到記錄。
func TestChatLogLinesIsACopy(t *testing.T) {
	var g ChatLog
	g.Append(0, "原文")
	got := g.Lines()
	got[0].Text = "被改掉"
	if g.Lines()[0].Text != "原文" {
		t.Error("Lines 回的應是複本")
	}
}

// 送出去的訊息帶 KindChat 與截斷後的內文。
func TestChatMessageCarriesTruncatedText(t *testing.T) {
	m := ChatMessage(3, strings.Repeat("y", 200))
	if m.Kind != KindChat {
		t.Errorf("種類應是 %q,得到 %q", KindChat, m.Kind)
	}
	if m.Player != 3 {
		t.Errorf("發話者應是 3,得到 %d", m.Player)
	}
	if len(m.Text) != ChatTextMax {
		t.Errorf("內文應已截到 %d byte,得到 %d", ChatTextMax, len(m.Text))
	}
}

// 版面常數與「等待其他玩家」那張畫面的輸入列必須不打架:
// 首行 243+14 = 257,14 行 × 12,最後一行底部要在輸入列 430 之上。
//
// 這一條是**兩個獨立來源的交叉驗證**:偏移量出自繪製端(sub_F1075),
// 輸入列出自 `Add_Net_Next_Turn_Fields_` @ 0xEFCEA。兩邊不知道彼此,結果卻剛好放得下。
func TestChatLayoutFitsAboveTheInputRow(t *testing.T) {
	const panelY = 243 // 資產 40 的 y(netnextturn.go 算出來的)
	const inputY = 430 // Add_Net_Next_Turn_Fields_ 的真值
	first := panelY + ChatFirstDY
	last := first + (ChatLogMax-1)*ChatLineStep + ChatEraseHeight
	if first != 257 {
		t.Errorf("首行應在 257,得到 %d", first)
	}
	if last > inputY {
		t.Errorf("第 %d 行底部 %d 撞到輸入列 %d", ChatLogMax, last, inputY)
	}
	if inputY-last > ChatLineStep {
		t.Errorf("聊天區底部 %d 與輸入列 %d 之間留了 %d px——超過一行的間距,"+
			"表示偏移量或行數讀錯了", last, inputY, inputY-last)
	}
}
