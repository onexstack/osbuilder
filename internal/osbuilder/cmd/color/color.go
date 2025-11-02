// Package color print colors supported by the current terminal.
package color

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"
	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// ColorOptions is an options struct to support color subcommands.
type ColorOptions struct {
	Type        []string // color types to display
	Example     bool     // show code examples
	Format      string   // output format: table, simple, demo
	NoColor     bool     // disable colored output
	Interactive bool     // interactive mode
	Test        bool     // test terminal color support

	genericiooptions.IOStreams
}

// ColorInfo represents color information.
type ColorInfo struct {
	Category    string
	Name        string
	Description string
	Code        string
	Function    string
	Example     string
}

var (
	colorLong = templates.LongDesc(`
		Print the colors supported by the current terminal and demonstrate color usage.

		Color lets you use colorized outputs in terms of ANSI Escape Codes in Go (Golang). 
		It has support for Windows too! The API can be used in several ways, pick one that suits you.

		This command helps you:
		- Explore available colors and text effects
		- Learn color usage with practical examples
		- Test color support in your terminal
		- Generate color code snippets for development
		- Validate color combinations and readability

		Find more information at: https://github.com/fatih/color
	`)

	colorExample = templates.Examples(`
		# Print default foreground and background colors
		osbuilder color

		# Print specific color types
		osbuilder color -t fg-hi,bg

		# Print all supported colors and effects
		osbuilder color -t all

		# Show usage examples and code snippets
		osbuilder color --example

		# Simple output format without table
		osbuilder color --format simple

		# Interactive color demonstration
		osbuilder color --interactive

		# Test terminal color support
		osbuilder color --test

		# Demo format with color combinations
		osbuilder color --format demo
	`)

	availableTypes = []string{"fg", "fg-hi", "bg", "bg-hi", "effects", "all"}
	outputFormats  = []string{"table", "simple", "demo"}

	colorCodeExample = templates.Examples(`
### 1. Standard Colors

// Import color package
import "github.com/fatih/color"

// Print with default helper functions
color.Cyan("Prints text in cyan.")
color.Blue("Prints %s in blue.", "text")

// These are using the default foreground colors
color.Red("We have red")
color.Magenta("And many others ..")

### 2. Mix and Reuse Colors

// Create a new color object
c := color.New(color.FgCyan).Add(color.Underline)
c.Println("Prints cyan text with an underline.")

// Or just add them to New()
d := color.New(color.FgCyan, color.Bold)
d.Printf("This prints bold cyan %s\n", "too!")

// Mix up foreground and background colors, create new mixes!
red := color.New(color.FgRed)

boldRed := red.Add(color.Bold)
boldRed.Println("This will print text in bold red.")

whiteBackground := red.Add(color.BgWhite)
whiteBackground.Println("Red text with white background.")

### 3. Use Your Own Output (io.Writer)

// Use your own io.Writer output
color.New(color.FgBlue).Fprintln(myWriter, "blue color!")

blue := color.New(color.FgBlue)
blue.Fprint(writer, "This will print text in blue.")

### 4. Custom Print Functions (PrintFunc)

// Create a custom print function for convenience
red := color.New(color.FgRed).PrintfFunc()
red("Warning")
red("Error: %s", err)

// Mix up multiple attributes
notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
notice("Don't forget this...")

### 5. Custom Fprint Functions (FprintFunc)

blue := color.New(color.FgBlue).FprintfFunc()
blue(myWriter, "important notice: %s", stars)

// Mix up with multiple attributes
success := color.New(color.Bold, color.FgGreen).FprintlnFunc()
success(myWriter, "Don't forget this...")

### 6. Insert Into Non-color Strings (SprintFunc)

// Create SprintXxx functions to mix strings with other non-colorized strings:
yellow := color.New(color.FgYellow).SprintFunc()
red := color.New(color.FgRed).SprintFunc()
fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))

info := color.New(color.FgWhite, color.BgGreen).SprintFunc()
fmt.Printf("This %s rocks!\n", info("package"))

// Use helper functions
fmt.Println("This", color.RedString("warning"), "should be not neglected.")
fmt.Printf("%v %v\n", color.GreenString("Info:"), "an important message.")

// Windows supported too! Just don't forget to change the output to color.Output
fmt.Fprintf(color.Output, "Windows support: %s", color.GreenString("PASS"))

### 7. Plug Into Existing Code

// Use handy standard colors
color.Set(color.FgYellow)

fmt.Println("Existing text will now be in yellow")
fmt.Printf("This one %s\n", "too")

color.Unset() // Don't forget to unset

// You can mix up parameters
color.Set(color.FgMagenta, color.Bold)
defer color.Unset() // Use it in your function

fmt.Println("All text will now be bold magenta.")

### 8. Disable/Enable Color
 
// Disable color output globally
var flagNoColor = flag.Bool("no-color", false, "Disable color output")

if *flagNoColor {
    color.NoColor = true // disables colorized output
}

// Disable/enable color output on the fly:
c := color.New(color.FgCyan)
c.Println("Prints cyan text")

c.DisableColor()
c.Println("This is printed without any color")

c.EnableColor()
c.Println("This prints again cyan...")

### 9. Practical Color Schemes

// Success, Warning, Error patterns
var (
    Success = color.New(color.FgGreen, color.Bold).SprintFunc()
    Warning = color.New(color.FgYellow, color.Bold).SprintFunc()
    Error   = color.New(color.FgRed, color.Bold).SprintFunc()
    Info    = color.New(color.FgCyan).SprintFunc()
)

fmt.Printf("%s Operation completed successfully\n", Success("✓"))
fmt.Printf("%s This is a warning message\n", Warning("⚠"))
fmt.Printf("%s An error occurred\n", Error("✗"))
fmt.Printf("%s Additional information\n", Info("ℹ"))

### 10. Terminal Capability Detection

// Check if colors are supported
if color.NoColor {
    fmt.Println("Colors are disabled")
} else {
    fmt.Println("Colors are supported")
}
`)
)

