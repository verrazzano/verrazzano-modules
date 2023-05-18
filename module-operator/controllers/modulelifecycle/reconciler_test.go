// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllercore/statemachine"
	"github.com/verrazzano/verrazzano-modules/common/handlerspi"
	"github.com/verrazzano/verrazzano-modules/common/pkg/controller/util"
	"github.com/verrazzano/verrazzano-modules/common/pkg/vzlog"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
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
}

var _ handlerspi.StateMachineHandler = &handler{}

// TestReconcile tests that the Reconcile implementation works correctly
// GIVEN a Reconciler
// WHEN the Reconcile method is called for various scenarios
// THEN ensure that it works correctly
func TestReconcile(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                       string
		action                     moduleapi.ModuleLifecycleActionType
		expectedStatemachineCalled bool
		statemachineError          bool
		expectedRequeue            bool
		expectedError              bool
		startingStatusState        moduleapi.ModuleLifecycleState
		statusGeneration           int64
	}{
		{
			name:                       "test-install",
			action:                     moduleapi.InstallAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "install-statemachine-error",
			action:                     moduleapi.InstallAction,
			startingStatusState:        "",
			statemachineError:          true,
			expectedStatemachineCalled: true,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "install-state-completed",
			action:                     moduleapi.InstallAction,
			startingStatusState:        moduleapi.StateCompleted,
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "install-state-not-needed",
			action:                     moduleapi.InstallAction,
			startingStatusState:        moduleapi.StateNotNeeded,
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "test-action-upgrade",
			action:                     moduleapi.UpgradeAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "test-action-update",
			action:                     moduleapi.UpdateAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "test-action-uninstall",
			action:                     moduleapi.DeleteAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "test-missing-action",
			action:                     "",
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &handler{
				statemachineError: test.statemachineError,
			}
			executeStateMachine = h.testExecuteStateMachine
			defer func() { executeStateMachine = defaultExecuteStateMachine }()

			rctx := controllerspi.ReconcileContext{
				Log:       vzlog.DefaultLogger(),
				ClientCtx: context.TODO(),
			}
			cr := &moduleapi.ModuleLifecycle{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Spec: moduleapi.ModuleLifecycleSpec{
					Action:             test.action,
					LifecycleClassName: moduleapi.LifecycleClassType(moduleapi.CalicoLifecycleClass),
				},
				Status: moduleapi.ModuleLifecycleStatus{
					State:              test.startingStatusState,
					ObservedGeneration: test.statusGeneration,
				},
			}

			scheme := initScheme()
			clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
			r := Reconciler{
				Client: clientBuilder.Build(),
				Scheme: initScheme(),
				handlerInfo: handlerspi.ModuleLifecycleHandlerInfo{
					InstallActionHandler: &handler{},
				},
			}
			uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
			asserts.NoError(err)
			res, err := r.Reconcile(rctx, &unstructured.Unstructured{Object: uObj})

			asserts.Equal(test.expectedStatemachineCalled, h.statemachineCalled)
			asserts.Equal(test.expectedError, err != nil)
			asserts.Equal(test.expectedRequeue, res.Requeue)
		})
	}
}

func initScheme() *runtime.Scheme {
	// Create a scheme then add each GKV group to the scheme
	scheme := runtime.NewScheme()

	utilruntime.Must(moduleapi.AddToScheme(scheme))
	return scheme
}

func (h *handler) testExecuteStateMachine(sm statemachine.StateMachine, ctx handlerspi.HandlerContext) ctrl.Result {
	h.statemachineCalled = true
	if h.statemachineError {
		return util.NewRequeueWithShortDelay()
	}
	return ctrl.Result{}
}

func (h handler) GetActionName() string {
	return "install"
}

func (h handler) Init(context handlerspi.HandlerContext, config handlerspi.StateMachineHandlerConfig) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsActionNeeded(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PreAction(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PreActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPreActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) ActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) DoAction(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PostActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PostAction(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPostActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) CompletedActionUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
