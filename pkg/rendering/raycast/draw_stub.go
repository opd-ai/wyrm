//go:build noebiten

package raycast

// DrawStub is a placeholder type for headless builds without Ebitengine.
// This file allows the package to compile and tests to run without X11.
type DrawStub struct{}

// Note: In headless mode, the Draw method is not available.
// This stub file exists solely to allow tests of core.go functions
// (NewRenderer, castRay, getWallColor, SetPlayerPos) to run without
// requiring an X11 display.
