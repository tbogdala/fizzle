// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package ui

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	renderer "github.com/tbogdala/fizzle/renderer"
)

// These are constants used in UILayout Anchor positions.
const (
	_ = iota
	UIAnchorTopLeft
	UIAnchorTopMiddle
	UIAnchorTopRight
	UIAnchorMiddle
	UIAnchorMiddleLeft
	UIAnchorMiddleRight
	UIAnchorBottomLeft
	UIAnchorBottomMiddle
	UIAnchorBottomRight
)

// These are the min and max Z depth used for the UI ortho projection.
const (
	minZDepth = -200.0
	maxZDepth = 200.0
)

// UILayout determines how the UI widget will get positioned by
// specifying a corner or mid-point to anchor to and an offset to
// that anchor.
type UILayout struct {
	// The offset in pixels from the Anchor point. Can be positive or negative.
	Offset mgl.Vec3

	// The anchor point for the layout, which should be one of the UIANchor constants
	// like UIAnchorTopLeft.
	Anchor int
}

// UILabel is a text label widget in the user interface.
type UILabel struct {
	// Text is the text string that was used to create the renderable.
	Text string

	// Layout specifies the way the label should be positioned on screen.
	Layout UILayout

	// Renderable is the drawable object
	Renderable *fizzle.Renderable

	// manager is a pointer back to the owner UIManager.
	manager *UIManager
}

// UIImage is an image widget which shows a texture in the user interface.
type UIImage struct {
	// Texture is the OpenGL texture id to use to draw this widget
	Texture graphics.Texture

	// Layout specifies the way the label should be positioned on screen.
	Layout UILayout

	// Renderable is the drawable object
	Renderable *fizzle.Renderable

	// manager is a pointer back to the owner UIManager.
	manager *UIManager
}

// UIWidget is a common interface for all of the user interface widgets
// that represents common functionality.
type UIWidget interface {
	// Destroy should release any OpenGL or other data specific to widget.
	Destroy()

	// Draw should render the widget to screen.
	Draw(renderer renderer.Renderer, binder renderer.RenderBinder, projection mgl.Mat4, view mgl.Mat4, camera fizzle.Camera)

	// GetLayout should return the UILayout for the widget that's used for positioning.
	GetLayout() *UILayout

	// GetRenderable should return the drawable Renderable object.
	GetRenderable() *fizzle.Renderable
}

// UIManager is the primary owner for all of the widgets created by it and
// has methods to layout the widgets on screen and draw them.
type UIManager struct {
	// width is used to construct the ortho projection matrix and is probably
	// best set to the width of the window.
	width int32

	// height is used to construct the ortho projection matrix and is probably
	// best set to the height of the window.
	height int32

	// widget is a slice that contains all of the widgets to be rendered
	// to the screen by the UIManager on Draw().
	widgets []UIWidget
}

// NewUIManager creates a new UIManager object.
func NewUIManager() *UIManager {
	uim := new(UIManager)
	return uim
}

// Destroy removes all widgets from the UIManager and Destroy()s them all
// one by one.
func (ui *UIManager) Destroy() {
	for _, w := range ui.widgets {
		w.Destroy()
	}

	// make an empty slice
	ui.widgets = []UIWidget{}
}

// Add puts the widget into the internal slice of widgets to use in the
// user interface.
func (ui *UIManager) Add(w UIWidget) {
	ui.widgets = append(ui.widgets, w)
}

// AdviseResolution will change the resolution the UIManager uses to draw
// widgets and will also adjust the layouts of any exiting wigets.
func (ui *UIManager) AdviseResolution(w int32, h int32) {
	ui.width = w
	ui.height = h
	ui.LayoutWidgets()
}

// LayoutWidgets repositions the widgets according to the anchor and offset.
// This can be useful if the size of the root window changes.
// Note: currently only the root window is supported, but this could be
// changed in the future
func (ui *UIManager) LayoutWidgets() {
	for _, w := range ui.widgets {
		layout := w.GetLayout()
		renderable := w.GetRenderable()

		// for now use the root window as the reference point
		var minX, minY, maxX, maxY float32
		minX = 0.0
		minY = 0.0
		maxX = float32(ui.width)
		maxY = float32(ui.height)

		switch layout.Anchor {
		case UIAnchorTopLeft:
			renderable.Location[0] = minX + layout.Offset[0]
			renderable.Location[1] = maxY - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorTopMiddle:
			renderable.Location[0] = maxX/2.0 - renderable.BoundingRect.DeltaX()/2.0 + layout.Offset[0]
			renderable.Location[1] = maxY - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorTopRight:
			renderable.Location[0] = maxX - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = maxY - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorMiddle:
			renderable.Location[0] = maxX/2.0 - renderable.BoundingRect.DeltaX()/2.0 + layout.Offset[0]
			renderable.Location[1] = maxY/2.0 - renderable.BoundingRect.DeltaY()/2.0 + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorMiddleLeft:
			renderable.Location[0] = minX + layout.Offset[0]
			renderable.Location[1] = maxY/2.0 - renderable.BoundingRect.DeltaY()/2.0 + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorMiddleRight:
			renderable.Location[0] = maxX - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = maxY/2.0 - renderable.BoundingRect.DeltaY()/2.0 + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorBottomLeft:
			renderable.Location[0] = minX + layout.Offset[0]
			renderable.Location[1] = minY + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorBottomMiddle:
			renderable.Location[0] = maxX/2.0 - renderable.BoundingRect.DeltaX()/2.0 + layout.Offset[0]
			renderable.Location[1] = minY + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchorBottomRight:
			renderable.Location[0] = maxX - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = minY + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		}
	}
}

