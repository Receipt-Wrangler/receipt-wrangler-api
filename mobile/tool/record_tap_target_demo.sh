#!/usr/bin/env bash
# Records the category/tag tap-target demo GIF. See mobile/CLAUDE.md ->
# "Recording the tap-target demo GIF" for prerequisites and the full story.
#
# Drives real mouse clicks (xdotool) against the demo binary running on a
# headless X server (Xvfb), screen-captures the result (ffmpeg) and converts it
# to a GIF. Usage: tool/record_tap_target_demo.sh [output.gif]
set -euo pipefail

cd "$(dirname "$0")/.."

OUT="${1:-tool/tap-target-demo.gif}"
DISPLAY_NUM="${TAP_DEMO_DISPLAY:-:99}"
RECTS="${TAP_DEMO_RECTS:-/tmp/tap_demo_rects.json}"
BIN="build/linux/x64/release/bundle/receipt_wrangler_mobile"
MP4=$(mktemp /tmp/tap_demo_XXXX.mp4)

[ -x "$BIN" ] || {
  echo "Missing $BIN -- run:" >&2
  echo "  flutter build linux --release --target=tool/tap_target_demo.dart" >&2
  exit 1
}

cleanup() {
  for pid in "${FFMPEG_PID:-}" "${APP_PID:-}" "${XVFB_PID:-}"; do
    [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null && kill "$pid" 2>/dev/null
  done
  rm -f "$MP4"
  return 0
}
trap cleanup EXIT

export DISPLAY="$DISPLAY_NUM"

# Refuse a display that is already served. Starting Xvfb on it fails, and every
# later step -- the demo, xdotool, ffmpeg -- would silently attach to that other
# session instead and record whatever it is showing. Checking after the fact
# cannot catch this: the readiness probe below would be answered by the very
# server we are trying to avoid.
if xdotool getdisplaygeometry >/dev/null 2>&1; then
  echo "$DISPLAY_NUM is already served by another X server." >&2
  echo "Free it (killall Xvfb) or point TAP_DEMO_DISPLAY at another display." >&2
  exit 1
fi

Xvfb "$DISPLAY_NUM" -screen 0 1280x720x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
XVFB_PID=$!

# Xvfb exits immediately when another X server already owns the display -- a
# stale server left by an aborted run is the usual cause. Without this check the
# script would sail on and record whatever that other session is showing.
xvfb_ready=false
for _ in $(seq 20); do
  if ! kill -0 "$XVFB_PID" 2>/dev/null; then
    echo "Xvfb exited -- is $DISPLAY_NUM already in use? /tmp/xvfb.log says:" >&2
    tail -3 /tmp/xvfb.log >&2
    echo "Free that display, or point TAP_DEMO_DISPLAY at another one." >&2
    exit 1
  fi
  # Waits for the server to actually answer, which a fixed sleep does not.
  xdotool getdisplaygeometry >/dev/null 2>&1 && { xvfb_ready=true; break; }
  sleep 0.25
done
"$xvfb_ready" || { echo "Xvfb never came up on $DISPLAY_NUM" >&2; exit 1; }

rm -f "$RECTS"
TAP_DEMO_RECTS="$RECTS" "$BIN" >/tmp/tap_demo_app.log 2>&1 &
APP_PID=$!

# The app publishes its field rects after first paint; that is also the signal
# that it is up and drawn.
for _ in $(seq 40); do [ -f "$RECTS" ] && break; sleep 0.5; done
[ -f "$RECTS" ] || { echo "demo app never published $RECTS" >&2; exit 1; }
sleep 1

# Park the cursor off the panels so it does not sit on a field in frame one.
xdotool mousemove 640 560

# Crop to the panels: the window is 1280x720 but the content ends around y=410.
ffmpeg -loglevel error -f x11grab -draw_mouse 1 -framerate 20 \
  -video_size 1280x720 -i "$DISPLAY_NUM" \
  -vf "crop=1280:412:0:0" -c:v libx264 -preset ultrafast -qp 0 -y "$MP4" &
FFMPEG_PID=$!
sleep 2

python3 tool/tap_target_click_plan.py "$RECTS"

sleep 2.5
kill -INT "$FFMPEG_PID"; wait "$FFMPEG_PID" 2>/dev/null || true

mkdir -p "$(dirname "$OUT")"
# Two-pass palette: a flat UI GIF bands badly on the default palette.
ffmpeg -loglevel error -i "$MP4" \
  -vf "fps=10,scale=960:-1:flags=lanczos,split[a][b];[a]palettegen=max_colors=128[p];[b][p]paletteuse=dither=bayer:bayer_scale=3" \
  -loop 0 -y "$OUT"

echo "wrote $OUT ($(du -h "$OUT" | cut -f1))"
