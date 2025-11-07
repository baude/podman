package types

import (
	"github.com/opencontainers/go-digest"
	libartifact "go.podman.io/common/pkg/libartifact/store"
)

type ArtifactInspectReport struct {
	*libartifact.Artifact
	Digest string
}

type ArtifactAddReport struct {
	ArtifactDigest *digest.Digest
}

type ArtifactRemoveReport struct {
	ArtifactDigests []*digest.Digest
}

type ArtifactListReport struct {
	*libartifact.Artifact
}

type ArtifactPushReport struct {
	ArtifactDigest *digest.Digest
}

type ArtifactPullReport struct {
	ArtifactDigest *digest.Digest
}
