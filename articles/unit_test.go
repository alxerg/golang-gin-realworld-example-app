package articles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/recoilme/golang-gin-realworld-example-app/common"
	"github.com/recoilme/golang-gin-realworld-example-app/users"
	sp "github.com/recoilme/slowpoke"
	"github.com/stretchr/testify/assert"
	"gopkg.in/gin-gonic/gin.v1"
)

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	str := "RandString"
	asserts.Equal(len(str), 10, "length should be 10")
}

func TestNewGob(t *testing.T) {
	asserts := assert.New(t)

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(42))
	err := sp.SetGob("db/test", b, 13)
	asserts.Nil(err)
	keys, err := sp.Keys("db/test", nil, uint32(0), uint32(0), false)
	asserts.Nil(err)
	//fmt.Println(keys)

	vals := sp.Gets("db/test", keys)
	//fmt.Println(vals)
	var val int
	sp.GetGob("db/test", vals[0], &val)
	//fmt.Println(val)
	asserts.Equal(13, val)
	sp.DeleteFile("db/test")
}

func HeaderTokenMock(req *http.Request, u uint32) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", common.GenToken(u)))
}

func TestWithoutAuth(t *testing.T) {
	asserts := assert.New(t)
	//You could write the reset database code here if you want to create a database for this block
	//resetDB()

	r := gin.New()

	users.UsersRegister(r.Group("/users"))

	ArticlesAnonymousRegister(r.Group("/articles"))
	r.Use(users.AuthMiddleware(true))
	users.UserRegister(r.Group("/user"))
	users.ProfileRegister(r.Group("/profiles"))

	ArticlesRegister(r.Group("/articles"))

	for num, testData := range unauthRequestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest(testData.method, testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		testData.init(req)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		res := asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
		if !res {
			fmt.Println("num", num)
		}
	}
}

var unauthRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	//Testing will run one by one, so you can combine it to a user story till another init().
	//And you can modified the header or body in the func(req *http.Request) {}

	//---------------------   Testing for user register   ---------------------
	{
		func(req *http.Request) {
			common.ResetUsersDBWithMock()
		},
		"/users/",
		"POST",
		`{"user":{"username": "user1","email": "e@mail.ru","password": "password","image":"http://image/1.jpg"}}`,
		http.StatusCreated,
		`{"user":{"username":"user1","email":"e@mail.ru","bio":"","image":"http://image/1.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"","image":"http://image/1.jpg","following":false}}`,
		"request should return self profile",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles/",
		"POST",
		`{
			"article": {
				"title": "How to train your dragon",
				"description": "Ever wonder how?",
				"body": "You have to believe"
			}
		}`,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"","image":"http://image/1.jpg","following":false}}`,
		"request should return self profile",
	},
}
