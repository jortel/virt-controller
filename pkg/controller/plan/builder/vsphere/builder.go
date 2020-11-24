package vsphere

import (
	"context"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	libitr "github.com/konveyor/controller/pkg/itinerary"
	api "github.com/konveyor/virt-controller/pkg/apis/virt/v1alpha1"
	"github.com/konveyor/virt-controller/pkg/apis/virt/v1alpha1/plan"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web/vsphere"
	vmio "github.com/kubevirt/vm-import-operator/pkg/apis/v2v/v1beta1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25"
	"gopkg.in/yaml.v2"
	core "k8s.io/api/core/v1"
	"net/http"
	liburl "net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
// vSphere builder.
type Builder struct {
	// Client.
	Client client.Client
	// Provider API client.
	Inventory web.Client
	// Source provider.
	Provider *api.Provider
	// Host map.
	HostMap map[string]*api.Host
}

//
// Build the VMIO secret.
func (r *Builder) Secret(vmID string, in, object *core.Secret) (err error) {
	url := r.Provider.Spec.URL
	hostID, err := r.hostID(vmID)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if host, found := r.HostMap[hostID]; found {
		hostURL := liburl.URL{
			Scheme: "https",
			Host:   host.Spec.IpAddress,
			Path:   vim25.Path,
		}
		url = hostURL.String()
		hostSecret, nErr := r.hostSecret(host)
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		h, nErr := r.host(hostID)
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		hostSecret.Data["thumbprint"] = []byte(h.Thumbprint)
		in = hostSecret
	}
	content, mErr := yaml.Marshal(
		map[string]string{
			"apiUrl":     url,
			"username":   string(in.Data["user"]),
			"password":   string(in.Data["password"]),
			"thumbprint": string(in.Data["thumbprint"]),
		})
	if mErr != nil {
		err = liberr.Wrap(mErr)
		return
	}
	object.StringData = map[string]string{
		"vmware": string(content),
	}

	return
}

//
// Find host ID for VM.
func (r *Builder) hostID(vmID string) (hostID string, err error) {
	vm := &vsphere.VM{}
	status, pErr := r.Inventory.Get(vm, vmID)
	if pErr != nil {
		err = liberr.Wrap(pErr)
		return
	}
	switch status {
	case http.StatusOK:
		hostID = vm.Host.ID
	default:
		err = liberr.New(
			fmt.Sprintf(
				"VM %s lookup failed: %s",
				vmID,
				http.StatusText(status)))
	}

	return
}

//
// Find host CR secret.
func (r *Builder) hostSecret(host *api.Host) (secret *core.Secret, err error) {
	ref := host.Spec.Secret
	secret = &core.Secret{}
	err = r.Client.Get(
		context.TODO(),
		client.ObjectKey{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		secret)
	if err != nil {
		err = liberr.Wrap(err)
	}

	return
}

//
// Find host in the inventory.
func (r *Builder) host(hostID string) (host *vsphere.Host, err error) {
	host = &vsphere.Host{}
	status, err := r.Inventory.Get(host, hostID)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	switch status {
	case http.StatusOK:
	default:
		err = liberr.New(
			fmt.Sprintf(
				"Host %s lookup failed: %s",
				hostID,
				http.StatusText(status)))
		return
	}

	return
}

//
// Build the VMIO ResourceMapping CR.
func (r *Builder) mapping(vmID string, mp *plan.Map, object *vmio.VirtualMachineImportSpec) (err error) {
	translator, err := r.hostTranslator(vmID)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	netMap := []vmio.NetworkResourceMappingItem{}
	dsMap := []vmio.StorageResourceMappingItem{}
	for i := range mp.Networks {
		network := &mp.Networks[i]
		id := &network.Source.ID
		if translator != nil {
			netID, tErr := translator.networkID(*id)
			if tErr != nil {
				err = liberr.Wrap(tErr)
				return
			}
			id = &netID
		}
		netMap = append(
			netMap,
			vmio.NetworkResourceMappingItem{
				Source: vmio.Source{
					ID: id,
				},
				Target: vmio.ObjectIdentifier{
					Namespace: &network.Destination.Namespace,
					Name:      network.Destination.Name,
				},
				Type: &network.Destination.Type,
			})
	}
	for i := range mp.Datastores {
		ds := &mp.Datastores[i]
		id := &ds.Source.ID
		if translator != nil {
			dsID, tErr := translator.DatastoreID(*id)
			if tErr != nil {
				err = liberr.Wrap(tErr)
				return
			}
			id = &dsID
		}
		dsMap = append(
			dsMap,
			vmio.StorageResourceMappingItem{
				Source: vmio.Source{
					ID: &ds.Source.ID,
				},
				Target: vmio.ObjectIdentifier{
					Name: ds.Destination.StorageClass,
				},
			})
	}
	object.Source.Vmware.Mappings = &vmio.VmwareMappings{
		NetworkMappings: &netMap,
		StorageMappings: &dsMap,
	}

	return
}

//
// Build a host translator.
func (r *Builder) hostTranslator(vmID string) (tr *HostTranslator, err error) {
	hostID, err := r.hostID(vmID)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if host, found := r.HostMap[hostID]; found {
		hostURL := liburl.URL{
			Scheme: "https",
			Host:   host.Spec.IpAddress,
			Path:   vim25.Path,
		}
		secret, nErr := r.hostSecret(host)
		if nErr != nil {
			err = liberr.Wrap(nErr)
			return
		}
		tr = &HostTranslator{
			url:       hostURL.String(),
			inventory: r.Inventory,
			secret:    secret,
		}
	}

	return
}

//
// Build the VMIO VM Import Spec.
func (r *Builder) Import(vmID string, mp *plan.Map, object *vmio.VirtualMachineImportSpec) (err error) {
	vm := &vsphere.VM{}
	status, pErr := r.Inventory.Get(vm, vmID)
	if pErr != nil {
		err = liberr.Wrap(pErr)
		return
	}
	switch status {
	case http.StatusOK:
		uuid := vm.UUID
		object.TargetVMName = &vm.Name
		object.Source.Vmware = &vmio.VirtualMachineImportVmwareSourceSpec{
			VM: vmio.VirtualMachineImportVmwareSourceVMSpec{
				ID: &uuid,
			},
		}
		err = r.mapping(vmID, mp, object)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	default:
		err = liberr.New(
			fmt.Sprintf(
				"VM %s lookup failed: %s",
				vmID,
				http.StatusText(status)))
	}

	return
}

//
// Build tasks.
func (r *Builder) Tasks(vmID string) (list []*plan.Task, err error) {
	vm := &vsphere.VM{}
	status, pErr := r.Inventory.Get(vm, vmID)
	if pErr != nil {
		err = liberr.Wrap(pErr)
		return
	}
	switch status {
	case http.StatusOK:
		for _, disk := range vm.Disks {
			mB := disk.Capacity / 0x100000
			list = append(
				list,
				&plan.Task{
					Name: disk.File,
					Progress: libitr.Progress{
						Total: mB,
					},
					Annotations: map[string]string{
						"unit": "MB",
					},
				})
		}
	default:
		err = liberr.New(
			fmt.Sprintf(
				"VM %s lookup failed: %s",
				vmID,
				http.StatusText(status)))
	}

	return
}

//
// Host translator.
type HostTranslator struct {
	// Host url.
	url string
	// Host secret.
	secret *core.Secret
	// Inventory client.
	inventory web.Client
	// Host client.
	client *vim25.Client
	// Finder
	finder *find.Finder
}

//
// Translate network ID.
func (r *HostTranslator) networkID(in string) (out string, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = r.connect(ctx)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	network := &vsphere.Network{}
	status, pErr := r.inventory.Get(network, in)
	if pErr != nil {
		err = liberr.Wrap(pErr)
		return
	}
	if status != http.StatusOK {
		err = liberr.New(http.StatusText(status))
		return
	}
	object, fErr := r.finder.Network(ctx, network.Path)
	if fErr != nil {
		err = liberr.Wrap(fErr)
		return
	}
	out = object.Reference().Value

	return
}

//
// Translate datastore ID.
func (r *HostTranslator) DatastoreID(in string) (out string, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = r.connect(ctx)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	ds := &vsphere.Datastore{}
	status, pErr := r.inventory.Get(ds, in)
	if pErr != nil {
		err = liberr.Wrap(pErr)
		return
	}
	if status != http.StatusOK {
		err = liberr.New(http.StatusText(status))
		return
	}
	object, fErr := r.finder.Datastore(ctx, ds.Path)
	if fErr != nil {
		err = liberr.Wrap(fErr)
		return
	}
	out = object.Reference().Value

	return
}

//
// Build the client and finder.
func (r *HostTranslator) connect(ctx context.Context) (err error) {
	insecure := true
	if r.client != nil {
		return
	}
	url, err := liburl.Parse(r.url)
	if err != nil {
		return liberr.Wrap(err)
	}
	url.User = liburl.UserPassword(
		r.user(),
		r.password())

	client, err := govmomi.NewClient(ctx, url, insecure)
	if err != nil {
		return liberr.Wrap(err)
	}
	r.client, err = vim25.NewClient(ctx, client)
	if err != nil {
		err = liberr.Wrap(err)
	}

	r.finder = find.NewFinder(r.client)

	return nil
}

//
// User.
func (r *HostTranslator) user() string {
	if user, found := r.secret.Data["user"]; found {
		return string(user)
	}

	return ""
}

//
// Password.
func (r *HostTranslator) password() string {
	if password, found := r.secret.Data["password"]; found {
		return string(password)
	}

	return ""
}
