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
	width           uint
	height          uint
	scaleMode       ScaleMode
	transform       []float32
	opts            *C.resvg_options
	once            sync.Once
}

// New creates a new resvg.
func New(opts ...Option) *Resvg {
	r := &Resvg{
		loadSystemFonts: true,
		shapeRendering:  shapeRenderingNotSet,
		textRendering:   textRenderingNotSet,
		imageRendering:  imageRenderingNotSet,
		background:      color.Transparent,
	}
	for _, o := range opts {
		o(r)
	}
	runtime.SetFinalizer(r, (*Resvg).finalize)
	return r
}

// ParseConfig parses the svg, returning an image config.
func (r *Resvg) ParseConfig(data []byte) (image.Config, error) {
	tree, width, height, _, _, err := r.parse(data)
	if err != nil {
		return image.Config{}, err
	}
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
	tree, width, height, scaleX, scaleY, err := r.parse(data)
	if err != nil {
		return nil, err
	}
	// build transform
	ts := C.resvg_transform_identity()
	if r.transform == nil {
		ts.a, ts.d = C.float(scaleX), C.float(scaleY)
	} else {
		ts.a = C.float(r.transform[0])
		ts.b = C.float(r.transform[1])
		ts.c = C.float(r.transform[2])
		ts.d = C.float(r.transform[3])
		ts.e = C.float(r.transform[4])
		ts.f = C.float(r.transform[5])
	}
	// background
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	if c := color.RGBAModel.Convert(r.background).(color.RGBA); c.R != 0 || c.G != 0 || c.B != 0 || c.A != 0 {
		for i := range width {
			for j := range height {
				img.SetRGBA(i, j, c)
			}
		}
	}
	// render
	C.render(tree, C.int(width), C.int(height), ts, img.Pix)
	// destroy
	C.resvg_tree_destroy(tree)
	return img, nil
}

