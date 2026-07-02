#!/usr/bin/env bash
# 以 moo2-ebiten image(CGO + GL)測試/編譯依賴 ebiten 的套件(internal/uifont、cmd/moo2)。
# 需在 xvfb 下跑(ebiten 測試會初始化圖形)。可用 MOO2_FONT_TEST 指定 CJK 字型跑解析探測。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="moo2-ebiten"

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  docker build -t "$IMAGE" -f "$REPO_ROOT/docker/Dockerfile.ebiten" "$REPO_ROOT"
fi

MOUNT_ARGS=()
ENV_ARGS=()
if [[ -n "${MOO2_FONT_TEST:-}" ]]; then
  # 保留原副檔名(.ttc 需走 LoadCollection);以 basename 對映容器內路徑。
  base="/testdata/$(basename "${MOO2_FONT_TEST}")"
  MOUNT_ARGS+=(-v "${MOO2_FONT_TEST}:${base}:ro")
  ENV_ARGS+=(-e "MOO2_FONT_TEST=${base}")
fi

# ebiten 的 internal/ui 在 import 時即 init GLFW,故需 DISPLAY。用手動 Xvfb
# (xvfb-run -a 曾在此環境卡住,改自起 Xvfb + 設 DISPLAY)。
exec docker run --rm \
  -v "${REPO_ROOT}:/src" \
  -v "${REPO_ROOT}/.docker-cache/go:/go" \
  -w /src \
  "${ENV_ARGS[@]}" "${MOUNT_ARGS[@]}" \
  "$IMAGE" \
  bash -c "Xvfb :99 -screen 0 640x480x24 >/dev/null 2>&1 & sleep 2; export DISPLAY=:99; go test -buildvcs=false ${*:-./internal/uifont/ ./cmd/...}"
