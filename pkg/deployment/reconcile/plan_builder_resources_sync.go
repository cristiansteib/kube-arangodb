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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/sync"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
)

func createArangoMembersResourceSyncPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, cid := range context.ACS().RemoteClusters() {
		c, ok := context.ACS().Cluster(cid)
		if !ok {
			continue
		}

		if !c.Ready() {
			continue
		}

		for _, member := range status.Members.AsList() {
			name := member.ArangoMember(apiObject.GetName())
			_, _, a := sync.GetArangoMemberActionType(context.ACS().CurrentClusterCache(), c.Cache(), name)
			switch a {
			case sync.ActionCreate, sync.ActionDelete:
				plan = append(plan, newResourceSyncClusterAction("", cid, name, inspector.ArangoMemberKind))
			}
		}
	}

	return plan
}

func createSecretsResourceSyncPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	var plan api.Plan
	var secrets []string

	if spec.IsAuthenticated() {
		secrets = append(secrets, pod.JWTSecretFolder(apiObject.GetName()))
	}

	if spec.TLS.IsSecure() {
		secrets = append(secrets, resources.GetCASecretName(apiObject))

		for _, member := range status.Members.AsList() {
			secrets = append(secrets, k8sutil.AppendTLSKeyfileSecretPostfix(member.ArangoMember(apiObject.GetName())))
		}
	}

	for _, cid := range context.ACS().RemoteClusters() {
		c, ok := context.ACS().Cluster(cid)
		if !ok {
			continue
		}

		if !c.Ready() {
			continue
		}

		for _, name := range secrets {
			_, _, a := sync.GetSecretActionType(context.ACS().CurrentClusterCache(), c.Cache(), name)
			switch a {
			case sync.ActionCreate, sync.ActionUpdate, sync.ActionDelete:
				plan = append(plan, newResourceSyncClusterAction("", cid, name, inspector.SecretKind))
			}
		}
	}

	return plan
}

func newResourceSyncClusterAction(from, to types.UID, name, kind string) api.Action {
	return actions.NewClusterAction(api.ActionTypeResourceSync).
		AddParam(ActionResourcesParamSource, string(from)).
		AddParam(ActionResourcesParamDestination, string(to)).
		AddParam(ActionResourcesParamName, name).
		AddParam(ActionResourcesParamKind, kind)
}
