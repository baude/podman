//go:build amd64 || arm64

package machine

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/libpod/events"
	"github.com/containers/podman/v5/pkg/machine"
	"github.com/containers/podman/v5/pkg/machine/shim"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.podman.io/common/pkg/config"
)

var (
	startCmd = &cobra.Command{
		Use:               "start [options] [MACHINE]",
		Short:             "Start an existing machine",
		Long:              "Start a managed virtual machine ",
		PersistentPreRunE: machinePreRunE,
		RunE:              start,
		Args:              cobra.MaximumNArgs(1),
		Example:           `podman machine start podman-machine-default`,
		ValidArgsFunction: autocompleteMachine,
	}
	startOpts            = machine.StartOptions{}
	setDefaultSystemConn bool
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: startCmd,
		Parent:  machineCmd,
	})

	flags := startCmd.Flags()
	noInfoFlagName := "no-info"
	flags.BoolVar(&startOpts.NoInfo, noInfoFlagName, false, "Suppress informational tips")

	quietFlagName := "quiet"
	flags.BoolVarP(&startOpts.Quiet, quietFlagName, "q", false, "Suppress machine starting status output")

	setDefaultConnectionFlagName := "update-connection"
	flags.BoolVarP(&setDefaultSystemConn, setDefaultConnectionFlagName, "u", false, "Set default system connection for this machine")
}

func start(cmd *cobra.Command, args []string) error {
	startOpts.NoInfo = startOpts.Quiet || startOpts.NoInfo

	vmName := defaultMachineName
	if len(args) > 0 && len(args[0]) > 0 {
		vmName = args[0]
	}

	mc, vmProvider, err := shim.VMExists(vmName)
	if err != nil {
		return err
	}

	if !startOpts.Quiet {
		fmt.Printf("Starting machine %q\n", vmName)
	}

	if err := shim.Start(mc, vmProvider, startOpts); err != nil {
		return err
	}
	fmt.Printf("Machine %q started successfully\n", vmName)
	newMachineEvent(events.Start, events.Event{Name: vmName})
	return checkAndSetDefConnection(cmd, vmName, mc.HostUser.Rootful, setDefaultSystemConn)
}

func checkAndSetDefConnection(cmd *cobra.Command, machineName string, isRootful bool, shouldUpdate bool) error {
	// shouldUpdate is the value from the CLI option "update-connection"
	if cmd.Flags().Changed("update-connection") {
		// the user has explicitly opted to NOT update anything
		// i.e. "--update-connection=false"
		if !shouldUpdate {
			return nil
		}
	}
	// this needs to determine rootful and rootless
	if isRootful {
		machineName += "-root"
	}
	conn, err := registry.PodmanConfig().ContainersConfDefaultsRO.GetConnection(machineName, false)
	if err != nil {
		return err
	}
	// early return, nothing to do here.
	if conn.Default {
		return nil
	}
	// Prompt the user if they would like to switch
	if !shouldUpdate {
		fmt.Print("\nWarning: The machine being started is not set as your default Podman connection.\n")
		fmt.Printf("As such, Podman commands may not work correctly.\n")
		fmt.Printf("Set the default Podman connection to this machine? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.ToLower(answer)[0] != 'y' {
			return nil
		}
	}
	return config.EditConnectionConfig(func(cfg *config.ConnectionsFile) error {
		// In the cmd/podman/system/default.go code, there is a check to verify the
		// connection exists prior to attempting to set it as the default.  I do not
		// think we need that as we have just checked for the connection earlier.
		logrus.Infof("Setting default Podman connection to %s", machineName)
		cfg.Connection.Default = machineName
		return nil
	})
}
