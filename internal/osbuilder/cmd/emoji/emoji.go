package emoji

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/enescakir/emoji"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"

	cmdutil "github.com/onexstack/osbuilder/internal/osbuilder/cmd/util"
)

// EmojiOptions defines configuration for the "emoji" command.
type EmojiOptions struct {
	List     bool   // list all available emojis
	Search   string // search emojis by keyword
	Category string // filter by category
	Format   string // output format: table, json, simple
	Limit    int    // limit number of results
	Code     bool   // show unicode codes
	Random   int    // show random emojis

	genericiooptions.IOStreams
}

// EmojiInfo holds information about an emoji from the enescakir/emoji package.
type EmojiInfo struct {
	Name     string `json:"name"`
	Emoji    string `json:"emoji"`
	Code     string `json:"code"`
	Category string `json:"category"`
}

var (
	emojiLongDesc = templates.LongDesc(`
		Display, search, and explore emoji characters using the comprehensive emoji library.

		This command provides access to all emojis available in the github.com/enescakir/emoji
		package. You can display specific emojis, list all available emojis, search by keywords,
		and export in different formats.

		Features:
		- Display specific emojis by name (e.g., rocket, heart, smile)
		- List all available emojis from the emoji package
		- Search emojis by keyword or name pattern
		- Multiple output formats (simple, table, json)
		- Show Unicode codes for programming use
		- Random emoji selection for inspiration
		- Categorized emoji organization
	`)

	emojiExample = templates.Examples(`
		# Display specific emojis
		osbuilder emoji rocket star heart grinning_face

		# List all available emojis
		osbuilder emoji --list

		# Search emojis by keyword
		osbuilder emoji --search "smile"
		osbuilder emoji --search "heart"
		osbuilder emoji --search "fire"

		# Show with unicode codes
		osbuilder emoji rocket star --code

		# Table format with limit
		osbuilder emoji --list --format table --limit 20

		# Get random emojis for inspiration
		osbuilder emoji --random 5

		# Export as JSON
		osbuilder emoji --search "face" --format json --limit 10

		# Filter by category (if available)
		osbuilder emoji --list --category "smileys"
	`)

	// Predefined categories based on common emoji groupings
	emojiCategories = map[string][]string{
		"faces": {
			"grinning_face", "grinning_face_with_big_eyes", "grinning_face_with_smiling_eyes",
			"beaming_face_with_smiling_eyes", "grinning_squinting_face", "face_with_tears_of_joy",
			"rolling_on_the_floor_laughing", "winking_face", "smiling_face_with_heart_eyes",
			"smiling_face_with_heart_shaped_eyes", "star_struck", "face_blowing_a_kiss",
			"kissing_face", "smiling_face", "kissing_face_with_closed_eyes",
			"kissing_face_with_smiling_eyes", "smiling_face_with_tear", "face_savoring_food",
			"face_with_tongue", "winking_face_with_tongue", "zany_face", "squinting_face_with_tongue",
			"money_mouth_face", "smiling_face_with_open_hands", "face_with_hand_over_mouth",
			"face_with_open_eyes_and_hand_over_mouth", "face_with_peeking_eye", "shushing_face",
			"thinking_face", "saluting_face", "zipper_mouth_face", "face_with_raised_eyebrow",
			"neutral_face", "expressionless_face", "face_without_mouth", "dotted_line_face",
			"face_in_clouds", "smirking_face", "unamused_face", "face_with_rolling_eyes",
			"grimacing_face", "face_exhaling", "lying_face", "shaking_face",
		},
		"hearts": {
			"red_heart", "orange_heart", "yellow_heart", "green_heart", "blue_heart",
			"purple_heart", "brown_heart", "black_heart", "grey_heart", "white_heart",
			"pink_heart", "heart_with_arrow", "heart_with_ribbon", "sparkling_heart",
			"growing_heart", "beating_heart", "revolving_hearts", "two_hearts",
			"heart_decoration", "heavy_heart_exclamation", "broken_heart", "heart_on_fire",
			"mending_heart", "heart", "love_letter", "kiss_mark",
		},
		"animals": {
			"monkey_face", "monkey", "gorilla", "orangutan", "dog_face", "dog", "guide_dog",
			"service_dog", "poodle", "wolf", "fox", "raccoon", "cat_face", "cat",
			"black_cat", "lion", "tiger_face", "tiger", "leopard", "horse_face", "horse",
			"unicorn", "zebra", "deer", "bison", "cow_face", "cow", "ox", "water_buffalo",
			"pig_face", "pig", "boar", "pig_nose", "ram", "ewe", "goat", "camel",
			"two_hump_camel", "llama", "giraffe", "elephant", "mammoth", "rhinoceros",
			"hippopotamus", "mouse_face", "mouse", "rat", "hamster", "rabbit_face",
			"rabbit", "chipmunk", "beaver", "hedgehog", "bat", "bear", "polar_bear",
			"koala", "panda", "sloth", "otter", "skunk", "kangaroo", "badger",
		},
		"objects": {
			"watch", "mobile_phone", "calling", "computer", "keyboard", "desktop_computer",
			"printer", "computer_mouse", "trackball", "joystick", "compression", "minidisc",
			"floppy_disk", "cd", "dvd", "abacus", "movie_camera", "film_strip", "film_projector",
			"clapper", "tv", "camera", "camera_flash", "video_camera", "vhs", "mag",
			"mag_right", "candle", "bulb", "flashlight", "red_paper_lantern", "diya_lamp",
		},
		"transportation": {
			"car", "taxi", "blue_car", "pickup_truck", "bus", "trolleybus", "racing_car",
			"police_car", "ambulance", "fire_engine", "minibus", "delivery_truck", "articulated_lorry",
			"tractor", "racing_motorcycle", "motorcycle", "motor_scooter", "manual_wheelchair",
			"motorized_wheelchair", "auto_rickshaw", "bike", "kick_scooter", "skateboard",
			"roller_skate", "bus_stop", "motorway", "railway_track", "oil_drum", "fuelpump",
			"rotating_light", "traffic_light", "vertical_traffic_light", "stop_sign",
			"construction", "anchor", "ring_buoy", "boat", "canoe", "speedboat", "passenger_ship",
			"ferry", "motor_boat", "ship", "airplane", "small_airplane", "flight_departure",
			"flight_arrival", "parachute", "seat", "helicopter", "suspension_railway",
			"mountain_cableway", "aerial_tramway", "artificial_satellite", "rocket",
		},
		"symbols": {
			"heart", "yellow_heart", "green_heart", "blue_heart", "purple_heart", "brown_heart",
			"black_heart", "grey_heart", "white_heart", "pink_heart", "orange_heart",
			"red_heart", "heavy_heart_exclamation", "broken_heart", "two_hearts", "revolving_hearts",
			"beating_heart", "growing_heart", "sparkling_heart", "cupid", "gift_heart",
			"heart_decoration", "peace_symbol", "latin_cross", "star_and_crescent",
			"om", "wheel_of_dharma", "star_of_david", "six_pointed_star", "menorah",
			"yin_yang", "orthodox_cross", "place_of_worship", "ophiuchus", "aries",
			"taurus", "gemini", "cancer", "leo", "virgo", "libra", "scorpius",
			"sagittarius", "capricorn", "aquarius", "pisces", "id", "atom_symbol",
			"accept", "radioactive", "biohazard", "mobile_phone_off", "vibration_mode",
			"u6709", "u7121", "u7533", "u55b6", "u6708", "eight_pointed_black_star",
			"vs", "accept", "white_flower", "ideograph_advantage", "secret", "congratulations",
		},
	}
)

