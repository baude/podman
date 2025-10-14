//go:build amd64 || arm64

package os

import (
	"github.com/blang/semver/v4"
)

// Manager is the interface for operations on a Podman machine's OS
type Manager interface {
	// Apply machine OS changes from an OCI image.
	Apply(image string, opts ApplyOptions) error
	// Upgrade the machine OS
	Upgrade(hostVersion string, opts UpgradeOptions) error
}

// ApplyOptions are the options for applying an image into a Podman machine VM
type ApplyOptions struct {
	Image string
}

type UpgradeOptions struct {
	HostVersion semver.Version
}
