package repo

import (
	"gorm.io/gorm"
	"pdm/repo/entity"
)

// CaseRepository 接口分类支持层
type CaseRepository struct {
}

func NewCaseRepository() *CaseRepository {
	return &CaseRepository{}
}

// ExistName 检查接口用例名称在同级是否存在
func (r *CaseRepository) ExistName(name string, categorizeId int) (bool, error) {
	if name == "" {
		return false, nil
	}
	res := &entity.ApiCase{}
	err := DB.First(res, "name = ? AND categorize_id = ?", name, categorizeId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}
