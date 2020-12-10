package github.com/jamesBan/sensitive/rpc

import (
	hrpc "github.com/hprose/hprose-golang/rpc"
	rpc "github.com/hprose/hprose-golang/rpc/fasthttp"
	"github.com/valyala/fasthttp"
	"log"
	"reflect"
	"github.com/jamesBan/sensitive"
	"time"
)

func logInvokeHandler(name string, args []reflect.Value, context hrpc.Context, next hrpc.NextInvokeHandler) (results []reflect.Value, err error) {
	//log.Printf("%s(%v) = ", name, args)
	startTime := time.Now()
	results, err = next(name, args, context)
	log.Printf("%s(%v) %v ns\r\n", name, args, time.Now().Sub(startTime).Nanoseconds())
	return
}

func RpcServer(manager *sensitive.Manager, port string) {
	service := rpc.NewFastHTTPService()
	service.AddInvokeHandler(logInvokeHandler)

	service.AddFunction("add", manager.GetStore().Write)
	service.AddFunction("remove", manager.GetStore().Remove)
	service.AddFunction("filter", manager.GetFilter().Find)
	service.AddFunction("replace", manager.GetFilter().Replace)

	fasthttp.ListenAndServe(":"+port, service.ServeFastHTTP)
}
