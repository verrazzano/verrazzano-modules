// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package module

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/handlerspi"
	"github.com/verrazzano/verrazzano-modules/module-operator/internal/statemachine"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const namespace = "testns"
const name = "test"

type handler struct {
	statemachineError  bool
	statemachineCalled bool
	loadHelmInfoErr    string
	smHandler          handlerspi.StateMachineHandler
}

var _ handlerspi.StateMachineHandler = &handler{}

// TestReconcileSuccess tests that the Reconcile implementation works correctly
// GIVEN a Reconciler
// WHEN the Reconcile method is called for the life cycle operations
// THEN ensure that it works correctly
func TestReconcileSuccess(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                       string
		statemachineError          bool
		specVersion                string
		statusVersion              string
		statusGeneration           int64
		moduleInfo                 handlerspi.ModuleHandlerInfo
		conditions                 []moduleapi.ModuleCondition
		expectedStatemachineCalled bool
		expectedRequeue            bool
		expectedError              bool
		expectNilHandler           bool
	}{
		{
			name:              "test-install",
			statemachineError: false,
			moduleInfo: handlerspi.ModuleHandlerInfo{
				InstallActionHandler: &handler{},
			},
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:              "test-update",
			statemachineError: false,
			moduleInfo: handlerspi.ModuleHandlerInfo{
				UpdateActionHandler: &handler{},
			},
			conditions: []moduleapi.ModuleCondition{{
				Type: moduleapi.CondInstallComplete,
			}},
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:              "test-upgrade",
			statemachineError: false,
			specVersion:       "1.0.1",
			statusVersion:     "1.0.0",
			moduleInfo: handlerspi.ModuleHandlerInfo{
				UpgradeActionHandler: &handler{},
			},
			conditions: []moduleapi.ModuleCondition{{
				Type: moduleapi.CondInstallComplete,
			}},
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &handler{
				statemachineError: test.statemachineError,
			}
			funcExecuteStateMachine = h.testExecuteStateMachine
			defer func() { funcExecuteStateMachine = defaultExecuteStateMachine }()

			funcLoadHelmInfo = h.fakeLoadHelmInfo
			defer func() { funcLoadHelmInfo = loadHelmInfo }()

			rctx := controllerspi.ReconcileContext{
				Log:       vzlog.DefaultLogger(),
				ClientCtx: context.TODO(),
			}
			cr := &moduleapi.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Spec: moduleapi.ModuleSpec{
					Version: test.specVersion,
				},
				Status: moduleapi.ModuleStatus{
					Conditions:         test.conditions,
					Version:            test.statusVersion,
					ObservedGeneration: test.statusGeneration,
				},
			}

			scheme := initScheme()
			clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
			r := Reconciler{
				Client:      clientBuilder.Build(),
				Scheme:      initScheme(),
				HandlerInfo: test.moduleInfo,
			}
			uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
			asserts.NoError(err)
			res, err := r.Reconcile(rctx, &unstructured.Unstructured{Object: uObj})

			asserts.Equal(test.expectedStatemachineCalled, h.statemachineCalled)
			asserts.Equal(test.expectedError, err != nil)
			asserts.Equal(test.expectedRequeue, res.Requeue)
			asserts.NotNil(h.smHandler)
		})
	}
}

