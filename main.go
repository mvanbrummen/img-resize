package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nfnt/resize"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, struct {
			Status string
		}{
			"GREEN",
		})
	})

	r.GET("/resize/:dimensions/:key", func(c *gin.Context) {
		h, w, err := parseDimensions(c.Param("dimensions"))
		if err != nil {
			panic(err)
		}

		log.Printf("resizing to %dx%d\n", h, w)

		key := c.Param("key")

		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String("ap-southeast-2")},
		)

		downloader := s3manager.NewDownloader(sess)

		buffer := aws.NewWriteAtBuffer([]byte{})

		_, err = downloader.Download(buffer,
			&s3.GetObjectInput{
				Bucket: aws.String("dasless-images"),
				Key:    aws.String("/raw/" + key),
			})

		buf := bytes.NewBuffer(buffer.Bytes())
		image, err := jpeg.Decode(buf)
		if err != nil {
			panic(err)
		}

		m := resize.Resize(uint(w), uint(h), image, resize.Lanczos3)

		go func() {
			b := new(bytes.Buffer)
			err = jpeg.Encode(b, m, nil)
			reader := bytes.NewReader(b.Bytes())

			uploader := s3manager.NewUploader(sess)
			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String("dasless-images"),
				Key:    aws.String(fmt.Sprintf("/%dx%d/%s", h, w, key)),
				Body:   reader,
			})
			if err != nil {
				log.Println(err)
			}
		}()

		c.Header("Content-Type", "image/jpeg")
		err = jpeg.Encode(c.Writer, m, nil)
		if err != nil {
			panic(err)
		}
	})

	osPort := os.Getenv("PORT")
	var port string
	if osPort == "" {
		port = ":8080"
	} else {
		port = ":" + osPort
	}

	r.Run(port)
}

func parseDimensions(s string) (int, int, error) {
	vars := strings.Split(s, "x")
	if len(vars) < 2 {
		return 0, 0, errors.New("dimensions not right")
	}

	i, err := strconv.Atoi(vars[0])
	if err != nil {
		return 0, 0, err
	}

	j, err := strconv.Atoi(vars[1])
	if err != nil {
		return 0, 0, err
	}

	return i, j, nil
}
