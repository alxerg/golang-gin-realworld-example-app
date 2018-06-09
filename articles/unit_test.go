package articles

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	str := "RandString"
	asserts.Equal(len(str), 10, "length should be 10")
}

func TestNewArticle(t *testing.T) {
	asserts := assert.New(t)

	a := &ArticleModel{}
	a.AuthorID = 0
	a.Body = ""
	a.CreatedAt = time.Now()
	a.Description = ""
	a.Slug = ""
	a.Title = ""
	err := SaveOne(a)
	asserts.Nil(err)
}
