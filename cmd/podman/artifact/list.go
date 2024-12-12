package artifact

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/pkg/domain/entities"
	units "github.com/docker/go-units"
	"github.com/spf13/cobra"
)

var (
	ListCmd = &cobra.Command{
		Use:               "list [options]",
		Aliases:           []string{"ls"},
		Short:             "List OCI artifacts",
		Long:              "List OCI artifacts in local store",
		RunE:              list,
		Args:              cobra.NoArgs,
		ValidArgsFunction: completion.AutocompleteNone,
		Example:           `podman artifact ls`,
	}
	ListFlag = inspectFlagType{}
)

type listFlagType struct {
	format string
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: ListCmd,
		Parent:  artifactCmd,
	})
	flags := ListCmd.Flags()
	formatFlagName := "format"
	flags.StringVar(&inspectFlag.format, formatFlagName, "", "Format volume output using JSON or a Go template")

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func list(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("format") {
		return fmt.Errorf("not implemented")
	}
	reports, err := registry.ImageEngine().ArtifactList(registry.GetContext(), entities.ArtifactListOptions{})
	if err != nil {
		return err
	}

	return writeTemplate(cmd, reports)
}

func writeTemplate(cmd *cobra.Command, reports []*entities.ArtifactListReport) error {
	fmt.Println("")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSIZE")
	for _, art := range reports {
		splitsees := strings.SplitN(art.List.Reference.StringWithinTransport(), ":", 2)
		fmt.Fprintf(tw, "%s\t\t%s\n", splitsees[1], units.HumanSize(float64(art.TotalSize())))
	}
	tw.Flush()
	fmt.Println("")
	return nil
}
