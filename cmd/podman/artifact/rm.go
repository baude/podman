package artifact

import (
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/spf13/cobra"
)

var (
	rmCmd = &cobra.Command{
		Use:     "remove [options] [ARTIFACT...]",
		Short:   "Remove an OCI artifact",
		Long:    "Remove an OCI from local storage",
		RunE:    rm,
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		Example: `podman artifact remove quay.io/myimage/myartifact:latest`,
	}
	rmFlag = rmFlagType{}
)

// TODO at some point force will be a required option; but this cannot be
// until we have artifacts being consumed by other parts of libpod like
// volumes
type rmFlagType struct {
	force bool
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: rmCmd,
		Parent:  artifactCmd,
	})

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func rm(cmd *cobra.Command, args []string) error {
	_, err := registry.ImageEngine().ArtifactRm(registry.GetContext(), args[0], entities.ArtifactRemoveOptions{})
	return err
}
