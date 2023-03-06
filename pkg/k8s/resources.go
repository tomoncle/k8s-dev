package k8s

type Resource string

func (r *Resource) toString() string {
	return string(*r)
}

const (
	DefaultNamespace = "default"
	POD              = Resource("pods")
	SVC              = Resource("services")
	DEPLOY           = Resource("deployments")
)
