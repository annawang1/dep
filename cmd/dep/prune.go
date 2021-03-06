// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"log"

	"github.com/golang/dep"
	"github.com/golang/dep/internal/gps"
	"github.com/golang/dep/internal/gps/pkgtree"
	"github.com/pkg/errors"
)

const pruneShortHelp = `Prune the vendor tree of unused packages`
const pruneLongHelp = `
Prune is used to remove unused packages from your vendor tree.

STABILITY NOTICE: this command creates problems for vendor/ verification. As
such, it may be removed and/or moved out into a separate project later on.
`

type pruneCommand struct {
}

func (cmd *pruneCommand) Name() string      { return "prune" }
func (cmd *pruneCommand) Args() string      { return "" }
func (cmd *pruneCommand) ShortHelp() string { return pruneShortHelp }
func (cmd *pruneCommand) LongHelp() string  { return pruneLongHelp }
func (cmd *pruneCommand) Hidden() bool      { return false }

func (cmd *pruneCommand) Register(fs *flag.FlagSet) {
}

func (cmd *pruneCommand) Run(ctx *dep.Ctx, args []string) error {
	p, err := ctx.LoadProject()
	if err != nil {
		return err
	}

	sm, err := ctx.SourceManager()
	if err != nil {
		return err
	}
	sm.UseDefaultSignalHandling()
	defer sm.Release()

	// While the network churns on ListVersions() requests, statically analyze
	// code from the current project.
	ptree, err := pkgtree.ListPackages(p.AbsRoot, string(p.ImportRoot))
	if err != nil {
		return errors.Wrap(err, "analysis of local packages failed: %v")
	}

	// Set up a solver in order to check the InputHash.
	params := p.MakeParams()
	params.RootPackageTree = ptree

	if ctx.Verbose {
		params.TraceLogger = ctx.Err
	}

	s, err := gps.Prepare(params, sm)
	if err != nil {
		return errors.Wrap(err, "could not set up solver for input hashing")
	}

	if p.Lock == nil {
		return errors.Errorf("Gopkg.lock must exist for prune to know what files are safe to remove.")
	}

	if !bytes.Equal(s.HashInputs(), p.Lock.SolveMeta.InputsDigest) {
		return errors.Errorf("Gopkg.lock is out of sync; run dep ensure before pruning.")
	}

	var pruneLogger *log.Logger
	if ctx.Verbose {
		pruneLogger = ctx.Err
	}
	return dep.PruneProject(p, sm, pruneLogger)
}
