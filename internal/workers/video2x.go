package workers

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

		url, err := url.Parse(source.(string))
		if err != nil {
			return fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
		}

		ctx.Tracker.Info("connecting to storage at", "source", source)
		store, err := storage.ConnectStorage(ctx.Environment, url)
		if err != nil {
			return fmt.Errorf("failed to connect to storage: %s", err.Error())
		}

		// download file
		dirname, filename := filepath.Split(url.Path)
		ctx.Tracker.Info("downloading from bucket", "file", url.Path)
		err = store.DownloadFile(url.Path, filename)
		if err != nil {
			return fmt.Errorf("failed to download file from storage: %s", err.Error())
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
		outfilename := fmt.Sprintf("%s_upscaled%s", strings.TrimSuffix(filename, filepath.Ext(filename)), filepath.Ext(filename))
		err = services.ExecuteVideo2x(ctx.ServiceContext, conf, driver, ratio, filename, outfilename)
		if err != nil {
			return fmt.Errorf("video2x failed: %s", err.Error())
		}

		// upload file
		ctx.Tracker.Info("uploading to bucket", "file", outfilename)
		err = store.UploadFile(outfilename, path.Join(dirname, outfilename))
		if err != nil {
			return fmt.Errorf("failed to upload file to storage: %s", err.Error())
		}

		url.Path = path.Join(dirname, outfilename)
		ctx.Variables["output"] = url.String()
		ctx.Tracker.Info("video2x successful", "output", ctx.Variables["output"])
		return nil
	}
}
