#!/bin/bash
# Copyright 2018 Istio Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
################################################################################

set -o errexit
set -o nounset
set -o pipefail
set -x

# shellcheck disable=SC1091
source "/workspace/gcb_env.sh"

# This script downloads creates a static file on GCS which has the download link of lnux tar gz

DAILY_HTTPS_PATH="https://storage.googleapis.com/${CB_GCS_STAGING_BUCKET}/daily-build/${CB_VERSION}/istio-${CB_VERSION}-linux.tar.gz"

TEMP_FILE=$(mktemp)
echo -n "${DAILY_HTTPS_PATH}" > "${TEMP_FILE}"
cat "${TEMP_FILE}"

# this file contains the linux download URL of the latest successful daily build for a particular branch
gsutil -q cp "${TEMP_FILE}" "gs://${CB_GCS_STAGING_BUCKET}/daily-build/${CB_BRANCH}-latest.txt"
