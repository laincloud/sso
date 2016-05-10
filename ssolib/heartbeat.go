package ssolib

import (
	"strconv"
	"time"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/iuser"
)

const (
	interval = 100 * time.Second
)

// 在某些网络环境下，如果 ssoserver 和 mysql 之间的长连接在 idle 一段时间后，此连接上发送的数据包会被丢弃，导致查询超时。
// 为避免这种情况的发生，可以定时查询数据库，以保持连接 live 。
func HeartbeatToDB(ctx context.Context) {
	for {
		startTime := time.Now()
		ub := getUserBackend(ctx)
		_, err := ub.GetUserByName("unknown")
		if err != iuser.ErrUserNotFound {
			log.Warnf("sso_debug error: GetUserByName: %s", err)
		}
		queryTime := time.Since(startTime).Seconds()
		if queryTime > 1 {
			log.Warnf("sso_debug heartbeat slow : " + strconv.FormatFloat(queryTime, 'f', 5, 64) + "s")
		} else {
			log.Debugf("sso_debug heartbeat normal " + strconv.FormatFloat(queryTime, 'f', 5, 64) + "s")
		}
		time.Sleep(interval)
	}

}
