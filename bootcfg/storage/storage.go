package storage

import (
	"errors"

	"github.com/mikeynap/coreos-baremetal/bootcfg/storage/storagepb"
)

// Storage errors
var (
	ErrGroupNotFound   = errors.New("storage: No Group found")
	ErrProfileNotFound = errors.New("storage: No Profile found")
)

// A Store stores machine Groups, Profiles, and Configs.
type Store interface {
	// GroupPut creates or updates a Group.
	GroupPut(group *storagepb.Group) error
	// GroupGet returns a machine Group by id.
	GroupGet(id string) (*storagepb.Group, error)
	// GroupList lists all machine Groups.
	GroupList() ([]*storagepb.Group, error)

	// ProfilePut creates or updates a Profile.
	ProfilePut(profile *storagepb.Profile) error
	// ProfileGet gets a profile by id.
	ProfileGet(id string) (*storagepb.Profile, error)
	// ProfileList lists all profiles.
	ProfileList() ([]*storagepb.Profile, error)

	// IgnitionPut creates or updates an Ignition template.
	IgnitionPut(name string, config []byte) error
	// IgnitionGet gets an Ignition template by name.
	IgnitionGet(name string) (string, error)

	// CloudPut creates or updates an Cloud-Config template.
	CloudPut(name string, config []byte) error
	// CloudGet gets a Cloud-Config template by name.
	CloudGet(name string) (string, error)

	// GenericGet gets a generic template by name.
	GenericGet(name string) (string, error)
}
