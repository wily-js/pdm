package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/dto"
	"pdm/controller/middle"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"pdm/reuint"
	"pdm/reuint/jwt"
	"strconv"
	"strings"
)

// UserController 用户控制器
type UserController struct {
}

// NewUserController 创建用户控制器
func NewUserController(router gin.IRouter) *UserController {
	res := &UserController{}
	r := router.Group("/user")
	// 创建用户
	r.POST("/create", Admin, res.create)
	// 查找用户
	r.GET("/search", Admin, res.search)
	// 修改口令
	r.POST("/modifyPwd", User, res.modifyPwd)
	// 重置口令
	r.POST("/resetPwd", Admin, res.resetPwd)
	// 更换头像
	r.POST("/updateAvatar", User, res.updateAvatar)
	// 删除用户
	r.DELETE("/delete", Admin, res.delete)
	// 展示下拉框名称列表
	r.GET("/nameList", Authed, res.nameList)
	// 查询用户个人信息
	r.GET("/info", Authed, res.info)
	// 用户编辑个人信息
	r.POST("/edit", User, res.edit)
	// 管理员编辑用户信息
	r.POST("/adminEdit", Admin, res.adminEdit)
	return res
}

/**
@api {POST} /api/user/create 用户创建
@apiDescription 创建用户，用户名不能重复。
创建用户时要求输入工号，若输入姓名，则生成姓名的拼音缩写。
@apiName UserCreate
@apiGroup User

@apiPermission 管理员

@apiParam {String} openid 工号。
@apiParam {String{8..16}} password 口令，长度至少大于8。
@apiParam {String} [name] 用户姓名。
@apiParam {String} [username] 用户名（若用户未填写用户名，则默认工号）。
@apiParam {String} [phone] 手机号。
@apiParam {String} [email] 邮箱。
@apiParam {String} [sn] 身份证号。

@apiSuccess {Integer} id 用户ID。
@apiSuccess {String} openid 工号。
@apiSuccess {String} name 用户姓名。
@apiSuccess {String} createdAt 创建时间，格式为"YYYY-MM-DD HH:mm:ss"。

@apiParamExample {json} 请求示例
{
	"openid":"1001",
    "password": "Gm123qwe",
    "name": "张三",
	"username":"1001",
    "phone":"13855555555",
	"email":"123@qq.com",
	"sn":""
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 1,
	"openid":"1001",
    "name": "张三",
    "createdAt": "2020-08-24 16:26:16"
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

用户名已经存在
*/

// create 创建用户
func (c *UserController) create(ctx *gin.Context) {
	var info entity.User
	err := ctx.BindJSON(&info)
	applog.L(ctx, "创建用户", map[string]interface{}{
		"name": info.Name,
	})
	// 工号不能为空
	if len(strings.Trim(info.Openid, " ")) == 0 {
		ErrIllegal(ctx, "工号不能为空")
		return
	}
	// 工号唯一
	exist, err := repo.UserRepo.ExistOpenid(info.Openid)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "工号已经存在")
		return
	}

	// 若用户名为空，则将工号设置为用户名
	if len(strings.Trim(info.Username, " ")) == 0 {
		info.Username = info.Openid
	}

	exist, err = repo.UserRepo.ExistUsername(info.Username)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "用户名已经存在")
		return
	}

	//姓名转化为拼音首字母
	if len(info.Name) > 0 {
		str, err := reuint.PinyinConversion(info.Name)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		info.NamePinyin = str
	}

	// 手机号格式校验
	if len(info.Phone) != 0 && !reuint.PhoneValidate(info.Phone) {
		ErrIllegal(ctx, "手机号格式错误")
		return
	}
	// 手机号唯一
	if len(info.Phone) != 0 {
		exist, err := repo.UserRepo.ExistPhone(info.Phone)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "手机号已经存在")
			return
		}
	}

	// 邮箱格式校验
	if len(info.Email) != 0 && !reuint.EmailValidate(info.Email) {
		ErrIllegal(ctx, "邮箱格式错误")
		return
	}
	// 邮箱唯一
	if len(info.Email) != 0 {
		exist, err := repo.UserRepo.ExistEmail(info.Email)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "邮箱已经存在")
			return
		}
	}

	// 身份证号格式校验
	if len(info.Sn) != 0 && !reuint.SnValidate(info.Sn) {
		ErrIllegal(ctx, "身份证号格式错误")
		return
	}

	// 口令长度 大于等8位
	if len(strings.Trim(info.Password.String(), " ")) < 8 {
		ErrIllegal(ctx, "口令长度不少于8位")
		return
	}
	pwd, salt, err := reuint.GenPasswordSalt(info.Password.String())
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 密码和盐值
	info.Password = entity.Pwd(pwd)
	info.Salt = salt
	if err = repo.DB.Create(&info).Error; err != nil {
		ErrSys(ctx, err)
		return
	}

	reqInfo := dto.UserCreateDto{}
	reqInfo.Transform(&info)
	ctx.JSON(200, reqInfo)
}

