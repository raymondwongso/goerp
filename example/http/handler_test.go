package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raymondwongso/goerp/domain"
	"github.com/raymondwongso/goerp/domain/xerror"
	examplemock "github.com/raymondwongso/goerp/example/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type handlerTestSuite struct {
	submoduleACreate *examplemock.MockSubmoduleACreate
	submoduleBGet    *examplemock.MockSubmoduleBGet
}

func newHandlerTestSuite(t *testing.T) *handlerTestSuite {
	ctrl := gomock.NewController(t)
	return &handlerTestSuite{
		submoduleACreate: examplemock.NewMockSubmoduleACreate(ctrl),
		submoduleBGet:    examplemock.NewMockSubmoduleBGet(ctrl),
	}
}

func (ts *handlerTestSuite) newHandler() *Handler {
	return NewHandler(HandlerParam{
		SubmoduleACreate: ts.submoduleACreate,
		SubmoduleBGet:    ts.submoduleBGet,
	})
}

func TestHandler_CreateSubmoduleA(t *testing.T) {
	validInput := domain.Example{ID: 1, Name: "test"}
	validResult := domain.Example{ID: 2, Name: "abc"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleACreate.EXPECT().
			Invoke(gomock.Any(), validInput).
			Return(validResult, nil)

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/example", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ts.newHandler().CreateSubmoduleA(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var res domain.Example
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		assert.Equal(t, validResult, res)
	})

	t.Run("error — invalid request body", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		req := httptest.NewRequest(http.MethodPost, "/example", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		ts.newHandler().CreateSubmoduleA(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error — usecase xerror unprocessable", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleACreate.EXPECT().
			Invoke(gomock.Any(), validInput).
			Return(domain.Example{}, xerror.New(xerror.CodeUnprocessable, "invalid"))

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/example", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ts.newHandler().CreateSubmoduleA(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("error — usecase unknown error", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleACreate.EXPECT().
			Invoke(gomock.Any(), validInput).
			Return(domain.Example{}, errors.New("unexpected error"))

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/example", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ts.newHandler().CreateSubmoduleA(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetSubmoduleB(t *testing.T) {
	validResult := domain.Example{ID: 1, Name: "test"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleBGet.EXPECT().
			Invoke(gomock.Any(), int64(1)).
			Return(validResult, nil)

		req := httptest.NewRequest(http.MethodGet, "/example/1", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		ts.newHandler().GetSubmoduleB(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res domain.Example
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&res))
		assert.Equal(t, validResult, res)
	})

	t.Run("error — invalid id param", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		req := httptest.NewRequest(http.MethodGet, "/example/abc", nil)
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()

		ts.newHandler().GetSubmoduleB(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error — usecase xerror not found", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleBGet.EXPECT().
			Invoke(gomock.Any(), int64(999)).
			Return(domain.Example{}, xerror.New(xerror.CodeNotFound, "not found"))

		req := httptest.NewRequest(http.MethodGet, "/example/999", nil)
		req.SetPathValue("id", "999")
		w := httptest.NewRecorder()

		ts.newHandler().GetSubmoduleB(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("error — usecase unknown error", func(t *testing.T) {
		t.Parallel()
		ts := newHandlerTestSuite(t)

		ts.submoduleBGet.EXPECT().
			Invoke(gomock.Any(), int64(1)).
			Return(domain.Example{}, errors.New("unexpected error"))

		req := httptest.NewRequest(http.MethodGet, "/example/1", nil)
		req.SetPathValue("id", "1")
		w := httptest.NewRecorder()

		ts.newHandler().GetSubmoduleB(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
