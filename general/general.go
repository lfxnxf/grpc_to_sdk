package general

import (
	"errors"
	"fmt"
	"github.com/lfxnxf/protobuf_to_sdk/tpl"
	"github.com/lfxnxf/protobuf_to_sdk/utils"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type General struct {
	OutModel           string
	OutSdk             string
	Content            []string
	NeedCommon         int64
	ModelPackage       string
	SdkPackage         string
	wg                 *sync.WaitGroup
	modelList          []string
	sdkList            []string
	pk                 string
	modelMap           map[string]map[string]string
	sdkMap             map[string]map[string]string
	bracesStack        []string //message括号栈
	serviceBracesStack []string //service括号栈
	serviceName        string
}

func New(filePath, outModel, outSdk, modelPackage, sdkPackage string, needCommon int64) *General {
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
		OutModel:     outModel + ".go",
		OutSdk:       outSdk + ".go",
		Content:      strings.Split(string(bytes), "\n"),
		NeedCommon:   needCommon,
		ModelPackage: modelPackage,
		SdkPackage:   sdkPackage,
		wg:           new(sync.WaitGroup),
		modelMap:     make(map[string]map[string]string),
		sdkMap:       make(map[string]map[string]string),
	}
}

func (g *General) Start() {
	//解析数据生成List
	g.SetData()
	//生成 model
	g.GenModel()
	//生成 sdk
	g.GenSdk()
	//输出common
	if g.NeedCommon >= 0 {
		g.GenCommon()
	}
}

func (g *General) SetData() {
	var modelName string
	var i int
	for i = 0; i < len(g.Content); i++ {
		v := strings.TrimRight(strings.TrimSpace(g.Content[i]), ";")
		if IsEmpty(v) {
			continue
		}
		switch true {
		case IsPackage(v):
			s := strings.Split(v, " ")
			g.pk = s[1]
			break
		case IsMessage(v):
			g.PushBraces()
			if len(g.bracesStack) > 1 {
				newModelName := modelName + GetItemName(v)
				i = g.SetMessageChildData(i, newModelName)
				g.modelMap[modelName][newModelName] = newModelName
			} else {
				modelName = GetItemName(v)
				if _, ok := g.modelMap[modelName]; !ok {
					g.modelMap[modelName] = make(map[string]string)
				}
			}
			break
		case IsRightBraces(v) && !g.BracesOver():
			g.PopBraces()
			break
		case IsService(v):
			g.PushServiceBraces()
			g.serviceName = GetItemName(v)
			break
		case IsRightBraces(v) && !g.ServiceBracesOver():
			g.PopServiceBraces()
			break
		}
		//设置message to model
		s := strings.Split(v, " ")
		if !g.BracesOver() && !IsMessage(v) {
			if s[0] == "repeated" {
				g.modelMap[modelName][GetEqual(s[2])] = "[]" + s[1]
			} else {
				g.modelMap[modelName][GetEqual(s[1])] = s[0]
			}
		}

		//设置service to sdk
		if !g.ServiceBracesOver() && !IsService(v) {
			if _, ok := g.sdkMap[s[1]]; !ok {
				g.sdkMap[s[1]] = make(map[string]string)
			}
			g.sdkMap[s[1]]["in"] = TrimParentheses(s[2])
			g.sdkMap[s[1]]["out"] = TrimParentheses(s[4])
			annotation := strings.Split(v, "//")
			if len(annotation) >= 2 {
				g.sdkMap[s[1]]["annotation"] = annotation[1]
			}
		}
	}
}

func (g *General) SetMessageChildData(index int, modelName string) int {
	modelName = strings.Title(modelName)
	var i int
	contents := g.Content[index+1:]
	for i = 0; i < len(contents); i++ {
		v := strings.TrimSpace(contents[i])
		if IsMessage(v) {
			g.PushBraces()
			newModelName := modelName + strings.Title(GetItemName(v))
			i = g.SetMessageChildData(i, newModelName)
			if _, ok := g.modelMap[newModelName]; !ok {
				g.modelMap[newModelName] = make(map[string]string)
			}
			g.modelMap[modelName][newModelName] = newModelName
		} else if IsRightBraces(v) {
			g.PopBraces()
			break
		} else {
			//初始数据
			if _, ok := g.modelMap[modelName]; !ok {
				g.modelMap[modelName] = make(map[string]string)
			}
			s := strings.Split(strings.TrimSpace(v), " ")
			if s[0] == "repeated" {
				g.modelMap[modelName][GetEqual(s[2])] = "[]" + s[1]
			} else {
				g.modelMap[modelName][GetEqual(s[1])] = s[0]
			}
		}
	}
	return index + i
}

func (g *General) GenModel() {
	m := strings.ReplaceAll(tpl.ModelTpl, "{{model}}", g.ModelPackage)
	for k, v := range g.modelMap {
		str := fmt.Sprintf("\ntype %s struct{\n", k)
		for field, fieldType := range v {
			str += fmt.Sprintf("    %s %s `%s` \n", ToHump(field), fieldType, fmt.Sprintf("json:\"%s\"", field))
		}
		str += "}\n"
		m += str
	}
	WriteFile(g.OutModel, m)
}

