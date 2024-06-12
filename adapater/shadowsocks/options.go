package shadowsocks

import "github.com/juju/ratelimit"

type SSOptionHandler func(*SSOption)

type SSOption struct {
	EnableTrafficControl bool
	TxBucket             *ratelimit.Bucket
	RxBucket             *ratelimit.Bucket
}

func withTrafficControl(txBucket, rxBucket *ratelimit.Bucket) SSOptionHandler {
	return func(opt *SSOption) {
		opt.EnableTrafficControl = true
		opt.TxBucket = txBucket
		opt.RxBucket = rxBucket
	}
}
