package vsphere

import (
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	api "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1alpha1"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1alpha1/ref"
	"github.com/konveyor/forklift-controller/pkg/controller/provider/web"
	model "github.com/konveyor/forklift-controller/pkg/controller/provider/web/vsphere"
)

//
// vSphere validator.
type Validator struct {
	plan *api.Plan
	inventory web.Client
}

//
// Load.
func (r *Validator) Load() (err error) {
	r.inventory, err = web.NewClient(r.plan.Referenced.Provider.Source)
	return
}

//
// Validate that a VM's networks have been mapped.
func (r *Validator) NetworksMapped(vmRef ref.Ref) (ok bool, err error) {
	vm := &model.VM{}
	pErr := r.inventory.Find(vm, vmRef)
	if pErr != nil {
		err = liberr.New(
			fmt.Sprintf(
				"VM %s lookup failed: %s",
				vmRef.String(),
				pErr.Error()))
		return
	}

	for _, net := range vm.Networks {
		if !r.plan.Referenced.Map.Network.Status.Refs.Find(ref.Ref{ID: net.ID}) {
			return
		}
	}
	ok = true
	return
}

//
// Validate that a VM's disk backing storage has been mapped.
func (r *Validator) StorageMapped(vmRef ref.Ref) (ok bool, err error) {
	vm := &model.VM{}
	pErr := r.inventory.Find(vm, vmRef)
	if pErr != nil {
		err = liberr.New(
			fmt.Sprintf(
				"VM %s lookup failed: %s",
				vmRef.String(),
				pErr.Error()))
		return
	}

	for _, disk := range vm.Disks {
		if !r.plan.Referenced.Map.Storage.Status.Refs.Find(ref.Ref{ID: disk.Datastore.ID}) {
			return
		}
	}
	ok = true
	return
}