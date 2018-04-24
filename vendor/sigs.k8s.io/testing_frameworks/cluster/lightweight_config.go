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
	"io"
	"net/url"
	"time"
)

// LightWeightMasterConfig is a struct used as a nested struct in `Config` to
// add aditional configuration properties needed by the "lightweight"
// implementation to the main `Config` struct.
type LightWeightMasterConfig struct {
	// APIServerProcessConfig hold configuration properties related to the
	// APIServer process.
	APIServerProcessConfig ProcessConfig
}

// LightWeightEtcd is a struct used as a nested struct in `Etcd` to
// add aditional configuration properties needed by the "lightweight"
// implementation to the main `Etcd` struct.
type LightWeightEtcd struct {
	// ProcessConfig holds configuration properties releated to the Etcd progress
	ProcessConfig ProcessConfig

	// BindURL is the URL Etcd should bind to
	BindURL *url.URL
}

// LightWeightAPI is a struct used as a nested struct in `API` to
// add aditional configuration properties needed by the "lightweight"
// implementation to the main `API` struct.
type LightWeightAPI struct {
	// BindURL is a URL the API should listen on.
	//
	// If this is kept empty, it will be defaulted to a free port on "localhost".
	BindURL *url.URL
}

// ProcessConfig is configuring certain properties for processes.
type ProcessConfig struct {
	// Path is the path to the binary.
	//
	// If this is left as the empty string, we will attempt to locate a binary,
	// by checking for the TEST_ASSET_KUBE_<COMPONENT> (e.g.
	// TEST_ASSET_KUBE_APISERVER, TEST_ASSET_KUBE_ETCD, ...) environment
	// variable, and the default test assets directory. See the "Binaries"
	// section above (in doc.go) for details.
	Path string

	// StartTimeout, StopTimeout specify the time the the process is allowed to
	// take when starting and stoppping before an error is emitted.
	//
	// If not specified, these default to 20 seconds.
	StartTimeout time.Duration
	StopTimeout  time.Duration

	// Out, Err specify where the process should write its StdOut, StdErr to.
	//
	// If not specified, the output will be discarded.
	Out io.Writer
	Err io.Writer
}
