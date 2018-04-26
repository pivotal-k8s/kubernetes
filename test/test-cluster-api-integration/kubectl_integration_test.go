package kubectl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/testing_frameworks/cluster"
	"sigs.k8s.io/testing_frameworks/lightweight"
)

var _ = Describe("KubectlIntegration", func() {
	var (
		fixture cluster.Fixture
	)
	BeforeEach(func() {
		fixture = &lightweight.ControlPlane{}

		Expect(fixture.Setup(clusterConfig)).To(Succeed())

		configUpdate(fixture)
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

	It("can create, use and default to namespaces", func() {
		createAndUseNamespace()
		kubeCtl.
			WithNamespace(Namespace{AutoCreate: true}).
			WithArgs("get", "namespace").To(Succeed())
	})
})
