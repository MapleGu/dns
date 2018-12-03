// Copy FROM https://github.com/owlwalks/rind/blob/master/store.go
// Change some code that looks not beauty

package store

import (
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	storeName   string = "store"
	storeBkName string = "store_bk"
)

func init() {
	gob.Register(&dnsmessage.AResource{})
	gob.Register(&dnsmessage.CNAMEResource{})
}

// Store store struct
type Store struct {
	sync.RWMutex
	data      map[string]Entry
	rwDirPath string
}

// NewStore new store
func NewStore(rwDirPath string) *Store {
	return &Store{
		rwDirPath: rwDirPath,
		data:      make(map[string]Entry),
	}
}

// Entry entry struct
type Entry struct {
	Resources []dnsmessage.Resource
	TTL       uint32
	Created   int64
}

// Get get daya by key
func (s *Store) Get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	if !ok {
		return nil, false
	}

	now := time.Now().Unix()
	if e.TTL > 1 && (e.Created+int64(e.TTL) < now) {
		s.Remove(key, nil)
		return nil, false
	}
	return e.Resources, ok
}

// Set data
func (s *Store) Set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.data[key]; !ok {
		e := Entry{
			Resources: []dnsmessage.Resource{resource},
			TTL:       resource.Header.TTL,
			Created:   time.Now().Unix(),
		}
		s.data[key] = e
		return true
	}

	if old == nil {
		e := s.data[key]
		e.Resources = append(e.Resources, resource)
		s.data[key] = e
		return true
	}

	for i, rec := range s.data[key].Resources {
		if RString(rec) == RString(*old) {
			s.data[key].Resources[i] = resource
			return true
		}
	}

	return false
}

// Override override data
func (s *Store) Override(key string, resources []dnsmessage.Resource) {
	s.Lock()
	defer s.Unlock()

	e := Entry{
		Resources: resources,
		Created:   time.Now().Unix(),
	}
	if len(resources) > 0 {
		e.TTL = resources[0].Header.TTL
	}
	s.data[key] = e
}

// Remove remove data
func (s *Store) Remove(key string, r *dnsmessage.Resource) bool {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.data[key]; !ok {
		return false
	}

	if r == nil {
		delete(s.data, key)
		return true
	}

	rs := RString(*r)
	for i, rec := range s.data[key].Resources {
		if RString(rec) == rs {
			e := s.data[key]
			copy(e.Resources[i:], e.Resources[i+1:])
			var blank dnsmessage.Resource
			e.Resources[len(e.Resources)-1] = blank // to be garbage collected
			e.Resources = e.Resources[:len(e.Resources)-1]
			s.data[key] = e
			return true
		}
	}

	return false
}

// Save save to file
func (s *Store) Save() error {
	bk, err := os.OpenFile(filepath.Join(s.rwDirPath, storeBkName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return errors.Wrap(err, err.Error())
	}
	defer bk.Close()

	dst, err := os.OpenFile(filepath.Join(s.rwDirPath, storeName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return errors.Wrap(err, err.Error())
	}
	defer dst.Close()

	// backing up current store
	_, err = io.Copy(bk, dst)
	if err != nil {
		return errors.Wrap(err, err.Error())
	}

	enc := gob.NewEncoder(dst)
	book := s.Clone()
	err = enc.Encode(book)
	if err != nil {
		return errors.Wrap(err, err.Error())
	}
	return nil
}

// Load load from file
func (s *Store) Load() error {
	fReader, err := os.Open(filepath.Join(s.rwDirPath, storeName))
	if err != nil {
		return errors.Wrap(err, err.Error())
	}
	defer fReader.Close()

	dec := gob.NewDecoder(fReader)

	s.Lock()
	defer s.Unlock()

	err = dec.Decode(&s.data)
	if err != nil {
		return errors.Wrap(err, err.Error())
	}
	return nil
}

// Clone clone all data
func (s *Store) Clone() map[string]Entry {
	cp := make(map[string]Entry)
	s.RLock()
	for k, v := range s.data {
		cp[k] = v
	}
	s.RUnlock()
	return cp
}

// RString resource to string
func RString(r dnsmessage.Resource) string {
	var sb strings.Builder
	sb.Write(r.Header.Name.Data[:])
	sb.WriteString(r.Header.Type.String())
	sb.WriteString(r.Body.GoString())

	return sb.String()
}
