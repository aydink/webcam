package main

import "strings"

// Users hold an in memory user database
var Users = []User{
	User{"aydink", "Aydın KILIÇ", "password", false, true, false},	
	User{"user1", "John DOE", "password", false, false, false},
}

type User struct {
	Username   string
	FullName   string
	Password   string
	IsDisabled bool
	IsAdmin    bool
	IsLoggedin bool
}

// FindUser Girilen kullanıcı adı şifresine uygun bir kullanıcı varsa o kulanıcıyı,
// yok ise boş bir kullanıcı ve hata döndürür.
func FindUser(username, password string) (User, bool) {
	for _, user := range Users {
		if strings.Compare(user.Username, username) == 0 && strings.Compare(user.Password, password) == 0 {
			return user, true
		}
	}

	return User{}, false
}
