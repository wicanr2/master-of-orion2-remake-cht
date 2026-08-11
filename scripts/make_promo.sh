#!/usr/bin/env bash
set -euo pipefail

# Docker 內可重跑的 MOO2 remake 推廣片流程。
# 用法：make_promo.sh <截圖目錄> <音樂檔或 -> <輸出 mp4>
#
# 音樂檔應由使用者提供並確認發布權。傳入 - 時輸出無聲預覽，避免把
# 私有原版音訊誤帶進公開 Release。

if [[ $# -ne 3 ]]; then
  echo "用法：$0 <截圖目錄> <音樂檔或 -> <輸出 mp4>" >&2
  exit 2
fi

SHOT_DIR=$1
MUSIC=$2
OUT_FILE=$3
W=1280
H=720
FPS=30
BG_DEEP="#07111f"
BG_LITE="#123b55"
ACCENT="#d5a83a"
ACCENT_SHADOW="#6b4516"
TEXT="#f4ead2"
MUTED="#b6c7cf"
FONT_TITLE="/usr/share/fonts/opentype/noto/NotoSerifCJK-Bold.ttc"
FONT_BODY="/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc"

for required in "$FONT_TITLE" "$FONT_BODY"; do
  [[ -f "$required" ]] || { echo "找不到字型：$required" >&2; exit 1; }
done

for shot in \
  01_menu.png 02_raceselect.png 04_galaxy.png 07_fleet.png \
  10_colonyscreen.png 14_info_tech.png 15_diplomacy.png \
  16_tactical.png 17_groundcombat.png 20_bombing.png \
  23_multiplayer.png 24_hotseat.png 25_shipdesign.png 26_buildqueue.png \
  27_commandpoints.png 28_measure.png 33_netgames.png; do
  [[ -f "$SHOT_DIR/$shot" ]] || { echo "找不到截圖：$SHOT_DIR/$shot" >&2; exit 1; }
done

mkdir -p "$(dirname "$OUT_FILE")"
TMP=$(mktemp -d /tmp/moo2-promo-XXXXXX)
trap 'rm -rf "$TMP"' EXIT

card() {
  local out=$1 title=$2 subtitle=$3 detail=$4
  convert -size "${W}x${H}" "radial-gradient:#1b4862-${BG_DEEP}" \
    -font "$FONT_TITLE" -gravity center \
    -fill "$ACCENT_SHADOW" -pointsize 88 -annotate +4+4 "$title" \
    -fill "$ACCENT" -pointsize 88 -annotate +0+0 "$title" \
    -fill "$TEXT" -pointsize 50 -annotate +0+115 "$subtitle" \
    -font "$FONT_BODY" -fill "$MUTED" -pointsize 27 -annotate +0+190 "$detail" \
    "$out"
}

frame_slide() {
  local out=$1 shot=$2 caption=$3 label=$4
  convert -size "${W}x${H}" "gradient:${BG_LITE}-${BG_DEEP}" "$TMP/bg.png"
  convert "$SHOT_DIR/$shot" -resize 900x650\> -strip -bordercolor "$ACCENT" -border 4 "$TMP/frame.png"
  convert "$TMP/bg.png" "$TMP/frame.png" -gravity center -geometry +0-24 -composite \
    -fill "#06101ee6" -draw "rectangle 0,594 ${W},${H}" \
    -font "$FONT_BODY" -gravity southwest -fill "$ACCENT" -pointsize 21 -annotate +44+83 "$label" \
    -fill "$TEXT" -pointsize 34 -annotate +44+40 "$caption" \
    "$out"
}

full_slide() {
  local out=$1 shot=$2 caption=$3 label=$4
  convert "$SHOT_DIR/$shot" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" -strip "$TMP/full.png"
  convert -size "${W}x210" "gradient:transparent-${BG_DEEP}e8" "$TMP/veil.png"
  convert "$TMP/full.png" "$TMP/veil.png" -gravity south -composite \
    -font "$FONT_BODY" -gravity southwest -fill "$ACCENT" -pointsize 21 -annotate +44+94 "$label" \
    -fill "$TEXT" -pointsize 36 -annotate +44+48 "$caption" \
    "$out"
}

split_slide() {
  local out=$1 left=$2 right=$3 caption=$4
  convert -size "${W}x${H}" "gradient:#112d3d-${BG_DEEP}" "$TMP/split-bg.png"
  convert "$SHOT_DIR/$left" -resize 555x416\> -strip -bordercolor "$ACCENT" -border 4 "$TMP/left.png"
  convert "$SHOT_DIR/$right" -resize 555x416\> -strip -bordercolor "$ACCENT" -border 4 "$TMP/right.png"
  convert "$TMP/split-bg.png" \
    \( "$TMP/left.png" \) -gravity west -geometry +34+0 -composite \
    \( "$TMP/right.png" \) -gravity east -geometry +34+0 -composite \
    -fill "$ACCENT" -draw "rectangle 632,112 648,528" \
    -fill "#06101ee8" -draw "rectangle 0,594 ${W},${H}" \
    -font "$FONT_BODY" -gravity south -fill "$TEXT" -pointsize 34 -annotate +0+42 "$caption" \
    "$out"
}

quote_slide() {
  local out=$1 shot=$2 quote=$3 detail=$4
  convert "$SHOT_DIR/$shot" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" \
    -modulate 48,70,100 -blur 0x2 -strip "$TMP/quote-bg.png"
  convert "$TMP/quote-bg.png" \
    -fill "#07111fe8" -draw "rectangle 0,0 770,${H}" \
    -font "$FONT_TITLE" -gravity northwest -fill "$ACCENT" -pointsize 126 -annotate +68+92 "「" \
    -font "$FONT_BODY" -fill "$TEXT" -pointsize 48 -annotate +92+220 "$quote" \
    -fill "$MUTED" -pointsize 27 -annotate +94+420 "$detail" \
    "$out"
}

cta() {
  local out=$1
  convert -size "${W}x${H}" "radial-gradient:#23546b-${BG_DEEP}" \
    -font "$FONT_TITLE" -gravity center \
    -fill "$ACCENT_SHADOW" -pointsize 74 -annotate +3-70 "Master of Orion 2" \
    -fill "$ACCENT" -pointsize 74 -annotate +0-73 "Master of Orion 2" \
    -font "$FONT_BODY" -fill "$TEXT" -pointsize 47 -annotate +0+28 "繁體中文重製版" \
    -fill "$MUTED" -pointsize 28 -annotate +0+103 "Linux · Windows · macOS  |  開源 remake  |  立即加入銀河" \
    "$out"
}

# 敘事骨架：定位 → 世界 → 經營 → 外交 → 艦隊 → 戰鬥 → 多人 → 跨平台 → CTA。
# 版面刻意輪換：標題卡、框內、滿版、左右對照、大引號，避免 12 張同模板投影片。
card        "$TMP/00.png" "銀河霸主 2" "安塔瑞斯之戰" "經典 4X 策略遊戲的繁體中文 remake"
full_slide  "$TMP/01.png" "04_galaxy.png" "你的星圖，從第一個殖民地開始改寫" "探索／擴張"
frame_slide "$TMP/02.png" "02_raceselect.png" "選擇種族，讓每一個優勢成為你的戰略" "種族與開局"
frame_slide "$TMP/03.png" "10_colonyscreen.png" "殖民地、產能與士氣：每一回合都要取捨" "帝國經營"
split_slide  "$TMP/04.png" "14_info_tech.png" "27_commandpoints.png" "研究科技，將資源轉成艦隊優勢"
quote_slide  "$TMP/05.png" "15_diplomacy.png" "外交是另一座戰場" "談判、間諜與特殊行動，讓局勢在開火前改變"
split_slide  "$TMP/06.png" "25_shipdesign.png" "07_fleet.png" "自訂艦艇，組成真正屬於你的戰鬥編制"
full_slide   "$TMP/07.png" "16_tactical.png" "在戰術畫面親自下令，讓每一枚武器發揮作用" "艦隊戰"
split_slide  "$TMP/08.png" "17_groundcombat.png" "20_bombing.png" "地面戰與轟炸：勝負延伸到星球表面"
frame_slide  "$TMP/09.png" "23_multiplayer.png" "多人與熱座流程，和對手爭奪同一片星海" "多人遊戲"
full_slide   "$TMP/10.png" "01_menu.png" "繁體中文介面，並支援 Linux、Windows、macOS" "跨平台與中文化"
cta          "$TMP/11.png"

NAMES=(00 01 02 03 04 05 06 07 08 09 10 11)
DURS=(5 6 6 6 6 6 6 7 6 6 5 7)
LIST="$TMP/list.txt"
: > "$LIST"
for i in "${!NAMES[@]}"; do
  name=${NAMES[$i]}
  dur=${DURS[$i]}
  frames=$((FPS * dur))
  fade_out=$((dur - 1))
  ffmpeg -y -loglevel error -loop 1 -i "$TMP/$name.png" \
    -vf "fps=${FPS},fade=t=in:st=0:d=0.6,fade=t=out:st=${fade_out}:d=1,format=yuv420p" \
    -frames:v "$frames" -threads 2 -c:v libx264 -preset veryfast -crf 19 \
    -pix_fmt yuv420p -an "$TMP/$name.mp4"
  printf "file '%s'\n" "$TMP/$name.mp4" >> "$LIST"
done

ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" \
  -map_chapters -1 -threads 2 -c:v libx264 -preset veryfast -crf 19 -pix_fmt yuv420p "$TMP/video.mp4"
DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/video.mp4")

if [[ "$MUSIC" != "-" ]]; then
  [[ -f "$MUSIC" ]] || { echo "找不到音樂：$MUSIC" >&2; exit 1; }
  FADE_OUT=$(awk -v d="$DUR" 'BEGIN { printf "%.3f", d-3 }')
  ffmpeg -y -loglevel error -i "$TMP/video.mp4" -stream_loop -1 -i "$MUSIC" \
    -filter_complex "[1:a]aresample=48000,volume=0.78,afade=t=in:st=0:d=2,afade=t=out:st=${FADE_OUT}:d=3,atrim=duration=${DUR}[a]" \
    -map 0:v:0 -map "[a]" -map_chapters -1 -c:v copy -c:a aac -b:a 192k -ar 48000 -ac 2 \
    -metadata title="Master of Orion 2 remake — 繁體中文推廣片" \
    -metadata comment="本機預覽；音樂權利依輸入素材，不代表公開散布授權" \
    -movflags +faststart "$OUT_FILE"
else
  cp "$TMP/video.mp4" "$OUT_FILE"
fi

ffprobe -v error -show_entries format=duration:stream=index,codec_type,width,height,r_frame_rate,sample_rate,channels \
  -of default=noprint_wrappers=1 "$OUT_FILE"
