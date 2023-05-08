// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package modulelifecycle

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano-modules/common/actionspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/controllerspi"
	"github.com/verrazzano/verrazzano-modules/common/controllers/base/statemachine"
	moduleapi "github.com/verrazzano/verrazzano-modules/module-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	vzspi "github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
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

var _ actionspi.LifecycleActionHandler = &handler{}

func Test(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		name                       string
		startingStatusState        moduleapi.ModuleLifecycleState
		expectedStatemachineCalled bool
		statemachineError          bool
		expectedRequeue            bool
		expectedError              bool
	}{
		{
			name:                       "test1",
			startingStatusState:        "",
			statemachineError:          false,
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
			executeStateMachine = h.testExecuteStateMachine
			defer func() { executeStateMachine = defaultExecuteStateMachine }()

			rctx := controllerspi.ReconcileContext{
				Log:       vzlog.DefaultLogger(),
				ClientCtx: context.TODO(),
			}

			cr := newModuleLifecycleCR(namespace, name, "", moduleapi.InstallAction, test.startingStatusState)
			scheme := initScheme()
			clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
			r := Reconciler{
				Client: clientBuilder.Build(),
				Scheme: initScheme(),
				handlers: actionspi.ActionHandlers{
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

func newModuleLifecycleCR(namespace string, name string, className string, action moduleapi.ActionType, startingStatusState moduleapi.ModuleLifecycleState) *moduleapi.ModuleLifecycle {
	m := &moduleapi.ModuleLifecycle{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: moduleapi.ModuleLifecycleSpec{
			Action:             action,
			LifecycleClassName: moduleapi.LifecycleClassType(className),
		},
		Status: moduleapi.ModuleLifecycleStatus{
			State: startingStatusState,
		},
	}
	return m
}

func initScheme() *runtime.Scheme {
	// Create a scheme then add each GKV group to the scheme
	scheme := runtime.NewScheme()

	utilruntime.Must(moduleapi.AddToScheme(scheme))
	return scheme
}

func (h handler) testExecuteStateMachine(sm statemachine.StateMachine, ctx vzspi.ComponentContext) ctrl.Result {
	return ctrl.Result{}
}

func (h handler) GetActionName() string {
	return "install"
}

func (h handler) Init(context vzspi.ComponentContext, config actionspi.HandlerConfig) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsActionNeeded(context vzspi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PreAction(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PreActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPreActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) ActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) DoAction(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) PostActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) PostAction(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h handler) IsPostActionDone(context vzspi.ComponentContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h handler) CompletedActionUpdateStatus(context vzspi.ComponentContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
