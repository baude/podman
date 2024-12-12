package artifact

import (
	"fmt"
	"os"

	"github.com/containers/podman/v5/cmd/podman/common"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/spf13/cobra"
)

var (
	inspectCmd = &cobra.Command{
		Use:               "inspect [options] [ARTIFACT...]",
		Short:             "Inspect an OCI artifact",
		Long:              "Provide details on an OCI artifact",
		RunE:              inspect,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: common.AutocompleteArtifacts,
		Example:           `podman artifact inspect quay.io/myimage/myartifact:latest`,
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
	if inspectFlag.remote {
		return fmt.Errorf("not implemented")
	}

	if inspectFlag.format != "" {
		return fmt.Errorf("not implemented")
	}

	artifactOptions := entities.ArtifactInspectOptions{}
	inspectData, err := registry.ImageEngine().ArtifactInspect(registry.GetContext(), args[0], artifactOptions)
	if err != nil {
		return err
	}
	return printJSON(inspectData)
}

func printJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	// by default, json marshallers will force utf=8 from
	// a string.
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "     ")
	return enc.Encode(data)
}
