package web

import (
	"github.com/konveyor/controller/pkg/inventory/container"
	libweb "github.com/konveyor/controller/pkg/inventory/web"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web/vmware"
)

//
// Build all handlers.
func All(container *container.Container) []libweb.RequestHandler {
	return []libweb.RequestHandler{
		&libweb.SchemaHandler{},
		&vmware.VMHandler{
			Base: vmware.Base{
				Container: container,
			},
		},
	}
}

//
// Set the package logger.
func SetLogger(logger *logging.Logger) {
	vmware.Log = logger
}
