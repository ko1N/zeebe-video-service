package workers

import (
	"io"
	"fmt"
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

		// source store
		ctx.Tracker.Info("connecting to source storage at", "source", source)
		sourceStore, err := storage.ConnectStorage(sourceUrl)
		if err != nil {
			return fmt.Errorf("failed to connect to source storage: %s", err.Error())
		}
		defer sourceStore.Close()

		// target store
		ctx.Tracker.Info("connecting to target storage at", "target", target)
		targetStore, err := storage.ConnectStorage(targetUrl)
		if err != nil {
			return fmt.Errorf("failed to connect to target storage: %s", err.Error())
		}
		defer targetStore.Close()

		reader, err := sourceStore.GetFileReader(sourceUrl)
		if err != nil {
			return fmt.Errorf("failed to get reader for source storage: %s", err.Error())
		}
		defer reader.Close()

		writer, err := targetStore.GetFileWriter(targetUrl)
		if err != nil {
			return fmt.Errorf("failed to get writer for target storage: %s", err.Error())
		}
		defer writer.Close()

		_, err = io.Copy(writer, reader)
		if err != nil {
			return fmt.Errorf("failed to copy between storages: %s", err.Error())
		}

		ctx.Tracker.Info("file copy successful")
		return nil
	}
}
