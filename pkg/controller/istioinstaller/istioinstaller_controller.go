/*
Copyright 2019 The Istio Authors.

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

package istioinstaller

import (
	"context"
	"sync"

	installerv1alpha1 "github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/ostromart/istio-installer/pkg/component/installation"
	"fmt"
)

type IstioInstallationInitializerType func()*installation.IstioInstallation
var (
	istioInstallations = make(map[string]*installation.IstioInstallation)
	istioInstallationsMu sync.Mutex
	istioInstallationInitializer IstioInstallationInitializerType
)

var log = logf.Log.WithName("controller")

func SetIstioInstallationInitializer(f IstioInstallationInitializerType) {
	istioInstallationInitializer = f
}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IstioInstaller Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIstioInstaller{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("istioinstaller-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to IstioInstaller
	err = c.Watch(&source.Kind{Type: &installerv1alpha1.IstioInstaller{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		fmt.Println("Error in watch")
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileIstioInstaller{}

// ReconcileIstioInstaller reconciles a IstioInstaller object
type ReconcileIstioInstaller struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a IstioInstaller object and makes changes based on the state read
// and what is in the IstioInstaller.Spec
func (r *ReconcileIstioInstaller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the IstioInstaller istioInstaller
	istioInstaller := &installerv1alpha1.IstioInstaller{}
	err := r.Get(context.TODO(), request.NamespacedName, istioInstaller)
	nsName := namespaceNameString(request)
	istioInstallationsMu.Lock()
	defer istioInstallationsMu.Unlock()
	if err != nil {
		if errors.IsNotFound(err) {
			delete(istioInstallations, nsName)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if istioInstallationInitializer == nil {
		return reconcile.Result{}, fmt.Errorf("istioInstallationInitializer not set")
	}

	// For the time being, throw away all the old controllers and create a new batch.
	// TODO: check performance and optimize if needed.
	i := istioInstallationInitializer()
	i.Build(&istioInstaller.Spec)

	istioInstallations[nsName] = i
	i.RunApplyLoop(context.Background())

	return reconcile.Result{}, nil
}

func namespaceNameString(request reconcile.Request) string {
	return fmt.Sprintf("%s-%s", request.Namespace, request.Name)
}