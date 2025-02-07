package monitoring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	modelsapp "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/models/appconfiguration/monitoring"
	"kusionstack.io/kusion/pkg/projectstack"
)

func Test_monitoringPatcher_Patch(t *testing.T) {
	spec := &models.Intent{}
	err := appconfiguration.AppendToSpec(models.Kubernetes, "id", spec, buildMockDeployment())
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		appName string
		app     *modelsapp.AppConfiguration
		project *projectstack.Project
	}
	type args struct {
		resources map[string][]*models.Resource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "operatorModeTrue",
			fields: fields{
				appName: "test",
				app: &modelsapp.AppConfiguration{
					Monitoring: &monitoring.Monitor{},
				},
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Prometheus: &projectstack.PrometheusConfig{
							OperatorMode: true,
						},
					},
				},
			},
			args: args{
				resources: spec.Resources.GVKIndex(),
			},
			wantErr: assert.NoError,
		},
		{
			name: "operatorModeFalse",
			fields: fields{
				appName: "test",
				app: &modelsapp.AppConfiguration{
					Monitoring: &monitoring.Monitor{},
				},
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Prometheus: &projectstack.PrometheusConfig{
							OperatorMode: false,
						},
					},
				},
			},
			args: args{
				resources: spec.Resources.GVKIndex(),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &monitoringPatcher{
				appName: tt.fields.appName,
				app:     tt.fields.app,
				project: tt.fields.project,
			}
			tt.wantErr(t, p.Patch(tt.args.resources), fmt.Sprintf("Patch(%v)", tt.args.resources))
			// check if the deployment is patched
			var deployment appsv1.Deployment
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(spec.Resources[0].Attributes, &deployment); err != nil {
				t.Fatal(err)
			}
			if tt.fields.project.Prometheus.OperatorMode {
				assert.NotNil(t, deployment.Labels)
				assert.NotNil(t, deployment.Spec.Template.Labels)
				assert.Equal(t, deployment.Labels["kusion_monitoring_appname"], tt.fields.appName)
				assert.Equal(t, deployment.Spec.Template.Labels["kusion_monitoring_appname"], tt.fields.appName)
			} else {
				assert.NotNil(t, deployment.Annotations)
				assert.NotNil(t, deployment.Spec.Template.Annotations)
				assert.Equal(t, deployment.Annotations["prometheus.io/scrape"], "true")
				assert.Equal(t, deployment.Annotations["prometheus.io/scheme"], tt.fields.app.Monitoring.Scheme)
				assert.Equal(t, deployment.Annotations["prometheus.io/path"], tt.fields.app.Monitoring.Path)
				assert.Equal(t, deployment.Annotations["prometheus.io/port"], tt.fields.app.Monitoring.Port)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/scrape"], "true")
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/scheme"], tt.fields.app.Monitoring.Scheme)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/path"], tt.fields.app.Monitoring.Path)
				assert.Equal(t, deployment.Spec.Template.Annotations["prometheus.io/port"], tt.fields.app.Monitoring.Port)
			}
		})
	}
}

// generate a mock Deployment
func buildMockDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mock-deployment",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		Spec: appsv1.DeploymentSpec{},
	}
}

func TestNewMonitoringPatcherFunc(t *testing.T) {
	type args struct {
		appName string
		app     *modelsapp.AppConfiguration
		project *projectstack.Project
	}
	tests := []struct {
		name string
		args args
		want appconfiguration.NewPatcherFunc
	}{
		{
			name: "NewMonitoringPatcherFunc",
			args: args{
				appName: "test",
				app:     &modelsapp.AppConfiguration{},
				project: &projectstack.Project{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcherFunc := NewMonitoringPatcherFunc(tt.args.appName, tt.args.app, tt.args.project)
			assert.NotNil(t, patcherFunc)
			patcher, err := patcherFunc()
			assert.NoError(t, err)
			assert.Equal(t, tt.args.appName, patcher.(*monitoringPatcher).appName)
		})
	}
}
