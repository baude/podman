package integration

import (
	"testing"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"path/filepath"
	"fmt"
	"strings"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)
// - CRIO_ROOT=/var/tmp/checkout PODMAN_BINARY=/usr/bin/podman CONMON_BINARY=/usr/libexec/crio/conmon PAPR=1 sh .papr.sh
// PODMAN_OPTIONS="--root $TESTDIR/crio $STORAGE_OPTIONS --runroot $TESTDIR/crio-run --runtime ${RUNTIME_BINARY} --conmon ${CONMON_BINARY} --cni-config-dir ${LIBPOD_CNI_CONFIG}"

//TODO do the image caching
// "$COPYIMG_BINARY" --root "$TESTDIR/crio" $STORAGE_OPTIONS --runroot "$TESTDIR/crio-run" --image-name=${IMAGES[${key}]} --import-from=dir:"$ARTIFACTS_PATH"/${key} --add-name=${IMAGES[${key}]}
//TODO whats the best way to clean up after a test

var (
	PODMAN_BINARY="/home/bbaude/git/libpod/bin/podman"
	CONMON_BINARY="/home/bbaude/git/libpod/bin/conmon"
	CRIO_ROOT="/home/bbaude/git/libpod/"
	CNI_CONFIG_DIR="/etc/cni/net.d"
	RUNC_BINARY="/usr/bin/runc"
)
// PodmanSession wrapps the gexec.session so we can extend it
type PodmanSession struct {
    *gexec.Session
}

// PodmanTest struct for command line options
type PodmanTest struct {
	PodmanBinary string
	ConmonBinary string
	CrioRoot string
	CNIConfigDir string
	RunCBinary string
	RunRoot string
}

// TestLibpod ginkgo master function
func TestLibpod(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Libpod Suite")
}

// CreateTempDirin
func CreateTempDirInTempDir() (string, error){
	return ioutil.TempDir("","podman_test")
}

// PodmanCreate creates a PodmanTest instance for the tests
func PodmanCreate(podmanBinary, conmonBinary, CNIConfigDir, tempDir, runCBinary string) PodmanTest {
	return PodmanTest{
		PodmanBinary: podmanBinary,
		ConmonBinary: conmonBinary,
		CrioRoot: filepath.Join(tempDir, "crio"),
		CNIConfigDir: CNIConfigDir,
		RunCBinary: runCBinary,
		RunRoot: filepath.Join(tempDir, "crio-run"),
	}
}

//MakeOptions assembles all the podman main options
func (p *PodmanTest) MakeOptions() []string{
	return strings.Split(fmt.Sprintf("--root %s --runroot %s --runtime %s --conmon %s --cni-config-dir %s",
		p.CrioRoot, p.RunRoot, p.RunCBinary, p.ConmonBinary, p.CNIConfigDir), " ")
}

// Podman is the exec call to podman on the filesystem
func (p *PodmanTest) Podman(args []string) (*PodmanSession){
	podmanOptions := p.MakeOptions()
	podmanOptions = append(podmanOptions, args...)
	command := exec.Command(PODMAN_BINARY, podmanOptions...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	if err != nil {
		Fail(fmt.Sprintf("unable to run podman command: %s", strings.Join(podmanOptions, " ")))
	}
	return &PodmanSession{session}
}

// Cleanup cleans up the temporary store
func (p *PodmanTest) Cleanup() {
	// does nothing yet
}

// GrepString takes session output and behaves like grep. it returns a bool
// if successful and an array of strings on positive matches
func (s *PodmanSession) GrepString(term string) (bool, []string) {
	// not implement
	return false, nil
}
// Pull Images pulls multiple images
func (p *PodmanTest) PullImages(images []string) error {
	for _, i := range images {
		p.PullImage(i)
	}
	return nil
}

// Pull Image a single image
// TODO should the timeout be configurable?
func (p *PodmanTest) PullImage(image string) error {
	session := p.Podman([]string{"pull", image})
	session.Wait(60)
	Expect(session.ExitCode()).To(Equal(0))
	return nil
}

// OutputToString formats session output to string
func (s *PodmanSession) OutputToString() string {
	return fmt.Sprintf("%s", s.Out.Contents())
}

// SystemExec is used to exec a system command to check its exit code or output
func (p *PodmanSession) SystemExec(command string, args []string) (*PodmanSession, error){
	// not implemented yet
	return nil, nil
}