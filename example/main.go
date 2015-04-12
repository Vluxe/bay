package main

import (
	"fmt"

	"github.com/Vluxe/bay"
)

func main() {
	id, err := bay.BuildWithGitRepo("https://github.com/maverickames/test_repo")
	if err != nil {
		fmt.Println("horse factory of sadness:", err)
	}
	fmt.Println("hi there!", id)
}
