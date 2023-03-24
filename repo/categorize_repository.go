package repo

import (
	"gorm.io/gorm"
	"pdm/repo/entity"
)

// CategorizeRepository 接口分类支持层
type CategorizeRepository struct {
}

func NewCategorizeRepository() *CategorizeRepository {
	return &CategorizeRepository{}
}

// ExistName 检查接口分类名称在同级是否存在
func (r *CategorizeRepository) ExistName(name string, parentId int) (bool, error) {
	if name == "" {
		return false, nil
	}
	res := &entity.ApiCategorize{}
	err := DB.First(res, "name = ? AND parent_id = ?", name, parentId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}
