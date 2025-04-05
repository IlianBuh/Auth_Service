package main

import (
	"Service/internal/config"
	"fmt"
)

func main() {
	cfg := config.New()

	fmt.Printf("%+v\n", cfg)

	// TODO: init logger

	// TODO: init application

	// TODO: start application
}
