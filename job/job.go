package job

import (
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

type JobItem struct {
	guid         string
	fn           func() //调用函数
	actionTime   int64  //执行时间
	intervalTime int64  //间隔时间
	execNum      int    //执行次数
}

type Cron struct {
	sync.Mutex
	cronInterface
	running bool
	entries []*JobItem
}

var Crontab = newCrontab()

func newCrontab() *Cron {
	this := &Cron{
		running: true,
		entries: make([]*JobItem, 10),
	}
	this.exec()
	return this
}

type cronInterface interface {
	AddCronTask(_time int64, execNum int, handler func())
}

//加入工作列表
// t = 秒
func (this *Cron) AddCronTask(_time int64, execNum int, fn_ func()) {
	this.Lock()
	defer this.Unlock()
	job := &JobItem{
		guid:         uuid.NewV4().String(),
		fn:           fn_,
		intervalTime: _time,
		actionTime:   time.Now().Unix(),
		execNum:      execNum,
	}
	this.entries = append(this.entries, job)
}

func (this *Cron) exec() {
	timer := time.NewTicker(1 * time.Second)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				for k := 0; k < len(this.entries); k++ {
					curTime := time.Now().Unix()
					if (this.entries[k].actionTime + this.entries[k].intervalTime) <= curTime {
						go this.entries[k].fn()
						this.entries[k].actionTime = curTime
						if this.entries[k].execNum != -1 {
							this.entries[k].execNum--
						}
						if this.entries[k].execNum == 0 {
							this.Lock()
							this.entries = append(this.entries[:k], this.entries[k+1:]...)
							this.Unlock()
						}
					}
				}
			}
		}
	}()
}
