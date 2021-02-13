package crawler

import (
	"net/url"
)

type workerInput struct {
	parent *url.URL
}

type workerOutput struct {
	result Result
	err    error
}

type worker struct {
	urlFinder      urlFinder
	workerInputCh  chan workerInput
	workerOutputCh chan workerOutput
}

func (w *worker) work() {
	for workerIn := range w.workerInputCh {
		children, err := w.urlFinder.find(workerIn.parent)
		w.workerOutputCh <- workerOutput{
			err: err,
			result: Result{
				Parent:   workerIn.parent,
				Children: children,
			},
		}
	}
}
