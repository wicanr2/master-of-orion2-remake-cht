# 跨平台打包

> 記錄 `cmd/moo2`(ebiten GUI)與 `cmd/moo2sim`(純 Go headless 模擬器)的三平台打包做法,
> 分兩條互補路徑:
> - **CI(GitHub Actions)**:`.github/workflows/build-macos.yml`、`.github/workflows/build-desktop.yml`——macOS 必走(cgo + Apple SDK 限制),Linux/Windows 順便補一份雲端建置。
> - **本機 docker 腳本**(CLAUDE.md [HARD]:編譯走 docker):`scripts/package-appimage.sh`(Linux AppImage)、`scripts/package-windows.sh`(Windows zip)——實際跑過、產出檔已驗證存在,見 §5。macOS 因 cgo+Apple SDK 限制無法用本機 docker 產出,只能靠 CI。

## 0. 為什麼 macOS 要獨立一份 workflow

ebiten 的 macOS backend(`internal/glfw`)是 **CGO + Cocoa/OpenGL**(`cocoa_monitor_darwin.m` 等 Objective-C 檔),`go build` 對 `GOOS=darwin` 時一定要 `CGO_ENABLED=1` 並链接 Cocoa framework。這代表:

- **不能從 Linux 乾淨跨編**:cgo 需要 macOS SDK 的 headers/frameworks(`Cocoa.framework`、`OpenGL.framework`…),Linux 上沒有,`osxcross` 之類的方案也踩 Apple SDK EULA 的灰色地帶。
- **必須用真正的 macOS host 編**,因此走 GitHub Actions 的 `macos-14`(Apple Silicon)runner——這點與本 repo 用 docker 編 Linux 版的策略(`docker/Dockerfile.ebiten`)不同,是唯一「不能全部塞進 docker」的例外。

參考:`mac-app-cross-pack` skill 記錄的是 SDL1.2/C++ 老遊戲(需要 dylibbundler 包 SDL 動態庫),本專案是 **Go/ebiten**,情況更單純——ebiten 只連結 macOS **系統內建 framework**(Cocoa/OpenGL/IOKit,由 ebiten 原始碼的 cgo `LDFLAGS` 指令自動連,見 `internal/glfw` 各 `*_darwin.*` 檔),**不需要 dylibbundler**、不需要額外裝 SDL 系列函式庫。

## 1. macOS:universal binary + .app + .dmg/.tar.gz

`build-macos.yml` 流程:

1. **`macos-14` runner 分別編 arm64(原生)與 amd64(cross-arch)**:
   - arm64:`GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build`——runner 本身是 Apple Silicon,原生編。
   - amd64:同樣在 arm64 runner 上編,但設 `CC="clang -arch x86_64"`。Xcode 的 clang 是 **universal 前端**,不需要 Rosetta 就能直接產生 x86_64 object(Rosetta 是拿來「跑」x86_64 binary,不是拿來「編」);這比另開一個 Intel runner(`macos-13`,已於 2025-12-04 除役)簡單且更省 CI 額度。
2. **`lipo -create` 合併成 universal binary**,並用 `lipo -info` 斷言兩個 slice(arm64 + x86_64)都在,防止「CI 綠燈但玩家機器打不開」這種靜默壞掉。
3. **組 `.app` bundle**,關鍵設計是 **launcher shell script**:
   ```
   MOO2.app/Contents/MacOS/MOO2        ← 真正的 CFBundleExecutable,是個 bash script
   MOO2.app/Contents/Resources/moo2-bin      ← 真正的 universal binary
   MOO2.app/Contents/Resources/moo2sim-bin
   MOO2.app/Contents/Resources/assets/i18n/  ← 譯文 TSV(我們自己的資產,無版權疑慮)
   ```
   `Contents/MacOS/MOO2` 內容只有三行:`cd` 進 `Contents/Resources` 再 `exec moo2-bin "$@"`。
   **原因**:`internal/i18n/registry.go` 的用法示範是 `os.DirFS("assets/i18n")`——**相對路徑**;`cmd/moo2` 目前也沒有 `-i18n <dir>` 這種可覆寫路徑的 flag。Finder 雙擊 `.app` 啟動時的 cwd 通常是使用者家目錄,不是 bundle 內部,若直接把 universal binary 放在 `Contents/MacOS/` 當 `CFBundleExecutable`,程式會找不到 `assets/i18n` 而无法載入譯文。Launcher script 是最小改動的繞過法,不用動 `cmd/moo2` 原始碼。
   - **待辦**(不在本輪 CI 範圍,記在這裡避免下次忘記):`cmd/moo2` 應該加 `-assets <dir>` 或改用 `os.Executable()` 反推 bundle 路徑,才是长期正解;或等 WORKLIST 的字型/i18n `go:embed` 落地後這個問題直接消失(embed 進 binary 就沒有相對路徑問題)。
