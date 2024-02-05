//go:build amd64 || arm64

package machine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/podman/v4/pkg/machine/compression"
	"github.com/containers/podman/v4/pkg/machine/define"
	"github.com/containers/podman/v4/pkg/machine/vmconfigs"
	"github.com/containers/storage/pkg/homedir"
	"github.com/containers/storage/pkg/lockfile"
	"github.com/sirupsen/logrus"
)

const (
	DefaultMachineName string = "podman-machine-default"
	apiUpTimeout              = 20 * time.Second
)

var (
	DefaultIgnitionUserName = "core"
	ForwarderBinaryName     = "gvproxy"
)

type Download struct {
	Arch                  string
	Artifact              define.Artifact
	CacheDir              string
	CompressionType       compression.ImageCompression
	DataDir               string
	Format                define.ImageFormat
	ImageName             string
	LocalPath             string
	LocalUncompressedFile string
	Sha256sum             string
	Size                  int64
	URL                   *url.URL
	VMKind                define.VMType
	VMName                string
}

type ListOptions struct{}

type ListResponse struct {
	Name               string
	CreatedAt          time.Time
	LastUp             time.Time
	Running            bool
	Starting           bool
	Stream             string
	VMType             string
	CPUs               uint64
	Memory             uint64
	DiskSize           uint64
	Port               int
	RemoteUsername     string
	IdentityPath       string
	UserModeNetworking bool
}

type SetOptions struct {
	CPUs               *uint64
	DiskSize           *uint64
	Memory             *uint64
	Rootful            *bool
	UserModeNetworking *bool
	USBs               *[]string
}

type SSHOptions struct {
	Username string
	Args     []string
}

type StartOptions struct {
	NoInfo bool
	Quiet  bool
}

type StopOptions struct{}

type RemoveOptions struct {
	Force        bool
	SaveImage    bool
	SaveIgnition bool
}

type InspectOptions struct{}

type VM interface {
	Init(opts define.InitOptions) (bool, error)
	Inspect() (*InspectInfo, error)
	Remove(name string, opts RemoveOptions) (string, func() error, error)
	Set(name string, opts SetOptions) ([]error, error)
	SSH(name string, opts SSHOptions) error
	Start(name string, opts StartOptions) error
	State(bypass bool) (define.Status, error)
	Stop(name string, opts StopOptions) error
}

type DistributionDownload interface {
	HasUsableCache() (bool, error)
	Get() *Download
	CleanCache() error
}
type InspectInfo struct {
	ConfigPath         define.VMFile
	ConnectionInfo     ConnectionConfig
	Created            time.Time
	Image              ImageConfig
	LastUp             time.Time
	Name               string
	Resources          vmconfigs.ResourceConfig
	SSHConfig          vmconfigs.SSHConfig
	State              define.Status
	UserModeNetworking bool
	Rootful            bool
}

// GetCacheDir returns the dir where VM images are downloaded into when pulled
func GetCacheDir(vmType define.VMType) (string, error) {
	dataDir, err := GetDataDir(vmType)
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(dataDir, "cache")
	if _, err := os.Stat(cacheDir); !errors.Is(err, os.ErrNotExist) {
		return cacheDir, nil
	}
	return cacheDir, os.MkdirAll(cacheDir, 0755)
}

// GetDataDir returns the filepath where vm images should
// live for podman-machine.
func GetDataDir(vmType define.VMType) (string, error) {
	dataDirPrefix, err := DataDirPrefix()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(dataDirPrefix, vmType.String())
	if _, err := os.Stat(dataDir); !errors.Is(err, os.ErrNotExist) {
		return dataDir, nil
	}
	mkdirErr := os.MkdirAll(dataDir, 0755)
	return dataDir, mkdirErr
}

// GetGLobalDataDir returns the root of all backends
// for shared machine data.
func GetGlobalDataDir() (string, error) {
	dataDir, err := DataDirPrefix()
	if err != nil {
		return "", err
	}

	return dataDir, os.MkdirAll(dataDir, 0755)
}

