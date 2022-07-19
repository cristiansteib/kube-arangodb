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

package crd

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func init() {
	registerCRDWithPanic("arangobackuppolicies.backup.arangodb.com", crd{
		version: "1.0.1",
		spec: apiextensions.CustomResourceDefinitionSpec{
			Group: "backup.arangodb.com",
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "arangobackuppolicies",
				Singular: "arangobackuppolicy",
				Kind:     "ArangoBackupPolicy",
				ListKind: "ArangoBackupPolicyList",
				ShortNames: []string{
					"arangobackuppolicy",
					"arangobp",
				},
			},
			Scope: apiextensions.NamespaceScoped,
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name: "v1",
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: util.NewBool(true),
						},
					},
					Served:  true,
					Storage: true,
					AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
						{
							JSONPath:    ".spec.schedule",
							Description: "Schedule",
							Name:        "Schedule",
							Type:        "string",
						},
						{
							JSONPath:    ".status.scheduled",
							Description: "Scheduled",
							Name:        "Scheduled",
							Type:        "string",
						},
						{
							JSONPath:    ".status.message",
							Priority:    1,
							Description: "Message of the ArangoBackupPolicy object",
							Name:        "Message",
							Type:        "string",
						},
					},
					Subresources: &apiextensions.CustomResourceSubresources{
						Status: &apiextensions.CustomResourceSubresourceStatus{},
					},
				},
				{
					Name: "v1alpha",
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: util.NewBool(true),
						},
					},
					Served:  true,
					Storage: false,
					AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
						{
							JSONPath:    ".spec.schedule",
							Description: "Schedule",
							Name:        "Schedule",
							Type:        "string",
						},
						{
							JSONPath:    ".status.scheduled",
							Description: "Scheduled",
							Name:        "Scheduled",
							Type:        "string",
						},
						{
							JSONPath:    ".status.message",
							Priority:    1,
							Description: "Message of the ArangoBackupPolicy object",
							Name:        "Message",
							Type:        "string",
						},
					},
					Subresources: &apiextensions.CustomResourceSubresources{
						Status: &apiextensions.CustomResourceSubresourceStatus{},
					},
				},
			},
		},
	})
}