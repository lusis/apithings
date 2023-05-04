package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lusis/apithings/internal/statusthing/providers"
	"github.com/lusis/apithings/internal/statusthing/types"
	"github.com/stretchr/testify/require"
)

func TestConstructor(t *testing.T) {
	t.Parallel()
	t.Run("test-good", func(t *testing.T) {
		sh, err := NewStatusThingHandler(&testProvider{}, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, sh, "should not be nil")
	})

	t.Run("test-bad", func(t *testing.T) {
		sh, err := NewStatusThingHandler(&testProvider{}, "", nil, "")
		require.Error(t, err, "should error")
		require.Nil(t, sh, "should be nil")
	})
}

func TestInvalidContentType(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	p := &testProvider{}
	h, err := NewStatusThingHandler(p, "/", nil, "")
	require.NoError(t, err, "should not error")
	require.NotNil(t, h, "should not be nil")

	h.ServeHTTP(w, r)

	result := w.Result()
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	require.NoError(t, err, "body should be read")
	require.NotEmpty(t, body, "body should not be empty")
	require.Equal(t, http.StatusBadRequest, result.StatusCode)
	require.Equal(t, "invalid content type\n", string(body))
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	t.Run("internal-error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusInternalServerError, result.StatusCode, "test provider should immediately error")
	})

	t.Run("empty-results", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			allFunc: func() ([]*types.StatusThing, error) {
				return []*types.StatusThing{}, nil
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusOK, result.StatusCode, "should be okay")
		require.Equal(t, "[]\n", string(body), "should return empty result set")
	})
	t.Run("one-result", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			allFunc: func() ([]*types.StatusThing, error) {
				return []*types.StatusThing{
					{
						ID:          t.Name() + "_id",
						Name:        t.Name() + "_name",
						Description: t.Name() + "_desc",
						Status:      types.StatusGreen,
					},
				}, nil
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusOK, result.StatusCode, "should be okay")
		require.Equal(t, `[{"id":"TestGetAll/one-result_id","name":"TestGetAll/one-result_name","description":"TestGetAll/one-result_desc","status":"STATUS_GREEN"}]`, strings.TrimSuffix(string(body), "\n"), "should return empty result set")
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("unexpected-error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			getFunc: func(s string) (*types.StatusThing, error) { return nil, fmt.Errorf("snarf") },
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusInternalServerError, result.StatusCode, "test provider should immediately error")
	})

	t.Run("not-found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			getFunc: func(s string) (*types.StatusThing, error) { return nil, types.ErrNotFound },
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusNotFound, result.StatusCode, "should be not found")
		require.Equal(t, "not found\n", string(body), "should return not found")
	})

	t.Run("good", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			getFunc: func(s string) (*types.StatusThing, error) {
				return &types.StatusThing{
					ID:          t.Name() + "_id",
					Name:        t.Name() + "_name",
					Description: t.Name() + "_desc",
					Status:      types.StatusGreen,
				}, nil
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusOK, result.StatusCode, "should be ok")
		require.Equal(t, `{"id":"TestGet/good_id","name":"TestGet/good_name","description":"TestGet/good_desc","status":"STATUS_GREEN"}`, strings.TrimSuffix(string(body), "\n"), "should return not found")
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("unexpected-error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			// doesn't matter what we return here
			getFunc: func(s string) (*types.StatusThing, error) { return nil, fmt.Errorf("snarf") },
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusInternalServerError, result.StatusCode, "test provider should immediately error")
	})

	t.Run("not-found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			getFunc: func(s string) (*types.StatusThing, error) { return nil, types.ErrNotFound },
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusNotFound, result.StatusCode, "should be not found")
		require.Equal(t, "no such record\n", string(body), "should return not found")
	})

	t.Run("good", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/foobar", nil)
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		p := &testProvider{
			getFunc: func(s string) (*types.StatusThing, error) {
				return &types.StatusThing{
					ID:          t.Name() + "_id",
					Name:        t.Name() + "_name",
					Description: t.Name() + "_desc",
					Status:      types.StatusGreen,
				}, nil
			},
			removeFunc: func(s string) error { return nil },
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err, "body should read")
		require.Equal(t, http.StatusOK, result.StatusCode, "should be ok")
		require.Equal(t, `{"id":"TestDelete/good_id","name":"TestDelete/good_name","description":"TestDelete/good_desc","status":"STATUS_GREEN"}`, strings.TrimSuffix(string(body), "\n"), "should return not found")
	})
}

