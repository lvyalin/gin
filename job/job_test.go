package job

import (
	"errors"
	"fmt"
	"gin"
	"sync"
	"testing"
	"time"
)

func PanicFunc(context *gin.Context) error {
	panic("this is a panic job")
	return nil
}

func TestJobPanicRecover(t *testing.T) {
	job := New(gin.New())
	job.Run(PanicFunc)
	time.Sleep(time.Second * 2)
}

func TestBeforeFuncPanic(t *testing.T) {
	job := New(gin.New())
	job.AddBeforeRun(func(ctx *gin.Context) bool {
		panic("panic in before func")
	})
	job.Run(PanicFunc)
	time.Sleep(time.Second * 2)
}

func TestAfterFuncPanic(t *testing.T) {
	job := New(gin.New())
	job.AddAfterRun(func(ctx *gin.Context) {
		panic("panic in after func")
	})
	job.Run(PanicFunc)
	time.Sleep(time.Second * 2)
}

func TestBeforeFuncStop(t *testing.T) {
	job := New(gin.New())
	job.AddBeforeRun(func(ctx *gin.Context) bool {
		return false
	})
	job.Run(PanicFunc)
	time.Sleep(time.Second * 2)
}

func DemoJob(ctx *gin.Context) error {
	fmt.Println("this is a dome job")
	return nil
}

// 正常job
func TestRunJob(t *testing.T) {
	job := New(gin.New())
	job.Run(DemoJob)
	// todo 判断方法
	time.Sleep(time.Second * 2)
}

func ErrorJob(ctx *gin.Context) error {
	fmt.Println("this is a error job")
	return errors.New("this is a error job")
}

func TestRunErrorJob(t *testing.T) {
	job := New(gin.New())
	job.Run(ErrorJob)
	// todo 判断方法
	time.Sleep(time.Second * 2)
}

func TestConcurrentRead(t *testing.T) {
	job := New(gin.New())
	wg := sync.WaitGroup{}
	job.Run(func(ctx *gin.Context) error {
		ctx.Set("test", 1)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for j := 0; j < 10; j++ {
					a := ctx.MustGet("test")
					fmt.Printf("测试 %v \n", a)
				}
			}()
		}
		return nil
	})

	wg.Wait()
	time.Sleep(time.Second * 2)
}

func TestConcurrentWrite(t *testing.T) {
	job := New(gin.New())
	wg := sync.WaitGroup{}
	job.Run(func(ctx *gin.Context) error {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					ctx.Set("test", j)
					a := ctx.MustGet("test")
					fmt.Printf("测试 %v \n", a)
				}
			}()
		}
		return nil
	})

	wg.Wait()
	time.Sleep(time.Second * 2)
}
