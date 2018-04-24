package kubectl_test

import (
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/testing_frameworks/cluster"
	"sigs.k8s.io/testing_frameworks/lightweight"
)

var _ = Describe("KubectlIntegration", func() {
	var (
		kubeCtl *testKubeCtl
		fixture cluster.Fixture
	)
	BeforeEach(func() {
		// admissionPluginsEnabled := "Initializers,LimitRanger,ResourceQuota"
		// admissionPluginsDisabled := "ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook"
		admissionPluginsDisabled := "ServiceAccount"

		clusterConfig := cluster.Config{}
		clusterConfig.APIServerExtraArgs = map[string]string{
			// This will get a bit nicer as soon as
			// https://github.com/kubernetes-sigs/testing_frameworks/pull/41 is
			// merged
			"--etcd-servers":              "{{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
			"--cert-dir":                  "{{ .CertDir }}",
			"--insecure-port":             "{{ if .URL }}{{ .URL.Port }}{{ end }}",
			"--insecure-bind-address":     "{{ if .URL }}{{ .URL.Hostname }}{{ end }}",
			"--secure-port":               "0",
			"--disable-admission-plugins": admissionPluginsDisabled,
		}
		clusterConfig.APIServerProcessConfig.Path = apiServerPath
		clusterConfig.Etcd.ProcessConfig.Path = etcdPath

		fixture = &lightweight.ControlPlane{}

		Expect(fixture.Setup(clusterConfig)).To(Succeed())

		k := fixture.(*lightweight.ControlPlane).KubeCtl()
		k.Path = kubeCtlPath
		kubeCtl = &testKubeCtl{kubeCtl: k}
	})
	AfterEach(func() {
		Expect(fixture.TearDown()).To(Succeed())
	})

	It("can run 'get pods'", func() {
		stdout, stderr := kubeCtl.Run("get", "pods")
		Expect(stderr).To(ContainSubstring("No resources found."))
		Expect(stdout).To(BeEmpty())
	})
})

type testKubeCtl struct {
	kubeCtl  *lightweight.KubeCtl
	template string
}

func (k *testKubeCtl) Run(args ...string) (string, string) {
	callArgs := []string{}
	callArgs = append(callArgs, args...)
	if k.template != "" {
		callArgs = append(callArgs, "-o", k.template)
	}

	stdout, stderr, err := k.kubeCtl.Run(callArgs...)
	Expect(err).NotTo(HaveOccurred(), "Stdout: %s\nStderr: %s", stdout, stderr)
	return readToString(stdout), readToString(stderr)
}

func readToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(b)
}
