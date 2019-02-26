package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/influxdata/influxdb"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// DocumentBackend is all services and associated parameters required to construct
// the DocumentHandler.
type DocumentBackend struct {
	Logger *zap.Logger

	DocumentService influxdb.DocumentService
}

// NewDocumentBackend returns a new instance of DocumentBackend.
func NewDocumentBackend(b *APIBackend) *DocumentBackend {
	return &DocumentBackend{
		Logger:          b.Logger.With(zap.String("handler", "document")),
		DocumentService: b.DocumentService,
	}
}

// DocumentHandler represents an HTTP API handler for documents.
type DocumentHandler struct {
	*httprouter.Router

	Logger *zap.Logger

	DocumentService influxdb.DocumentService
}

const (
	documentsPath            = "/api/v2/documents"
	documentsNamespacePath   = "/api/v2/documents/:ns"
	documentsNamespaceIDPath = "/api/v2/documents/:ns/:id"
)

// NewDocumentHandler returns a new instance of DocumentHandler.
func NewDocumentHandler(b *DocumentBackend) *DocumentHandler {
	h := &DocumentHandler{
		Router: NewRouter(),
		Logger: b.Logger,

		DocumentService: b.DocumentService,
	}

	h.HandlerFunc("POST", documentsPath, h.handlePostDocuments)
	h.HandlerFunc("POST", documentsNamespacePath, h.handlePostDocument)
	h.HandlerFunc("GET", documentsNamespacePath, h.handleGetDocuments)
	h.HandlerFunc("GET", documentsNamespaceIDPath, h.handleGetDocument)

	return h
}

// handlePostDocument is the HTTP handler for the POST /api/v2/documents route.
func (h *DocumentHandler) handlePostDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodePostDocumentsRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if _, err := h.DocumentService.CreateDocumentStore(ctx, req.Namespace); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusCreated, req); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

type postDocumentsRequest struct {
	Namespace string `json:"namespace"`
}

func decodePostDocumentsRequest(ctx context.Context, r *http.Request) (*postDocumentsRequest, error) {
	req := &postDocumentsRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	return req, nil
}

// handlePostDocument is the HTTP handler for the POST /api/v2/documents/:ns route.
func (h *DocumentHandler) handlePostDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodePostDocumentRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	s, err := h.DocumentService.FindDocumentStore(ctx, req.Namespace)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := s.CreateDocument(ctx, req.Document); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusCreated, req.Document); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

type postDocumentRequest struct {
	Namespace string
	Document  *influxdb.Document
}

func decodePostDocumentRequest(ctx context.Context, r *http.Request) (*postDocumentRequest, error) {
	doc := &influxdb.Document{}
	if err := json.NewDecoder(r.Body).Decode(doc); err != nil {
		return nil, err
	}

	params := httprouter.ParamsFromContext(ctx)
	ns := params.ByName("ns")
	if ns == "" {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing namespace",
		}
	}

	return &postDocumentRequest{
		Namespace: ns,
		Document:  doc,
	}, nil
}

// handleGetDocuments is the HTTP handler for the GET /api/v2/documents/:ns route.
func (h *DocumentHandler) handleGetDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeGetDocumentsRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	s, err := h.DocumentService.FindDocumentStore(ctx, req.Namespace)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	ds, err := s.FindDocuments(ctx, influxdb.FindOptions{})
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, ds); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

type getDocumentsRequest struct {
	Namespace string
}

func decodeGetDocumentsRequest(ctx context.Context, r *http.Request) (*getDocumentsRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	ns := params.ByName("ns")
	if ns == "" {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing namespace",
		}
	}

	return &getDocumentsRequest{
		Namespace: ns,
	}, nil
}

// handleGetDocument is the HTTP handler for the GET /api/v2/documents/:ns/:id route.
func (h *DocumentHandler) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeGetDocumentRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	s, err := h.DocumentService.FindDocumentStore(ctx, req.Namespace)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	d, err := s.FindDocumentsByID(ctx, req.ID)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, d); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}
}

type getDocumentRequest struct {
	Namespace string
	ID        influxdb.ID
}

func decodeGetDocumentRequest(ctx context.Context, r *http.Request) (*getDocumentRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	ns := params.ByName("ns")
	if ns == "" {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing namespace",
		}
	}

	i := params.ByName("id")
	if i == "" {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "url missing id",
		}
	}

	var id influxdb.ID
	if err := id.DecodeFromString(i); err != nil {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "bad id in url",
		}
	}

	return &getDocumentRequest{
		Namespace: ns,
		ID:        id,
	}, nil
}
