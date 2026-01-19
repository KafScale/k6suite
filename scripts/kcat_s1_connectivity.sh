#!/bin/sh

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

set -e

BROKER="${BROKER:-127.0.0.1:39092}"

echo "S1 connectivity check (kcat)"
echo "Broker: ${BROKER}"

if ! command -v kcat >/dev/null 2>&1; then
  echo "kcat not found. Install with: brew install kcat (macOS) or apt-get install kcat (Linux)."
  exit 1
fi

echo "Fetching broker metadata..."
kcat -b "${BROKER}" -X broker.address.family=v4 -L >/dev/null
echo "OK: metadata fetched"
