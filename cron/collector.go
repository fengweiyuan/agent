package cron

import (
	"time"

	"github.com/open-falcon/agent/funcs"
	"github.com/open-falcon/agent/g"
	"github.com/open-falcon/common/model"
)

/**
 * 秒级收集cpu与disk统计
 */
func InitDataHistory() {
	for {
		/*更新cpu统计信息，上一次的信息会保留*/
		funcs.UpdateCpuStat()
		/*更新Disk统计信息*/
		funcs.UpdateDiskStats()
		/*睡眠1秒*/
		time.Sleep(g.COLLECT_INTERVAL)
	}
}

/**
 * 收集
 */
func Collect() {

	/**
	 * 如果没有开启Transfer，那么直接返回
	 */
	if !g.Config().Transfer.Enabled {
		return
	}

	/**
	 * 如果没有定义Transfer地址，也直接返回
	 */
	if len(g.Config().Transfer.Addrs) == 0 {
		return
	}

	/**
	 *
	 */
	for _, v := range funcs.Mappers {
		go collect(int64(v.Interval), v.Fs)
	}
}

/**
 *
 */
func collect(sec int64, fns []func() []*model.MetricValue) {
	t := time.NewTicker(time.Second * time.Duration(sec)).C
	for {
		<-t

		hostname, err := g.Hostname()
		if err != nil {
			continue
		}

		mvs := []*model.MetricValue{}
		ignoreMetrics := g.Config().IgnoreMetrics

		for _, fn := range fns {
			items := fn()
			if items == nil {
				continue
			}

			if len(items) == 0 {
				continue
			}

			for _, mv := range items {
				if b, ok := ignoreMetrics[mv.Metric]; ok && b {
					continue
				} else {
					mvs = append(mvs, mv)
				}
			}
		}

		now := time.Now().Unix()
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = sec
			mvs[j].Endpoint = hostname
			mvs[j].Timestamp = now
		}

		g.SendToTransfer(mvs)

	}
}
