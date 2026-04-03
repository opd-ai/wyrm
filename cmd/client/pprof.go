//go:build !noebiten

// pprof.go imports net/http/pprof to register profiling handlers.
// This import registers pprof handlers with http.DefaultServeMux, which is used
// by startProfileServer when profiling is enabled via config.Debug.ProfilingEnabled.

package main

import (
	_ "net/http/pprof"
)