/**
@api {GET} /api/user/search 查找用户
@apiDescription 查找用户，可以通过关键模糊查找。
@apiName UserSearch
@apiGroup User

@apiPermission 管理员

@apiParam {String} [keyword] 用户名、姓名、姓名拼音缩写。
@apiParam {Integer} [page=1] 分页查询页码，表示第几页，默认 1。
@apiParam {Integer} [limit=20] 单页多少数据，默认 20。

@apiParamExample {get} 请求示例
GET /api/user/search?keyword=zs&page=1&limit=20

@apiSuccess {User[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {User} User 用户数据结构。
@apiSuccess {Integer} User.id 用户ID。
@apiSuccess {String} User.openid 工号，不可重复。
@apiSuccess {String} User.name 姓名。
@apiSuccess {String} User.createdAt 创建时间。
@apiSuccess {String} User.updatedAt 更新时间。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
	"records": [
		{
		    "id": 5,
		    "createdAt": "2020-09-26 11:29:44",
		    "updatedAt": "2020-09-26 11:29:44",
		    "openid": "1001",
		    "name": "测试人名",
		}
    ],
	"total": 19,
    "size": 2,
    "current": 1,
    "pages": 10
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

用户不存在
*/

// search 查找用户
func (c *UserController) search(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	reqInfo, tx := repo.NewPageQueryFnc(repo.DB, &entity.User{}, page, limit, func(db *gorm.DB) *gorm.DB {
		//拼音模糊条件
		queryPinyin := db.Where("name_pinyin like ?", fmt.Sprintf("%%%s%%", keyword))
		//姓名模糊条件
		queryName := db.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
		//用户名模糊条件
		queryUserName := db.Where("username like ?", fmt.Sprintf("%%%s%%", keyword))
		//模糊查询
		db = db.Where("is_delete", 0).Where(queryPinyin.Or(queryName).Or(queryUserName))
		return db
	})
	res := []entity.User{}
	err = tx.Find(&res).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Records = res
	ctx.JSON(200, &reqInfo)
}

/**
@api {POST} /api/user/modifyPwd 修改口令
@apiDescription 用户修改口令。
@apiName UserModifyPwd
@apiGroup User

@apiPermission 管理员,用户

@apiParam {Integer} id 用户ID。
@apiParam {String} oldPwd 原口令。
@apiParam {String} newPwd 新口令，8字符以上。

@apiParamExample {json} 请求示例
{
    "id": 1,
    "oldPwd": "12345678",
    "newPwd": "Gm123qwe"
}

@apiSuccess {String} body 修改通过成功状态码200，否则返还错误码。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

新口令为空
*/

