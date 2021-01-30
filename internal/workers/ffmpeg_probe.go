package workers

import (
	"fmt"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterFFmpegProbeWorker(client zbc.Client, conf *config.FFmpegConfig) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("ffmpeg-probe-service").
		Handler(WorkerHandler(client, ffmpegProbeHandler(conf))).
		Timeout(1 * time.Hour).
		Concurrency(8).
		Open()
}

func ffmpegProbeHandler(conf *config.FFmpegConfig) func(ctx *WorkerContext) error {
	return func(ctx *WorkerContext) error {
		source := ctx.Variables["source"]
		if source == "" {
			return fmt.Errorf("`source` variable must not be empty")
		}

		url, err := storage.ParseFileUrl(source.(string))
		if err != nil {
			return fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
		}

		err = ctx.FileSystem.AddInput(url)
		if err != nil {
			return fmt.Errorf("unable to add input file '%s': %s", url.URL.String(), err.Error())
		}

		//filesystem, _ := storage.CreateVirtualFS()
		//filesystem, _ := filesystem.CreateDiskFS()
		//filesystem.AddInput(source.(string))
		//filesystem.AddOutput("minio://minio:miniominio@172.17.0.1:9000/test/test2file.txt")
		//defer filesystem.Close()
		// ............

		// ffprobe
		probe, err := services.ExecuteFFmpegProbe(ctx.ServiceContext, conf, url.FilePath)
		if err != nil {
			return fmt.Errorf("ffprobe failed: %s", err.Error())
		}

		ctx.Tracker.Info("ffmpeg-probe successful")
		ctx.Variables["probe"] = probe
		return nil
	}
}
