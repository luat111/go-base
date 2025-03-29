package common

import "time"

const (
	CheckPortTimeout time.Duration = 2 * time.Second
	DefaultTimeOut   time.Duration = 10 * time.Second
	DefaultHTTPPort  int           = 3000
)
