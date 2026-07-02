#!/usr/bin/env bash
# 在 docker + xvfb 下 headless 跑 moo2 並截圖(承襲 mom「編譯綠 ≠ 畫面對」驗證紀律)。
# 需要玩家自備的遊戲資料夾(含 *.lbx)。
#
# 用法:
#   scripts/screenshot.sh <遊戲資料夾> <輸出.png> [-- 額外的 moo2 參數]
# 例:
#   scripts/screenshot.sh /path/to/mastori2 out.png -- -lbx mainmenu.lbx -asset 21
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten"

if [[ $# -lt 2 ]]; then
  echo "用法: scripts/screenshot.sh <遊戲資料夾> <輸出.png> [-- 額外 moo2 參數]" >&2
  exit 2
fi
DATA_DIR="$(cd "$1" && pwd)"; shift
OUT="$1"; shift
EXTRA=()
if [[ "${1:-}" == "--" ]]; then shift; EXTRA=("$@"); fi
OUT_DIR="$(cd "$(dirname "$OUT")" && pwd)"; OUT_BASE="$(basename "$OUT")"

# 確保 image 存在。
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  docker build -t "$IMAGE" -f "$REPO_ROOT/docker/Dockerfile.ebiten" "$REPO_ROOT"
fi

docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -v "${DATA_DIR}:/data:ro" \
  -v "${OUT_DIR}:/out" \
  -w /src "$IMAGE" \
  bash -c "go build -buildvcs=false -o /tmp/moo2 ./cmd/moo2 && \
    xvfb-run -a -s '-screen 0 800x600x24' \
    /tmp/moo2 -data /data -shot /out/${OUT_BASE} ${EXTRA[*]:-}"

echo "截圖輸出:${OUT}"
