package main

import (
	"os"
	"io"
	"io/ioutil"
	"time"
	"regexp"
	"path/filepath"
	"strings"
	"reflect"
	"mime/multipart"
	"database/sql"

	"gopkg.in/yaml.v2"

	"github.com/kataras/iris"

	"./ses"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"database/sql/driver"
	"github.com/lib/pq"
)

/*
	go get ./...
*/

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
	Id                     int64             `gorm:"primary_key"`
	Passport               Photo
	PassportId             int64             `gorm:"not null; unique"`
	Selfie                 Photo
	SelfieId               sql.NullInt64
	Name                   string
	Email                  string            `gorm:"not null; unique"`
	Phone                  string
	Birthday               string
	Country                string
	VerificationStage      VerificationStage `gorm:"not null; default:0"`
	EmailVerificationToken string
	CreatedAt              time.Time
}

// Photo is photo table structure.
type Photo struct {
	Id        int64
	Path      string `gorm:"not null; unique"`
	Extension string `gorm:"varchar(5); not null"`
	CreatedAt time.Time
}

type WhitelistToken struct {
	Whitelist   Whitelist
	WhitelistId int64
	Token       string `gorm:"not null; unique"`
	CreatedAt   time.Time
	ExpiredAt   time.Time
	UsedAt      pq.NullTime
}

// Config file structure
type Config struct {
	AwsKey    string `yaml:"AwsKey"`
	AwsSecret string `yaml:"AwsSecret"`
	AwsRegion string `yaml:"AwsRegion"`

	NoReplyEmail string `yaml:"NoReplyEmail"`
	ReplyEmail   string `yaml:"ReplyEmail"`

	DatabaseDriver string `yaml:"DatabaseDriver"`
	DatabaseDSN    string `yaml:"DatabaseDSN"`

	MaxFileUploadSizeMb int64 `yaml:"MaxFileUploadSizeMb"`

	Port string `yaml:"Port"`
}

var (
	config = &Config{}

	nameValidatorRegex = regexp.MustCompile("(?:(\\pL|[-])+((?:\\s)+)?)")
	// YYYY-MM-DD // YYYY >= 1000 matches correct dates in months
	dateValidatorRegex = regexp.MustCompile("^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-9])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31))$")
	// phone number, simple
	phoneNumberValidatorRegex = regexp.MustCompile("[0-9()\\pL\\s-+#]+")

	tokenRegex = regexp.MustCompile("[0-9a-zA-Z]+")
)

func init() {
	loadConfig()

	// Amazon SES setup
	ses.SetConfiguration(config.AwsKey, config.AwsSecret, config.AwsRegion)
}

func main() {
	app := iris.New()

	db, err := gorm.Open(config.DatabaseDriver, config.DatabaseDSN)
	if err != nil {
		app.Logger().Fatalf("db failed to initialized: %v", err)
	}

	iris.RegisterOnInterrupt(func() {
		db.Close()
	})

	db.AutoMigrate(&Whitelist{}, &Photo{}, &WhitelistToken{})

	routes(app, db)

	app.Run(iris.Addr(config.Port),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithPostMaxMemory(config.MaxFileUploadSizeMb<<20))
}

func routes(app *iris.Application, db *gorm.DB) {

	app.Get("/whitelist/confirm_email", func(ctx iris.Context) {
		token := ctx.Params().GetTrim("token")

		if !tokenRegex.MatchString(token) {
			ctx.HTML("Invalid token data")
			return
		}

		// todo check errors
		whitelistToken := &WhitelistToken{}
		if db.Where("token = ? AND used_at IS NULL", token).First(whitelistToken).Error != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Token db search error")
			return
		}

		whitelistToken.UsedAt = pq.NullTime{Time: time.Now(), Valid: true}
		db.Update(whitelistToken)

		whitelist := &Whitelist{}
		db.First(whitelist, whitelistToken.WhitelistId)

		whitelist.VerificationStage = EMAIL_VERIFIED
		db.Update(whitelist)

		ctx.HTML("Your email " + whitelist.Email + " has confirmed")
	})

	app.Post("/whitelist/request", func(ctx iris.Context) {
		birthday := ctx.FormValue("birthday")
		if birthday == "" {
			birthday = combineDatetime(ctx.FormValue("year"), ctx.FormValue("month"), ctx.FormValue("day"))
		}

		whitelist := &Whitelist{
			Name:     ctx.FormValue("name"),
			Email:    ctx.FormValue("email"),
			Country:  ctx.FormValue("country"),
			Birthday: birthday,
		}

		// Get the file from the request.
		passportFile, passportInfo, passportErr := ctx.FormFile("passport")

		if err := whitelist.Validate(); err != nil {
			ctx.StatusCode(iris.StatusUnprocessableEntity)
			errVal := reflect.ValueOf(err)
			if passportErr != nil && errVal.Kind() == reflect.Map {
				errVal.SetMapIndex(reflect.ValueOf("passport"), reflect.ValueOf("Add image of your passport"))
			}
			ctx.JSON(map[string]interface{}{"errors": err})
			return
		}

		if !db.Where("email = ?", whitelist.Email).First(&Whitelist{}).RecordNotFound() {
			ctx.StatusCode(iris.StatusUnprocessableEntity)
			ctx.JSON(map[string]interface{}{"errors": map[string]string{"email": "This email already registered."}})
			return
		}

		filePath, fileExt, err := saveFile(passportFile, passportInfo)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't save passport!\n\t" + err.Error())
			return
		}

		photo := &Photo{
			Path:      filePath,
			Extension: fileExt,
		}

		db.NewRecord(photo)
		if err := db.Create(photo).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't insert passport in database " + err.Error())
			return
		}

		whitelist.PassportId = photo.Id

		// Get selfie file from the request.
		selfieFile, selfieInfo, selfieErr := ctx.FormFile("selfie")
		if selfieErr == nil {
			filePath, fileExt, err := saveFile(selfieFile, selfieInfo)
			if err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				println("Can't save selfie!\n\t" + err.Error())
				return
			}

			selfie := &Photo{
				Path:      filePath,
				Extension: fileExt,
			}

			db.NewRecord(selfie)
			if err := db.Create(selfie).Error; err != nil {
				ctx.StatusCode(iris.StatusInternalServerError)
				println("Can't insert selfie in database " + err.Error())
				return
			}

			whitelist.SelfieId = sql.NullInt64{Int64: selfie.Id, Valid: true}
		}

		token, err := whitelist.StoreData(db)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			println("Can't insert whitelist " + err.Error())
			return
		}

		sendEmail(whitelist.Email, token)

		ctx.JSON(map[string]bool{"success": true})
	})
}

