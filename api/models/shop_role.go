package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Definition du role utilisateur
type ShopRole int

const (
	SHOP_DEV ShopRole = iota
	SHOP_MANAGER
	SHOP_ADMIN
)

// String - Créer un comportement commun - attribuer au type une fonction String
func (s ShopRole) String() string {
	return [...]string{"SHOP_DEV", "SHOP_MANAGER", "SHOP_ADMIN"}[s]
}

// EnumIndex - Créer un comportement commun - donner au type une fonction EnumIndex
func (s ShopRole) EnumIndex() int {
	return int(s)
}

func (ShopRole) GormDataType() string {
	return "text"
}

func (ShopRole) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

func (s *ShopRole) Scan(src interface{}) error {
	if src == nil {
		*s = SHOP_DEV
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
	case SHOP_ADMIN.String():
		*s = SHOP_ADMIN

	case SHOP_MANAGER.String():
		*s = SHOP_MANAGER

	default:
		*s = SHOP_DEV
	}

	return nil
}

func (s ShopRole) Value() (driver.Value, error) {
	return s.String(), nil
}
