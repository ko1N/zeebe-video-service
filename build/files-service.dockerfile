# build
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o video-service.out

# deploy
FROM alpine:latest
COPY --from=builder /app/cmd/video-service.out /usr/local/bin/video-service

# add entrypoint script
WORKDIR /app
ADD build/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["video-service", "--services", "files"]

ENV NVIDIA_DRIVER_CAPABILITIES all