package common

import (
	"strings"
)

const (
	AppName       = "Reflow"
	ServiceDomain = "reflow.thescore.is"
)

var MainCLIName = strings.ToLower(AppName)

var APIVersionV1Alpha1 = ServiceDomain + "/" + "v1alpha1"
