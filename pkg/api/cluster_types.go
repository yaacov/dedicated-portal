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

// This file contains the API types used by the clusters service.

package api

// ClusterState represents the state of a cluster
type ClusterState string

const (
	// ClusterStateInstalling - the cluster is still installing
	ClusterStateInstalling ClusterState = "Installing"
	// ClusterStateReady - cluster is ready for use
	ClusterStateReady ClusterState = "Ready"
	// ClusterStateError - error during installation
	ClusterStateError ClusterState = "Error"
)

// Cluster represents a cluster.
type Cluster struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Region  string          `json:"region"`
	Nodes   ClusterNodes    `json:"nodes"`
	Memory  ClusterResource `json:"memory"`
	CPU     ClusterResource `json:"cpu"`
	Storage ClusterResource `json:"storage"`
	State   ClusterState    `json:"state"`
}

// ClusterStatus represents the status of a cluster
type ClusterStatus struct {
	ID    string       `json:"id,omitempty"`
	State ClusterState `json:"state,omitempty"`
}

// ClusterNodes represents the node count inside a cluster.
type ClusterNodes struct {
	Total   int `json:"total"`
	Master  int `json:"master"`
	Infra   int `json:"infra"`
	Compute int `json:"compute"`
}

// ClusterList is a list of clusters.
type ClusterList struct {
	Page  int        `json:"page"`
	Size  int        `json:"size"`
	Total int        `json:"total"`
	Items []*Cluster `json:"items"`
}

// ClusterResource represents a resource availability in the cluster.
type ClusterResource struct {
	Used  int `json:"used"`
	Total int `json:"total"`
}
