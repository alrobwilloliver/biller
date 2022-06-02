package project

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"biller/svc/compute/store"

	"github.com/gogo/status"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

type FakeFieldMask struct {
	*fieldmaskpb.FieldMask
	paths []string
	err   error
}

func (f FakeFieldMask) Normalize() error {
	return nil
}

func (f FakeFieldMask) GetPaths() []string {
	return f.paths
}

type FakeTx struct {
	pgx.Tx
	err error
}

func (tx FakeTx) Rollback(context.Context) error {
	return nil
}

func (tx FakeTx) Commit(ctx context.Context) error {
	return tx.err
}

type FakeTxQuerier struct {
	store.TxQuerier
	billingAccount         store.BillingAccount
	billingAccountError    error
	createdProject         store.Project
	createProjectError     error
	deleteProjectInt       int64
	deleteProjectError     error
	listProjects           []store.Project
	listProjectsError      error
	findProjectByIdError   error
	project                store.Project
	selectProjectForUpdate store.Project
	selectError            error
	tx                     FakeTx
	txErr                  error
	updateProject          store.Project
	updateProjectError     error
}

func (q FakeTxQuerier) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, store.Querier, error) {
	return q.tx, q, q.txErr
}

func (q FakeTxQuerier) ExecWithTx(context.Context, pgx.TxOptions, func(store.Querier) error) error {
	return nil
}

func (q FakeTxQuerier) FindBillingAccountById(ctx context.Context, id string) (store.BillingAccount, error) {
	return q.billingAccount, q.billingAccountError
}

func (q FakeTxQuerier) CreateProject(ctx context.Context, arg store.CreateProjectParams) (store.Project, error) {
	return q.createdProject, q.createProjectError
}

func (q FakeTxQuerier) FindProjectById(ctx context.Context, id string) (store.Project, error) {
	return q.project, q.findProjectByIdError
}

func (q FakeTxQuerier) DeleteProject(ctx context.Context, id string) (int64, error) {
	return q.deleteProjectInt, q.deleteProjectError
}

func (q FakeTxQuerier) ListProjects(ctx context.Context, arg store.ListProjectsParams) ([]store.Project, error) {
	return q.listProjects, q.listProjectsError
}

func (q FakeTxQuerier) SelectProjectForUpdate(ctx context.Context, id string) (store.Project, error) {
	return q.selectProjectForUpdate, q.selectError
}

func (q FakeTxQuerier) UpdateProject(ctx context.Context, arg store.UpdateProjectParams) (store.Project, error) {
	return q.updateProject, q.updateProjectError
}
func Test_CreateProject(t *testing.T) {
	t.Run("should fail if no id is passed in the request", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id: "",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail if the name is invalid", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id: "invalid&$!@",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail if the billing account id is invalid", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "invalid£$^",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail when BeginTx returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.txErr = errors.New("failure to begin transaction")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "b1b391aa-5755-4ca8-966e-8ef8b0f30664",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Fatalf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to begin transaction" {
			t.Errorf("expected: %s, got: %s", "failure to begin transaction", err.Error())
		}
	})
	t.Run("should fail when FindBillingAccountById query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = errors.New("failure to find billing account")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "b1b391aa-5755-4ca8-966e-8ef8b0f30664",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when FindBillingAccountById query returns no result", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = pgx.ErrNoRows
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "b1b391aa-5755-4ca8-966e-8ef8b0f30664",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.NotFound {
			t.Errorf("expected: %s, got: %s", codes.NotFound, st.Code())
		}
	})
	t.Run("should fail when FindBillingAccountById returns an account with DemandEnabled set to false", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = nil
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: false,
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "b1b391aa-5755-4ca8-966e-8ef8b0f30664",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.PermissionDenied {
			t.Errorf("expected: %s, got: %s", codes.PermissionDenied, st.Code())
		}
	})
	t.Run("should fail when CreateProject query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.createProjectError = errors.New("failure to create project")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "valid-billing-account-id",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		fmt.Print("err: ", err)
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when commiting the transaction returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.createdProject = store.Project{
			ID: "test-project-id",
		}
		querier.tx = FakeTx{
			err: errors.New("failure to commit transaction"),
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "test-billing-id",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should return a project when everything is successful", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.createdProject = store.Project{
			ID: "test-project-id",
		}
		querier.tx = FakeTx{}
		server := NewServer(querier, zaptest.NewLogger(t))
		resp, err := server.CreateProject(context.Background(), &CreateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "b1b391aa-5755-4ca8-966e-8ef8b0f30664",
			},
		})
		if err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
		if resp.Id != querier.createdProject.ID {
			t.Errorf("expected: %s, got: %s", querier.createdProject.ID, resp.Id)
		}
		if resp.BillingAccountId != querier.createdProject.BillingAccountID {
			t.Errorf("expected: %s, got: %s", querier.createdProject.BillingAccountID, resp.BillingAccountId)
		}
		if resp.Id != querier.createdProject.ID {
			t.Errorf("expected: %s, got: %s", querier.createdProject.ID, resp.Id)
		}
	})
}