func TestPut(t *testing.T) {
	t.Parallel()

	t.Run("bad-request-invalid-body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("[]"))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		addCalled := false
		p := &testProvider{
			addFunc: func(p providers.Params) (*types.StatusThing, error) {
				addCalled = true
				return nil, fmt.Errorf("should not be called")
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusBadRequest, result.StatusCode, "test provider should immediately error")
		require.False(t, addCalled, "add should not be called")
	})

	t.Run("bad-request-required-values", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{}"))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		addCalled := false
		p := &testProvider{
			addFunc: func(p providers.Params) (*types.StatusThing, error) {
				addCalled = true
				return nil, types.ErrRequiredValueMissing
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusBadRequest, result.StatusCode, "should fail validation error")
		require.True(t, addCalled, "should have called our add func")
	})

	t.Run("already-exists", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"status":"STATUS_GREEN", "name":"test service 6","description":"foo"}`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		addCalled := false
		p := &testProvider{
			addFunc: func(p providers.Params) (*types.StatusThing, error) {
				addCalled = true
				return nil, types.ErrAlreadyExists
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusConflict, result.StatusCode, "should fail with conflict error")
		require.True(t, addCalled, "should have called our add func")
	})

	t.Run("internal-error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"status":"STATUS_GREEN", "name":"test service 6","description":"foo"}`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		addCalled := false
		p := &testProvider{
			addFunc: func(p providers.Params) (*types.StatusThing, error) {
				addCalled = true
				return nil, fmt.Errorf("snarf")
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusInternalServerError, result.StatusCode, "should fail with internal error")
		require.True(t, addCalled, "should have called our add func")
	})

	t.Run("good", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"status":"STATUS_GREEN", "name":"test service 6","description":"foo"}`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		addCalled := false
		p := &testProvider{
			addFunc: func(p providers.Params) (*types.StatusThing, error) {
				addCalled = true
				return &types.StatusThing{ID: t.Name()}, nil
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		stringBody := strings.TrimSuffix(string(body), "\n")
		require.Equal(t, http.StatusOK, result.StatusCode, "should pass")
		require.True(t, addCalled, "should have called our add func")
		require.Equal(t, `{"id":"TestPut/good","name":"","description":"","status":"STATUS_UNKNOWN"}`, stringBody)
	})
}

func TestPost(t *testing.T) {
	t.Run("bad-request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/abcdefg", strings.NewReader(`[]`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		statusCalled := false
		p := &testProvider{
			statusFunc: func(s1 string, s2 types.Status) error {
				statusCalled = true
				return fmt.Errorf("snarf")
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusBadRequest, result.StatusCode, "should bad request error")
		require.False(t, statusCalled, "should have not called our status func")
	})
	t.Run("internal-error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/abcdefg", strings.NewReader(`{"status":"STATUS_GREEN", "name":"test service 6","description":"foo"}`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		statusCalled := false
		p := &testProvider{
			statusFunc: func(s1 string, s2 types.Status) error {
				statusCalled = true
				return fmt.Errorf("snarf")
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusInternalServerError, result.StatusCode, "should throw interal error")
		require.True(t, statusCalled, "should have called our status func")
	})
	t.Run("good", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/abcdefg", strings.NewReader(`{"status":"STATUS_GREEN", "name":"test service 6","description":"foo"}`))
		r.Header.Set(contentTypeHeader, applicationJSON)
		w := httptest.NewRecorder()
		statusCalled := false
		p := &testProvider{
			statusFunc: func(s1 string, s2 types.Status) error {
				statusCalled = true
				return nil
			},
		}
		h, err := NewStatusThingHandler(p, "/", nil, "")
		require.NoError(t, err, "should not error")
		require.NotNil(t, h, "should not be nil")

		h.ServeHTTP(w, r)
		result := w.Result()
		defer result.Body.Close()
		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		stringBody := strings.TrimSuffix(string(body), "\n")
		require.Equal(t, http.StatusOK, result.StatusCode, "should pass")
		require.True(t, statusCalled, "should have called our status func")
		require.Empty(t, stringBody, "should not return a body")
	})
}

type testProvider struct {
	providers.UnimplementedProvider
	allFunc    func() ([]*types.StatusThing, error)
	getFunc    func(string) (*types.StatusThing, error)
	addFunc    func(providers.Params) (*types.StatusThing, error)
	removeFunc func(string) error
	statusFunc func(string, types.Status) error
}

// All gets all [types.StatusThing]
func (tp *testProvider) All(ctx context.Context) ([]*types.StatusThing, error) {
	if tp.allFunc == nil {
		return nil, fmt.Errorf("missing allfunc")
	}
	return tp.allFunc()
}

// Get gets a [types.StatusThing] by its id
func (tp *testProvider) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	if tp.getFunc == nil {
		return nil, fmt.Errorf("missing getfunc")
	}
	return tp.getFunc(id)
}

// Add adds a [types.StatusThing]
func (tp *testProvider) Add(ctx context.Context, newThing providers.Params) (*types.StatusThing, error) {
	if tp.addFunc == nil {
		return nil, fmt.Errorf("missing addfunc")
	}
	return tp.addFunc(newThing)
}

// Remove removes a [types.StatusThing] by its id
func (tp *testProvider) Remove(ctx context.Context, id string) error {
	if tp.removeFunc == nil {
		return fmt.Errorf("missing removefunc")
	}
	return tp.removeFunc(id)
}

// SetStatus sets the status of a [types.StatusThing] by its id
func (tp *testProvider) SetStatus(ctx context.Context, id string, status types.Status) error {
	if tp.statusFunc == nil {
		return fmt.Errorf("missing statusfunc")
	}
	return tp.statusFunc(id, status)
}
