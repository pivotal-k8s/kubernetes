package kubectl_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/client-go/rest"
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

type KubeCtl struct {
	Path             string
	ConnectionConfig rest.Config

	args         []string
	outMatcher   types.GomegaMatcher
	errMatcher   types.GomegaMatcher
	outputFormat outputFormat
}

func (k KubeCtl) run() (io.Reader, io.Reader, error) {
	if k.Path == "" {
		k.Path = "kubectl"
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runArgs := append(k.connectionArgs(), k.args...)

	if f := k.outputFormat; f != (outputFormat{}) {
		runArgs = append(runArgs, "-o", fmt.Sprintf("%s=%s", f.format, f.template))
	}

	cmd := exec.Command(k.Path, runArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	return stdout, stderr, err
}

func (k KubeCtl) connectionArgs() []string {
	connArgs := []string{}
	if host := k.ConnectionConfig.Host; host != "" {
		connArgs = []string{"-s", host}
	}
	return connArgs
}

func (k KubeCtl) should(matcher types.GomegaMatcher) {
	stdoutReader, stderrReader, err := k.run()

	stdout, stderr := toString(stdoutReader), toString(stderrReader)

	if m := k.outMatcher; m != nil {
		ExpectWithOffset(2, stdout).To(m)
	}

	if m := k.errMatcher; m != nil {
		ExpectWithOffset(2, stderr).To(m)
	}

	ExpectWithOffset(2, err).To(
		matcher,
		"---[ stdout ]---\n%s\n---[ stderr ]---\n%s\n----------------\n", stdout, stderr,
	)
}

func (k KubeCtl) Should(matcher types.GomegaMatcher) {
	k.should(matcher)
}
func (k KubeCtl) To(matcher types.GomegaMatcher) {
	k.should(matcher)
}

func (k KubeCtl) ShouldNot(matcher types.GomegaMatcher) {
	k.should(Not(matcher))
}
func (k KubeCtl) NotTo(matcher types.GomegaMatcher) {
	k.should(Not(matcher))
}

func (k KubeCtl) ExpectStderrTo(matcher types.GomegaMatcher) KubeCtl {
	k.errMatcher = matcher
	return k
}
func (k KubeCtl) ExpectStderrNotTo(matcher types.GomegaMatcher) KubeCtl {
	return k.ExpectStderrTo(Not(matcher))
}

func (k KubeCtl) ExpectStdoutTo(matcher types.GomegaMatcher) KubeCtl {
	k.outMatcher = matcher
	return k
}
func (k KubeCtl) ExpectStdoutNotTo(matcher types.GomegaMatcher) KubeCtl {
	return k.ExpectStdoutTo(Not(matcher))
}

func (k KubeCtl) WithFormat(f outputFormat) KubeCtl {
	k.outputFormat = f
	return k
}

func (k KubeCtl) WithArgs(args ...string) KubeCtl {
	k.args = append(k.args, args...)
	return k
}

func toString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}

type outputFormatType string

const (
	goTemplate outputFormatType = "go-template"
	jsonPath   outputFormatType = "jsonpath"
)

type outputFormat struct {
	format   outputFormatType
	template string
}

func GoTemplate(tmpl string) outputFormat {
	return outputFormat{
		format:   goTemplate,
		template: tmpl,
	}
}

func JsonPath(tmpl string) outputFormat {
	return outputFormat{
		format:   jsonPath,
		template: tmpl,
	}
}