// NewColorOptions returns an initialized ColorOptions instance.
func NewColorOptions(ioStreams genericiooptions.IOStreams) *ColorOptions {
	return &ColorOptions{
		Type:        []string{},
		Example:     false,
		Format:      "table",
		NoColor:     false,
		Interactive: false,
		Test:        false,
		IOStreams:   ioStreams,
	}
}

// NewCmdColor returns new initialized instance of color sub command.
func NewCmdColor(f cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	o := NewColorOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "color",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"colours", "colors"},
		Short:                 "Print colors supported by the current terminal",
		TraverseChildren:      true,
		Long:                  colorLong,
		Example:               colorExample,
		ValidArgs:             availableTypes,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run(args))
		},
		SuggestFor: []string{"colours", "colors", "terminal", "ansi", "style"},
	}

	cmd.Flags().StringSliceVarP(&o.Type, "type", "t", o.Type,
		fmt.Sprintf("Specify the color type, available types: %s", strings.Join(availableTypes, ", ")))
	cmd.Flags().BoolVar(&o.Example, "example", o.Example, "Print code examples showing how to use color package")
	cmd.Flags().StringVar(&o.Format, "format", o.Format,
		fmt.Sprintf("Output format (%s)", strings.Join(outputFormats, ", ")))
	cmd.Flags().BoolVar(&o.NoColor, "no-color", o.NoColor, "Disable colored output (for testing)")
	cmd.Flags().BoolVar(&o.Interactive, "interactive", o.Interactive, "Interactive color demonstration")
	cmd.Flags().BoolVar(&o.Test, "test", o.Test, "Test terminal color support")

	return cmd
}

