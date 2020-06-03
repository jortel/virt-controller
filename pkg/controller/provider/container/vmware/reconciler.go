package vmware

import (
	"context"
	"errors"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	"github.com/konveyor/controller/pkg/logging"
	model "github.com/konveyor/virt-controller/pkg/controller/provider/model/vmware"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	core "k8s.io/api/core/v1"
	liburl "net/url"
	"time"
)

var Log *logging.Logger

const (
	VirtualMachine = "VirtualMachine"
	DataCenter     = "Datacenter"
)

const (
	Folder          = "Folder"
	VmFolder        = "vmFolder"
	ChildEntity     = "childEntity"
	TraverseFolders = "traverseFolders"
)

//
// Actions
const (
	Enter  = "enter"
	Leave  = "leave"
	Modify = "modify"
)

//
// Datacenter traversal Spec.
var TsDataCenter = &types.TraversalSpec{
	Type: DataCenter,
	Path: VmFolder,
	SelectSet: []types.BaseSelectionSpec{
		&types.SelectionSpec{
			Name: TraverseFolders,
		},
	},
}

//
// Root Folder traversal Spec
var TsRootFolder = &types.TraversalSpec{
	SelectionSpec: types.SelectionSpec{
		Name: TraverseFolders,
	},
	Type: Folder,
	Path: ChildEntity,
	SelectSet: []types.BaseSelectionSpec{
		TsDataCenter,
	},
}

func init() {
	logger := logging.WithName("")
	logger.Reset()
	Log = &logger
}

//
// A VMWare reconciler.
type Reconciler struct {
	// The vcenter host.
	host string
	// Credentials secret: {user:,password}.
	secret *core.Secret
	// DB client.
	db libmodel.DB
	// client.
	client *govmomi.Client
	// cancel function.
	cancel func()
	// has consistency
	consistent bool
}

//
// New reconciler.
func New(host string, secret *core.Secret, db libmodel.DB) *Reconciler {
	return &Reconciler{
		host:   host,
		secret: secret,
		db:     db,
	}
}

//
// The name.
func (r *Reconciler) Name() string {
	return r.host
}

//
// Get the DB.
func (r *Reconciler) DB() libmodel.DB {
	return r.db
}

//
// Reset.
func (r *Reconciler) Reset() {
	r.consistent = false
}

//
// Reset.
func (r *Reconciler) HasConsistency() bool {
	return r.consistent
}

//
// Start the reconciler.
func (r *Reconciler) Start() error {
	err := r.db.Open(true)
	if err != nil {
		Log.Trace(err)
		return err
	}
	ctx := context.Background()
	ctx, r.cancel = context.WithCancel(ctx)
	err = r.connect(ctx)
	if err != nil {
		Log.Trace(err)
		return err
	}
	run := func() {
		Log.Info("Started.", "name", r.Name())
		r.getUpdates(ctx)
		r.client.Logout(ctx)
		Log.Info("Shutdown.", "name", r.Name())
	}

	go run()

	return nil
}

//
// Shutdown the reconciler.
func (r *Reconciler) Shutdown(purge bool) {
	r.db.Close(true)
	if r.cancel != nil {
		r.cancel()
	}
}

//
// Get updates.
func (r *Reconciler) getUpdates(ctx context.Context) error {
	pc := property.DefaultCollector(r.client.Client)
	pc, err := pc.Create(ctx)
	if err != nil {
		return err
	}
	defer pc.Destroy(context.Background())
	filter := r.filter(pc)
	err = pc.CreateFilter(ctx, filter.CreateFilter)
	if err != nil {
		return err
	}
	req := types.WaitForUpdatesEx{
		This:    pc.Reference(),
		Options: filter.Options,
	}
	for {
		res, err := methods.WaitForUpdatesEx(ctx, r.client, &req)
		if err != nil {
			if ctx.Err() == context.Canceled {
				pc.CancelWaitForUpdates(context.Background())
				break
			}
			return err
		}
		updateSet := res.Returnval
		if updateSet == nil {
			break
		}
		req.Version = updateSet.Version
		for _, fs := range updateSet.FilterSet {
			r.apply(ctx, fs.ObjectSet)
		}
		if updateSet.Truncated == nil || !*updateSet.Truncated {
			r.consistent = true
		}
	}

	return nil
}

