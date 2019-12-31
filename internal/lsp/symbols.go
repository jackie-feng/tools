// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"github.com/jackie-feng/tools/internal/lsp/protocol"
	"github.com/jackie-feng/tools/internal/lsp/source"
	"github.com/jackie-feng/tools/internal/lsp/telemetry"
	"github.com/jackie-feng/tools/internal/span"
	"github.com/jackie-feng/tools/internal/telemetry/log"
	"github.com/jackie-feng/tools/internal/telemetry/trace"
)

func (s *Server) documentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]protocol.DocumentSymbol, error) {
	ctx, done := trace.StartSpan(ctx, "lsp.Server.documentSymbol")
	defer done()

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
	var symbols []protocol.DocumentSymbol
	switch fh.Identity().Kind {
	case source.Go:
		symbols, err = source.DocumentSymbols(ctx, snapshot, fh)
	case source.Mod:
		return []protocol.DocumentSymbol{}, nil
	}

	if err != nil {
		log.Error(ctx, "DocumentSymbols failed", err, telemetry.URI.Of(uri))
		return []protocol.DocumentSymbol{}, nil
	}
	return symbols, nil
}
