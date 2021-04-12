package main

import (
    "flag"
    "fmt"
    "github.com/open-falcon/agent/cron"
    "github.com/open-falcon/agent/funcs" /* funcs本来就是项目带的，为啥要加github.com，看来这些古典项目就是这样 */
    "github.com/open-falcon/agent/g"
    "github.com/open-falcon/agent/http"
    "os"
)

func main() {

    /**
     * -c 指定配置文件，默认值是cfg.json
     * -v 打印版本
     * -check 检查收集器
     */
    cfg := flag.String("c", "cfg.json", "configuration file")
    version := flag.Bool("v", false, "show version")
    check := flag.Bool("check", false, "check collector")
    flag.Parse()

    /**
     * 只要有-v，就打印版本便退出，那怕有其他选项参数也不再生效
     */
    if *version {
        fmt.Println(g.VERSION)
        os.Exit(0)
    }

    /**
     * Exporter里面有一个模块叫Collector，专门负责收集系统信息，本函数就是检查它是否能顺利工作
     * 用户可以先看看，如果不顺利，说明这个程序不适合在当前的系统跑。
     * 比如在 Mac 下是不成功的。
     */
    if *check {
        funcs.CheckCollector()
        os.Exit(0)
    }

    /*解析配置文件的内容*/
    g.ParseConfig(*cfg)

    /*取得当前所处目录路径*/
    g.InitRootDir()
    /*取得当前主机ip*/
    g.InitLocalIp()
    /*还得初始化rpc的上报通道*/
    g.InitRpcClients()

    /*不懂*/
    funcs.BuildMappers()

    /*收集cpu与disk统计*/
    go cron.InitDataHistory()

    /*上报Agent状态*/
    cron.ReportAgentStatus()
    /**/
    cron.SyncMinePlugins()
    /*内置指标？*/
    cron.SyncBuiltinMetrics()
    cron.SyncTrustableIps()

    /*这个是主动收集并发送Metrics?*/
    cron.Collect()

    go http.Start()

    /*永久阻塞*/
    select {}

}
