package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	uuid "github.com/satori/go.uuid"

	"gopkg.in/go-playground/validator.v9"
)

// Post is
type Post struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Text         string `json:"text" validate:"min=1,max=100"`
	ParentPostID string `json:"parent_post_id"`
	CommentCount int    `json:"comment_count"`
	PostedAt     string `json:"posted_at"`
}

// DB migration
func dbInit() {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbInit")
	}
	defer db.Close()
	db.AutoMigrate(&Post{})
}

// DB Get All
func dbGetAll() []*Post {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbGetAll")
	}
	defer db.Close()
	var posts []*Post
	db.Order("posted_at desc").Find(&posts)
	return posts
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	dbInit()

	// 投稿一覧
	router.GET("/posts", func(c *gin.Context) {
		posts := dbGetAll()
		c.JSON(200, posts)
	})

	// 投稿へのコメント一覧
	router.GET("/posts/:post_id/comments", func(c *gin.Context) {
		db, err := gorm.Open("sqlite3", "post.sqlite3")
		if err != nil {
			panic("You failed to dbGetAll")
		}
		defer db.Close()
		var posts []*Post

		// パスに記述されている post_id を取得している
		postID := c.Param("post_id")
		ell := db.Find(&posts, "parent_post_id = ?", postID)
		c.JSON(200, ell)
	})

	// 投稿作成
	router.POST("/posts/:path", func(c *gin.Context) {
		if c.Param("path") == "create" {

			fmt.Println("投稿作成の方に入れたよ")
			db, err := gorm.Open("sqlite3", "post.sqlite3")
			if err != nil {
				panic("You can't open DB (dbInsert())")
			}
			defer db.Close()
			var post *Post
			c.Bind(&post)

			if (post.UserID != "11111111-1111-1111-1111-111111111111") && (post.UserID != "22222222-2222-2222-2222-222222222222") && (post.UserID != "33333333-3333-3333-3333-333333333333") {
				c.JSON(400, gin.H{
					"result": "ユーザIDが不適切です",
				})
				return
			}

			validate := validator.New()
			err2 := validate.Struct(post)
			if err2 != nil {
				c.JSON(400, gin.H{
					"result": "不正なコンテンツです",
				})
				return
			}

			// satoriのuuidを利用して一意な値を生成している
			uuid, err := uuid.NewV4()
			post.ID = uuid.String()
			post.ParentPostID = ""
			post.CommentCount = 0
			post.PostedAt = time.Now().Format("2018-05-01 11:20:10")

			db.Create(&post)
			c.JSON(200, gin.H{
				"result": "OK",
			})
		}
	})

	// 投稿へのコメント作成
	// /posts/:path と途中まで同様にしてやらないとコンフリクトが返される
	router.POST("/posts/:path/*action", func(c *gin.Context) {
		fmt.Println(c.Param("path"))
		fmt.Println("AAAAAAAAAA")
		fmt.Println(c.Param("action"))
		if c.Param("action") == "/comments/create" {
			fmt.Println("投稿のコメント作成の方へ入れたよ〜")
			db, err := gorm.Open("sqlite3", "post.sqlite3")
			if err != nil {
				panic("You can't open DB (dbInsertComment())")
			}
			defer db.Close()
			var post *Post
			//var post2 Post
			c.Bind(&post)

			// ユーザのIDが適切であるかを判断している
			if (post.UserID != "11111111-1111-1111-1111-111111111111") && (post.UserID != "22222222-2222-2222-2222-222222222222") && (post.UserID != "33333333-3333-3333-3333-333333333333") {
				c.JSON(400, gin.H{
					"result": "ユーザIDが不適切です",
				})
				return
			}

			// Textが validation をクリアしているかを判断している
			validate := validator.New()
			err2 := validate.Struct(post)
			if err2 != nil {
				c.JSON(400, gin.H{
					"result": "不正なコンテンツです",
				})
				return
			}

			// satoriのuuidを利用して一意な値を生成している
			uuid, err := uuid.NewV4()
			post.ID = uuid.String()

			post.ParentPostID = c.Param("path")

			var post2 Post
			result := db.Where("ID = ?", post.ParentPostID).First(&post2).Update("CommentCount", (post2.CommentCount + 1))
			if result.Error != nil {
				c.JSON(400, gin.H{
					"result": "PostIDが不適切です",
				})
				return
			}

			post.CommentCount = 0
			post.PostedAt = time.Now().Format("2018-05-01 11:20:10")

			db.Create(&post)
			c.JSON(200, gin.H{
				"result": "OK",
			})
		}
	})

	router.Run()
}
