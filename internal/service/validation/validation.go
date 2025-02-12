package validation

import "regexp"

func ValidateLogin(login string) bool {
	re := regexp.MustCompile(`^[A-Za-zА-Яа-яЁё0-9][A-Za-zА-Яа-яЁё0-9-_.!@#$%^&*()+=-]{3,20}[A-Za-zА-Яа-яЁё0-9]$`)
	return re.MatchString(login)
}

func ValidatePassword(password string) bool {
	re := regexp.MustCompile(`^[a-zA-ZА-Яа-яЁё0-9!@#$%^&*()_+=-]{8,16}$`)
	return re.MatchString(password)
}
