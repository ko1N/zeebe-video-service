package services

import (
	"strconv"
	"strings"

	"github.com/ko1N/zeebe-video-service/internal/config"
)

func ExecuteRife(ctx *ServiceContext, conf *config.RifeConfig, ratio int, uhd bool, skip bool, filename string, outputfilename string) error {
	ctx.Tracker.Info("rife files", "input", filename, "output", outputfilename)

	// upsample input
	executable := "python3 /rife/inference_video.py"
	if conf != nil {
		executable = conf.Executable
	}
	cmdline := strings.Split(executable, " ")

	args := []string{"--exp", strconv.Itoa(ratio), "--video", filename, "--output", outputfilename}
	if skip {
		args = append(args, "--skip")
	}
	if uhd {
		args = append(args, "--UHD")
	}

	_, err := ctx.Environment.Execute(
		cmdline[0], append(cmdline[1:], args...),
		func(outmsg string) {
			ctx.Tracker.Info(outmsg, "stream", "stdout")
		},
		func(errmsg string) {
			ctx.Tracker.Info(errmsg, "stream", "stderr")
		})
	if err != nil {
		ctx.Tracker.Crit("unable to execute rife", "error", err)
		return err
	}

	return nil
}