// Complete completes all the required options.
func (o *ColorOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	// Set default types if none specified
	if len(o.Type) == 0 {
		o.Type = []string{"fg", "bg"}
	}

	// Handle "all" type
	if stringsutil.StringIn("all", o.Type) {
		o.Type = []string{"fg", "fg-hi", "bg", "bg-hi", "effects"}
	}

	// Normalize format
	o.Format = strings.ToLower(o.Format)
	if !stringsutil.StringIn(o.Format, outputFormats) {
		o.Format = "table"
	}

	// Apply no-color setting
	if o.NoColor {
		color.NoColor = true
	}

	return nil
}

// Validate makes sure there is no discrepancy in command options.
func (o *ColorOptions) Validate(cmd *cobra.Command, args []string) error {
	// Validate color types
	for _, t := range o.Type {
		if !stringsutil.StringIn(t, availableTypes) {
			return cmdutil.UsageErrorf(cmd, "--type must be one of: %s", strings.Join(availableTypes, ", "))
		}
	}

	// Validate output format
	if !stringsutil.StringIn(o.Format, outputFormats) {
		return cmdutil.UsageErrorf(cmd, "--format must be one of: %s", strings.Join(outputFormats, ", "))
	}

	return nil
}

// Run executes a color subcommand using the specified options.
func (o *ColorOptions) Run(args []string) error {
	// Handle code examples
	if o.Example {
		return o.printCodeExamples()
	}

	// Handle terminal color support test
	if o.Test {
		return o.testColorSupport()
	}

	// Handle interactive mode
	if o.Interactive {
		return o.runInteractiveMode()
	}

	// Generate and display color data
	return o.displayColors()
}

// displayColors generates and displays color information.
func (o *ColorOptions) displayColors() error {
	colorData := o.generateColorData()

	switch o.Format {
	case "simple":
		return o.printSimple(colorData)
	case "demo":
		return o.printDemo(colorData)
	default: // table
		return o.printTable(colorData)
	}
}

// generateColorData generates color information based on selected types.
func (o *ColorOptions) generateColorData() []ColorInfo {
	var data []ColorInfo

	for _, t := range o.Type {
		switch t {
		case "fg":
			data = append(data, o.getForegroundColors()...)
		case "fg-hi":
			data = append(data, o.getHighIntensityForegroundColors()...)
		case "bg":
			data = append(data, o.getBackgroundColors()...)
		case "bg-hi":
			data = append(data, o.getHighIntensityBackgroundColors()...)
		case "effects":
			data = append(data, o.getEffects()...)
		}
	}

	// Sort by category and name
	sort.Slice(data, func(i, j int) bool {
		if data[i].Category != data[j].Category {
			return data[i].Category < data[j].Category
		}
		return data[i].Name < data[j].Name
	})

	return data
}

// getForegroundColors returns standard foreground colors.
func (o *ColorOptions) getForegroundColors() []ColorInfo {
	return []ColorInfo{
		{"fg", "black", "Standard black text", "color.FgBlack", "color.BlackString", color.BlackString("Sample text")},
		{"fg", "red", "Standard red text", "color.FgRed", "color.RedString", color.RedString("Sample text")},
		{"fg", "green", "Standard green text", "color.FgGreen", "color.GreenString", color.GreenString("Sample text")},
		{"fg", "yellow", "Standard yellow text", "color.FgYellow", "color.YellowString", color.YellowString("Sample text")},
		{"fg", "blue", "Standard blue text", "color.FgBlue", "color.BlueString", color.BlueString("Sample text")},
		{"fg", "magenta", "Standard magenta text", "color.FgMagenta", "color.MagentaString", color.MagentaString("Sample text")},
		{"fg", "cyan", "Standard cyan text", "color.FgCyan", "color.CyanString", color.CyanString("Sample text")},
		{"fg", "white", "Standard white text", "color.FgWhite", "color.WhiteString", color.WhiteString("Sample text")},
	}
}

