package cron

import (
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

type ICron interface {
	AddCronTask(_time int64, execNum int, handler func())
	Stop()
}

type cronItem struct {
	guid         string
	fn           func() //调用函数
	actionTime   int64  //执行时间
	intervalTime int64  //间隔时间
	execNum      int    //执行次数
}

type cron struct {
	running bool
	ticker  *time.Ticker
	entries []*cronItem
}

var CronMgr = _NewCron()

func _NewCron() ICron {
	this := &cron{
		running: true,
		ticker:  time.NewTicker(1 * time.Second),
		entries: make([]*cronItem, 10),
	}
	go this.exec()
	return this
}

//加入工作列表
// t = 秒
func (this *cron) AddCronTask(_time int64, execNum int, fn_ func()) {
	var _sync sync.Mutex
	_sync.Lock()
	defer _sync.Unlock()
	job := &cronItem{
		guid:         uuid.NewV4().String(),
		fn:           fn_,
		intervalTime: _time,
		actionTime:   time.Now().Unix(),
		execNum:      execNum,
	}
	this.entries = append(this.entries, job)
}

func (this *cron) Stop() {
	this.ticker.Stop()
}

func (this *cron) exec() {
	for {
		select {
		case <-this.ticker.C:
			for k := 0; k < len(this.entries); k++ {
				if this.entries[k] == nil {
					continue
				}
				curTime := time.Now().Unix()
				if (this.entries[k].actionTime + this.entries[k].intervalTime) <= curTime {
					go this.entries[k].fn()
					this.entries[k].actionTime = curTime
					if this.entries[k].execNum != -1 {
						this.entries[k].execNum--
					}
					if this.entries[k].execNum == 0 {
						this.entries = append(this.entries[:k], this.entries[k+1:]...)
					}
				}
			}
		}
	}
}
