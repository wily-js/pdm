package controller

import (
	"errors"
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

// NewProjectController 创建项目控制器
func NewProjectController(router gin.IRouter) *ProjectController {
	res := &ProjectController{}
	r := router.Group("/project")
	// 创建项目
	r.POST("/create", Authed, res.create)
	// 搜索项目
	r.GET("/search", Authed, res.search)
	// 编辑项目
	r.POST("/edit", Admin, res.edit)
	// 删除项目
	r.DELETE("/delete", Admin, res.delete)
	// 查询项目信息
	r.GET("/info", Admin, res.info)

	return res
}

// ProjectController 项目控制器
type ProjectController struct {
}

/**
@api {POST} /api/project/create 创建
@apiDescription 创建项目，项目名称不能重复，项目描述文本即可，
在写入数据库时需要生成项目名称的拼音缩写。
创建项目时同时在项目成员中添加项目负责人记录。
@apiName ProjectCreate
@apiGroup Project

@apiPermission 管理员、用户

@apiParam {String} name 项目名称，不能重复。
@apiParam {String} description 项目描述。
@apiParam {Integer} manager 项目负责人ID（用户ID）
@apiParam {String} version 版本号。

@apiSuccess {Integer} id 项目ID
@apiSuccess {String} name 项目名称，不能重复。
@apiSuccess {Integer} manager 项目负责人ID（用户ID）

@apiParamExample {json} 请求示例
{
    "name": "研发项目管理系统",
    "description": "管理维护与项目开发相关的各类文档以及资料，提供项目生命周期的管理。",
    "manager": 13,
	"version":V1.0.0
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 1,
    "name": "研发项目管理系统",
    "description": "管理维护与项目开发相关的各类文档以及资料，提供项目生命周期的管理。",
    "manager": 13
}

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

项目名已经存在
*/

// create 创建项目
func (c *ProjectController) create(ctx *gin.Context) {
	var info entity.Project
	var member entity.ProjectMember
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "创建项目", map[string]interface{}{
		"name":    info.Name,
		"manager": info.Manager,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 项目名唯一
	if info.Name == "" {
		ErrIllegal(ctx, "项目名不能为空")
		return
	}

	exist, err := repo.ProjectRepo.NameExist(info.Name)
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if exist {
		ErrIllegal(ctx, "项目名已经存在")
		return
	}

	//项目名称拼音缩写生成
	str, err := reuint.PinyinConversion(info.Name)
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
	info.NamePinyin = str
	info.IsDelete = 0

	if info.Version != "" && !strings.HasPrefix(info.Version, "V") {
		info.Version = fmt.Sprintf("V%s", info.Version)
	}

	// 事务处理  创建项目记录 项目负责人记录
	if err = repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&info).Error; err != nil {
			return err
		}
		// 项目成员：负责人
		member.Role = 3
		member.ProjectId = info.ID
		member.UserId = info.Manager
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		ErrSys(ctx, err)
		return
	}

	var reqInfo dto.ProjectDto
	reqInfo.Transform(&info)
	ctx.JSON(200, reqInfo)
}

/**
@api {GET} /api/project/search 搜索
@apiDescription 搜索项目，支持分页查询，
查询条件支持项目的名称或拼音缩写，以及项目的状态。
项目管理支持查询所有项目，普通用户仅支持查询与自己有关的项目。
@apiName ProjectSearch
@apiGroup Project

@apiPermission 管理员,用户

@apiParam {String} [keyword] 项目名、项目名拼音、项目简介缩写
@apiParam {Integer} [page=1] 分页查询页码，表示第几页，默认 1。
@apiParam {Integer} [limit=20] 单页多少数据，默认 20。

@apiParamExample {get} 请求示例
GET /api/project/search?keyword=cs

@apiSuccess {project[]} records 查询结果列表。
@apiSuccess {Integer} total 记录总数。
@apiSuccess {Integer} size 每页显示条数，默认 20。
@apiSuccess {Integer} current 当前页。
@apiSuccess {Integer} pages 总页数。

@apiSuccess {Object} Project 用户数据结构。
@apiSuccess {Integer} Project.id 项目ID。
@apiSuccess {String} Project.name 用户姓名。
@apiSuccess {String} Project.createdAt 创建时间。
@apiSuccess {String} Project.updatedAt 更新时间。
@apiSuccess {String} [Project.version] 版本号。
@apiSuccess {Object} [Project.manager] 负责人信息。
@apiSuccess {Integer} [Project.manager.id] 用户ID。
@apiSuccess {String} [Project.manager.name] 姓名。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
	"records": [
		{
		    "id": 5,
		    "createdAt": "2020-09-26 11:29:44",
		    "updatedAt": "2020-09-26 11:29:44",
		    "name": "测试项目",
            "version": "",
            "manager": { "id" : 13, "name":"张三"}
		}
    ],
	"total": 19,
    "size": 2,
    "current": 1,
    "pages": 10
}
@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

权限错误
*/

