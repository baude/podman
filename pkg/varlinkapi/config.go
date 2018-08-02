package varlinkapi

import (
	iocontainerspodman "github.com/projectatomic/libpod/cmd/podman/varlink"
	"github.com/urfave/cli"
)

// LibpodAPI is the basic varlink struct for libpod
type LibpodAPI struct {
	Cli *cli.Context
	iocontainerspodman.VarlinkInterface
}

// New creates a new varlink client
func New(cli *cli.Context) *iocontainerspodman.VarlinkInterface {
	lp := LibpodAPI{Cli: cli}
	return iocontainerspodman.VarlinkNew(&lp)
}
