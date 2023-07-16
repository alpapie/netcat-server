package lib

import (
	"fmt"
	"os"
)

func Errorstr(err error, er string) {
	if err != nil {
		fmt.Println("\033[31m", err, "\033[0m")
		os.Exit(0)
	} else if er != "" {
		fmt.Println("\033[31m", er, "\033[0m")
		os.Exit(0)
	}
}


func GetString(tab []byte) string {
	s := ""
	for _, v := range tab {
		if v == 0 {
			break
		}
		s += string(v)
	}
	return s
}