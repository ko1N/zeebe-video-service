# Zeebe Video Service

## Local Development

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
docker run -it --gpus all -v /dev/dri:/dev/dri --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true ffmpeg-service
docker run -it --gpus all -v /dev/dri:/dev/dri --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true video2x-service
docker run -it --gpus all -v /dev/dri:/dev/dri --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true rife-service
```

The `--gpus all -v /dev/dri:/dev/dri` flag adds the host gpu to the container (if the driver is installed properly).

## License

Licensed under GPL3 License, see [LICENSE](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, shall be licensed as above, without any additional terms or conditions.
