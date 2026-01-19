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

SHELL := /bin/sh

K6_BIN ?= ./k6
K6_PROFILE ?= local-service
K6_TARGET ?= kafscale
.DEFAULT_GOAL := help

.PHONY: help k6-select k6-diagnose k6-smoke-single k6-smoke-concurrent k6-smoke-shared k6-smoke-topic-autocreate k6-smoke-consumer-group k6-smoke-multi-producer-single-consumer k6-smoke-metrics k6-suite

help:
	@echo "Targets:"
	@echo "  k6-diagnose                      - Run diagnose.js"
	@echo "  k6-select                        - Select k6 scenarios/profiles interactively"
	@echo "  k6-smoke-single                  - Run smoke_single.js"
	@echo "  k6-smoke-concurrent              - Run smoke_concurrent.js"
	@echo "  k6-smoke-shared                  - Run smoke_shared.js"
	@echo "  k6-smoke-topic-autocreate        - Run smoke_topic_autocreate.js"
	@echo "  k6-smoke-consumer-group          - Run smoke_consumer_group.js"
	@echo "  k6-smoke-multi-producer-single-consumer - Run smoke_multi_producer_single_consumer.js"
	@echo "  k6-smoke-metrics                 - Run smoke_metrics.js"
	@echo "  k6-suite                         - Run all k6 smoke tests"

k6-diagnose:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/diagnose.js

k6-select:
	./kaf6-runner k6-select tests/k6

k6-smoke-single:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_single.js

k6-smoke-concurrent:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_concurrent.js

k6-smoke-shared:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_shared.js

k6-smoke-topic-autocreate:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_topic_autocreate.js

k6-smoke-consumer-group:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_consumer_group.js

k6-smoke-multi-producer-single-consumer:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_multi_producer_single_consumer.js

k6-smoke-metrics:
	K6_PROFILE=$(K6_PROFILE) K6_TARGET=$(K6_TARGET) $(K6_BIN) run tests/k6/smoke_metrics.js

k6-suite:
	$(MAKE) k6-diagnose
	$(MAKE) k6-smoke-metrics
	$(MAKE) k6-smoke-single
	$(MAKE) k6-smoke-concurrent
	$(MAKE) k6-smoke-shared
	$(MAKE) k6-smoke-topic-autocreate
	$(MAKE) k6-smoke-consumer-group
	$(MAKE) k6-smoke-multi-producer-single-consumer
