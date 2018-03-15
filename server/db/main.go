package db

import (
	"github.com/go-xorm/xorm"
	"github.com/go-xorm/core"
	_ "github.com/mattn/go-sqlite3"

	"../config"
)

var Engine *xorm.Engine

func Init() (db *xorm.Engine, err error) {
	db, err = xorm.NewEngine(config.Config.DatabaseDriver, config.Config.DatabaseDSN)

	if err != nil {
		return nil, err
	}

	if config.Config.Debug {
		db.ShowSQL(true) // Show SQL statement on standard output;
		db.Logger().SetLevel(core.LOG_DEBUG)
	}
	//db.SetMaxOpenConns(60)
	//db.SetMaxIdleConns(5)

	Engine = db

	return Engine, nil
}
