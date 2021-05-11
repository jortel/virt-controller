package ovirt

import (
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	"github.com/konveyor/forklift-controller/pkg/controller/provider/model/base"
)

//
// Errors
var NotFound = libmodel.NotFound

type InvalidRefError = base.InvalidRefError

//
// Types
type Model = base.Model
type ListOptions = base.ListOptions
type Ref = base.Ref

//
// Base oVirt model.
type Base struct {
	// Managed object ID.
	ID string `sql:"pk"`
	// Parent
	Parent Ref `sql:"d0,index(parent)"`
	// Name
	Name string `sql:"d0,index(name)"`
	// Revision
	Description string `sql:"d0"`
	// Revision
	Revision int64 `sql:"d0,index(revision)"`
}

//
// Get the PK.
func (m *Base) Pk() string {
	return m.ID
}

//
// String representation.
func (m *Base) String() string {
	return m.ID
}

//
// Get labels.
func (m *Base) Labels() libmodel.Labels {
	return nil
}

func (m *Base) Equals(other libmodel.Model) bool {
	if vm, cast := other.(*Base); cast {
		return m.ID == vm.ID
	}

	return false
}

//
// Updated.
// Increment revision. Should ONLY be called by
// the reconciler.
func (m *Base) Updated() {
	m.Revision++
}

type DataCenter struct {
	Base
}

type Cluster struct {
	Base
}

type Network struct {
	Base
	VLan         Ref      `sql:""`
	Usages       []string `sql:""`
	VNICProfiles []Ref    `sql:""`
}

type VNICProfile struct {
	Base
	QoS Ref `sql:""`
}

type StorageDomain struct {
	Base
	Type    string `sql:""`
	Storage struct {
		Type string
	} `sql:""`
	Available int64 `sql:""`
	Used      int64 `sql:""`
}

type Host struct {
	Base
}

type VM struct {
	Base
}
