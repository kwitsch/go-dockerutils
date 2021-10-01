package process

import (
	"fmt"
	reaper "github.com/ramr/go-reaper"
)

func init() {
	fmt.Println("Process reaper initialization")
	go reaper.Reap()
	fmt.Println("---------------------")
}