//
// Build the client.
func (r *Reconciler) connect(ctx context.Context) error {
	insecure := true
	url := &liburl.URL{
		Scheme: "https",
		User:   liburl.UserPassword(r.user(), r.password()),
		Host:   r.host,
		Path:   vim25.Path,
	}
	client, err := govmomi.NewClient(ctx, url, insecure)
	if err != nil {
		return err
	}

	r.client = client

	return nil
}

//
// User.
func (r *Reconciler) user() string {
	user := string(r.secret.Data["user"])
	return user
}

//
// Password.
func (r *Reconciler) password() string {
	password := string(r.secret.Data["password"])
	return password
}

//
// Build the object Spec filter.
func (r *Reconciler) filter(pc *property.Collector) *property.WaitFilter {
	return &property.WaitFilter{
		CreateFilter: types.CreateFilter{
			This: pc.Reference(),
			Spec: types.PropertyFilterSpec{
				ObjectSet: []types.ObjectSpec{
					r.objectSpec(),
				},
				PropSet: r.propertySpec(),
			},
		},
		Options: &types.WaitOptions{},
	}
}

//
// Build the object Spec.
func (r *Reconciler) objectSpec() types.ObjectSpec {
	return types.ObjectSpec{
		Obj: r.client.ServiceContent.RootFolder,
		SelectSet: []types.BaseSelectionSpec{
			TsRootFolder,
		},
	}
}

//
// Build the property Spec.
func (r *Reconciler) propertySpec() []types.PropertySpec {
	return []types.PropertySpec{
		{
			Type: VirtualMachine,
			PathSet: []string{
				"name",
				"summary",
			},
		},
	}
}

//
// Apply updates.
func (r Reconciler) apply(ctx context.Context, updates []types.ObjectUpdate) {
	for _, u := range updates {
		update := Update{
			kind:      string(u.Kind),
			changeSet: u.ChangeSet,
			db:        r.db,
		}
		switch u.Obj.Type {
		case VirtualMachine:
			update.model = &model.VM{
				Base: model.Base{
					ID: u.Obj.Value,
				},
			}
			update.Apply(ctx)
		default:
		}
	}
}

//
// Provide model update.
type Update struct {
	// Database client.
	db libmodel.DB
	// The `kind` of update.
	kind string
	// The changes to be applied.
	changeSet []types.PropertyChange
	// The model to be updated.
	model model.Model
}

//
// Apply the update.
func (u *Update) Apply(ctx context.Context) {
	switch u.kind {
	case Enter:
		u.insert()
	case Modify:
		u.update(ctx)
	case Leave:
		u.delete()
	}
}

//
// Insert the model.
func (u *Update) insert() {
	db := u.db
	u.model.With(u.changeSet)
	err := db.Insert(u.model)
	if err != nil {
		Log.Trace(err)
	}
}

//
// Update the model.
// Handles `Conflict` errors.
func (u *Update) update(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	db := u.db
	for {
		err := db.Get(u.model)
		if err != nil {
			Log.Trace(err)
			break
		}
		u.model.With(u.changeSet)
		err = db.Update(u.model)
		if err == nil {
			break
		}
		if errors.Is(err, libmodel.Conflict) {
			Log.Info(err.Error())
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				break
			}
			continue
		} else {
			Log.Trace(err)
			break
		}
	}
}

//
// Delete the model.
func (u *Update) delete() {
	db := u.db
	err := db.Delete(u.model)
	if err != nil {
		Log.Trace(err)
	}
}
