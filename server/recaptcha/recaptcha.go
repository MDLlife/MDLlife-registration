package recaptcha

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"time"
	"github.com/kataras/iris/core/netutil"
	"github.com/kataras/iris/context"
)

const (
	ResponseFormValue = "g-recaptcha-response"
	apiURL            = "https://www.google.com/recaptcha/api/siteverify"
)

var secret string

// Response is the google's recaptcha response as JSON.
type Response struct {
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Success     bool      `json:"success"`
}

// Client is the default `net/http#Client` instance which
// is used to send requests to the Google API.
//
// Change Client only if you know what you're doing.
var Client = netutil.Client(time.Duration(20 * time.Second))

// Check accepts  context and returns the google's recaptcha response,
// if `response.Success` is true then validation passed.
func Check(ctx context.Context) (response Response) {
	generatedResponseID := ctx.FormValue(ResponseFormValue)
	if generatedResponseID == "" {
		response.ErrorCodes = append(response.ErrorCodes,
			"form value[g-recaptcha-response] is empty")
		return
	}

	r, err := Client.PostForm(apiURL,
		url.Values{
			"secret":   {secret},
			"response": {generatedResponseID},
			"remoteip": {ctx.RemoteAddr()},
		},
	)

	if err != nil {
		response.ErrorCodes = append(response.ErrorCodes, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		response.ErrorCodes = append(response.ErrorCodes, err.Error())
		return
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		response.ErrorCodes = append(response.ErrorCodes, err.Error())
		return
	}

	return response
}

// Confirm returns Success reCaptcha or not
func Confirm(ctx context.Context) bool {
	return Check(ctx).Success
}

// Init allows the webserver or code evaluating the reCaptcha form input to set the
// reCaptcha private key (string) value, which will be different for every domain.
func Init(key string) {
	secret = key
}
