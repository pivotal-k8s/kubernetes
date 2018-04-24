package lightweight

import (
	"fmt"
	"net/url"

	"sigs.k8s.io/testing_frameworks/cluster"
	"sigs.k8s.io/testing_frameworks/lightweight/internal"
)

// APIServer knows how to run a kubernetes apiserver.
type APIServer struct {
	// ClusterConfig is the kubeadm-compatible configuration for
	// clusters, which is partially supported by this framework.
	ClusterConfig cluster.Config

	// EtcdURL is the URL of the Etcd the APIServer should use.
	//
	// If this is not specified, the Start() method will return an error.
	EtcdURL *url.URL

	processState *internal.ProcessState
}

// Start starts the apiserver, waits for it to come up, and returns an error,
// if occurred.
func (s *APIServer) Start() error {
	if s.EtcdURL == nil {
		return fmt.Errorf("expected EtcdURL to be configured")
	}

	var err error

	s.processState = &internal.ProcessState{}

	s.processState.DefaultedProcessInput, err = internal.DoDefaulting(
		"kube-apiserver",
		s.ClusterConfig.API.BindURL,
		s.ClusterConfig.CertificatesDir,
		s.ClusterConfig.APIServerProcessConfig.Path,
		s.ClusterConfig.APIServerProcessConfig.StartTimeout,
		s.ClusterConfig.APIServerProcessConfig.StopTimeout,
	)
	if err != nil {
		return err
	}

	s.processState.HealthCheckEndpoint = "/healthz"

	tmplData := struct {
		EtcdURL *url.URL
		URL     *url.URL
		CertDir string
	}{
		s.EtcdURL,
		&s.processState.URL,
		s.processState.Dir,
	}

	args := flattenArgs(s.ClusterConfig.APIServerExtraArgs)

	s.processState.Args, err = internal.RenderTemplates(
		internal.DoAPIServerArgDefaulting(args), tmplData,
	)
	if err != nil {
		return err
	}

	return s.processState.Start(
		s.ClusterConfig.APIServerProcessConfig.Out,
		s.ClusterConfig.APIServerProcessConfig.Err,
	)
}

// Stop stops this process gracefully, waits for its termination, and cleans up
// the CertDir if necessary.
func (s *APIServer) Stop() error {
	return s.processState.Stop()
}

func flattenArgs(mappedArgs map[string]string) []string {
	args := []string{}
	for k, v := range mappedArgs {
		if v == "" {
			args = append(args, k)
		} else {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return args
}