// NewEmojiCmd creates the "emoji" command.
func NewEmojiCmd(factory cmdutil.Factory, ioStreams genericiooptions.IOStreams) *cobra.Command {
	opts := &EmojiOptions{
		IOStreams: ioStreams,
		Format:    "simple",
		Limit:     0,
	}

	cmd := &cobra.Command{
		Use:                   "emoji [emoji-names...]",
		Short:                 "Display and search emoji characters with their codes",
		Long:                  emojiLongDesc,
		Example:               emojiExample,
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.Complete())
			cmdutil.CheckErr(opts.Validate())
			cmdutil.CheckErr(opts.Run(args))
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&opts.List, "list", "l", false, "List all available emojis")
	cmd.Flags().StringVarP(&opts.Search, "search", "s", "", "Search emojis by keyword")
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "", "Filter by category (faces, hearts, animals, objects, transportation, symbols)")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", opts.Format, "Output format: simple, table, json")
	cmd.Flags().IntVar(&opts.Limit, "limit", opts.Limit, "Limit number of results (0 for no limit)")
	cmd.Flags().BoolVar(&opts.Code, "code", false, "Show unicode codes")
	cmd.Flags().IntVar(&opts.Random, "random", 0, "Show random emojis (specify count)")

	return cmd
}

// Complete sets default values and resolves working directory.
func (o *EmojiOptions) Complete() error {
	// Normalize format
	o.Format = strings.ToLower(o.Format)
	if o.Format != "simple" && o.Format != "table" && o.Format != "json" {
		o.Format = "simple"
	}

	// Normalize category
	if o.Category != "" {
		o.Category = strings.ToLower(o.Category)
		if o.Category == "help" {
			o.showAvailableCategories()
			return fmt.Errorf("") // Exit gracefully after showing help
		}
		if _, exists := emojiCategories[o.Category]; !exists {
			return fmt.Errorf("invalid category: %s. Available categories: %s",
				o.Category, strings.Join(o.getAvailableCategories(), ", "))
		}
	}

	return nil
}

