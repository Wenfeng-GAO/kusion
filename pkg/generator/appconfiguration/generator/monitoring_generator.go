package generator

import (
	"fmt"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/monitoring"
	"kusionstack.io/kusion/pkg/projectstack"
)

type monitoringGenerator struct {
	project *projectstack.Project
	monitor *monitoring.Monitor
	appName string
}

func NewMonitoringGenerator(project *projectstack.Project, monitor *monitoring.Monitor, appName string) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	if len(appName) == 0 {
		return nil, fmt.Errorf("app name must not be empty")
	}
	return &monitoringGenerator{
		project: project,
		monitor: monitor,
		appName: appName,
	}, nil
}

func NewMonitoringGeneratorFunc(project *projectstack.Project, monitor *monitoring.Monitor, appName string) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewMonitoringGenerator(project, monitor, appName)
	}
}

func (g *monitoringGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	// If Prometheus runs as an operator, it relies on Custom Resources to
	// manage the scrape configs. CRs (ServiceMonitors and PodMonitors) rely on
	// corresponding resources (Services and Pods) to have labels that can be
	// used as part of the label selector for the CR to determine which
	// service/pods to scrape from.
	// Here we choose the label name kusion_monitoring_appname for two reasons:
	// 1. Unlike the label validation in Kubernetes, the label name accepted by
	// Prometheus cannot contain non-alphanumeric characters except underscore:
	// https://github.com/prometheus/common/blob/main/model/labels.go#L94
	// 2. The name should be unique enough that is only created by Kusion and
	// used to identify a certain application
	monitoringLabels := map[string]string{
		"kusion_monitoring_appname": g.appName,
	}

	if g.project.ProjectConfiguration.Prometheus != nil && g.project.ProjectConfiguration.Prometheus.OperatorMode && g.monitor != nil {
		if g.project.ProjectConfiguration.Prometheus.MonitorType == projectstack.ServiceMonitorType {
			serviceEndpoint := prometheusv1.Endpoint{
				Interval:      g.monitor.Interval,
				ScrapeTimeout: g.monitor.Timeout,
				Port:          g.monitor.Port,
				Path:          g.monitor.Path,
				Scheme:        g.monitor.Scheme,
			}
			serviceEndpointList := []prometheusv1.Endpoint{serviceEndpoint}
			serviceMonitor := &prometheusv1.ServiceMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceMonitor",
					APIVersion: prometheusv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-service-monitor", g.appName), Namespace: g.project.Name},
				Spec: prometheusv1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: monitoringLabels,
					},
					Endpoints: serviceEndpointList,
				},
			}
			err := appconfiguration.AppendToSpec(
				models.Kubernetes,
				appconfiguration.KubernetesResourceID(serviceMonitor.TypeMeta, serviceMonitor.ObjectMeta),
				spec,
				serviceMonitor,
			)
			if err != nil {
				return err
			}
		} else if g.project.ProjectConfiguration.Prometheus.MonitorType == projectstack.PodMonitorType {
			podMetricsEndpoint := prometheusv1.PodMetricsEndpoint{
				Interval:      g.monitor.Interval,
				ScrapeTimeout: g.monitor.Timeout,
				Port:          g.monitor.Port,
				Path:          g.monitor.Path,
				Scheme:        g.monitor.Scheme,
			}
			podMetricsEndpointList := []prometheusv1.PodMetricsEndpoint{podMetricsEndpoint}

			podMonitor := &prometheusv1.PodMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodMonitor",
					APIVersion: prometheusv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-pod-monitor", g.appName), Namespace: g.project.Name},
				Spec: prometheusv1.PodMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: monitoringLabels,
					},
					PodMetricsEndpoints: podMetricsEndpointList,
				},
			}

			err := appconfiguration.AppendToSpec(
				models.Kubernetes,
				appconfiguration.KubernetesResourceID(podMonitor.TypeMeta, podMonitor.ObjectMeta),
				spec,
				podMonitor,
			)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("MonitorType should either be service or pod %s", g.project.ProjectConfiguration.Prometheus.MonitorType)
		}
	}

	return nil
}
