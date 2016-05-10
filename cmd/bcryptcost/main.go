package main

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := []byte("password")
	fmt.Println("You should choose a cost number which costs 200ms~500ms hash time:")
	for cost := bcrypt.MinCost; cost <= bcrypt.MaxCost; cost++ {
		now := time.Now()
		bcrypt.GenerateFromPassword(password, cost)
		timeElasped := time.Since(now)
		fmt.Println(cost, timeElasped)
		if timeElasped > time.Second*10 {
			fmt.Println("Skip slower ones.")
			break
		}
	}
}