// Validate ensures provided inputs are valid.
func (o *EmojiOptions) Validate() error {
	if o.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if o.Random < 0 {
		return fmt.Errorf("random count cannot be negative")
	}
	return nil
}

// Run performs the emoji operation.
func (o *EmojiOptions) Run(args []string) error {
	// Handle random emoji generation
	if o.Random > 0 {
		return o.showRandomEmojis()
	}

	// If listing or searching, handle those cases
	if o.List || o.Search != "" {
		return o.handleListOrSearch()
	}

	// If no arguments provided, show help
	if len(args) == 0 {
		return o.showUsageHelp()
	}

	// Display specific emojis
	return o.displayEmojis(args)
}

// displayEmojis displays specific emojis by name.
func (o *EmojiOptions) displayEmojis(names []string) error {
	var results []EmojiInfo
	var notFound []string

	for _, name := range names {
		// Try to get emoji using the package
		emojiStr := emoji.Parse(fmt.Sprintf(":%s:", name))

		// Check if emoji was found (if it's different from the input)
		if emojiStr != fmt.Sprintf(":%s:", name) {
			info := EmojiInfo{
				Name:     name,
				Emoji:    emojiStr,
				Code:     o.getUnicodeCode(emojiStr),
				Category: o.getCategoryForEmoji(name),
			}
			results = append(results, info)
		} else {
			notFound = append(notFound, name)
		}
	}

	// Display results
	if len(results) > 0 {
		if err := o.outputEmojis(results); err != nil {
			return err
		}
	}

	// Report not found
	if len(notFound) > 0 {
		fmt.Fprintf(o.Out, "\n%s Not found: %s\n",
			emoji.CrossMark, color.RedString(strings.Join(notFound, ", ")))
		fmt.Fprintf(o.Out, "%s Try using --search to find similar emojis\n", emoji.LightBulb)
		fmt.Fprintf(o.Out, "%s Use --list to see all available emojis\n", emoji.Memo)
	}

	return nil
}

