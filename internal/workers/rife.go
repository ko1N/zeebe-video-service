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

func RegisterRifeWorker(client zbc.Client, conf *config.RifeConfig) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("rife-service").
		Handler(WorkerHandler(client, rifeHandler(conf))).
		Timeout(24 * time.Hour).
		Concurrency(1).
		Open()
}

func rifeHandler(conf *config.RifeConfig) func(ctx *WorkerContext) error {
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
		ratio, err := strconv.Atoi(ctx.Headers["ratio"])
		if err != nil {
			return fmt.Errorf("failed to convert 'ratio' header to integer: %s", err.Error())
		}
		if ratio < 2 {
			return fmt.Errorf("invalid 'ratio' header: ratio must be greater or equal to 2")
		}
		uhd := false
		if ctx.Headers["uhd"] == "true" {
			uhd = true
		}
		skip := false
		if ctx.Headers["skip"] == "true" {
			skip = true
		}
		ctx.Tracker.Info("rife settings", "ratio", ratio, "uhd", uhd, "skip", skip)

		// rife
		err = services.ExecuteRife(ctx.ServiceContext, conf, ratio-1, uhd, skip, sourceUrl.FilePath, targetUrl.FilePath)
		if err != nil {
			return fmt.Errorf("rife failed: %s", err.Error())
		}

		ctx.Tracker.Info("rife successful", "output", ctx.Variables["output"])
		return nil
	}
}
