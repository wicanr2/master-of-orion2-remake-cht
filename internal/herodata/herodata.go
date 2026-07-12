// Package herodata 解析原版 HERODATA.LBX(英雄/領袖定義表)。
//
// 格式(逆向,交叉驗證 openorion2 gamestate.cpp Leader::load + gamestate.h 常數):
// HERODATA.LBX asset 0 = 4-byte header(count uint16LE、recordSize uint16LE)+ count 筆
// 固定長記錄。實測真檔:count=67、recordSize=59,4 + 67×59 = 3957 bytes 完全吻合。
//
// 每筆 59-byte 記錄欄位佈局(移植自 openorion2 Leader::load,gamestate.cpp:581):
//
//	name          [15] NUL 補齊字串(LEADER_NAME_SIZE 0x0f)
//	title         [20] NUL 補齊字串(LEADER_TITLE_SIZE 0x14)
//	type          uint8    // 0=commander/captain(艦艇軍官)、非 0=admin(殖民地領袖)
//	experience    uint16LE
//	commonSkills  uint32LE // 通用技能 bitmask
//	specialSkills uint32LE // 專屬技能 bitmask
//	techs         [3]uint8
//	picture       uint8    // 肖像索引
//	skillValue    uint16LE // 雇用費/技能強度基準
//	level         uint8
//	location      sint16LE // 以下 remake 用不到,略讀
//	eta           uint8
//	displayLevelUp uint8
//	status        sint8
//	playerIndex   sint8
//
// 資料本身(英雄名/技能)為版權物,不入 repo;由玩家自備的 LBX 於執行期解析。
package herodata

import (
	"encoding/binary"
	"fmt"
)

const (
	leaderNameSize  = 15 // openorion2 LEADER_NAME_SIZE 0x0f
	leaderTitleSize = 20 // openorion2 LEADER_TITLE_SIZE 0x14
	recordSize      = 59 // 15 + 20 + 24(其餘定長欄位),與真檔 header recordSize 吻合
)

// Leader 是 HERODATA.LBX 一筆英雄記錄(只保留 remake 用得到的欄位)。
type Leader struct {
	Name          string
	Title         string
	Type          uint8  // 0=艦艇軍官(commander/captain)、非 0=殖民地領袖(admin)
	Experience    uint16
	CommonSkills  uint32 // 通用技能 bitmask(對照 openorion2 LeaderSkills enum COMMON_SKILLS)
	SpecialSkills uint32
	Picture       uint8
	SkillValue    uint16
	Level         uint8
}

// Ship 回傳此領袖是否為艦艇軍官(Type==0 為 commander/captain 類)。
func (l Leader) Ship() bool { return l.Type == 0 }

// Parse 解析 HERODATA.LBX asset 0 的 bytes,回傳英雄清單。
func Parse(data []byte) ([]Leader, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("herodata: 資料過短(%d bytes)", len(data))
	}
	count := int(binary.LittleEndian.Uint16(data[0:2]))
	rsize := int(binary.LittleEndian.Uint16(data[2:4]))
	if rsize != recordSize {
		return nil, fmt.Errorf("herodata: 記錄長 %d,預期 %d(格式不符或非 HERODATA)", rsize, recordSize)
	}
	need := 4 + count*rsize
	if len(data) < need {
		return nil, fmt.Errorf("herodata: 資料 %d bytes 不足 %d(count=%d)", len(data), need, count)
	}
	out := make([]Leader, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, parseRecord(data[4+i*rsize:4+(i+1)*rsize]))
	}
	return out, nil
}

func parseRecord(r []byte) Leader {
	o := 0
	name := cstr(r[o : o+leaderNameSize])
	o += leaderNameSize
	title := cstr(r[o : o+leaderTitleSize])
	o += leaderTitleSize
	typ := r[o]
	o++
	exp := binary.LittleEndian.Uint16(r[o : o+2])
	o += 2
	common := binary.LittleEndian.Uint32(r[o : o+4])
	o += 4
	special := binary.LittleEndian.Uint32(r[o : o+4])
	o += 4
	o += 3 // techs[3]uint8,remake 未用
	pic := r[o]
	o++
	skillVal := binary.LittleEndian.Uint16(r[o : o+2])
	o += 2
	level := r[o]
	// 其餘欄位(location/eta/displayLevelUp/status/playerIndex)remake 用不到,不再讀。
	return Leader{Name: name, Title: title, Type: typ, Experience: exp,
		CommonSkills: common, SpecialSkills: special, Picture: pic, SkillValue: skillVal, Level: level}
}

// cstr 取定長欄位到第一個 NUL 為止的字串,並去尾端空白。
func cstr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			b = b[:i]
			break
		}
	}
	// 去尾端空白(部分 title 以空白補齊)。
	end := len(b)
	for end > 0 && (b[end-1] == ' ' || b[end-1] == 0) {
		end--
	}
	return string(b[:end])
}