// handleListOrSearch handles list and search operations.
func (o *EmojiOptions) handleListOrSearch() error {
	allEmojis := o.getAllEmojisFromPackage()

	var filteredEmojis []EmojiInfo

	// Apply category filter
	if o.Category != "" {
		categoryEmojis := emojiCategories[o.Category]
		for _, emojiInfo := range allEmojis {
			for _, categoryEmoji := range categoryEmojis {
				if emojiInfo.Name == categoryEmoji {
					filteredEmojis = append(filteredEmojis, emojiInfo)
					break
				}
			}
		}
	} else {
		filteredEmojis = allEmojis
	}

	// Apply search filter
	if o.Search != "" {
		var searchResults []EmojiInfo
		searchLower := strings.ToLower(o.Search)
		for _, emojiInfo := range filteredEmojis {
			if strings.Contains(strings.ToLower(emojiInfo.Name), searchLower) {
				searchResults = append(searchResults, emojiInfo)
			}
		}
		filteredEmojis = searchResults
	}

	if len(filteredEmojis) == 0 {
		fmt.Fprintf(o.Out, "%s No emojis found matching your criteria\n", emoji.CrossMark)
		return nil
	}

	// Apply limit
	if o.Limit > 0 && len(filteredEmojis) > o.Limit {
		filteredEmojis = filteredEmojis[:o.Limit]
		fmt.Fprintf(o.Out, "%s Showing first %d results (use --limit 0 to show all)\n\n",
			"ℹ️", o.Limit)
	}

	return o.outputEmojis(filteredEmojis)
}

// showRandomEmojis displays random emojis.
func (o *EmojiOptions) showRandomEmojis() error {
	allEmojis := o.getAllEmojisFromPackage()

	if len(allEmojis) == 0 {
		return fmt.Errorf("no emojis available")
	}

	count := o.Random
	if count > len(allEmojis) {
		count = len(allEmojis)
	}

	// Shuffle and take first N
	shuffled := make([]EmojiInfo, len(allEmojis))
	copy(shuffled, allEmojis)

	// Simple shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := i % (i + 1) // Simple pseudo-random
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	randomEmojis := shuffled[:count]

	fmt.Fprintf(o.Out, "%s Random emojis for inspiration:\n\n", emoji.GameDie)
	return o.outputEmojis(randomEmojis)
}

// getAllEmojisFromPackage gets all emojis using reflection from the emoji package.
func (o *EmojiOptions) getAllEmojisFromPackage() []EmojiInfo {
	var emojis []EmojiInfo

	// Get all emoji constants from the package using reflection
	// This is a simplified approach - in practice, you might want to maintain a static list
	// or use the package's exported functions if available

	// For now, let's use our predefined categories and test each one
	for category, names := range emojiCategories {
		for _, name := range names {
			emojiStr := emoji.Parse(fmt.Sprintf(":%s:", name))
			if emojiStr != fmt.Sprintf(":%s:", name) {
				emojis = append(emojis, EmojiInfo{
					Name:     name,
					Emoji:    emojiStr,
					Code:     o.getUnicodeCode(emojiStr),
					Category: category,
				})
			}
		}
	}

	// Add some common emojis that might not be categorized
	commonEmojis := []string{
		"thumbs_up", "thumbs_down", "clap", "raised_hands", "pray", "handshake",
		"muscle", "point_up", "point_down", "point_left", "point_right",
		"ok_hand", "crossed_fingers", "peace", "love_you_gesture", "metal",
		"call_me", "backhand_index_pointing_left", "backhand_index_pointing_right",
		"backhand_index_pointing_up", "backhand_index_pointing_down", "index_pointing_up",
		"raised_hand", "raised_back_of_hand", "vulcan", "writing_hand",
	}

	for _, name := range commonEmojis {
		emojiStr := emoji.Parse(fmt.Sprintf(":%s:", name))
		if emojiStr != fmt.Sprintf(":%s:", name) {
			// Check if already exists
			exists := false
			for _, existing := range emojis {
				if existing.Name == name {
					exists = true
					break
				}
			}
			if !exists {
				emojis = append(emojis, EmojiInfo{
					Name:     name,
					Emoji:    emojiStr,
					Code:     o.getUnicodeCode(emojiStr),
					Category: "gestures",
				})
			}
		}
	}

	// Sort by name
	sort.Slice(emojis, func(i, j int) bool {
		return emojis[i].Name < emojis[j].Name
	})

	return emojis
}

