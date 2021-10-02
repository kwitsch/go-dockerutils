package dockerutils

import "fmt"

func init() {
	fmt.Println("No process reaper initialization under windows")
	fmt.Println("---------------------")
}