// TestReconcileErrors tests that the Reconcile implementation handles errors correctly
// GIVEN a Reconciler
// WHEN the Reconcile method is called for the life cycle operations
// THEN ensure that various errors are handled correctly
func TestReconcileErrors(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                       string
		statemachineError          bool
		specVersion                string
		statusVersion              string
		statusGeneration           int64
		loadHelmInfoErr            string
		moduleInfo                 handlerspi.ModuleHandlerInfo
		conditions                 []moduleapi.ModuleCondition
		funcUpgradeNeeded          func(desiredVersion, installedVersion string) (bool, error)
		expectedStatemachineCalled bool
		expectedRequeue            bool
		expectedError              bool
		expectNilHandler           bool
	}{
		{
			name:              "test-state-machine-error",
			statemachineError: true,
			moduleInfo: handlerspi.ModuleHandlerInfo{
				InstallActionHandler: &handler{},
			},
			expectedStatemachineCalled: true,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "test-same-generation",
			statemachineError:          false,
			statusGeneration:           1,
			expectedStatemachineCalled: false,
			expectedRequeue:            false,
			expectedError:              false,
			expectNilHandler:           true,
		},
		{
			name:                       "test-uprade-needed-failed",
			statemachineError:          false,
			funcUpgradeNeeded:          fakeIsUpgradeNeeded,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
			expectNilHandler:           true,
			conditions: []moduleapi.ModuleCondition{{
				Type: moduleapi.CondInstallComplete,
			}},
		},
		{
			name:                       "test-nil-handler",
			statemachineError:          false,
			funcUpgradeNeeded:          fakeIsUpgradeNeeded,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
			expectNilHandler:           true,
		},
		{
			name:              "test-helm-not-found",
			statemachineError: false,
			loadHelmInfoErr:   "FileNotFound",
			moduleInfo: handlerspi.ModuleHandlerInfo{
				InstallActionHandler: &handler{},
			},
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              true,
			expectNilHandler:           true,
		},
		{
			name:              "test-helm-error",
			statemachineError: false,
			loadHelmInfoErr:   "ReadError",
			moduleInfo: handlerspi.ModuleHandlerInfo{
				InstallActionHandler: &handler{},
			},
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              true,
			expectNilHandler:           true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &handler{
				statemachineError: test.statemachineError,
				loadHelmInfoErr:   test.loadHelmInfoErr,
			}
			funcExecuteStateMachine = h.testExecuteStateMachine
			defer func() { funcExecuteStateMachine = defaultExecuteStateMachine }()

			funcLoadHelmInfo = h.fakeLoadHelmInfo
			defer func() { funcLoadHelmInfo = loadHelmInfo }()

			if test.funcUpgradeNeeded != nil {
				funcIsUpgradeNeeded = fakeIsUpgradeNeeded
			}
			defer func() { funcIsUpgradeNeeded = IsUpgradeNeeded }()

			rctx := controllerspi.ReconcileContext{
				Log:       vzlog.DefaultLogger(),
				ClientCtx: context.TODO(),
			}
			cr := &moduleapi.Module{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Spec: moduleapi.ModuleSpec{
					Version: test.specVersion,
				},
				Status: moduleapi.ModuleStatus{
					Conditions:         test.conditions,
					Version:            test.statusVersion,
					ObservedGeneration: test.statusGeneration,
				},
			}

			scheme := initScheme()
			clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
			r := Reconciler{
				Client:      clientBuilder.Build(),
				Scheme:      initScheme(),
				HandlerInfo: test.moduleInfo,
			}
			uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
			asserts.NoError(err)
			res, err := r.Reconcile(rctx, &unstructured.Unstructured{Object: uObj})

			asserts.Equal(test.expectedStatemachineCalled, h.statemachineCalled)
			asserts.Equal(test.expectedError, err != nil)
			asserts.Equal(test.expectedRequeue, res.Requeue)
			asserts.Equal(test.expectNilHandler, h.smHandler == nil)
		})
	}
}

func initScheme() *runtime.Scheme {
	// Create a scheme then add each GKV group to the scheme
	scheme := runtime.NewScheme()

	utilruntime.Must(moduleapi.AddToScheme(scheme))
	return scheme
}

func (h *handler) testExecuteStateMachine(ctx handlerspi.HandlerContext, sm statemachine.StateMachine) ctrl.Result {
	h.statemachineCalled = true
	h.smHandler = sm.Handler
	if h.statemachineError {
		return util.NewRequeueWithShortDelay()
	}
	return ctrl.Result{}
}

func (h handler) GetWorkName() string {
	return "install"
}

func (h handler) Init(context handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsWorkNeeded(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PreWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PreWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPreActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) DoWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) DoWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsWorkDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PostWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PostWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPostActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) WorkCompletedUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) fakeLoadHelmInfo(cr *moduleapi.Module) (handlerspi.HelmInfo, error) {
	if h.loadHelmInfoErr == "" {
		return handlerspi.HelmInfo{}, nil
	}
	return handlerspi.HelmInfo{}, errors.New(h.loadHelmInfoErr)
}

func fakeIsUpgradeNeeded(desiredVersion, installedVersion string) (bool, error) {
	return false, errors.New("fake-error")
}
