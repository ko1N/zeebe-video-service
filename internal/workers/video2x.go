package workers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterVideo2xWorker(client zbc.Client, conf *config.Video2xConfig) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("video2x-service").
		Handler(WorkerHandler(client, video2xHandler(conf))).
		Timeout(24 * time.Hour).
		Concurrency(1).
		Open()
}

func video2xHandler(conf *config.Video2xConfig) func(ctx *WorkerContext) error {
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

		// parse arguments
		driver := ctx.Headers["driver"]
		if driver == "" {
			driver = "anime4kcpp"
		}

		ratio, err := strconv.Atoi(ctx.Headers["ratio"])
		if err != nil {
			return fmt.Errorf("failed to convert 'ratio' header to integer: %s", err.Error())
		}
		if ratio < 2 {
			return fmt.Errorf("invalid 'ratio' header: ratio must be greater or equal to 2")
		}
		ctx.Tracker.Info("video2x settings", "driver", driver, "ratio", ratio)

		// run video2x
		err = services.ExecuteVideo2x(ctx.ServiceContext, conf, driver, ratio, sourceUrl.FilePath, targetUrl.FilePath)
		if err != nil {
			return fmt.Errorf("video2x failed: %s", err.Error())
		}

		ctx.Tracker.Info("video2x successful", "output", ctx.Variables["output"])
		return nil
	}
}
