package fancyhttpclient

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDoerWithDelay struct {
	mock.Mock
	Delay time.Duration
}

func (md *mockDoerWithDelay) Do(*http.Request) (*http.Response, error) {
	time.Sleep(md.Delay)
	args := md.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*http.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestFancyHTTPClient_Do(t *testing.T) {
	t.Run("one request taking 100ms, one worker, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer)
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		res, err := fhc.Do(req)
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, took > 100*time.Millisecond && took < 110*time.Millisecond)
	})

	t.Run("3 request taking 100ms, one worker, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer)
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)

		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				res, err := fhc.Do(req)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				wg.Done()
			}()
		}
		wg.Wait()

		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 300*time.Millisecond && took < 310*time.Millisecond)
	})

	t.Run("3 request taking 100ms, 3 workers, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer, WithMaxConn(3))
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				res, err := fhc.Do(req)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				wg.Done()
			}()
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 100*time.Millisecond && took < 110*time.Millisecond)
	})

	t.Run("3 request taking 100ms, 3 workers, 50ms delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer, WithMaxConn(3), WithDelay(50*time.Millisecond))
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				res, err := fhc.Do(req)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				wg.Done()
			}()
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 200*time.Millisecond && took < 210*time.Millisecond)
	})
}

func TestFancyHTTPClient_DoBunch(t *testing.T) {
	t.Run("one request taking 100ms, one worker, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer)
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		responsers := fhc.DoBunch([]*http.Request{req})
		res, err := responsers[0].Response()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, took > 100*time.Millisecond && took < 110*time.Millisecond)
	})

	t.Run("3 request taking 100ms, one worker, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer)
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		responsers := fhc.DoBunch([]*http.Request{req, req, req})
		res, err := responsers[0].Response()
		assert.NoError(t, err)
		assert.NotNil(t, res)
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 300*time.Millisecond && took < 310*time.Millisecond)
	})

	t.Run("3 request taking 100ms, 3 workers, 0 delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer, WithMaxConn(3))
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		responsers := fhc.DoBunch([]*http.Request{req, req, req})
		res, err := responsers[0].Response()
		assert.NoError(t, err)
		assert.NotNil(t, res)
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 100*time.Millisecond && took < 110*time.Millisecond)
	})

	t.Run("3 request taking 100ms, 3 workers, 50ms delay", func(t *testing.T) {
		doer := &mockDoerWithDelay{Delay: 100 * time.Millisecond}
		doer.On("Do", mock.Anything).Return(httpmock.NewStringResponse(http.StatusOK, "body"), nil)
		fhc := New(doer, WithMaxConn(3), WithDelay(50*time.Millisecond))
		start := time.Now()
		req, _ := http.NewRequest("GET", "https://example.com", nil)
		responsers := fhc.DoBunch([]*http.Request{req, req, req})
		res, err := responsers[0].Response()
		assert.NoError(t, err)
		assert.NotNil(t, res)
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 200*time.Millisecond && took < 210*time.Millisecond)
	})

}
