package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/raymondwoongso/goerp/domain"
	"github.com/raymondwoongso/goerp/domain/xerror"
	"github.com/raymondwoongso/goerp/example"
)

type HandlerParam struct {
	SubmoduleACreate example.SubmoduleACreate
	SubmoduleBGet    example.SubmoduleBGet
}

type Handler struct {
	SubmoduleACreate example.SubmoduleACreate
	SubmoduleBGet    example.SubmoduleBGet
}

func NewHandler(param HandlerParam) *Handler {
	if param.SubmoduleACreate == nil {
		panic("SubmoduleA/Create is empty")
	}

	if param.SubmoduleBGet == nil {
		panic("SubmoduleB/Get is empty")
	}

	return &Handler{
		SubmoduleACreate: param.SubmoduleACreate,
		SubmoduleBGet:    param.SubmoduleBGet,
	}
}

func (h *Handler) CreateSubmoduleA(w http.ResponseWriter, req *http.Request) {
	var body domain.Example
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.SubmoduleACreate.Invoke(req.Context(), body)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) GetSubmoduleB(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.SubmoduleBGet.Invoke(req.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func writeError(w http.ResponseWriter, err error) {
	var xerr xerror.Error
	if errors.As(err, &xerr) {
		w.WriteHeader(mapXErrorCodeToHTTP(xerr.Code))
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func mapXErrorCodeToHTTP(code xerror.Code) int {
	switch code {
	case xerror.CodeNotFound:
		return http.StatusNotFound
	case xerror.CodeForbidden:
		return http.StatusForbidden
	case xerror.CodeDuplicate:
		return http.StatusConflict
	case xerror.CodeUnprocessable:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
