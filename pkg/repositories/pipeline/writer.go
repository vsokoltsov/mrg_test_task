package pipeline

import (
	"fmt"
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/storage"
	"sync"
)

// Writer represents Pipe interface for writing
type Writer struct {
	wg            *sync.WaitGroup
	jobs          chan models.WriteJob
	results       chan models.WriteResult
	errors        chan error
	mu            *sync.Mutex
	categoryFiles map[string]*os.File
	store         storage.Int
}

// NewWriter returns new Writer pipe
func NewWriter(
	wg *sync.WaitGroup,
	jobs chan models.WriteJob,
	results chan models.WriteResult,
	errors chan error,
	store storage.Int) Pipe {
	return Writer{
		wg:            wg,
		jobs:          jobs,
		results:       results,
		categoryFiles: make(map[string]*os.File),
		mu:            &sync.Mutex{},
		store:         store,
	}
}

// Call executes pipe action for writing results to file
func (w Writer) Call(in, out chan interface{}) {
	go func(in chan interface{}, w *Writer) {
		defer close(w.jobs)
		for data := range in {
			resultData := data.(*models.ResultData)
			for _, category := range resultData.Categories {
				var (
					catFile *os.File
				)
				catFile = w.getCategoryFile(category)
				if catFile == nil {
					catFile, _ = w.setCategoryFile(category)
				}
				w.jobs <- models.WriteJob{File: catFile, ResultData: resultData, Category: category}
			}
		}
	}(in, &w)

	go func(wg *sync.WaitGroup, results chan models.WriteResult) {
		wg.Wait()
		close(results)
	}(w.wg, w.results)

	for res := range w.results {
		out <- res
	}
}

func (w Writer) getCategoryFile(category string) *os.File {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.categoryFiles[category]
}

func (w Writer) setCategoryFile(category string) (*os.File, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fp := w.store.ResultPath(category)
	categoryFile, err := w.store.CreateFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error of creating %s file: %s", category, err)
	}
	return categoryFile, nil
}
