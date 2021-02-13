package crawler

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

const defaultTimeout time.Duration = 5 * time.Second

func TestWorker(t *testing.T) {
	t.Run("worker populates parent in results", func(t *testing.T) {
		inCh := make(chan workerInput)
		outCh := make(chan workerOutput)

		mockedURLFinder := mockURLFinder{}

		w := worker{
			urlFinder:      &mockedURLFinder,
			workerInputCh:  inCh,
			workerOutputCh: outCh,
		}

		go w.work()
		defer close(inCh)

		parentURL := &url.URL{Host: "someparent"}

		mockedURLFinder.On("find", parentURL).Return([]*url.URL{}, nil)

		inCh <- workerInput{
			parent: parentURL,
		}

		var out workerOutput
		select {
		case out = <-outCh:
		case <-time.After(defaultTimeout):
			assert.Fail(t, "timeout")
		}
		assert.NoError(t, out.err)
		require.NotEmpty(t, out.result)
		assert.Equal(t, parentURL, out.result.Parent)
	})

	t.Run("worker returns found children if successful", func(t *testing.T) {
		inCh := make(chan workerInput)
		outCh := make(chan workerOutput)

		mockedURLFinder := mockURLFinder{}

		w := worker{
			urlFinder:      &mockedURLFinder,
			workerInputCh:  inCh,
			workerOutputCh: outCh,
		}

		go w.work()
		defer close(inCh)

		parentURL := &url.URL{}
		children := []*url.URL{
			{Host: "child1"},
			{Host: "child2"},
		}
		mockedURLFinder.On("find", parentURL).Return(children, nil)

		inCh <- workerInput{
			parent: parentURL,
		}

		var out workerOutput
		select {
		case out = <-outCh:
		case <-time.After(defaultTimeout):
			assert.Fail(t, "timeout")
		}
		assert.NoError(t, out.err)
		require.NotEmpty(t, out.result)
		assert.Equal(t, children, out.result.Children)
	})

	t.Run("worker returns error if could not find children", func(t *testing.T) {
		inCh := make(chan workerInput)
		outCh := make(chan workerOutput)

		mockedURLFinder := mockURLFinder{}

		w := worker{
			urlFinder:      &mockedURLFinder,
			workerInputCh:  inCh,
			workerOutputCh: outCh,
		}

		go w.work()
		defer close(inCh)

		parentURL := &url.URL{}
		mockedURLFinder.On("find", parentURL).Return([]*url.URL{}, errors.New("some random error"))

		inCh <- workerInput{
			parent: parentURL,
		}

		var out workerOutput
		select {
		case out = <-outCh:
		case <-time.After(defaultTimeout):
			assert.Fail(t, "timeout")
		}
		assert.Error(t, out.err)
		require.NotEmpty(t, out.result)
	})

}

type mockURLFinder struct {
	mock.Mock
}

func (m *mockURLFinder) find(u *url.URL) ([]*url.URL, error) {
	args := m.Called(u)
	return args.Get(0).([]*url.URL), args.Error(1)
}
