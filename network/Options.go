package network

type IOptions interface {
	GetMaxRetires() int32
}

type Options struct {
	maxRetries int32
}

func WithMax(maxRetries int32) IOptions {
	return &Options{
		maxRetries: maxRetries,
	}
}

func (o *Options) GetMaxRetires() int32 {
	return o.maxRetries
}
