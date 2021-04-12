package cron

import (
	"fmt"
	"github.com/open-falcon/agent/g"
	"github.com/open-falcon/common/model"
	"log"
	"time"
)

/**
 * 上报Agent的状态
 */
func ReportAgentStatus() {
	/*打开了心跳上报，并且有记录服务器地址*/
	if g.Config().Heartbeat.Enabled && g.Config().Heartbeat.Addr != "" {
		/*以协程形式，上报Agent状态*/
		go reportAgentStatus(time.Duration(g.Config().Heartbeat.Interval) * time.Second)
	}
}

/**
 * 上报Agent状态
 */
func reportAgentStatus(interval time.Duration) {
	/**
	 * 无限循环
	 */
	for {
		/*取得主机名*/
		hostname, err := g.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}
		/*model是一个包名，里面有一个结构体，叫AgentReportRequest，进行一个上报*/
		req := model.AgentReportRequest{
			Hostname:      hostname,     /*主机名*/
			IP:            g.IP(),       /*IP*/
			AgentVersion:  g.VERSION,    /*版本*/
			PluginVersion: g.GetCurrPluginVersion(),  /**/
		}

		/*Http请求上报，并获得响应，存储的resp中*/
		var resp model.SimpleRpcResponse
		err = g.HbsClient.Call("Agent.ReportStatus", req, &resp)
		if err != nil || resp.Code != 0 {
			log.Println("call Agent.ReportStatus fail:", err, "Request:", req, "Response:", resp)
		}

		/*睡眠一段时间，就是函数的循环间隔。*/
		time.Sleep(interval)
	}
}
