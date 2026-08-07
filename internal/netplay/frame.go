// Package netplay 是**網路對戰的傳輸層與同步協定**。
//
// 這個套件刻意不相依 `internal/shell`:它只搬位元組與「這一回合誰做了什麼」,
// 規則怎麼跑是 shell 的事。這樣兩邊可以各自測——傳輸層用 `net.Pipe()` 就測得完,
// 不需要真的開一局遊戲。
//
// ============ 原版的形狀(反組譯,不是猜的)============
//
// `Net_Next_Turn_` @ 0xFC470 的骨架:
//
//	byte_1AAF7E[本方玩家] = 1                 ; 標記「我這回合結束了」
//	for 每個其他玩家:
//	    if 玩家還在線(`[player+0x28] == 0x64`):
//	        sub_F6816(...)                     ; 把自己的狀態送過去
//	Wait_Until_Net_Opponent_Finished_ @ 0x3FBFE ; 等對手
//
// 也就是**鎖步(lock-step)**:每個人各自下完指令 → 廣播「我好了」→ 全部到齊才推進回合。
// `byte_1AAF76` / `byte_1AAF7E` 是逐玩家的旗標陣列(玩家結構 stride 0xEA9)。
//
// remake 照這個形狀做,但**傳輸換成 TCP + JSON**:原版走 IPX / 數據機 / 序列埠,
// 那三種在現在的機器上都不存在。這是**移植決策不是還原**,誠實標在這裡。
//
// ============ 為什麼要有長度前綴 ============
//
// TCP 是位元組流,沒有訊息邊界。`Write` 一次寫進去的東西,對方可能分兩次讀到,
// 也可能與下一則黏在一起。所以每一則訊息前面放 4 位元組大端長度,讀的時候先讀長度再讀內容。
package netplay

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// MaxFrameBytes 是單則訊息的上限。
//
// 有上限的理由不是省記憶體,是**不讓對面一個壞掉(或惡意)的長度欄位要求我們配置 4 GB**。
// 一則回合封包裝的是幾十條指令,4 MB 遠遠夠用。
const MaxFrameBytes = 4 << 20

// ErrFrameTooLarge 是長度前綴超過 MaxFrameBytes。
var ErrFrameTooLarge = errors.New("netplay: 訊息長度超過上限")

// WriteFrame 把 v 序列化成 JSON,加上 4 位元組大端長度前綴後寫出去。
func WriteFrame(w io.Writer, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("netplay: 序列化失敗:%w", err)
	}
	if len(body) > MaxFrameBytes {
		return ErrFrameTooLarge
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(body)))
	// 一次寫出去:分兩次 Write 在 net.Pipe（無緩衝）上會讓對面先讀到長度再阻塞等內容,
	// 單執行緒的測試會就這樣死鎖。
	if _, err := w.Write(append(hdr[:], body...)); err != nil {
		return fmt.Errorf("netplay: 寫入失敗:%w", err)
	}
	return nil
}

// ReadFrame 讀一則訊息並反序列化到 v。
//
// 讀到 EOF 時原樣回傳 io.EOF,呼叫端據此判斷「對面關了」而不是「壞掉了」。
func ReadFrame(r io.Reader, v any) error {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("netplay: 讀長度失敗:%w", err)
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > MaxFrameBytes {
		return ErrFrameTooLarge
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return fmt.Errorf("netplay: 讀內容失敗(長度 %d):%w", n, err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("netplay: 反序列化失敗:%w", err)
	}
	return nil
}