// parse parses the svg data, returning the width, height, and scaling factors.
func (r *Resvg) parse(data []byte) (*C.resvg_render_tree, int, int, float32, float32, error) {
	r.once.Do(r.buildOpts)
	if r.opts == nil {
		return nil, 0, 0, 0.0, 0.0, ErrOptionsNotInitialized
	}
	// parse
	tree, err := C.parse(data, r.opts)
	if err != nil {
		return nil, 0, 0, 0.0, 0.0, newErrNo(err)
	}
	// dimensions
	size := C.resvg_get_image_size(tree)
	if size.width == 0 || size.height == 0 {
		return nil, 0, 0, 0.0, 0.0, ErrInvalidWidthOrHeight
	}
	// determine height, width, scaleX, scaleY
	width, height, scaleX, scaleY := r.scaleMode.Scale(uint(size.width), uint(size.height), r.width, r.height)
	switch {
	case width == 0:
		return nil, 0, 0, 0.0, 0.0, ErrInvalidWidth
	case height == 0:
		return nil, 0, 0, 0.0, 0.0, ErrInvalidHeight
	case scaleX == 0.0:
		return nil, 0, 0, 0.0, 0.0, ErrInvalidXScale
	case scaleY == 0.0:
		return nil, 0, 0, 0.0, 0.0, ErrInvalidYScale
	}
	return tree, width, height, scaleX, scaleY, nil
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
	if r.shapeRendering != shapeRenderingNotSet {
		C.resvg_options_set_shape_rendering_mode(opts, C.resvg_shape_rendering(r.shapeRendering))
	}
	if r.textRendering != textRenderingNotSet {
		C.resvg_options_set_text_rendering_mode(opts, C.resvg_text_rendering(r.textRendering))
	}
	if r.imageRendering != imageRenderingNotSet {
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

// ShapeRendering is the shape rendering mode.
type ShapeRendering int

// Shape rendering modes.
const (
	ShapeRenderingOptimizeSpeed      ShapeRendering = C.RESVG_SHAPE_RENDERING_OPTIMIZE_SPEED
	ShapeRenderingCrispEdges         ShapeRendering = C.RESVG_SHAPE_RENDERING_CRISP_EDGES
	ShapeRenderingGeometricPrecision ShapeRendering = C.RESVG_SHAPE_RENDERING_GEOMETRIC_PRECISION
	shapeRenderingNotSet             ShapeRendering = 0xff
)

// TextRendering is the text rendering mode.
type TextRendering int

// Text rendering modes.
const (
	TextRenderingOptimizeSpeed      TextRendering = C.RESVG_TEXT_RENDERING_OPTIMIZE_SPEED
	TextRenderingOptimizeLegibility TextRendering = C.RESVG_TEXT_RENDERING_OPTIMIZE_LEGIBILITY
	TextRenderingGeometricPrecision TextRendering = C.RESVG_TEXT_RENDERING_GEOMETRIC_PRECISION
	textRenderingNotSet             TextRendering = 0xff
)

// ImageRendering is the image rendering mode.
type ImageRendering int

// Image rendering modes.
const (
	ImageRenderingOptimizeQuality ImageRendering = C.RESVG_IMAGE_RENDERING_OPTIMIZE_QUALITY
	ImageRenderingOptimizeSpeed   ImageRendering = C.RESVG_IMAGE_RENDERING_OPTIMIZE_SPEED
	imageRenderingNotSet          ImageRendering = 0xff
)

// ScaleMode is a scale mode.
type ScaleMode uint8

// Scale modes.
const (
	ScaleNone ScaleMode = iota
	ScaleMinWidth
	ScaleMinHeight
	ScaleMaxWidth
	ScaleMaxHeight
	ScaleBestFit
)

// Scale calculates the scale for the width, height.
func (mode ScaleMode) Scale(width, height, w, h uint) (int, int, float32, float32) {
	switch mode {
	case ScaleMinWidth:
		return scaleWidth(width, height, w, h, width < w)
	case ScaleMinHeight:
		return scaleHeight(width, height, w, h, height < h)
	case ScaleMaxWidth:
		return scaleWidth(width, height, w, h, width > w)
	case ScaleMaxHeight:
		return scaleHeight(width, height, w, h, height > h)
	case ScaleBestFit:
		return scaleBestFit(width, height, w, h)
	}
	scaleX, scaleY := float32(1.0), float32(1.0)
	if w != 0 {
		width, scaleX = w, float32(w)/float32(width)
	}
	if h != 0 {
		height, scaleY = h, float32(h)/float32(height)
	}
	return int(width), int(height), scaleX, scaleY
}

// scaleWidth calculates the scale for a width.
func scaleWidth(width, height, w, _ uint, scale bool) (int, int, float32, float32) {
	if scale {
		scaleX := float32(w) / float32(width)
		return int(w), int(math.Round(float64(float32(height) * scaleX))), scaleX, scaleX
	}
	return int(width), int(height), 1.0, 1.0
}

// scaleHeight calculates the scale for a height.
func scaleHeight(width, height, _, h uint, scale bool) (int, int, float32, float32) {
	if scale {
		scaleY := float32(h) / float32(height)
		return int(math.Round(float64(float32(width) * scaleY))), int(h), scaleY, scaleY
	}
	return int(width), int(height), 1.0, 1.0
}

// scaleBestFit calculates the best fit scale for the width, height.
func scaleBestFit(width, height, w, h uint) (int, int, float32, float32) {
	var scale float32
	switch {
	case w == 0:
		scale = float32(h) / float32(height)
	case h == 0:
		scale = float32(w) / float32(width)
	default:
		scale = min(float32(w)/float32(width), float32(h)/float32(height))
	}
	return int(math.Round(float64(float32(width) * scale))), int(math.Round(float64(float32(height) * scale))), scale, scale
}

// Error is a package error.
type Error string

// Errors.
const (
	ErrOptionsNotInitialized Error = "options not initialized"
	ErrInvalidWidthOrHeight  Error = "invalid width or height"
	ErrInvalidWidth          Error = "invalid width"
	ErrInvalidHeight         Error = "invalid height"
	ErrInvalidXScale         Error = "invalid x scale"
	ErrInvalidYScale         Error = "invalid y scale"
)

// Error satisfies the [error] interface.
func (err Error) Error() string {
	return string(err)
}

// ErrNo wraps a resvg error.
type ErrNo int

// newErrNo creates a new error.
func newErrNo(err error) error {
	if err == nil {
		return nil
	}
	if se, ok := err.(syscall.Errno); ok {
		return ErrNo(int(se))
	}
	panic(fmt.Sprintf("invalid error type %T", err))
}

// Error satisfies the [error] interface.
func (err ErrNo) Error() string {
	switch err {
	case C.RESVG_OK:
		return "resvg: ok"
	case C.RESVG_ERROR_NOT_AN_UTF8_STR:
		return "resvg: not a utf8 string"
	case C.RESVG_ERROR_FILE_OPEN_FAILED:
		return "resvg: file open failed"
	case C.RESVG_ERROR_MALFORMED_GZIP:
		return "resvg: malformed gzip data"
	case C.RESVG_ERROR_ELEMENTS_LIMIT_REACHED:
		return "resvg: element limit reached"
	case C.RESVG_ERROR_INVALID_SIZE:
		return "resvg: invalid size"
	case C.RESVG_ERROR_PARSING_FAILED:
		return "resvg: parsing failed"
	}
	return fmt.Sprintf("resvg: unknown error %d", int(err))
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
		r.width = uint(width)
	}
}

// WithHeight is a resvg option to set the height.
func WithHeight(height int) Option {
	return func(r *Resvg) {
		r.height = uint(height)
	}
}

// WithScaleMode is a resvg option to set scale mode.
func WithScaleMode(scaleMode ScaleMode) Option {
	return func(r *Resvg) {
		r.scaleMode = scaleMode
	}
}

// WithTransform is a resvg option to set the transform used.
func WithTransform(a, b, c, d, e, f float32) Option {
	return func(r *Resvg) {
		r.transform = []float32{a, b, c, d, e, f}
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
