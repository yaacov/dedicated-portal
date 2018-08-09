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
	"database/sql"
	"fmt"

	// Register the postgresql sql backend.
	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"

	"github.com/container-mgmt/dedicated-portal/pkg/api"
)

// SQLClustersService is a ClusterService placeholder implementation.
type SQLClustersService struct {
	provisioner ClusterProvisioner
	db          *sql.DB
}

// NewSQLClustersService is a constructor for the SQLCustomersService struct.
func NewSQLClustersService(connStr string, provisioner ClusterProvisioner) (*SQLClustersService, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set DB
	service := new(SQLClustersService)
	service.db = db

	// Set provisioner
	service.provisioner = provisioner

	return service, nil
}

// Close closes the sql clusters service client.
func (s *SQLClustersService) Close() error {
	return s.db.Close()
}

// List returns lists of clusters.
func (s SQLClustersService) List(args ListArguments) (result api.ClusterList, err error) {
	result.Items = make([]*api.Cluster, 0)
	total, err := s.getClusterCount()
	if err != nil {
		return api.ClusterList{}, err
	}

	rows, err := s.db.Query(`SELECT
		id,
		name,
		region,
		master_nodes,
		infra_nodes,
		compute_nodes,
		memory,
		cpu_cores,
		storage,
		state
		FROM clusters
		ORDER BY id
		LIMIT $1
		OFFSET $2`,
		args.Size,
		args.Page*args.Size,
	)
	if err != nil {
		return api.ClusterList{}, fmt.Errorf("Error executing query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id           string
			name         string
			region       string
			masterNodes  int
			infraNodes   int
			computeNodes int
			memory       int
			cpuCores     int
			storage      int
			state        api.ClusterState
		)

		err = rows.Scan(
			&id,
			&name,
			&region,
			&masterNodes,
			&infraNodes,
			&computeNodes,
			&memory,
			&cpuCores,
			&storage,
			&state)
		if err != nil {
			return api.ClusterList{}, err
		}

		totalNodes := masterNodes + infraNodes + computeNodes
		result.Items = append(result.Items, &api.Cluster{
			Name:   name,
			Region: region,
			ID:     id,
			Nodes: api.ClusterNodes{
				Total:   totalNodes,
				Master:  masterNodes,
				Infra:   infraNodes,
				Compute: computeNodes,
			},
			Memory: api.ClusterResource{
				Total: memory,
				Used:  0,
			},
			CPU: api.ClusterResource{
				Total: cpuCores,
				Used:  0,
			},
			Storage: api.ClusterResource{
				Total: storage,
				Used:  0,
			},
			State: state,
		})
	}
	err = rows.Err() // get any error encountered during iteration
	if err != nil {
		return api.ClusterList{}, err
	}
	result.Page = args.Page
	result.Size = len(result.Items)
	result.Total = total
	return result, nil
}

// Create saves a new cluster definition in the Database
func (s SQLClustersService) Create(spec api.Cluster, provision bool) (result api.Cluster, err error) {
	id, err := ksuid.NewRandom()
	if err != nil {
		return api.Cluster{}, err
	}

	stmt, err := s.db.Prepare(`
		INSERT INTO clusters (
			id,
			name,
			region,
			master_nodes,
			infra_nodes,
			compute_nodes,
			memory,
			cpu_cores,
			storage,
			state
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10)
	`)
	if err != nil {
		return api.Cluster{}, err
	}
	defer stmt.Close()
	queryResult, err := stmt.Exec(
		id,
		spec.Name,
		spec.Region,
		spec.Nodes.Master,
		spec.Nodes.Infra,
		spec.Nodes.Compute,
		spec.Memory.Total,
		spec.CPU.Total,
		spec.Storage.Total,
		api.ClusterStateInstalling)
	if err != nil {
		return api.Cluster{}, err
	}
	inserted, err := queryResult.RowsAffected()
	if err != nil {
		return api.Cluster{}, err
	}
	if inserted != 1 {
		return api.Cluster{}, fmt.Errorf("Error: [%d] rows inserted. 1 expected",
			inserted,
		)
	}

	totalNodes := spec.Nodes.Master + spec.Nodes.Infra + spec.Nodes.Compute
	newCluster := api.Cluster{
		Name:   spec.Name,
		ID:     id.String(),
		Region: spec.Region,
		Nodes: api.ClusterNodes{
			Total:   totalNodes,
			Master:  spec.Nodes.Master,
			Infra:   spec.Nodes.Infra,
			Compute: spec.Nodes.Compute,
		},
		Memory: api.ClusterResource{
			Total: spec.Memory.Total,
			Used:  0,
		},
		CPU: api.ClusterResource{
			Total: spec.CPU.Total,
			Used:  0,
		},
		Storage: api.ClusterResource{
			Total: spec.Storage.Total,
			Used:  0,
		},
		State: api.ClusterStateInstalling,
	}

	if provision {
		// Use cluster provisioner to Provision a cluster.
		err = s.provisioner.Provision(newCluster)
		if err != nil {
			return api.Cluster{}, fmt.Errorf("An error occurred while trying to provision cluster %s: %s",
				spec.Name, err)
		}
	}

	return newCluster, nil

}

// Get returns a single cluster by id
func (s SQLClustersService) Get(id string) (result api.Cluster, err error) {
	var (
		name         string
		region       string
		masterNodes  int
		infraNodes   int
		computeNodes int
		memory       int
		cpuCores     int
		storage      int
		state        api.ClusterState
	)

	err = s.db.QueryRow(`
	SELECT
		id,
		name,
		region,
		master_nodes,
		infra_nodes,
		compute_nodes,
		memory,
		cpu_cores,
		storage,
		state
	FROM clusters
	WHERE id = $1`, id).Scan(
		&id,
		&name,
		&region,
		&masterNodes,
		&infraNodes,
		&computeNodes,
		&memory,
		&cpuCores,
		&storage,
		&state)

	if err != nil {
		return api.Cluster{}, err
	}
	totalNodes := masterNodes + infraNodes + computeNodes
	return api.Cluster{
			Name:   name,
			Region: region,
			ID:     id,
			Nodes: api.ClusterNodes{
				Total:   totalNodes,
				Master:  masterNodes,
				Infra:   infraNodes,
				Compute: computeNodes,
			},
			Memory: api.ClusterResource{
				Total: memory,
				Used:  0,
			},
			CPU: api.ClusterResource{
				Total: cpuCores,
				Used:  0,
			},
			Storage: api.ClusterResource{
				Total: storage,
				Used:  0,
			},
			State: state,
		},
		nil
}

func (s SQLClustersService) getClusterCount() (total int, err error) {
	// retrieve total number of clusters.
	err = s.db.QueryRow("select count(*) from clusters").Scan(&total)
	return
}

// GetStatus returns a cluster status by id
func (s SQLClustersService) GetStatus(id string) (result api.ClusterStatus, err error) {
	var name string
	err = s.db.QueryRow(`
	SELECT
		id,
		name
	FROM clusters
	WHERE
		id = $1
	`, id,
	).Scan(
		&id,
		&name,
	)

	if err != nil {
		return api.ClusterStatus{}, err
	}

	var state api.ClusterState

	// Check state with provisioner
	state, err = s.provisioner.GetState(id)
	if err != nil {
		return api.ClusterStatus{}, fmt.Errorf("An error occurred while trying get status for cluster %s: %s",
			id, err)
	}

	return api.ClusterStatus{
		ID:    id,
		State: state,
	}, nil
}
