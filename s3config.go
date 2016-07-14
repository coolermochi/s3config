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
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type s3InfoType int

// s3InfoTypes.
const (
	s3InfoTypeRole s3InfoType = iota + 1
	s3InfoTypeEnv
	s3InfoTypeKey
)

// minimum interval.
const (
	minInterval = 10 * time.Second
)

// S3Info(aws properties).
type S3Info struct {
	Type      s3InfoType
	Region    *string
	AccessKey string
	SecretKey string
	Bucket    string
	Folder    string
	FileName  string
	Interval  time.Duration
}

// NewS3Info set from server env or role.
func NewS3InfoRole(region string, bucket string, folder string, fileName string, interval time.Duration) *S3Info {
	return &S3Info{
		Type:     s3InfoTypeRole,
		Region:   aws.String(region),
		Bucket:   bucket,
		Folder:   folder,
		FileName: fileName,
		Interval: interval,
	}
}

// NewS3Info set from server env or role.
func NewS3InfoEnv(region string, bucket string, folder string, fileName string, interval time.Duration) *S3Info {
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
func NewS3InfoKey(region string, accessKey string, secretKey string, bucket string, folder string, fileName string, interval time.Duration) *S3Info {
	return &S3Info{
		Type:      s3InfoTypeKey,
		Region:    aws.String(region),
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		Folder:    folder,
		FileName:  fileName,
		Interval:  interval,
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
	case s3InfoTypeRole:
		ec2 := ec2metadata.New(session.New(), &aws.Config{
			HTTPClient: &http.Client{Timeout: 10 * time.Second},
		})
		creds = credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
			Client: ec2,
		})
	case s3InfoTypeEnv:
		creds = credentials.NewEnvCredentials()
	case s3InfoTypeKey:
		creds = credentials.NewStaticCredentials(s3Info.AccessKey, s3Info.SecretKey, "")
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
	case s3InfoTypeRole:
		if s3Info.Region == nil {
			return errors.New("s3Info not region")
		}
	case s3InfoTypeEnv:
		if s3Info.Region == nil {
			return errors.New("s3Info not region")
		}
	case s3InfoTypeKey:
		if s3Info.Region == nil || s3Info.AccessKey == "" || s3Info.SecretKey == "" {
			return errors.New("s3Info not enougth")
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
