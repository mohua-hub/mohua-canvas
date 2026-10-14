package ports

import (
	"context"
	"io"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)

type ProjectStore interface {
	ListProjects(context.Context) ([]domain.Project, error)
	OpenProject(context.Context, string) (domain.Project, error)
	CreateProject(context.Context, domain.Project) error
	SaveProject(context.Context, domain.Project, int64) error
	DeleteProject(context.Context, string) error
}
type ChannelStore interface {
	ListChannels(context.Context) ([]domain.Channel, error)
	GetChannel(context.Context, string) (domain.Channel, error)
	SaveChannel(context.Context, domain.Channel) error
	DeleteChannel(context.Context, string) error
}
type CredentialVault interface {
	PutCredential(context.Context, string, string) error
	GetCredential(context.Context, string) (string, error)
	DeleteCredential(context.Context, string) error
}
type AssetStore interface {
	ListAssets(context.Context) ([]domain.Asset, error)
	GetAsset(context.Context, string) (domain.Asset, error)
	SaveAsset(context.Context, domain.Asset) error
	DeleteAsset(context.Context, string) error
}
type MediaStore interface {
	Put(context.Context, string, []byte) error
	Open(context.Context, string) (io.ReadSeekCloser, error)
	Remove(context.Context, string) error
}
type GeneratedAssets interface {
	ImportGenerated(context.Context, string, []byte) (domain.Asset, error)
	Delete(context.Context, string) error
}
type JobStore interface {
	GetJobByRequest(context.Context, string) (domain.Job, error)
	CreateIdempotent(context.Context, domain.Job, domain.Channel, bool) (domain.Job, bool, error)
	GetJob(context.Context, string) (domain.Job, error)
	ListJobs(context.Context, string) ([]domain.Job, error)
	SetJobStatus(context.Context, domain.Job, []string) (bool, error)
	InterruptPending(context.Context) error
}
type LibraryStore interface {
	ListLibrary(context.Context) ([]domain.LibraryEntry, error)
	SaveLibraryEntry(context.Context, domain.LibraryEntry) (domain.LibraryEntry, error)
	DeleteLibraryEntry(context.Context, string) error
}
type ModelProvider interface {
	ID() string
	Capabilities() []domain.Kind
	Generate(context.Context, domain.GenerationRequest, domain.Channel) (domain.GenerationResult, error)
}
