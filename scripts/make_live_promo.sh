#!/usr/bin/env bash
set -euo pipefail

# 將 scripts/capture_promo_gameplay.sh 錄下的連續實機畫面，後製成有片頭、章節識別與
# 片尾的推廣片。章節識別只使用 4:3 遊戲畫面外的側欄，不遮蓋遊戲 UI。
#
# 用法：make_live_promo.sh <實機 mp4> <音樂檔或 -> <一般 CJK 字型> <粗體 CJK 字型> <輸出 mp4>
#
# 若音樂傳入 -，輸出無聲影片。提供音樂時必須由呼叫者明確標示權利；目前專案的原版
# STREAM 音軌只能用於本機私有預覽，不得把成片直接上傳公開 Release。

if [[ $# -ne 5 ]]; then
  echo "用法：$0 <實機 mp4> <音樂檔或 -> <一般 CJK 字型> <粗體 CJK 字型> <輸出 mp4>" >&2
  exit 2
fi

LIVE_VIDEO=$1
MUSIC_FILE=$2
FONT_REGULAR=$3
FONT_BOLD=$4
OUT_FILE=$5
WIDTH=1280
HEIGHT=720
FPS=30
TITLE_SECONDS=4
OUTRO_SECONDS=5
MAX_LIVE_SECONDS=64

for required in "$LIVE_VIDEO" "$FONT_REGULAR" "$FONT_BOLD"; do
  [[ -f "$required" ]] || { echo "找不到必要檔案：$required" >&2; exit 1; }
done
if [[ "$MUSIC_FILE" != "-" ]]; then
  [[ -f "$MUSIC_FILE" ]] || { echo "找不到音樂檔：$MUSIC_FILE" >&2; exit 1; }
  [[ "${MOO2_PROMO_MUSIC_RIGHTS:-}" == "local-private-preview" ]] || {
    echo "拒絕合成：音樂輸入必須設定 MOO2_PROMO_MUSIC_RIGHTS=local-private-preview；此片不可公開散布。" >&2
    exit 1
  }
fi

mkdir -p "$(dirname "$OUT_FILE")"
WORK_DIR=$(mktemp -d /tmp/moo2-live-promo-XXXXXX)
trap 'rm -rf "$WORK_DIR"' EXIT INT TERM

BG="#07111f"
PANEL="#10263b"
GRID="#24425a"
CYAN="#78c9d8"
GOLD="#e5c85e"
TEXT="#edf3f4"
MUTED="#a9bdc9"

LIVE_SECONDS=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$LIVE_VIDEO")
USED_LIVE_SECONDS=$(awk -v d="$LIVE_SECONDS" -v max="$MAX_LIVE_SECONDS" 'BEGIN { if (d < max) printf "%.3f", d; else printf "%.3f", max }')
TOTAL_SECONDS=$(awk -v a="$TITLE_SECONDS" -v b="$USED_LIVE_SECONDS" -v c="$OUTRO_SECONDS" 'BEGIN { printf "%.3f", a+b+c }')

# 片頭底圖取自實機第一幕再大幅壓暗、模糊；它只作識別卡背景，正片仍是連續的實機錄影。
ffmpeg -y -loglevel error -ss 0.2 -i "$LIVE_VIDEO" -frames:v 1 "$WORK_DIR/live-first.png"
convert "$WORK_DIR/live-first.png" -resize "${WIDTH}x${HEIGHT}^" -gravity center -extent "${WIDTH}x${HEIGHT}" \
  -modulate 35,55,100 -blur 0x7 -fill '#07111fc0' -colorize 55% \
  -stroke "$GRID" -strokewidth 1 \
  -draw 'line 0,120 1280,120 line 0,600 1280,600 line 160,0 160,720 line 1120,0 1120,720' \
  -stroke "$CYAN" -strokewidth 3 -fill none \
  -draw 'line 76,76 206,76 line 76,76 76,146 line 1074,644 1204,644 line 1204,574 1204,644' \
  -font "$FONT_BOLD" -gravity center \
  -fill '#02070c' -stroke '#02070c' -strokewidth 5 -pointsize 82 -annotate +4-72 '銀河霸主 II' \
  -fill "$GOLD" -stroke none -pointsize 82 -annotate +0-76 '銀河霸主 II' \
  -font "$FONT_REGULAR" -fill "$CYAN" -pointsize 24 -annotate +0+4 'MASTER OF ORION II' \
  -fill "$TEXT" -pointsize 32 -annotate +0+62 '繁體中文 Go／Ebiten 重製' \
  -fill "$MUTED" -pointsize 22 -annotate +0+114 '探索・經營・外交・征服' \
  "$WORK_DIR/title.png"

convert -size "${WIDTH}x${HEIGHT}" "radial-gradient:#18364c-${BG}" \
  -stroke "$GRID" -strokewidth 1 \
  -draw 'line 0,120 1280,120 line 0,600 1280,600 line 160,0 160,720 line 1120,0 1120,720' \
  -stroke "$CYAN" -strokewidth 3 -fill none \
  -draw 'line 76,76 206,76 line 76,76 76,146 line 1074,644 1204,644 line 1204,574 1204,644' \
  -font "$FONT_BOLD" -gravity center -fill "$GOLD" -stroke none -pointsize 58 -annotate +0-92 '經典 4X・繁體中文重製' \
  -font "$FONT_REGULAR" -fill "$TEXT" -pointsize 31 -annotate +0-22 'Linux・Windows・macOS' \
  -fill "$CYAN" -pointsize 25 -annotate +0+43 '開源重製工程' \
  -fill "$MUTED" -pointsize 22 -annotate +0+94 'github.com/wicanr2/master-of-orion2-remake-cht' \
  "$WORK_DIR/outro.png"

make_badge() {
  local out=$1 side=$2 number=$3 line1=$4 line2=$5
  local x1=14 x2=146 gravity=northwest
  if [[ "$side" == "right" ]]; then
    x1=1134; x2=1266; gravity=northeast
  fi
  convert -size "${WIDTH}x${HEIGHT}" xc:none \
    -fill '#07111fe8' -stroke "$CYAN" -strokewidth 2 -draw "roundrectangle ${x1},28 ${x2},116 8,8" \
    -fill "$GOLD" -stroke none -draw "rectangle ${x1},28 ${x2},34" \
    -font "$FONT_REGULAR" -gravity "$gravity" -fill "$CYAN" -pointsize 14 \
    -annotate "$([[ "$side" == right ]] && echo +14+43 || echo +26+43)" "章節 ${number}" \
    -font "$FONT_BOLD" -fill "$TEXT" -pointsize 18 \
    -annotate "$([[ "$side" == right ]] && echo +14+66 || echo +26+66)" "$line1" \
    -font "$FONT_REGULAR" -fill "$MUTED" -pointsize 15 \
    -annotate "$([[ "$side" == right ]] && echo +14+91 || echo +26+91)" "$line2" \
    "$out"
}

make_badge "$WORK_DIR/chapter-1.png" left  01 '建立帝國' '種族與開局'
make_badge "$WORK_DIR/chapter-2.png" right 02 '經營殖民地' '人口與星圖'
make_badge "$WORK_DIR/chapter-3.png" left  03 '間諜外交' '行動與談判'
make_badge "$WORK_DIR/chapter-4.png" right 04 '艦隊戰術' '移動與射擊'
make_badge "$WORK_DIR/chapter-5.png" left  05 '戰果回寫' '帝國持續運轉'

ffmpeg -y -loglevel error -loop 1 -i "$WORK_DIR/title.png" \
  -vf "fps=${FPS},fade=t=in:st=0:d=0.5,fade=t=out:st=3.35:d=0.65,format=yuv420p" \
  -t "$TITLE_SECONDS" -threads 2 -c:v libx264 -preset veryfast -crf 19 -an "$WORK_DIR/title.mp4"

# 每個章節標籤只顯示 2.4 秒；交替左右側欄，保持遊戲畫面完整可讀。
ffmpeg -y -loglevel error -i "$LIVE_VIDEO" \
  -i "$WORK_DIR/chapter-1.png" -i "$WORK_DIR/chapter-2.png" -i "$WORK_DIR/chapter-3.png" \
  -i "$WORK_DIR/chapter-4.png" -i "$WORK_DIR/chapter-5.png" \
  -filter_complex "[0:v]trim=duration=${USED_LIVE_SECONDS},setpts=PTS-STARTPTS[base];\
[base][1:v]overlay=0:0:enable='between(t,0.4,2.8)'[v1];\
[v1][2:v]overlay=0:0:enable='between(t,11,13.4)'[v2];\
[v2][3:v]overlay=0:0:enable='between(t,20,22.4)'[v3];\
[v3][4:v]overlay=0:0:enable='between(t,28,30.4)'[v4];\
[v4][5:v]overlay=0:0:enable='between(t,52,54.4)',fade=t=out:st=$(awk -v d="$USED_LIVE_SECONDS" 'BEGIN { printf "%.3f", d-0.5 }'):d=0.5,format=yuv420p[v]" \
  -map '[v]' -an -threads 2 -c:v libx264 -preset veryfast -crf 19 -pix_fmt yuv420p "$WORK_DIR/live.mp4"

ffmpeg -y -loglevel error -loop 1 -i "$WORK_DIR/outro.png" \
  -vf "fps=${FPS},fade=t=in:st=0:d=0.55,fade=t=out:st=4.25:d=0.75,format=yuv420p" \
  -t "$OUTRO_SECONDS" -threads 2 -c:v libx264 -preset veryfast -crf 19 -an "$WORK_DIR/outro.mp4"

printf "file '%s'\nfile '%s'\nfile '%s'\n" "$WORK_DIR/title.mp4" "$WORK_DIR/live.mp4" "$WORK_DIR/outro.mp4" > "$WORK_DIR/concat.txt"
ffmpeg -y -loglevel error -f concat -safe 0 -i "$WORK_DIR/concat.txt" \
  -map 0:v:0 -map_chapters -1 -map_metadata -1 -threads 2 -c:v libx264 -preset veryfast -crf 19 \
  -pix_fmt yuv420p "$WORK_DIR/video.mp4"

if [[ "$MUSIC_FILE" == "-" ]]; then
  ffmpeg -y -loglevel error -i "$WORK_DIR/video.mp4" -map 0:v:0 -c:v copy -movflags +faststart "$OUT_FILE"
else
  FADE_OUT=$(awk -v d="$TOTAL_SECONDS" 'BEGIN { printf "%.3f", d-3 }')
  ffmpeg -y -loglevel error -i "$WORK_DIR/video.mp4" -stream_loop -1 -i "$MUSIC_FILE" \
    -filter_complex "[1:a]aresample=48000,volume=0.72,atrim=duration=${TOTAL_SECONDS},afade=t=in:st=0:d=2,afade=t=out:st=${FADE_OUT}:d=3[a]" \
    -map 0:v:0 -map '[a]' -map_chapters -1 -map_metadata -1 -c:v copy -c:a aac -b:a 192k -ar 48000 -ac 2 \
    -metadata title='銀河霸主 II 繁體中文重製實機推廣片' \
    -metadata comment='本機預覽；原版音樂不代表公開散布授權' -movflags +faststart "$OUT_FILE"
fi

ffprobe -v error -show_entries format=duration:stream=index,codec_type,width,height,r_frame_rate,sample_rate,channels \
  -of default=noprint_wrappers=1 "$OUT_FILE"
