//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package v1

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultMaintenanceGracePeriod = 30 * time.Minute
)

type Timeouts struct {

	// MaintenanceGracePeriod action timeout
	MaintenanceGracePeriod *Timeout `json:"maintenanceGracePeriod,omitempty"`

	// Actions
	Actions ActionTimeouts `json:"actions,omitempty"`

	// deprecated
	AddMember *Timeout `json:"-"`

	// deprecated
	RuntimeContainerImageUpdate *Timeout `json:"-"`
}

func (t *Timeouts) GetMaintenanceGracePeriod() time.Duration {
	if t == nil {
		return DefaultMaintenanceGracePeriod
	}

	return t.MaintenanceGracePeriod.Get(DefaultMaintenanceGracePeriod)
}

func (t *Timeouts) Get() Timeouts {
	if t == nil {
		return Timeouts{}
	}

	return *t
}

type ActionTimeouts map[ActionType]Timeout

func NewTimeout(timeout time.Duration) Timeout {
	return Timeout(meta.Duration{Duration: timeout})
}

type Timeout meta.Duration

func (t *Timeout) Get(d time.Duration) time.Duration {
	if t == nil {
		return d
	}

	return t.Duration
}
