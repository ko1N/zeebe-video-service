# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o video-service.out

# build anime4kcpp
FROM k4yt3x/video2x:latest as cppbuilder
RUN git clone https://github.com/TianZerL/Anime4KCPP /tmp/anime4kcpp
WORKDIR /tmp/anime4kcpp
RUN cmake . && make

# deploy
FROM k4yt3x/video2x:latest
COPY --from=builder /app/cmd/video-service.out /usr/local/bin/video-service
COPY --from=cppbuilder /tmp/anime4kcpp/bin/Anime4KCPP_CLI /video2x/src/dependencies/anime4kcpp/anime4kcpp
COPY --from=cppbuilder /tmp/anime4kcpp/bin/libAnime4KCPPCore.so /video2x/src/dependencies/anime4kcpp/libAnime4KCPPCore.so

# add entrypoint script
WORKDIR /app
ADD build/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["video-service", "--services", "video2x"]

ENV NVIDIA_DRIVER_CAPABILITIES all