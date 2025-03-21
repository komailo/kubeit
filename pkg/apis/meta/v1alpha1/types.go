package v1alpha1

type ObjectMeta struct {
	Name string `json:"name" validate:"required"`
}
