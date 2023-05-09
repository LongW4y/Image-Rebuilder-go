package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Global variables
var size int = 4

// input arguments Flag struct
type Flag struct {
	source        string
	output        string
	input_format  string
	output_format string
	help          bool
}

// Source image struct
type Image struct {
	path       string
	dimensions []int
	format     string
	image      image.Image
}

func parse_flags() Flag {
	// Define the flag
	var flags Flag
	flag.StringVar(&flags.source, "source", "", "Source image path, cannot be empty.")
	flag.StringVar(&flags.output, "", "output", "Output image path, optional. If not provided, the output will be named output_<source>.<output_format> in the current directory.")
	flag.StringVar(&flags.output_format, "output-format", "", "Output image format, optional. If not provided, the output format will be the same as the source image.")
	flag.BoolVar(&flags.help, "help", false, "Show this help message")

	flag.Parse()

	// Show help if requested
	if flags.help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate the flags
	valid, flags := validate_flags(flags)
	if !valid {
		flag.Usage()
		os.Exit(1)
	}

	if flags.output == "" {
		flags.output = "output_" + flags.source[:len(flags.source)-4] + flags.output_format
	}

	return flags
}

func validate_flags(flags Flag) (bool, Flag) {
	// Only source cannot be empty
	if flags.source == "" {
		return false, flags
	}
	// Check if source file exists
	if _, err := os.Stat(flags.source); os.IsNotExist(err) {
		fmt.Println("Source file does not exist")
		return false, flags
	}
	// Check if the source file is either a png or a jpg
	if flags.source[len(flags.source)-4:] != ".png" && flags.source[len(flags.source)-4:] != ".jpg" {
		fmt.Println("Source file must be either a png or a jpg")
		return false, flags
	}
	// Set the input format, lower case
	flags.input_format = strings.ToLower(flags.source[len(flags.source)-4:])

	// Set the output format if not provided
	if flags.output_format == "" {
		flags.output_format = flags.input_format
	} else if flags.output_format != ".png" && flags.output_format != ".jpg" {
		fmt.Println("Output format must be either a png or a jpg")
		return false, flags
	}

	// Check if output file exists, if it does, warn the user
	if flags.output != "" {
		if _, err := os.Stat(flags.output); !os.IsNotExist(err) {
			fmt.Println("The file specified as output: " + flags.output + " already exists.")
			fmt.Println("File will be named " + flags.output[:len(flags.output)-4] + "_1" + flags.output_format + " instead.")
			flags.output = flags.output[:len(flags.output)-4] + "_1" + flags.output_format
		}
	}

	return true, flags
}

func load_image(path string, format string) Image {
	new_image := Image{}
	new_image.path = path
	new_image.format = format
	// Load the image
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error while opening the source file: " + err.Error())
	}
	defer file.Close()
	if format == ".png" {
		new_image.image, err = png.Decode(file)
	} else if format == ".jpg" || format == "jpeg" {
		new_image.image, err = jpeg.Decode(file)
	} else {
		fmt.Println("Error: unknown format: " + format)
		os.Exit(1)
	}
	if err != nil {
		fmt.Println("Error while decoding the source file: " + err.Error())
	}
	// Get the dimensions
	new_image.dimensions = []int{new_image.image.Bounds().Max.X, new_image.image.Bounds().Max.Y}

	return new_image
}

func get_possible_cuts(image Image) [][]int {
	// Calculate into how many squares the image can be cut
	// The square has to be X by X pixels large
	// The squares in the last row and column can be up to 5% smaller than the rest
	// The image has to be cut into as many squares as possible, but calculate all the options
	// Return a list of possible cuts, where an element in a list is made up of the squares dimensions (just need one since it's X by X) and whether it's a perfect cut or not

	possible_cuts := [][]int{}
	// Append possible_cuts with 1x1 squares with 'true'
	possible_cuts = append(possible_cuts, []int{size, size, 1})

	// TODO: Implement the rest of the possible cuts
	// // Calculate how many squares can fit into the image
	// for i := 2; i <= image.dimensions[0]; i++ {
	// 	for j := 2; j <= image.dimensions[1]; j++ {
	// 		// Check if the square is a perfect cut
	// 		if image.dimensions[0]%i == 0 && image.dimensions[1]%j == 0 {
	// 			possible_cuts = append(possible_cuts, []int{i, j, 1})
	// 		} else if image.dimensions[0]%i <= int(float64(image.dimensions[0])*0.05) && image.dimensions[1]%j <= int(float64(image.dimensions[1])*0.05) {
	// 			possible_cuts = append(possible_cuts, []int{i, j, 0})
	// 		}
	// 	}
	// }

	return possible_cuts
}

func get_num_of_squares(image Image, square_dimensions []int) int {
	// Calculate how many squares can fit into the image
	num_of_squares := 0
	for i := 0; i < image.dimensions[0]; i += square_dimensions[0] {
		for j := 0; j < image.dimensions[1]; j += square_dimensions[1] {
			num_of_squares++
		}
	}

	return num_of_squares
}

