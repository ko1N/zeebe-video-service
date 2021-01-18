package services

import (
	"path/filepath"
	"strconv"

	"github.com/ko1N/zeebe-video-service/internal/config"
)

func ExecuteRife(ctx *ServiceContext, conf *config.RifeConfig, ratio int, uhd bool, skip bool, filename string, outputfilename string) error {
	fullfilename, err := filepath.Abs(ctx.Environment.FullPath(filename))
	if err != nil {
		ctx.Tracker.Crit("unable to get fullpath of file", "error", err)
		return err
	}

	fulloutputfilename, err := filepath.Abs(ctx.Environment.FullPath(outputfilename))
	if err != nil {
		ctx.Tracker.Crit("unable to get fullpath of file", "error", err)
		return err
	}

	ctx.Tracker.Info("rife files", "input", fullfilename, "output", fulloutputfilename)

	skipcmd := ""
	if skip {
		skipcmd = "--skip"
	}

	uhdcmd := ""
	if uhd {
		uhdcmd = "--UHD"
	}

	// upsample input
	executable := "python3 /rife/inference_video.py"
	if conf != nil {
		executable = conf.Executable
	}
	_, err = ctx.Environment.Execute(
		append([]string{}, executable+" --exp="+strconv.Itoa(ratio)+" "+uhdcmd+" "+skipcmd+" --video=\""+fullfilename+"\" --output=\""+fulloutputfilename+"\""),
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
