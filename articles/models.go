package articles

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	_ "fmt"
	"strconv"
	"time"

	"github.com/recoilme/golang-gin-realworld-example-app/users"
	sp "github.com/recoilme/slowpoke"
)

const (
	dbSlug    = "db/article/slug"
	dbCounter = "db/article/counter"
	dbArticle = "db/article/article"
	dbTag     = "db/article/tag"
)

type ArticleModel struct {
	ID          uint32
	Slug        string `gorm:"unique_index"`
	Title       string
	Description string `gorm:"size:2048"`
	Body        string `gorm:"size:2048"`
	Author      ArticleUserModel
	AuthorID    uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []TagModel     `gorm:"many2many:article_tags;"`
	Comments    []CommentModel `gorm:"ForeignKey:ArticleID"`
}

type ArticleUserModel struct {
	UserModel      users.UserModel
	UserModelID    uint32
	ArticleModels  []ArticleModel  `gorm:"ForeignKey:AuthorID"`
	FavoriteModels []FavoriteModel `gorm:"ForeignKey:FavoriteByID"`
}

type FavoriteModel struct {
	ID           uint32
	Favorite     ArticleModel
	FavoriteID   uint32
	FavoriteBy   ArticleUserModel
	FavoriteByID uint32
}

type TagModel struct {
	ID  uint32
	Tag string `gorm:"unique_index"`
	//ArticleModels []ArticleModel `gorm:"many2many:article_tags;"`
}