// getHighIntensityForegroundColors returns high-intensity foreground colors.
func (o *ColorOptions) getHighIntensityForegroundColors() []ColorInfo {
	return []ColorInfo{
		{"fg-hi", "black", "High-intensity black text", "color.FgHiBlack", "color.HiBlackString", color.HiBlackString("Sample text")},
		{"fg-hi", "red", "High-intensity red text", "color.FgHiRed", "color.HiRedString", color.HiRedString("Sample text")},
		{"fg-hi", "green", "High-intensity green text", "color.FgHiGreen", "color.HiGreenString", color.HiGreenString("Sample text")},
		{"fg-hi", "yellow", "High-intensity yellow text", "color.FgHiYellow", "color.HiYellowString", color.HiYellowString("Sample text")},
		{"fg-hi", "blue", "High-intensity blue text", "color.FgHiBlue", "color.HiBlueString", color.HiBlueString("Sample text")},
		{"fg-hi", "magenta", "High-intensity magenta text", "color.FgHiMagenta", "color.HiMagentaString", color.HiMagentaString("Sample text")},
		{"fg-hi", "cyan", "High-intensity cyan text", "color.FgHiCyan", "color.HiCyanString", color.HiCyanString("Sample text")},
		{"fg-hi", "white", "High-intensity white text", "color.FgHiWhite", "color.HiWhiteString", color.HiWhiteString("Sample text")},
	}
}

// getBackgroundColors returns background colors.
func (o *ColorOptions) getBackgroundColors() []ColorInfo {
	return []ColorInfo{
		{"bg", "black", "Black background", "color.BgBlack", "color.New(color.FgWhite, color.BgBlack)", color.New(color.FgWhite, color.BgBlack).Sprint("Sample text")},
		{"bg", "red", "Red background", "color.BgRed", "color.New(color.FgWhite, color.BgRed)", color.New(color.FgWhite, color.BgRed).Sprint("Sample text")},
		{"bg", "green", "Green background", "color.BgGreen", "color.New(color.FgBlack, color.BgGreen)", color.New(color.FgBlack, color.BgGreen).Sprint("Sample text")},
		{"bg", "yellow", "Yellow background", "color.BgYellow", "color.New(color.FgBlack, color.BgYellow)", color.New(color.FgBlack, color.BgYellow).Sprint("Sample text")},
		{"bg", "blue", "Blue background", "color.BgBlue", "color.New(color.FgWhite, color.BgBlue)", color.New(color.FgWhite, color.BgBlue).Sprint("Sample text")},
		{"bg", "magenta", "Magenta background", "color.BgMagenta", "color.New(color.FgWhite, color.BgMagenta)", color.New(color.FgWhite, color.BgMagenta).Sprint("Sample text")},
		{"bg", "cyan", "Cyan background", "color.BgCyan", "color.New(color.FgBlack, color.BgCyan)", color.New(color.FgBlack, color.BgCyan).Sprint("Sample text")},
		{"bg", "white", "White background", "color.BgWhite", "color.New(color.FgBlack, color.BgWhite)", color.New(color.FgBlack, color.BgWhite).Sprint("Sample text")},
	}
}

// getHighIntensityBackgroundColors returns high-intensity background colors.
func (o *ColorOptions) getHighIntensityBackgroundColors() []ColorInfo {
	return []ColorInfo{
		{"bg-hi", "black", "High-intensity black background", "color.BgHiBlack", "color.New(color.FgWhite, color.BgHiBlack)", color.New(color.FgWhite, color.BgHiBlack).Sprint("Sample text")},
		{"bg-hi", "red", "High-intensity red background", "color.BgHiRed", "color.New(color.FgWhite, color.BgHiRed)", color.New(color.FgWhite, color.BgHiRed).Sprint("Sample text")},
		{"bg-hi", "green", "High-intensity green background", "color.BgHiGreen", "color.New(color.FgBlack, color.BgHiGreen)", color.New(color.FgBlack, color.BgHiGreen).Sprint("Sample text")},
		{"bg-hi", "yellow", "High-intensity yellow background", "color.BgHiYellow", "color.New(color.FgBlack, color.BgHiYellow)", color.New(color.FgBlack, color.BgHiYellow).Sprint("Sample text")},
		{"bg-hi", "blue", "High-intensity blue background", "color.BgHiBlue", "color.New(color.FgWhite, color.BgHiBlue)", color.New(color.FgWhite, color.BgHiBlue).Sprint("Sample text")},
		{"bg-hi", "magenta", "High-intensity magenta background", "color.BgHiMagenta", "color.New(color.FgWhite, color.BgHiMagenta)", color.New(color.FgWhite, color.BgHiMagenta).Sprint("Sample text")},
		{"bg-hi", "cyan", "High-intensity cyan background", "color.BgHiCyan", "color.New(color.FgBlack, color.BgHiCyan)", color.New(color.FgBlack, color.BgHiCyan).Sprint("Sample text")},
		{"bg-hi", "white", "High-intensity white background", "color.BgHiWhite", "color.New(color.FgBlack, color.BgHiWhite)", color.New(color.FgBlack, color.BgHiWhite).Sprint("Sample text")},
	}
}

