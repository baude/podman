package entities

import (
	"io"

	"github.com/containers/image/v5/types"
	encconfig "github.com/containers/ocicrypt/config"
	"github.com/containers/podman/v5/pkg/libartifact"
)

type ArtifactAddoptions struct{}

type ArtifactInspectOptions struct {
	Remote    bool
	StorePath string
}

type ArtifactListOptions struct {
	ImagePushOptions
	StorePath string
}

type ArtifactPullOptions struct {
	Architecture          string
	AuthFilePath          string
	CertDirPath           string
	InsecureSkipTLSVerify types.OptionalBool
	MaxRetries            *uint
	OciDecryptConfig      *encconfig.DecryptConfig
	Password              string
	Quiet                 bool
	RetryDelay            string
	SignaturePolicyPath   string
	StorePath             string
	Username              string
	Writer                io.Writer
}

type ArtifactPushOptions struct {
	ImagePushOptions
	CredentialsCLI             string
	DigestFile                 string
	EncryptLayers              []int
	EncryptionKeys             []string
	SignBySigstoreParamFileCLI string
	SignPassphraseFileCLI      string
	StorePath                  string
	TLSVerifyCLI               bool // CLI only
}

type ArtifactRemoveOptions struct {
	StorePath string
}

type ArtifactPullReport struct{}

type ArtifactPushReport struct{}

type ArtifactInspectReport struct {
	*libartifact.Artifact
}

type ArtifactListReport struct {
	*libartifact.Artifact
}

type ArtifactRemoveReport struct{}
