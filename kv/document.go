package kv

import (
	"context"
	"encoding/json"
	"path"

	"github.com/influxdata/influxdb"
)

const (
	documentDataBucket = "/documents/data"
	documentMetaBucket = "/documents/meta"
)

func (s *Service) CreateDocumentStore(ctx context.Context, ns string) (influxdb.DocumentStore, error) {
	// TODO(desa): keep track of which namespaces exist.
	var ds influxdb.DocumentStore

	err := s.kv.Update(func(tx Tx) error {
		if _, err := tx.Bucket([]byte(path.Join(ns, documentDataBucket))); err != nil {
			return err
		}

		if _, err := tx.Bucket([]byte(path.Join(ns, documentMetaBucket))); err != nil {
			return err
		}

		ds = &DocumentStore{
			namespace: ns,
			service:   s,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (s *Service) FindDocumentStore(ctx context.Context, ns string) (influxdb.DocumentStore, error) {
	var ds influxdb.DocumentStore

	err := s.kv.View(func(tx Tx) error {
		if _, err := tx.Bucket([]byte(path.Join(ns, documentDataBucket))); err != nil {
			return err
		}

		if _, err := tx.Bucket([]byte(path.Join(ns, documentMetaBucket))); err != nil {
			return err
		}

		ds = &DocumentStore{
			namespace: ns,
			service:   s,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ds, nil
}

type DocumentStore struct {
	service   *Service
	namespace string
}

func (s *DocumentStore) CreateDocument(ctx context.Context, d *influxdb.Document) error {
	return s.service.kv.Update(func(tx Tx) error {
		return s.service.createDocument(ctx, tx, s.namespace, d)
	})
}

func (s *Service) createDocument(ctx context.Context, tx Tx, ns string, d *influxdb.Document) error {
	d.ID = s.IDGenerator.ID()

	if err := s.putDocument(ctx, tx, ns, d); err != nil {
		return err
	}

	return nil
}

func (s *Service) putDocument(ctx context.Context, tx Tx, ns string, d *influxdb.Document) error {
	if err := s.putDocumentMeta(ctx, tx, ns, d.ID, &d.Meta); err != nil {
		return err
	}

	if err := s.putDocumentData(ctx, tx, ns, d.ID, d.Data); err != nil {
		return err
	}

	// TODO(desa): implement
	//if err := s.indexDocumentMeta(ctx, tx, d.ID, d.Meta); err != nil {
	//}

	return nil
}

func (s *Service) putAtID(ctx context.Context, tx Tx, bucket string, id influxdb.ID, i interface{}) error {
	v, err := json.Marshal(i)
	if err != nil {
		return err
	}

	k, err := id.Encode()
	if err != nil {
		return err
	}

	b, err := tx.Bucket([]byte(bucket))
	if err != nil {
		return err
	}

	if err := b.Put(k, v); err != nil {
		return err
	}

	return nil
}

func (s *Service) putDocumentData(ctx context.Context, tx Tx, ns string, id influxdb.ID, data interface{}) error {
	return s.putAtID(ctx, tx, path.Join(ns, documentDataBucket), id, data)
}

func (s *Service) putDocumentMeta(ctx context.Context, tx Tx, ns string, id influxdb.ID, m *influxdb.DocumentMeta) error {
	return s.putAtID(ctx, tx, path.Join(ns, documentMetaBucket), id, m)
}

func (s *DocumentStore) FindDocumentsByID(ctx context.Context, ids ...influxdb.ID) ([]*influxdb.Document, error) {
	ds := make([]*influxdb.Document, 0, len(ids))

	err := s.service.kv.View(func(tx Tx) error {
		for _, id := range ids {
			d, err := s.service.findDocumentByID(ctx, tx, s.namespace, id)
			if err != nil {
				return err
			}

			ds = append(ds, d)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (s *Service) findDocumentByID(ctx context.Context, tx Tx, ns string, id influxdb.ID) (*influxdb.Document, error) {
	m, err := s.findDocumentMetaByID(ctx, tx, ns, id)
	if err != nil {
		return nil, err
	}

	d, err := s.findDocumentDataByID(ctx, tx, ns, id)
	if err != nil {
		return nil, err
	}

	return &influxdb.Document{
		ID:   id,
		Meta: *m,
		Data: d,
	}, nil
}

func (s *Service) findByID(ctx context.Context, tx Tx, bucket string, id influxdb.ID, i interface{}) error {
	b, err := tx.Bucket([]byte(bucket))
	if err != nil {
		return err
	}

	k, err := id.Encode()
	if err != nil {
		return err
	}

	v, err := b.Get(k)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(v, i); err != nil {
		return err
	}

	return nil
}

func (s *Service) findDocumentMetaByID(ctx context.Context, tx Tx, ns string, id influxdb.ID) (*influxdb.DocumentMeta, error) {
	m := &influxdb.DocumentMeta{}

	if err := s.findByID(ctx, tx, path.Join(ns, documentMetaBucket), id, m); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Service) findDocumentDataByID(ctx context.Context, tx Tx, ns string, id influxdb.ID) (interface{}, error) {
	var data interface{}
	if err := s.findByID(ctx, tx, path.Join(ns, documentDataBucket), id, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *DocumentStore) FindDocuments(ctx context.Context, opts influxdb.FindOptions) ([]*influxdb.Document, error) {
	var ds []*influxdb.Document
	err := s.service.kv.View(func(tx Tx) error {
		return s.service.findDocuments(ctx, tx, s.namespace, &ds)
	})

	if err != nil {
		return nil, err
	}

	return ds, nil
}

// TODO(desa): support find options
func (s *Service) findDocuments(ctx context.Context, tx Tx, ns string, ds *[]*influxdb.Document) error {
	b, err := tx.Bucket([]byte(path.Join(ns, documentMetaBucket)))
	if err != nil {
		return err
	}

	cur, err := b.Cursor()
	if err != nil {
		return err
	}

	for k, v := cur.First(); len(k) != 0; k, v = cur.Next() {
		d := &influxdb.Document{}
		if err := d.ID.Decode(k); err != nil {
			return err
		}

		if err := json.Unmarshal(v, &d.Meta); err != nil {
			return err
		}

		*ds = append(*ds, d)
	}

	return nil
}

//func (s *Service) UpdateDocumentData(ctx context.Context, id influxdb.ID, data interface{}) error {
//	return s.kv.Update(func(tx Tx) error {
//		return nil
//	})
//}

func (s *DocumentStore) DeleteDocuments(ctx context.Context, ids ...influxdb.ID) error {
	return s.service.kv.Update(func(tx Tx) error {
		for _, id := range ids {
			if err := s.service.deleteDocument(ctx, tx, s.namespace, id); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) deleteDocument(ctx context.Context, tx Tx, ns string, id influxdb.ID) error {
	m, err := s.findDocumentMetaByID(ctx, tx, ns, id)
	if err != nil {
		return err
	}

	if err := s.deleteDocumentMeta(ctx, tx, ns, id); err != nil {
		return err
	}

	if err := s.deleteDocumentData(ctx, tx, ns, id); err != nil {
		return err
	}

	// TODO(desa): do this eventually
	_ = m
	//if err := s.deindexDocumentMeta(ctx, tx, id, m); err != nil {
	//	return err
	//}

	return nil
}

func (s *Service) deleteAtID(ctx context.Context, tx Tx, bucket string, id influxdb.ID) error {
	k, err := id.Encode()
	if err != nil {
		return err
	}

	b, err := tx.Bucket([]byte(bucket))
	if err != nil {
		return err
	}

	if err := b.Delete(k); err != nil {
		return err
	}

	return nil
}

func (s *Service) deleteDocumentData(ctx context.Context, tx Tx, ns string, id influxdb.ID) error {
	return s.deleteAtID(ctx, tx, path.Join(ns, documentDataBucket), id)
}

func (s *Service) deleteDocumentMeta(ctx context.Context, tx Tx, ns string, id influxdb.ID) error {
	return s.deleteAtID(ctx, tx, path.Join(ns, documentMetaBucket), id)
}

func ErrInvalidDocument() *influxdb.Error {
	// TODO(desa): get the json result errors
	return &influxdb.Error{
		Code: influxdb.EInvalid,
		Msg:  "document does not fit schema",
	}
}
