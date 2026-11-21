// Package all registers every built-in plugin. Only blackpie's main imports
// it: core never imports a plugin, and a plugin never imports another.
package all

import (
	_ "github.com/hmjz100/blackpie/plugins/budget"
	_ "github.com/hmjz100/blackpie/plugins/gateway"
	_ "github.com/hmjz100/blackpie/plugins/groups"
	_ "github.com/hmjz100/blackpie/plugins/library"
	_ "github.com/hmjz100/blackpie/plugins/middleware"
	_ "github.com/hmjz100/blackpie/plugins/otel"
	_ "github.com/hmjz100/blackpie/plugins/redact"
	_ "github.com/hmjz100/blackpie/plugins/rotation"
	_ "github.com/hmjz100/blackpie/plugins/routing"
	_ "github.com/hmjz100/blackpie/plugins/search"
	_ "github.com/hmjz100/blackpie/plugins/sessions"
	_ "github.com/hmjz100/blackpie/plugins/tools"
	_ "github.com/hmjz100/blackpie/plugins/usage"
	_ "github.com/hmjz100/blackpie/plugins/warmups"
)
