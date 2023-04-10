package main

import (
	"github.com/sc-js/core_backend/src/bundles/hardwarebundle"
	"github.com/sc-js/core_backend/src/bundles/initbundle"
	"github.com/sc-js/core_backend/src/bundles/websocketbundle"
)

func main() {
	initbundle.InitializeCoreWithBundles([]initbundle.Bundle{
		{Handler: hardwarebundle.InitBundle, Settings: nil},
		{Handler: websocketbundle.InitBundle, Settings: map[string]string{"permission": websocketbundle.PERM_ADMIN}},
	}, nil)
	initbundle.RunTLS(nil, true)
}
