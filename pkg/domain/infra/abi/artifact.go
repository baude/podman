package abi

import (
	"context"
	"os"

	"github.com/containers/common/libimage"
	"github.com/containers/podman/v5/pkg/domain/entities"
	"github.com/containers/podman/v5/pkg/libartifact"
)

func ArtifactAdd(ctx context.Context, path, name string, opts entities.ArtifactAddoptions) error {
	return nil
}

func (ir *ImageEngine) ArtifactInspect(ctx context.Context, name string, opts entities.ArtifactInspectOptions) (*entities.ArtifactInspectReport, error) {
	artStore, err := libartifact.NewArtifactStore(opts.StorePath, ir.Libpod.SystemContext())
	if err != nil {
		return nil, err
	}
	art, err := artStore.Inspect(ctx, name)
	if err != nil {
		return nil, err
	}
	artInspectReport := entities.ArtifactInspectReport{
		Artifact: art,
	}
	return &artInspectReport, nil
}

func (ir *ImageEngine) ArtifactList(ctx context.Context, opts entities.ArtifactListOptions) ([]*entities.ArtifactListReport, error) {
	var reports []*entities.ArtifactListReport
	artStore, err := libartifact.NewArtifactStore(opts.StorePath, ir.Libpod.SystemContext())
	if err != nil {
		return nil, err
	}
	lrs, err := artStore.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, lr := range lrs {
		artListReport := entities.ArtifactListReport{
			Artifact: lr,
		}
		reports = append(reports, &artListReport)
	}
	return reports, nil
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

	if !opts.Quiet && pullOptions.Writer == nil {
		pullOptions.Writer = os.Stderr
	}

	artStore, err := libartifact.NewArtifactStore(opts.StorePath, ir.Libpod.SystemContext())
	if err != nil {
		return nil, err
	}
	return nil, artStore.Pull(ctx, name)
}

func (ir *ImageEngine) ArtifactRm(ctx context.Context, name string, opts entities.ArtifactRemoveOptions) (*entities.ArtifactRemoveReport, error) {
	artStore, err := libartifact.NewArtifactStore(opts.StorePath, ir.Libpod.SystemContext())
	if err != nil {
		return nil, err
	}
	err = artStore.Remove(ctx, artStore.SystemContext, name)
	return nil, err
}

func (ir *ImageEngine) ArtifactPush(ctx context.Context, name string, opts entities.ArtifactPushOptions) (*entities.ArtifactPushReport, error) {
	artStore, err := libartifact.NewArtifactStore(opts.StorePath, ir.Libpod.SystemContext())
	if err != nil {
		return nil, err
	}

	err = artStore.Push(ctx, name, name)
	return &entities.ArtifactPushReport{}, err
}
