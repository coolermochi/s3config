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
		Port    string `yaml:"port"`
	}
	mysql struct {
		URL      string `yaml:"url"`
		Schema   string `yaml:"schema"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
	}
)

func TestBind(t *testing.T) {
	// folder fileName from command line etc...
	s3Info := NewS3InfoRole("ap-northeast-1", "bucket", "folder", "sample.yml", 10*time.Minute)
	//s3Info := NewS3InfoKey("ap-northeast-1", "AWS_ACCESS_KEY", "AWS_SECRET_KEY", "bucket", "folder", "fileName", 10*time.Minute)

	config := &Config{}

	// bind config
	if err := Bind(s3Info, config); err != nil || config.Server == nil || config.MySQL == nil {
		t.Errorf("error TestBind %s", err.Error())
		return
	}

	// log
	t.Logf("Server %+v\n", config.Server)
	t.Logf("MySQL %+v\n", config.MySQL)
}
