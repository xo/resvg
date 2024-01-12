// Package resvg is a wrapper around rust's resvg c-api crate.
package resvg

/*
#cgo LDFLAGS: -Llibresvg -lresvg -lm

#include <errno.h>

#include "libresvg/resvg.h"

resvg_render_tree* parse(_GoBytes_ data) {
	// create options
	resvg_options* opts = resvg_options_create();
	resvg_options_load_system_fonts(opts);
	// parse
	resvg_render_tree* tree;
	errno = resvg_parse_tree_from_data(data.p, data.n, opts, &tree);
	resvg_options_destroy(opts);
	if (errno != 0) {
		return 0;
	}
	return tree;
}

void render(resvg_render_tree* tree, int width, int height, _GoBytes_ buf) {
	resvg_render(tree, resvg_transform_identity(), width, height, buf.p);
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"syscall"
)

// Render renders svg data as a RGBA image.
func Render(data []byte) (*image.RGBA, error) {
	tree, errno := C.parse(data)
	if errno != nil {
		return nil, NewParseError(errno)
	}
	// height/width
	size := C.resvg_get_image_size(tree)
	width, height := int(size.width), int(size.height)
	// create
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			img.Set(i, j, color.Transparent)
		}
	}
	// render
	C.render(tree, C.int(width), C.int(height), img.Pix)
	// destroy
	C.resvg_tree_destroy(tree)
	return img, nil
}

// ParseError is a parse error.
type ParseError int

// NewParseError creates a new error.
func NewParseError(e error) error {
	if e == nil {
		return nil
	}
	if se, ok := e.(syscall.Errno); ok {
		return ParseError(int(se))
	}
	panic(fmt.Sprintf("invalid error type: %T", e))
}

// Error satisfies the [error] interface.
func (err ParseError) Error() string {
	switch err {
	case C.RESVG_OK:
		return "OK"
	case C.RESVG_ERROR_NOT_AN_UTF8_STR:
		return "only UTF-8 content are supported"
	case C.RESVG_ERROR_FILE_OPEN_FAILED:
		return "failed to open the provided file"
	case C.RESVG_ERROR_MALFORMED_GZIP:
		return "compressed SVG must use the GZip algorithm"
	case C.RESVG_ERROR_ELEMENTS_LIMIT_REACHED:
		return "we do not allow SVG with more than 1_000_000 elements for security reasons"
	case C.RESVG_ERROR_INVALID_SIZE:
		return "SVG doesn't have a valid size"
	case C.RESVG_ERROR_PARSING_FAILED:
		return "failed to parse SVG data"
	}
	return ""
}
