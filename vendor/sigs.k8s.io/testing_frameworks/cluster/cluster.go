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

import (
	"net/url"
)

// Fixture is some kind of test cluster fixture, which can be started, interacted with, and stopped.
type Fixture interface {
	// Setup starts the test cluster according to the provided
	// configuration. If the config asks for a feature that is not
	// supported by this Fixture implementation, or if the test
	// cluster fails to start, return an error.
	//
	// This should block until the test cluster control plane is
	// ready to recieve client connections.
	Setup(config Config) error

	// TearDown cleanly stops the test cluster. If we can't stop
	// cleanly, return an error.
	//
	// TearDown should block until the test cluster has stopped,
	// or we have given up on stopping it and returned an error.
	//
	// TearDown should be idempotent. If a user calls TearDown
	// twice in a row, and the first call succeded, then the
	// second call should also succeed.
	TearDown() error

	// ClientConfig returns the URL at which you can find the APIServer
	// of the test cluster. In future, this will likely become a
	// rest.Config from
	// https://github.com/kubernetes/client-go/blob/master/rest/config.go.
	ClientConfig() *url.URL
}

// Config is a struct into which you can parse a YAML or JSON config
// file (which should always be compatible with
// https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init/#config-file
// ) to describe your test cluster.
//
// To maintain compatibility with kubeadm, we should follow the
// patterns established in
// https://github.com/kubernetes/kubernetes/blob/c8cded58d71e36665bd345a70fbe404e7523abb8/cmd/kubeadm/app/apis/kubeadm/types.go#L30
type Config struct {
	// Etcd holds configuration for etcd.
	Etcd Etcd

	// API holds configuration for the k8s apiserver.
	API API

	// KubeadmMasterConfig is the nested struct holding all the configuration
	// supported by kubeadm
	KubeadmMasterConfig

	// LightWeightMasterConfig is the nested struct holding all the configuration
	// additionally supported by the "lightweight" framework
	LightWeightMasterConfig
}

// Etcd contains elements describing Etcd configuration.
// See also https://github.com/kubernetes/kubernetes/blob/c8cded58d71e36665bd345a70fbe404e7523abb8/cmd/kubeadm/app/apis/kubeadm/types.go#L163
type Etcd struct {
	// KubeadmEtcd is the nexted struct holding all the configuration for etcd
	// supported by kubeadm
	KubeadmEtcd

	// LightWeightEtcd is the nested struct holding all the configuration
	// additionally supported by the "lightweight" framework
	LightWeightEtcd
}

// API struct contains elements of API server address.
type API struct {
	// KubeadmAPI is the nested struct holding all the configuration for the API
	// server address supported by kubeadm.
	KubeadmAPI

	// LightWeightAPI is the nested struct holding all addirional configuration
	// for the API server adress supported by the "lightweight" framework.
	LightWeightAPI
}
