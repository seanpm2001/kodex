package web

import (
	"fmt"
	. "github.com/kiprotect/gospel"
	"github.com/kiprotect/kodex"
	"github.com/kiprotect/kodex/api"
)

func UserForm(c Context) Element {

	// this creates a state variable in the context
	name := Var(c, "this is an example")
	choice := Var(c, "salad")
	err := Var(c, "")

	// retrieve the user
	// user := Var(c, getUser(c), Cached(name))

	// How can we get stuff from the context, like e.g. a Router object?
	// We should e.g. be able to use it

	// get the router, which itself defines variables like the route
	// router := Router(c)

	// this creates a callback function in the context
	createUser := Func(c, func() {

		kodex.Log.Infof("Creating user with name '%s' and choice '%s'...", name.Get(), choice.Get())
		// newUser := controller.create(name)
		// change state according to result...
		// Gospel will re-render the resulting HTML...
		// user.Set(newUser)

		err.Set(fmt.Sprintf("this is my errossr: %s - %s", name.Get(), choice.Get()))

		// redirect to the new view, which will be either rendered directly
		// or we'll instruct the browser to really redirect...
		// router.RedirectTo("/user/13")
	})

	var errValue Element

	if err.Get() != "" {
		errValue = Div(Class("bulma-message", "bulma-is-warning"), err.Get())
	}

	// Form should recognize the 'OnSubmit' and create a trigger for
	// it, which either triggers for the current route or via JS
	// and then takes care to call the 'createUser' function, which
	// will modify variables
	return Form(
		Method("POST"),
		Class("kip-user-form", "bulma-form"),
		OnSubmit(createUser),
		errValue,
		H1(name.Get()),
		Input(Class("bulma-control"), Value(name)),
		Input(Class("bulma-control"), Value(choice)),
		Button(Class("bulma-button", "bulma-is-primary"), "submit"),
	)
}

// <h1 class="bulma-navbar-item bulma-navbar-title">Projects › My Example Project</h1><div aria-label="menu" aria-expanded="false" class="bulma-navbar-burger bulma-burger is-hidden-desktop" data-target="sidebar" role="button"><span aria-hidden="true"></span><span aria-hidden="true"></span><span aria-hidden="true"></span></div></div><div class="bulma-navbar-menu"><div class="bulma-navbar-end"><div class="kip-navbar-dropdown-menu bulma-navbar-item bulma-has-dropdown"><a aria-haspopup="true" aria-expanded="false" class="bulma-navbar-link" role="button" tabindex="0"><span><span class="icon is-small"><i class="fas fa-th-large"></i></span><span class="bulma-is-hidden-navbar">Apps</span></span></a><div class="kip-navbar-dropdown bulma-navbar-dropdown bulma-is-right"><a class="kip-navbar-dropdown__item bulma-dropdown-item" href="/klaro"><span><span class="icon is-small"><i class="fas fa-check-circle"></i></span>Klaro</span></a><a class="kip-navbar-dropdown__item bulma-dropdown-item" href="/kodex"><span><span class="icon is-small"><i class="fas fa-book-open"></i></span>Kodex</span></a><a class="kip-navbar-dropdown__item bulma-dropdown-item" href="/admin"><span><span class="icon is-small"><i class="fas fa-cogs"></i></span>Administration</span></a></div></div><div class="kip-navbar-dropdown-menu bulma-navbar-item bulma-has-dropdown"><a aria-haspopup="true" aria-expanded="false" class="bulma-navbar-link" role="button" tabindex="0"><div class="kip-nowrap"><span class="icon is-small"><i class="fas fa-user-circle"></i></span><span class="kip-overflow-ellipsis bulma-is-hidden-navbar">azure@kiprotect.com</span></div></a><div class="kip-navbar-dropdown bulma-navbar-dropdown bulma-is-right"><a class="kip-navbar-dropdown__item bulma-dropdown-item" href="/logout"><span><span class="icon is-small"><i class="fas fa-sign-out-alt"></i></span>Log out</span></a></div></div></div></div></header>

func AuthorizedContent(c Context) Element {

	userProvider := UseUserProvider(c)
	controller := UseController(c)
	router := UseRouter(c)

	// we get the user from the provider...
	externalUser, _ := userProvider.Get(controller, c.Request())

	// we redirect to the login page
	if externalUser == nil {
		router.RedirectTo("/login")
		return nil
	}

	SetExternalUser(c, externalUser)

	return F(
		c.Element("navHeader", Navbar),
		c.Element("contentWithSidebar", WithSidebar(
			c.Element("sidebar", Sidebar),
			c.Element("mainContent", MainContent),
		)),
	)

}

func Logout(c Context) Element {
	return Div()
}

func Login(c Context) Element {
	return Div(
		Class("kip-with-app-selector"),
		A(
			Class("kip-with-app-selector-link"),
			Href("/#"),
			Div(
				Class("kip-logo-wrapper"),
				Img(
					Class("kip-logo", Alt("projects")),
					Src("/static/images/kodexlogo-white.png"),
				),
			),
		),
		Section(
			Class("kip-centered-card", "kip-is-info", "kip-is-fullheight"),
			Div(
				Class("kip-card", "kip-is-centered", "kip-account"),
				Div(
					Class("kip-card-header"),
					Div(
						Class("kip-card-title"),
						H2("Login"),
					),
				),
				Div(
					Class("kip-card-content", "kip-card-centered"),
					Div(
						Class("kip-login"),
						Div(
							Class("kip-provider-list"),
							Ul(
								Li(
									A(
										Href("/api/v1/login"),
										Button(
											Class("bulma-button", "bulma-is-success", "bulma-is-flex"),
											"Log in via SSO",
										),
									),
								),
							),
						),
					),
				),
			),
		),
	)
}

func AppContent(c Context) Element {

	router := UseRouter(c)

	return router.Match(
		c,
		Route("/login", Login),
		Route("/logout", Logout),
		Route("", AuthorizedContent),
	)

}

func Root(controller api.Controller) (func(c Context) Element, error) {

	userProvider, err := controller.UserProvider()

	if err != nil {
		return nil, err
	}

	return func(c Context) Element {

		// we set the user provider
		SetUserProvider(c, userProvider)

		// we set the controller
		SetController(c, controller)

		title := GlobalVar(c, "title", "Kodex")

		return F(
			Doctype("html"),
			Html(
				Lang("en"),
				Head(
					Meta(Charset("utf-8")),
					Title(title.Get()),
					// Link(Rel("apple-touch-icon"), Sizes("180x180"), Href("/icons/apple-touch-icon.png")),
					// Link(Rel("icon"), Type("image/png"), Sizes("32x32"), Href("/icons/favicon-32x32.png")),
					Link(Rel("stylesheet"), Href("/static/main.css")),
					Script(Src("/static/gospel.js"), Type("module")),
				),
				Body(
					Class("kip-fonts", "bulma-body"),
					Div(
						Class("kip"),
						c.Element("appContent", AppContent),
					),
				),
			),
		)

	}, nil
}
