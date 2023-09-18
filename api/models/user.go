package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Declare related constants for each user role starting with index 1
type UserRole int

const (
	USER_STANDARD UserRole = iota
	USER_MANAGER
	USER_ADMIN
)

// String - Creating common behavior - give the type a String function
func (d UserRole) String() string {
	return [...]string{"USER_STANDARD", "USER_MANAGER", "USER_ADMIN"}[d]
}

// jwtCustomClaims are custom claims extending default ones.
type JwtCustomClaims struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`

	// Role
	IsAdmin    bool `json:"is_admin"`
	IsManager  bool `json:"is_manager"`
	IsStandard bool `json:"is_standard"`

	jwt.StandardClaims
}

type (
	UserRoles []UserRole

	UserModel struct {
		Model

		// Information
		FirstName string `json:"first_name" form:"first_name" validate:"required"`
		LastName  string `json:"last_name" form:"last_name" validate:"required"`
		Birthday  string `json:"birthday" form:"birthday" validate:"omitempty"`

		// Phone
		PhoneIndicator string `json:"phone_indicator" form:"phone_indicator"`
		PhoneNumber    string `json:"phone_number" form:"phone_number"`
		Email          string `json:"email" form:"email" validate:"omitempty,email" gorm:"unique"`

		Role     UserRole `json:"role" form:"role" gorm:"type:text"`
		Password string   `json:"-" form:"password" validate:"omitempty"`
	}
)

func (UserModel) TableName() string {
	return "users"
}

func (user *UserModel) BeforeCreate(tx *gorm.DB) (err error) {
	user.Model.BeforeCreate(tx)

	// Traitement des roles
	password := user.Password

	ciphertext, err := encrypt([]byte(user.Password), PASS_SECRET)
	if err != nil {
		return
	}

	user.Password = fmt.Sprintf("%0x", ciphertext)
	ok, err := user.CheckPass(password)
	if err != nil {
		return
	}

	if !ok {
		err = fmt.Errorf("password not valide")
		return
	}

	return
}

func (u *UserModel) BeforeDelete(tx *gorm.DB) (err error) {
	if u.isGrant(USER_ADMIN) {
		return errors.New("admin user not allowed to delete")
	}

	return
}

func (u *UserModel) CheckPass(password string) (bool, error) {
	if len(u.Password) <= 0 || len(password) <= 0 {
		return false, fmt.Errorf("error login")
	}

	src := []byte(u.Password)

	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(dst, src)
	if err != nil {
		return false, err
	}

	// Crypate du password
	ciphertext, err := decrypt(dst, PASS_SECRET)
	if err != nil {
		return false, err
	}

	return string(ciphertext) == password, nil
}

func (u *UserModel) GenerateJWT() (string, error) {
	// Set custom claims
	claims := &JwtCustomClaims{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.PhoneNumber,
		Email:     u.Email,

		// KMT Role
		IsAdmin:    u.isGrant(USER_ADMIN),
		IsManager:  u.isGrant(USER_MANAGER),
		IsStandard: u.isGrant(USER_STANDARD),

		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString([]byte(JWT_SECRET))
}

func (u *UserModel) isGrant(role UserRole) bool {
	if u.Role == USER_ADMIN || u.Role == role {
		return true
	}

	return false
}

// / https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) ([]byte, error) {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decrypt(data []byte, passphrase string) ([]byte, error) {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
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
		*u = USER_STANDARD
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
		*u = USER_STANDARD
	}

	return nil
}

func (u UserRole) Value() (driver.Value, error) {
	return u.String(), nil
}
