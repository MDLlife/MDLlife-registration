package admin

import (
	"github.com/kataras/iris"
	"strconv"
	"github.com/go-ozzo/ozzo-validation"
	"os"
	"fmt"
	"bufio"
	"encoding/base64"

	"../../model"
	"../../db"
	"regexp"
)

var (
	SortByRegex = regexp.MustCompile("^(id|name|country|birthday)$")
	StageFilterRegex = regexp.MustCompile("^(all|unconfirmed|confirmed|declined|question|accepted)$")
)

func GetWhitelistList(ctx iris.Context) {
	var whitelists []model.WhitelistPassport
	descending, _ := strconv.ParseBool(ctx.FormValue("descending"))
	page, _ := strconv.Atoi(ctx.FormValueDefault("page", "1"))
	rowsPerPage, _ := strconv.Atoi(ctx.FormValue("rowsPerPage"))
	sortBy := ctx.FormValueDefault("sortBy", "id")
	search := ctx.FormValue("search")
	stageFilter := ctx.FormValueDefault("stage", "all")

	query := db.Engine.NewSession()

	if err := validation.Validate(sortBy, validation.Match(SortByRegex)); err != nil {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.JSON(map[string]interface{}{"errors": err})
		return
	}

	if err := validation.Validate(stageFilter, validation.Match(StageFilterRegex)); err != nil {
		ctx.StatusCode(iris.StatusUnprocessableEntity)
		ctx.JSON(map[string]interface{}{"errors": err})
		return
	}

	query = query.Table("whitelists").Alias("w")
	if stageFilter == "all" {
		query = query.Where("w.verification_stage >= ?", int(model.STAGE_EMAIL_CONFIRMED))
	} else {
		query = query.Where("w.verification_stage = ?", int(model.NewVerificationStageFromString(stageFilter)))
	}

	if search != "" {
		query = query.And("w.name LIKE ?", "%" + search + "%")
	}

	rowsNumber, err := query.Clone().Count(&model.Whitelist{})
	if err != nil {
		println("Can't count whitelists. " + err.Error())
	}

	// move below because it breaks count
	query = query.Select("w.id, w.name, w.email, w.birthday, w.country, w.verification_stage, w.passport_id, p.id, p.path, p.extension")
	query = query.Join("INNER", []string{"photos", "p"}, "p.id = w.passport_id")
	if descending {
		query = query.Desc("w."+sortBy)
	} else {
		query = query.Asc("w."+sortBy)
	}
	if rowsPerPage > 0 {
		query = query.Limit(rowsPerPage, (page-1)*rowsPerPage)
	}

	if err := query.Find(&whitelists); err != nil {
		println("Can't receive whitelists. " + err.Error())
	}

	for i := 0; i < len(whitelists); i++ {
		imgFile, err := os.Open(whitelists[i].Passport.Path)

		if err != nil {
			fmt.Printf("Can't open photoId: %v \n\t %s", whitelists[i].Passport.Id, err.Error())
			continue
		}

		defer imgFile.Close()

		// create a new buffer base on file size
		fInfo, _ := imgFile.Stat()
		var size int64 = fInfo.Size()
		buf := make([]byte, size)

		// read file content into buffer
		fReader := bufio.NewReader(imgFile)
		fReader.Read(buf)

		// convert the buffer bytes to base64 string
		imgBase64Str := base64.StdEncoding.EncodeToString(buf)

		whitelists[i].Passport.Src = "data:image/" + whitelists[i].Passport.Extension + ";base64," + imgBase64Str
	}

	ctx.JSON(map[string]interface{}{"data": whitelists, "pagination": map[string]interface{}{
		"descending":  descending,
		"page":        page,
		"rowsPerPage": rowsPerPage,
		"rowsNumber":  rowsNumber,
		"sortBy":      sortBy,
	}})
}

func WhitelistAccept(ctx iris.Context) {
	id := ctx.Params().Get("id")

	whitelist := &model.Whitelist{VerificationStage: model.STAGE_ACCEPTED}
	i, err := db.Engine.ID(id).Cols("verification_stage").Update(whitelist)
	if err != nil || i == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		fmt.Printf("Can't accept whitelist id: %v \n\t %s", id, err)
		return
	}
}

func WhitelistDecline(ctx iris.Context) {
	id := ctx.Params().Get("id")

	whitelist := &model.Whitelist{VerificationStage: model.STAGE_DECLINED}
	i, err := db.Engine.
		ID(id).
		Where("verification_stage < ?", int(model.STAGE_ACCEPTED)).
		Cols("verification_stage").
		Update(whitelist)
	if err != nil || i == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		fmt.Printf("Can't decline whitelist id: %v \n\t %s", id, err)
		return
	}
}

func WhitelistQuestion(ctx iris.Context) {
	id := ctx.Params().Get("id")

	whitelist := &model.Whitelist{VerificationStage: model.STAGE_QUESTION}
	i, err := db.Engine.
		ID(id).
		Where("verification_stage < ?", int(model.STAGE_ACCEPTED)).
		Cols("verification_stage").
		Update(whitelist)
	if err != nil || i == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		fmt.Printf("Can't mark by a question whitelist id: %v \n\t %s", id, err)
		return
	}
}
