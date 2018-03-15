package model

import (
	"time"
	"mime/multipart"
	"path/filepath"
	"strings"
	"errors"
	"os"
	"io"

	"../utils"
	"../db"
)

// Photo is photo table structure.
type Photo struct {
	Id        int64
	Path      string    `xorm:"varchar(255) not null unique"`
	Extension string    `xorm:"varchar(5) not null"`
	CreatedAt time.Time `xorm:"created"`
}

func (p *Photo) TableName() string {
	return "photos"
}

// CRUD
func (p *Photo) StoreFile(file multipart.File, fileInfo *multipart.FileHeader) error {
	defer file.Close()
	var (
		path string
		ext string
		filename string
		imgPath  string
	)

	ext = filepath.Ext(fileInfo.Filename)
	ext = strings.ToLower(ext[1:]) // remove dot and cast to lower case

	// generate a new name if file exists
	for {
		filename = utils.RandomString(48)
		imgPath = filename[0:3] + "/" + filename[3:6]

		// create path / bug with 0644
		if err := os.MkdirAll("./uploads/"+imgPath, 0744); err != nil {
			return errors.New("Can't create image path: " + err.Error())
		}

		path = "./uploads/" + imgPath + "/" + filename + "." + ext
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	// Create a file
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0744)

	if err != nil {
		return errors.New("Can't save file: " + err.Error())
	}
	defer out.Close()

	io.Copy(out, file)

	p.Path = path
	p.Extension = ext

	if _, err := db.Engine.InsertOne(p); err != nil {
		return errors.New("Can't insert photo into database: " + err.Error())
	}

	return nil
}