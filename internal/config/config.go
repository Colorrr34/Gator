package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct{
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read()Config{
	configPath := getConfigPath()
	configJson,err := os.ReadFile(configPath)
	if err!= nil{
		log.Fatal(err)
	} 
	
	var config Config
	errM := json.Unmarshal(configJson,&config)
	if errM != nil{
		log.Fatal(errM)
	}
	if config == (Config{}){
		log.Fatal("nil Config", configJson)
	}
	return config
}

func (c *Config) SetUser(username string){
	(*c).CurrentUserName = username
	WriteConfig(*c)
}

func WriteConfig(c Config){
	configPath := getConfigPath()
	configJson,err := json.Marshal(c)
	if err != nil{
		log.Fatal(err)
	}
	os.WriteFile(configPath,configJson,os.ModePerm)
}

func getConfigPath()string{
	homePath,err := os.UserHomeDir()
	if err != nil{
		log.Fatal(err)
	}
	path := filepath.Join(homePath,".gatorconfig.json")
	return path
}