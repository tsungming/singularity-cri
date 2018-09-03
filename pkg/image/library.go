package image

import (
	"path/filepath"
	"strings"

	library "github.com/singularityware/singularity/src/pkg/library/client"
	"github.com/sylabs/cri/pkg/singularity"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

type libraryImageInfo struct {
	ref        string
	user       string
	collection string
	container  string
	tags       []string
}

// parseLibraryRef parses provided reference to an image and
// fetches all possible information from it. Reference must be in form
// [library://][repo/collection/]|[collection/]container[:tag]
func parseLibraryRef(ref string) libraryImageInfo {
	ref = strings.TrimPrefix(ref, "library://")
	refParts := strings.Split(ref, "/")

	info := libraryImageInfo{
		ref:       "library://" + ref,
		tags:      []string{"latest"},
		container: refParts[len(refParts)-1],
	}

	switch len(refParts) {
	case 3:
		info.user = refParts[0]
		info.collection = refParts[1]
	case 2:
		info.collection = refParts[0]
	}

	imageParts := strings.Split(info.container, ":")
	if len(imageParts) != 1 {
		info.container = imageParts[0]
		info.tags = strings.Split(imageParts[1], ",")
	}

	return info
}

func (i libraryImageInfo) Remote() string {
	return i.ref
}

func (i libraryImageInfo) Id() string {
	var parts []string
	if i.user != "" {
		parts = append(parts, i.user)
	}
	if i.collection != "" {
		parts = append(parts, i.collection)
	}
	parts = append(parts, i.container)
	return strings.Join(parts, "_") + ".sif"
}

func (i libraryImageInfo) Tags() []string {
	return i.tags
}

func (i libraryImageInfo) Digests() []string {
	return nil
}

func (i libraryImageInfo) Pull(auth *v1alpha2.AuthConfig, dir string) error {
	path := filepath.Join(dir, i.Id())
	server := singularity.LibraryURL
	if auth != nil && auth.ServerAddress != "" {
		server = auth.ServerAddress
	}
	return library.DownloadImage(path, i.Remote(), server, true, "")
}
