package utils

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"
)

// RunWithLimitedConcurrency 以限制並發數的方式執行多個任務
// 參數：
//   - tasks: 任務列表，每個任務是一個無參數無返回值的函數
//   - maxConcurrency: 最大並發數
func RunWithLimitedConcurrency(tasks []func(), maxConcurrency int) {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for _, task := range tasks {
		wg.Add(1)

		// 捕獲 task 變數
		taskCopy := task

		go func() {
			defer wg.Done()

			// 獲取信號量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 執行任務
			taskCopy()
		}()
	}

	wg.Wait()
}

// NewTimeoutContext 創建一個帶超時的 Context
// 參數：
//   - seconds: 超時秒數
//
// 返回：
//   - 上下文對象和取消函數
func NewTimeoutContext(seconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
}

// RetryWithBackoff 使用指數退避策略重試函數執行
// 參數：
//   - fn: 要執行的函數，返回錯誤表示執行失敗
//   - maxRetries: 最大重試次數
//   - initialBackoff: 初始退避時間（毫秒）
//
// 返回：
//   - 最後一次執行的錯誤
func RetryWithBackoff(fn func() error, maxRetries int, initialBackoff int) error {
	var err error
	backoff := time.Duration(initialBackoff) * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2 // 指數增長
		}
	}

	return err
}

// SafeGoroutine 安全地執行 goroutine，捕獲並處理 panic
// 參數：
//   - fn: 要執行的函數
//   - onPanic: 發生 panic 時的處理函數，接收 panic 值
func SafeGoroutine(fn func(), onPanic ...func(any)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 記錄 panic 堆棧
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, false)]
				PrettyPrintf("Goroutine panic: %v\n%s", r, stack)

				// 調用自定義 panic 處理
				if len(onPanic) > 0 && onPanic[0] != nil {
					onPanic[0](r)
				}
			}
		}()

		fn()
	}()
}

// WorkerPool 實現一個可取消的工作池
type WorkerPool struct {
	tasks  chan func()
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewWorkerPool 創建一個新的工作池
// 參數：
//   - numWorkers: 工作者數量
//   - queueSize: 任務隊列大小
//
// 返回：
//   - 工作池實例
func NewWorkerPool(numWorkers, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		tasks:  make(chan func(), queueSize),
		ctx:    ctx,
		cancel: cancel,
	}

	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}

	return pool
}

// worker 工作者循環處理任務
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						PrettyPrintf("工作池工作 panic: %v", r)
					}
				}()

				task()
			}()

		case <-p.ctx.Done():
			return
		}
	}
}

// Submit 提交任務到工作池
// 參數：
//   - task: 任務函數
//
// 返回：
//   - 錯誤信息，若工作池已關閉則返回錯誤
func (p *WorkerPool) Submit(task func()) error {
	select {
	case p.tasks <- task:
		return nil
	case <-p.ctx.Done():
		return errors.New("工作池已關閉")
	}
}

// Stop 停止工作池並等待所有任務完成
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
}
