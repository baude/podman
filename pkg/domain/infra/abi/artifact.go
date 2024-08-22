package abi

import (
	"context"
	"os"

	"github.com/containers/common/libimage"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/containers/podman/v5/pkg/libartifact"
)

func (ir *ImageEngine) ArtifactInspect(ctx context.Context, name string, opts entities.ArtifactInspectOptions) (*entities.ArtifactInspectReport, error) {
	return nil, nil
}

func (ir *ImageEngine) ArtifactPull(ctx context.Context, name string, opts entities.ArtifactPullOptions) (*entities.ArtifactPullReport, error) {
	pullOptions := &libimage.PullOptions{}
	pullOptions.AuthFilePath = opts.AuthFilePath
	pullOptions.CertDirPath = opts.CertDirPath
	pullOptions.Username = opts.Username
	pullOptions.Password = opts.Password
	//pullOptions.Architecture = opts.Arch
	pullOptions.SignaturePolicyPath = opts.SignaturePolicyPath
	pullOptions.InsecureSkipTLSVerify = opts.InsecureSkipTLSVerify
	pullOptions.Writer = opts.Writer
	pullOptions.OciDecryptConfig = opts.OciDecryptConfig
	pullOptions.MaxRetries = opts.MaxRetries

	//if opts.RetryDelay != "" {
	//	duration, err := time.ParseDuration(opts.RetryDelay)
	//	if err != nil {
	//		return nil, err
	//	}
	//	pullOptions.RetryDelay = &duration
	//}
	storeConfig := ir.Libpod.StorageConfig()
	foo := storeConfig.ImageStore
	_ = foo
	if !opts.Quiet && pullOptions.Writer == nil {
		pullOptions.Writer = os.Stderr
	}
	artStore, err := libartifact.NewArtifactStore(opts.StorePath)
	if err != nil {
		return nil, err
	}
	return nil, artStore.Pull(ctx, name)
}
