/*
Copyright 2022 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	resolutioncommon "github.com/tektoncd/resolution/pkg/common"
	"github.com/tektoncd/resolution/pkg/resolver/framework"
)

// LabelValueGitResolverType is the value to use for the
// resolution.tekton.dev/type label on resource requests
const LabelValueGitResolverType string = "git"

// GitResolverName is the name that the git resolver should be
// associated with
const GitResolverName string = "Git"

// YAMLContentType is the content type to use when returning yaml
const YAMLContentType string = "application/x-yaml"

var _ framework.Resolver = &Resolver{}

// Resolver implements a framework.Resolver that can fetch files from git.
type Resolver struct{}

// Initialize performs any setup required by the gitresolver.
func (r *Resolver) Initialize(ctx context.Context) error {
	return nil
}

// GetName returns the string name that the gitresolver should be
// associated with.
func (r *Resolver) GetName(_ context.Context) string {
	return GitResolverName
}

// GetSelector returns the labels that resource requests are required to have for
// the gitresolver to process them.
func (r *Resolver) GetSelector(_ context.Context) map[string]string {
	return map[string]string{
		resolutioncommon.LabelKeyResolverType: LabelValueGitResolverType,
	}
}

// ValidateParams returns an error if the given parameter map is not
// valid for a resource request targeting the gitresolver.
func (r *Resolver) ValidateParams(_ context.Context, params map[string]string) error {
	required := []string{
		URLParam,
		PathParam,
	}
	missing := []string{}
	if params == nil {
		missing = required
	} else {
		for _, p := range required {
			v, has := params[p]
			if !has || v == "" {
				missing = append(missing, p)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing %v", strings.Join(missing, ", "))
	}

	if params[CommitParam] != "" && params[BranchParam] != "" {
		return fmt.Errorf("supplied both %q and %q", CommitParam, BranchParam)
	}

	// TODO(sbwsg): validate repo url is well-formed, git:// or https://
	// TODO(sbwsg): validate path is valid relative path

	return nil
}

// Resolve performs the work of fetching a file from git given a map of
// parameters.
func (r *Resolver) Resolve(_ context.Context, params map[string]string) (framework.ResolvedResource, error) {
	repo := params[URLParam]
	commit := params[CommitParam]
	branch := params[BranchParam]
	path := params[PathParam]
	cloneOpts := &git.CloneOptions{
		URL: repo,
	}
	filesystem := memfs.New()
	if branch != "" {
		cloneOpts.SingleBranch = true
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}
	repository, err := git.Clone(memory.NewStorage(), filesystem, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("clone error: %w", err)
	}
	if commit == "" {
		headRef, err := repository.Head()
		if err != nil {
			return nil, fmt.Errorf("error reading repository HEAD value: %w", err)
		}
		commit = headRef.Hash().String()
	}

	w, err := repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree error: %v", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return nil, fmt.Errorf("checkout error: %v", err)
	}

	f, err := filesystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %v", path, err)
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, fmt.Errorf("error reading file %q: %v", path, err)
	}

	return &ResolvedGitResource{
		Commit:  commit,
		Content: buf.Bytes(),
	}, nil
}

// ResolvedGitResource implements framework.ResolvedResource and returns
// the resolved file []byte data and an annotation map for any metadata.
type ResolvedGitResource struct {
	Commit  string
	Content []byte
}

var _ framework.ResolvedResource = &ResolvedGitResource{}

// Data returns the bytes of the file resolved from git.
func (r *ResolvedGitResource) Data() []byte {
	return r.Content
}

// Annotations returns the metadata that accompanies the file fetched
// from git.
func (r *ResolvedGitResource) Annotations() map[string]string {
	return map[string]string{
		AnnotationKeyCommitHash:                   r.Commit,
		resolutioncommon.AnnotationKeyContentType: YAMLContentType,
	}
}
