package models

type ServicePermissionModel struct {
	Model

	// User
	UserId string     `json:"user_id,omitempty" form:"user_id" gorm:"index:user_service_role,unique"`
	User   *UserModel `json:"user,omitempty" form:"-" validate:"-"`

	// Service
	ServiceId string        `json:"service_id,omitempty" form:"service_id" gorm:"index:user_service_role,unique"`
	Service   *ServiceModel `json:"service,omitempty" form:"-" validate:"-"`

	Role ServiceRole `json:"role" form:"role" gorm:"type:text" validate:"required"`
}

// TableName changement du nom de la table
func (ServicePermissionModel) TableName() string {
	return "services_permissions"
}
