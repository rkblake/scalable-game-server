package main

type Match struct {
	Code       string `json:"code"`
	NumPlayers int    `json:"numPlayers"`
	MaxPlayers int    `json:"maxPlayers"`
	Private    bool   `json:"private"`
}

type MatchInfo struct {
	NumMatches int     `json:"numMatches"`
	Matches    []Match `json:"matches"`
}

var matches = MatchInfo{Matches: []Match{}, NumMatches: 0}

func AddMatch(code string, maxPlayers int, private bool) {
	matches.NumMatches += 1
	matches.Matches = append(matches.Matches, Match{
		Code:       code,
		NumPlayers: 1,
		MaxPlayers: maxPlayers,
		Private:    private,
	})
}

func RemoveMatch(code string) {
	for i := 0; i < matches.NumMatches; i++ {
		if matches.Matches[i].Code == code {
			// NOTE: how memory safe is this? any leaks?
			matches.Matches[i] = matches.Matches[len(matches.Matches)-1]
			matches.Matches = matches.Matches[:len(matches.Matches)-1]
			return
		}
	}
}
