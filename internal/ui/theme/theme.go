package theme

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ramonvermeulen/whosthere/internal/core/config"
	"github.com/rivo/tview"
)

var log = slog.Default()

var noColor = os.Getenv("NO_COLOR") != ""

var registry = map[string]tview.Theme{
	config.DefaultThemeName: {
		PrimitiveBackgroundColor:    tcell.GetColor("#000000"),
		ContrastBackgroundColor:     tcell.GetColor("#0c0c0c"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e1e1e"),
		BorderColor:                 tcell.GetColor("#6666ff"),
		TitleColor:                  tcell.GetColor("#ffffff"),
		GraphicsColor:               tcell.GetColor("#ff66ff"),
		PrimaryTextColor:            tcell.GetColor("#ffffff"),
		SecondaryTextColor:          tcell.GetColor("#66ffff"),
		TertiaryTextColor:           tcell.GetColor("#ffff66"),
		InverseTextColor:            tcell.GetColor("#000000"),
		ContrastSecondaryTextColor:  tcell.GetColor("#66ff66"),
	},

	"midnight-blue": {
		PrimitiveBackgroundColor:    tcell.GetColor("#000a1a"),
		ContrastBackgroundColor:     tcell.GetColor("#001a33"),
		MoreContrastBackgroundColor: tcell.GetColor("#003366"),
		BorderColor:                 tcell.GetColor("#0088ff"),
		TitleColor:                  tcell.GetColor("#00ffff"),
		GraphicsColor:               tcell.GetColor("#00ffaa"),
		PrimaryTextColor:            tcell.GetColor("#cceeff"),
		SecondaryTextColor:          tcell.GetColor("#6699ff"),
		TertiaryTextColor:           tcell.GetColor("#ffaa00"),
		InverseTextColor:            tcell.GetColor("#000a1a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#88ddff"),
	},

	"dracula": {
		PrimitiveBackgroundColor:    tcell.GetColor("#282a36"),
		ContrastBackgroundColor:     tcell.GetColor("#44475a"),
		MoreContrastBackgroundColor: tcell.GetColor("#6272a4"),
		BorderColor:                 tcell.GetColor("#bd93f9"),
		TitleColor:                  tcell.GetColor("#8be9fd"),
		GraphicsColor:               tcell.GetColor("#ff79c6"),
		PrimaryTextColor:            tcell.GetColor("#f8f8f2"),
		SecondaryTextColor:          tcell.GetColor("#bd93f9"),
		TertiaryTextColor:           tcell.GetColor("#ffb86c"),
		InverseTextColor:            tcell.GetColor("#44475a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#50fa7b"),
	},

	"nord": {
		PrimitiveBackgroundColor:    tcell.GetColor("#2e3440"),
		ContrastBackgroundColor:     tcell.GetColor("#3b4252"),
		MoreContrastBackgroundColor: tcell.GetColor("#434c5e"),
		BorderColor:                 tcell.GetColor("#81a1c1"),
		TitleColor:                  tcell.GetColor("#88c0d0"),
		GraphicsColor:               tcell.GetColor("#bf616a"),
		PrimaryTextColor:            tcell.GetColor("#d8dee9"),
		SecondaryTextColor:          tcell.GetColor("#81a1c1"),
		TertiaryTextColor:           tcell.GetColor("#ebcb8b"),
		InverseTextColor:            tcell.GetColor("#4c566a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a3be8c"),
	},

	"solarized-dark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#002b36"),
		ContrastBackgroundColor:     tcell.GetColor("#073642"),
		MoreContrastBackgroundColor: tcell.GetColor("#586e75"),
		BorderColor:                 tcell.GetColor("#268bd2"),
		TitleColor:                  tcell.GetColor("#2aa198"),
		GraphicsColor:               tcell.GetColor("#d33682"),
		PrimaryTextColor:            tcell.GetColor("#839496"),
		SecondaryTextColor:          tcell.GetColor("#268bd2"),
		TertiaryTextColor:           tcell.GetColor("#b58900"),
		InverseTextColor:            tcell.GetColor("#fdf6e3"),
		ContrastSecondaryTextColor:  tcell.GetColor("#859900"),
	},

	"solarized-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#fdf6e3"),
		ContrastBackgroundColor:     tcell.GetColor("#eee8d5"),
		MoreContrastBackgroundColor: tcell.GetColor("#93a1a1"),
		BorderColor:                 tcell.GetColor("#268bd2"),
		TitleColor:                  tcell.GetColor("#2aa198"),
		GraphicsColor:               tcell.GetColor("#d33682"),
		PrimaryTextColor:            tcell.GetColor("#657b83"),
		SecondaryTextColor:          tcell.GetColor("#268bd2"),
		TertiaryTextColor:           tcell.GetColor("#b58900"),
		InverseTextColor:            tcell.GetColor("#002b36"),
		ContrastSecondaryTextColor:  tcell.GetColor("#859900"),
	},

	"gruvbox-dark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#282828"),
		ContrastBackgroundColor:     tcell.GetColor("#3c3836"),
		MoreContrastBackgroundColor: tcell.GetColor("#504945"),
		BorderColor:                 tcell.GetColor("#d65d0e"),
		TitleColor:                  tcell.GetColor("#689d6a"),
		GraphicsColor:               tcell.GetColor("#cc241d"),
		PrimaryTextColor:            tcell.GetColor("#ebdbb2"),
		SecondaryTextColor:          tcell.GetColor("#fe8019"),
		TertiaryTextColor:           tcell.GetColor("#fabd2f"),
		InverseTextColor:            tcell.GetColor("#3c3836"),
		ContrastSecondaryTextColor:  tcell.GetColor("#b8bb26"),
	},

	"onedark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#282c34"),
		ContrastBackgroundColor:     tcell.GetColor("#2c323c"),
		MoreContrastBackgroundColor: tcell.GetColor("#3e4452"),
		BorderColor:                 tcell.GetColor("#61afef"),
		TitleColor:                  tcell.GetColor("#56b6c2"),
		GraphicsColor:               tcell.GetColor("#e06c75"),
		PrimaryTextColor:            tcell.GetColor("#abb2bf"),
		SecondaryTextColor:          tcell.GetColor("#61afef"),
		TertiaryTextColor:           tcell.GetColor("#d19a66"),
		InverseTextColor:            tcell.GetColor("#3e4452"),
		ContrastSecondaryTextColor:  tcell.GetColor("#98c379"),
	},

	"tokyonight": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1a1b26"),
		ContrastBackgroundColor:     tcell.GetColor("#24283b"),
		MoreContrastBackgroundColor: tcell.GetColor("#414868"),
		BorderColor:                 tcell.GetColor("#7aa2f7"),
		TitleColor:                  tcell.GetColor("#2ac3de"),
		GraphicsColor:               tcell.GetColor("#f7768e"),
		PrimaryTextColor:            tcell.GetColor("#c0caf5"),
		SecondaryTextColor:          tcell.GetColor("#7aa2f7"),
		TertiaryTextColor:           tcell.GetColor("#e0af68"),
		InverseTextColor:            tcell.GetColor("#414868"),
		ContrastSecondaryTextColor:  tcell.GetColor("#9ece6a"),
	},

	"catppuccin-mocha": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1e1e2e"),
		ContrastBackgroundColor:     tcell.GetColor("#313244"),
		MoreContrastBackgroundColor: tcell.GetColor("#45475a"),
		BorderColor:                 tcell.GetColor("#89b4fa"),
		TitleColor:                  tcell.GetColor("#89dceb"),
		GraphicsColor:               tcell.GetColor("#f38ba8"),
		PrimaryTextColor:            tcell.GetColor("#cdd6f4"),
		SecondaryTextColor:          tcell.GetColor("#b4befe"),
		TertiaryTextColor:           tcell.GetColor("#f9e2af"),
		InverseTextColor:            tcell.GetColor("#313244"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a6e3a1"),
	},

	"rose-pine": {
		PrimitiveBackgroundColor:    tcell.GetColor("#191724"),
		ContrastBackgroundColor:     tcell.GetColor("#1f1d2e"),
		MoreContrastBackgroundColor: tcell.GetColor("#26233a"),
		BorderColor:                 tcell.GetColor("#31748f"),
		TitleColor:                  tcell.GetColor("#9ccfd8"),
		GraphicsColor:               tcell.GetColor("#eb6f92"),
		PrimaryTextColor:            tcell.GetColor("#e0def4"),
		SecondaryTextColor:          tcell.GetColor("#c4a7e7"),
		TertiaryTextColor:           tcell.GetColor("#f6c177"),
		InverseTextColor:            tcell.GetColor("#1f1d2e"),
		ContrastSecondaryTextColor:  tcell.GetColor("#9ccfd8"),
	},

	"monokai": {
		PrimitiveBackgroundColor:    tcell.GetColor("#272822"),
		ContrastBackgroundColor:     tcell.GetColor("#3e3d32"),
		MoreContrastBackgroundColor: tcell.GetColor("#75715e"),
		BorderColor:                 tcell.GetColor("#66d9ef"),
		TitleColor:                  tcell.GetColor("#a6e22e"),
		GraphicsColor:               tcell.GetColor("#f92672"),
		PrimaryTextColor:            tcell.GetColor("#f8f8f2"),
		SecondaryTextColor:          tcell.GetColor("#fd971f"),
		TertiaryTextColor:           tcell.GetColor("#ae81ff"),
		InverseTextColor:            tcell.GetColor("#3e3d32"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a6e22e"),
	},

	"material": {
		PrimitiveBackgroundColor:    tcell.GetColor("#263238"),
		ContrastBackgroundColor:     tcell.GetColor("#37474f"),
		MoreContrastBackgroundColor: tcell.GetColor("#546e7a"),
		BorderColor:                 tcell.GetColor("#82b1ff"),
		TitleColor:                  tcell.GetColor("#80deea"),
		GraphicsColor:               tcell.GetColor("#ff5252"),
		PrimaryTextColor:            tcell.GetColor("#cfd8dc"),
		SecondaryTextColor:          tcell.GetColor("#b388ff"),
		TertiaryTextColor:           tcell.GetColor("#ffd740"),
		InverseTextColor:            tcell.GetColor("#37474f"),
		ContrastSecondaryTextColor:  tcell.GetColor("#69f0ae"),
	},

	"high-contrast": {
		PrimitiveBackgroundColor:    tcell.GetColor("#000000"),
		ContrastBackgroundColor:     tcell.GetColor("#0a0a0a"),
		MoreContrastBackgroundColor: tcell.GetColor("#1a1a1a"),
		BorderColor:                 tcell.GetColor("#00ffff"),
		TitleColor:                  tcell.GetColor("#ffff00"),
		GraphicsColor:               tcell.GetColor("#ff00ff"),
		PrimaryTextColor:            tcell.GetColor("#ffffff"),
		SecondaryTextColor:          tcell.GetColor("#00ffff"),
		TertiaryTextColor:           tcell.GetColor("#ffff00"),
		InverseTextColor:            tcell.GetColor("#ffffff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ff00"),
	},

	"papercolor-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#eeeeee"),
		ContrastBackgroundColor:     tcell.GetColor("#afafaf"),
		MoreContrastBackgroundColor: tcell.GetColor("#878787"),
		BorderColor:                 tcell.GetColor("#0087af"),
		TitleColor:                  tcell.GetColor("#00afaf"),
		GraphicsColor:               tcell.GetColor("#d7005f"),
		PrimaryTextColor:            tcell.GetColor("#444444"),
		SecondaryTextColor:          tcell.GetColor("#005f87"),
		TertiaryTextColor:           tcell.GetColor("#d75f00"),
		InverseTextColor:            tcell.GetColor("#eeeeee"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00af87"),
	},

	"ayu-dark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0e14"),
		ContrastBackgroundColor:     tcell.GetColor("#0f1419"),
		MoreContrastBackgroundColor: tcell.GetColor("#1a1f29"),
		BorderColor:                 tcell.GetColor("#39bae6"),
		TitleColor:                  tcell.GetColor("#95e6cb"),
		GraphicsColor:               tcell.GetColor("#ff3333"),
		PrimaryTextColor:            tcell.GetColor("#b3b1ad"),
		SecondaryTextColor:          tcell.GetColor("#59c2ff"),
		TertiaryTextColor:           tcell.GetColor("#ffb454"),
		InverseTextColor:            tcell.GetColor("#1a1f29"),
		ContrastSecondaryTextColor:  tcell.GetColor("#c2d94c"),
	},

	"everforest": {
		PrimitiveBackgroundColor:    tcell.GetColor("#2b3339"),
		ContrastBackgroundColor:     tcell.GetColor("#3c474d"),
		MoreContrastBackgroundColor: tcell.GetColor("#4b565c"),
		BorderColor:                 tcell.GetColor("#7fbbb3"),
		TitleColor:                  tcell.GetColor("#83c092"),
		GraphicsColor:               tcell.GetColor("#e67e80"),
		PrimaryTextColor:            tcell.GetColor("#d3c6aa"),
		SecondaryTextColor:          tcell.GetColor("#a7c080"),
		TertiaryTextColor:           tcell.GetColor("#dbbc7f"),
		InverseTextColor:            tcell.GetColor("#3c474d"),
		ContrastSecondaryTextColor:  tcell.GetColor("#83c092"),
	},

	"ayu-mirage": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1f2430"),
		ContrastBackgroundColor:     tcell.GetColor("#232834"),
		MoreContrastBackgroundColor: tcell.GetColor("#2a3042"),
		BorderColor:                 tcell.GetColor("#73d0ff"),
		TitleColor:                  tcell.GetColor("#5ccfe6"),
		GraphicsColor:               tcell.GetColor("#ff6666"),
		PrimaryTextColor:            tcell.GetColor("#cbccc6"),
		SecondaryTextColor:          tcell.GetColor("#73d0ff"),
		TertiaryTextColor:           tcell.GetColor("#ffd580"),
		InverseTextColor:            tcell.GetColor("#232834"),
		ContrastSecondaryTextColor:  tcell.GetColor("#bae67e"),
	},

	"catppuccin-latte": {
		PrimitiveBackgroundColor:    tcell.GetColor("#eff1f5"),
		ContrastBackgroundColor:     tcell.GetColor("#e6e9ef"),
		MoreContrastBackgroundColor: tcell.GetColor("#dce0e8"),
		BorderColor:                 tcell.GetColor("#1e66f5"),
		TitleColor:                  tcell.GetColor("#04a5e5"),
		GraphicsColor:               tcell.GetColor("#d20f39"),
		PrimaryTextColor:            tcell.GetColor("#4c4f69"),
		SecondaryTextColor:          tcell.GetColor("#8839ef"),
		TertiaryTextColor:           tcell.GetColor("#fe640b"),
		InverseTextColor:            tcell.GetColor("#dce0e8"),
		ContrastSecondaryTextColor:  tcell.GetColor("#40a02b"),
	},

	"catppuccin-frappe": {
		PrimitiveBackgroundColor:    tcell.GetColor("#303446"),
		ContrastBackgroundColor:     tcell.GetColor("#292c3c"),
		MoreContrastBackgroundColor: tcell.GetColor("#414559"),
		BorderColor:                 tcell.GetColor("#8caaee"),
		TitleColor:                  tcell.GetColor("#81c8be"),
		GraphicsColor:               tcell.GetColor("#e78284"),
		PrimaryTextColor:            tcell.GetColor("#c6d0f5"),
		SecondaryTextColor:          tcell.GetColor("#ca9ee6"),
		TertiaryTextColor:           tcell.GetColor("#ef9f76"),
		InverseTextColor:            tcell.GetColor("#414559"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a6d189"),
	},

	"catppuccin-macchiato": {
		PrimitiveBackgroundColor:    tcell.GetColor("#24273a"),
		ContrastBackgroundColor:     tcell.GetColor("#1e2030"),
		MoreContrastBackgroundColor: tcell.GetColor("#363a4f"),
		BorderColor:                 tcell.GetColor("#8aadf4"),
		TitleColor:                  tcell.GetColor("#8bd5ca"),
		GraphicsColor:               tcell.GetColor("#ed8796"),
		PrimaryTextColor:            tcell.GetColor("#cad3f5"),
		SecondaryTextColor:          tcell.GetColor("#c6a0f6"),
		TertiaryTextColor:           tcell.GetColor("#f5a97f"),
		InverseTextColor:            tcell.GetColor("#363a4f"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a6da95"),
	},

	"tokyonight-storm": {
		PrimitiveBackgroundColor:    tcell.GetColor("#24283b"),
		ContrastBackgroundColor:     tcell.GetColor("#1f2335"),
		MoreContrastBackgroundColor: tcell.GetColor("#292e42"),
		BorderColor:                 tcell.GetColor("#7aa2f7"),
		TitleColor:                  tcell.GetColor("#7dcfff"),
		GraphicsColor:               tcell.GetColor("#f7768e"),
		PrimaryTextColor:            tcell.GetColor("#c0caf5"),
		SecondaryTextColor:          tcell.GetColor("#9aa5ce"),
		TertiaryTextColor:           tcell.GetColor("#e0af68"),
		InverseTextColor:            tcell.GetColor("#292e42"),
		ContrastSecondaryTextColor:  tcell.GetColor("#9ece6a"),
	},

	"tokyonight-moon": {
		PrimitiveBackgroundColor:    tcell.GetColor("#222436"),
		ContrastBackgroundColor:     tcell.GetColor("#1e2030"),
		MoreContrastBackgroundColor: tcell.GetColor("#2f334d"),
		BorderColor:                 tcell.GetColor("#82aaff"),
		TitleColor:                  tcell.GetColor("#86e1fc"),
		GraphicsColor:               tcell.GetColor("#ff757f"),
		PrimaryTextColor:            tcell.GetColor("#c8d3f5"),
		SecondaryTextColor:          tcell.GetColor("#c099ff"),
		TertiaryTextColor:           tcell.GetColor("#ffc777"),
		InverseTextColor:            tcell.GetColor("#2f334d"),
		ContrastSecondaryTextColor:  tcell.GetColor("#c3e88d"),
	},

	"rose-pine-moon": {
		PrimitiveBackgroundColor:    tcell.GetColor("#232136"),
		ContrastBackgroundColor:     tcell.GetColor("#2a273f"),
		MoreContrastBackgroundColor: tcell.GetColor("#393552"),
		BorderColor:                 tcell.GetColor("#3e8fb0"),
		TitleColor:                  tcell.GetColor("#9ccfd8"),
		GraphicsColor:               tcell.GetColor("#eb6f92"),
		PrimaryTextColor:            tcell.GetColor("#e0def4"),
		SecondaryTextColor:          tcell.GetColor("#c4a7e7"),
		TertiaryTextColor:           tcell.GetColor("#f6c177"),
		InverseTextColor:            tcell.GetColor("#2a273f"),
		ContrastSecondaryTextColor:  tcell.GetColor("#9ccfd8"),
	},

	"rose-pine-dawn": {
		PrimitiveBackgroundColor:    tcell.GetColor("#faf4ed"),
		ContrastBackgroundColor:     tcell.GetColor("#fffaf3"),
		MoreContrastBackgroundColor: tcell.GetColor("#f2e9e1"),
		BorderColor:                 tcell.GetColor("#286983"),
		TitleColor:                  tcell.GetColor("#56949f"),
		GraphicsColor:               tcell.GetColor("#b4637a"),
		PrimaryTextColor:            tcell.GetColor("#575279"),
		SecondaryTextColor:          tcell.GetColor("#907aa9"),
		TertiaryTextColor:           tcell.GetColor("#ea9d34"),
		InverseTextColor:            tcell.GetColor("#fffaf3"),
		ContrastSecondaryTextColor:  tcell.GetColor("#56949f"),
	},

	"onedark-vivid": {
		PrimitiveBackgroundColor:    tcell.GetColor("#21252b"),
		ContrastBackgroundColor:     tcell.GetColor("#282c34"),
		MoreContrastBackgroundColor: tcell.GetColor("#3a3f4b"),
		BorderColor:                 tcell.GetColor("#61afef"),
		TitleColor:                  tcell.GetColor("#56b6c2"),
		GraphicsColor:               tcell.GetColor("#e06c75"),
		PrimaryTextColor:            tcell.GetColor("#abb2bf"),
		SecondaryTextColor:          tcell.GetColor("#c678dd"),
		TertiaryTextColor:           tcell.GetColor("#d19a66"),
		InverseTextColor:            tcell.GetColor("#3a3f4b"),
		ContrastSecondaryTextColor:  tcell.GetColor("#98c379"),
	},

	"gruvbox-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#fbf1c7"),
		ContrastBackgroundColor:     tcell.GetColor("#ebdbb2"),
		MoreContrastBackgroundColor: tcell.GetColor("#d5c4a1"),
		BorderColor:                 tcell.GetColor("#b57614"),
		TitleColor:                  tcell.GetColor("#427b58"),
		GraphicsColor:               tcell.GetColor("#cc241d"),
		PrimaryTextColor:            tcell.GetColor("#3c3836"),
		SecondaryTextColor:          tcell.GetColor("#8f3f71"),
		TertiaryTextColor:           tcell.GetColor("#b57614"),
		InverseTextColor:            tcell.GetColor("#ebdbb2"),
		ContrastSecondaryTextColor:  tcell.GetColor("#797403"),
	},

	"nordic": {
		PrimitiveBackgroundColor:    tcell.GetColor("#2e3440"),
		ContrastBackgroundColor:     tcell.GetColor("#3b4252"),
		MoreContrastBackgroundColor: tcell.GetColor("#434c5e"),
		BorderColor:                 tcell.GetColor("#81a1c1"),
		TitleColor:                  tcell.GetColor("#8fbcbb"),
		GraphicsColor:               tcell.GetColor("#bf616a"),
		PrimaryTextColor:            tcell.GetColor("#d8dee9"),
		SecondaryTextColor:          tcell.GetColor("#88c0d0"),
		TertiaryTextColor:           tcell.GetColor("#ebcb8b"),
		InverseTextColor:            tcell.GetColor("#4c566a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#a3be8c"),
	},

	"purple-haze": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1a1020"),
		ContrastBackgroundColor:     tcell.GetColor("#24182c"),
		MoreContrastBackgroundColor: tcell.GetColor("#2f2038"),
		BorderColor:                 tcell.GetColor("#9d4edd"),
		TitleColor:                  tcell.GetColor("#c77dff"),
		GraphicsColor:               tcell.GetColor("#ff6d6d"),
		PrimaryTextColor:            tcell.GetColor("#e0d6eb"),
		SecondaryTextColor:          tcell.GetColor("#bb86fc"),
		TertiaryTextColor:           tcell.GetColor("#ffb74d"),
		InverseTextColor:            tcell.GetColor("#24182c"),
		ContrastSecondaryTextColor:  tcell.GetColor("#03dac6"),
	},

	"matrix": {
		PrimitiveBackgroundColor:    tcell.GetColor("#000000"),
		ContrastBackgroundColor:     tcell.GetColor("#001100"),
		MoreContrastBackgroundColor: tcell.GetColor("#002200"),
		BorderColor:                 tcell.GetColor("#00ff00"),
		TitleColor:                  tcell.GetColor("#00ff00"),
		GraphicsColor:               tcell.GetColor("#00ff00"),
		PrimaryTextColor:            tcell.GetColor("#00ff00"),
		SecondaryTextColor:          tcell.GetColor("#00cc00"),
		TertiaryTextColor:           tcell.GetColor("#00ff88"),
		InverseTextColor:            tcell.GetColor("#000000"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ff00"),
	},

	"cyberpunk": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0a18"),
		ContrastBackgroundColor:     tcell.GetColor("#141428"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e1e38"),
		BorderColor:                 tcell.GetColor("#ff00ff"),
		TitleColor:                  tcell.GetColor("#00ffff"),
		GraphicsColor:               tcell.GetColor("#ffff00"),
		PrimaryTextColor:            tcell.GetColor("#ffffff"),
		SecondaryTextColor:          tcell.GetColor("#ff00ff"),
		TertiaryTextColor:           tcell.GetColor("#00ff00"),
		InverseTextColor:            tcell.GetColor("#141428"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ffff"),
	},

	"synthwave": {
		PrimitiveBackgroundColor:    tcell.GetColor("#241b3f"),
		ContrastBackgroundColor:     tcell.GetColor("#2d224c"),
		MoreContrastBackgroundColor: tcell.GetColor("#362959"),
		BorderColor:                 tcell.GetColor("#ff00ff"),
		TitleColor:                  tcell.GetColor("#00ffff"),
		GraphicsColor:               tcell.GetColor("#ff5500"),
		PrimaryTextColor:            tcell.GetColor("#ffffff"),
		SecondaryTextColor:          tcell.GetColor("#ff00ff"),
		TertiaryTextColor:           tcell.GetColor("#ffff00"),
		InverseTextColor:            tcell.GetColor("#2d224c"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ffff"),
	},

	"oceanic": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0c1c2c"),
		ContrastBackgroundColor:     tcell.GetColor("#142838"),
		MoreContrastBackgroundColor: tcell.GetColor("#1c3444"),
		BorderColor:                 tcell.GetColor("#66b3ff"),
		TitleColor:                  tcell.GetColor("#4dd2ff"),
		GraphicsColor:               tcell.GetColor("#ff6666"),
		PrimaryTextColor:            tcell.GetColor("#e0f0ff"),
		SecondaryTextColor:          tcell.GetColor("#66b3ff"),
		TertiaryTextColor:           tcell.GetColor("#ffcc66"),
		InverseTextColor:            tcell.GetColor("#142838"),
		ContrastSecondaryTextColor:  tcell.GetColor("#66ffb3"),
	},

	"forest": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1a251c"),
		ContrastBackgroundColor:     tcell.GetColor("#243028"),
		MoreContrastBackgroundColor: tcell.GetColor("#2e3b34"),
		BorderColor:                 tcell.GetColor("#6a994e"),
		TitleColor:                  tcell.GetColor("#84a98c"),
		GraphicsColor:               tcell.GetColor("#bc4749"),
		PrimaryTextColor:            tcell.GetColor("#dad7cd"),
		SecondaryTextColor:          tcell.GetColor("#a7c957"),
		TertiaryTextColor:           tcell.GetColor("#ee9b00"),
		InverseTextColor:            tcell.GetColor("#243028"),
		ContrastSecondaryTextColor:  tcell.GetColor("#90be6d"),
	},

	"fire": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1c0c0c"),
		ContrastBackgroundColor:     tcell.GetColor("#281414"),
		MoreContrastBackgroundColor: tcell.GetColor("#341c1c"),
		BorderColor:                 tcell.GetColor("#ff4400"),
		TitleColor:                  tcell.GetColor("#ff8800"),
		GraphicsColor:               tcell.GetColor("#ff2200"),
		PrimaryTextColor:            tcell.GetColor("#ffddcc"),
		SecondaryTextColor:          tcell.GetColor("#ff6600"),
		TertiaryTextColor:           tcell.GetColor("#ffaa00"),
		InverseTextColor:            tcell.GetColor("#281414"),
		ContrastSecondaryTextColor:  tcell.GetColor("#ffff00"),
	},

	"midnight-purple": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0a1a"),
		ContrastBackgroundColor:     tcell.GetColor("#14142a"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e1e3a"),
		BorderColor:                 tcell.GetColor("#9d4edd"),
		TitleColor:                  tcell.GetColor("#c77dff"),
		GraphicsColor:               tcell.GetColor("#ff6d6d"),
		PrimaryTextColor:            tcell.GetColor("#e0d6ff"),
		SecondaryTextColor:          tcell.GetColor("#bb86fc"),
		TertiaryTextColor:           tcell.GetColor("#ffb74d"),
		InverseTextColor:            tcell.GetColor("#14142a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#03dac6"),
	},

	"zenburn": {
		PrimitiveBackgroundColor:    tcell.GetColor("#3f3f3f"),
		ContrastBackgroundColor:     tcell.GetColor("#4f4f4f"),
		MoreContrastBackgroundColor: tcell.GetColor("#5f5f5f"),
		BorderColor:                 tcell.GetColor("#7cb8bb"),
		TitleColor:                  tcell.GetColor("#8cd0d3"),
		GraphicsColor:               tcell.GetColor("#cc9393"),
		PrimaryTextColor:            tcell.GetColor("#dcdccc"),
		SecondaryTextColor:          tcell.GetColor("#7cb8bb"),
		TertiaryTextColor:           tcell.GetColor("#f0dfaf"),
		InverseTextColor:            tcell.GetColor("#4f4f4f"),
		ContrastSecondaryTextColor:  tcell.GetColor("#bfebbf"),
	},

	// LIGHT THEMES
	"github-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#ffffff"),
		ContrastBackgroundColor:     tcell.GetColor("#f6f8fa"),
		MoreContrastBackgroundColor: tcell.GetColor("#eaeef2"),
		BorderColor:                 tcell.GetColor("#d0d7de"),
		TitleColor:                  tcell.GetColor("#0969da"),
		GraphicsColor:               tcell.GetColor("#cf222e"),
		PrimaryTextColor:            tcell.GetColor("#1f2328"),
		SecondaryTextColor:          tcell.GetColor("#656d76"),
		TertiaryTextColor:           tcell.GetColor("#e36c00"),
		InverseTextColor:            tcell.GetColor("#ffffff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#1a7f37"),
	},

	"one-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#fafafa"),
		ContrastBackgroundColor:     tcell.GetColor("#f0f0f0"),
		MoreContrastBackgroundColor: tcell.GetColor("#e6e6e6"),
		BorderColor:                 tcell.GetColor("#383a42"),
		TitleColor:                  tcell.GetColor("#4078f2"),
		GraphicsColor:               tcell.GetColor("#e45649"),
		PrimaryTextColor:            tcell.GetColor("#383a42"),
		SecondaryTextColor:          tcell.GetColor("#a0a1a7"),
		TertiaryTextColor:           tcell.GetColor("#c18401"),
		InverseTextColor:            tcell.GetColor("#f0f0f0"),
		ContrastSecondaryTextColor:  tcell.GetColor("#50a14f"),
	},

	"ayu-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#fcfcfc"),
		ContrastBackgroundColor:     tcell.GetColor("#f3f4f5"),
		MoreContrastBackgroundColor: tcell.GetColor("#e6e8eb"),
		BorderColor:                 tcell.GetColor("#5ccfe6"),
		TitleColor:                  tcell.GetColor("#ffae57"),
		GraphicsColor:               tcell.GetColor("#f07171"),
		PrimaryTextColor:            tcell.GetColor("#5c6773"),
		SecondaryTextColor:          tcell.GetColor("#a37acc"),
		TertiaryTextColor:           tcell.GetColor("#86b300"),
		InverseTextColor:            tcell.GetColor("#f3f4f5"),
		ContrastSecondaryTextColor:  tcell.GetColor("#aad94c"),
	},

	"paper": {
		PrimitiveBackgroundColor:    tcell.GetColor("#f5f5f5"),
		ContrastBackgroundColor:     tcell.GetColor("#eeeeee"),
		MoreContrastBackgroundColor: tcell.GetColor("#e0e0e0"),
		BorderColor:                 tcell.GetColor("#757575"),
		TitleColor:                  tcell.GetColor("#212121"),
		GraphicsColor:               tcell.GetColor("#f44336"),
		PrimaryTextColor:            tcell.GetColor("#424242"),
		SecondaryTextColor:          tcell.GetColor("#616161"),
		TertiaryTextColor:           tcell.GetColor("#ff9800"),
		InverseTextColor:            tcell.GetColor("#eeeeee"),
		ContrastSecondaryTextColor:  tcell.GetColor("#4caf50"),
	},

	"minimal-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#ffffff"),
		ContrastBackgroundColor:     tcell.GetColor("#f8f8f8"),
		MoreContrastBackgroundColor: tcell.GetColor("#f0f0f0"),
		BorderColor:                 tcell.GetColor("#dddddd"),
		TitleColor:                  tcell.GetColor("#000000"),
		GraphicsColor:               tcell.GetColor("#666666"),
		PrimaryTextColor:            tcell.GetColor("#333333"),
		SecondaryTextColor:          tcell.GetColor("#666666"),
		TertiaryTextColor:           tcell.GetColor("#999999"),
		InverseTextColor:            tcell.GetColor("#ffffff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#888888"),
	},

	"terminal": {
		PrimitiveBackgroundColor:    tcell.GetColor("#000000"),
		ContrastBackgroundColor:     tcell.GetColor("#0a0a0a"),
		MoreContrastBackgroundColor: tcell.GetColor("#1a1a1a"),
		BorderColor:                 tcell.GetColor("#00ff00"),
		TitleColor:                  tcell.GetColor("#00ff00"),
		GraphicsColor:               tcell.GetColor("#00ff00"),
		PrimaryTextColor:            tcell.GetColor("#00ff00"),
		SecondaryTextColor:          tcell.GetColor("#00aa00"),
		TertiaryTextColor:           tcell.GetColor("#ffff00"),
		InverseTextColor:            tcell.GetColor("#000000"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ff00"),
	},

	"windows-95": {
		PrimitiveBackgroundColor:    tcell.GetColor("#008080"),
		ContrastBackgroundColor:     tcell.GetColor("#c0c0c0"),
		MoreContrastBackgroundColor: tcell.GetColor("#ffffff"),
		BorderColor:                 tcell.GetColor("#000000"),
		TitleColor:                  tcell.GetColor("#000000"),
		GraphicsColor:               tcell.GetColor("#000000"),
		PrimaryTextColor:            tcell.GetColor("#000000"),
		SecondaryTextColor:          tcell.GetColor("#808080"),
		TertiaryTextColor:           tcell.GetColor("#000080"),
		InverseTextColor:            tcell.GetColor("#c0c0c0"),
		ContrastSecondaryTextColor:  tcell.GetColor("#008000"),
	},

	// MONOCHROME & SPECIAL THEMES
	"monochrome": {
		PrimitiveBackgroundColor:    tcell.GetColor("#000000"),
		ContrastBackgroundColor:     tcell.GetColor("#1a1a1a"),
		MoreContrastBackgroundColor: tcell.GetColor("#333333"),
		BorderColor:                 tcell.GetColor("#666666"),
		TitleColor:                  tcell.GetColor("#ffffff"),
		GraphicsColor:               tcell.GetColor("#999999"),
		PrimaryTextColor:            tcell.GetColor("#cccccc"),
		SecondaryTextColor:          tcell.GetColor("#999999"),
		TertiaryTextColor:           tcell.GetColor("#666666"),
		InverseTextColor:            tcell.GetColor("#000000"),
		ContrastSecondaryTextColor:  tcell.GetColor("#b3b3b3"),
	},

	"monochrome-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#ffffff"),
		ContrastBackgroundColor:     tcell.GetColor("#f0f0f0"),
		MoreContrastBackgroundColor: tcell.GetColor("#e0e0e0"),
		BorderColor:                 tcell.GetColor("#333333"),
		TitleColor:                  tcell.GetColor("#000000"),
		GraphicsColor:               tcell.GetColor("#666666"),
		PrimaryTextColor:            tcell.GetColor("#000000"),
		SecondaryTextColor:          tcell.GetColor("#333333"),
		TertiaryTextColor:           tcell.GetColor("#666666"),
		InverseTextColor:            tcell.GetColor("#ffffff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#999999"),
	},

	"sepia": {
		PrimitiveBackgroundColor:    tcell.GetColor("#f4ecd8"),
		ContrastBackgroundColor:     tcell.GetColor("#e9ddc7"),
		MoreContrastBackgroundColor: tcell.GetColor("#decbb6"),
		BorderColor:                 tcell.GetColor("#8b7355"),
		TitleColor:                  tcell.GetColor("#5d4c3c"),
		GraphicsColor:               tcell.GetColor("#8b4513"),
		PrimaryTextColor:            tcell.GetColor("#5d4c3c"),
		SecondaryTextColor:          tcell.GetColor("#8b7355"),
		TertiaryTextColor:           tcell.GetColor("#a0522d"),
		InverseTextColor:            tcell.GetColor("#e9ddc7"),
		ContrastSecondaryTextColor:  tcell.GetColor("#228b22"),
	},

	"red-dark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#1a0000"),
		ContrastBackgroundColor:     tcell.GetColor("#260000"),
		MoreContrastBackgroundColor: tcell.GetColor("#330000"),
		BorderColor:                 tcell.GetColor("#ff3333"),
		TitleColor:                  tcell.GetColor("#ff6666"),
		GraphicsColor:               tcell.GetColor("#ff0000"),
		PrimaryTextColor:            tcell.GetColor("#ffcccc"),
		SecondaryTextColor:          tcell.GetColor("#ff6666"),
		TertiaryTextColor:           tcell.GetColor("#ff9933"),
		InverseTextColor:            tcell.GetColor("#260000"),
		ContrastSecondaryTextColor:  tcell.GetColor("#ff3333"),
	},

	"green-dark": {
		PrimitiveBackgroundColor:    tcell.GetColor("#001a00"),
		ContrastBackgroundColor:     tcell.GetColor("#002600"),
		MoreContrastBackgroundColor: tcell.GetColor("#003300"),
		BorderColor:                 tcell.GetColor("#33ff33"),
		TitleColor:                  tcell.GetColor("#66ff66"),
		GraphicsColor:               tcell.GetColor("#00ff00"),
		PrimaryTextColor:            tcell.GetColor("#ccffcc"),
		SecondaryTextColor:          tcell.GetColor("#66ff66"),
		TertiaryTextColor:           tcell.GetColor("#ffff33"),
		InverseTextColor:            tcell.GetColor("#002600"),
		ContrastSecondaryTextColor:  tcell.GetColor("#33ff33"),
	},

	"blue-light": {
		PrimitiveBackgroundColor:    tcell.GetColor("#f0f8ff"),
		ContrastBackgroundColor:     tcell.GetColor("#e6f2ff"),
		MoreContrastBackgroundColor: tcell.GetColor("#d9ebff"),
		BorderColor:                 tcell.GetColor("#0066cc"),
		TitleColor:                  tcell.GetColor("#004080"),
		GraphicsColor:               tcell.GetColor("#cc0000"),
		PrimaryTextColor:            tcell.GetColor("#003366"),
		SecondaryTextColor:          tcell.GetColor("#0066cc"),
		TertiaryTextColor:           tcell.GetColor("#cc7a00"),
		InverseTextColor:            tcell.GetColor("#e6f2ff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#008000"),
	},

	"pastel": {
		PrimitiveBackgroundColor:    tcell.GetColor("#fff0f5"),
		ContrastBackgroundColor:     tcell.GetColor("#f8f8ff"),
		MoreContrastBackgroundColor: tcell.GetColor("#f0fff0"),
		BorderColor:                 tcell.GetColor("#dda0dd"),
		TitleColor:                  tcell.GetColor("#9370db"),
		GraphicsColor:               tcell.GetColor("#ff69b4"),
		PrimaryTextColor:            tcell.GetColor("#696969"),
		SecondaryTextColor:          tcell.GetColor("#ba55d3"),
		TertiaryTextColor:           tcell.GetColor("#ffa500"),
		InverseTextColor:            tcell.GetColor("#f8f8ff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#90ee90"),
	},

	"retro": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0f0f23"),
		ContrastBackgroundColor:     tcell.GetColor("#1c1c34"),
		MoreContrastBackgroundColor: tcell.GetColor("#292945"),
		BorderColor:                 tcell.GetColor("#00ff9f"),
		TitleColor:                  tcell.GetColor("#ff6b9f"),
		GraphicsColor:               tcell.GetColor("#ffff00"),
		PrimaryTextColor:            tcell.GetColor("#ffffff"),
		SecondaryTextColor:          tcell.GetColor("#00ff9f"),
		TertiaryTextColor:           tcell.GetColor("#ff6b9f"),
		InverseTextColor:            tcell.GetColor("#1c1c34"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ffff"),
	},

	"steampunk": {
		PrimitiveBackgroundColor:    tcell.GetColor("#3c2f2f"),
		ContrastBackgroundColor:     tcell.GetColor("#4b3a3a"),
		MoreContrastBackgroundColor: tcell.GetColor("#5a4545"),
		BorderColor:                 tcell.GetColor("#b87333"),
		TitleColor:                  tcell.GetColor("#d4af37"),
		GraphicsColor:               tcell.GetColor("#8b0000"),
		PrimaryTextColor:            tcell.GetColor("#f5f5dc"),
		SecondaryTextColor:          tcell.GetColor("#b87333"),
		TertiaryTextColor:           tcell.GetColor("#d4af37"),
		InverseTextColor:            tcell.GetColor("#4b3a3a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#808000"),
	},

	"hacker": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0a0a"),
		ContrastBackgroundColor:     tcell.GetColor("#121212"),
		MoreContrastBackgroundColor: tcell.GetColor("#1a1a1a"),
		BorderColor:                 tcell.GetColor("#00ff00"),
		TitleColor:                  tcell.GetColor("#00ff00"),
		GraphicsColor:               tcell.GetColor("#ff0000"),
		PrimaryTextColor:            tcell.GetColor("#00ff00"),
		SecondaryTextColor:          tcell.GetColor("#00cc00"),
		TertiaryTextColor:           tcell.GetColor("#ffff00"),
		InverseTextColor:            tcell.GetColor("#0a0a0a"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ff00"),
	},

	"nebula": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0014"),
		ContrastBackgroundColor:     tcell.GetColor("#14001e"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e0028"),
		BorderColor:                 tcell.GetColor("#9d00ff"),
		TitleColor:                  tcell.GetColor("#00e5ff"),
		GraphicsColor:               tcell.GetColor("#ff006e"),
		PrimaryTextColor:            tcell.GetColor("#e0c3ff"),
		SecondaryTextColor:          tcell.GetColor("#9d00ff"),
		TertiaryTextColor:           tcell.GetColor("#ffbe0b"),
		InverseTextColor:            tcell.GetColor("#14001e"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ff88"),
	},

	"galaxy": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0a1e"),
		ContrastBackgroundColor:     tcell.GetColor("#141428"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e1e32"),
		BorderColor:                 tcell.GetColor("#6a00f4"),
		TitleColor:                  tcell.GetColor("#00d4ff"),
		GraphicsColor:               tcell.GetColor("#ff0055"),
		PrimaryTextColor:            tcell.GetColor("#d4d4ff"),
		SecondaryTextColor:          tcell.GetColor("#9370db"),
		TertiaryTextColor:           tcell.GetColor("#ffaa00"),
		InverseTextColor:            tcell.GetColor("#141428"),
		ContrastSecondaryTextColor:  tcell.GetColor("#00ffaa"),
	},

	"midnight": {
		PrimitiveBackgroundColor:    tcell.GetColor("#0a0a14"),
		ContrastBackgroundColor:     tcell.GetColor("#14141e"),
		MoreContrastBackgroundColor: tcell.GetColor("#1e1e28"),
		BorderColor:                 tcell.GetColor("#4a4aff"),
		TitleColor:                  tcell.GetColor("#00ffff"),
		GraphicsColor:               tcell.GetColor("#ff4a4a"),
		PrimaryTextColor:            tcell.GetColor("#e0e0ff"),
		SecondaryTextColor:          tcell.GetColor("#8888ff"),
		TertiaryTextColor:           tcell.GetColor("#ffff4a"),
		InverseTextColor:            tcell.GetColor("#14141e"),
		ContrastSecondaryTextColor:  tcell.GetColor("#4aff4a"),
	},

	"arctic": {
		PrimitiveBackgroundColor:    tcell.GetColor("#e6f7ff"),
		ContrastBackgroundColor:     tcell.GetColor("#d9f2ff"),
		MoreContrastBackgroundColor: tcell.GetColor("#cceeff"),
		BorderColor:                 tcell.GetColor("#3399ff"),
		TitleColor:                  tcell.GetColor("#0066cc"),
		GraphicsColor:               tcell.GetColor("#ff3366"),
		PrimaryTextColor:            tcell.GetColor("#003366"),
		SecondaryTextColor:          tcell.GetColor("#3399ff"),
		TertiaryTextColor:           tcell.GetColor("#ff9933"),
		InverseTextColor:            tcell.GetColor("#d9f2ff"),
		ContrastSecondaryTextColor:  tcell.GetColor("#33cc33"),
	},

	"desert": {
		PrimitiveBackgroundColor:    tcell.GetColor("#f4e7d3"),
		ContrastBackgroundColor:     tcell.GetColor("#e8d4b9"),
		MoreContrastBackgroundColor: tcell.GetColor("#dcc19f"),
		BorderColor:                 tcell.GetColor("#8b7355"),
		TitleColor:                  tcell.GetColor("#a0522d"),
		GraphicsColor:               tcell.GetColor("#8b0000"),
		PrimaryTextColor:            tcell.GetColor("#5d4037"),
		SecondaryTextColor:          tcell.GetColor("#8b7355"),
		TertiaryTextColor:           tcell.GetColor("#d2691e"),
		InverseTextColor:            tcell.GetColor("#e8d4b9"),
		ContrastSecondaryTextColor:  tcell.GetColor("#228b22"),
	},
}

