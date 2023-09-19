package models

type ShopPermissionModel struct {
	Model

	// User
	UserId string     `json:"user_id,omitempty" form:"user_id" gorm:"index:user_shop_role,unique"`
	User   *UserModel `json:"user,omitempty" form:"-" validate:"-"`

	// Shop
	ShopId string     `json:"shop_id,omitempty" form:"shop_id" gorm:"index:user_shop_role,unique"`
	Shop   *ShopModel `json:"shop,omitempty" form:"-" validate:"-"`

	Role ShopRole `json:"role" form:"role" gorm:"type:text" validate:"required"`
}
