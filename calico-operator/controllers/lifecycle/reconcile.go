// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package lifecycle

import (
	"errors"
	"fmt"
	vzerrors "github.com/verrazzano/verrazzano/pkg/controller/errors"
	networkapi "github.com/vz-app/operator/apis/network/v1beta1"
	"github.com/vz-app/operator/constants"
	"github.com/vz-app/operator/microcontrollers/common/spi"
	corev1 "k8s.io/api/core/v1"
	k8err "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Reconcile updates the Certificate
func (r Reconciler) Reconcile(ctx spi.ReconcileContext, u *unstructured.Unstructured) error {
	dns := &networkapi.DNS{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, dns); err != nil {
		return err
	}
	// Build the domain name and update the status
	domain, err := r.buildDomainName(ctx, dns)
	if err != nil {
		return err
	}
	if domain != dns.Status.HostName {
		dns.Status.HostName = domain
		err = r.Client.Status().Update(ctx.ClientCtx, dns)
		if err != nil {
			return ctx.Log.ErrorfNewErr("Failed to update the Verrazzano DNS resource: %v", err)
		}
	}
	return nil
}

// buildDomainName generates a domain name
func (r *Reconciler) buildDomainName(ctx spi.ReconcileContext, dns *networkapi.DNS) (string, error) {
	if len(dns.Spec.Subdomain) == 0 {
		return "", ctx.Log.ErrorfNewErr("Failed: empty Subdomain field in Verrazzano DNS resource")
	}

	if dns.Spec.OCI != nil {
		return "", errors.New("Failed, OCI DNS not yet supported")
	}

	if dns.Spec.External != nil {
		return "", errors.New("Failed, External DNS not yet supported")
	}

	return r.buildDomainNameForWildcard(ctx, dns)
}

// buildDomainNameForWildcard generates a domain name in the format of "<IP>.<wildcard-domain>"
// Get the IP from Istio resources
func (r *Reconciler) buildDomainNameForWildcard(ctx spi.ReconcileContext, dns *networkapi.DNS) (string, error) {
	var IP string

	wildcard := "nip.io"
	if dns.Spec.Wildcard != nil {
		wildcard = dns.Spec.Wildcard.Domain
	}

	// Need to discover a service with an IP that can be used
	var err error
	IP, err = r.discoverIngressIP(ctx, dns.Spec.ID)
	if err != nil {
		return "", err
	}

	domain := fmt.Sprintf("%s.%s.%s", dns.Spec.Subdomain, IP, wildcard)
	return domain, nil
}

// Find a service with an IP that provides ingress into the cluster
func (r *Reconciler) discoverIngressIP(ctx spi.ReconcileContext, requiredDNSID string) (string, error) {
	serviceLlist := corev1.ServiceList{}

	ctx.Log.Progress("DNS IP discovery looking for services in any namespace")
	err := r.Client.List(ctx.ClientCtx, &serviceLlist)
	if k8err.IsNotFound(err) {
		ctx.Log.Progress("DNS IP discovery cannot find any matching services")
		return "", vzerrors.RetryableError{}
	}
	if err != nil {
		ctx.Log.ErrorfNewErr("Failed in DNS IP discovery: %v", err)
		return "", err
	}
	for _, service := range serviceLlist.Items {
		if service.Labels != nil {
			dnsID, ok := service.Labels[constants.DnsIdLabel]
			if !ok {
				continue
			}
			if dnsID != requiredDNSID {
				continue
			}
		}
		IP := getIPFromService(&service)
		if len(IP) > 0 {
			ctx.Log.Oncef("Using IP %s in service %s/%s", IP, service.Namespace, service.Name)
			return IP, nil
		}
	}
	ctx.Log.Progress("Waiting for a service with matching DNS ID labels that has status IP set")
	return "", vzerrors.RetryableError{}
}

// getIPFromService gets the External IP or Ingress IP from a service
func getIPFromService(service *corev1.Service) string {
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer || service.Spec.Type == corev1.ServiceTypeNodePort {
		if len(service.Spec.ExternalIPs) > 0 {
			return service.Spec.ExternalIPs[0]
		}
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			return service.Status.LoadBalancer.Ingress[0].IP
		}
	}
	return ""
}
