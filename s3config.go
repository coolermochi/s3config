// Package s3config .
package s3config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v2"
)

type Type int

// Types.
const (
	TypeRole Type = iota + 1
	TypeEnv
	TypeKey
)

// minimum interval.
const (
	minInterval = 30 * time.Second
)

// S3Info(aws properties).
type S3Info struct {
	Type   Type
	Region *string
	*Bucket
	AccessKey string
	SecretKey string
	Interval  time.Duration
}

type Bucket struct {
	Name   string
	Folder string
	File   string
}

type Option func(*S3Info) error

// New.
func New(typo Type, region string, bucket *Bucket, opts ...Option) (*S3Info, error) {

	if bucket == nil || bucket.Name == "" || bucket.File == "" {
		return nil, errors.New("Empty Bucket data")
	}

	s3Info := &S3Info{
		Type:   typo,
		Region: aws.String(region),
		Bucket: &Bucket{
			Name:   bucket.Name,
			Folder: bucket.Folder,
			File:   bucket.File,
		},
	}

	for _, opt := range opts {
		if err := opt(s3Info); err != nil {
			return nil, err
		}
	}
	if s3Info.Type == TypeKey && (s3Info.AccessKey == "" || s3Info.SecretKey == "") {
		return nil, errors.New("Empty data")
	}

	return s3Info, nil
}

func Keys(accessKey string, secretKey string) Option {
	return func(s3Info *S3Info) error {
		if accessKey == "" || secretKey == "" {
			return errors.New("Empty Keys")
		}
		s3Info.AccessKey = accessKey
		s3Info.SecretKey = secretKey
		return nil
	}
}

func Interval(interval time.Duration) Option {
	return func(s3Info *S3Info) error {
		if interval > minInterval {
			s3Info.Interval = interval
		} else {
			s3Info.Interval = minInterval
		}
		return nil
	}
}

// Bind.
// config bind yaml from s3.
func Bind(s3Info *S3Info, config interface{}) error {

	if config == nil {
		return errors.New("config is nil")
	}

	// create aws credentials
	var creds *credentials.Credentials
	switch s3Info.Type {
	case TypeRole:
		ec2 := ec2metadata.New(session.New(), &aws.Config{
			HTTPClient: &http.Client{Timeout: 30 * time.Second},
		})
		creds = credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{
			Client: ec2,
		})
	case TypeEnv:
		creds = credentials.NewEnvCredentials()
	case TypeKey:
		creds = credentials.NewStaticCredentials(s3Info.AccessKey, s3Info.SecretKey, "")
	}

	// create S3 service
	sess, err := session.NewSession(&aws.Config{
		Region:      s3Info.Region,
		Credentials: creds,
	})
	if err != nil {
		return err
	}

	// create svc
	s3Svc := s3.New(sess)

	// load file
	if err := loadFile(s3Svc, s3Info.Name, s3Info.Folder, s3Info.File, config); err != nil {
		return err
	}

	// reload
	go func() {
		for {
			// interval
			time.Sleep(s3Info.Interval)
			if err := loadFile(s3Svc, s3Info.Name, s3Info.Folder, s3Info.File, config); err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	return nil
}

// LoadFile.
// load yaml file bind config.
func loadFile(s3Svc *s3.S3, bucket string, folder string, file string, config interface{}) error {

	// get file
	var key string
	if folder != "" {
		key = folder + "/" + file // full fileName
	} else {
		key = file
	}
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
