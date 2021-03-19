package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ko1N/zeebe-video-service/internal/config"
)

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func ExecuteVideo2x(ctx *ServiceContext, conf *config.Video2xConfig, driver string, ratio int, filename string, outputfilename string) error {
	if !contains([]string{"waifu2x_caffe", "waifu2x_converter_cpp", "waifu2x_ncnn_vulkan", "srmd_ncnn_vulkan", "realsr_ncnn_vulkan", "anime4kcpp"}, driver) {
		ctx.Tracker.Crit("invalid video2x driver", "driver", driver)
		return fmt.Errorf("invalid video2x driver %s", driver)
	}

	ctx.Tracker.Info("video2x files", "input", filename, "output", outputfilename)

	// upscale input
	executable := "python3.8 /video2x/src/video2x.py"
	if conf != nil {
		executable = conf.Executable
	}
	cmdline := strings.Split(executable, " ")

	_, err := ctx.Environment.Execute(
		cmdline[0], append(cmdline[1:], []string{"-d", driver, "-r", strconv.Itoa(ratio), "-i", filename, "-o", outputfilename}...),
		func(outmsg string) {
			ctx.Tracker.Info(outmsg, "stream", "stdout")
		},
		func(errmsg string) {
			ctx.Tracker.Info(errmsg, "stream", "stderr")
		})
	if err != nil {
		ctx.Tracker.Crit("unable to execute video2x", "error", err)
		return err
	}

	return nil
}