// search 搜索项目
func (c *ProjectController) search(ctx *gin.Context) {

	// 检查用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

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

	var reqInfo []dto.ProjectSearchDto
	projectIdList := make([]int, 0)

	// 若为用户，则获取其参与的所有项目的ID
	if claims.Type == "user" {
		var projectInfo []entity.ProjectMember
		err = repo.DB.Select("project_id").Where("user_id", claims.Sub).Find(&projectInfo).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		// 获取项目id
		for _, member := range projectInfo {
			projectIdList = append(projectIdList, member.ProjectId)
		}
	}

	// 查询 项目表
	query, tx := repo.NewPageQueryFnc(repo.DB, []entity.Project{}, page, limit, func(db *gorm.DB) *gorm.DB {
		// 请求头 keyword 是否为空 (项目名称，拼音)
		if keyword != "" {
			// 项目名称模糊查询
			queryName := db.Where("name like ?", fmt.Sprintf("%%%s%%", keyword))
			// 项目名称拼音模糊查询
			queryPinyin := db.Where("name_pinyin like ?", fmt.Sprintf("%%%s%%", keyword))
			// 项目简介模糊查询
			queryDescription := db.Where("description like ?", fmt.Sprintf("%%%s%%", keyword))
			// 模糊查询
			db = db.Where(queryName.Or(queryPinyin).Or(queryDescription))
		}
		db = db.Where("is_delete", 0)
		// 项目未被删除
		if claims.Type == "user" {
			db = db.Where("id in ?", projectIdList)
			return db
		}
		return db
	})

	projects := []entity.Project{}

	err = tx.Find(&projects).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	// 遍历查询结果
	for _, project := range projects {
		user := entity.User{}
		// 查询 负责人信息
		err := repo.DB.Select("name").Where("id", project.Manager).Find(&user).Error
		if err != nil {
			ErrSys(ctx, err)
			return
		}
		template := dto.ProjectSearchDto{}
		template.Transform(&project, &user)
		reqInfo = append(reqInfo, template)
	}
	query.Records = reqInfo
	ctx.JSON(200, query)

}

/**
@api {POST} /api/project/edit 编辑
@apiDescription 编辑项目，项目名称不能重复，项目描述文本即可，
可以关键字（拼音缩写、姓名、用户名）查询项目负责人。
在写入数据库时需要生成项目名称的拼音缩写。
在项目负责人发生变化是同时更新项目成员表（project_members）
@apiName ProjectEdit
@apiGroup Project

@apiPermission 管理员

@apiParam {Integer} id 项目ID。
@apiParam {String} name 项目名称，不能重复。
@apiParam {String} description 项目描述。
@apiParam {Integer} manager 项目负责人ID（用户ID）。

@apiParamExample {json} 请求示例
{
    "id": 1,
    "name": "研发项目管理系统",
    "description": "管理维护与项目开发相关的各类文档以及资料，提供项目生命周期的管理。",
    "manager": 13
}

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

项目名(name)已经存在
*/

