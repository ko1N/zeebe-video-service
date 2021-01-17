package workers

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterFFmpegProbeWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("ffmpeg-probe-service").
		Handler(WorkerHandler(client, ffmpegProbeHandler)).
		Concurrency(8).
		Open()
}

func ffmpegProbeHandler(ctx *WorkerContext) error {
	source := ctx.Variables["source"]
	url, err := url.Parse(source.(string))
	if err != nil {
		return fmt.Errorf("unable to get `source` variable: %s", err.Error())
	}

	ctx.Tracker.Info("connecting to storage at", "source", source)
	store, err := storage.ConnectStorage(ctx.Environment, url)
	if err != nil {
		return fmt.Errorf("failed to connect to storage: %s", err.Error())
	}

	dir, file := filepath.Split(url.Path)
	bucket := strings.TrimLeft(path.Clean(dir), "/")
	ctx.Tracker.Info("downloading from bucket", "bucket", bucket, "file", file)

	err = store.GetFile(bucket, file, file)
	if err != nil {
		return fmt.Errorf("failed to download file from storage: %s", err.Error())
	}

	// ffprobe
	probe, err := services.ExecuteFFmpegProbe(ctx.ServiceContext, file)
	if err != nil {
		return fmt.Errorf("ffprobe failed: %s", err.Error())
	}

	ctx.Tracker.Info("ffmpeg-probe successful")
	ctx.Variables["probe"] = probe
	return nil
}