package sso

import (
	v1Sso "github.com/KubeOperator/kubepi/internal/model/v1/sso"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/KubeOperator/kubepi/internal/service/v1/sso"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

type Handler struct {
	ssoService sso.Service
}

func NewHandler() *Handler {
	return &Handler{
		ssoService: sso.NewService(),
	}
}

func (h *Handler) AddSso() iris.Handler {
	return func(ctx *context.Context) {
		var req v1Sso.Sso
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
		}
		err := h.ssoService.Create(&req, common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", &req)
	}
}

func (h *Handler) ListSso() iris.Handler {
	return func(ctx *context.Context) {
		ssos, err := h.ssoService.List(common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", ssos)
	}
}

func (h *Handler) UpdateSso() iris.Handler {
	return func(ctx *context.Context) {
		var req v1Sso.Sso
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
		}
		err := h.ssoService.Update(req.UUID, &req, common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", &req)
	}
}

func (h *Handler) TestConnect() iris.Handler {
	return func(ctx *context.Context) {
		var req v1Sso.Sso
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.Values().Set("message", err.Error())
		}
		if err := h.ssoService.TestConnect(&req); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		ctx.Values().Set("data", "SSO测试连接成功")
	}
}

func (h *Handler) StatusSso() iris.Handler {
	return func(ctx *context.Context) {
		if ssoSwitch := h.ssoService.Status(common.DBOptions{}); !ssoSwitch {
			ctx.Values().Set("data", false)
			return
		}
		ctx.Values().Set("data", true)
	}
}

func Install(parent iris.Party) {
	handler := NewHandler()
	sp := parent.Party("/sso")
	sp.Get("/", handler.ListSso())
	sp.Post("/", handler.AddSso())
	sp.Put("/", handler.UpdateSso())
	sp.Post("/test/connect", handler.TestConnect())
	sp.Get("/status", handler.StatusSso())
}
