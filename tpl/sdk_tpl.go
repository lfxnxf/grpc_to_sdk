package tpl

const SdkTpl = `package {{sdk}}

import (
	//todo 替换为自己的地址
	"git.inke.cn/coolive/socialgame/social/server/commen/commen.project.base/rpc_sdk/pb/{{pk}}"
	"git.inke.cn/coolive/socialgame/social/server/commen/commen.project.base/rpc_sdk/rpc_common"
	"git.inke.cn/inkelogic/daenerys"
	"git.inke.cn/nvwa/server/commlib/nvwa_errors"
	"golang.org/x/net/context"
)

const (
	//todo 替换为config中配置的服务名称
	{{service}} = ""
)

func Get{{service}}() {{pk}}.{{service}} {
	client := rpc_common.GetService({{service}}, func() interface{} {
		return {{pk}}.New{{service}}(daenerys.RPCFactory(context.TODO(), {{service}}))
	})
	return client.({{pk}}.{{service}})
}
`

const SdkFuncTpl = `
//{{annotation}}
func {{func}}(ctx context.Context, in {{model}}.{{InModel}}) ({{out}}err error) {
	req := {{req_data}}
	resp, err := Get{{service}}.{{func}}(ctx, req)
	if err != nil {
		return {{res_data}}err
	}
	if resp.GetDmError() != nil {
		err = nvwa_errors.AddError(resp.GetDmError(), resp.GetErrorMsg())
		return {{res_data}}err
	}
	{{out_data}}
	return {{return_out_name}}nil
}
`