func generate_n_colors(num int) []color.Color {
	// Generate n colors
	colors := []color.Color{}
	for i := 0; i < num; i++ {
		colors = append(colors, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
	}

	return colors
}

func assign_color_to_pixel(source_image Image, colors []color.Color, dimensions []int) image.Image {
	// Assign a color to each pixel in the image
	// The new should be as close as possible to the original
	// If a color is used, remove it from the colors list
	// Run it in parallel to speed it up & print the progress

	// Create a new black image
	new_image := image.NewRGBA(image.Rect(0, 0, source_image.dimensions[0], source_image.dimensions[1]))
	// Color the new_image black
	for i := 0; i < source_image.dimensions[0]; i++ {
		for j := 0; j < source_image.dimensions[1]; j++ {
			new_image.Set(i, j, color.RGBA{0, 0, 0, 255})
		}
	}

	pixels_colored := 0
	// Assign a color to each pixel, use get_closest_color function to find the closest color
	for i := 0; i < source_image.dimensions[0]; i += dimensions[0] {
		for j := 0; j < source_image.dimensions[1]; j += dimensions[1] {
			square_colors := []color.Color{}
			for k := i; k < i+dimensions[0]; k++ {
				for l := j; l < j+dimensions[1]; l++ {
					square_colors = append(square_colors, source_image.image.At(k, l))
				}
			}
			closest_color, index := get_closest_color(square_colors, colors)
			colors = append(colors[:index], colors[index+1:]...)
			for k := i; k < i+dimensions[0]; k++ {
				for l := j; l < j+dimensions[1]; l++ {
					new_image.Set(k, l, closest_color)
				}
			}
			// Print the progress by listing the number of pixels colored
			pixels_colored += dimensions[0] * dimensions[1]
			fmt.Printf("\rPixels colored: %d/%d", pixels_colored, source_image.dimensions[0]*source_image.dimensions[1])
		}
	}

	return new_image
}

func get_closest_color(square_colors []color.Color, list_of_colors []color.Color) (color.Color, int) {
	// Get the closest color from the colors list to the avg_square_color
	// Return the closest color and the index of the color in the colors list
	// Use the get_color_distance function to find a color with the smallest distance to the avg_square_color
	avg_square_color := calculate_avg_color(square_colors)
	closest_color := list_of_colors[0]
	index := 0
	for i, color := range list_of_colors {
		if get_color_distance(avg_square_color, color) < get_color_distance(avg_square_color, closest_color) {
			closest_color = color
			index = i
		}
	}

	return closest_color, index
}

func get_color_distance(source_color color.Color, target_color color.Color) int {
	// Get the distance between two colors
	// The distance is calculated as the sum of the differences between the red, green and blue values
	R1, G1, B1, A1 := source_color.RGBA()
	R2, G2, B2, A2 := target_color.RGBA()
	return int(math.Abs(float64(R1-R2)) + math.Abs(float64(G1-G2)) + math.Abs(float64(B1-B2)) + math.Abs(float64(A1-A2)))
}

func calculate_avg_color(colors []color.Color) color.Color {
	// Calculate average RGBA values for a list of colors

	// Define a list
	Rs := []int{}
	Gs := []int{}
	Bs := []int{}
	As := []int{}

	// Get the RGBA values for each color
	for _, color := range colors {
		R, G, B, A := color.RGBA()
		Rs = append(Rs, int(R))
		Gs = append(Gs, int(G))
		Bs = append(Bs, int(B))
		As = append(As, int(A))
	}

	// Calculate the average RGBA values
	return color.RGBA{uint8(get_avg(Rs)), uint8(get_avg(Gs)), uint8(get_avg(Bs)), uint8(get_avg(As))}
}

func get_avg(list []int) int {
	// Calculate the average of a list of integers
	sum := 0
	for _, num := range list {
		sum += num
	}

	return sum / len(list)
}

func save_image(output image.Image, format string, filename string) {
	// Save the image in format `format` as `filename`
	file, err := os.Create(filename + format)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	switch format {
	case ".jpg":
		jpeg.Encode(file, output, nil)
	case ".png":
		png.Encode(file, output)
	}
}

func main() {
	// Start the timer
	start := time.Now()
	flags := parse_flags()
	image := load_image(flags.source, flags.input_format)
	possible_cuts := get_possible_cuts(image)
	num_of_squares := get_num_of_squares(image, possible_cuts[0])
	colors := generate_n_colors(num_of_squares)

	// Color the new image
	output := assign_color_to_pixel(image, colors, possible_cuts[0])

	// Save the new image
	save_image(output, flags.output_format, flags.output)
	// Stop the timer
	elapsed := time.Since(start)
	fmt.Printf("Time elapsed: %s\n", elapsed)
}

// TODOs:
// - Implement the rest of the possible cuts
// - Color the image in parallel with goroutines
// - Implement rounding of the closest color, e.g.: if the distance is smaller than X, it can be classified as close enough
// - GUI? Web app? If bored
