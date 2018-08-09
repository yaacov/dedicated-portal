/*
Copyright (c) 2018 Red Hat, Inc.

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

package service

import (
	"fmt"
	"strings"

	"github.com/container-mgmt/dedicated-portal/pkg/api"
	"github.com/golang/glog"
	"github.com/openshift/cluster-operator/pkg/controller"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	v1alpha1 "github.com/openshift/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	clientset "github.com/openshift/cluster-operator/pkg/client/clientset_generated/clientset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capiv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	capiclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

// ClusterProvisioner is the interface used by cluster service to
// provision clusters.
type ClusterProvisioner interface {
	Provision(spec api.Cluster) error
	GetState(id string) (api.ClusterState, error)
}

// ClusterOperatorProvisioner is the struct implementing ClusterProvisioner
// using Cluster Operator.
type ClusterOperatorProvisioner struct {
	clusterOperatorClient *clientset.Clientset
	k8sClient             *kubernetes.Clientset
	clusterAPIClient      *capiclient.Clientset
	awsSecretRef          string
	sshSecretRef          string
}

const clusterNameSpace = "unified-hybrid-cloud"

// NewClusterOperatorProvisioner A constructor for ClusterOperatorProvisioner struct.
func NewClusterOperatorProvisioner(k8sConfig *rest.Config,
	awsSecretRef string, sshSecretRef string) (*ClusterOperatorProvisioner, error) {
	metav1.AddToGroupVersion(scheme.Scheme, schema.GroupVersion{Version: "v1"})
	// Register ClusterDeployment, ClusterVersion, and other CRD's to k8s scheme.
	err := v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("An error occurred trying to add ClusterDeployment to scheme: %s", err)
	}
	clusterOperatorClient, err := clientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cluster operator client: %s", err)
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create kubernetes client: %s", err)
	}

	clusterAPIClient, err := capiclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create cluster operator client: %s", err)
	}

	return &ClusterOperatorProvisioner{
		clusterOperatorClient: clusterOperatorClient,
		k8sClient:             k8sClient,
		clusterAPIClient:      clusterAPIClient,
		awsSecretRef:          awsSecretRef,
		sshSecretRef:          sshSecretRef,
	}, nil
}

// Provision provisions a cluster on aws using cluster operator.
func (p *ClusterOperatorProvisioner) Provision(spec api.Cluster) error {
	// Create cluster version object.
	err := p.createClusterVersionIfNotExist(spec)
	if err != nil {
		return fmt.Errorf("Failed to create ClusterVersion object: %s", err)
	}
	// Create the cluster deployment custom resource
	clusterDeployment := p.clusterDeploymentFromSpec(spec)
	_, err = p.clusterOperatorClient.
		ClusteroperatorV1alpha1().
		ClusterDeployments(clusterNameSpace).
		Create(&clusterDeployment)
	if err != nil {
		return fmt.Errorf("Failed to create ClusterDeployment object: %s", err)
	}
	return nil
}

func (p *ClusterOperatorProvisioner) clusterDeploymentFromSpec(spec api.Cluster) v1alpha1.ClusterDeployment {
	clusterName := strings.ToLower(spec.Name)
	clusterDeploymentSpec := v1alpha1.ClusterDeploymentSpec{
		ClusterName: clusterName,
		ClusterVersionRef: v1alpha1.ClusterVersionReference{
			Name:      "origin-v3-10",
			Namespace: clusterNameSpace,
		},
		NetworkConfig: capiv1.ClusterNetworkingConfig{
			Services: capiv1.NetworkRanges{
				CIDRBlocks: []string{"172.30.0.0/16"},
			},
			Pods: capiv1.NetworkRanges{
				CIDRBlocks: []string{"10.128.0.0/14"},
			},
		},
		Hardware: v1alpha1.ClusterHardwareSpec{
			AWS: &v1alpha1.AWSClusterSpec{
				AccountSecret: corev1.LocalObjectReference{
					Name: p.awsSecretRef,
				},
				SSHSecret: corev1.LocalObjectReference{
					Name: p.sshSecretRef,
				},
				SSHUser: "centos",
				SSLSecret: corev1.LocalObjectReference{
					Name: "default-cluster-tls-certs",
				},
				Region:      spec.Region,
				KeyPairName: "default-ssh-key-pair",
			},
		},
		DefaultHardwareSpec: &v1alpha1.MachineSetHardwareSpec{
			AWS: &v1alpha1.MachineSetAWSHardwareSpec{
				InstanceType: "t2.xlarge",
			},
		},
		MachineSets: p.machineSetsFromSpec(spec),
	}

	clusterDeployment := v1alpha1.ClusterDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "clusteroperator.openshift.io/v1alpha1",
			Kind:       "ClusterDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterNameSpace,
			Labels: map[string]string{
				"uhc.openshift.com/cluster_id": spec.ID,
			},
		},
		Spec: clusterDeploymentSpec,
	}
	return clusterDeployment
}

func (p *ClusterOperatorProvisioner) machineSetsFromSpec(spec api.Cluster) []v1alpha1.ClusterMachineSet {
	infra := v1alpha1.ClusterMachineSet{
		ShortName: "infra",
		MachineSetConfig: v1alpha1.MachineSetConfig{
			Infra:    true,
			Size:     spec.Nodes.Infra,
			NodeType: v1alpha1.NodeTypeCompute,
		},
	}
	compute := v1alpha1.ClusterMachineSet{
		ShortName: "compute",
		MachineSetConfig: v1alpha1.MachineSetConfig{
			Infra:    false,
			Size:     spec.Nodes.Compute,
			NodeType: v1alpha1.NodeTypeCompute,
		},
	}
	master := v1alpha1.ClusterMachineSet{
		MachineSetConfig: v1alpha1.MachineSetConfig{
			Infra:    false,
			Size:     spec.Nodes.Master,
			NodeType: v1alpha1.NodeTypeMaster,
		},
	}
	return []v1alpha1.ClusterMachineSet{master, compute, infra}
}

func (p *ClusterOperatorProvisioner) createClusterVersionIfNotExist(spec api.Cluster) error {
	openshiftAnsibleImage := "registry.svc.ci.openshift.org/openshift-cluster-operator/cluster-operator-ansible:latest"
	clusterAPIImage := "registry.svc.ci.openshift.org/openshift-cluster-operator/kubernetes-cluster-api:latest"
	machineControllerImgae := "registry.svc.ci.openshift.org/openshift-cluster-operator/cluster-operator:latest"
	pullPolicy := corev1.PullIfNotPresent
	clusterVersionName := "origin-v3-10"
	clusterVersion := v1alpha1.ClusterVersion{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "clusteroperator.openshift.io/v1alpha1",
			Kind:       "ClusterVersion",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterVersionName,
			Namespace: clusterNameSpace,
		},
		Spec: v1alpha1.ClusterVersionSpec{
			DeploymentType: v1alpha1.ClusterDeploymentTypeOrigin,
			Version:        "v3.10.0",
			VMImages: v1alpha1.VMImages{
				AWSImages: &v1alpha1.AWSVMImages{
					RegionAMIs: []v1alpha1.AWSRegionAMIs{
						{
							Region: spec.Region,
							AMI:    "ami-0dd8ad483cef75c18",
						},
					},
				},
			},
			Images: v1alpha1.ClusterVersionImages{
				ImageFormat:                      "openshift/origin-${component}:v3.10.0",
				OpenshiftAnsibleImage:            &openshiftAnsibleImage,
				OpenshiftAnsibleImagePullPolicy:  &pullPolicy,
				ClusterAPIImage:                  &clusterAPIImage,
				ClusterAPIImagePullPolicy:        &pullPolicy,
				MachineControllerImage:           &machineControllerImgae,
				MachineControllerImagePullPolicy: &pullPolicy,
			},
		},
	}

	// Attempt to retrieve ClusterVersion object.
	_, err := p.clusterOperatorClient.
		ClusteroperatorV1alpha1().
		ClusterVersions(clusterNameSpace).
		Get(clusterVersionName, metav1.GetOptions{})

	// If ClusterVersion does not exit - create it;
	// Otherwise, return.
	if errors.IsNotFound(err) {
		_, err = p.clusterOperatorClient.
			ClusteroperatorV1alpha1().
			ClusterVersions(clusterNameSpace).
			Create(&clusterVersion)
		if err != nil {
			return err
		}
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return fmt.Errorf("Error getting cluster version %s in namespace %s: %v",
			clusterVersionName, clusterNameSpace, statusError.ErrStatus.Message)
	} else if err != nil {
		return err
	}
	return nil
}

// GetState returns the state of the cluster
func (p *ClusterOperatorProvisioner) GetState(id string) (api.ClusterState, error) {

	labelSelector := fmt.Sprintf("uhc.openshift.com/cluster_id=%s", id)

	// get the clusterdeployment object with the corresponding ID
	clusterDeployments, err := p.clusterOperatorClient.
		ClusteroperatorV1alpha1().
		ClusterDeployments(clusterNameSpace).
		List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		return api.ClusterStateError, err
	}

	// Sanity checks
	if len(clusterDeployments.Items) == 0 {
		// ClusterDeployment object doesn't exist.
		glog.Warningf("Couldn't find cluster deployment object with ID=%s", id)
		return api.ClusterStateError, fmt.Errorf("Couldn't find cluster deployment object with ID=%s", id)
	} else if len(clusterDeployments.Items) > 1 {
		// There should be only one deployment since the ID is unique
		glog.Errorf("Internal error: There is more than one deployment with ID=%s", id)
		return api.ClusterStateError, fmt.Errorf("There is more than one deployment with ID=%s", id)
	}

	clusterDeployment := clusterDeployments.Items[0]

	// Get the cluster that was created by that deployment
	cluster, err := p.clusterAPIClient.
		ClusterV1alpha1().
		Clusters(clusterNameSpace).
		Get(clusterDeployment.Spec.ClusterName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Failed to get cluster named:", clusterDeployment.Spec.ClusterName)
		return api.ClusterStateError, err
	}

	providerStatus, err := controller.ClusterProviderStatusFromCluster(cluster)
	if err != nil {
		fmt.Println("Failed to read cluster status for", clusterDeployment.Spec.ClusterName)
		return api.ClusterStateError, err
	}

	return p.parseClusterStatus(*providerStatus)
}

func (p *ClusterOperatorProvisioner) parseClusterStatus(
	providerStatus v1alpha1.ClusterProviderStatus) (state api.ClusterState, err error) {
	// Currently we just check the "Ready" flag. In the future we should return more detailed data
	if providerStatus.Ready {
		state = api.ClusterStateReady
	} else {
		state = api.ClusterStateInstalling
	}

	return
}
