package services

import "path/filepath"

func ExecuteRife(ctx *ServiceContext, filename string) error {
	fullfilename, err := filepath.Abs(ctx.Environment.FullPath(filename))
	if err != nil {
		ctx.Tracker.Crit("unable to get fullpath of file", "error", err)
		return err
	}

	ctx.Tracker.Info("rife input file", "fullfilename", fullfilename)

	// probe inputs
	_, err = ctx.Environment.Execute(
		append([]string{}, "rife --exp=1 --video=\""+fullfilename+"\""), nil, nil)
	if err != nil {
		ctx.Tracker.Crit("unable to execute rife", "error", err)
		return err
	}

	return nil
}