func Test_DeleteProject(t *testing.T) {
	t.Run("should fail when the project has no Id passed in the request", func(t *testing.T) {
		server := NewServer(nil, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail when BeginTx returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.txErr = errors.New("failure to begin transaction")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Fatalf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to begin transaction" {
			t.Errorf("expected: %s, got: %s", "failure to begin transaction", err.Error())
		}
	})
	t.Run("should fail when FindProjectById query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.findProjectByIdError = errors.New("failure to find project")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Fatalf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to find project" {
			t.Errorf("expected: %s, got: %s", "failure to find project", err.Error())
		}
	})
	t.Run("should fail when DeleteProject query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.deleteProjectError = errors.New("failure to delete project")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when DeleteProject returns 0 deleted rows", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.deleteProjectInt = 0
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when commiting the transaction returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.tx = FakeTx{
			err: errors.New("failure to commit transaction"),
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should successfully delete a project", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.deleteProjectInt = 1
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.DeleteProject(context.Background(), &DeleteProjectRequest{
			Id: "test",
		})
		if err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
	})
}

func Test_GetProject(t *testing.T) {
	t.Run("should fail when the name passed in the request is empty", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail when the BeginTx query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.txErr = errors.New("failure to begin transaction")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Errorf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to begin transaction" {
			t.Errorf("expected: %s, got: %s", "failure to begin transaction", err.Error())
		}
	})
	t.Run("should fail when the FindProjectById query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.findProjectByIdError = errors.New("failure to find project")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when the FindProjectById query returns no rows", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.findProjectByIdError = pgx.ErrNoRows
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.NotFound {
			t.Errorf("expected: %s, got: %s", codes.NotFound, st.Code())
		}
	})
	t.Run("should fail when the authorization checker Check returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "test",
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Errorf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to check authorization" {
			t.Errorf("expected: %s, got: %s", "failure to check authorization", err.Error())
		}
	})
	t.Run("should successfully return a project", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.project = store.Project{
			ID:               "project-id",
			BillingAccountID: "billing-account-id",
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		project, err := server.GetProject(context.Background(), &GetProjectRequest{
			Id: "test",
		})
		if err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
		if project.Id != "test" {
			t.Errorf("expected: %s, got: %s", "test", project.Id)
		}
		if project.BillingAccountId != "billing-account-id" {
			t.Errorf("expected: %s, got: %s", "billing-account-id", project.BillingAccountId)
		}
	})
}

func Test_ListProjects(t *testing.T) {
	t.Run("should return an error when BeginTx returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.txErr = errors.New("failure to begin transaction")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.ListProjects(context.Background(), &ListProjectsRequest{})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Errorf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "failure to begin transaction" {
			t.Errorf("expected: %s, got: %s", "failure to begin transaction", err.Error())
		}
	})
	t.Run("should fail when parsing an invalid permitted id to a uuid", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.ListProjects(context.Background(), &ListProjectsRequest{})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when ListProjects query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.listProjectsError = errors.New("failure to list projects")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.ListProjects(context.Background(), &ListProjectsRequest{})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should successfully return a List of Projects", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.listProjects = []store.Project{
			{
				ID:               "test-1",
				BillingAccountID: "billing-account-id",
				CreateTime:       time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:               "test-2",
				BillingAccountID: "billing-account-id",
				CreateTime:       time.Date(2020, time.January, 1, 0, 0, 0, 1, time.UTC),
			},
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		projects, err := server.ListProjects(context.Background(), &ListProjectsRequest{})
		if err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
		if len(projects.Projects) != 2 {
			t.Errorf("expected: %d, got: %d", 2, len(projects.Projects))
		}
		if projects.Projects[0].Id != "test-1" {
			t.Errorf("expected: %s, got: %s", "test-1", projects.Projects[0].Id)
		}
		if projects.Projects[0].BillingAccountId != "billing-account-id" {
			t.Errorf("expected: %s, got: %s", "billing-account-id", projects.Projects[0].BillingAccountId)
		}
		if projects.Projects[1].Id != "test-2" {
			t.Errorf("expected: %s, got: %s", "test-2", projects.Projects[1].Id)
		}
		if projects.Projects[1].BillingAccountId != "billing-account-id" {
			t.Errorf("expected: %s, got: %s", "billing-account-id", projects.Projects[1].BillingAccountId)
		}
	})
}

