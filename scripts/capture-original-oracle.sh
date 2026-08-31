#!/usr/bin/env bash
# 在隔離的 DOSBox-X 容器內擷取原版 640x480 oracle。
# 原版目錄固定唯讀；輸出只寫入使用者指定目錄。
set -euo pipefail

IMAGE="${MOO2_DOSBOX_IMAGE:-civ1-dosboxx-input:20260830}"

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "用法：$0 <含 Orion2.exe 的原版目錄> <輸出目錄> [menu|continue]" >&2
  exit 2
fi

DATA_DIR="$(cd "$1" && pwd)"
OUT_DIR="$(mkdir -p "$2" && cd "$2" && pwd)"
MODE="${3:-menu}"

if [[ ! -f "$DATA_DIR/Orion2.exe" && ! -f "$DATA_DIR/ORION2.EXE" ]]; then
  echo "找不到 Orion2.exe：$DATA_DIR" >&2
  exit 2
fi
if [[ "$MODE" != "menu" && "$MODE" != "continue" ]]; then
  echo "未知模式：$MODE（只接受 menu 或 continue）" >&2
  exit 2
fi
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "缺少既有 DOSBox-X 映像：$IMAGE" >&2
  exit 1
fi

docker run --rm --network none \
  --memory 768m --cpus 2 --pids-limit 256 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/tmp/moo2-oracle-home \
  -e ORACLE_MODE="$MODE" \
  -v "$DATA_DIR:/game:ro" \
  -v "$OUT_DIR:/out" \
  "$IMAGE" bash -lc '
    set -euo pipefail
    mkdir -p "$HOME"
    Xvfb :99 -screen 0 1280x1024x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    xvfb_pid=$!
    export DISPLAY=:99
    dosbox-x -defaultconf -noautoexec \
      -c "mount c /game" -c "c:" -c "orion2" >/tmp/dosbox.log 2>&1 &
    dosbox_pid=$!
    trap "kill $dosbox_pid $xvfb_pid 2>/dev/null || true" EXIT

    # 關閉兩個 DOSBox-X 首次啟動提示，再跳過開場動畫。
    sleep 3
    window_id=$(xdotool search --onlyvisible --name "DOSBox" | tail -1)
    xdotool windowfocus "$window_id"
    xdotool key --window "$window_id" Return
    sleep 2
    window_id=$(xdotool search --onlyvisible --name "DOSBox" | tail -1)
    xdotool windowfocus "$window_id"
    xdotool key --window "$window_id" Return
    sleep 12
    window_id=$(xdotool search --onlyvisible --name "DOSBox" | tail -1)
    xdotool windowfocus "$window_id"
    xdotool key --window "$window_id" Escape
    sleep 6

    if [[ "$ORACLE_MODE" == "continue" ]]; then
      xdotool mousemove --window "$window_id" 490 200 click 1
      sleep 8
    fi

    import -window "$window_id" /tmp/original-window.png
    # DOSBox-X 的 17px 選單列不是遊戲 framebuffer。
    convert /tmp/original-window.png -crop 640x480+0+17 +repage /out/original-640x480.png
    sha256sum /out/original-640x480.png | sed "s#  /out/#  #" > /out/SHA256SUMS
    geometry=$(identify -format "%wx%h" /out/original-640x480.png)
    dosbox-x --version >/tmp/dosbox-version.txt 2>&1 || true
    dosbox_version=$(sed -n "/DOSBox-X version/{p;q;}" /tmp/dosbox-version.txt)
    printf "image=%s\nmode=%s\nclient_crop=%s\ndosbox=%s\n" \
      "$geometry" "$ORACLE_MODE" "640x480+0+17" "$dosbox_version" \
      > /out/metadata.txt
    sha256sum /game/Orion2.exe 2>/dev/null >> /out/metadata.txt || \
      sha256sum /game/ORION2.EXE >> /out/metadata.txt
    if [[ -f /game/SAVE10.GAM ]]; then
      sha256sum /game/SAVE10.GAM >> /out/metadata.txt
    fi
  '

echo "原版 oracle 已輸出：$OUT_DIR/original-640x480.png"
