package services

import (
	"path/filepath"
	"strconv"
)

func ExecuteRife(ctx *ServiceContext, ratio int, skip bool, filename string, outputfilename string) error {
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

	// upsample input
	_, err = ctx.Environment.Execute(
		append([]string{}, "inference_video --exp="+strconv.Itoa(ratio)+" "+skipcmd+" --video=\""+fullfilename+"\" --output=\""+fulloutputfilename+"\""),
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
