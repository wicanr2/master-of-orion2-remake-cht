package audio

// realfile_test.go:在環境變數 MOO2_AUDIO_TEST 指向真實遊戲資料目錄時才跑,
// 驗證解碼器能吃下原版 STREAMHD.LBX / SOUND.LBX(測資為版權物,不入 repo)。
//
//	MOO2_AUDIO_TEST=/path/to/gamedata go test ./internal/audio/...

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

func openRealLBX(t *testing.T, dir, name string) *lbx.Archive {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("開檔 %s 失敗: %v", path, err)
	}
	t.Cleanup(func() { f.Close() })
	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("stat %s 失敗: %v", path, err)
	}
	a, err := lbx.Open(f, fi.Size())
	if err != nil {
		t.Fatalf("解析 %s 失敗: %v", path, err)
	}
	return a
}

func TestRealMusicAndSoundBank(t *testing.T) {
	dir := os.Getenv("MOO2_AUDIO_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_AUDIO_TEST,跳過真實音訊資料測試")
	}

	t.Run("STREAMHD.LBX music", func(t *testing.T) {
		a := openRealLBX(t, dir, "STREAMHD.LBX")
		clips, entryIDs, err := LoadMusic(a)
		if err != nil {
			t.Fatalf("LoadMusic 失敗: %v", err)
		}
		if len(clips) == 0 {
			t.Fatal("LoadMusic 解出 0 條音樂 clip")
		}
		t.Logf("STREAMHD.LBX: 解出 %d 條音樂,entry IDs 前幾筆: %v", len(clips), firstN(entryIDs, 5))
		for i, c := range clips {
			if len(c.PCM) == 0 {
				t.Errorf("clip %d PCM 為空", i)
			}
			if c.SampleRate <= 0 {
				t.Errorf("clip %d SampleRate = %d,不合理", i, c.SampleRate)
			}
		}
	})

	t.Run("SOUND.LBX sound bank", func(t *testing.T) {
		a := openRealLBX(t, dir, "SOUND.LBX")
		sb, err := LoadSoundBank(a)
		if err != nil {
			t.Fatalf("LoadSoundBank 失敗: %v", err)
		}
		if sb.Len() == 0 {
			t.Fatal("LoadSoundBank 解出 0 個具名音效")
		}
		t.Logf("SOUND.LBX: 解出 %d 個具名音效", sb.Len())
		if c := sb.Clip("BUTTON1"); c == nil {
			t.Error("SoundBank 應含 \"BUTTON1\",卻取不到")
		} else if len(c.PCM) == 0 {
			t.Error("BUTTON1 的 PCM 為空")
		}
	})
}

func firstN(s []int, n int) []int {
	if len(s) < n {
		return s
	}
	return s[:n]
}
