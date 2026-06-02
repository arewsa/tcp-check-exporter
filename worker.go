package main

import (
    "strconv"
    "time"
)

type ProbeTask struct {
    Target  TargetConfig
    Timeout time.Duration
}

type ProbeResult struct {
    Target  TargetConfig
    Up      bool
    Latency float64
    Error   error
}

type WorkerPool struct {
    numWorkers int
    taskChan   chan ProbeTask
    resultChan chan ProbeResult
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    return &WorkerPool{
        numWorkers: numWorkers,
        taskChan:   make(chan ProbeTask, 100),
        resultChan: make(chan ProbeResult, 100),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.numWorkers; i++ {
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    for task := range wp.taskChan {
        up, latency := checkHost(task.Target.Host, strconv.Itoa(task.Target.Port), task.Timeout)
        wp.resultChan <- ProbeResult{
            Target:  task.Target,
            Up:      up,
            Latency: latency,
        }
    }
}

func (wp *WorkerPool) AddTask(task ProbeTask) {
    wp.taskChan <- task
}

func (wp *WorkerPool) Results() <-chan ProbeResult {
    return wp.resultChan
}

func (wp *WorkerPool) Close() {
    close(wp.taskChan)
    close(wp.resultChan)
}