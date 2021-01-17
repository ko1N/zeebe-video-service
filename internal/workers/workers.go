package workers

import (
	"context"
	"fmt"
	"log"

	"github.com/ko1N/zeebe-video-service/internal/environment"
	"github.com/ko1N/zeebe-video-service/internal/services"
	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
)

type WorkerContext struct {
	// headers, variables, close handlers, etc
	Headers        map[string]string
	Variables      map[string]interface{}
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

		// create environment
		// TODO: configurable path + sandboxing :)
		env, err := environment.CreateNativeEnvironment()
		if err != nil {
			// failed to parse source url
			fmt.Println("Failed to setup environment")
			failJob(jobClient, job)
			return
		}
		defer env.Close()

		// create context
		tracker := services.NewTracker(client, job)
		serviceContext := services.NewServiceContext(env, tracker)

		workerContext := WorkerContext{
			Headers:        headers,
			Variables:      variables,
			Environment:    env,
			Tracker:        tracker,
			ServiceContext: serviceContext,
		}
		err = handler(&workerContext)
		if err != nil {
			fmt.Printf("Failed to execute job: %s\n", err.Error())
			failJob(jobClient, job)
			return
		}

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
