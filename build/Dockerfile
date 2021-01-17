# build
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go get
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o video-service.out

# deploy
FROM python:3.8-slim
COPY --from=builder /app/cmd/video-service.out /usr/local/bin/video-service

# install deps
RUN apt-get update && apt-get -y install \
    git ffmpeg

# setup RIFE
WORKDIR /app
RUN git clone https://github.com/hzwer/arXiv2020-RIFE rife
WORKDIR /app/rife
RUN pip3 install -r requirements.txt

ADD build/rife.sh /usr/local/bin/rife
RUN chmod +x /usr/local/bin/rife

# add entrypoint script
WORKDIR /app
ADD build/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["video-service"]
