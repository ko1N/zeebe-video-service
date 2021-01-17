package services

import (
	"fmt"
	"path/filepath"
	"strconv"
)

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func ExecuteVideo2x(ctx *ServiceContext, driver string, ratio int, filename string, outputfilename string) error {
	if !contains([]string{"waifu2x_caffe", "waifu2x_converter_cpp", "waifu2x_ncnn_vulkan", "srmd_ncnn_vulkan", "realsr_ncnn_vulkan", "anime4kcpp"}, driver) {
		ctx.Tracker.Crit("invalid video2x driver", "driver", driver)
		return fmt.Errorf("invalid video2x driver %s", driver)
	}

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

	ctx.Tracker.Info("video2x files", "input", fullfilename, "output", fulloutputfilename)

	// upscale input
	_, err = ctx.Environment.Execute(
		append([]string{}, "python3.8 /video2x/src/video2x.py -d "+driver+" -r "+strconv.Itoa(ratio)+" -i \""+fullfilename+"\" -o \""+fulloutputfilename+"\""), nil, nil)
	if err != nil {
		ctx.Tracker.Crit("unable to execute video2x", "error", err)
		return err
	}

	return nil
}
