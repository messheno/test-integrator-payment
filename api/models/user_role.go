package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Definition du role utilisateur
type UserRole int

const (
	USER_MERCHANT UserRole = iota
	USER_MANAGER
	USER_ADMIN
)

// String - Créer un comportement commun - attribuer au type une fonction String
func (u UserRole) String() string {
	return [...]string{"USER_MERCHANT", "USER_MANAGER", "USER_ADMIN"}[u]
}

// EnumIndex - Créer un comportement commun - donner au type une fonction EnumIndex
func (u UserRole) EnumIndex() int {
	return int(u)
}

func (UserRole) GormDataType() string {
	return "text"
}

func (UserRole) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// returns different database type based on driver name
	return "text"
}

func (u *UserRole) Scan(src interface{}) error {
	if src == nil {
		*u = USER_MERCHANT
		return nil
	}

	var bytes []byte
	switch v := src.(type) {
	case []byte:
		if len(v) > 0 {
			bytes = make([]byte, len(v))
			copy(bytes, v)
		}
	case int:
		bytes = []byte(fmt.Sprintf("%v", v))
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("échec de l'analyse du champ:", src))
	}

	switch string(bytes) {
	case USER_ADMIN.String():
		*u = USER_ADMIN

	case USER_MANAGER.String():
		*u = USER_MANAGER

	default:
		*u = USER_MERCHANT
	}

	return nil
}

func (u UserRole) Value() (driver.Value, error) {
	return u.String(), nil
}
