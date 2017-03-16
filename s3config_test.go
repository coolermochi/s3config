package s3config

import (
	"testing"
	"time"
)

type (
	Config struct {
		Server *server `yaml:"server"`
		MySQL  *mysql  `yaml:"mysql"`
	}
	server struct {
		Port string `yaml:"port"`
	}
	mysql struct {
		URL    string `yaml:"url"`
		Schema string `yaml:"schema"`
		User   string `yaml:"user"`
		Pass   string `yaml:"pass"`
	}
)

func TestBind(t *testing.T) {
	// type role
	//s3Info, err := New(TypeRole, "ap-northeast-1", &S3Bucket{"bucket", "folder", "sample.yml"})
	// type env
	//s3Info, err := New(TypeEnv, "ap-northeast-1", &S3Bucket{"bucket", "folder", "sample.yml"})
	// type key
	s3Info, err := New(TypeKey,
		"ap-northeast-1",
		&S3Bucket{"bucket", "folder", "sample.yml"},
		Keys("accessKey", "secretKey"),
		Interval(10*time.Minute),
	)
	if err != nil {
		t.Fatalf("error Create S3Info %s", err.Error())
		return
	}

	config := &Config{}

	// bind config
	if err := Bind(s3Info, config); err != nil {
		t.Fatalf("error TestBind %s", err.Error())
		return
	}

	// log
	t.Logf("Server %+v\n", config.Server)
	t.Logf("MySQL %+v\n", config.MySQL)
}
