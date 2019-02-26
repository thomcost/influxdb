package influxdb

import (
	"context"
)

type DocumentService interface {
	CreateDocumentStore(ctx context.Context, name string) (DocumentStore, error)
	FindDocumentStore(ctx context.Context, name string) (DocumentStore, error)
}

type Document struct {
	ID   ID           `json:"id"`
	Meta DocumentMeta `json:"meta"`
	Data interface{}  `json:"data,omitempty"` // TODO(desa): maybe this needs to be json.Marshaller & json.Unmarshaler
}

type DocumentMeta struct {
	Name string `json:"name"`
}

type DocumentStore interface {
	CreateDocument(ctx context.Context, d *Document) error
	FindDocumentsByID(ctx context.Context, ids ...ID) ([]*Document, error)
	FindDocuments(ctx context.Context, opts FindOptions) ([]*Document, error) // TODO(desa): add additional search parameters.
	//UpdateDocument(ctx context.Context, id ID, d *Document) error
	DeleteDocuments(ctx context.Context, ids ...ID) error
}
