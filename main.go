package main

import (
	"github.com/sc-js/backend_core/src/bundles/hardwarebundle"
	"github.com/sc-js/backend_core/src/bundles/initbundle"
	"github.com/sc-js/backend_core/src/bundles/websocketbundle"
)

func main() {
	initbundle.InitializeCoreWithBundles([]initbundle.Bundle{
		{Handler: hardwarebundle.InitBundle, Settings: nil},
		{Handler: websocketbundle.InitBundle, Settings: map[string]string{"permission": websocketbundle.PERM_ADMIN}},
	}, nil)
	initbundle.RunTLS(nil, true)
}
