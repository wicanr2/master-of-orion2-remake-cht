#!/usr/bin/env bash
# 把 cmd/moo2(ebiten GUI)+ cmd/moo2sim(headless)跨編成 Windows amd64,打包成 .zip。
# 全程在 docker 內執行(CLAUDE.md [HARD]:編譯走 docker)。
#
# 用法: scripts/package-windows.sh
# 產出: dist/MasterOfOrion2-cht-windows-amd64.zip
#
# 實測記錄(2026-07-03,ebiten v2.9.9):
#   原本預期跟 macOS 一樣「ebiten Windows backend 是 CGO,需要 mingw-w64」,
#   但實測 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/moo2` **直接成功**,
#   產出正常的 PE32+ GUI 執行檔。核對 ebiten v2.9.9 原始碼
#   (`internal/glfwwin/*.go`)後確認:該版本的 Windows backend 已改寫成純 Go
#   (glfwwin 套件,靠 golang.org/x/sys/windows 呼叫 Win32 API,OpenGL 走 purego
#   動態載入 opengl32.dll),整個目錄找不到任何 `import "C"`。
#   → **不需要 mingw-w64**,本腳本因此比原本規劃簡單(不裝交叉編譯工具鏈)。
#   這點與 macOS backend(見 docs/tech/packaging.md,cgo + Cocoa.framework,
#   一定要真 macOS host)不同,不要誤用同一套假設。
#   若之後升級 ebiten 版本導致此路徑失效(cgo 依賴又回來),應改回:
#     docker 內裝 gcc-mingw-w64-x86-64、CC=x86_64-w64-mingw32-gcc、CGO_ENABLED=1
#   並在本檔更新記錄。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="golang:1.25-bookworm"
DIST_DIR="${REPO_ROOT}/dist"
APP_NAME="MasterOfOrion2-cht"
STAGE_SUBDIR="${APP_NAME}-windows-amd64"

mkdir -p "${DIST_DIR}" "${REPO_ROOT}/.docker-cache/go"

docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${DIST_DIR}:/dist" \
  -w /src \
  -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 \
  "${IMAGE}" \
  bash -eu -o pipefail -c '
    STAGE="/tmp/'"${STAGE_SUBDIR}"'"
    mkdir -p "$STAGE"

    echo "== 跨編 cmd/moo2(GUI,-H=windowsgui 避免彈黑窗)=="
    if go build -buildvcs=false -ldflags="-s -w -H=windowsgui" -o "$STAGE/moo2.exe" ./cmd/moo2; then
      echo "moo2.exe: OK"
    else
      echo "!! cmd/moo2 跨編失敗(ebiten 版本可能改回需要 CGO/mingw-w64),moo2.exe 略過" >&2
    fi

    echo "== 跨編 cmd/moo2sim(純 Go headless,穩定可跨編)=="
    go build -buildvcs=false -ldflags="-s -w" -o "$STAGE/moo2sim.exe" ./cmd/moo2sim

    echo "== 附帶 assets(i18n 譯文,無版權疑慮)=="
    cp -r assets "$STAGE/assets"

    echo "== 打包 zip =="
    apt-get update -qq && apt-get install -y -qq zip >/dev/null
    cd /tmp
    zip -qr "/dist/'"${APP_NAME}"'-windows-amd64.zip" "'"${STAGE_SUBDIR}"'"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-windows-amd64.zip"
ls -la "${DIST_DIR}"

cat <<'EOF'

備註:GUI 版(moo2.exe)本機 docker 內只能做「跨編是否成功 + PE 格式檢查」,
無法在 Linux 容器內實際執行 Windows GUI 驗證行為(需要真 Windows 或
GitHub Actions windows-latest runner 原生跑一次,見 docs/tech/packaging.md)。
EOF
