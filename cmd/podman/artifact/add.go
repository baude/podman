package artifact

import (
	"fmt"

	"github.com/containers/podman/v5/pkg/domain/entities"

	"github.com/containers/podman/v5/cmd/podman/common"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:               "add [options] PATH ARTIFACT",
		Short:             "Add an OCI artifact to the local store",
		Long:              "Add an OCI artifact to the local store from the local filesystem",
		RunE:              add,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: common.AutocompleteArtifactAdd,
		Example:           `podman artifact add /tmp/foobar.txt quay.io/myimage/myartifact:latest`,
	}
	addFlag = addFlagType{}
)

// TODO at some point force will be a required option; but this cannot be
// until we have artifacts being consumed by other parts of libpod like
// volumes
type addFlagType struct {
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: addCmd,
		Parent:  artifactCmd,
	})

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func add(cmd *cobra.Command, args []string) error {
	//_, err := registry.ImageEngine().ArtifactRm(registry.GetContext(), args[0], entities.ArtifactRemoveOptions{})
	//return err
	report, err := registry.ImageEngine().ArtifactAdd(registry.Context(), args[0], args[1], entities.ArtifactAddoptions{})
	if err != nil {
		return err
	}
	fmt.Println(report.NewBlobDigest)
	return nil
}
