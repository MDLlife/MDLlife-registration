package model

import (
	"database/sql/driver"
	"database/sql"
	"time"

	"./validation_rules"
	"../utils"
	"../db"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// VerificationStage Type enumeration
type VerificationStage uint8

const (
	EMAIL_NOT_VERIFIED VerificationStage = iota
	EMAIL_VERIFIED
	FIRST_STAGE
)

func (u *VerificationStage) Scan(value interface{}) error { *u = VerificationStage(value.(uint8)); return nil }
func (u VerificationStage) Value() (driver.Value, error)  { return uint8(u), nil }

// Whitelist is whitelist table structure.
type Whitelist struct {
	Id                int64
	PassportId        int64             `xorm:"not null unique"`
	SelfieId          sql.NullInt64
	Name              string            `xorm:"varchar(255) not null"`
	Email             string            `xorm:"varchar(255) not null unique"`
	Phone             string            `xorm:"varchar(255) not null"`
	Birthday          string            `xorm:"varchar(255) not null"`
	Country           string            `xorm:"varchar(255) not null"`
	VerificationStage VerificationStage `xorm:"not null default 0"`
	CreatedAt         time.Time         `xorm:"created"`
	UpdatedAt         time.Time         `xorm:"updated"`
}

func (w *Whitelist) TableName() string {
	return "whitelists"
}

// validation
func (w Whitelist) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Name, validation.Required, validation.Match(validation_rules.NameRegex)),
		validation.Field(&w.Email, validation.Required, is.Email),
		validation.Field(&w.Phone, validation.Match(validation_rules.PhoneNumberRegex)),
		validation.Field(&w.Birthday, validation.Required, validation.Match(validation_rules.DateRegex)),
		validation.Field(&w.Country, validation.Required, validation.Match(validation_rules.NameRegex)),
	)
}

// CRUD
func (w *Whitelist) StoreData() (emailToken string, err error) {
	tx := db.Engine.NewSession()
	defer tx.Close()

	if err = tx.Begin(); err != nil {
		return "", err
	}

	if _, err = tx.InsertOne(w); err != nil {
		return "", err
	}

	var token string
	// regenerate if not unique
	for {
		token = utils.SecureRandomString(35)
		has, err := tx.Where("whitelist_id = ? AND token = ?", w.Id, token).Exist(&WhitelistToken{})
		if err != nil {
			return "", err
		}
		if !has {
			break
		}
	}

	wt := &WhitelistToken{
		WhitelistId: w.Id,
		Token:       token,
		ExpiredAt:   time.Now().AddDate(0, 0, 7),
	}

	if _, err = tx.InsertOne(wt); err != nil {
		return "", err
	}

	return token, tx.Commit()
}

func (w *Whitelist) EmailExist() (has bool, err error) {
	return db.Engine.Where("email = ?", w.Email).Exist(&Whitelist{})
}