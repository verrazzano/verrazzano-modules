// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package moduleaction

import (
	"context"
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
}

type moduleHandler struct {
	actualVersion string
	handlerspi.ModuleActualState
}

var _ handlerspi.StateMachineHandler = &handler{}
var _ handlerspi.ModuleActualStateInCluster = &moduleHandler{}

// TestReconcile tests that the Reconcile implementation works correctly
// GIVEN a Reconciler
// WHEN the Reconcile method is called for various scenarios
// THEN ensure that it works correctly
func TestReconcile(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                       string
		action                     moduleapi.ModuleActionType
		expectedStatemachineCalled bool
		statemachineError          bool
		expectedRequeue            bool
		expectedError              bool
		startingStatusState        moduleapi.ModuleActionState
		statusGeneration           int64
		desiredVersion             string
		actualVersion              string
		handlerspi.ModuleActualState
	}{
		{
			name:                       "test-install",
			ModuleActualState:          handlerspi.ModuleStateNotInstalled,
			action:                     moduleapi.ReconcileAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "install-statemachine-error",
			ModuleActualState:          handlerspi.ModuleStateNotInstalled,
			action:                     moduleapi.ReconcileAction,
			startingStatusState:        "",
			statemachineError:          true,
			expectedStatemachineCalled: true,
			expectedRequeue:            true,
			expectedError:              false,
		},
		{
			name:                       "install-state-completed",
			ModuleActualState:          handlerspi.ModuleStateNotInstalled,
			action:                     moduleapi.ReconcileAction,
			startingStatusState:        moduleapi.StateCompleted,
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "install-state-not-needed",
			ModuleActualState:          handlerspi.ModuleStateNotInstalled,
			action:                     moduleapi.ReconcileAction,
			startingStatusState:        moduleapi.StateNotNeeded,
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:                       "test-action-upgrade",
			ModuleActualState:          handlerspi.ModuleStateReady,
			action:                     moduleapi.ReconcileAction,
			startingStatusState:        "",
			statemachineError:          false,
			expectedStatemachineCalled: false,
			expectedRequeue:            true,
			expectedError:              false,
			actualVersion:              "v1",
			desiredVersion:             "v2",
		},
		{
			name:                       "test-action-update",
			ModuleActualState:          handlerspi.ModuleStateReady,
			action:                     moduleapi.ReconcileAction,
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
			cr := &moduleapi.ModuleAction{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Spec: moduleapi.ModuleActionSpec{
					Action:          test.action,
					Version:         test.desiredVersion,
					ModuleClassName: moduleapi.ModuleClassType(moduleapi.CalicoModuleClass),
				},
				Status: moduleapi.ModuleActionStatus{
					State:              test.startingStatusState,
					ObservedGeneration: test.statusGeneration,
				},
			}

			scheme := initScheme()
			clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
			r := Reconciler{
				Client: clientBuilder.Build(),
				Scheme: initScheme(),
				handlerInfo: handlerspi.ModuleActionHandlerInfo{
					ModuleActualStateInCluster: moduleHandler{
						ModuleActualState: test.ModuleActualState,
						actualVersion:     test.actualVersion,
					},
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

func (h *handler) testExecuteStateMachine(ctx handlerspi.HandlerContext, sm statemachine.StateMachine) ctrl.Result {
	h.statemachineCalled = true
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

func (m moduleHandler) GetActualModuleState(context handlerspi.HandlerContext, cr *moduleapi.ModuleAction) (handlerspi.ModuleActualState, ctrl.Result, error) {
	return m.ModuleActualState, ctrl.Result{}, nil
}

func (m moduleHandler) IsUpgradeNeeded(context handlerspi.HandlerContext, cr *moduleapi.ModuleAction) (bool, ctrl.Result, error) {
	return m.actualVersion != cr.Spec.Version, ctrl.Result{}, nil
}
