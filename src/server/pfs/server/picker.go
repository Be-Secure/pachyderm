package server

import (
	"context"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pkg/errors"
)

// picker resolves a resource given
type picker struct {
	db *pachsql.DB
}

func (p *picker) PickProject(ctx context.Context, projectPicker *pfs.ProjectPicker) (*pfsdb.ProjectWithID, error) {
	if projectPicker == nil || projectPicker.Picker == nil {
		return nil, errors.New("project picker cannot be nil")
	}
	switch projectPicker.Picker.(type) {
	case *pfs.ProjectPicker_Name:
		var projectWithID *pfsdb.ProjectWithID
		var err error
		if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
			projectWithID, err = pfsdb.GetProjectWithID(ctx, tx, projectPicker.GetName())
			if err != nil {
				return errors.New("picking project")
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return projectWithID, nil
	default:
		return nil, errors.New(fmt.Sprintf("project picker is of an unknown type: %T", projectPicker.Picker))
	}
}

func (p *picker) PickRepo(ctx context.Context, repoPicker *pfs.RepoPicker) (*pfsdb.RepoInfoWithID, error) {
	if repoPicker == nil || repoPicker.Picker == nil {
		return nil, errors.New("repo picker cannot be nil")
	}
	switch repoPicker.Picker.(type) {
	case *pfs.RepoPicker_Name:
		picker := repoPicker.GetName()
		proj, err := p.PickProject(ctx, picker.Project)
		if err != nil {
			return nil, errors.Wrap(err, "picking repo")
		}
		repo := &pfs.Repo{
			Project: proj.ProjectInfo.Project,
			Type:    pfs.UserRepoType,
			Name:    picker.Name,
		}
		if picker.Type != "" {
			repo.Type = picker.Type
		}
		var repoInfoWithID *pfsdb.RepoInfoWithID
		if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
			repoInfoWithID, err = pfsdb.GetRepoInfoWithID(ctx, tx, repo.Project.Name, repo.Name, repo.Type)
			if err != nil {
				return errors.New("picking repo")
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return repoInfoWithID, nil
	default:
		return nil, errors.New(fmt.Sprintf("repo picker is of an unknown type: %T", repoPicker.Picker))
	}
}

func (p *picker) PickBranch(ctx context.Context, branchPicker *pfs.BranchPicker) (*pfsdb.BranchInfoWithID, error) {
	if branchPicker == nil || branchPicker.Picker == nil {
		return nil, errors.New("branch picker cannot be nil")
	}
	switch branchPicker.Picker.(type) {
	case *pfs.BranchPicker_Name:
		picker := branchPicker.GetName()
		repo, err := p.PickRepo(ctx, picker.Repo)
		if err != nil {
			return nil, errors.Wrap(err, "picking branch")
		}
		var branchInfoWithID *pfsdb.BranchInfoWithID
		if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
			branchInfoWithID, err = pfsdb.GetBranchInfoWithID(ctx, tx, &pfs.Branch{
				Repo: repo.RepoInfo.Repo,
				Name: picker.Name,
			})
			if err != nil {
				return errors.New("picking branch")
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return branchInfoWithID, nil
	default:
		return nil, errors.New(fmt.Sprintf("branch picker is of an unknown type: %T", branchPicker.Picker))
	}
}

func (p *picker) PickCommit(ctx context.Context, commitPicker *pfs.CommitPicker) (*pfsdb.CommitWithID, error) {
	if commitPicker == nil || commitPicker.Picker == nil {
		return nil, errors.New("commit picker cannot be nil")
	}
	switch commitPicker.Picker.(type) {
	case *pfs.CommitPicker_Id:
		return p.pickCommitGlobalID(ctx, commitPicker.GetId())
	case *pfs.CommitPicker_BranchHead:
		return p.pickCommitBranchHead(ctx, commitPicker.GetBranchHead())
	case *pfs.CommitPicker_Ancestor:
		return p.pickCommitAncestorOf(ctx, commitPicker.GetAncestor())
	case *pfs.CommitPicker_BranchRoot_:
		return p.pickCommitBranchRoot(ctx, commitPicker.GetBranchRoot())
	default:
		return nil, errors.New(fmt.Sprintf("commit picker is of an unknown type: %T", commitPicker.Picker))
	}
}

func (p *picker) pickCommitGlobalID(ctx context.Context, picker *pfs.CommitPicker_CommitByGlobalId) (*pfsdb.CommitWithID, error) {
	repo, err := p.PickRepo(ctx, picker.Repo)
	if err != nil {
		return nil, errors.Wrap(err, "picking commit")
	}
	var commitWithID *pfsdb.CommitWithID
	if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		commitWithID, err = pfsdb.GetCommitWithIDByKey(ctx, tx, &pfs.Commit{
			Repo: repo.RepoInfo.Repo,
			Id:   picker.Id,
		})
		return errors.Wrap(err, "picking commit")
	}); err != nil {
		return nil, err
	}
	return commitWithID, nil
}

func (p *picker) pickCommitBranchHead(ctx context.Context, branchHead *pfs.BranchPicker) (*pfsdb.CommitWithID, error) {
	branchInfoWithID, err := p.PickBranch(ctx, branchHead)
	if err != nil {
		return nil, errors.Wrap(err, "picking commit")
	}
	var commitWithID *pfsdb.CommitWithID
	if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		commitWithID, err = pfsdb.GetCommitWithIDByKey(ctx, tx, branchInfoWithID.Head)
		return errors.Wrap(err, "picking commit")
	}); err != nil {
		return nil, err
	}
	return commitWithID, nil
}

