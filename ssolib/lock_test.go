package ssolib

import (
	"testing"
	"time"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/lock"
	"github.com/laincloud/sso/ssolib/models/testhelper"
)

func TestLockAndUnlockWorksIn20SecondsAsLocalLock(t *testing.T) {
	mysqlDSN := testhelper.GetTestMysqlDSN()

	testLock := lock.New(mysqlDSN, "testLock")
	log.Info(testLock)
	t.Log(testLock)

	a := 1
	b := 1

	go func() {
		t.Log("begin 1")
		testLock.Lock()
		t.Log("enter 1")
		defer testLock.Unlock()
		if a == 1 && b == 1 {
			time.Sleep(8 * time.Second)
			a = 2
		}
		t.Log("before leaving 1", a, b)
	}()

	go func() {
		t.Log("begin 2")
		testLock.Lock()
		t.Log("enter 2")
		defer testLock.Unlock()
		if a == 1 && b == 1 {
			time.Sleep(8 * time.Second)
			b = 2
		}
		t.Log("before leaving 2:", a, b)
	}()

	time.Sleep(19 * time.Second)

	if (a == 1 && b == 2) || (a == 2 && b == 1) {
		t.Log(a, b)
	} else {
		t.Fatal("lock failed:", a, b)
	}
}

func TestLockAndUnlockWorksIn20SecondsAsDistributedLock(t *testing.T) {
	mysqlDSN := testhelper.GetTestMysqlDSN()

	testLock1 := lock.New(mysqlDSN, "testLock")
	testLock2 := lock.New(mysqlDSN, "testLock")

	a := 1
	b := 1

	go func() {
		t.Log("begin 1")
		testLock1.Lock()
		t.Log("enter 1")
		defer testLock1.Unlock()
		if a == 1 && b == 1 {
			time.Sleep(8 * time.Second)
			a = 2
		}
		t.Log("before leaving 1", a, b)
	}()

	go func() {
		t.Log("begin 2")
		testLock2.Lock()
		t.Log("enter 2")
		defer testLock2.Unlock()
		if a == 1 && b == 1 {
			time.Sleep(8 * time.Second)
			b = 2
		}
		t.Log("before leaving 2:", a, b)
	}()

	time.Sleep(19 * time.Second)

	if (a == 1 && b == 2) || (a == 2 && b == 1) {
		t.Log(a, b)
	} else {
		t.Fatal("lock failed:", a, b)
	}
}