func GetMachineDirs(vmType define.VMType) (*define.MachineDirs, error) {
	rtDir, err := getRuntimeDir()
	if err != nil {
		return nil, err
	}

	rtDir = filepath.Join(rtDir, "podman")
	configDir, err := GetConfDir(vmType)
	if err != nil {
		return nil, err
	}

	configDirFile, err := define.NewMachineFile(configDir, nil)
	if err != nil {
		return nil, err
	}
	dataDir, err := GetDataDir(vmType)
	if err != nil {
		return nil, err
	}

	dataDirFile, err := define.NewMachineFile(dataDir, nil)
	if err != nil {
		return nil, err
	}

	rtDirFile, err := define.NewMachineFile(rtDir, nil)
	if err != nil {
		return nil, err
	}

	dirs := define.MachineDirs{
		ConfigDir:  configDirFile,
		DataDir:    dataDirFile,
		RuntimeDir: rtDirFile,
	}

	// make sure all machine dirs are present
	if err := os.MkdirAll(rtDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	err = os.MkdirAll(dataDir, 0755)

	return &dirs, err
}

// DataDirPrefix returns the path prefix for all machine data files
func DataDirPrefix() (string, error) {
	data, err := homedir.GetDataHome()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(data, "containers", "podman", "machine")
	return dataDir, nil
}

// GetConfigDir returns the filepath to where configuration
// files for podman-machine should live
func GetConfDir(vmType define.VMType) (string, error) {
	confDirPrefix, err := ConfDirPrefix()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(confDirPrefix, vmType.String())
	if _, err := os.Stat(confDir); !errors.Is(err, os.ErrNotExist) {
		return confDir, nil
	}
	mkdirErr := os.MkdirAll(confDir, 0755)
	return confDir, mkdirErr
}

// ConfDirPrefix returns the path prefix for all machine config files
func ConfDirPrefix() (string, error) {
	conf, err := homedir.GetConfigHome()
	if err != nil {
		return "", err
	}
	confDir := filepath.Join(conf, "containers", "podman", "machine")
	return confDir, nil
}

// GetSSHIdentityPath returns the path to the expected SSH private key
func GetSSHIdentityPath(name string) (string, error) {
	datadir, err := GetGlobalDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(datadir, name), nil
}

// ImageConfig describes the bootable image for the VM
type ImageConfig struct {
	// IgnitionFile is the path to the filesystem where the
	// ignition file was written (if needs one)
	IgnitionFile define.VMFile `json:"IgnitionFilePath"`
	// ImageStream is the update stream for the image
	ImageStream string
	// ImageFile is the fq path to
	ImagePath define.VMFile `json:"ImagePath"`
}

// ConnectionConfig contains connections like sockets, etc.
type ConnectionConfig struct {
	// PodmanSocket is the exported podman service socket
	PodmanSocket *define.VMFile `json:"PodmanSocket"`
	// PodmanPipe is the exported podman service named pipe (Windows hosts only)
	PodmanPipe *define.VMFile `json:"PodmanPipe"`
}

type APIForwardingState int

const (
	NoForwarding APIForwardingState = iota
	ClaimUnsupported
	NotInstalled
	MachineLocal
	DockerGlobal
)

type Virtualization struct {
	artifact    define.Artifact
	compression compression.ImageCompression
	format      define.ImageFormat
	vmKind      define.VMType
}

func (p *Virtualization) Artifact() define.Artifact {
	return p.artifact
}

func (p *Virtualization) Compression() compression.ImageCompression {
	return p.compression
}

func (p *Virtualization) Format() define.ImageFormat {
	return p.format
}

func (p *Virtualization) VMType() define.VMType {
	return p.vmKind
}

func (p *Virtualization) NewDownload(vmName string) (Download, error) {
	cacheDir, err := GetCacheDir(p.VMType())
	if err != nil {
		return Download{}, err
	}

	dataDir, err := GetDataDir(p.VMType())
	if err != nil {
		return Download{}, err
	}

	return Download{
		Artifact:        p.Artifact(),
		CacheDir:        cacheDir,
		CompressionType: p.Compression(),
		DataDir:         dataDir,
		Format:          p.Format(),
		VMKind:          p.VMType(),
		VMName:          vmName,
	}, nil
}

func NewVirtualization(artifact define.Artifact, compression compression.ImageCompression, format define.ImageFormat, vmKind define.VMType) Virtualization {
	return Virtualization{
		artifact,
		compression,
		format,
		vmKind,
	}
}

func WaitAndPingAPI(sock string) {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(context.Context, string, string) (net.Conn, error) {
				con, err := net.DialTimeout("unix", sock, apiUpTimeout)
				if err != nil {
					return nil, err
				}
				if err := con.SetDeadline(time.Now().Add(apiUpTimeout)); err != nil {
					return nil, err
				}
				return con, nil
			},
		},
	}

	resp, err := client.Get("http://host/_ping")
	if err == nil {
		defer resp.Body.Close()
	}
	if err != nil || resp.StatusCode != 200 {
		logrus.Warn("API socket failed ping test")
	}
}