var primitives []tview.Primitive

// Resolve returns the theme by name. Unknown names fall back to default; "custom" applies overrides atop default.
func Resolve(tc *config.ThemeConfig) tview.Theme {
	name := strings.ToLower(strings.TrimSpace(config.DefaultThemeName))
	if tc != nil {
		if n := strings.TrimSpace(tc.Name); n != "" {
			name = strings.ToLower(n)
		}
	}

	base, ok := registry[name]
	if name == config.CustomThemeName {
		defaultTheme := registry[config.DefaultThemeName]
		base = applyOverrides(&defaultTheme, tc)
	} else if !ok {
		log.Warn("theme not found, falling back to default", "name", name)
		base = registry[config.DefaultThemeName]
	}

	tview.Styles = base
	return base
}

// applyOverrides starts from base and applies overrides from config.
func applyOverrides(base *tview.Theme, tc *config.ThemeConfig) tview.Theme {
	if base == nil {
		return registry[config.DefaultThemeName]
	}

	th := *base
	if tc == nil {
		return th
	}

	if c := parseColor(tc.PrimitiveBackgroundColor); c != nil {
		th.PrimitiveBackgroundColor = *c
	}
	if c := parseColor(tc.ContrastBackgroundColor); c != nil {
		th.ContrastBackgroundColor = *c
	}
	if c := parseColor(tc.MoreContrastBackgroundColor); c != nil {
		th.MoreContrastBackgroundColor = *c
	}
	if c := parseColor(tc.BorderColor); c != nil {
		th.BorderColor = *c
	}
	if c := parseColor(tc.TitleColor); c != nil {
		th.TitleColor = *c
	}
	if c := parseColor(tc.GraphicsColor); c != nil {
		th.GraphicsColor = *c
	}
	if c := parseColor(tc.PrimaryTextColor); c != nil {
		th.PrimaryTextColor = *c
	}
	if c := parseColor(tc.SecondaryTextColor); c != nil {
		th.SecondaryTextColor = *c
	}
	if c := parseColor(tc.TertiaryTextColor); c != nil {
		th.TertiaryTextColor = *c
	}
	if c := parseColor(tc.InverseTextColor); c != nil {
		th.InverseTextColor = *c
	}
	if c := parseColor(tc.ContrastSecondaryTextColor); c != nil {
		th.ContrastSecondaryTextColor = *c
	}

	return th
}

