package pipeline

import (
	"bufio"
	"os"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/record"
	"sync"
)

// Reader implements Pipe interface for reading data
type Reader struct {
	file    *os.File
	results chan ReadResult
	wg      *sync.WaitGroup
	jobs    chan ReadJob
	errors  chan error
}

// ReadJob defines job for read workers pool
type ReadJob struct {
	Record *record.Row
}

// Result represents operation outcome
type ReadResult struct {
	WorkerID int
	Result   *handler.ResultData
	Err      error
}

// type Record struct {
// 	URL             string   `json:"url"`
// 	State           string   `json:"state"`
// 	Categories      []string `json:"categories"`
// 	CategoryAnother string   `json:"category_another"`
// 	ForMainPage     bool     `json:"for_main_page"`
// 	Ctime           int      `json:"ctime"`
// }

// NewReader returns new instance of Reader pipe
func NewReader(file *os.File, results chan ReadResult, wg *sync.WaitGroup, jobs chan ReadJob, errors chan error) Pipe {
	return Reader{
		file:    file,
		results: results,
		jobs:    jobs,
		errors:  errors,
		wg:      wg,
	}
}

// Call executes main Pipe action for reading
func (r Reader) Call(in, out chan interface{}) {
	mainWg := &sync.WaitGroup{}
	recordWg := &sync.WaitGroup{}
	go func(file *os.File, jobs chan ReadJob, errors chan error, wg *sync.WaitGroup, mainWg *sync.WaitGroup) {
		defer close(jobs)
		defer close(errors)

		scanner := bufio.NewScanner(file)
		var writesNum int
		for scanner.Scan() {
			bytes := scanner.Bytes()
			row, decodeError := record.DecodeLine(bytes)
			if decodeError != nil {
				errors <- decodeError
				break
			}

			if len(row.Categories) > 0 {
				writesNum++
				jobs <- ReadJob{Record: row}
			}
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			errors <- scannerErr
		}
	}(r.file, r.jobs, r.errors, recordWg, mainWg)

	go func(wg *sync.WaitGroup, results chan ReadResult) {
		wg.Wait()
		close(results)
	}(r.wg, r.results)

	for res := range r.results {
		if res.Err == nil {
			out <- res.Result
		}
	}
}
