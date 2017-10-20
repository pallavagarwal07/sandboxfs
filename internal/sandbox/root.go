// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License.  You may obtain a copy
// of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations
// under the License.

package sandbox

// TODO(pallavag): Root data type is necessary since we want to reconfigure the
// node present at the root. However, FUSE API caches the value obtained from
// filesystem.Root() and does not allow us to change it.
// If the FUSE API allowed changing root and reserving, this file wouldn't be
// needed.

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// cacheInvalidator represents a node with kernel cache invalidation abilities.
type cacheInvalidator interface {
	fs.Node

	invalidate(*fs.Server)
}

// Dir defines the interfaces satisfied by all directory types.
type Dir interface {
	fs.Node
	fs.NodeCreater
	fs.NodeStringLookuper
	fs.NodeMkdirer
	fs.NodeMknoder
	fs.NodeOpener
	fs.NodeRemover
	fs.NodeRenamer
	fs.NodeSymlinker

	invalidateEntries(*fs.Server, fs.Node)
}

// Root is the container for all types that are valid to be at the root level
// of a filesystem tree.
type Root struct {
	Dir
}

// NewRoot returns a new instance of Root with the appropriate underlying node.
func NewRoot(node Dir) *Root {
	return &Root{node}
}

// Reconfigure resets the filesystem tree to the tree pointed to by root.
func (r *Root) Reconfigure(server *fs.Server, root Dir) {
	err := server.InvalidateNodeData(r)
	logCacheInvalidationError(err, "Could not invalidate root: ", r)

	// Invalidate the cache of the entries that are present before reconfiguration. This
	// essentially gets rid of entries that will be no longer available.
	r.Dir.invalidateEntries(server, r)

	r.Dir = root

	// Invalidate the cache of entries that were previously returning ENOENT.
	r.Dir.invalidateEntries(server, r)
}

// Rename delegates the Rename operation to the underlying node.
// NOTE: When renaming a file within the same directory, in root, we want
// newDir passed to be the underlying type and not the *Root type..
func (r *Root) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	if newDir == r {
		newDir = r.Dir
	}
	return r.Dir.Rename(ctx, req, newDir)
}
