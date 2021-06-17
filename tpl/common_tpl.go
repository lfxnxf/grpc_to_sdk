package tpl

const CommonTpl = `
package rpc_common

import (
	"sync"
)

var rpcClientMap sync.Map

func GetService(serviceName string, f func() interface{}) interface{} {
	client, _ := rpcClientMap.LoadOrStore(serviceName, f())
	return client
}
`