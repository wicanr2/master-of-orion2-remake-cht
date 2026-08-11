package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// versionAssetDirs 是兩個遊戲版本各自的 LBX 搜尋路徑。
// 每一欄都可用逗號串多層目錄，左邊優先；空欄由共用 -data 回退。
// 這讓正版資料留在工作樹之外，同時重現 DOS patch「覆蓋基礎安裝」的載入方式。
type versionAssetDirs struct {
	classic13   []string
	community15 []string
}

func splitAssetDirs(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func cloneAssetDirs(in []string) []string {
	return append([]string(nil), in...)
}

// buildVersionAssetDirs 建立兩版資產路徑。
// -data 是相容舊用法的共用資料根；指定 -data13/-data15 後，該版本改用自己的路徑。
// 未指定的版本不會偷偷借用另一個「版本專用」目錄，避免把 1.31 資產誤標成 1.5。
func buildVersionAssetDirs(shared, classic13, community15 string) (versionAssetDirs, error) {
	base := splitAssetDirs(shared)
	d := versionAssetDirs{
		classic13:   splitAssetDirs(classic13),
		community15: splitAssetDirs(community15),
	}
	if len(d.classic13) == 0 {
		d.classic13 = cloneAssetDirs(base)
	}
	if len(d.community15) == 0 {
		d.community15 = cloneAssetDirs(base)
	}
	if len(d.classic13) == 0 && len(d.community15) == 0 {
		return versionAssetDirs{}, fmt.Errorf("至少要指定 -data、-data13 或 -data15 其中一個遊戲資料路徑")
	}
	return d, nil
}

func buildGameVersionAssets(shared, classic13, community15, requested string) (versionAssetDirs, gamedata.GameVersion, error) {
	dirs, err := buildVersionAssetDirs(shared, classic13, community15)
	if err != nil {
		return versionAssetDirs{}, 0, err
	}
	v, err := initialVersion(requested, dirs, shared, classic13, community15)
	if err != nil {
		return versionAssetDirs{}, 0, err
	}
	return dirs, v, nil
}

func (d versionAssetDirs) forVersion(v gamedata.GameVersion) ([]string, bool) {
	switch v {
	case gamedata.VersionClassic13:
		return d.classic13, len(d.classic13) > 0
	case gamedata.VersionCommunity15:
		return d.community15, len(d.community15) > 0
	default:
		return nil, false
	}
}

// detectVersionFromData 用資料本身的 README 標記決定 auto 預設。
// 目前私有 mastori2 資料的 README 與 ORION95.EXE 都標示 1.31；找不到標記時，
// 延續原有預設 1.5，讓沒有 README 的合法資料目錄仍可運作。
func detectVersionFromData(dirs []string) gamedata.GameVersion {
	for _, dir := range dirs {
		for _, name := range []string{"README.TXT", "README.txt", "readme.txt"} {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			text := strings.ToLower(string(data))
			switch {
			case strings.Contains(text, "version 1.31"), strings.Contains(text, "version 1.3"):
				return gamedata.VersionClassic13
			case strings.Contains(text, "version 1.50"), strings.Contains(text, "version 1.5"):
				return gamedata.VersionCommunity15
			}
		}
	}
	return gamedata.VersionCommunity15
}

// initialVersion 選出主選單第一次顯示的版本。使用者明指定的 -version 優先；
// auto 先採 1.5 專用路徑，再採 1.3 專用路徑，最後對共用資料讀 README 標記。
func initialVersion(requested string, dirs versionAssetDirs, shared, classic13, community15 string) (gamedata.GameVersion, error) {
	switch strings.ToLower(strings.TrimSpace(requested)) {
	case "", "auto":
		if strings.TrimSpace(community15) != "" {
			return gamedata.VersionCommunity15, nil
		}
		if strings.TrimSpace(classic13) != "" && strings.TrimSpace(shared) == "" {
			return gamedata.VersionClassic13, nil
		}
		if strings.TrimSpace(shared) != "" {
			return detectVersionFromData(splitAssetDirs(shared)), nil
		}
		if _, ok := dirs.forVersion(gamedata.VersionCommunity15); ok {
			return gamedata.VersionCommunity15, nil
		}
		return gamedata.VersionClassic13, nil
	case "1.3", "1.31", "13":
		if _, ok := dirs.forVersion(gamedata.VersionClassic13); !ok {
			return 0, fmt.Errorf("未提供 1.31 資產路徑(-data 或 -data13)")
		}
		return gamedata.VersionClassic13, nil
	case "1.5", "1.50", "15":
		if _, ok := dirs.forVersion(gamedata.VersionCommunity15); !ok {
			return 0, fmt.Errorf("未提供 1.5 資產路徑(-data 或 -data15)")
		}
		return gamedata.VersionCommunity15, nil
	default:
		return 0, fmt.Errorf("-version 只接受 auto、1.3 或 1.5,得到 %q", requested)
	}
}

// selectGameVersion 在尚未進入新局前切換規則與資產版本。
// GameSession 的 RuleProfile 仍在新遊戲建立時注入；這裡只換主選單使用的資產，
// 不會把已存在的對局狀態偷偷改成另一版。
func (b *sceneBuilder) selectGameVersion(v gamedata.GameVersion) error {
	dirs, ok := b.versionAssets.forVersion(v)
	if !ok {
		return fmt.Errorf("未提供 %s 的遊戲資產", versionShort(v))
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return fmt.Errorf("載入 %s 遊戲資產: %w", versionShort(v), err)
	}
	b.res = res
	b.gameVersion = v
	// 這些快取可能含有舊版 LBX 的影像或遮罩；版本切換只能發生在主選單，
	// 但清掉仍能避免日後從存檔返回主選單時混用兩版資料。
	b.colChrome = nil
	b.colBldgCache = nil
	b.colVegSizeCache = nil
	b.nebMaskCache = nil
	b.herodataMercs = loadHerodataMercs(b, res)
	if b.session != nil && len(b.herodataMercs) > 0 {
		b.session.SetMercCandidates(b.herodataMercs)
	}
	return nil
}
