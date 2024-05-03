# Chess Puzzles

These puzzles are intended to be difficult to calculate efficiently for a computer but be relatively easy for a human to do quickly. The intent is that two computer of similar capabilities will end up at the same solution given the state of a chess board. There are a few bugs but it mostly works.

The library finds large swings in evaulations on the board and labels them as "puzzle points". Then has the user calculate the best move in the position and uses that best move to form a puzzle key.

# Bugs

The evaluation of a position may fluctuate slightly which may cause positions close to the cutoff point be lost.

The difference in capabilities for computers can show up drastically when calculating "puzzle points"
