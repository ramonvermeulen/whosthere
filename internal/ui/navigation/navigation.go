package navigation

import "github.com/rivo/tview"

// Page is a UI page that can be registered with the Router.
type Page interface {
	GetName() string
	GetPrimitive() tview.Primitive
	FocusTarget() tview.Primitive
	Refresh()
}

const (
	RouteDashboard   = "dashboard"
	RouteSplash      = "splash"
	RouteDetail      = "detail"
	RouteThemePicker = "theme-picker"
)

// Router is both the visual pages container and the logical router.
type Router struct {
	*tview.Pages
	pages       map[string]Page
	currentPage string
}

func NewRouter() *Router {
	return &Router{
		Pages: tview.NewPages(),
		pages: make(map[string]Page),
	}
}

func (r *Router) Register(p Page) {
	name := p.GetName()
	r.pages[name] = p
	r.AddPage(name, p.GetPrimitive(), true, false)
}

func (r *Router) NavigateTo(name string) {
	if _, ok := r.pages[name]; !ok {
		return
	}
	r.currentPage = name
	r.SwitchToPage(name) // SwitchToPage hides all others and shows this one
	r.pages[name].Refresh()
}

// ShowOverlay shows a page as an overlay on top of the current page.
// Use this for modals/dialogs that should not hide the underlying page.
func (r *Router) ShowOverlay(name string) {
	page, ok := r.pages[name]
	if !ok {
		return
	}
	r.ShowPage(name) // ShowPage makes it visible without hiding others
	page.Refresh()
}

// HideOverlay hides an overlay page, revealing the page underneath.
func (r *Router) HideOverlay(name string) {
	r.HidePage(name)
}

func (r *Router) FocusCurrent(app *tview.Application) {
	if app == nil {
		return
	}
	p, ok := r.pages[r.currentPage]
	if !ok || p == nil {
		app.SetFocus(r)
		return
	}
	if ft := p.FocusTarget(); ft != nil {
		app.SetFocus(ft)
		return
	}
	app.SetFocus(p.GetPrimitive())
}

func (r *Router) Current() string { return r.currentPage }

func (r *Router) Page(name string) Page {
	return r.pages[name]
}
