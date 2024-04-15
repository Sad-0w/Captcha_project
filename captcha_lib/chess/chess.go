package chess

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

// Do a POST request and return the result
func doPostRequest(postURL string, postContents []byte) (int, []byte, error) {
	// Initialize a client
	client := &http.Client{}
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(postContents))
	if err != nil {
		return 0, nil, err
	}

	req.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"User-Agent":   {"Mozilla/5.0 (Macintosh"},
	}

	// Make the POST request
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	// Extract the body contents
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	return resp.StatusCode, body, nil
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
func Test() {
	getChessPuzzles()
}

func getChessPuzzles() {
	// set up engine to use stockfish exe
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}
	defer eng.Close()
	// initialize uci with new game
	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}
	// have stockfish play speed chess against itself (10 msec per move)
	game := chess.NewGame()
	for game.Outcome() == chess.NoOutcome {
		cmdPos := uci.CmdPosition{Position: game.Position()}
		cmdGo := uci.CmdGo{MoveTime: time.Second / 100}
		if err := eng.Run(cmdPos, cmdGo); err != nil {
			panic(err)
		}
		move := eng.SearchResults().BestMove
		if err := game.Move(move); err != nil {
			panic(err)
		}
	}
	fmt.Println(game.String())
	fmt.Println(game.Position().Board().Draw())
	fmt.Printf("Game completed. %s by %s.\n", game.Outcome(), game.Method())
	fmt.Println(game.String())
	// Output:
	// 1.c4 c5 2.Nf3 e6 3.Nc3 Nc6 4.d4 cxd4 5.Nxd4 Nf6 6.a3 d5 7.cxd5 exd5 8.Bf4 Bc5 9.Ndb5 O-O 10.Nc7 d4 11.Na4 Be7 12.Nxa8 Bf5 13.g3 Qd5 14.f3 Rxa8 15.Bg2 Rd8 16.b4 Qe6 17.Nc5 Bxc5 18.bxc5 Nd5 19.O-O Nc3 20.Qd2 Nxe2+ 21.Kh1 d3 22.Bd6 Qd7 23.Rab1 h6 24.a4 Re8 25.g4 Bg6 26.a5 Ncd4 27.Qb4 Qe6 28.Qxb7 Nc2 29.Qxa7 Ne3 30.Rb8 Nxf1 31.Qb6 d2 32.Rxe8+ Qxe8 33.Qb3 Ne3 34.h3 Bc2 35.Qxc2 Nxc2 36.Kh2 d1=Q 37.h4 Qg1+ 38.Kh3 Ne1 39.h5 Qxg2+ 40.Kh4 Nxf3#  0-1

}
