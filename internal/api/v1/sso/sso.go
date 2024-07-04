package sso

import (
	"github.com/KubeOperator/kubepi/internal/api/v1/user"
	v1Sso "github.com/KubeOperator/kubepi/internal/model/v1/sso"
	v1User "github.com/KubeOperator/kubepi/internal/model/v1/user"
	"github.com/KubeOperator/kubepi/internal/server"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/KubeOperator/kubepi/internal/service/v1/sso"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"strings"
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

func (h *Handler) LoginSso() iris.Handler {
	return func(ctx *context.Context) {
		ssos, err := h.ssoService.List(common.DBOptions{})
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}

		// 根据协议设置重定向URL
		r := ctx.Request()
		redirectURL := ""
		if strings.HasPrefix(strings.ToLower(r.Proto), "https") {
			redirectURL = "https://" + r.Host + "/callback"
		} else if strings.HasPrefix(strings.ToLower(r.Proto), "http") {
			redirectURL = "http://" + r.Host + "/callback"
		}

		// 目前只支持OpenID
		switch ssos[0].Protocol {
		case "openid":
			openid := &v1Sso.OpenID{
				ClientId:     ssos[0].ClientId,
				ClientSecret: ssos[0].ClientSecret,
				RedirectURL:  redirectURL,
				IssuerURL:    ssos[0].InterfaceAddress,
				IsConfig:     true,
			}
			oauth2Config, err := h.ssoService.OpenID(openid)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				ctx.Values().Set("message", err.Error())
				return
			}
			ctx.Redirect(oauth2Config.AuthCodeURL("state"), iris.StatusFound)
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", "目前只支持OpenID")
			return
		}
	}
}

func (h *Handler) CallbackSso() iris.Handler {
	return func(ctx *context.Context) {
		var req user.User
		language := ctx.GetHeader("Accept-Language")
		if strings.Contains(language, "zh-CN") {
			req.Language = "zh-CN"
		} else {
			req.Language = "en-US"
		}

		//tx
		tx, err := server.DB().Begin(true)
		_ = tx.Rollback()
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.Values().Set("message", err.Error())
			return
		}
		req.Type = v1User.LOCAL
		ctx.Values().Set("data", &req.User)
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
	sp.Get("/login", handler.LoginSso())
	sp.Get("/callback", handler.CallbackSso())
	sp.Post("/test/connect", handler.TestConnect())
	sp.Get("/status", handler.StatusSso())
}
