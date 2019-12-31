package other

import "github.com/jackie-feng/tools/internal/lsp/rename/crosspkg"

func Other() {
	crosspkg.Bar
	crosspkg.Foo()
}
