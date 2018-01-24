package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
)


var _ = Describe("Podman rmi", func() {
	var (
		tempdir string
		err error
		podmanTest PodmanTest
		image1 = "docker.io/library/alpine:latest"
		image2 = "docker.io/library/busybox:latest"
		image3 = "docker.io/library/busybox:glibc"
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

	It("podman rmi bogus image", func() {
		session := podmanTest.Podman([]string{"rmi", "debian:6.0.10"})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(125))

	})

	It("podman rmi with fq name", func() {
		podmanTest.PullImage(image1)
		session := podmanTest.Podman([]string{"rmi", image1})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(0))

	})

	It("podman rmi with short name", func() {
		podmanTest.PullImage(image1)
		session := podmanTest.Podman([]string{"rmi", "alpine"})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(0))

	})

	It("podman rmi all images", func() {
		podmanTest.PullImages([]string{image1, image2, image3})
		session := podmanTest.Podman([]string{"rmi", "-a"})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(0))

	})

	It("podman rmi all images forceably with short options", func() {
		podmanTest.PullImages([]string{image1, image2, image3})
		podmanTest.PullImage(image1)
		session := podmanTest.Podman([]string{"rmi", "-fa"})
		session.Wait()
		Expect(session.ExitCode()).To(Equal(0))

	})

})


