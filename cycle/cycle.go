package cycle

import (
	"encoding/json"
	"log"
	"runtime"
	"time"

	"github.com/lvyalin/gin"
)

type Cycle struct {
	entries   []*Entry
	gin       *gin.Engine
	beforeRun func(*gin.Context) bool
	afterRun  func(*gin.Context)
}

type Job interface {
	Run(ctx *gin.Context) error
}

type Entry struct {
	Duration time.Duration
	Job      Job
}

func New(engine *gin.Engine) *Cycle {
	return &Cycle{
		entries: nil,
		gin:     engine,
	}
}

type FuncJob func(*gin.Context) error

func (f FuncJob) Run(ctx *gin.Context) error { return f(ctx) }

// add cron before func
func (c *Cycle) AddBeforeRun(beforeRun func(*gin.Context) bool) *Cycle {
	c.beforeRun = beforeRun
	return c
}

// add cron after func
func (c *Cycle) AddAfterRun(afterRun func(*gin.Context)) *Cycle {
	c.afterRun = afterRun
	return c
}

func (c *Cycle) AddFunc(duration time.Duration, cmd func(*gin.Context) error) {
	entry := &Entry{
		Duration: duration,
		Job:      FuncJob(cmd),
	}
	c.entries = append(c.entries, entry)
}

func (c *Cycle) Start() {
	for _, e := range c.entries {
		go c.run(e)
	}
}

//死循环
func (c *Cycle) run(e *Entry) {
	for {
		c.runWithRecovery(e)
	}
}

func (c *Cycle) runWithRecovery(entry *Entry) {
	ctx := gin.CreateNewContext(c.gin)
	// todo 这里判断是否存在，可以复用和减少部分gc
	cusctx := gin.CustomContext{
		Handle:    entry.Job.(FuncJob),
		Desc:      string(entry.Duration),
		Type:      "Cycle",
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
		ok := c.beforeRun(ctx)
		if !ok {
			return
		}
	}

	error := entry.Job.Run(ctx)
	ctx.CustomContext.Error = error
	ctx.CustomContext.EndTime = time.Now()

	if c.afterRun != nil {
		c.afterRun(ctx)
	}

	time.Sleep(entry.Duration)
}
