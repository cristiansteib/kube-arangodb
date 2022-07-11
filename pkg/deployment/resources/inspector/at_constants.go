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

package inspector

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deploymentv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

// ArangoTask
const (
	ArangoTaskGroup     = deployment.ArangoDeploymentGroupName
	ArangoTaskResource  = deployment.ArangoTaskResourcePlural
	ArangoTaskKind      = deployment.ArangoTaskResourceKind
	ArangoTaskVersionV1 = deploymentv1.ArangoDeploymentVersion
)

func ArangoTaskGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoTaskGroup,
		Kind:  ArangoTaskKind,
	}
}

func ArangoTaskGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoTaskGroup,
		Kind:    ArangoTaskKind,
		Version: ArangoTaskVersionV1,
	}
}

func ArangoTaskGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoTaskGroup,
		Resource: ArangoTaskResource,
	}
}

func ArangoTaskGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoTaskGroup,
		Resource: ArangoTaskResource,
		Version:  ArangoTaskVersionV1,
	}
}

func (p *arangoTasksInspectorV1) GroupVersionKind() schema.GroupVersionKind {
	return ArangoTaskGKv1()
}

func (p *arangoTasksInspectorV1) GroupVersionResource() schema.GroupVersionResource {
	return ArangoTaskGRv1()
}

func (p *arangoTasksInspector) GroupKind() schema.GroupKind {
	return ArangoTaskGK()
}

func (p *arangoTasksInspector) GroupResource() schema.GroupResource {
	return ArangoTaskGR()
}