// modifyPwd 修改口令
func (c *UserController) modifyPwd(ctx *gin.Context) {
	var info dto.PasswordDto
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "用户修改口令", map[string]interface{}{
		"userId": info.ID,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	// 普通用户仅能修改个人密码
	if claims.Sub != info.ID {
		ErrForbidden(ctx, "权限错误")
		return
	}

	//数据库搜索用户
	reqInfo := &entity.User{}
	err = repo.DB.First(reqInfo, "id = ? AND is_delete = 0", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该用户")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	//旧口令正确性校验
	if reuint.VerifyPasswordSalt(info.OldPwd.String(), reqInfo.Password.String(), reqInfo.Salt) == false {
		ErrIllegal(ctx, "旧口令错误")
		return
	}

	//新口令长度校验
	if len(info.NewPwd.String()) < 8 {
		ErrIllegal(ctx, "口令长度不少于8位")
		return
	}

	pwd, salt, err := reuint.GenPasswordSalt(info.NewPwd.String())
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Password = entity.Pwd(pwd)
	reqInfo.Salt = salt

	err = repo.DB.Save(reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/user/resetPwd 重置口令
@apiDescription 重置用户口令
@apiName UserResetPwd
@apiGroup User

@apiPermission 管理员

@apiParam {Integer} id 用户ID。
@apiParam {String} newPwd 新口令，8字符以上。

@apiParamExample {json} 请求示例
{
    "id": 1,
    "newPwd": "Gm123qwe"
}

@apiSuccess {String} body 修改通过成功状态码200，否则返还错误码。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误

*/

// resetPwd 重置口令
func (c *UserController) resetPwd(ctx *gin.Context) {
	var info dto.PasswordDto
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 记录日志
	applog.L(ctx, "管理员重置口令", map[string]interface{}{
		"userId": info.ID,
	})

	if len(info.NewPwd) == 0 {
		ErrIllegal(ctx, "请输入口令")
		return
	}

	//数据库搜索用户
	reqInfo := &entity.User{}
	err := repo.DB.First(reqInfo, "id = ? AND is_delete = 0", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "不存在该用户")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	//新口令长度校验
	if len(info.NewPwd.String()) < 8 {
		ErrIllegal(ctx, "口令长度不少于8位")
		return

	}

	pwd, salt, err := reuint.GenPasswordSalt(info.NewPwd.String())
	if err != nil {
		ErrSys(ctx, err)
		return

	}
	reqInfo.Password = entity.Pwd(pwd)
	reqInfo.Salt = salt

	if err = repo.DB.Save(reqInfo).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/user/updateAvatar 更换头像
@apiDescription 更新用户头像，json传输base64，仅支持支持小于64k jpeg、png。
@apiName UserUpdateAvatar
@apiGroup User

@apiPermission 用户


@apiParam {String} username 用户名。
@apiParam {String} avatar 头像文件，仅支持支持小于64k jpeg、png,例如格式：data:image/jpeg;base64

@apiParamExample {json} 请求示例
{
    "id": "13",
    "avatar": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCABZAGYDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwDLuJ5hcygTSZ3n+I+tR/aJwf8AXyf99mn3C/6TP/vt/OowvXPfvXZc8NseLicD/XSf99mjz5wf9dJ/32aZg59aUKc9TTFclFxPj/XSf99ml+0T9fOf/vs0wClxxS6hceLmc9ZZP++jUgnnBz50n/fRqHp25p+OP1pBcf58+f8AXSf99Gj7RP3mk4/2j0pnXqKcB7UBceJ5+vnSf99GnCefOfOk4/2jzTAMe9OxzkigLl7TZJTctmRyNh6sfUUUaV/x8vjpsP8AMUVLNYPQyLgYupv98/zpgB6H1qa4H+ly4/vn+dM/pTMmNFKBx0p4HNLgY64piEAox396XBp+O+DigYwJjovNOxz0pwH1pcc8DmgBuMdadjI560oGDS4OaBXDHp2NG3r7U4Dil20AW9MBFywzj5D/ADFFO0w4uW7/ACH+YoqWaw2MufH2mb2c/wA6jA68c1LOB9pl/wB8/wA6bjAHAzTM2NwQKdjP5Uo5znpTh9OnFMVxNopQPb/69OwKMfSgLjQv0p3QD0pev+GKMUCuHbpTsUU7Gc0AhPwpeBRzT8HHQk/SgZa0z/j6f/cP8xRT9MH+kN/uH+YoqWaw2MidT9plPX5z/OkxyKkuAPtMpx/y0P8AOmjNMxkxAMD2pwGPwpe1HHvzTQriGl69hSgDvShe9AXADvijBJ70oAz9KcAOnNAmJjvSr0p2OBRgZ+lAABxz1p2OfegLk08ChlJ6FvThi4b/AHD29xRT9PA+0N/uf1FFSzSD0Mecf6VLx0c/zpoqScD7RL/vn+dMA6+1UjF7hgemB0pcD34pcZ4oHPegA59qXBpfXpS49KdhNiY/OlAGOlKB3wacF6+1FguH8qdto5znAzTgD60rBcAMf56U4D8qADzSjnpQMt6aM3DcEjZ/UUUum4E7f7v9RRUs1hsZNxgXUnHG8/zqPHfGa6OT7z/U00dPxNVcTp+ZgdRxSj8K6AdBSnpT5iXT8znwAfwpQOOgroO5pw+6aOYfs/MwAvJyKUewrfH+FOHU0ri9n5mAAaUdelb4pKEw9n5mIPy7UpxnkVup/WlboKLh7MzdPH+kNx/D/UUVsQ/6w/SiobNIwsj/2Q=="
}


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

用户记录ID非法

@apiErrorExample 失败响应2
HTTP/1.1 400 Bad Request

图片格式错误，无法解析
*/

// updateAvatar 更换头像
func (c *UserController) updateAvatar(ctx *gin.Context) {

}

/**
@api {DELETE} /api/user/delete 删除用户
@apiDescription 删除用户，如果存在多个用户，其中某个用户删除失败，依然返回200状态码。
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName UserDelete
@apiGroup User

@apiPermission 管理员

@apiParam {String} ids 待删除的ID序列，多个ID用","隔开，如：ids=1,99。

@apiParamExample 请求示例
DELETE /api/user/delete?ids=12,24

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// delete 删除用户
func (c *UserController) delete(ctx *gin.Context) {
	ids := ctx.Query("ids")
	//将string转化为[]int
	idArray := reuint.StrToIntSlice(ids)

	// 记录日志
	applog.L(ctx, "删除用户", map[string]interface{}{
		"ids": ids,
	})

	// 将is_delete字段赋值为1
	err := repo.DB.Model(&entity.User{}).Where("id in ?", idArray).Update("is_delete", 1).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

}

/**
@api {GET} /api/user/nameList 所有成员
@apiDescription 查询所有的成员
@apiName UserNameList
@apiGroup User

@apiPermission 管理员，项目负责人

@apiParam {String} keyword 关键字。

@apiParamExample 请求示例
GET /api/user/nameList?keyword=zs

@apiSuccess {Integer} id 用户ID。
@apiSuccess {String} openid 工号。
@apiSuccess {String} name 用户姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    { "id": 1,"openid":1001, "name": "张三"},
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// nameList 展示下拉框名称列表
func (c *UserController) nameList(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	reqInfo := &[]dto.NameListDto{}

	//拼音模糊条件
	queryPinyin := repo.DB.Where("name_pinyin like ?", fmt.Sprintf("%%%s%%", keyword))
	//姓名模糊条件
	queryName := repo.DB.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
	//用户名模糊条件
	queryUserName := repo.DB.Where("username like ?", fmt.Sprintf("%%%s%%", keyword))
	//模糊查询
	err := repo.DB.Model(&entity.User{}).Select("id,name,avatar").Where("is_delete", 0).Where(queryPinyin.Or(queryName).Or(queryUserName)).Find(reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	ctx.JSON(200, *reqInfo)
}

/**
@api {GET} /api/user/info 用户个人信息
@apiDescription 查询用户个人信息
@apiName UserInfo
@apiGroup User

@apiPermission 管理员，用户

@apiParam {Integer} id 用户ID

@apiParamExample 请求示例
GET /api/user/info?id=1

@apiSuccess {Integer} id 用户ID。
@apiSuccess {String} username 用户名，不可重复。
@apiSuccess {String} [name] 用户姓名。
@apiSuccess {String} openid 用户工号。
@apiSuccess {String} phone 手机号。
@apiSuccess {String} email 邮箱。
@apiSuccess {String} sn 身份证号。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    { "id":1,"openid": 1001,"username":"zs", "name": "张三", "phone":"13855555555","email":"123@qq.com","sn":""}
]

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// info 查询用户个人信息
func (c *UserController) info(ctx *gin.Context) {
	var reqInfo dto.UserInfoDto
	userId, _ := strconv.Atoi(ctx.Query("id"))
	if userId <= 0 {
		ErrIllegal(ctx, "参数异常，无法解析")
		return
	}

	// 获取用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	if claims.Type == "user" && claims.Sub != userId {
		ErrIllegal(ctx, "权限错误")
		return
	}
	err := repo.DB.Model(&entity.User{}).Select("id,openid,username,name,phone,email,sn").First(&reqInfo, "id = ? AND is_delete = ?", userId, 0).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, reqInfo)
}

/**
@api {POST} /api/user/edit 用户编辑个人信息
@apiDescription 修改用户信息，用户仅可以修改自己的部分信息，如：手机号，邮箱。
@apiName UserEdit
@apiGroup User

@apiPermission 用户

@apiParam {String} [phone] 手机号。
@apiParam {String} [email] 邮箱。

@apiParamExample {json} 请求示例
{
	"phone":"13855555555",
	"email":"123@qq.com",
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// edit 用户编辑个人信息
func (c *UserController) edit(ctx *gin.Context) {
	var info entity.User
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	// 记录日志
	applog.L(ctx, "用户编辑个人信息", map[string]interface{}{
		"id": claims.Sub,
	})

	// 登录者为普通用户 仅可修改自己的信息（手机号、邮箱）且不可修改工号、姓名
	var reqInfo entity.User
	err := repo.DB.First(&reqInfo, "id = ? AND is_delete = 0", claims.Sub).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	// 手机号格式校验
	if info.Phone != "" && !reuint.PhoneValidate(info.Phone) {
		ErrIllegal(ctx, "手机号格式错误")
		return
	}

	if info.Phone != "" && info.Phone != reqInfo.Phone {
		// 手机号唯一
		exist, err := repo.UserRepo.ExistPhone(info.Phone)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "手机号已经存在")
			return
		}
	}
	reqInfo.Phone = info.Phone

	// 邮箱格式校验
	if info.Email != "" && !reuint.EmailValidate(info.Email) {
		ErrIllegal(ctx, "邮箱格式错误")
		return
	}
	if info.Email != "" && info.Email != reqInfo.Email {
		// 邮箱唯一
		exist, err := repo.UserRepo.ExistEmail(info.Email)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "邮箱已经存在")
			return
		}
	}
	reqInfo.Email = info.Email

	if err = repo.DB.Save(&reqInfo).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {POST} /api/user/adminEdit 管理员编辑用户信息
@apiDescription 修改用户信息，管理员具有修改用户全部信息的权限。
@apiName UserAdminEdit
@apiGroup User

@apiPermission 管理员

@apiParam {Integer} id 用户ID。
@apiParam {String} [openid] 工号。
@apiParam {String} [username] 用户名（若用户未填写用户名，则默认工号）。
@apiParam {String} [name] 姓名。
@apiParam {String} [phone] 手机号。
@apiParam {String} [email] 邮箱。
@apiParam {String} [sn] 身份证号。

@apiParamExample {json} 请求示例
{
    "id": 1,
	"openid":"1001",
	"username":"1001",
	"name":"张三",
	"phone":"13855555555",
	"email":"123@qq.com",
	"sn":""
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 500

系统内部错误
*/

// adminEdit 管理员编辑用户信息
func (c *UserController) adminEdit(ctx *gin.Context) {
	var info entity.User
	if err := ctx.BindJSON(&info); err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 记录日志
	applog.L(ctx, "管理员编辑用户信息", map[string]interface{}{
		"id": info.ID,
	})

	if info.Openid == "" {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	var reqInfo entity.User
	err := repo.DB.First(&reqInfo, "id = ? AND is_delete = 0", info.ID).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "用户不存在或被删除")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	reqInfo.Name = info.Name
	if info.Name != "" {
		str, err := reuint.PinyinConversion(info.Name)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		reqInfo.NamePinyin = str
	} else {
		reqInfo.NamePinyin = ""
	}

	// 若输入的用户名为空，则将其设置为工号
	if info.Username == "" {
		info.Username = info.Openid
	}
	// 若发生用户名修改，用户名唯一检查
	if reqInfo.Username != info.Username {
		existUsername, err := repo.UserRepo.ExistUsername(info.Username)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if existUsername {
			ErrIllegal(ctx, "用户名已经存在")
			return
		}
	}
	reqInfo.Username = info.Username

	// 若发生工号修改，工号唯一检查
	if reqInfo.Openid != info.Openid {
		existOpenId, err := repo.UserRepo.ExistOpenid(info.Openid)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if existOpenId {
			ErrIllegal(ctx, "工号已经存在")
			return
		}
	}
	reqInfo.Openid = info.Openid

	// 手机号格式校验
	if info.Phone != "" && !reuint.PhoneValidate(info.Phone) {
		ErrIllegal(ctx, "手机号格式错误")
		return
	}
	// 若发生手机号修改，手机号唯一检查
	if info.Phone != "" && reqInfo.Phone != info.Phone {
		// 手机号唯一
		exist, err := repo.UserRepo.ExistPhone(info.Phone)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "手机号已经存在")
			return
		}
	}
	reqInfo.Phone = info.Phone

	// 邮箱格式校验
	if info.Email != "" && !reuint.EmailValidate(info.Email) {
		ErrIllegal(ctx, "邮箱格式错误")
		return
	}
	// 若发生邮箱变化，且新邮箱不为空，邮箱唯一检查
	if info.Email != "" && reqInfo.Email != info.Email {
		// 邮箱唯一
		exist, err := repo.UserRepo.ExistEmail(info.Email)
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		if exist {
			ErrIllegal(ctx, "邮箱已经存在")
			return
		}
	}
	reqInfo.Email = info.Email

	// 身份证号格式校验
	if info.Sn != "" && !reuint.SnValidate(info.Sn) {
		ErrIllegal(ctx, "身份证号格式错误")
		return
	}
	reqInfo.Sn = info.Sn
	if err = repo.DB.Save(&reqInfo).Error; err != nil {
		ErrSys(ctx, err)
		return
	}
}