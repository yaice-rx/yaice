package network

type IOptions interface {
	GetMaxRetires() uint32
	GetMaxConnCount() uint32
	CallBackFunc() func(interface{})
}

type Options struct {
	maxRetries   uint32
	maxConnCount uint32
	callBack     func(data interface{})
}

func WithMax(maxRetries uint32, callBackFunc func(interface{})) IOptions {
	return &Options{
		maxRetries: maxRetries,
		callBack:   callBackFunc,
	}
}

func (o *Options) GetMaxRetires() uint32 {
	return o.maxRetries
}

func (o *Options) GetMaxConnCount() uint32 {
	return o.maxConnCount
}

func (o *Options) CallBackFunc() func(interface{}) {
	return o.callBack
}
