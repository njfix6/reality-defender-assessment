package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"

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

type SpeechToTextResponse struct {
	Text string
}

type LanguageResponse struct {
	Language string
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
	bucketName := "reality-defender-assessment-nick"
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
	objectName := "reality-defender-assessment-nick-" + filePath

	contentType := "application/octet-stream"

	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	return "success", nil
}

func main() {

	// dsn := "host=host.docker.internal user=server password=server dbname=server port=5432 sslmode=disable"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", os.Getenv("DB_HOST"), os.Getenv("DB_SERVER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

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
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

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

	r.GET("/process/speech-to-text", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		fileName := c.Query("filename")

		// Upload the file to specific dst.
		// c.SaveUploadedFile(file, dst)

		conn.WriteMessage(websocket.TextMessage, []byte("Calling AI Processer"))

		client := http.Client{}

		var jsonStr = []byte(`{"filename": "` + fileName + `"}`)
		req, err := http.NewRequest("POST", os.Getenv("AI_SERVER")+"/speech-to-text", bytes.NewBuffer(jsonStr))

		fmt.Println("making request")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		req.Header = http.Header{
			"Content-Type": {"application/json"},
		}

		res, err := client.Do(req)
		fmt.Println("making call")

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		decoder := json.NewDecoder(res.Body)

		var data SpeechToTextResponse
		err = decoder.Decode(&data)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fmt.Println("response: ", data.Text)

		conn.WriteMessage(websocket.TextMessage, []byte("AI Processer Completed"))

		// c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))

		conn.WriteMessage(websocket.TextMessage, []byte("text: "+data.Text))

		c.JSON(http.StatusOK, gin.H{
			"text": data.Text,
		})
	})

	r.GET("/process/language", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		fileName := c.Query("filename")

		// Upload the file to specific dst.
		// c.SaveUploadedFile(file, dst)

		conn.WriteMessage(websocket.TextMessage, []byte("Calling AI Processer"))

		client := http.Client{}

		var jsonStr = []byte(`{"filename": "` + fileName + `"}`)
		req, err := http.NewRequest("POST", os.Getenv("AI_SERVER")+"/language", bytes.NewBuffer(jsonStr))

		fmt.Println("making request")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		req.Header = http.Header{
			"Content-Type": {"application/json"},
		}

		res, err := client.Do(req)
		fmt.Println("making call")

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		decoder := json.NewDecoder(res.Body)

		var data LanguageResponse
		err = decoder.Decode(&data)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fmt.Println("response: ", data.Language)

		conn.WriteMessage(websocket.TextMessage, []byte("AI Processer Completed"))

		// c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))

		conn.WriteMessage(websocket.TextMessage, []byte("language: "+data.Language))

		c.JSON(http.StatusOK, gin.H{
			"language": data.Language,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
