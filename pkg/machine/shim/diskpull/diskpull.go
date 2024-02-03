package diskpull

import (
	"context"
	"strings"

	"github.com/containers/podman/v4/pkg/machine/define"
	"github.com/containers/podman/v4/pkg/machine/ocipull"
	"github.com/containers/podman/v4/pkg/machine/stdpull"
)

func GetDisk(userInputPath string, dirs *define.MachineDirs, imagePath *define.VMFile, vmType define.VMType, name string) error {
	var (
		err    error
		mydisk ocipull.Disker
	)

	if userInputPath == "" {
		mydisk, err = ocipull.NewVersioned(context.Background(), dirs.DataDir, name, vmType.String(), imagePath)
	} else {
		if strings.HasPrefix(userInputPath, "http") {
			// TODO probably should use tempdir instead of datadir
			mydisk, err = stdpull.NewDiskFromURL(userInputPath, imagePath, dirs.DataDir, nil)
		} else {
			mydisk, err = stdpull.NewStdDiskPull(userInputPath, imagePath)
		}
	}
	if err != nil {
		return err
	}
	return mydisk.Get()
}
