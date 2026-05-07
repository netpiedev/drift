package main

import "fmt"

func main() {
	fmt.Println(`{"sql":["ALTER TABLE users ADD COLUMN IF NOT EXISTS note TEXT;"]}`)
}
