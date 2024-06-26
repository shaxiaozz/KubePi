package sso

import v1 "github.com/KubeOperator/kubepi/internal/model/v1"

type Sso struct {
	v1.BaseModel     `storm:"inline"`
	v1.Metadata      `storm:"inline"`
	Enable           bool   `json:"enable"`
	Protocol         string `json:"protocol"`
	InterfaceAddress string `json:"interfaceAddress"`
	ClientId         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
}
