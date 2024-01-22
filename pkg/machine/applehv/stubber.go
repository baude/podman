package applehv

import (
	gvproxy "github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/containers/podman/v4/pkg/machine/define"
	"github.com/containers/podman/v4/pkg/machine/vmconfigs"
	"github.com/containers/podman/v4/pkg/strongunits"
)

type AppleHVStubber struct {
	vmconfigs.AppleHVConfig
}

func (a AppleHVStubber) CreateVM(opts define.CreateVMOpts, mc *vmconfigs.MachineConfig) error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) GetHyperVisorVMs() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) MountType() vmconfigs.VolumeMountType {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) MountVolumesToVM(mc *vmconfigs.MachineConfig, quiet bool) error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) RemoveAndCleanMachines() error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) SetProviderAttrs(mc *vmconfigs.MachineConfig, cpus, memory *uint64, newDiskSize *strongunits.GiB) error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) StartNetworking(mc *vmconfigs.MachineConfig, cmd *gvproxy.GvproxyCommand) error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) StartVM(mc *vmconfigs.MachineConfig) (func() error, func() error, error) {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) StopHostNetworking() error {
	//TODO implement me
	panic("implement me")
}

func (a AppleHVStubber) VMType() define.VMType {
	return define.AppleHvVirt
}