// helper to transform user defined color strings into tcell.Color pointers.
func parseColor(s string) *tcell.Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	c := tcell.GetColor(s)
	return &c
}

// Names returns the currently registered theme names (built-ins plus any custom registrations).
func Names() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	names = append(names, "custom")
	sort.Strings(names)
	return names
}

// RegisterPrimitive registers a primitive for theme updates.
func RegisterPrimitive(p tview.Primitive) {
	ApplyToPrimitive(p)
	primitives = append(primitives, p)
}

// ApplyThemeToAllRegisteredPrimitives applies the current theme to all registered primitives.
func ApplyThemeToAllRegisteredPrimitives() {
	for _, p := range primitives {
		ApplyToPrimitive(p)
	}
}

// SaveToConfig saves the current theme to the config file.
func SaveToConfig(themeName string, cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	cfg.Theme.Name = themeName

	if err := config.Save(cfg, ""); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ApplyToPrimitive applies theme colors to any tview primitive.
// Supports all official tview primitives with proper styling methods.
func ApplyToPrimitive(p tview.Primitive) {
	if p == nil {
		return
	}

	switch v := p.(type) {
	case *tview.TextView:
		v.SetTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.TextArea:
		v.SetTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.PrimaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Table:
		v.SetBordersColor(tview.Styles.BorderColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		if noColor {
			if tview.Styles.PrimaryTextColor == tcell.ColorDefault {
				selectedStyle := tcell.StyleDefault.Reverse(true)
				v.SetSelectedStyle(selectedStyle)
			}
		}

	case *tview.TreeView:
		v.SetGraphicsColor(tview.Styles.GraphicsColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.List:
		v.SetMainTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.PrimaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetSecondaryTextStyle(tcell.StyleDefault.
			Foreground(tview.Styles.SecondaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetShortcutStyle(tcell.StyleDefault.
			Foreground(tview.Styles.TertiaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetSelectedStyle(tcell.StyleDefault.
			Foreground(tview.Styles.InverseTextColor).
			Background(tview.Styles.SecondaryTextColor))
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)
		if noColor {
			selectedStyle := tcell.StyleDefault.Reverse(true)
			v.SetSelectedStyle(selectedStyle)
		}

	case *tview.InputField:
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.DropDown:
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Checkbox:
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Image:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Button:
		v.SetLabelColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Form:
		v.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetFieldTextColor(tview.Styles.PrimaryTextColor)
		v.SetLabelColor(tview.Styles.SecondaryTextColor)
		v.SetButtonBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetButtonTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Modal:
		v.SetTextColor(tview.Styles.PrimaryTextColor)
		v.SetButtonStyle(tcell.StyleDefault.
			Foreground(tview.Styles.PrimaryTextColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetButtonTextColor(tview.Styles.PrimaryTextColor)
		v.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		v.SetBorderStyle(tcell.StyleDefault.
			Foreground(tview.Styles.BorderColor).
			Background(tview.Styles.PrimitiveBackgroundColor))
		v.SetTitleColor(tview.Styles.TitleColor)
		if noColor {
			buttonStyle := tcell.StyleDefault.Reverse(true)
			v.SetButtonActivatedStyle(buttonStyle)
		}

	case *tview.Grid:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBordersColor(tview.Styles.BorderColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Flex:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	case *tview.Pages:
		v.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		v.SetBorderColor(tview.Styles.BorderColor)
		v.SetTitleColor(tview.Styles.TitleColor)

	default:
		if box, ok := p.(interface{ SetBackgroundColor(tcell.Color) *tview.Box }); ok {
			box.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		}
	}
}

// NoColorTheme returns a theme with all colors set to tcell.ColorDefault to disable ANSI colors.
// so NO_COLOR is supported properly
// see https://no-color.org/
// see https://github.com/rivo/tview/issues/859#issuecomment-1848992716
func NoColorTheme() tview.Theme {
	return tview.Theme{}
}

// TviewDefaultTheme returns the default tview theme.
// in case the users has themes disabled, this will take terminal default ANDI colors.
func TviewDefaultTheme() tview.Theme {
	return tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,
		ContrastBackgroundColor:     tcell.ColorBlue,
		MoreContrastBackgroundColor: tcell.ColorGreen,
		BorderColor:                 tcell.ColorWhite,
		TitleColor:                  tcell.ColorWhite,
		GraphicsColor:               tcell.ColorWhite,
		PrimaryTextColor:            tcell.ColorWhite,
		SecondaryTextColor:          tcell.ColorYellow,
		TertiaryTextColor:           tcell.ColorGreen,
		InverseTextColor:            tcell.ColorBlue,
		ContrastSecondaryTextColor:  tcell.ColorNavy,
	}
}

// IsNoColor returns whether NO_COLOR is set.
func IsNoColor() bool {
	return noColor
}
