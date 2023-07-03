package main

import (
	"avatar"
	"fmt"
	"github.com/go-mods/initials"
	"image"
	"image/draw"
	"image/png"
	"os"
)

var names = []string{"Abdul Zahir",
	"Abena Gyasi",
	"Ahmed Hassan",
	"Aisha Qureshi",
	"Alexei Nikitin",
	"Amalia Zavala",
	"Amira Mahmoud",
	"Ana Cardoso",
	"Anisa Bello",
	"Anjali Iyer",
	"Chen Mei",
	"Elina Ivanova",
	"Erik Varga",
	"Fatima Saad",
	"Gabriel Almeida",
	"Gita Singh",
	"Hana El-Sayed",
	"Hassan Jamil",
	"Hassan Rashid",
	"Ivan Zhdanov",
	"Jan Turek",
	"Johan Fransson",
	"John Doe",
	"Jose Ramos",
	"Karina Gonzalez",
	"Khaled Dabashi",
	"Laila Mahmood",
	"Leif Persson",
	"Mariam Diop",
	"Miguel Lopez",
	"Mohammed Ali",
	"Nadia Qadir",
	"Nao Okamoto",
	"Nizar Fadel",
	"Paula Neves",
	"Ricardo Pereira",
	"Santiago Cortez",
	"Sariya Omer",
	"Seo-jin Cho",
	"Seung-ho Cho",
	"Simone Esposito",
	"Siti Zahara",
	"Sven Bjorn",
	"Tatsuo Hayashi",
	"Tomasz Wójcik",
	"Vanessa Uzoma",
	"Wahida Xiong",
	"Yasmin Zaidi",
	"Yoko Kato",
	"Yulia Kovalenko",
	"Ziad Yusuf",
}

func main() {

	// Path
	avatarPath := "sample/avatar"
	circlePath := "sample/avatar/circle"
	squarePath := "sample/avatar/square"
	customPath := "sample/avatar/custom"
	palettePath := "sample/palette"

	// create avatar directory
	_ = os.Mkdir(avatarPath, 0755) // #nosec:G301
	_ = os.Mkdir(circlePath, 0755) // #nosec:G301
	_ = os.Mkdir(squarePath, 0755) // #nosec:G301
	_ = os.Mkdir(customPath, 0755) // #nosec:G301
	// create palette directory
	_ = os.Mkdir(palettePath, 0755) // #nosec:G301

	// Generate palettes
	for _, name := range names {
		generatePalette(palettePath, name)
	}

	// generate circle avatars
	for _, name := range names {
		generateAvatar(circlePath, name, avatar.WithShape("circle"))
	}

	// generate square avatars
	for _, name := range names {
		generateAvatar(squarePath, name, avatar.WithShape("square"))
	}

	// generate custom avatars
	generateAvatar(customPath, names[1], avatar.WithShape("square"), avatar.WithBorderWidth(2))
	generateAvatar(customPath, names[2], avatar.WithShape("square"), avatar.WithBorderWidth(2), avatar.WithBorderRadius(4))
	generateAvatar(customPath, names[3], avatar.WithShape("square"), avatar.WithBorderWidth(2), avatar.WithBorderRadius(4), avatar.WithBorderDash("6 2"))
	generateAvatar(customPath, names[4], avatar.WithShape("square"), avatar.WithBorderColor("#AFA318"), avatar.WithBorderWidth(2), avatar.WithBorderRadius(4), avatar.WithBorderDash("6 2"))

	generateAvatar(customPath, names[5], avatar.WithShape("circle"), avatar.WithBorderWidth(2))
	generateAvatar(customPath, names[6], avatar.WithShape("circle"), avatar.WithBorderWidth(2), avatar.WithPadding(2))
	generateAvatar(customPath, names[7], avatar.WithShape("circle"), avatar.WithBorderWidth(2), avatar.WithBorderDash("3 1"))
	generateAvatar(customPath, names[8], avatar.WithShape("circle"), avatar.WithBorderColor("#AFA318"), avatar.WithBorderWidth(2), avatar.WithBorderDash("3 1"))

	generateAvatar(customPath, names[9], avatar.WithFont("Roboto", 400))
	generateAvatar(customPath, names[10], avatar.WithFont("Tangerine", 400))
	generateAvatar(customPath, names[11], avatar.WithFont("Inconsolata", 400))
	generateAvatar(customPath, names[12], avatar.WithFont("Droid Sans", 400))
	generateAvatar(customPath, names[13], avatar.WithFont("Rancho", 400))
	generateAvatar(customPath, names[14], avatar.WithFont("Josefin Sans", 400))
	generateAvatar(customPath, names[15], avatar.WithFont("Lalezar", 400))

	generateAvatar(customPath, "John Doe", avatar.WithShape("square"),
		avatar.WithWidth(200), avatar.WithHeight(200),
		avatar.WithBackgroundColor("red"),
		avatar.WithFontColor("white"),
		avatar.WithFont("Caprasimo", 400),
		avatar.WithBorderWidth(10),
		avatar.WithBorderColor("blue"),
		avatar.WithBorderDash("10 10"),
		avatar.WithBorderRadius(50),
		avatar.WithPadding(10),
	)
}

func generatePalette(destination string, name string) {

	// Get initials from name
	i := initials.GetInitials(name)

	// Create a new color generator
	g := avatar.NewColorGenerator(i)

	// get the background color from the generator
	background := g.GetBackgroundColor()

	// get the text color from the generator
	text := g.GetTextColor()

	// get the border color from the generator
	border := g.GetBorderColor()

	// Generate palette image
	width := 40
	img := image.NewRGBA(image.Rect(0, 0, 3*width, width))
	draw.Draw(img, image.Rect(0*width, 0, (0+1)*width-2, 40-2), &image.Uniform{C: background}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(1*width, 0, (1+1)*width-2, 40-2), &image.Uniform{C: text}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(2*width, 0, (2+1)*width-2, 40-2), &image.Uniform{C: border}, image.Point{}, draw.Src)

	// Save image
	file, err := os.Create(fmt.Sprintf("%s/%s.png", destination, name))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_ = png.Encode(file, img)
}

func generateAvatar(destination string, name string, options ...avatar.GeneratorOption) {

	// Get initials from name
	i := initials.GetInitials(name)

	// Generate the avatar
	svg := avatar.GetSVG(i, options...)

	// save the svg to a file
	_ = os.WriteFile(fmt.Sprintf("%s/%s.svg", destination, name), svg.Bytes(), 0644) // #nosec:G306
}
