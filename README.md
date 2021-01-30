# Zeebe Video Services

<p align="center">
<img src="docs/screenshot.png" alt="screenshot" title="screenshot" />
</p>

This project aims to provide services that allow for working with files (especially video files) in the [zeebe microservice orchestration workflow engine](https://github.com/zeebe-io/zeebe). The services are built to work in a serverless environment.

## Filestores

Filestores are backends that persistently store files. The file URL-Scheme let's the services decide what backend store to use.

File URL's are specified in the following format:
```
scheme://user:password@host/[storage]/[filepath]
```

Currently the following stores are implemented:

- `smb`

    Connects to a samba/cifs share.
    The `storage` path of the url must match the samba share on the server.
    Seeking is supported for both reading and writing files.

- `minio`

    Connects to a minio object store.
    The `storage` path of the url must match the bucket on the minio server.
    Seeking is only supported when reading files.

## Filesystems

The filesystem is the local storage for each service. When a file is accessed from one of the filestores it is made available through the selected filesystem.
The filesystem for each service can be customized directly in the BPMN. 

Simply set the `filesystem` Header to one of the available filesystems:

- `disk`

    The disk filesystem can be used when seek support is crucial for reading and writing files. This filesystem simply copies all files from the storage into a local temp folder and also uploads processed files directly from the filesystem.

- `virtual`

    The virtual filesystem utilizes fuse to create virtual files for reading and writing. Depending on the backend file store in use it can support file seeking. If a tool tries to seek while reading/writing a file with a backend store that doesn't support it, it will currently return an error.

## Available Services

### file-copy-service

The file-copy-service service will copy a file from one storage to another. It directly streams the files from one store to another without storing a local copy. The service simply acts as a proxy.

Inputs:
- source - path to the file (as uri)
- target - the output file (as uri)

### ffmpeg-probe-service

The probe service will run `ffprobe` on a file and return a json encoded result.

Inputs:
- source - path to the file (as uri)

Outputs:
- probe - the metadata object

### ffmpeg-transcode-service

The transcode service will run ``ffmpeg` on a file and also tracks the transcoding progress.

Inputs:
- source - path to the file (as uri)
- target - the output file (as uri)

Outputs:
- none

Headers:
- args
    - An argument template that is passed to ffmpeg. The following variables can be used to replace parts of the string:
        
        {{.Source}} - The source input file
        {{.Target}} - The target input file

### video2x-service

The video2x-service can upscale videos using several different backends. More information about the video2x project can be found [here](https://github.com/k4yt3x/video2x).

Inputs:
- source - path to the file (as uri)
- target - the output file (as uri)

Headers:
- driver
    - The upscale backend to use. Valid values are:
        - waifu2x_caffe
        - waifu2x_converter_cpp
        - waifu2x_ncnn_vulkan
        - srmd_ncnn_vulkan
        - realsr_ncnn_vulkan
        - anime4kcpp
- ratio - the ratio by which the image will be upscaled

### rife-service

The rife-service can generate intermediate frames for a video to artifically increase the framerate or generate super-slowmotion videos. More information about the RIFE project can be found [here](https://github.com/hzwer/arXiv2020-RIFE).

Inputs:
- source - path to the file (as uri)
- target - the output file (as uri)

Headers:
- ratio - the ratio by which the number of frames will be increased
- uhd - enables ultra high definition for 4k videos.
- skip - skip duplicate frames (does not keep audio)

## Local Development

Starting the zeebe services:
```bash
cd deployments/development
docker-compose up
```

Building the containers locally
```bash
docker build -t files-service -f build/files-service.dockerfile .
docker build -t ffmpeg-service -f build/ffmpeg-service.dockerfile .
docker build -t video2x-service -f build/video2x-service.dockerfile .
docker build -t rife-service -f build/rife-service.dockerfile .
```

Running the containers locally
```bash
docker run --env ZEEBE_HOST="172.17.0.1:26500" --env ZEEBE_PLAINTEXT=true files-service
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
