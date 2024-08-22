package entities

import (
	"github.com/containers/image/v5/types"
	encconfig "github.com/containers/ocicrypt/config"
	"io"
)

type ArtifactInspectOptions struct{}

type ArtifactPullOptions struct {
	StorePath             string
	AuthFilePath          string
	CertDirPath           string
	Username              string
	Password              string
	Architecture          string
	SignaturePolicyPath   string
	InsecureSkipTLSVerify types.OptionalBool
	Writer                io.Writer
	OciDecryptConfig      *encconfig.DecryptConfig
	MaxRetries            *uint
	RetryDelay            string
	Quiet                 bool
}

type ArtifactPullReport struct{}

type ArtifactInspectReport struct{}
