// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"github.com/jackie-feng/tools/internal/lsp/protocol"
	"github.com/jackie-feng/tools/internal/lsp/source"
	"github.com/jackie-feng/tools/internal/span"
)

func (s *Server) implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	uri := span.NewURI(params.TextDocument.URI)
	view, err := s.session.ViewOf(uri)
	if err != nil {
		return nil, err
	}
	snapshot := view.Snapshot()
	fh, err := snapshot.GetFile(ctx, uri)
	if err != nil {
		return nil, err
	}
	if fh.Identity().Kind != source.Go {
		return nil, nil
	}
	return source.Implementation(ctx, snapshot, fh, params.Position)
}
