package worker

import (
	"context"
	"log"
	"sync"
)

// Worker 管理器
type Worker struct {
	batch        int  //启动worker数量
	exec         Exec // 业务处理方法
	wg           *sync.WaitGroup
	finishedChan chan struct{} //
}

// Exec 业务处理器
type Exec func(data interface{})

// New
func New(batch int, exec Exec, wg *sync.WaitGroup) *Worker {
	return &Worker{
		batch: batch,
		exec:  exec,
		wg:    wg,
	}
}

// Finished 返回finishedChan
func (w *Worker) Finished() chan struct{} {
	w.finishedChan = make(chan struct{})
	go w.checkFinished()
	return w.finishedChan
}

// checkFinished 等待所有的goroutine已经结束
func (w *Worker) checkFinished() {
	w.wg.Wait()
	w.finishedChan <- struct{}{}
}

// Run 运行worker
func (w *Worker) Run(data chan interface{}, ctx context.Context) {
	for i := 0; i < w.batch; i++ {
		w.wg.Add(1)

		// 启动处理器
		go func() {
			defer w.wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case value, ok := <-data:
					// channel已经关闭且数据已经读完，则终止goroutine
					if !ok && len(data) == 0 {
						return
					}

					// 处理接收的数据
					w.exec(value)
				default:
					// 没有新数据进入下一次循环
					log.Println("no receive data")
				}
			}
		}()
	}
}
