package middlewares

func CheckLogin(login, password string) bool {
	if login == "" || password == "" || len(login) < 3 || len(password) < 8 {
		return false
	}

	return true
}
