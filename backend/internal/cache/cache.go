package cache //nolint:gofmt

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var Store *gocache.Cache

func Init() {
	Store = gocache.New(1*time.Hour, 15*time.Minute)
}
