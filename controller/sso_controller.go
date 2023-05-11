package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"pdm/controller/dto"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
	"time"
)

// NewSsoController 创建单点登录控制器
func NewSsoController(router gin.IRouter, baseUrl string) *SsoController {
	res := &SsoController{}
	router.GET("/redirect", res.redirect)
	res.ssoBaseUrl = baseUrl
	return res
}

// SsoController 单点登录控制器
type SsoController struct {
	ssoBaseUrl string
}

/**
@api {get} /api/redirect 重定向页面

@apiName OauthRedirect
@apiGroup Oauth

@apiParam {String} code  授权码

@apiParamExample {get} 请求示例
GET /oauth/redirect?code=b9502e98e4e3adf1dd400b39c60e272f7c9458db146d22229f3796c0f503a81c

@apiExample html调用示例
<a id="login"></a>
<script>
	var host = "https://nantemen.hzauth.com"
	var login = document.getElementById("login")
	var redirect_uri = "http://example.com/"
	var client_id = "CLIENT_ID"
	login.href = host+"/oauth/authorize?client_id=" + client_id + "&redirect_uri=" + redirect_uri
</script>
*/

// redirect 单点登录回调接口
func (c *SsoController) redirect(ctx *gin.Context) {
	const clientId = "14b40ad8-30bc-4ee1-b381-8a43cfcf185c"
	const clientSecret = "2cd12ccbf873e5d491b47c3e6d75af1ba818f25a8b27580607ce20a33fbd8d96"
	const grantType = "authorization_code"
	code := ctx.Query("code")
	log.Println(code)
	// 获取access_token
	tokenRes, _ := http.NewRequest("GET", c.ssoBaseUrl+"/oauth/token?client_id="+
		clientId+"&client_secret="+clientSecret+"&grant_type="+grantType+"&code="+code, nil)
	tokenResp, _ := http.DefaultClient.Do(tokenRes)
	body, _ := io.ReadAll(tokenResp.Body)
	defer tokenResp.Body.Close()
	tokenInfo := dto.AccessTokenDto{}
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		ErrSys(ctx, err)
		return
	}

	// 获取用户信息
	accessToken := fmt.Sprintf("Bearer %s", tokenInfo.AccessToken)
	infoRes, _ := http.NewRequest("GET", c.ssoBaseUrl+"/info", nil)
	infoRes.Header.Add("accept", "application/json")
	infoRes.Header.Add("Authorization", accessToken)
	infoResp, _ := http.DefaultClient.Do(infoRes)
	body, _ = io.ReadAll(infoResp.Body)
	defer infoResp.Body.Close()
	info := dto.InfoDto{}
	if err := json.Unmarshal(body, &info); err != nil {
		ErrSys(ctx, err)
		return
	}

	openid := info.Data.Openid
	user := entity.User{}
	err := repo.DB.First(&user, "openid = ? AND is_delete = 0", openid).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 生成用户token进入主页
	claims := jwt.Claims{Type: "user", Sub: user.ID, Exp: time.Now().Add(8 * time.Hour).UnixMilli()}
	token := tokenManager.GenToken(&claims)
	// 设置头部 Cookies 有效时间为8小时
	ctx.SetCookie("token", token, 8*3600, "", "", false, true)

	ctx.Redirect(http.StatusMovedPermanently, "/ui/#/index/projectList")
}
