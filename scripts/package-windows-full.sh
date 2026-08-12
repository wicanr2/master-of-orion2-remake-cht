#!/usr/bin/env bash
# Windows amd64 本機完整版；含使用者提供的正版資料與字型，僅供本機授權測試。
# 用法：MOO2_DATA=/path/to/mastori2 MOO2_FONT=/path/to/font.ttc scripts/package-windows-full.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten:latest"
DIST_DIR="${MOO2_DIST_DIR:-${REPO_ROOT}/dist-all}"
DATA_DIR="${MOO2_DATA:-}"
FONT_FILE="${MOO2_FONT:-}"
APP_NAME="MasterOfOrion2-cht-full-windows-amd64"

[ -n "${DATA_DIR}" ] && [ -d "${DATA_DIR}" ] || { echo "請提供有效的 MOO2_DATA 遊戲資料目錄。" >&2; exit 1; }
[ -n "${FONT_FILE}" ] && [ -f "${FONT_FILE}" ] || { echo "請提供有效的 MOO2_FONT 字型檔。" >&2; exit 1; }
[ -d "${DIST_DIR}" ] || { echo "找不到輸出目錄: ${DIST_DIR}" >&2; exit 1; }
[ -d "${REPO_ROOT}/.docker-cache/go" ] || { echo "缺少 Go 快取；拒絕暗開網路。" >&2; exit 1; }
docker image inspect "${IMAGE}" >/dev/null 2>&1 || { echo "找不到 Docker image ${IMAGE}。" >&2; exit 1; }

# 與 Linux／macOS 完整包同源的 55 個正常玩家路徑資料。這個集合由靜態消費端與
# 封裝畫廊交叉確認；`stardb.lbx` 只在 resolver 測試作假檔，正版資料不存在。
LBX_LIST="amebafin anatkfin antaroom anwinfin bldg0 bldg1 bldg2 bldg3 bldg4 bldg5 buffer0 cmbtsfx cmbtshp colbldg colgcbt colony colony2 colpups colroads colsum colveggi combat confirm council design dimtvfin diplomat fleet game genwinfn help herodata inbox info intro loserfin mainmenu multigm newgame officer orionfin planets plntdfin plntsum raceopt races racesel science sound starbg stream streamhd techsel turnsum wininfin"

docker run --rm \
  --network none \
  --memory 3g --cpus 2 --pids-limit 256 \
  -u "$(id -u):$(id -g)" \
  -v "${REPO_ROOT}:/src:ro" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${DIST_DIR}:/dist" \
  -v "${DATA_DIR}:/gamedata:ro" \
  -v "${FONT_FILE}:/font.ttc:ro" \
  -w /src \
  -e GOPATH=/go -e GOMODCACHE=/go/pkg/mod \
  -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 \
  -e GOCACHE=/go/build-cache \
  -e "LBX_LIST=${LBX_LIST}" \
  -e "APP_NAME=${APP_NAME}" \
  "${IMAGE}" bash -eu -o pipefail -c '
    export PATH=/usr/local/go/bin:$PATH
    mkdir -p /go/build-cache
    STAGE="/tmp/'"${APP_NAME}"'"
    rm -rf "$STAGE"
    mkdir -p "$STAGE/assets" "$STAGE/gamedata"

    echo "== 編譯 Windows GUI 與 headless 模擬器 =="
    go build -buildvcs=false -ldflags="-s -w -H=windowsgui" -o "$STAGE/moo2.exe" ./cmd/moo2
    go build -buildvcs=false -ldflags="-s -w" -o "$STAGE/moo2sim.exe" ./cmd/moo2sim

    echo "== 複製自製 assets 與本機字型 =="
    cp -r assets/i18n "$STAGE/assets/i18n"
    cp -r assets/fonts "$STAGE/assets/fonts" 2>/dev/null || true
    cp /font.ttc "$STAGE/font.ttc"

    echo "== 複製與 Linux 完整版同源的 LBX =="
    for name in ${LBX_LIST}; do
      found="$(find /gamedata -maxdepth 1 -type f -iname "${name}.lbx" -print -quit)"
      [ -n "$found" ] || { echo "缺少必要 LBX: ${name}.lbx" >&2; exit 1; }
      cp "$found" "$STAGE/gamedata/$(basename "$found" | tr "a-z" "A-Z")"
    done

    cat > "$STAGE/run-full.bat" <<"EOF"
@echo off
setlocal
cd /d "%~dp0"
echo [本機授權完整版] 使用隨包正版資料與字型，請勿重新散布。
moo2.exe -game -lang zh -data "%~dp0gamedata" -font "%~dp0font.ttc" %*
endlocal
EOF

    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go run scripts/zipdir.go \
      -root "$STAGE" -prefix "'"${APP_NAME}"'" \
      -output "/dist/'"${APP_NAME}"'.zip"
  '

echo "本機授權完整版已建立: ${DIST_DIR}/${APP_NAME}.zip"
