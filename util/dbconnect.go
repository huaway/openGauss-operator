package util

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"
)

type DbConnect struct {
	client *sql.DB
	connStr string
	ctx context.Context
	stop context.CancelFunc
}

func NewDbConnect(host string) (*DbConnect, error) {
	var err error
	db := &DbConnect{}
	db.connStr = fmt.Sprintf("user=root dbname=postgres password=root host=%s port=5432 sslmode=disable", host)
	db.client, err = sql.Open("postgres", db.connStr)
	if err != nil {
		klog.Error("Open opengauss error: ", err)
		return nil, err
	}

	db.ctx, db.stop = context.WithCancel(context.Background())

	return db, nil
}

func (db *DbConnect) Reconnect() error {
	klog.Info("Try to reconnect shardingsphere")
	var err error
	db.client, err = sql.Open("postgres", db.connStr)
	if err != nil {
		klog.Error("Open opengauss error: ", err)
		return err
	}

	db.ctx, db.stop = context.WithCancel(context.Background())
	klog.Info("Reconnect to shardingsphere")
	return nil
}

func (db *DbConnect) Close() {
	db.client.Close()
	db.stop()
}

func (db *DbConnect) ExecCmd(cmd string) error {
	cnt := 0
	for {
		_, err := db.client.ExecContext(db.ctx,
			cmd,
		)
		if err != nil {
			klog.Error(fmt.Sprintf("Error: %s when executing cmd:\n %s", err, cmd))
			if strings.Contains(err.Error(), "This connection has been closed") {
				db.Reconnect()
			}
			time.Sleep(500 * time.Millisecond)
			if cnt < 5 {
				cnt++
				continue
			}
			return err
		}
		break		
	}
	return nil
}