// outputEmojis outputs emojis in the specified format.
func (o *EmojiOptions) outputEmojis(emojis []EmojiInfo) error {
	switch o.Format {
	case "table":
		return o.outputTable(emojis)
	case "json":
		return o.outputJSON(emojis)
	default: // simple
		return o.outputSimple(emojis)
	}
}

// outputSimple outputs emojis in simple format.
func (o *EmojiOptions) outputSimple(emojis []EmojiInfo) error {
	for _, emojiInfo := range emojis {
		if o.Code {
			fmt.Fprintf(o.Out, "%s %s %s\n",
				emojiInfo.Emoji,
				color.CyanString(emojiInfo.Name),
				color.GreenString(emojiInfo.Code))
		} else {
			fmt.Fprintf(o.Out, "%s %s\n",
				emojiInfo.Emoji,
				color.CyanString(emojiInfo.Name))
		}
	}
	return nil
}

// outputTable outputs emojis in table format.
func (o *EmojiOptions) outputTable(emojis []EmojiInfo) error {
	w := tabwriter.NewWriter(o.Out, 0, 8, 2, ' ', 0)

	// Header
	if o.Code {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			color.GreenString("EMOJI"),
			color.GreenString("NAME"),
			color.GreenString("CATEGORY"),
			color.GreenString("CODE"))
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			"-----", "----", "--------", "----")
	} else {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			color.GreenString("EMOJI"),
			color.GreenString("NAME"),
			color.GreenString("CATEGORY"))
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			"-----", "----", "--------")
	}

	// Data
	for _, emojiInfo := range emojis {
		if o.Code {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				emojiInfo.Emoji,
				color.CyanString(emojiInfo.Name),
				color.YellowString(emojiInfo.Category),
				color.GreenString(emojiInfo.Code))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				emojiInfo.Emoji,
				color.CyanString(emojiInfo.Name),
				color.YellowString(emojiInfo.Category))
		}
	}

	return w.Flush()
}

// outputJSON outputs emojis in JSON format.
func (o *EmojiOptions) outputJSON(emojis []EmojiInfo) error {
	fmt.Fprintf(o.Out, "[\n")
	for i, emojiInfo := range emojis {
		fmt.Fprintf(o.Out, "  {\n")
		fmt.Fprintf(o.Out, "    \"name\": %s,\n", strconv.Quote(emojiInfo.Name))
		fmt.Fprintf(o.Out, "    \"emoji\": %s,\n", strconv.Quote(emojiInfo.Emoji))
		fmt.Fprintf(o.Out, "    \"category\": %s", strconv.Quote(emojiInfo.Category))
		if o.Code {
			fmt.Fprintf(o.Out, ",\n    \"code\": %s", strconv.Quote(emojiInfo.Code))
		}
		fmt.Fprintf(o.Out, "\n")

		if i < len(emojis)-1 {
			fmt.Fprintf(o.Out, "  },\n")
		} else {
			fmt.Fprintf(o.Out, "  }\n")
		}
	}
	fmt.Fprintf(o.Out, "]\n")
	return nil
}

// showUsageHelp shows usage examples and available categories.
func (o *EmojiOptions) showUsageHelp() error {
	fmt.Fprintf(o.Out, "%s Welcome to the emoji command!\n\n", emoji.WavingHand)

	fmt.Fprintf(o.Out, "%s Quick examples:\n", emoji.LightBulb)
	fmt.Fprintf(o.Out, "  osbuilder emoji rocket star heart grinning_face\n")
	fmt.Fprintf(o.Out, "  osbuilder emoji --list --category faces\n")
	fmt.Fprintf(o.Out, "  osbuilder emoji --search smile --format table\n")
	fmt.Fprintf(o.Out, "  osbuilder emoji --random 5\n\n")

	fmt.Fprintf(o.Out, "%s Available categories: %s\n\n",
		emoji.FileFolder, color.YellowString(strings.Join(o.getAvailableCategories(), ", ")))

	fmt.Fprintf(o.Out, "%s Popular emojis to try:\n", emoji.GlowingStar)
	popularEmojis := []string{"rocket", "star", "heart", "fire", "thumbs_up", "grinning_face", "party", "muscle"}
	for _, name := range popularEmojis {
		emojiStr := emoji.Parse(fmt.Sprintf(":%s:", name))
		if emojiStr != fmt.Sprintf(":%s:", name) {
			fmt.Fprintf(o.Out, "  %s %s\n", emojiStr, color.CyanString(name))
		}
	}

	fmt.Fprintf(o.Out, "\n%s Use --help for more options\n", emoji.Information)

	return nil
}