// getEffects returns text effects and attributes.
func (o *ColorOptions) getEffects() []ColorInfo {
	return []ColorInfo{
		{"effects", "reset", "Reset all attributes", "color.Reset", "color.New(color.Reset)", color.New(color.FgGreen, color.Reset).Sprint("Sample text")},
		{"effects", "bold", "Bold text", "color.Bold", "color.New(color.Bold)", color.New(color.FgGreen, color.Bold).Sprint("Sample text")},
		{"effects", "faint", "Faint text", "color.Faint", "color.New(color.Faint)", color.New(color.FgGreen, color.Faint).Sprint("Sample text")},
		{"effects", "italic", "Italic text", "color.Italic", "color.New(color.Italic)", color.New(color.FgGreen, color.Italic).Sprint("Sample text")},
		{"effects", "underline", "Underlined text", "color.Underline", "color.New(color.Underline)", color.New(color.FgGreen, color.Underline).Sprint("Sample text")},
		{"effects", "blink-slow", "Slow blinking text", "color.BlinkSlow", "color.New(color.BlinkSlow)", color.New(color.FgGreen, color.BlinkSlow).Sprint("Sample text")},
		{"effects", "blink-rapid", "Rapid blinking text", "color.BlinkRapid", "color.New(color.BlinkRapid)", color.New(color.FgGreen, color.BlinkRapid).Sprint("Sample text")},
		{"effects", "reverse", "Reverse video", "color.ReverseVideo", "color.New(color.ReverseVideo)", color.New(color.FgGreen, color.ReverseVideo).Sprint("Sample text")},
		{"effects", "concealed", "Concealed text", "color.Concealed", "color.New(color.Concealed)", color.New(color.FgGreen, color.Concealed).Sprint("Sample text")},
		{"effects", "crossed-out", "Crossed out text", "color.CrossedOut", "color.New(color.CrossedOut)", color.New(color.FgGreen, color.CrossedOut).Sprint("Sample text")},
	}
}

// printTable prints colors in table format.
func (o *ColorOptions) printTable(data []ColorInfo) error {
	table := tablewriter.NewWriter(o.Out)

	// Configure table
	table.SetAutoMergeCells(true)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Category", "Color Name", "Description", "Code", "Example"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgRedColor},
	)

	// Convert to table data
	var tableData [][]string
	for _, info := range data {
		tableData = append(tableData, []string{
			info.Category,
			info.Name,
			info.Description,
			info.Code,
			info.Example,
		})
	}

	table.AppendBulk(tableData)
	table.Render()
	return nil
}

// printSimple prints colors in simple format.
func (o *ColorOptions) printSimple(data []ColorInfo) error {
	currentCategory := ""
	for _, info := range data {
		if info.Category != currentCategory {
			currentCategory = info.Category
			fmt.Fprintf(o.Out, "\n%s %s:\n", emoji.FileFolder, strings.ToUpper(currentCategory))
		}
		fmt.Fprintf(o.Out, "  %-15s %s\n", info.Name, info.Example)
	}
	return nil
}

