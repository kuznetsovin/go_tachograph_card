package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Settings struct {
	Server    string
	Port      string
	Queue     string
	UploadDir string
	Log       string
}

func (settings *Settings) init() error {
	setting_json, err := ioutil.ReadFile("setting.json")
	if err != nil {
		return err
	}
	if err = json.Unmarshal(setting_json, settings); err != nil {
		return err
	}

	if err := os.MkdirAll(settings.UploadDir, os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Stat(settings.Log); os.IsNotExist(err) {
		path := filepath.Dir(settings.Log)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
		f, err := os.Create(settings.Log)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return err
}
