package gofbauth

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

type MockClient struct {
	Response *http.Response
}

func (c *MockClient) Get(url string) (res *http.Response, err error) {
	err = nil
	if c.Response == nil {
		err = http.ErrShortBody
	}
	return c.Response, err
}

func TestClient(t *testing.T) {
	client := Client{"APP_ID", "APP_SECRET", "REDIRECT_URI", "PERMISSIONS"}
	Convey("TestUrl", t, func() {
		url := client.DialogURL()
		should := "https://www.facebook.com/dialog/oauth?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&scope=PERMISSIONS"
		So(url.String(), ShouldEqual, should)
	})

	Convey("Test getName", t, func() {
		res := &http.Response{}
		body := `{
		  "id": "1487207854850126", 
		  "email": "ernado@ya.ru", 
		  "first_name": "Alexander", 
		  "gender": "male", 
		  "last_name": "Razumov", 
		  "link": "https://www.facebook.com/app_scoped_user_id/1487207854850126/", 
		  "locale": "ru_RU", 
		  "name": "Alexander Razumov", 
		  "timezone": 4, 
		  "updated_time": "2014-07-12T06:19:29+0000", 
		  "verified": true
		}`
		res.Body = ioutil.NopCloser(bytes.NewBufferString(body))
		httpClient = &MockClient{res}
		user, err := client.GetUser("2")
		So(err, ShouldBeNil)
		So(user.FirstName, ShouldEqual, "Alexander")
		So(user.LastName, ShouldEqual, "Razumov")
		So(user.Email, ShouldEqual, "ernado@ya.ru")
		So(user.Gender, ShouldEqual, "male")

		Convey("Http error", func() {
			httpClient = &MockClient{nil}
			_, err := client.GetUser("1")
			So(err, ShouldNotBeNil)
		})

		Convey("json error", func() {
			body := `[[[]}`
			res.Body = ioutil.NopCloser(bytes.NewBufferString(body))
			_, err := client.GetUser("1")
			So(err, ShouldNotBeNil)
		})
		Convey("Server error", func() {
			body := `{"response": {"error": "500"}}`
			res.Body = ioutil.NopCloser(bytes.NewBufferString(body))
			_, err := client.GetUser("1")
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Test accessTokenUrl", t, func() {
		Convey("Request url ok", func() {
			urlStr := "https://graph.facebook.com/oauth/access_token?client_id=APP_ID&client_secret=APP_SECRET&code=CODE&redirect_uri=REDIRECT_URI"
			url := client.accessTokenURL("CODE")
			So(url.String(), ShouldEqual, urlStr)
		})

		urlStr := "http://REDIRECT_URI?code=7a6fa4dff77a228eeda56603b8f53806c883f011c40b72630bb50df056f6479e52a"
		req, _ := http.NewRequest("GET", urlStr, nil)

		resTok := &http.Response{}
		body := `access_token=533bacf01e11f55b536a565b57531ac114461ae8736d6506a3&expires=43200`
		resTok.Body = ioutil.NopCloser(bytes.NewBufferString(body))
		httpClient = &MockClient{resTok}

		tok, err := client.GetAccessToken(req)
		So(err, ShouldBeNil)
		So(tok.AccessToken, ShouldEqual, "533bacf01e11f55b536a565b57531ac114461ae8736d6506a3")
		So(tok.Expires, ShouldEqual, 43200)

		Convey("Bad response", func() {
			resTok.Body = ioutil.NopCloser(bytes.NewBufferString("asdfasdf"))
			httpClient = &MockClient{resTok}
			_, err := client.GetAccessToken(req)
			So(err, ShouldNotBeNil)
		})

		Convey("Bad urk", func() {
			req, _ = http.NewRequest("GET", "http://REDIRECT_URI?error=kek", nil)
			_, err := client.GetAccessToken(req)
			So(err, ShouldNotBeNil)
		})

		Convey("Http error", func() {
			httpClient = &MockClient{nil}
			_, err := client.GetAccessToken(req)
			So(err, ShouldNotBeNil)
		})
	})
}