func Test_UpdateProject(t *testing.T) {
	t.Run("should fail when an invalid project uid is passed in the request", func(t *testing.T) {
		querier := FakeTxQuerier{}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id: "invalid-uid",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail when the BeginTx query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.txErr = fmt.Errorf("test error")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id: "f863f339-9f9c-4863-aa02-73e059060b9a",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Errorf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "test error" {
			t.Errorf("expected: %s, got: %s", "test error", err.Error())
		}
	})
	t.Run("should fail when SelectProjectForUpdate returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.selectError = fmt.Errorf("test error")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id: "test-id",
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if ok {
			t.Errorf("expected a non grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Unknown {
			t.Errorf("expected: %s, got: %s", codes.Unknown, st.Code())
		}
		if err.Error() != "test error" {
			t.Errorf("expected: %s, got: %s", "test error", err.Error())
		}
	})
	t.Run("should fail when using the mask updating the account id is invalid", func(t *testing.T) {
		querier := FakeTxQuerier{}

		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "f863f339-9f9c-4863-aa02-73e059060b9a",
				BillingAccountId: "invalid-account-id",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected: %s, got: %s", codes.InvalidArgument, st.Code())
		}
	})
	t.Run("should fail when FindBillingAccountByID returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = fmt.Errorf("test error")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test-id",
				BillingAccountId: "billing-account-id",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when FindBillingAccountByID returns no rows", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = pgx.ErrNoRows
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test-id",
				BillingAccountId: "billing-account-id",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.NotFound {
			t.Errorf("expected: %s, got: %s", codes.NotFound, st.Code())
		}
	})
	t.Run("should fail when updating by account id with an account with DemandEnabled set to false", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = nil
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: false,
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test-id",
				BillingAccountId: "billing-account-id",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.PermissionDenied {
			t.Errorf("expected: %s, got: %s", codes.PermissionDenied, st.Code())
		}
	})
	t.Run("should fail when the UpdateProject query returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = nil
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.updateProjectError = fmt.Errorf("update project error")
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "billing-account-id",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id", "name"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should fail when transaction commit returns an error", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = nil
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.updateProjectError = nil
		querier.tx = FakeTx{
			err: fmt.Errorf("commit error"),
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		_, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "f863f339-9f9c-4863-aa02-73e059060b9a",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id", "name"},
			},
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("expected a grpc error, got: %s", err.Error())
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected: %s, got: %s", codes.Internal, st.Code())
		}
	})
	t.Run("should successfully return an updated project", func(t *testing.T) {
		querier := FakeTxQuerier{}
		querier.billingAccountError = nil
		querier.billingAccount = store.BillingAccount{
			DemandEnabled: true,
		}
		querier.updateProject = store.Project{
			ID:               "test-update",
			BillingAccountID: "billing-account-id-update",
		}
		server := NewServer(querier, zaptest.NewLogger(t))
		updatedProject, err := server.UpdateProject(context.Background(), &UpdateProjectRequest{
			Project: &Project{
				Id:               "test",
				BillingAccountId: "billing-account-id-update",
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"account_id", "name"},
			},
		})
		if err != nil {
			t.Errorf("expected no error, got: %s", err.Error())
		}
		if updatedProject == nil {
			t.Errorf("expected project, got nil")
		}
		if updatedProject.Id != "test-update" {
			t.Errorf("expected project name to be %s, got: %s", "test-update", updatedProject.Id)
		}
		if updatedProject.BillingAccountId != "billing-account-id-update" {
			t.Errorf("expected project billing account id to be billing-account-id-update, got: %s", updatedProject.BillingAccountId)
		}
	})
}
