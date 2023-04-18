package repo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/middle"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

// CategorizeRepository 接口分类支持层
type CategorizeRepository struct {
}

func NewCategorizeRepository() *CategorizeRepository {
	return &CategorizeRepository{}
}

// ExistName 检查接口分类名称在同级是否存在
func (r *CategorizeRepository) ExistName(ctx *gin.Context, name string, parentId int) (bool, error) {
	if name == "" {
		return false, nil
	}
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)
	res := &entity.ApiCategorize{}
	err := DB.First(res, "name = ? AND parent_id = ? AND project_id = ?", name, parentId, claims.PID).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}
