package avatar_test

import (
	"errors"
	"fmt"

	"github.com/go-mods/avatar"
)

func ExampleGetSVG() {
	svg, err := avatar.GetSVG("JD")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(svg)
	// Output:
	// <?xml version="1.0"?>
	// <svg width="48" height="48"
	//      xmlns="http://www.w3.org/2000/svg"
	//      xmlns:xlink="http://www.w3.org/1999/xlink">
	// <circle cx="24" cy="24" r="24" fill="#9fd6a8" stroke="#89948b" stroke-width="0" stroke-dasharray="none" />
	// <text x="24" y="31" text-anchor="middle" font-family="sans-serif" font-size="21" fill="#05911d">JD</text>
	// </svg>
}

func ExampleGetSVG_options() {
	svg, err := avatar.GetSVG("JD",
		avatar.WithShape("square"),
		avatar.WithWidth(64),
		avatar.WithHeight(64),
		avatar.WithBackgroundColor("navy"),
		avatar.WithFontColor("#ffffff"),
		avatar.WithBorderColor("gold"),
		avatar.WithBorderWidth(2),
		avatar.WithBorderRadius(8),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(svg)
	// Output:
	// <?xml version="1.0"?>
	// <svg width="64" height="64"
	//      xmlns="http://www.w3.org/2000/svg"
	//      xmlns:xlink="http://www.w3.org/1999/xlink">
	// <rect x="2" y="2" width="60" height="60" rx="8" ry="8" fill="#000080" stroke="#ffd700" stroke-width="2" stroke-dasharray="none" />
	// <text x="32" y="41" text-anchor="middle" font-family="sans-serif" font-size="28" fill="#ffffff">JD</text>
	// </svg>
}

func ExampleGetSVG_invalidOption() {
	_, err := avatar.GetSVG("JD", avatar.WithBackgroundColor("not a color"))

	fmt.Println(errors.Is(err, avatar.ErrInvalidColor))
	fmt.Println(err)
	// Output:
	// true
	// invalid color: "not a color"
}

func ExampleNewColorGenerator() {
	colors := avatar.NewColorGenerator("JD")

	fmt.Println(colors.GetBackgroundColor().Hex())
	fmt.Println(colors.GetTextColor().Hex())
	fmt.Println(colors.GetBorderColor().Hex())
	// Output:
	// #9fd6a8
	// #05911d
	// #89948b
}
