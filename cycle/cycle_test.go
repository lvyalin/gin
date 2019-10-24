package cycle

import (
	"fmt"
	"github.com/lvyalin/gin"
	"testing"
	"time"
)

func PanicFunc(context *gin.Context) error {
	panic("this is a panic job")
}

func TestJobPanicRecover(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddFunc(time.Second, PanicFunc)
	cycle.Start()
	time.Sleep(time.Second * 2)
}

func TestBeforeFuncPanic(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddBeforeRun(func(ctx *gin.Context) bool {
		panic("panic in before func")
	})
	cycle.AddFunc(time.Second, PanicFunc)
	cycle.Start()
	time.Sleep(time.Second * 2)
}

func TestAfterFuncPanic(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddAfterRun(func(ctx *gin.Context) {
		panic("panic in after func")
	})
	cycle.AddFunc(time.Second, PanicFunc)
	cycle.Start()
	time.Sleep(time.Second * 2)
}

func TestBeforeFuncStop(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddBeforeRun(func(ctx *gin.Context) bool {
		return false
	})
	cycle.AddFunc(time.Second, PanicFunc)
	cycle.Start()
	time.Sleep(time.Second * 2)
}

func DemoJob(ctx *gin.Context) error {
	fmt.Println("this is a dome job")
	return nil
}

// 正常job
func TestRunJob(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddFunc(time.Second, DemoJob)
	cycle.Start()
	// todo 判断方法
}

// 多次start
func TestStartTwice(t *testing.T) {
	cycle := New(gin.New())
	cycle.AddFunc(time.Second, DemoJob)
	cycle.Start()
	cycle.Start()
	// todo 判断
}
