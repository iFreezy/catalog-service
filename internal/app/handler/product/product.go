package hproduct

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"

	"github.com/iFreezy/catalog-service/internal/app/entity"
	rhandler "github.com/iFreezy/catalog-service/internal/app/handler"
	"github.com/iFreezy/catalog-service/internal/app/service"
	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
)

type handler struct {
	svcProduct service.Product
}

func NewHandler(svcProduct service.Product) rhandler.Product {
	return &handler{svcProduct: svcProduct}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestProductCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.SendError(w, http.StatusBadRequest, err)
		return
	}

	product, err := h.svcProduct.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrAlreadyExists):
			httph.SendError(w, http.StatusBadRequest, err)
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusBadRequest, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := entity.ResponseProduct{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusCreated, resp)
}

func (h *handler) GetByGUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid, err := uuid.FromString(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	product, err := h.svcProduct.GetByGUID(r.Context(), guid)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusNotFound, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := entity.ResponseProduct{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusOK, resp)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid, err := uuid.FromString(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestProductUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.SendError(w, http.StatusBadRequest, err)
		return
	}

	product, err := h.svcProduct.Update(r.Context(), guid, req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusNotFound, err)
		case errors.Is(err, entity.ErrAlreadyExists):
			httph.SendError(w, http.StatusBadRequest, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	resp := entity.ResponseProduct{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}

	httph.SendJSON(w, http.StatusOK, resp)
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid, err := uuid.FromString(vars["guid"])
	if err != nil {
		httph.SendError(w, http.StatusBadRequest, entity.ErrIncorrectParameters)
		return
	}

	err = h.svcProduct.Delete(r.Context(), guid)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			httph.SendError(w, http.StatusNotFound, err)
		default:
			httph.SendError(w, http.StatusInternalServerError, err)
		}
		return
	}

	httph.SendEmpty(w, http.StatusNoContent)
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.svcProduct.List(r.Context())
	if err != nil {
		httph.SendError(w, http.StatusInternalServerError, err)
		return
	}

	resp := make([]entity.ResponseProduct, len(products))
	for i, p := range products {
		resp[i] = entity.ResponseProduct{
			GUID:         p.GUID,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			CategoryGUID: p.CategoryGUID,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		}
	}

	httph.SendJSON(w, http.StatusOK, resp)
}
