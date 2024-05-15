package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type User struct {
	gorm.Model
	Username string
	File     []File
}

type File struct {
	gorm.Model
	UserId           uint
	Name             string
	WebUrl           string
	DetectionStatus  string
	TextSpeechStatus string
}

func uploadFile(filePath string) (string, error) {

	ctx := context.Background()
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return "", err
	}

	// Make a new bucket called testbucket.
	bucketName := "testbucket"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			// c.AbortWithError(http.StatusInternalServerError, err)
			return "", err
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	// Upload the test file
	// Change the value of filePath if the file is in another location
	objectName := "testdata"

	contentType := "application/octet-stream"

	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	return "success", nil
}

func main() {

	dsn := "host=localhost user=server password=server dbname=server port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&User{})

	if err != nil {
		panic("migration failed for user")
	}

	err = db.AutoMigrate(&File{})

	if err != nil {
		panic("migration failed for file")
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	type userJson struct {
		Username string `json:"username"`
	}

	r.POST("/create-user", func(c *gin.Context) {

		var user userJson

		// Call BindJSON to bind the received JSON to
		// newAlbum.
		if err := c.BindJSON(&user); err != nil {
			return
		}

		createUser := User{Username: user.Username}

		result := db.Create(&createUser)

		fmt.Println("user created", result)

		fmt.Println("create user", user.Username)

		c.JSON(http.StatusOK, gin.H{
			"message": "user created: " + createUser.Username,
		})
	})

	r.POST("/upload", func(c *gin.Context) {

		username := c.Query("username")

		var user User

		db.First(&user, "username = ?", username)

		if user.Username == "" {
			c.JSON(http.StatusOK, gin.H{
				"message": "user not found",
			})
			return
		}

		fmt.Println("username", user.Username)

		file, _ := c.FormFile("file")

		filePath := "/tmp/test/" + file.Filename

		var fileQuery File
		db.First(&fileQuery, "name = ?", filePath)

		if fileQuery.ID != 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "file already exists",
			})
			return
		}

		err = c.SaveUploadedFile(file, filePath)

		_, err = uploadFile(filePath)
		if err != nil {
			fmt.Println("this error", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		dbFile := File{UserId: user.ID, Name: filePath, DetectionStatus: "init", TextSpeechStatus: "init"}

		result := db.Create(&dbFile)

		fmt.Println("db file save", result.Name())

		fmt.Print("uploaded: ", file.Filename, file.Header)

		c.JSON(http.StatusOK, gin.H{
			"message": "upload successful",
		})
	})

	r.GET("/file-status", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/process/text-to-string", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Upload the file to specific dst.
		// c.SaveUploadedFile(file, dst)

		for i := 0; i < 10; i++ {
			conn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
			time.Sleep(time.Second)
		}

		// c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/process/language", func(c *gin.Context) {

		fileType := c.Query("type")
		fmt.Println("type:", fileType)

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
