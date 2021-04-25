package common

import "errors"

var (
	ERR_LOCK_ALREAY_REQUIRED = errors.New("锁已占用")
)
