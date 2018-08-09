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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/container-mgmt/dedicated-portal/cmd/clusters-service/service"
	"github.com/container-mgmt/dedicated-portal/pkg/api"
)

func (s Server) listClusters(w http.ResponseWriter, r *http.Request) {
	page, err := getQueryParamInt("page", 0, r)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	size, err := getQueryParamInt("size", 100, r)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	results, err := s.clusterService.List(service.ListArguments{Page: page, Size: size})
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	writeJSONResponse(w, http.StatusOK, results)
}

func (s Server) createCluster(w http.ResponseWriter, r *http.Request) {
	provision, err := getQueryParamBool("provision", true, r)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	var spec api.Cluster
	err = json.Unmarshal(bytes, &spec)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	if spec.ID != "" {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "id must be empty"})
		return
	}
	result, err := s.clusterService.Create(spec, provision)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	writeJSONResponse(w, http.StatusCreated, result)
}

func (s Server) getCluster(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "no id provided"})
		return
	}
	cluster, err := s.clusterService.Get(id)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	writeJSONResponse(w, http.StatusOK, cluster)
}

func (s Server) getClusterStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": "no id provided"})
		return
	}

	status, err := s.clusterService.GetStatus(id)
	if err != nil {
		writeJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("%v", err)})
		return
	}
	writeJSONResponse(w, http.StatusOK, status)

}

func getQueryParamInt(param string, defaultValue int, r *http.Request) (value int, err error) {
	valueString, ok := r.URL.Query()[param]

	if !ok || len(valueString) < 1 {
		return defaultValue, nil
	}
	var result int64
	// This needs to be ParseInt and not Atoi because the interface asks for int64
	result, err = strconv.ParseInt(valueString[0], 10, 32)
	return int(result), err
}

func getQueryParamBool(param string, defaultValue bool, r *http.Request) (value bool, err error) {
	valueString, ok := r.URL.Query()[param]

	if !ok || len(valueString) < 1 {
		return defaultValue, nil
	}
	var result bool
	result, err = strconv.ParseBool(valueString[0])
	return result, err
}

func writeJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.MarshalIndent(payload, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
