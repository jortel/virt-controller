package ovirt

import (
	liberr "github.com/konveyor/controller/pkg/error"
	fb "github.com/konveyor/controller/pkg/filebacked"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	model "github.com/konveyor/forklift-controller/pkg/controller/provider/model/ovirt"
)

//
// All adapters.
var adapterList []Adapter

//
// Event (type) mapped to adapter.
var adapterMap = map[int]Adapter{}

func init() {
	adapterList = []Adapter{
		&DataCenterAdapter{},
		&StorageDomainAdapter{},
		&VNICProfileAdapter{},
		&NetworkAdapter{},
		&ClusterAdapter{},
		&HostAdapter{},
		&VMAdapter{},
	}
	for _, adapter := range adapterList {
		for _, event := range adapter.Event() {
			adapterMap[event] = adapter
		}
	}
}

type Adapter interface {
	List(client *Client) (itr fb.Iterator, err error)
	Apply(client *Client, tx *libmodel.Tx, event *Event) (err error)
	Event() []int
}

//
// DataCenter.
type DataCenterAdapter struct {
}

func (r *DataCenterAdapter) List(client *Client) (itr fb.Iterator, err error) {
	dataCenterList := DataCenterList{}
	err = client.list("datacenters", &dataCenterList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range dataCenterList.Items {
		m := &model.DataCenter{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

//
// Handled events.
func (r *DataCenterAdapter) Event() []int {
	return []int{
		DataCenterAdded,
		DataCenterUpdated,
		DataCenterDeleted,
	}
}

func (r *DataCenterAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	case DataCenterAdded:
		object := &DataCenter{}
		err = client.get(event.DataCenter.Ref, object)
		if err != nil {
			break
		}
		m := &model.DataCenter{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Insert(m)
		if err != nil {
			break
		}
	case DataCenterUpdated:
		object := &DataCenter{}
		err = client.get(event.DataCenter.Ref, object)
		if err != nil {
			break
		}
		m := &model.DataCenter{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Update(m)
		if err != nil {
			break
		}
	case DataCenterDeleted:
		object := &DataCenter{}
		err = client.get(event.DataCenter.Ref, object)
		if err != nil {
			break
		}
		m := &model.DataCenter{
			Base: model.Base{ID: object.ID},
		}
		err = tx.Delete(m)
		if err != nil {
			break
		}
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// Network adapter.
type NetworkAdapter struct {
}

//
// Handled events.
func (r *NetworkAdapter) Event() []int {
	return []int{}
}

func (r *NetworkAdapter) List(client *Client) (itr fb.Iterator, err error) {
	networkList := NetworkList{}
	err = client.list("networks", &networkList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range networkList.Items {
		m := &model.Network{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *NetworkAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// VNICProfileAdapter adapter.
type VNICProfileAdapter struct {
}

//
// Handled events.
func (r *VNICProfileAdapter) Event() []int {
	return []int{}
}

func (r *VNICProfileAdapter) List(client *Client) (itr fb.Iterator, err error) {
	pList := VNICProfileList{}
	err = client.list("vnicprofiles", &pList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range pList.Items {
		m := &model.VNICProfile{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *VNICProfileAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// StorageDomain adapter.
type StorageDomainAdapter struct {
}

//
// Handled events.
func (r *StorageDomainAdapter) Event() []int {
	return []int{}
}

func (r *StorageDomainAdapter) List(client *Client) (itr fb.Iterator, err error) {
	sdList := StorageDomainList{}
	err = client.list("storagedomains", &sdList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range sdList.Items {
		m := &model.StorageDomain{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *StorageDomainAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// Cluster adapter.
type ClusterAdapter struct {
}

//
// Handled events.
func (r *ClusterAdapter) Event() []int {
	return []int{
		ClusterAdded,
		ClusterUpdated,
		ClusterDeleted,
	}
}

func (r *ClusterAdapter) List(client *Client) (itr fb.Iterator, err error) {
	clusterList := ClusterList{}
	err = client.list("clusters", &clusterList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range clusterList.Items {
		m := &model.Cluster{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *ClusterAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	case ClusterAdded:
		object := &Cluster{}
		err = client.get(event.Cluster.Ref, object)
		if err != nil {
			break
		}
		m := &model.Cluster{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Insert(m)
		if err != nil {
			break
		}
	case ClusterUpdated:
		object := &Cluster{}
		err = client.get(event.Cluster.Ref, object)
		if err != nil {
			break
		}
		m := &model.Cluster{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Update(m)
		if err != nil {
			break
		}
	case ClusterDeleted:
		m := &model.Cluster{
			Base: model.Base{ID: event.Cluster.Ref},
		}
		err = tx.Delete(m)
		if err != nil {
			break
		}
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// Host adapter.
type HostAdapter struct {
}

//
// Handled events.
func (r *HostAdapter) Event() []int {
	return []int{
		HostAdded,
		HostUpdated,
		HostDeleted,
	}
}

func (r *HostAdapter) List(client *Client) (itr fb.Iterator, err error) {
	hostList := HostList{}
	err = client.list("hosts", &hostList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range hostList.Items {
		m := &model.Host{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *HostAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	case HostAdded:
		object := &Host{}
		err = client.get(event.Host.Ref, object)
		if err != nil {
			break
		}
		m := &model.Host{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Insert(m)
		if err != nil {
			break
		}
	case HostUpdated:
		object := &Host{}
		err = client.get(event.Host.Ref, object)
		if err != nil {
			break
		}
		m := &model.Host{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Update(m)
		if err != nil {
			break
		}
	case HostDeleted:
		m := &model.Host{
			Base: model.Base{ID: event.Host.Ref},
		}
		err = tx.Delete(m)
		if err != nil {
			break
		}
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}

//
// VM adapter.
type VMAdapter struct {
}

//
// Handled events.
func (r *VMAdapter) Event() []int {
	return []int{
		VmAdded,
		VmUpdated,
		VmDeleted,
	}
}

func (r *VMAdapter) List(client *Client) (itr fb.Iterator, err error) {
	vmList := VMList{}
	err = client.list("vms", &vmList)
	if err != nil {
		return
	}
	list := fb.NewList()
	for _, object := range vmList.Items {
		m := &model.VM{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = list.Append(m)
		if err != nil {
			return
		}
	}

	itr = list.Iter()

	return
}

func (r *VMAdapter) Apply(client *Client, tx *libmodel.Tx, event *Event) (err error) {
	switch event.code() {
	case VmAdded:
		object := &VM{}
		err = client.get(event.VM.Ref, object)
		if err != nil {
			break
		}
		m := &model.VM{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Insert(m)
		if err != nil {
			break
		}
	case VmUpdated:
		object := &VM{}
		err = client.get(event.VM.Ref, object)
		if err != nil {
			break
		}
		m := &model.VM{
			Base: model.Base{ID: object.ID},
		}
		object.ApplyTo(m)
		err = tx.Update(m)
		if err != nil {
			break
		}
	case VmDeleted:
		m := &model.VM{
			Base: model.Base{ID: event.VM.Ref},
		}
		err = tx.Delete(m)
		if err != nil {
			break
		}
	default:
		err = liberr.New("unknown event", "event", event)
	}

	return
}
