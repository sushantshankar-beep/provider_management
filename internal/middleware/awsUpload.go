package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

type S3Uploader struct {
	s3Client   *s3.S3
	bucketName string
}

type FieldConfig struct {
	FormFieldName string
	ContextKey    string
}

func NewS3Uploader(sess *session.Session, bucketName string) *S3Uploader {
	return &S3Uploader{
		s3Client:   s3.New(sess),
		bucketName: bucketName,
	}
}

func (u *S3Uploader) UploadMiddleware(fieldConfigs []FieldConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		form, err := c.MultipartForm()
		if err != nil {
			c.Next()
			return
		}

		uploadedURLs := make(map[string][]string)
        log.Println("hbhjvh",uploadedURLs)
		for _, config := range fieldConfigs {
			files := form.File[config.FormFieldName]
			if len(files) == 0 {
				continue
			}

			var urls []string
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": fmt.Sprintf("Failed to open file: %v", err),
					})
					c.Abort()
					return
				}
				defer file.Close()

				buf := new(bytes.Buffer)
				if _, err := buf.ReadFrom(file); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": fmt.Sprintf("Failed to read file: %v", err),
					})
					c.Abort()
					return
				}

				key := fmt.Sprintf("%d_%s", time.Now().UnixNano()/int64(time.Millisecond), fileHeader.Filename)

				uploadInput := &s3.PutObjectInput{
					Bucket:      aws.String(u.bucketName),
					Key:         aws.String(key),
					Body:        bytes.NewReader(buf.Bytes()),
					ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
					ACL:         aws.String("public-read"),
				}

				result, err := u.s3Client.PutObject(uploadInput)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": fmt.Sprintf("Failed to upload to S3: %v", err),
					})
					c.Abort()
					return
				}

				url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.bucketName, key)
				urls = append(urls, url)

				fmt.Printf("Uploaded file: %s, ETag: %s\n", url, *result.ETag)
			}

			if len(urls) > 0 {
				uploadedURLs[config.ContextKey] = urls
			}
		}

		for key, urls := range uploadedURLs {
			
				c.Set(key, urls)
		}

		c.Next()
	}
}

func GetUploadedURL(c *gin.Context, key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}

	url, ok := value.(string)
	return url, ok
}

func GetUploadedURLs(c *gin.Context, key string) ([]string, bool) {
	log.Println("skjabcjxsba",key)
	value, exists := c.Get(key)
	log.Println("valueee")
	if !exists {
		return nil, false
	}

	urls, ok := value.([]string)
	return urls, ok
}
