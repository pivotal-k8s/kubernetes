package kubectl_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubectl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubectl Suite")
}

var (
	etcdPath      string
	apiServerPath string
	kubeCtlPath   string
)

var _ = BeforeSuite(func() {
	etcdPath = getEtcdPath()
	apiServerPath = getK8sPath("kube-apiserver")
	kubeCtlPath = getK8sPath("kubectl")
})

func getK8sPath(name string) string {
	return resolveToExecutable(
		filepath.Join(getKubeRoot(), "_output", "bin", name),
		fmt.Sprintf("Have you run `make WHAT=\"cmd/%s\"`?", name),
	)
}

func getEtcdPath() string {
	return resolveToExecutable(
		filepath.Join(getKubeRoot(), "third_party", "etcd", "etcd"),
		"Have you run `./hack/install-etcd.sh`?",
	)
}

func getKubeRoot() string {
	_, filename, _, ok := runtime.Caller(1)
	Expect(ok).To(BeTrue())
	return cdUp(filepath.Dir(filename), 2)
}

func cdUp(path string, count int) string {
	for i := 0; i < count; i++ {
		path = filepath.Dir(path)
	}
	return path
}

func resolveToExecutable(path, message string) string {
	Expect(path).To(BeAnExistingFile(),
		fmt.Sprintf("Expected to find a binary at '%s'. %s", path, message),
	)

	realBin, err := filepath.EvalSymlinks(path)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Could not find link target for '%s'", path),
	)

	stat, err := os.Stat(realBin)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Could not get permissions for '%s'", realBin),
	)

	isExecutable := ((stat.Mode() | 0111) != 0)
	Expect(isExecutable).To(BeTrue(),
		fmt.Sprintf("'%s' is not executable", realBin),
	)

	return realBin
}
