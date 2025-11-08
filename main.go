package main

import (
	"vcs.technonext.com/carrybee/ride_engine/cmd"
)

// @title Ride Engine API
// @version 1.0
// @description This is a comprehensive Ride Engine API server for managing customers, drivers, and rides.
// @description The API provides endpoints for user registration, authentication, location tracking, and ride management.

// @contact.name Mohammad Kaium
// @contact.email mohammadkaiom79@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// start root command

	cmd.Execute()
}