func (g *General) GenSdk() {
	s := strings.ReplaceAll(tpl.SdkTpl, "{{sdk}}", g.SdkPackage)
	s = strings.ReplaceAll(s, "{{pk}}", g.pk)
	s = strings.ReplaceAll(s, "{{service}}", g.serviceName)

	for k, v := range g.sdkMap {
		if _, ok := g.modelMap[v["in"]]; !ok {
			utils.PrintError(errors.New(fmt.Sprintf("model error : %s", v["in"])))
		}

		if _, ok := g.modelMap[v["out"]]; !ok {
			utils.PrintError(errors.New(fmt.Sprintf("model error : %s", v["out"])))
		}

		funcTpl := strings.ReplaceAll(tpl.SdkFuncTpl, "{{pk}}", g.pk)
		funcTpl = strings.ReplaceAll(funcTpl, "{{func}}", k)
		funcTpl = strings.ReplaceAll(funcTpl, "{{model}}", g.ModelPackage)
		funcTpl = strings.ReplaceAll(funcTpl, "{{InModel}}", v["in"])
		funcTpl = strings.ReplaceAll(funcTpl, "{{OutModel}}", v["out"])
		funcTpl = strings.ReplaceAll(funcTpl, "{{service}}", g.serviceName)
		if _, ok := v["annotation"]; ok {
			funcTpl = strings.ReplaceAll(funcTpl, "{{annotation}}", v["annotation"])
		}
		//数据处理
		inData := fmt.Sprintf("&%s.%s{\n", g.pk, v["in"])
		for field, _ := range g.modelMap[v["in"]] {
			inData += fmt.Sprintf("        %s: %s,\n", ToHump(field), fmt.Sprintf("in.%s", ToHump(field)))
		}
		inData += "    }\n"
		funcTpl = strings.ReplaceAll(funcTpl, "{{req_data}}", inData)
		outData := fmt.Sprintf("var outData %s.%s\n", g.ModelPackage, v["out"])
		outData += "    bytesVal, _ := json.Marshal(resp)\n"
		outData += "    _ = json.Unmarshal(bytesVal, &outData)\n"
		respData := ""
		for field, _ := range g.modelMap[v["out"]] {
			if field == "dm_error" || field == "error_msg" {
				continue
			}
			respData = ToHump(field)
		}
		var out string
		var resReplaceData string
		var returnOutName string
		if respData != "" {
			outData += fmt.Sprintf("    out = outData.%s", respData)
			out = fmt.Sprintf("out %s.%s, ", g.ModelPackage, respData)
			resReplaceData = fmt.Sprintf("%s.%s{}, ", g.ModelPackage, respData)
			returnOutName = "out, "
		}

		funcTpl = strings.ReplaceAll(funcTpl, "{{in_data}}", inData)
		funcTpl = strings.ReplaceAll(funcTpl, "{{out_data}}", outData)
		funcTpl = strings.ReplaceAll(funcTpl, "{{out}}", out)
		funcTpl = strings.ReplaceAll(funcTpl, "{{res_data}}", resReplaceData)
		funcTpl = strings.ReplaceAll(funcTpl, "{{return_out_name}}", returnOutName)
		s += fmt.Sprintf("%s", funcTpl)
	}
	WriteFile(g.OutSdk, s)
}

func (g *General) GenCommon() {
	WriteFile("common.go", tpl.CommonTpl)
}

func (g *General) PushBraces() {
	g.bracesStack = append(g.bracesStack, "{")
}

func (g *General) PopBraces() {
	g.bracesStack = g.bracesStack[:len(g.bracesStack)-1]
}

func (g *General) BracesOver() bool {
	if len(g.bracesStack) > 0 {
		return false
	}
	return true
}

func (g *General) PushServiceBraces() {
	g.serviceBracesStack = append(g.serviceBracesStack, "{")
}

func (g *General) PopServiceBraces() {
	g.serviceBracesStack = g.serviceBracesStack[:len(g.serviceBracesStack)-1]
}

func (g *General) ServiceBracesOver() bool {
	if len(g.serviceBracesStack) > 0 {
		return false
	}
	return true
}

func IsPackage(msg string) bool {
	return strings.Contains(msg, "package")
}

func IsMessage(msg string) bool {
	return strings.Contains(msg, "message")
}

func IsLeftBraces(msg string) bool {
	return strings.Contains(msg, "{")
}

func IsService(msg string) bool {
	return strings.Contains(msg, "service")
}

func IsEmpty(msg string) bool {
	if msg == "" {
		return true
	}
	return false
}

func IsRightBraces(msg string) bool {
	return strings.Contains(msg, "}")
}

func GetItemName(msg string) string {
	s := strings.Split(msg, " ")
	return strings.Title(s[1])
}

func TrimParentheses(msg string) string {
	return strings.TrimRight(strings.TrimRight(strings.TrimLeft(msg, "("), ";"), ")")
}

func GetEqual(msg string) string {
	if strings.Contains(msg, "=") {
		s := strings.Split(msg, "=")
		return s[0]
	}
	return msg
}

func ToHump(msg string) string {
	s := strings.Split(msg, "_")
	for k, v := range s {
		s[k] = strings.Title(v)
	}
	return strings.Join(s, "")
}

func WriteFile(file, content string) {
	msg := []byte(content)
	err := ioutil.WriteFile(file, msg, 0644)
	if err != nil {
		utils.PrintError(err)
	}
}
