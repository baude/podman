package libartifact

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
)

type ArtifactStore struct {
	storePath string
	store     storage.Store
	runtime   *libimage.Runtime
}

var (
	// annotatedName is the label name where the artifact tag reference lives
	annotatedName string = "org.opencontainers.image.ref.name"
)

func NewArtifactStore(storePath string) (*ArtifactStore, error) {
	sp := storage.StoreOptions{ImageStore: storePath}
	store, err := storage.GetStore(sp)
	if err != nil {
		return nil, err
	}
	rt, err := libimage.RuntimeFromStore(store, nil)
	if err != nil {
		return nil, err
	}
	artifactStore := &ArtifactStore{
		runtime:   rt,
		store:     store,
		storePath: storePath,
	}
	return artifactStore, nil
}

func (a *ArtifactStore) Remove(ctx context.Context, sys types.SystemContext, name string) error {
	ir, err := layout.NewReference(a.storePath, name)
	if err != nil {
		return err
	}
	return ir.DeleteImage(ctx, &sys)
}

func (a *ArtifactStore) Inspect(name string) (*layout.ListResult, error) {
	lrs, err := layout.List(a.storePath)
	if err != nil {
		return nil, err
	}
	for _, l := range lrs {
		if val, ok := l.ManifestDescriptor.Annotations[annotatedName]; !ok || val != name {
			continue
		}
		return &l, nil
	}
	return nil, fmt.Errorf("no artifact found with name %s", name)
}

func (a *ArtifactStore) List() ([]layout.ListResult, error) {
	// note: this does not work on non-existent store nor a store
	// without a json file.
	return layout.List(a.storePath)
}

func (a *ArtifactStore) Pull(ctx context.Context, name string) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", name))
	if err != nil {
		return err
	}
	destRef, err := layout.NewReference(a.storePath, name)
	if err != nil {
		return err
	}
	copyOpts := libimage.CopyOptions{
		Writer: os.Stdout,
	}
	copyer, err := a.runtime.NewCopier(&copyOpts)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}

func (a ArtifactStore) Push(ctx context.Context, sys types.SystemContext, name string) error {
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", name))
	if err != nil {
		return err
	}
	srcRef, err := layout.NewReference(a.storePath, name)
	if err != nil {
		return err
	}
	copyOpts := libimage.CopyOptions{
		Writer: os.Stdout,
	}
	copyer, err := a.runtime.NewCopier(&copyOpts)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}
