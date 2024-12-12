package artifact

import (
	"fmt"
	"os"

	"github.com/containers/buildah/pkg/cli"
	"github.com/containers/common/pkg/auth"
	"github.com/containers/common/pkg/completion"
	"github.com/containers/image/v5/types"
	"github.com/containers/podman/v5/cmd/podman/common"
	"github.com/containers/podman/v5/cmd/podman/registry"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/containers/podman/v5/pkg/util"
	"github.com/spf13/cobra"
)

// pullOptionsWrapper wraps entities.ImagePullOptions and prevents leaking
// CLI-only fields into the API types.
type pullOptionsWrapper struct {
	entities.ArtifactPullOptions
	TLSVerifyCLI   bool // CLI only
	CredentialsCLI string
	DecryptionKeys []string
}

var (
	pullOptions     = pullOptionsWrapper{}
	pullDescription = `Pulls an artifact from a registry and stores it locally.`

	pullCmd = &cobra.Command{
		Use:   "pull [options] ARTIFACT",
		Short: "Pull an OCI artifact",
		Long:  pullDescription,
		RunE:  artifactPull,
		//PersistentPreRunE: devOnly,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: common.AutocompleteArtifacts,
		Example:           `podman artifact pull quay.io/myimage/myartifact:latest`,
		// TODO Autocomplete function needs to be done
	}
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: pullCmd,
		Parent:  artifactCmd,
	})
	pullFlags(pullCmd)

	// TODO When the inspect structure has been defined, we need to uncommand and redirect this.  Reminder, this
	// will also need to be reflected in the podman-artifact-inspect man page
	// _ = inspectCmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteFormat(&machine.InspectInfo{}))
}

// pullFlags set the flags for the pull command.
func pullFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	credsFlagName := "creds"
	flags.StringVar(&pullOptions.CredentialsCLI, credsFlagName, "", "`Credentials` (USERNAME:PASSWORD) to use for authenticating to a registry")
	_ = cmd.RegisterFlagCompletionFunc(credsFlagName, completion.AutocompleteNone)

	// TODO I think these can be removed
	/*
		archFlagName := "arch"
		flags.StringVar(&pullOptions.Arch, archFlagName, "", "Use `ARCH` instead of the architecture of the machine for choosing images")
		_ = cmd.RegisterFlagCompletionFunc(archFlagName, completion.AutocompleteArch)


		osFlagName := "os"
		flags.StringVar(&pullOptions.OS, osFlagName, "", "Use `OS` instead of the running OS for choosing images")
		_ = cmd.RegisterFlagCompletionFunc(osFlagName, completion.AutocompleteOS)

			variantFlagName := "variant"
			flags.StringVar(&pullOptions.Variant, variantFlagName, "", "Use VARIANT instead of the running architecture variant for choosing images")
			_ = cmd.RegisterFlagCompletionFunc(variantFlagName, completion.AutocompleteNone)

			platformFlagName := "platform"
			flags.String(platformFlagName, "", "Specify the platform for selecting the image.  (Conflicts with arch and os)")
			_ = cmd.RegisterFlagCompletionFunc(platformFlagName, completion.AutocompleteNone)


			flags.Bool("disable-content-trust", false, "This is a Docker specific option and is a NOOP")
	*/

	flags.BoolVarP(&pullOptions.Quiet, "quiet", "q", false, "Suppress output information when pulling images")
	flags.BoolVar(&pullOptions.TLSVerifyCLI, "tls-verify", true, "Require HTTPS and verify certificates when contacting registries")

	authfileFlagName := "authfile"
	flags.StringVar(&pullOptions.AuthFilePath, authfileFlagName, auth.GetDefaultAuthFile(), "Path of the authentication file. Use REGISTRY_AUTH_FILE environment variable to override")
	_ = cmd.RegisterFlagCompletionFunc(authfileFlagName, completion.AutocompleteDefault)

	decryptionKeysFlagName := "decryption-key"
	flags.StringArrayVar(&pullOptions.DecryptionKeys, decryptionKeysFlagName, nil, "Key needed to decrypt the image (e.g. /path/to/key.pem)")
	_ = cmd.RegisterFlagCompletionFunc(decryptionKeysFlagName, completion.AutocompleteDefault)

	retryFlagName := "retry"
	flags.Uint(retryFlagName, registry.RetryDefault(), "number of times to retry in case of failure when performing pull")
	_ = cmd.RegisterFlagCompletionFunc(retryFlagName, completion.AutocompleteNone)
	retryDelayFlagName := "retry-delay"
	flags.String(retryDelayFlagName, registry.RetryDelayDefault(), "delay between retries in case of pull failures")
	_ = cmd.RegisterFlagCompletionFunc(retryDelayFlagName, completion.AutocompleteNone)

	if registry.IsRemote() {
		_ = flags.MarkHidden(decryptionKeysFlagName)
	} else {
		certDirFlagName := "cert-dir"
		flags.StringVar(&pullOptions.CertDirPath, certDirFlagName, "", "`Pathname` of a directory containing TLS certificates and keys")
		_ = cmd.RegisterFlagCompletionFunc(certDirFlagName, completion.AutocompleteDefault)

		signaturePolicyFlagName := "signature-policy"
		flags.StringVar(&pullOptions.SignaturePolicyPath, signaturePolicyFlagName, "", "`Pathname` of signature policy file (not usually used)")
		_ = flags.MarkHidden(signaturePolicyFlagName)
	}
}

