#!/usr/bin/env bash
# 把 cmd/moo2(ebiten GUI,需 CGO+X11/OpenGL)+ cmd/moo2sim(headless)打包成
# Linux x86_64 AppImage。全程在 moo2-ebiten docker image 內執行(CLAUDE.md [HARD]:編譯走 docker)。
#
# 用法: scripts/package-appimage.sh
# 產出: dist/MasterOfOrion2-cht-x86_64.AppImage
#
# 做法:
#   1. 容器內用 go build(CGO_ENABLED=1,繼承 Dockerfile.ebiten)編出 moo2 + moo2sim。
#   2. 組 AppDir(.desktop + 佔位圖示,見 scripts/gen-icon.go —— 不含任何版權遊戲美術)。
#   3. 下載 linuxdeploy + appimagetool(快取進 .docker-cache/appimage-tools,避免每次重抓)。
#      容器內無 FUSE,兩者皆以 --appimage-extract-and-run 執行。
#   4. linuxdeploy 自動掃描 moo2 的動態依賴(libGL/libX11 等)塞進 AppDir/usr/lib。
#   5. appimagetool 打包成最終 .AppImage。
#
# 執行期需求:玩家需自備正版 .lbx 遊戲資料夾,用 `moo2 -data <path>` 執行(AppImage
# 本身不含遊戲資產,同 .gitignore 的版權隔離原則)。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten"
DIST_DIR="${REPO_ROOT}/dist"
TOOLS_CACHE="${REPO_ROOT}/.docker-cache/appimage-tools"
APP_NAME="MasterOfOrion2-cht"

mkdir -p "${DIST_DIR}" "${TOOLS_CACHE}" "${REPO_ROOT}/.docker-cache/go"

if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  docker build -t "${IMAGE}" -f "${REPO_ROOT}/docker/Dockerfile.ebiten" "${REPO_ROOT}"
fi

docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${TOOLS_CACHE}:/tools" \
  -v "${DIST_DIR}:/dist" \
  -w /src \
  "${IMAGE}" \
  bash -eu -o pipefail -c '
    APP_NAME="'"${APP_NAME}"'"
    APPDIR=/tmp/AppDir

    echo "== [1/5] go build cmd/moo2 + cmd/moo2sim (CGO_ENABLED=${CGO_ENABLED}) =="
    mkdir -p "${APPDIR}/usr/bin"
    go build -buildvcs=false -ldflags="-s -w" -o "${APPDIR}/usr/bin/moo2" ./cmd/moo2
    go build -buildvcs=false -ldflags="-s -w" -o "${APPDIR}/usr/bin/moo2sim" ./cmd/moo2sim

    echo "== [2/5] 組 AppDir (.desktop + 佔位圖示) =="
    mkdir -p "${APPDIR}/usr/share/applications" \
             "${APPDIR}/usr/share/icons/hicolor/256x256/apps"
    go run scripts/gen-icon.go "${APPDIR}/usr/share/icons/hicolor/256x256/apps/moo2.png"
    cat > "${APPDIR}/usr/share/applications/moo2.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=Master of Orion 2 (中文化 Remake)
Comment=銀河霸主2 go/ebiten remake(需自備正版 .lbx 遊戲資料)
Exec=moo2
Icon=moo2
Categories=Game;StrategyGame;
Terminal=false
EOF

    echo "== [3/5] 下載 linuxdeploy + appimagetool(快取)=="
    LD=/tools/linuxdeploy-x86_64.AppImage
    AT=/tools/appimagetool-x86_64.AppImage
    if [ ! -x "$LD" ]; then
      curl -sSL -o "$LD" https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
      chmod +x "$LD"
    fi
    if [ ! -x "$AT" ]; then
      curl -sSL -o "$AT" https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage
      chmod +x "$AT"
    fi

    echo "== [4/5] linuxdeploy 掃依賴(libGL/libX11 等)=="
    cd /tmp
    "$LD" --appimage-extract-and-run \
      --appdir "${APPDIR}" \
      --executable "${APPDIR}/usr/bin/moo2" \
      --desktop-file "${APPDIR}/usr/share/applications/moo2.desktop" \
      --icon-file "${APPDIR}/usr/share/icons/hicolor/256x256/apps/moo2.png"
    # moo2sim 是純 Go headless 工具,依賴極少(僅 libc),手動確認可執行即可,
    # linuxdeploy 的 --executable 只需指向會用到 GL/X11 的 moo2。

    echo "== [5/5] appimagetool 打包 =="
    "$AT" --appimage-extract-and-run "${APPDIR}" "/dist/${APP_NAME}-x86_64.AppImage"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
ls -la "${DIST_DIR}"
