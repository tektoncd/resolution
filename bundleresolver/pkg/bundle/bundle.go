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

package bundle

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// RequestOptions are the options used to request a resource from
// a remote bundle.
type RequestOptions struct {
	Namespace      string
	ServiceAccount string
	Bundle         string
	EntryName      string
	Kind           string
}

// ResolvedResource wraps the content of a matched entry in a bundle.
type ResolvedResource struct {
	data        []byte
	annotations map[string]string
}

// Data returns the bytes of the resource fetched from the bundle.
func (br *ResolvedResource) Data() []byte {
	return br.data
}

// Annotations returns the annotations from the bundle that are relevant
// to resolution.
func (br *ResolvedResource) Annotations() map[string]string {
	return br.annotations
}

// GetEntry accepts a keychain and options for the request and returns
// either a successfully resolved bundle entry or an error.
func GetEntry(ctx context.Context, keychain authn.Keychain, opts RequestOptions) (*ResolvedResource, error) {
	imgRef, err := name.ParseReference(opts.Bundle)
	if err != nil {
		return nil, fmt.Errorf("invalid bundle reference: %w", err)
	}

	image, err := remote.Image(imgRef, remote.WithAuthFromKeychain(keychain), remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error retrieving image: %w", err)
	}

	manifest, err := image.Manifest()
	if err != nil {
		return nil, fmt.Errorf("error parsing bundle manifest: %w", err)
	}

	for idx, manifestLayer := range manifest.Layers {
		layerKind, ok := manifestLayer.Annotations[BundleAnnotationKind]
		if !ok {
			return nil, fmt.Errorf("kind annotation not found in bundle layer %d", idx)
		}

		layerName, ok := manifestLayer.Annotations[BundleAnnotationName]
		if !ok {
			return nil, fmt.Errorf("name annotation not found in bundle layer %d", idx)
		}

		layerAPIVersion, ok := manifestLayer.Annotations[BundleAnnotationAPIVersion]
		if !ok {
			return nil, fmt.Errorf("apiVersion annotation not found in bundle layer %d", idx)
		}

		if layerKind == opts.Kind && layerName == opts.EntryName {
			manifestLayerDigest := manifestLayer.Digest.String()
			layers, err := image.Layers()
			if err != nil {
				return nil, fmt.Errorf("error reading layers: %w", err)
			}
			for _, imageLayer := range layers {
				digest, err := imageLayer.Digest()
				if err != nil {
					return nil, fmt.Errorf("error reading layer digest: %w", err)
				}
				if digest.String() == manifestLayerDigest {
					layerReader, err := imageLayer.Uncompressed()
					if err != nil {
						return nil, fmt.Errorf("error decompressing layer: %w", err)
					}
					defer layerReader.Close()
					tarReader := tar.NewReader(layerReader)
					header, err := tarReader.Next()
					if err != nil {
						return nil, fmt.Errorf("error reading tarball header: %w", err)
					}
					data := make([]byte, header.Size)
					if n, err := tarReader.Read(data); err != nil && err != io.EOF {
						return nil, fmt.Errorf("invalid tarball: %w", err)
					} else if int64(n) != header.Size {
						return nil, fmt.Errorf("layer data does not match size reported in header: expected %d received %d", header.Size, n)
					}
					resource := ResolvedResource{
						data: data,
						annotations: map[string]string{
							BundleAnnotationKind:       layerKind,
							BundleAnnotationName:       layerName,
							BundleAnnotationAPIVersion: layerAPIVersion,
						},
					}
					return &resource, nil
				}
			}
			// No image layer with matching digest, but may
			// be a raw image layer instead.
			if idx < len(layers) {
				imageLayer := layers[idx]
				layerReader, err := imageLayer.Uncompressed()
				if err != nil {
					return nil, fmt.Errorf("error decompressing layer: %w", err)
				}
				defer layerReader.Close()
				data, err := ioutil.ReadAll(layerReader)
				if err != nil {
					return nil, fmt.Errorf("error reading layer content: %w", err)
				}
				return &ResolvedResource{
					data: data,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("no matching image layer")
}
