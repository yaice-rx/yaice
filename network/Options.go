package network

import "sync/atomic"

type ServeType int32

const (
	Serve_Client ServeType = 0
	Serve_Server ServeType = 1
)

type IOptions interface {
	GetMaxRetires() int32
	GetTimesTotal() *int64
	GetHttpCreateTime() int64
}

type Options struct {
	maxRetries int32
	timesTotal *int64
	httpCreateTime int64
}

func WithMax(maxRetries int32,timesTotal *int64,httpCreateTime int64) IOptions {
	return &Options{
		maxRetries: maxRetries,
		timesTotal:timesTotal,
		httpCreateTime:httpCreateTime,
	}
}

func (o *Options) GetMaxRetires() int32 {
	return o.maxRetries
}

func (o *Options) GetTimesTotal() *int64 {
	return o.timesTotal
}

func (o *Options) AddTimesTotal(times int64) int64 {
	return atomic.AddInt64(o.timesTotal,times)
}

func (o *Options) GetHttpCreateTime() int64 {
	return o.httpCreateTime
}
