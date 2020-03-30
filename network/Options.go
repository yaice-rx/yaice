package network

type IOptions interface {
	GetMaxRetires() int32
	GetMaxConnCount() int32
	CallBackFunc() func(interface{})
}

type Options struct {
	maxRetries   int32
	maxConnCount int32
	callBack     func(data interface{})
}

func WithMax(maxRetries int32, maxConnCount int32, callBackFunc func(interface{})) IOptions {
	return &Options{
		maxRetries:   maxRetries,
		maxConnCount: maxConnCount,
		callBack:     callBackFunc,
	}
}

func (o *Options) GetMaxRetires() int32 {
	return o.maxRetries
}

func (o *Options) GetMaxConnCount() int32 {
	return o.maxConnCount
}

func (o *Options) CallBackFunc() func(interface{}) {
	return o.callBack
}
