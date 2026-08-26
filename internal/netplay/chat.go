package netplay

import (
	"fmt"
	"strings"
)

// chat.go:網路對局的**聊天列**——原版的訊息環與格式,照抄語意。
//
// ============ 這四個數字都是一手的 ============
//
// `Receive_Chat_Msg_` @ 0xDD351 的收訊端只有 39 行,而那 39 行把整個資料結構寫死了:
//
//	cmp   dword ptr [eax+47Ch], 0Eh   ; 已經有 14 則?
//	jnz   short loc_DD383
//	mov   ebx, 42Ah                    ; 0x42A = 13 × 0x52
//	lea   edx, [eax+52h]               ; 從第 2 則開始
//	call  memmove_                     ; ← 整批往前搬一格
//	mov   dword ptr [eax+47Ch], 0Dh    ; 剩 13 則
//	...
//	imul  edx, [eax+47Ch], 52h
//	mov   [edx+eax], cl                ; ← 第一個 byte 是「誰說的」
//	lea   edi, [eax+1]                 ; ← 第二個 byte 起是字串
//	loop: lodsb / mov [edi],al / inc edi / cmp al,0 / jnz
//
// 讀得出來的:
//
//	| 值 | 出處 | 意思 |
//	|---|---|---|
//	| 14 | `cmp [+47Ch], 0Eh` | 保留幾則;滿了丟**最舊**的一則 |
//	| 82 | `imul …, 52h` | 每則佔幾個 byte |
//	| 80 | 82 − 發話者 1 byte − 結尾 NUL 1 byte | 一則最多幾個字元 |
//	| 8 | `cmp ax, 8 / jge` (繪製端) | 發話者 ≥ 8 是 GNN 新聞,不是玩家 |
//
// 計數欄位落在 `[+0x47C]` 也對得上:14 × 82 = 1148 = 0x47C ——**陣列剛好塞滿到那裡**,
// 計數就接在後面。這是第二條獨立的線,不是同一個數字換句話說。
//
// ============ 兩種前綴 ============
//
// 繪製端(`sub_F1075`)照發話者分兩路,兩個格式字串直接躺在 .data 裡:
//
//	發話者 < 8  → sprintf("(%s)  %s", 玩家名, 內文)     ; 名字取自玩家陣列 +1
//	發話者 ≥ 8  → sprintf("( GNN )  %s", 內文)          ; 固定色 byte_199F34 = 0x10
//
// 注意 `(%s)` 後面是**兩個空格**,`( GNN )` 括號內側也各有一個空格——照抄,不順手改成一個。
// 對齊靠的就是這兩格:`( GNN )` 是 7 個字元,而多數玩家名短於此,兩空格讓內文起點接近。
//
// ============ 版面(相對於承載聊天的面板)============
//
//	x      = 面板x + 0x18 = +24
//	首行 y = 面板y + 0x0E = +14
//	行距   = 0x0C = 12
//	每行先擦一塊 0x23A × 0x0B = 570 × 11 再寫字(拿資產 40 重畫該列)
//
// 這四個數放進「等待其他玩家」那張畫面會自己對上:資產 40 畫在 y=243,
// 首行就是 257,14 行 × 12 = 168,最後一行落在 413(+11 = 424),
// 而 `Add_Net_Next_Turn_Fields_` 給的輸入列在 **430**。中間剩 6 px。
// 版面是算出來的、輸入列是另一個函式給的,兩邊卻嚴絲合縫——這是第二個來源。
//
// ============ 送訊 ============
//
// `Send_Chat_Msg_` @ 0xDD3B8 逐一走過玩家陣列(stride 0xEA9),
// 只發給 `[player+0x28] == 'd'` 的那些,封包型別 `edx = 27h`。
// 那個 `27h = 39` 在 `sub_F5A9F` 的跳表註解裡對上了:`; jumptable 000F5AD4 case 39`
// ——那個 case 正是呼叫 `Receive_Chat_Msg_` 的那一條。**送與收兩端各自指到同一個 39。**
//
// remake 走 JSON over TCP,不需要數字 opcode,但把它記在這裡:
// 之後若要和原版的封包對接,這是唯一的入口號碼。
//
// ============ ⚠ GNN 那一邊訂正過(2026-08-08,第 54 項(三個寫入端))============
//
// 這裡原本寫「GNN 走的是 case 43(`Send_GNN_Chat_Msg_` @ 0xDD42A)」,而那句話
// **把送訊端和收訊 case 湊成了一對,但它們不是同一個號碼**:
//
//	Send_GNN_Chat_Msg_ @ 0xDD42A: mov edx, 2Dh (45)   ; asm 315237
//	收訊跳表 sub_F5A9F:            case 43 (0x2B) 才是「speaker 固定 8 呼叫 Receive_Chat_Msg_」
//
// 而那張 68 格的跳表**沒有 case 45**(實際存在的是 0..67 扣掉 4/45/49/51/53/54)。
// 也就是說:一個 `Send_GNN_Chat_Msg_` 送出去的封包,在收端落到 default 被丟掉。
//
// ⚠ **不要把這句讀成「原版的 GNN 廣播是壞的」**——我核到的是「這條路徑上收發號碼對不起來」,
// 沒有核 `sub_F6816` 是不是唯一的傳送出口、也沒有核有沒有第二個派送器。
// 唯一能確定的是:先前那個 `0x2B` 是**收訊 case 號**,被當成送訊型別號記下來了。
//
// ============ 誠實留白 ============
//
// `[player+0x28] == 'd'`(0x64)是「這個位置該收訊」的條件,但 **0x28 這個欄位是什麼、
// 為什麼是字母 d,沒有查到寫入端**——所以 remake 不照抄這個判斷,改成「發給所有已連線的對手」。
// 兩者在正常對局裡結果相同(連上的就是要收的),差別只在原版某些位置(AI?空位?)被排除。
// 不編一個名字給它。

