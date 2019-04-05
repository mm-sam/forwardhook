// config
package main

import (
	"encoding/json"
	"os"
	"runtime"
	"path/filepath"
)

func getLocalConfigPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

type Mapping struct {
	Path  string   `json:"path"`
	Sites []string `json:"sites"`
}

type Config struct {
	saveFile string
	Listen   string `json:"listen"`
	MaxRetries  int `json:"retries"`
	Mappings []Mapping `json:"mappings"`
}

func NewConfig(filename string) (err error, c *Config) {
	c = &Config{}
	c.saveFile = filename
	err = c.load(filename)
	return
}

func InitConfig(savePath string) *Config {
	return &Config{
		saveFile: savePath,
		Listen: "127.0.0.1:8899",
		MaxRetries: 10,
		Mappings: []Mapping{
			Mapping{
				Path: "/",
				Sites: []string{
					"http://www.google.com/",
				},
			},
		},
	}
}

func (c *Config) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		logger.Error.Println(err)
	}
	return err
}

func (c *Config) Export() string {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		logger.Error.Println(err)
		return ""
	}
	return string(data)
}

func (c *Config) Save() error {
	file, err := os.Create(c.saveFile)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	defer file.Close()
	data, err2 := json.MarshalIndent(c, "", "    ")
	if err2 != nil {
		logger.Error.Println(err2)
		return err2
	}
	_, err3 := file.Write(data)
	if err3 != nil {
		logger.Error.Println(err3)
	}
	return err3
}
