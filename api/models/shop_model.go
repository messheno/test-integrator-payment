package models

import (
	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type ShopModel struct {
	Model

	Amount        float64 `json:"amount" form:"amount" validate:"-"`
	CurrentAmount float64 `json:"current_amount" form:"current_amount" validate:"-"`

	Name        string `json:"name" form:"name" validate:"required" gorm:"index"`
	NameSlug    string `json:"name_slug" gorm:"unique" form:"-" validate:"-"`
	Description string `json:"description" form:"description" validate:"required" gorm:"index"`
	SiteWeb     string `json:"site_web,omitempty" form:"site_web" validate:"omitempty" gorm:"index"`
	Logo        string `json:"logo" form:"logo" validate:"omitempty"`
	Country     string `json:"country,omitempty" form:"country" gorm:"index"`

	ClientId  string `json:"-" form:"client_id" validate:"required"`
	ClientKey string `json:"-" form:"client_key" validate:"required"`

	// Liste de permission des utilisateurs
	Permissions []ShopPermissionModel `json:"permissions,omitempty" form:"permissions" gorm:"foreignKey:ShopId;constraint:OnDelete:CASCADE;"`

	// Liste des transaction
	Transactions []TransactionModel `json:"transactions,omitempty" form:"transactions" gorm:"foreignKey:ShopId;constraint:OnDelete:CASCADE;"`
}

func (s *ShopModel) BeforeCreate(tx *gorm.DB) (err error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	s.ID = uuid.String()
	s.NameSlug = slug.Make(s.Name)

	// Génération des clés
	s.ClientId = RandStr(32)
	s.ClientKey = RandStr(64)

	return
}

func (s *ShopModel) BeforeSave(tx *gorm.DB) (err error) {
	s.NameSlug = slug.Make(s.Name)

	return
}

func (s *ShopModel) GenerateClient() {
	s.ClientId = RandStr(32)
	s.ClientKey = RandStr(64)
}
