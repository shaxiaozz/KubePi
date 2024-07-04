package sso

import (
	"context"
	"errors"
	"fmt"
	v1Sso "github.com/KubeOperator/kubepi/internal/model/v1/sso"
	"github.com/KubeOperator/kubepi/internal/service/v1/common"
	"github.com/KubeOperator/kubepi/internal/service/v1/user"
	ssoClient "github.com/KubeOperator/kubepi/pkg/util/sso"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"time"
)

type Service interface {
	common.DBService
	TestConnect(sso *v1Sso.Sso) error
	List(options common.DBOptions) ([]v1Sso.Sso, error)
	Create(sso *v1Sso.Sso, options common.DBOptions) error
	Update(id string, sso *v1Sso.Sso, options common.DBOptions) error
	Status(options common.DBOptions) bool
	OpenID(openid *v1Sso.OpenID) (*oauth2.Config, error)
}

func NewService() Service {
	return &service{}
}

type service struct {
	common.DefaultDBService
	userService user.Service
}

func (s *service) TestConnect(sso *v1Sso.Sso) error {
	// 测试连接，理论上不应强制用户开启SSO认证
	//if !sso.Enable {
	//	return errors.New("请先启用SSO")
	//}

	sc := ssoClient.NewSsoClient(sso.Protocol, sso.InterfaceAddress, sso.ClientId, sso.ClientSecret, sso.Enable)
	if err := sc.TestConnect(sso.InterfaceAddress); err != nil {
		return err
	}

	return nil
}

func (s *service) Create(sso *v1Sso.Sso, options common.DBOptions) error {
	sc := ssoClient.NewSsoClient(sso.Protocol, sso.InterfaceAddress, sso.ClientId, sso.ClientSecret, sso.Enable)
	// 当用户进行SSO配置时，应该为用户检测目标是否可连接
	if err := sc.TestConnect(sso.InterfaceAddress); err != nil {
		return err
	}

	db := s.GetDB(options)
	sso.UUID = uuid.New().String()
	sso.CreateAt = time.Now()
	sso.UpdateAt = time.Now()
	return db.Save(sso)
}

func (s *service) Update(id string, sso *v1Sso.Sso, options common.DBOptions) error {
	sc := ssoClient.NewSsoClient(sso.Protocol, sso.InterfaceAddress, sso.ClientId, sso.ClientSecret, sso.Enable)
	// 当用户进行SSO配置时，应该为用户检测目标是否可连接
	if err := sc.TestConnect(sso.InterfaceAddress); err != nil {
		return err
	}

	old, err := s.GetById(id, options)
	if err != nil {
		return err
	}
	sso.UUID = old.UUID
	sso.CreateAt = old.CreateAt
	sso.UpdateAt = time.Now()
	db := s.GetDB(options)
	if sso.Enable != old.Enable {
		err = db.UpdateField(sso, "Enable", sso.Enable)
		if err != nil {
			return err
		}
	}
	return db.Update(sso)
}

func (s *service) List(options common.DBOptions) ([]v1Sso.Sso, error) {
	db := s.GetDB(options)
	sso := make([]v1Sso.Sso, 0)
	if err := db.All(&sso); err != nil {
		return nil, err
	}
	return sso, nil
}

func (s *service) Status(options common.DBOptions) bool {
	db := s.GetDB(options)
	sso := make([]v1Sso.Sso, 0)
	if err := db.All(&sso); err != nil {
		return false
	}

	if len(sso) == 0 {
		return false
	}

	return sso[0].Enable
}

func (s *service) GetById(id string, options common.DBOptions) (*v1Sso.Sso, error) {
	db := s.GetDB(options)
	var sso v1Sso.Sso
	query := db.Select(q.Eq("UUID", id))
	if err := query.First(&sso); err != nil {
		return nil, err
	}
	return &sso, nil
}

func (s *service) OpenID(openid *v1Sso.OpenID) (*oauth2.Config, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, openid.IssuerURL)
	if err != nil {
		return nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     openid.ClientId,
		ClientSecret: openid.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  openid.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	if openid.IsConfig {
		return oauth2Config, nil
	}

	// 用 code 换取 token
	token, err := oauth2Config.Exchange(ctx, openid.Code)
	if err != nil {
		return nil, errors.New("交换Token失败: " + err.Error())
	}

	// 使用 token 获取用户信息
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, errors.New("获取用户信息失败: " + err.Error())
	}

	localUser, err := s.userService.GetByNameOrEmail(userInfo.Email, openid.Options)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			// 创建本地账号，密码默认设置为`@=7kvi-$l*Pj+,s`，用户角色默认为ReadOnly，默认不开启MFA
			//userProfile := &v1User.User{
			//	NickName:     "",
			//	Email:        "",
			//	Language:     "zh-CN",
			//	IsAdmin:      false,
			//	Authenticate: v1User.Authenticate{},
			//	Type:         v1User.LOCAL,
			//	Mfa:          v1User.Mfa{Enable: false},
			//}
			//s.userService.Create()
			fmt.Println(localUser.Name + "用户不存在")
			return nil, nil
		}
		return nil, err
	}

	return nil, err
}
