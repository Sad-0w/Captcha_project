package chess

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

// this is the cutoff point for a chess puzzle point
// it is measured in centipawns or 1/100 of a pawn (so 700 is 7 pawns)
// this variable can be changed to alter the difficulty of the chess puzzles
const CUTOFF = 700

// this is the run time the engine uses to perform a estimation of the position, it should be low
// currently it is 10 ms
const engRuntime = time.Second / 100

// this is the run time to calculate the solution to the puzzle, this will give the *best* move in the position
const solutionTime = time.Second * 10

// this is the number of puzzles to have the user solve
// increasing this will cause the computation to increase greatly
const PuzzleKeyLen = 2

// this function handles having the user find the solution to a chess puzzle
func promptUserInput(game *chess.Game, solution *chess.Move, retrieveKey bool) bool {
	if retrieveKey {
		for {
			fmt.Println(game.Position().Board().Draw())
			fmt.Println("it is ", game.Position().Turn().Name(), " to move")
			fmt.Println("Please enter the next best move in the format:")
			fmt.Println("piece to move location || location moved to")
			fmt.Println("example: h8g8 moves the piece at h8 to g8")
			fmt.Println("if you would like to skip this puzzle then type 'skip'")
			fmt.Print("Please enter move: ")
			var w1 string
			fmt.Scanln(&w1)
			fmt.Println()
			if strings.ToLower(w1) == "skip" {
				return false
			}
			if strings.ToLower(w1) == solution.String() {
				return true
			} else {
				fmt.Println("\nThat is not the correct solution")
			}
		}
	} else {
		for {
			fmt.Println(game.Position().Board().Draw())
			fmt.Println("it is ", game.Position().Turn().Name(), " to move")
			fmt.Println("Please enter the next best move in the format:")
			fmt.Println("piece to move location || location moved to")
			fmt.Println("example: h8g8 moves the piece at h8 to g8")
			fmt.Print("Please enter move: ")
			var w1 string
			fmt.Scanln(&w1)
			fmt.Println()
			if strings.ToLower(w1) == solution.String() {
				return true
			} else {
				fmt.Println("\nThat is not the correct solution")
			}
		}
	}
}

// same as Hashs but for byte strings
func Hashb(bs []byte, salt []byte) []byte {
	h := sha256.New()
	h.Write(bs)
	h.Write(salt)
	bsr := h.Sum(nil)
	return bsr
}

// same as HashNs but for byte strings
// requires N >= 0
func HashNb(bs []byte, N uint16, salt []byte) []byte {
	bsr := bs
	var h hash.Hash
	h = sha256.New()
	h.Write(bsr)
	h.Write(salt)
	bsr = h.Sum(nil)
	for i := uint16(1); i < N; i++ {
		h = sha256.New()
		h.Write(bsr)
		bsr = h.Sum(nil)
	}
	return bsr
}

// to export a function just capitalize the first letter
func GetPuzzleKey(pwd string, offsets []int) (string, []int) {
	bpwd := Hashb([]byte(pwd), nil)
	return getChessPuzzles(bpwd, engRuntime, offsets)
}

// function that takes in the byte string password, a timescale for estimations, and a skipped array (for recovering the key from the byte string)
func getChessPuzzles(pwd []byte, timeScale time.Duration, skip []int) (string, []int) {
	// set up engine to use stockfish exe
	salt := make([]byte, 16)
	key := HashNb(pwd, 12, salt)
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	defer eng.Close()
	// initialize uci with new game
	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	// create a seeded pseudorandom function to be used to generate chess moves
	Srand := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(key))))

	var result string
	// maintain state of the number of skipped puzzles for each index
	skipped := make([]int, PuzzleKeyLen)
	i := 0

	for i < PuzzleKeyLen {
		game := chess.NewGame()
		// randomly perform moves until the game is decided or the puzzle key is finished
		for game.Outcome() == chess.NoOutcome {
			// select a random move
			moves := game.ValidMoves()
			move := moves[Srand.Intn(len(moves))]
			game.Move(move)

			// have the engine perform a quick evaluation to see if anything interesting is happening
			cmdPos := uci.CmdPosition{Position: game.Position()}
			cmdGo := uci.CmdGo{MoveTime: timeScale}
			if err := eng.Run(cmdPos, cmdGo); err != nil {
				panic(err)
			}
			stat := eng.SearchResults()

			// compare the current state with the cutoff point for a "puzzle point"
			if math.Abs(float64(stat.Info.Score.CP)) > CUTOFF && i < PuzzleKeyLen {
				if skip != nil && skip[i] > 0 {
					skip[i]--
					continue
				}

				// have the engine evaluate the best move in the position
				cmdPos := uci.CmdPosition{Position: game.Position()}
				cmdGo := uci.CmdGo{MoveTime: solutionTime}
				if err := eng.Run(cmdPos, cmdGo); err != nil {
					panic(err)
				}
				solution_move := stat.BestMove

				// this next line is for testing purposes as it will display the solution
				// fmt.Println("Best move: ", solution_move)

				// prompt the user to find the best move
				guess := promptUserInput(game, solution_move, skip == nil)
				if guess {
					result += solution_move.String()
					i++
				} else {
					skipped[i]++
				}
			}
		}
	}
	return result, skipped
}
