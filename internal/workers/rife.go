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
		outfilename := fmt.Sprintf("%s_upsampled%s", strings.TrimSuffix(filename, filepath.Ext(filename)), filepath.Ext(filename))
		err = services.ExecuteRife(ctx.ServiceContext, conf, ratio-1, uhd, skip, filename, outfilename)
		if err != nil {
			return fmt.Errorf("rife failed: %s", err.Error())
		}

		// upload file
		ctx.Tracker.Info("uploading to bucket", "file", outfilename)
		err = store.UploadFile(outfilename, path.Join(dirname, outfilename))
		if err != nil {
			return fmt.Errorf("failed to upload file to storage: %s", err.Error())
		}

		url.Path = path.Join(dirname, outfilename)
		ctx.Variables["output"] = url.String()
		ctx.Tracker.Info("rife successful", "output", ctx.Variables["output"])
		return nil
	}
}
