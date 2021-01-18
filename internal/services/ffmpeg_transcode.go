package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/ko1N/zeebe-video-service/internal/config"
	"github.com/ko1N/zeebe-video-service/internal/environment"
)

// pipeline module for ffmpeg

// ExecuteFFmpeg -
func ExecuteFFmpegTranscode(ctx *ServiceContext, conf *config.FFmpegConfig, cmd string) (*environment.ExecutionResult, error) {
	ctx.Tracker.Info("probing input files")
	duration, err := estimateDuration(ctx, conf, cmd)
	if err != nil {
		ctx.Tracker.Crit("unable to estimate file duration")
		return nil, err
	}

	// run ffmpeg and track progress
	executable := "ffprobe"
	if conf != nil {
		executable = conf.FFmpegExecutable
	}
	ctx.Tracker.Info("executing ffmpeg", "cmd", cmd)
	result, err := ctx.Environment.Execute(
		append([]string{}, executable+" -v warning -progress /dev/stdout "+cmd),
		func(outmsg string) {
			//fmt.Println(outmsg)
			s := strings.Split(outmsg, "=")
			if len(s) == 2 && s[0] == "out_time_us" {
				time, err := strconv.Atoi(s[1])
				if err == nil {
					progress := float64(time) / (duration * 1000.0 * 1000.0)
					ctx.Tracker.Progress(progress)
				}
			}
		},
		func(errmsg string) {
			//fmt.Println(errmsg)
		})
	if err != nil {
		ctx.Tracker.Crit("execution of ffmpeg failed")
		return nil, err
	}

	if result.ExitCode == 0 {
		ctx.Tracker.Progress(1.0)
	} else {
		// TODO: handle error
		return nil, errors.New("unable to transcode video")
	}

	return &environment.ExecutionResult{}, nil
}

func estimateDuration(ctx *ServiceContext, conf *config.FFmpegConfig, cmd string) (float64, error) {
	// parse argument list and figure out the input file(s)
	var opts struct {
		Input string `short:"i" long:"input"`
		// TODO: handle shorted flag, -t, etc
	}
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(strings.Split(cmd, " "))
	if err != nil {
		ctx.Tracker.Crit("unable to parse input command line `" + cmd + "`")
		return 0, err
	}

	// probe inputs
	probe, err := ExecuteFFmpegProbe(ctx, conf, opts.Input)
	if err != nil {
		ctx.Tracker.Crit("unable to probe result", "error", err)
		return 0, err
	}

	format, ok := (*probe)["format"]
	if !ok {
		ctx.Tracker.Crit("could not locate `format`in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	durationStr, ok := format.(map[string]interface{})["duration"].(string)
	if !ok {
		ctx.Tracker.Crit("could not locate `dration`in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	duration, err := strconv.ParseFloat(durationStr, 32)
	if err != nil {
		ctx.Tracker.Crit("could not parse duration `" + durationStr + "` as number in ffprobe result")
		return 0, errors.New("unable to parse ffprobe result")
	}

	ctx.Tracker.Info("input file length", "file", opts.Input, "duration", duration)
	return duration, nil
}