// printDemo prints colors in demo format with combinations.
func (o *ColorOptions) printDemo(data []ColorInfo) error {
	fmt.Fprintf(o.Out, "%s Color Demonstration\n\n",
		color.New(color.Bold, color.FgCyan).Sprint(emoji.ArtistPalette))

	// Group by category
	categories := make(map[string][]ColorInfo)
	for _, info := range data {
		categories[info.Category] = append(categories[info.Category], info)
	}

	for category, infos := range categories {
		fmt.Fprintf(o.Out, "%s %s\n", emoji.FileFolder, 
			color.New(color.Bold, color.FgYellow).Sprintf("%s Colors", strings.Title(category)))
		fmt.Fprintf(o.Out, "%s\n", strings.Repeat("-", 50))

		for _, info := range infos {
			fmt.Fprintf(o.Out, "%-12s │ %s │ %s\n",
				info.Name,
				info.Example,
				color.New(color.Faint).Sprint(info.Code))
		}
		fmt.Fprintf(o.Out, "\n")
	}

	// Show practical combinations
	o.showColorCombinations()
	return nil
}

// showColorCombinations displays practical color combinations.
func (o *ColorOptions) showColorCombinations() {
	fmt.Fprintf(o.Out, "%s Practical Color Combinations\n",
		color.New(color.Bold, color.FgMagenta).Sprint(emoji.ArtistPalette))
	fmt.Fprintf(o.Out, "%s\n", strings.Repeat("-", 50))

	combinations := []struct {
		name   string
		create func() *color.Color
		usage  string
	}{
		{"Success", func() *color.Color { return color.New(color.FgGreen, color.Bold) }, "✓ Operation completed"},
		{"Warning", func() *color.Color { return color.New(color.FgYellow, color.Bold) }, "⚠ Warning message"},
		{"Error", func() *color.Color { return color.New(color.FgRed, color.Bold) }, "✗ Error occurred"},
		{"Info", func() *color.Color { return color.New(color.FgCyan, color.Bold) }, "ℹ Information"},
		{"Highlight", func() *color.Color { return color.New(color.FgBlack, color.BgYellow) }, "Important text"},
		{"Urgent", func() *color.Color { return color.New(color.FgWhite, color.BgRed, color.Bold) }, "URGENT MESSAGE"},
		{"Code", func() *color.Color { return color.New(color.FgGreen, color.BgBlack) }, "code snippet"},
		{"Link", func() *color.Color { return color.New(color.FgBlue, color.Underline) }, "https://example.com"},
	}

	for _, combo := range combinations {
		c := combo.create()
		fmt.Fprintf(o.Out, "%-12s │ %s │ %s\n",
			combo.name,
			c.Sprint(combo.usage),
			color.New(color.Faint).Sprint("Common use case"))
	}
	fmt.Fprintf(o.Out, "\n")
}

// printCodeExamples prints code usage examples.
func (o *ColorOptions) printCodeExamples() error {
	fmt.Fprintf(o.Out, "%s Color Package Usage Examples\n",
		color.New(color.Bold, color.FgCyan).Sprint(emoji.Laptop))
	fmt.Fprintf(o.Out, "%s\n\n", strings.Repeat("=", 60))

	fmt.Fprint(o.Out, colorCodeExample)
	return nil
}

