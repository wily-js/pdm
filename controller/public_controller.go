package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"pdm/appconf/dir"
	"strconv"
	"strings"
)

func NewPublicController(r gin.IRouter) *PublicController {
	res := &PublicController{}
	// 获取头像
	r.GET("/avatar", res.avatar)
	return res
}

type PublicController struct {
}

/**
@api {GET} /api/avatar 获取头像
@apiDescription 获取用户头像
@apiName PublicAvatar
@apiGroup Public

@apiPermission 匿名

@apiParam {String} type 类型
<ul>
	    <li>admin</li>
	    <li>audit</li>
 		<li>user</li>
</ul>
@apiParam {Integer} id ID

@apiParamExample 请求示例
GET /api/avatar?type=user&id=3

@apiSuccess {[]byte} data 图片二进制数据。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK


@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// avatar 获取头像
func (c *PublicController) avatar(ctx *gin.Context) {
	avatarType := ctx.Query("type")
	id, err := strconv.Atoi(ctx.Query("id"))

	if avatarType == "" || err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	avatarName := fmt.Sprintf("%s-%d", avatarType, id)

	avatarPath := filepath.Join(dir.AvatarDir, avatarName)
	// 防止用户通过 ../../ 的方式读取到操作系统内的重要文件
	if !strings.HasPrefix(avatarPath, dir.AvatarDir) {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	// 打开头像文件
	avatar, err := os.Open(avatarPath)
	defer avatar.Close()
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	_, err = io.Copy(ctx.Writer, avatar)
	if err != nil {
		ErrSys(ctx, err)
	}
}
