// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package metrics

import (
	"math"
	"sort"
	"sync"
	"time"
)

type Summary struct {
	Produced int64
	Consumed int64
	Errors   int64

	ProduceLatencies []time.Duration
	ConsumeLatencies []time.Duration
	ConsumePollLatencies []time.Duration

	mu sync.Mutex
}

func (s *Summary) AddProduce(lat time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Produced++
	if lat > 0 {
		s.ProduceLatencies = append(s.ProduceLatencies, lat)
	}
}

func (s *Summary) AddConsume(lat time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Consumed++
	if lat > 0 {
		s.ConsumeLatencies = append(s.ConsumeLatencies, lat)
	}
}

func (s *Summary) AddConsumePoll(lat time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if lat > 0 {
		s.ConsumePollLatencies = append(s.ConsumePollLatencies, lat)
	}
}

func (s *Summary) AddError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Errors++
}

type Percentiles struct {
	P50 float64
	P95 float64
	P99 float64
}

func LatencyPercentiles(durations []time.Duration) Percentiles {
	if len(durations) == 0 {
		return Percentiles{}
	}
	values := make([]float64, 0, len(durations))
	for _, d := range durations {
		values = append(values, float64(d.Milliseconds()))
	}
	sort.Float64s(values)
	return Percentiles{
		P50: percentile(values, 0.50),
		P95: percentile(values, 0.95),
		P99: percentile(values, 0.99),
	}
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 1 {
		return values[len(values)-1]
	}
	pos := p * float64(len(values)-1)
	lower := int(math.Floor(pos))
	upper := int(math.Ceil(pos))
	if lower == upper {
		return values[lower]
	}
	frac := pos - float64(lower)
	return values[lower] + (values[upper]-values[lower])*frac
}
