package sudoku

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

/***

use case: K = sudoku.Test(*keystr, uint16(*N))

***/

const (
	N = 9
)

type Grid [N][N]int

//	func (g *Grid) Print() {
//		for i := 0; i < N; i++ {
//			if i%3 == 0 && i != 0 {
//				fmt.Println("-----------")
//			}
//			for j := 0; j < N; j++ {
//				if j%3 == 0 && j != 0 {
//					fmt.Print("|")
//				}
//				if g[i][j] == 0 {
//					print(" ")
//				} else {
//					fmt.Printf("%d", g[i][j])
//				}
//			}
//			fmt.Println()
//		}
//		fmt.Println()
//	}

func HashNs(s string, N uint16, salt []byte) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	h.Write(salt)
	bs := h.Sum(nil)
	for i := uint16(1); i < N; i++ {
		h = sha256.New()
		h.Write(bs)
		bs = h.Sum(nil)
	}
	return bs
}

// check if new number fits sudoku rule
func (g *Grid) isValid(row, col, num int) bool {
	blockRow, blockCol := (row/3)*3, (col/3)*3
	for i := 0; i < N; i++ {
		if g[row][i] == num || g[i][col] == num {
			return false
		}
		if g[blockRow+i/3][blockCol+i%3] == num {
			return false
		}
	}
	return true
}

// count puzzle solutions
func (g *Grid) countSolutions(count *int, row int, col int) {
	if row == N { // If row exceeds grid size, we've filled the grid correctly
		*count++
		return
	}

	nextRow, nextCol := row, col+1
	if nextCol == N {
		nextRow++
		nextCol = 0
	}

	if g[row][col] != 0 { // Skip filled cells
		g.countSolutions(count, nextRow, nextCol)
	} else {
		for num := 1; num <= N; num++ {
			if g.isValid(row, col, num) {
				g[row][col] = num
				g.countSolutions(count, nextRow, nextCol)
				g[row][col] = 0 // Unmake the move
			}
		}
	}

	if *count > 1 { // verify only one solution
		return
	}
}

// check if unique solution
func (g *Grid) ensureUniqueSolution() bool {
	var solutionCount int
	g.countSolutions(&solutionCount, 0, 0)
	return solutionCount == 1
}

// attempt to solve puzzle
func (g *Grid) Solve() bool {
	for row := 0; row < N; row++ {
		for col := 0; col < N; col++ {
			if g[row][col] == 0 {
				for num := 1; num <= N; num++ {
					if g.isValid(row, col, num) {
						g[row][col] = num
						if g.Solve() {
							return true
						}
						g[row][col] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

// generate puzzle
func (g *Grid) generator(seed int64) {

	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			g[i][j] = 0
		}
	}
	for i := 0; i < N*N; i++ {
		row := rng.Intn(N)
		col := rng.Intn(N)
		num := rng.Intn(N) + 1
		if g[row][col] == 0 && g.isValid(row, col, num) {
			g[row][col] = num
			copiedG := *g
			if !copiedG.Solve() {
				g[row][col] = 0
			}
		}

	}
	g.Solve()
	// g.Print()
	// remove numbers and check for uniqueness
	tries := 81
	for tries > 0 {
		row := rng.Intn(N)
		col := rng.Intn(N)
		if g[row][col] != 0 {
			backup := g[row][col]
			g[row][col] = 0
			copiedGrid := *g
			if !copiedGrid.ensureUniqueSolution() {
				g[row][col] = backup
			}
			tries--
		}
	}

}

func generateHashedPartialKey(key string, n uint16) []byte {

	salt := make([]byte, 16)
	key1 := HashNs(key, N, salt)
	return key1
}
func generateSeed(key string, n uint16) int64 {

	key1 := generateHashedPartialKey(key, n)

	reader := bytes.NewReader(key1)

	// Variable to hold the converted int64
	var hashedPartiaKey int64
	// Read the bytes into num
	err := binary.Read(reader, binary.LittleEndian, &hashedPartiaKey)
	if err != nil {
		fmt.Println("converting to num failed:", err)
	}
	// fmt.Println(hashedPartiaKey)
	return hashedPartiaKey
}
func generateHashedPuzzleKey(g Grid, key string, n uint16) ([]byte, [N * N]int, string) {

	key1 := generateSeed(key, n)
	g.generator(key1)

	// fmt.Println("Generated Sudoku Puzzle:")
	// g.Print()

	puzzle := [N * N]int{}
	// Copy elements from the 2D array to the 1D array
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			puzzle[i*N+j] = g[i][j]
		}
	}

	solution := g
	solution.Solve()
	//verify only one solution
	oneD := [N * N]int{}
	// Copy elements from the 2D array to the 1D array
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			oneD[i*N+j] = solution[i][j]
		}
	}

	var solutionStr string
	for _, value := range oneD {
		solutionStr += strconv.Itoa(value) // Convert each int to a string and append
	}
	// fmt.Println("solution:", solutionStr)
	salt := make([]byte, 16)
	hashedPuzzleKey := HashNs(solutionStr, n, salt)
	// fmt.Println("hashedPuzzleKey", hashedPuzzleKey)
	return hashedPuzzleKey, puzzle, solutionStr

}

