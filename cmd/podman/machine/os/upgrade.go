//go:build amd64 || arm64

package os

import (
	"github.com/containers/podman/v5/cmd/podman/common"
	"github.com/containers/podman/v5/cmd/podman/machine"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/cmd/podman/validate"
	"github.com/containers/podman/v5/pkg/machine/define"
	"github.com/containers/podman/v5/pkg/machine/os"
	provider2 "github.com/containers/podman/v5/pkg/machine/provider"
	"github.com/containers/podman/v5/version"
	"github.com/spf13/cobra"
)

var (
	upgradeCmd = &cobra.Command{
		Use:               "upgrade [options] IMAGE [NAME]",
		Short:             "Upgrade machine os",
		Long:              "Upgrade the machine operating system to a newer version",
		PersistentPreRunE: validate.NoOp,
		Args:              cobra.MaximumNArgs(1),
		RunE:              upgrade,
		ValidArgsFunction: common.AutocompleteImages,
		Example:           `podman machine os upgrade`,
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: upgradeCmd,
		Parent:  machine.OSCmd,
	})
	flags := upgradeCmd.Flags()

	restartFlagName := "restart"
	flags.BoolVar(&restart, restartFlagName, false, "Restart VM to apply changes")
}

func upgrade(_ *cobra.Command, args []string) error {
	vmName := define.DefaultMachineName
	if len(args) == 1 {
		vmName = args[0]
	}

	managerOpts := ManagerOpts{
		VMName:  vmName,
		CLIArgs: args,
		Restart: restart,
	}

	provider, err := provider2.Get()
	if err != nil {
		return err
	}
	osManager, err := NewOSManager(managerOpts, provider)
	if err != nil {
		return err
	}

	upgradeOpts := os.UpgradeOptions{HostVersion: version.Version}
	return osManager.Upgrade(vmName, upgradeOpts)
}
