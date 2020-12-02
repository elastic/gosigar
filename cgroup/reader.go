package cgroup

import (
	"path/filepath"
)

// Stats contains metrics and limits from each of the cgroup subsystems.
type Stats struct {
	Metadata
	CPU           *CPUSubsystem           `json:"cpu"`
	CPUAccounting *CPUAccountingSubsystem `json:"cpuacct"`
	Memory        *MemorySubsystem        `json:"memory"`
	BlockIO       *BlockIOSubsystem       `json:"blkio"`
}

// Metadata contains metadata associated with cgroup stats.
type Metadata struct {
	ID   string `json:"id,omitempty"`   // ID of the cgroup.
	Path string `json:"path,omitempty"` // Path to the cgroup relative to the cgroup subsystem's mountpoint.
}

type mount struct {
	subsystem  string // Subsystem name (e.g. cpuacct).
	mountpoint string // Mountpoint of the subsystem (e.g. /cgroup/cpuacct).
	path       string // Relative path to the cgroup (e.g. /docker/<id>).
	id         string // ID of the cgroup.
	fullPath   string // Absolute path to the cgroup. It's the mountpoint joined with the path.
}

// Reader reads cgroup metrics and limits.
type Reader struct {
	// Mountpoint of the root filesystem. Defaults to / if not set. This can be
	// useful for example if you mount / as /rootfs inside of a container.
	rootfsMountpoint         string
	ignoreRootCgroups        bool // Ignore a cgroup when its path is "/".
	cgroupsHierarchyOverride string
	cgroupMountpoints        map[string]string // Mountpoints for each subsystem (e.g. cpu, cpuacct, memory, blkio).
}

// ReaderOptions holds options for NewReaderOptions.
type ReaderOptions struct {
	// RootfsMountpoint holds the mountpoint of the root filesystem.
	//
	// If unspecified, "/" is assumed.
	RootfsMountpoint string

	// IgnoreRootCgroups ignores cgroup subsystem with the path "/".
	IgnoreRootCgroups bool

	// CgroupsHierarchyOverride is an optional path override for cgroup
	// subsystem paths. If non-empty, this will be used instead of the
	// paths specified in /proc/<pid>/cgroup.
	//
	// This should be set to "/" when running within a Docker container,
	// where the paths in /proc/<pid>/cgroup do not correspond to any
	// paths under /sys/fs/cgroup.
	CgroupsHierarchyOverride string
}

// NewReader creates and returns a new Reader.
func NewReader(rootfsMountpoint string, ignoreRootCgroups bool) (*Reader, error) {
	return NewReaderOptions(ReaderOptions{
		RootfsMountpoint:  rootfsMountpoint,
		IgnoreRootCgroups: ignoreRootCgroups,
	})
}

// NewReaderOptions creates and returns a new Reader with the given options.
func NewReaderOptions(opts ReaderOptions) (*Reader, error) {
	if opts.RootfsMountpoint == "" {
		opts.RootfsMountpoint = "/"
	}

	// Determine what subsystems are supported by the kernel.
	subsystems, err := SupportedSubsystems(opts.RootfsMountpoint)
	if err != nil {
		return nil, err
	}

	// Locate the mountpoints of those subsystems.
	mountpoints, err := SubsystemMountpoints(opts.RootfsMountpoint, subsystems)
	if err != nil {
		return nil, err
	}

	return &Reader{
		rootfsMountpoint:         opts.RootfsMountpoint,
		ignoreRootCgroups:        opts.IgnoreRootCgroups,
		cgroupsHierarchyOverride: opts.CgroupsHierarchyOverride,
		cgroupMountpoints:        mountpoints,
	}, nil
}

// GetStatsForProcess returns cgroup metrics and limits associated with a process.
func (r *Reader) GetStatsForProcess(pid int) (*Stats, error) {
	// Read /proc/[pid]/cgroup to get the paths to the cgroup metrics.
	paths, err := ProcessCgroupPaths(r.rootfsMountpoint, pid)
	if err != nil {
		return nil, err
	}

	// Build the full path for the subsystems we are interested in.
	mounts := map[string]mount{}
	for _, interestedSubsystem := range []string{"blkio", "cpu", "cpuacct", "memory"} {
		path, found := paths[interestedSubsystem]
		if !found {
			continue
		}

		if path == "/" && r.ignoreRootCgroups {
			continue
		}

		subsystemMount, found := r.cgroupMountpoints[interestedSubsystem]
		if !found {
			continue
		}

		id := filepath.Base(path)
		if r.cgroupsHierarchyOverride != "" {
			path = r.cgroupsHierarchyOverride
		}
		mounts[interestedSubsystem] = mount{
			subsystem:  interestedSubsystem,
			mountpoint: subsystemMount,
			id:         id,
			path:       path,
			fullPath:   filepath.Join(subsystemMount, path),
		}
	}

	stats := Stats{Metadata: getCommonCgroupMetadata(mounts)}

	// Collect stats from each cgroup subsystem associated with the task.
	if mount, found := mounts["blkio"]; found {
		stats.BlockIO = &BlockIOSubsystem{}
		err := stats.BlockIO.get(mount.fullPath)
		if err != nil {
			return nil, err
		}
		stats.BlockIO.Metadata.ID = mount.id
		stats.BlockIO.Metadata.Path = mount.path
	}
	if mount, found := mounts["cpu"]; found {
		stats.CPU = &CPUSubsystem{}
		err := stats.CPU.get(mount.fullPath)
		if err != nil {
			return nil, err
		}
		stats.CPU.Metadata.ID = mount.id
		stats.CPU.Metadata.Path = mount.path
	}
	if mount, found := mounts["cpuacct"]; found {
		stats.CPUAccounting = &CPUAccountingSubsystem{}
		err := stats.CPUAccounting.get(mount.fullPath)
		if err != nil {
			return nil, err
		}
		stats.CPUAccounting.Metadata.ID = mount.id
		stats.CPUAccounting.Metadata.Path = mount.path
	}
	if mount, found := mounts["memory"]; found {
		stats.Memory = &MemorySubsystem{}
		err := stats.Memory.get(mount.fullPath)
		if err != nil {
			return nil, err
		}
		stats.Memory.Metadata.ID = mount.id
		stats.Memory.Metadata.Path = mount.path
	}

	// Return nil if no metrics were collected.
	if stats.BlockIO == nil && stats.CPU == nil && stats.CPUAccounting == nil && stats.Memory == nil {
		return nil, nil
	}

	return &stats, nil
}

// getCommonCgroupMetadata returns Metadata containing the cgroup path and ID
// iff all subsystems share a common path and ID. This is common for
// containerized processes. If there is no common path and ID then the returned
// values are empty strings.
func getCommonCgroupMetadata(mounts map[string]mount) Metadata {
	var path string
	for _, m := range mounts {
		if path == "" {
			path = m.path
		} else if path != m.path {
			// All paths are not the same.
			return Metadata{}
		}
	}

	return Metadata{Path: path, ID: filepath.Base(path)}
}
