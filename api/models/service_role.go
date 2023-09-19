package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Definition du role utilisateur
type ServiceRole int

const (
	SERVICE_DEV ServiceRole = iota
	SERVICE_MANAGER
	SERVICE_ADMIN
)

// String - Créer un comportement commun - attribuer au type une fonction String
func (s ServiceRole) String() string {
	return [...]string{"SERVICE_DEV", "SERVICE_MANAGER", "SERVICE_ADMIN"}[s]
}

// EnumIndex - Créer un comportement commun - donner au type une fonction EnumIndex
func (s ServiceRole) EnumIndex() int {
	return int(s)
}

func (ServiceRole) GormDataType() string {
	return "text"
}

func (ServiceRole) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

func (s *ServiceRole) Scan(src interface{}) error {
	if src == nil {
		*s = SERVICE_DEV
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
	case SERVICE_ADMIN.String():
		*s = SERVICE_ADMIN

	case SERVICE_MANAGER.String():
		*s = SERVICE_MANAGER

	default:
		*s = SERVICE_DEV
	}

	return nil
}

func (s ServiceRole) Value() (driver.Value, error) {
	return s.String(), nil
}