4. **不含遊戲資料**:`original_game/`、`moo2_patch1.31/`、`moo2_patch1.5/` 是版權檔,`.gitignore` 已排除、CI 也不會去讀取或打包這些目錄。打包出的 `.app` 只有我們自己的程式碼與譯文,玩家啟動時仍需自備正版 `.lbx` 並透過 `-data` 指定路徑(GUI 尚未做「啟動時選資料夾」的對話框,見 WORKLIST 待補項)。
5. **Ad-hoc codesign**(`codesign --sign -`):不是正式簽署(沒有 Apple Developer 憑證),只是讓 bundle 結構符合 codesign 要求、避免某些 Gatekeeper 檢查直接拒絕載入。玩家第一次執行仍需要:
   ```bash
   xattr -dr com.apple.quarantine /Applications/MOO2.app
   ```
   解除 quarantine 隔離(未簽署 app 的標準操作,見 `mac-app-cross-pack` skill 段 3)。
6. **雙格式輸出**:`.dmg`(`hdiutil create -format UDZO`,給 Mac 用戶雙擊安裝)+ `.tar.gz`(給任何平台的開發者取用/CI 驗證——**APFS/UDZO 格式的 .dmg 在非 Mac 平台常讀不到**,7-Zip 舊版看得到 GPT partition 但解不開內層 APFS,`.tar.gz` 是保證能解的備援)。

### Universal binary 驗證重點

`lipo -info` 必須同時列出 `arm64` 與 `x86_64`,workflow 內已用 `grep` 斷言擋 CI。若之後改成「每弧各自 `-with-...-prefix` 分開編、事後 lipo」這種更複雜的路線(本專案目前不需要,因為沒有第三方 dylib 依賴),要留意 `mac-app-cross-pack` skill §1.5 記錄的雷:**dylibbundler 對 universal binary 可能只解析/複製其中一個 slice 的動態庫**,導致主程式是雙架構但依賴庫是單架構——本專案因為 ebiten 只連結系統 framework(不是第三方 dylib),不會踩到這個雷,但如果未來加了需要動態庫的第三方 CGO 依賴,要重新檢查。

## 2. Linux / Windows(`build-desktop.yml`,補足本機跨編困難)

CLAUDE.md 規定「編譯一律走 docker」,本機(Linux dev box)用 `docker/Dockerfile.ebiten` 可以編 **Linux** 版沒問題(docker image 本身就是 Linux)。

> **修正(2026-07-03,實測)**:下面原本假設「跨編 Windows GUI 版需要 mingw-w64 + CGO_ENABLED=1」——這個假設**對 ebiten v2.9.9 不成立**,已用本機 docker 實測 + 核對原始碼推翻,詳見 §5.2。`build-desktop.yml` 目前仍裝 `egor-tensin/setup-mingw` 是保守作法(不影響正確性,只是多裝了用不到的工具鏈),之後可簡化成純 `CGO_ENABLED=0` 跨編、拿掉 mingw 安裝步驟,但這屬於 CI workflow 調整,不在本輪「本機 docker 打包腳本」授權範圍內,先記錄於此供下輪處理。

- 需要 `mingw-w64` 交叉工具鏈(`x86_64-w64-mingw32-gcc`)且要正確餵給 `CC`,版本/ABI 對不齊容易產生連不動或執行期崩潰的 binary。
- ebiten Windows backend 一樣是 CGO(win32 API),純 `GOOS=windows` 不開 CGO 會編出「能編但執行期立刻崩」或功能殘缺的版本。

`build-desktop.yml` 用 `windows-latest` runner 原生編(`egor-tensin/setup-mingw` 裝好 mingw-w64,`CGO_ENABLED=1`),繞開本機交叉編譯的組態地獄;`ubuntu-latest` 編 Linux 版則對齊 `Dockerfile.ebiten` 的套件清單(X11/OpenGL headers)。兩者都只是「CI 上補一份可下載的建置」,不影響本機 docker 開發流程。

- **Linux**:`.tar.gz`(`moo2` + `moo2sim` + `assets/`)。
- **Windows**:`.zip`,`go build` 加 `-ldflags "-H=windowsgui"` 讓雙擊執行時不彈黑底主控台視窗(GUI app 慣例;`moo2sim` 是 CLI 工具所以不加這個 flag)。

