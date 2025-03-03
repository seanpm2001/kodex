package web

import (
	"bytes"
	. "github.com/gospel-dev/gospel"
	"github.com/kiprotect/kodex"
	"github.com/kiprotect/kodex/web/ui"
	"time"
	//	"github.com/kiprotect/kodex/api"
)

func StreamConfigs(stream kodex.Stream, onUpdate func(ChangeInfo, string)) ElementFunction {

	return func(c Context) Element {

		router := UseRouter(c)

		return F(
			router.Match(
				c,
				Route("", StreamConfigsList(stream, onUpdate)),
			),
		)
	}
}

func ConfigSettings(config kodex.Config, onUpdate func(ChangeInfo, string)) func(c Context) Element {
	return func(c Context) Element {
		return F(
			H2(
				Class("bulma-subtitle"),
				"API URL",
			),
			P(
				"To transform data using this config, use the following URL in a POST request, e.g. ", Code("curl -X POST -h \"Content-Type: application/json\" -d '{\"items\": [...]}' ..."),
			),
			Hr(),
			Pre(
				Span(Id("host")), Fmt("/api/v1/configs/%s/transform", Hex(config.ID())),
			),
			Script(`host.innerText = location.protocol + '//' + location.host;`),
		)
	}
}

func StreamConfigDetails(stream kodex.Stream, onUpdate func(ChangeInfo, string)) func(c Context, configId, tab string) Element {

	return func(c Context, configId, tab string) Element {

		if tab == "" {
			tab = "actions"
		}

		config, err := stream.Config(Unhex(configId))

		if err != nil {
			return Div(Fmt("Cannot get config %s...", configId))
		}

		// make sure this config belongs to the stream...
		if !bytes.Equal(config.Stream().ID(), stream.ID()) {
			return nil
		}

		AddBreadcrumb(c, "Configs", "")
		AddBreadcrumb(c, config.Name(), Fmt("/details/%s", Hex(config.ID())))

		router := UseRouter(c)

		name := Var(c, config.Name())
		error := Var(c, "")

		onSubmit := Func[any](c, func() {

			if name.Get() == "" {
				error.Set("please enter a name")
				return
			}

			if err := config.Update(map[string]any{"name": name.Get()}); err != nil {
				error.Set(Fmt("cannot set name: %v", err))
				return
			}

			if err := config.Save(); err != nil {
				error.Set(Fmt("cannot save: %v", err))
				return
			}

			onUpdate(ChangeInfo{}, router.LastPath())
		})

		var errorNotice Element

		if error.Get() != "" {
			errorNotice = P(
				Class("bulma-help", "bulma-is-danger"),
				error.Get(),
			)
		}

		// edit the name of the config
		editStreamConfigName := func(c Context) Element {
			return Form(
				Method("POST"),
				OnSubmit(onSubmit),
				Fieldset(
					errorNotice,
					Div(
						Class("bulma-field", "bulma-has-addons"),
						P(
							Class("bulma-control"),
							Input(Class("bulma-control", "bulma-input"), Value(name)),
						),
						P(
							Class("bulma-control"),
							Button(
								Class("bulma-button", "bulma-is-success"),
								Type("submit"),
								"Change",
							),
						),
					),
				),
			)
		}

		var content Element

		switch tab {
		case "actions":
			content = c.Element("configActions", ConfigActionsList(config, onUpdate))
		case "settings":
			content = c.Element("configSettings", ConfigSettings(config, onUpdate))
		}

		basePath := Fmt("/projects/%s/streams/details/%s/configs/details/%s", Hex(stream.Project().ID()), Hex(stream.ID()), configId)

		return Div(
			H2(
				Class("bulma-title"),
				router.Match(
					c,
					If(onUpdate != nil,
						Route("/name/edit",
							c.ElementFunction("editName", editStreamConfigName),
						),
					),
					Route("",
						F(
							"Config: ",
							config.Name(),
							If(onUpdate != nil,
								A(
									Style("float: right"),
									Href(router.CurrentRoute().Path+"/name/edit"),
									Nbsp,
									Nbsp,
									I(Class("fas fa-edit")),
								),
							),
						),
					),
				),
			),
			Div(
				Class("bulma-tags"),
				Span(
					Class("bulma-tag", "bulma-is-info", "bulma-is-light"),
					Fmt("id: %s", Hex(config.ID())),
				),
				Span(
					Class("bulma-tag", "bulma-is-info", "bulma-is-light"),
					Fmt("last modified: %s", HumanDuration(time.Now().Sub(config.CreatedAt()))),
				),
			),
			Div(Class("bulma-content"), IfElse(config.Description() != "", config.Description(), "(no description given)")),
			ui.Tabs(
				ui.Tab(ui.ActiveTab(tab == "actions"), A(Href(basePath+"/actions"), "Actions")),
				ui.Tab(ui.ActiveTab(tab == "settings"), A(Href(basePath+"/settings"), "Settings")),
			),
			content,
		)
	}

}

