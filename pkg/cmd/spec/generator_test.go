package spec

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/require"
	kclgo "kcl-lang.io/kcl-go"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/kcl"
	"kusionstack.io/kusion/pkg/models"
	appconfigmodel "kusionstack.io/kusion/pkg/models/appconfiguration"
	"kusionstack.io/kusion/pkg/projectstack"
)

var (
	spec1 = `
resources:
- id: v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
      creationTimestamp: null
    spec: {}
    status: {}
`
	specModel1 = &models.Spec{
		Resources: []models.Resource{
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"spec":       make(map[string]interface{}),
					"status":     make(map[string]interface{}),
					"metadata": map[string]interface{}{
						"name":              "default",
						"creationTimestamp": nil,
					},
				},
			},
		},
	}

	spec2 = `
resources:
- id: v1:Namespace:default
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: default
- id: v1:Namespace:kube-system
  type: Kubernetes
  attributes:
    apiVersion: v1
    kind: Namespace
    metadata:
      name: kube-system
`

	specModel2 = &models.Spec{
		Resources: []models.Resource{
			{
				ID:   "v1:Namespace:default",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "default",
					},
				},
			},
			{
				ID:   "v1:Namespace:kube-system",
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": "kube-system",
					},
				},
			},
		},
	}
)

func TestGenerateSpecFromFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    *models.Spec
		wantErr bool
	}{
		{
			name:    "test1",
			path:    "kusion_spec.yaml",
			content: spec1,
			want:    specModel1,
		},
		{
			name:    "test2",
			path:    "kusion_spec.yaml",
			content: spec2,
			want:    specModel2,
		},
		{
			name:    "test3",
			path:    "kusion_spec.yaml",
			content: `k1: v1`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.path)
			file.Write([]byte(tt.content))
			defer os.Remove(tt.path)
			got, err := GenerateSpecFromFile(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateSpec(t *testing.T) {
	apc := &appconfigmodel.AppConfiguration{}
	var apcMap map[string]interface{}
	tmp, _ := json.Marshal(apc)
	_ = json.Unmarshal(tmp, &apcMap)

	type args struct {
		o       *generator.Options
		project *projectstack.Project
		stack   *projectstack.Stack
		mocker  *mockey.MockBuilder
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Spec
		wantErr bool
	}{
		{
			name: "nil generator", args: struct {
				o       *generator.Options
				project *projectstack.Project
				stack   *projectstack.Stack
				mocker  *mockey.MockBuilder
			}{
				o:       &generator.Options{},
				project: &projectstack.Project{},
				stack:   &projectstack.Stack{},
				mocker:  mockey.Mock((*kcl.Generator).GenerateSpec).Return(nil, nil),
			},
			want: nil,
		},
		{
			name: "kcl generator", args: struct {
				o       *generator.Options
				project *projectstack.Project
				stack   *projectstack.Stack
				mocker  *mockey.MockBuilder
			}{
				o: &generator.Options{},
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Generator: &projectstack.GeneratorConfig{
							Type: projectstack.KCLGenerator,
						},
					},
				},
				stack:  &projectstack.Stack{},
				mocker: mockey.Mock((*kcl.Generator).GenerateSpec).Return(nil, nil),
			},
			want: nil,
		},
		{
			name: "app generator", args: struct {
				o       *generator.Options
				project *projectstack.Project
				stack   *projectstack.Stack
				mocker  *mockey.MockBuilder
			}{
				o: &generator.Options{Arguments: map[string]string{}},
				project: &projectstack.Project{
					ProjectConfiguration: projectstack.ProjectConfiguration{
						Name: "default",
						Generator: &projectstack.GeneratorConfig{
							Type: projectstack.AppConfigurationGenerator,
						},
					},
				},
				stack: &projectstack.Stack{},
				mocker: mockey.Mock(kcl.Run).Return(&kcl.CompileResult{
					Documents: []kclgo.KCLResult{apcMap},
				}, nil),
			},
			want: specModel1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.args.mocker.Build()
			defer m.UnPatch()

			got, err := GenerateSpec(tt.args.o, tt.args.project, tt.args.stack)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}
