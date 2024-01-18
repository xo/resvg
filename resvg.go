// Package resvg is a wrapper around rust's [resvg] c-api crate.
//
// [resvg]: https://github.com/RazrFalcon/resvg
package resvg

/*
#cgo CFLAGS: -I${SRCDIR}/libresvg
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/libresvg/darwin_amd64 -lresvg -lm
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/libresvg/darwin_arm64 -lresvg -lm
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libresvg/linux_amd64 -lresvg -lm
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/libresvg/linux_arm64 -lresvg -lm
#cgo linux,arm LDFLAGS: -L${SRCDIR}/libresvg/linux_arm -lresvg -lm
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/libresvg/windows_amd64 -lresvg -lm -lkernel32 -ladvapi32 -lbcrypt -lntdll -luserenv -lws2_32

#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include "resvg.h"

char* version() {
	char* s = malloc(sizeof(RESVG_VERSION));
	strncpy(s, RESVG_VERSION, sizeof(RESVG_VERSION));
	return s;
}

resvg_render_tree* parse(_GoBytes_ data, resvg_options* opts) {
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
	"strings"
	"syscall"
	"unsafe"
)

// Render renders svg data as a RGBA image.
func Render(data []byte, opts ...Option) (*image.RGBA, error) {
	return New(opts...).Render(data)
}

// Version returns the resvg version.
func Version() string {
	v := C.version()
	ver := C.GoString(v)
	C.free(unsafe.Pointer(v))
	return ver
}

// Resvg wraps the [resvg c-api].
//
// [resvg c-api]: https://github.com/RazrFalcon/resvg
type Resvg struct {
	LoadSystemFonts bool
	ResourcesDir    string
	DPI             float32
	FontFamily      string
	FontSize        float32
	SerifFamily     string
	SansSerifFamily string
	CursiveFamily   string
	FantasyFamily   string
	MonospaceFamily string
	Languages       []string
	ShapeRendering  ShapeRendering
	TextRendering   TextRendering
	ImageRendering  ImageRendering
	Fonts           [][]byte
	FontFiles       []string
}

// New creates a new resvg.
func New(opts ...Option) *Resvg {
	r := &Resvg{
		LoadSystemFonts: true,
		ShapeRendering:  ShapeRenderingNotSet,
		TextRendering:   TextRenderingNotSet,
		ImageRendering:  ImageRenderingNotSet,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// buildOpts builds the resvg options.
func (r *Resvg) buildOpts() *C.resvg_options {
	opts := C.resvg_options_create()
	if r.LoadSystemFonts {
		C.resvg_options_load_system_fonts(opts)
	}
	if r.ResourcesDir != "" {
		s := C.CString(r.ResourcesDir)
		C.resvg_options_set_resources_dir(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.DPI != 0.0 {
		C.resvg_options_set_dpi(opts, C.float(r.DPI))
	}
	if r.FontFamily != "" {
		s := C.CString(r.FontFamily)
		C.resvg_options_set_font_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.FontSize != 0.0 {
		C.resvg_options_set_font_size(opts, C.float(r.FontSize))
	}
	if r.SerifFamily != "" {
		s := C.CString(r.SerifFamily)
		C.resvg_options_set_serif_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.SansSerifFamily != "" {
		s := C.CString(r.SansSerifFamily)
		C.resvg_options_set_sans_serif_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.CursiveFamily != "" {
		s := C.CString(r.CursiveFamily)
		C.resvg_options_set_cursive_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.FantasyFamily != "" {
		s := C.CString(r.FantasyFamily)
		C.resvg_options_set_fantasy_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.MonospaceFamily != "" {
		s := C.CString(r.MonospaceFamily)
		C.resvg_options_set_monospace_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if len(r.Languages) != 0 {
		s := C.CString(strings.Join(r.Languages, ","))
		C.resvg_options_set_languages(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.ShapeRendering != ShapeRenderingNotSet {
		C.resvg_options_set_shape_rendering_mode(opts, C.resvg_shape_rendering(r.ShapeRendering))
	}
	if r.TextRendering != TextRenderingNotSet {
		C.resvg_options_set_text_rendering_mode(opts, C.resvg_text_rendering(r.TextRendering))
	}
	if r.ImageRendering != ImageRenderingNotSet {
		C.resvg_options_set_image_rendering_mode(opts, C.resvg_image_rendering(r.ImageRendering))
	}
	for _, font := range r.Fonts {
		s := C.CString(string(font))
		C.resvg_options_load_font_data(opts, s, C.uintptr_t(len(font)))
		C.free(unsafe.Pointer(s))
	}
	for _, fontFile := range r.FontFiles {
		if fontFile != "" {
			s := C.CString(fontFile)
			C.resvg_options_load_font_file(opts, s)
			C.free(unsafe.Pointer(s))
		}
	}
	return opts
}

// Render renders svg data as a RGBA image.
func (r *Resvg) Render(data []byte) (*image.RGBA, error) {
	tree, errno := C.parse(data, r.buildOpts())
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

// ShapeRendering is the shape rendering mode.
type ShapeRendering int

// Shape rendering modes.
const (
	ShapeRenderingOptimizeSpeed      ShapeRendering = C.RESVG_SHAPE_RENDERING_OPTIMIZE_SPEED
	ShapeRenderingCrispEdges         ShapeRendering = C.RESVG_SHAPE_RENDERING_CRISP_EDGES
	ShapeRenderingGeometricPrecision ShapeRendering = C.RESVG_SHAPE_RENDERING_GEOMETRIC_PRECISION
	ShapeRenderingNotSet             ShapeRendering = 0xff
)

// TextRendering is the text rendering mode.
type TextRendering int

// Text rendering modes.
const (
	TextRenderingOptimizeSpeed      TextRendering = C.RESVG_TEXT_RENDERING_OPTIMIZE_SPEED
	TextRenderingOptimizeLegibility TextRendering = C.RESVG_TEXT_RENDERING_OPTIMIZE_LEGIBILITY
	TextRenderingGeometricPrecision TextRendering = C.RESVG_TEXT_RENDERING_GEOMETRIC_PRECISION
	TextRenderingNotSet             TextRendering = 0xff
)

// ImageRendering is the image rendering mode.
type ImageRendering int

// Image rendering modes.
const (
	ImageRenderingOptimizeQuality ImageRendering = C.RESVG_IMAGE_RENDERING_OPTIMIZE_QUALITY
	ImageRenderingOptimizeSpeed   ImageRendering = C.RESVG_IMAGE_RENDERING_OPTIMIZE_SPEED
	ImageRenderingNotSet          ImageRendering = 0xff
)

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

// Option is a resvg rendering option.
type Option func(*Resvg)

// WithLoadSystemFonts is a resvg option to load system fonts.
func WithLoadSystemFonts(loadSystemFonts bool) Option {
	return func(r *Resvg) {
		r.LoadSystemFonts = loadSystemFonts
	}
}

// WithResourcesDir is a resvg option to set the resources dir.
func WithResourcesDir(resourcesDir string) Option {
	return func(r *Resvg) {
		r.ResourcesDir = resourcesDir
	}
}

// WithDPI is a resvg option to set the DPI.
func WithDPI(dpi float32) Option {
	return func(r *Resvg) {
		r.DPI = dpi
	}
}

// WithFontFamily is a resvg option to set the font family.
func WithFontFamily(fontFamily string) Option {
	return func(r *Resvg) {
		r.FontFamily = fontFamily
	}
}

// WithFontSize is a resvg option to set the font size.
func WithFontSize(fontSize float32) Option {
	return func(r *Resvg) {
		r.FontSize = fontSize
	}
}

// WithSerifFamily is a resvg option to set the serif family.
func WithSerifFamily(serifFamily string) Option {
	return func(r *Resvg) {
		r.SerifFamily = serifFamily
	}
}

// WithCursiveFamily is a resvg option to set the cursive family.
func WithCursiveFamily(cursiveFamily string) Option {
	return func(r *Resvg) {
		r.CursiveFamily = cursiveFamily
	}
}

// WithFantasyFamily is a resvg option to set the fantasy family.
func WithFantasyFamily(fantasyFamily string) Option {
	return func(r *Resvg) {
		r.FantasyFamily = fantasyFamily
	}
}

// WithMonospaceFamily is a resvg option to set the monospace family.
func WithMonospaceFamily(monospaceFamily string) Option {
	return func(r *Resvg) {
		r.MonospaceFamily = monospaceFamily
	}
}

// WithLanguages is a resvg option to set the languages.
func WithLanguages(languages ...string) Option {
	return func(r *Resvg) {
		r.Languages = languages
	}
}

// WithShapeRendering is a resvg option to set the shape rendering mode.
func WithShapeRendering(shapeRendering ShapeRendering) Option {
	return func(r *Resvg) {
		r.ShapeRendering = shapeRendering
	}
}

// WithTextRendering is a resvg option to set the text rendering mode.
func WithTextRendering(textRendering TextRendering) Option {
	return func(r *Resvg) {
		r.TextRendering = textRendering
	}
}

// WithImageRendering is a resvg option to set the image rendering mode.
func WithImageRendering(imageRendering ImageRendering) Option {
	return func(r *Resvg) {
		r.ImageRendering = imageRendering
	}
}

// WithFonts is a resvg option to set font data.
func WithFonts(fonts ...[]byte) Option {
	return func(r *Resvg) {
		r.Fonts = fonts
	}
}

// WithFontFiles is a resvg option to set font files.
func WithFontFiles(fontFiles ...string) Option {
	return func(r *Resvg) {
		r.FontFiles = fontFiles
	}
}
