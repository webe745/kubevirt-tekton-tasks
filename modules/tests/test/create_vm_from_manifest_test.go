package test

import (
	"context"
	"fmt"

	"github.com/kubevirt/kubevirt-tekton-tasks/modules/sharedtest/testobjects"
	testtemplate "github.com/kubevirt/kubevirt-tekton-tasks/modules/sharedtest/testobjects/template"
	. "github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/constants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/framework"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/runner"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/testconfigs"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/tests/test/vm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	instancetypev1alpha2 "kubevirt.io/api/instancetype/v1alpha2"
)

var _ = Describe("Create VM from manifest", func() {
	f := framework.NewFramework().
		OnBeforeTestSetup(func(config framework.TestConfig) {
			if createVMConfig, ok := config.(*testconfigs.CreateVMTestConfig); ok {
				createVMConfig.TaskData.CreateMode = CreateVMVMManifestMode
			}
		})

	BeforeEach(func() {
		if f.TestOptions.SkipCreateVMFromManifestTests {
			Skip("skipCreateVMFromManifestTests is set to true, skipping tests")
		}
	})

	DescribeTable("taskrun fails and no VM is created", func(config *testconfigs.CreateVMTestConfig) {
		f.TestSetup(config)

		expectedVM := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVM) // in case it succeeds

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectFailure().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(nil)

		_, err := vm.WaitForVM(f.KubevirtClient, expectedVM.Namespace, expectedVM.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).Should(HaveOccurred())
	},
		Entry("no vm manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ExpectedLogs: "only one of vm-manifest, template-name or virtctl should be specified",
			},
			TaskData: testconfigs.CreateVMTaskData{},
		}),
		Entry("invalid manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "could not read VM manifest: error unmarshaling",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VMManifest: "invalid manifest",
			},
		}),
		Entry("create vm with non matching disk fails", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "admission webhook \"virtualmachine-validator.kubevirt.io\" denied the request: spec.template.spec.domain.devices.disks[0].Name",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("vm-with-non-existent-pvc").WithNonMatchingDisk().Build(),
			},
		}),
		Entry("cannot create a VM in different namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountNameNamespaced,
				ExpectedLogs:   "cannot create resource \"virtualmachines\" in API group \"kubevirt.io\"",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                testobjects.NewTestAlpineVM("different-ns-namespace-scope").Build(),
				VMTargetNamespace: SystemTargetNS,
			},
		}),
		Entry("cannot create a VM in different namespace in manifest", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountNameNamespaced,
				ExpectedLogs:   "cannot create resource \"virtualmachines\" in API group \"kubevirt.io\"",
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                        testobjects.NewTestAlpineVM("different-ns-namespace-scope-in-manifest").Build(),
				VMManifestTargetNamespace: SystemTargetNS,
			},
		}),
		Entry("manifest and virtctl are specified", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "only one of vm-manifest, template-name or virtctl should be specified",
			},
			TaskData: testconfigs.CreateVMTaskData{
				Virtctl:                            "--volume-containerdisk src:my.registry/my-image:my-tag",
				VM:                                 testobjects.NewTestAlpineVM("vm-with-manifest-namespace").Build(),
				VMManifestTargetNamespace:          DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
			},
		}),
		Entry("manifest, template and virtctl are specified", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "only one of vm-manifest, template-name or virtctl should be specified",
			},
			TaskData: testconfigs.CreateVMTaskData{
				Virtctl:                            "--volume-containerdisk src:my.registry/my-image:my-tag",
				VM:                                 testobjects.NewTestAlpineVM("vm-with-manifest-namespace").Build(),
				VMManifestTargetNamespace:          DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
				Template:                           testtemplate.NewCirrosServerTinyTemplate().Build(),
				TemplateParams: []string{
					testtemplate.TemplateParam(testtemplate.NameParam, E2ETestsRandomName("simple-vm")),
				},
			},
		}),
		Entry("should fail with invalid params", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   "unknown flag: --invalid",
			},
			TaskData: testconfigs.CreateVMTaskData{
				Virtctl: "--invalid params",
			},
		}),
	)

	DescribeTable("VM is created successfully", func(config *testconfigs.CreateVMTestConfig) {
		f.TestSetup(config)

		expectedVM := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVM)

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectSuccess().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(map[string]string{
				CreateVMResults.Name:      expectedVM.Name,
				CreateVMResults.Namespace: expectedVM.Namespace,
			})

		_, err := vm.WaitForVM(f.KubevirtClient, expectedVM.Namespace, expectedVM.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).ShouldNot(HaveOccurred())
	},
		Entry("simple vm", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("simple-vm").Build(),
			},
		}),
		Entry("vm to deploy namespace by default", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                                 testobjects.NewTestAlpineVM("vm-to-deploy-by-default").Build(),
				VMTargetNamespace:                  DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
			},
		}),
		Entry("vm with manifest namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                                 testobjects.NewTestAlpineVM("vm-with-manifest-namespace").Build(),
				VMManifestTargetNamespace:          DeployTargetNS,
				UseDefaultVMNamespacesInTaskParams: true,
			},
		}),

		Entry("vm with overridden manifest namespace", &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM:                        testobjects.NewTestAlpineVM("vm-with-overridden-manifest-namespace").Build(),
				VMManifestTargetNamespace: DeployTargetNS,
			},
		}),
	)

	It("VM is created from manifest properly ", func() {
		config := &testconfigs.CreateVMTestConfig{
			TaskRunTestConfig: testconfigs.TaskRunTestConfig{
				ServiceAccount: CreateVMFromManifestServiceAccountName,
				ExpectedLogs:   ExpectedSuccessfulVMCreation,
			},
			TaskData: testconfigs.CreateVMTaskData{
				VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
					WithLabel("app", "my-custom-app").
					WithVMILabel("name", "test").
					WithVMILabel("ra", "rara").
					Build(),
			},
		}
		f.TestSetup(config)

		expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
		f.ManageVMs(expectedVMStub)

		runner.NewTaskRunRunner(f, config.GetTaskRun()).
			CreateTaskRun().
			ExpectSuccess().
			ExpectLogs(config.GetAllExpectedLogs()...).
			ExpectResults(map[string]string{
				CreateVMResults.Name:      expectedVMStub.Name,
				CreateVMResults.Namespace: expectedVMStub.Namespace,
			})

		vm, err := vm.WaitForVM(f.KubevirtClient, expectedVMStub.Namespace, expectedVMStub.Name,
			"", config.GetTaskRunTimeout(), false)
		Expect(err).ShouldNot(HaveOccurred())

		vmName := expectedVMStub.Name
		expectedVM := config.TaskData.VM
		// fill VM accordingly
		expectedVM.Spec.Template.Spec.Domain.Machine = vm.Spec.Template.Spec.Domain.Machine // ignore Machine
		expectedVM.Spec.Template.Spec.Architecture = vm.Spec.Template.Spec.Architecture     // ignore Architecture
		expectedVM.Spec.Template.ObjectMeta.Labels["vm.kubevirt.io/name"] = vm.Spec.Template.ObjectMeta.Name

		Expect(vm.Spec.Template.Spec).Should(Equal(expectedVM.Spec.Template.Spec))
		// check VM labels
		Expect(vm.Labels).Should(Equal(expectedVM.Labels))
		// check VMI labels
		Expect(vm.Spec.Template.ObjectMeta.Labels).Should(Equal(map[string]string{
			"name":                "test",
			"ra":                  "rara",
			"vm.kubevirt.io/name": vmName,
		}))
	})

	Context("virtctl create vm", func() {
		It("should succeed with specified namespace and name", func() {
			vmName := "my-vm-0"
			config := &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					Virtctl:     fmt.Sprintf("--name %s --memory 256Mi", vmName),
					VMNamespace: f.TestOptions.DeployNamespace,
				},
			}
			f.TestSetup(config)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...)

			vm, err := vm.WaitForVM(f.KubevirtClient, f.TestOptions.DeployNamespace, vmName,
				"", config.GetTaskRunTimeout(), false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(vm.Name).To(Equal(vmName))
			Expect(*vm.Spec.RunStrategy).To(Equal(kubevirtv1.RunStrategyAlways))
		})

		It("should succeed without specified namespace", func() {
			config := &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					Virtctl: "--run-strategy Halted --memory 256Mi",
				},
			}
			f.TestSetup(config)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			taskrun := runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...)

			vmName := taskrun.GetResults()["name"]

			vm, err := vm.WaitForVM(f.KubevirtClient, f.TestOptions.DeployNamespace, vmName,
				"", config.GetTaskRunTimeout(), false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(vm.Name).To(Equal(vmName))
			Expect(*vm.Spec.RunStrategy).To(Equal(kubevirtv1.RunStrategyHalted))
		})

		It("should succeed with instancetype specified", func() {
			instancetypeName := "instancetype-2"
			config := &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					Virtctl: fmt.Sprintf("--instancetype %s", instancetypeName),
				},
			}
			f.TestSetup(config)

			instancetype := createInstancetype(f, instancetypeName)
			f.ManageClusterInstancetypes(instancetype)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			taskrun := runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...)

			vmName := taskrun.GetResults()["name"]

			vm, err := vm.WaitForVM(f.KubevirtClient, f.TestOptions.DeployNamespace, vmName,
				"", config.GetTaskRunTimeout(), false)

			Expect(err).ShouldNot(HaveOccurred())

			Expect(vm.Name).To(Equal(vmName))
			Expect(*vm.Spec.RunStrategy).To(Equal(kubevirtv1.RunStrategyAlways))
		})

		It("should start with startVM set to true", func() {
			instancetypeName := "instancetype-3"
			config := &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					Virtctl: fmt.Sprintf("--run-strategy Halted --instancetype %s", instancetypeName),
					StartVM: "true",
				},
			}
			f.TestSetup(config)

			instancetype := createInstancetype(f, instancetypeName)
			f.ManageClusterInstancetypes(instancetype)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			taskrun := runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...)

			vmName := taskrun.GetResults()["name"]

			vm, err := vm.WaitForVM(f.KubevirtClient, f.TestOptions.DeployNamespace, vmName,
				"", config.GetTaskRunTimeout(), false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(vm.Name).To(Equal(vmName))
			// startVM should set RunStrategy to Always
			Expect(*vm.Spec.RunStrategy).To(Equal(kubevirtv1.RunStrategyAlways))
		})
	})

	Context("with StartVM", func() {
		DescribeTable("VM is created successfully", func(config *testconfigs.CreateVMTestConfig, phase kubevirtv1.VirtualMachineInstancePhase, running bool) {
			f.TestSetup(config)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...).
				ExpectResults(map[string]string{
					CreateVMResults.Name:      expectedVMStub.Name,
					CreateVMResults.Namespace: expectedVMStub.Namespace,
				})

			vm, err := vm.WaitForVM(f.KubevirtClient, expectedVMStub.Namespace, expectedVMStub.Name,
				phase, config.GetTaskRunTimeout(), false)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(*vm.Spec.Running).To(Equal(running), "vm should be in correct running phase")
		},
			Entry("with false StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "false",
				},
			}, kubevirtv1.VirtualMachineInstancePhase(""), false),
			Entry("with invalid StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "invalid_value",
				},
			}, kubevirtv1.VirtualMachineInstancePhase(""), false),
			Entry("with true StartVM value", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					StartVM: "true",
				},
			}, kubevirtv1.Running, true),
		)
	})

	Context("with RunStrategy", func() {
		DescribeTable("VM is created successfully", func(config *testconfigs.CreateVMTestConfig, expectedRunStrategy kubevirtv1.VirtualMachineRunStrategy) {
			f.TestSetup(config)

			expectedVMStub := config.TaskData.GetExpectedVMStubMeta()
			f.ManageVMs(expectedVMStub)

			runner.NewTaskRunRunner(f, config.GetTaskRun()).
				CreateTaskRun().
				ExpectSuccess().
				ExpectLogs(config.GetAllExpectedLogs()...).
				ExpectResults(map[string]string{
					CreateVMResults.Name:      expectedVMStub.Name,
					CreateVMResults.Namespace: expectedVMStub.Namespace,
				})

			vm, err := f.KubevirtClient.VirtualMachine(expectedVMStub.Namespace).Get(expectedVMStub.Name, &v1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())

			Expect(*vm.Spec.RunStrategy).To(Equal(expectedRunStrategy), "vm should have correct run strategy")
		},
			Entry("with RunStrategy always", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					RunStrategy: "Always",
					StartVM:     "true",
				},
			}, kubevirtv1.RunStrategyAlways),
			Entry("with RunStrategy halted", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					RunStrategy: "Halted",
				},
			}, kubevirtv1.RunStrategyHalted),
			Entry("with RunStrategy halted and startVM", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					RunStrategy: "Halted",
					StartVM:     "true",
				},
			}, kubevirtv1.RunStrategyAlways),
			Entry("with RunStrategy Manual", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					RunStrategy: "Manual",
					StartVM:     "true",
				},
			}, kubevirtv1.RunStrategyManual),
			Entry("with RunStrategy RerunOnFailure", &testconfigs.CreateVMTestConfig{
				TaskRunTestConfig: testconfigs.TaskRunTestConfig{
					ServiceAccount: CreateVMFromManifestServiceAccountName,
					ExpectedLogs:   ExpectedSuccessfulVMCreation,
				},
				TaskData: testconfigs.CreateVMTaskData{
					VM: testobjects.NewTestAlpineVM("vm-from-manifest-data").
						WithLabel("app", "my-custom-app").
						WithVMILabel("name", "test").
						WithVMILabel("ra", "rara").
						Build(),
					RunStrategy: "RerunOnFailure",
					StartVM:     "true",
				},
			}, kubevirtv1.RunStrategyRerunOnFailure),
		)
	})
})

func createInstancetype(f *framework.Framework, instancetypeName string) *instancetypev1alpha2.VirtualMachineClusterInstancetype {
	instancetype := &instancetypev1alpha2.VirtualMachineClusterInstancetype{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instancetypeName,
			Namespace: f.TestOptions.DeployNamespace,
		},
		Spec: instancetypev1alpha2.VirtualMachineInstancetypeSpec{
			CPU: instancetypev1alpha2.CPUInstancetype{
				Guest: uint32(1),
			},
			Memory: instancetypev1alpha2.MemoryInstancetype{
				Guest: resource.MustParse("128Mi"),
			},
		},
	}
	createdInstancetype, err := f.Clients.KubevirtClient.VirtualMachineClusterInstancetype().Create(context.Background(), instancetype, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred())

	return createdInstancetype
}
