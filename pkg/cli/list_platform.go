package cli

import (
	"context"
	"strings"
	"time"

	"github.com/loft-sh/log"
	"github.com/loft-sh/vcluster/pkg/cli/flags"
	"github.com/loft-sh/vcluster/pkg/platform"
	"k8s.io/client-go/tools/clientcmd"
)

func ListPlatform(ctx context.Context, options *ListOptions, globalFlags *flags.GlobalFlags, logger log.Logger) error {
	rawConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).RawConfig()
	if err != nil {
		return err
	}
	currentContext := rawConfig.CurrentContext

	if globalFlags.Context == "" {
		globalFlags.Context = currentContext
	}

	platformClient, err := platform.CreatePlatformClient()
	if err != nil {
		return err
	}

	proVClusters, err := platformClient.ListVClusters(ctx, "", "")
	if err != nil {
		return err
	}

	err = printVClusters(ctx, options, proToVClusters(proVClusters, currentContext), globalFlags, false, logger)
	if err != nil {
		return err
	}

	return nil
}

func proToVClusters(vClusters []platform.VirtualClusterInstanceProject, currentContext string) []ListVCluster {
	var output []ListVCluster
	for _, vCluster := range vClusters {
		status := string(vCluster.VirtualCluster.Status.Phase)
		if vCluster.VirtualCluster.DeletionTimestamp != nil {
			status = "Terminating"
		} else if status == "" {
			status = "Pending"
		}

		connected := strings.HasPrefix(currentContext, "vcluster-platform_"+vCluster.VirtualCluster.Name+"_"+vCluster.Project.Name)
		vClusterOutput := ListVCluster{
			Name:       vCluster.VirtualCluster.Spec.ClusterRef.VirtualCluster,
			Namespace:  vCluster.VirtualCluster.Spec.ClusterRef.Namespace,
			Connected:  connected,
			Created:    vCluster.VirtualCluster.CreationTimestamp.Time,
			AgeSeconds: int(time.Since(vCluster.VirtualCluster.CreationTimestamp.Time).Round(time.Second).Seconds()),
			Status:     status,
			Version:    vCluster.VirtualCluster.Status.VirtualCluster.HelmRelease.Chart.Version,
		}
		output = append(output, vClusterOutput)
	}
	return output
}
