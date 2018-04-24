// Package cluster is a Test Cluster Framework -- see the motivating
// document here:
// https://docs.google.com/document/d/13bMjmWpsdkgbY-JayrcU-e_QNwRJCP-rHjtqdeeoQHo/edit?ts=5aa005c9#heading=h.75awtuvlo3ad
//
// This package aims to provide a consistent abstraction over a number
// of different ways of creating kubernetes clusters for testing.
//
// To make your test cluster compatible with this framework, just
// implement the Fixture interface.  If you require more config
// options than we have so far, add them to the Config struct. Bear in
// mind while doing so that we aim to maintain compatibility with
// kubeadm, so if your new config options describe properties that
// already exist in the kubeadm config, you should use
// kubeadm-compatibile terminology.
package cluster

// KubeadmMasterConfig is a struct which is used as a nested struct in `Config`.
//
// We use this indirection to be prepared if we'd be vendored into k/k -- then
// we could remove `KubeadmMasterConfig` and use the actual one from within k/k.
type KubeadmMasterConfig struct {
	// CertificatesDir specifies where to store or look for all required certificates.
	// See also https://github.com/kubernetes/kubernetes/blob/c8cded58d71e36665bd345a70fbe404e7523abb8/cmd/kubeadm/app/apis/kubeadm/types.go#L104
	CertificatesDir string

	// APIServerExtraArgs is a set of extra flags to pass to the API Server or override
	// default ones in form of <flagname>=<value>.
	// See also https://github.com/kubernetes/kubernetes/blob/c8cded58d71e36665bd345a70fbe404e7523abb8/cmd/kubeadm/app/apis/kubeadm/types.go#L81
	APIServerExtraArgs map[string]string
}

// KubeadmEtcd is a struct which is used as a nested struct in `Etcd`.
//
// We use this indirection to be prepared if we'd be vendored into k/k -- then
// we could remove `KubeadmEtcd` and use the actual one from within k/k.
type KubeadmEtcd struct {
	// DataDir is the directory etcd will place its data.
	// Defaults to "/var/lib/etcd".
	DataDir string

	// ExtraArgs are extra arguments provided to the etcd binary
	// when run inside a static pod.
	ExtraArgs map[string]string
}

// KubeadmAPI is a struct which is used as a nested struct in `API`.
//
// We use this indirection to be prepared if we'd be vendored into k/k -- then
// we could remove `KubeadmEtcd` and use the actual one from within k/k.
//
// For now this is empty, as none of the currently supported framework supports
// any of these settings.
type KubeadmAPI struct{}
