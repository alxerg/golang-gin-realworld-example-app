package articles

import (
	"bytes"
	"errors"
	"fmt"
	_ "fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/recoilme/golang-gin-realworld-example-app/common"
	"github.com/recoilme/golang-gin-realworld-example-app/users"

	sp "github.com/recoilme/slowpoke"
)

const (
	dbSlug       = "db/article/slug"
	dbCounter    = "db/article/counter"
	dbArticle    = "db/article/article"
	dbTag        = "db/article/tag"
	dbComment    = "db/article/comment"
	dbArticleUid = "db/article/uid/%d"
	dbFavMS      = "db/article/favms"
	dbFavSM      = "db/article/favsm"
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
	CommentsIds []uint32
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
	//TODO why query user?
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

	aid32 := common.Uint32toBin(article.ID)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, aid32...)
	masterstar = append(masterstar, '*')
	keys, _ := sp.Keys(dbFavSM, masterstar, 0, 0, true)
	/*
		db.Model(&FavoriteModel{}).Where(FavoriteModel{
			FavoriteID: article.ID,
		}).Count(&count)
	*/
	return uint(len(keys))
}

func (article ArticleModel) isFavoriteBy(user ArticleUserModel) bool {
	//db := common.GetDB()

	master := user.UserModel.ID
	slave := article.ID
	_, slavemaster := common.GetMasterSlave(master, slave)
	has, _ := sp.Has(dbFavSM, slavemaster)
	return has
	/*
		db.Where(FavoriteModel{
			FavoriteID:   article.ID,
			FavoriteByID: user.ID,
		}).First(&favorite)
	*/
	//return favorite.ID != 0
}

func (article ArticleModel) favoriteBy(user ArticleUserModel) (err error) {
	log.Println("favoriteBy")
	masterslave, slavemaster := common.GetMasterSlave(user.UserModel.ID, article.ID)
	err = sp.Set(dbFavMS, masterslave, nil)
	if err != nil {
		return err
	}
	err = sp.Set(dbFavSM, slavemaster, nil)
	if err != nil {
		return err
	}

	return err
	//	db := common.GetDB()
	//var favorite FavoriteModel
	/*
		err := db.FirstOrCreate(&favorite, &FavoriteModel{
			FavoriteID:   article.ID,
			FavoriteByID: user.ID,
		}).Error
	*/

}

