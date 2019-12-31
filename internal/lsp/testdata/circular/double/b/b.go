package b

import (
	_ "github.com/jackie-feng/tools/internal/lsp/circular/double/one" //@diag("_ \"github.com/jackie-feng/tools/internal/lsp/circular/double/one\"", "go list", "import cycle not allowed")
)
