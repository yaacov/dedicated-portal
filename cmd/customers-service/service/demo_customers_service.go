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

// DemoCustomersService is a struct implementing the customer service interface.
type DemoCustomersService struct {
}

// NewDemoCustomersService is a constructor for the SQLCustomersService struct.
func NewDemoCustomersService() (*DemoCustomersService, error) {
	service := new(DemoCustomersService)
	return service, nil
}

// Close closes the sql customers service client.
func (s *DemoCustomersService) Close() error {
	return nil
}

// Add adds a single customer to psql database.
func (s *DemoCustomersService) Add(customer api.Customer) (*api.Customer, error) {
	return &api.Customer{}, nil
}

// Get retrieves a single customer from psql database.
func (s *DemoCustomersService) Get(id string) (*api.Customer, error) {
	return &api.Customer{
		ID:            "UNIQEID",
		Name:          "Ari",
		OwnedClusters: []string{"Eeast_cluster", "West_cluster"},
	}, nil
}

// List retrieves a list of current customers stored in datastore.
func (s *DemoCustomersService) List(args *ListArguments) (*api.CustomerList, error) {
	result := &api.CustomerList{
		Items: []*api.Customer{
			{
				ID:            "UNIQEID_1",
				Name:          "Ari",
				OwnedClusters: []string{"Eeast_cluster", "West_cluster"},
			},
			{
				ID:            "UNIQEID_2",
				Name:          "Bni",
				OwnedClusters: []string{"West_cluster"},
			},
		},
		Page:  0,
		Size:  10,
		Total: 2,
	}

	return result, nil
}
