package repo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pdm/controller/middle"
	"pdm/repo/entity"
	"pdm/reuint/jwt"
)

// UserRepository 用户支持层
type UserRepository struct {
}

// Exist 检查用户是否已经存在
func (r *UserRepository) Exist(userId int) (bool, error) {
	if userId == 0 {
		return false, nil
	}
	res := &entity.User{}
	err := DB.First(res, "id = ? AND is_delete = 0 ", userId).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistOpenid 检查工号是否已经存在
func (r *UserRepository) ExistOpenid(openid string) (bool, error) {
	if openid == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DB.First(res, "openid = ? AND is_delete = 0 ", openid).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistPhone 检查手机号是否已经存在
func (r *UserRepository) ExistPhone(phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DB.First(res, "phone = ? AND is_delete = 0 ", phone).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistUsername 检查用户名是否已经存在
func (r *UserRepository) ExistUsername(username string) (bool, error) {
	if username == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DB.First(res, "username = ? AND is_delete = 0 ", username).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// ExistEmail 检查邮箱是否已经存在
func (r *UserRepository) ExistEmail(email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	res := &entity.User{}
	err := DB.First(res, "email = ? AND is_delete = 0 ", email).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

// IsValid 判断登录用户是否有效
func (r *UserRepository) IsValid(ctx *gin.Context) (bool, error) {
	claimsValue, _ := ctx.Get(middle.FlagClaims)
	claims := claimsValue.(*jwt.Claims)

	res := &entity.User{}
	err := DB.First(res, "id = ? AND is_delete = 0 ", claims.Sub).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return true, err
	}

	return false, nil
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}
