package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

const CONNECTION_TIMEOUT = time.Second * 10
const JWT_SECRET = "~bFlgzsG7<:IHV:.yIo[BH(^<3yw*D"
const PASS_SECRET = "+7tkvy*dMFJ]8~(trf+|l'/T^:'?/k"

// Declare related constants for each user role starting with index 1
type DBType int

const (
	DB_TYPE_POSTGRESQL DBType = iota + 1
	DB_TYPE_MYSQL
	DB_TYPE_SQLSERVER
	DB_TYPE_SQLITE
)

// String - Creating common behavior - give the type a String function
func (d DBType) String() string {
	return [...]string{"POSTGRESQL", "MYSQL", "SQLSERVER", "SQLITE"}[d-1]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex functio
func (d DBType) EnumIndex() int {
	return int(d)
}

func CreateUpdateTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserModel{},
	)
}

func ResetTable(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&UserModel{},
	)
}

type Model struct {
	ID        string    `json:"id,omitempty" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if len(base.ID) <= 0 {
		uuid, err := uuid.NewV4()
		if err != nil {
			return err
		}

		base.ID = uuid.String()
	}

	return
}

func StringToDate(dateString string, layouts []string) (bool, time.Time) {
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateString)
		if err == nil {
			return true, t
		}
	}

	return false, time.Time{}
}

var LAYOUTS_TIME = []string{
	// yyyy-mm-dd
	"2006-01-02",
	"06-01-02",
	"2006/01/02",
	"2006 01 02",

	// mm-dd-yyyy
	"01-02-2006",
	"01-02-06",
	"01/02/06",

	// dd-mm-yyyy
	"02-01-2006",
	"02-01-06",
	"02/01/2006",
	"02 01 2006",
}

func arrayUnique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}
