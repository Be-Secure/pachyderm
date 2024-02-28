package server

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"testing"
)

func withDB(t *testing.T, testCase func(context.Context, *testing.T, *pachsql.DB)) {
	t.Helper()
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	testCase(ctx, t, db)
}

func withPicker(t *testing.T, testCase func(context.Context, *testing.T, *picker)) {
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		p := &picker{db: db}
		testCase(ctx, t, p)
	})
}

func TestPicker_PickProject(t *testing.T) {
	pickerName := &pfs.ProjectPicker{
		Picker: &pfs.ProjectPicker_Name{
			Name: "default",
		},
	}
	expected := &pfsdb.ProjectWithID{
		ID: 1,
		ProjectInfo: &pfs.ProjectInfo{
			Project: &pfs.Project{
				Name: "default",
			},
		},
	}
	withPicker(t, func(ctx context.Context, t *testing.T, picker *picker) {
		got, err := picker.PickProject(ctx, pickerName)
		require.NoError(t, err, "should be able to pick project")
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf("project info with id (-got +want):\n%s", diff)
		}
	})
}
