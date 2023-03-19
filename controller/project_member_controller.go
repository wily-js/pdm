package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/dto"
	"pdm/logg/applog"
	"pdm/repo"
	"pdm/repo/entity"
	"strconv"
)

// ProjectMemberController 项目成员控制器
type ProjectMemberController struct {
}

func NewProjectMemberController(router gin.IRouter) *ProjectMemberController {
	res := &ProjectMemberController{}
	r := router.Group("/member")
	// 添加成员
	r.POST("/add", Manager, res.add)
	// 修改角色
	r.POST("/change", Manager, res.change)
	// 删除成员
	r.DELETE("/delete", Manager, res.delete)
	// 查询所有成员
	r.GET("/all", ProjectMember, res.all)
	// 修改成员任务描述
	return res
}

/**
@api {POST} /api/project/member/add 添加成员
@apiDescription 添加成员。
注意项目的 负责人 类型不运行添加，该类型仅由管理员在创建项目时指定。
@apiName MemberAdd
@apiGroup Member

@apiPermission 项目负责人

@apiParam {Integer} role 角色类型：
<ul>
    <li>0 - 访客</li>
    <li>1 - 测试</li>
    <li>2 - 开发</li>
    <li>3 - 维护</li>
</ul>

@apiParam {Integer} projectId 项目ID。
@apiParam {Integer[]} userIds 用户ID。

@apiParamExample {json} 请求示例
{
    "role": 2,
    "projectId": 1,
    "userId": [2]
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

权限错误
*/

