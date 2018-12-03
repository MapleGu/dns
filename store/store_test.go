package store

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	ttl = 100
)

var (
	dirPath = ""
)

func TestMain(m *testing.M) {
	log.Println("test start")
	m.Run()
	log.Println("test end")
}

func newStoreNil() *Store {
	return NewStore("")
}

func newStore() *Store {
	book := new(Store)
	expired, _ := dnsmessage.NewName("expired")
	unexpired, _ := dnsmessage.NewName("unexpired")
	book.data = map[string]Entry{
		"expired": {
			Resources: []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  expired,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
						TTL:   ttl,
					},
					Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
				},
			},
			TTL:     ttl,
			Created: time.Now().Unix() - ttl*2,
		},
		"unexpired": {
			Resources: []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  unexpired,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
						TTL:   ttl,
					},
					Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
				},
			},
			TTL:     ttl,
			Created: time.Now().Unix(),
		},
	}
	book.rwDirPath = ""

	return book
}

func Test_get_key_not_exist(t *testing.T) {
	store := newStore()
	data, ok := store.Get("not_exist")
	if data != nil && ok != false {
		t.Errorf("%s \n", "Get a key that doesn't exist error")
	}
}

func Test_get_key_not_exist_but_expired(t *testing.T) {
	store := newStore()
	data, ok := store.Get("expired")
	if data != nil || ok != false {
		t.Errorf("%s \n", "Get a key that doesn't exist but expired error")
	}
}

func Test_get_key_exist_and_unexpired(t *testing.T) {
	store := newStore()
	data, ok := store.Get("unexpired")
	if data == nil || ok != true {
		t.Errorf("%s %+v \n", "Get a key that exist and unexpired error", data)
	}
}

func Test_set_key_not_exist(t *testing.T) {
	store := newStoreNil()
	exitStr := "not_exist"
	exit, _ := dnsmessage.NewName(exitStr)
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  exit,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
	}
	ok := store.Set(exitStr, res, nil)
	if ok == false {
		t.Errorf("%s \n", "set a key that doesn't exist before: set error")
	}

	data, ok := store.Get(exitStr)
	if ok != true || data == nil || RString(data[0]) != RString(res) {
		t.Errorf("%s %+v \n", "set a key that doesn't exist before error: get error", data)
	}
}

func Test_set_key_exist_but_old_not_exist(t *testing.T) {
	store := newStore()
	exitStr := "unexpired"
	exit, _ := dnsmessage.NewName(exitStr)
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  exit,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 2}},
	}
	ok := store.Set(exitStr, res, nil)
	if ok == false {
		t.Errorf("%s \n", "set a key that doesn't exist but old parm is nil: set error")
	}

	data, ok := store.Get(exitStr)
	if ok != true || data == nil || RString(data[len(data)-1]) != RString(res) {
		t.Errorf("%s %+v \n", "set a key that doesn't exist before error: get error", data)
	}
}

func Test_set_key_exist_and_old_exist(t *testing.T) {
	store := newStore()
	exitStr := "unexpired"
	exit, _ := dnsmessage.NewName(exitStr)
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  exit,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
	}
	ok := store.Set(exitStr, res, &res)
	if ok == false {
		t.Errorf("%s \n", "set a key that exist and old parm exist: set error")
	}

	data, ok := store.Get(exitStr)
	if ok != true || data == nil || RString(data[0]) != RString(res) {
		t.Errorf("%s %+v \n", "set a key that exist before error: get error", data)
	}
}

func Test_set_key_exist_and_old_exist_but_store_not_exist(t *testing.T) {
	store := newStore()
	exitStr := "unexpired"
	exit, _ := dnsmessage.NewName(exitStr)
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  exit,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 2}},
	}
	ok := store.Set(exitStr, res, &res)
	if ok == true {
		t.Errorf("%s \n", "set a key that exist and old parm exist but store not exist: set error")
	}

	data, ok := store.Get(exitStr)
	if ok != true || data == nil || RString(data[0]) == RString(res) {
		t.Errorf("%s %+v \n", "set a key that exist and old parm exist but store not exist: get error", data)
	}
}

func Test_override(t *testing.T) {
	store := newStore()
	exitStr := "unexpired"
	exit, _ := dnsmessage.NewName(exitStr)
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  exit,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 2}},
	}
	ms := []dnsmessage.Resource{res}
	store.Override(exitStr, ms)

	data, ok := store.Get(exitStr)
	if ok != true || data == nil || RString(data[0]) != RString(res) {
		t.Errorf("%s %+v \n", "set a key that exist and old parm exist but store not exist: get error", data)
	}
}

func Test_remote_key_not_exist(t *testing.T) {
	store := newStore()
	exitStr := "not_exist"
	ok := store.Remove(exitStr, nil)
	if ok == true {
		t.Errorf("%s\n", "remove a key that not exist")
	}
}

func Test_remote_key_exist_and_res_exist(t *testing.T) {
	store := newStore()
	unexpiredStr := "unexpired"
	unexpired, _ := dnsmessage.NewName("unexpired")
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  unexpired,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
	}

	ok := store.Remove(unexpiredStr, &res)
	if ok != true {
		t.Errorf("%s\n", "remove a key that not exist")
	}
}

func Test_remote_key_exist_but_res_not_exist(t *testing.T) {
	store := newStore()
	unexpiredStr := "unexpired"
	unexpired, _ := dnsmessage.NewName("unexpired")
	res := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  unexpired,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 2}},
	}

	ok := store.Remove(unexpiredStr, &res)
	if ok == true {
		t.Errorf("%s\n", "remove a key that not exist")
	}
}

func Test_store_save_load(t *testing.T) {
	dirPath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Skip(err)
	}
	name, _ := dnsmessage.NewName("test")
	data := map[string]Entry{
		"test": {
			Resources: []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
				},
			},
		},
	}
	book := Store{data: data, rwDirPath: dirPath}
	err = book.Save()
	if err != nil {
		t.Errorf("%s\n", err.Error())
	}
	main, _ := os.Stat(filepath.Join(dirPath, storeName))
	bk, _ := os.Stat(filepath.Join(dirPath, storeBkName))
	if main.Size() != 664 || bk.Size() != 0 {
		t.Fail()
	}
	bookNew := Store{rwDirPath: dirPath}
	err = bookNew.Load()
	if err != nil {
		t.Errorf("%s\n", err.Error())
	}
	r, _ := bookNew.Get("test")
	body, _ := r[0].Body.(*dnsmessage.AResource)
	if body.A != [4]byte{127, 0, 0, 1} {
		t.Fail()
	}
}
