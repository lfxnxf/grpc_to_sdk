package main

import (
	"flag"
	"protbuf_to_sdk/general"
)

var g *general.General

func init() {
	//解析参数
	in := flag.String("in", "", "protobuf file")
	outModelName := flag.String("model", "", "out model file")
	outSdkName := flag.String("sdk", "", "out sdk file")
	flag.Parse()
	g = general.New(*in, *outModelName, *outSdkName)
}

func main() {
	g.Start()
}
