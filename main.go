//main.go
package main

import (
	"github.com/dieehard/forex-be/lib"
	"os"
)

func main() {
	// Initialize our handler
	a := lib.AppHandler{}

	a.Initialize(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"))

	//a.Initialize(
	//	"root",
	//	"",
	//	"localhost",
	//	"3306",
	//	"exchange_rate")

	// Listen to port 8080
	a.Run(":8080")
}