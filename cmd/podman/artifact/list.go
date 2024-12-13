package artifact

import (
	"fmt"
	"github.com/containers/common/pkg/report"
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
	listFlag = listFlagType{}
)

type listFlagType struct {
	format string
}

type artifactListOutput struct {
	Name string
	Size int64
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: ListCmd,
		Parent:  artifactCmd,
	})
	flags := ListCmd.Flags()
	formatFlagName := "format"
	flags.StringVar(&listFlag.format, formatFlagName, "", "Format volume output using JSON or a Go template")

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func list(cmd *cobra.Command, args []string) error {
	reports, err := registry.ImageEngine().ArtifactList(registry.GetContext(), entities.ArtifactListOptions{})
	if err != nil {
		return err
	}

	return outputTemplate(cmd, reports)
}

func outputTemplate(cmd *cobra.Command, lrs []*entities.ArtifactListReport) error {
	var artifacts []artifactListOutput
	var err error
	for _, lr := range lrs {
		artifactName, err := lr.Artifact.Name()
		// if we want to implement the concept of None, the above func
		// call might be where to do it
		if err != nil {
			return err
		}
		artifacts = append(artifacts, artifactListOutput{
			Name: artifactName,
			Size: lr.Artifact.TotalSize(),
		})
	}
	for _, artifact := range artifacts {
		fmt.Println(artifact.Name)
	}
	headers := report.Headers(artifactListOutput{}, map[string]string{
		"Name": "NAME",
		"Size": "SIZE",
	})
	rpt := report.New(os.Stdout, cmd.Name())
	defer rpt.Flush()

	switch {
	case cmd.Flag("format").Changed:
		rpt, err = rpt.Parse(report.OriginPodman, listFlag.format)
	default:
		fmt.Println("<--")
		rpt, err = rpt.Parse(report.OriginUser, listFlag.format)
	}
	if err != nil {
		return err
	}

	//if rpt.RenderHeaders {
	fmt.Println("-->")
	if err := rpt.Execute(headers); err != nil {
		return fmt.Errorf("failed to write report column headers: %w", err)
	}
	//}
	return rpt.Execute(artifacts)
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
