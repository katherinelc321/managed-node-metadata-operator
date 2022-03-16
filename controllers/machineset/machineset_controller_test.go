package machineset

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/golang/mock/gomock"

	machinev1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type mocks struct {
	fakeKubeClient client.Client
	mockCtrl       *gomock.Controller
}

var _ = Describe("MachinesetController", func() {

	var (
		machineSet     machinev1.MachineSet
		machine        machinev1.Machine
		node           v1.Node
		updatedNode    v1.Node
		updatedMachine machinev1.Machine
		mockObjects    *mocks
		r              *ReconcileMachineSet
		ctx            context.Context
	)

	err := machinev1.AddToScheme(scheme.Scheme)
	if err != nil {
		fmt.Printf("failed adding apis to scheme in machineset controller tests")
	}

	Describe("Updating labels in machine", func() {
		var (
			newLabelsInMachineSet   map[string]string
			existingLabelsInMachine map[string]string
			updatedLabelsInMachine  map[string]string
		)

		BeforeEach(func() {
			machineSet = machinev1.MachineSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test machineset",
					Namespace: "test",
				},
				Spec: machinev1.MachineSetSpec{
					Template: machinev1.MachineTemplateSpec{
						Spec: machinev1.MachineSpec{
							ObjectMeta: machinev1.ObjectMeta{
								Labels: newLabelsInMachineSet,
							},
						},
					},
				},
			}
			machine = machinev1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test machineset",
					Namespace: "test",
				},
				Spec: machinev1.MachineSpec{
					ObjectMeta: machinev1.ObjectMeta{
						Labels: existingLabelsInMachine,
					},
				},
			}
			updatedMachine = machinev1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test machineset",
					Namespace: "test",
				},
				Spec: machinev1.MachineSpec{
					ObjectMeta: machinev1.ObjectMeta{
						Labels: updatedLabelsInMachine,
					},
				},
			}

			localObjects := []runtime.Object{
				&machineSet,
				&machine,
			}
			mockObjects = &mocks{
				fakeKubeClient: fake.NewFakeClient(localObjects...),
				mockCtrl:       gomock.NewController(GinkgoT()),
			}

			r = &ReconcileMachineSet{
				mockObjects.fakeKubeClient,
				scheme.Scheme,
				record.NewFakeRecorder(32),
			}
		})

		AfterEach(func() {
			mockObjects.mockCtrl.Finish()
		})

		Context("When new label is added to machineset", func() {
			newLabelsInMachineSet = map[string]string{"foo": "bar"}
			existingLabelsInMachine = map[string]string{}
			updatedLabelsInMachine = map[string]string{"foo": "bar"}

			It("should update labels in machine", func() {
				err = r.updateLabelsInMachine(ctx, &machineSet, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedMachine.Spec.Labels))
			})
		})

		Context("When label is deleted from machinset", func() {
			newLabelsInMachineSet = map[string]string{}
			existingLabelsInMachine = map[string]string{"foo": "bar"}
			updatedLabelsInMachine = map[string]string{}

			It("should delete label in machine", func() {
				err = r.updateLabelsInMachine(ctx, &machineSet, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedMachine.Spec.Labels))
			})
		})

		Context("When no new label is added to machineset", func() {
			newLabelsInMachineSet = map[string]string{}
			existingLabelsInMachine = map[string]string{}
			updatedLabelsInMachine = map[string]string{}

			It("should not change labels", func() {
				err = r.updateLabelsInMachine(ctx, &machineSet, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedMachine.Spec.Labels))
			})
		})

	})

	Describe("Updating labels in node", func() {
		var (
			newLabelsInMachine   map[string]string
			existingLabelsInNode map[string]string
			updatedLabelsInNode  map[string]string
		)

		BeforeEach(func() {
			machine = machinev1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test machineset",
					Namespace: "test",
				},
				Spec: machinev1.MachineSpec{
					ObjectMeta: machinev1.ObjectMeta{
						Labels: newLabelsInMachine,
					},
				},
			}
			node = v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Namespace:   "test",
					Labels:      existingLabelsInNode,
					Annotations: existingLabelsInNode,
				},
			}

			updatedNode = v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Namespace:   "test",
					Labels:      updatedLabelsInNode,
					Annotations: updatedLabelsInNode,
				},
			}

			localObjects := []runtime.Object{
				&machine,
				&node,
			}
			mockObjects = &mocks{
				fakeKubeClient: fake.NewFakeClient(localObjects...),
				mockCtrl:       gomock.NewController(GinkgoT()),
			}

			r = &ReconcileMachineSet{
				mockObjects.fakeKubeClient,
				scheme.Scheme,
				record.NewFakeRecorder(32),
			}
		})

		AfterEach(func() {
			mockObjects.mockCtrl.Finish()
		})

		Context("When new label is added to machine", func() {
			newLabelsInMachine = map[string]string{"foo": "bar"}
			existingLabelsInNode = map[string]string{}
			updatedLabelsInNode = map[string]string{"foo": "bar"}

			It("should update labels in node", func() {
				err = r.updateLabelsInNode(ctx, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedNode.Labels))
			})
		})

		Context("When label is deleted from machine", func() {
			newLabelsInMachine = map[string]string{}
			existingLabelsInNode = map[string]string{"foo": "bar"}
			updatedLabelsInNode = map[string]string{}

			It("should update labels in node", func() {
				err = r.updateLabelsInNode(ctx, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedNode.Labels))
			})
		})

		Context("When no new label is added to machine", func() {
			newLabelsInMachine = map[string]string{}
			existingLabelsInNode = map[string]string{}
			updatedLabelsInNode = map[string]string{}

			It("should not change labels in node", func() {
				err = r.updateLabelsInNode(ctx, &machine)
				Expect(err).NotTo(HaveOccurred())
				Expect(machine.Spec.Labels).To(Equal(updatedNode.Labels))
			})
		})

	})

})
