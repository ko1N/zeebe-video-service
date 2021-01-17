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

func RegisterRifeWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("rife-service").
		Handler(WorkerHandler(client, rifeHandler)).
		Concurrency(1).
		Open()
}

func rifeHandler(ctx *WorkerContext) error {
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

	// rife
	err = services.ExecuteRife(ctx.ServiceContext, file)
	if err != nil {
		return fmt.Errorf("rife failed: %s", err.Error())
	}

	ctx.Tracker.Info("rife successful")
	return nil
}
