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
	"github.com/container-mgmt/dedicated-portal/pkg/api"
)

// ListArguments are arguments relevant for listing objects.
type ListArguments struct {
	Page int
	Size int
}

// ClustersService performs operations on clusters.
type ClustersService interface {
	List(args ListArguments) (clusters api.ClusterList, err error)
	Create(spec api.Cluster, provision bool) (result api.Cluster, err error)
	Get(id string) (result api.Cluster, err error)
	GetStatus(id string) (result api.ClusterStatus, err error)
}
