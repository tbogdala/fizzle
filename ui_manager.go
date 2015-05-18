// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzle

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

// Define UILayout Anchor positions
const (
	_ = iota
	UIAnchor_TopLeft
	UIAnchor_TopMiddle
	UIAnchor_TopRight
	UIAnchor_MiddleLeft
	UIAnchor_MiddleRight
	UIAnchor_BottomLeft
	UIAnchor_BottomMiddle
	UIAnchor_BottomRight
)

const (
	minZDepth = -200.0
	maxZDepth = 200.0
)

// UILayout determines how the UI widget will get positioned
type UILayout struct {
	Offset mgl.Vec3
	Anchor int
}

// UILabel is a text label in the user interface
type UILabel struct {
	Text       string
	Layout     UILayout
	Renderable *Renderable

	manager *UIManager
}

type UIWidget interface {
	Destroy()
	Draw(perspective mgl.Mat4, view mgl.Mat4)
	GetLayout() *UILayout
	GetRenderable() *Renderable
}

type UIManager struct {
	renderer *DeferredRenderer
	width    int32
	height   int32
	widgets  []UIWidget
}

func NewUIManager(renderer *DeferredRenderer) *UIManager {
	uim := new(UIManager)
	uim.renderer = renderer
	return uim
}

func (ui *UIManager) Destroy() {
	for _, w := range ui.widgets {
		w.Destroy()
	}
}

func (ui *UIManager) AdviseResolution(w int32, h int32) {
	ui.width = w
	ui.height = h
	ui.LayoutWidgets()
}

// LayoutWidgets repositions the widgets according to the anchor and offset.
// This can be useful if the size of the root window changes.
func (ui *UIManager) LayoutWidgets() {
	for _, w := range ui.widgets {
		layout := w.GetLayout()
		renderable := w.GetRenderable()

		switch layout.Anchor {
		case UIAnchor_TopLeft:
			renderable.Location[0] = layout.Offset[0]
			renderable.Location[1] = float32(ui.height) - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_TopMiddle:
			renderable.Location[0] = float32(ui.width)/2.0 - renderable.BoundingRect.DeltaX()/2.0 + layout.Offset[0]
			renderable.Location[1] = float32(ui.height) - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_TopRight:
			renderable.Location[0] = float32(ui.width) - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = float32(ui.height) - renderable.BoundingRect.DeltaY() + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_MiddleLeft:
			renderable.Location[0] = layout.Offset[0]
			renderable.Location[1] = float32(ui.height)/2.0 + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_MiddleRight:
			renderable.Location[0] = float32(ui.width) - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = float32(ui.height)/2.0 + layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_BottomLeft:
			renderable.Location[0] = layout.Offset[0]
			renderable.Location[1] = layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_BottomMiddle:
			renderable.Location[0] = float32(ui.width)/2.0 - renderable.BoundingRect.DeltaX()/2.0 + layout.Offset[0]
			renderable.Location[1] = layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		case UIAnchor_BottomRight:
			renderable.Location[0] = float32(ui.width) - renderable.BoundingRect.DeltaX() + layout.Offset[0]
			renderable.Location[1] = layout.Offset[1]
			renderable.Location[2] = layout.Offset[2]
		}
	}
}

// Draw renders all of widgets on the screen.
func (ui *UIManager) Draw() {
	// calculate the perspective and view
	ortho := mgl.Ortho(0, float32(ui.width), 0, float32(ui.height), minZDepth, maxZDepth)
	view := mgl.Ident4()

	// draw all of the widgets
	for _, w := range ui.widgets {
		w.Draw(ortho, view)
	}
}

// CreateLabel creates the label widget and the text renderable.
func (ui *UIManager) CreateLabel(font *GLFont, anchor int, offset mgl.Vec3, msg string) *UILabel {
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

func (l *UILabel) Destroy() {
	l.Renderable.Destroy()
}

func (l *UILabel) Draw(perspective mgl.Mat4, view mgl.Mat4) {
	l.manager.renderer.DrawRenderable(l.Renderable, perspective, view)
}

func (l *UILabel) GetLayout() *UILayout {
	return &l.Layout
}

func (l *UILabel) GetRenderable() *Renderable {
	return l.Renderable
}
