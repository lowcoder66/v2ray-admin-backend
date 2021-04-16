package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/model"
	proxyCmd "v2ray.com/core/app/proxyman/command"
	statsCmd "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

var handlerClient proxyCmd.HandlerServiceClient
var statsClient statsCmd.StatsServiceClient
var tag string

func init() {
	log.Println("初始化v2ray远程连接...")

	c := conf.App.V2ray
	tag = c.Tag

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	handlerClient = proxyCmd.NewHandlerServiceClient(cmdConn)
	statsClient = statsCmd.NewStatsServiceClient(cmdConn)

	log.Println("v2ray远程连接初始化完成")
}

func AddUser(user *model.User) error {
	resp, err := handlerClient.AlterInbound(context.Background(), &proxyCmd.AlterInboundRequest{
		Tag: tag,
		Operation: serial.ToTypedMessage(&proxyCmd.AddUserOperation{
			User: &protocol.User{
				Level: user.Level,
				Email: user.Email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               user.UId,
					AlterId:          user.AlterId,
					SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
				}),
			},
		}),
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		log.Printf("ok: %v", resp)
	}

	return err
}

func RemoveUser(user *model.User) error {
	resp, err := handlerClient.AlterInbound(context.Background(), &proxyCmd.AlterInboundRequest{
		Tag: tag,
		Operation: serial.ToTypedMessage(&proxyCmd.RemoveUserOperation{
			Email: user.Email,
		}),
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	} else {
		log.Printf("ok: %v", resp)
	}

	return err
}

func QueryUserTraffic(email string) {
	resp, err := statsClient.QueryStats(context.Background(), &statsCmd.QueryStatsRequest{
		Pattern: fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email),
		Reset_:  false, // 查询完成后是否重置流量
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	}

	stat := resp.GetStat()
	// 返回的是一个数组，对其进行遍历输出
	for _, e := range stat {
		fmt.Println(e)
	}
}
