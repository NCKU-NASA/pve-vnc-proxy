package config

import (
//    "os"
    "log"
    "github.com/spf13/viper"
//    "github.com/joho/godotenv"
)

func init() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    err := viper.ReadInConfig()
    if err != nil {
        log.Panicln("Error loading .env file")
    }
}
