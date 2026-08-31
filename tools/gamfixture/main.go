// gamfixture 建立原版 MOO2 的最小戰術 oracle 存檔副本。
// 它只改既有 Ship 記錄的 Star/X/Y 與同一存檔的戰鬥模式欄位，不會改寫原始輸入檔。
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

const (
	shipRecordsOffset = 139096
	shipWireSize      = 129
	shipStarOffset    = 101
	// Load_Game_ 先讀 46 bytes，再把 553-byte settings block 讀入 v51；
	// byte_199CB4 取 v51[216]，所以檔案 offset 是 262。
	strategicModeOffset = 46 + 216
)

type shipLocation struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Owner uint8  `json:"owner"`
	Star  int16  `json:"star"`
	X     uint16 `json:"x"`
	Y     uint16 `json:"y"`
}

type changedByte struct {
	Offset int   `json:"offset"`
	Before uint8 `json:"before"`
	After  uint8 `json:"after"`
}

type manifest struct {
	InputSHA256  string        `json:"input_sha256"`
	OutputSHA256 string        `json:"output_sha256"`
	Before       shipLocation  `json:"before"`
	After        shipLocation  `json:"after"`
	Changes      []changedByte `json:"changes"`
	Evidence     string        `json:"evidence"`
	TacticalMode changedByte   `json:"tactical_mode"`
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func location(index int, ship save.Ship) shipLocation {
	return shipLocation{index, ship.Design.Name, ship.Owner, ship.Star, ship.X, ship.Y}
}

func main() {
	in := flag.String("in", "", "輸入 SAVE*.GAM")
	out := flag.String("out", "", "輸出 fixture SAVE*.GAM")
	manifestPath := flag.String("manifest", "", "輸出 JSON 變更清單")
	shipIndex := flag.Int("ship", -1, "要移動的既有 ship index")
	starIndex := flag.Int("star", -1, "目的 star index")
	flag.Parse()
	if *in == "" || *out == "" || *manifestPath == "" || *shipIndex < 0 || *starIndex < 0 {
		flag.Usage()
		os.Exit(2)
	}

	input, err := os.ReadFile(*in)
	if err != nil {
		panic(err)
	}
	state, err := save.Load(input)
	if err != nil {
		panic(err)
	}
	if *shipIndex >= state.ShipCount || *starIndex >= state.StarCount {
		panic("ship 或 star index 超出有效計數")
	}

	before := location(*shipIndex, state.Ships[*shipIndex])
	target := state.Stars[*starIndex]
	output := append([]byte(nil), input...)
	base := shipRecordsOffset + *shipIndex*shipWireSize + shipStarOffset
	binary.LittleEndian.PutUint16(output[base:base+2], uint16(int16(*starIndex)))
	binary.LittleEndian.PutUint16(output[base+2:base+4], target.X)
	binary.LittleEndian.PutUint16(output[base+4:base+6], target.Y)
	modeBefore := output[strategicModeOffset]
	output[strategicModeOffset] = 0

	patched, err := save.Load(output)
	if err != nil {
		panic(err)
	}
	after := location(*shipIndex, patched.Ships[*shipIndex])
	changes := make([]changedByte, 0, 7)
	for i := range input {
		if input[i] != output[i] {
			changes = append(changes, changedByte{i, input[i], output[i]})
		}
	}
	if len(changes) == 0 || len(changes) > 7 {
		panic(fmt.Sprintf("fixture 差異應介於 1..7 bytes，實得 %d", len(changes)))
	}

	report := manifest{
		InputSHA256: digest(input), OutputSHA256: digest(output),
		Before: before, After: after, Changes: changes,
		Evidence:     "Ship wire size 129; Star/X/Y at record +101/+103/+105. Load_Game_ reads the SAVE settings block at file offset 46; block[216] selects tactical(0)/strategic(1).",
		TacticalMode: changedByte{strategicModeOffset, modeBefore, 0},
	}
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*out, output, 0o600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(*manifestPath, append(reportJSON, '\n'), 0o600); err != nil {
		panic(err)
	}
}
