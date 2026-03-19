package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "TestPassword123!"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Printf("Plain: %s\n", password)
	fmt.Printf("Hash : %s\n", string(hashed))
	
	err := bcrypt.CompareHashAndPassword(hashed, []byte(password))
	if err == nil {
		fmt.Println("Match: YES")
	} else {
		fmt.Println("Match: NO")
	}
}
