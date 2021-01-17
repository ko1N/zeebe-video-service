# build
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go get
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o video-service.out

# deploy
FROM rife:latest
COPY --from=builder /app/cmd/video-service.out /usr/local/bin/video-service

# install deps
RUN apt-get update && apt-get -y install \
    bash git

# add entrypoint script
WORKDIR /app
ADD build/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["video-service", "--services", "rife"]

ENV NVIDIA_DRIVER_CAPABILITIES all