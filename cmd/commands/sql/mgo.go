package sql

var (
	mgoTemplates = map[string]string{
		"model": `
package models

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	{{.Name}} *_{{.Name}}

	{{.Name|lowercase}}Collection = "{{.Name|lowercase}}"
	{{.Name|lowercase}}Indexes = []mgo.Index{}
)

type {{.Name}}Model struct {
	ID			bson.ObjectId	` + "`" + `bson:"_id"` + "`" + ` 
	CreatedAt	time.Time 		` + "`" + `bson:"created_at"` + "`" + `
	UpdatedAt	time.Time 		` + "`" + `bson:"updated_at"` + "`" + `

	isNewRecord bool   ` + "`" + `bson:"-"` + "`" + `
}

func New{{.Name}}Model() *{{.Name}}Model {
	return &{{.Name}}Model{
		ID:          bson.NewObjectId(),
		isNewRecord: true,
	}
}

func (m *{{.Name}}Model) IsNewRecord() bool {
	return m.isNewRecord
}

func (m *{{.Name}}Model) Save() (err error) {
	if !m.ID.Valid() {
		err = errors.New("Invalid bson id")

		return
	}

	m.UpdatedAt = time.Now()

	{{.Name}}.Query(func(c *mgo.Collection){
		if m.IsNewRecord() {
			m.CreatedAt = m.UpdatedAt

			err = c.Insert(m)
			if err == nil {
				m.isNewRecord = false
			}
		} else {
			update := bson.M{
				"updated_at": m.UpdatedAt
			}

			err = c.UpdateId(m.ID, bson.M{ 
				"$set":  update,
			})
		}
	})

	return
}

// {{.Name}} model helpers
type _{{.Name}} struct{}

// Find returns {{.Name}}Model with id supplied
func (_ *_{{.Name}}) Find(id string) (m *{{.Name}}Model, err error) {
	if !bson.IsObjectIdHex(id) {
		err = errors.New("Invalid bson id")

		return
	}

	query := bson.M{
		"_id": bson.ObjectIdHex(id),
	}

	{{.Name}}.Query(func(c *mgo.Collection){
		err = c.Find(query).One(&m)
	})
	
	return
}

func (_ *_{{.Name}}) Query(query func(c *mgo.Collection)) {
	model.Query({{.Name|lowercase}}Collection, {{.Name|lowercase}}Indexes, query)
}
`,
		"model_test": `
package models

import(
	"testing"

	"github.com/golib/assert"
)

func Test_{{.Name}}(t *testing.T) {
	assertion := assert.New(t)

	m := New{{.Name}}Model()
	assertion.True(m.IsNewRecord())

	err := m.Save()
	assertion.Nil(err)
	assertion.False(m.IsNewRecord())
}

func Test_{{.Name}}_Find(t *testing.T) {
	assertion := assert.New(t)

	m := New{{.Name}}Model()

	// it should be error
	tmp, err := {{.Name}}.Find(m.ID.Hex())
	assertion.NotNil(err)
	assertion.Nil(tmp)

	// it should work
	m.Save()

	tmp, err = {{.Name}}.Find(m.ID.Hex())
	assertion.Nil(err)
	assertion.Equal(m.ID.Hex(), tmp.ID.Hex())
}
`}
)
