package main

import (
	"fmt"
	"net/http"
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

// dbInit はDBの初期化を行う
func dbInit() {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbInit")
	}
	defer db.Close()
	db.AutoMigrate(&Post{})
}

// dbGetAll はデータベースにある全てのPostを取得する
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

// dbGetComment はデータベースにある特定のPost_idに関するコメントを取得する
func dbGetComment(param string) []*Post {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbGetAll")
	}
	defer db.Close()
	var posts []*Post

	// パスに記述されている post_id を取得している
	postID := param
	db.Order("posted_at desc").Find(&posts, "parent_post_id = ?", postID)
	return posts
}

// createPost は新規投稿を作成する
func createPost(c *gin.Context) {
	fmt.Println("投稿作成の方に入れたよ")
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You can't open DB (dbInsert())")
	}
	defer db.Close()
	var post Post
	c.BindJSON(&post)
	if (post.UserID != "11111111-1111-1111-1111-111111111111") && (post.UserID != "22222222-2222-2222-2222-222222222222") && (post.UserID != "33333333-3333-3333-3333-333333333333") {
		c.JSON(400, gin.H{
			"result": "ユーザIDが不適切です",
		})
		return
	}

	validate := validator.New()
	err2 := validate.Struct(post)
	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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
	c.JSON(http.StatusOK, gin.H{
		"result": "OK",
	})
}

// createPostComment はあるPost_idに関する新規コメントを作成する
func createPostComment(c *gin.Context) {
	fmt.Println(c.Param("post_id"))
	fmt.Println("AAAAAAAAAA")
	fmt.Println(c.Param("action"))
	if c.Param("action") == "/comments/create" {
		fmt.Println("投稿のコメント作成の方へ入れたよ〜")
		db, err := gorm.Open("sqlite3", "post.sqlite3")
		if err != nil {
			panic("You can't open DB (dbInsertComment())")
		}
		defer db.Close()
		var post Post
		//var post2 Post
		c.BindJSON(&post)

		// ユーザのIDが適切であるかを判断している
		if (post.UserID != "11111111-1111-1111-1111-111111111111") && (post.UserID != "22222222-2222-2222-2222-222222222222") && (post.UserID != "33333333-3333-3333-3333-333333333333") {
			c.JSON(http.StatusBadRequest, gin.H{
				"result": "ユーザIDが不適切です",
			})
			return
		}

		// Textが validation をクリアしているかを判断している
		validate := validator.New()
		err2 := validate.Struct(post)
		if err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"result": "不正なコンテンツです",
			})
			return
		}

		// satoriのuuidを利用して一意な値を生成している
		uuid, err := uuid.NewV4()
		post.ID = uuid.String()

		post.ParentPostID = c.Param("post_id")

		var post2 Post
		result := db.Where("ID = ?", post.ParentPostID).First(&post2).Update("CommentCount", (post2.CommentCount + 1))
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"result": "PostIDが不適切です",
			})
			return
		}

		post.CommentCount = 0
		post.PostedAt = time.Now().Format("2018-05-01 11:20:10")

		db.Create(&post)
		c.JSON(http.StatusOK, gin.H{
			"result": "OK",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"result": "不適切なURLです",
		})
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// db初期化
	dbInit()

	// 投稿一覧
	router.GET("/posts", func(c *gin.Context) {
		posts := dbGetAll()
		fmt.Println("AAAAAA")
		for i := 0; i < len(posts); i++ {
			fmt.Println(*posts[i])
		}
		c.JSON(200, gin.H{
			"posts": posts,
		})
	})

	// 投稿へのコメント一覧
	router.GET("/posts/:post_id/comments", func(c *gin.Context) {
		postComments := dbGetComment(c.Param("post_id"))
		c.JSON(200, postComments)
	})

	// 投稿作成
	router.POST("/posts/:post_id", func(c *gin.Context) {
		if c.Param("post_id") == "create" {
			createPost(c)
		}
	})

	// 投稿へのコメント作成
	// /posts/:path と途中まで同様にしてやらないとコンフリクトが返される
	router.POST("/posts/:post_id/*action", func(c *gin.Context) {
		createPostComment(c)
	})

	router.Run()
}
