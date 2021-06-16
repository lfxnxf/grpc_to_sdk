package general

import (
	"fmt"
	"io/ioutil"
	"os"
	"protbuf_to_sdk/utils"
	"strings"
)

type General struct {
	OutModel string   `json:"out_model"`
	OutSdk   string   `json:"out_sdk"`
	Content  []string `json:"content"`
}

func New(filePath, outModel, outSdk string) *General {
	file, err := os.Open(filePath)
	if err != nil {
		utils.PrintError(err)
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		utils.PrintError(err)
	}
	return &General{
		OutModel: outModel,
		OutSdk:   outSdk,
		Content:  strings.Split(string(bytes), "\n"),
	}
}

func (g *General) Start() {
	fmt.Println(g.Content)
}
