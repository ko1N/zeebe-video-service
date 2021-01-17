package workers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterFFmpegTranscodeWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("ffmpeg-transcode-service").
		Handler(WorkerHandler(client, ffmpegTranscodeHandler)).
		Concurrency(1).
		Open()
}

type FFMpegArgs struct {
	Source string
}

func ffmpegTranscodeHandler(ctx *WorkerContext) error {
	source := ctx.Variables["source"]
	if source == "" {
		return fmt.Errorf("`source` variable must not be empty")
	}

	url, err := url.Parse(source.(string))
	if err != nil {
		return fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
	}

	ctx.Tracker.Info("connecting to storage at", "source", source)
	store, err := storage.ConnectStorage(ctx.Environment, url)
	if err != nil {
		return fmt.Errorf("failed to connect to storage: %s", err.Error())
	}

	dir, file := filepath.Split(url.Path)
	bucket := strings.TrimLeft(path.Clean(dir), "/")

	// download file
	ctx.Tracker.Info("downloading from bucket", "bucket", bucket, "file", file)
	err = store.GetFile(bucket, file, file)
	if err != nil {
		return fmt.Errorf("failed to download file from storage: %s", err.Error())
	}

	// ffmpeg
	argopts := FFMpegArgs{Source: file}
	argtpl, err := template.New("args").Parse(ctx.Headers["args"])
	if err != nil {
		return fmt.Errorf("invalid ffmpeg args")
	}

	var args bytes.Buffer
	err = argtpl.Execute(&args, argopts)
	if err != nil {
		return fmt.Errorf("malformed ffmpeg args")
	}

	// ffmpeg
	_, err = services.ExecuteFFmpegTranscode(ctx.ServiceContext, args.String())
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %s", err.Error())
	}

	ctx.Tracker.Info("ffmpeg-transcode successful")
	return nil
}
