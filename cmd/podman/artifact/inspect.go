package artifact

import (
	"fmt"

	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/spf13/cobra"
)

var (
	inspectCmd = &cobra.Command{
		Use:               "inspect [options] [ARTIFACT...]",
		Short:             "Inspect an OCI artifact",
		Long:              "Provide details on an OCI artifact",
		RunE:              inspect,
		PersistentPreRunE: devOnly,
		Args:              cobra.MinimumNArgs(1),
		Example:           `podman artifact inspect quay.io/myimage/myartifact:latest`,
		// TODO Autocomplete function needs to be done
	}
	inspectFlag = inspectFlagType{}
)

type inspectFlagType struct {
	format string
	remote bool
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: inspectCmd,
		Parent:  artifactCmd,
	})
	flags := inspectCmd.Flags()
	formatFlagName := "format"
	flags.StringVar(&inspectFlag.format, formatFlagName, "", "Format volume output using JSON or a Go template")
	remoteFlagName := "remote"
	flags.BoolVar(&inspectFlag.remote, remoteFlagName, false, "Inspect the image on a container image registry")

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func inspect(cmd *cobra.Command, args []string) error {
	return errNotImplemented()
}

func devOnly(_ *cobra.Command, _ []string) error {
	fmt.Printf("\n** Artifacts are in development.  It can change at any time. **\n\n")
	return nil
}

func errNotImplemented() error {
	return fmt.Errorf("not implemented yet")
}
