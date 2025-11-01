package cache

import (
	gocache "github.com/patrickmn/go-cache"
	"time"
)

var Store *gocache.Cache

func Init() {
	Store = gocache.New(1*time.Hour, 15*time.Minute)
}
