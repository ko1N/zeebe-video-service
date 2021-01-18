#!/bin/sh

export ZEEBE_HOST="${ZEEBE_HOST:-0.0.0.0:26500}"

export ZEEBE_PLAINTEXT="${ZEEBE_PLAINTEXT:-false}"

cat << EOT > ./config.yml
zeebe:
  host: "$ZEEBE_HOST"
  plaintext: $ZEEBE_PLAINTEXT

ffmpeg:
  ffprobe: "/usr/bin/ffprobe"
  ffmpeg: "/usr/bin/ffmpeg"

video2x:
  executable: "python3.8 /video2x/src/video2x.py"

rife:
  executable: "python3 -m trace --trace /rife/inference_video.py"
EOT

exec "$@"
