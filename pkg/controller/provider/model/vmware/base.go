package vmware

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/vmware/govmomi/vim25/types"
)

// Shared logger.
var Log *logging.Logger

// Errors
var NotFound = libmodel.NotFound
var Conflict = libmodel.Conflict

func init() {
	logger := logging.WithName("")
	logger.Reset()
	Log = &logger
}

//
// Model interface
type Model interface {
	libmodel.Model
	With([]types.PropertyChange)
}

//
// Base VMWare model.
type Base struct {
	// Primary key (digest).
	PK string `sql:"pk"`
	// Provider
	ID string `sql:"key,unique(a)"`
	// The raw json-encoded object.
	Object string `sql:""`
}

//
// Get the PK.
func (m *Base) Pk() string {
	return m.PK
}

//
// Set the primary key.
func (m *Base) SetPk() {
	h := sha1.New()
	h.Write([]byte(m.ID))
	m.PK = fmt.Sprintf("%x", h.Sum(nil))
}

func (m *Base) String() string {
	return m.ID
}

func (m *Base) Labels() libmodel.Labels {
	return nil
}

func (m *Base) With(changeSet []types.PropertyChange) {
	object := map[string]interface{}{}
	for _, p := range changeSet {
		switch p.Op {
		// TODO:
		default:
			object[p.Name] = p.Val
		}
	}
	j, _ := json.Marshal(object)
	m.Object = string(j)
}