## 3. 觸發條件與 artifact 保留

兩份 workflow 目前都用:

- `workflow_dispatch`:手動觸發,方便單獨重跑某平台驗證。
- `push` tag `v*`:之後正式釋出版本時用 tag 觸發完整三平台打包。
- `pull_request`(限定 `cmd/**`、`internal/**`、`assets/i18n/**`、`go.mod/go.sum` 路徑):PR 階段先確認三平台都編得過,及早抓到「本機只測過 Linux docker,結果 macOS/Windows build 壞掉」這種平台特定 regression。

`actions/upload-artifact` 保留 14 天(GitHub Actions 預設值上限前的常見保守值),之後若要長期保存需另外接 GitHub Releases(本輪 CI workflow 不含這段——建立/發布 Release 屬於「對外發布」,依 `~/.claude/rules/30-lcy-agent-boundaries.md` 需先回報再確認,不在本次「只新增 CI workflow + 文件」的授權範圍內)。

## 4. 已知限制 / 待辦

- **相對路徑 i18n**(見上方 §1.3):launcher script 只是繞過,長期應改 `cmd/moo2` 支援可覆寫的 assets 路徑或 `go:embed`。
- **無圖示**:`.app` 目前沒有 `.icns`,Info.plist 未設 `CFBundleIconFile`,Finder 會用系統預設圖示。之後有美術資產(見 `docs/tech/sprite-tile-quality.md`)可以補。
- **未做 Apple 正式簽署/公證(notarization)**:需要付費 Apple Developer 帳號,不在本專案範圍;玩家需自行 `xattr -dr com.apple.quarantine` 解除隔離。
- **未實際在真 Mac 上跑過**(本機無 Mac 裝置):workflow 語法已用 `actionlint` + `python3 -c "import yaml; yaml.safe_load(...)"` 驗證通過,但 build/lipo/codesign/hdiutil 的實際行為要等第一次 CI 跑在 `macos-14` runner 上才能驗證,對應 `retro-game-playtest` skill 精神——CI 綠燈不等於玩家能玩,之後有 Mac 測試機時應補一輪真機驗證。

## 5. 本機 docker 打包腳本(Linux AppImage / Windows zip,實際跑過)

跟 §1–§4 的 CI workflow 不同,這節記錄**本機可重跑、已實際執行並確認產物存在**的兩支 docker 化腳本:
`scripts/package-appimage.sh`、`scripts/package-windows.sh`。兩者都遵守 CLAUDE.md [HARD](編譯走 docker),
不裝系統套件、不污染本機環境。

### 5.1 Linux AppImage(`scripts/package-appimage.sh`)

流程(全部在既有的 `moo2-ebiten` image 內執行,`docker/Dockerfile.ebiten` 額外補了 `file`/`desktop-file-utils`/`ca-certificates` 供 appimagetool 使用):

1. `go build`(image 預設 `CGO_ENABLED=1`)編出 `cmd/moo2`(GUI)與 `cmd/moo2sim`(headless)兩個 Linux amd64 binary,放進 `AppDir/usr/bin/`。
2. 組 `.desktop` + 佔位圖示(`scripts/gen-icon.go`,純 Go stdlib 畫一個深空藍配金環的圖,**不含任何版權遊戲美術**——原版素材依 `.gitignore` 一律不進 repo,圖示也不例外)。
3. 下載 `linuxdeploy` + `appimagetool`(快取進 `.docker-cache/appimage-tools/`,避免每次重抓),兩者皆用 `--appimage-extract-and-run` 執行(容器內無 FUSE,不能直接跑 AppImage 本體)。
4. `linuxdeploy --executable AppDir/usr/bin/moo2` 掃描 `moo2` 的 ELF 動態依賴,把非系統標配的庫(`libXau.so.6`/`libXdmcp.so.6`/`libbsd.so.0`/`libmd.so.0`)複製進 `AppDir/usr/lib/` 並設 rpath;`libX11.so.6`/`libGL` 等被 linuxdeploy 判定為「blacklisted」(視為目標系統一定有的基礎庫,不重複打包,這是 AppImage 官方建議的標準行為,避免驅動相關的 libGL 打包進去反而在異機衝突)。`moo2sim` 是純 Go 靜態連結(`ldd` 顯示 not dynamically linked),linuxdeploy 略過依賴掃描,但檔案本身仍留在 AppDir 內一起打包。
5. `appimagetool` 打包成 `dist/MasterOfOrion2-cht-x86_64.AppImage`。

