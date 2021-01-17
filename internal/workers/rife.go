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

func RegisterRifeWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("rife-service").
		Handler(WorkerHandler(client, rifeHandler)).
		Timeout(24 * time.Hour).
		Concurrency(1).
		Open()
}

func rifeHandler(ctx *WorkerContext) error {
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
	// exp (ratio): 1,2,3,...
	// skip: true/false (skip static frames)

	// parse arguments
	ratio, err := strconv.Atoi(ctx.Headers["ratio"])
	if err != nil {
		return fmt.Errorf("failed to convert 'ratio' header to integer: %s", err.Error())
	}
	if ratio < 2 {
		return fmt.Errorf("invalid 'ratio' header: ratio must be greater or equal to 2")
	}
	skip := false
	if ctx.Headers["skip"] == "true" {
		skip = true
	}
	ctx.Tracker.Info("rife settings", "ratio", ratio, "skip", skip)

	// rife - hardcoded output file name (see build/rife-service.dockerfile) for the corresponding hack!
	outfile := fmt.Sprintf("%s_upsampled%s", strings.TrimSuffix(file, filepath.Ext(file)), filepath.Ext(file))
	err = services.ExecuteRife(ctx.ServiceContext, ratio-1, skip, file, outfile)
	if err != nil {
		return fmt.Errorf("rife failed: %s", err.Error())
	}

	// upload file
	ctx.Tracker.Info("uploading to bucket", "bucket", bucket, "file", outfile)
	err = store.PutFile(bucket, outfile, outfile)
	if err != nil {
		return fmt.Errorf("failed to upload file to storage: %s", err.Error())
	}

	url.Path = path.Join(dir, outfile)
	ctx.Variables["output"] = url.String()
	ctx.Tracker.Info("rife successful", "output", ctx.Variables["output"])
	return nil
}