const (
	// ChatLogMax 是聊天記錄保留幾則(原版 `cmp [chat+47Ch], 0Eh`)。
	ChatLogMax = 14
	// ChatEntryStride 是原版每則佔的 byte 數(`imul …, 52h`)。remake 不用這個佈局,
	// 留著是因為 ChatTextMax 是從它推出來的——刪掉就只剩一個沒有來歷的 80。
	ChatEntryStride = 82
	// ChatTextMax 是一則訊息的字元上限:82 − 發話者 1 − 結尾 NUL 1。
	ChatTextMax = ChatEntryStride - 2
	// ChatGNNSpeaker 是「發話者編號到這裡以上算 GNN 新聞」的門檻(原版 `cmp ax, 8 / jge`)。
	ChatGNNSpeaker = 8

	// ChatOpcode 是原版一般聊天的封包型別號(`Send_Chat_Msg_` 的 `edx = 27h`;
	// 收訊跳表 case 39 對得上)。remake 用 KindChat,這個數只在要和原版對接時才有意義。
	ChatOpcode = 0x27
	// ChatGNNSendOpcode 是 `Send_GNN_Chat_Msg_` 實際送出的型別號(`edx = 2Dh`)。
	// ChatGNNRecvCase 是收訊跳表裡「speaker 固定 8」那一格的 case 號。
	// **兩者不相等**,而且跳表沒有 case 45——詳見檔頭的訂正段落。
	ChatGNNSendOpcode = 0x2D
	ChatGNNRecvCase   = 0x2B
)

// KindChat 是聊天訊息的種類。
const KindChat MsgKind = "chat"

// 版面常數(相對於承載聊天的面板左上角,見檔頭)。
const (
	ChatTextDX      = 0x18 // +24
	ChatFirstDY     = 0x0E // +14
	ChatLineStep    = 0x0C // 12
	ChatEraseWidth  = 0x23A
	ChatEraseHeight = 0x0B
)

// ChatLine 是記錄裡的一則。
//
// Speaker 直接存原版那個 byte 的語意:< 8 是玩家編號,>= 8 是 GNN。
// 不拆成 `IsGNN bool + Player int` 是因為那樣就得決定 GNN 的 Player 填什麼,
// 而原版根本沒有那個欄位。
type ChatLine struct {
	Speaker int    `json:"speaker"`
	Text    string `json:"text"`
}

// IsGNN 回報這則是不是 GNN 新聞而非玩家發言。
func (l ChatLine) IsGNN() bool { return l.Speaker >= ChatGNNSpeaker }

// ChatLog 是聊天記錄環。零值可用。
type ChatLog struct {
	lines []ChatLine
}

// Append 加一則。滿了就丟掉**最舊**的一則(原版的 memmove)。
//
// 空字串不加:原版送訊端在 `Chat_Box_Input_Loop_` 就先 `cmp byte_1AAC54, 0 / jz` 擋掉了,
// 按 Enter 但沒打字不會產生一則空白。
func (g *ChatLog) Append(speaker int, text string) {
	text = ChatTruncate(text)
	if text == "" {
		return
	}
	if speaker < 0 {
		speaker = ChatGNNSpeaker
	}
	g.lines = append(g.lines, ChatLine{Speaker: speaker, Text: text})
	if len(g.lines) > ChatLogMax {
		g.lines = append(g.lines[:0], g.lines[len(g.lines)-ChatLogMax:]...)
	}
}

// AppendGNN 加一則 GNN 新聞。
func (g *ChatLog) AppendGNN(text string) { g.Append(ChatGNNSpeaker, text) }

// Lines 回傳目前的記錄(由舊到新)。回傳的是複本,呼叫端改它不會動到記錄本身。
func (g *ChatLog) Lines() []ChatLine {
	out := make([]ChatLine, len(g.lines))
	copy(out, g.lines)
	return out
}

// Len 回傳目前有幾則。
func (g *ChatLog) Len() int { return len(g.lines) }

// ChatTruncate 把一則訊息截到原版的長度上限,並去掉頭尾空白。
//
// ⚠ **上限是 byte 數,不是字元數**——那是原版的緩衝區大小,不是設計選擇。
// 但截斷點取在 **rune 邊界**:原版是單 byte ASCII,切在哪裡都還是合法字串;
// UTF-8 切在半個中文字上會變成亂碼。守住緩衝區、不產生壞字串,兩件事都要。
//
// 實務上這代表中文一則約 26 個字。那不是 remake 加的限制,是原版 82 byte 的格子。
func ChatTruncate(s string) string {
	s = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1 // 控制字元(含換行)不進聊天列
		}
		return r
	}, s))
	if len(s) <= ChatTextMax {
		return s
	}
	cut := 0
	for i := range s {
		if i > ChatTextMax {
			break
		}
		cut = i
	}
	return strings.TrimSpace(s[:cut])
}

// FormatChatPrefix 依呼叫端提供的外部格式回傳一則前綴。
//
// name 為空時退回顯示發話者編號——不是編一個「玩家」出來:名字缺了是名冊沒同步,
// 蓋掉它只會讓那個問題更難查。
func FormatChatPrefix(speaker int, name, gnnPrefix, playerFormat, unknownNameFormat string) string {
	if speaker >= ChatGNNSpeaker {
		return gnnPrefix
	}
	if name == "" {
		name = fmt.Sprintf(unknownNameFormat, speaker)
	}
	return fmt.Sprintf(playerFormat, name)
}

// ChatMessage 造一則要送出去的聊天訊息。
func ChatMessage(player int, text string) Message {
	return Message{Kind: KindChat, Player: player, Text: ChatTruncate(text)}
}
