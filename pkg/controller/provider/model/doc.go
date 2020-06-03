package model

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/virt-controller/pkg/controller/provider/model/vmware"
)

//
// Build all models.
func All() []interface{} {
	return []interface{}{
		&vmware.VM{},
	}
}

//
// Set the package logger.
func SetLogger(logger *logging.Logger) {
	vmware.Log = logger
}
