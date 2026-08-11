#!/usr/bin/env bash
set -euo pipefail

# 在 game-video Docker image 內錄製「真實互動」的 MOO2 remake 推廣片。
# 外層必須以 --network none、有限資源、唯讀 AppImage／資料／字型掛載執行。
#
# 用法：capture_promo_gameplay.sh <AppImage> <遊戲資料目錄> <CJK 字型> <音樂檔或 -> <輸出 mp4>
#
# 錄影本身是 Xvfb 中的實際遊玩流程：依序開新局、選種族、進星圖、檢視殖民地／
# 科技／外交，再宣戰進入戰術戰鬥。遊戲用 -promo-demo 重播正常 UI 點擊；沒有把
# PNG 當作影片來源，也不使用 -gamegallery 的展示狀態注入。
# 無音訊裝置的容器用程式的 -noaudio 錄畫面；若提供音樂，才在合成階段鋪上
# 已獲發布權的音檔。傳入 - 會輸出無聲影片。

if [[ $# -ne 5 ]]; then
  echo "用法：$0 <AppImage> <遊戲資料目錄> <CJK 字型> <音樂檔或 -> <輸出 mp4>" >&2
  exit 2
fi

APPIMAGE=$1
DATA_DIR=$2
FONT_FILE=$3
MUSIC_FILE=$4
OUT_FILE=$5
WIDTH=1280
HEIGHT=960
FPS=30
CAPTURE_SECONDS=60

for required in "$APPIMAGE" "$FONT_FILE"; do
  [[ -f "$required" ]] || { echo "找不到必要檔案：$required" >&2; exit 1; }
done
[[ -d "$DATA_DIR" ]] || { echo "找不到遊戲資料目錄：$DATA_DIR" >&2; exit 1; }
if [[ "$MUSIC_FILE" != "-" ]]; then
  [[ -f "$MUSIC_FILE" ]] || { echo "找不到音樂檔：$MUSIC_FILE" >&2; exit 1; }
fi

mkdir -p "$(dirname "$OUT_FILE")"
WORK_DIR=$(mktemp -d /tmp/moo2-promo-gameplay-XXXXXX)
RAW_VIDEO="$WORK_DIR/live-gameplay-4x3.mp4"
GAME_LOG="$WORK_DIR/game.log"
XVFB_LOG="$WORK_DIR/xvfb.log"

cleanup() {
  if [[ -n "${CAPTURE_PID:-}" ]]; then kill "$CAPTURE_PID" 2>/dev/null || true; fi
  if [[ -n "${GAME_PID:-}" ]]; then kill "$GAME_PID" 2>/dev/null || true; fi
  if [[ -n "${XVFB_PID:-}" ]]; then kill "$XVFB_PID" 2>/dev/null || true; fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT INT TERM

cd "$WORK_DIR"
"$APPIMAGE" --appimage-extract >/dev/null 2>&1
[[ -x squashfs-root/AppRun ]] || { echo "AppImage 解開後沒有 AppRun" >&2; exit 1; }

export DISPLAY=:99
export LIBGL_ALWAYS_SOFTWARE=1
Xvfb :99 -screen 0 "${WIDTH}x${HEIGHT}x24" -nolisten tcp >"$XVFB_LOG" 2>&1 &
XVFB_PID=$!
sleep 1

# -noaudio 是錄製環境專用；合成配樂與遊戲內音訊解碼保持分開，避免無 ALSA 裝置使
# 實際遊玩畫面無法啟動。-promo-demo 只重播正常 UI 點擊，不會換入截圖廊專用狀態。
./squashfs-root/AppRun -game -data "$DATA_DIR" -font "$FONT_FILE" -uiscale 2 -promo-demo -noaudio >"$GAME_LOG" 2>&1 &
GAME_PID=$!

WINDOW_ID=""
for _ in $(seq 1 30); do
  WINDOW_ID=$(xdotool search --name "Master of Orion II" 2>/dev/null | head -n 1 || true)
  [[ -n "$WINDOW_ID" ]] && break
  kill -0 "$GAME_PID" 2>/dev/null || { cat "$GAME_LOG" >&2; exit 1; }
  sleep 0.5
done
[[ -n "$WINDOW_ID" ]] || { cat "$GAME_LOG" >&2; echo "找不到遊戲視窗" >&2; exit 1; }
xdotool windowfocus "$WINDOW_ID" 2>/dev/null || true

ffmpeg -y -loglevel error \
  -f x11grab -video_size "${WIDTH}x${HEIGHT}" -framerate "$FPS" -draw_mouse 0 -i :99.0 \
  -t "$CAPTURE_SECONDS" -threads 2 -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p -an \
  "$RAW_VIDEO" &
CAPTURE_PID=$!

wait "$CAPTURE_PID"
CAPTURE_PID=""

DURATION=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$RAW_VIDEO")
FADE_OUT=$(awk -v d="$DURATION" 'BEGIN { printf "%.3f", d-3 }')
VIDEO_FILTER="[0:v]scale=960:720:flags=lanczos,pad=1280:720:160:0:color=0x07111f,format=yuv420p[v]"

if [[ "$MUSIC_FILE" == "-" ]]; then
  ffmpeg -y -loglevel error -i "$RAW_VIDEO" -filter_complex "$VIDEO_FILTER" -map "[v]" \
    -threads 2 -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p -movflags +faststart "$OUT_FILE"
else
  ffmpeg -y -loglevel error -i "$RAW_VIDEO" -stream_loop -1 -i "$MUSIC_FILE" \
    -filter_complex "$VIDEO_FILTER;[1:a]aresample=48000,volume=0.72,atrim=0:${DURATION},afade=t=in:st=0:d=2,afade=t=out:st=${FADE_OUT}:d=3[a]" \
    -map "[v]" -map "[a]" -threads 2 -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p \
    -c:a aac -b:a 192k -ar 48000 -ac 2 -movflags +faststart "$OUT_FILE"
fi

ffprobe -v error -show_entries format=duration:stream=index,codec_type,width,height,r_frame_rate,sample_rate,channels \
  -of default=noprint_wrappers=1 "$OUT_FILE"
