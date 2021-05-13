package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/wychl/worker"
)

func main() {
	wg := sync.WaitGroup{}
	batch := 5
	result := 0
	mux := sync.RWMutex{}
	sum := newSum(&result, &mux)
	dataChan := make(chan interface{}, 10)

	// goroutine写入数据
	go func() {
		step := 500  // 步长
		max := 10000 //最大值

		for i := 0; i < 20; i++ {
			// 截止值
			endValue := (i + 1) * step
			if i == 19 {
				endValue = max + 1
			}

			// 数据写入channel
			dataChan <- sumPayload{
				startValue: i * step,
				endValue:   endValue,
			}
			fmt.Println(i, i*step, (i+1)*step)
		}

		// 数据已经写完，关闭channel
		close(dataChan)
	}()

	worker := worker.New(batch, sum, &wg)
	ctx := context.Background()
	worker.Run(dataChan, ctx)

	// 等待计算任务完成
	<-worker.Finished()

	fmt.Println("sum:", result)
}

// 返回和计算函数
func newSum(total *int, mux *sync.RWMutex) worker.Exec {
	return func(data interface{}) {
		p := data.(sumPayload)
		for i := p.startValue; i < p.endValue; i++ {
			mux.Lock()
			*total += i
			mux.Unlock()
		}
	}
}

// 和计算函数，输入参数类型
type sumPayload struct {
	startValue int
	endValue   int
}
