package persistent_volume_test

import (
	"fmt"
	. "tests/test_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("NFS", func() {
	var (
		kubectl *KubectlRunner
		iaas    string
	)

	BeforeEach(func() {
		kubectl = NewKubectlRunner()
		kubectl.Setup()

		var err error
		iaas, err = IaaS()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		kubectl.Teardown()
	})

	Context("when creating an NFS PV", func() {
		var (
			storageClassSpec     string
			nfsServerSpec        string
			nfsServerServiceSpec string
			nfsPvSpec            string
			nfsPvcSpec           string
			nfsPodRcSpec         string
		)

		BeforeEach(func() {
			storageClassSpec = PathFromRoot(fmt.Sprintf("specs/storage-class-%s.yml", iaas))
			nfsServerSpec = PathFromRoot("specs/nfs-server-statefulset.yml")
			nfsServerServiceSpec = PathFromRoot("specs/nfs-server-service.yml")
			nfsPvSpec = PathFromRoot("specs/nfs-pv.yml")
			nfsPvcSpec = PathFromRoot("specs/nfs-pvc.yml")
			nfsPodRcSpec = PathFromRoot("specs/nfs-pod-rc.yml")
			Eventually(kubectl.RunKubectlCommand("apply", "-f", storageClassSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("apply", "-f", nfsServerSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("apply", "-f", nfsServerServiceSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("apply", "-f", nfsPvSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("apply", "-f", nfsPvcSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("apply", "-f", nfsPodRcSpec), "60s").Should(gexec.Exit(0))
		})

		AfterEach(func() {
			Eventually(kubectl.RunKubectlCommand("delete", "-f", nfsPodRcSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "-f", nfsPvcSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "-f", nfsPvSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "-f", nfsServerServiceSpec), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "-f", nfsServerSpec), "60s").Should(gexec.Exit(0))
			// Some pv(c)s aren't being cleaned
			Eventually(kubectl.RunKubectlCommand("delete", "pvc", "--all"), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "pv", "--all"), "60s").Should(gexec.Exit(0))
			Eventually(kubectl.RunKubectlCommand("delete", "-f", storageClassSpec), "60s").Should(gexec.Exit(0))
		})

		FIt("should mount an NFS PV to a workload", func() {
			podName := kubectl.GetResourceNameBySelector(kubectl.Namespace(), "pod", "name=nfs-busybox")

			Eventually(func() int {
				session := kubectl.RunKubectlCommand("exec", podName, "--", "cat", "/mnt/index.html")
				fmt.Println(session.ExitCode())
				return session.ExitCode()
			}, "300s", "5s").Should(Equal(0))
		})
	})
})
