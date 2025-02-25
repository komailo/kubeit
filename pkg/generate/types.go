package generate

import "github.com/komailo/kubeit/common"

type GenerateOptions struct {
	OutputDir string
	WorkDir   string
}

type ManifestSource struct {
	Type   string
	Uri    string
	RawUri string
}

type CommonK8sLabels struct {
}

type CommonK8sAnnotations struct {
	AppName     string
	AppType     string
	GeneratedBy string
}

func (c *CommonK8sAnnotations) GenerateAnnotations() map[string]string {
	return map[string]string{
		common.KubeitDomain + "/app-name":     c.AppName,
		common.KubeitDomain + "/app-type":     c.AppType,
		common.KubeitDomain + "/generated-by": c.GeneratedBy,
	}
}

func (c *CommonK8sLabels) GenerateLabels() map[string]string {
	// TODO: Implement label generation logic here
	return map[string]string{}
}
