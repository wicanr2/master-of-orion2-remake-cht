#!/usr/bin/env bash
# 把 cmd/moo2(ebiten GUI,需 CGO+X11/OpenGL)+ cmd/moo2sim(headless)打包成
# Linux x86_64 AppImage。全程在 moo2-ebiten docker image 內執行(CLAUDE.md [HARD]:編譯走 docker)。
#
# 用法: scripts/package-appimage.sh
# 產出: dist-all/MasterOfOrion2-cht-x86_64.AppImage
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
DIST_DIR="${MOO2_DIST_DIR:-${REPO_ROOT}/dist-all}"
TOOLS_CACHE="${REPO_ROOT}/.docker-cache/appimage-tools"
APP_NAME="MasterOfOrion2-cht"

for required_dir in "${DIST_DIR}" "${TOOLS_CACHE}" "${REPO_ROOT}/.docker-cache/go"; do
  if [[ ! -d "${required_dir}" ]]; then
    echo "缺少既有目錄: ${required_dir}；拒絕在主機端自行建立。" >&2
    exit 1
  fi
done

if ! docker image inspect "${IMAGE}" >/dev/null 2>&1; then
  echo "找不到既有打包映像: ${IMAGE}" >&2
  echo "請先依 docker/Dockerfile.ebiten 準備可重現工具鏈；本腳本不會自行下載或另建映像。" >&2
  exit 1
fi
for tool in linuxdeploy-x86_64.AppImage appimagetool-x86_64.AppImage runtime-x86_64; do
  if [[ ! -x "${TOOLS_CACHE}/${tool}" ]]; then
    echo "缺少離線 AppImage 工具快取: ${TOOLS_CACHE}/${tool}" >&2
    echo "請以受控、明確授權的下載步驟補齊快取後重跑。" >&2
    exit 1
  fi
done

docker run --rm --network none --memory 4g --cpus 2 --pids-limit 256 \
  -u "$(id -u):$(id -g)" \
  -e GOPATH=/go -e GOMODCACHE=/go/pkg/mod -e GOCACHE=/go/build-cache \
  -v "${REPO_ROOT}:/src:ro" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${TOOLS_CACHE}:/tools" \
  -v "${DIST_DIR}:/dist" \
  -w /src \
  "${IMAGE}" \
  bash -eu -o pipefail -c '
    export PATH=/usr/local/go/bin:$PATH
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

    echo "== [3/5] 使用已驗證的離線 AppImage 工具快取 =="
    LD=/tools/linuxdeploy-x86_64.AppImage
    AT=/tools/appimagetool-x86_64.AppImage
    RT=/tools/runtime-x86_64
    [ -x "$LD" ] && [ -x "$AT" ] && [ -x "$RT" ]

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
    OUT="/dist/.${APP_NAME}-x86_64.AppImage.tmp"
    rm -f "$OUT"
    "$AT" --appimage-extract-and-run --runtime-file "$RT" "${APPDIR}" "$OUT"
    mv "$OUT" "/dist/${APP_NAME}-x86_64.AppImage"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-x86_64.AppImage"
ls -la "${DIST_DIR}"
