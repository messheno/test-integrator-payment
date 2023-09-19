package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type StringArray []string

func (StringArray) GormDataType() string {
	return "text"
}

func (StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

func (s *StringArray) Scan(src interface{}) error {
	if src == nil {
		*s = []string{}
		return nil
	}

	var bytes []byte
	switch v := src.(type) {
	case []byte:
		if len(v) > 0 {
			bytes = make([]byte, len(v))
			copy(bytes, v)
		}
	case string:
		bytes = []byte(v)

	case []string:
		*s = v
		return nil
	default:
		return errors.New(fmt.Sprint("échec de l'analyse du champ:", src))
	}

	val := string(bytes)
	*s = strings.Split(val, ",")

	return nil
}

func (s StringArray) Value() (driver.Value, error) {
	strings.Join(s, ",")
	return strings.Join(s, ","), nil
}

type ProviderModel struct {
	Model

	Name           string `json:"name" form:"name" validate:"required" gorm:"index"`
	NameSlug       string `json:"name_slug" gorm:"unique" form:"-" validate:"-"`
	Description    string `json:"description" form:"description" validate:"required" gorm:"index"`
	AsynchroneMode bool   `json:"asynchrone_mode" form:"asynchrone_mode" validate:"-"`

	PayUrl      string `json:"pay_url" form:"pay_url" validate:"required"`
	PayCheckUrl string `json:"pay_check_url" form:"pay_check_url" validate:"required"`
	HealthUrl   string `json:"health_url" form:"health_url" validate:"required"`

	SupportCountry StringArray `json:"support_country" form:"support_country" gorm:"type:text[]" validate:"required"` // CIV

	Transactions []TransactionModel `json:"transactions,omitempty" form:"transactions" gorm:"foreignKey:ProviderId"`
}

// TableName changement du nom de la table
func (ProviderModel) TableName() string {
	return "providers"
}
