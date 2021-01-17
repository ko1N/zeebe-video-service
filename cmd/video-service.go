package main

import (
	"fmt"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/workers"
)

func main() {
	conf, err := config.ReadConfig("config.yml")
	if err != nil {
		panic(err)
	}

	fmt.Println("connecting to", conf.Zeebe.Host)
	client, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         conf.Zeebe.Host,
		UsePlaintextConnection: conf.Zeebe.Plaintext,
	})
	if err != nil {
		panic(err)
	}

	ffprobeWorker := workers.RegisterFFmpegProbeWorker(client)
	defer ffprobeWorker.AwaitClose()

	ffmpegWorker := workers.RegisterFFmpegTranscodeWorker(client)
	defer ffmpegWorker.AwaitClose()

	rifeWorker := workers.RegisterRifeWorker(client)
	defer rifeWorker.AwaitClose()

	fmt.Println("workers started")
	for {
		time.Sleep(1 * time.Second)
	}
}