// Draw renders all of widgets on the screen.
func (ui *UIManager) Draw(renderer renderer.Renderer, binder renderer.RenderBinder) {
	// calculate the perspective and view
	ortho := mgl.Ortho(0, float32(ui.width), 0, float32(ui.height), minZDepth, maxZDepth)
	view := mgl.Ident4()

	// draw all of the widgets
	for _, w := range ui.widgets {
		w.Draw(renderer, binder, ortho, view, nil)
	}
}

// RemoveWidget removes the supplied widget from the internal collection
// of widgets so that it's no longer modified on layout changes or drawn to
// screen.
func (ui *UIManager) RemoveWidget(widgetToRemove UIWidget) {
	for i, w := range ui.widgets {
		if w == widgetToRemove {
			ui.widgets = append(ui.widgets[:i], ui.widgets[i+1:]...)
			return
		}
	}
}

// -------------------------------------------------------------------------
// LABEL WIDGET

// CreateLabel creates the label widget and the text renderable.
func (ui *UIManager) CreateLabel(font *fizzle.GLFont, anchor int, offset mgl.Vec3, msg string) *UILabel {
	label := new(UILabel)
	label.Text = msg
	label.Layout.Anchor = anchor
	label.Layout.Offset = offset
	label.Renderable = font.CreateLabel(msg)
	label.manager = ui

	// add it to the widget list before returning
	ui.widgets = append(ui.widgets, label)

	return label
}

// Destroy releases the OpenGL data specific to the widget
// but DOES NOT remove the widget from the parent UIManager.
func (l *UILabel) Destroy() {
	l.Renderable.Destroy()
}

// Draw renders the widget onto the screen. Layout should have already
// modified the positioning of the renderable.
func (l *UILabel) Draw(renderer renderer.Renderer, binder renderer.RenderBinder, projection mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	renderer.DrawRenderable(l.Renderable, binder, projection, view, camera)
}

// GetLayout returns a pointer to the layout object of the widget.
func (l *UILabel) GetLayout() *UILayout {
	return &l.Layout
}

// GetRenderable returns the Renderable object for the label widget
func (l *UILabel) GetRenderable() *fizzle.Renderable {
	return l.Renderable
}

// -------------------------------------------------------------------------
// IMAGE WIDGET

// CreateImage creates the image widget and renderable.
func (ui *UIManager) CreateImage(anchor int, offset mgl.Vec3, texId graphics.Texture, width float32, height float32, shader *fizzle.RenderShader) *UIImage {
	img := new(UIImage)
	img.Texture = texId
	img.Layout.Anchor = anchor
	img.Layout.Offset = offset
	img.Renderable = fizzle.CreatePlaneXY("color_textured", 0, 0, width, height)
	img.Renderable.Core.Shader = shader
	img.Renderable.Core.Tex0 = texId
	img.Renderable.Core.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	img.manager = ui

	// add it to the widget list before returning
	ui.widgets = append(ui.widgets, img)

	return img
}

// Destroy releases the OpenGL data specific to the widget
// but DOES NOT remove the widget from the parent UIManager.
func (img *UIImage) Destroy() {
	img.Renderable.Destroy()
}

// Draw renders the widget onto the screen. Layout should have already
// modified the positioning of the renderable.
func (img *UIImage) Draw(renderer renderer.Renderer, binder renderer.RenderBinder, projection mgl.Mat4, view mgl.Mat4, camera fizzle.Camera) {
	renderer.DrawRenderable(img.Renderable, binder, projection, view, camera)
}

// GetLayout returns a pointer to the layout object of the widget.
func (img *UIImage) GetLayout() *UILayout {
	return &img.Layout
}

// GetRenderable returns the Renderable object for the label widget
func (img *UIImage) GetRenderable() *fizzle.Renderable {
	return img.Renderable
}
