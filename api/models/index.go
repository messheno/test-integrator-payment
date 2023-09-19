package models

import (
	"math/rand"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

const ConnectTimeout = time.Second * 10
const JWT_SECRET = "~bFlgzsG7<:IHV:.yIo[BH(^<3yw*D"
const PASS_SECRET = "+7tkvy*dMFJ]8~(trf+|l'/T^:'?/k"

type Model struct {
	ID        string    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Model) BeforeCreate(tx *gorm.DB) error {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	base.ID = uuid.String()

	return nil
}

func RandStr(n int) string {
	var charset = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-@&")

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
