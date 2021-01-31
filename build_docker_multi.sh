#!/bin/bash

# setup docker buildx env
#docker buildx create --use

# build multi-arch containers
docker buildx build --platform linux/amd64,linux/arm64 -t files-service -f build/files-service.dockerfile .
docker buildx build --platform linux/amd64,linux/arm64 -t ffmpeg-service -f build/ffmpeg-service.dockerfile .
docker buildx build --platform linux/amd64,linux/arm64 -t video2x-service -f build/video2x-service.dockerfile .
docker buildx build --platform linux/amd64,linux/arm64 -t rife-service -f build/rife-service.dockerfile .
