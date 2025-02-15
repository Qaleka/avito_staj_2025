package dsn

import (
	"fmt"
	"os"
)

// FromEnv собирает DSN строку из переменных окружения
func FromEnv() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}

func FromEnvE2E() string {
	host := os.Getenv("DB_HOST_TEST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT_TEST")
	user := os.Getenv("DB_USER_TEST")
	pass := os.Getenv("DB_PASS_TEST")
	dbname := os.Getenv("DB_NAME_TEST")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
