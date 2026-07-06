package main

import (
	"os"

	"github.com/onexstack/shop-app/cmd/shop-apiserver/app"
)

// @title           Shop App API
// @version         1.0
// @description     基于 OneX 技术栈的电商后端服务，提供用户登录注册等功能。
// @host            localhost:5555
// @BasePath        /api
// @schemes         http
// @license.name    MIT
// @contact.name    mungdong

// The default entry point of a Go program. Serves as the starting point
// for reading the project code.
func main() {
	command := app.NewWebServerCommand()

	// Execute the command and handle errors.
	if err := command.Execute(); err != nil {
		// Exit the program if an error occurs.
		// Return an exit code so that other programs (e.g., bash scripts)
		// can determine the service status based on the exit code.
		os.Exit(1)
	}
}
