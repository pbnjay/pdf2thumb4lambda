//go:build lambda

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Upload struct {
	Records []struct {
		S3 struct {
			Bucket struct {
				Name string `json:"name"`
				ARN  string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
				Size      int64  `json:"size"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func lambdaRenderer_tempFiles(ctx context.Context, event *S3Upload) (string, error) {
	bucketName := event.Records[0].S3.Bucket.Name
	keyName := event.Records[0].S3.Object.Key

	if !strings.HasSuffix(keyName, ".pdf") {
		log.Println(bucketName + "/" + keyName)
		return "no", errors.New("invalid request")
	}

	// The session the S3 Downloader will use
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)

	////
	f, err := os.CreateTemp("/tmp", "src*.pdf")
	if err != nil {
		return "unable to create temp file", err
	}

	filePath := f.Name()

	// Write the contents of S3 Object to the file
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download file, %v", err)
	}
	log.Println("file is ", n, "bytes in size")

	output := strings.TrimSuffix(filePath, ".pdf") + ".png"
	err = renderPage(filePath, output)
	if err != nil {
		return filePath, err
	}

	f.Close()
	///////

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err = os.Open(output)
	if err != nil {
		return "missing", fmt.Errorf("failed to open file %q, %v", output, err)
	}

	mime := "image/png"
	// Upload the file to S3.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(DestinationBucket),
		Key:         aws.String(strings.TrimSuffix(keyName, ".pdf") + ".png"),
		Body:        f,
		ContentType: &mime,
	})
	if err != nil {
		return "boo", fmt.Errorf("failed to upload file, %v", err)
	}

	return filePath, nil
}

func lambdaRenderer(ctx context.Context, event *S3Upload) (string, error) {
	fmt.Println(len(event.Records), "potential records to thumbnail")
	for _, rec := range event.Records {
		start := time.Now()

		bucketName := rec.S3.Bucket.Name
		keyName := rec.S3.Object.Key
		if bucketName == "" || keyName == "" {
			continue
		}

		if !strings.HasSuffix(keyName, ".pdf") {
			continue
		}

		// The session the S3 Downloader will use
		sess := session.Must(session.NewSession())
		downloader := s3manager.NewDownloader(sess)

		src := aws.NewWriteAtBuffer([]byte{})

		// TODO: if file is too large stop here
		_, err := downloader.Download(src, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
		if err != nil {
			return "", fmt.Errorf("failed to download file, %v", err)
		}

		dest := &bytes.Buffer{}
		err = renderPageFromBytes(src.Bytes(), dest)
		if err != nil {
			return "", err
		}

		// Create an uploader with the session and default options
		uploader := s3manager.NewUploader(sess)

		// Upload the file to S3.
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(DestinationBucket),
			Key:         aws.String(strings.TrimSuffix(keyName, ".pdf") + ".png"),
			Body:        dest,
			ContentType: aws.String("image/png"),
		})
		if err != nil {
			return "", fmt.Errorf("failed to upload file, %v", err)
		}

		elap := time.Since(start)
		fmt.Println("completed thumbnail for ", keyName, "in", elap)
	}

	return "ok", nil
}

var DestinationBucket = "destination-bucket-not-defined"

func main() {
	DestinationBucket = os.Getenv("DESTINATION_BUCKET")
	lambda.Start(lambdaRenderer)
}
