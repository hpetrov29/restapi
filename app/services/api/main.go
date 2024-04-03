package main

import (
	"github.com/hpetrov29/restapi/app/services/api/v1/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Main(cmd.Routes())
}