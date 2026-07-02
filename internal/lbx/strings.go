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

// diploBufSize 是 DIPLOMSE 每則字串的固定寬度(DIPLOMSG_BUFSIZE)。
const diploBufSize = 200

// ParseDiploStrings 解析 DIPLOMSE 單一資產:u16(=1)+ u16 + u8(variant)+ u8 count
// + count 個 diploBufSize 固定寬字串。對照 openorion2 TextManager::load 的 diplomsg 區塊。
func ParseDiploStrings(data []byte) ([]string, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("lbx diplo: 資產太短")
	}
	if binary.LittleEndian.Uint16(data[0:2]) != 1 {
		return nil, fmt.Errorf("lbx diplo: 非 diplomsg 資產")
	}
	// data[2:4] 忽略;data[4] variant;data[5] count。
	count := int(data[5])
	if count == 0 {
		return nil, nil
	}
	if 6+count*diploBufSize > len(data) {
		return nil, fmt.Errorf("lbx diplo: 資料不足")
	}
	out := make([]string, count)
	for i := 0; i < count; i++ {
		out[i] = cutNul(data[6+i*diploBufSize : 6+(i+1)*diploBufSize])
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

// HELP.LBX asset 0 的記錄佈局(對照 openorion2 lbx.cpp 的 HelpText::load 常數)。
const (
	helpTitleSize    = 80   // HELP_TITLE_SIZE
	helpFilenameSize = 14   // HELP_FILENAME_SIZE(插圖所在 lbx 檔名)
	helpTextSize     = 1300 // HELP_TEXT_SIZE
	helpEntrySize    = 1403 // HELP_ENTRY_SIZE = 80+14+2+2+1+4+1300
)

// HelpEntry 是一則百科條目:標題、插圖連結(archive/asset/frame/section)與本文。
// 本文可能含 \x07 欄位定位碼與 \n 換行(顯示層負責排版)。
type HelpEntry struct {
	Title   string
	Archive string // 插圖所在 .lbx 檔名(空字串表示無插圖)
	AssetID uint16
	Frame   uint16
	Section uint8
	Text    string
}

// ParseHelp 解析 HELP.LBX 的 asset 0:u16 count + u16 size(記錄寬,≥helpEntrySize)+
// count 個記錄。每記錄:title[80] + filename[14] + assetID(u16) + frame(u16) + section(u8)
// + nextParagraph(u32,段落串接,顯示層自理,此處略) + text[1300] + (size-1403 padding)。
func ParseHelp(data []byte) ([]HelpEntry, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("lbx help: 資產太短")
	}
	count := int(binary.LittleEndian.Uint16(data[0:2]))
	size := int(binary.LittleEndian.Uint16(data[2:4]))
	if size < helpEntrySize {
		return nil, fmt.Errorf("lbx help: 記錄寬 %d < %d(非 help 資產)", size, helpEntrySize)
	}
	if 4+count*size > len(data) {
		return nil, fmt.Errorf("lbx help: 資料不足(需 %d,有 %d)", 4+count*size, len(data))
	}
	out := make([]HelpEntry, count)
	for i := 0; i < count; i++ {
		rec := data[4+i*size : 4+(i+1)*size]
		p := 0
		title := cutNul(rec[p : p+helpTitleSize])
		p += helpTitleSize
		archive := cutNul(rec[p : p+helpFilenameSize])
		p += helpFilenameSize
		assetID := binary.LittleEndian.Uint16(rec[p : p+2])
		p += 2
		frame := binary.LittleEndian.Uint16(rec[p : p+2])
		p += 2
		section := rec[p]
		p++
		p += 4 // nextParagraph(u32)略
		text := cutNul(rec[p : p+helpTextSize])
		out[i] = HelpEntry{
			Title: title, Archive: archive, AssetID: assetID,
			Frame: frame, Section: section, Text: text,
		}
	}
	return out, nil
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
