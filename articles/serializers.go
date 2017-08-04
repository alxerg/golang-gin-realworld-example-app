package articles

import (
    "github.com/gosimple/slug"
    "gopkg.in/gin-gonic/gin.v1"
    "golang-gin-starter-kit/users"
)

type ArticleSerializer struct {
    C *gin.Context
    ArticleModel
}

type TagSerializer struct {
    TagModel
}

func (self *TagSerializer) Response() string {
    return self.TagModel.Tag
}

type ArticleResponse struct {
    ID             uint        `json:"-"`
    Title          string      `json:"title"`
    Slug           string      `json:"slug"`
    Description    string      `json:"description"`
    Body           string      `json:"body"`
    CreatedAt      string      `json:"createdAt"`
    UpdatedAt      string      `json:"updatedAt"`
    Author         users.ProfileResponse `json:"author"`
    Tags           []string    `json:"tagList"`
    Favorite       bool        `json:"favorited"`
    FavoritesCount uint        `json:"favoritesCount"`
}

type ArticlesSerializer struct {
    C        *gin.Context
    Articles []ArticleModel
}


func (self *ArticleSerializer) Response() ArticleResponse {
    myUserModel := self.C.MustGet("my_user_model").(users.UserModel)
    authorSerializer := users.ProfileSerializer{self.C, self.Author}
    response := ArticleResponse{
        ID:          self.ID,
        Slug:        slug.Make(self.Title),
        Title:       self.Title,
        Description: self.Description,
        Body:        self.Body,
        CreatedAt:   self.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        //UpdatedAt:      self.UpdatedAt.UTC().Format(time.RFC3339Nano),
        UpdatedAt:      self.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        Author:         authorSerializer.Response(),
        Favorite:       self.isFavoriteBy(myUserModel),
        FavoritesCount: self.favoritesCount(),
    }
    response.Tags = make([]string, 0)
    for _, tag := range self.Tags {
        serializer := TagSerializer{tag}
        response.Tags = append(response.Tags, serializer.Response())
    }
    return response
}

func (self *ArticlesSerializer) Response() []ArticleResponse {
    response := []ArticleResponse{}
    for _, article := range self.Articles {
        serializer := ArticleSerializer{self.C, article}
        response = append(response, serializer.Response())
    }
    return response
}