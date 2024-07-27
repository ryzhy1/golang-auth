package middlewares

func CheckRegister(login, email, password string) bool {
	if login == "" || email == "" || password == "" || CorrectEmailChecker(email) == false || len(login) < 3 || len(password) < 8 {
		return false
	}
	return true
}
