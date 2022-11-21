package main

import (
	"fmt"
	"os"
	"strconv"
)

func requiredEnv(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("you must define %s env var", name)
	}
	return val, nil
}

func getPort() (int, error) {
	portStr, err := requiredEnv("PORT")
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("strconv.Atoi failed; %w", err)
	}
	return port, nil
}
