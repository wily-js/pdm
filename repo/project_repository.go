package repo

import (
	"gorm.io/gorm"
	"pdm/repo/entity"
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

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}