// edit 项目编辑
func (c *ProjectController) edit(ctx *gin.Context) {
	var info dto.ProjectDto
	var project entity.Project
	err := ctx.BindJSON(&info)
	// 记录日志
	applog.L(ctx, "编辑项目", map[string]interface{}{
		"id":      info.ID,
		"name":    info.Name,
		"manager": info.Manager,
	})
	if err != nil {
		ErrIllegal(ctx, "参数非法，无法解析")
		return
	}

	// 查找
	err = repo.DB.Where("id = ? AND is_delete = 0", info.ID).Find(&project).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if project.ID == 0 {
		ErrIllegal(ctx, "项目不存在或项目已经被删除")
		return
	}

	// 事务处理 ， 确定 项目成员表 和 项目表 都完成更新
	err = repo.DB.Transaction(func(tx *gorm.DB) error {
		// 如果对项目名称进行了修改
		if project.Name != info.Name {
			// 项目名唯一
			if info.Name == "" {
				return errors.New("项目名(name)不能为空")
			}
			exist, err := repo.ProjectRepo.NameExist(info.Name)
			if err != nil {
				return err
			}
			if exist {
				return errors.New("项目名(name)已经存在")
			}
			project.Name = info.Name
			str, err := reuint.PinyinConversion(project.Name)
			if err != nil {
				return err
			}
			project.NamePinyin = str
		}

		// 项目描述进行修改
		project.Description = info.Description

		// 如果对项目负责人进行了修改
		if info.Manager != project.Manager {
			// 找到 该项目的项目负责人 并  更新项目负责人
			err := tx.Model(&entity.ProjectMember{}).Where("project_id = ? AND role = ?", info.ID, UserRoleProjectManager).Update("user_id", info.Manager).Error
			if err != nil {
				return err
			}
			project.Manager = info.Manager
		}
		err := tx.Model(&entity.Project{}).Where("id", info.ID).Updates(project).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ErrIllegalE(ctx, err)
		return
	}
}

/**
@api {DELETE} /api/project/delete 删除项目
@apiDescription 删除项目，该接口仅在数据库操作异常时返回500系统错误的状态码，其他情况均返回200。
@apiName ProjectDelete
@apiGroup Project

@apiPermission 管理员

@apiParam {String} id 项目ID。

@apiParamExample 请求示例
DELETE /api/project/delete?id=12

@apiSuccessExample 成功响应
HTTP/1.1 200 OK

@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

参数非法，无法解析
*/

// delete 删除项目
func (c *ProjectController) delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	// 记录日志
	applog.L(ctx, "删除项目", map[string]interface{}{
		"id": id,
	})
	if id <= 0 {
		ErrIllegal(ctx, "参数非法,无法解析")
		return
	}

	err := repo.DB.Model(&entity.Project{}).Where("id", id).Update("is_delete", 1).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
}

/**
@api {GET} /api/project/info 查询项目信息
@apiDescription 查询项目信息
管理员可以查询所有项目信息，普通用户只可以查找和自己有关项目的项目信息
@apiName ProjectInfo
@apiGroup Project

@apiPermission 管理员

@apiParam {String} ID 项目ID。


@apiParamExample {get} 请求示例
GET /api/project/info?id=5

@apiSuccess {Integer} id 项目ID。
@apiSuccess {String} createdAt 创建时间。
@apiSuccess {String} name 项目名称，不能重复。
@apiSuccess {String} description 项目描述。
@apiSuccess {Object} [manager] 负责人信息。
@apiSuccess {Integer} [manager.id] 用户ID。
@apiSuccess {String} [manager.name] 姓名。
@apiSuccess {String} [Project.version] 版本号。

@apiSuccessExample 成功响应
HTTP/1.1 200 OK
{
    "id": 5,
    "createdAt": "2022-10-12 01:43:30",
    "name": "口令测试工具",
    "description": "",
    "manager": {"id": 13 , "name": "王伟" },
    "version": ""
}
@apiErrorExample 失败响应
HTTP/1.1 400 Bad Request

项目不存在或已经被删除
*/

// info 查询项目信息
func (c *ProjectController) info(ctx *gin.Context) {
	var reqInfo dto.ProjectInfoDto
	var project entity.Project
	var user entity.User
	id, _ := strconv.Atoi(ctx.Query("id"))
	if id <= 0 {
		ErrIllegal(ctx, "参数非法,无法解析")
		return
	}

	// 查询项目信息
	err := repo.DB.Where("id = ? AND is_delete = ?", id, 0).Find(&project).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	if project.ID == 0 {
		ErrIllegal(ctx, "项目不存在或已经被删除")
		return
	}

	// 查询负责人信息
	err = repo.DB.Select("name").Where("id", project.Manager).Find(&user).Error
	if err != nil {
		ErrSys(ctx, err)
		return
	}
	reqInfo.Transform(&project, &user)

	ctx.JSON(200, reqInfo)
}
