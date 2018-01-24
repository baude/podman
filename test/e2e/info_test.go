package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)


var _ = Describe("Podman Info", func() {
	var (
		tempdir string
		err error
		podmanTest PodmanTest
	)

	BeforeEach(func() {
		tempdir, err = CreateTempDirInTempDir()
		if err != nil {
			os.Exit(1)
		}
		podmanTest = PodmanCreate(PODMAN_BINARY, CONMON_BINARY, CNI_CONFIG_DIR, tempdir, RUNC_BINARY)
	})

	AfterEach(func() {
		podmanTest.Cleanup()
	})

	It("podman info json output", func() {
		session := podmanTest.Podman([]string{"info", "--format=json"})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(0))

	})
})