func (dl Download) NewFcosDownloader(imageStream FCOSStream) (DistributionDownload, error) {
	info, err := dl.GetFCOSDownload(imageStream)
	if err != nil {
		return nil, err
	}
	urlSplit := strings.Split(info.Location, "/")
	dl.ImageName = urlSplit[len(urlSplit)-1]
	downloadURL, err := url.Parse(info.Location)
	if err != nil {
		return nil, err
	}

	// Complete the download struct
	dl.Arch = GetFcosArch()
	// This could be eliminated as a struct and be a generated()
	dl.LocalPath = filepath.Join(dl.CacheDir, dl.ImageName)
	dl.Sha256sum = info.Sha256Sum
	dl.URL = downloadURL
	fcd := FcosDownload{
		Download: dl,
	}
	dataDir, err := GetDataDir(dl.VMKind)
	if err != nil {
		return nil, err
	}
	fcd.Download.LocalUncompressedFile = fcd.GetLocalUncompressedFile(dataDir)
	return fcd, nil
}

// AcquireVMImage determines if the image is already in a FCOS stream. If so,
// retrieves the image path of the uncompressed file. Otherwise, the user has
// provided an alternative image, so we set the image path and download the image.
func (dl Download) AcquireVMImage(imagePath string) (*define.VMFile, FCOSStream, error) {
	var (
		err           error
		imageLocation *define.VMFile
		fcosStream    FCOSStream
	)

	switch imagePath {
	// TODO these need to be re-typed as FCOSStreams
	case Testing.String(), Next.String(), Stable.String(), "":
		// Get image as usual
		fcosStream, err = FCOSStreamFromString(imagePath)
		if err != nil {
			return nil, 0, err
		}

		dd, err := dl.NewFcosDownloader(fcosStream)
		if err != nil {
			return nil, 0, err
		}

		imageLocation, err = define.NewMachineFile(dd.Get().LocalUncompressedFile, nil)
		if err != nil {
			return nil, 0, err
		}

		if err := DownloadImage(dd); err != nil {
			return nil, 0, err
		}
	default:
		// The user has provided an alternate image which can be a file path
		// or URL.
		fcosStream = CustomStream
		imgPath, err := dl.AcquireAlternateImage(imagePath)
		if err != nil {
			return nil, 0, err
		}
		imageLocation = imgPath
	}
	return imageLocation, fcosStream, nil
}

// Deprecated: GetLock
func GetLock(name string, vmtype define.VMType) (*lockfile.LockFile, error) {
	// FIXME: there's a painful amount of `GetConfDir` calls scattered
	// across the code base.  This should be done once and stored
	// somewhere instead.
	vmConfigDir, err := GetConfDir(vmtype)
	if err != nil {
		return nil, err
	}

	lockPath := filepath.Join(vmConfigDir, name+".lock")
	lock, err := lockfile.GetLockFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("creating lockfile for VM: %w", err)
	}

	return lock, nil
}
