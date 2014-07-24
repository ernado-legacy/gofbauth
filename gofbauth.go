package gofbauth

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	host                  = "www.facebook.com"
	scheme                = "https"
	accessTokenAction     = "oauth/access_token"
	fbGraphHost           = "graph.facebook.com"
	appIDParameter        = "client_id"
	accessTokenParameter  = "access_token"
	expiresParameter      = "expires"
	appSecretParameter    = "client_secret"
	version               = "5.23"
	fieldsParameter       = "fields"
	fiealdsValue          = "id,name,birthday,gender,picture.type(large),email"
	versionParameter      = "v"
	responseTypeParameter = "response_type"
	responseTypeCode      = "code"
	redirectParameter     = "redirect_uri"
	scopeParameter        = "scope"
	authAction            = "dialog/oauth"
	codeParameter         = "code"
	usersGetAction        = "me"
)

var (
	// ErrorBadCode occures when server returns blank code or error
	ErrorBadCode = errors.New("bad code")
	// ErrorBadResponse occures when server returns unexpected response
	ErrorBadResponse                = errors.New("bad server response")
	httpClient       mockHTTPClient = &http.Client{}
)

// Client for vkontakte oauth
type Client struct {
	ID          string
	Secret      string
	RedirectURL string
	Scope       string
}

type User struct {
	ID          int64     `json:"uid"`
	Name        string    `json:"name"`
	Gender      string    `json:"gender"`
	StrBirthday string    `json:"birthday"`
	Birthday    time.Time `json:"-"`
	Email       string    `json:"email"`
	Photo       string    `json:"-"`
	Picture     struct {
		Data struct {
			Url string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// AccessToken describes oath server response
type AccessToken struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires"`
}

type mockHTTPClient interface {
	Get(url string) (res *http.Response, err error)
}

func (client *Client) base(action string) url.URL {
	u := url.URL{}
	u.Host = host
	u.Scheme = scheme
	u.Path = action

	query := u.Query()
	query.Add(appIDParameter, client.ID)
	query.Add(redirectParameter, client.RedirectURL)

	u.RawQuery = query.Encode()
	return u
}

// DialogURL is url for vk auth dialog
func (client *Client) DialogURL() url.URL {
	u := client.base(authAction)

	query := u.Query()
	query.Add(scopeParameter, client.Scope)
	query.Add(responseTypeParameter, responseTypeCode)

	u.RawQuery = query.Encode()
	return u
}

func (client *Client) accessTokenURL(code string) url.URL {
	u := client.base(accessTokenAction)
	u.Host = fbGraphHost

	query := u.Query()
	query.Add(appSecretParameter, client.Secret)
	query.Add(codeParameter, code)

	u.RawQuery = query.Encode()
	return u
}

// GetAccessToken is handler for redirect, gets and returns access token
func (client *Client) GetAccessToken(req *http.Request) (token *AccessToken, err error) {
	query := req.URL.Query()
	code := query.Get(codeParameter)
	if code == "" {
		err = ErrorBadCode
		return nil, err
	}

	requestURL := client.accessTokenURL(code)
	res, err := httpClient.Get(requestURL.String())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	q, err := url.ParseQuery(string(data))
	if err != nil {
		return
	}
	token = &AccessToken{}
	token.AccessToken = q.Get(accessTokenParameter)
	if token.AccessToken == "" {
		err = ErrorBadResponse
		return
	}
	token.Expires, err = strconv.Atoi(q.Get(expiresParameter))
	return
}

func (client *Client) GetUser(token string) (user User, err error) {
	u := client.base(usersGetAction)
	u.Host = fbGraphHost
	q := u.Query()
	q.Del(appIDParameter)
	q.Del(redirectParameter)
	q.Add(accessTokenParameter, token)
	q.Add(fieldsParameter, fiealdsValue)
	u.RawQuery = q.Encode()
	res, err := httpClient.Get(u.String())
	if err != nil {
		return
	}
	decoder := json.NewDecoder(res.Body)
	if err = decoder.Decode(&user); err != nil {
		return
	}
	if user.Email == "" || user.Name == "" {
		err = ErrorBadResponse
		return
	}
	user.Birthday, err = time.Parse("02/01/2006", user.StrBirthday)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%+v", user)
	user.Photo = user.Picture.Data.Url
	return user, nil
}
