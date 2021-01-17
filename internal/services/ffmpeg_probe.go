package services

import (
	"encoding/json"
)

func ExecuteFFmpegProbe(ctx *ServiceContext, filename string) (*map[string]interface{}, error) {
	ctx.Tracker.Info("probing input file", "filename", filename)

	// probe inputs
	probeResult, err := ctx.Environment.Execute(
		append([]string{}, "ffprobe -v quiet -print_format json -show_format -show_streams -i "+filename), nil, nil)
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
