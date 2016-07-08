// Package s3config .
package s3config

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type s3InfoType int

// s3InfoTypes.
const (
	s3InfoTypeKey s3InfoType = iota + 1
	s3InfoTypeEnv
)

// minimum interval.
const (
	minInterval = 10 * time.Second
)

// S3Info(aws properties).
type S3Info struct {
	Type            s3InfoType
	Region          *string
	AccessKey       string
	SecretAccessKey string
	Bucket          string
	Folder          string
	FileName        string
	Interval        time.Duration
}

// NewS3Info set from server env or role.
func NewS3Info(region string, bucket string, folder string, fileName string, interval time.Duration) *S3Info {
	return &S3Info{
		Type:     s3InfoTypeEnv,
		Region:   aws.String(region),
		Bucket:   bucket,
		Folder:   folder,
		FileName: fileName,
		Interval: interval,
	}
}

// NewS3InfoKey set from credentials.csv.
func NewS3InfoKey(region string, accessKey string, secretAccessKey string, bucket string, folder string, fileName string, interval time.Duration) *S3Info {
	return &S3Info{
		Type:            s3InfoTypeKey,
		Region:          aws.String(region),
		AccessKey:       accessKey,
		SecretAccessKey: secretAccessKey,
		Bucket:          bucket,
		Folder:          folder,
		FileName:        fileName,
		Interval:        interval,
	}
}

// Bind yaml from s3 -> config.
func Bind(s3Info *S3Info, config interface{}) error {

	// check
	if err := checkS3Info(s3Info); err != nil {
		return err
	}
	if config == nil {
		return errors.New("config is nil")
	}

	// create aws credentials
	var creds *credentials.Credentials
	switch s3Info.Type {
	case s3InfoTypeKey:
		creds = credentials.NewStaticCredentials(s3Info.AccessKey, s3Info.SecretAccessKey, "")
	case s3InfoTypeEnv:
		creds = credentials.NewEnvCredentials()
	}

	// create S3 service
	s3Svc := s3.New(session.New(), &aws.Config{
		Region:      s3Info.Region,
		Credentials: creds,
	})

	// load file
	if err := loadFile(s3Svc, s3Info.Bucket, s3Info.Folder, s3Info.FileName, config); err != nil {
		return err
	}

	// reload
	go func() {
		for {
			// interval
			time.Sleep(s3Info.Interval)
			if err := loadFile(s3Svc, s3Info.Bucket, s3Info.Folder, s3Info.FileName, config); err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	return nil
}

// loadFile file(yaml) -> config.
func loadFile(s3Svc *s3.S3, bucket string, folder string, fileName string, config interface{}) error {

	// get file
	key := folder + "/" + fileName // full fileName
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	resp, err := s3Svc.GetObject(params)
	if err != nil {
		return errors.New(fmt.Sprintf("file not found [%s/%s %s]", bucket, key, err.Error()))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("file error data %+v", resp.Body))
	}

	// yaml -> config
	if err := yaml.Unmarshal(body, config); err != nil {
		fmt.Println(err.Error())
		return errors.New(fmt.Sprintf("file not yaml %+v", body))
	}

	return nil
}

// checkS3Info.
func checkS3Info(s3Info *S3Info) error {
	switch s3Info.Type {
	case s3InfoTypeKey:
		if s3Info.Region == nil || s3Info.AccessKey == "" || s3Info.SecretAccessKey == "" {
			return errors.New("s3Info not enougth")
		}
	case s3InfoTypeEnv:
		if s3Info.Region == nil {
			return errors.New("s3Info not region")
		}
	default:
		return errors.New("s3Info not type")
	}

	if s3Info.Bucket == "" || s3Info.FileName == "" {
		return errors.New("s3Info not enougth(bucket/folder/file)")
	}

	if s3Info.Interval < minInterval {
		// set minInterval
		s3Info.Interval = minInterval
	}

	return nil
}