// add 添加成员
func (c ProjectMemberController) add(ctx *gin.Context) {
	var reqInfo entity.ProjectMember
	var info dto.MemberDTO
	err := ctx.BindJSON(&info)

	// 记录日志
	applog.L(ctx, "添加成员", map[string]interface{}{
		"projectId": info.ProjectId,
		"userId":    info.UserId,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	if len(info.UserId) == 0 {
		ErrIllegal(ctx, "请添加成员名称")
		return
	}

	if info.Role == entity.RoleCreator {
		ErrIllegal(ctx, "无法添加项目负责人")
		return
	}
	// 判断项目是否存在
	exist, err := repo.ProjectRepo.Exist(info.ProjectId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "项目不存在或已被删除，无法添加")
		return
	}

	reqInfo.ProjectId = info.ProjectId
	reqInfo.Role = info.Role

	// userId 数组
	idArr := info.UserId

	//事务处理
	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range idArr {
			reqInfo.UserId = id
			// 插入新记录时ID自增，不可有值
			reqInfo.ID = 0
			// 判断用户是否被删除
			exist, err := repo.UserRepo.Exist(id)
			if err != nil {
				return err
			}
			if !exist {
				return errors.New("用户不存在或被删除")
			}
			// 判断用户在该项目中是否存在角色
			exist, err = repo.ProjectMemberRepo.Exist(reqInfo.ProjectId, reqInfo.UserId)
			if err != nil {
				return err
			}
			if exist {
				return errors.New("项目中已存在该用户")
			}
			err = tx.Create(&reqInfo).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
}

/**
@api {POST} /api/project/member/change 修改角色
@apiDescription 修改成员的角色，注意不允许修改为 负责人角色。
@apiName MemberChange
@apiGroup Member

@apiPermission 项目负责人

@apiParam {Integer} id 记录ID。
@apiParam {Integer} projectId 项目ID。

@apiParam {Integer} role 角色类型：
<ul>
    <li>0 - 访客</li>
    <li>1 - 测试</li>
    <li>2 - 开发</li>
    <li>3 - 维护</li>
</ul>


@apiParamExample {json} 请求示例

{
    "id": 2,
	"projectId":11,
    "role": 3
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

权限错误
*/

// change 修改角色
func (c ProjectMemberController) change(ctx *gin.Context) {
	var reqInfo entity.ProjectMember
	err := ctx.BindJSON(&reqInfo)
	// 记录日志
	applog.L(ctx, "修改成员的角色", map[string]interface{}{
		"id":   reqInfo.ID,
		"role": reqInfo.Role,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	if reqInfo.Role == entity.RoleCreator {
		ErrIllegal(ctx, "不可修改为项目负责人")
		return
	}

	memberInfo := &entity.ProjectMember{}
	err = repo.DB.First(memberInfo, "id = ? AND project_id = ?", reqInfo.ID, reqInfo.ProjectId).Error
	if err == gorm.ErrRecordNotFound {
		ErrIllegal(ctx, "项目中不存在该成员")
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if memberInfo.Role == entity.RoleCreator {
		ErrIllegal(ctx, "项目负责人不可修改")
		return
	}
	memberInfo.Role = reqInfo.Role
	err = repo.DB.Save(memberInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {DELETE} /api/project/member/delete 删除成员
@apiDescription 删除成员，注意 负责人 不允许删除！
该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName MemberDelete
@apiGroup Member

@apiPermission 项目负责人

@apiParam {String} id 项目成员记录ID。
@apiParam {String} projectId 项目ID。

@apiParamExample 请求示例
DELETE /api/project/member/delete?projectId=11&id=12

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

权限错误
*/

// delete 删除成员
func (c ProjectMemberController) delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	// 记录日志
	applog.L(ctx, "删除成员", map[string]interface{}{
		"id": id,
	})

	res := &entity.ProjectMember{}
	err := repo.DB.First(res, "id = ? AND project_id = ?", id, projectId).Error
	// 没有该记录
	if err == gorm.ErrRecordNotFound {
		return
	}
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 用户为项目负责人不进行删除操作
	if res.Role == 4 {
		return
	}
	err = repo.DB.Delete(&entity.ProjectMember{}, id).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/project/member/all 所有成员
@apiDescription 查询所有的成员
@apiName MemberAll
@apiGroup Member

@apiPermission 项目负责人，项目成员

@apiParam {String} projectId 项目ID
@apiParam {String} keyword 用户名、姓名、姓名拼音缩写。
@apiParamExample 请求示例
GET /api/project/member/all?projectId=12&keyword=zs

@apiParam {MemberUser[]} body 响应体。

@apiParam {Object} MemberUser 成员用户信息。
@apiParam {Integer} MemberUser.id 记录ID。
@apiParam {Integer} MemberUser.userId 用户ID。
@apiParam {Integer} MemberUser.role 成员角色。
@apiParam {String} MemberUser.name 姓名。


@apiSuccessExample 成功响应
HTTP/1.1 200 OK

[
    {"id": 12, "userId": 1, "role": 4, "name": "张三"},
    {"id": 17, "userId": 2, "role": 2, "name": "郭小菊"}
]

@apiErrorExample 失败响应1
HTTP/1.1 500

系统内部错误

@apiErrorExample 失败响应2
HTTP/1.1 400

权限错误
*/

// all 查询所有成员
func (c ProjectMemberController) all(ctx *gin.Context) {
	var reqInfo []dto.MemberAllDTO
	projectId, _ := strconv.Atoi(ctx.Query("projectId"))
	if projectId <= 0 {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}
	keyword := ctx.Query("keyword")

	// 判断项目是否存在
	exist, err := repo.ProjectRepo.Exist(projectId)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if !exist {
		ErrIllegal(ctx, "该项目不存在或被删除")
		return
	}
	//拼音模糊条件
	queryPinyin := repo.DB.Where("name_pinyin like ?", fmt.Sprintf("%%%s%%", keyword))
	//姓名模糊条件
	queryName := repo.DB.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
	//用户名模糊条件
	queryUserName := repo.DB.Where("username like ?", fmt.Sprintf("%%%s%%", keyword))

	// 联表后条件查询
	//SELECT project_members.id,user_id,project_id,role,username,name,name_pinyin,is_delete
	//	FROM `project_members` left join users
	//	on project_members.user_id = users.id
	//	WHERE (is_delete = 0 AND project_id = 1) AND (name_pinyin like '%zs%' OR name like '%zs%' OR username like '%zs%')
	err = repo.DB.Table("project_members").
		Select("project_members.id,user_id,project_id,role,username,name,name_pinyin,is_delete").
		Joins("left join users  on project_members.user_id = users.id").
		Where("is_delete = ? AND project_id = ?", 0, projectId).Where(queryPinyin.Or(queryName).Or(queryUserName)).Find(&reqInfo).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}

	ctx.JSON(200, reqInfo)
}
