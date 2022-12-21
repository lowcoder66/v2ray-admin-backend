package service

import (
	"context"
	"fmt"
	proxyCmd "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	statsCmd "github.com/v2fly/v2ray-core/v4/app/stats/command"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/vmess"
	"google.golang.org/grpc"
	"log"
	"strings"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/model"
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

func QueryUserTraffic(email string, reset bool) (uint64, uint64) {
	resp, err := statsClient.QueryStats(context.Background(), &statsCmd.QueryStatsRequest{
		Pattern: fmt.Sprintf("user>>>%s>>>traffic>>>", email),
		Reset_:  reset,
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	}

	stat := resp.GetStat()

	up, down := uint64(0), uint64(0)
	for _, e := range stat {
		if strings.HasSuffix(e.Name, ">>>uplink") {
			up = uint64(e.Value)
		}
		if strings.HasSuffix(e.Name, ">>>downlink") {
			down = uint64(e.Value)
		}
	}

	return up, down
}

func QueryGlobalTraffic(reset bool, tag string) (uint64, uint64) {
	queryPattern := "inbound>>>"
	if &tag != nil && tag != "" {
		queryPattern += tag + ">>>"
	}

	resp, err := statsClient.QueryStats(context.Background(), &statsCmd.QueryStatsRequest{
		Pattern: queryPattern,
		Reset_:  reset,
	})
	if err != nil {
		log.Printf("failed to call grpc command: %v", err)
	}

	stat := resp.GetStat()

	up, down := uint64(0), uint64(0)
	for _, e := range stat {
		if strings.HasSuffix(e.Name, ">>>uplink") {
			up = uint64(e.Value)
		}
		if strings.HasSuffix(e.Name, ">>>downlink") {
			down = uint64(e.Value)
		}
	}

	return up, down
}
