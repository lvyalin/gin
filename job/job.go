package job

import (
	"encoding/json"
	"gin"
	"log"
	"runtime"
	"time"
)

type Job struct {
	gin           *gin.Engine
	beforeRun     func(*gin.Context, interface{}) bool
	afterRun      func(*gin.Context)
	getJobContext func(*gin.Context) interface{}
}

type Entry struct {
	span interface{}
	Job  FuncJob
}

type FuncJob func(*gin.Context) error

func (f FuncJob) Run(ctx *gin.Context) error { return f(ctx) }

func New(engine *gin.Engine) *Job {
	return &Job{
		gin: engine,
	}
}

// add cron before func
func (c *Job) AddBeforeRun(beforeRun func(*gin.Context, interface{}) bool) *Job {
	c.beforeRun = beforeRun
	return c
}

// add cron after func
func (c *Job) AddAfterRun(afterRun func(*gin.Context)) *Job {
	c.afterRun = afterRun
	return c
}

func (c *Job) AddJobContext(jobContext func(*gin.Context) interface{}) *Job {
	c.getJobContext = jobContext
	return c
}

func (c *Job) Run(ctx *gin.Context, f func(ctx *gin.Context) error) {
	e := &Entry{
		Job: FuncJob(f),
	}

	if c.getJobContext != nil {
		e.span = c.getJobContext(ctx)
	}
	go c.runWithRecovery(e)
}

func (c *Job) runWithRecovery(e *Entry) {
	ctx := gin.CreateNewContext(c.gin)
	cusctx := gin.CustomContext{
		Handle:    e.Job,
		Type:      "Job",
		StartTime: time.Now(),
	}
	ctx.CustomContext = cusctx
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			requestId, _ := ctx.Get("requestId")
			handleName := ctx.CustomContext.HandlerName()
			info, _ := json.Marshal(map[string]interface{}{
				"time":      time.Now().Format("2006-01-02 15:04:05"),
				"level":     "error",
				"module":    "stack",
				"handle":    handleName,
				"requestId": requestId,
			})
			log.Printf("%s\n-------------------stack-start-------------------\n%v\n%s\n-------------------stack-end-------------------\n", string(info), r, buf)
		}
		gin.RecycleContext(c.gin, ctx)
	}()

	if c.beforeRun != nil {
		ok := c.beforeRun(ctx, e.span)
		if !ok {
			return
		}
	}

	error := e.Job.Run(ctx)
	ctx.CustomContext.Error = error
	ctx.CustomContext.EndTime = time.Now()

	if c.afterRun != nil {
		c.afterRun(ctx)
	}
}
