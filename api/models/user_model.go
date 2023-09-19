package models

import (
	"errors"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	Model

	// Identifiant unique keycloak
	AuthId string `json:"auth_id" form:"auth_id" validate:"required" gorm:"index,unique"`

	// Information
	FirstName   string `json:"first_name" form:"first_name" validate:"required" gorm:"index"`
	LastName    string `json:"last_name" form:"last_name" validate:"required" gorm:"index"`
	PhonePrefix string `json:"phone_prefix" form:"phone_prefix" gorm:"index"`
	PhoneNumber string `json:"phone_number" form:"phone_number" gorm:"index"`
	Email       string `json:"email" form:"email" validate:"omitempty,email" gorm:"index"`
	Country     string `json:"country" form:"country" gorm:"index"`

	// Role
	Role            UserRole              `json:"role" form:"role" gorm:"type:text" validate:"-"`
	ShopPermissions []ShopPermissionModel `json:"shop_permissions" form:"shop_permissions" gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE;"`
}

func (u *UserModel) BeforeCreate(tx *gorm.DB) (err error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	u.ID = uuid.String()
	return
}

func (u *UserModel) BeforeDelete(tx *gorm.DB) (err error) {
	if u.IsGrant(USER_ADMIN) {
		return errors.New("admin user not allowed to delete")
	}

	return
}

func (u *UserModel) IsGrant(role UserRole) bool {
	if u.Role == USER_ADMIN || u.Role == role {
		return true
	}

	if role == USER_MERCHANT && u.Role == USER_MANAGER {
		return true
	}

	return false
}

func (u *UserModel) IsShopGrant(shopId string, role ShopRole) bool {
	if u.Role == USER_ADMIN {
		return true
	}

	for _, perm := range u.ShopPermissions {
		if shopId == perm.ShopId {
			if perm.Role == SHOP_ADMIN || perm.Role == role {
				return true
			}

			if role == SHOP_MANAGER && perm.Role == SHOP_DEV {
				return true
			}
		}
	}

	return false
}
