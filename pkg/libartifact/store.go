package libartifact

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/common/libimage"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/opencontainers/go-digest"
	specV1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// indexName is the name of the JSON file in root of the artifact store
	// that describes the store's contents
	indexName = "index.json"
)

type ArtifactStore struct {
	SystemContext *types.SystemContext
	storePath     string
}

// NewArtifactStore is a constructor for artifact stores.  Most artifact dealings depend on this. Store path is
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

	// if the storage dir does not exist, we need to create it.
	baseDir := filepath.Dir(artifactStore.indexPath())
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}
	// if the index file is not present we need to create an empty one
	_, err := os.Stat(artifactStore.indexPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if createErr := artifactStore.createEmptyManifest(); createErr != nil {
				return nil, createErr
			}
		}
	}

	return artifactStore, nil
}

func (as ArtifactStore) Remove(ctx context.Context, sys *types.SystemContext, name string) error {
	ir, err := layout.NewReference(as.storePath, name)
	if err != nil {
		return err
	}
	return ir.DeleteImage(ctx, sys)
}

func (as ArtifactStore) Inspect(ctx context.Context, name string) (*Artifact, error) {
	artifacts, err := as.getArtifacts(ctx, nil)
	if err != nil {
		return nil, err
	}
	return artifacts.getByName(name)
}

func (as ArtifactStore) List(ctx context.Context) (ArtifactList, error) {
	return as.getArtifacts(ctx, nil)
}

func (as ArtifactStore) Pull(ctx context.Context, name string) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", name))
	if err != nil {
		return err
	}
	destRef, err := layout.NewReference(as.storePath, name)
	if err != nil {
		return err
	}
	copyOpts := libimage.CopyOptions{
		Writer: os.Stdout,
	}

	copyer, err := libimage.NewCopier(&copyOpts, as.SystemContext, nil)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}

func (as ArtifactStore) Push(ctx context.Context, src, dest string) error {
	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", dest))
	if err != nil {
		return err
	}
	srcRef, err := layout.NewReference(as.storePath, src)
	if err != nil {
		return err
	}
	copyOpts := libimage.CopyOptions{
		Writer: os.Stdout,
	}
	copyer, err := libimage.NewCopier(&copyOpts, as.SystemContext, nil)
	if err != nil {
		return err
	}
	_, err = copyer.Copy(ctx, srcRef, destRef)
	if err != nil {
		return err
	}
	return copyer.Close()
}

func (as ArtifactStore) Add(ctx context.Context, dest string, path, artifactType string) (*digest.Digest, error) {
	// TODO
	// 1. Check to make sure the artifact does not otherwise exist

	var (
		// TODO Needs to be located somewhere that makes sense
		MediaTypeOCIArtifact = "application/vnd.github.com.containers.artifact"
	)

	var annotations = map[string]string{}
	annotations[specV1.AnnotationTitle] = filepath.Base(path)

	ir, err := layout.NewReference(as.storePath, dest)
	if err != nil {
		return nil, err
	}

	// Initialize artifact if not present
	// sourcePath == dest
	//ociDest, err := openOrCreateSourceImage(ctx, sourcePath)
	//if err != nil {
	//	return err
	//}
	//defer ociDest.Close()

	// ociDest replaces imageDest
	imageDest, err := ir.NewImageDestination(ctx, as.SystemContext)
	if err != nil {
		return nil, err
	}

	// get the new artifact into the local store
	newBlobDigest, newBlobSize, err := layout.PutBlobFromLocalFile(ctx, imageDest, path)
	if err != nil {
		return nil, err
	}

	//https://github.com/containers/buildah/blob/main/internal/source/add.go#L46
	is, err := ir.NewImageSource(ctx, as.SystemContext)
	if err != nil {
		return nil, err
	}

	//ociSource, err := ociRef.NewImageSource(ctx, &types.SystemContext{})
	// initial digest is used to define whats need removed after we've written the updated
	// manifest
	artifactManifest, initialDigest, _, err := readManifestFromImageSource(ctx, is)
	if err != nil {
		return nil, err
	}

	artifactManifest.Layers = append(artifactManifest.Layers,
		specV1.Descriptor{
			MediaType:   MediaTypeOCIArtifact,
			Digest:      newBlobDigest,
			Size:        newBlobSize,
			Annotations: annotations,
		},
	)

	updatedDigest, updatedSize, err := writeManifest(ctx, artifactManifest, imageDest)
	if err != nil {
		return nil, err
	}

	// Remove of the old blobs JSON file by the initialDigest name
	// https://github.com/containers/buildah/blob/main/internal/source/source.go#L117-L121

	manifestDescriptor := specV1.Descriptor{
		MediaType: specV1.MediaTypeImageManifest, // TODO: the media type should be configurable
		Digest:    *updatedDigest,
		Size:      updatedSize,
	}

	// Update the artifact in index.json
	//// We need to update the index.json on local storage
	storeIndex, err := as.readIndex()
	if err != nil {
		return nil, err
	}

	for i, m := range storeIndex.Manifests {
		if &m.Digest == initialDigest {
			storeIndex.Manifests[i] = manifestDescriptor
			break
		}
	}

	if err := as.writeIndex(storeIndex); err != nil {
		return nil, err
	}
	return updatedDigest, nil
}

