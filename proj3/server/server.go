package server

import (
	"encoding/json"
	"io"
	"log"
	"proj3/data/db"
	"proj3/queue"
	"proj3/utils"
	. "proj3/utils"
	"sync"
	"sync/atomic"
)

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	Threadcount    int // Represents the number of threads
	IsWorkstealing bool
}
type ServerContext struct {
	mu     *sync.Mutex
	cond   *sync.Cond
	prodwg *sync.WaitGroup
	conswg *sync.WaitGroup
	done   int64
}

func NewServerContext() *ServerContext {
	var m sync.Mutex
	var cw sync.WaitGroup
	var pw sync.WaitGroup
	c := sync.NewCond(&m)
	return &ServerContext{mu: &m, cond: c, prodwg: &pw, conswg: &cw, done: 0}
}

// Run starts the forcaster based on cofiguration information
func Run(config Config) {
	//I can add the max number of requests, in this case its 50
	//in temporary version let's just start with producer then consumer
	MotherQueue := queue.NewLockFreeQueue()
	ctx := NewServerContext()
	ctx.conswg.Add(1)
	go func() {
		defer ctx.conswg.Done()
		config.Consumer(MotherQueue, true, ctx)
	}()
	if config.Mode == "p" {
		var j int64
		for i := 0; i < config.Threadcount-1; i++ {
			ctx.conswg.Add(1)
			atomic.AddInt64(&j, 1)
			go func() {
				defer ctx.conswg.Done()
				if config.IsWorkstealing {
					config.Consumer(MotherQueue, false, ctx)
				} else {
					config.Consumer(MotherQueue, true, ctx)
				}
			}()
		}
	}
	config.producer(MotherQueue, ctx)
	ctx.conswg.Wait()
}

// keeps looping through the input stream to convert requests into tasks
func (c Config) producer(taskqueue *queue.LockFreeQueue, ctx *ServerContext) {
	log.Println("in producer")
	var r Request
	for atomic.LoadInt64(&ctx.done) == 0 {
		if err := c.Decoder.Decode(&r); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		} else {
			rcopy := r
			ctx.prodwg.Add(1)
			go func(r Request) {
				defer ctx.prodwg.Done()
				if r.Symbol == "TERM_SIG" {
					atomic.CompareAndSwapInt64(&ctx.done, 0, 1)
					log.Printf("Detected TERM_SIG, will terminate after this batch.")
					return
				}
				normdist := db.CheckDB(r.Symbol)
				log.Printf("PRODUCER -> %v\n", r)
				taskqueue.Enqueue(r.RequestToTask(normdist))
				ctx.mu.Lock()
				ctx.cond.Signal()
				ctx.mu.Unlock()
			}(rcopy)
		}
	}
	ctx.prodwg.Wait()
}

func (c Config) Consumer(taskqueue *queue.LockFreeQueue, isprimary bool, ctx *ServerContext) {
	var completed CompletedTask
	var task *Task
	gettask := func() *Task {
		if isprimary {
			task = taskqueue.PopBack()
		} else {
			task = taskqueue.PopFront()
		}
		return task
	}
	for atomic.LoadInt64(&ctx.done) == 0 || taskqueue.Count > 0 {
		ctx.mu.Lock()
		task = gettask()
		for task == nil && atomic.LoadInt64(&ctx.done) == 0 {
			ctx.cond.Wait()
			task = gettask()
		}
		ctx.mu.Unlock()
		if task != nil {
			log.Println("inside loop now")
			completed.RequestInfo = task.Request
			log.Printf("%t TASK: %v\n", isprimary, task)
			for i := 0; i < c.Threadcount; i++ {
				go utils.GetNormInv(task)
			}
			mean, stdv := utils.ConsolidateCumulative(task)
			completed.Probability = GetHypothesisProbability(mean, stdv, task.HypothesisPercentageChange)
			log.Printf("[1] FINAL RESULT -> %v\n", completed)
			if err := c.Encoder.Encode(completed); err != nil {
				panic(err)
			}
		} else {
			break
		}
	}
	if atomic.LoadInt64(&ctx.done) == 1 {
		ctx.mu.Lock()
		ctx.cond.Broadcast()
		ctx.mu.Unlock()
	}
}
