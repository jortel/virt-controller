/*
Copyright 2019 Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	libcontainer "github.com/konveyor/controller/pkg/inventory/container"
	libmodel "github.com/konveyor/controller/pkg/inventory/model"
	libweb "github.com/konveyor/controller/pkg/inventory/web"
	"github.com/konveyor/controller/pkg/logging"
	api "github.com/konveyor/virt-controller/pkg/apis/virt/v1alpha1"
	rlnr "github.com/konveyor/virt-controller/pkg/controller/provider/container"
	"github.com/konveyor/virt-controller/pkg/controller/provider/model"
	"github.com/konveyor/virt-controller/pkg/controller/provider/web"
	"github.com/konveyor/virt-controller/pkg/settings"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	clienterror "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logging.WithName("provider")

//
// Application settings.
var Settings = &settings.Settings

func init() {
	model.SetLogger(&log)
	web.SetLogger(&log)
	rlnr.SetLogger(&log)
}

//
// Creates a new Inventory Controller and adds it to the Manager.
func Add(mgr manager.Manager) error {
	restCfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	nClient, err := client.New(
		restCfg,
		client.Options{
			Scheme: scheme.Scheme,
		})
	if err != nil {
		panic(err)
	}
	container := libcontainer.New()
	web := libweb.New(container, web.All(container)...)
	reconciler := &Reconciler{
		Client:    nClient,
		scheme:    mgr.GetScheme(),
		container: container,
		web:       web,
	}

	web.Start()

	cnt, err := controller.New(
		"provider-controller",
		mgr,
		controller.Options{
			Reconciler: reconciler,
		})
	if err != nil {
		log.Trace(err)
		return err
	}
	err = cnt.Watch(
		&source.Kind{Type: &api.Provider{}},
		&handler.EnqueueRequestForObject{},
		&ProviderPredicate{})
	if err != nil {
		log.Trace(err)
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &Reconciler{}

//
// Reconciles an provider object.
type Reconciler struct {
	client.Client
	scheme    *runtime.Scheme
	container *libcontainer.Container
	web       *libweb.WebServer
}

//
// Reconcile a Inventory CR.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error

	// Reset the logger.
	log.Reset()

	// Fetch the CR.
	provider := &api.Provider{}
	err = r.Get(context.TODO(), request.NamespacedName, provider)
	if err != nil {
		if clienterror.IsNotFound(err) {
			deleted := &api.Provider{
				ObjectMeta: meta.ObjectMeta{
					Namespace: request.Namespace,
					Name:      request.Name,
				},
			}
			if r, found := r.container.Get(deleted); found {
				r.Shutdown(true)
			}
			return reconcile.Result{}, nil
		}
		log.Trace(err)
		return reconcile.Result{}, err
	}

	// Validations.
	err = r.validate(provider)
	if err != nil {
		log.Trace(err)
		return reconcile.Result{Requeue: true}, nil
	}

	// Ready condition.
	if !provider.Status.HasBlockerCondition() {
		provider.Status.SetReady(true, ReadyMessage)
	}

	// Update the container.
	err = r.updateContainer(provider)
	if err != nil {
		log.Trace(err)
		return reconcile.Result{Requeue: true}, nil
	}

	// Apply changes.
	provider.Status.ObservedGeneration = provider.Generation
	err = r.Status().Update(context.TODO(), provider)
	if err != nil {
		log.Trace(err)
		return reconcile.Result{Requeue: true}, nil
	}

	// Done
	return reconcile.Result{}, nil
}

//
// Update the container.
func (r *Reconciler) updateContainer(provider *api.Provider) error {
	db := r.getDB(provider)
	secret, err := r.getSecret(provider)
	if err != nil {
		log.Trace(err)
		return err
	}
	ds := rlnr.New(provider, secret, db)
	if ds == nil {
		return errors.New("provider not supported")
	}

	r.container.Add(provider, ds)

	return nil
}

//
// Build DB for provider.
func (r *Reconciler) getDB(provider *api.Provider) libmodel.DB {
	dir := "/tmp"
	file := string(provider.UID)
	file = file + ".db"
	path := filepath.Join(dir, file)
	db := libmodel.New(path, model.All()...)
	return db
}

//
// Get the secret referenced by the provider.
func (r *Reconciler) getSecret(provider *api.Provider) (*core.Secret, error) {
	ref := provider.Spec.Secret
	secret := &core.Secret{}
	key := client.ObjectKey{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}
	err := r.Get(context.TODO(), key, secret)
	if err != nil {
		log.Trace(err)
		return nil, err
	}

	return secret, nil
}
