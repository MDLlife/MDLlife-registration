package model

import (
	"time"
	"fmt"
	"errors"

	"../db"

	"github.com/lib/pq"
)

type WhitelistToken struct {
	WhitelistId int64
	Token       string    `xorm:"varchar(128) not null pk"`
	CreatedAt   time.Time `xorm:"created"`
	ExpiredAt   time.Time
	UsedAt      pq.NullTime
}

func (wt *WhitelistToken) TableName() string {
	return "whitelist_tokens"
}

// CRUD
func (wt *WhitelistToken) GetValidToken(token string) (has bool, err error) {
	return db.Engine.Where("token = ? AND used_at IS NULL AND expired_at > ?", token, time.Now()).Get(wt)
}

func (wt *WhitelistToken) TokenConfirmed() (w *Whitelist, err error) {
	tx := db.Engine.NewSession()
	defer tx.Close()

	if err = tx.Begin(); err != nil {
		return nil, err
	}

	wt.UsedAt = pq.NullTime{Time: time.Now(), Valid: true}
	if _, err = tx.ID(wt.Token).Cols("used_at").Update(wt); err != nil {
		return nil, err
	}

	w = &Whitelist{}
	has, err := tx.ID(wt.WhitelistId).Get(w)
	if !has {
		return nil, errors.New(fmt.Sprintf("Can't find whitelist with id: %v", wt.WhitelistId))
	}
	if err != nil {
		return nil, err
	}

	w.VerificationStage = EMAIL_VERIFIED
	if _, err = tx.ID(w.Id).Cols("verification_stage").Update(w); err != nil {
		return nil, err
	}

	return w, tx.Commit()
}