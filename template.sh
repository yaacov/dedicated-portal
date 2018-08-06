#!/bin/bash -ex

#
# Copyright (c) 2018 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script uses the OpenShift template to create a deployment of
# the application.

# Create the namespace:
oc new-project unified-hybrid-cloud || oc project unified-hybrid-cloud || true

# Use the template to create the objects:
oc process \
  --filename="template.yml" \
  --param=NAMESPACE="${TEMPLATE_NAMESPACE:-unified-hybrid-cloud}" \
  --param=VERSION="${TEMPLATE_VERSION:-latest}" \
  --param=DOMAIN="${TEMPLATE_DOMAIN:-example.com}" \
  --param=PASSWORD="${TEMPLATE_PASSWORD:-redhat123}" \
  --param=DEMO_MODE="${TEMPLATE_DEMO_MODE:false}" \
  --param=SSH_PUBLIC_KEY="${TEMPLATE_SSH_PUBLIC_KEY}" \
  --param=SSH_PRIVATE_KEY="${TEMPLATE_SSH_PRIVATE_KEY}" \
  --param=AWS_ACCESS_KEY_ID="${TEMPLATE_AWS_ACCESS_KEY_ID}" \
  --param=AWS_SECRET_ACCESS_KEY="${TEMPLATE_AWS_SECRET_ACCESS_KEY}" \
  --param=DEFAULT_CLUSTER_TLS_CERTIFICATE="${TEMPLATE_DEFAULT_CLUSTER_TLS_CERTIFICATE}" \
  --param=DEFAULT_CLUSTER_TLS_KEY="${TEMPLATE_DEFAULT_CLUSTER_TLS_KEY}" \
| \
oc apply \
  --filename=-
