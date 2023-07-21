package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error connecting to .env")
	}

	//GIN App Setup
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 8 << 20

	//S3 Uploader Setup - Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(ctx *gin.Context) {
		//Get the file
		file, err := ctx.FormFile("image")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}

		//Save the File
		f, openErr := file.Open()

		if openErr != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}

		res, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("airosmithdemobucket"),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		})

		fmt.Println("Location", res.Location)

		if uploadErr != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to upload Image 2" + uploadErr.Error(),
			})
		}

		// Render the page
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"image": res.Location,
		})

		//Update Image
		_, updateErr := client.CopyObject(context.TODO(), &s3.CopyObjectInput{
			Bucket:     aws.String("airosmithdemobucket"),
			CopySource: aws.String("airosmithdemobucket/" + file.Filename),
			Key:        aws.String(file.Filename),
		})

		if updateErr != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"error": "Failed to update old image" + updateErr.Error(),
			})
		}
	})

	//Listen and Serve on 0.0.0.0:8080
	r.Run()
}