// check if user answer is accepted
func validateSudoku(input string, solution string) bool {
	return input == solution
}

// customTheme extends the base theme provided by Fyne
type customTheme struct {
	fyne.Theme
}

// newCustomTheme creates a new theme based on the dark theme with overridden colors
func newCustomTheme() fyne.Theme {
	baseTheme := theme.DarkTheme() // Start with the dark theme
	return &customTheme{Theme: baseTheme}
}

// Color overrides the color settings for the theme
func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground: // Used by entry widgets for text
		return color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange color for text
	case theme.ColorNameDisabled:
		return color.RGBA{R: 180, G: 180, B: 180, A: 255}
	default:
		return t.Theme.Color(name, variant)
	}
}

func AcceptUserInput(initialGrid [N * N]int, solution string, resultChan chan<- bool) {
	a := app.New()
	w := a.NewWindow("SUDOKU PUZZLE")
	a.Settings().SetTheme(newCustomTheme())

	entries := make([]*widget.Entry, N*N)

	for i := range entries {
		entries[i] = widget.NewEntry()
		entries[i].Validator = nil
		if initialGrid[i] != 0 {
			entries[i].SetText(strconv.Itoa(initialGrid[i]))
			entries[i].Disable()
		} else {
			entries[i].SetPlaceHolder("")
		}
	}
	// Main container for the Sudoku grid
	blocks := container.NewGridWithColumns(3)

	// Adding each 3x3 subgrid
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			subGrid := container.NewGridWithColumns(3)
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					index := (i*3+k)*N + (j*3 + l)
					subGrid.Add(entries[index])
				}
			}

			paddedSubGrid := container.NewPadded(subGrid)
			blocks.Add(paddedSubGrid)
		}
	}

	submitButton := widget.NewButton("Submit", func() {
		var result string
		for _, entry := range entries {
			result += entry.Text
		}
		// fmt.Println("Current Grid State:", result)
		solved := validateSudoku(result, solution)

		resultChan <- solved
		if !solved {
			d := dialog.NewError(errors.New("Solve failed"), w)
			d.SetOnClosed(func() {
				w.Close()
			})
			d.Show()

		} else {
			info := dialog.NewInformation("Success", "Sudoku solved successfully!", w)
			info.SetOnClosed(func() {
				w.Close()
				a.Quit()
			})
			info.Show()
		}

	})

	w.SetContent(container.NewVBox(
		// grid,
		blocks,
		submitButton,
	))

	w.SetOnClosed(func() {
		w.Close()
		a.Quit()
	})

	w.Resize(fyne.NewSize(480, 430))
	w.ShowAndRun()
}

// generate final key
func combineTwoKeys(key string, n uint16) string {
	HashedPartialKey := generateHashedPartialKey(key, n)
	var g Grid
	HashedPuzzleKey, puzzle, solutionStr := generateHashedPuzzleKey(g, key, n)
	resultChan := make(chan bool, 1)
	AcceptUserInput(puzzle, solutionStr, resultChan)
	solved := <-resultChan
	if solved {
		fmt.Println("Sudoku solved successfully.")
	} else {
		fmt.Println("Sudoku solving failed.")
		os.Exit(-2)
	}

	EncryptionKey := append(HashedPartialKey, HashedPuzzleKey...)
	// fmt.Println("Key:", EncryptionKey)
	return string(EncryptionKey)
}

// main
func Test(key string, n uint16) string {

	return combineTwoKeys(key, n)
}
