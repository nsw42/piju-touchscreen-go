package screenblankmgr

import (
	"fmt"
	"os/exec"
	"strconv"
)

// Functions to call xset

func blankScreenNow() {
	runXset("activate")
}

func setTimeout(timeout int) {
	runXset(strconv.Itoa(timeout))
}

func runXset(s_arg string) {
	fmt.Println("xset", "s", s_arg)
	cmd := exec.Command("xset", "s", s_arg)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running xset: ", err)
	}
}
