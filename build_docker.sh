#!/bin/bash
docker build -t ffmpeg-service -f build/ffmpeg-service.dockerfile .
docker build -t video2x-service -f build/video2x-service.dockerfile .
docker build -t rife-service -f build/rife-service.dockerfile .