**實測產出**(2026-07-03):`dist/MasterOfOrion2-cht-x86_64.AppImage`,5.16 MB。解壓驗證 `squashfs-root/usr/bin/` 內同時有 `moo2`(動態連結,x86-64 ELF)與 `moo2sim`(靜態連結)兩個可執行檔,`.desktop` 內容正確。

**執行期需求**:AppImage 本身不含遊戲資產(同 `.gitignore` 版權隔離原則),玩家需自備正版 `.lbx` 資料夾,用 `./MasterOfOrion2-cht-x86_64.AppImage -data <path>` 執行(GUI 尚無「啟動時選資料夾」對話框,見 WORKLIST)。

### 5.2 Windows(`scripts/package-windows.sh`)——實測不需要 mingw-w64

原始規劃(比照 macOS/CI 的假設)是「ebiten Windows backend 是 CGO,需要 `mingw-w64` 交叉工具鏈」。
**實測結果推翻這個假設**:

```
docker run --rm -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.25-bookworm \
  go build -o moo2.exe ./cmd/moo2
```

直接成功,產出正常的 `PE32+ executable (GUI) x86-64` 檔案。核對 `ebiten v2.9.9` 原始碼
(`internal/glfwwin/*.go`,go module cache 內)確認:**該版本的 Windows backend 已整包改寫成純 Go**
(`glfwwin` 套件靠 `golang.org/x/sys/windows` 呼叫 Win32 API;OpenGL 透過 purego 動態 `LoadLibrary` 載入
`opengl32.dll`),整個目錄內**找不到任何 `import "C"`**。用 `objdump -p` 檢查產出的 `moo2.exe` import table,
只看到 `kernel32.dll`(purego 用它的 `LoadLibraryA`/`GetProcAddress` 動態解析其他 DLL,不會出現在靜態
import table 裡)——與「傳統 cgo 連結 user32/gdi32/opengl32」的樣貌不同,但這正是預期行為。

這與 §0/§2 描述的 **macOS backend 不同**——macOS 仍是 cgo + Cocoa framework,兩者不要套同一套假設。
若之後升級 ebiten 版本導致此路徑失效(cgo 依賴又回來),`scripts/package-windows.sh` 內建這段記錄,
應改回裝 `gcc-mingw-w64-x86-64`、設 `CC=x86_64-w64-mingw32-gcc`、`CGO_ENABLED=1`。

流程:

1. `docker run golang:1.25-bookworm`(不需要 `moo2-ebiten` image,純 Go 跨編不需要 X11/OpenGL headers)。
2. `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -H=windowsgui"` 編 `moo2.exe`(`-H=windowsgui` 避免雙擊彈黑底主控台視窗)。
3. 同環境變數編 `moo2sim.exe`(不加 `-H=windowsgui`,是 CLI 工具)。
4. 附帶 `assets/i18n/`(自製譯文,無版權疑慮)。
5. `zip` 打包成 `dist/MasterOfOrion2-cht-windows-amd64.zip`。

**實測產出**(2026-07-03):`dist/MasterOfOrion2-cht-windows-amd64.zip`(4.8 MB),內含 `moo2.exe`(9.9 MB,GUI subsystem)、
`moo2sim.exe`(1.9 MB,console subsystem)、`assets/i18n/*.tsv`。

**已知限制**:本機 docker 只能驗證「跨編成功 + PE 格式/import table 正確」,**無法在 Linux 容器內實際執行
Windows GUI**(沒有 Wine + 真的 Win32 訊息迴圈測試)。要驗證雙擊後真的能開窗、貼圖正常,仍需要真 Windows
機器或 CI 的 `windows-latest` runner(見 §2)實跑一次——這點與 §1 macOS 的「CI 綠燈不等於玩家能玩」是同一個提醒。

### 5.3 快速重跑

```bash
scripts/package-appimage.sh   # → dist/MasterOfOrion2-cht-x86_64.AppImage
scripts/package-windows.sh    # → dist/MasterOfOrion2-cht-windows-amd64.zip
```

兩支腳本都是 idempotent(可重跑覆蓋),`dist/` 已加進 `.gitignore`(打包產物不入 repo,同原則見 §3 的
「建立 Release 需另外授權」)。`.docker-cache/appimage-tools/` 快取 linuxdeploy/appimagetool 的下載,
避免每次重跑都打一次 GitHub Releases。
