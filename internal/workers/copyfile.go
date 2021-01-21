package workers

import (
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/ko1N/zeebe-video-service/internal/storage"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

func RegisterCopyFileWorker(client zbc.Client) worker.JobWorker {
	return client.
		NewJobWorker().
		JobType("file-copy-service").
		Handler(WorkerHandler(client, copyFileHandler())).
		Timeout(12 * time.Hour).
		Concurrency(8).
		Open()
}

func copyFileHandler() func(ctx *WorkerContext) error {
	return func(ctx *WorkerContext) error {
		source := ctx.Variables["source"]
		if source == "" {
			return fmt.Errorf("`source` variable must not be empty")
		}

		dest := ctx.Variables["dest"]
		if dest == "" {
			return fmt.Errorf("`dest` variable must not be empty")
		}

		// download source file
		filename := ""
		{
			url, err := url.Parse(source.(string))
			if err != nil {
				return fmt.Errorf("unable to parse url in `source` variable: %s", err.Error())
			}

			ctx.Tracker.Info("connecting to storage at", "source", source)
			store, err := storage.ConnectStorage(ctx.Environment, url)
			if err != nil {
				return fmt.Errorf("failed to connect to storage: %s", err.Error())
			}
			defer store.Close()

			// download file
			_, filename = filepath.Split(url.Path)
			ctx.Tracker.Info("downloading from storage", "src", url.Path, "dest", filename)
			err = store.DownloadFile(url.Path, filename)
			if err != nil {
				return fmt.Errorf("failed to download file from storage: %s", err.Error())
			}
		}

		// upload file
		{
			url, err := url.Parse(dest.(string))
			if err != nil {
				return fmt.Errorf("unable to parse url in `dest` variable: %s", err.Error())
			}

			ctx.Tracker.Info("connecting to storage at", "source", source)
			store, err := storage.ConnectStorage(ctx.Environment, url)
			if err != nil {
				return fmt.Errorf("failed to connect to storage: %s", err.Error())
			}
			defer store.Close()

			// download file
			ctx.Tracker.Info("uploading to storage", "src", filename, "dest", url.Path)
			err = store.UploadFile(filename, url.Path)
			if err != nil {
				return fmt.Errorf("failed to download file from storage: %s", err.Error())
			}
		}

		//url.Path = path.Join(dirname, outfilename)
		//ctx.Variables["output"] = url.String()
		ctx.Tracker.Info("file copy successful", "dest", dest)
		return nil
	}
}
