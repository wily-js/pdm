package repo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/middle"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

type ProjectMemberRepository struct {
}

func NewProjectMemberRepository() *ProjectMemberRepository {
	return &ProjectMemberRepository{}
}

// Exist 检查项目成员在该项目是否已经存在
func (r *ProjectMemberRepository) Exist(projectId int, userId int) (bool, error) {
	res := &entity.ProjectMember{}
	err := DB.Where("project_id", projectId).Where("user_id", userId).Find(res).Error

	if res.ID == 0 {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// HeadRole 检查登录用户是否为项目负责人
func (r *ProjectMemberRepository) HeadRole(ctx *gin.Context) (bool, error) {

	// 获取当前用户信息
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	projectId := claims.PID

	if projectId <= 0 || claims.Sub <= 0 {
		return false, nil
	}
	res := &entity.ProjectMember{}
	err := DB.First(res, "project_id = ? AND user_id = ? AND role = ?", projectId, claims.Sub, 4).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// IsProjectManager 是否是项目管理员
func (r *ProjectMemberRepository) IsProjectManager(projectId, userId int) (bool, error) {
	var role int
	err := DB.Model(&entity.ProjectMember{}).Select("role").
		First(&role, "project_id = ? AND user_id = ?", projectId, userId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return entity.RoleCreator == role, nil
}
