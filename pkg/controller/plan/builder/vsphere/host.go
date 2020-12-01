package vsphere

import (
	"context"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/virt-controller/pkg/apis/virt/v1alpha1/mapped"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web/vsphere"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25"
	core "k8s.io/api/core/v1"
	"net/http"
	liburl "net/url"
)

//
// ESX Host.
type EsxHost struct {
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
func (r *EsxHost) networkID(in mapped.SourceObject) (out string, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = r.connect(ctx)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	network := &vsphere.Network{}
	status, pErr := r.inventory.Get(network, in.ID)
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
func (r *EsxHost) DatastoreID(in mapped.SourceObject) (out string, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = r.connect(ctx)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	ds := &vsphere.Datastore{}
	status, pErr := r.inventory.Get(ds, in.ID)
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
func (r *EsxHost) connect(ctx context.Context) (err error) {
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
func (r *EsxHost) user() string {
	if user, found := r.secret.Data["user"]; found {
		return string(user)
	}

	return ""
}

//
// Password.
func (r *EsxHost) password() string {
	if password, found := r.secret.Data["password"]; found {
		return string(password)
	}

	return ""
}
