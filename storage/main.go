package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	minio "github.com/minio/minio-go"
)

var (
	minioClient *minio.Client
)

type File struct {
	Name         string    `json:"name"`
	Size         string    `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

type ListFilesResponse struct {
	Data []File `json:"data"`
}

func main() {
	// Init Minio
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	var err error
	minioClient, err = minio.New("minio-minio-svc:9000", minioAccessKey, minioSecretKey, false)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.GET("/buckets", listBucketsHandler)
	r.GET("/buckets/:name/files", listFilesHandler)
	r.POST("/buckets/:name", createBucketHandler)
	r.POST("/buckets/:name/files", createFileHandler)
	if err := r.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}

func listBucketsHandler(c *gin.Context) {
	if buckets, err := minioClient.ListBuckets(); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		c.JSON(http.StatusOK, buckets)
	}
}

func listFilesHandler(c *gin.Context) {
	bucketName := c.Param("name")
	done := make(chan struct{})
	defer close(done)
	objects := minioClient.ListObjectsV2(bucketName, "", true, done)

	res := ListFilesResponse{}
	for object := range objects {
		res.Data = append(res.Data, File{
			Name:         object.Key,
			Size:         fmt.Sprint(object.Size),
			LastModified: object.LastModified,
		})
	}
	c.JSON(http.StatusOK, res)
}

func createBucketHandler(c *gin.Context) {
	bucketName := c.Param("name")
	if err := minioClient.MakeBucket(bucketName, ""); err != nil {
		if exists, err := minioClient.BucketExists(bucketName); err == nil && exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Bucket already exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not create bucket",
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Bucket was created",
	})
}

func createFileHandler(c *gin.Context) {
	bucketName := c.Param("name")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file found",
		})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not read file",
		})
		return
	}
	_, err = minioClient.PutObject(bucketName, fileHeader.Filename, file, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not save file",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "File saved",
	})
}
