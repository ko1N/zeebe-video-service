package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/workers"
)

func containsService(services []string, service string) bool {
	for _, s := range services {
		if s == service {
			return true
		}
	}
	return false
}

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

	var servicestr string
	flag.StringVar(&servicestr, "services", "files,ffmpeg,video2x,rife", "the services to activate")
	flag.Parse()
	services := strings.Split(servicestr, ",")

	handlers := []worker.JobWorker{}

	if containsService(services, "ffmpeg") {
		fmt.Println("adding ffmpeg services")
		handlers = append(handlers, workers.RegisterFFmpegProbeWorker(client))
		handlers = append(handlers, workers.RegisterFFmpegTranscodeWorker(client))
	}

	if containsService(services, "video2x") {
		fmt.Println("adding video2x service")
		handlers = append(handlers, workers.RegisterVideo2xWorker(client))
	}

	if containsService(services, "rife") {
		fmt.Println("adding rife service")
		handlers = append(handlers, workers.RegisterRifeWorker(client))
	}

	fmt.Println("workers started")
	for {
		time.Sleep(1 * time.Second)
	}
}
