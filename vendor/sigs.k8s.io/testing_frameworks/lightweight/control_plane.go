package lightweight

import (
	"fmt"
	"net/url"

	"sigs.k8s.io/testing_frameworks/cluster"
)

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in
// future.
type ControlPlane struct {
	APIServer *APIServer
	Etcd      *Etcd
}

// Setup will start your control plane processes according to the
// supplied configuration.
func (f *ControlPlane) Setup(config cluster.Config) error {
	f.Etcd = &Etcd{}
	f.Etcd.ClusterConfig = config

	if f.APIServer == nil {
		f.APIServer = &APIServer{}
	}
	f.APIServer.ClusterConfig = config

	return f.Start()
}

// Start will start your control plane processes. To stop them, call Stop().
func (f *ControlPlane) Start() error {
	if f.Etcd == nil {
		f.Etcd = &Etcd{}
	}
	if err := f.Etcd.Start(); err != nil {
		return err
	}

	if f.APIServer == nil {
		f.APIServer = &APIServer{}
	}
	f.APIServer.EtcdURL = &f.Etcd.processState.URL
	return f.APIServer.Start()
}

// TearDown will stop your control plane processes. This is an alias for Stop()
func (f *ControlPlane) TearDown() error {
	return f.Stop()
}

// Stop will stop your control plane processes, and clean up their data.
func (f *ControlPlane) Stop() error {
	if f.APIServer != nil {
		if err := f.APIServer.Stop(); err != nil {
			return err
		}
	}
	if f.Etcd != nil {
		if err := f.Etcd.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// ClientConfig returns the URL of your APIServer. This is an alias for APIURL()
func (f *ControlPlane) ClientConfig() *url.URL {
	return f.APIURL()
}

// APIURL returns the URL you should connect to to talk to your API.
func (f *ControlPlane) APIURL() *url.URL {
	return &f.APIServer.processState.URL
}

// KubeCtl returns a pre-configured KubeCtl, ready to connect to this
// ControlPlane.
func (f *ControlPlane) KubeCtl() *KubeCtl {
	k := &KubeCtl{}
	k.Opts = append(k.Opts, fmt.Sprintf("--server=%s", f.APIURL()))
	return k
}
