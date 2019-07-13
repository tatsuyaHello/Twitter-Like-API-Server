package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	uuid "github.com/satori/go.uuid"

	"gopkg.in/go-playground/validator.v9"
)

// Post は投稿に関する構造体
type Post struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Text         string `json:"text" validate:"min=1,max=100"`
	ParentPostID string `json:"parent_post_id"`
	CommentCount int    `json:"comment_count"`
	PostedAt     string `json:"posted_at"`
}

// TestUser はテストユーザに関する構造体である
type TestUser struct {
	ID   string
	Name string
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
func dbGetAll() ([]*Post, error) {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbGetAll")
	}
	defer db.Close()
	var posts []*Post
	db.Order("posted_at desc").Find(&posts)
	return posts, err
}

// dbGetComment はデータベースにある特定のPost_idに関するコメントを取得する
func dbGetComment(param string) ([]*Post, error) {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You failed to dbGetAll")
	}
	defer db.Close()
	var posts []*Post

	// パスに記述されている post_id を取得している
	postID := param
	db.Order("posted_at desc").Find(&posts, "parent_post_id = ?", postID)
	return posts, err
}

// createPost は新規投稿を作成する
func createPost(c *gin.Context) {
	db, err := gorm.Open("sqlite3", "post.sqlite3")
	if err != nil {
		panic("You can't open DB (dbInsert())")
	}
	defer db.Close()
	var post Post
	c.BindJSON(&post)

	// usersにはDBのTestUserに格納されている全ユーザの情報が格納される
	users := []TestUser{}
	db.Find(&users)

	// 有効なユーザであるか確認している
	checkValid := false
	for i := 0; i < len(users); i++ {
		if users[i].ID == post.UserID {
			checkValid = true
		}
	}

	if !checkValid {
		c.JSON(400, gin.H{
			"result":  "NG",
			"message": "ユーザIDが不適切です",
		})
		return
	}

	// Textが validation をクリアしているかを判断している
	validate := validator.New()
	validateErr := validate.Struct(post)
	if validateErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  "NG",
			"message": "テキストの長さが不適切です",
		})
		return
	}

	// satoriのuuidを利用して一意な値を生成している
	uuid, err := uuid.NewV4()
	post.ID = uuid.String()
	post.ParentPostID = ""
	post.CommentCount = 0
	post.PostedAt = time.Now().Format("2006-01-02 15:04:05")

	db.Create(&post)

	// サーバ側でエラーが発生した場合の処理
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "server error",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "OK",
	})
}

// createPostComment はあるPost_idに関する新規コメントを作成する
func createPostComment(c *gin.Context) {
	if c.Param("action") == "/comments/create" {
		db, err := gorm.Open("sqlite3", "post.sqlite3")
		if err != nil {
			panic("You can't open DB (dbInsertComment())")
		}
		defer db.Close()
		var post Post
		c.BindJSON(&post)

		// usersにはDBのTestUserに格納されている全ユーザの情報が格納される
		users := []TestUser{}
		db.Find(&users)

		// 有効なユーザであるか確認している
		checkValid := false
		for i := 0; i < len(users); i++ {
			if users[i].ID == post.UserID {
				checkValid = true
			}
		}

		if !checkValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"result":  "NG",
				"message": "ユーザIDが不適切です",
			})
			return
		}

		// Textが validation をクリアしているかを判断している
		validate := validator.New()
		validateErr := validate.Struct(post)
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"result":  "NG",
				"message": "テキストの長さが不適切です",
			})
			return
		}

		// satoriのuuidを利用して一意な値を生成している
		uuid, err := uuid.NewV4()
		post.ID = uuid.String()

		post.ParentPostID = c.Param("post_id")

		// parentPostはコメントポストの元となる親元のポストが入る
		var parentPost Post

		// コメントを作成するので、その元となる投稿のCommentCountを増加させる
		result := db.Where("ID = ?", post.ParentPostID).First(&parentPost).Update("CommentCount", (parentPost.CommentCount + 1))
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"result":  "NG",
				"message": "PostIDが不適切です",
			})
			return
		}

		post.CommentCount = 0
		post.PostedAt = time.Now().Format("2006-01-02 15:04:05")

		db.Create(&post)

		// サーバ側でエラーが発生した場合の処理
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "server error",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"result": "OK",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  "NG",
			"message": "不適切なURLです",
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
		posts, err := dbGetAll()

		// サーバ側でエラーが発生した場合の処理
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "server error",
			})
		}

		c.JSON(200, gin.H{
			"posts": posts,
		})
	})

	// 投稿へのコメント一覧
	router.GET("/posts/:post_id/comments", func(c *gin.Context) {
		postComments, err := dbGetComment(c.Param("post_id"))

		// サーバ側でエラーが発生した場合の処理
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "server error",
			})
		}

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
