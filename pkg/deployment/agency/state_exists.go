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

package agency

import "github.com/arangodb/kube-arangodb/pkg/util"

type StateExists []byte

func (d StateExists) Hash() string {
	if d == nil {
		return ""
	}

	return util.SHA256(d)
}

func (d StateExists) Exists() bool {
	return d != nil
}

func (d *StateExists) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		*d = nil
		return nil
	}

	z := make([]byte, len(bytes))

	copy(z, bytes)

	*d = z
	return nil
}
