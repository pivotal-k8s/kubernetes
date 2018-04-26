package kubectl_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/testing_frameworks/cluster"
)

const (
	CONTEXT = "test"
	CLUSTER = "test_cluser"
)

func TestKubectl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubectl Suite")
}

var (
	kubeCtl       KubeCtl
	clusterConfig cluster.Config
	configCleanup configCleaner
	configUpdate  configUpdater
)

var _ = BeforeSuite(func() {
	kubeCtl, configUpdate, configCleanup = createTestEnvironment()
	clusterConfig = getClusterConfig()
})

var _ = AfterSuite(func() {
	configCleanup()
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

func getClusterConfig() cluster.Config {
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
	clusterConfig.APIServerProcessConfig.Path = getK8sPath("kube-apiserver")
	clusterConfig.Etcd.ProcessConfig.Path = getEtcdPath()

	return clusterConfig
}

type configCleaner func()
type configUpdater func(f cluster.Fixture)

func createTestEnvironment() (KubeCtl, configUpdater, configCleaner) {
	tmpFile, err := ioutil.TempFile("", "test_kube_conf_")
	Expect(err).NotTo(HaveOccurred())
	kubeConfigPath := tmpFile.Name()
	Expect(tmpFile.Close()).To(Succeed())

	k := KubeCtl{
		Path:       getK8sPath("kubectl"),
		KubeConfig: kubeConfigPath,
	}
	k.WithArgs("config", "set-context", CONTEXT, "--cluster", CLUSTER).Should(Succeed())
	k.WithArgs("config", "use-context", CONTEXT).Should(Succeed())

	updater := func(f cluster.Fixture) {
		// TODO extend when Fixture.ClientConfig() returns a rest.Config or such
		server := f.ClientConfig().String()
		k.WithArgs("config", "set-cluster", CLUSTER, "--server", server).Should(Succeed())
	}

	cleaner := func() {
		Expect(os.RemoveAll(kubeConfigPath)).To(Succeed())
	}

	return k, updater, cleaner
}

type KubeCtl struct {
	Path       string
	KubeConfig string

	args         []string
	outMatcher   types.GomegaMatcher
	errMatcher   types.GomegaMatcher
	outputFormat outputFormat
	namespace    string
}

func (k KubeCtl) run() (io.Reader, io.Reader, error) {
	if k.Path == "" {
		k.Path = "kubectl"
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runArgs := append(k.connectionArgs(), k.args...)

	if k.namespace != "" {
		runArgs = append(runArgs, "--namespace", k.namespace)
	}

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
	if c := k.KubeConfig; c != "" {
		connArgs = []string{"--kubeconfig", c}
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

type Namespace struct {
	Name       string
	AutoCreate bool
}

func (k KubeCtl) WithNamespace(ns Namespace) KubeCtl {
	if ns.Name == "" {
		ns.Name = createRandomNamespaceName()
	}

	if ns.AutoCreate {
		out, _, _ := k.WithArgs("get", "namespace", ns.Name, "-o", "yaml").run()
		if toString(out) == "" {
			k.WithArgs("create", "namespace", ns.Name).Should(Succeed())
		}
	}

	k.namespace = ns.Name

	return k
}

func createRandomNamespaceName() string {
	allowedChars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	postfix := randomString(5, allowedChars...)
	timestamp := time.Now().UnixNano()
	namespace := fmt.Sprintf("%s-%d-%s", "namespace", timestamp, postfix)
	return namespace
}

func createAndUseNamespace() string {
	namespace := createRandomNamespaceName()

	kubeCtl.WithArgs("create", "namespace", namespace).Should(Succeed())
	kubeCtl.WithArgs("config", "set-context", CONTEXT, "--namespace", namespace).Should(Succeed())

	return namespace
}

func randomString(n int, allowedChars ...rune) string {
	if len(allowedChars) == 0 {
		allowedChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	}

	b := make([]rune, n)
	for i := range b {
		b[i] = allowedChars[rand.Intn(len(allowedChars))]
	}
	return string(b)
}
