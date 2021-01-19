# Zeebe Video Service

## Local Development

Starting the zeebe services:
```bash
cd deployments/development
docker-compose up
```

Building the containers locally
```bash
docker build -t ffmpeg-service -f build/ffmpeg-service.dockerfile .
docker build -t video2x-service -f build/video2x-service.dockerfile .
docker build -t rife-service -f build/rife-service.dockerfile .
```

Running the containers locally
```bash
docker run --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true ffmpeg-service
docker run --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true video2x-service
docker run --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true rife-service
```

Running containers with gpu support:
```bash
docker run -it --gpus all -v /dev/dri:/dev/dri -v $PWD/temp:/app/temp --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true ffmpeg-service
docker run -it --gpus all -v /dev/dri:/dev/dri -v $PWD/temp:/app/temp --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true video2x-service
docker run -it --gpus all -v /dev/dri:/dev/dri -v $PWD/temp:/app/temp --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true rife-service
```

It is recommended to mount a temp directory to /app/temp as the docker images by default can only scale up to 10gb in runtime size.
The `--gpus all -v /dev/dri:/dev/dri` flag adds the host gpu to the container (if the driver is installed properly).

Deploying workflows:
```bash
zbctl --insecure deploy workflows/ffmpeg-transcode-test.bpmn
zbctl --insecure deploy workflows/upscale-upsample-test.bpmn
```

Run the workflows (make sure the file exists on the minio instance):
```bash
zbctl --insecure create instance ffmpeg-transcode-test --variables "{\"filename\": \"minio://minio:miniominio@172.17.0.1:9000/test/test.mp4\"}"
zbctl --insecure create instance upscale-upsample-test --variables "{\"filename\": \"minio://minio:miniominio@172.17.0.1:9000/test/test.mp4\"}"
```

## License

Licensed under GPL3 License, see [LICENSE](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, shall be licensed as above, without any additional terms or conditions.
