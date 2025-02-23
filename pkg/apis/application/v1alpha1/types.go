package v1alpha1

const GroupVersion = "kubeit.komailo.github.io/v1alpha1"
const Kind = "Application"

type Application struct {
	Metadata Metadata    `json:"metadata" yaml:"metadata"`
	Spec     interface{} `json:"spec" yaml:"spec"`
}

type Metadata struct {
	Name string `json:"name" yaml:"name" validate:"required"`
}
