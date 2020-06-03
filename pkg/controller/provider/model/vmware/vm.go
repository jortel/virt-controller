package vmware

import (
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
)

type VM struct {
	Base
}

func (m *VM) Equals(other libmodel.Model) bool {
	if vm, cast := other.(*VM); cast {
		return m.ID == vm.ID
	}

	return false
}
