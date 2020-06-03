package vmware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	model "github.com/konveyor/virt-controller/pkg/controller/provider/model/vmware"
	"net/http"
)

const (
	VMsRoot = Root + "/vms"
	VMRoot  = VMsRoot + "/:vm"
)

//
// Virtual Machine handler.
type VMHandler struct {
	Base
}

//
// Add routes to the `gin` router.
func (h *VMHandler) AddRoutes(e *gin.Engine) {
	e.GET(VMsRoot, h.List)
	e.GET(VMsRoot+"/", h.List)
	e.GET(VMRoot, h.Get)
}

//
// Prepare to handle the request.
func (h *VMHandler) Prepare(ctx *gin.Context) int {
	return h.Base.Prepare(ctx)
}

//
// List resources in a REST collection.
func (h VMHandler) List(ctx *gin.Context) {
	status := h.Prepare(ctx)
	if status != http.StatusOK {
		ctx.Status(status)
		return
	}
	db := h.Reconciler.DB()
	selector := &model.VM{}
	options := libmodel.ListOptions{
		Page: &h.Page,
	}
	list := []model.VM{}
	err := db.List(selector, options, &list)
	if err != nil {
		Log.Trace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	content := []*VM{}
	for _, m := range list {
		r := &VM{}
		r.With(&m, false)
		content = append(content, r)
	}

	ctx.JSON(http.StatusOK, content)
}

//
// Get a specific REST resource.
func (h VMHandler) Get(ctx *gin.Context) {
	status := h.Prepare(ctx)
	if status != http.StatusOK {
		ctx.Status(status)
		return
	}
	m := &model.VM{
		Base: model.Base{
			ID: ctx.Param("vm"),
		},
	}
	db := h.Reconciler.DB()
	err := db.Get(m)
	if err != nil {
		Log.Trace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	r := &VM{}
	r.With(m, true)

	ctx.JSON(http.StatusOK, r)
}

//
// REST Resource.
type VM struct {
	ID     string      `json:"id"`
	Name   interface{} `json:"name"`
	Object interface{} `json:"object,omitempty"`
}

//
// Build the resource using the model.
func (r *VM) With(m *model.VM, detail bool) {
	r.ID = m.ID
	object := map[string]interface{}{}
	json.Unmarshal([]byte(m.Object), &object)
	r.Name = object["name"]
	if detail {
		r.Object = object
	}
}
