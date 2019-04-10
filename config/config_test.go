package config

import (
	"testing"
)

func TestGenerateAndLoadConfig(t *testing.T) {
	testconf := "./config.json"
	//defer os.Remove(testconf)
	if err := GenerateConfig(testconf); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(testconf); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateConfigFailure(t *testing.T) {
	testConf := "/root/toor/config.json"
	if err := GenerateConfig(testConf); err == nil {
		t.Fatal("error expected")
	}
}

func TestLoadConfigFailure(t *testing.T) {
	testFileExists := "./README.md"
	if _, err := LoadConfig(testFileExists); err == nil {
		t.Fatal("error expected")
	}
	testFileNotExists := "/root/toor/config.json"
	if _, err := LoadConfig(testFileNotExists); err == nil {
		t.Fatal("error expected")
	}
}
