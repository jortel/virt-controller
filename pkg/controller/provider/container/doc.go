package provider

import (
	libcontainer "github.com/konveyor/controller/pkg/inventory/container"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	"github.com/konveyor/controller/pkg/logging"
	api "github.com/konveyor/virt-controller/pkg/apis/virt/v1alpha1"
	"github.com/konveyor/virt-controller/pkg/controller/provider/container/vmware"
	core "k8s.io/api/core/v1"
)

//
// Create a new reconciler.
func New(p *api.Provider, s *core.Secret, db libmodel.DB) libcontainer.Reconciler {
	switch p.Spec.Type {
	case api.VMWare:
		return vmware.New(p.Spec.URL, s, db)
	default:
		return nil
	}
}

//
// Set the package logger.
func SetLogger(logger *logging.Logger) {
	vmware.Log = logger
}