func (p *picker) pickCommitAncestorOf(ctx context.Context, ancestorOf *pfs.CommitPicker_AncestorOf) (*pfsdb.CommitWithID, error) {
	startCommit, err := p.PickCommit(ctx, ancestorOf.Start)
	if err != nil {
		return nil, errors.Wrap(err, "picking commit")
	}
	var commitWithID *pfsdb.CommitWithID
	if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		ancestry, err := pfsdb.GetCommitAncestry(ctx, tx, startCommit.ID, uint(ancestorOf.Offset))
		if err != nil {
			return errors.Wrap(err, "picking commit")
		}
		commitInfo, err := pfsdb.GetCommit(ctx, tx, ancestry.EarliestDiscovered)
		if err != nil {
			return errors.Wrap(err, "picking commit")
		}
		commitWithID = &pfsdb.CommitWithID{
			ID:         ancestry.EarliestDiscovered,
			CommitInfo: commitInfo,
		}
		return errors.Wrap(err, "picking commit")
	}); err != nil {
		return nil, err
	}
	return commitWithID, nil
}

func (p *picker) pickCommitBranchRoot(ctx context.Context, branchRoot *pfs.CommitPicker_BranchRoot) (*pfsdb.CommitWithID, error) {
	branchInfoWithID, err := p.PickBranch(ctx, branchRoot.GetBranch())
	if err != nil {
		return nil, errors.Wrap(err, "picking commit")
	}
	var commitWithID *pfsdb.CommitWithID
	if err := dbutil.WithTx(ctx, p.db, func(cbCtx context.Context, tx *pachsql.Tx) error {
		headCommitId, err := pfsdb.GetCommitID(ctx, tx, branchInfoWithID.Head)
		if err != nil {
			return errors.Wrap(err, "picking commit")
		}
		ancestry, err := pfsdb.GetCommitAncestry(ctx, tx, headCommitId, 0)
		if err != nil {
			return errors.Wrap(err, "picking commit")
		}
		commitPtr := ancestry.EarliestDiscovered
		for i := branchRoot.Offset; i > 0; i-- {
			commitPtr = ancestry.Lineage[commitPtr]
		}
		commitInfo, err := pfsdb.GetCommit(ctx, tx, commitPtr)
		if err != nil {
			return errors.Wrap(err, "picking commit")
		}
		commitWithID = &pfsdb.CommitWithID{
			ID:         commitPtr,
			CommitInfo: commitInfo,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return commitWithID, nil
}
