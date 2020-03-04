package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type ManagedSecretSpec struct {
	Project    string `json:"project"`
	Secret     string `json:"secret"`
	Generation int64  `json:"generation"`
}

type ManagedSecretStatus struct {
	ManagedSecretSpec
	Error string
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedSecretSpec   `json:"spec"`
	Status ManagedSecretStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ManagedSecret `json:"items"`
}

func (ms ManagedSecret) String() string {
	return fmt.Sprintf("%s/%s/%d", ms.Spec.Project, ms.Spec.Secret, ms.Spec.Generation)
}
