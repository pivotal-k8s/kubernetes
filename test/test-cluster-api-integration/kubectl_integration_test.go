package kubectl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/testing_frameworks/cluster"
	"sigs.k8s.io/testing_frameworks/lightweight"
)

var _ = Describe("KubectlIntegration", func() {
	var (
		kubeCtl KubeCtl
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

		kubeCtl = KubeCtl{
			Path: kubeCtlPath,
			ConnectionConfig: rest.Config{
				Host: fixture.ClientConfig().Host,
			},
		}
	})
	AfterEach(func() {
		Expect(fixture.TearDown()).To(Succeed())
	})

	It("can run 'get pods'", func() {
		kubeCtl.
			WithArgs("get", "pods").
			WithFormat(GoTemplate("{{.Id}}")).
			ExpectStdoutTo(Equal("<no value>")).
			ExpectStderrTo(BeEmpty()).
			Should(Succeed())
	})
})
