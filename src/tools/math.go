package tools

import (
	"math"
)

// Function to upscale a 2d float array using bicubic interpolation with corner boundary checks
func BicubicInterpolation(arr [][]float64, newWidth int, newHeight int) [][]float64 {
	// Get the original width and height of the array
	width := len(arr[0])
	height := len(arr)

	// Create a new array with the given new width and height
	newArr := make([][]float64, newHeight)
	for i := range newArr {
		newArr[i] = make([]float64, newWidth)
	}

	// Loop through the new array and calculate the bicubic interpolation for each element
	for i := 0; i < newHeight; i++ {
		for j := 0; j < newWidth; j++ {
			// Calculate the x and y coordinates in the original array
			x := float64(j) * (float64(width) / float64(newWidth))
			y := float64(i) * (float64(height) / float64(newHeight))

			// Get the coordinates of the 4 corners of the square in the original array
			x0 := int(math.Floor(x))
			y0 := int(math.Floor(y))
			x1 := int(math.Ceil(x))
			y1 := int(math.Ceil(y))

			// Check if the coordinates are within the bounds of the original array
			if x0 < 0 {
				x0 = 0
			}
			if x1 > width-1 {
				x1 = width - 1
			}
			if y0 < 0 {
				y0 = 0
			}
			if y1 > height-1 {
				y1 = height - 1
			}

			// Get the 4 corner points
			p00 := arr[y0][x0]
			p01 := arr[y1][x0]
			p10 := arr[y0][x1]
			p11 := arr[y1][x1]

			// Calculate the bicubic interpolation
			newArr[i][j] = bicubic(x, y, p00, p01, p10, p11)
		}
	}

	return newArr
}

// Helper function to calculate the bicubic interpolation
func bicubic(x, y float64, p00, p01, p10, p11 float64) float64 {
	// Calculate the coefficients
	c00 := p00
	c01 := -0.5*p00 + 0.5*p01
	c10 := -0.5*p00 + 0.5*p10
	c11 := 0.25*p00 - 0.25*p01 - 0.25*p10 + 0.25*p11

	// Calculate the interpolated value
	val := c00 + c01*y + c10*x + c11*x*y
	return val
}
