# 技術文件知識庫

逆向 + 移植過程中確認的格式、數值資料與工程心得。每輪把新看到/翻譯的數值資料整理進來。

## 資料格式(逆向紀錄)

- [lbx-format.md](lbx-format.md) — `.lbx` 資產封存檔:容器(magic `0xfead`)、影像 header/flags、frame offset 表、內嵌調色盤(6-bit→8-bit)、scan-line RLE 解碼。
- [savegame-format.md](savegame-format.md) — `save?.gam` 存檔:版本 `0xe0`、關鍵 offset(colonyCount `0x25b`、galaxy `0x31be4`)、各實體結構大小與欄位佈局、真實檔驗證數據。
- [enums.md](enums.md) — 28 個資料枚舉對照(技術 212/研究主題 83/建築 49/種族特性 32/特殊裝備…),英文名即 gameplay 邏輯 key,中文欄待填(中文化術語表基礎)。**自動生成**(`scripts/gen-enums.py` 讀 openorion2 gamestate.h),對應 Go:`internal/gamedata/enums.go`。
- [formulas.md](formulas.md) — 唯讀衍生公式與查表精簡版(艦艇戰力/HP/戰速、行星產出、雇用費),對應 Go:`internal/gamedata/formulas.go`。
- [moo2-formulas-reference.md](moo2-formulas-reference.md) — **遊戲公式參考(完整版)**:殖民地成長、生產/污染、研究樹、軍官、艦艇衍生值、光束命中、飛彈防禦、間諜共 8 個系統,逐條附 openorion2 行號/手冊頁碼出處與驗證範例,含手冊自相矛盾記錄(AMR 命中率、飛彈速度)。對應 `internal/gamedata/*.go` 全部檔案。
- [ebiten-notes.md](ebiten-notes.md) — Phase 2 移植筆記:MOO2=640×480、資料層→ebiten 全鏈路、docker headless(CGO/xvfb/buildvcs)、ReadPixels 截圖。對應:`cmd/moo2`、`docker/Dockerfile.ebiten`、`scripts/screenshot.sh`。
- [music-integration.md](music-integration.md) — 音樂/音效整合可行性:原版音訊架構(Miles AIL/XMIDI)+ **關鍵發現:配樂實為預渲染 PCM WAV**(STREAM/STREAMHD.LBX,22050Hz/8-bit/立體聲,實測)非即時合成 → 直接抽 WAV 餵 ebiten,不需 XMI→SoundFont/OPL。含音效資產(SOUND/CMBTSFX/SPHERSFX)、ebiten 音訊限制、整合路線與版權鐵則。
- [packaging.md](packaging.md) — 跨平台打包,CI + 本機 docker 腳本兩條路徑:macOS 走 GitHub Actions(`macos-14` runner 原生編 arm64+amd64 → `lipo` universal binary → `.app`/`.dmg`/`.tar.gz`,launcher script 繞過相對路徑 i18n 問題,對應 `.github/workflows/build-macos.yml`/`build-desktop.yml`);Linux/Windows 除 CI 外另有**已實測跑過**的本機 docker 打包腳本 `scripts/package-appimage.sh`(linuxdeploy+appimagetool 產 AppImage)、`scripts/package-windows.sh`(**實測 ebiten v2.9.9 Windows backend 已純 Go/purego 化,`CGO_ENABLED=0` 免 mingw-w64 即可跨編**,推翻原本 cgo 假設)。

## 待補(後續輪次)

- 枚舉 enums.md 的中文譯名欄(中文化階段填)。
- `Player::researchCost`(依 LBX research_choices 資料表)。
- ebiten porting 心得、patch 1.3/1.5 差異、選單擴展。

> kick-off 階段的策略/可行性文件在 `../kickoff/`。
