package apply

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/states/local"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
)

func TestApplyOptions_Run(t *testing.T) {
	mockey.PatchConvey("Detail is true", t, func() {
		mockeyPatchDetectProjectAndStack()
		mockeyPatchGenerateSpec()
		mockeyPatchNewKubernetesRuntime()
		mockeyPatchOperationPreview()

		o := NewApplyOptions()
		o.Detail = true
		o.All = true
		o.NoStyle = true
		err := o.Run()
		assert.Nil(t, err)
	})

	mockey.PatchConvey("DryRun is true", t, func() {
		mockeyPatchDetectProjectAndStack()
		mockeyPatchGenerateSpec()
		mockeyPatchNewKubernetesRuntime()
		mockeyPatchOperationPreview()
		mockOperationApply(opsmodels.Success)

		o := NewApplyOptions()
		o.DryRun = true
		mockPromptOutput("yes")
		err := o.Run()
		assert.Nil(t, err)
	})
}

var (
	project = &projectstack.Project{
		ProjectConfiguration: projectstack.ProjectConfiguration{
			Name:   "testdata",
			Tenant: "admin",
		},
	}
	stack = &projectstack.Stack{
		StackConfiguration: projectstack.StackConfiguration{
			Name: "dev",
		},
	}
)

func mockeyPatchDetectProjectAndStack() *mockey.Mocker {
	return mockey.Mock(projectstack.DetectProjectAndStack).To(func(stackDir string) (*projectstack.Project, *projectstack.Stack, error) {
		project.Path = stackDir
		stack.Path = stackDir
		return project, stack, nil
	}).Build()
}

func mockeyPatchGenerateSpec() *mockey.Mocker {
	return mockey.Mock(spec.GenerateSpec).To(func(
		o *generator.Options,
		project *projectstack.Project,
		stack *projectstack.Stack,
	) (*models.Spec, error) {
		return &models.Spec{Resources: []models.Resource{sa1, sa2, sa3}}, nil
	}).Build()
}

func mockeyPatchNewKubernetesRuntime() *mockey.Mocker {
	return mockey.Mock(kubernetes.NewKubernetesRuntime).To(func() (runtime.Runtime, error) {
		return &fakerRuntime{}, nil
	}).Build()
}

var _ runtime.Runtime = (*fakerRuntime)(nil)

type fakerRuntime struct{}

func (f *fakerRuntime) Import(ctx context.Context, request *runtime.ImportRequest) *runtime.ImportResponse {
	return &runtime.ImportResponse{Resource: request.PlanResource}
}

func (f *fakerRuntime) Apply(ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
	return &runtime.ApplyResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Read(ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
	if request.PlanResource.ResourceKey() == "fake-id" {
		return &runtime.ReadResponse{
			Resource: nil,
			Status:   nil,
		}
	}
	return &runtime.ReadResponse{
		Resource: request.PlanResource,
		Status:   nil,
	}
}

func (f *fakerRuntime) Delete(ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
	return nil
}

func (f *fakerRuntime) Watch(ctx context.Context, request *runtime.WatchRequest) *runtime.WatchResponse {
	return nil
}

func mockeyPatchOperationPreview() *mockey.Mocker {
	return mockey.Mock((*operation.PreviewOperation).Preview).To(func(*operation.PreviewOperation, *operation.PreviewRequest) (rsp *operation.PreviewResponse, s status.Status) {
		return &operation.PreviewResponse{
			Order: &opsmodels.ChangeOrder{
				StepKeys: []string{sa1.ID, sa2.ID, sa3.ID},
				ChangeSteps: map[string]*opsmodels.ChangeStep{
					sa1.ID: {
						ID:     sa1.ID,
						Action: opsmodels.Create,
						From:   &sa1,
					},
					sa2.ID: {
						ID:     sa2.ID,
						Action: opsmodels.UnChanged,
						From:   &sa2,
					},
					sa3.ID: {
						ID:     sa3.ID,
						Action: opsmodels.Undefined,
						From:   &sa1,
					},
				},
			},
		}, nil
	}).Build()
}

const (
	apiVersion = "v1"
	kind       = "ServiceAccount"
	namespace  = "test-ns"
)

var (
	sa1 = newSA("sa1")
	sa2 = newSA("sa2")
	sa3 = newSA("sa3")
)

func newSA(name string) models.Resource {
	return models.Resource{
		ID:   engine.BuildID(apiVersion, kind, namespace, name),
		Type: "Kubernetes",
		Attributes: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func Test_apply(t *testing.T) {
	stateStorage := &local.FileSystemState{Path: filepath.Join("", local.KusionState)}
	mockey.PatchConvey("dry run", t, func() {
		planResources := &models.Spec{Resources: []models.Resource{sa1}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: opsmodels.Create,
					From:   sa1,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)
		o := NewApplyOptions()
		o.DryRun = true
		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply success", t, func() {
		mockOperationApply(opsmodels.Success)
		o := NewApplyOptions()
		planResources := &models.Spec{Resources: []models.Resource{sa1, sa2}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID, sa2.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: opsmodels.Create,
					From:   &sa1,
				},
				sa2.ID: {
					ID:     sa2.ID,
					Action: opsmodels.UnChanged,
					From:   &sa2,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)

		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.Nil(t, err)
	})
	mockey.PatchConvey("apply failed", t, func() {
		mockOperationApply(opsmodels.Failed)

		o := NewApplyOptions()
		planResources := &models.Spec{Resources: []models.Resource{sa1}}
		order := &opsmodels.ChangeOrder{
			StepKeys: []string{sa1.ID},
			ChangeSteps: map[string]*opsmodels.ChangeStep{
				sa1.ID: {
					ID:     sa1.ID,
					Action: opsmodels.Create,
					From:   &sa1,
				},
			},
		}
		changes := opsmodels.NewChanges(project, stack, order)

		err := Apply(o, stateStorage, planResources, changes, os.Stdout)
		assert.NotNil(t, err)
	})
}

func mockOperationApply(res opsmodels.OpResult) {
	mockey.Mock((*operation.ApplyOperation).Apply).To(
		func(o *operation.ApplyOperation, request *operation.ApplyRequest) (*operation.ApplyResponse, status.Status) {
			var err error
			if res == opsmodels.Failed {
				err = errors.New("mock error")
			}
			for _, r := range request.Spec.Resources {
				// ing -> $res
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   "",
					OpErr:      nil,
				}
				o.MsgCh <- opsmodels.Message{
					ResourceID: r.ResourceKey(),
					OpResult:   res,
					OpErr:      err,
				}
			}
			close(o.MsgCh)
			if res == opsmodels.Failed {
				return nil, status.NewErrorStatus(err)
			}
			return &operation.ApplyResponse{}, nil
		}).Build()
}

func Test_prompt(t *testing.T) {
	mockey.PatchConvey("prompt error", t, func() {
		mockey.Mock(survey.AskOne).Return(errors.New("mock error")).Build()
		_, err := prompt()
		assert.NotNil(t, err)
	})

	mockey.PatchConvey("prompt yes", t, func() {
		mockPromptOutput("yes")
		_, err := prompt()
		assert.Nil(t, err)
	})
}

func mockPromptOutput(res string) {
	mockey.Mock(survey.AskOne).To(func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		reflect.ValueOf(response).Elem().Set(reflect.ValueOf(res))
		return nil
	}).Build()
}
