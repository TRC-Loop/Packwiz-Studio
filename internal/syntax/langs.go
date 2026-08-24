package syntax

// words turns a list into the set the scanner looks words up in.
func words(list ...string) map[string]bool {
	set := make(map[string]bool, len(list))
	for _, w := range list {
		set[w] = true
	}
	return set
}

// toml is the packwiz index and any other TOML file.
func toml() Lang {
	return Lang{
		Name:         "TOML",
		lineComment:  []string{"#"},
		quotes:       `"'`,
		assign:       "=",
		tableHeaders: true,
		constants:    words("true", "false"),
	}
}

// packToml adds the keys a pack.toml holds, so they can be completed.
func packToml() Lang {
	l := toml()
	l.Name = "pack.toml"
	l.completions = []string{
		"name", "author", "version", "description", "pack-format",
		"[index]", "file", "hash-format", "hash",
		"[versions]", "minecraft", "fabric", "forge", "neoforge", "quilt",
		"[options]",
	}
	return l
}

// metafileToml adds the keys a mod's .pw.toml holds.
func metafileToml() Lang {
	l := toml()
	l.Name = "metafile"
	l.completions = []string{
		"name", "filename", "side", "pin", "preserve",
		"[download]", "url", "hash", "hash-format", "mode",
		"[update]", "[update.modrinth]", "mod-id", "version",
		"[option]", "optional", "default", "description",
		"both", "client", "server",
	}
	return l
}

// jsonLang is strict JSON: no comments, double quotes only.
func jsonLang() Lang {
	return Lang{
		Name:      "JSON",
		quotes:    `"`,
		assign:    ":",
		constants: words("true", "false", "null"),
	}
}

// json5 is JSON with comments, single quotes and bare keys, which is what
// most mod configs that call themselves JSON actually are.
func json5() Lang {
	l := jsonLang()
	l.Name = "JSON5"
	l.lineComment = []string{"//"}
	l.blockOpen, l.blockClose = "/*", "*/"
	l.quotes = `"'`
	l.constants = words("true", "false", "null", "Infinity", "NaN")
	return l
}

// yaml colours keys, comments and the usual scalars. Block scalars are
// left as plain text rather than guessed at.
func yaml() Lang {
	return Lang{
		Name:        "YAML",
		lineComment: []string{"#"},
		quotes:      `"'`,
		assign:      ":",
		constants: words("true", "false", "null", "yes", "no", "on", "off",
			"True", "False", "Null", "~"),
	}
}

// javascript covers KubeJS scripts, which is the JavaScript a pack holds.
// The KubeJS event names are offered for completion because they are what
// a script starts with and they are tedious to type exactly.
func javascript() Lang {
	return Lang{
		Name:        "JavaScript",
		lineComment: []string{"//"},
		blockOpen:   "/*",
		blockClose:  "*/",
		quotes:      "\"'`",
		assign:      ":",
		keywords: words("const", "let", "var", "function", "return", "if",
			"else", "for", "while", "do", "switch", "case", "default",
			"break", "continue", "new", "class", "extends", "of", "in",
			"typeof", "instanceof", "delete", "throw", "try", "catch",
			"finally", "await", "async", "yield", "export", "import",
			"from", "this"),
		constants: words("true", "false", "null", "undefined", "NaN"),
		completions: []string{
			"ServerEvents.recipes", "ServerEvents.tags", "ServerEvents.commandRegistry",
			"ServerEvents.loaded", "ClientEvents.init", "StartupEvents.registry",
			"StartupEvents.postInit", "ItemEvents.tooltip", "BlockEvents.rightClicked",
			"PlayerEvents.loggedIn", "event.recipes", "event.remove", "event.shaped",
			"event.shapeless", "event.smelting", "event.custom", "event.add",
			"console.log", "function", "const", "let",
		},
	}
}

// snbt is the text form of NBT: JSON shaped, with typed number suffixes
// and bare keys.
func snbt() Lang {
	return Lang{
		Name:      "SNBT",
		quotes:    `"'`,
		assign:    ":",
		constants: words("true", "false"),
	}
}

// properties covers .properties, .cfg and .ini, where a line is a key,
// a separator and everything after it.
func properties() Lang {
	return Lang{
		Name:         "properties",
		lineComment:  []string{"#", "!", ";"},
		assign:       "=:",
		tableHeaders: true,
		constants:    words("true", "false"),
	}
}

// generic is the fallback: comments and quotes as most configuration
// formats spell them, and nothing assumed beyond that.
func generic() Lang {
	return Lang{
		Name:        "text",
		lineComment: []string{"#"},
		quotes:      `"'`,
		assign:      "=:",
		constants:   words("true", "false"),
	}
}
