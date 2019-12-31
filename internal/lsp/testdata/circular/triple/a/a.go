package a

import (
	_ "github.com/jackie-feng/tools/internal/lsp/circular/triple/b" //@diag("_ \"github.com/jackie-feng/tools/internal/lsp/circular/triple/b\"", "go list", "import cycle not allowed")
)
