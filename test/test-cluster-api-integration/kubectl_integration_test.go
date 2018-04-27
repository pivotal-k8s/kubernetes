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

	Context("CRD Tests", func() {
		var ns string

		BeforeEach(func() {
			ns = createAndUseNamespace()
		})
		AfterEach(func() {
			kubeCtl.WithArgs("delete", "namespace", ns).Should(Succeed())
		})

		It("can create multiple crds", func() {
			kubeCtl.Create(crdFooSpec)
			kubeCtl.
				WithArgs("get", "customresourcedefinitions").
				WithFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectStdoutTo(Equal("foos.company.com:")).
				Should(Succeed())

			kubeCtl.Create(crdBarSpec)
			kubeCtl.
				WithArgs("get", "customresourcedefinitions").
				WithFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectStdoutTo(Equal("bars.company.com:foos.company.com:")).
				Should(Succeed())

			kubeCtl.Create(crdResourcesSpec)
			kubeCtl.
				WithArgs("get", "customresourcedefinitions").
				WithFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectStdoutTo(Equal("bars.company.com:foos.company.com:resources.mygroup.example.com:")).
				Should(Succeed())
		})
	})
})

var crdFooSpec = `
{
  "kind": "CustomResourceDefinition",
  "apiVersion": "apiextensions.k8s.io/v1beta1",
  "metadata": {
    "name": "foos.company.com"
  },
  "spec": {
    "group": "company.com",
    "version": "v1",
    "names": {
      "plural": "foos",
      "kind": "Foo"
    }
  }
}`

var crdBarSpec = `
{
  "kind": "CustomResourceDefinition",
  "apiVersion": "apiextensions.k8s.io/v1beta1",
  "metadata": {
    "name": "bars.company.com"
  },
  "spec": {
    "group": "company.com",
    "version": "v1",
    "names": {
      "plural": "bars",
      "kind": "Bar"
    }
  }
}`

var crdResourcesSpec = `
{
  "kind": "CustomResourceDefinition",
  "apiVersion": "apiextensions.k8s.io/v1beta1",
  "metadata": {
    "name": "resources.mygroup.example.com"
  },
  "spec": {
    "group": "mygroup.example.com",
    "version": "v1alpha1",
    "scope": "Namespaced",
    "names": {
      "plural": "resources",
      "singular": "resource",
      "kind": "Kind",
      "listKind": "KindList"
    }
  }
}`