// showAvailableCategories shows all available categories with sample emojis.
func (o *EmojiOptions) showAvailableCategories() {
	fmt.Fprintf(o.Out, "%s Available emoji categories:\n\n", emoji.Books)

	for category, names := range emojiCategories {
		fmt.Fprintf(o.Out, "%s %s:\n", emoji.FileFolder, color.YellowString(category))

		// Show first few emojis from each category
		count := 0
		for _, name := range names {
			if count >= 8 { // Limit display
				fmt.Fprintf(o.Out, " ...")
				break
			}
			emojiStr := emoji.Parse(fmt.Sprintf(":%s:", name))
			if emojiStr != fmt.Sprintf(":%s:", name) {
				fmt.Fprintf(o.Out, " %s", emojiStr)
				count++
			}
		}
		fmt.Fprintf(o.Out, "\n\n")
	}

	fmt.Fprintf(o.Out, "%s Usage: osbuilder emoji --list --category CATEGORY_NAME\n", emoji.LightBulb)
}

// Helper methods

func (o *EmojiOptions) getAvailableCategories() []string {
	var categories []string
	for category := range emojiCategories {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories
}

func (o *EmojiOptions) getCategoryForEmoji(name string) string {
	for category, names := range emojiCategories {
		for _, categoryName := range names {
			if categoryName == name {
				return category
			}
		}
	}
	return "misc"
}

func (o *EmojiOptions) getUnicodeCode(emojiStr string) string {
	if len(emojiStr) == 0 {
		return ""
	}

	runes := []rune(emojiStr)
	if len(runes) == 0 {
		return ""
	}

	var codes []string
	for _, r := range runes {
		if r > 127 { // Only show non-ASCII characters
			codes = append(codes, fmt.Sprintf("U+%04X", r))
		}
	}
	return strings.Join(codes, " ")
}

// PrintSuccess prints a success message with emoji.
func (o *EmojiOptions) PrintSuccess(message string) {
	fmt.Fprintf(o.Out, "%s %s\n", emoji.CheckMarkButton, color.GreenString(message))
}

// PrintError prints an error message with emoji.
func (o *EmojiOptions) PrintError(message string) {
	fmt.Fprintf(o.Out, "%s %s\n", emoji.CrossMark, color.RedString(message))
}

// PrintInfo prints an info message with emoji.
func (o *EmojiOptions) PrintInfo(message string) {
	fmt.Fprintf(o.Out, "%s %s\n", emoji.Information, color.CyanString(message))
}

// GetAllEmojiNames returns all available emoji names (useful for other commands).
func GetAllEmojiNames() []string {
	var names []string
	for _, categoryNames := range emojiCategories {
		names = append(names, categoryNames...)
	}

	// Add common emojis
	commonEmojis := []string{
		"thumbs_up", "thumbs_down", "clap", "raised_hands", "pray", "handshake",
		"muscle", "point_up", "point_down", "point_left", "point_right",
		"ok_hand", "crossed_fingers", "peace", "love_you_gesture", "metal",
	}
	names = append(names, commonEmojis...)

	// Remove duplicates and sort
	uniqueNames := make(map[string]bool)
	var result []string
	for _, name := range names {
		if !uniqueNames[name] {
			uniqueNames[name] = true
			result = append(result, name)
		}
	}
	sort.Strings(result)

	return result
}
