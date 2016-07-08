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
		CPU     int    `yaml:"cpu"`
		Port    string `yaml:"port"`
		Mode    string `yaml:"mode"`
		LogFile string `yaml:"log_file"`
	}
	mysql struct {
		URL      string `yaml:"url"`
		Schema   string `yaml:"schema"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		IdleConn int    `yaml:"idle_conn"`
		MaxConn  int    `yaml:"max_conn"`
	}
)

func TestBind(t *testing.T) {
	// folder fileName from command line etc...
	s3Info := NewS3Info("ap-northeast-1", "bucket", "folder", "config.yml", 10*time.Second)
	//s3Info := NewS3InfoKey("ap-northeast-1", "access_key", "access_secret_key", "bucket", "folder", "fileName", 10*time.Second)

	myConfig := &Config{}

	// bind config
	if err := Bind(s3Info, myConfig); err != nil || myConfig.Server == nil || myConfig.MySQL == nil {
		t.Errorf("error TestBind %s", err.Error())
		return
	}

	// log
	t.Logf("Server %+v\n", myConfig.Server)
	t.Logf("MySQL %+v\n", myConfig.MySQL)

	// edit now s3 file
	time.Sleep(30 * time.Second)

	t.Logf("Server %+v\n", myConfig.Server)
	t.Logf("MySQL %+v\n", myConfig.MySQL)
}