func (article ArticleModel) unFavoriteBy(user ArticleUserModel) (err error) {
	log.Println("unFavoriteBy")
	masterslave, slavemaster := common.GetMasterSlave(user.UserModel.ID, article.ID)
	_, err = sp.Delete(dbFavMS, masterslave)
	if err != nil {
		return err
	}
	_, err = sp.Delete(dbFavSM, slavemaster)
	if err != nil {
		return err
	}
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
	if article.ID == 0 {
		// new article
		err = checkArticleConstr(article)
		if err != nil {
			return err
		}
		aid, err := sp.Counter(dbCounter, []byte("aid"))
		if err != nil {
			return err
		}
		article.ID = uint32(aid)
		article.CreatedAt = time.Now()
		article.UpdatedAt = time.Now()
		article.AuthorID = article.Author.UserModel.ID

	} else {
		err = checkSlug(article)
		if err != nil {
			return err
		}
		article.UpdatedAt = time.Now()
		//TODO update slug if change
	}

	id32 := common.Uint32toBin(article.ID)

	// store slug
	if err = sp.Set(dbSlug, []byte(article.Slug), id32); err != nil {
		return err
	}

	// store articleid -> userid
	//fmt.Println(id32, article.AuthorID)
	//fmt.Printf("%+v\n", article.Author)
	if err = sp.Set(dbArticle, id32, common.Uint32toBin(article.AuthorID)); err != nil {
		return err
	}

	// Store every article by uid in separate file

	f := fmt.Sprintf(dbArticleUid, article.AuthorID)
	//log.Println(f)
	if err = sp.SetGob(f, id32, article); err != nil {
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
	fmt.Println("\n\nSaveOneComment")
	if comment.ID == 0 {
		// new comment
		cid, err := sp.Counter(dbCounter, []byte("cid"))
		if err != nil {
			return err
		}
		comment.ID = uint32(cid)
		comment.CreatedAt = time.Now()
		comment.UpdatedAt = time.Now()
		// workaround for sp crash
		sp.Close(dbCounter)
	}

	id32 := common.Uint32toBin(comment.ID)

	// store comment
	//fmt.Println(id32, article)
	if err = sp.SetGob(dbComment, id32, comment); err != nil {
		return err
	}
	var cids = comment.Article.CommentsIds
	//fmt.Printf("\n\ncomment.Article %+v\n", comment.Article)

	//fmt.Printf("\n\ncomment.Author.UserModel %+v\n", comment.Author.UserModel)

	var isNew = true
	for _, val := range comment.Article.CommentsIds {
		if val == comment.ID {
			isNew = false
			break
		}
	}

	if isNew {
		cids = append(cids, comment.ID)
		//log.Println("is new cids", cids)
		comment.Article.CommentsIds = cids
		aid32 := common.Uint32toBin(comment.Article.ID)

		f := fmt.Sprintf(dbArticleUid, comment.Article.ID)
		if err = sp.SetGob(f, aid32, comment.Article); err != nil {
			return err
		}

	}
	return err
}

func FindOneArticle(article *ArticleModel) (model ArticleModel, err error) {
	log.Println("FindOneArticle")
	err = checkSlug(article)
	if err != nil {
		return model, err
	}
	aid, err := sp.Get(dbSlug, []byte(article.Slug))
	if err != nil {
		return model, err
	}
	//log.Println(aid)
	// Get uid
	uid, err := sp.Get(dbArticle, aid)
	if err != nil {
		return model, err
	}
	//log.Println("uid", uid)
	f := fmt.Sprintf(dbArticleUid, common.BintoUint32(uid))
	err = sp.GetGob(f, aid, &model)
	//log.Printf("model:%+v\n", model)

	if err != nil {
		return model, err
	}
	return model, err
}

func (self *ArticleModel) getComments() (err error) {
	cids := self.CommentsIds
	log.Println("getComments:", cids)
	var comments []CommentModel
	for _, cid := range cids {
		cid32 := common.Uint32toBin(cid)
		var com CommentModel
		if errCom := sp.GetGob(dbComment, cid32, &com); errCom == nil {
			comments = append(comments, com)
		}
	}
	self.Comments = comments
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
	log.Println("FindManyArticle")
	var models []ArticleModel
	var err error
	var cnt uint64
	var offset_int, limit_int int

	limit_int = 20

	offset_int, _ = strconv.Atoi(offset)

	limit_int, _ = strconv.Atoi(limit)

	if tag != "" {

	} else if author != "" {
		userModel, err := users.FindOneUser(&users.UserModel{Username: author})
		if err == nil {
			file := fmt.Sprintf(dbArticleUid, userModel.ID)
			keys, err := sp.Keys(file, nil, uint32(limit_int), uint32(offset_int), false)
			for _, key := range keys {
				var model ArticleModel
				if err = sp.GetGob(file, key, &model); err != nil {
					fmt.Println("kerr", err)
					break
				}
				models = append(models, model)
			}
		}
		//var userModel users.UserModel
		//tx.Where(users.UserModel{Username: author}).First(&userModel)
	} else if favorited != "" {
	} else {
		//no params
		log.Println("no params")
		keys, err := sp.Keys(dbArticle, nil, uint32(limit_int), uint32(offset_int), false)

		if err != nil {
			return models, 0, err
		}
		for _, key := range keys {
			var model ArticleModel

			uidb, err := sp.Get(dbArticle, key)
			if err != nil {
				break
			}
			uid := common.BintoUint32(uidb)

			if err = sp.GetGob(fmt.Sprintf(dbArticleUid, uid), key, &model); err != nil {
				fmt.Println("kerr", err)
				break
			}
			models = append(models, model)
		}
		cnt, _ = sp.Count(dbArticle)
	}
	//log.Println("models", err, models)
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

	followings := self.UserModel.GetFollowings()
	var allIds = make([][]byte, 0, 0)
	var aidUid = make(map[string]uint32)
	for _, following := range followings {
		//get limit + offset posts id from each user
		ids, err := sp.Keys(fmt.Sprintf(dbArticleUid, following.ID), nil, uint32(limit_int+offset_int), 0, false)
		if err == nil {

			allIds = append(allIds, ids...)
			for _, aids := range ids {
				aidUid[string(aids)] = following.ID
			}
		}
		cnt, err := sp.Count(fmt.Sprintf(dbArticleUid, following.ID))
		if err == nil {
			count += int(cnt)
		}

	}
	// sort desc
	sort.Slice(allIds, func(i, j int) bool {
		return bytes.Compare(allIds[i], allIds[j]) >= 0
	})

	curLimit := 0
	for num, id := range allIds {
		if num >= offset_int {
			if curLimit <= limit_int {

				userId := aidUid[string(id)]
				f := fmt.Sprintf(dbArticleUid, userId)
				var model = ArticleModel{}
				err = sp.GetGob(f, id, &model)
				if err == nil {
					models = append(models, model)
				}
				curLimit++
			}
		}
	}

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
	/*
		var tagList []TagModel
		for _, tag := range tags {
			var tagModel TagModel
			var err error
			var tagList []TagModel

			exists, err := sp.Has(dbTag, []byte(tag))
			if exists == false && err == nil {
				err = sp.Set(dbTag, []byte(tag), nil)
			}
			if err == nil {
				tagModel.Tag = tag
				tagList = append(tagList, tagModel)
			}
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

	log.Println("DeleteArticleModel")
	article := condition.(*ArticleModel)

	err = checkSlug(article)
	if err != nil {
		return err
	}
	aid, err := sp.Get(dbSlug, []byte(article.Slug))
	if err != nil {
		return err
	}

	_, err = sp.Delete(dbSlug, []byte(article.Slug))
	if err != nil {
		return err
	}

	_, err = sp.Delete(dbArticle, aid)
	if err != nil {
		return err
	}

	//todo comments?

	return err
}

func DeleteCommentModel(id uint) (err error) {

	_, err = sp.Delete(dbComment, common.Uint32toBin(uint32(id)))
	return err
}
