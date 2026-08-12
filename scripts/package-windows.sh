#!/usr/bin/env bash
# 公開 Windows amd64 包；原版資料與字型由玩家自行提供。
# 依賴：既有 moo2-ebiten:latest、.docker-cache/go、可寫 dist-all/。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten:latest"
DIST_DIR="${MOO2_DIST_DIR:-${REPO_ROOT}/dist-all}"
APP_NAME="MasterOfOrion2-cht"
STAGE_SUBDIR="${APP_NAME}-windows-amd64"

[ -d "${DIST_DIR}" ] || { echo "找不到輸出目錄: ${DIST_DIR}" >&2; exit 1; }
[ -d "${REPO_ROOT}/.docker-cache/go" ] || {
  echo "缺少 Go 快取: ${REPO_ROOT}/.docker-cache/go；拒絕暗開網路。" >&2
  exit 1
}
docker image inspect "${IMAGE}" >/dev/null 2>&1 || {
  echo "找不到 Docker image ${IMAGE}；請先建立既有工具鏈。" >&2
  exit 1
}

docker run --rm \
  --network none \
  --memory 3g --cpus 2 --pids-limit 256 \
  -u "$(id -u):$(id -g)" \
  -v "${REPO_ROOT}:/src:ro" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${DIST_DIR}:/dist" \
  -w /src \
  -e GOPATH=/go -e GOMODCACHE=/go/pkg/mod \
  -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 \
  -e GOCACHE=/go/build-cache \
  "${IMAGE}" bash -eu -o pipefail -c '
    export PATH=/usr/local/go/bin:$PATH
    mkdir -p /go/build-cache
    STAGE="/tmp/'"${STAGE_SUBDIR}"'"
    rm -rf "$STAGE"
    mkdir -p "$STAGE"

    echo "== 編譯 Windows GUI =="
    go build -buildvcs=false -ldflags="-s -w -H=windowsgui" -o "$STAGE/moo2.exe" ./cmd/moo2
    echo "moo2.exe: OK"

    echo "== 編譯 headless 模擬器 =="
    go build -buildvcs=false -ldflags="-s -w" -o "$STAGE/moo2sim.exe" ./cmd/moo2sim

    echo "== 附帶自製 assets =="
    cp -r assets "$STAGE/assets"

    echo "== 以 Go 標準庫打包 zip =="
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go run scripts/zipdir.go \
      -root "$STAGE" -prefix "'"${STAGE_SUBDIR}"'" \
      -output "/dist/'"${APP_NAME}"'-windows-amd64.zip"
  '

echo "產出: ${DIST_DIR}/${APP_NAME}-windows-amd64.zip"
ls -la "${DIST_DIR}/${APP_NAME}-windows-amd64.zip"
