package main

import (
	"bytes"
	"errors"
	"image/jpeg"
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

	r.GET("/prod/resize/:dimensions/:key", func(c *gin.Context) {
		h, w, err := parseDimensions(c.Param("dimensions"))
		if err != nil {
			panic(err)
		}

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

		c.Header("Content-Type", "image/jpeg")
		err = jpeg.Encode(c.Writer, m, nil)
		if err != nil {
			panic(err)
		}
	})

	r.Run(":8080")
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

	j, err := strconv.Atoi(vars[0])
	if err != nil {
		return 0, 0, err
	}

	return i, j, nil
}
