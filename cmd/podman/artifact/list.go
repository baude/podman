package artifact

import (
	"fmt"
	"os"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/common/pkg/report"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/podman/v5/cmd/podman/common"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/cmd/podman/validate"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

var (
	ListCmd = &cobra.Command{
		Use:               "ls [options]",
		Aliases:           []string{"list"},
		Short:             "List OCI artifacts",
		Long:              "List OCI artifacts in local store",
		RunE:              list,
		Args:              validate.NoArgs,
		ValidArgsFunction: completion.AutocompleteNone,
		Example:           `podman artifact ls`,
	}
	listFlag = listFlagType{}
)

type listFlagType struct {
	format string
}

type artifactListOutput struct {
	Digest     string
	Repository string
	Size       string
	Tag        string
}

var (
	defaultArtifactListOutputFormat = "{{range .}}{{.Repository}}\t{{.Tag}}\t{{.Digest}}\t{{.Size}}\n{{end -}}"
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: ListCmd,
		Parent:  artifactCmd,
	})
	flags := ListCmd.Flags()
	formatFlagName := "format"
	flags.StringVar(&listFlag.format, formatFlagName, defaultArtifactListOutputFormat, "Format volume output using JSON or a Go template")
	_ = ListCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&artifactListOutput{}))
	// TODO When the inspect structure has been defined, we need to uncomment and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

func list(cmd *cobra.Command, _ []string) error {
	reports, err := registry.ImageEngine().ArtifactList(registry.GetContext(), entities.ArtifactListOptions{})
	if err != nil {
		return err
	}

	return outputTemplate(cmd, reports)
}

func outputTemplate(cmd *cobra.Command, lrs []*entities.ArtifactListReport) error {
	var err error
	artifacts := make([]artifactListOutput, 0)
	for _, lr := range lrs {
		var (
			tag string
		)
		artifactName, err := lr.Artifact.GetName()
		if err != nil {
			return err
		}
		repo, err := reference.Parse(artifactName)
		if err != nil {
			return err
		}
		named, ok := repo.(reference.Named)
		if !ok {
			return fmt.Errorf("%q is an invalid artifact name", artifactName)
		}
		if tagged, ok := named.(reference.Tagged); ok {
			tag = tagged.Tag()
		}

		// Note: Right now we only support things that are single manifests
		// We should certainly expand this support for things like arch, etc
		// as we move on
		artifactDigest, err := lr.Artifact.GetDigest()
		if err != nil {
			return err
		}
		// TODO when we default to shorter ids, i would foresee a switch
		// like images that will show the full ids.
		artifacts = append(artifacts, artifactListOutput{
			Digest:     artifactDigest.Encoded(),
			Repository: named.Name(),
			Size:       units.HumanSize(float64(lr.Artifact.TotalSizeBytes())),
			Tag:        tag,
		})
	}

	headers := report.Headers(artifactListOutput{}, map[string]string{
		"REPOSITORY": "REPOSITORY",
		"Tag":        "TAG",
		"Size":       "SIZE",
		"Digest":     "DIGEST",
	})

	rpt := report.New(os.Stdout, cmd.Name())
	defer rpt.Flush()

	switch {
	case cmd.Flag("format").Changed:
		rpt, err = rpt.Parse(report.OriginUser, listFlag.format)
	default:
		rpt, err = rpt.Parse(report.OriginPodman, listFlag.format)
	}
	if err != nil {
		return err
	}

	if rpt.RenderHeaders {
		if err := rpt.Execute(headers); err != nil {
			return fmt.Errorf("failed to write report column headers: %w", err)
		}
	}
	return rpt.Execute(artifacts)
}
