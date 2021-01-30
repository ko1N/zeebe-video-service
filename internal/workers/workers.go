package workers

import (
	"context"
	"log"

	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"

	"github.com/ko1N/zeebe-video-service/internal/environment"
	"github.com/ko1N/zeebe-video-service/internal/environment/filesystem"
	"github.com/ko1N/zeebe-video-service/internal/services"
)

type WorkerContext struct {
	// headers, variables, close handlers, etc
	Headers        map[string]string
	Variables      map[string]interface{}
	FileSystem     filesystem.FileSystem
	Environment    environment.Environment
	Tracker        *services.Tracker
	ServiceContext *services.ServiceContext
}

func WorkerHandler(client zbc.Client, handler func(ctx *WorkerContext) error) func(worker.JobClient, entities.Job) {
	return func(jobClient worker.JobClient, job entities.Job) {
		// read headers
		headers, err := job.GetCustomHeadersAsMap()
		if err != nil {
			// failed to handle job as we require the custom job headers
			failJob(jobClient, job)
			return
		}

		// read variables
		variables, err := job.GetVariablesAsMap()
		if err != nil {
			// failed to handle job as we require the variables
			failJob(jobClient, job)
			return
		}

		// create tracker
		tracker := services.NewTracker(client, job)

		var fs filesystem.FileSystem
		{
			fsConf := "disk"
			if fileSystem, ok := headers["filesystem"]; ok {
				fsConf = fileSystem
			}

			tracker.Info("loading filesystem", "fs", fsConf)

			// TODO: configurable path
			switch fsConf {
			case "virtual", "fuse":

				fs, err = filesystem.CreateVirtualFS()
				break

			//case "disk":
			default:
				fs, err = filesystem.CreateDiskFS()
				break
			}

			if err != nil {
				tracker.Crit("failure loading filesystem", "fs", fsConf, "err", err)
				failJob(jobClient, job)
				return
			}
		}
		defer fs.Close()

		// create environment
		env, err := environment.CreateNativeEnvironment(fs)
		if err != nil {
			tracker.Crit("failure loading native environment", "err", err)
			failJob(jobClient, job)
			return
		}
		defer env.Close()

		// create context
		serviceContext := services.NewServiceContext(env, tracker)

		workerContext := WorkerContext{
			Headers:        headers,
			Variables:      variables,
			FileSystem:     fs,
			Environment:    env,
			Tracker:        tracker,
			ServiceContext: serviceContext,
		}
		err = handler(&workerContext)
		if err != nil {
			tracker.Crit("job failed with error", "error", err)
			failJob(jobClient, job)
			return
		}

		// flush all filesystem operations
		fs.Flush()

		request, err := client.
			NewCompleteJobCommand().
			JobKey(job.GetKey()).
			VariablesFromMap(workerContext.Variables)
		if err != nil {
			// failed to set the updated variables
			failJob(jobClient, job)
			return
		}

		ctx := context.Background()
		_, err = request.Send(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func failJob(client worker.JobClient, job entities.Job) {
	log.Println("Failed to complete job", job.GetKey())

	ctx := context.Background()
	_, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
	if err != nil {
		panic(err)
	}
}
