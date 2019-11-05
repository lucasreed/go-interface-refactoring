// Copyright © 2019Luke Reed
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"regexp"

	extensionsv1beta "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	// Importing gcp auth client so we can get to GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// TODO: Add Services to the GetHosts() function

const (
	annotationHostnameKey string = "external-dns.alpha.kubernetes.io/hostname"
)

// Kube is a wrapper around the kubernetes interface and holds all our relevant info
type Kube struct {
	Client            kubernetes.Interface
	Namespaces        []string
	Domain            string // Domain is the DNS domain we want to match against for hostname checking
	IgnoredSubDomains []string
	ValidTargets      map[string][]string
}

// Hostname represents a public facing hostname along with the resource within the cluster that it points to
type Hostname string

// Resource maps to a single kube resource e.g. ingress
type Resource struct {
	Name      string
	Namespace string
	Kind      string
	Target    string
}

// Hostnames map a Hostname to a Resource
type Hostnames map[Hostname][]Resource

// New provides the kube client and a few other pieces of information needed when interacting with the cluster
func New(domain string, ignoredSubdomains []string) *Kube {
	var ret = Kube{
		Client:            getKubeClient(),
		Domain:            domain,
		IgnoredSubDomains: ignoredSubdomains,
		ValidTargets:      make(map[string][]string),
	}
	ret.Namespaces = ret.getNamespaces()
	return &ret
}

func getKubeClient() kubernetes.Interface {
	kubeConf, err := config.GetConfig()
	if err != nil {
		klog.Fatalf("Error getting kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		klog.Fatalf("Error creating kubernetes client: %v", err)
	}
	return clientset
}

func (k *Kube) getNamespaces() []string {
	var ret []string
	namespaces, err := k.Client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		klog.Fatalf("Error getting namespaces: %v", err)
	}
	for _, ns := range namespaces.Items {
		ret = append(ret, ns.Name)
	}
	return ret
}

// GetHosts returns a map of hostnames that are present in ingresses indexing them to resources
func (k *Kube) GetHosts() Hostnames {
	var hosts = make(Hostnames)
	for _, ns := range k.Namespaces {
		i, err := k.Client.ExtensionsV1beta1().Ingresses(ns).List(metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("Error getting ingresses in namespace %s: %v", ns, err)
		}
		s, err := k.Client.CoreV1().Services(ns).List(metav1.ListOptions{})
		if err != nil {
			klog.Fatalf("Error getting services in namespace %s: %v", ns, err)
		}
		for _, service := range s.Items {
			var target string
			if len(service.Status.LoadBalancer.Ingress) == 1 {
				if service.Status.LoadBalancer.Ingress[0].IP != "" {
					target = service.Status.LoadBalancer.Ingress[0].IP
				}
				if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
					target = service.Status.LoadBalancer.Ingress[0].Hostname
				}
			}
			k.ValidTargets[target] = append(k.ValidTargets[target], "service/"+service.Name)
			if host := hostFromAnnotation(service.Annotations, k.Domain, k.IgnoredSubDomains); host != "" {
				hosts[host] = append(hosts[host], Resource{
					Name:      service.Name,
					Namespace: ns,
					Kind:      "service",
					Target:    target,
				})
			}
		}
		for _, ingress := range i.Items {
			var target string
			if len(ingress.Status.LoadBalancer.Ingress) == 1 {
				if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
					target = ingress.Status.LoadBalancer.Ingress[0].IP
				}
				if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
					target = ingress.Status.LoadBalancer.Ingress[0].Hostname
				}
			}
			k.ValidTargets[target] = append(k.ValidTargets[target], "ingress/"+ingress.Name)
			if host := hostFromAnnotation(ingress.Annotations, k.Domain, k.IgnoredSubDomains); host != "" {
				hosts[host] = append(hosts[host], Resource{
					Name:      ingress.Name,
					Namespace: ns,
					Kind:      "ingress",
					Target:    target,
				})
			} else {
				for _, host := range hostsFromIngressRules(ingress.Spec.Rules, k.Domain, k.IgnoredSubDomains) {
					hosts[host] = append(hosts[host], Resource{
						Name:      ingress.Name,
						Namespace: ns,
						Kind:      "ingress",
						Target:    target,
					})
				}
			}
		}
	}
	return hosts
}

func hostFromAnnotation(annotations map[string]string, domain string, ignoredSubdomains []string) Hostname {
	var ret Hostname
	dom, _ := regexp.Compile(`.*` + domain + `\.??`)
	if host, exists := annotations[annotationHostnameKey]; exists {
		for _, sub := range ignoredSubdomains {
			if isIgnored(host, sub) {
				return ret
			}
		}
		if dom.Match([]byte(host)) {
			ret = Hostname(host)
		}
	}
	return ret
}

func hostsFromIngressRules(rules []extensionsv1beta.IngressRule, domain string, ignoredSubdomains []string) []Hostname {
	var hosts []Hostname
	re, _ := regexp.Compile(`.*` + domain + `\.??`)
Rules:
	for _, rule := range rules {
		if rule.Host == "" {
			continue
		}
		if !re.Match([]byte(rule.Host)) {
			continue
		}
		for _, sub := range ignoredSubdomains {
			if isIgnored(rule.Host, sub) {
				continue Rules
			}
		}
		hosts = append(hosts, Hostname(rule.Host))
	}
	return hosts
}

func isIgnored(host string, ignoredSubdomain string) bool {
	sub, _ := regexp.Compile(`.*` + ignoredSubdomain + `\.??`)
	return sub.Match([]byte(host))
}
