package services

import (
	"encoding/json"

	"github.com/ko1N/zeebe-video-service/internal/config"
)

func ExecuteFFmpegProbe(ctx *ServiceContext, conf *config.FFmpegConfig, filename string) (*map[string]interface{}, error) {
	ctx.Tracker.Info("probing input file", "filename", filename)

	// probe inputs
	executable := "ffprobe"
	if conf != nil {
		executable = conf.FFprobeExecutable
	}
	probeResult, err := ctx.Environment.Execute(
		append([]string{}, executable+" -v quiet -print_format json -show_format -show_streams -i "+filename), nil, nil)
	if err != nil {
		ctx.Tracker.Crit("unable to execute ffprobe", "error", err)
		return nil, err
	}

	var probe interface{}
	err = json.Unmarshal([]byte(probeResult.StdOut), &probe)
	if err != nil {
		ctx.Tracker.Crit("unable to unmarshal ffprobe result")
		return nil, err
	}

	probeMap := probe.(map[string]interface{})
	return &probeMap, nil
}
