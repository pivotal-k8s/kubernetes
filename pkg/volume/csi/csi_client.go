/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package csi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	csipb "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	api "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
)

type csiClient interface {
	NodeGetInfo(ctx context.Context) (
		nodeID string,
		maxVolumePerNode int64,
		accessibleTopology *csipb.Topology,
		err error)
	NodePublishVolume(
		ctx context.Context,
		volumeid string,
		readOnly bool,
		stagingTargetPath string,
		targetPath string,
		accessMode api.PersistentVolumeAccessMode,
		volumeInfo map[string]string,
		volumeAttribs map[string]string,
		nodePublishSecrets map[string]string,
		fsType string,
	) error
	NodeUnpublishVolume(
		ctx context.Context,
		volID string,
		targetPath string,
	) error
	NodeStageVolume(ctx context.Context,
		volID string,
		publishVolumeInfo map[string]string,
		stagingTargetPath string,
		fsType string,
		accessMode api.PersistentVolumeAccessMode,
		nodeStageSecrets map[string]string,
		volumeAttribs map[string]string,
	) error
	NodeUnstageVolume(ctx context.Context, volID, stagingTargetPath string) error
	NodeGetCapabilities(ctx context.Context) ([]*csipb.NodeServiceCapability, error)
}

// csiClient encapsulates all csi-plugin methods
type csiDriverClient struct {
	driverName string
	nodeClient csipb.NodeClient
}

var _ csiClient = &csiDriverClient{}

func newCsiDriverClient(driverName string) *csiDriverClient {
	c := &csiDriverClient{driverName: driverName}
	return c
}

func (c *csiDriverClient) NodeGetInfo(ctx context.Context) (
	nodeID string,
	maxVolumePerNode int64,
	accessibleTopology *csipb.Topology,
	err error) {

	return "", 0, &csipb.Topology{}, errors.New("NOOOO")
}

func (c *csiDriverClient) NodePublishVolume(
	ctx context.Context,
	volID string,
	readOnly bool,
	stagingTargetPath string,
	targetPath string,
	accessMode api.PersistentVolumeAccessMode,
	volumeInfo map[string]string,
	volumeAttribs map[string]string,
	nodePublishSecrets map[string]string,
	fsType string,
) error {
	return nil
}

func (c *csiDriverClient) NodeUnpublishVolume(ctx context.Context, volID string, targetPath string) error {

	return nil
}

func (c *csiDriverClient) NodeStageVolume(ctx context.Context,
	volID string,
	publishInfo map[string]string,
	stagingTargetPath string,
	fsType string,
	accessMode api.PersistentVolumeAccessMode,
	nodeStageSecrets map[string]string,
	volumeAttribs map[string]string,
) error {

	return nil
}

func (c *csiDriverClient) NodeUnstageVolume(ctx context.Context, volID, stagingTargetPath string) error {

	return nil
}

func (c *csiDriverClient) NodeGetCapabilities(ctx context.Context) ([]*csipb.NodeServiceCapability, error) {
	glog.V(4).Info(log("calling NodeGetCapabilities rpc"))

	conn, err := newGrpcConn(c.driverName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	nodeClient := csipb.NewNodeClient(conn)

	req := &csipb.NodeGetCapabilitiesRequest{}
	resp, err := nodeClient.NodeGetCapabilities(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetCapabilities(), nil
}

func asCSIAccessMode(am api.PersistentVolumeAccessMode) csipb.VolumeCapability_AccessMode_Mode {
	switch am {
	case api.ReadWriteOnce:
		return csipb.VolumeCapability_AccessMode_SINGLE_NODE_WRITER
	case api.ReadOnlyMany:
		return csipb.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY
	case api.ReadWriteMany:
		return csipb.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER
	}
	return csipb.VolumeCapability_AccessMode_UNKNOWN
}

func newGrpcConn(driverName string) (*grpc.ClientConn, error) {
	if driverName == "" {
		return nil, fmt.Errorf("driver name is empty")
	}
	addr := fmt.Sprintf(csiAddrTemplate, driverName)
	// TODO once KubeletPluginsWatcher graduates to beta, remove FeatureGate check
	if utilfeature.DefaultFeatureGate.Enabled(features.KubeletPluginsWatcher) {
		driver, ok := csiDrivers.driversMap[driverName]
		if !ok {
			return nil, fmt.Errorf("driver name %s not found in the list of registered CSI drivers", driverName)
		}
		addr = driver.driverEndpoint
	}
	network := "unix"
	glog.V(4).Infof(log("creating new gRPC connection for [%s://%s]", network, addr))

	return grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithDialer(func(target string, timeout time.Duration) (net.Conn, error) {
			return net.Dial(network, target)
		}),
	)
}
