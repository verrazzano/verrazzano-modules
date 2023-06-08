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
	"github.com/verrazzano/verrazzano-modules/pkg/controller/result"
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	fakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

type finalizerHandler struct {
	statemachineError  bool
	statemachineCalled bool
	loadHelmInfoErr    string
	smHandler          handlerspi.StateMachineHandler
}

var _ handlerspi.StateMachineHandler = &handler{}

// TestPreRemoveFinalizer tests that the PreRemoveFinalizer implementation works correctly
// GIVEN a Finalizer
// WHEN the PreRemoveFinalizer method is called
// THEN ensure that it works correctly
func TestPreRemoveFinalizer(t *testing.T) {
	asserts := assert.New(t)

	const namespace = "testns"
	const name = "test"

	tests := []struct {
		name                       string
		statemachineError          bool
		specVersion                string
		statusVersion              string
		statusGeneration           int64
		moduleInfo                 handlerspi.ModuleHandlerInfo
		expectedStatemachineCalled bool
		expectedRequeue            bool
		expectedError              bool
		expectNilHandler           bool
	}{
		{
			name:              "test-no-error",
			statemachineError: false,
			moduleInfo: handlerspi.ModuleHandlerInfo{
				DeleteActionHandler: &handler{},
			},
			expectedStatemachineCalled: true,
			expectedRequeue:            false,
			expectedError:              false,
		},
		{
			name:              "test-state-machine-error",
			statemachineError: true,
			moduleInfo: handlerspi.ModuleHandlerInfo{
				DeleteActionHandler: &handler{},
			},
			expectedStatemachineCalled: true,
			expectedRequeue:            true,
			expectedError:              false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &finalizerHandler{
				statemachineError: test.statemachineError,
			}
			funcExecuteStateMachine = h.fakeExecuteStateMachine
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
					LastSuccessfulVersion: test.statusVersion,
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
			res, err := r.PreRemoveFinalizer(rctx, &unstructured.Unstructured{Object: uObj})

			asserts.Equal(test.expectedStatemachineCalled, h.statemachineCalled)
			asserts.Equal(test.expectedError, err != nil)
			asserts.Equal(test.expectedRequeue, res.Requeue)
			asserts.NotNil(h.smHandler)
		})
	}
}

// TestPostRemoveFinalizer tests that the PostRemoveFinalizer implementation works correctly
// GIVEN a Finalizer
// WHEN the PostRemoveFinalizer method is called
// THEN ensure that it works correctly
func TestPostRemoveFinalizer(t *testing.T) {
	asserts := assert.New(t)

	const namespace = "testns"
	const name = "test"

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
	}

	scheme := initScheme()
	clientBuilder := fakes.NewClientBuilder().WithScheme(scheme)
	r := Reconciler{
		Client: clientBuilder.Build(),
		Scheme: initScheme(),
	}

	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	asserts.NoError(err)
	r.PostRemoveFinalizer(rctx, &unstructured.Unstructured{Object: uObj})
}

// TestGetName tests that the GetName implementation works correctly
// GIVEN a Finalizer
// WHEN the GetName method is called
// THEN ensure that the correct name is returned
func TestGetName(t *testing.T) {
	asserts := assert.New(t)
	const expectedName = "module.platform.verrazzano.io/finalizer"
	asserts.Equal(expectedName, Reconciler{}.GetName())
}

func (h *finalizerHandler) fakeExecuteStateMachine(ctx handlerspi.HandlerContext, sm statemachine.StateMachine) ctrl.Result {
	h.statemachineCalled = true
	h.smHandler = sm.Handler
	if h.statemachineError {
		return result.NewRequeueWithShortDelay()
	}
	return ctrl.Result{}
}

func (h finalizerHandler) GetWorkName() string {
	return "install"
}

func (h finalizerHandler) IsWorkNeeded(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h finalizerHandler) PreWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) PreWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) IsPreActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h finalizerHandler) DoWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) DoWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) IsWorkDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h finalizerHandler) PostWorkUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) PostWork(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) IsPostActionDone(context handlerspi.HandlerContext) (bool, ctrl.Result, error) {
	return true, ctrl.Result{}, nil
}

func (h finalizerHandler) WorkCompletedUpdateStatus(context handlerspi.HandlerContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (h finalizerHandler) fakeLoadHelmInfo(cr *moduleapi.Module) (handlerspi.HelmInfo, error) {
	if h.loadHelmInfoErr == "" {
		return handlerspi.HelmInfo{}, nil
	}
	return handlerspi.HelmInfo{}, errors.New(h.loadHelmInfoErr)
}