func artifactPull(cmd *cobra.Command, args []string) error {
	// TLS verification in c/image is controlled via a `types.OptionalBool`
	// which allows for distinguishing among set-true, set-false, unspecified
	// which is important to implement a sane way of dealing with defaults of
	// boolean CLI flags.
	if cmd.Flags().Changed("tls-verify") {
		pullOptions.InsecureSkipTLSVerify = types.NewOptionalBool(!pullOptions.TLSVerifyCLI)
	}

	if cmd.Flags().Changed("retry") {
		retry, err := cmd.Flags().GetUint("retry")
		if err != nil {
			return err
		}

		pullOptions.MaxRetries = &retry
	}

	if cmd.Flags().Changed("retry-delay") {
		val, err := cmd.Flags().GetString("retry-delay")
		if err != nil {
			return err
		}

		pullOptions.RetryDelay = val
	}

	if cmd.Flags().Changed("authfile") {
		if err := auth.CheckAuthFile(pullOptions.AuthFilePath); err != nil {
			return err
		}
	}

	// TODO Once we have a decision about the flag removal above, this should be safe to delete
	/*
		platform, err := cmd.Flags().GetString("platform")
		if err != nil {
			return err
		}
		if platform != "" {
			if pullOptions.Arch != "" || pullOptions.OS != "" {
				return errors.New("--platform option can not be specified with --arch or --os")
			}

			specs := strings.Split(platform, "/")
			pullOptions.OS = specs[0] // may be empty
			if len(specs) > 1 {
				pullOptions.Arch = specs[1]
				if len(specs) > 2 {
					pullOptions.Variant = specs[2]
				}
			}
		}
	*/

	if pullOptions.CredentialsCLI != "" {
		creds, err := util.ParseRegistryCreds(pullOptions.CredentialsCLI)
		if err != nil {
			return err
		}
		pullOptions.Username = creds.Username
		pullOptions.Password = creds.Password
	}

	decConfig, err := cli.DecryptConfig(pullOptions.DecryptionKeys)
	if err != nil {
		return fmt.Errorf("unable to obtain decryption config: %w", err)
	}
	pullOptions.OciDecryptConfig = decConfig

	if !pullOptions.Quiet {
		pullOptions.Writer = os.Stderr
	}

	_, err = registry.ImageEngine().ArtifactPull(registry.GetContext(), args[0], pullOptions.ArtifactPullOptions)
	return err

}
