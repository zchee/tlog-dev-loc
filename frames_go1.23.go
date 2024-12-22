//go:build go1.23
// +build go1.23

package loc

import "runtime"

type (
	runtimeFrame = runtime.Frame

	runtimeFrames struct {
		ptr *PC
		len int
		buf PC // cap

		nextPC PC

		frames     []runtimeFrame
		frameStore [2]runtimeFrame
	}
)
