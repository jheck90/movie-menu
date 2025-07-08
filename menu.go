package main

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type MenuPage struct {
	app.Compo
	selectedList string
}

func (p *MenuPage) OnMount(ctx app.Context) {
	// Set default list if none selected
	if p.selectedList == "" && len(MovieMenus) > 0 {
		for k := range MovieMenus {
			p.selectedList = k
			break
		}
	}
}

// Render shows the currently selected list of movies
func (p *MenuPage) Render() app.UI {
	list, ok := MovieMenus[p.selectedList]
	if !ok {
		return app.Div().Body(
			app.H1().Text("Movie Menu"),
			app.P().Text("No movie list found."),
		)
	}

	return app.Div().Class("app-dark").Body(
		app.H1().Text("Pick a Movie!"),
		p.renderListSelector(),
		app.Div().Class("poster-grid").Body(
			app.Range(list.Movies).Slice(func(i int) app.UI {
				movie := list.Movies[i]
				return app.Div().Class("poster-card").Body(
					app.Img().
						Src(movie.PosterURL).
						Alt(movie.Title).
						Class("poster-img"),
					app.P().Text(movie.Title).Class("poster-title"),
				)
			}),
		),
	)
}

// renderListSelector allows user to choose between different lists
func (p *MenuPage) renderListSelector() app.UI {
	return app.Select().
		Class("list-selector").
		OnChange(p.onListChange).
		Body(
			app.Range(getListNames()).Slice(func(i int) app.UI {
				name := getListNames()[i]
				return app.Option().
					Value(name).
					Selected(name == p.selectedList).
					Text(name)
			}),
		)
}

// onListChange updates the selected list
func (p *MenuPage) onListChange(ctx app.Context, e app.Event) {
	p.selectedList = ctx.JSSrc().Get("value").String()
	p.Update()
}

// getListNames returns sorted names of available lists
func getListNames() []string {
	names := make([]string, 0, len(MovieMenus))
	for name := range MovieMenus {
		names = append(names, name)
	}
	return names
}
