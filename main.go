package main

import (
	"flag"
	"github.com/lfxnxf/protobuf_to_sdk/general"
)

var g *general.General

func init() {
	//解析参数

	//protobuf文件
	in := flag.String("in", "", "protobuf file")
	//输出模型文件名称
	outModelName := flag.String("om", "model", "out model file")
	//输出sdk文件名称
	outSdkName := flag.String("os", "sdk", "out sdk file")
	//模型package
	modelPackage := flag.String("mp", "model", "model package")
	//sdk package
	sdkPackage := flag.String("sp", "sdk", "sdk package")
	//是否需要生成common
	needCommon := flag.Int64("common", 1, "是否需要生成common")
	flag.Parse()
	g = general.New(*in, *outModelName, *outSdkName, *modelPackage, *sdkPackage, *needCommon)
}

func main() {
	g.Start()
}
