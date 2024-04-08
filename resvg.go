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
	if (errno != 0) {
		return 0;
	}
	return tree;
}

void render(resvg_render_tree* tree, int width, int height, resvg_transform ts, _GoBytes_ buf) {
	resvg_render(tree, ts, width, height, buf.p);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// Resvg wraps the [resvg c-api] to render svgs as standard a [image.RGBA].
//
// [resvg c-api]: https://github.com/RazrFalcon/resvg
type Resvg struct {
	loadSystemFonts bool
	resourcesDir    string
	dp              float32
	fontFamily      string
	fontSize        float32
	serifFamily     string
	sansSerifFamily string
	cursiveFamily   string
	fantasyFamily   string
	monospaceFamily string
	languages       []string
	shapeRendering  ShapeRendering
	textRendering   TextRendering
	imageRendering  ImageRendering
	fonts           [][]byte
	fontFiles       []string
	background      color.Color
	width           int
	height          int
	bestFit         bool
	transform       []float64
	opts            *C.resvg_options
	once            sync.Once
}

// New creates a new resvg.
func New(opts ...Option) *Resvg {
	r := &Resvg{
		loadSystemFonts: true,
		shapeRendering:  ShapeRenderingNotSet,
		textRendering:   TextRenderingNotSet,
		imageRendering:  ImageRenderingNotSet,
		background:      color.Transparent,
	}
	for _, o := range opts {
		o(r)
	}
	runtime.SetFinalizer(r, (*Resvg).finalize)
	return r
}

// buildOpts builds the resvg options.
func (r *Resvg) buildOpts() {
	opts := C.resvg_options_create()
	if r.loadSystemFonts {
		C.resvg_options_load_system_fonts(opts)
	}
	if r.resourcesDir != "" {
		s := C.CString(r.resourcesDir)
		C.resvg_options_set_resources_dir(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.dp != 0.0 {
		C.resvg_options_set_dpi(opts, C.float(r.dp))
	}
	if r.fontFamily != "" {
		s := C.CString(r.fontFamily)
		C.resvg_options_set_font_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.fontSize != 0.0 {
		C.resvg_options_set_font_size(opts, C.float(r.fontSize))
	}
	if r.serifFamily != "" {
		s := C.CString(r.serifFamily)
		C.resvg_options_set_serif_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.sansSerifFamily != "" {
		s := C.CString(r.sansSerifFamily)
		C.resvg_options_set_sans_serif_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.cursiveFamily != "" {
		s := C.CString(r.cursiveFamily)
		C.resvg_options_set_cursive_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.fantasyFamily != "" {
		s := C.CString(r.fantasyFamily)
		C.resvg_options_set_fantasy_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.monospaceFamily != "" {
		s := C.CString(r.monospaceFamily)
		C.resvg_options_set_monospace_family(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if len(r.languages) != 0 {
		s := C.CString(strings.Join(r.languages, ","))
		C.resvg_options_set_languages(opts, s)
		C.free(unsafe.Pointer(s))
	}
	if r.shapeRendering != ShapeRenderingNotSet {
		C.resvg_options_set_shape_rendering_mode(opts, C.resvg_shape_rendering(r.shapeRendering))
	}
	if r.textRendering != TextRenderingNotSet {
		C.resvg_options_set_text_rendering_mode(opts, C.resvg_text_rendering(r.textRendering))
	}
	if r.imageRendering != ImageRenderingNotSet {
		C.resvg_options_set_image_rendering_mode(opts, C.resvg_image_rendering(r.imageRendering))
	}
	for _, font := range r.fonts {
		s := C.CString(string(font))
		C.resvg_options_load_font_data(opts, s, C.uintptr_t(len(font)))
		C.free(unsafe.Pointer(s))
	}
	for _, fontFile := range r.fontFiles {
		if fontFile != "" {
			s := C.CString(fontFile)
			C.resvg_options_load_font_file(opts, s)
			C.free(unsafe.Pointer(s))
		}
	}
	r.opts = opts
}

// finalize finalizes the C allocations.
func (r *Resvg) finalize() {
	if r.opts != nil {
		C.resvg_options_destroy(r.opts)
	}
	r.opts = nil
	runtime.SetFinalizer(r, nil)
}

// ParseConfig parses the svg, returning an image config.
func (r *Resvg) ParseConfig(data []byte) (image.Config, error) {
	r.once.Do(r.buildOpts)
	if r.opts == nil {
		return image.Config{}, errors.New("options not initialized")
	}
	tree, errno := C.parse(data, r.opts)
	if errno != nil {
		return image.Config{}, NewParseError(errno)
	}
	// height/width
	size := C.resvg_get_image_size(tree)
	width, height := int(size.width), int(size.height)
	// destroy
	C.resvg_tree_destroy(tree)
	return image.Config{
		ColorModel: color.RGBAModel,
		Width:      width,
		Height:     height,
	}, nil
}

// Render renders svg data as a RGBA image.
func (r *Resvg) Render(data []byte) (*image.RGBA, error) {
	r.once.Do(r.buildOpts)
	if r.opts == nil {
		return nil, errors.New("options not initialized")
	}
	tree, errno := C.parse(data, r.opts)
	if errno != nil {
		return nil, NewParseError(errno)
	}
	// determine height, width, scaleX, scaleY
	size := C.resvg_get_image_size(tree)
	if size.width == 0 || size.height == 0 {
		return nil, errors.New("invalid width or height")
	}
	width, height, scaleX, scaleY := r.calc(int(size.width), int(size.height))
	switch {
	case width == 0:
		return nil, errors.New("invalid width")
	case height == 0:
		return nil, errors.New("invalid height")
	case scaleX == 0.0:
		return nil, errors.New("invalid x scale")
	case scaleY == 0.0:
		return nil, errors.New("invalid y scale")
	}
	// build transform
	ts := C.resvg_transform_identity()
	if r.transform != nil {
		ts.a = C.float(r.transform[0])
		ts.b = C.float(r.transform[1])
		ts.c = C.float(r.transform[2])
		ts.d = C.float(r.transform[3])
		ts.e = C.float(r.transform[4])
		ts.f = C.float(r.transform[5])
	} else {
		ts.a, ts.d = C.float(scaleX), C.float(scaleY)
	}
	c := color.RGBAModel.Convert(r.background).(color.RGBA)
	// build out
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			img.SetRGBA(i, j, c)
		}
	}
	// render
	C.render(tree, C.int(width), C.int(height), ts, img.Pix)
	// destroy
	C.resvg_tree_destroy(tree)
	return img, nil
}

// calc determines the width/height and scales in the x/y direction to use.
func (r *Resvg) calc(width, height int) (int, int, float32, float32) {
	hasWidth, hasHeight := r.width != 0, r.height != 0
	switch {
	case hasWidth && hasHeight:
		return r.width, r.height, float32(r.width) / float32(width), float32(r.height) / float32(height)
	case !r.bestFit && hasWidth:
		return r.width, height, float32(r.width) / float32(width), 1.0
	case !r.bestFit && hasHeight:
		return width, r.height, 1.0, float32(r.height) / float32(height)
	case !r.bestFit:
		return width, height, 1.0, 1.0
	}
	if hasWidth {
		scaleX := float32(r.width) / float32(width)
		return r.width, int(math.Round(float64(float32(height) * scaleX))), scaleX, scaleX
	}
	scaleY := float32(r.height) / float32(height)
	return int(math.Round(float64(float32(width) * scaleY))), r.height, scaleY, scaleY
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
		r.loadSystemFonts = loadSystemFonts
	}
}

// WithResourcesDir is a resvg option to set the resources dir.
func WithResourcesDir(resourcesDir string) Option {
	return func(r *Resvg) {
		r.resourcesDir = resourcesDir
	}
}

// WithDPI is a resvg option to set the DPI.
func WithDPI(dpi float32) Option {
	return func(r *Resvg) {
		r.dp = dpi
	}
}

// WithFontFamily is a resvg option to set the font family.
func WithFontFamily(fontFamily string) Option {
	return func(r *Resvg) {
		r.fontFamily = fontFamily
	}
}

// WithFontSize is a resvg option to set the font size.
func WithFontSize(fontSize float32) Option {
	return func(r *Resvg) {
		r.fontSize = fontSize
	}
}

// WithSerifFamily is a resvg option to set the serif family.
func WithSerifFamily(serifFamily string) Option {
	return func(r *Resvg) {
		r.serifFamily = serifFamily
	}
}

// WithCursiveFamily is a resvg option to set the cursive family.
func WithCursiveFamily(cursiveFamily string) Option {
	return func(r *Resvg) {
		r.cursiveFamily = cursiveFamily
	}
}

// WithFantasyFamily is a resvg option to set the fantasy family.
func WithFantasyFamily(fantasyFamily string) Option {
	return func(r *Resvg) {
		r.fantasyFamily = fantasyFamily
	}
}

// WithMonospaceFamily is a resvg option to set the monospace family.
func WithMonospaceFamily(monospaceFamily string) Option {
	return func(r *Resvg) {
		r.monospaceFamily = monospaceFamily
	}
}

// WithLanguages is a resvg option to set the languages.
func WithLanguages(languages ...string) Option {
	return func(r *Resvg) {
		r.languages = languages
	}
}

// WithShapeRendering is a resvg option to set the shape rendering mode.
func WithShapeRendering(shapeRendering ShapeRendering) Option {
	return func(r *Resvg) {
		r.shapeRendering = shapeRendering
	}
}

// WithTextRendering is a resvg option to set the text rendering mode.
func WithTextRendering(textRendering TextRendering) Option {
	return func(r *Resvg) {
		r.textRendering = textRendering
	}
}

// WithImageRendering is a resvg option to set the image rendering mode.
func WithImageRendering(imageRendering ImageRendering) Option {
	return func(r *Resvg) {
		r.imageRendering = imageRendering
	}
}

// WithFonts is a resvg option to set font data.
func WithFonts(fonts ...[]byte) Option {
	return func(r *Resvg) {
		r.fonts = fonts
	}
}

// WithFontFiles is a resvg option to set font files.
func WithFontFiles(fontFiles ...string) Option {
	return func(r *Resvg) {
		r.fontFiles = fontFiles
	}
}

// WithBackground is a resvg option to set the fill background color.
func WithBackground(background color.Color) Option {
	return func(r *Resvg) {
		r.background = background
	}
}

// WithWidth is a resvg option to set the width.
func WithWidth(width int) Option {
	return func(r *Resvg) {
		r.width = width
	}
}

// WithHeight is a resvg option to set the height.
func WithHeight(height int) Option {
	return func(r *Resvg) {
		r.height = height
	}
}

// WithBestFit is a resvg option to set best fit.
func WithBestFit(bestFit bool) Option {
	return func(r *Resvg) {
		r.bestFit = bestFit
	}
}

// WithTransform is a resvg option to set the transform used.
func WithTransform(a, b, c, d, e, f float64) Option {
	return func(r *Resvg) {
		r.transform = []float64{a, b, c, d, e, f}
	}
}

// Default is the default renderer.
var Default = New()

func init() {
	image.RegisterFormat("svg", `<?xml`, Decode, DecodeConfig)
	image.RegisterFormat("svg", "<svg", Decode, DecodeConfig)
}

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

// Decode decodes a svg from the reader.
func Decode(r io.Reader) (image.Image, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	img, err := Default.Render(buf)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// DecodeConfig decodes a svg config from the reader.
func DecodeConfig(r io.Reader) (image.Config, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return image.Config{}, err
	}
	return Default.ParseConfig(buf)
}
