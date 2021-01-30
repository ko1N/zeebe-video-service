package workers

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterFFmpegTranscodeWorker(client zbc.Client, conf *config.FFmpegConfig) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("ffmpeg-transcode-service").
		Handler(WorkerHandler(client, ffmpegTranscodeHandler(conf))).
		Timeout(24 * time.Hour).
		Concurrency(1).
		Open()
}

type FFMpegArgs struct {
	Source string
	Target string
}

func ffmpegTranscodeHandler(conf *config.FFmpegConfig) func(ctx *WorkerContext) error {
	return func(ctx *WorkerContext) error {
		source := ctx.Variables["source"]
		if source == "" {
			return fmt.Errorf("`source` variable must not be empty")
		}

		target := ctx.Variables["target"]
		if target == "" {
			return fmt.Errorf("`target` variable must not be empty")
		}

		sourceUrl, err := storage.ParseFileUrl(source.(string))
		if err != nil {
			return fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
		}

		targetUrl, err := storage.ParseFileUrl(target.(string))
		if err != nil {
			return fmt.Errorf("unable to parse url in `target` variable: %s", err.Error())
		}

		// add input + output
		err = ctx.FileSystem.AddInput(sourceUrl)
		if err != nil {
			return fmt.Errorf("unable to add input file '%s': %s", sourceUrl.URL.String(), err.Error())
		}

		err = ctx.FileSystem.AddOutput(targetUrl)
		if err != nil {
			return fmt.Errorf("unable to add output file '%s': %s", targetUrl.URL.String(), err.Error())
		}

		// ffmpeg
		argopts := FFMpegArgs{
			Source: sourceUrl.FilePath,
			Target: targetUrl.FilePath,
		}
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
		_, err = services.ExecuteFFmpegTranscode(ctx.ServiceContext, conf, args.String())
		if err != nil {
			return fmt.Errorf("ffmpeg failed: %s", err.Error())
		}

		ctx.Tracker.Info("ffmpeg-transcode successful")
		return nil
	}
}
