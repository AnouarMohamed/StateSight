package kubeid

import "fmt"

// ID is a compact identifier for a Kubernetes resource.
type ID struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	Name      string
}

// String renders a stable identity key for logs and storage.
func (id ID) String() string {
	return fmt.Sprintf("%s/%s/%s:%s/%s", id.Group, id.Version, id.Kind, id.Namespace, id.Name)
}
