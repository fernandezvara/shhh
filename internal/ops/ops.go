package ops

import (
	"encoding/json"

	"github.com/asdine/storm/v3"
)

type Ops struct {
	db   *storm.DB
	salt []byte
	key  []byte
}

func Open(filename string, isNew bool) (ops *Ops, err error) {

	if isNew && exists(filename) {
		err = ErrAlreadyExist
	}

	if !isNew {
		if !exists(filename) {
			err = ErrNotExist
		}
	}

	if err != nil {
		return
	}

	ops = &Ops{}
	ops.db, err = storm.Open(filename)
	if err != nil {
		return
	}

	err = ops.db.Init(&Entry{})

	return

}

func (o *Ops) Close() error {

	return o.db.Close()

}

func (o *Ops) GetSalt() (err error) {

	err = o.db.Get("config", "salt", &o.salt)
	return

}

func (o *Ops) SetSalt() (err error) {

	o.salt, err = newSalt()
	if err != nil {
		return
	}

	err = o.db.Set("config", "salt", o.salt)
	return

}

func (o *Ops) SetKey(password string) (err error) {

	var (
		cipheredSTDIN []byte
	)

	o.key, err = keyFromPassword([]byte(password), o.salt)
	if err != nil {
		return
	}

	cipheredSTDIN, err = o.encrypt([]byte(passwdOK))
	if err != nil {
		return
	}

	err = o.db.Set("config", "ok", &cipheredSTDIN)
	return

}

func (o *Ops) GetKey(password string) (err error) {

	o.key, err = keyFromPassword([]byte(password), o.salt)
	if err != nil {
		return
	}

	err = o.passMatch()

	return

}

func (o *Ops) passMatch() (err error) {

	var (
		cipheredDB []byte
		plainText  []byte
	)

	err = o.db.Get("config", "ok", &cipheredDB)
	if err != nil {
		return
	}

	plainText, err = o.decrypt([]byte(cipheredDB))
	if err != nil {
		err = ErrPasswordNotMatch
		return
	}

	if string(plainText) != passwdOK {
		err = ErrPasswordNotMatch
	}

	return

}

func (o *Ops) Get(id, key string) (values map[string][]byte, err error) {

	var (
		entry        Entry
		idCipher     []byte
		valCipher    []byte
		valuesSource map[string][]byte
	)

	values = make(map[string][]byte)

	idCipher = toSHA256(id)

	err = o.db.One("ID", idCipher, &entry)
	if err != nil {
		return
	}

	valCipher, err = o.decrypt(entry.Value)
	if err != nil {
		return
	}

	err = json.Unmarshal(valCipher, &valuesSource)
	if err != nil {
		return
	}

	for k, v := range valuesSource {
		if key == "" || k == key {
			values[k] = v
		}
	}

	return

}

func (o *Ops) Set(id, key string, value []byte) (err error) {

	var (
		values    map[string][]byte
		val       []byte
		valCipher []byte
		entry     Entry
		idCipher  []byte
	)

	// first get current entries
	values, err = o.Get(id, "")
	switch err {
	case storm.ErrNotFound:
		values = make(map[string][]byte)
	case nil:
		// ok
	default:
		return
	}

	values[key] = value

	val, err = json.Marshal(values)
	if err != nil {
		return
	}

	valCipher, err = o.encrypt(val)
	if err != nil {
		return
	}

	idCipher = toSHA256(id)

	entry.ID = idCipher
	entry.Value = valCipher

	err = o.db.Save(&entry) // TODO: verify if Update must be used
	return

}

func (o *Ops) Delete(id string) (err error) {

	var idCipher []byte = toSHA256(id)

	err = o.db.Delete("Entry", idCipher)
	return

}

func (o *Ops) DeleteKey(id, key string) (err error) {

	var (
		values    map[string][]byte
		val       []byte
		valCipher []byte
		entry     Entry
		idCipher  []byte
	)

	// first get current entries
	values, err = o.Get(id, "")
	if err != nil {
		return
	}

	_, ok := values[key]
	if !ok {
		err = ErrNotExist
		return
	}

	delete(values, key)

	val, err = json.Marshal(values)
	if err != nil {
		return
	}

	valCipher, err = o.encrypt(val)
	if err != nil {
		return
	}

	idCipher = toSHA256(id)

	entry.ID = idCipher
	entry.Value = valCipher

	err = o.db.Save(&entry)
	return

}