// testColorSupport tests terminal color support.
func (o *ColorOptions) testColorSupport() error {
	fmt.Fprintf(o.Out, "%s Terminal Color Support Test\n",
		color.New(color.Bold, color.FgCyan).Sprint(emoji.Syringe))
	fmt.Fprintf(o.Out, "%s\n\n", strings.Repeat("=", 50))

	// Basic color support
	fmt.Fprintf(o.Out, "Basic Colors: ")
	if color.NoColor {
		fmt.Fprintf(o.Out, "%s Colors are disabled\n", "❌")
	} else {
		fmt.Fprintf(o.Out, "%s Colors are enabled\n", "✅")
	}

	// Test 8 colors
	fmt.Fprintf(o.Out, "\n8-Color Test: ")
	for _, c := range []func(string, ...interface{}) string{
		color.BlackString, color.RedString, color.GreenString, color.YellowString,
		color.BlueString, color.MagentaString, color.CyanString, color.WhiteString,
	} {
		fmt.Fprintf(o.Out, "%s", c("█"))
	}
	fmt.Fprintf(o.Out, "\n")

	// Test 16 colors
	fmt.Fprintf(o.Out, "16-Color Test: ")
	for _, c := range []func(string, ...interface{}) string{
		color.HiBlackString, color.HiRedString, color.HiGreenString, color.HiYellowString,
		color.HiBlueString, color.HiMagentaString, color.HiCyanString, color.HiWhiteString,
	} {
		fmt.Fprintf(o.Out, "%s", c("█"))
	}
	fmt.Fprintf(o.Out, "\n")

	// Test effects
	fmt.Fprintf(o.Out, "\nText Effects Test:\n")
	effects := map[string]func() *color.Color{
		"Bold":      func() *color.Color { return color.New(color.Bold) },
		"Underline": func() *color.Color { return color.New(color.Underline) },
		"Italic":    func() *color.Color { return color.New(color.Italic) },
		"Faint":     func() *color.Color { return color.New(color.Faint) },
	}

	for name, createColor := range effects {
		c := createColor()
		fmt.Fprintf(o.Out, "  %s: %s\n", name, c.Sprint("Sample Text"))
	}

	// Environment info
	fmt.Fprintf(o.Out, "\nEnvironment Information:\n")
	fmt.Fprintf(o.Out, "  TERM: %s\n", color.New(color.FgCyan).Sprint("Not available in this context"))
	fmt.Fprintf(o.Out, "  Color support: %s\n",
		color.New(color.FgGreen).Sprintf("%t", !color.NoColor))

	return nil
}

// runInteractiveMode runs interactive color demonstration.
func (o *ColorOptions) runInteractiveMode() error {
	fmt.Fprintf(o.Out, "%s Interactive Color Demonstration\n",
		color.New(color.Bold, color.FgMagenta).Sprint(emoji.VideoGame))
	fmt.Fprintf(o.Out, "%s\n", strings.Repeat("=", 50))

	// Color palette
	fmt.Fprintf(o.Out, "\n%s Color Palette:\n", emoji.ArtistPalette)
	colors := []struct {
		name string
		fn   func(string, ...interface{}) string
	}{
		{"Red", color.RedString},
		{"Green", color.GreenString},
		{"Blue", color.BlueString},
		{"Yellow", color.YellowString},
		{"Magenta", color.MagentaString},
		{"Cyan", color.CyanString},
	}

	for _, c := range colors {
		fmt.Fprintf(o.Out, "  %s: %s %s\n", c.name, c.fn("████"), c.fn("Sample Text"))
	}

	// Rainbow effect
	fmt.Fprintf(o.Out, "\n%s Rainbow Text:\n", emoji.Rainbow)
	rainbow := []func(string, ...interface{}) string{
		color.RedString, color.YellowString, color.GreenString,
		color.CyanString, color.BlueString, color.MagentaString,
	}

	text := "RAINBOW"
	fmt.Fprintf(o.Out, "  ")
	for i, char := range text {
		colorFn := rainbow[i%len(rainbow)]
		fmt.Fprintf(o.Out, "%s", colorFn(string(char)))
	}
	fmt.Fprintf(o.Out, "\n")

	// Progress bar simulation
	fmt.Fprintf(o.Out, "\n%s Progress Bar Demo:\n", emoji.BarChart)
	progress := 75
	total := 20
	filled := int(float64(progress) / 100 * float64(total))

	fmt.Fprintf(o.Out, "  Progress [%d%%]: ", progress)
	for i := 0; i < total; i++ {
		if i < filled {
			fmt.Fprintf(o.Out, "%s", color.GreenString("█"))
		} else {
			fmt.Fprintf(o.Out, "%s", color.New(color.Faint).Sprint("░"))
		}
	}
	fmt.Fprintf(o.Out, "\n")

	return nil
}
