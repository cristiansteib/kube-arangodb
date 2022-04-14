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

package shared

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"unicode"
)

var (
	arangodPrefixes = []string{"CRDN-", "PRMR-", "AGNT-", "SNGL-"}
)

// stripArangodPrefix removes well know arangod ID prefixes from the given id.
func stripArangodPrefix(id string) string {
	for _, prefix := range arangodPrefixes {
		if strings.HasPrefix(id, prefix) {
			return id[len(prefix):]
		}
	}
	return id
}

// FixupResourceName ensures that the given name
// complies with kubernetes name requirements.
// If the name is to long or contains invalid characters,
// if will be adjusted and a hash with be added.
func FixupResourceName(name string) string {
	maxLen := 63

	sb := strings.Builder{}
	needHash := len(name) > maxLen
	for _, ch := range name {
		if unicode.IsDigit(ch) || unicode.IsLower(ch) || ch == '-' {
			sb.WriteRune(ch)
		} else if unicode.IsUpper(ch) {
			sb.WriteRune(unicode.ToLower(ch))
			needHash = true
		} else {
			needHash = true
		}
	}
	result := sb.String()
	if needHash {
		hash := sha1.Sum([]byte(name))
		h := fmt.Sprintf("-%0x", hash[:3])
		if len(result)+len(h) > maxLen {
			result = result[:maxLen-(len(h))]
		}
		result = result + h
	}
	return result
}

// CreatePodHostName returns the hostname of the pod for a member with
// a given id in a deployment with a given name.
func CreatePodHostName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + stripArangodPrefix(id)
}

// CreatePersistentVolumeClaimName returns the name of the persistent volume claim for a member with
// a given id in a deployment with a given name.
func CreatePersistentVolumeClaimName(deploymentName, role, id string) string {
	return deploymentName + "-" + role + "-" + stripArangodPrefix(id)
}