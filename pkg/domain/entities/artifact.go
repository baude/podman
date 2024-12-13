package entities

import (
	"github.com/opencontainers/go-digest"
	"io"

	"github.com/containers/image/v5/types"
	encconfig "github.com/containers/ocicrypt/config"
	"github.com/containers/podman/v5/pkg/libartifact"
)

type ArtifactAddoptions struct {
	ArtifactType string
}

type ArtifactInspectOptions struct {
	Remote bool
}

type ArtifactListOptions struct {
	ImagePushOptions
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
	TLSVerifyCLI               bool // CLI only
}

type ArtifactRemoveOptions struct {
}

type ArtifactPullReport struct{}

type ArtifactPushReport struct{}

type ArtifactInspectReport struct {
	*libartifact.Artifact
}

type ArtifactListReport struct {
	Name string
	Size int64
	*libartifact.Artifact
}

type ArtifactAddReport struct {
	NewBlobDigest *digest.Digest
}

type ArtifactRemoveReport struct{}
