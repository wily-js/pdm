package repo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/middle"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

// ProjectRepository 项目支持层
type ProjectRepository struct {
}

// Exist 判断项目是否已经存在
func (r *ProjectRepository) Exist(id int) (bool, error) {
	if id == 0 {
		return false, nil
	}
	res := &entity.Project{}
	err := DB.First(res, "id = ? AND is_delete = ?", id, 0).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// NameExist 判断项目名是否已经存在
func (r *ProjectRepository) NameExist(name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	res := &entity.Project{}
	err := DB.First(res, "name = ? AND is_delete = ?", name, 0).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

func (r *ProjectRepository) GetProjectName(ctx *gin.Context) (string, error) {
	res := entity.Project{}
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	err := DB.First(&res, "id = ? AND is_delete = 0", claims.PID).Error
	if err == gorm.ErrRecordNotFound {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return res.Name, nil
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}
