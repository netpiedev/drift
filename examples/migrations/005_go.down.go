package main

import "fmt"

func main() {
	fmt.Println(`{"sql":["ALTER TABLE users DROP COLUMN IF EXISTS note;"]}`)
}
