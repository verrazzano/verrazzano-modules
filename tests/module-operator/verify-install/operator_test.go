// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package operator_test

import (
	"context"

	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano-modules/pkg/k8sutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *OperatorTestSuite) TestOperatorRunning() {
	client, err := k8sutil.GetKubernetesClientset()
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	deployment, err := client.AppsV1().Deployments("verrazzano-install").Get(context.TODO(), "verrazzano-module-operator", v1.GetOptions{})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(int(deployment.Status.ReadyReplicas)).To(gomega.Equal(1))
}

func (suite *OperatorTestSuite) TestCRDsInstalled() {
	crdInstalled, err := k8sutil.CheckCRDsExist([]string{"modules.platform.verrazzano.io", "moduleactions.platform.verrazzano.io"})
	suite.gomega.Expect(err).NotTo(gomega.HaveOccurred())
	suite.gomega.Expect(crdInstalled).To(gomega.BeTrue())
}
