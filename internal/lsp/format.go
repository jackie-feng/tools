// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"github.com/jackie-feng/tools/internal/lsp/protocol"
	"github.com/jackie-feng/tools/internal/lsp/source"
	"github.com/jackie-feng/tools/internal/span"
)

func (s *Server) formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
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
	var edits []protocol.TextEdit
	switch fh.Identity().Kind {
	case source.Go:
		edits, err = source.Format(ctx, snapshot, fh)
	case source.Mod:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return edits, nil
}