type CommentModel struct {
	ID        uint32
	Article   ArticleModel
	ArticleID uint32
	Author    ArticleUserModel
	AuthorID  uint32
	Body      string `gorm:"size:2048"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetArticleUserModel(userModel users.UserModel) ArticleUserModel {
	var articleUserModel ArticleUserModel
	if userModel.ID == 0 {

		return articleUserModel
	}
	//TODO query user?
	/*
		db := common.GetDB()
		db.Where(&ArticleUserModel{
			UserModelID: userModel.ID,
		}).FirstOrCreate(&articleUserModel)
	*/
	articleUserModel.UserModel = userModel
	return articleUserModel
}

func (article ArticleModel) favoritesCount() uint {
	//db := common.GetDB()
	var count uint
	/*
		db.Model(&FavoriteModel{}).Where(FavoriteModel{
			FavoriteID: article.ID,
		}).Count(&count)
	*/
	return count
}

func (article ArticleModel) isFavoriteBy(user ArticleUserModel) bool {
	//db := common.GetDB()
	var favorite FavoriteModel
	/*
		db.Where(FavoriteModel{
			FavoriteID:   article.ID,
			FavoriteByID: user.ID,
		}).First(&favorite)
	*/
	return favorite.ID != 0
}

func (article ArticleModel) favoriteBy(user ArticleUserModel) (err error) {
	//	db := common.GetDB()
	//var favorite FavoriteModel
	/*
		err := db.FirstOrCreate(&favorite, &FavoriteModel{
			FavoriteID:   article.ID,
			FavoriteByID: user.ID,
		}).Error
	*/
	return err
}

func (article ArticleModel) unFavoriteBy(user ArticleUserModel) (err error) {
	/*
		db := common.GetDB()
		err := db.Where(FavoriteModel{
			FavoriteID:   article.ID,
			FavoriteByID: user.ID,
		}).Delete(FavoriteModel{}).Error
	*/
	return err
}

func checkSlug(article *ArticleModel) error {
	// check slug
	if article == nil || article.Slug == "" {
		return errors.New("UNIQUE constraint failed: article_model.slug")
	}
	return nil
}

// checkArticleConstr - check new article slug
func checkArticleConstr(article *ArticleModel) (err error) {
	err = checkSlug(article)
	if err != nil {
		return err
	}
	has, err := sp.Has(dbSlug, []byte(article.Slug))
	if err != nil {
		return err
	}
	if has {
		return errors.New("UNIQUE constraint failed: article_model.slug")
	}
	return err
}

func SaveOne(article *ArticleModel) (err error) {
	err = checkArticleConstr(article)
	if err != nil {
		return err
	}

	if article.ID == 0 {
		// new article
		aid, err := sp.Counter(dbCounter, []byte("aid"))
		if err != nil {
			return err
		}
		article.ID = uint32(aid)
		// workaround for sp crash
		sp.Close(dbCounter)
	}

	id32 := make([]byte, 4)
	binary.BigEndian.PutUint32(id32, article.ID)

	// store slug
	if err = sp.Set(dbSlug, []byte(article.Slug), id32); err != nil {
		return err
	}

	// store article
	//fmt.Println(id32, article)
	if err = sp.SetGob(dbArticle, id32, article); err != nil {
		return err
	}
	/*
		fmt.Println("id32", id32)
		var a ArticleModel
		sp.GetGob(dbArticle, id32, &a)
		fmt.Println("a", a)
		k, _ := sp.Keys(dbArticle, nil, uint32(0), uint32(0), false)
		fmt.Println("k", k)
	*/
	return err
}

func SaveOneComment(comment *CommentModel) (err error) {
	return err
}

func FindOneArticle(article *ArticleModel) (model ArticleModel, err error) {
	err = checkSlug(article)
	if err != nil {
		return model, err
	}
	aid, err := sp.Get(dbSlug, []byte(article.Slug))
	if err != nil {
		return model, err
	}
	//fmt.Println("aid", aid)
	if err = sp.GetGob(dbArticle, aid, &model); err != nil {
		return model, err
	}

	// update author? but author may change username..
	return model, err
}

func (self *ArticleModel) getComments() (err error) {
	/*
		db := common.GetDB()
		tx := db.Begin()
		tx.Model(self).Related(&self.Comments, "Comments")
		for i, _ := range self.Comments {
			tx.Model(&self.Comments[i]).Related(&self.Comments[i].Author, "Author")
			tx.Model(&self.Comments[i].Author).Related(&self.Comments[i].Author.UserModel)
		}
		err := tx.Commit().Error
	*/
	return err
}

func getAllTags() (models []TagModel, err error) {
	/*
		db := common.GetDB()
		var models []TagModel
		err := db.Find(&models).Error
	*/
	return models, err
}

func FindManyArticle(tag, author, limit, offset, favorited string) ([]ArticleModel, int, error) {
	//db := common.GetDB()
	var models []ArticleModel
	var err error

	offset_int, err := strconv.Atoi(offset)
	if err != nil {
		offset_int = 0
	}

	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 20
	}
	_ = limit_int
	_ = offset_int
	keys, err := sp.Keys(dbArticle, nil, uint32(limit_int), uint32(offset_int), false)
	fmt.Println(keys, err)
	if err != nil {
		return models, 0, err
	}
	for _, key := range keys {
		var model ArticleModel
		k := make([]byte, 4)
		buf := bytes.Buffer{}
		buf.Write(key)
		if err := gob.NewDecoder(&buf).Decode(&k); err == nil {
			err := sp.GetGob(dbArticle, k, &model)
			//fmt.Println(key, err, model)
			if err != nil {
				break
			}
		} else {
			fmt.Println("kerr", err)
		}

		models = append(models, model)
	}
	cnt, _ := sp.Count(dbArticle)
	return models, int(cnt), err
	/*
		tx := db.Begin()
		if tag != "" {
			var tagModel TagModel
			tx.Where(TagModel{Tag: tag}).First(&tagModel)
			if tagModel.ID != 0 {
				tx.Model(&tagModel).Offset(offset_int).Limit(limit_int).Related(&models, "ArticleModels")
				count = tx.Model(&tagModel).Association("ArticleModels").Count()
			}
		} else if author != "" {
			var userModel users.UserModel
			tx.Where(users.UserModel{Username: author}).First(&userModel)
			articleUserModel := GetArticleUserModel(userModel)

			if articleUserModel.ID != 0 {
				count = tx.Model(&articleUserModel).Association("ArticleModels").Count()
				tx.Model(&articleUserModel).Offset(offset_int).Limit(limit_int).Related(&models, "ArticleModels")
			}
		} else if favorited != "" {
			var userModel users.UserModel
			tx.Where(users.UserModel{Username: favorited}).First(&userModel)
			articleUserModel := GetArticleUserModel(userModel)
			if articleUserModel.ID != 0 {
				var favoriteModels []FavoriteModel
				tx.Where(FavoriteModel{
					FavoriteByID: articleUserModel.ID,
				}).Offset(offset_int).Limit(limit_int).Find(&favoriteModels)

				count = tx.Model(&articleUserModel).Association("FavoriteModels").Count()
				for _, favorite := range favoriteModels {
					var model ArticleModel
					tx.Model(&favorite).Related(&model, "Favorite")
					models = append(models, model)
				}
			}
		} else {
			db.Model(&models).Count(&count)
			db.Offset(offset_int).Limit(limit_int).Find(&models)
		}

		for i, _ := range models {
			tx.Model(&models[i]).Related(&models[i].Author, "Author")
			tx.Model(&models[i].Author).Related(&models[i].Author.UserModel)
			tx.Model(&models[i]).Related(&models[i].Tags, "Tags")
		}
		err = tx.Commit().Error
	*/
}

func (self *ArticleUserModel) GetArticleFeed(limit, offset string) ([]ArticleModel, int, error) {

	//db := common.GetDB()
	var models []ArticleModel
	var count int

	offset_int, err := strconv.Atoi(offset)
	if err != nil {
		offset_int = 0
	}
	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 20
	}
	_ = limit_int
	_ = offset_int
	/*
		tx := db.Begin()
		followings := self.UserModel.GetFollowings()
		var articleUserModels []uint
		for _, following := range followings {
			articleUserModel := GetArticleUserModel(following)
			articleUserModels = append(articleUserModels, articleUserModel.ID)
		}

		tx.Where("author_id in (?)", articleUserModels).Order("updated_at desc").Offset(offset_int).Limit(limit_int).Find(&models)

		for i, _ := range models {
			tx.Model(&models[i]).Related(&models[i].Author, "Author")
			tx.Model(&models[i].Author).Related(&models[i].Author.UserModel)
			tx.Model(&models[i]).Related(&models[i].Tags, "Tags")
		}
		err = tx.Commit().Error
	*/
	return models, count, err
}

func (model *ArticleModel) setTags(tags []string) error {
	/*
		db := common.GetDB()
		var tagList []TagModel
		for _, tag := range tags {
			var tagModel TagModel
			err := db.FirstOrCreate(&tagModel, TagModel{Tag: tag}).Error
			if err != nil {
				return err
			}
			tagList = append(tagList, tagModel)
		}
		model.Tags = tagList
	*/
	return nil
}

func (model *ArticleModel) Update(article ArticleModel) (err error) {
	SaveOne(&article)
	/*
		db := common.GetDB()
		err := db.Model(model).Update(data).Error
	*/
	return SaveOne(&article)
}

func DeleteArticleModel(condition interface{}) (err error) {
	/*
		db := common.GetDB()
		err := db.Where(condition).Delete(ArticleModel{}).Error
	*/
	return err
}

func DeleteCommentModel(condition interface{}) (err error) {
	/*
		db := common.GetDB()
		err := db.Where(condition).Delete(CommentModel{}).Error
	*/
	return err
}
