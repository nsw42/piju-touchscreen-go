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

func runXset(sArg string) {
	fmt.Println("xset", "s", sArg)
	cmd := exec.Command("xset", "s", sArg)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running xset: ", err)
	}
}