func NewStreamConfig(stream kodex.Stream, onUpdate func(ChangeInfo, string)) ElementFunction {
	return func(c Context) Element {

		name := Var(c, "")
		error := Var(c, "")
		router := UseRouter(c)

		onSubmit := Func[any](c, func() {

			if name.Get() == "" {
				error.Set("Please enter a name")
				return
			}

			config := stream.MakeConfig(nil)

			config.SetName(name.Get())

			if err := config.Save(); err != nil {
				error.Set("Cannot save config")
			} else {
				onUpdate(ChangeInfo{}, router.CurrentPath())
			}

		})

		var errorNotice Element

		if error.Get() != "" {
			errorNotice = P(
				Class("bulma-help", "bulma-is-danger"),
				error.Get(),
			)
		}

		return Form(
			Method("POST"),
			OnSubmit(onSubmit),
			Div(
				Class("bulma-field"),
				errorNotice,
				Label(
					Class("bulma-label", "Name"),
					Input(
						Class("bulma-input", If(error.Get() != "", "bulma-is-danger")),
						Type("text"),
						Value(name),
						Placeholder("config name"),
					),
				),
			),
			Div(
				Class("bulma-field"),
				P(
					Class("bulma-control"),
					Button(
						Class("bulma-button", "bulma-is-success"),
						Type("submit"),
						"Create StreamConfig",
					),
				),
			),
		)
	}
}

func StreamConfigsList(stream kodex.Stream, onUpdate func(ChangeInfo, string)) ElementFunction {

	return func(c Context) Element {

		// we retrieve the config configs of the stream...
		configs, err := stream.Configs()

		if err != nil {
			// to do: error handling
			return nil
		}

		ais := make([]Element, 0, len(configs))

		for _, config := range configs {
			configItem := A(
				Href(Fmt("/projects/%s/streams/details/%s/configs/details/%s", Hex(stream.Project().ID()), Hex(stream.ID()), Hex(config.ID()))),
				ui.ListItem(
					ui.ListColumn("md", config.Name()),
					ui.ListColumn("sm", HumanDuration(time.Now().Sub(config.CreatedAt()))),
				),
			)
			ais = append(ais, configItem)
		}

		router := UseRouter(c)

		return F(
			router.Match(
				c,
				If(onUpdate != nil, Route("/new", c.Element("newStreamConfig", NewStreamConfig(stream, onUpdate)))),
				Route("", F(
					IfElse(
						len(ais) > 0,
						ui.List(
							ui.ListHeader(
								ui.ListColumn("md", "Name"),
								ui.ListColumn("sm", "Created At"),
							),
							ais,
						),
						ui.Message(
							"info",
							"No existing configs.",
						),
					),
					If(onUpdate != nil,
						A(
							Href(router.CurrentRoute().Path+"/new"),
							Class("bulma-button", "bulma-is-success"),
							"New StreamConfig",
						),
					),
				),
				),
			),
		)
	}
}
