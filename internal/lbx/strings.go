package lbx

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// LBX 字串資源有多種封裝。對照 openorion2 lbx.cpp TextManager::StringList:
//   - Fixed:u16 count + u16 bufsize + count 個固定寬(bufsize)字串(loadAsset)。
//     用於科技名/種族名/建築名等名稱表。
//   - CStrings:自 offset 起連續的 NUL 結尾字串(loadStrings)。
//
// 字串以第一個 NUL 截斷、去尾端空白(i18n 查找亦 TrimSpace,故 key 一致)。

// ParseFixedStrings 解析 loadAsset 格式(u16 count, u16 bufsize, count×bufsize)。
func ParseFixedStrings(data []byte) ([]string, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("lbx strings: 資料太短")
	}
	count := int(binary.LittleEndian.Uint16(data[0:2]))
	bufsize := int(binary.LittleEndian.Uint16(data[2:4]))
	if count == 0 || bufsize == 0 {
		return nil, fmt.Errorf("lbx strings: count/bufsize 為 0(非固定寬字串資產)")
	}
	if 4+count*bufsize > len(data) {
		return nil, fmt.Errorf("lbx strings: 資料不足(需 %d,有 %d)", 4+count*bufsize, len(data))
	}
	out := make([]string, count)
	for i := 0; i < count; i++ {
		seg := data[4+i*bufsize : 4+(i+1)*bufsize]
		out[i] = cutNul(seg)
	}
	return out, nil
}

// ParseCStrings 自 offset 起讀連續 NUL 結尾字串,直到資料尾。
func ParseCStrings(data []byte, offset int) []string {
	var out []string
	i := offset
	for i < len(data) {
		start := i
		for i < len(data) && data[i] != 0 {
			i++
		}
		out = append(out, strings.TrimRight(string(data[start:i]), " "))
		i++ // 跳過 NUL
	}
	// 去尾端連續空字串(padding)。
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

// ParseFileEntry 解析 loadFile 格式的單一資產:u16 count(應為 1)+ u16 size + C-string。
// 用於 EVENTMSE/DIPLOMSE/COUNCMSG/ANTARMSG/MAINTEXT 等「每資產一則訊息」的封裝。
func ParseFileEntry(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("lbx strings: loadFile 資產太短")
	}
	if binary.LittleEndian.Uint16(data[0:2]) != 1 {
		return "", fmt.Errorf("lbx strings: 非單則 loadFile 資產")
	}
	// data[2:4] 為字串長度,忽略。
	return cutNul(data[4:]), nil
}

func cutNul(b []byte) string {
	if idx := indexByte(b, 0); idx >= 0 {
		b = b[:idx]
	}
	return strings.TrimRight(string(b), " ")
}

func indexByte(b []byte, c byte) int {
	for i := range b {
		if b[i] == c {
			return i
		}
	}
	return -1
}
