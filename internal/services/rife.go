package services

import (
	"fmt"
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
	ctx.Tracker.Info("rife frame skipping", "skip", skip)

	// upsample input
	_, err = ctx.Environment.Execute(
		append([]string{}, "inference_video --exp="+strconv.Itoa(ratio)+" "+skipcmd+" --video=\""+fullfilename+"\" --output=\""+fulloutputfilename+"\""),
		func(outmsg string) {
			fmt.Println("rife stdout:", outmsg)
		},
		func(errmsg string) {
			fmt.Println("rife stderr:", errmsg)
		})
	if err != nil {
		ctx.Tracker.Crit("unable to execute rife", "error", err)
		return err
	}

	return nil
}
