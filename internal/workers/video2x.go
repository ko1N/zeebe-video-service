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

	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/ko1N/zeebe-video-service/internal/storage"
)

func RegisterVideo2xWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("video2x-service").
		Handler(WorkerHandler(client, video2xHandler)).
		Timeout(24 * time.Hour).
		Concurrency(1).
		Open()
}

func video2xHandler(ctx *WorkerContext) error {
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

	// Options:
	// driver: ...
	// ratio: 2,3,4,...

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
	outfile := fmt.Sprintf("%s_upscaled%s", strings.TrimSuffix(file, filepath.Ext(file)), filepath.Ext(file))
	err = services.ExecuteVideo2x(ctx.ServiceContext, driver, ratio, file, outfile)
	if err != nil {
		return fmt.Errorf("video2x failed: %s", err.Error())
	}

	// upload file
	ctx.Tracker.Info("uploading to bucket", "bucket", bucket, "file", outfile)
	err = store.PutFile(bucket, outfile, outfile)
	if err != nil {
		return fmt.Errorf("failed to upload file to storage: %s", err.Error())
	}

	url.Path = path.Join(dir, outfile)
	ctx.Variables["output"] = url.String()
	ctx.Tracker.Info("video2x successful", "output", ctx.Variables["output"])
	return nil
}
