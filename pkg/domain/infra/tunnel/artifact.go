package tunnel

import (
	"context"
	"fmt"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

// TODO For now, no remote support has been added. We need the API to firm up first.

func (ir *ImageEngine) ArtifactInspect(ctx context.Context, name string, opts entities.ArtifactInspectOptions) (*entities.ArtifactInspectReport, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ir *ImageEngine) ArtifactPull(ctx context.Context, name string, opts entities.ArtifactPullOptions) (*entities.ArtifactPullReport, error) {
	return nil, fmt.Errorf("not implemented")
}
