package lightweight

import (
	"net/url"

	"sigs.k8s.io/testing_frameworks/cluster"
	"sigs.k8s.io/testing_frameworks/lightweight/internal"
)

// Etcd knows how to run an etcd server.
type Etcd struct {
	// ClusterConfig is the kubeadm-compatible configuration for
	// clusters, which is partially supported by this framework.
	ClusterConfig cluster.Config

	processState *internal.ProcessState
}

// Start starts the etcd, waits for it to come up, and returns an error, if one
// occoured.
func (e *Etcd) Start() error {
	var err error

	e.processState = &internal.ProcessState{}

	e.processState.DefaultedProcessInput, err = internal.DoDefaulting(
		"etcd",
		e.ClusterConfig.Etcd.BindURL,
		e.ClusterConfig.Etcd.DataDir,
		e.ClusterConfig.Etcd.ProcessConfig.Path,
		e.ClusterConfig.Etcd.ProcessConfig.StartTimeout,
		e.ClusterConfig.Etcd.ProcessConfig.StopTimeout,
	)
	if err != nil {
		return err
	}

	e.processState.StartMessage = internal.GetEtcdStartMessage(e.processState.URL)

	tmplData := struct {
		URL     *url.URL
		DataDir string
	}{
		&e.processState.URL,
		e.processState.Dir,
	}

	args := flattenArgs(e.ClusterConfig.Etcd.ExtraArgs)

	e.processState.Args, err = internal.RenderTemplates(
		internal.DoEtcdArgDefaulting(args), tmplData,
	)
	if err != nil {
		return err
	}

	return e.processState.Start(
		e.ClusterConfig.Etcd.ProcessConfig.Out,
		e.ClusterConfig.Etcd.ProcessConfig.Err,
	)
}

// Stop stops this process gracefully, waits for its termination, and cleans up
// the DataDir if necessary.
func (e *Etcd) Stop() error {
	return e.processState.Stop()
}
