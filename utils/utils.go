package utils

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

func PrintError(err error) {
	fmt.Printf("%s: "+err.Error()+"\n", color.RedString("Error"))
	os.Exit(0)
}

func PrintLog(logString, replaceStr string) {
	fmt.Printf(logString+"\n", color.RedString(replaceStr))
}

