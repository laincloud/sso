package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var (
	cost = flag.Int("cost", bcrypt.DefaultCost, "bcrypt cost")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Generate bcrypt hash.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "\t%s [options] {password}\n", os.Args[0])
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	password := []byte(flag.Arg(0))
	hash, err := bcrypt.GenerateFromPassword(password, *cost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(hash))
}
