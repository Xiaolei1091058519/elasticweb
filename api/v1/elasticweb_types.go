/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"strconv"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticWebSpec defines the desired state of ElasticWeb.
type ElasticWebSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ElasticWeb. Edit elasticweb_types.go to remove/update
	// Foo string `json:"foo,omitempty"`

	// 业务服务对应的镜像，包括tag
	SinglePodQPS *int32                 `json:"singlePodQPS"`
	TotalQPS     *int32                 `json:"totalQPS"`
	Deploy       []ElasticWebSpecDeploy `json:"deploy"`
	Service      ElasticWebSpecSvc      `json:"service"`
}

type ElasticWebSpecDeploy struct {
	Name  string                      `json:"name"`
	Image string                      `json:"image"`
	Ports []ElasticWebSpecDeployPorts `json:"ports"`
}

type ElasticWebSpecDeployPorts struct {
	Name string `json:"name"`
	Port *int32 `json:"port"`
}

// type ElasticWebSpecDeployResources struct {
// 	Requests []ElasticWebSpecDeployResourcesRequests
// 	Limits
// }

type ElasticWebSpecSvc struct {
	Type  string                   `json:"type"`
	Ports []ElasticWebSpecSvcPorts `json:"ports"`
}

type ElasticWebSpecSvcPorts struct {
	Name       string `json:"name"`
	Port       *int32 `json:"port"`
	TargetPort *int32 `json:"targetport"`
}

// ElasticWebStatus defines the observed state of ElasticWeb.
type ElasticWebStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	RealQPS *int32 `json:"realQPS"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ElasticWeb is the Schema for the elasticwebs API.
type ElasticWeb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticWebSpec   `json:"spec,omitempty"`
	Status ElasticWebStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ElasticWebList contains a list of ElasticWeb.
type ElasticWebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticWeb `json:"items"`
}

func (in *ElasticWeb) String() string {
	var realQPS string
	if nil == in.Status.RealQPS {
		realQPS = "nil"
	} else {
		realQPS = strconv.Itoa(int(*(in.Status.RealQPS)))
	}

	var Str []string

	for _, v0 := range in.Spec.Deploy {
		for _, v1 := range v0.Ports {
			Str = append(Str, fmt.Sprintf("Image [%s],Port [%d]", v0.Image, *(&v1.Port)))
		}
	}
	Str = append(Str, realQPS)
	return strings.Join(Str, " ")
}

func init() {
	SchemeBuilder.Register(&ElasticWeb{}, &ElasticWebList{})
}
