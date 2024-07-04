package sso

import (
	v1 "github.com/KubeOperator/kubepi/internal/model/v1"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
)

type Sso struct {
	v1.BaseModel     `storm:"inline"`
	v1.Metadata      `storm:"inline"`
	Enable           bool   `json:"enable"`
	Protocol         string `json:"protocol"`
	InterfaceAddress string `json:"interfaceAddress"`
	ClientId         string `json:"clientId"`
	ClientSecret     string `json:"clientSecret"`
}

type OpenID struct {
	ClientId     string
	ClientSecret string
	RedirectURL  string
	IssuerURL    string
	IsConfig     bool
	Code         string
	Options      common.DBOptions
}
