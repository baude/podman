package libartifact

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/opencontainers/go-digest"
)

type ArtifactStore struct {
	SystemContext *types.SystemContext
	storePath     string
}

type Artifact struct {
	List      layout.ListResult
	Manifests []manifest.OCI1
}

type GetArtifactOptions struct{}

type ArtifactList []*Artifact

var (
	// annotatedName is the label name where the artifact tag reference lives
	annotatedName = "org.opencontainers.image.ref.name"
)

// NewArtifactStore is a contructor for artifact stores.  Most artifact dealings depend on this. Store path is
// the filesystem location.
func NewArtifactStore(storePath string, sc *types.SystemContext) (*ArtifactStore, error) {
	// storePath here is an override
	if storePath == "" {
		storeOptions, err := storage.DefaultStoreOptions()
		if err != nil {
			return nil, err
		}
		storePath = filepath.Join(storeOptions.GraphRoot, "artifacts")
	}
	artifactStore := &ArtifactStore{
		storePath:     storePath,
		SystemContext: sc,
	}
	return artifactStore, nil
}

func (a ArtifactStore) Remove(ctx context.Context, sys *types.SystemContext, name string) error {
	ir, err := layout.NewReference(a.storePath, name)
	if err != nil {
		return err
	}
	return ir.DeleteImage(ctx, sys)
}

func (a ArtifactStore) Inspect(ctx context.Context, name string) (*Artifact, error) {
	artifacts, err := a.getArtifacts(ctx, nil)
	if err != nil {
		return nil, err
	}
	return artifacts.getByName(name)
}

func (a ArtifactStore) List(ctx context.Context) (ArtifactList, error) {
	return a.getArtifacts(ctx, nil)
}

func (a ArtifactStore) Pull(ctx context.Context, name string) error {
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

	copyer, err := libimage.NewCopier(&copyOpts, a.SystemContext)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}

func (a ArtifactStore) Push(ctx context.Context, src, dest string) error {
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", dest))
	if err != nil {
		return err
	}
	srcRef, err := layout.NewReference(a.storePath, src)
	if err != nil {
		return err
	}
	copyOpts := libimage.CopyOptions{
		Writer: os.Stdout,
	}
	copyer, err := libimage.NewCopier(&copyOpts, a.SystemContext)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}

func (a ArtifactStore) Add(ctx context.Context, src, dest string) error {
	if _, err := os.Stat(src); err != nil {
		// I don't think that handling file not found here specifically
		// and will return the error as is for the caller to handle
		return err
	}
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", dest))
	if err != nil {
		return err
	}
	_ = destRef
	// create a manifest with the correct store
	//list := manifests.Create()

	// add artifact to the manifest
	return err
}

// TotalSize returns the total bytes of the all the artifact layers
func (a Artifact) TotalSize() int64 {
	var s int64
	for _, artifact := range a.Manifests {
		for _, layer := range artifact.Layers {
			s += layer.Size
		}
	}
	return s
}

// getArtifacts returns an ArtifactList based on the artifact's store.  The return error and
// unused opts is meant for future growth like filters, etc so the API does not change.
func (as ArtifactStore) getArtifacts(ctx context.Context, _ *GetArtifactOptions) (ArtifactList, error) {
	var (
		al ArtifactList
	)
	lrs, err := layout.List(as.storePath)
	if err != nil {
		return nil, err
	}
	for _, l := range lrs {
		imgSrc, err := l.Reference.NewImageSource(ctx, as.SystemContext)
		if err != nil {
			return nil, err
		}
		manifests, err := getManifests(ctx, imgSrc, nil)
		if err != nil {
			return nil, err
		}

		artifact := Artifact{
			List:      l,
			Manifests: manifests,
		}
		al = append(al, &artifact)
	}
	return al, nil
}

// getByName returns an artifact, if present, by a given name
func (al ArtifactList) getByName(name string) (*Artifact, error) {
	for _, artifact := range al {
		if val, ok := artifact.List.ManifestDescriptor.Annotations[annotatedName]; ok && val == name {
			return artifact, nil
		}
	}
	return nil, fmt.Errorf("no artifact found with name %s", name)
}

// getManifests takes an imgSrc and starting digest (nil means "top") and collects all the manifests "under"
// it.  this func calls itself recursively with a new startingDigest assuming that we are dealing with
// and index list
func getManifests(ctx context.Context, imgSrc types.ImageSource, startingDigest *digest.Digest) ([]manifest.OCI1, error) {
	var (
		manifests []manifest.OCI1
	)
	b, manifestType, err := imgSrc.GetManifest(ctx, startingDigest)
	if err != nil {
		return nil, err
	}
	// this assumes that there are only single, and multi-images
	if !manifest.MIMETypeIsMultiImage(manifestType) {
		// these are the keepers
		mani, err := manifest.OCI1FromManifest(b)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, *mani)
		return manifests, nil
	}
	// We are dealing with an oci index list
	maniList, err := manifest.OCI1IndexFromManifest(b)
	if err != nil {
		return nil, err
	}
	for _, m := range maniList.Manifests {
		iterManifests, err := getManifests(ctx, imgSrc, &m.Digest)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, iterManifests...)
	}
	return manifests, nil
}