func (as ArtifactStore) readIndex() (specV1.Index, error) {
	index := specV1.Index{}
	rawData, err := os.ReadFile(as.indexPath())
	if err != nil {
		return specV1.Index{}, err
	}
	err = json.Unmarshal(rawData, &index)
	return index, err
}

func (as ArtifactStore) writeIndex(index specV1.Index) error {
	rawData, err := json.Marshal(&index)
	if err != nil {
		return err
	}
	return os.WriteFile(as.indexPath(), rawData, 0o644)
}

func (as ArtifactStore) createEmptyManifest() error {
	index := specV1.Index{}
	rawData, err := json.Marshal(&index)
	if err != nil {
		return err
	}

	return os.WriteFile(as.indexPath(), rawData, 0o644)
}

func (as ArtifactStore) indexPath() string {
	return filepath.Join(as.storePath, indexName)
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

// readManifestFromImageSource reads the manifest from the specified image
// source.  Note that the manifest is expected to be an OCI v1 manifest.
// Taken from buildah source.go
func readManifestFromImageSource(ctx context.Context, src types.ImageSource) (*specV1.Manifest, *digest.Digest, int64, error) {
	rawData, mimeType, err := src.GetManifest(ctx, nil)
	if err != nil {
		return nil, nil, -1, err
	}
	if mimeType != specV1.MediaTypeImageManifest {
		return nil, nil, -1, fmt.Errorf("image %q is of type %q (expected: %q)", strings.TrimPrefix(src.Reference().StringWithinTransport(), "//"), mimeType, specV1.MediaTypeImageManifest)
	}

	readManifest := specV1.Manifest{}
	if err := json.Unmarshal(rawData, &readManifest); err != nil {
		return nil, nil, -1, fmt.Errorf("reading manifest: %w", err)
	}

	manifestDigest := digest.FromBytes(rawData)
	return &readManifest, &manifestDigest, int64(len(rawData)), nil
}

// writeManifest writes the specified OCI `manifest` to the source image at
// `ociDest`.
// Taken from buildah source.go
func writeManifest(ctx context.Context, manifest *specV1.Manifest, ociDest types.ImageDestination) (*digest.Digest, int64, error) {
	rawData, err := json.Marshal(&manifest)
	if err != nil {
		return nil, -1, fmt.Errorf("marshalling manifest: %w", err)
	}

	if err := ociDest.PutManifest(ctx, rawData, nil); err != nil {
		return nil, -1, fmt.Errorf("writing manifest: %w", err)
	}

	manifestDigest := digest.FromBytes(rawData)
	return &manifestDigest, int64(len(rawData)), nil
}
