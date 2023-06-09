// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package operator_test

import (
	"context"

	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestOperatorRunning tests the ruuning status of module-operator.
// GIVEN an installation of module-operator in a cluster
// WHEN the status of ready replicas for the module-operator are checked
// THEN 1 replica is found to be ready.
func (suite *OperatorTestSuite) XTestOperatorRunning() {
	client, err := k8sutil.GetKubernetesClientset()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	deployment, err := client.AppsV1().Deployments("verrazzano-install").Get(context.TODO(), "verrazzano-module-operator", v1.GetOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(int(deployment.Status.ReadyReplicas)).To(gomega.Equal(1))
}

// TestCRDsInstalled tests the installation status of modules crds.
// GIVEN an installation of module-operator in a cluster
// WHEN the status of installtion of module crd is checked
// THEN module crd is found to be installed.
func (suite *OperatorTestSuite) XTestCRDsInstalled() {
	crdInstalled, err := k8sutil.CheckCRDsExist([]string{"modules.platform.verrazzano.io"})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(crdInstalled).To(gomega.BeTrue())
}
