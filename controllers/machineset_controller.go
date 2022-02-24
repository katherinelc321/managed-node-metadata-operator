/*


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

package controllers

import (
	"context"
	"fmt"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	machinev1 "github.com/openshift/api/machine/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	controllerName = "machineset_controller"
)

// Add creates a new MachineSet Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler.
func newReconciler(mgr manager.Manager) *MachineSetReconciler {
	r := &MachineSetReconciler{
		Client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(controllerName)}
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller.
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to MachineSet.
	err = c.Watch(
		&source.Kind{Type: &machinev1.MachineSet{}},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return err
	}

	return nil
}

// MachineSetReconciler reconciles a MachineSet object
type MachineSetReconciler struct {
	client.Client

	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=machine.openshift.io,resources=machinesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=machine.openshift.io,resources=machinesets/status,verbs=get;update;patch

func (r *MachineSetReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	// Fetch the MachineSet instance
	machineSet := &machinev1.MachineSet{}
	err := r.Get(ctx, request.NamespacedName, machineSet)
	if err != nil {
		if k8serr.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if machineSet.Labels != nil {
		machinesetLabel := machineSet.Labels

		allMachines := &machinev1.MachineList{}

		err := r.Client.List(context.Background(), allMachines, client.InNamespace(machineSet.Namespace))
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to list machines: %w", err)
		}

		for idx := range allMachines.Items {
			machine := &allMachines.Items[idx]
			if machine.Labels == nil {
				machine.Labels = machinesetLabel
			}

			node := &v1.Node{}
			key := client.ObjectKey{Namespace: metav1.NamespaceNone, Name: machine.Status.NodeRef.Name}
			err := r.Client.Get(context.TODO(), key, node)
			if err != nil {
				fmt.Errorf("failed to fetch node for machine %s", machine.Name)
				return reconcile.Result{}, err
			}

			if node.Labels == nil {
				node.Labels = machinesetLabel
			}
		}
	}

	return reconcile.Result{}, nil
}
