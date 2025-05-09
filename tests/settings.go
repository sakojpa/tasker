package tests

import (
	"fmt"
	"os"
	"strconv"
)

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = false
var Search = false
var Token = ``

func init() {
	if envVal := os.Getenv("TODO_FULLNEXTDATE"); len(envVal) != 0 {
		FullNextDate, _ = strconv.ParseBool(envVal)
	}
	if envVal := os.Getenv("TODO_SEARCH"); len(envVal) != 0 {
		Search, _ = strconv.ParseBool(envVal)
	}
	if envVal := os.Getenv("TODO_TOKEN"); len(envVal) != 0 {
		Token = envVal
	}

	fmt.Printf(
		"Running tests with FullNextDate=%t, Search=%t, Auth=%t\n", FullNextDate, Search, len(Token) > 0,
	)
}