func sendEmail(to string, token string) {
	emailData := ses.Email{
		To:   to,
		From: config.NoReplyEmail,
		Text: "Your whitelist submission is well received.\n\n" +
			"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.\n\n" +
			"To confirm your email please paste this line in your browser url address line:\n" +
			"http://mdl.life/whitelist/confirm_email?token=" + token + "\n\n" +
			"For inquiries and support please contact support@mdl.life",
		HTML: "<h3 style=\"color:purple;\">Your whitelist submission is well received.</h3><br>" +
			"The instructions of how to purchase the MDL Tokens to be send soon is confirmation that you have passed the whitelist.<br><br>" +
			"To confirm your email please follow this link:<br>" +
			"<a href=\"http://mdl.life/whitelist/confirm_email?token=" + token + "\">" + "http://mdl.life/whitelist/confirm_email?token=" + token + "</a><br><br>" +
			"For inquiries and support please contact <a mailto=\"support@mdl.life\">support@mdl.life</a>",
		Subject: "MDL Talent Hub: Whitelist application received",
		ReplyTo: config.ReplyEmail,
	}

	println(ses.SendEmail(emailData))
}

func combineDatetime(y string, m string, d string) string {
	str := y + "-"

	if len(m) == 1 {
		str += "0"
	}

	str += m + "-"

	if len(d) == 1 {
		str += "0"
	}

	str += d

	return str
}

func loadConfig() {
	// which will try to find the 'filename' from current working dir too.
	yamlAbsPath, err := filepath.Abs("config.yml")
	if err != nil {
		println("Can't find example.config.yml " + err.Error())
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		println("Can't read example.config.yml " + err.Error())
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func (w Whitelist) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Name, validation.Required, validation.Match(nameValidatorRegex)),
		validation.Field(&w.Email, validation.Required, is.Email),
		validation.Field(&w.Phone, validation.Match(phoneNumberValidatorRegex)),
		validation.Field(&w.Birthday, validation.Required, validation.Match(dateValidatorRegex)),
		validation.Field(&w.Country, validation.Required, validation.Match(nameValidatorRegex)),
	)
}

func saveFile(file multipart.File, fileInfo *multipart.FileHeader) (path string, ext string, err error) {
	defer file.Close()
	ext = filepath.Ext(fileInfo.Filename)
	ext = strings.ToLower(ext[1:]) // remove dot and cast to lower case
	var (
		filename string
		imgPath  string
	)

	// generate a new name if file exists
	for {
		filename = RandomString(48)
		imgPath = filename[0:3] + "/" + filename[3:6]

		// create path / bug with 0644
		if err := os.MkdirAll("./uploads/"+imgPath, 0744); err != nil {
			return "", "", NewError("Can't create image path: " + err.Error())
		}

		path = "./uploads/" + imgPath + "/" + filename + "." + ext
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
	}

	// Create a file
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0744)

	if err != nil {
		return "", "", NewError("Can't save image: " + err.Error())
	}
	defer out.Close()

	io.Copy(out, file)

	return path, ext, nil
}

func (w *Whitelist) StoreData(db *gorm.DB) (emailToken string, err error) {
	tx := db.Begin()
	if err = tx.Error; err != nil {
		return "", err
	}

	db.NewRecord(w)
	if err = db.Create(w).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	var token string
	// regenerate if not unique
	for {
		token = SecureRandomString(35)
		if db.Where("whitelist_id = ? AND token = ?", w.Id, token).First(&WhitelistToken{}).RecordNotFound() {
			break
		}
	}

	wt := &WhitelistToken{
		Token:     token,
		ExpiredAt: time.Now().AddDate(0, 0, 7),
	}

	db.NewRecord(wt)
	if err = db.Create(wt).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	if err = tx.Commit().Error; err != nil {
		return "", err
	}

	return token, nil
}
